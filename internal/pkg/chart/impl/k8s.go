package impl

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/kramerul/shalm/internal/pkg/chart/api"
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
func (k *k8sImpl) Apply(output func(io.Writer) error) error {
	return k.run("apply", output)
}
func (k *k8sImpl) ForNamespace(namespace string) api.K8s {
	result := &k8sImpl{namespace: namespace}
	return result
}

// Delete -
func (k *k8sImpl) Delete(output func(io.Writer) error) error {
	return k.run("delete", output, "--ignore-not-found")
}

// Delete -
func (k *k8sImpl) DeleteObject(kind string, name string) error {
	return k.kubectl("delete", kind, name, "--ignore-not-found").Run()
}

// RolloutStatus -
func (k *k8sImpl) RolloutStatus(typ string, name string, timeout time.Duration) error {
	return k.kubectl("rollout", "status", typ, name, "--timeout", fmt.Sprintf("%.0fs", timeout.Seconds())).Run()
}

func (k *k8sImpl) kubectl(command string, flags ...string) *exec.Cmd {
	cmd := exec.Command("kubectl", append([]string{"-n", k.namespace, command}, flags...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (k *k8sImpl) run(command string, output func(io.Writer) error, flags ...string) error {
	cmd := k.kubectl(command, append([]string{"-f", "-"}, flags...)...)

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
