/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package container

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"encoding/json"
	"time"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schedTypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"

	dockerclient "github.com/fsouza/go-dockerclient"
)

// DockerInterface is an abstract interface for testability.  It abstracts the interface of docker.Client.
// type Container interface {
// 	ListContainers(options docker.ListContainersOptions) ([]docker.APIContainers, error)
// 	InspectContainer(id string) (*docker.Container, error)
// 	CreateContainer(docker.CreateContainerOptions) (*docker.Container, error)
// 	StartContainer(id string, hostConfig *docker.HostConfig) error
// 	StopContainer(id string, timeout uint) error
// 	RemoveContainer(opts docker.RemoveContainerOptions) error
// 	InspectImage(image string) (*docker.Image, error)
// 	ListImages(opts docker.ListImagesOptions) ([]docker.APIImages, error)
// 	PullImage(opts docker.PullImageOptions, auth docker.AuthConfiguration) error
// 	RemoveImage(image string) error
// 	Logs(opts docker.LogsOptions) error
// 	Version() (*docker.Env, error)
// 	Info() (*docker.Env, error)
// 	CreateExec(docker.CreateExecOptions) (*docker.Exec, error)
// 	StartExec(string, docker.StartExecOptions) error
// 	InspectExec(id string) (*docker.ExecInspect, error)
// 	AttachToContainer(opts docker.AttachToContainerOptions) error
// }

const (
	// StopContainerGraceTime executor default gracetime
	StopContainerGraceTime = 10
	// DefaultDockerCPUPeriod default cpu period for executor
	DefaultDockerCPUPeriod = 100000
)

//DockerContainer implement container interface
//to handle all operator with docker command line.
type DockerContainer struct {
	user     string               //login docker user name
	passwd   string               //login docker user password
	client   *dockerclient.Client //docker operator client
	endpoint string               //docker daemon endpoint for client
}

//NewDockerContainer create DockerContainer manager tool
//endpoint can be http, https, tcp, or unix domain sock
func NewDockerContainer(endpoint, user, passwd string) Container {
	client, err := dockerclient.NewClient(endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create DockerContainer Manager failed: %s", err.Error())
		return nil
	}
	container := &DockerContainer{
		user:     user,
		passwd:   passwd,
		client:   client,
		endpoint: endpoint,
	}
	//check connection for remote docker daemon
	return container
}

//RunCommand running command in Container
func (docker *DockerContainer) RunCommand(containerID string, command []string) error {
	//create exec with command
	createOpt := dockerclient.CreateExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          command,
		Container:    containerID,
		User:         "root",
		Privileged:   true,
	}
	exeInst, err := docker.client.CreateExec(createOpt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create docker exec failed: %s", err.Error())
		return err
	}
	//start exec
	startOpt := dockerclient.StartExecOptions{
		Detach: true,
		Tty:    true,
	}
	if startErr := docker.client.StartExec(exeInst.ID, startOpt); startErr != nil {
		fmt.Fprintf(os.Stderr, "Start docker exec failed: %s", startErr.Error())
		return startErr
	}
	return nil
}

