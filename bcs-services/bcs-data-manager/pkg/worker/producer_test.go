/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package worker

import (
	"context"
	"testing"

	"github.com/robfig/cron/v3"
)

func TestProducer_ClusterProducer(t *testing.T) {

}

func TestProducer_NamespaceProducer(t *testing.T) {

}

func TestProducer_ProjectProducer(t *testing.T) {

}

func TestProducer_PublicProducer(t *testing.T) {

}

func TestProducer_Run(t *testing.T) {
	ctx := context.Background()
	newcron := cron.New()
	p := &Producer{
		cron: newcron,
		ctx:  ctx,
	}
	p.Run()
}

func TestProducer_SendJob(t *testing.T) {

}

func TestProducer_WorkloadProducer(t *testing.T) {

}

func TestProducer_getCronList(t *testing.T) {

}
