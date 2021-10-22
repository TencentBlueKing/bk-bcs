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
from django.contrib import admin

from .models import Conf, FunctionController


class FunctionControllerAdmin(admin.ModelAdmin):
    list_display = ('func_code', 'func_name', 'enabled', 'created', 'updated')
    exclude = ('creator', 'updator', 'is_deleted', 'deleted_time')


class ConfAdmin(admin.ModelAdmin):
    list_display = ('key', 'name', 'value')
    search_fields = ('key', 'name', 'value')


admin.site.register(FunctionController, FunctionControllerAdmin)
admin.site.register(Conf, ConfAdmin)
