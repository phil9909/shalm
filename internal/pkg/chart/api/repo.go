package api

import (
	"go.starlark.net/starlark"
)

// Repo -
type Repo interface {
	// Get -
	Get(thread *starlark.Thread, name string, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error)
}
