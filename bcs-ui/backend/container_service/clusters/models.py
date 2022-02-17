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

from django.db import models
from django.utils.translation import ugettext_lazy as _

from backend.templatesets.legacy_apps.configuration.models import BaseModel

logger = logging.getLogger(__name__)


class NodeOperType:
    BkeInstall = "bke_install"
    NodeInstall = "initialize"
    NodeRemove = "remove"
    InitialCheck = 'initial_check'
    SoInitial = 'so_initial'
    NodeReinstall = 'reinstall'


class ClusterOperType:
    ClusterInstall = 'initialize'
    ClusterRemove = 'remove'
    InitialCheck = 'initial_check'
    SoInitial = 'so_initial'
    ClusterReinstall = 'reinstall'
    ClusterUpgrade = "upgrade"
    ClusterReupgrade = "reupgrade"


class CommonStatus:
    InitialChecking = "initial_checking"
    InitialCheckFailed = "check_failed"
    Uninitialized = "uninitialized"
    Initializing = "initializing"
    InitialFailed = "initial_failed"
    Normal = "normal"
    Removing = "removing"
    Removed = "removed"
    RemoveFailed = "remove_failed"
    SoInitial = "so_initializing"
    SoInitialFailed = "so_init_failed"
    ScheduleFailed = "schedule_failed"
    Scheduling = "scheduling"
    DeleteFailed = "delete_failed"


class ClusterStatus:
    Uninitialized = "uninitialized"
    Initializing = "initializing"
    InitialFailed = "initial_failed"
    Normal = "normal"
    Upgrading = "upgrading"
    UpgradeFailed = "upgrade_failed"


class NodeStatus:
    Uninitialized = "uninitialized"
    Initializing = "initializing"
    InitialFailed = "initial_failed"
    Normal = "normal"
    ToRemoved = "to_removed"
    Removable = "removable"
    Removing = "removing"
    RemoveFailed = "remove_failed"
    Removed = "removed"
    BkeInstall = "bke_installing"
    BkeFailed = "bke_failed"
    NotReady = "not_ready"


class GcloudPollingTask(models.Model):
    project_id = models.CharField(max_length=64)
    task_id = models.CharField(max_length=64, null=True)
    token = models.CharField(max_length=64, null=True)
    operator = models.CharField(max_length=64, null=True)
    params = models.TextField()
    is_finished = models.BooleanField(default=False)
    is_polling = models.BooleanField(default=False)
    create_at = models.DateTimeField(auto_now_add=True)
    update_at = models.DateTimeField(auto_now=True)
    cluster_id = models.CharField(max_length=32)
    log = models.TextField()

    class Meta:
        abstract = True

    def set_is_finish(self, flag):
        self.is_finished = flag
        self.save()

    def set_is_polling(self, flag):
        self.is_polling = flag
        self.save()

    def set_status(self, status):
        self.status = status
        self.save()

    def set_finish_polling_status(self, finish_flag, polling_flag, status):
        self.is_finished = finish_flag
        self.is_polling = polling_flag
        self.status = status
        self.save()

    def set_task_id(self, task_id):
        self.task_id = task_id
        self.save()

    def set_params(self, params):
        self.params = json.dumps(params)
        self.save()

    @property
    def log_params(self):
        try:
            return json.loads(self.params)
        except Exception:
            return {}

    def activate_polling(self):
        from celery import chain

        from backend.celery_app.tasks import cluster

        chain(
            cluster.polling_task.s(self.__class__.__name__, self.pk), cluster.chain_polling_bke_status.s()
        ).apply_async()

    def polling_task(self):
        """轮训任务"""
        from backend.container_service.clusters import tasks

        tasks.ClusterOrNodeTaskPoller.start(
            {"model_type": self.__class__.__name__, "pk": self.pk}, tasks.TaskStatusResultHandler
        )


