package builtin

import "github.com/juliangruber/go-intersect"

func Intersect(x, y interface{}) interface{} {
	return intersect.Simple(x, y)
}
