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

from .models import (
    Application,
    ConfigMap,
    Deplpyment,
    K8sConfigMap,
    K8sDaemonSet,
    K8sDeployment,
    K8sJob,
    K8sSecret,
    K8sService,
    K8sStatefulSet,
    ResourceFile,
    Secret,
    Service,
    ShowVersion,
    Template,
    VersionedEntity,
)


class ShowVersionAdmin(admin.ModelAdmin):
    list_display = ('id', 'template_id', 'real_version_id', 'name')
    search_fields = ('id', 'template_id', 'real_version_id', 'name')


class VersionedEntityAdmin(admin.ModelAdmin):
    list_display = ('id', 'version', 'entity', 'last_version_id')
    search_fields = ('id', 'version', 'entity', 'last_version_id')


class TemplateAdmin(admin.ModelAdmin):
    list_display = ('id', 'name', 'project_id', 'category', 'desc')
    search_fields = ('id', 'name', 'project_id')


class ApplicationAdmin(admin.ModelAdmin):
    list_display = ('id', 'name', 'app_id')
    search_fields = ('id', 'name', 'app_id')


class DeplpymentAdmin(admin.ModelAdmin):
    list_display = ('id', 'name', 'app_id')
    search_fields = ('id', 'name', 'app_id')


class ServiceAdmin(admin.ModelAdmin):
    list_display = ('id', 'name', 'app_id')
    search_fields = ('id', 'name', 'app_id')


class ConfigMapAdmin(admin.ModelAdmin):
    list_display = ('id', 'name')
    search_fields = ('id', 'name')


class SecretAdmin(admin.ModelAdmin):
    list_display = ('id', 'name')
    search_fields = ('id', 'name')


class K8sDeploymentAdmin(admin.ModelAdmin):
    list_display = ('id', 'name', 'deploy_tag')
    search_fields = ('id', 'name', 'deploy_tag')


class K8sConfigMapAdmin(admin.ModelAdmin):
    list_display = ('id', 'name')
    search_fields = ('id', 'name')


class K8sSecretAdmin(admin.ModelAdmin):
    list_display = ('id', 'name')
    search_fields = ('id', 'name')


class K8sDaemonSetAdmin(admin.ModelAdmin):
    list_display = ('id', 'name')
    search_fields = ('id', 'name')


class K8sJobAdmin(admin.ModelAdmin):
    list_display = ('id', 'name')
    search_fields = ('id', 'name')


class K8sServiceAdmin(admin.ModelAdmin):
    list_display = ('id', 'name', 'service_tag')
    search_fields = ('id', 'name', 'service_tag')


class K8sStatefulSetAdmin(admin.ModelAdmin):
    list_display = ('id', 'name', 'service_tag')
    search_fields = ('id', 'name', 'service_tag')


class ResourceFileAdmin(admin.ModelAdmin):
    list_display = ('id', 'name', 'template_id')
    search_fields = ('id', 'name', 'template_id')


admin.site.register(VersionedEntity, VersionedEntityAdmin)
admin.site.register(ShowVersion, ShowVersionAdmin)

admin.site.register(Template, TemplateAdmin)
admin.site.register(Application, ApplicationAdmin)
admin.site.register(Deplpyment, DeplpymentAdmin)
admin.site.register(Service, ServiceAdmin)
admin.site.register(ConfigMap, ConfigMapAdmin)
admin.site.register(Secret, SecretAdmin)

admin.site.register(K8sDeployment, K8sDeploymentAdmin)
admin.site.register(K8sConfigMap, K8sConfigMapAdmin)
admin.site.register(K8sSecret, K8sSecretAdmin)
admin.site.register(K8sDaemonSet, K8sDaemonSetAdmin)
admin.site.register(K8sJob, K8sJobAdmin)
admin.site.register(K8sService, K8sServiceAdmin)
admin.site.register(K8sStatefulSet, K8sStatefulSetAdmin)

admin.site.register(ResourceFile, ResourceFileAdmin)