// RunCommandV2 v2 version run command
func (docker *DockerContainer) RunCommandV2(ops *schedTypes.RequestCommandTask) (*schedTypes.ResponseCommandTask, error) {
	if ops.User == "" {
		ops.User = "root"
	}
	by, _ := json.Marshal(ops)
	logs.Infof("docker run command %s", string(by))
	//create exec with command
	createOpt := dockerclient.CreateExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          ops.Cmd,
		Container:    ops.ContainerId,
		User:         ops.User,
		Privileged:   ops.Privileged,
		Env:          ops.Env,
	}

	resp := &schedTypes.ResponseCommandTask{
		ID:          ops.ID,
		TaskId:      ops.TaskId,
		ContainerId: ops.ContainerId,
	}
	exeInst, err := docker.client.CreateExec(createOpt)
	if err != nil {
		logs.Errorf("docker create exec error %s", err.Error())
		resp.Status = commtypes.TaskCommandStatusFailed
		resp.Message = err.Error()
		return resp, nil
	}
	//start exec
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	startOpt := dockerclient.StartExecOptions{
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	}
	err = docker.client.StartExec(exeInst.ID, startOpt)
	if err != nil {
		logs.Errorf("docker start exec error %s", err.Error())
		resp.Status = commtypes.TaskCommandStatusFailed
		resp.Message = err.Error()
		return resp, nil
	}

	ins, err := docker.client.InspectExec(exeInst.ID)
	if err != nil {
		logs.Errorf("docker inspect exec error %s", err.Error())
		resp.Status = commtypes.TaskCommandStatusFailed
		resp.Message = err.Error()
		return resp, nil
	}
	logs.Infof("docker container %s run command success", ops.ContainerId)

	var out, errB []byte
	if outBuf.Len() > 256 {
		out = outBuf.Bytes()[:256]
	} else {
		out = outBuf.Bytes()
	}
	if errBuf.Len() > 256 {
		errB = errBuf.Bytes()[:256]
	} else {
		errB = errBuf.Bytes()
	}

	resp.Status = commtypes.TaskCommandStatusFinish
	resp.CommInspect = &commtypes.CommandInspectInfo{}
	resp.CommInspect.ExitCode = ins.ExitCode
	resp.CommInspect.Stderr = string(errB)
	resp.CommInspect.Stdout = string(out)

	return resp, nil
}

//UploadToContainer upload file from host to Container
func (docker *DockerContainer) UploadToContainer(containerID string, source, dest string) error {
	//create tarfile by source
	tarFile, err := os.Create(source + ".tar")
	defer tarFile.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create tar file for %s failed.\n", source)
		return err
	}
	tw := tar.NewWriter(tarFile)
	defer tw.Close()
	//get fileinfo from source
	fileInfo, statErr := os.Stat(source)
	if statErr != nil {
		fmt.Fprintf(os.Stderr, "State %s failed.\n", source)
		return statErr
	}
	hdr, hErr := tar.FileInfoHeader(fileInfo, "")
	if hErr != nil {
		fmt.Fprintf(os.Stderr, "Create tar file Header failed!\n")
		return hErr
	}
	if err := tw.WriteHeader(hdr); err != nil {
		fmt.Fprintln(os.Stderr, "Write tar file Header failed!")
		return err
	}
	//open source file
	sfile, sErr := os.Open(source)
	if sErr != nil {
		fmt.Fprintf(os.Stderr, "Open source %s failed!\n", source)
		return sErr
	}
	if _, copyErr := io.Copy(tw, sfile); copyErr != nil {
		fmt.Fprintf(os.Stderr, "Copy source file info to tar failed!\n")
		return copyErr
	}
	tw.Flush()
	input, iErr := os.Open(source + ".tar")
	defer input.Close()
	if iErr != nil {
		fmt.Fprintf(os.Stderr, "Open tar file for input stream failed\n")
		return iErr
	}
	//upload to Container
	destPath, absErr := filepath.Abs(filepath.Dir(dest))
	if absErr != nil {
		fmt.Fprintf(os.Stderr, "Get dest %s abs Err: %s\n", dest, absErr.Error())
		return absErr
	}
	uploadOpt := dockerclient.UploadToContainerOptions{
		InputStream:          input,
		Path:                 destPath,
		NoOverwriteDirNonDir: true,
	}
	return docker.client.UploadToContainer(containerID, uploadOpt)
}

//ListContainer list all running containner info
func (docker *DockerContainer) ListContainer() {

}

