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

from django.utils.translation import ugettext_lazy as _
from jsonschema import SchemaError
from jsonschema import ValidationError as JsonValidationError
from jsonschema import validate as json_validate
from rest_framework.exceptions import ValidationError

from .constants import KEY_PATTERN, NUM_VAR_ERROR_MSG, REAL_NUM_VAR_PATTERN
from .models import VersionedEntity, get_model_class_by_resource_name


def get_name_from_config(config):
    return config.get('metadata', {}).get('name') or ''


def is_name_duplicate(resource_name, resource_id, name, version_id):
    """同一类资源的名称不能重复"""
    # 判断新名称与老名称是否一致，如果一致，则不会重复
    model_class = get_model_class_by_resource_name(resource_name)
    try:
        resource = model_class.objects.get(id=resource_id)
        if name == resource.name:
            return False
    except model_class.DoesNotExist:
        pass

    # 只校验当前版本内是否重复
    try:
        version_entity = VersionedEntity.objects.get(id=version_id)
    except VersionedEntity.DoesNotExist:
        return False
    else:
        entity = version_entity.get_entity()
        resource_ids = entity.get(resource_name, '')
        if not resource_ids:
            return False
        if model_class.objects.filter(name=name, id__in=resource_ids.split(',')):
            return True
        return False


def validate_variable_inconfig(config):
    """校验配置文件中的变量名是否合法"""
    search_list = KEY_PATTERN.findall(json.dumps(config))
    search_keys = set(search_list)
    for ikey in search_keys:
        if not REAL_NUM_VAR_PATTERN.match(ikey):
            raise ValidationError(_('变量[{}]不合法, {}').format(ikey, NUM_VAR_ERROR_MSG))


def validate_res_config(config, resource_name, schema):
    err_prefix = '{resource_name} {suffix_msg}'.format(resource_name=resource_name, suffix_msg=_("配置信息格式错误"))
    try:
        json_validate(config, schema)
    except JsonValidationError as e:
        raise ValidationError(f'{err_prefix}:{e.message}')
    except SchemaError as e:
        raise ValidationError(f'{err_prefix}:{e}')


def validate_name_duplicate(data):
    resource_id = data.get('resource_id', None)
    version_id = data.get('version_id', None)
    if resource_id is None or version_id is None:
        return

    resource_name = data['resource_name']
    name = data['name']
    is_duplicate = is_name_duplicate(resource_name, resource_id, name, version_id)
    if is_duplicate:
        raise ValidationError(_('{}名称:{}已经在项目模板中被占用,请重新填写').format(resource_name, name))
