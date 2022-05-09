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
import pytest

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.resource import ResourceClient, ResourceList, ResourceObj
from backend.resources.utils.format import ResourceDefaultFormatter
from backend.tests.conftest import TEST_NAMESPACE


@pytest.fixture
def ctx_cluster(project_id, cluster_id):
    return CtxCluster.create(cluster_id, project_id, token='token')


class PodResourceObj(ResourceObj):
    @property
    def image_name(self) -> str:
        return self.data.spec.containers[0]['image']


class MyPod(ResourceClient):
    kind = 'Pod'
    formatter = ResourceDefaultFormatter()
    result_type = PodResourceObj


@pytest.fixture(
    params=[
        # method kwargs, expected_type
        # default `is_format` value is True
        ({}, dict),
        ({'is_format': True}, dict),
        ({'is_format': False}, PodResourceObj),
    ]
)
def type_assertion_pair(request):
    return request.param


class TestResourceClient:
    @pytest.fixture(autouse=True)
    def init_resources(self, random_name, ctx_cluster):
        MyPod(ctx_cluster).update_or_create(
            namespace=TEST_NAMESPACE, name=random_name, body=make_pod_body(random_name)
        )
        yield
        MyPod(ctx_cluster).delete(namespace=TEST_NAMESPACE, name=random_name)

    def test_list_formatted(self, random_name, ctx_cluster):
        pods = MyPod(ctx_cluster).list(namespace=TEST_NAMESPACE, is_format=True)

        assert isinstance(pods, list)
        assert isinstance(pods[0], dict)

    def test_list_not_formatted(self, random_name, ctx_cluster):
        pods = MyPod(ctx_cluster).list(namespace=TEST_NAMESPACE, is_format=False)

        assert isinstance(pods, ResourceList)
        assert isinstance(pods.metadata, dict)
        assert random_name in [pod.name for pod in pods.items]
        for pod in pods.items:
            if pod.name == random_name:
                assert pod.image_name == 'busybox'

    def test_get_is_format(self, type_assertion_pair, random_name, ctx_cluster):
        kwargs, expected_type = type_assertion_pair
        pod = MyPod(ctx_cluster).get(namespace=TEST_NAMESPACE, name=random_name, **kwargs)
        assert isinstance(pod, expected_type)

    def test_get_none(self, random_name, ctx_cluster):
        pod = MyPod(ctx_cluster).get(namespace=TEST_NAMESPACE, name=random_name + '-non-existent')
        assert pod is None

    def test_patch(self, type_assertion_pair, random_name, ctx_cluster):
        body = make_pod_body(random_name)
        body['metadata']['labels'] = {'foo': 'bar'}

        kwargs, expected_type = type_assertion_pair
        pod = MyPod(ctx_cluster).patch(namespace=TEST_NAMESPACE, name=random_name, body=body, **kwargs)
        assert isinstance(pod, expected_type)


class TestResourceClientCreation:
    def test_create(self, type_assertion_pair, random_name, ctx_cluster):
        kwargs, expected_type = type_assertion_pair
        pod = MyPod(ctx_cluster).create(
            namespace=TEST_NAMESPACE, name=random_name, body=make_pod_body(random_name), **kwargs
        )
        assert isinstance(pod, expected_type)

    def test_update_or_create(self, type_assertion_pair, random_name, ctx_cluster):
        kwargs, expected_type = type_assertion_pair
        pod, created = MyPod(ctx_cluster).update_or_create(
            namespace=TEST_NAMESPACE, name=random_name, body=make_pod_body(random_name), **kwargs
        )
        assert isinstance(created, bool)
        assert isinstance(pod, expected_type)


def make_pod_body(name: str):
    """Make simple Pod body for testing"""
    return {
        'apiVersion': 'v1',
        'kind': 'Pod',
        'metadata': {'name': name},
        'spec': {'containers': [{'name': "main", 'image': "busybox"}]},
    }