//CreateContainer Start Container with docker client.
func (docker *DockerContainer) CreateContainer(containerName string, containerTask *BcsContainerTask) (*BcsContainerInfo, error) {
	//Container Config
	config := &dockerclient.Config{
		Hostname:        containerTask.HostName,
		ExposedPorts:    make(map[dockerclient.Port]struct{}),
		AttachStdin:     false,
		AttachStdout:    true,
		AttachStderr:    true,
		Tty:             false,
		Labels:          make(map[string]string),
		OpenStdin:       false,
		StdinOnce:       false,
		NetworkDisabled: false,
		Image:           containerTask.Image,
	}
	//HostConfig
	hostConfig := &dockerclient.HostConfig{
		CapAdd:          []string{"SYS_PTRACE"},
		ExtraHosts:      containerTask.Hosts,
		PortBindings:    make(map[dockerclient.Port][]dockerclient.PortBinding),
		Privileged:      containerTask.Privileged,
		PublishAllPorts: containerTask.PublishAllPorts,
		OOMKillDisable:  &containerTask.OOMKillDisabled,
		ShmSize:         containerTask.ShmSize,
		IpcMode:         containerTask.Ipc,
	}
	//setting ulimit from docker parameter
	for _, item := range containerTask.Ulimits {
		value, _ := strconv.Atoi(item.Value)
		u := dockerclient.ULimit{
			Name: item.Key,
			Soft: int64(value),
			Hard: int64(value),
		}
		hostConfig.Ulimits = append(hostConfig.Ulimits, u)
	}

	//check images
	imageList, _ := docker.ListImage(containerTask.Image)
	//ready to pull if user setting pullImageForce
	if len(imageList) == 0 || containerTask.ForcePullImage {
		if pullErr := docker.PullImage(containerTask.Image); pullErr != nil {
			return nil, pullErr
		}
	}
	fmt.Fprintf(os.Stdout, "CreateContainer %s, hostname: %s\n", containerName, containerTask.HostName)
	//from BcsContainerTask.Volumes to dockerclient.Config
	for _, volumn := range containerTask.Volums {
		mount := dockerclient.Mount{
			Source:      volumn.HostPath,
			Destination: volumn.ContainerPath,
			RW:          !volumn.ReadOnly,
		}
		config.Mounts = append(config.Mounts, mount)
		//done(developerJim): setting HostConfig.Binds
		bind := fmt.Sprintf("%s:%s", volumn.HostPath, volumn.ContainerPath)
		hostConfig.Binds = append(hostConfig.Binds, bind)
	}
	//setting network
	if containerTask.NetworkName == "bridge" {
		hostConfig.NetworkMode = "default"
	} else {
		hostConfig.NetworkMode = containerTask.NetworkName
	}
	//enviroments
	if containerTask.Env != nil {
		for _, env := range containerTask.Env {
			item := fmt.Sprintf("%s=%s", env.Key, env.Value)
			config.Env = append(config.Env, item)
		}
	}
	//setting Cmd in config, if docker image setting entrypoint
	//empty CommandInfo.Value is acceptable, and arguments will become
	//entrypoint's arguments
	if containerTask.Command != "" {
		config.Cmd = append(config.Cmd, containerTask.Command)
	}
	if containerTask.Args != nil && len(containerTask.Args) > 0 {
		config.Cmd = append(config.Cmd, containerTask.Args...)
	}
	//resource setting
	//overlay2 do not support storage limitation
	// if containerTask.Resource.Disk != 0 {
	// 	hostConfig.StorageOpt = make(map[string]string)
	// 	hostConfig.StorageOpt["size"] = strconv.Itoa(int(containerTask.Resource.Disk)) + "M"
	// }
	if containerTask.Resource.Cpus <= 0 {
		hostConfig.CPUShares = 1024
	} else {
		hostConfig.CPUShares = int64(containerTask.Resource.Cpus * 1024)
	}
	if containerTask.Resource.Mem >= 4 {
		hostConfig.Memory = int64(containerTask.Resource.Mem * 1024 * 1024)
		hostConfig.MemorySwap = int64(containerTask.Resource.Mem * 1024 * 1024)
	}

	if containerTask.LimitResource != nil && containerTask.LimitResource.Cpus > 0 {
		hostConfig.CPUPeriod = DefaultDockerCPUPeriod
		hostConfig.CPUQuota = int64(containerTask.LimitResource.Cpus * DefaultDockerCPUPeriod)
	}
	if containerTask.LimitResource != nil && containerTask.LimitResource.Mem >= 4 {
		hostConfig.Memory = int64(containerTask.LimitResource.Mem * 1024 * 1024)
		hostConfig.MemorySwap = int64(containerTask.LimitResource.Mem * 1024 * 1024)
	}

	//done(developerJim): setting portMapping
	for key, value := range containerTask.PortBindings {
		var tmp struct{}
		config.ExposedPorts[dockerclient.Port(key)] = tmp
		bind := dockerclient.PortBinding{
			HostPort: value.HostPort,
		}
		var list []dockerclient.PortBinding
		list = append(list, bind)
		hostConfig.PortBindings[dockerclient.Port(key)] = list
	}
	//setting lables
	for _, kv := range containerTask.Labels {
		config.Labels[kv.Key] = kv.Value
	}
	//CreateContainer
	option := dockerclient.CreateContainerOptions{
		Name:       containerName,
		Config:     config,
		HostConfig: hostConfig,
	}

	if containerTask.NetworkName != "none" && containerTask.NetworkName != "bridge" && containerTask.NetworkName != "host" {
		if containerTask.NetworkIPAddr != "" {
			EndpointsConfig := make(map[string]*dockerclient.EndpointConfig, 0)
			EndpointsConfig[containerTask.NetworkName] = &dockerclient.EndpointConfig{
				IPAMConfig: &dockerclient.EndpointIPAMConfig{
					IPv4Address: containerTask.NetworkIPAddr,
				},
			}

			option.NetworkingConfig = &dockerclient.NetworkingConfig{
				EndpointsConfig: EndpointsConfig,
			}
		}
	}

	fmt.Fprintf(os.Stdout, "ready to create container: %s, image: %s\n", containerName, config.Image)
	by, _ := json.Marshal(option)
	logs.Infof("create container %s data %s", containerName, string(by))
	dockerContainer, err := docker.client.CreateContainer(option)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create Container %s failed: %s\n", containerName, err.Error())
		return nil, err
	}
	bcs := &BcsContainerInfo{
		ID:   dockerContainer.ID,
		Name: dockerContainer.Name,
	}

	fmt.Fprintf(os.Stdout, "Success to create container(ID:%s)\n", dockerContainer.ID)
	return bcs, nil
}

