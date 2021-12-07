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

from backend.templatesets.legacy_apps.configuration import models
from backend.templatesets.legacy_apps.configuration.constants import TemplateEditMode

from . import res_manifest

pytestmark = pytest.mark.django_db


@pytest.fixture
def form_version_entity(form_template):
    deploy1 = models.K8sDeployment.perform_create(
        name='nginx-deployment1',
        config=res_manifest.NGINX_DEPLOYMENT1_JSON,
    )
    deploy2 = models.K8sDeployment.perform_create(
        name='nginx-deployment2',
        config=res_manifest.NGINX_DEPLOYMENT2_JSON,
    )

    svc = models.K8sService.perform_create(name='nginx-service', config=res_manifest.NGINX_SVC_JSON)
    ventity = models.VersionedEntity.objects.create(
        template_id=form_template.id, entity={'K8sDeployment': f'{deploy1.id},{deploy2.id}', 'K8sService': f'{svc.id}'}
    )
    return ventity


@pytest.fixture
def form_show_version(project_id, form_template, form_version_entity):
    return models.ShowVersion.objects.create(
        name='v1', template_id=form_template.id, real_version_id=form_version_entity.id
    )


@pytest.fixture
def yaml_template(project_id):
    template = models.Template.objects.create(project_id=project_id, name='gw', edit_mode=TemplateEditMode.YAML.value)
    return template


@pytest.fixture
def yaml_version_entity(yaml_template):
    res_file1 = models.ResourceFile.objects.create(
        name='nginx',
        resource_name='Deployment',
        content=res_manifest.NGINX_DEPLOYMENT_YAML,
        template_id=yaml_template.id,
    )
    res_file2 = models.ResourceFile.objects.create(
        name="redis",
        resource_name='CustomManifest',
        content=res_manifest.CUSTOM_MANIFEST2_YAML,
        template_id=yaml_template.id,
    )

    res_file3 = models.ResourceFile.objects.create(
        name="gw",
        resource_name='CustomManifest',
        content=res_manifest.CUSTOM_MANIFEST1_YAML,
        template_id=yaml_template.id,
    )

    ventity = models.VersionedEntity.objects.create(
        template_id=yaml_template.id,
        entity={'Deployment': f'{res_file1.id}', 'CustomManifest': f'{res_file2.id},{res_file3.id}'},
    )
    return ventity


@pytest.fixture
def yaml_show_version(project_id, yaml_template, yaml_version_entity):
    return models.ShowVersion.objects.create(
        name='v1', template_id=yaml_template.id, real_version_id=yaml_version_entity.id
    )
