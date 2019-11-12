package fakes

import (
	"io"
	"time"

	"github.com/kramerul/shalm/internal/pkg/k8s"
)

// K8sFake -
type K8sFake struct {
	Writer             io.Writer
	RolloutStatusCalls []string
}

var (
	_ k8s.K8s = (*K8sFake)(nil)
)

func (k *K8sFake) RolloutStatus(namespace string, typ string, name string, timeout time.Duration) error {
	k.RolloutStatusCalls = append(k.RolloutStatusCalls, name)
	return nil
}

// Apply -
func (k *K8sFake) Apply(namespace string, output func(io.Writer) error) error {
	return output(k.Writer)
}

// Delete -
func (k *K8sFake) Delete(namespace string, output func(io.Writer) error) error {
	return output(k.Writer)
}
