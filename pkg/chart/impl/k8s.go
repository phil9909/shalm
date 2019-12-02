package impl

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kramerul/shalm/pkg/chart"
	"github.com/pkg/errors"
)

// NewK8s create new instance to interact with kubernetes
func NewK8s() chart.K8s {
	return &k8sImpl{}
}

// k8sImpl -
type k8sImpl struct {
	namespace string
}

var (
	_ chart.K8s = (*k8sImpl)(nil)
)

// Apply -
func (k *k8sImpl) Apply(output func(io.Writer) error, options *chart.K8sOptions) error {
	return k.run("apply", output, options)
}
func (k *k8sImpl) ForNamespace(namespace string) chart.K8s {
	result := &k8sImpl{namespace: namespace}
	return result
}

// Delete -
func (k *k8sImpl) Delete(output func(io.Writer) error, options *chart.K8sOptions) error {
	return k.run("delete", output, options, "--ignore-not-found")
}

// Delete -
func (k *k8sImpl) DeleteObject(kind string, name string, options *chart.K8sOptions) error {
	return k.kubectl("delete", options, kind, name, "--ignore-not-found").Run()
}

// RolloutStatus -
func (k *k8sImpl) RolloutStatus(kind string, name string, options *chart.K8sOptions) error {
	start := time.Now()
	for {
		err := k.kubectl("rollout", options, "status", kind, name).Run()
		if err == nil {
			return nil
		}
		if !k.IsNotExist(err) {
			return err
		}
		if options.Timeout > 0 {
			if time.Since(start) > options.Timeout {
				return fmt.Errorf("Timeout during waiting for %s %s", kind, name)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// Get -
func (k *k8sImpl) Get(kind string, name string, writer io.Writer, options *chart.K8sOptions) error {
	cmd := k.kubectl("get", options, kind, name, "-o", "yaml")
	buffer := bytes.Buffer{}
	cmd.Stdout = writer
	cmd.Stderr = &buffer
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, string(buffer.Bytes()))
	}
	return nil
}

// IsNotExist -
func (k *k8sImpl) IsNotExist(err error) bool {
	return strings.Contains(err.Error(), "NotFound")
}

func (k *k8sImpl) kubectl(command string, options *chart.K8sOptions, flags ...string) *exec.Cmd {
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

type writeCounter struct {
	counter int
	writer  io.Writer
}

func (w *writeCounter) Write(data []byte) (int, error) {
	w.counter++
	return w.writer.Write(data)
}

func (k *k8sImpl) run(command string, output func(io.Writer) error, options *chart.K8sOptions, flags ...string) error {
	cmd := k.kubectl(command, options, append([]string{"-f", "-"}, flags...)...)

	writer, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting %s: %s", cmd.String(), err.Error())
	}
	w := &writeCounter{writer: writer}
	err = output(w)
	if err != nil {
		return err
	}
	writer.Close()
	err = cmd.Wait()
	if err != nil {
		if w.counter == 0 {
			return nil
		}
		return fmt.Errorf("error running %s: %s", cmd.String(), err.Error())
	}
	return nil

}