//StartContainer starting container with container id return by CreateContainer
func (docker *DockerContainer) StartContainer(containerID string) error {
	//to start container
	fmt.Fprintf(os.Stdout, "begin to start container(ID:%s)\n", containerID)
	err := docker.client.StartContainer(containerID, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Start Container(ID:%s) failed: %s\n", containerID, err.Error())
		return err
	}
	fmt.Fprintf(os.Stdout, "Success to start container(ID:%s)\n", containerID)
	return nil
}

//StopContainer with container name. container will be killed when timeout
func (docker *DockerContainer) StopContainer(containerName string, timeout int) error {
	duration := time.Duration(timeout + StopContainerGraceTime)
	ticker := time.Tick(time.Second * duration)

	done := make(chan error, 1)
	func() {
		fmt.Fprintf(os.Stdout, "start stop container %s\n", containerName)
		err := docker.client.StopContainer(containerName, uint(timeout))
		fmt.Fprintf(os.Stdout, "stop container %s done\n", containerName)
		done <- err
	}()

	select {
	case <-ticker:
		fmt.Fprintf(os.Stdout, "stop container %s timeout\n", containerName)
		return fmt.Errorf("stop container %s timeout", containerName)
	case err := <-done:
		return err
	}
}

//RemoveContainer remove container by name
func (docker *DockerContainer) RemoveContainer(containerName string, force bool) error {
	opt := dockerclient.RemoveContainerOptions{
		ID:    containerName,
		Force: force,
	}
	if err := docker.client.RemoveContainer(opt); err != nil {
		fmt.Fprintf(os.Stderr, "Remove Container %s failed, err %s\n", containerName, err.Error())
		return err
	}
	fmt.Fprintf(os.Stdout, "Success to remove container %s\n", containerName)
	return nil
}

//KillContainer kill container by name with designate signal
func (docker *DockerContainer) KillContainer(containerName string, signal int) error {
	option := dockerclient.KillContainerOptions{
		ID:     containerName,
		Signal: dockerclient.SIGKILL,
	}
	return docker.client.KillContainer(option)
}

