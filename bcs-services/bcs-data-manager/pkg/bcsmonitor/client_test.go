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

package bcsmonitor

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var (
	header        = http.Header{}
	testOps       = BcsMonitorClientOpt{Schema: "", Endpoint: "", Password: "", UserName: "", AppCode: "", AppSecret: ""}
	testRequester = NewRequester()
)

func TestBcsMonitorClient_LabelValues(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		labelName string
		selectors []string
		startTime time.Time
		endTime   time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *LabelResponse
		wantErr bool
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args:    args{labelName: "job", selectors: nil, startTime: time.Time{}, endTime: time.Time{}},
			want:    &LabelResponse{CommonResponse: CommonResponse{Status: "success"}, Data: []string{"apiserver", "bcs-kube-agent", "kube-controller-manager", "kube-dns", "kube-etcd", "kube-scheduler", "kube-state-metrics", "kubelet", "node-exporter", "prometheus", "prometheus-operator", "thanos-sidecar"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			c.SetCompleteEndpoint()
			got, err := c.LabelValues(tt.args.labelName, tt.args.selectors, tt.args.startTime, tt.args.endTime)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBcsMonitorClient_Labels(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		selectors []string
		startTime time.Time
		endTime   time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *LabelResponse
		wantErr bool
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args: args{selectors: nil, startTime: time.Time{}, endTime: time.Time{}},
			want: &LabelResponse{CommonResponse: CommonResponse{Status: "success"},
				Data: []string{"From", "To", "__name__", "action", "address", "annotation_kubectl_kubernetes_io_last_applied_configuration", "branch", "broadcast", "bucket", "buildDate", "cache", "call", "cause", "check", "client", "cluster_ip", "code", "collector", "compiler", "component", "condition", "config", "configmap", "container", "container_id", "container_name", "container_runtime_version", "contentType", "controller", "cpu", "created_by_kind", "created_by_name", "daemonset", "deployment", "device", "dialer_name", "dockerVersion", "domainname", "duplex", "effect", "endpoint", "error", "event", "exported_endpoint", "exported_namespace", "failure_type", "fstype", "gitCommit", "gitTreeState", "gitVersion", "goVersion", "goversion", "group", "grpc_code", "grpc_method", "grpc_service", "grpc_type", "handler", "host", "host_ip", "id", "image", "image_id", "instance", "interface", "interval", "job", "kernelVersion", "kernel_version", "key", "kind", "kubelet_version", "kubeproxy_version", "label_addonmanager_kubernetes_io_mode", "label_app", "label_app_kubernetes_io_instance", "label_app_kubernetes_io_managed_by", "label_app_kubernetes_io_name", "label_app_kubernetes_io_platform", "label_app_kubernetes_io_version", "label_bcs_webhook", "label_beta_kubernetes_io_arch", "label_beta_kubernetes_io_os", "label_chart", "label_component", "label_controller_revision_hash", "label_helm_sh_chart", "label_heritage", "label_io_tencent_bcs_app_appid", "label_io_tencent_bcs_cluster", "label_io_tencent_bcs_clusterid", "label_io_tencent_bcs_controller_name", "label_io_tencent_bcs_controller_type", "label_io_tencent_bcs_kind", "label_io_tencent_bcs_monitor_level", "label_io_tencent_bcs_namespace", "label_io_tencent_bkdata_baseall_dataid", "label_io_tencent_bkdata_container_stdlog_dataid", "label_io_tencent_paas_projectid", "label_io_tencent_paas_source_type", "label_io_tencent_paas_version", "label_jobLabel", "label_k8s_app", "label_kubernetes_io_arch", "label_kubernetes_io_cluster_service", "label_kubernetes_io_hostname", "label_kubernetes_io_name", "label_kubernetes_io_os", "label_kubespray", "label_managed_by", "label_modifiedAt", "label_module", "label_name", "label_node_role_kubernetes_io_master", "label_node_role_kubernetes_io_node", "label_operated_prometheus", "label_owner", "label_platform", "label_pod_template_generation", "label_pod_template_hash", "label_prometheus", "label_provider", "label_release", "label_role", "label_self_monitor", "label_statefulset_kubernetes_io_pod_name", "label_status", "label_tier", "label_topology_com_tencent_cloud_csi_cbs_zone", "label_version", "le", "listener_name", "machine", "major", "method", "minor", "mode", "mountpoint", "name", "namespace", "node", "nodename", "operation", "operation_name", "operation_type", "operstate", "osVersion", "os_image", "owner_is_controller", "owner_kind", "owner_name", "persistentvolumeclaim", "phase", "platform", "plugin_name", "pod", "pod_ip", "pod_name", "proto", "quantile", "queue_name", "reason", "rejected", "release", "replicaset", "requestKind", "resource", "resource_version", "result", "revision", "role", "scope", "scrape_job", "secret", "server_go_version", "server_id", "server_version", "service", "slice", "state", "statefulset", "status", "status_code", "storageclass", "subresource", "sysname", "system", "triggered_by", "type", "uid", "unit", "url", "username", "verb", "version", "volume_plugin"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			c.SetCompleteEndpoint()
			got, err := c.Labels(tt.args.selectors, tt.args.startTime, tt.args.endTime)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBcsMonitorClient_Query(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		promql string
		time   time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *QueryResponse
		wantErr bool
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args: args{promql: "http_requests_total{instance=\"111\"}", time: time.Time{}},
			want: &QueryResponse{CommonResponse: CommonResponse{Status: "success"},
				Data: QueryData{ResultType: "vector", Result: []VectorResult{}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			c.SetCompleteEndpoint()
			got, err := c.Query(tt.args.promql, tt.args.time)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBcsMonitorClient_QueryByPost(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		promql string
		time   time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *QueryResponse
		wantErr bool
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args: args{promql: "http_requests_total{instance=\"111\"}", time: time.Time{}},
			want: &QueryResponse{CommonResponse: CommonResponse{Status: "success"},
				Data: QueryData{ResultType: "vector", Result: []VectorResult{}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			c.SetCompleteEndpoint()
			got, err := c.QueryByPost(tt.args.promql, tt.args.time)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBcsMonitorClient_QueryRange(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		promql    string
		startTime time.Time
		endTime   time.Time
		step      time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *QueryRangeResponse
		wantErr bool
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args: args{promql: "http_requests_total{instance=\"111\"}", startTime: time.Now(), endTime: time.Now(), step: 15 * time.Second},
			want: &QueryRangeResponse{CommonResponse: CommonResponse{Status: "success"},
				Data: QueryRangeData{ResultType: "matrix", Result: []MatrixResult{}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			c.SetCompleteEndpoint()
			got, err := c.QueryRange(tt.args.promql, tt.args.startTime, tt.args.endTime, tt.args.step)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBcsMonitorClient_QueryRangeByPost(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		promql    string
		startTime time.Time
		endTime   time.Time
		step      time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *QueryRangeResponse
		wantErr bool
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args: args{promql: "http_requests_total{instance=\"111\"}", startTime: time.Now(), endTime: time.Now(), step: 15 * time.Second},
			want: &QueryRangeResponse{CommonResponse: CommonResponse{Status: "success"},
				Data: QueryRangeData{ResultType: "matrix", Result: []MatrixResult{}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			c.SetCompleteEndpoint()
			got, err := c.QueryRangeByPost(tt.args.promql, tt.args.startTime, tt.args.endTime, tt.args.step)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBcsMonitorClient_Series(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		selectors []string
		startTime time.Time
		endTime   time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SeriesResponse
		wantErr bool
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args: args{selectors: []string{"http_requests_total{instance=\"111\"}"}, startTime: time.Now(), endTime: time.Now()},
			want: &SeriesResponse{CommonResponse: CommonResponse{Status: "success"},
				Data: []interface{}{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			c.SetCompleteEndpoint()
			got, err := c.Series(tt.args.selectors, tt.args.startTime, tt.args.endTime)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBcsMonitorClient_SeriesByPost(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		selectors []string
		startTime time.Time
		endTime   time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SeriesResponse
		wantErr bool
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args: args{selectors: []string{"http_requests_total{instance=\"111\"}"}, startTime: time.Now(), endTime: time.Now()},
			want: &SeriesResponse{CommonResponse: CommonResponse{Status: "success"},
				Data: []interface{}{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			c.SetCompleteEndpoint()
			got, err := c.SeriesByPost(tt.args.selectors, tt.args.startTime, tt.args.endTime)
			assert.Equal(t, nil, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBcsMonitorClient_setQuery(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		queryString string
		key         string
		value       string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args: args{queryString: "http_requests_total{instance=\"111\"", key: "test", value: "test"},
			want: "http_requests_total{instance=\"111\"&test=test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			got := c.setQuery(tt.args.queryString, tt.args.key, tt.args.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBcsMonitorClient_setSelectors(t *testing.T) {
	type fields struct {
		opts             BcsMonitorClientOpt
		defaultHeader    http.Header
		completeEndpoint string
		requestClient    Requester
	}
	type args struct {
		queryString string
		selectors   []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{name: "1",
			fields: fields{opts: testOps,
				defaultHeader: header, requestClient: testRequester},
			args: args{queryString: "", selectors: []string{"test"}},
			want: "match[]=test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BcsMonitorClient{
				opts:             tt.fields.opts,
				defaultHeader:    tt.fields.defaultHeader,
				completeEndpoint: tt.fields.completeEndpoint,
				requestClient:    tt.fields.requestClient,
			}
			got := c.setSelectors(tt.args.queryString, tt.args.selectors)
			assert.Equal(t, tt.want, got)
		})
	}
}
