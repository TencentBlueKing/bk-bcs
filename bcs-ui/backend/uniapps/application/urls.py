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

from django.conf.urls import include, url

from . import instance_views, views
from .all_views import views as ns_views
from .common_views import operation
from .common_views import query as common_query
from .filters import views as filter_views
from .other_views import templates

urlpatterns = [
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/retry/$",  # noqa
        views.CreateInstance.as_view(),
        name="api.project.retry_instance",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/info/$",  # noqa
        instance_views.GetInstanceInfo.as_view(),
        name="api.project.cluster.instance_status",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/(?P<instance_name>[\w\-\.]+)/info/$",  # noqa
        instance_views.GetInstanceStatus.as_view(),
        name="api.project.cluster.instance_status",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/new_update/$",  # noqa
        views.UpdateInstanceNew.as_view(),
        name="api.application.update_new",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/(?P<instance_name>[\w\-\.]+)/scale/$",  # noqa
        views.ScaleInstance.as_view(),
        name="api.application.scale",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/(?P<instance_name>[\w\-\.]+)/cancel/$",  # noqa
        views.CancelUpdateInstance.as_view(),
        name="api.application.cancel",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/(?P<instance_name>[\w\-\.]+)/pause/$",  # noqa
        views.PauseUpdateInstance.as_view(),
        name="api.application.pause",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/(?P<instance_name>[\w\-\.]+)/resume/$",  # noqa
        views.ResumeUpdateInstance.as_view(),
        name="api.application.resume",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/(?P<instance_name>[\w\-\.]+)/delete/$",  # noqa
        views.DeleteInstance.as_view(),
        name="api.application.delete",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/(?P<instance_name>[\w\-\.]+)/recreate/$",  # noqa
        views.ReCreateInstance.as_view(),
        name="api.application.recreate",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/(?P<instance_name>[\w\-\.]+)/labels/$",  # noqa
        instance_views.GetInstanceLabels.as_view(),
        name="api.application.labels",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/(?P<instance_name>[\w\-\.]+)/annotations/$",  # noqa
        instance_views.GetInstanceAnnotations.as_view(),
        name="api.application.annotations",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/rescheduler/$",  # noqa
        instance_views.ReschedulerTaskgroup.as_view(),
        name="api.application.rescheduler",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/events/$",  # noqa
        instance_views.TaskgroupEvents.as_view(),
        name="api.application.events",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/taskgroups/(?P<taskgroup_name>[\w\-\.]+)/containers/(?P<container_id>[\w\-]+)/info/$",  # noqa
        instance_views.ContainerInfo.as_view(),
        name="api.application.container_info",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/taskgroups/(?P<name>[\w\-\.]+)/containers/info/$",  # noqa
        common_query.K8SContainerInfo.as_view({'post': 'container_info'}),
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/pods/(?P<name>[\w\-\.]+)/containers/env_info/$",  # noqa
        common_query.K8SContainerInfo.as_view({'post': 'env_info'}),
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/containers/(?P<container_id>[\w\-]+)/logs/$",  # noqa
        instance_views.ContainerLogs.as_view(),
        name="api.application.container_logs",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/configs/$",  # noqa
        instance_views.InstanceConfigInfo.as_view(),
        name="api.application.instance_config",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/versions/$",  # noqa
        instance_views.GetVersionList.as_view(),
        name="api.application.version_list",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/batch/",
        instance_views.BatchInstances.as_view(),
    ),
    #######################################################
    # 优化项，针对模板集、模板、实例三方面进行调优
    ########################################################
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/musters/$",  # noqa
        views.GetProjectMuster.as_view(),
        # tmpl_set_views.TemplateSet.as_view(),
        name="api.application.muster",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/musters/(?P<muster_id>\d+)/templates/$",  # noqa
        views.GetMusterTemplate.as_view(),
        name="api.application.muster,template",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/musters/(?P<muster_id>\d+)/templates/(?P<template_id>\d+)/instances/$",  # noqa
        views.AppInstance.as_view(),
        name="api.application.template.instance",
    ),
    #######################################################
    # 针对taskgroup的优化
    ######################################################
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/taskgroups/$",  # noqa
        instance_views.QueryAllTaskgroups.as_view(),
        name="api.application.taskgroups",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/taskgroups/(?P<taskgroup_name>[\w\-\.]+)/containers/$",  # noqa
        instance_views.QueryContainersByTaskgroup.as_view(),
        name="api.application.taskgroup.containers",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/taskgroups/(?P<taskgroup_name>[\w\-\.]+)/info/$",  # noqa
        instance_views.QueryTaskgroupInfo.as_view(),
        name="api.application.taskgroup.info",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/containers/$",  # noqa
        instance_views.QueryApplicationContainers.as_view(),
        name="api.application.containers",
    ),
    #######################################################
    # 针对模板的view
    #######################################################
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/musters/(?P<muster_id>\d+)/instances/namespaces/",
        templates.TemplateNamespace.as_view(),
        name="api.application.template.namespace",
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/musters/(?P<muster_id>\d+)/instances/resources/$",  # noqa
        templates.DeleteTemplateInstance.as_view(),
        name="api.application.template.instances",
    ),
    ######################################################
    # 节点详情页跳转得到容器详情
    ######################################################
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/clusters/(?P<cluster_id>[\w\-]+)/container/",  # noqa
        instance_views.QueryContainerInfo.as_view(),
    ),
    #####################################################
    # 命名空间相关
    ####################################################
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/namespaces/$",
        ns_views.GetProjectNamespace.as_view(),
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/namespaces/(?P<ns_id>\d+)/instances/$",
        ns_views.GetInstances.as_view(),
    ),
    ####################################################
    # 实例版本相关
    ###################################################
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<inst_id>\d+)/version_conf/$",
        instance_views.GetInstanceVersionConf.as_view(),
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<inst_id>\d+)/all_versions/$",
        instance_views.GetInstanceVersions.as_view(),
    ),
    ###################################################
    # 过滤器相关
    ###################################################
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/all_musters/$",
        filter_views.GetAllMusters.as_view(),
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/all_instances/$",
        filter_views.GetAllInstances.as_view(),
    ),
    url(
        r"^api/app/projects/(?P<project_id>[\w\-]+)/all_namespaces/$",
        filter_views.GetAllNamespaces.as_view(),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/all_configs/$',
        operation.RollbackPreviousVersion.as_view({'get': 'get'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/rollback/$',
        operation.RollbackPreviousVersion.as_view({'put': 'update'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/pods/reschedule/$',
        operation.ReschedulePodsViewSet.as_view({'put': 'reschedule_pods'}),
    ),
]
