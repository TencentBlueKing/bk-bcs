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

import mock
import pytest
from django.conf import settings
from django.utils.crypto import get_random_string

from backend.container_service.clusters.base.models import CtxCluster
from backend.dashboard.examples.constants import RANDOM_SUFFIX_LENGTH, SUFFIX_ALLOWED_CHARS
from backend.resources.custom_object.crd import CustomResourceDefinition
from backend.tests.conftest import TEST_CLUSTER_ID, TEST_PROJECT_ID

# 用于测试解析逻辑的文件路径
TEST_CONFIG_DIR = f'{settings.BASE_DIR}/backend/tests/dashboard/custom_object/contents'

with open(f'{TEST_CONFIG_DIR}/crd_manifest.json') as fr:
    crd_manifest = json.load(fr)

with open(f'{TEST_CONFIG_DIR}/cobj_manifest.json') as fr:
    cobj_manifest = json.load(fr)
    # 为测试用的 CustomObject 配置增加随机后缀
    random_suffix = get_random_string(length=RANDOM_SUFFIX_LENGTH, allowed_chars=SUFFIX_ALLOWED_CHARS)
    cobj_manifest['metadata']['name'] = f"{cobj_manifest['metadata']['name']}-{random_suffix}"


# custom_object 相关 api 单元测试使用同一个 CRD
@pytest.fixture(autouse=True, scope='package')
def update_or_create_crd():
    client = CustomResourceDefinition(
        CtxCluster.create(token='token', project_id=TEST_PROJECT_ID, id=TEST_CLUSTER_ID),
        api_version=crd_manifest["apiVersion"],
    )
    name = crd_manifest['metadata']['name']
    client.create(body=crd_manifest, namespace='default', name=name)
    yield
    client.delete_wait_finished(namespace="default", name=name)


@pytest.fixture(autouse=True, scope='package')
def custom_resource_api_common_patch():
    with mock.patch(
        'backend.dashboard.custom_object_v2.views.gen_cobj_web_annotations', new=lambda *args, **kwargs: {}
    ), mock.patch('backend.dashboard.custom_object_v2.views.gen_base_web_annotations', new=lambda *args, **kwargs: {}):
        yield
