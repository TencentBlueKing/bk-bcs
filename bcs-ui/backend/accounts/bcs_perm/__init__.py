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

import abc
import logging

from django.utils.translation import ugettext_lazy as _

from backend.components.enterprise.bk_login import get_all_users
from backend.components.ssm import get_client_access_token

# 与资源无关
NO_RES = "**"
# 任意资源
ANY_RES = "*"

logger = logging.getLogger(__name__)


class PermissionMeta(object):
    """权限元类"""

    __metaclass__ = abc.ABCMeta

    # 服务类型，常量
    RESOURCE_TYPE = ""
    RES_TYPE_NAME = ""

    # 功能列表
    POLICY_LIST = ["create", "delete", "view", "edit", "use"]

    CMD_NAME = {
        "delete": _("删除"),
        "create": _("创建"),
        "use": _("使用"),
        "edit": _("编辑"),
        "view": _("查看"),
        "list": _("列表"),
    }

    def __init__(self, request, project_id, resource_id, resource_name=None):
        pass

    def had_perm(self, action_id):
        """判断是否有权限"""
        return True

    def can_create(self, raise_exception):
        """创建权限不做判断"""
        # 创建权限都默认开放
        return True

    def can_edit(self, raise_exception):
        """是否编辑权限"""
        return True

    def can_delete(self, raise_exception):
        """是否使用删除"""
        return True

    def can_view(self, raise_exception):
        return True

    def can_use(self, raise_exception):
        """是否使用权限"""
        return True

    def get_msg_key(self, cmd):
        return f"{cmd}_msg"

    def get_msg(self, cmd):
        """获取消息"""
        return ""

    def register(self, resource_id, resource_name):
        """注册资源到权限中心"""
        return {"code": 0}

    def delete(self):
        """删除资源"""
        return {"code": 0}

    def update_name(self, resource_name, raise_exception=False):
        return {"code": 0}

    def hook_perms(self, data_list, filter_use=False, id_flag="id"):
        """资源列表，添加permissions"""
        # NOTE: 现阶段有项目权限，那么就有所有权限
        default_perms = {perm: True for perm in self.POLICY_LIST}
        data_list = data_list or []
        for data in data_list:
            data["permissions"] = default_perms
        return data_list

    def get_err_data(self, policy_code):
        return []


class Cluster(PermissionMeta):
    """集群权限"""

    # 资源类型
    RESOURCE_TYPE = "cluster"
    RES_TYPE_NAME = "集群"

    POLICY_LIST = ["create", "edit", "cluster-readonly", "cluster-manager"]

    def __init__(self, request, project_id, resource_id, resource_type=None):
        pass

    @classmethod
    def hook_perms(cls, request, project_id, cluster_list, filter_use=False):
        default_perms = {perm: True for perm in cls.POLICY_LIST}
        default_perms.update({"view": True, "use": True, "delete": True})
        cluster_list = cluster_list or []
        for data in cluster_list:
            data["permissions"] = default_perms
        return cluster_list

    def register(self, cluster_id, cluster_name, environment=None):
        """注册集群"""
        return super(Cluster, self).register(cluster_id, cluster_name)

    def delete_cluster(self, cluster_id, environment=None):
        """删除集群"""
        return {"code": 0}

    def update_cluster(self, cluster_id, cluster_name):
        """更新注册集群的名称"""
        return {"code": 0}


class Namespace(PermissionMeta):
    """命名空间权限
    命名空间的 resource_id格式为：cluster:cluster1/namespace:namespace2
    """

    # 资源类型
    RESOURCE_TYPE = "namespace"
    RES_TYPE_NAME = "命名空间"

    # 功能列表
    POLICY_LIST = ["edit", "cluster-readonly", "cluster-manager"]

    def __init__(self, request, project_id, namespace_id, cluster_id=None, namespace_name=None):
        pass

    def register(self, resource_id, resource_name):
        """注册资源到权限中心"""
        return {"code": 0}

    def delete(self):
        return {"code": 0}

    def can_create(self, raise_exception=True):
        """是否编辑命名空间权限"""
        return True

    def can_use(self, raise_exception=True):
        """命名空间的使用权限，需要先判断集群的使用权限"""
        return True

    def can_edit(self, raise_exception=True):
        """命名空间的编辑权限，需要先判断集群的使用权限"""
        return True

    def hook_base_perms(
        self, ns_list, ns_id_flag="id", cluster_id_flag="cluster_id", ns_name_flag="name", **filter_parms
    ):  # noqa
        default_perms = {perm: True for perm in self.POLICY_LIST}
        default_perms.update({"create": True, "view": True, "use": True, "delete": True, "edit_msg": ""})
        ns_list = ns_list or []
        for data in ns_list:
            data["permissions"] = default_perms
        return ns_list

    def hook_perms(
        self, ns_list, filter_use=False, ns_id_flag="id", cluster_id_flag="cluster_id", ns_name_flag="name"
    ):  # noqa
        """批量添加权限"""
        filter_parms = {}
        if filter_use:
            filter_parms["is_filter"] = True
            filter_parms["filter_type"] = "cluster-manager"
        return self.hook_base_perms(
            ns_list, ns_id_flag=ns_id_flag, cluster_id_flag=cluster_id_flag, ns_name_flag=ns_name_flag, **filter_parms
        )


class Templates(PermissionMeta):
    """模板集权限"""

    # 资源类型
    RESOURCE_TYPE = "templates"
    RES_TYPE_NAME = _("模板集")

    # 功能列表
    POLICY_LIST = ["create", "delete", "view", "edit", "use"]


class Metric(PermissionMeta):
    """metric权限"""

    # 资源类型
    RESOURCE_TYPE = "metric"
    RES_TYPE_NAME = "Metric"

    # 功能列表
    POLICY_LIST = ["edit", "create", "delete", "use"]


PERMS_DICT = {
    "cluster_prod": Cluster,
    "cluster_test": Cluster,
    "namespace": Namespace,
    "metric": Metric,
    "templates": Templates,
}


def get_perm_cls(resource_type, request, project_id, resource_id, resource_name):
    """获取权限实例
    给前端API校验使用
    """
    perm_cls = PERMS_DICT[resource_type]
    if perm_cls in (Namespace, Cluster):
        return perm_cls(request, project_id, resource_id)
    else:
        return perm_cls(request, project_id, resource_id, resource_name)


def get_access_token():
    return get_client_access_token()


def verify_project_by_user(access_token, project_id, project_code, user_id):
    """
    验证用户是否有项目权限
    """
    return True


def get_all_user():
    resp = get_all_users()
    data = resp.get("data") or []
    users = []
    for _d in data:
        users.append({"id": _d.get("bk_username", ""), "name": _d.get("chname", "")})
    return users


try:
    from .perm_ext import *  # noqa
except ImportError as e:
    logger.debug('Load extension failed: %s', e)
