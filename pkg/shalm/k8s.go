package shalm

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// NewK8s create new instance to interact with kubernetes
func NewK8s() K8s {
	return &k8sImpl{}
}

// NewK8sFromContent create new instance to interact with kubernetes
func NewK8sFromContent(kubeConfig string) (K8s, error) {
	if kubeConfig == "" {
		return NewK8s(), nil
	}
	kubeconfig, err := kubeConfigFromContent(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &k8sImpl{kubeconfig: &kubeconfig}, nil
}

// k8sImpl -
type k8sImpl struct {
	namespace  string
	kubeconfig *string
	cmd        string
}

var (
	_ K8s = (*k8sImpl)(nil)
)

func (k *k8sImpl) Inspect() string {
	if k.kubeconfig != nil {
		return "kubeconfig = " + *k.kubeconfig + " namespace = " + k.namespace
	}
	return "namespace = " + k.namespace
}

// Apply -
func (k *k8sImpl) Apply(output func(io.Writer) error, options *K8sOptions) error {
	return k.run("apply", output, options)
}
func (k *k8sImpl) ForNamespace(namespace string) K8s {
	result := &k8sImpl{namespace: namespace, kubeconfig: k.kubeconfig}
	return result
}

// Delete -
func (k *k8sImpl) Delete(output func(io.Writer) error, options *K8sOptions) error {
	return k.run("delete", output, options, "--ignore-not-found")
}

// Delete -
func (k *k8sImpl) DeleteObject(kind string, name string, options *K8sOptions) error {
	return run(k.kubectl("delete", options, kind, name, "--ignore-not-found"))
}

// RolloutStatus -
func (k *k8sImpl) RolloutStatus(kind string, name string, options *K8sOptions) error {
	start := time.Now()
	for {
		err := run(k.kubectl("rollout", options, "status", kind, name))
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

func (k *k8sImpl) Wait(kind string, name string, condition string, options *K8sOptions) error {
	return run(k.kubectl("wait", options, kind, name, "--for", condition))
}

// Get -
func (k *k8sImpl) Get(kind string, name string, writer io.Writer, options *K8sOptions) error {
	cmd := k.kubectl("get", options, kind, name, "-o", "json")
	cmd.Stdout = writer
	return run(cmd)
}

func (k *k8sImpl) Watch(kind string, name string, options *K8sOptions) (io.ReadCloser, error) {
	cmd := k.kubectl("get", options, kind, name, "-o", "json", "--watch")
	reader, writer := io.Pipe()
	cmd.Stdout = writer
	return reader, cmd.Start()
}

// IsNotExist -
func (k *k8sImpl) IsNotExist(err error) bool {
	return strings.Contains(err.Error(), "NotFound")
}

// IsNotExist -
func (k *k8sImpl) KubeConfigContent() *string {
	return k.kubeconfig
}

func run(cmd *exec.Cmd) error {
	buffer := bytes.Buffer{}
	cmd.Stderr = &buffer
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, string(buffer.Bytes()))
	}
	return nil

}

func (k *k8sImpl) kubectl(command string, options *K8sOptions, flags ...string) *exec.Cmd {
	if k.kubeconfig != nil {
		flags = append([]string{command, "--kubeconfig", *k.kubeconfig}, flags...)
	} else {
		flags = append([]string{command}, flags...)
	}
	if options.Namespaced {
		flags = append(flags, "-n", k.namespace)
	}
	if options.Timeout > 0 {
		flags = append(flags, "--timeout", fmt.Sprintf("%.0fs", options.Timeout.Seconds()))
	}
	kubectl := k.cmd
	if kubectl == "" {
		kubectl = "kubectl"
	}
	cmd := exec.Command(kubectl, flags...)
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

func (k *k8sImpl) run(command string, output func(io.Writer) error, options *K8sOptions, flags ...string) error {
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
