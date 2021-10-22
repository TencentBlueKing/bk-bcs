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
from django.template.loader import render_to_string

from ..constants import FileResourceName


def render_tpls_to_string(resource_name):
    return render_to_string(f'{resource_name}.yaml')


INITIAL_TEMPLATES = {
    resource_name: render_tpls_to_string(resource_name) for resource_name in FileResourceName.choice_values()
}


def get_initial_templates():
    return {'resource_names': FileResourceName.choice_values(), 'initial_templates': INITIAL_TEMPLATES}
