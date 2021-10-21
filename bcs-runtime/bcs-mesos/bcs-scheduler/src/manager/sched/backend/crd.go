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

package backend

import (
	"fmt"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

//custom resource register
func (b *backend) RegisterCustomResource(crr *commtypes.Crr) error {
	crrs, err := b.store.ListCustomResourceRegister()
	if err != nil {
		return err
	}

	for _, c := range crrs {
		if c.Spec.Names.Kind == crr.Spec.Names.Kind {
			return nil
		}
	}

	return b.store.SaveCustomResourceRegister(crr)
}

func (b *backend) UnregisterCustomResource(string) error {
	//TODO

	return nil
}

func (b *backend) CreateCustomResource(crd *commtypes.Crd) error {
	crrs, err := b.store.ListCustomResourceRegister()
	if err != nil {
		return err
	}

	var kindOk bool
	for _, c := range crrs {
		if c.Spec.Names.Kind == string(crd.Kind) {
			kindOk = true
		}
	}

	if !kindOk {
		return fmt.Errorf("custom resource kind %s is invalid", crd.Kind)
	}

	return b.store.SaveCustomResourceDefinition(crd)
}

func (b *backend) UpdateCustomResource(crd *commtypes.Crd) error {
	crrs, err := b.store.ListCustomResourceRegister()
	if err != nil {
		return err
	}

	var kindOk bool
	for _, c := range crrs {
		if c.Spec.Names.Kind == string(crd.Kind) {
			kindOk = true
		}
	}

	if !kindOk {
		return fmt.Errorf("custom resource kind %s is invalid", crd.Kind)
	}

	return b.store.SaveCustomResourceDefinition(crd)
}

func (b *backend) DeleteCustomResource(kind, ns, name string) error {
	return b.store.DeleteCustomResourceDefinition(kind, ns, name)
}

func (b *backend) ListCustomResourceDefinition(kind, ns string) ([]*commtypes.Crd, error) {
	return b.store.ListCustomResourceDefinition(kind, ns)
}

func (b *backend) ListAllCrds(kind string) ([]*commtypes.Crd, error) {
	return b.store.ListAllCrds(kind)
}

func (b *backend) FetchCustomResourceDefinition(kind, ns, name string) (*commtypes.Crd, error) {
	return b.store.FetchCustomResourceDefinition(kind, ns, name)
}
