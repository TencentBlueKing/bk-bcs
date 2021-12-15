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

package gce

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	gce "google.golang.org/api/compute/v1"
	"k8s.io/klog"
)

// MachineTypeKey is used to identify MachineType.
type MachineTypeKey struct {
	Zone        string
	MachineType string
}

// GceCache is used for caching cluster resources state.
//
// It is needed to:
// - keep track of autoscaled MIGs in the cluster,
// - keep track of instances and which MIG they belong to,
// - limit repetitive GCE API calls.
//
// Cached resources:
// 1) MIG configuration,
// 2) instance->MIG mapping,
// 3) resource limits (self-imposed quotas),
// 4) machine types.
//
// How it works:
// - migs (1), resource limits (3) and machine types (4) are only stored in this cache,
// not updated by it.
// - instanceRefToMigRef (2) is based on registered migs (1). For each mig, its instances
// are fetched from GCE API using gceService.
// - instanceRefToMigRef (2) is NOT updated automatically when migs field (1) is updated. Calling
// RegenerateInstancesCache is required to sync it with registered migs.
type GceCache struct {
	cacheMutex sync.Mutex

	// Cache content.
	migs                   map[GceRef]Mig
	instanceRefToMigRef    map[GceRef]GceRef
	resourceLimiter        *cloudprovider.ResourceLimiter
	machinesCache          map[MachineTypeKey]*gce.MachineType
	migTargetSizeCache     map[GceRef]int64
	migBaseNameCache       map[GceRef]string
	instanceTemplatesCache map[GceRef]*gce.InstanceTemplate

	// Service used to refresh cache.
	GceService AutoscalingGceClient
}

// NewGceCache creates empty GceCache.
func NewGceCache(gceService AutoscalingGceClient) *GceCache {
	return &GceCache{
		migs:                   map[GceRef]Mig{},
		instanceRefToMigRef:    map[GceRef]GceRef{},
		machinesCache:          map[MachineTypeKey]*gce.MachineType{},
		migTargetSizeCache:     map[GceRef]int64{},
		migBaseNameCache:       map[GceRef]string{},
		instanceTemplatesCache: map[GceRef]*gce.InstanceTemplate{},
		GceService:             gceService,
	}
}

//  Methods locking on migsMutex.

// RegisterMig returns true if the node group wasn't in cache before, or its config was updated.
func (gc *GceCache) RegisterMig(newMig Mig) bool {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	oldMig, found := gc.migs[newMig.GceRef()]
	if found {
		if !reflect.DeepEqual(oldMig, newMig) {
			gc.migs[newMig.GceRef()] = newMig
			klog.V(4).Infof("Updated Mig %s", newMig.GceRef().String())
			return true
		}
		return false
	}

	klog.V(1).Infof("Registering %s", newMig.GceRef().String())
	gc.migs[newMig.GceRef()] = newMig
	return true
}

// UnregisterMig returns true if the node group has been removed, and false if it was already missing from cache.
func (gc *GceCache) UnregisterMig(toBeRemoved Mig) bool {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	_, found := gc.migs[toBeRemoved.GceRef()]
	if found {
		klog.V(1).Infof("Unregistered Mig %s", toBeRemoved.GceRef().String())
		delete(gc.migs, toBeRemoved.GceRef())
		gc.removeInstancesForMig(toBeRemoved.GceRef())
		return true
	}
	return false
}

// GetMigs returns a copy of migs list.
func (gc *GceCache) GetMigs() []Mig {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	migs := make([]Mig, 0, len(gc.migs))
	for _, mig := range gc.migs {
		migs = append(migs, mig)
	}
	return migs
}

// GetMigs returns a copy of migs list.
func (gc *GceCache) getMigRefs() []GceRef {
	migRefs := make([]GceRef, 0, len(gc.migs))
	for migRef := range gc.migs {
		migRefs = append(migRefs, migRef)
	}
	return migRefs
}

