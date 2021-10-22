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
from kubernetes.dynamic.resource import ResourceInstance

from backend.resources.utils.kube_client import CoreDynamicClient


class TestCoreDynamicClient:
    def test_get_or_none_nonexistent(self, testing_kubernetes_apiclient, random_name):
        resource = CoreDynamicClient(testing_kubernetes_apiclient).resources.get(kind='ConfigMap')
        result = resource.get_or_none(namespace='default', name=random_name)
        assert result is None

    def test_delete_ignore_nonexistent_existed(self, testing_kubernetes_apiclient, random_name):
        resource = CoreDynamicClient(testing_kubernetes_apiclient).resources.get(kind='ConfigMap')
        body = {"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"name": random_name}, "spec": {}}
        resource.create(body=body, name=random_name, namespace='default')

        result = resource.delete_ignore_nonexistent(namespace='default', name=random_name)
        assert isinstance(result, ResourceInstance)

    def test_delete_ignore_nonexistent_nonexistent(self, testing_kubernetes_apiclient, random_name):
        resource = CoreDynamicClient(testing_kubernetes_apiclient).resources.get(kind='ConfigMap')
        result = resource.delete_ignore_nonexistent(namespace='default', name=random_name)
        assert result is None

    def test_update_or_create_create(self, testing_kubernetes_apiclient, random_name):
        resource = CoreDynamicClient(testing_kubernetes_apiclient).resources.get(kind='ConfigMap')
        body = {"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"name": random_name}, "spec": {}}
        obj, created = resource.update_or_create(body=body, name=random_name, namespace='default')

        assert obj.metadata.name == random_name
        assert created is True

    def test_update_or_create_update(self, testing_kubernetes_apiclient, random_name):
        resource = CoreDynamicClient(testing_kubernetes_apiclient).resources.get(kind='ConfigMap')
        body = {"apiVersion": "v1", "kind": "ConfigMap", "metadata": {"name": random_name}, "data": {'foo': 'bar'}}

        # Create resource first
        resource.update_or_create(body=body, name=random_name, namespace='default')

        new_body = {
            "apiVersion": "v1",
            "kind": "ConfigMap",
            "metadata": {"name": random_name},
            "data": {'foo': 'barzzz'},
        }
        obj, created = resource.update_or_create(body=new_body, name=random_name, namespace='default')
        assert obj.data.foo == 'barzzz'
        assert created is False
