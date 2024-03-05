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
 */

package ipt

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/controller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/iptables"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/metrics"
	docker "github.com/fsouza/go-dockerclient"
)

// iptablesHandler handler iptables for all the containers at this node.
type iptablesHandler struct {
	version string

	ipSetHandler *iptables.IPSet

	needSyncChains bool
	policyInfos    []controller.NetworkPolicyInfo

	dockerClient *docker.Client

	activeIPSet map[string]string
}

// NewHandler return iptablesHandler object
func NewHandler(dockerClient *docker.Client, ipSetHandler *iptables.IPSet,
	infos []controller.NetworkPolicyInfo, version string) *iptablesHandler {
	h := &iptablesHandler{
		dockerClient: dockerClient,
		ipSetHandler: ipSetHandler,
		version:      version,
		policyInfos:  infos,
		activeIPSet:  make(map[string]string),
	}
	return h
}

// NewCleanupHandler used to cleanup ipSets and iptables for every container
func NewCleanupHandler(dockerClient *docker.Client, ipSetHandler *iptables.IPSet) *iptablesHandler {
	return &iptablesHandler{
		dockerClient: dockerClient,
		ipSetHandler: ipSetHandler,
	}
}

// Refresh will refresh the iptables for all containers at this node.
// This func will for-loop every container,
//   - Skip pause container
//   - Skip hostNetwork container
//   - Only one container of a pod will be refreshed iptables
//
// For-loop every container with every networkPolicy, if networkPolicy
// hit the container, refresh iptables with policy for it.
func (h *iptablesHandler) Refresh() error {
	start := time.Now()
	defer func() {
		endTime := time.Since(start)
		metrics.ControllerPolicyChainsSyncTime.Observe(endTime.Seconds())
		blog.Infof("Syncing iptables chains took %v", endTime)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(controller.DockerTimeout)*time.Second)
	defer cancel()
	containers, err := h.dockerClient.ListContainers(docker.ListContainersOptions{
		Context: ctx,
	})
	if err != nil {
		blog.Errorf("List containers failed, err: %s", err.Error())
		return err
	}

	syncedPods := make(map[string]string)
	// sync iptables for all containers
	for i := range containers {
		// not care about pause container
		if containers[i].Command == controller.PauseContainerCommand {
			continue
		}
		// not care host-network container
		if networks := containers[i].Networks.Networks; networks != nil {
			if _, ok := networks["host"]; ok {
				blog.Warnf("Container %s with host network, skipped.", containers[i].Names)
				continue
			}
		}

		// filtered multi-containers of one pod, only one container in same pod
		// needs to sync iptables.
		if labels := containers[i].Labels; labels != nil {
			if podName, ok := labels["pod_name"]; ok {
				if _, ok := syncedPods[podName]; ok {
					blog.Warnf("Container %s in pod %s synced, don't need sync again.",
						containers[i].Names, podName)
					continue
				} else {
					syncedPods[podName] = podName
				}
			}
		}

		iptablesClient, err := h.createIptablesClient(&containers[i])
		if err != nil {
			return err
		}
		containerHandler := &containerHandler{
			iptablesClient: iptablesClient,
			version:        h.version,
			container:      &containers[i],
			ipSetHandler:   h.ipSetHandler,
			activeChains:   make(map[string]string),
			activeIPSets:   h.activeIPSet,
		}
		blog.Infof("Syncing for container: %s", containers[i].Names)
		if err := containerHandler.Sync(h.policyInfos); err != nil {
			return err
		}
		blog.Infof("Synced for container successfully.")
	}
	if err := h.cleanupStaleIPSets(); err != nil {
		return err
	}
	return nil
}

// Cleanup will cleanup iptables for all containers at this node.
func (h *iptablesHandler) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(controller.DockerTimeout)*time.Second)
	defer cancel()
	containers, err := h.dockerClient.ListContainers(docker.ListContainersOptions{
		Context: ctx,
	})
	if err != nil {
		blog.Errorf("List containers failed, err: %s", err.Error())
		return err
	}

	for i := range containers {
		if containers[i].Command == controller.PauseContainerCommand {
			continue
		}

		iptablesClient, err := h.createIptablesClient(&containers[i])
		if err != nil {
			return err
		}
		chains, err := iptablesClient.ListChains("filter")
		if err != nil {
			blog.Errorf("List chains failed, err: %s", err.Error())
			return err
		}
		for i := range chains {
			if !strings.HasPrefix(chains[i], controller.KubeNetworkPolicyChainPrefix) &&
				chains[i] != "INPUT" && chains[i] != "OUTPUT" {
				continue
			}
			if err := iptablesClient.ClearChain("filter", chains[i]); err != nil {
				blog.Errorf("Clear chain filter/%s failed, err: %s", chains[i], err.Error())
				return err
			}
			if chains[i] != "INPUT" && chains[i] != "OUTPUT" {
				if err := iptablesClient.DeleteChain("filter", chains[i]); err != nil {
					blog.Errorf("Delete chain filter/%s failed, err: %s", chains[i], err.Error())
					return err
				}
			}
		}
		blog.Infof("Clear iptables for container %s successfully.", containers[i].Names)
	}
	for k, set := range h.ipSetHandler.Sets {
		if !strings.HasPrefix(k, controller.KubeSourceIPSetPrefix) &&
			!strings.HasPrefix(k, controller.KubeDestinationIPSetPrefix) {
			continue
		}
		if err := set.Destroy(); err != nil {
			blog.Errorf("Destroy ipSet %s failed, err: %s", k, err.Error())
			return err
		}
		blog.Infof("Destroy ipSet %s successfully.", k)
	}
	return nil
}

// cleanupStaleIPSets used to cleanup stale ipSets after iptables synced.
func (h *iptablesHandler) cleanupStaleIPSets() error {
	for _, set := range h.ipSetHandler.Sets {
		if _, ok := h.activeIPSet[set.Name]; ok {
			continue
		}
		if !strings.HasPrefix(set.Name, controller.KubeSourceIPSetPrefix) {
			continue
		}
		if err := set.Destroy(); err != nil {
			blog.Errorf("Destroy ipSet %s failed, err: %s", set.Name, err.Error())
			return err
		}
		blog.Infof("IPSet %s is destroyed.", set.Name)
	}
	return nil
}

// createIptablesClient used to create iptablesClient for container
func (h *iptablesHandler) createIptablesClient(c *docker.APIContainers) (iptables.Interface, error) {
	inspectResult, err := h.dockerClient.InspectContainerWithOptions(docker.InspectContainerOptions{
		ID:      c.ID,
		Context: context.Background(),
	})
	if err != nil {
		blog.Errorf("Failed to inspect container: %s, err: %s", c.ID, err.Error())
		return nil, err
	}

	pid := inspectResult.State.Pid
	iptablesClient, err := iptables.NewNSIPTable(strconv.Itoa(pid))
	if err != nil {
		blog.Errorf("Failed to init iptables client, pid: %s, err: %s", pid, err.Error())
		return nil, err
	}
	return iptablesClient, nil
}
