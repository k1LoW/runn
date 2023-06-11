package builtin

import "github.com/juliangruber/go-intersect"

func Intersect(x, y any) any {
	return intersect.Simple(x, y)
}
