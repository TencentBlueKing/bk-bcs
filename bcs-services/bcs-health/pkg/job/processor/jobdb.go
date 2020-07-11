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

package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/util"

	etcdc "github.com/coreos/etcd/client"
)

type JobDBInterf interface {
	WriteJob(j *util.Job) error
	ListJobs() map[string]map[string]*util.Job
}

func NewJobDb(rootPath string, cli etcdc.KeysAPI) (*JobDB, error) {
	jdb := &JobDB{
		jobPathPrefix: fmt.Sprintf("%s/jobs/", rootPath),
		eCli:          cli,
		jobCache: &jobCache{
			cache: make(map[string]map[string]*util.Job),
		},
	}

	setOpts := &etcdc.SetOptions{
		PrevExist: etcdc.PrevNoExist,
		Dir:       true,
	}
	_, err := jdb.eCli.Set(context.Background(), jdb.jobPathPrefix, "", setOpts)
	if err != nil {
		if eerr, ok := err.(etcdc.Error); ok && eerr.Code != etcdc.ErrorCodeNodeExist {
			return nil, fmt.Errorf("initial etcd job path[%s] failed. err: %v", jdb.jobPathPrefix, err)
		}
	}

	go jdb.watchJobs()
	go jdb.syncEtcdJobs()
	go jdb.jobCache.syncExpireJobs()
	return jdb, nil
}

type JobDB struct {
	// etcd root path
	jobPathPrefix string
	eCli          etcdc.KeysAPI
	jobCache      *jobCache
}

const default_ttl = 10 * 60 // seconds

func (d *JobDB) WriteJob(j *util.Job) error {
	path := d.getJobPath(j)
	opt := &etcdc.SetOptions{
		PrevExist: etcdc.PrevIgnore,
		TTL:       default_ttl * time.Second,
		Dir:       false,
	}

	js, err := json.Marshal(j)
	if err != nil {
		return err
	}

	_, err = d.eCli.Set(context.Background(), path, string(js), opt)
	return err
}

func (d *JobDB) watchJobs() {
	setOpts := &etcdc.SetOptions{
		PrevExist: etcdc.PrevNoExist,
		Dir:       true,
	}
	_, err := d.eCli.Set(context.Background(), d.jobPathPrefix, "", setOpts)
	if nil != err {
		if eerr, ok := err.(etcdc.Error); ok && eerr.Code != etcdc.ErrorCodeNodeExist {
			blog.Fatalf("initial etcd job path failed. err: %v", err)
		}
	}

restart:
	opts := &etcdc.WatcherOptions{
		AfterIndex: 0,
		Recursive:  true,
	}
	watcher := d.eCli.Watcher(d.jobPathPrefix, opts)
	blog.Warnf("-> new etcd watcher and prepare to watch event from etcd.")
	for {
		resp, err := watcher.Next(context.Background())
		if err != nil {
			// why this error should be handled like this, please refer to issue:
			// https://github.com/coreos/etcd/pull/8240
			if e, ok := err.(etcdc.Error); ok && e.Code == etcdc.ErrorCodeEventIndexCleared {
				goto restart
			}

			time.Sleep(1 * time.Second)
			blog.Errorf("watch etcd failed. will try again later. err: %v", err)
			// TODO: before restart the etcd watch again, we should try to resync the losted events
			// during the former watch, otherwise we may lost event forever or lost the alarm message.
			continue
		}
		// all the action is as follows:
		// get, set, delete, update, create, compareAndSwap,
		// compareAndDelete and expire
		var jobs []*util.Job
		switch resp.Action {
		case "set", "create":
			jobs = loopNode(resp.Node)
			for _, j := range jobs {
				d.jobCache.Add(j)
			}
		case "update", "compareAndSwap":
			jobs = loopNode(resp.Node)
			for _, j := range jobs {
				d.jobCache.Update(j)
			}
		case "delete", "expire", "compareAndDelete":
			jobs = loopNode(resp.PrevNode)
			for _, j := range jobs {
				d.jobCache.Delete(j)
			}
		}
		//blog.V(5).Infof("watch etcd jobs status, got *%d* jobs, action: %s", len(jobs), resp.Action)
	}

}