//InspectContainer inspect container by name
func (docker *DockerContainer) InspectContainer(containerName string) (*BcsContainerInfo, error) {
	container, err := docker.client.InspectContainer(containerName)
	if err != nil {
		return nil, err
	}

	bcsContainer := &BcsContainerInfo{
		ID:          container.ID,
		Name:        container.Name,
		Pid:         container.State.Pid,
		StartAt:     container.State.StartedAt,
		FinishAt:    container.State.FinishedAt,
		Status:      container.State.Status,
		ExitCode:    container.State.ExitCode,
		Hostname:    container.Config.Hostname,
		NetworkMode: container.HostConfig.NetworkMode,
		OOMKilled:   container.State.OOMKilled,
		Resource: &schedTypes.Resource{
			Cpus: float64(container.HostConfig.CPUShares / 1024),
			Mem:  float64(container.HostConfig.Memory / 1024 / 1024),
		},
	}
	bcsContainer.Name = strings.Replace(bcsContainer.Name, "/", "", 1)
	networkMap := container.NetworkSettings.Networks
	if networkMap != nil && len(networkMap) > 0 {
		//ip address is in networkMap
		//fix: change default to bridge
		if bcsContainer.NetworkMode == "default" {
			bcsContainer.NetworkMode = "bridge"
		}
		if info, ok := networkMap[bcsContainer.NetworkMode]; ok {
			bcsContainer.IPAddress = info.IPAddress
		}
		//Get no IPAddress for this
	}
	return bcsContainer, nil
}

//PullImage pull image from hub
func (docker *DockerContainer) PullImage(image string) error {
	repo, tag := dockerclient.ParseRepositoryTag(image)
	pullOpt := dockerclient.PullImageOptions{
		Repository:        repo,
		Tag:               tag,
		InactivityTimeout: time.Minute * 3,
	}
	auth := dockerclient.AuthConfiguration{
		Username: docker.user,
		Password: docker.passwd,
	}
	return docker.client.PullImage(pullOpt, auth)
}

//ListImage list all image from local
func (docker *DockerContainer) ListImage(filter string) ([]*BcsImage, error) {
	opt := dockerclient.ListImagesOptions{
		All:    true,
		Filter: filter,
	}
	var bcsImages []*BcsImage
	images, listErr := docker.client.ListImages(opt)
	if listErr != nil {
		return bcsImages, listErr
	}
	for _, image := range images {
		bcs := &BcsImage{
			ID:         image.ID,
			Repository: image.RepoTags,
			Created:    image.Created,
			Size:       image.Size,
		}
		bcsImages = append(bcsImages, bcs)
	}
	return bcsImages, nil
}

// UpdateResources update container resource in runtime
func (docker *DockerContainer) UpdateResources(id string, resource *schedTypes.TaskResources) error {
	if resource == nil {
		return fmt.Errorf("container resource to update cannot be empty")
	}
	if *resource.Cpu < 0 || *resource.Mem < 4 || *resource.ReqCpu < 0 || *resource.ReqMem < 4 {
		return fmt.Errorf("container resource reqCpu %f reqMem %f cpu %f memory %f is invalid",
			*resource.ReqCpu, *resource.ReqMem, *resource.Cpu, *resource.Mem)
	}

	logs.Infof("update container %s resources cpu %f mem %f", id, *resource.Cpu, *resource.Mem)

	options := dockerclient.UpdateContainerOptions{
		CPUShares:  int(*resource.ReqCpu * 1024),
		CPUPeriod:  DefaultDockerCPUPeriod,
		CPUQuota:   int(*resource.Cpu * DefaultDockerCPUPeriod),
		Memory:     int(*resource.Mem * 1024 * 1024),
		MemorySwap: int(*resource.Mem * 1024 * 1024),
	}

	return docker.client.UpdateContainer(id, options)
}

// CommitImage create image from running container
func (docker *DockerContainer) CommitImage(id, image string) error {
	repo, tag := dockerclient.ParseRepositoryTag(image)

	options := dockerclient.CommitContainerOptions{
		Container:  id,
		Repository: repo,
		Tag:        tag,
	}

	_, err := docker.client.CommitContainer(options)
	if err != nil {
		return err
	}

	opts := dockerclient.PushImageOptions{
		Name: repo,
		Tag:  tag,
	}

	auth := dockerclient.AuthConfiguration{
		Username: docker.user,
		Password: docker.passwd,
	}

	return docker.client.PushImage(opts, auth)
}
