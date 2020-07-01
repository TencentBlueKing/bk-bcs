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

package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/util"

	"github.com/emicklei/go-restful"
	"golang.org/x/net/context"
)

// watch jobs return all the jobs belonged by the zones that this slave have.
func (r *HttpAlarm) WatchJobs(req *restful.Request, resp *restful.Response) {
	body, err := ioutil.ReadAll(req.Request.Body)
	if nil != err {
		resp.WriteAsJson(util.SvrResponse{Error: err})
		blog.Errorf("read reqest body failed. err: %v", err)
		return
	}

	blog.Infof("received watch job request, source: [ %s ] data: %s.", req.Request.RemoteAddr, string(body))
	slave := new(util.Slave)
	if err := json.Unmarshal(body, &slave); nil != err {
		resp.WriteAsJson(util.SvrResponse{Error: err})
		blog.Errorf("received an watch job request, but unmarshal failed, err: %v", err)
		return
	}

	jobChan := make(chan *util.Job, 10)
	defer close(jobChan)

	wCtx, cancelWatch := context.WithCancel(context.Background())
	r.s.jobCtrl.WatchSlaveJobs(wCtx, slave, jobChan)

	notifier := resp.ResponseWriter.(http.CloseNotifier)
	flusher := resp.ResponseWriter.(http.Flusher)
	reqCtx := req.Request.Context()
outter:
	for {
		select {
		case job := <-jobChan:
			js, err := json.Marshal(util.SvrResponse{Jobs: []*util.Job{job}})
			if err != nil {
				blog.Errorf("write watch jobs, but marshal jobs failed. err: %v", err)
				continue
			}

			if _, err := resp.Write(js); err != nil {
				blog.Errorf("write watch jobs to slave[ClusterName:%s, IP:%s] failed. err: %v",
					slave.SlaveClusterName, slave.IP, err)
				break outter
			}

			flusher.Flush()

		case <-reqCtx.Done():
			// check the request error
			if reqCtx.Err() != nil {
				blog.Errorf("watch job zone[%v] from slave[ClusterName:%s, IP:%s] finished with an err: %v.",
					slave.Zones, slave.SlaveClusterName, slave.IP, reqCtx.Err())
			}

			// cancel the watch job to jobdb.
			cancelWatch()
			break outter

		case <-notifier.CloseNotify():
			blog.Errorf("watch job zone[%v] from slave[ClusterName:%s, IP:%s] finished, because of client canceled. ",
				slave.Zones, slave.SlaveClusterName, slave.IP)
			// cancel the watch job to jobdb.
			cancelWatch()
			break outter
		}
	}

	blog.Errorf("watch job zone[%v] from slave[ClusterName:%s, IP:%s] finished.",
		slave.Zones, slave.SlaveClusterName, slave.IP)
}

func (r *HttpAlarm) ListJobs(req *restful.Request, resp *restful.Response) {
	data, err := ioutil.ReadAll(req.Request.Body)
	if nil != err {
		resp.WriteAsJson(util.SvrResponse{Error: err})
		blog.Errorf("read reqest body failed. err: %v", err)
		return
	}

	blog.V(5).Infof("received list job request, source: [ %s ] data: %s.", req.Request.RemoteAddr, string(data))
	slave := util.Slave{}
	if err := json.Unmarshal(data, &slave); nil != err {
		resp.WriteAsJson(util.SvrResponse{Error: err})
		blog.Errorf("received an list job request, but unmarshal failed, err: %v", err)
		return
	}

	jobs := r.s.jobCtrl.ListSlaveJobs(slave.Zones)
	if err := resp.WriteAsJson(util.SvrResponse{Jobs: jobs}); err != nil {
		blog.Errorf("write list jobs to slave[ClusterName:%s, IP:%s] failed. err: %v",
			slave.SlaveClusterName, slave.IP, err)
		return
	}

	blog.V(5).Infof("list job zone[%v] to slave[ClusterName:%s, IP:%s] finished.",
		slave.Zones, slave.SlaveClusterName, slave.IP)
}

func (r *HttpAlarm) ReportJobs(req *restful.Request, resp *restful.Response) {
	data, err := ioutil.ReadAll(req.Request.Body)
	if nil != err {
		resp.WriteAsJson(util.SvrResponse{Error: err})
		blog.Errorf("read reqest body failed. err: %v", err)
		return
	}

	//blog.Infof("received report job request, source: [ %s ] data: %s.", req.Request.RemoteAddr, string(data))
	job := new(util.Job)
	if err := json.Unmarshal(data, job); nil != err {
		resp.WriteAsJson(util.SvrResponse{Error: err})
		blog.Errorf("received an report job request, but unmarshal failed, err: %v", err)
		return
	}
	if err := r.s.jobProcessor.WriteJob(job); err != nil {
		resp.WriteAsJson(util.SvrResponse{Error: err})
		blog.Errorf("received an report job request, but write etcd failed, err: %v", err)
		return
	}
	//blog.V(3).Infof("received job report, job: %s", job.String())
}
