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
from kubernetes.dynamic import resource


class Resource(resource.Resource):
    """修复官方bug，见pr: https://github.com/kubernetes-client/python-base/pull/251"""

    def to_dict(self):
        d = {
            '_type': 'Resource',
            'prefix': self.prefix,
            'group': self.group,
            'api_version': self.api_version,
            'kind': self.kind,
            'namespaced': self.namespaced,
            'verbs': self.verbs,
            'name': self.name,
            'preferred': self.preferred,
            'singularName': self.singular_name,
            'shortNames': self.short_names,
            'categories': self.categories,
            'subresources': {k: sr.to_dict() for k, sr in self.subresources.items()},
        }
        d.update(self.extra_args)
        return d


class Subresource(resource.Subresource):
    """修复官方bug，见pr: https://github.com/kubernetes-client/python-base/pull/251"""

    def __init__(self, parent, **kwargs):
        self.parent = parent
        self.prefix = parent.prefix
        self.group = parent.group
        self.api_version = parent.api_version
        self.kind = kwargs.pop('kind')
        self.name = kwargs.pop('name')
        self.subresource = kwargs.pop('subresource', None) or self.name.split('/')[1]
        self.namespaced = kwargs.pop('namespaced', False)
        self.verbs = kwargs.pop('verbs', None)
        self.extra_args = kwargs

    def to_dict(self):
        d = {
            'kind': self.kind,
            'name': self.name,
            'subresource': self.subresource,
            'namespaced': self.namespaced,
            'verbs': self.verbs,
        }
        d.update(self.extra_args)
        return d
