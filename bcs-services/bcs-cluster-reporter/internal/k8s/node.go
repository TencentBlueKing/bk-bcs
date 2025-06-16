/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package k8s xxx
package k8s

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/remotecommand"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/interrupt"
)

// NodeController XXX
type NodeController struct {
	clusterID string
	nodeName  string
	config    *rest.Config
	pod       *corev1.Pod
	lock      sync.Mutex
	cs        *kubernetes.Clientset
}

const (
	defaultNamespace = "bkmonitor-operator"
	defaultImage     = "alpine:latest"
)

// NewNodeController XXX
func NewNodeController(nodeName string, config *rest.Config, clusterID string, image string) (*NodeController, error) {
	podName := fmt.Sprintf("cluster-reporter-%s", nodeName)
	if image == "" {
		image = defaultImage
	}

	// create pod and control node by it
	pod := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: podName,
		},
		Spec: corev1.PodSpec{
			NodeName:    nodeName,
			HostIPC:     true,
			HostPID:     true,
			HostNetwork: true,
			Containers: []corev1.Container{
				{
					Name:    "alpine",
					Image:   image,
					Command: []string{"nsenter", "--target", "1", "--mount", "--uts", "--ipc", "--net", "--pid", "bash", "-l"},
					SecurityContext: &corev1.SecurityContext{
						Privileged: ptrBool(true),
					},
					TTY:       false,
					Stdin:     true,
					StdinOnce: true,
				},
			},
			Tolerations: []corev1.Toleration{
				{Key: "CriticalAddonsOnly", Operator: "Exists"},
				{Key: "NoExecute", Operator: "Exists"},
			},

			RestartPolicy: corev1.RestartPolicyNever, // 单次执行可设置为 Never
		},
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("%s NodeExecCommand failed: %s", nodeName, err.Error())
		return nil, err
	}

	// create pod
	_, err = cs.CoreV1().Pods(defaultNamespace).Create(context.Background(), pod, v1.CreateOptions{})
	if err != nil {
		klog.Errorf("%s NodeExecCommand failed: %s", nodeName, err.Error())
		return nil, err
	}

	nc := &NodeController{
		clusterID: clusterID,
		pod:       pod,
		nodeName:  nodeName,
		config:    config,
		cs:        cs,
	}
	// 等待pod状态正常
	pod, err = waitForPod(cs.CoreV1(), defaultNamespace, pod.Name, 30*time.Second)
	if err != nil {
		nc.Close()
		klog.Error(err.Error())
		return nil, err
	} else {
		nc.pod = pod
	}

	return nc, nil
}

// Close delete controller pod
func (c *NodeController) Close() {
	err := c.cs.CoreV1().Pods(c.pod.Namespace).Delete(context.Background(), c.pod.Name, v1.DeleteOptions{})
	if err != nil {
		klog.Errorf("%s delete pod failed: %s", c.nodeName, err.Error())
	}
}

// NodeGetFile copy file from node from pod srcPath to local dstPath
func (c *NodeController) NodeGetFile(srcPath string, dstPath string) error {
	result, err := c.NodeExecCommand(map[string]string{"filePaths": fmt.Sprintf("ls %s", srcPath)})
	if err != nil {
		klog.Errorf(err.Error())
		return err
	}

	filePaths := strings.Split(result["filePaths"], "\n")

	// get one file only
	for _, filePath := range filePaths {
		if filePath == "" {
			continue
		}

		// init a exec req to tar file
		req := c.cs.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(c.pod.Name).
			Namespace(c.pod.Namespace).
			SubResource("exec").
			VersionedParams(&corev1.PodExecOptions{
				Container: "alpine",
				Command:   []string{"tar", "cf", "-", filePath},
				Stdin:     false,
				Stdout:    true,
				Stderr:    true,
			}, scheme.ParameterCodec)

		executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
		if err != nil {
			klog.Errorf(err.Error())
			return err
		}

		pipReader, pipWriter := io.Pipe()
		defer func() {
			pipReader.Close()
			pipWriter.Close()
		}()

		go func() {
			// exec req
			err = executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
				Stdout: pipWriter,
				Stderr: os.Stderr,
				Tty:    false,
			})

			if err != nil {
				klog.Errorf(err.Error())
			}
			pipWriter.Close()
		}()

		time.Sleep(time.Second)

		// untar file to local dstPath from pipReader
		err = untar(pipReader, path.Join(dstPath, filepath.Base(filePath)))
		if err != nil {
			klog.Errorf(err.Error())
			return err
		}

		// read from pipReader to end executor
		for {
			tmp := make([]byte, 4)
			_, err := pipReader.Read(tmp)
			if err == io.EOF {
				break
			}
		}
	}

	return nil
}

