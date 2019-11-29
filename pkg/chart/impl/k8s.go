package impl

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/kramerul/shalm/pkg/chart/api"
)

// NewK8s create new instance to interact with kubernetes
func NewK8s() api.K8s {
	return &k8sImpl{}
}

// k8sImpl -
type k8sImpl struct {
	namespace string
}

var (
	_ api.K8s = (*k8sImpl)(nil)
)

// Apply -
func (k *k8sImpl) Apply(output func(io.Writer) error, options *api.K8sOptions) error {
	return k.run("apply", output, options)
}
func (k *k8sImpl) ForNamespace(namespace string) api.K8s {
	result := &k8sImpl{namespace: namespace}
	return result
}

// Delete -
func (k *k8sImpl) Delete(output func(io.Writer) error, options *api.K8sOptions) error {
	return k.run("delete", output, options, "--ignore-not-found")
}

// Delete -
func (k *k8sImpl) DeleteObject(kind string, name string, options *api.K8sOptions) error {
	return k.kubectl("delete", options, kind, name, "--ignore-not-found").Run()
}

// RolloutStatus -
func (k *k8sImpl) RolloutStatus(typ string, name string, options *api.K8sOptions) error {
	return k.kubectl("rollout", options, "status", typ, name).Run()
}

func (k *k8sImpl) kubectl(command string, options *api.K8sOptions, flags ...string) *exec.Cmd {
	flags = append([]string{command}, flags...)
	if options.Namespaced {
		flags = append(flags, "-n", k.namespace)
	}
	if options.Timeout > 0 {
		flags = append(flags, "--timeout", fmt.Sprintf("%.0fs", options.Timeout.Seconds()))
	}
	cmd := exec.Command("kubectl", flags...)
	fmt.Println(cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (k *k8sImpl) run(command string, output func(io.Writer) error, options *api.K8sOptions, flags ...string) error {
	cmd := k.kubectl(command, options, append([]string{"-f", "-"}, flags...)...)

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
