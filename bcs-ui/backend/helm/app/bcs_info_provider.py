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
from dataclasses import dataclass

from backend.components import paas_cc
from backend.templatesets.legacy_apps.instance import constants as instance_constants

try:
    from backend.container_service.observability.datalog.utils import get_data_id_by_project_id
except ImportError:
    from backend.container_service.observability.datalog_ce.utils import get_data_id_by_project_id

logger = logging.getLogger(__name__)

""" bcs_info is used for provide bcs info for helm app, so that it can be inject to resource yaml
"""


def resource_kind_getter(resource, context):
    try:
        return resource["kind"]
    except KeyError:
        return None


def resource_name_getter(resource, context):
    try:
        return resource["metadata"]["name"]
    except KeyError:
        return None


def chart_version_getter(resource, context):
    return context.get("version", "")


def get_custom_labels(resource, context):
    return '{}'


@dataclass
class BcsInfoProvider:
    # note: all data it provide must be pickleable
    project_id: str
    cluster_id: str
    access_token: str  # 注意：该值可能为空（通过api调用helm功能时）
    namespace: str
    namespace_id: int
    context: dict
    ignore_empty_access_token: bool = False

    @property
    def bkdata_container_stdlog_dataid(self):
        # should return str
        data_info = get_data_id_by_project_id(self.project_id)
        return str(data_info.get('standard_data_id'))

    @property
    def cc_app_id(self):
        return str(self.project_info.get("cc_app_id"))

    @property
    def project_info(self):
        """
        {
            "cc_app_id": 764,
            "cc_app_name": "blueking",
        }
        """
        if not self.access_token and self.ignore_empty_access_token:
            return self.context.get("project_info", dict())

        resp = paas_cc.get_project(self.access_token, self.project_id)
        if resp.get('code') != 0:
            logger.error(
                "查询project的信息出错(project_id:{project_id}):{message}".format(
                    project_id=self.project_id, message=resp.get('message')
                )
            )
            return None

        return resp.get('data')

    @property
    def namespace_info(self):
        """
        return:
        {
            cluster_id: "BCS-K8S-15007"
            name: "helm1"
            project_id: "b37778ec757544868a01e1f01f07037f"
        }
        """
        if not self.access_token and self.ignore_empty_access_token:
            return {
                "name": self.namespace,
                "cluster_id": self.cluster_id,
                "project_id": self.project_id,
            }

        resp = paas_cc.get_namespace(self.access_token, self.project_id, self.namespace_id)
        if resp.get('code') != 0:
            logger.error(
                "查询命名空间的信息出错(namespace_id:{project_id}-{namespace_id}):{message}".format(
                    namespace_id=self.namespace_id, project_id=self.project_id, message=resp.get('message')
                )
            )
            return None

        return resp.get('data')

    @property
    def monitor_level(self):
        return "general"

    @property
    def bkdataid(self):
        # should return str
        return "6566"

    def provide_annotations(self, source_type):
        """
        io.tencent.paas.creator  # 创建者rtx名
        io.tencent.paas.updator  # 更新着rtx名
        io.tencent.paas.createTime  # 创建时间
        io.tencent.paas.updateTime  # 更新时间
        """
        data = {
            "io.tencent.paas.creator": self.context["creator"],  # 创建者rtx名
            "io.tencent.paas.updator": self.context["updator"],  # 更新着rtx名
            # 通过getter方法设置，可以保证应用更新时，版本号得到正确的更新
            "io.tencent.paas.version": chart_version_getter,  # 版本号（应用页面显示用）
        }
        return data

    def provide_pod_labels(self, source_type='helm'):
        labels = self.provide_labels(source_type)
        labels.pop("io.tencent.paas.version", None)
        return labels

    def provide_labels(self, source_type='helm'):
        """
        io.tencent.paas.source_type: helm/template  # 来源
        io.tencent.paas.version: test1              # 版本号（应用页面显示用）
        io.tencent.paas.projectid                   # 项目ID

        io.tencent.bcs.clusterid  # 集群ID（兼容lol老数据）
        io.tencent.bcs.namespace  # 命名空间

        io.tencent.bcs.controller.type  # 配置文件类型，Deployment/DaemonSet/Job/StatefulSet
        io.tencent.bcs.controller.name  # 应用名称, metadata.name 中的值
        """
        labels = {
            "io.tencent.paas.source_type": source_type,  # 来源
            "io.tencent.paas.projectid": self.project_id,  # 项目ID
            "io.tencent.bcs.clusterid": self.cluster_id,  # 集群ID（兼容lol老数据）
            "io.tencent.bcs.namespace": self.namespace,  # 命名空间
            "io.tencent.bcs.controller.type": resource_kind_getter,
            # 配置文件类型，Deployment/DaemonSet/Job/StatefulSet, it will be injected with value from danymic getter  # noqa
            "io.tencent.bcs.controller.name": resource_name_getter,
            # 应用名称, metadata.name 中的值, it will be injected with injector with value from danymic getter  # noqa
        }
        return labels

    def provide_container_env(self):
        """
        io_tencent_bcs_namespace    # 命名空间
        io_tencent_bcs_custom_labels    # 附加日志数据，如：'{"set": "set1", "module": "m1"}'
        """
        data = [
            {"name": "io_tencent_bcs_namespace", "value": self.namespace},
            {"name": "io_tencent_bcs_custom_labels", "value": get_custom_labels},
        ]

        return data