// untar untar file from reader to dest
func untar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		//target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dest, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
	return nil
}

// NodeExecCommand exec command on host
func (c *NodeController) NodeExecCommand(commands map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for key, command := range commands {
		// init a exec req to exec command
		req := c.cs.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(c.pod.Name).
			Namespace(c.pod.Namespace).
			SubResource("exec").
			VersionedParams(&corev1.PodExecOptions{
				Container: "alpine",
				Command:   []string{"/bin/sh", "-c", command},
				Stdin:     false,
				Stdout:    true,
				Stderr:    true,
			}, scheme.ParameterCodec)

		executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
		if err != nil {
			klog.Errorf(err.Error())
			return nil, err
		}

		// get command result
		stdinReader, stdinWriter := io.Pipe()
		defer func() {
			stdinWriter.Close()
			stdinReader.Close()
		}()

		go func() {
			err = executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
				Stdout: stdinWriter,
				Stderr: os.Stderr,
				Tty:    false,
			})
			if err != nil {
				klog.Errorf("%s NodeExecCommand failed: %s", c.nodeName, err.Error())
			}
			stdinWriter.Close()
		}()

		time.Sleep(time.Second)

		commandResult := make([]byte, 0)
		for {
			tmp := make([]byte, 4)
			n, err := stdinReader.Read(tmp)
			if err == io.EOF {
				break
			}
			commandResult = append(commandResult, tmp[:n]...)
		}
		result[key] = string(commandResult)
	}

	return result, nil
}

// execCommand exec command by sendint command to writer and get result from reader
func execCommand(commmands map[string]string, in *io.PipeWriter, out *bytes.Buffer) (map[string]string, error) {
	result := make(map[string]string)
	length := 0
	for key, command := range commmands {
		if command == "exit" {
			_, err := in.Write([]byte(fmt.Sprintf("%s\n", command)))
			if err != nil {
				return nil, err
			}
			break
		}

		_, err := in.Write([]byte(fmt.Sprintf(" %s\n echo 'bcs command done'\n", command)))
		if err != nil {
			return nil, err
		}
		time.Sleep(time.Second)

		out.Next(length)
		commandResult := make([]byte, 0)
		for {
			tmp := make([]byte, 4)
			n, err := out.Read(tmp)
			if err == io.EOF && strings.Contains(string(commandResult), "bcs command done") {
				break
			}
			commandResult = append(commandResult, tmp[:n]...)
		}
		length += len(commandResult)
		commandStr := string(commandResult)
		commandLines := strings.Split(commandStr, "\n")
		result[key] = strings.Join(commandLines[:len(commandLines)-2], "\n")
	}

	return result, nil
}

// waitForPod wait for pod to be normal status(Running)
func waitForPod(podClient typedv1.PodsGetter, ns, name string, timeout time.Duration) (*corev1.Pod, error) {
	ctx, cancel := watchtools.ContextWithOptionalTimeout(context.Background(), timeout)
	defer cancel()

	fieldSelector := fields.OneTermEqualSelector("metadata.name", name).String()
	lw := &cache.ListWatch{
		ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			return podClient.Pods(ns).List(context.TODO(), options)
		},
		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			return podClient.Pods(ns).Watch(context.TODO(), options)
		},
	}

	intr := interrupt.New(nil, cancel)
	var result *corev1.Pod
	err := intr.Run(func() error {
		ev, err := watchtools.UntilWithSync(ctx, lw, &corev1.Pod{}, nil, podRunningAndReady)
		if ev != nil {
			result = ev.Object.(*corev1.Pod)
		}
		return err
	})

	return result, err
}

// podRunningAndReady get pod status from event
func podRunningAndReady(event watch.Event) (bool, error) {
	switch event.Type {
	case watch.Deleted:
		return false, errors.NewNotFound(schema.GroupResource{Resource: "pods"}, "")
	}
	switch t := event.Object.(type) {
	case *corev1.Pod:
		switch t.Status.Phase {
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, fmt.Errorf("ErrPodCompleted")
		case corev1.PodRunning:
			conditions := t.Status.Conditions
			if conditions == nil {
				return false, nil
			}
			for i := range conditions {
				if conditions[i].Type == corev1.PodReady &&
					conditions[i].Status == corev1.ConditionTrue {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func ptrBool(v bool) *bool {
	return &v
}
