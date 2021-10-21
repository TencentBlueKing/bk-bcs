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

package register

import "errors"

var (
	//ErrNotExist data not exist err
	ErrNotExist = errors.New("resource does not exist")
)

//Register interface for gateway-discovery to register all necessary
//bcs services to specified api-gateway
type Register interface {
	//CreateService create Service interface, if service already exists, return error
	CreateService(svc *Service) error
	//UpdateService update specified Service, if service does not exist, return error
	UpdateService(svc *Service) error
	//GetService get specified service by name, if no service, return nil
	GetService(svc string) (*Service, error)
	//DeleteService delete specified service, success even if no such service
	DeleteService(svc *Service) error
	//ListServices get all existence services
	ListServices() ([]*Service, error)
	//GetTargetByService get service relative backends
	GetTargetByService(svc *Service) ([]Backend, error)
	//ReplaceTargetByService replace specified service backend list
	// so we don't care what original backend list are
	ReplaceTargetByService(svc *Service, backends []Backend) error
	//DeleteTargetByService clean all backend list for service
	DeleteTargetByService(svc *Service) error
	//GetRoutesByService(name string) ([]Route, error)
}
