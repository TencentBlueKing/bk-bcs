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
import ipaddress
import json
import logging

from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from rest_framework import response, viewsets
from rest_framework.exceptions import ValidationError
from rest_framework.renderers import BrowsableAPIRenderer

from backend.accounts.bcs_perm import Cluster
from backend.bcs_web.audit_log import client
from backend.components import paas_cc
from backend.container_service.clusters import constants as cluster_constants
from backend.container_service.clusters import serializers as cluster_serializers
from backend.container_service.clusters.base import utils as cluster_utils
from backend.container_service.clusters.base.constants import ClusterCOES
from backend.container_service.clusters.constants import ClusterNetworkType, ClusterStatusName
from backend.container_service.clusters.models import ClusterInstallLog, CommonStatus
from backend.container_service.clusters.module_apis import get_cluster_mod
from backend.container_service.clusters.utils import cluster_env_transfer, status_transfer
from backend.resources.utils.kube_client import get_dynamic_client
from backend.utils.basic import normalize_datetime, normalize_metric
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer

# 导入cluster模块
cluster = get_cluster_mod()

DEFAULT_OPER_USER = settings.DEFAULT_OPER_USER
# 1表示gse agent正常
AGENT_NORMAL_STATUS = 1

logger = logging.getLogger(__name__)


class ClusterBase:
    def get_cluster(self, request, project_id, cluster_id):
        """get cluster info"""
        cluster_resp = paas_cc.get_cluster(request.user.token.access_token, project_id, cluster_id)
        if cluster_resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(cluster_resp.get("message"))
        cluster_data = cluster_resp.get("data") or {}
        return cluster_data

    def get_cluster_node(self, request, project_id, cluster_id):
        """get cluster node list"""
        cluster_node_resp = paas_cc.get_node_list(
            request.user.token.access_token,
            project_id,
            cluster_id,
            params={"limit": cluster_constants.DEFAULT_NODE_LIMIT},
        )
        if cluster_node_resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(cluster_node_resp.get("message"))
        data = cluster_node_resp.get("data") or {}
        return data.get("results") or []


class ClusterPermBase:
    def can_view_cluster(self, request, project_id, cluster_id):
        perm = Cluster(request, project_id, cluster_id)
        perm.can_view(raise_exception=True)


class ClusterCreateListViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def cluster_has_node(self, request, project_id):
        """cluster has node
        format: {cluster_id: True/False}
        """
        cluster_node_resp = paas_cc.get_node_list(
            request.user.token.access_token, project_id, None, params={"limit": cluster_constants.DEFAULT_NODE_LIMIT}
        )
        data = cluster_node_resp.get("data") or {}
        results = data.get("results") or []
        # compose the map for cluste and node
        return {
            info["cluster_id"]: True for info in results if info["status"] not in cluster_constants.FILTER_NODE_STATUS
        }

    def get_cluster_list(self, request, project_id):
        cluster_resp = paas_cc.get_all_clusters(request.user.token.access_token, project_id, desire_all_data=1)
        if cluster_resp.get("code") != ErrorCode.NoError:
            logger.error("get cluster error, %s", cluster_resp)
            return {}
        return cluster_resp.get("data") or {}

    def get_cluster_create_perm(self, request, project_id):
        test_cluster_perm = Cluster(request, project_id, cluster_constants.NO_RES, resource_type="cluster_test")
        can_create_test = test_cluster_perm.can_create(raise_exception=False)
        prod_cluster_perm = Cluster(request, project_id, cluster_constants.NO_RES, resource_type="cluster_prod")
        can_create_prod = prod_cluster_perm.can_create(raise_exception=False)
        return can_create_test, can_create_prod

    def list(self, request, project_id):
        """get project cluster list"""
        cluster_info = self.get_cluster_list(request, project_id)
        cluster_data = cluster_info.get("results") or []
        cluster_node_map = self.cluster_has_node(request, project_id)
        # add allow delete perm
        for info in cluster_data:
            info["environment"] = cluster_env_transfer(info["environment"])
            # allow delete cluster
            allow_delete = False if cluster_node_map.get(info["cluster_id"]) else True
            info["allow"] = info["allow_delete"] = allow_delete
        perm_can_use = True if request.GET.get("perm_can_use") == "1" else False

        cluster_results = Cluster.hook_perms(request, project_id, cluster_data, filter_use=perm_can_use)
        # add can create cluster perm for prod/test
        can_create_test, can_create_prod = self.get_cluster_create_perm(request, project_id)

        cluster_results = cluster_utils.append_shared_clusters(cluster_results)
        return response.Response(
            {
                "code": ErrorCode.NoError,
                "data": {"count": len(cluster_results), "results": cluster_results},
                "permissions": {
                    "test": can_create_test,
                    "prod": can_create_prod,
                    "create": can_create_test or can_create_prod,
                },
            }
        )

    def list_clusters(self, request, project_id):
        cluster_info = self.get_cluster_list(request, project_id)
        cluster_data = cluster_info.get("results") or []
        return response.Response({"clusters": cluster_data})

    def create(self, request, project_id):
        """create cluster"""
        cluster_client = cluster.CreateCluster(request, project_id)
        return cluster_client.create()


