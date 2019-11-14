package impl

import (
	"fmt"

	"go.starlark.net/starlark"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
)

type releaseValue struct {
	*api.Release
}

var (
	_ starlark.HasAttrs = (*releaseValue)(nil)
)

// NewReleaseValue -
func NewReleaseValue(release *api.Release) starlark.Value {
	return &releaseValue{release}
}

// String -
func (r *releaseValue) String() string {
	return "release"
}

// Type -
func (r *releaseValue) Type() string {
	return "release"
}

// Freeze -
func (r *releaseValue) Freeze() {
}

// Truth -
func (r *releaseValue) Truth() starlark.Bool {
	return false
}

// Hash -
func (r *releaseValue) Hash() (uint32, error) {
	panic("implement me")
}

// Attr -
func (r *releaseValue) Attr(name string) (starlark.Value, error) {
	if name == "namespace" {
		return starlark.String(r.Namespace), nil
	}
	if name == "name" {
		return starlark.String(r.Name), nil
	}
	if name == "service" {
		return starlark.String(r.Service), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("release has no .%s attribute", name))
}

// AttrNames -
func (r *releaseValue) AttrNames() []string {
	return []string{"namespace", "name", "service"}
}
