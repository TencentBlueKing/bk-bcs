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
import logging
import tempfile
from dataclasses import dataclass, field
from pathlib import Path
from typing import Dict, List

from celery import shared_task
from django.utils.translation import ugettext_lazy as _

from backend.helm.helm import models
from backend.helm.helm.providers.constants import PUBLIC_REPO_URL
from backend.helm.toolkit.kubehelm.options import RawFlag
from backend.utils import client as bcs_client
from backend.utils.basic import ChoicesEnum
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)

VALUESFILE_KEY = "--valuesfile"


class HelmOperations(ChoicesEnum):
    INSTALL = "install"
    UPGRADE = "upgrade"
    UNINSTALL = "uninstall"
    ROLLBACK = "rollback"


@dataclass
class BaseArgs:
    project_id: str
    cluster_id: str
    name: str  # release name
    namespace: str
    operator: str


@dataclass
class ReleaseArgs(BaseArgs):
    """
    options 有序示例 [
    {"--valuesfile": {"file_name": "", "file_content": ""}},
    {"--set": "a=1"},
    {"--valuesfile": {"file_name": "", "file_content": ""}}
    {"--post-renderer": {"bcs_inject_data": ""}}
    ]
    其他参考官方options https://helm.sh/docs/helm/helm_install/#options
    """

    chart_url: str
    options: List[RawFlag] = field(default_factory=list)


class HelmTool:
    def __init__(self, access_token, helm_args):
        self.access_token = access_token
        self.helm_args = helm_args

    def _run_with_helm(self, client, operation):
        pass

    def run_with_helm(self, operation):
        err_msg = ""

        with bcs_client.make_helm_client(
            project_id=self.helm_args.project_id,
            cluster_id=self.helm_args.cluster_id,
            access_token=self.access_token,
        ) as (client, err):
            if err is not None:
                err_msg = f"make helm client failed: {err}"

            if not err_msg:
                try:
                    self._run_with_helm(client, operation)
                except Exception as e:
                    err_msg = f"helm {operation} failed: {e}"

        if err_msg:
            raise error_codes.ComponentError(err_msg)


class HelmDeployer(HelmTool):
    """
    only for helm3 use absolute url as chart args
    """

    def install(self):
        self.run_with_helm(HelmOperations.INSTALL.value)

    def upgrade(self):
        self.run_with_helm(HelmOperations.UPGRADE.value)

    def _inject_auth_options(self):
        # chart_url示例如下
        # "http://repo.example.com/bcs-public-project/helm-public-repo/charts/bcs-gamestatefulset-operator-0.5.0.tgz"
        chart_url = self.helm_args.chart_url
        repo_url, _, _ = chart_url.partition("/charts/")
        if f"{repo_url}/" == PUBLIC_REPO_URL:  # 平台公共仓库不需要auth信息
            return

        repository = models.Repository.objects.filter(url__startswith=repo_url).first()
        if not repository:  # 非平台chart repo url，由用户决定--username xxx --password xxx
            return

        credentials = models.RepositoryAuth.objects.get(repo=repository).credentials_decoded
        self.helm_args.options.extend(
            [{"--username": credentials["username"]}, {"--password": credentials["password"]}]
        )

    def _run_with_helm(self, helm_client, operation):
        self._inject_auth_options()
        with self._reconfigure_values_and_post_renderer():
            helm_client.do_install_or_upgrade(
                operation,
                self.helm_args.name,
                self.helm_args.namespace,
                self.helm_args.chart_url,
                self.helm_args.options,
            )

    def _write_values(self, temp_dir: str, valuesfile: Dict[str, str]) -> str:
        """
        :param valuesfile: {"file_name": "", "file_content": ""}
        """
        path = Path(temp_dir) / valuesfile["file_name"]
        path.write_text(valuesfile["file_content"])
        return str(path.resolve())

    def _find_valuesfile(self) -> List[int]:
        indx_list = []
        for indx, flag in enumerate(self.helm_args.options):
            if isinstance(flag, dict) and VALUESFILE_KEY in flag.keys():
                indx_list.append(indx)
        return indx_list

    @contextlib.contextmanager
    def _reconfigure_values_and_post_renderer(self):
        """
        如果options包含--valuesfile, 需要生成临时values文件
        """
        # TODO 增加ytt的post-renderer支持
        indx_list = self._find_valuesfile()
        if not indx_list:
            yield
        else:
            with tempfile.TemporaryDirectory() as temp_dir:
                # 为了保留原始options的顺序，通过找到索引重新赋值
                for indx in indx_list:
                    # --valuesfile非helm官方flag, 需要转换成--values
                    self.helm_args.options[indx] = {
                        "--values": self._write_values(temp_dir, self.helm_args.options[indx][VALUESFILE_KEY])
                    }
                yield


def _generate_err_msg(operation, release_args, exception_msg):
    err_msg = _("集群{} helm {} {} -n {} 失败: {}").format(
        operation,
        release_args.get("cluster_id"),
        release_args.get("name"),
        release_args.get("namespace"),
        exception_msg,
    )
    logger.error(err_msg)
    return err_msg


@shared_task
def helm_install(access_token, release_args):
    try:
        deployer = HelmDeployer(access_token=access_token, helm_args=ReleaseArgs(**release_args))
        deployer.install()
    except Exception as e:
        return _generate_err_msg(HelmOperations.INSTALL.value, release_args, e)


@shared_task
def helm_upgrade(access_token, release_args):
    try:
        deployer = HelmDeployer(access_token=access_token, helm_args=ReleaseArgs(**release_args))
        deployer.upgrade()
    except Exception as e:
        return _generate_err_msg(HelmOperations.UPGRADE.value, release_args, e)
