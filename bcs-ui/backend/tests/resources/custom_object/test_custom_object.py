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
import time
from unittest import mock

import pytest

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.constants import PatchType
from backend.resources.custom_object.crd import CustomResourceDefinition
from backend.resources.custom_object.custom_object import get_cobj_client_by_crd
from backend.utils.basic import getitems

from ..conftest import FakeBcsKubeConfigurationService

# https://github.com/kubernetes/sample-controller
sample_crd = {
    "apiVersion": "apiextensions.k8s.io/v1beta1",
    "kind": "CustomResourceDefinition",
    "metadata": {
        "name": "foos.samplecontroller.k8s.io",
        "annotations": {"api-approved.kubernetes.io": "unapproved, experimental-only"},
    },
    "spec": {
        "group": "samplecontroller.k8s.io",
        "version": "v1alpha1",
        "versions": [{"name": "v1alpha1", "served": True, "storage": True}],
        "names": {"kind": "Foo", "plural": "foos"},
        "scope": "Namespaced",
    },
}

sample_custom_object = {
    "apiVersion": "samplecontroller.k8s.io/v1alpha1",
    "kind": "Foo",
    "metadata": {"name": "example-foo"},
    "spec": {"deploymentName": "example-foo", "replicas": 1},
}


class TestCRDAndCustomObject:
    @pytest.fixture(autouse=True)
    def use_faked_configuration(self):
        """Replace ConfigurationService with fake object"""
        with mock.patch(
            'backend.resources.utils.kube_client.BcsKubeConfigurationService',
            new=FakeBcsKubeConfigurationService,
        ):
            yield

    @pytest.fixture
    def crd_client(self, project_id, cluster_id):
        return CustomResourceDefinition(
            CtxCluster.create(token='token', project_id=project_id, id=cluster_id),
            api_version=sample_crd["apiVersion"],
        )

    @pytest.fixture
    def update_or_create_crd(self, crd_client):
        name = sample_crd['metadata']['name']
        crd_client.update_or_create(body=sample_crd, namespace='default', name=name)
        # NOTE 创建 CRD 后立即使用 可能出现 NotFoundError(404) 异常，需等待就绪
        time.sleep(3)
        yield
        crd_client.delete_wait_finished(namespace="default", name=name)

    @pytest.fixture
    def cobj_client(self, project_id, cluster_id):
        return get_cobj_client_by_crd(
            CtxCluster.create(token='token', project_id=project_id, id=cluster_id),
            crd_name=getitems(sample_crd, "metadata.name"),
        )

    @pytest.fixture
    def update_or_create_custom_object(self, cobj_client):
        cobj_client.update_or_create(
            body=sample_custom_object, namespace="default", name=getitems(sample_custom_object, "metadata.name")
        )
        yield
        cobj_client.delete_wait_finished(name=getitems(sample_custom_object, "metadata.name"), namespace="default")

    def test_crd_list(self, crd_client, update_or_create_crd):
        crd_lists = crd_client.list()
        assert isinstance(crd_lists, list)

    def test_crd_get(self, crd_client, update_or_create_crd):
        crd = crd_client.get(name=getitems(sample_crd, "metadata.name"), is_format=False)
        assert crd.data.spec.scope == "Namespaced"

        crd = crd_client.get(name="no.k3s.cattle.io", is_format=False)
        assert crd is None

    def test_custom_object_patch(self, update_or_create_crd, cobj_client, update_or_create_custom_object):
        cobj = cobj_client.patch(
            name=getitems(sample_custom_object, "metadata.name"),
            namespace="default",
            body={"spec": {"replicas": 2}},
            is_format=False,
            content_type=PatchType.MERGE_PATCH_JSON.value,
        )
        assert cobj.data.spec.replicas == 2
