package k8s

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"go.starlark.net/starlark"
)

// K8s kubernetes API
type K8s struct {
}

var (
	_ starlark.HasAttrs = (*K8s)(nil)
)

// String -
func (k *K8s) String() string { return os.Getenv("KUBECONFIG") }

// Type -
func (k *K8s) Type() string { return "k8s" }

// Freeze -
func (k *K8s) Freeze() {}

// Truth -
func (k *K8s) Truth() starlark.Bool { return false }

// Hash -
func (k *K8s) Hash() (uint32, error) { panic("implement me") }

// Attr -
func (k *K8s) Attr(name string) (starlark.Value, error) {
	if name == "wait_crds" {
		return starlark.NewBuiltin("wait_crds", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (value starlark.Value, e error) {
			return starlark.None, nil
		}), nil
	}
	return starlark.None, starlark.NoSuchAttrError(fmt.Sprintf("k8s has no .%s attribute", name))
}

// AttrNames -
func (k *K8s) AttrNames() []string { return []string{"wait_crd"} }

// Apply -
func (k *K8s) Apply(namespace string, output func(io.Writer) error) error {
	return k.run("apply", namespace, output)
}

// Delete -
func (k *K8s) Delete(namespace string, output func(io.Writer) error) error {
	return k.run("delete", namespace, output, "--ignore-not-found")
}

func (k *K8s) run(command string, namespace string, output func(io.Writer) error, flags ...string) error {
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
