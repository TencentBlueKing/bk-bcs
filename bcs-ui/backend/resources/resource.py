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
import pprint
import time
from typing import Dict, List, Optional, Tuple, Type, TypeVar, Union

from kubernetes.dynamic.resource import ResourceInstance

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.utils.kube_client import get_dynamic_client, get_resource_api

from .constants import PatchType
from .utils.format import ResourceDefaultFormatter, ResourceFormatter

logger = logging.getLogger(__name__)

T = TypeVar('T', bound='ResourceObj')


class ResourceList:
    """包装了 Kubernetes 资源列表的自定义类型

    :param data: DynamicClient 返回的 ResourceInstance 对象
    :param item_type: 包装原始返回的资源类型，应当为 `ResourceObj` 或其子类
    """

    def __init__(self, data: ResourceInstance, item_type: Type[T]):
        self.data = data
        self.item_type = item_type
        self.metadata = dict(data.metadata)
        self._set_items(data)

    def _set_items(self, data: ResourceInstance):
        """Set `self.items` property based on current data"""
        items = getattr(data, 'items')
        try:
            # 检查 "items" 是否可迭代
            iter(items)
            self.items: List[T] = [self.item_type(item) for item in items]
        except TypeError:
            # 当对 `CustomResourceDefinition` 等资源调用 list 接口时, Client 会返回普通的资源而不是 `*List` 类，
            # 针对这种情况，用原结果造一个“假”资源列表
            logger.warning(f"'data.items' is not a valid list type, current type: {type(items)}")
            self.items: List[T] = [self.item_type(data)]

    def __repr__(self) -> str:
        return pprint.pformat({'metadata': self.metadata, 'items': self.items})


class ResourceObj:
    """包装了 Kubernetes 资源对象的自定义类型

    :param data: DynamicClient 返回的 ResourceInstance 对象
    """

    def __init__(self, data: ResourceInstance):
        self.data = data
        self.metadata = dict(data.metadata)

    @property
    def name(self) -> str:
        """资源名称"""
        return self.data.metadata.name

    def __repr__(self) -> str:
        return pprint.pformat({'metadata': self.metadata, 'data': self.data})


