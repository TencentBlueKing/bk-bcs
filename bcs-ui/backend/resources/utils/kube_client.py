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
from contextlib import contextmanager
from functools import lru_cache
from typing import Any, Dict, Optional, Tuple

from kubernetes.client import ApiClient
from kubernetes.client.exceptions import ApiException
from kubernetes.dynamic import DynamicClient, Resource, ResourceInstance
from kubernetes.dynamic.exceptions import ResourceNotUniqueError

from backend.container_service.clusters.base.models import CtxCluster
from backend.utils.error_codes import error_codes

from ..client import BcsKubeConfigurationService
from .dynamic.discovery import BcsLazyDiscoverer, DiscovererCache

logger = logging.getLogger(__name__)


class CoreDynamicClient(DynamicClient):
    """为官方 SDK 里的 DynamicClient 追加新功能：

    - 使用 sanitize_for_serialization 处理 body
    - 提供获取 preferred resource 方法
    - 包装请求失败时的 ApiException
    - 提供 get_or_none、update_or_create 等方法
    """

    def serialize_body(self, body: Any) -> Dict:
        """使用 sanitize 方法剔除 OpenAPI 对象里的 None 值"""
        body = self.client.sanitize_for_serialization(body)
        return body or {}

    def get_preferred_resource(self, kind: str) -> Resource:
        """尝试获取动态 Resource 对象，优先使用 preferred=True 的 ApiGroup

        :param kind: 资源种类，比如 Deployment
        :raises: ResourceNotUniqueError 匹配到多个不同版本资源，ResourceNotFoundError 没有找到资源
        """
        try:
            return self.resources.get(kind=kind, preferred=True)
        except ResourceNotUniqueError:
            # 如果使用 preferred=True 仍然能匹配到多个 ApiGroup，使用第一个结果
            resources = self.resources.search(kind=kind, preferred=True)
            return resources[0]

    def get_or_none(
        self, resource: Resource, name: Optional[str] = None, namespace: Optional[str] = None, **kwargs
    ) -> Optional[ResourceInstance]:
        """查询资源，当资源不存在抛出 404 错误时返回 None"""
        try:
            return self.get(resource, name=name, namespace=namespace, **kwargs)
        except ApiException as e:
            if e.status == 404:
                return None
            raise

    def delete_ignore_nonexistent(
        self,
        resource: Resource,
        name: Optional[str] = None,
        namespace: Optional[str] = None,
        body: Optional[Dict] = None,
        label_selector: Optional[str] = None,
        field_selector: Optional[str] = None,
        **kwargs,
    ) -> Optional[ResourceInstance]:
        """删除资源，但是当资源不存在时忽略错误"""
        try:
            return resource.delete(
                name=name,
                namespace=namespace,
                body=body,
                label_selector=label_selector,
                field_selector=field_selector,
                **kwargs,
            )
        except ApiException as e:
            if e.status == 404:
                logger.info(
                    f"Delete a non-existent resource {resource.kind}:{name} in namespace:{namespace}, error captured."
                )
                return
            raise

    def update_or_create(
        self,
        resource: Resource,
        body: Optional[Dict] = None,
        name: Optional[str] = None,
        namespace: Optional[str] = None,
        update_method: str = "replace",
        auto_add_version: bool = False,
        **kwargs,
    ) -> Tuple[ResourceInstance, bool]:
        """创建或修改一个 Kubernetes 资源

        :param update_method: 修改类型，默认为 replace，可选值 patch
        :param auto_add_version: 当 update_method=replace 时，是否自动添加 metadata.resourceVersion 字段，默认为 False
        :returns: (instance, created)
        :raises: 当 update_method 不正确时，抛出 ValueError。调用 API 错误时，抛出 ApiException
        """
        if update_method not in ["replace", "patch"]:
            raise ValueError("Invalid update_method {}".format(update_method))

        obj = self.get_or_none(resource, name=name, namespace=namespace, **kwargs)
        if not obj:
            logger.info(f"Resource {resource.kind}:{name} not exists, continue creating")
            return resource.create(body=body, namespace=namespace, **kwargs), True

        # 资源已存在，执行后续的 update 逻辑
        if update_method == 'replace' and auto_add_version:
            self._add_resource_version(obj, body)

        update_func_obj = getattr(resource, update_method)
        return update_func_obj(body=body, name=name, namespace=namespace, **kwargs), False

    def replace(
        self,
        resource: Resource,
        body: Optional[Dict] = None,
        name: Optional[str] = None,
        namespace: str = None,
        auto_add_version: bool = False,
        **kwargs,
    ) -> ResourceInstance:
        if auto_add_version:
            # get 方法若找不到对应的资源，会抛出 ApiException 给上层捕获
            obj = self.get(resource, name=name, namespace=namespace, **kwargs)
            self._add_resource_version(obj, body)
        return super().replace(resource, body, name, namespace, **kwargs)

    def request(self, method, path, body=None, **params):
        # TODO: 包装转换请求异常
        return super().request(method, path, body=body, **params)

    def _add_resource_version(self, obj: ResourceInstance, body: Optional[Dict] = None):
        if isinstance(body, dict):
            body['metadata'].setdefault('resourceVersion', obj.metadata.resourceVersion)


