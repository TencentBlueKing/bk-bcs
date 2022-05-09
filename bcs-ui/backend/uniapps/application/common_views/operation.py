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
import logging

from django.utils.translation import ugettext_lazy as _
from rest_framework import response, viewsets
from rest_framework.exceptions import ValidationError
from rest_framework.renderers import BrowsableAPIRenderer

from backend.bcs_web.audit_log import client
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.utils.basic import getitems
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer

from .. import constants as app_constants
from ..base_views import InstanceAPI
from ..common_views.utils import delete_pods, get_project_namespaces
from ..serializers import ReschedulePodsSLZ

logger = logging.getLogger(__name__)


class RollbackPreviousVersion(InstanceAPI, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_config(self, config):
        try:
            return json.loads(config)
        except Exception as err:
            logger.error("解析实例配置异常，配置: %s，错误详情: %s", config, err)
            return {}

    def from_template(self, instance_id):
        if not self._from_template(instance_id):
            raise error_codes.CheckFailed(_("非模板集实例化的应用不允许进行回滚操作"))

    def get_last_config(self, instance_detail):
        last_config = self.get_config(instance_detail.last_config)
        last_config = last_config.get('old_conf') or {}
        if not last_config:
            raise error_codes.CheckFailed(_("请确认已经执行过更新或滚动升级"))
        return last_config

    def get_current_config(self, instance_detail):
        current_config = self.get_config(instance_detail.config)
        if not current_config:
            raise error_codes.CheckFailed(_("获取实例配置为空"))

        return current_config

    def get(self, request, project_id, instance_id):
        # 检查是否是模板集创建
        self.from_template(instance_id)
        instance_detail = self.get_instance_info(instance_id).first()
        # 校验权限
        self.can_use_instance(request, project_id, instance_detail.namespace)
        # 获取实例的config
        current_config = self.get_current_config(instance_detail)
        last_config = self.get_last_config(instance_detail)

        data = {
            'current_config': current_config,
            'current_config_yaml': self.json2yaml(last_config),
            'last_config': last_config,
            'last_config_yaml': self.json2yaml(last_config),
        }

        return response.Response(data)

    def update_resource(self, request, project_id, cluster_id, namespace, config, instance_detail):
        resp = self.update_deployment(
            request,
            project_id,
            cluster_id,
            namespace,
            config,
            kind=request.project.kind,
            category=instance_detail.category,
            app_name=instance_detail.name,
        )
        is_bcs_success = True if resp.data.get('code') == ErrorCode.NoError else False
        # 更新状态
        instance_detail.oper_type = app_constants.ROLLING_UPDATE_INSTANCE
        instance_detail.is_bcs_success = is_bcs_success
        if not is_bcs_success:
            # 出异常时，保存一次；如果正常，最后保存；减少save次数
            instance_detail.save()
            raise error_codes.APIError(_("回滚上一版本失败，{}").format(resp.data.get('message')))
        # 更新配置
        instance_last_config = json.loads(instance_detail.last_config)
        instance_last_config['old_conf'] = json.loads(instance_detail.config)
        instance_detail.last_config = json.dumps(instance_last_config)
        instance_detail.config = json.dumps(config)
        instance_detail.save()

    def update(self, request, project_id, instance_id):
        """回滚上一版本，只有模板集实例化的才会
        1. 判断当前实例允许回滚
        2. 对应实例的配置
        3. 下发更新操作
        """
        # 检查是否来源于模板集
        self.from_template(instance_id)
        instance_detail = self.get_instance_info(instance_id).first()
        # 校验权限
        self.can_use_instance(request, project_id, instance_detail.namespace)
        # 获取实例的config
        current_config = self.get_current_config(instance_detail)
        last_config = self.get_last_config(instance_detail)
        # 兼容annotation和label
        cluster_id = getitems(current_config, ['metadata', 'annotations', 'io.tencent.bcs.cluster'], '')
        if not cluster_id:
            cluster_id = getitems(current_config, ['metadata', 'labels', 'io.tencent.bcs.cluster'], '')
        namespace = getitems(current_config, ['metadata', 'namespace'], '')
        desc = _("集群:{}, 命名空间:{}, 应用:[{}] 回滚上一版本").format(cluster_id, namespace, instance_detail.name)
        # 下发配置
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=instance_detail.name,
            resource_id=instance_id,
            description=desc,
        ).log_modify():
            self.update_resource(request, project_id, cluster_id, namespace, last_config, instance_detail)

        return response.Response()


class ReschedulePodsViewSet(InstanceAPI, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def _can_use_namespaces(self, request, project_id, data, ns_name_id_map):
        permission = NamespaceScopedPermission()
        for info in data:
            # 通过集群id和namespace名称确认namespace id，判断是否有namespace权限
            ns_id = ns_name_id_map.get((info["cluster_id"], info["namespace"]))
            if not ns_id:
                raise ValidationError(_("集群:{}下没有查询到namespace:{}").format(info["cluster_id"], info["namespace"]))

            perm_ctx = NamespaceScopedPermCtx(
                username=request.user.username,
                project_id=project_id,
                cluster_id=info["cluster_id"],
                name=info["namespace"],
            )
            permission.can_use(perm_ctx)

    def _get_pod_names(self, request, project_id, resource_list):
        """通过应用名称获取相应的"""
        pod_names = {}
        for info in resource_list:
            cluster_id = info["cluster_id"]
            namespace = info["namespace"]
            resource_kind = info["resource_kind"]
            name = info["name"]
            is_bcs_success, data = self.get_pod_or_taskgroup(
                request,
                project_id,
                cluster_id,
                field=["resourceName"],  # 这里仅需要得到podname即可
                app_name=name,
                ns_name=namespace,
                category=resource_kind,
                kind=request.project.kind,
            )
            if not is_bcs_success:
                raise error_codes.APIError(_("查询资源POD出现异常"))
            # data 结构: [{"resourceName": "test1"}, {"resourceName": "test2"}]
            key = (cluster_id, namespace, name, resource_kind)
            for info in data:
                if key in pod_names:
                    pod_names[key].append(info["resourceName"])
                else:
                    pod_names[key] = [info["resourceName"]]
        return pod_names

    def reschedule_pods(self, request, project_id):
        """批量重新调度pod，实现deployment等应用的重建
        NOTE: 这里需要注意，因为前端触发，用户需要在前端展示调用成功后，确定任务都已经下发了
        因此，这里采用同步操作
        """
        # 获取请求参数
        data_slz = ReschedulePodsSLZ(data=request.data)
        data_slz.is_valid(raise_exception=True)
        data = data_slz.validated_data["resource_list"]

        access_token = request.user.token.access_token
        # 判断是否有命名空间权限
        # 因为操作肯定是同一个项目下的，所以获取项目下的namespace信息，然后判断是否有namespace权限
        ns_data = get_project_namespaces(access_token, project_id)
        ns_name_id_map = {(info["cluster_id"], info["name"]): info["id"] for info in ns_data}
        self._can_use_namespaces(request, project_id, data, ns_name_id_map)

        # 查询应用下面的pod
        pod_names = self._get_pod_names(request, project_id, data)
        # 删除pod
        delete_pods(access_token, project_id, pod_names)

        return response.Response()
