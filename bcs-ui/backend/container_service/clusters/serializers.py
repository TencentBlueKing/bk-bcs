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
import json
import time

import arrow
from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.components import data as data_api
from backend.components import paas_cc
from backend.container_service.clusters import constants
from backend.container_service.clusters.base.utils import get_cluster_nodes
from backend.container_service.clusters.models import ClusterInstallLog, NodeLabel, NodeStatus, NodeUpdateLog
from backend.utils.errcodes import ErrorCode

# metrics 默认时间 1小时
METRICS_DEFAULT_TIMEDELTA = 3600


class NodeLabelSLZ(serializers.ModelSerializer):
    project_id = serializers.CharField(max_length=32, required=True)
    cluster_id = serializers.CharField(max_length=32, required=True)
    node_id = serializers.IntegerField(required=True)
    labels = serializers.JSONField(required=False)

    class Meta:
        model = NodeLabel
        fields = ('id', 'project_id', 'cluster_id', 'node_id', 'labels', 'creator', 'updator')


class NodeLabelUpdateSLZ(serializers.ModelSerializer):
    node_id = serializers.IntegerField(required=True)
    labels = serializers.JSONField(required=False)

    class Meta:
        model = NodeLabel
        fields = ('id', 'project_id', 'cluster_id', 'node_id', 'labels', 'creator', 'updator')


class CreateClusterSLZ(serializers.Serializer):
    cluster_state = serializers.ChoiceField(
        choices=constants.ClusterState.get_choices(), default=constants.ClusterState.BCSNew.value
    )
    name = serializers.CharField(max_length=64)
    description = serializers.CharField(default="")
    area_id = serializers.IntegerField(default=1)
    environment = serializers.CharField(max_length=8)
    master_ips = serializers.ListField(child=serializers.CharField(min_length=1), min_length=1)
    need_nat = serializers.BooleanField(default=True)
    coes = serializers.CharField(default="")

    def validate_master_ips(self, value):
        # 现阶段k8s和mesos的master数量最大限制为5个
        if len(value) % 2 == 0 or len(value) > 5:
            raise ValidationError(_("集群Master节点数量必须为不大于5的奇数"))
        # 可能有多IP的节点，只取第一个即可
        ip_list = [info.split(',')[0] for info in value if info]
        resp = paas_cc.get_project_nodes(self.context['access_token'], self.context['project_id'], is_master=True)
        # 检查是否被占用
        intersection = set(resp.keys()) & set(ip_list)
        if intersection:
            raise ValidationError(_('IP: {ip_list}已经被占用，请重新选择节点').format(ip_list=','.join(intersection)))

        return ip_list

    def validate_name(self, value):
        resp = paas_cc.verify_cluster_exist(self.context['access_token'], self.context['project_id'], value)
        if resp.get('data', {}).get('count'):
            raise ValidationError(_("集群名称已经存在，请修改后重试"))
        return value

    def validate_coes(self, coes):
        if coes:
            return coes
        return self.context["default_coes"]


class UpdateClusterSLZ(serializers.Serializer):
    name = serializers.CharField(required=False, max_length=60)
    status = serializers.IntegerField(required=False)
    description = serializers.CharField(required=False)
    cluster_type = serializers.CharField(required=False, default='private')


class UpdateNodeSLZ(serializers.Serializer):
    name = serializers.CharField(required=False)
    description = serializers.CharField(required=False)
    status = serializers.ChoiceField(
        required=False,
        choices=[
            constants.ClusterManagerNodeStatus.RUNNING.value,
            constants.ClusterManagerNodeStatus.REMOVABLE.value,
        ],
    )


class BatchUpdateNodesSLZ(UpdateNodeSLZ):
    inner_ip_list = serializers.ListField(child=serializers.CharField(), required=True)


class BatchReinstallNodesSLZ(serializers.Serializer):
    node_id_list = serializers.ListField(child=serializers.IntegerField(), required=True)

    def validate_node_id_list(self, node_id_list):
        cluster_nodes = self.context['cluster_nodes']
        # 检查存在
        if set(node_id_list) - cluster_nodes.keys():
            raise ValidationError(_("部分节点不属于当前集群，请确认后重试"))
        # 状态必须为初始化失败
        for node_id in node_id_list:
            if cluster_nodes[node_id]['status'] not in constants.NODE_FAILED_STATUS:
                raise ValidationError(_("重试节点必须处于初始化失败状态，请确认后重试"))
        return node_id_list


class BatchDeleteNodesSLZ(serializers.Serializer):
    node_id_list = serializers.ListField(child=serializers.IntegerField(), required=True)


class CreateNodeSLZ(serializers.Serializer):
    ip = serializers.ListField(child=serializers.CharField(min_length=1), min_length=1)


