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

package expectations

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

// ScaleAction is the action of scale, like create and delete.
type ScaleAction string

const (
	// Create action
	Create ScaleAction = "create"
	// Delete action
	Delete ScaleAction = "delete"
)

// ScaleExpectations is an interface that allows users to set and wait on expectations of pods scale.
type ScaleExpectations interface {
	ExpectScale(controllerKey string, action ScaleAction, name string)
	ObserveScale(controllerKey string, action ScaleAction, name string)
	SatisfiedExpectations(controllerKey string) (bool, map[ScaleAction][]string)
	DeleteExpectations(controllerKey string)
	GetExpectations(controllerKey string) map[ScaleAction]sets.String
}

// NewScaleExpectations returns a common ScaleExpectations.
func NewScaleExpectations() ScaleExpectations {
	return &realScaleExpectations{
		controllerCache: make(map[string]*realControllerScaleExpectations),
	}
}

type realScaleExpectations struct {
	sync.RWMutex
	// key: parent key, workload namespace/name
	controllerCache map[string]*realControllerScaleExpectations
}

type realControllerScaleExpectations struct {
	// item: name for this object
	objsCache map[ScaleAction]sets.String
}

func (r *realScaleExpectations) GetExpectations(controllerKey string) map[ScaleAction]sets.String {
	r.Lock()
	defer r.Unlock()

	expectations := r.controllerCache[controllerKey]
	if expectations == nil {
		return nil
	}

	res := make(map[ScaleAction]sets.String, len(expectations.objsCache))
	for k, v := range expectations.objsCache {
		res[k] = sets.NewString(v.List()...)
	}

	return res
}

func (r *realScaleExpectations) ExpectScale(controllerKey string, action ScaleAction, name string) {
	r.Lock()
	defer r.Unlock()

	expectations := r.controllerCache[controllerKey]
	if expectations == nil {
		expectations = &realControllerScaleExpectations{
			objsCache: make(map[ScaleAction]sets.String),
		}
		r.controllerCache[controllerKey] = expectations
	}

	if s := expectations.objsCache[action]; s != nil {
		s.Insert(name)
	} else {
		expectations.objsCache[action] = sets.NewString(name)
	}
}

func (r *realScaleExpectations) ObserveScale(controllerKey string, action ScaleAction, name string) {
	r.Lock()
	defer r.Unlock()

	expectations := r.controllerCache[controllerKey]
	if expectations == nil {
		return
	}

	s := expectations.objsCache[action]
	if s == nil {
		return
	}
	s.Delete(name)

	for _, s := range expectations.objsCache {
		if s.Len() > 0 {
			return
		}
	}
	delete(r.controllerCache, controllerKey)
}

func (r *realScaleExpectations) SatisfiedExpectations(controllerKey string) (bool, map[ScaleAction][]string) {
	r.Lock()
	defer r.Unlock()

	expectations := r.controllerCache[controllerKey]
	if expectations == nil {
		return true, nil
	}

	for a, s := range expectations.objsCache {
		if s.Len() > 0 {
			return false, map[ScaleAction][]string{a: s.List()}
		}
	}
	delete(r.controllerCache, controllerKey)
	return true, nil
}

func (r *realScaleExpectations) DeleteExpectations(controllerKey string) {
	r.Lock()
	defer r.Unlock()
	delete(r.controllerCache, controllerKey)
}
