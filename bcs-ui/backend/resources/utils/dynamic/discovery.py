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
from collections import defaultdict
from functools import partial

from kubernetes import __version__
from kubernetes.dynamic.discovery import CacheDecoder, CacheEncoder, LazyDiscoverer
from kubernetes.dynamic.exceptions import NotFoundError

from backend.utils.cache import rd_client

logger = logging.getLogger(__name__)


class DiscovererCache:
    def __init__(self, cache_key):
        self.cache_key = cache_key
        self.rd_client = rd_client

    def exists(self) -> bool:
        return self.rd_client.exists(self.cache_key)

    def get_content(self) -> bytes:
        return self.rd_client.get(self.cache_key)

    def set_content(self, content: str):
        self.rd_client.set(self.cache_key, content)

    def invalidate(self):
        self.rd_client.delete(self.cache_key)


class BcsLazyDiscoverer(LazyDiscoverer):
    """
    - override Discoverer 中的 __init_cache 方法，修复 'CacheDecoder' object is not callable
    - 用redis替代文件缓存
    """

    def get_resources_for_api_version(self, prefix, group, version, preferred):
        """忽略 NotFoundError，直接返回默认值，避免使用 缓存中存在但不存在于集群中的 group 请求 resources 导致报错"""
        resources = defaultdict(list)
        try:
            resources = super().get_resources_for_api_version(prefix, group, version, preferred)
        except NotFoundError:
            logger.warning('Ignore get_resources_for_api_version failed, group: %s, version: %s', group, version)
        return resources

    def _Discoverer__init_cache(self, refresh=False):
        discoverer_cache = self._Discoverer__cache_file

        if refresh or not discoverer_cache.exists():
            self._cache = {'library_version': __version__}
            refresh = True
        else:
            try:
                cache_content = discoverer_cache.get_content()
                self._cache = json.loads(cache_content, cls=partial(CacheDecoder, self.client))
                if self._cache.get('library_version') != __version__:
                    # Version mismatch, need to refresh cache
                    self.invalidate_cache()
            except Exception as e:
                logger.error("load cache error: %s", e)
                self.invalidate_cache()
        self._load_server_info()
        self.discover()
        if refresh:
            self._write_cache()

    def _write_cache(self):
        try:
            discoverer_cache = self._Discoverer__cache_file
            cache_content = json.dumps(self._cache, cls=CacheEncoder)
            discoverer_cache.set_content(cache_content)
        except Exception as e:
            # Failing to write the cache isn't a big enough error to crash on
            logger.exception("write cache error: %s", e)
