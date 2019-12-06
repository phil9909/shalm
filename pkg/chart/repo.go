package chart

import (
	"go.starlark.net/starlark"
)

// Repo -
type Repo interface {
	// Get -
	Get(thread *starlark.Thread, url string, namespace string, args starlark.Tuple, kwargs []starlark.Tuple) (ChartValue, error)
}
