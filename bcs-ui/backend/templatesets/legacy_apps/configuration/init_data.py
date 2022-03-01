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
import time

from celery import shared_task
from django.db import transaction
from django.utils import timezone
from django.utils.translation import ugettext_lazy as _

from backend.accounts import bcs_perm
from backend.iam.permissions.resources.templateset import TemplatesetPermCtx, TemplatesetPermission
from backend.utils.basic import RequestClass

from .fixture.template_k8s import K8S_TEMPLATE
from .models import MODULE_DICT, ShowVersion, Template, VersionedEntity, get_default_version

logger = logging.getLogger(__name__)

TEMPLATE_NAME = _("示例模板集")
TEMPLATE_DESC = _("这是一款开源的吃豆游戏的配置模板集，您可以通过这个示例来熟悉微服务架构下系统的配置、发布、变更过程")


def _delete_old_init_templates(template_qsets, project_id, project_code, access_token, username):
    for template in template_qsets:
        template_id = template.id

        perm_ctx = TemplatesetPermCtx(username=username, project_id=project_id, template_id=template_id)
        TemplatesetPermission().can_delete(perm_ctx)

        VersionedEntity.objects.filter(template_id=template_id).delete()
        ShowVersion.objects.filter(template_id=template_id, name='init_version').delete()

        template.name = f'[deleted_{int(time.time())}]{template.name}'
        template.is_deleted = True
        template.deleted_time = timezone.now()
        template.save(update_fields=['name', 'is_deleted', 'deleted_time'])


@shared_task
@transaction.atomic
def init_template(project_id, project_code, access_token, username):
    """创建项目时，初始化示例模板集
    request.project.english_name
    """
    # 判断模板集是否已经创建, 如果已经创建, 删除旧模板
    exit_templates = Template.objects.filter(project_id=project_id, name=TEMPLATE_NAME)
    if exit_templates.exists():
        _delete_old_init_templates(exit_templates, project_id, project_code, access_token, username)

    template_data = K8S_TEMPLATE.get('data', {})

    logger.info(f'init_template [begin] project_id: {project_id}')
    # 新建模板集
    init_template = Template.objects.create(
        project_id=project_id,
        name=TEMPLATE_NAME,
        desc=TEMPLATE_DESC,
        creator=username,
        updator=username,
    )

    new_entity = {}
    for cate in template_data:
        new_item_id_list = []
        data_list = template_data[cate]
        for _data in data_list:
            _save_data = {}
            for _d_key in _data:
                # 新建，忽略 id 字段
                if _d_key == 'id':
                    continue
                # 目前只有 dict、list这两类非字符格式
                _d_value = _data[_d_key]
                if isinstance(_d_value, list) or isinstance(_d_value, dict):
                    _save_data[_d_key] = json.dumps(_d_value)
                else:
                    _save_data[_d_key] = _d_value
            _ins = MODULE_DICT.get(cate).objects.create(**_save_data)
            new_item_id_list.append(str(_ins.id))
        new_entity[cate] = ','.join(new_item_id_list)

    # 新建version
    new_ver = VersionedEntity.objects.create(
        template_id=init_template.id,
        entity=json.dumps(new_entity),
        version=get_default_version(),
        creator=username,
        updator=username,
    )
    # 新建可见版本
    ShowVersion.objects.create(
        template_id=init_template.id,
        real_version_id=new_ver.id,
        name='init_version',
    )

    logger.info(f'init_template [end] project_id: {project_id}')
