package api

import (
	"go.starlark.net/starlark"
)

// Repo -
type Repo interface {
	// Get -
	Get(thread *starlark.Thread, parent Chart, name string, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error)
	// Push -
	Push(chart Chart, ref string) error
}
