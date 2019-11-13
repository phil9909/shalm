package api

import "go.starlark.net/starlark"

// Release -
type Release struct {
	Name      string
	Namespace string
	Service   string
}

// HelmChart -
type HelmChart struct {
	APIVersion  string   `json:"apiVersion,omitempty"`
	Name        string   `json:"name,omitempty"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Home        string   `json:"home,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	Icon        string   `json:"icon,omitempty"`
}

// Chart -
type Chart interface {
	ApplyFunction() starlark.Callable
	DeleteFunction() starlark.Callable
	TemplateFunction() starlark.Callable
}

// ChartValue -
type ChartValue interface {
	starlark.HasSetField
	Chart
}
