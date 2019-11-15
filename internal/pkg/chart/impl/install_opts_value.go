package impl

import (
	"fmt"

	"go.starlark.net/starlark"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
)

type installOptsValue struct {
	*api.InstallOpts
}

var (
	_ starlark.HasAttrs = (*installOptsValue)(nil)
)

// NewInstallOptsValue -
func NewInstallOptsValue(installOpts *api.InstallOpts) starlark.Value {
	return &installOptsValue{installOpts}
}

// String -
func (r *installOptsValue) String() string {
	return "installOpts"
}

// Type -
func (r *installOptsValue) Type() string {
	return "installOpts"
}

// Freeze -
func (r *installOptsValue) Freeze() {
}

// Truth -
func (r *installOptsValue) Truth() starlark.Bool {
	return false
}

// Hash -
func (r *installOptsValue) Hash() (uint32, error) {
	panic("implement me")
}

// Attr -
func (r *installOptsValue) Attr(name string) (starlark.Value, error) {
	if name == "namespace" {
		return starlark.String(r.Namespace), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("installOpts has no .%s attribute", name))
}

// AttrNames -
func (r *installOptsValue) AttrNames() []string {
	return []string{"namespace", "name", "service"}
}
