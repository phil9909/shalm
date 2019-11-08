package k8s

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"go.starlark.net/starlark"
)

// K8s kubernetes API
type K8s interface {
	starlark.HasAttrs
	Apply(namespace string, output func(io.Writer) error) error
	Delete(namespace string, output func(io.Writer) error) error
}

// New create new instance to interact with kubernetes
func New() K8s {
	return &k8sImpl{}
}

// k8sImpl -
type k8sImpl struct {
}

var (
	_ K8s = (*k8sImpl)(nil)
)

// String -
func (k *k8sImpl) String() string { return os.Getenv("KUBECONFIG") }

// Type -
func (k *k8sImpl) Type() string { return "k8s" }

// Freeze -
func (k *k8sImpl) Freeze() {}

// Truth -
func (k *k8sImpl) Truth() starlark.Bool { return false }

// Hash -
func (k *k8sImpl) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (k *k8sImpl) Attr(name string) (starlark.Value, error) {
	if name == "wait_crds" {
		return starlark.NewBuiltin("wait_crds", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			return starlark.None, nil
		}), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("k8s has no .%s attribute", name))
}

// AttrNames -
func (k *k8sImpl) AttrNames() []string { return []string{"wait_crd"} }

// Apply -
func (k *k8sImpl) Apply(namespace string, output func(io.Writer) error) error {
	return k.run("apply", namespace, output)
}

// Delete -
func (k *k8sImpl) Delete(namespace string, output func(io.Writer) error) error {
	return k.run("delete", namespace, output, "--ignore-not-found")
}

func (k *k8sImpl) run(command string, namespace string, output func(io.Writer) error, flags ...string) error {
	args := append([]string{"-n", namespace, command, "-f", "-"}, flags...)
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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
