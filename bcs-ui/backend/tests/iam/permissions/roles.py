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

ADMIN_USER = 'admin'

ANONYMOUS_USER = 'anonymous_user'

PROJECT_NO_VIEW_USER = 'project_no_view_user'  # 无 project_view 权限
NO_PROJECT_USER = 'no_project_user'  # 无 project 任何权限

CLUSTER_USER = 'cluster_user'  # 有 cluster 所有权限
CLUSTER_MANAGE_NOT_VIEW_USER = 'cluster_manage_not_view_user'  # 有 cluster_manage 无 cluster_view 权限
PROJECT_CLUSTER_USER = 'project_cluster_user'  # 有 cluster 和 project 所有权限
PROJECT_NO_CLUSTER_USER = 'project_no_cluster_user'  # 有 project 但无 cluster 权限
CLUSTER_NO_PROJECT_USER = 'cluster_no_project_user'  # 有 cluster 但无 project 权限

NAMESPACE_NO_CLUSTER_PROJECT_USER = 'namespace_no_cluster_project_user'  # 有 namespace 但无 project 和 cluster 权限

TEMPLATESET_USER = 'templateset_user'  # 有 templateset 所有权限
PROJECT_TEMPLATESET_USER = 'project_templateset_user'  # 有 templateset 和 project 权限
TEMPLATESET_NO_PROJECT_USER = 'templateset_no_project_user'  # 有 templateset 但无 project 权限

CLUSTER_SCOPED_NO_CLUSTER_USER = 'cluster_scoped_no_cluster_user'  # 有 cluster_scoped 但无 cluster 权限

NAMESPACE_SCOPED_NO_VIEW_USER = (
    'namespace_scoped_no_view_user'  # 有 namespace_scoped_update 但无 namespace_scoped_view 权限
)