func (d *JobDB) syncEtcdJobs() {
	blog.Infof("start syncing job' status from etcd.")
	d.doSync()
	ticker := time.Tick(300 * time.Second)
	for {
		select {
		case <-ticker:
			blog.V(4).Infof("start sync jobs from etcd.")
		}
		d.doSync()
	}
}
func (d *JobDB) doSync() {
	opts := &etcdc.GetOptions{Recursive: true}
	r, err := d.eCli.Get(context.Background(), d.jobPathPrefix, opts)
	if err != nil {
		blog.Errorf("sync jobs from etcd failed. err: %v", err)
		return
	}

	jobs := loopNode(r.Node)
	blog.V(5).Infof("sync jobs from etcd, got *%d* jobs.", len(jobs))
	tmp := make(map[string]map[string]*util.Job)
	for _, j := range jobs {
		if _, exist := tmp[d.jobCache.Key(j)]; !exist {
			tmp[d.jobCache.Key(j)] = make(map[string]*util.Job)
		}
		tmp[d.jobCache.Key(j)][d.jobCache.SubKey(j)] = j

		if !d.jobCache.Exist(j) {
			blog.Warnf("sync jobs from etcd, find a *new* one, job:[%s]", j.Name())
			d.jobCache.Add(j)
		}
	}

	cachJobs := d.jobCache.List()
	blog.V(5).Infof("sync jobs from cache, jobs: %v", cachJobs)
	for k, v := range cachJobs {
		if _, exist := tmp[k]; !exist {
			for _, j := range v {
				blog.Warnf("sync jobs form etcd, find a *redundant* one, job:[%s]", j.Name())
				d.jobCache.Delete(j)
			}
			continue
		}
		for sk, j := range v {
			if _, exist := tmp[k][sk]; !exist {
				blog.Warnf("sync jobs form etcd, find a *redundant* one, job:[%s]", j.Name())
				d.jobCache.Delete(j)
			}
		}
	}
}

func (d *JobDB) ListJobs() map[string]map[string]*util.Job {
	return d.jobCache.List()
}

type jobCache struct {
	locker sync.Mutex
	cache  map[string]map[string]*util.Job
}

func (c *jobCache) syncExpireJobs() {
	blog.Infof("start sync expire jobs")
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <-ticker:
		}
		now := time.Now().Unix()
		c.locker.Lock()
		for mk, jobs := range c.cache {
			for k, job := range jobs {
				if now-job.Status.FinishedAt > job_expire_seconds {
					delete(c.cache[mk], k)
					blog.V(5).Infof("delete expired job: %s", job.Name())
				}
			}
		}
		c.locker.Unlock()
	}
}

func (c *jobCache) Add(j *util.Job) {
	c.locker.Lock()
	defer c.locker.Unlock()
	if _, exist := c.cache[c.Key(j)]; !exist {
		c.cache[c.Key(j)] = make(map[string]*util.Job)
	}
	c.cache[c.Key(j)][c.SubKey(j)] = j
}

func (c *jobCache) Update(j *util.Job) {
	c.Add(j)
}

func (c *jobCache) Delete(j *util.Job) {
	c.locker.Lock()
	defer c.locker.Unlock()
	if _, exist := c.cache[c.Key(j)]; !exist {
		return
	}
	delete(c.cache[c.Key(j)], c.SubKey(j))
	if len(c.cache[c.Key(j)]) == 0 {
		delete(c.cache, c.Key(j))
	}
}

func (c *jobCache) Exist(j *util.Job) bool {
	c.locker.Lock()
	defer c.locker.Unlock()
	if _, exist := c.cache[c.Key(j)]; !exist {
		return false
	}
	_, exist := c.cache[c.Key(j)][c.SubKey(j)]
	return exist
}

func (c *jobCache) List() map[string]map[string]*util.Job {
	c.locker.Lock()
	defer c.locker.Unlock()
	newCache := make(map[string]map[string]*util.Job)
	for mkey, mv := range c.cache {
		newCache[mkey] = make(map[string]*util.Job)
		for skey, j := range mv {
			newCache[mkey][skey] = j.DeepCopy()
		}
	}
	return newCache
}

func (c *jobCache) Key(j *util.Job) string {
	return fmt.Sprintf("%s::%s::%s", j.Zone, j.Protocol, j.Url)
}

func (c *jobCache) SubKey(j *util.Job) string {
	return j.Status.SlaveInfo.SlaveClusterName
}

func loopNode(node *etcdc.Node) []*util.Job {
	jobs := make([]*util.Job, 0)
	if node.Dir {
		for _, n := range node.Nodes {
			jobs = append(jobs, loopNode(n)...)
		}
		return jobs
	}

	j := new(util.Job)
	blog.V(4).Infof("loop etcd node with key: %s, value: %s", node.Key, node.Value)
	if err := json.Unmarshal([]byte(node.Value), j); err != nil {
		blog.Errorf("[ERROR] unmarshal key[%s] with value: [%s] failed. err :%v", node.Key, node.Value, err)
		return jobs
	}
	jobs = append(jobs, j)

	return jobs
}

// etcd storage path format
// /rootpath/jobs/zone/jobname/slavename
func (d JobDB) getJobPath(j *util.Job) string {
	if j == nil {
		return ""
	}
	u := strings.Replace(j.Url, "/", "::", -1)
	jobname := fmt.Sprintf("%s-%s", strings.ToTitle(string(j.Protocol)), u)
	return fmt.Sprintf("%s/%s/%s/%s",
		d.jobPathPrefix,
		j.Zone,
		jobname,
		j.Status.SlaveInfo.SlaveClusterName)
}
