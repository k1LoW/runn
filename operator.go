package runbk

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
)

type Step struct {
	httpRunner  *httpRunner
	httpRequest map[string]interface{}
	dbRunner    *dbRunner
	dbQuery     map[string]interface{}
	testRunner  *testRunner
	testCond    string
}

type Store struct {
	Steps []map[string]interface{}
	Vars  map[string]string
}

type Operator struct {
	httpRunners map[string]*httpRunner
	dbRunners   map[string]*dbRunner
	steps       []*Step
	store       Store
}

func New(opts ...Option) (*Operator, error) {
	c := &book{Vars: map[string]string{}}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	o := &Operator{
		httpRunners: map[string]*httpRunner{},
		dbRunners:   map[string]*dbRunner{},
		store: Store{
			Steps: []map[string]interface{}{},
			Vars:  c.Vars,
		},
	}

	for k, v := range c.Runners {
		switch {
		case strings.Index(v, "https://") == 0 || strings.Index(v, "http://") == 0:
			hc, err := newHttpRunner(v, o)
			if err != nil {
				return nil, err
			}
			o.httpRunners[k] = hc
		default:
			dc, err := newDBRunner(v, o)
			if err != nil {
				return nil, err
			}
			o.dbRunners[k] = dc
		}
	}
	for k, v := range c.httpRunners {
		v.operator = o
		o.httpRunners[k] = v
	}
	for k, v := range c.dbRunners {
		v.operator = o
		o.dbRunners[k] = v
	}

	for i, s := range c.Steps {
		if len(s) != 1 {
			return nil, fmt.Errorf("invalid steps[%d]: %v", i, s)
		}
		if err := o.AppendStep(s); err != nil {
			return nil, err
		}
	}

	return o, nil
}

func (o *Operator) AppendStep(s map[string]interface{}) error {
	step := &Step{}
	for k, v := range s {
		if k == testRunnerKey {
			tr, err := newTestRunner(o)
			if err != nil {
				return err
			}
			step.testRunner = tr
			vv, ok := v.(string)
			if !ok {
				return fmt.Errorf("invalid test condition: %v", v)
			}
			step.testCond = vv
			continue
		}
		h, ok := o.httpRunners[k]
		if ok {
			step.httpRunner = h
			vv, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid http request: %v", v)
			}
			step.httpRequest = vv
			continue
		}
		db, ok := o.dbRunners[k]
		if ok {
			step.dbRunner = db
			vv, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid http request: %v", v)
			}
			step.dbQuery = vv
			continue
		}
		return fmt.Errorf("can not find client: %s", k)
	}
	o.steps = append(o.steps, step)
	return nil
}

func (o *Operator) Run(ctx context.Context) error {
	for i, s := range o.steps {
		switch {
		case s.httpRunner != nil && s.httpRequest != nil:
			e, err := o.expand(s.httpRequest)
			if err != nil {
				return err
			}
			r, ok := e.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid steps[%d]: %v", i, e)
			}
			req, err := o.parseHTTPRequest(r)
			if err != nil {
				return err
			}
			if err := s.httpRunner.Run(ctx, req); err != nil {
				return fmt.Errorf("http request failed on steps[%d]: %v", i, err)
			}
		case s.dbRunner != nil && s.dbQuery != nil:
			e, err := o.expand(s.dbQuery)
			if err != nil {
				return err
			}
			q, ok := e.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid steps[%d]: %v", i, e)
			}
			query, err := o.parseDBQuery(q)
			if err != nil {
				return fmt.Errorf("invalid steps[%d]: %v", i, q)
			}
			if err := s.dbRunner.Run(ctx, query); err != nil {
				return fmt.Errorf("db query failed on steps[%d]: %v", i, err)
			}
		case s.testRunner != nil && s.testCond != "":
			if err := s.testRunner.Run(ctx, s.testCond); err != nil {
				return fmt.Errorf("test failed on steps[%d]: %s", i, s.testCond)
			}
		default:
			return fmt.Errorf("invalid steps[%d]: %v", i, s)
		}
	}
	return nil
}

var expandRe = regexp.MustCompile(`"?{{\s*([^}]+)\s*}}"?`)

func (o *Operator) expand(in interface{}) (interface{}, error) {
	store := map[string]interface{}{
		"steps": o.store.Steps,
		"vars":  o.store.Vars,
	}
	b, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	var reperr error
	replacefunc := func(in string) string {
		if !strings.Contains(in, "{{") {
			return in
		}
		matches := expandRe.FindAllStringSubmatch(in, -1)
		oldnew := []string{}
		for _, m := range matches {
			o, err := expr.Eval(m[1], store)
			if err != nil {
				reperr = err
			}
			var s string
			switch v := o.(type) {
			case string:
				s = v
			case int64:
				s = strconv.Itoa(int(v))
			default:
				reperr = fmt.Errorf("invalid expand format: %v", o)
			}
			oldnew = append(oldnew, m[0], s)
		}
		rep := strings.NewReplacer(oldnew...)
		return rep.Replace(in)
	}
	e := expand.ReplaceYAML(string(b), replacefunc, true)
	if reperr != nil {
		return nil, reperr
	}
	var out interface{}
	if err := yaml.Unmarshal([]byte(e), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (o *Operator) parseHTTPRequest(v map[string]interface{}) (*httpRequest, error) {
	req := &httpRequest{
		headers: map[string]string{},
	}
	if len(v) != 1 {
		return nil, fmt.Errorf("invalid request: %v", v)
	}
	for k, vv := range v {
		req.path = k
		vvv, ok := vv.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid request: %v", v)
		}
		if len(vvv) != 1 {
			return nil, fmt.Errorf("invalid request: %v", v)
		}
		for kk, vvvv := range vvv {
			req.method = strings.ToUpper(kk)
			vvvvv, ok := vvvv.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid request: %v", v)
			}
			hm, ok := vvvvv["headers"]
			if ok {
				hm, ok := hm.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid request: %v", v)
				}
				for k, v := range hm {
					req.headers[k] = v.(string)
				}
			}
			bm, ok := vvvvv["body"]
			if ok {
				switch v := bm.(type) {
				case map[string]interface{}:
					if len(v) != 1 {
						return nil, fmt.Errorf("invalid request: %v", v)
					}
					for kkk, vvvvvv := range v {
						req.mediaType = kkk
						req.body = vvvvvv
						break
					}
				default:
					if v != nil {
						return nil, fmt.Errorf("invalid request: %v", v)
					}
				}
			}
		}

		break
	}
	return req, nil
}

func (o *Operator) parseDBQuery(v map[string]interface{}) (*dbQuery, error) {
	q := &dbQuery{}
	if len(v) != 1 {
		return nil, fmt.Errorf("invalid query: %v", v)
	}
	s, ok := v["query"]
	if !ok {
		return nil, fmt.Errorf("invalid query: %v", v)
	}
	stmt, ok := s.(string)
	if !ok {
		return nil, fmt.Errorf("invalid query: %v", v)
	}
	q.stmt = stmt
	return q, nil
}
