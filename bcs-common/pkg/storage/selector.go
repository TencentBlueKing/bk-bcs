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

package storage

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"strings"
)

//Selector is a filter when reading data from event-storage
//if data object is matched, object will push to watch tunnel
type Selector interface {
	String() string
	Matchs(obj meta.Object) (bool, error)
	MatchList(objs []meta.Object) ([]meta.Object, error)
}

//NewSelectors create multiple selector
func NewSelectors(s ...Selector) Selector {
	ss := &Selectors{}
	for _, se := range s {
		if se == nil {
			continue
		}
		ss.list = append(ss.list, se)
	}
	if len(ss.list) == 0 {
		return &Everything{}
	}
	return ss
}

//Selectors compose multiple Selector
type Selectors struct {
	list []Selector
}

//String match string
func (ss *Selectors) String() string {
	var strs []string
	for _, s := range ss.list {
		strs = append(strs, s.String())
	}
	return strings.Join(strs, "|")
}

//Matchs match object
func (ss *Selectors) Matchs(obj meta.Object) (bool, error) {
	if obj == nil {
		return false, nil
	}
	for _, s := range ss.list {
		ok, err := s.Matchs(obj)
		if err != nil {
			return false, err
		}
		if ok {
			continue
		}
		return false, nil
	}
	return true, nil
}

//MatchList match object list
func (ss *Selectors) MatchList(objs []meta.Object) ([]meta.Object, error) {
	var targets []meta.Object
	for _, obj := range objs {
		if ok, _ := ss.Matchs(obj); ok {
			targets = append(targets, obj)
		}
	}
	return targets, nil
}

//Everything filter nothing
type Everything struct{}

//String match string
func (e *Everything) String() string {
	return "Everything"
}

//Matchs match object
func (e *Everything) Matchs(obj meta.Object) (bool, error) {
	if obj == nil {
		return false, nil
	}
	return true, nil
}

//MatchList match object list
func (e *Everything) MatchList(objs []meta.Object) ([]meta.Object, error) {
	return objs, nil
}

//LabelAsSelector create selector from labels
func LabelAsSelector(l meta.Labels) Selector {
	if l != nil {
		return &LabelSelector{
			labels: l,
		}
	}
	return &Everything{}
}

//LabelSelector implements selector interface with Labels
type LabelSelector struct {
	labels meta.Labels
}

//String match string
func (ls *LabelSelector) String() string {
	//modified by marsjma
	//return ls.labels.String()
	return "labelSelector=" + ls.labels.String()
}

//Matchs match object
func (ls *LabelSelector) Matchs(obj meta.Object) (bool, error) {
	if obj == nil {
		return false, nil
	}
	if ls.labels == nil {
		return true, nil
	}
	target := obj.GetLabels()
	if target == nil {
		return false, nil
	}
	matched := meta.LabelsAllMatch(ls.labels, meta.Labels(target))
	return matched, nil
}

//MatchList match object list
func (ls *LabelSelector) MatchList(objs []meta.Object) ([]meta.Object, error) {
	var targets []meta.Object
	for _, obj := range objs {
		if ok, _ := ls.Matchs(obj); ok {
			targets = append(targets, obj)
		}
	}
	return targets, nil
}

//NamespaceSelector select expected namespace
type NamespaceSelector struct {
	Namespace string
}

//String match string
func (ns *NamespaceSelector) String() string {
	return ns.Namespace
}

//Matchs match object
func (ns *NamespaceSelector) Matchs(obj meta.Object) (bool, error) {
	if obj == nil {
		return false, nil
	}
	target := obj.GetNamespace()
	if target == ns.Namespace {
		return true, nil
	}
	return false, nil
}

//MatchList match object list
func (ns *NamespaceSelector) MatchList(objs []meta.Object) ([]meta.Object, error) {
	var targets []meta.Object
	for _, obj := range objs {
		if ok, _ := ns.Matchs(obj); ok {
			targets = append(targets, obj)
		}
	}
	return targets, nil
}