class ClusterCheckDeleteViewSet(ClusterBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def check_cluster(self, request, project_id, cluster_id):
        """检查集群是否允许删除
        - 检查集群的状态是创建失败的
        - 集群下没有node节点
        """
        cluster_node_list = self.get_cluster_node(request, project_id, cluster_id)
        allow = False if len(cluster_node_list) else True
        return response.Response({"allow": allow})

    def delete(self, request, project_id, cluster_id):
        """删除项目下集群"""
        cluster_client = cluster.DeleteCluster(request, project_id, cluster_id)
        return cluster_client.delete()


class ClusterFilterViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get(self, request, project_id):
        """check cluster name exist"""
        name = request.GET.get("name")
        cluster_resp = paas_cc.get_cluster_by_name(request.user.token.access_token, project_id, name)
        if cluster_resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(cluster_resp.get("message"))
        data = cluster_resp.get("data") or {}
        return response.Response({"is_exist": True if data.get("count") else False})


class ClusterCreateGetUpdateViewSet(ClusterBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def retrieve(self, request, project_id, cluster_id):
        cluster_data = self.get_cluster(request, project_id, cluster_id)

        cluster_data["environment"] = cluster_env_transfer(cluster_data["environment"])

        return response.Response({"code": ErrorCode.NoError, "data": cluster_data})

    def reinstall(self, request, project_id, cluster_id):
        cluster_client = cluster.ReinstallCluster(request, project_id, cluster_id)
        return cluster_client.reinstall()

    def get_params(self, request):
        slz = cluster_serializers.UpdateClusterSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        return dict(slz.validated_data)

    def update_cluster(self, request, project_id, cluster_id, data):
        result = paas_cc.update_cluster(request.user.token.access_token, project_id, cluster_id, data)
        if result.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(result.get("message"))
        return result.get("data") or {}

    def update_data(self, data, project_id, cluster_id, cluster_perm):
        if data["cluster_type"] == "public":
            data["related_projects"] = [project_id]
            cluster_perm.register(cluster_id, "共享集群", "prod")
        elif data.get("name"):
            cluster_perm.update_cluster(cluster_id, data["name"])
        return data

    def update(self, request, project_id, cluster_id):
        cluster_perm = Cluster(request, project_id, cluster_id)
        cluster_perm.can_edit(raise_exception=True)
        data = self.get_params(request)
        data = self.update_data(data, project_id, cluster_id, cluster_perm)
        # update cluster info
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="cluster",
            resource_id=cluster_id,
        ).log_modify():
            cluster_info = self.update_cluster(request, project_id, cluster_id, data)
        # render environment for frontend
        cluster_info["environment"] = cluster_env_transfer(cluster_info["environment"])

        return response.Response(cluster_info)


class ClusterInstallLogView(ClusterBase, viewsets.ModelViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    queryset = ClusterInstallLog.objects.all()

    def get_queryset(self, project_id, cluster_id):
        return super().get_queryset().filter(project_id=project_id, cluster_id=cluster_id).order_by("-create_at")

    def get_display_status(self, curr_status):
        return status_transfer(
            curr_status, cluster_constants.CLUSTER_RUNNING_STATUS, cluster_constants.CLUSTER_FAILED_STATUS
        )

    def get_log_data(self, logs, project_id, cluster_id):
        if not logs:
            return {"status": "none"}
        # 获取最新的一条记录的状态
        latest_log = logs[0]
        status = self.get_display_status(latest_log.status)
        data = {
            "project_id": project_id,
            "cluster_id": cluster_id,
            "status": status,
            "log": [],
            "task_url": logs.first().log_params.get("task_url") or "",
            "error_msg_list": [],
        }
        for info in logs:
            info.status = self.get_display_status(info.status)
            slz = cluster_serializers.ClusterInstallLogSLZ(instance=info)
            data["log"].append(slz.data)
        return data

    def can_view_cluster(self, request, project_id, cluster_id):
        """has view cluster perm"""
        # when cluster exist, check view perm
        try:
            self.get_cluster(request, project_id, cluster_id)
        except Exception as err:
            logger.error("request cluster info, detial is %s", err)
            return
        cluster_perm = Cluster(request, project_id, cluster_id)
        cluster_perm.can_view(raise_exception=True)

    def get(self, request, project_id, cluster_id):
        # view perm
        self.can_view_cluster(request, project_id, cluster_id)
        # get log
        logs = self.get_queryset(project_id, cluster_id)
        data = self.get_log_data(logs, project_id, cluster_id)

        return response.Response(data)


class ClusterInfo(ClusterPermBase, ClusterBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_master_count(self, request, project_id, cluster_id):
        """获取集群master信息"""
        master_info = paas_cc.get_master_node_list(request.user.token.access_token, project_id, cluster_id)
        if master_info.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(master_info.get("message"))
        data = master_info.get("data") or {}
        return data.get("count") or 0

    def get_node_count(self, request, project_id, cluster_id):
        # get node count
        node_results = self.get_cluster_node(request, project_id, cluster_id)
        return len([info for info in node_results if info["status"] not in [CommonStatus.Removed]])

    def get_area(self, request, area_id):
        """get area info"""
        area_info = paas_cc.get_area_info(request.user.token.access_token, area_id)
        if area_info.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(area_info.get("message"))
        return area_info.get("data") or {}

    def get_tke_cluster_config(self, request, project_id, cluster_id):
        """获取tke集群的快照，展示集群所有的配置信息"""
        data = cluster_utils.get_cluster_snapshot(request.user.token.access_token, project_id, cluster_id)
        snapshot = data.get("configure") or "{}"
        snapshot = json.loads(snapshot)
        cidr_settings = snapshot.get("ClusterCIDRSettings") or {}
        cidr = cidr_settings.get("ClusterCIDR")
        advanced_settings = snapshot.get("ClusterAdvancedSettings") or {}
        kube_proxy = cluster_constants.KubeProxy
        config = {
            "max_pod_num": 0,
            "max_service_num": cidr_settings.get("MaxClusterServiceNum") or 0,
            "max_node_pod_num": cidr_settings.get("MaxNodePodNum") or 0,
            "version": snapshot.get("version"),
            "vpc_id": snapshot.get("vpc_id"),
            "network_type": snapshot.get("network_type") or ClusterNetworkType.OVERLAY.value,
            "kube_proxy": kube_proxy.IPVS if advanced_settings.get("IPVS") else kube_proxy.IPTABLES,
            "snapshot": snapshot,  # 添加一个集群快照字段，用于前端快速使用
        }
        if cidr:
            config["cluster_cidr"] = cidr
            config["max_pod_num"] = ipaddress.ip_network(cidr).num_addresses
        return config

    def cluster_info(self, request, project_id, cluster_id):
        # can view cluster
        self.can_view_cluster(request, project_id, cluster_id)
        cluster = self.get_cluster(request, project_id, cluster_id)
        cluster["cluster_name"] = cluster.get("name")
        cluster["created_at"] = normalize_datetime(cluster["created_at"])
        cluster["updated_at"] = normalize_datetime(cluster["updated_at"])
        status = cluster.get("status", "normal")
        cluster["chinese_status_name"] = ClusterStatusName[status].value
        # get area info
        area_info = self.get_area(request, cluster.get("area_id"))
        cluster["area_name"] = _(area_info.get("chinese_name"))
        # get master count
        cluster["master_count"] = self.get_master_count(request, project_id, cluster_id)
        # get node count
        cluster["node_count"] = self.get_node_count(request, project_id, cluster_id)
        total_mem = normalize_metric(cluster["total_mem"])
        cluster["total_mem"] = total_mem

        # 获取集群调度引擎
        coes = cluster["type"]
        # 补充tke和bcs k8s相关配置
        if coes == ClusterCOES.TKE.value:
            cluster.update(self.get_tke_cluster_config(request, project_id, cluster_id))

        cluster_version = self.query_cluster_version(request.user.token.access_token, project_id, cluster_id)
        # 通过集群查询集群版本，如果查询集群异常，则返回集群快照中的数据
        if cluster_version:
            cluster["version"] = cluster_version

        return response.Response(cluster)

    def query_cluster_version(self, access_token: str, project_id: str, cluster_id: str) -> str:
        """查询集群版本
        NOTE: 调用接口出现异常时，返回为空字符串，其它信息可以通过集群快照中获取
        """
        try:
            client = get_dynamic_client(access_token, project_id, cluster_id)
            return client.version["kubernetes"]["gitVersion"]
        except Exception as e:
            logger.error("query cluster version error, %s", e)
            # N/A 表示集群不可用
            return "N/A"


class ClusterVersionViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def versions(self, request, project_id):
        coes = request.query_params.get("coes")
        # 校验集群类型
        if coes not in ClusterCOES.choice_values():
            raise ValidationError(_("集群类型不正确"))
        version_list = cluster_utils.get_cluster_versions(request.user.token.access_token, kind=coes)

        return response.Response(version_list)
