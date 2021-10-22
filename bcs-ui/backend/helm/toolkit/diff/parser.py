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
from typing import List

import yaml

"""
a resource parser which extract resource from raw text,
reference <https://github.com/databus23/helm-diff/blob/master/manifest/parse.go>
"""

logger = logging.getLogger(__name__)
yaml_seperator = b"\n---\n"


class MappingResult(object):
    def __init__(self, name, kind, content):
        self.name = name
        self.kind = kind
        self.content = content

    def __dict__(self):
        return dict(name=self.name, kind=self.kind, content=self.content)


class Resource(object):
    def __init__(self, apiVersion, kind, metadata):
        self.apiVersion = apiVersion
        self.kind = kind
        self.metadata = metadata

    def __str__(self):
        return "%s, %s, %s (%s)" % (self.metadata.namespace, self.metadata.name, self.kind, self.apiVersion)


def split_manifest(manifest: bytes) -> List[bytes]:
    """
    :param manifest: yaml格式的文件，可能包含分隔符---
    :return: 按照---分割的yaml文件列表
    """
    # 过滤掉分割出来可能存在的空文件
    return list(filter(lambda m: m.strip(b"-\t\n "), manifest.split(yaml_seperator)))


def parse(manifest, default_namespace):
    if not isinstance(manifest, bytes):
        if isinstance(manifest, str):
            manifest = manifest.encode("utf-8")
        elif manifest is None:
            manifest = b""
        else:
            logger.warning("unexpect type of %s, with value: %s" % (type(manifest), manifest))
            manifest = bytes(manifest, "utf-8")

    contents = split_manifest(manifest)
    result = dict()
    for content in contents:
        if not content.strip():
            continue

        resource = yaml.load(content)
        if resource is None:
            continue

        if not resource["metadata"].get("namespace"):
            resource["metadata"]["namespace"] = default_namespace

        name = "{kind}/{name}".format(
            name=resource["metadata"]["name"],
            kind=resource["kind"],
        )
        if name in result:
            logger.info("Error: Found duplicate key %s in manifest" % name)
            continue

        result[name] = MappingResult(
            name=name,
            kind=resource["kind"],
            content=content,
        )

    return result
