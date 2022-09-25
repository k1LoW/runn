package runn

import (
	"fmt"

	goyaml "gopkg.in/yaml.v2"
)

type usingMappedSteps2 struct {
	Desc     string                 `yaml:"desc,omitempty"`
	Runners  map[string]interface{} `yaml:"runners,omitempty"`
	Vars     map[string]interface{} `yaml:"vars,omitempty"`
	Steps    goyaml.MapSlice        `yaml:"steps,omitempty"`
	Debug    bool                   `yaml:"debug,omitempty"`
	Interval string                 `yaml:"interval,omitempty"`
	If       string                 `yaml:"if,omitempty"`
	SkipTest bool                   `yaml:"skipTest,omitempty"`
}

func newMapped2() usingMappedSteps2 {
	return usingMappedSteps2{
		Runners: map[string]interface{}{},
		Vars:    map[string]interface{}{},
		Steps:   goyaml.MapSlice{},
	}
}

func unmarshalAsListedSteps2(b []byte, bk *book) error {
	var ok bool
	l := newListed()
	if err := goyaml.Unmarshal(b, &l); err != nil {
		return err
	}
	bk.useMap = false
	bk.desc = l.Desc
	bk.runners, ok = normalizeTo2(l.Runners).(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to normalizeTo2: %v", l.Runners)
	}
	bk.vars, ok = normalizeTo2(l.Vars).(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to normalizeTo2: %v", l.Vars)
	}
	bk.debug = l.Debug
	bk.intervalStr = l.Interval
	bk.ifCond = l.If
	bk.skipTest = l.SkipTest
	bk.rawSteps, ok = normalizeTo2(l.Steps).([]map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to normalizeTo2: %v", l.Steps)
	}
	return nil
}

func unmarshalAsMappedSteps2(b []byte, bk *book) error {
	var ok bool
	m := newMapped2()
	if err := goyaml.Unmarshal(b, &m); err != nil {
		return err
	}
	bk.useMap = true
	bk.desc = m.Desc
	bk.runners, ok = normalizeTo2(m.Runners).(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to normalizeTo2: %v", m.Runners)
	}
	bk.vars, ok = normalizeTo2(m.Vars).(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to normalizeTo2: %v", m.Vars)
	}
	bk.debug = m.Debug
	bk.intervalStr = m.Interval
	bk.ifCond = m.If
	bk.skipTest = m.SkipTest

	keys := map[string]struct{}{}
	for _, s := range m.Steps {
		v, ok := normalizeTo2(s.Value).(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to normalizeTo2: %v", s.Value)
		}
		bk.rawSteps = append(bk.rawSteps, v)
		var k string
		switch v := s.Key.(type) {
		case string:
			k = v
		case uint64:
			k = fmt.Sprintf("%d", v)
		default:
			k = fmt.Sprintf("%v", v)
		}
		bk.stepKeys = append(bk.stepKeys, k)
		if _, ok := keys[k]; ok {
			return fmt.Errorf("duplicate step keys: %s", k)
		}
		keys[k] = struct{}{}
	}
	return nil
}

// normalizeTo2 unmarshaled values
func normalizeTo2(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		res := make([]interface{}, len(v))
		for i, vv := range v {
			res[i] = normalizeTo2(vv)
		}
		return res
	case map[interface{}]interface{}:
		res := make(map[string]interface{})
		for k, vv := range v {
			res[fmt.Sprintf("%v", k)] = normalizeTo2(vv)
		}
		return res
	case map[string]interface{}:
		res := make(map[string]interface{})
		for k, vv := range v {
			res[k] = normalizeTo2(vv)
		}
		return res
	case []map[string]interface{}:
		res := make([]map[string]interface{}, len(v))
		for i, vv := range v {
			var ok bool
			res[i], ok = normalizeTo2(vv).(map[string]interface{})
			if !ok {
				return fmt.Errorf("failed to normalizeTo2: %v", vv)
			}
		}
		return res
	case goyaml.MapSlice:
		res := make(map[string]interface{})
		for _, i := range v {
			res[fmt.Sprintf("%v", i.Key)] = normalizeTo2(i.Value)
		}
		return res
	case int:
		if v < 0 {
			return int64(v)
		}
		return uint64(v)
	default:
		return v
	}
}
