package runn

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
	"github.com/samber/lo"
)

var varRep = regexp.MustCompile(`\{\{([^}]+)\}\}`)
var qRep = regexp.MustCompile(`\?.+$`)

type coverage struct {
	Specs []*specCoverage `json:"specs"`
}

type specCoverage struct {
	Key       string         `json:"key"`
	Coverages map[string]int `json:"coverages"`
}

func (o *operator) collectCoverage(ctx context.Context) (*coverage, error) {
	cov := &coverage{}
	// Collect coverage for openapi3
	for _, r := range o.httpRunners {
		ov, ok := r.validator.(*openApi3Validator)
		if !ok {
			continue
		}
		key := fmt.Sprintf("%s:%s", ov.doc.Info.Title, ov.doc.Info.Version)
		scov, ok := lo.Find(cov.Specs, func(scov *specCoverage) bool {
			return scov.Key == key
		})
		if !ok {
			scov = &specCoverage{
				Key:       key,
				Coverages: map[string]int{},
			}
			cov.Specs = append(cov.Specs, scov)
		}
		paths := map[*openapi3.PathItem]string{}
		for p, item := range ov.doc.Paths {
			paths[item] = p
			for m := range item.Operations() {
				mkey := fmt.Sprintf("%s %s", m, p)
				scov.Coverages[mkey] += 0
			}
		}
		for _, s := range o.steps {
			if s.httpRunner != r {
				continue
			}
		L:
			for p, m := range s.httpRequest {
				mm := m.(map[string]any)
				for mmm := range mm {
					method := strings.ToUpper(mmm)
					// Find path using openapi3 spec document (e.g. /v1/users/{id})
					i := ov.doc.Paths.Find(varRep.ReplaceAllString(qRep.ReplaceAllString(p, ""), "{x}"))
					if i == nil {
						// Find path using router (e.g. /v1/users/1)
						const host = "https://runn.test"
						for _, s := range ov.doc.Servers {
							s.URL = host
						}
						router, err := legacyrouter.NewRouter(ov.doc)
						if err != nil {
							return nil, err
						}
						req, err := http.NewRequest(method, host+p, nil)
						if err != nil {
							return nil, err
						}
						route, _, err := router.FindRoute(req)
						if err != nil {
							o.Warnf("%s %s was not matched in %s\n", method, p, key)
							continue
						}
						mkey := fmt.Sprintf("%s %s", method, route.Path)
						scov.Coverages[mkey]++
						continue
					}
					for m := range i.Operations() {
						if method == m {
							path, ok := paths[i]
							if !ok {
								return nil, fmt.Errorf("path not found in %s", p)
							}
							mkey := fmt.Sprintf("%s %s", method, path)
							scov.Coverages[mkey]++
							break L
						}
					}
					o.Warnf("%s %s was not matched in %s\n", method, p, key)
				}
			}
		}
	}

	// Collect coverage for protocol buffers
	for name, r := range o.grpcRunners {
		if err := r.resolveAllMethodsUsingProtos(ctx); err != nil {
			o.Warnf("%s was not resolved: %s\n", name, err)
			continue
		}
		for k := range r.mds {
			sm := strings.Split(k, "/")
			service := sm[0]
			method := sm[1]
			scov, ok := lo.Find(cov.Specs, func(scov *specCoverage) bool {
				return scov.Key == service
			})
			if !ok {
				scov = &specCoverage{
					Key:       service,
					Coverages: map[string]int{},
				}
				cov.Specs = append(cov.Specs, scov)
			}
			scov.Coverages[method] += 0
		}
		for _, s := range o.steps {
			if s.grpcRunner != r {
				continue
			}
			for k := range s.grpcRequest {
				sm := strings.Split(k, "/")
				service := sm[0]
				method := sm[1]
				scov, ok := lo.Find(cov.Specs, func(scov *specCoverage) bool {
					return scov.Key == service
				})
				if !ok {
					o.Warnf("%s/%s was not matched\n", service, method)
					continue
				}
				scov.Coverages[method]++
			}
		}
	}
	return cov, nil
}