// Methods locking on cacheMutex.

// GetMigForInstance returns Mig to which the given instance belongs.
// Attempts to regenerate cache if there is a Mig with matching prefix in migs list.
// TODO(aleksandra-malinowska): reconsider failing when there's a Mig with
// matching prefix, but instance doesn't belong to it.
func (gc *GceCache) GetMigForInstance(instanceRef GceRef) (Mig, error) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if migRef, found := gc.instanceRefToMigRef[instanceRef]; found {
		mig, found := gc.getMigNoLock(migRef)
		if !found {
			return nil, fmt.Errorf("instance %+v belongs to unregistered mig %+v", instanceRef, migRef)
		}
		return mig, nil
	}

	for _, migRef := range gc.getMigRefs() {

		// get mig basename - refresh if not found
		// todo[lukaszos] move this one as well as whole instance cache regeneration out of cache
		migBasename, found := gc.migBaseNameCache[migRef]
		var err error
		if !found {
			migBasename, err = gc.GceService.FetchMigBasename(migRef)
			if err != nil {
				return nil, err
			}
			gc.migBaseNameCache[migRef] = migBasename
		}

		if migRef.Project == instanceRef.Project &&
			migRef.Zone == instanceRef.Zone &&
			strings.HasPrefix(instanceRef.Name, migBasename) {
			if err := gc.regenerateInstanceCacheForMigNoLock(migRef); err != nil {
				return nil, fmt.Errorf("error while looking for MIG for instance %+v, error: %v", instanceRef, err)
			}

			migRef, found := gc.instanceRefToMigRef[instanceRef]
			if !found {
				return nil, fmt.Errorf("instance %+v belongs to unknown mig", instanceRef)
			}
			mig, found := gc.getMigNoLock(migRef)
			if !found {
				return nil, fmt.Errorf("instance %+v belongs to unregistered mig %+v", instanceRef, migRef)
			}
			return mig, nil
		}
	}
	// Instance doesn't belong to any configured mig.
	return nil, nil
}

func (gc *GceCache) removeInstancesForMig(migRef GceRef) {
	for instanceRef, instanceMigRef := range gc.instanceRefToMigRef {
		if migRef == instanceMigRef {
			delete(gc.instanceRefToMigRef, instanceRef)
		}
	}
}

func (gc *GceCache) getMigNoLock(migRef GceRef) (mig Mig, found bool) {
	mig, found = gc.migs[migRef]
	return
}

// RegenerateInstanceCacheForMig triggers instances cache regeneration for single MIG under lock.
func (gc *GceCache) RegenerateInstanceCacheForMig(migRef GceRef) error {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	return gc.regenerateInstanceCacheForMigNoLock(migRef)
}

func (gc *GceCache) regenerateInstanceCacheForMigNoLock(migRef GceRef) error {
	klog.V(4).Infof("Regenerating MIG information for %s", migRef.String())

	// cleanup old entries
	gc.removeInstancesForMig(migRef)

	instances, err := gc.GceService.FetchMigInstances(migRef)
	if err != nil {
		klog.V(4).Infof("Failed MIG info request for %s: %v", migRef.String(), err)
		return err
	}
	for _, instance := range instances {
		instanceRef, err := GceRefFromProviderId(instance.Id)
		if err != nil {
			return err
		}
		gc.instanceRefToMigRef[instanceRef] = migRef
	}
	return nil
}

// RegenerateInstancesCache triggers instances cache regeneration under lock.
func (gc *GceCache) RegenerateInstancesCache() error {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.instanceRefToMigRef = make(map[GceRef]GceRef)
	for _, migRef := range gc.getMigRefs() {
		err := gc.regenerateInstanceCacheForMigNoLock(migRef)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetResourceLimiter sets resource limiter.
func (gc *GceCache) SetResourceLimiter(resourceLimiter *cloudprovider.ResourceLimiter) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.resourceLimiter = resourceLimiter
}

// GetResourceLimiter returns resource limiter.
func (gc *GceCache) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	return gc.resourceLimiter, nil
}

