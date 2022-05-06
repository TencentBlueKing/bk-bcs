# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
import logging
import time
from itertools import groupby

import arrow
from rest_framework import response, viewsets
from rest_framework.renderers import BrowsableAPIRenderer

from backend.components import bcs_monitor as prometheus
from backend.components import data as apigw_data
from backend.components import paas_cc
from backend.container_service.clusters import serializers as cluster_serializers
from backend.container_service.clusters.utils import use_prometheus_source
from backend.container_service.clusters.views.metric_handler import get_namespace_metric, get_node_metric
from backend.iam.permissions.resources.cluster import ClusterPermCtx, ClusterPermission
from backend.utils.basic import normalize_metric
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.funutils import num_transform
from backend.utils.renderers import BKAPIRenderer

logger = logging.getLogger(__name__)


class ClusterMetricsBase:
    def can_view_cluster(self, request, project_id, cluster_id):
        perm_ctx = ClusterPermCtx(username=request.user.username, project_id=project_id, cluster_id=cluster_id)
        ClusterPermission().can_view(perm_ctx)

    def get_cluster(self, request, project_id, cluster_id):
        cluster_resp = paas_cc.get_cluster(request.user.token.access_token, project_id, cluster_id)
        if cluster_resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError.f(cluster_resp.get('message'))
        return cluster_resp.get('data') or {}


class MetricsParamsBase:
    def get_params(self, request):
        slz = cluster_serializers.MetricsSLZ(data=request.GET)
        slz.is_valid(raise_exception=True)
        return slz.validated_data


