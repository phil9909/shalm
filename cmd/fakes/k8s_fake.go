package fakes

import (
	"io"
	"time"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
)

// K8sFake -
type K8sFake struct {
	Writer             io.Writer
	RolloutStatusCalls []string
	Namespace          string
}

var (
	_ api.K8s = (*K8sFake)(nil)
)

// ForNamespace -
func (k *K8sFake) ForNamespace(namespace string) api.K8s {
	k.Namespace = namespace
	return k
}

// RolloutStatus -
func (k *K8sFake) RolloutStatus(kind string, name string, timeout time.Duration) error {
	k.RolloutStatusCalls = append(k.RolloutStatusCalls, name)
	return nil
}

// DeleteObject -
func (k *K8sFake) DeleteObject(kind string, name string) error {
	return nil
}

// Apply -
func (k *K8sFake) Apply(output func(io.Writer) error) error {
	return output(k.Writer)
}

// Delete -
func (k *K8sFake) Delete(output func(io.Writer) error) error {
	return output(k.Writer)
}
