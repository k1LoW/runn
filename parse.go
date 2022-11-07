package runn

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/duration"
	"google.golang.org/grpc/metadata"
)

func parseHTTPRequest(v map[string]interface{}) (*httpRequest, error) {
	v = trimDelimiter(v)
	req := &httpRequest{
		headers: map[string]string{},
	}
	part, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(v) != 1 {
		return nil, fmt.Errorf("invalid request: %s", string(part))
	}
	for k, vv := range v {
		req.path = k
		vvv, ok := vv.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid request: %s", string(part))
		}
		if len(vvv) != 1 {
			return nil, fmt.Errorf("invalid request: %s", string(part))
		}
		for kk, vvvv := range vvv {
			req.method = strings.ToUpper(kk)
			vvvvv, ok := vvvv.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid request: %s", string(part))
			}
			hm, ok := vvvvv["headers"]
			if ok {
				hm, ok := hm.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid request: %s", string(part))
				}
				for k, v := range hm {
					req.headers[k], ok = v.(string)
					if !ok {
						return nil, fmt.Errorf("invalid request: %s", string(part))
					}
				}
			}
			bm, ok := vvvvv["body"]
			if ok {
				switch v := bm.(type) {
				case map[string]interface{}:
					if len(v) != 1 {
						return nil, fmt.Errorf("invalid request: %s", string(part))
					}
					for kkk, vvvvvv := range v {
						req.mediaType = kkk
						req.body = vvvvvv
						break
					}
				default:
					if v != nil {
						return nil, fmt.Errorf("invalid request: %s", string(part))
					}
				}
			}
		}

		break
	}
	if err := req.validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func parseDBQuery(v map[string]interface{}) (*dbQuery, error) {
	q := &dbQuery{}
	part, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(v) != 1 {
		return nil, fmt.Errorf("invalid query: %s", string(part))
	}
	s, ok := v["query"]
	if !ok {
		return nil, fmt.Errorf("invalid query: %s", string(part))
	}
	stmt, ok := s.(string)
	if !ok || strings.Trim(stmt, " ") == "" {
		return nil, fmt.Errorf("invalid query: %s", string(part))
	}
	q.stmt = strings.Trim(stmt, " \n")
	return q, nil
}

func parseGrpcRequest(v map[string]interface{}, expand func(interface{}) (interface{}, error)) (*grpcRequest, error) {
	v = trimDelimiter(v)
	req := &grpcRequest{
		headers: metadata.MD{},
	}
	part, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(v) != 1 {
		return nil, fmt.Errorf("invalid request: %s", string(part))
	}
	for k, vv := range v {
		pe, err := expand(k)
		if err != nil {
			return nil, err
		}
		svc, mth, err := parseServiceAndMethod(pe.(string))
		if err != nil {
			return nil, err
		}
		req.service = svc
		req.method = mth
		vvv, ok := vv.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid request: %s", string(part))
		}
		hm, ok := vvv["headers"]
		if ok {
			hme, err := expand(hm)
			if err != nil {
				return nil, err
			}
			hm, ok := hme.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid request: %s", string(part))
			}
			for k, v := range hm {
				req.headers.Append(k, v.(string))
			}
		}
		// `message:` and `messages:` expand at run time so not here
		mm, ok := vvv["message"]
		if ok {
			mm, ok := mm.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid request: %s", string(part))
			}
			req.messages = append(req.messages, &grpcMessage{
				op:     GRPCOpMessage,
				params: mm,
			})
		} else {
			mms, ok := vvv["messages"]
			if ok {
				mms, ok := mms.([]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid request: %s", string(part))
				}
				for _, mm := range mms {
					switch v := mm.(type) {
					case string:
						op := GRPCOp(v)
						if op != GRPCOpClose && op != GRPCOpReceive {
							return nil, fmt.Errorf("invalid request: %s", string(part))
						}
						req.messages = append(req.messages, &grpcMessage{
							op: op,
						})
					case map[string]interface{}:
						req.messages = append(req.messages, &grpcMessage{
							op:     GRPCOpMessage,
							params: v,
						})
					default:
						return nil, fmt.Errorf("invalid request: %s", string(part))
					}
				}
			}
		}
	}
	return req, nil
}

