/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog"
)

var (
	defaultAsgCacheTTL = 3600 // Seconds
	virtualMachineRE   = regexp.MustCompile(`^azure://(?:.*)/providers/Microsoft.Compute/virtualMachines/(.+)$`)
)

type asgCache struct {
	registeredAsgs     []cloudprovider.NodeGroup
	instanceToAsg      map[azureRef]cloudprovider.NodeGroup
	notInRegisteredAsg map[azureRef]bool
	mutex              sync.Mutex
	interrupt          chan struct{}
}

func newAsgCache(asgCacheTTL int64) (*asgCache, error) {
	cache := &asgCache{
		registeredAsgs:     make([]cloudprovider.NodeGroup, 0),
		instanceToAsg:      make(map[azureRef]cloudprovider.NodeGroup),
		notInRegisteredAsg: make(map[azureRef]bool),
		interrupt:          make(chan struct{}),
	}

	go wait.Until(func() {
		cache.mutex.Lock()
		defer cache.mutex.Unlock()
		if err := cache.regenerate(); err != nil {
			klog.Errorf("Error while regenerating Asg cache: %v", err)
		}
	}, time.Duration(asgCacheTTL)*time.Second, cache.interrupt)

	return cache, nil
}

// Register registers a node group if it hasn't been registered.
func (m *asgCache) Register(asg cloudprovider.NodeGroup) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i := range m.registeredAsgs {
		if existing := m.registeredAsgs[i]; strings.EqualFold(existing.Id(), asg.Id()) {
			if reflect.DeepEqual(existing, asg) {
				return false
			}

			m.registeredAsgs[i] = asg
			klog.V(4).Infof("ASG %q updated", asg.Id())
			m.invalidateUnownedInstanceCache()
			return true
		}
	}

	klog.V(4).Infof("Registering ASG %q", asg.Id())
	m.registeredAsgs = append(m.registeredAsgs, asg)
	m.invalidateUnownedInstanceCache()
	return true
}

func (m *asgCache) invalidateUnownedInstanceCache() {
	klog.V(4).Info("Invalidating unowned instance cache")
	m.notInRegisteredAsg = make(map[azureRef]bool)
}

// Unregister ASG. Returns true if the ASG was unregistered.
func (m *asgCache) Unregister(asg cloudprovider.NodeGroup) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	updated := make([]cloudprovider.NodeGroup, 0, len(m.registeredAsgs))
	changed := false
	for _, existing := range m.registeredAsgs {
		if strings.EqualFold(existing.Id(), asg.Id()) {
			klog.V(1).Infof("Unregistered ASG %s", asg.Id())
			changed = true
			continue
		}
		updated = append(updated, existing)
	}
	m.registeredAsgs = updated
	return changed
}

func (m *asgCache) get() []cloudprovider.NodeGroup {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.registeredAsgs
}

// FindForInstance returns Asg of the given Instance
func (m *asgCache) FindForInstance(instance *azureRef, vmType string) (cloudprovider.NodeGroup, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.V(4).Infof("FindForInstance: starts, ref: %s", instance.Name)
	resourceID, err := convertResourceGroupNameToLower(instance.Name)
	klog.V(4).Infof("FindForInstance: resourceID: %s", resourceID)
	if err != nil {
		return nil, err
	}
	inst := azureRef{Name: resourceID}
	if m.notInRegisteredAsg[inst] {
		// We already know we don't own this instance. Return early and avoid
		// additional calls.
		klog.V(4).Infof("FindForInstance: Couldn't find NodeGroup of instance %q", inst)
		return nil, nil
	}

	if vmType == vmTypeVMSS {
		// Omit virtual machines not managed by vmss.
		if ok := virtualMachineRE.Match([]byte(inst.Name)); ok {
			klog.V(3).Infof("Instance %q is not managed by vmss, omit it in autoscaler", instance.Name)
			m.notInRegisteredAsg[inst] = true
			return nil, nil
		}
	}

	if vmType == vmTypeStandard {
		// Omit virtual machines with providerID not in Azure resource ID format.
		if ok := virtualMachineRE.Match([]byte(inst.Name)); !ok {
			klog.V(3).Infof("Instance %q is not in Azure resource ID format, omit it in autoscaler", instance.Name)
			m.notInRegisteredAsg[inst] = true
			return nil, nil
		}
	}

	// Look up caches for the instance.
	klog.V(4).Infof("FindForInstance: attempting to retrieve instance %v from cache", m.instanceToAsg)
	if asg := m.getInstanceFromCache(inst.Name); asg != nil {
		klog.V(4).Infof("FindForInstance: found asg %s in cache", asg.Id())
		return asg, nil
	}
	klog.V(4).Infof("FindForInstance: Couldn't find NodeGroup of instance %q", inst)
	return nil, nil
}

// Cleanup closes the channel to signal the go routine to stop that is handling the cache
func (m *asgCache) Cleanup() {
	close(m.interrupt)
}

func (m *asgCache) regenerate() error {
	newCache := make(map[azureRef]cloudprovider.NodeGroup)

	for _, nsg := range m.registeredAsgs {
		klog.V(6).Infof("regenerate: finding nodes for nsg %+v", nsg)
		instances, err := nsg.Nodes()
		if err != nil {
			return err
		}
		klog.V(6).Infof("regenerate: found nodes for nsg %v: %+v", nsg, instances)

		for _, instance := range instances {
			ref := azureRef{Name: instance.Id}
			newCache[ref] = nsg
		}
	}

	m.instanceToAsg = newCache

	// Incalidating unowned instance cache.
	m.invalidateUnownedInstanceCache()

	return nil
}

// Get node group from cache. nil would be return if not found.
// Should be call with lock protected.
func (m *asgCache) getInstanceFromCache(providerID string) cloudprovider.NodeGroup {
	for instanceID, asg := range m.instanceToAsg {
		if strings.EqualFold(instanceID.GetKey(), providerID) {
			return asg
		}
	}

	return nil
}