def generate_api_client(access_token: str, project_id: str, cluster_id: str) -> ApiClient:
    """根据指定参数，生成 api_client"""
    ctx_cluster = CtxCluster.create(id=cluster_id, project_id=project_id, token=access_token)
    config = BcsKubeConfigurationService(ctx_cluster).make_configuration()
    return ApiClient(
        config, header_name='X-BKAPI-AUTHORIZATION', header_value=json.dumps({"access_token": access_token})
    )


def generate_core_dynamic_client(access_token: str, project_id: str, cluster_id: str) -> CoreDynamicClient:
    """根据指定参数，生成 CoreDynamicClient"""
    api_client = generate_api_client(access_token, project_id, cluster_id)
    # TODO 考虑集群可能升级k8s版本的情况, 缓存文件会失效
    discoverer_cache = DiscovererCache(cache_key=f"osrcp-{cluster_id}.json")
    return CoreDynamicClient(api_client, cache_file=discoverer_cache, discoverer=BcsLazyDiscoverer)


def get_dynamic_client(
    access_token: str, project_id: str, cluster_id: str, use_cache: bool = True
) -> CoreDynamicClient:
    """
    根据 token、cluster_id 等参数，构建访问 Kubernetes 集群的 Client 对象

    :param access_token: bcs access_token
    :param project_id: 项目 ID
    :param cluster_id: 集群 ID
    :param use_cache: 是否使用缓存
    :return: 指定集群的 CoreDynamicClient
    """
    if use_cache:
        return _get_dynamic_client(access_token, project_id, cluster_id)
    # 若不使用缓存，则直接生成新的实例返回
    return generate_core_dynamic_client(access_token, project_id, cluster_id)


@lru_cache(maxsize=128)
def _get_dynamic_client(access_token: str, project_id: str, cluster_id: str) -> CoreDynamicClient:
    """获取 Kubernetes Client 对象（带缓存）"""
    return generate_core_dynamic_client(access_token, project_id, cluster_id)


@lru_cache(maxsize=128)
def get_resource_api(dynamic_client: CoreDynamicClient, kind: str, api_version: Optional[str] = None) -> Resource:
    """获取绑定到具体资源类型的 Resource API client"""
    if api_version:
        return dynamic_client.resources.get(kind=kind, api_version=api_version)
    return dynamic_client.get_preferred_resource(kind)


def make_labels_string(labels: Dict) -> str:
    """Turn a labels dict into string format

    :param labels: dict of labels
    """
    return ",".join("{}={}".format(key, value) for key, value in labels.items())


@contextmanager
def wrap_kube_client_exc():
    try:
        yield
    except ApiException as e:
        body = json.loads(e.body)
        raise error_codes.ResourceError(body['message'])
