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
	"fmt"
	"net/http"
	"testing"
	"time"

	typesplugin "bk-bcs/bcs-common/common/plugin"
	commtype "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/mesosproto/mesos"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

const (
	LostSlaveOfferId     = "lostoffer1"
	LostSlaveHostname    = "lostofferhostname1"
	DefaultSleepDuration = 50 * time.Millisecond

	OfferLifePeriod      = 60
	LostSlaveGracePeriod = 1
)

type scheduler struct{}

func (s *scheduler) GetHostAttributes(para *typesplugin.HostPluginParameter) (map[string]*typesplugin.HostAttributes, error) {
	attrs := make(map[string]*typesplugin.HostAttributes, 0)

	attr := &typesplugin.Attribute{
		Name:   "ip-resources",
		Type:   typesplugin.ValueScalar,
		Scalar: typesplugin.Value_Scalar{Value: 1},
	}

	hostAttr := &typesplugin.HostAttributes{
		Ip:         "127.0.0.1",
		Attributes: make([]*typesplugin.Attribute, 0),
	}

	hostAttr.Attributes = append(hostAttr.Attributes, attr)
	attrs["127.0.0.1"] = hostAttr

	return attrs, nil
}

func (s *scheduler) FetchAgentSetting(string) (*commtype.BcsClusterAgentSetting, error) {
	return nil, nil
}

func (s *scheduler) FetchAgentSchedInfo(string) (*types.AgentSchedInfo, error) {
	return nil, nil
}

func (s *scheduler) GetClusterId() string {
	return "BCS-TESTBCSTEST01-10001"
}

func (s *scheduler) DeclineResource(*string) (*http.Response, error) {
	return nil, nil
}

var offerpool OfferPool
var offerids []string

func init() {
	para := &OfferPara{
		Sched:                &scheduler{},
		OfferlifePeriod:      OfferLifePeriod,
		LostSlaveGracePeriod: LostSlaveGracePeriod,
	}

	offerpool = NewOfferPool(para)

	offerids = []string{"offer0", "offer1", "offer2", "offer3", "offer4", "offer5",
		"offer6", "offer7", "offer8", "offer9"}
}

func getOffers() []*mesos.Offer {
	offers := make([]*mesos.Offer, 0)

	for i := 0; i < 10; i++ {
		resource := &mesos.Resource{
			Name: proto.String("cpus"),
			Type: mesos.Value_SCALAR.Enum(),
			Scalar: &mesos.Value_Scalar{
				Value: proto.Float64(float64(10 - i)),
			},
		}

		attr := &mesos.Attribute{
			Name: proto.String("InnerIP"),
			Type: mesos.Value_TEXT.Enum(),
			Text: &mesos.Value_Text{
				Value: proto.String("127.0.0.1"),
			},
		}

		offer := &mesos.Offer{
			Id: &mesos.OfferID{
				Value: proto.String(offerids[i]),
			},
			FrameworkId: nil,
			AgentId: &mesos.AgentID{
				Value: proto.String(fmt.Sprintf("agentid.%d", i)),
			},
			Hostname:   proto.String(fmt.Sprintf("hostname.%d", i)),
			Resources:  []*mesos.Resource{resource},
			Attributes: []*mesos.Attribute{attr},
		}

		offers = append(offers, offer)
	}

	return offers
}

func getLostOffer() *mesos.Offer {
	resource := &mesos.Resource{
		Name: proto.String("cpus"),
		Type: mesos.Value_SCALAR.Enum(),
		Scalar: &mesos.Value_Scalar{
			Value: proto.Float64(float64(1)),
		},
	}

	attr := &mesos.Attribute{
		Name: proto.String("InnerIP"),
		Type: mesos.Value_TEXT.Enum(),
		Text: &mesos.Value_Text{
			Value: proto.String("127.0.0.1"),
		},
	}

	offer := &mesos.Offer{
		Id: &mesos.OfferID{
			Value: proto.String(LostSlaveOfferId),
		},
		FrameworkId: nil,
		AgentId: &mesos.AgentID{
			Value: proto.String(fmt.Sprintf("agentid.1")),
		},
		Hostname:   proto.String(LostSlaveHostname),
		Resources:  []*mesos.Resource{resource},
		Attributes: []*mesos.Attribute{attr},
	}

	return offer
}

func TestAddOffers(t *testing.T) {
	//get offers and add these offers in offer pool
	offers := getOffers()
	offerpool.AddOffers(offers)

	// because func AddOffers is asynchronous, so here sleep 1 seconds
	time.Sleep(time.Second)

	// get first offer
	offer := offerpool.GetFirstOffer()
	assert.NotNil(t, offer)

	length := len(offerids)
	for i, offerid := range offerids {
		assert.Equal(t, offerid, offer.Offer.GetId().GetValue())
		offer = offerpool.GetNextOffer(offer)

		if i+1 != length {
			assert.NotNil(t, offer)
		} else {
			assert.Nil(t, offer)
		}
	}
}

func TestLostSlaveOffer(t *testing.T) {
	//add lost slave, then add this slave offer
	offerpool.AddLostSlave(LostSlaveHostname)
	offerpool.AddOffers([]*mesos.Offer{getLostOffer()})

	// because func AddOffers is asynchronous, so sleep some times
	time.Sleep(DefaultSleepDuration)

	//because the slave just lost, so this slave's offer can't be added in
	//offer pool, the slave need LostSlaveGracePeriod seconds to recover
	offer := offerpool.GetFirstOffer()
	assert.NotNil(t, offer)
	assert.NotEqual(t, LostSlaveOfferId, offer.Offer.GetId().GetValue())

	length := offerpool.GetOffersLength()
	for i := 0; i < length; i++ {
		offer = offerpool.GetNextOffer(offer)

		if i+1 != length {
			assert.NotNil(t, offer)
			assert.NotEqual(t, LostSlaveOfferId, offer.Offer.GetId().GetValue())
		} else {
			assert.Nil(t, offer)
		}
	}

	//the lost slave need LostSlaveGracePeriod times to recover,so sleep the
	//times, then we can add this slave offer in offer pool
	time.Sleep((LostSlaveGracePeriod + 1) * time.Second)

	offerpool.AddOffers([]*mesos.Offer{getLostOffer()})
	time.Sleep(DefaultSleepDuration)

	offer = offerpool.GetFirstOffer()
	assert.NotNil(t, offer)

	length = offerpool.GetOffersLength()
	var ok bool
	var lostOffer *Offer

	for i := 0; i < length; i++ {
		if offer.Offer.GetId().GetValue() == LostSlaveOfferId {
			ok = true
			lostOffer = offer
			break
		}

		offer = offerpool.GetNextOffer(offer)
		if i+1 != length {
			assert.NotNil(t, offer)
		} else {
			assert.Nil(t, offer)
		}
	}

	offerpool.UseOffer(lostOffer)
	assert.Equal(t, ok, true)
}

func TestGetAllOffers(t *testing.T) {
	offers := offerpool.GetAllOffers()
	length := offerpool.GetOffersLength()

	for i, offer := range offers {
		assert.Equal(t, offerids[i], offer.offerId)
	}

	assert.Equal(t, length, len(offers))
}
