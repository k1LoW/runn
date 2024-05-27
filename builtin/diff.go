package builtin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/itchyny/gojq"
)

func Diff(x, y any, ignorePaths ...string) string {
	d, err := diff(x, y, ignorePaths...)
	if err != nil {
		panic(err)
	}

	return d
}

func diff(x, y any, ignorePaths ...string) (string, error) {
	// normalize values
	bx, err := json.Marshal(x)
	if err != nil {
		return "", err
	}
	by, err := json.Marshal(y)
	if err != nil {
		return "", err
	}

	var vx any
	if err := json.Unmarshal(bx, &vx); err != nil {
		return "", err
	}
	var vy any
	if err := json.Unmarshal(by, &vy); err != nil {
		return "", err
	}

	if len(ignorePaths) > 0 {
		query, err := buildIgnoreTransformJqQuery(ignorePaths)
		if err != nil {
			return "", err
		}

		code, err := gojq.Compile(query)
		if err != nil {
			return "", fmt.Errorf("diff ignorePaths query compile error: %w", err)
		}

		if v, err := applyJqTransformQueryCompiled(code, vx); err != nil {
			return "", fmt.Errorf("applying diff ignorePaths error: %w", err)
		} else {
			vx = v
		}

		if v, err := applyJqTransformQueryCompiled(code, vy); err != nil {
			return "", fmt.Errorf("applying diff ignorePaths error: %w", err)
		} else {
			vy = v
		}
	}

	return cmp.Diff(vx, vy), nil
}

func buildIgnoreTransformJqQuery(ignorePaths []string) (*gojq.Query, error) {
	qb := strings.Builder{}
	qb.WriteString("delpaths([")
	for i, pathExpr := range ignorePaths {
		if i > 0 {
			qb.WriteString(", ")
		}
		qb.WriteString("(try path(")
		if strings.HasPrefix(pathExpr, ".") {
			qb.WriteString(pathExpr)
		} else {
			// specified by key string for backward compatibility
			qb.WriteString(".[\"")
			qb.WriteString(strings.ReplaceAll(pathExpr, "\"", "\\\""))
			qb.WriteString("\"]")
		}
		qb.WriteString("))")
	}
	qb.WriteString("])")

	query, err := gojq.Parse(qb.String())
	if err != nil {
		return nil, fmt.Errorf("failed to build the ignorePaths query: %w", err)
	}

	return query, nil
}

func applyJqTransformQueryCompiled(code *gojq.Code, input any) (any, error) {
	iter := code.Run(input)
	for {
		out, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := out.(error); ok {
			var haltErr *gojq.HaltError
			if errors.As(err, &haltErr) && haltErr.Value() == nil {
				break
			}

			return nil, err
		}

		return out, nil
	}
	return input, nil
}
