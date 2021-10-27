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

from .models import InstanceConfig, InstanceEvent, MetricConfig, VersionInstance


class VersionInstanceAdmin(admin.ModelAdmin):
    list_display = (
        'id',
        'version_id',
        'show_version_name',
        'template_id',
        'ns_id',
        'is_bcs_success',
        'created',
        'updated',
        'is_deleted',
    )
    search_fields = (
        'id',
        'version_id',
        'show_version_name',
        'template_id',
        'ns_id',
        'is_bcs_success',
        'created',
        'updated',
        'is_deleted',
    )


class InstanceConfigAdmin(admin.ModelAdmin):
    list_display = ('instance_id', 'name', 'namespace', 'category', 'is_bcs_success', 'oper_type', 'status')
    search_fields = ('instance_id', 'name', 'namespace', 'category', 'is_bcs_success', 'oper_type', 'status')


admin.site.register(VersionInstance, VersionInstanceAdmin)
admin.site.register(InstanceConfig, InstanceConfigAdmin)
admin.site.register(MetricConfig)
admin.site.register(InstanceEvent)
