package runn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ajg/form"
	"github.com/goccy/go-json"
)

const (
	MediaTypeApplicationJSON           = "application/json"
	MediaTypeTextPlain                 = "text/plain"
	MediaTypeApplicationFormUrlencoded = "application/x-www-form-urlencoded"
	MediaTypeMultipartFormData         = "multipart/form-data"
)

var notFollowRedirectFn = func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

type httpRunner struct {
	name              string
	endpoint          *url.URL
	client            *http.Client
	handler           http.Handler
	operator          *operator
	validator         httpValidator
	multipartBoundary string
}

type httpRequest struct {
	path      string
	method    string
	headers   map[string]string
	mediaType string
	body      interface{}

	multipartWriter   *multipart.Writer
	multipartBoundary string
	// operator.root
	root string
}

func newHTTPRunner(name, endpoint string) (*httpRunner, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &httpRunner{
		name:     name,
		endpoint: u,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
		validator: newNopValidator(),
	}, nil
}

func newHTTPRunnerWithHandler(name string, h http.Handler) (*httpRunner, error) {
	return &httpRunner{
		name:      name,
		handler:   h,
		validator: newNopValidator(),
	}, nil
}

func (r *httpRequest) validate() error {
	switch r.method {
	case http.MethodPost, http.MethodPatch:
		if r.mediaType == "" {
			return fmt.Errorf("%s method requires mediaType", r.method)
		}
		if r.body == nil {
			return fmt.Errorf("%s method requires body", r.method)
		}
	}
	if r.isMultipartFormDataMediaType() {
		return nil
	}
	switch r.mediaType {
	case MediaTypeApplicationJSON, MediaTypeTextPlain, MediaTypeApplicationFormUrlencoded, "":
	default:
		return fmt.Errorf("unsupported mediaType: %s", r.mediaType)
	}
	return nil
}

func (r *httpRequest) encodeBody() (io.Reader, int64, error) {
	if r.body == nil {
		return nil, 0, nil
	}
	if r.isMultipartFormDataMediaType() {
		return r.encodeMultipart()
	}
	switch r.mediaType {
	case MediaTypeApplicationJSON:
		b, err := json.Marshal(r.body)
		if err != nil {
			return nil, 0, err
		}
		return bytes.NewBuffer(b), int64(len(b)), nil
	case MediaTypeApplicationFormUrlencoded:
		values, ok := r.body.(map[string]interface{})
		if !ok {
			return nil, 0, fmt.Errorf("invalid body: %v", r.body)
		}
		buf := new(bytes.Buffer)
		if err := form.NewEncoder(buf).Encode(values); err != nil {
			return nil, 0, err
		}
		return buf, int64(buf.Len()), nil
	case MediaTypeTextPlain:
		s, ok := r.body.(string)
		if !ok {
			return nil, 0, fmt.Errorf("invalid body: %v", r.body)
		}
		return strings.NewReader(s), int64(len(s)), nil
	default:
		return nil, 0, fmt.Errorf("unsupported mediaType: %s", r.mediaType)
	}
}

func (r httpRequest) isMultipartFormDataMediaType() bool {
	if r.mediaType == MediaTypeMultipartFormData {
		return true
	}
	return strings.HasPrefix(r.mediaType, MediaTypeMultipartFormData+"; boundary=")
}

func (r *httpRequest) encodeMultipart() (io.Reader, int64, error) {
	quoteEscaper := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	if r.multipartBoundary != "" {
		_ = mw.SetBoundary(r.multipartBoundary)
	}
	values := make([]map[string]interface{}, 0)
	switch v := r.body.(type) {
	case []interface{}:
		for _, vv := range v {
			rv, ok := vv.(map[string]interface{})
			if !ok {
				return nil, 0, fmt.Errorf("invalid body: %v", r.body)
			}
			values = append(values, rv)
		}
	case map[string]interface{}:
		for kk, vv := range v {
			if is, ok := vv.([]interface{}); ok {
				for _, vvv := range is {
					content := map[string]interface{}{
						kk: vvv,
					}
					values = append(values, content)
				}
			} else {
				content := map[string]interface{}{
					kk: vv,
				}
				values = append(values, content)
			}
		}
	default:
		return nil, 0, fmt.Errorf("invalid body: %v", r.body)
	}
	for _, value := range values {
		for fieldName, ifileName := range value {
			fileName, ok := ifileName.(string)
			if !ok {
				return nil, 0, fmt.Errorf("invalid body: %v", r.body)
			}
			b, err := os.ReadFile(filepath.Join(r.root, fileName))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, 0, err
			}
			h := make(textproto.MIMEHeader)
			if errors.Is(err, os.ErrNotExist) {
				b = []byte(fileName)
				h.Set("Content-Disposition",
					fmt.Sprintf(`form-data; name="%s"`, quoteEscaper.Replace(fieldName)))
			} else {
				h.Set("Content-Disposition",
					fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
						quoteEscaper.Replace(fieldName), quoteEscaper.Replace(filepath.Base(fileName))))
				h.Set("Content-Type", http.DetectContentType(b))
			}
			fw, err := mw.CreatePart(h)
			if err != nil {
				return nil, 0, err
			}
			if _, err = io.Copy(fw, bytes.NewReader(b)); err != nil {
				return nil, 0, err
			}
		}
	}
	// for Content-Type multipart/form-data with this Writer's Boundary
	r.multipartWriter = mw
	return buf, int64(buf.Len()), mw.Close()
}

