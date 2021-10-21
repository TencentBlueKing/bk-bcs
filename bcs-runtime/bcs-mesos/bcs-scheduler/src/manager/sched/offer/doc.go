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

/*
Package offer provides mesos offers resources pool implements.

struct Offer is responsible for the managements of mesos's offers, including add, use and delete.

Offer need interface SchedManager's function to manage offers, so we need struct
SchedManager to new struct offer. Interface SchedManager have four functions, for
details please look interface.go

	para := &OfferPara{
		Sched: &SchedManager{},
	}

	offerPool := NewOfferPool(para)
	//...
	var mesosOffers []*mesos.Offer
	//...

	//add mesos's offers to offer pool
	offerPool.AddOffers(mesosOffers)

	//get the first offer
	offer := offerPool.GetFirstOffer()
	for{
		if offer == nil {
			break
		}

		// if offer is suitable, then use this offer
		ok := offerPool.UseOffer(offer)
		if ok {
			break
		}

		//else get the next offer, until you get the right one
		offer = offerPool.GetNextOffer(offer)
	}
	//...

*/
package offer
