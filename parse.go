package runn

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/duration"
	"google.golang.org/grpc/metadata"
)

func parseHTTPRequest(v map[string]any) (*httpRequest, error) {
	v = trimDelimiter(v)
	req := &httpRequest{
		headers: http.Header{},
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
		vvv, ok := vv.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid request: %s", string(part))
		}
		if len(vvv) != 1 {
			return nil, fmt.Errorf("invalid request: %s", string(part))
		}
		for kk, vvvv := range vvv {
			req.method = strings.ToUpper(kk)
			vvvvv, ok := vvvv.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid request: %s", string(part))
			}
			hm, ok := vvvvv["headers"]
			if ok && hm != nil {
				hm, ok := hm.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid request: %s", string(part))
				}
				for k, v := range hm {
					switch v := v.(type) {
					case string:
						req.headers.Add(k, v)
					case []any:
						for _, vv := range v {
							req.headers.Add(k, vv.(string))
						}
					default:
						return nil, fmt.Errorf("invalid request: %s", string(part))
					}
				}
			}
			bm, ok := vvvvv["body"]
			if ok {
				switch v := bm.(type) {
				case map[string]any:
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
			um, ok := vvvvv["useCookie"]
			if ok {
				switch v := um.(type) {
				case bool:
					req.useCookie = &v
				default:
					if v != nil {
						return nil, fmt.Errorf("invalid request: %s", string(part))
					}
				}
			}
			tm, ok := vvvvv["trace"]
			if ok {
				switch v := tm.(type) {
				case bool:
					req.trace = &v
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

func parseDBQuery(v map[string]any) (*dbQuery, error) {
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
	tm, ok := v["trace"]
	if ok {
		switch v := tm.(type) {
		case bool:
			q.trace = &v
		default:
			if v != nil {
				return nil, fmt.Errorf("invalid query: %s", string(part))
			}
		}
	}
	return q, nil
}

func parseGrpcRequest(v map[string]any, expand func(any) (any, error)) (*grpcRequest, error) {
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
		vvv, ok := vv.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid request: %s", string(part))
		}
		hm, ok := vvv["headers"]
		if ok {
			hme, err := expand(hm)
			if err != nil {
				return nil, err
			}
			hm, ok := hme.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid request: %s", string(part))
			}
			for k, v := range hm {
				switch v := v.(type) {
				case string:
					req.headers.Append(k, v)
				case []any:
					for _, vv := range v {
						req.headers.Append(k, vv.(string))
					}
				default:
					return nil, fmt.Errorf("invalid request: %s", string(part))
				}
			}
		}
		tm, ok := vvv["timeout"]
		if ok {
			tme, err := expand(tm)
			if err != nil {
				return nil, err
			}
			tms, ok := tme.(string)
			if !ok {
				return nil, fmt.Errorf("invalid request: %s", string(part))
			}
			req.timeout, err = duration.Parse(tms)
			if err != nil {
				return nil, fmt.Errorf("invalid request: %s: %w", string(part), err)
			}
		}
		// `message:` and `messages:` expand at run time so not here
		mm, ok := vvv["message"]
		if ok {
			ms, ok := mm.(string)
			if ok {
				// Only for string, variable expansion is acceptable.
				mm, err = expand(ms)
				if err != nil {
					return nil, err
				}
			}
			mmm, ok := mm.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid request: %s", string(part))
			}
			req.messages = append(req.messages, &grpcMessage{
				op:     GRPCOpMessage,
				params: mmm,
			})
		} else {
			mm, ok := vvv["messages"]
			if ok {
				ms, ok := mm.(string)
				if ok {
					// Only for string, variable expansion is acceptable.
					mm, err = expand(ms)
					if err != nil {
						return nil, err
					}
				}
				mms, ok := mm.([]any)
				if !ok {
					return nil, fmt.Errorf("invalid request: %s", string(part))
				}
				for _, mm := range mms {
					// Only for string, variable expansion is acceptable.
					mms, ok := mm.(string)
					if ok {
						mm, err = expand(mms)
						if err != nil {
							return nil, err
						}
					}
					switch v := mm.(type) {
					case string:
						op := GRPCOp(v)
						if op != GRPCOpClose && op != GRPCOpReceive {
							return nil, fmt.Errorf("invalid request: %s", string(part))
						}
						req.messages = append(req.messages, &grpcMessage{
							op: op,
						})
					case map[string]any:
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
		tr, ok := vvv["trace"]
		if ok {
			switch v := tr.(type) {
			case bool:
				req.trace = &v
			default:
				if v != nil {
					return nil, fmt.Errorf("invalid request: %s", string(part))
				}
			}
		}
	}
	return req, nil
}

func parseCDPActions(v map[string]any, expand func(any) (any, error)) (CDPActions, error) {
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
	aa, ok := a.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid actions: %s", string(part))
	}
	for _, v := range aa {
		ca := CDPAction{
			Args: map[string]any{},
		}
		v, err := expand(v)
		if err != nil {
			return nil, err
		}
		switch vv := v.(type) {
		case string:
			if _, _, err := findCDPFn(vv); err != nil {
				return nil, fmt.Errorf("invalid action: %w", err)
			}
			ca.Fn = vv
		case map[string]any:
			if len(vv) != 1 {
				return nil, fmt.Errorf("invalid actions: %s", string(part))
			}
			for k, vvv := range vv {
				_, fn, err := findCDPFn(k)
				if err != nil {
					return nil, fmt.Errorf("invalid action: %w", err)
				}
				ca.Fn = k
				switch vvvv := vvv.(type) {
				case string:
					ca.Args[fn.Args[0].Key] = vvvv
				case map[string]any:
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

func parseSSHCommand(v map[string]any, expand func(any) (any, error)) (*sshCommand, error) {
	var err error
	part, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	v = trimDelimiter(v)
	vv, err := expand(v)
	if err != nil {
		return nil, err
	}
	vvv, ok := vv.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid command: %s", string(part))
	}
	sc := &sshCommand{}
	c, ok := vvv["command"]
	if !ok {
		return nil, fmt.Errorf("invalid command: %s", string(part))
	}
	sc.command, ok = c.(string)
	if !ok {
		return nil, fmt.Errorf("invalid command: %s", string(part))
	}
	return sc, nil
}

func parseServiceAndMethod(in string) (string, string, error) {
	splitted := strings.Split(strings.TrimPrefix(in, "/"), "/")
	if len(splitted) < 2 {
		return "", "", fmt.Errorf("invalid method: %s", in)
	}
	return strings.Join(splitted[:len(splitted)-1], "/"), splitted[len(splitted)-1], nil
}

func parseExecCommand(v map[string]any) (*execCommand, error) {
	v = trimDelimiter(v)
	c := &execCommand{}
	part, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(v) < 1 && len(v) > 3 {
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
	is, ok := v["stdin"]
	if ok {
		stdin, ok := is.(string)
		if !ok {
			return nil, fmt.Errorf("invalid stdin: %s", string(part))
		}
		c.stdin = stdin
	}
	ss, ok := v["shell"]
	if ok {
		sh, ok := ss.(string)
		if !ok {
			return nil, fmt.Errorf("invalid shell: %s", string(part))
		}
		c.shell = sh
	}
	return c, nil
}

func parseIncludeConfig(v any) (*includeConfig, error) {
	c := &includeConfig{vars: map[string]any{}}
	switch vv := v.(type) {
	case string:
		c.path = vv
		return c, nil
	case map[string]any:
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
			c.vars, ok = vars.(map[string]any)
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
		force, ok := vv["force"]
		if ok {
			c.force, ok = force.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid include condig: %v", v)
			}
		}
		return c, nil
	default:
		return nil, fmt.Errorf("invalid include condig: %v", v)
	}
}

func trimDelimiter(in map[string]any) map[string]any {
	for k, v := range in {
		switch vv := v.(type) {
		case string:
			in[k] = trimStringDelimiter(vv)
		case []any:
			for kk, vvv := range vv {
				switch vvvv := vvv.(type) {
				case string:
					vv[kk] = trimStringDelimiter(vvvv)
				}
			}
		case map[string]any:
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
