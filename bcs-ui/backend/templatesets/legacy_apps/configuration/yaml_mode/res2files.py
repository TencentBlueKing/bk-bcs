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
from rest_framework.exceptions import ValidationError

from .. import models


def get_resource_file(res_name, res_file_ids, *file_fields):
    resource_file = {'resource_name': res_name}
    resource_file['files'] = list(models.ResourceFile.objects.filter(id__in=res_file_ids).values(*file_fields))
    return resource_file


def get_template_files(version_id, *file_fields):
    try:
        ventity = models.VersionedEntity.objects.get(id=version_id)
    except models.VersionedEntity.DoesNotExist:
        raise ValidationError(f'template version(id:{version_id}) does not exist')

    entity = ventity.get_entity()

    template_files = []
    for res_name in sorted(entity.keys()):
        res_file_ids = entity[res_name].split(',')
        template_files.append(get_resource_file(res_name, res_file_ids, *file_fields))

    return template_files