class ClusterMetrics(ClusterMetricsBase, MetricsParamsBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_start_end_time(self, start_at, end_at):
        start_at = arrow.get(start_at / 1000).to('local').strftime('%Y-%m-%d %H:%M:%S')
        end_at = arrow.get(end_at / 1000).to('local').strftime('%Y-%m-%d %H:%M:%S')
        return start_at, end_at

    def get_history_data(self, request, project_id, cluster_id, metric, start_at, end_at):
        resp = paas_cc.get_cluster_history_data(
            request.user.token.access_token, project_id, cluster_id, metric, start_at, end_at
        )
        if resp.get('code') != 0:
            raise error_codes.APIError.f(resp.get('message'))
        return resp.get('data') or {}

    def get_prom_disk_data(self, request, cluster_id, metrics_disk):
        """update metric disk by prometheus"""
        try:
            metrics_disk = prometheus.fixed_disk_usage_history(cluster_id)
        except Exception as err:
            logger.error('prometheus error, %s', err)
        return metrics_disk

    def compose_metric(self, request, cluster_data):
        metrics_cpu, metrics_mem, metrics_disk = [], [], []
        for i in cluster_data['results']:
            time = arrow.get(i['capacity_updated_at']).timestamp * 1000
            metrics_cpu.append(
                {'time': time, 'remain_cpu': num_transform(i['remain_cpu']), 'total_cpu': i['total_cpu']}
            )
            total_mem = normalize_metric(i['total_mem'])
            remain_mem = normalize_metric(num_transform(i['remain_mem']))
            metrics_mem.append({'time': time, 'remain_mem': remain_mem, 'total_mem': total_mem})
            # add disk metric
            metrics_disk.append(
                {
                    'time': time,
                    'remain_disk': normalize_metric(num_transform(i['remain_disk']) / 1024),
                    'total_disk': normalize_metric(i['total_disk'] / 1024),
                }
            )
        return metrics_cpu, metrics_mem, metrics_disk

    def list(self, request, project_id):
        """get cluster metric info"""
        params = self.get_params(request)
        cluster_id = params['res_id']
        self.can_view_cluster(request, project_id, cluster_id)
        start_at, end_at = self.get_start_end_time(params['start_at'], params['end_at'])
        # get cluster history data between start_at and end_at
        cluster_data = self.get_history_data(request, project_id, cluster_id, params['metric'], start_at, end_at)
        if not cluster_data.get('results'):
            return response.Response({'cpu': [], 'mem': [], 'disk': []})
        metrics_cpu, metrics_mem, metrics_disk = self.compose_metric(request, cluster_data)

        if use_prometheus_source(request):
            metrics_disk = self.get_prom_disk_data(request, cluster_id, metrics_disk)

        return response.Response({'cpu': metrics_cpu, 'mem': metrics_mem, 'disk': metrics_disk})


class DockerMetrics(MetricsParamsBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def list(self, request, project_id):
        """get docker monitor info"""
        params = self.get_params(request)

        if use_prometheus_source(request):
            data = self.get_prom_data(params)
        else:
            data = self.get_bk_data(params, request.project.cc_app_id)

        return response.Response(data)

    def get_prom_data(self, data):
        _data = []
        if data['metric'] == 'disk':
            metric_data = prometheus.get_container_diskio_usage(data['res_id'], data['start_at'], data['end_at'])
            if metric_data:
                _data = {'list': [{'device_name': '', 'metrics': metric_data}]}

        elif data['metric'] == 'cpu_summary':
            metric_data = prometheus.get_container_cpu_usage(data['res_id'], data['start_at'], data['end_at'])
            if metric_data:
                _data = {'list': metric_data}

        elif data['metric'] == 'net':
            metric_data = prometheus.get_container_network_usage(data['res_id'], data['start_at'], data['end_at'])
            if metric_data:
                _data = {'list': [{'device_name': '', 'metrics': metric_data}]}

        elif data['metric'] == 'mem':
            metric_data = prometheus.get_container_memory_usage(data['res_id'], data['start_at'], data['end_at'])
            if metric_data:
                _data = {'list': metric_data}
        return _data

    def get_bk_data(self, data, cc_app_id):
        _data = apigw_data.get_docker_metrics(
            data['metric'], cc_app_id, data['res_id'], data['start_at'], data['end_at']
        )

        metrics_data = []
        if data['metric'] == 'disk':
            metric_list = groupby(
                sorted(_data['list'], key=lambda x: x['device_name']), key=lambda x: x['device_name']
            )
            for device_name, metrics in metric_list:
                metrics_data.append(
                    {
                        'device_name': device_name,
                        'metrics': [{'used_pct': normalize_metric(i['used_pct']), 'time': i['time']} for i in metrics],
                    }
                )
            _data['list'] = metrics_data
        elif data['metric'] == 'cpu_summary':
            for i in _data['list']:
                i['usage'] = normalize_metric(i.get('cpu_totalusage'))
                i.pop('cpu_totalusage', None)
        elif data['metric'] == 'mem':
            for i in _data['list']:
                i['rss_pct'] = normalize_metric(i['rss_pct'])
        elif data['metric'] == 'net':
            for i in _data['list']:
                i['rxpackets'] = int(i['rxpackets'])
                i['txbytes'] = int(i['txbytes'])
                i['rxbytes'] = int(i['rxbytes'])
                i['txpackets'] = int(i['txpackets'])
        return _data

    def get_multi_params(self, request):
        slz = cluster_serializers.MetricsMultiSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        return slz.validated_data

    def multi(self, request, project_id):
        """get multiple docker info"""
        data = self.get_multi_params(request)

        if use_prometheus_source(request):
            data = self.get_multi_prom_data(data)
        else:
            data = self.get_multi_bk_data(data, request.project.cc_app_id)

        return response.Response(data)

    def get_multi_prom_data(self, data):
        _data = []

        if data['metric'] == 'cpu_summary':
            metric_data = prometheus.get_container_cpu_usage(data['res_id_list'], data['start_at'], data['end_at'])
            if metric_data:
                _data = {'list': metric_data}

        elif data['metric'] == 'mem':
            metric_data = prometheus.get_container_memory_usage(data['res_id_list'], data['start_at'], data['end_at'])
            if metric_data:
                _data = {'list': metric_data}

        return _data

    def get_multi_bk_data(self, data, cc_app_id):
        _data = apigw_data.get_docker_metrics(
            data['metric'], cc_app_id, data['res_id_list'], data['start_at'], data['end_at']
        )

        metrics_data = []
        if data['metric'] == 'cpu_summary':
            metric_list = groupby(sorted(_data['list'], key=lambda x: x['id']), key=lambda x: x['id'])
            for _id, metrics in metric_list:
                container_name = ''
                _metrics = []
                for i in metrics:
                    container_name = i['container_name']
                    _metrics.append(
                        {
                            'usage': normalize_metric(i.get('cpu_totalusage')),
                            'time': i['time'],
                        }
                    )
                metrics_data.append({'id': _id, 'container_name': container_name, 'metrics': _metrics})
            _data['list'] = metrics_data

        elif data['metric'] == 'mem':
            metric_list = groupby(sorted(_data['list'], key=lambda x: x['id']), key=lambda x: x['id'])
            for _id, metrics in metric_list:
                container_name = ''
                _metrics = []
                for i in metrics:
                    container_name = i['container_name']
                    _metrics.append({'rss_pct': normalize_metric(i['rss_pct']), 'time': i['time']})
                metrics_data.append({'id': _id, 'container_name': container_name, 'metrics': _metrics})

            _data['list'] = metrics_data
        return _data


class NodeMetrics(ClusterMetricsBase, MetricsParamsBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_node_list(self, request, project_id):
        """get node list"""
        resp = paas_cc.get_node_list(request.user.token.access_token, project_id, None)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get('message'))
        return resp.get('data') or {}

    def validate_inner_ip(self, request, project_id, inner_ip):
        nodes = self.get_node_list(request, project_id)
        inner_ip_list = [i['inner_ip'] for i in nodes.get('results') or []]
        if inner_ip not in inner_ip_list:
            raise error_codes.CheckFailed.f('node ip is illegal')

    def list(self, request, project_id):
        """get node metrics info"""
        params = self.get_params(request)
        # compatible logic
        try:
            self.validate_inner_ip(request, project_id, params['res_id'])
        except Exception as err:
            return response.Response({'code': ErrorCode.NoError, 'data': {'list': []}, 'message': err})

        if use_prometheus_source(request):
            data = self.get_prom_data(params)
        else:
            data = self.get_bk_data(params, request.project.cc_app_id)
        return response.Response(data)

    def get_prom_data(self, data):
        """prometheus数据"""
        _data = []
        if data['metric'] == 'io':
            metric_data = prometheus.get_node_diskio_usage(data['res_id'], data['start_at'], data['end_at'])
            _data = {'list': [{'device_name': '', 'metrics': metric_data}]}

        elif data['metric'] == 'cpu_summary':
            metric_data = prometheus.get_node_cpu_usage(data['res_id'], data['start_at'], data['end_at'])
            _data = {'list': metric_data}

        elif data['metric'] == 'net':
            metric_data = prometheus.get_node_network_usage(data['res_id'], data['start_at'], data['end_at'])
            _data = {'list': [{'device_name': '', 'metrics': metric_data}]}

        elif data['metric'] == 'mem':
            metric_data = prometheus.get_node_memory_usage(data['res_id'], data['start_at'], data['end_at'])
            _data = {'list': metric_data}
        return _data

    def get_bk_data(self, data, cc_app_id):
        """数据平台返回"""
        _data = apigw_data.get_node_metrics(
            data['metric'], cc_app_id, data['res_id'], data['start_at'], data['end_at']
        )

        metrics_data = []
        if data['metric'] == 'io':
            metric_list = groupby(
                sorted(_data['list'], key=lambda x: x['device_name']), key=lambda x: x['device_name']
            )
            for device_name, metrics in metric_list:
                metrics_data.append(
                    {
                        'device_name': device_name,
                        'metrics': [
                            {
                                'rkb_s': normalize_metric(i['rkb_s']),
                                'wkb_s': normalize_metric(i['wkb_s']),
                                'time': i['time'],
                            }
                            for i in metrics
                        ],
                    }
                )
            _data['list'] = metrics_data
        elif data['metric'] == 'cpu_summary':
            for i in _data['list']:
                i['usage'] = normalize_metric(i['usage'])
        elif data['metric'] == 'net':
            metric_list = groupby(
                sorted(_data['list'], key=lambda x: x['device_name']), key=lambda x: x['device_name']
            )
            for device_name, metrics in metric_list:
                metrics_data.append(
                    {
                        'device_name': device_name,
                        'metrics': [
                            {'speedSent': i['speedSent'], 'speedRecv': i['speedRecv'], 'time': i['time']}
                            for i in metrics
                        ],
                    }
                )
            _data['list'] = metrics_data
        return _data


class NodeSummaryMetrics(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_params(self, request):
        slz = cluster_serializers.SearchResourceBaseSLZ(data=request.GET)
        slz.is_valid(raise_exception=True)
        return slz.validated_data

    def list(self, request, project_id):
        """get cpu/mem/disk info"""
        params = self.get_params(request)
        if use_prometheus_source(request):
            data = self.get_prom_data(params['res_id'])
        else:
            data = self.get_bk_data(params['res_id'], request.project.cc_app_id)

        return response.Response(data)

    def get_prom_data(self, res_id):
        """prometheus数据"""
        end_at = int(time.time()) * 1000
        start_at = end_at - 60 * 10 * 1000

        metric_data = prometheus.get_node_cpu_usage(res_id, start_at, end_at)
        if metric_data:
            cpu_metrics = metric_data[-1]['usage']
        else:
            cpu_metrics = 0

        metric_data = prometheus.get_node_disk_io_utils(res_id, start_at, end_at)
        if metric_data:
            io_metrics = metric_data[-1]['usage']
        else:
            io_metrics = 0

        metric_data = prometheus.get_node_memory_usage(res_id, start_at, end_at)
        if metric_data:
            mem_metrics = normalize_metric(metric_data[-1]['used'] * 100.0 / metric_data[-1]['total'])
        else:
            mem_metrics = 0

        data = {'cpu': cpu_metrics, 'mem': mem_metrics, 'io': io_metrics}
        return data

    def get_bk_data(self, res_id, cc_app_id):
        """数据平台"""
        cpu_metrics = apigw_data.get_node_metrics('cpu_summary', cc_app_id, res_id, limit=1)
        if cpu_metrics['list']:
            cpu_metrics = normalize_metric(cpu_metrics['list'][0]['usage'])
        else:
            cpu_metrics = 0

        mem_metrics = apigw_data.get_node_metrics('mem', cc_app_id, res_id, limit=1)
        if mem_metrics['list']:
            mem_metrics = normalize_metric(mem_metrics['list'][0]['used'] * 100.0 / mem_metrics['list'][0]['total'])
        else:
            mem_metrics = 0

        # device_name 有很多，需要处理
        io_metrics = apigw_data.get_node_metrics('io', cc_app_id, res_id, limit=1)

        if io_metrics['list']:
            io_metrics = normalize_metric(io_metrics['list'][0]['util'])
        else:
            io_metrics = 0

        data = {'cpu': cpu_metrics, 'mem': mem_metrics, 'io': io_metrics}
        return data


class ClusterSummaryMetrics(ClusterMetricsBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_params(self, request):
        slz = cluster_serializers.SummaryMetricsSLZ(data=request.GET)
        slz.is_valid(raise_exception=True)
        return slz.validated_data

    def compose_cluster_resource(self, request, project_id, cluster, ip_resource):
        """compose the cluster response with"""
        data = {'ip_resource': {'total': cluster['ip_resource_total'], 'used': cluster['ip_resource_used']}}
        if not ip_resource:
            node_data = get_node_metric(
                request, request.user.token.access_token, project_id, cluster['cluster_id'], cluster['type']
            )
            namespace_data = get_namespace_metric(request, project_id, cluster['cluster_id'])
            data.update({'node': node_data, 'namespace': namespace_data})

        return data

    def list(self, request, project_id):
        """get cluster summary info"""
        params = self.get_params(request)
        self.can_view_cluster(request, project_id, params['res_id'])
        # get cluster data by paas cc
        cluster = self.get_cluster(request, project_id, params['res_id'])

        return response.Response(
            self.compose_cluster_resource(request, project_id, cluster, params.get('ip_resource'))
        )