func parseCDPActions(v map[string]interface{}, expand func(interface{}) (interface{}, error)) (CDPActions, error) {
	v = trimDelimiter(v)
	cas := CDPActions{}
	part, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(v) != 1 {
		return nil, fmt.Errorf("invalid actions: %s", string(part))
	}
	a, ok := v["actions"]
	if !ok {
		return nil, fmt.Errorf("invalid actions: %s", string(part))
	}
	aa, ok := a.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid actions: %s", string(part))
	}
	for _, v := range aa {
		ca := CDPAction{
			Args: map[string]interface{}{},
		}
		switch vv := v.(type) {
		case string:
			if _, ok := CDPFnMap[vv]; !ok {
				return nil, fmt.Errorf("invalid action: %s", vv)
			}
			ca.Fn = vv
		case map[string]interface{}:
			if len(vv) != 1 {
				return nil, fmt.Errorf("invalid actions: %s", string(part))
			}
			for k, vvv := range vv {
				fn, ok := CDPFnMap[k]
				if !ok {
					return nil, fmt.Errorf("invalid action: %s", k)
				}
				ca.Fn = k
				switch vvvv := vvv.(type) {
				case string:
					ca.Args[fn.Args[0].key] = vvvv
				case map[string]interface{}:
					ca.Args = vvvv
				default:
					return nil, fmt.Errorf("invalid action args: %s(%v)", k, vvv)
				}
			}
		}
		cas = append(cas, ca)
	}
	return cas, nil
}

func parseServiceAndMethod(in string) (string, string, error) {
	splitted := strings.Split(strings.TrimPrefix(in, "/"), "/")
	if len(splitted) < 2 {
		return "", "", fmt.Errorf("invalid method: %s", in)
	}
	return strings.Join(splitted[:len(splitted)-1], "/"), splitted[len(splitted)-1], nil
}

func parseExecCommand(v map[string]interface{}) (*execCommand, error) {
	v = trimDelimiter(v)
	c := &execCommand{}
	part, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(v) != 1 && len(v) != 2 {
		return nil, fmt.Errorf("invalid command: %s", string(part))
	}
	cs, ok := v["command"]
	if !ok {
		return nil, fmt.Errorf("invalid command: %s", string(part))
	}
	command, ok := cs.(string)
	if !ok || strings.Trim(command, " ") == "" {
		return nil, fmt.Errorf("invalid command: %s", string(part))
	}
	c.command = strings.Trim(command, " \n")
	ss, ok := v["stdin"]
	if !ok {
		return c, nil
	}
	stdin, ok := ss.(string)
	if !ok {
		return nil, fmt.Errorf("invalid stdin: %s", string(part))
	}
	c.stdin = stdin
	return c, nil
}

func parseIncludeConfig(v interface{}) (*includeConfig, error) {
	c := &includeConfig{vars: map[string]interface{}{}}
	switch vv := v.(type) {
	case string:
		c.path = vv
		return c, nil
	case map[string]interface{}:
		path, ok := vv["path"]
		if !ok {
			return nil, fmt.Errorf("invalid include condig: %v", v)
		}
		c.path, ok = path.(string)
		if !ok {
			return nil, fmt.Errorf("invalid include condig: %v", v)
		}
		vars, ok := vv["vars"]
		if ok {
			c.vars, ok = vars.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid include condig: %v", v)
			}
		}
		skip, ok := vv["skipTest"]
		if ok {
			c.skipTest, ok = skip.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid include condig: %v", v)
			}
		}
		return c, nil
	default:
		return nil, fmt.Errorf("invalid include condig: %v", v)
	}
}

func trimDelimiter(in map[string]interface{}) map[string]interface{} {
	for k, v := range in {
		switch vv := v.(type) {
		case string:
			in[k] = trimStringDelimiter(vv)
		case []interface{}:
			for kk, vvv := range vv {
				switch vvvv := vvv.(type) {
				case string:
					vv[kk] = trimStringDelimiter(vvvv)
				}
			}
		case map[string]interface{}:
			in[k] = trimDelimiter(vv)
		}
	}
	return in
}

func trimStringDelimiter(in string) string {
L:
	for {
		switch {
		case strings.HasPrefix(in, "'") && strings.HasSuffix(in, "'"):
			in = strings.TrimSuffix(strings.TrimPrefix(in, "'"), "'")
		case strings.HasPrefix(in, "\"") && strings.HasSuffix(in, "\""):
			in = strings.Replace(strings.TrimSuffix(strings.TrimPrefix(in, "\""), "\""), "\\\"", "\"", -1)
		default:
			break L
		}
	}
	return in
}

var numOnlyRe = regexp.MustCompile(`^[0-9\.]+$`)

func parseDuration(v string) (time.Duration, error) {
	const defaultUnit = "sec"
	if numOnlyRe.MatchString(v) {
		v += defaultUnit
	}
	return duration.Parse(v)
}