class ClusterInstallLog(GcloudPollingTask):
    OPER_TYPE = (
        ("initialize", _("初始化集群")),
        ("reinstall", _("重新初始化集群")),
        ("initial_check", _("前置检查")),
        ("removing", _("删除集群")),
        ("so_initial", _("SO 机器初始化")),
        ("remove", _("删除集群")),
        ("upgrade", _("升级集群版本")),
    )
    status = models.CharField(max_length=32, null=True, blank=True)
    oper_type = models.CharField(max_length=16, choices=OPER_TYPE, default="initialize")

    def cluster_check_and_init_polling(self):
        """集群前置检查&初始化"""
        from celery import chain

        from backend.celery_app.tasks import cluster

        chain(
            cluster.polling_initial_task.s(self.__class__.__name__, self.pk),
            cluster.so_init.s(),
            cluster.polling_so_init.s(),
            cluster.exec_bcs_task.s(),
            cluster.chain_polling_task.s(self.__class__.__name__),
        ).apply_async()

    def cluster_so_and_init_polling(self):
        """集群so&初始化"""
        from celery import chain

        from backend.celery_app.tasks import cluster

        chain(
            cluster.polling_so_init.s(None, self.pk, self.__class__.__name__),
            cluster.exec_bcs_task.s(),
            cluster.chain_polling_task.s(self.__class__.__name__),
        ).apply_async()

    def delete_cluster(self):
        """删除集群"""
        from backend.celery_app.tasks import cluster

        cluster.delete_cluster_task.delay(self.__class__.__name__, self.pk)


class NodeUpdateLog(GcloudPollingTask):
    OPER_TYPE = (
        ("initialize", _("初始化节点")),
        ("reinstall", _("重新初始化节点")),
        ("removing", _("删除节点")),
        ("initial_check", _("前置检查")),
        ("so_initial", _("SO 机器初始化")),
        ("bke_install", _("安装BKE")),
        ("remove", _("删除节点")),
        ('bind_lb', _("绑定LB")),
        ('init_env', _("初始化环境")),
    )
    # node_id = models.CharField(max_length=32)
    node_id = models.TextField()
    status = models.CharField(max_length=32, null=True, blank=True)
    oper_type = models.CharField(max_length=16, choices=OPER_TYPE, default="initialize")

    def node_check_and_init_polling(self):
        """节点前置检查&初始化"""
        from celery import chain

        from backend.celery_app.tasks import cluster

        chain(
            cluster.polling_initial_task.s(self.__class__.__name__, self.pk),
            cluster.so_init.s(),
            cluster.polling_so_init.s(),
            cluster.node_exec_bcs_task.s(),
            cluster.chain_polling_task.s(self.__class__.__name__),
            cluster.chain_polling_bke_status.s(),
        ).apply_async()

    def node_so_and_init_polling(self):
        """节点so&初始化"""
        from celery import chain

        from backend.celery_app.tasks import cluster

        chain(
            cluster.polling_so_init.s(None, self.pk, self.__class__.__name__),
            cluster.node_exec_bcs_task.s(),
            cluster.chain_polling_task.s(self.__class__.__name__),
            cluster.chain_polling_bke_status.s(),
        ).apply_async()

    def bke_polling(self):
        from backend.celery_app.tasks import cluster

        cluster.polling_bke_status.delay(self.pk)

    def node_force_delete_polling(self):
        """强制删除节点"""
        from celery import chain

        from backend.celery_app.tasks import cluster

        chain(
            cluster.force_delete_node.s(self.__class__.__name__, self.pk),
            cluster.delete_cluster_node.s(),
            cluster.delete_cluster_node_polling.s(),
        ).apply_async()


def log_factory(log_type):
    if log_type == "ClusterInstallLog":
        return ClusterInstallLog
    elif log_type == "NodeUpdateLog":
        return NodeUpdateLog


class NodeLabel(BaseModel):
    project_id = models.CharField(_("项目ID"), max_length=32)
    cluster_id = models.CharField(_("集群ID"), max_length=32)
    node_id = models.IntegerField()
    labels = models.TextField()

    @property
    def node_labels(self):
        return json.loads(self.labels)

    class Meta:
        db_table = "node_label"
        unique_together = ("node_id",)
