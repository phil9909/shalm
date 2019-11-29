package api

import (
	"go.starlark.net/starlark"
)

// Repo -
type Repo interface {
	// Get -
	Get(thread *starlark.Thread, parent main.Chart, name string, args starlark.Tuple, kwargs []starlark.Tuple) (main.ChartValue, error)
	// Push -
	Push(chart main.Chart, ref string) error
}
