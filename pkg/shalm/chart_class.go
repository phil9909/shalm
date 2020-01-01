package shalm

import (
	"fmt"

	"go.starlark.net/starlark"
)

type chartClass struct {
	APIVersion  string   `json:"apiVersion,omitempty"`
	Name        string   `json:"name,omitempty"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Home        string   `json:"home,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	Icon        string   `json:"icon,omitempty"`
}

// String -
func (cc *chartClass) String() string { return cc.Name }

// Type -
func (cc *chartClass) Type() string { return "chart_class" }

// Freeze -
func (cc *chartClass) Freeze() {}

// Truth -
func (cc *chartClass) Truth() starlark.Bool { return false }

// Hash -
func (cc *chartClass) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (cc *chartClass) Attr(name string) (starlark.Value, error) {
	switch name {
	case "api_version":
		return starlark.String(cc.APIVersion), nil
	case "name":
		return starlark.String(cc.Name), nil
	case "version":
		return starlark.String(cc.Version), nil
	case "description":
		return starlark.String(cc.Description), nil
	case "keywords":
		return toStarlark(cc.Keywords), nil
	case "home":
		return starlark.String(cc.Home), nil
	case "sources":
		return toStarlark(cc.Sources), nil
	case "icon":
		return starlark.String(cc.Icon), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("chart_class has no .%s attribute", name))
}

// AttrNames -
func (cc *chartClass) AttrNames() []string {
	return []string{"api_version", "name", "version", "description", "keywords", "home", "sources", "icon"}
}
