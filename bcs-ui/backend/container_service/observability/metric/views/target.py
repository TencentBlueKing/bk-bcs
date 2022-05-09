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
import itertools
import logging
from urllib import parse

from django.conf import settings
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.components.bcs_monitor.prometheus import get_targets
from backend.container_service.clusters.base.utils import get_cluster_type, get_shared_cluster_proj_namespaces
from backend.container_service.clusters.constants import ClusterType
from backend.container_service.observability.metric.constants import FILTERED_ANNOTATION_PATTERN, JOB_PATTERN
from backend.container_service.observability.metric.serializers import FetchTargetsSLZ
from backend.utils.basic import getitems

logger = logging.getLogger(__name__)


class TargetsViewSet(SystemViewSet):
    """Metric Service 相关接口"""

    def list(self, request, project_id, cluster_id):
        """按 instance_id 聚合的 targets 列表"""
        params = self.params_validate(FetchTargetsSLZ)
        result = get_targets(project_id, cluster_id).get('data') or []
        targets = self._filter_targets(result, params['show_discovered'])

        targets_dict = {}
        for instance_id, targets in itertools.groupby(
            sorted(targets, key=lambda x: x['instance_id']), key=lambda y: y['instance_id']
        ):
            targets = list(targets)
            jobs = {t['job'] for t in targets}
            graph_url = self._gen_graph_url(project_id, cluster_id, jobs) if jobs else None
            targets_dict[instance_id] = {
                'targets': targets,
                'graph_url': graph_url,
                'total_count': len(targets),
                'health_count': len([t for t in targets if t['health'] == 'up']),
            }

        # 如果是共享集群，需要过滤出属于项目的命名空间的 Target
        # 过滤规则：targets_dict key: {namespace}/{name} 取 ns 进行检查
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            project_namespaces = get_shared_cluster_proj_namespaces(request.ctx_cluster, request.project.english_name)
            targets_dict = {
                inst_id: target_info
                for inst_id, target_info in targets_dict.items()
                if inst_id.split('/')[0] in project_namespaces
            }

        return Response(targets_dict)

    def _filter_targets(self, raw_targets, show_discovered):
        """
        按 Job 名称格式过滤符合条件的 Targets

        :param raw_targets: 原始 Target 信息
        :param show_discovered: 是否展示 Discovered
        :return: 符合条件的 Targets
        """
        targets = []
        for raw_t in raw_targets:
            active_targets = getitems(raw_t, 'targets.activeTargets', [])
            for t in active_targets:
                raw_job = t['discoveredLabels']['job']
                # 匹配 Job 名称，若不符合格式则跳过
                match_ret = JOB_PATTERN.match(raw_job)
                if not match_ret:
                    continue

                # 根据是否展示 Discovered 做特殊处理
                if show_discovered:
                    t['discoveredLabels'] = {
                        k: v for k, v in t['discoveredLabels'].items() if not FILTERED_ANNOTATION_PATTERN.match(k)
                    }
                else:
                    t.pop('discoveredLabels')

                job_info = match_ret.groupdict()
                t['instance_id'] = f"{job_info['namespace']}/{job_info['name']}"
                t['discovered_job'] = raw_job
                t['job'] = t['labels']['job']
                targets.append(t)
        return targets

    def _gen_graph_url(self, project_id: str, cluster_id: str, jobs: set):
        """
        生成图表展示的链接

        :param project_id: 项目 ID
        :param cluster_id: 集群 ID
        :param jobs: job 名称列表
        :return: 图表展示链接
        """
        jobs = "|".join(jobs)
        expr = f'{{cluster_id="{cluster_id}",job=~"{jobs}"}}'
        params = {"project_id": project_id, "expr": expr}
        query = parse.urlencode(params)

        return f"{settings.DEVOPS_HOST}/console/monitor/{self.request.project.project_code}/metric?{query}"
