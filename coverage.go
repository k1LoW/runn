package runn

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/pb33f/libopenapi-validator/paths"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/samber/lo"
)

var varRep = regexp.MustCompile(`\{\{([^}]+)\}\}`)
var qRep = regexp.MustCompile(`\?.+$`)

// Coverage is a coverage of runbooks.
type Coverage struct {
	Specs []*SpecCoverage `json:"specs"`
}

// SpecCoverage is a coverage of spec (e.g. OpenAPI Document, servive of protocol buffers).
type SpecCoverage struct {
	Key       string         `json:"key"`
	Coverages map[string]int `json:"coverages"`
}

func (o *operator) collectCoverage(ctx context.Context) (*Coverage, error) {
	cov := &Coverage{}
	// Collect coverage for openapi3
	for name, r := range o.httpRunners {
		ov, ok := r.validator.(*openAPI3Validator)
		if !ok {
			o.Debugf("%s does not have openapi3 spec document (%s)\n", name, o.bookPath)
			continue
		}
		doc := ov.doc
		v3m, err := doc.BuildV3Model()
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("%s:%s", v3m.Model.Info.Title, v3m.Model.Info.Version)
		scov, ok := lo.Find(cov.Specs, func(scov *SpecCoverage) bool {
			return scov.Key == key
		})
		if !ok {
			scov = &SpecCoverage{
				Key:       key,
				Coverages: map[string]int{},
			}
			cov.Specs = append(cov.Specs, scov)
		}
		pathm := map[*v3.PathItem]string{}
		for p := range orderedmap.Iterate(ctx, v3m.Model.Paths.PathItems) {
			pathm[p.Value()] = p.Key()
			for op := range orderedmap.Iterate(ctx, p.Value().GetOperations()) {
				mkey := fmt.Sprintf("%s %s", strings.ToUpper(op.Key()), p.Key())
				scov.Coverages[mkey] += 0
			}
		}
		regexCache := &sync.Map{}
	L:
		for _, s := range o.steps {
			if s.httpRunner != r {
				continue
			}
			for p, m := range s.httpRequest {
				mm, ok := m.(map[string]any)
				if !ok || len(mm) == 0 {
					continue L
				}
				var method string
				for mmm := range mm {
					method = strings.ToUpper(mmm)
					rp := varRep.ReplaceAllString(qRep.ReplaceAllString(p, ""), "{x}")
					req, err := http.NewRequest(method, strings.TrimSuffix(r.endpoint.String(), "/")+rp, nil)
					if err != nil {
						return nil, err
					}
					_, errs, pathValue := paths.FindPath(req, &v3m.Model, regexCache)
					if len(errs) > 0 {
						fmt.Println(req.URL.Path, errs)
						o.Debugf("%s %s was not matched in %s (%s)\n", method, p, key, o.bookPath)
						continue L
					}
					mkey := fmt.Sprintf("%s %s", method, pathValue)
					scov.Coverages[mkey]++
					continue L
				}
				o.Debugf("%s %s was not matched in %s (%s)\n", method, p, key, o.bookPath)
			}
		}
	}

	// Collect coverage for protocol buffers
	for name, r := range o.grpcRunners {
		if len(r.importPaths) > 0 || len(r.protos) > 0 || len(r.bufDirs) > 0 || len(r.bufLocks) > 0 || len(r.bufConfigs) > 0 || len(r.bufModules) > 0 {
			if err := r.resolveAllMethodsUsingProtos(ctx); err != nil {
				o.Debugf("%s was not resolved: %s (%s)\n", name, err, o.bookPath)
				continue
			}
		}
		if len(r.mds) == 0 {
			// Fallback to reflection
			if err := r.connectAndResolve(ctx, o); err != nil {
				o.Debugf("%s was not resolved: %s (%s)\n", name, err, o.bookPath)
				continue
			}
			if err := r.Close(); err != nil {
				return nil, err
			}
		}
		for k := range r.mds {
			sm := strings.Split(k, "/")
			service := sm[0]
			method := sm[1]
			scov, ok := lo.Find(cov.Specs, func(scov *SpecCoverage) bool {
				return scov.Key == service
			})
			if !ok {
				scov = &SpecCoverage{
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
				scov, ok := lo.Find(cov.Specs, func(scov *SpecCoverage) bool {
					return scov.Key == service
				})
				if !ok {
					o.Debugf("%s/%s was not matched (%s)\n", service, method, o.bookPath)
					continue
				}
				scov.Coverages[method]++
			}
		}
	}

	sort.Slice(cov.Specs, func(i, j int) bool {
		return cov.Specs[i].Key < cov.Specs[j].Key
	})

	return cov, nil
}
