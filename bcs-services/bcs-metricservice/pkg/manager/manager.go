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

package manager

import (
	"context"
	"hash/fnv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsMetric "github.com/Tencent/bk-bcs/bcs-common/common/metric"
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/route"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/watch"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/zk"
)

type MetricManager struct {
	watcher   *watch.Watcher
	event     chan *watch.MetricEvent
	roleEvent chan bcsMetric.RoleType

	ctx    context.Context
	cancel context.CancelFunc

	pipeline *Pipeline

	config  *config.Config
	storage storage.Storage
	route   route.Route
	zk      zk.Zk
}

func NewMetricManager(roleEvent chan bcsMetric.RoleType, config *config.Config, storage storage.Storage, route route.Route, zk zk.Zk) *MetricManager {
	return &MetricManager{roleEvent: roleEvent, config: config, storage: storage, route: route, zk: zk}
}

func (cli *MetricManager) Run() {
	cli.watcher = watch.NewWatcher(cli.storage, cli.zk, cli.config)

	for {
		select {
		case e := <-cli.roleEvent:
			blog.Infof("get new role change event: %s", e)
			switch e {
			case bcsMetric.MasterRole:
				cli.start()
			case bcsMetric.SlaveRole:
				cli.stop()
			default:
				blog.Warnf("unknown role to manager, will not change watcher state: %s", e)
			}
		}
	}
}

func (cli *MetricManager) start() {
	cli.event = cli.watcher.Start()
	cli.ctx, cli.cancel = context.WithCancel(context.Background())
	cli.pipeline = NewPipeLine(cli.ctx)
	cli.pipeline.Start()
	go cli.manager()
}

func (cli *MetricManager) stop() {
	if cli.watcher != nil {
		cli.watcher.Stop()
	}

	if cli.cancel != nil {
		cli.cancel()
	}
}

func (cli *MetricManager) manager() {
	for {
		select {
		case <-cli.ctx.Done():
			blog.Infof("MetricManger manager shut down")
			return
		case event := <-cli.event:
			blog.Infof("received metric event(%s): version(%s) clusterID(%s) namespace(%s) name(%s)",
				event.Type,
				event.Metric.Version,
				event.Metric.ClusterID,
				event.Metric.Namespace,
				event.Metric.Name)

			metricMeta, err := meta.NewMetricMeta(event.Metric, cli.config, cli.storage, cli.route, cli.zk)
			if err != nil {
				blog.Errorf("get meta failed: %v", event.Metric)
				continue
			}

			pipelineText := event.Metric.Namespace + event.Metric.Name
			var f func()

			switch event.Type {
			case watch.EventMetricUpd:
				f = func() {
					var ipMeta map[string]btypes.ObjectMeta
					ipMeta, err = metricMeta.GetIpMeta()
					if err != nil {
						blog.Errorf("get ipMeta failed: %v", event.Metric)
						return
					}
					if err = metricMeta.SetCollectorSettings(ipMeta); err != nil {
						blog.Errorf("set collector settings failed: %v", event.Metric)
						return
					}
					// first metric in this namespace should create the collector application
					if event.First {
						if err = metricMeta.CreateApplication(); err != nil {
							blog.Errorf("create application failed: %v", event.Metric)
							return
						}
					}
				}
			case watch.EventMetricDel:
				f = func() {
					if err = metricMeta.DeleteCollectorSettings(); err != nil {
						blog.Errorf("delete collector settings failed: %v", event.Metric)
						return
					}
					// last metric in this namespace should delete the collector after it quit
					if event.Last {
						if err = metricMeta.DeleteApplication(); err != nil {
							blog.Errorf("delete application failed: %v", event.Metric)
							return
						}
					}
				}
			case watch.EventDynamicUpd:
				f = func() {
					if err = metricMeta.SetCollectorSettings(event.Meta); err != nil {
						blog.Errorf("update collector settings failed: %v", event.Metric)
						return
					}
				}
			default:
				continue
			}

			cli.pipeline.Do(pipelineText, f)
		}
	}
}

type Pipeline struct {
	receivers []chan func()

	ctx    context.Context
	cancel context.CancelFunc
}

const pipelineWidth = 10

func NewPipeLine(ctx context.Context) *Pipeline {
	p := &Pipeline{}
	p.ctx, p.cancel = context.WithCancel(ctx)
	return p
}

func (pl *Pipeline) Start() {
	receivers := make([]chan func(), 0)
	for i := 0; i < pipelineWidth; i++ {
		receiver := make(chan func())
		go func(num int, r chan func()) {
			for {
				select {
				case <-pl.ctx.Done():
					blog.Infof("pipeline %d shutdown", num)
					return
				case f := <-r:
					blog.V(3).Infof("pipeline %d receives task", num)
					f()
				}
			}
		}(i, receiver)
		receivers = append(receivers, receiver)
	}
	pl.receivers = receivers
}

func (pl *Pipeline) Stop() {
	if pl.cancel != nil {
		pl.cancel()
	}
}

func (pl *Pipeline) Do(id string, f func()) {
	index := hashString2Number(id) % pipelineWidth
	pl.receivers[index] <- f
}

func hashString2Number(text string) uint32 {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(text))
	return algorithm.Sum32()
}