class ResourceClient:
    """资源基类

    使用方法：继承该类并重写 `kind`、formatter`、`result_type` 属性（如有必要）。

    默认情况下，本 Client 类的所有方法均会返回 ResourceObj 对象（一种包装了 `ResourceInstance` 的资源对象），但假如
    调用方指定了 `is_format` 参数为 `True`，返回结果将会被 `formatter` 格式化为标准字典类型。

    - 绝大多数方法都支持通过 `formatter` 覆盖默认 `self.formatter` 对象以调整格式化逻辑
    """

    kind = "Resource"

    # 视方法和结果类型不同，Client 将会尝试调用 formatter 的以下方法：
    #
    # 1. `format()`：格式化单个 `ResourceInstance` 对象，`.get()` 等方法使用
    # 2. `format_list()`：格式化列表类型的 `ResourceInstance` 对象，`.list()` 等方法使用
    # 3. `format_dict()`：格式化裸字典资源对象，`.watch()` 方法使用
    formatter: ResourceFormatter = ResourceDefaultFormatter()

    # 将结果转换为对应 ResourceObj 类型
    result_type: Type['ResourceObj'] = ResourceObj

    def __init__(
        self, ctx_cluster: CtxCluster, api_version: Optional[str] = None, cache_client: Optional[bool] = True
    ):
        self.dynamic_client = get_dynamic_client(
            ctx_cluster.context.auth.access_token, ctx_cluster.project_id, ctx_cluster.id, use_cache=cache_client
        )
        self.api = get_resource_api(self.dynamic_client, self.kind, api_version)
        # 保存 ctx_cluster 作为类对象，部分方法会使用
        self.ctx_cluster = ctx_cluster

    def list(
        self, is_format: bool = True, formatter: Optional[ResourceFormatter] = None, **kwargs
    ) -> Union[ResourceList, List, None]:
        resp = self.api.get_or_none(**kwargs)
        if resp is None:
            return resp

        if is_format:
            formatter = formatter or self.formatter
            return formatter.format_list(resp)
        return ResourceList(resp, self.result_type)

    def get(
        self, name: str, is_format: bool = True, formatter: Optional[ResourceFormatter] = None, **kwargs
    ) -> Union[ResourceObj, Dict, None]:
        obj = self.api.get_or_none(name=name, **kwargs)
        if obj is None:
            return obj

        if is_format:
            formatter = formatter or self.formatter
            return formatter.format(obj)
        return self.result_type(obj)

    def create(
        self,
        body: Optional[Dict] = None,
        name: Optional[str] = None,
        namespace: Optional[str] = None,
        is_format: bool = True,
        formatter: Optional[ResourceFormatter] = None,
        **kwargs,
    ) -> Union[ResourceObj, Dict]:
        obj = self.api.create(body=body, name=name, namespace=namespace, update_method="replace", **kwargs)
        if is_format:
            formatter = formatter or self.formatter
            return formatter.format(obj)
        return self.result_type(obj)

    def update_or_create(
        self,
        body: Optional[Dict] = None,
        name: Optional[str] = None,
        namespace: Optional[str] = None,
        is_format: bool = True,
        formatter: Optional[ResourceFormatter] = None,
        **kwargs,
    ) -> Tuple[Union[ResourceObj, Dict], bool]:
        obj, created = self.api.update_or_create(
            body=body, name=name, namespace=namespace, update_method="replace", **kwargs
        )
        if is_format:
            formatter = formatter or self.formatter
            return formatter.format(obj), created
        return self.result_type(obj), created

    def replace(
        self,
        body: Optional[Dict] = None,
        name: Optional[str] = None,
        namespace: Optional[str] = None,
        is_format: bool = True,
        formatter: Optional[ResourceFormatter] = None,
        **kwargs,
    ) -> Union[ResourceObj, Dict]:
        """使用 Replace 模式更新某个资源"""
        obj = self.api.replace(body=body, name=name, namespace=namespace, **kwargs)
        if is_format:
            formatter = formatter or self.formatter
            return formatter.format(obj)
        return self.result_type(obj)

    def patch(
        self,
        body: Optional[Dict] = None,
        name: Optional[str] = None,
        namespace: Optional[str] = None,
        is_format: bool = True,
        formatter: Optional[ResourceFormatter] = None,
        **kwargs,
    ) -> Union[ResourceObj, Dict]:
        # 参考kubernetes/client/rest.py中RESTClientObject类的request方法中对PATCH的处理
        # 如果指定的是json-patch+json但body不是list，则设置为strategic-merge-patch+json
        if kwargs.get("content_type") == PatchType.JSON_PATCH_JSON.value:
            if not isinstance(body, list):
                kwargs["content_type"] = PatchType.STRATEGIC_MERGE_PATCH_JSON.value

        obj, _ = self.api.update_or_create(body=body, name=name, namespace=namespace, update_method="patch", **kwargs)
        if is_format:
            formatter = formatter or self.formatter
            return formatter.format(obj)
        return self.result_type(obj)

    def delete_ignore_nonexistent(
        self,
        name: Optional[str] = None,
        namespace: Optional[str] = None,
        body: Optional[Dict] = None,
        label_selector: Optional[str] = None,
        field_selector: Optional[str] = None,
        **kwargs,
    ) -> Optional[ResourceInstance]:
        """删除某个资源，当目标不存在时忽略错误"""
        return self.api.delete_ignore_nonexistent(name, namespace, body, label_selector, field_selector, **kwargs)

    def delete_wait_finished(
        self,
        name: Optional[str] = None,
        namespace: Optional[str] = None,
        body: Optional[Dict] = None,
        label_selector: Optional[str] = None,
        field_selector: Optional[str] = None,
        max_wait_seconds: int = 10,
        **kwargs,
    ) -> Optional[ResourceInstance]:
        """删除某个资源，一直轮询等到删除成功后返回，默认忽略资源不存在的情况

        :param max_wait_seconds: 最长等待时间，默认为 10 秒
        :raises: 当超过最长等待时间，资源仍然没有删掉时抛出 RuntimeError
        """
        ret = self.api.delete_ignore_nonexistent(name, namespace, body, label_selector, field_selector, **kwargs)

        # 开始轮询检查资源删除情况，默认每 0.1 秒检查一次
        check_interval = 0.1
        start_time = time.time()
        while True:
            if time.time() - start_time >= max_wait_seconds:
                raise RuntimeError('Max wait seconds exceeded, resource(s) still existed after deletion')

            items = self.list(
                name=name,
                namespace=namespace,
                label_selector=label_selector,
                field_selector=field_selector,
                is_format=False,
            )
            if not items:
                return ret

            logger.debug(f'Resource(s) still exists, starting next check in {check_interval}')
            time.sleep(check_interval)

    def delete(
        self,
        name: Optional[str] = None,
        namespace: Optional[str] = None,
        body: Optional[Dict] = None,
        label_selector: Optional[str] = None,
        field_selector: Optional[str] = None,
        **kwargs,
    ) -> ResourceInstance:
        """删除某个资源"""
        return self.api.delete(name, namespace, body, label_selector, field_selector, **kwargs)

    def watch(self, formatter=None, **kwargs) -> List:
        """
        获取较指定的 ResourceVersion 更新的资源状态变更信息

        :param formatter: 指定的格式化器（自定义资源用）
        :return: 指定资源 watch 结果
        """
        formatter = formatter or self.formatter
        return [
            {
                'kind': r['object'].kind,
                'operate': r['type'],
                'uid': r['object'].metadata.uid,
                'manifest': r['raw_object'],
                'manifest_ext': formatter.format_dict(r['raw_object']),
            }
            for r in self.api.watch(**kwargs)
        ]
