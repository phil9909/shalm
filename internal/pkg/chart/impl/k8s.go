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
}

var (
	_ api.K8s = (*k8sImpl)(nil)
)

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
	return k.kubectl(namespace, "rollout", "status", typ, name, "--timeout", fmt.Sprintf("%.0fs", timeout.Seconds())).Run()
}

func (k *k8sImpl) kubectl(namespace string, command string, flags ...string) *exec.Cmd {
	cmd := exec.Command("kubectl", append([]string{"-n", namespace, command}, flags...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (k *k8sImpl) run(namespace string, command string, output func(io.Writer) error, flags ...string) error {
	cmd := k.kubectl(namespace, command, "-f", "-")

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
