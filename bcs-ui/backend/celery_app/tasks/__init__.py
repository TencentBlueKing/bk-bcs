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

from celery import shared_task

from .application import application_polling_task, delete_instance_task, update_create_error_record


@shared_task
def healthz(n):
    return -n


try:
    from . import cluster
except ImportError:
    pass
else:
    from .cluster import (
        chain_polling_bke_status,
        chain_polling_task,
        delete_cluster_node,
        delete_cluster_node_polling,
        delete_cluster_task,
        exec_bcs_task,
        force_delete_node,
        polling_bke_status,
        polling_initial_task,
        polling_so_init,
        polling_task,
        so_init,
    )