class NodeLabelParamsSLZ(serializers.Serializer):
    node_id_list = serializers.ListField(child=serializers.IntegerField(min_value=1, required=True), required=True)
    node_label_info = serializers.DictField()

    def validate_node_id_list(self, val):
        if not val:
            raise ValidationError(_("节点ID不能为空"))
        return val


class InstallLogBaseSLZ(serializers.ModelSerializer):
    def to_representation(self, instance):
        data = super().to_representation(instance)
        log = data.get('log') or '{}'
        data['prefix_message'] = '{oper_time}  {operator}  {oper_type_name}'.format(
            oper_time=data['create_at'],
            operator=data.get('operator') or '',
            oper_type_name=self.instance.get_oper_type_display(),
        )
        data['log'] = json.loads(log)
        return data


class ClusterInstallLogSLZ(InstallLogBaseSLZ):
    class Meta:
        model = ClusterInstallLog
        fields = ('project_id', 'cluster_id', 'is_finished', 'status', 'log', 'create_at', 'update_at', 'operator')


class NodeInstallLogSLZ(InstallLogBaseSLZ):
    class Meta:
        model = NodeUpdateLog
        fields = (
            'project_id',
            'cluster_id',
            'node_id',
            'is_finished',
            'status',
            'log',
            'create_at',
            'update_at',
            'operator',
        )


def get_order_choices():
    for m in data_api.NodeMetricFields.keys():
        # `-` represent the desc order
        yield m
        yield '-%s' % m


class ListNodeSLZ(serializers.Serializer):
    limit = serializers.IntegerField(required=False, default=constants.DEFAULT_NODE_LIMIT)
    offset = serializers.IntegerField(required=False, default=0)
    ip = serializers.CharField(required=False)
    ip_list = serializers.ListField(required=False)
    with_containers = serializers.BooleanField(required=False)
    ordering = serializers.ChoiceField(choices=list(get_order_choices()), required=False)
    labels = serializers.ListField(required=False)
    status_list = serializers.ListField(default=[])


class NodeSLZ(serializers.Serializer):
    res_id = serializers.CharField(required=True)

    def validate_res_id(self, res_id):
        request = self.context['request']
        project_id = self.context['project_id']
        cluster_id = self.context['cluster_id']
        ip_list = [i['inner_ip'] for i in get_cluster_nodes(request.user.token.access_token, project_id, cluster_id)]
        if res_id not in ip_list:
            raise ValidationError(f'inner_ip[{res_id}] not found')
        return res_id


class SearchResourceBaseSLZ(serializers.Serializer):
    res_id = serializers.CharField(required=True)


class SummaryMetricsSLZ(SearchResourceBaseSLZ):
    ip_resource = serializers.CharField(required=False)


class MetricsSLZBase(serializers.Serializer):
    metric = serializers.ChoiceField(choices=list(data_api.NodeMetricFields.keys()))
    start_at = serializers.DateTimeField(required=False)
    end_at = serializers.DateTimeField(required=False)

    def validate(self, data):
        now = int(time.time() * 1000)
        # handle the start_at
        if 'start_at' in data:
            data['start_at'] = arrow.get(data['start_at']).timestamp * 1000
        else:
            # default one hour
            data['start_at'] = now - METRICS_DEFAULT_TIMEDELTA * 1000
        # handle the end_at
        if 'end_at' in data:
            data['end_at'] = arrow.get(data['end_at']).timestamp * 1000
        else:
            data['end_at'] = now
        # start_at must be less than end_at
        if data['end_at'] <= data['start_at']:
            raise ValidationError(_('param[start_at] must be less than [end_at]'))
        return data


class MetricsSLZ(MetricsSLZBase):
    res_id = serializers.CharField(required=True)


class MetricsMultiSLZ(MetricsSLZBase):
    res_id_list = serializers.ListField(required=True)


class QueryLabelSLZ(serializers.Serializer):
    cluster_id = serializers.CharField(required=True)


class QueryLabelKeysSLZ(QueryLabelSLZ):
    pass


class QueryLabelValuesSLZ(QueryLabelSLZ):
    key_name = serializers.CharField(required=True)


class FetchCCHostSLZ(serializers.Serializer):
    """ 获取 CMDB 业务下可用主机列表 """

    limit = serializers.IntegerField(label=_('查询行数'), default=constants.DEFAULT_NODE_LIMIT)
    offset = serializers.IntegerField(label=_('偏移量'), default=0)
    ip_list = serializers.ListField(label=_('待过滤 IP 列表'), default=list)
    set_id = serializers.IntegerField(label=_('集群 ID'), default=None)
    module_id = serializers.IntegerField(label=_('模块 ID'), default=None)
    fuzzy = serializers.BooleanField(label=_('是否模糊匹配 IP'), default=False)
    desire_all_data = serializers.BooleanField(label=_('请求全量数据'), default=False)
