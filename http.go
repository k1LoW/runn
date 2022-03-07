package runbk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

const (
	MediaTypeApplicationJSON = "application/json"
)

type httpRunner struct {
	endpoint *url.URL
	client   *http.Client
	operator *operator
}

type httpRequest struct {
	path      string
	method    string
	headers   map[string]string
	mediaType string
	body      interface{}
}

func newHttpRunner(endpoint string, o *operator) (*httpRunner, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &httpRunner{
		endpoint: u,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
		operator: o,
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
	return nil
}

func (r *httpRequest) encodeBody() (io.Reader, error) {
	if r.body == nil {
		return nil, nil
	}
	switch r.mediaType {
	case MediaTypeApplicationJSON:
		b, err := json.Marshal(r.body)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(b), nil
	default:
		return nil, fmt.Errorf("unsupported mediaType: %s", r.mediaType)
	}
}

func (c *httpRunner) Run(ctx context.Context, r *httpRequest) error {
	u, err := mergeURL(c.endpoint, r.path)
	if err != nil {
		return err
	}
	reqBody, err := r.encodeBody()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, r.method, u.String(), reqBody)
	if err != nil {
		return err
	}
	if r.mediaType != "" {
		req.Header.Set("Content-Type", r.mediaType)
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	d := map[string]interface{}{}
	d["status"] = res.StatusCode

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if strings.Contains(res.Header.Get("Content-Type"), "json") {
		var b interface{}
		if err := json.Unmarshal(resBody, &b); err != nil {
			return err
		}
		d["body"] = b
	} else {
		d["rawBody"] = string(resBody)
	}

	d["headers"] = res.Header

	c.operator.store.steps = append(c.operator.store.steps, map[string]interface{}{
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
