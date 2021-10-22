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
import mock
import pytest

from backend.templatesets.release.generator import generator
from backend.templatesets.release.generator.res_context import ResContext
from backend.tests.bcs_mocks.misc import FakePaaSCCMod

pytestmark = pytest.mark.django_db


@pytest.fixture
def form_release_data(bk_user, cluster_id, form_template, form_version_entity, form_show_version):
    instance_entity = {res_name: ids.split(',') for res_name, ids in form_version_entity.resource_entity.items()}

    namespace = 'test'
    context = ResContext(
        access_token=bk_user.token.access_token,
        username=bk_user.username,
        cluster_id=cluster_id,
        project_id=form_template.project_id,
        namespace=namespace,
        template=form_template,
        show_version=form_show_version,
        instance_entity=instance_entity,
        is_preview=True,
        namespace_id=1,
    )

    with mock.patch(
        'backend.templatesets.release.generator.form_mode.get_ns_variable', return_value=(False, '1.12.3', {})
    ), mock.patch('backend.templatesets.legacy_apps.instance.generator.paas_cc', new=FakePaaSCCMod()):
        data_generator = generator.ReleaseDataGenerator(name="nginx", res_ctx=context)
        release_data = data_generator.generate()
        return release_data


@pytest.fixture
def yaml_release_data(bk_user, cluster_id, yaml_template, yaml_version_entity, yaml_show_version):
    instance_entity = {res_name: ids.split(',') for res_name, ids in yaml_version_entity.resource_entity.items()}
    namespace = 'test'
    context = ResContext(
        access_token=bk_user.token.access_token,
        username=bk_user.username,
        cluster_id=cluster_id,
        project_id=yaml_template.project_id,
        namespace=namespace,
        template=yaml_template,
        show_version=yaml_show_version,
        instance_entity=instance_entity,
        is_preview=True,
        namespace_id=1,
    )

    with mock.patch('backend.helm.app.bcs_info_provider.paas_cc', new=FakePaaSCCMod()), mock.patch(
        'backend.helm.helm.bcs_variable.paas_cc', new=FakePaaSCCMod()
    ):
        data_generator = generator.ReleaseDataGenerator(name="gw", res_ctx=context)
        release_data = data_generator.generate()
        return release_data
