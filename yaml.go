package runn

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/yaml.v2"
)

// CreateHTTPStepMapSlice creates yaml.MapSlice from *http.Request.
func CreateHTTPStepMapSlice(key string, req *http.Request) (yaml.MapSlice, error) {
	endpoint := req.URL.Path
	if req.URL.RawQuery != "" {
		endpoint = fmt.Sprintf("%s?%s", endpoint, req.URL.RawQuery)
	}
	if endpoint == "" {
		endpoint = "/"
	}

	hb := yaml.MapSlice{}
	// headers
	contentType := req.Header.Get("Content-Type")
	h := map[string]string{}
	for k, v := range req.Header {
		if k == "Content-Type" || k == "Host" {
			continue
		}
		h[k] = v[0]
	}
	if len(h) > 0 {
		hb = append(hb, yaml.MapItem{
			Key:   "headers",
			Value: h,
		})
	}

	// body
	var bd yaml.MapSlice
	var (
		save io.ReadCloser
		err  error
	)
	save, req.Body, err = drainBody(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to drainBody: %w", err)
	}
	switch {
	case save == http.NoBody || save == nil:
		if contentType == "" {
			bd = nil
		} else {
			bd = yaml.MapSlice{
				{Key: contentType, Value: nil},
			}
		}
	case strings.Contains(contentType, "json"):
		var v any
		if err := json.NewDecoder(save).Decode(&v); err != nil {
			return nil, fmt.Errorf("failed to decode: %w", err)
		}
		bd = yaml.MapSlice{
			{Key: contentType, Value: v},
		}
	case contentType == MediaTypeApplicationFormUrlencoded:
		b, err := io.ReadAll(save)
		if err != nil {
			return nil, fmt.Errorf("failed to io.ReadAll: %w", err)
		}
		vs, err := url.ParseQuery(string(b))
		if err != nil {
			return nil, fmt.Errorf("failed to url.ParseQuery: %w", err)
		}
		f := map[string]any{}
		for k, v := range vs {
			if len(v) == 1 {
				f[k] = v[0]
				continue
			}
			f[k] = v
		}
		bd = yaml.MapSlice{
			{Key: contentType, Value: f},
		}
	case strings.Contains(contentType, MediaTypeMultipartFormData):
		f := map[string]any{}
		mr, err := req.MultipartReader()
		if err != nil {
			return nil, err
		}
		for {
			part, err := mr.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			k := part.FormName()
			contentType := part.Header.Get("Content-Type")
			if contentType == "" {
				b, err := io.ReadAll(part)
				if err != nil {
					return nil, err
				}
				f[k] = string(b)
			} else {
				if part.FileName() != "" {
					if fn, ok := f[k]; ok {
						switch v := fn.(type) {
						case string:
							f[k] = []string{
								v,
								part.FileName(),
							}
						case []string:
							v = append(v, part.FileName())
							f[k] = v
						}
					} else {
						f[k] = part.FileName()
					}
				} else {
					exts, err := mime.ExtensionsByType(contentType)
					if err != nil {
						f[k] = "file"
					} else {
						f[k] = fmt.Sprintf("file%s", exts[0])
					}
				}
			}
		}
		req.Body = save

		bd = yaml.MapSlice{
			{Key: MediaTypeMultipartFormData, Value: f},
		}
	default:
		// case contentType == runn.MediaTypeTextPlain:
		b, err := io.ReadAll(save)
		if err != nil {
			return nil, fmt.Errorf("failed to io.ReadAll: %w", err)
		}
		bd = yaml.MapSlice{
			{Key: contentType, Value: string(b)},
		}
	}
	if len(bd) == 0 {
		hb = append(hb, yaml.MapItem{
			Key:   "body",
			Value: nil,
		})
	} else {
		hb = append(hb, yaml.MapItem{
			Key:   "body",
			Value: bd,
		})
	}

	m := yaml.MapItem{Key: strings.ToLower(req.Method), Value: nil}
	if len(hb) > 0 {
		m = yaml.MapItem{Key: strings.ToLower(req.Method), Value: hb}
	}

	step := yaml.MapSlice{
		{Key: key, Value: yaml.MapSlice{
			{Key: endpoint, Value: yaml.MapSlice{
				m,
			}},
		}},
	}

	return step, nil
}

// copy from net/http/httputil.
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == nil || b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