func (r *httpRequest) setContentTypeHeader(req *http.Request) {
	if r.mediaType == MediaTypeMultipartFormData {
		req.Header.Set("Content-Type", r.multipartWriter.FormDataContentType())
	} else if r.mediaType != "" {
		req.Header.Set("Content-Type", r.mediaType)
	}
}

func (rnr *httpRunner) Run(ctx context.Context, r *httpRequest) error {
	r.multipartBoundary = rnr.multipartBoundary
	r.root = rnr.operator.root
	reqBody, _, err := r.encodeBody()
	if err != nil {
		return err
	}

	var (
		req *http.Request
		res *http.Response
	)
	switch {
	case rnr.client != nil:
		u, err := mergeURL(rnr.endpoint, r.path)
		if err != nil {
			return err
		}
		req, err = http.NewRequestWithContext(ctx, r.method, u.String(), reqBody)
		if err != nil {
			return err
		}
		r.setContentTypeHeader(req)
		for k, v := range r.headers {
			req.Header.Set(k, v)
		}

		rnr.operator.capturers.captureHTTPRequest(rnr.name, req)

		if err := rnr.validator.ValidateRequest(ctx, req); err != nil {
			return err
		}
		// reset Request.Body
		reqBody, cl, err := r.encodeBody()
		if err != nil {
			return err
		}
		rc, ok := reqBody.(io.ReadCloser)
		if !ok && reqBody != nil {
			rc = io.NopCloser(reqBody)
		}
		req.Body = rc
		req.ContentLength = cl

		r.setContentTypeHeader(req)
		for k, v := range r.headers {
			req.Header.Set(k, v)
		}

		res, err = rnr.client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
	case rnr.handler != nil:
		req = httptest.NewRequest(r.method, r.path, reqBody)
		if r.mediaType != "" {
			req.Header.Set("Content-Type", r.mediaType)
		}
		for k, v := range r.headers {
			req.Header.Set(k, v)
		}

		rnr.operator.capturers.captureHTTPRequest(rnr.name, req)

		if err := rnr.validator.ValidateRequest(ctx, req); err != nil {
			return err
		}
		w := httptest.NewRecorder()
		rnr.handler.ServeHTTP(w, req)
		res = w.Result()
		defer res.Body.Close()
	default:
		return fmt.Errorf("invalid http runner: %s", rnr.name)
	}

	rnr.operator.capturers.captureHTTPResponse(rnr.name, res)

	if err := rnr.validator.ValidateResponse(ctx, req, res); err != nil {
		var target *UnsupportedError
		if errors.As(err, &target) {
			rnr.operator.Debugf("Skip validate response due to unsupported format: %s", err.Error())
		} else {
			return err
		}
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	d := map[string]interface{}{}
	d["status"] = res.StatusCode
	if strings.Contains(res.Header.Get("Content-Type"), "json") && len(resBody) > 0 {
		var b interface{}
		if err := json.Unmarshal(resBody, &b); err != nil {
			return err
		}
		d["body"] = b
	} else {
		d["body"] = nil
	}
	d["rawBody"] = string(resBody)
	d["headers"] = res.Header

	rnr.operator.record(map[string]interface{}{
		"res": d,
	})

	return nil
}

func mergeURL(u *url.URL, p string) (*url.URL, error) {
	if !strings.HasPrefix(p, "/") {
		return nil, fmt.Errorf("invalid path: %s", p)
	}
	m, err := url.Parse(u.String())
	if err != nil {
		return nil, err
	}
	a, err := url.Parse(p)
	if err != nil {
		return nil, err
	}
	m.Path = path.Join(m.Path, a.Path)
	q := u.Query()
	for k, vs := range a.Query() {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	m.RawQuery = q.Encode()

	return m, nil
}