// GetMigTargetSize returns the cached targetSize for a GceRef
func (gc *GceCache) GetMigTargetSize(ref GceRef) (int64, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	size, found := gc.migTargetSizeCache[ref]
	if found {
		klog.V(5).Infof("Target size cache hit for %s", ref)
	}
	return size, found
}

// SetMigTargetSize sets targetSize for a GceRef
func (gc *GceCache) SetMigTargetSize(ref GceRef, size int64) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.migTargetSizeCache[ref] = size
}

// InvalidateMigTargetSize clears the target size cache
func (gc *GceCache) InvalidateMigTargetSize(ref GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if _, found := gc.migTargetSizeCache[ref]; found {
		klog.V(5).Infof("Target size cache invalidated for %s", ref)
		delete(gc.migTargetSizeCache, ref)
	}
}

// InvalidateAllMigTargetSizes clears the target size cache
func (gc *GceCache) InvalidateAllMigTargetSizes() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("Target size cache invalidated")
	gc.migTargetSizeCache = map[GceRef]int64{}
}

// GetMigInstanceTemplate returns the cached gce.InstanceTemplate for a mig GceRef
func (gc *GceCache) GetMigInstanceTemplate(ref GceRef) (*gce.InstanceTemplate, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	instanceTemplate, found := gc.instanceTemplatesCache[ref]
	if found {
		klog.V(5).Infof("Instance template cache hit for %s", ref)
	}
	return instanceTemplate, found
}

// SetMigInstanceTemplate sets gce.InstanceTemplate for a mig GceRef
func (gc *GceCache) SetMigInstanceTemplate(ref GceRef, instanceTemplate *gce.InstanceTemplate) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.instanceTemplatesCache[ref] = instanceTemplate
}

// InvalidateMigInstanceTemplate clears the instance template cache for a mig GceRef
func (gc *GceCache) InvalidateMigInstanceTemplate(ref GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if _, found := gc.instanceTemplatesCache[ref]; found {
		klog.V(5).Infof("Instance template cache invalidated for %s", ref)
		delete(gc.instanceTemplatesCache, ref)
	}
}

// InvalidateAllMigInstanceTemplates clears the instance template cache
func (gc *GceCache) InvalidateAllMigInstanceTemplates() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("Instance template cache invalidated")
	gc.instanceTemplatesCache = map[GceRef]*gce.InstanceTemplate{}
}

// GetMachineFromCache retrieves machine type from cache under lock.
func (gc *GceCache) GetMachineFromCache(machineType string, zone string) *gce.MachineType {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	return gc.machinesCache[MachineTypeKey{zone, machineType}]
}

// AddMachineToCache adds machine to cache under lock.
func (gc *GceCache) AddMachineToCache(machineType string, zone string, machine *gce.MachineType) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.machinesCache[MachineTypeKey{zone, machineType}] = machine
}

// SetMachinesCache sets the machines cache under lock.
func (gc *GceCache) SetMachinesCache(machinesCache map[MachineTypeKey]*gce.MachineType) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.machinesCache = machinesCache
}

// SetMigBasename sets basename for given mig in cache
func (gc *GceCache) SetMigBasename(migRef GceRef, basename string) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	gc.migBaseNameCache[migRef] = basename
}

// GetMigBasename get basename for given mig from cache.
func (gc *GceCache) GetMigBasename(migRef GceRef) (basename string, found bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	basename, found = gc.migBaseNameCache[migRef]
	return
}

// InvalidateMigBasename invalidates basename entry for given mig.
func (gc *GceCache) InvalidateMigBasename(migRef GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	delete(gc.migBaseNameCache, migRef)
}

// InvalidateAllMigBasenames invalidates all basename entries.
func (gc *GceCache) InvalidateAllMigBasenames() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	gc.migBaseNameCache = make(map[GceRef]string)
}
