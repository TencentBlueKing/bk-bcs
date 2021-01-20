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

package offer

import (
	"container/list"
	"net/http"

	typesplugin "github.com/Tencent/bk-bcs/bcs-common/common/plugin"
	commtype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
)

//A SchedManager is a struct Scheduler, it is responsible for interacting with struct
//Offer.
type SchedManager interface {
	//GetHostAttributes is used to get variable mesos slave's attributes.
	//Examples for ip-resources, netflow.
	//If slave don't have variable attributes, it return nil
	GetHostAttributes(*typesplugin.HostPluginParameter) (map[string]*typesplugin.HostAttributes, error)

	//FetchAgentSetting is used to get user custom mesos slave's attributes.
	//input is slave's ip
	FetchAgentSetting(string) (*commtype.BcsClusterAgentSetting, error)

	//FetchAgentSchedInfo is used to get agent DeltaCPU, DeltaDisk, DeltaMem
	//input is slave's hostname
	FetchAgentSchedInfo(string) (*types.AgentSchedInfo, error)

	//Get mesos cluster id
	GetClusterId() string

	//decline mesos slave's offer, mesos will resubmit this offer after a few
	//seconds.
	//input is mesos offer's id
	DeclineResource(*string) (*http.Response, error)

	//fetch taskgroup
	FetchTaskGroup(taskGroupID string) (*types.TaskGroup, error)

	//update mesos agents
	UpdateMesosAgents()
}

//OfferPool is mesos offer pool, it is responsible for the managements of the mesos's offers.
//OfferPool maintains an ordered list of offers, we can use it by the following functions.
type OfferPool interface {
	//get the first offer from the offer pool.
	//if the offer pool don't have offer,it return nil
	GetFirstOffer() *Offer

	//get the specified offer's next offer.
	//if the offer don't have next offer,it return nil
	GetNextOffer(*Offer) *Offer

	//get all valid offers at the moment
	GetAllOffers() []*Offer

	//the offer list is a sequence sorted by id.
	//this function can return the repecified id's next offer.
	//if not have, it return nil
	//GetOfferGreaterThan(id int64) *Offer

	//scheduler can use the offer by function UseOffer.
	//offer pool don't manage the offer after scheduler use it.
	//at concurrency,it is possible that multiple threads will use the same offer.
	//so only return true, indicating use the offer successful.
	UseOffer(*Offer) bool

	//add mesos's offers in offer pool
	AddOffers([]*mesos.Offer) error

	//when mesos slave lost, the slave need grace period to recover after re-registration.
	//so sign the lost slave, it's offer is invalid for the moment
	AddLostSlave(string)

	//get offer pool's length
	GetOffersLength() int
}

type Offer struct {
	// offer id, int64
	Id int64

	element  *list.Element
	offerId  string
	hostname string

	//mesos slave offer
	Offer *mesos.Offer

	DeltaCPU  float64
	DeltaMem  float64
	DeltaDisk float64
}

//NewOfferPool's input parameter.
type OfferPara struct {
	// struct SchedManager
	Sched SchedManager

	//LostSlaveGracePeriod
	//if you don't specify, it will the const DefaultLostSlaveGracePeriod
	LostSlaveGracePeriod int

	//DefaultLostSlaveGracePeriod
	//if you don't specify, it will the const DefaultOfferLifePeriod
	OfferlifePeriod int

	//store
	Store store.Store
}
