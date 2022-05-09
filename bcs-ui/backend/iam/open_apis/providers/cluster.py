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
from typing import Dict, List

from iam.collection import FancyDict
from iam.resource.provider import ListResult, ResourceProvider
from iam.resource.utils import Page

from backend.components.base import ComponentAuth
from backend.components.cluster_manager import get_shared_clusters
from backend.components.paas_cc import PaaSCCClient
from backend.container_service.clusters.base.utils import get_clusters

from .utils import get_system_token

logger = logging.getLogger(__name__)


class ClusterProvider(ResourceProvider):
    """集群 Provider"""

    def list_instance(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        """
        获取集群列表

        :param filter_obj: 查询参数字典。 以下为必传 如: {"parent": {"id": 1}}
        :param page_obj: 分页对象
        :return: ListResult 类型的实例列表
        """
        project_id = filter_obj.parent['id']
        clusters = self._list_clusters_by_project(project_id)
        return ListResult(results=clusters[page_obj.slice_from : page_obj.slice_to], count=len(clusters))

    def fetch_instance_info(self, filter_obj: FancyDict, **options) -> ListResult:
        """
        批量获取集群属性详情

        :param filter_obj: 查询参数字典
        :return: ListResult 类型的实例列表
        """
        cluster_ids = filter_obj.ids
        paas_cc = PaaSCCClient(auth=ComponentAuth(get_system_token()))
        cluster_list = paas_cc.list_clusters(cluster_ids)
        results = [
            {'id': cluster['cluster_id'], 'display_name': cluster['name'], '_bk_iam_approver_': [cluster['creator']]}
            for cluster in cluster_list
        ]
        return ListResult(results=results, count=len(results))

    def list_instance_by_policy(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def list_attr(self, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def list_attr_value(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        return ListResult(results=[], count=0)

    def search_instance(self, filter_obj: FancyDict, page_obj: Page, **options) -> ListResult:
        """支持模糊搜索集群名"""
        clusters = self._list_clusters_by_project(project_id=filter_obj.parent['id'])
        # 针对搜索关键字过滤集群
        clusters = [cluster for cluster in clusters if filter_obj.keyword in cluster['display_name']]
        return ListResult(results=clusters[page_obj.slice_from : page_obj.slice_to], count=len(clusters))

    def _list_clusters_by_project(self, project_id: str) -> List[Dict[str, str]]:
        """根据项目 ID, 查询项目下的所有集群

        :param project_id: 项目 ID
        :return 集群信息列表. 单个元素结构 {'id': 集群 ID, 'display_name': 集群名}
        """
        cluster_names = {
            cluster['cluster_id']: cluster['name'] for cluster in get_clusters(get_system_token(), project_id)
        }
        shared_cluster_names = {cluster['cluster_id']: cluster['name'] for cluster in get_shared_clusters()}
        # merge 字典, 目的是去重集群(主要是公共集群)
        cluster_names = {**cluster_names, **shared_cluster_names}
        return [{'id': cluster_id, 'display_name': cluster_names[cluster_id]} for cluster_id in cluster_names]
