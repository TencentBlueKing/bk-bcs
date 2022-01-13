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
import contextlib
import json
import logging
import traceback
from dataclasses import dataclass

from rest_framework.exceptions import PermissionDenied, ValidationError

from backend.kube_core.toolkit.kubectl.exceptions import KubectlError, KubectlExecutionError
from backend.utils import client as bcs_client
from backend.utils.basic import ChoicesEnum
from backend.utils.client import make_kubectl_client, make_kubectl_client_from_kubeconfig

from ..helm.bcs_variable import get_valuefile_with_bcs_variable_injected
from ..toolkit import utils as bcs_helm_utils
from ..toolkit.kubehelm.exceptions import HelmError, HelmExecutionError
from .utils import get_cc_app_id

logger = logging.getLogger(__name__)


class ChartOperations(ChoicesEnum):
    INSTALL = "install"
    UPGRADE = "upgrade"
    UNINSTALL = "uninstall"
    ROLLBACK = "rollback"


@dataclass
class AppDeployer:
    """AppDeployEngine manages app's deploy operations"""

    app: object
    access_token: str
    kubeconfig_content: str = None
    ignore_empty_access_token: bool = False
    extra_inject_source: dict = None

    @contextlib.contextmanager
    def make_kubectl_client(self):
        with make_kubectl_client(
            project_id=self.app.project_id, cluster_id=self.app.cluster_id, access_token=self.access_token
        ) as (client, err):
            yield client, err

    @contextlib.contextmanager
    def make_helm_client(self):
        with bcs_client.make_helm_client(
            project_id=self.app.project_id, cluster_id=self.app.cluster_id, access_token=self.access_token
        ) as (client, err):
            yield client, err

    def install_app_by_helm(self):
        """通过helm实例化"""
        self.run_with_helm(ChartOperations.INSTALL.value)

    def install_app_by_kubectl(self):
        """通过kubectl实例化"""
        self.run_with_kubectl(ChartOperations.INSTALL.value)

    def upgrade_app_by_helm(self):
        """通过helm升级 app"""
        self.run_with_helm(ChartOperations.UPGRADE.value)

    def upgrade_app_by_kubectl(self):
        """通过kubectl升级 app"""
        self.run_with_kubectl(ChartOperations.INSTALL.value)

    def uninstall_app_by_helm(self):
        """通过helm删除/卸载 app"""
        self.run_with_helm(ChartOperations.UNINSTALL.value)

    def uninstall_app_by_kubectl(self):
        """通过kubectl删除/卸载 app"""
        self.run_with_kubectl(ChartOperations.UNINSTALL.value)

    def rollback_app_by_helm(self):
        """通过helm回滚app版本"""
        self.run_with_helm(ChartOperations.ROLLBACK.value)

    def rollback_app_by_kubectl(self):
        """通过kubectl回滚app版本"""
        self.run_with_kubectl(ChartOperations.INSTALL.value)

    def run_with_helm(self, operation):
        # NOTE: 兼容先前
        if operation in [ChartOperations.INSTALL.value, ChartOperations.UPGRADE.value]:
            content = self.app.render_app(
                access_token=self.access_token,
                username=self.app.updator,
                ignore_empty_access_token=self.ignore_empty_access_token,
                extra_inject_source=self.extra_inject_source,
            )[0]
        elif operation == ChartOperations.UNINSTALL.value:
            content = self.app.release.content
        elif operation == ChartOperations.ROLLBACK.value:
            content = self.app.release.content
        else:
            raise ValidationError("operation not allowed")
        if not content:
            return
        # 保存为release的content
        self.update_app_release_content(content)
        # 使用helm执行相应的命令
        with self.make_helm_client() as (client, err):
            if err is not None:
                transitioning_message = "make helm client failed, %s" % err
                self.app.set_transitioning(False, transitioning_message)
                return
            self._run_with_helm(client, self.app.name, self.app.namespace, operation)

    def get_release_revision(self, cmd_out):
        """解析执行命令的返回
        install和upgrade的返回格式类似:
        NAME: test-redis
        LAST DEPLOYED: Thu Mar 17 17:55:48 2020
        NAMESPACE: default
        STATUS: deployed
        REVISION: 1
        TEST SUITE: None
        """
        cmd_out_list = cmd_out.decode().split("\n")
        for item in cmd_out_list:
            if "REVISION:" not in item:
                continue
            return int(item.split(" ")[-1].strip())

        raise HelmError("parse helm cmd output error")

    def _run_with_helm(self, client, name, namespace, operation):
        transitioning_result = True
        try:
            if operation in [ChartOperations.INSTALL.value, ChartOperations.UPGRADE.value]:
                project_id = self.app.project_id
                namespace = self.app.namespace
                bcs_inject_data = bcs_helm_utils.BCSInjectData(
                    source_type="helm",
                    creator=self.app.creator,
                    updator=self.app.updator,
                    version=self.app.release.chartVersionSnapshot.version,
                    project_id=project_id,
                    app_id=get_cc_app_id(self.access_token, project_id),
                    cluster_id=self.app.cluster_id,
                    namespace=namespace,
                    stdlog_data_id=bcs_helm_utils.get_stdlog_data_id(project_id),
                )
                # 追加系统和用户渲染的变量
                values_with_bcs_variables = get_valuefile_with_bcs_variable_injected(
                    access_token=self.access_token,
                    project_id=project_id,
                    namespace_id=self.app.namespace_id,
                    valuefile=self.app.release.valuefile,
                    cluster_id=self.app.cluster_id,
                )
                # 获取执行的操作命令
                cmd_out = getattr(client, operation)(
                    name=name,
                    namespace=namespace,
                    files=self.app.release.chartVersionSnapshot.files,
                    chart_values=values_with_bcs_variables,
                    bcs_inject_data=bcs_inject_data,
                    cmd_flags=json.loads(self.app.cmd_flags),
                )[0]
                self.app.release.revision = self.get_release_revision(cmd_out)
                self.app.release.save()
            elif operation == ChartOperations.UNINSTALL.value:
                client.uninstall(name, namespace)
            elif operation == ChartOperations.ROLLBACK.value:
                client.rollback(name, namespace, self.app.release.revision)
        except HelmExecutionError as e:
            transitioning_result = False
            transitioning_message = (
                "helm command execute failed.\n" "Error code: {error_no}\nOutput:\n{output}"
            ).format(error_no=e.error_no, output=e.output)
            logger.warn(transitioning_message)
        except HelmError as e:
            err_msg = str(e)
            logger.warn(err_msg)
            # TODO: 现阶段针对删除release找不到的情况，认为是正常的
            if "not found" in err_msg and operation == ChartOperations.UNINSTALL.value:
                transitioning_result = True
                transitioning_message = "app success %s" % operation
            else:
                transitioning_result = False
                transitioning_message = err_msg
        except Exception as e:
            err_msg = str(e)
            transitioning_result = False
            logger.warning(err_msg)
            transitioning_message = self.collect_transitioning_error_message(e)
        else:
            transitioning_result = True
            transitioning_message = "app success %s" % operation

        self.app.set_transitioning(transitioning_result, transitioning_message)

    def run_with_kubectl(self, operation):
        if operation == "uninstall":
            # just load content from release, so that avoid unnecessary render exceptions
            content = self.app.release.content
        else:
            content, _ = self.app.render_app(
                access_token=self.access_token,
                username=self.app.updator,
                ignore_empty_access_token=self.ignore_empty_access_token,
                extra_inject_source=self.extra_inject_source,
            )

        if content is None:
            return

        self.update_app_release_content(content)

        if self.access_token:
            with self.make_kubectl_client() as (client, err):
                if err is not None:
                    transitioning_message = "make kubectl client failed, %s" % err
                    self.app.set_transitioning(False, transitioning_message)
                    return
                else:
                    self.run_with_kubectl_core(content, operation, client)
        elif self.ignore_empty_access_token:
            if self.kubeconfig_content:
                with make_kubectl_client_from_kubeconfig(self.kubeconfig_content) as client:
                    self.run_with_kubectl_core(content, operation, client)
            else:
                raise PermissionDenied("api access must supply valid kubeconfig")
        else:
            raise ValueError(self)

    def run_with_kubectl_core(self, content, operation, client):
        transitioning_result = True
        try:
            if operation == "install":
                client.ensure_namespace(self.app.namespace)
                client.apply(template=content, namespace=self.app.namespace)
            elif operation == "uninstall":
                client.ensure_namespace(self.app.namespace)
                client.delete_one_by_one(self.app.release.extract_structure(self.app.namespace), self.app.namespace)
                # client.delete(template=content, namespace=self.app.namespace)
            else:
                raise ValueError(operation)
        except KubectlExecutionError as e:
            transitioning_result = False
            transitioning_message = (
                "kubectl command execute failed.\n" "Error code: {error_no}\nOutput:\n{output}"
            ).format(error_no=e.error_no, output=e.output)
            logger.warn(transitioning_message)
        except KubectlError as e:
            transitioning_result = False
            logger.warn(e.message)
            transitioning_message = e.message
        except Exception as e:
            transitioning_result = False
            logger.warning(e.message)
            transitioning_message = self.collect_transitioning_error_message(e)
        else:
            transitioning_result = True
            transitioning_message = "app success %s" % operation

        self.app.set_transitioning(transitioning_result, transitioning_message)

    def collect_transitioning_error_message(self, error):
        return "{error}\n{stack}".format(error=error, stack=traceback.format_exc())

    def update_app_release_content(self, content):
        release = self.app.release
        release.content = content
        release.save(update_fields=["content"])
        release.refresh_structure(self.app.namespace)
