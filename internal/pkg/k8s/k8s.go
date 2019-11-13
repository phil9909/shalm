package k8s

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"go.starlark.net/starlark"
)

// K8s kubernetes API
type K8s interface {
	RolloutStatus(namespace string, typ string, name string, timeout time.Duration) error
	Apply(namespace string, output func(io.Writer) error) error
	Delete(namespace string, output func(io.Writer) error) error
}

// K8sValue -
type K8sValue interface {
	starlark.Value
	K8s
}

// New create new instance to interact with kubernetes
func New() K8sValue {
	return &k8sValueImpl{&k8sImpl{}}
}

// NewForTest create new instance to interact with kubernetes
func NewForTest(k K8s) K8sValue {
	return &k8sValueImpl{k}
}

type k8sValueImpl struct {
	K8s
}

// k8sImpl -
type k8sImpl struct {
}

var (
	_ starlark.HasAttrs = (*k8sValueImpl)(nil)
)

// String -
func (k *k8sValueImpl) String() string { return os.Getenv("KUBECONFIG") }

// Type -
func (k *k8sValueImpl) Type() string { return "k8s" }

// Freeze -
func (k *k8sValueImpl) Freeze() {}

// Truth -
func (k *k8sValueImpl) Truth() starlark.Bool { return false }

// Hash -
func (k *k8sValueImpl) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (k *k8sValueImpl) Attr(name string) (starlark.Value, error) {
	if name == "rollout_status" {
		return starlark.NewBuiltin("rollout_status", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			var timeout = 120
			var namespace string
			var typ string
			var name string
			if err := starlark.UnpackArgs("rollout_status", args, kwargs, "namespace", &namespace,
				"type", &typ, "name", &name, "timeout?", &timeout); err != nil {
				return nil, err
			}
			return starlark.None, k.RolloutStatus(namespace, typ, name, time.Duration(timeout)*time.Second)
		}), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("k8s has no .%s attribute", name))
}

// AttrNames -
func (k *k8sValueImpl) AttrNames() []string { return []string{"wait_crd"} }

// Apply -
func (k *k8sImpl) Apply(namespace string, output func(io.Writer) error) error {
	return k.run(namespace, "apply", output)
}

// Delete -
func (k *k8sImpl) Delete(namespace string, output func(io.Writer) error) error {
	return k.run(namespace, "delete", output, "--ignore-not-found")
}

// RolloutStatus -
func (k *k8sImpl) RolloutStatus(namespace string, typ string, name string, timeout time.Duration) error {
	return k.kubectl(namespace, "rollout", "status", typ, name, "--timeout", fmt.Sprint("%10.0fs", timeout.Seconds())).Run()
}

func (k *k8sImpl) kubectl(namespace string, command string, flags ...string) *exec.Cmd {
	cmd := exec.Command("kubectl", append([]string{"-n", namespace, command}, flags...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (k *k8sImpl) run(namespace string, command string, output func(io.Writer) error, flags ...string) error {
	cmd := k.kubectl(command, "-f", "-")

	writer, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting %s: %s", cmd.String(), err.Error())
	}
	err = output(writer)
	if err != nil {
		return err
	}
	writer.Close()
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error running %s: %s", cmd.String(), err.Error())
	}
	return nil

}
