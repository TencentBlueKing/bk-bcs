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
from collections import namedtuple
from contextlib import contextmanager
from unittest import mock

import pytest

from backend.iam.permissions.exceptions import PermissionDeniedError
from backend.iam.permissions.resources.project_scoped import (
    ClusterScopedPermission,
    NamespaceScopedPermission,
    ProjectScopedPermCtx,
    can_apply_in_cluster,
)
from backend.tests.iam import fake_iam

from . import roles


@pytest.fixture(autouse=True)
def patch_scoped_permission():
    namespace_scoped_patcher = mock.patch.object(
        NamespaceScopedPermission,
        '__bases__',
        (fake_iam.FakeNamespaceScopedPermission,),
    )
    cluster_scoped_patcher = mock.patch.object(
        ClusterScopedPermission,
        '__bases__',
        (fake_iam.FakeClusterScopedPermission,),
    )
    with namespace_scoped_patcher, cluster_scoped_patcher:
        namespace_scoped_patcher.is_local = True  # 标注为本地属性，__exit__ 的时候恢复成 patcher.temp_original
        cluster_scoped_patcher.is_local = True
        yield


@contextmanager
def does_not_raise():
    yield


Case = namedtuple('Case', ['username', 'expectation', 'no_auth_nums', 'project_id', 'cluster_id', 'namespace_name'])


@pytest.fixture(
    params=[
        [roles.ADMIN_USER, does_not_raise(), 0],
        [roles.ANONYMOUS_USER, pytest.raises(PermissionDeniedError), 8],
        [roles.NAMESPACE_SCOPED_NO_VIEW_USER, pytest.raises(PermissionDeniedError), 5],
        [roles.CLUSTER_SCOPED_NO_CLUSTER_USER, pytest.raises(PermissionDeniedError), 8],
    ]
)
def case(request, project_id, cluster_id, namespace_name):
    return Case(
        username=request.param[0],
        expectation=request.param[1],
        no_auth_nums=request.param[2],
        project_id=project_id,
        cluster_id=cluster_id,
        namespace_name=namespace_name,
    )


def test_can_not_instantiate_in_cluster(case):
    username, project_id, cluster_id, namespace_name = (
        case.username,
        case.project_id,
        case.cluster_id,
        case.namespace_name,
    )

    with case.expectation as exec:
        perm_ctx = ProjectScopedPermCtx(
            username=username, project_id=project_id, cluster_id=cluster_id, namespace=namespace_name
        )
        assert can_apply_in_cluster(perm_ctx)

    if exec:
        assert len(exec.value.data['apply_url']) == case.no_auth_nums
