package runn

import (
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

func (o *operator) collectCoverage() (*coverage, error) {
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
			if s.httpRunner == r {
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
									panic(fmt.Errorf("path not found in %s", p))
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
	}
	return cov, nil
}
