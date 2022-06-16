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
from pathlib import Path
from typing import Dict, List, NewType, Optional, Union

import attr
from celery import shared_task
from django.utils.crypto import get_random_string

from backend.helm.helm import models
from backend.helm.helm.providers.constants import PUBLIC_REPO_URL
from backend.helm.toolkit.kubehelm.helm import KubeHelmClient
from backend.helm.toolkit.kubehelm.options import RawFlag
from backend.packages.blue_krill.data_types.enum import StructuredEnum
from backend.utils import client as bcs_client
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)

ArgsDict = NewType('ArgsDict', Dict[str, Union[str, List[RawFlag]]])

VALUESFILE_KEY = '--valuesfile'


class HelmOperation(str, StructuredEnum):
    INSTALL = 'install'
    UPGRADE = 'upgrade'
    UNINSTALL = 'uninstall'
    ROLLBACK = 'rollback'


@attr.s(auto_attribs=True)
class ReleaseArgs:
    """HelmDeployer 的执行参数.

    options 有序示例 [
    {"--valuesfile": {"file_name": "", "file_content": ""}},
    {"--set": "a=1"},
    {"--valuesfile": {"file_name": "", "file_content": ""}}
    {"--post-renderer": {"bcs_inject_data": ""}}
    ]
    其他参考官方 options https://helm.sh/docs/helm/helm_install/#options
    """

    project_id: str
    cluster_id: str
    # release name
    name: str
    namespace: str
    operator: str
    chart_url: str
    options: List[RawFlag] = attr.field(factory=list)

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'ReleaseArgs':
        field_names = [f.name for f in attr.fields(cls)]
        return cls(**{k: v for k, v in init_data.items() if k in field_names})


class HelmDeployer:
    """
    only for helm3 use absolute url as chart args
    """

    def __init__(self, access_token: str, helm_args: ReleaseArgs):
        self.access_token = access_token
        self.helm_args = helm_args

    def install(self):
        self.run_with_helm(HelmOperation.INSTALL)

    def upgrade(self):
        self.run_with_helm(HelmOperation.UPGRADE)

    def uninstall(self):
        self.run_with_helm(HelmOperation.UNINSTALL)

    def run_with_helm(self, op: str):
        """执行 Helm 命令

        :param op: helm 操作, 如 install, upgrade 等
        """
        err_msg = ''

        with bcs_client.make_helm_client(
            project_id=self.helm_args.project_id,
            cluster_id=self.helm_args.cluster_id,
            access_token=self.access_token,
        ) as (client, err):
            if err is not None:
                err_msg = f'make helm client failed for cluster({self.helm_args.cluster_id}): {err}'

            if not err_msg:
                try:
                    self._run_with_helm(client, op)
                except Exception as e:
                    err_msg = f'helm {op} failed in cluster({self.helm_args.cluster_id}): {e}'

        if err_msg:
            raise error_codes.ComponentError(err_msg)

    def _run_with_helm(self, helm_client: KubeHelmClient, op: str):
        self._inject_auth_options()
        with self._reconfigure_values_and_post_renderer():
            if op == HelmOperation.UNINSTALL:
                helm_client.uninstall(self.helm_args.name, self.helm_args.namespace)
            elif op in [HelmOperation.INSTALL, HelmOperation.UPGRADE]:
                helm_client.do_install_or_upgrade(
                    op,
                    self.helm_args.name,
                    self.helm_args.namespace,
                    self.helm_args.chart_url,
                    self.helm_args.options,
                )
            else:
                raise NotImplementedError(f'unsupported op {op}')

    def _inject_auth_options(self):
        """注入用户凭证"""

        # chart_url 如 'http://repo.example.com/bcs-public-project/helm-public-repo/charts/xxx-operator-0.5.0.tgz'
        chart_url = self.helm_args.chart_url
        repo_url = chart_url.partition('/charts/')[0]
        if repo_url == PUBLIC_REPO_URL:  # 平台公共仓库不需要auth信息
            return

        repository = models.Repository.objects.filter(
            url__startswith=repo_url, project_id=self.helm_args.project_id
        ).first()
        if not repository:  # 非平台chart repo url，由用户决定--username xxx --password xxx
            return

        credentials = models.RepositoryAuth.objects.get(repo=repository).credentials_decoded
        self.helm_args.options.extend(
            [{'--username': credentials['username']}, {'--password': credentials['password']}]
        )

    @contextlib.contextmanager
    def _reconfigure_values_and_post_renderer(self):
        """整理 values, 增加 post-renderer

        TODO 增加 ytt 的 post-renderer 支持
        """
        idx_list = self._find_valuesfile()
        if not idx_list:
            yield
        else:
            # 如果 options 包含 --valuesfile, 需要生成临时 values 文件
            with tempfile.TemporaryDirectory() as temp_dir:
                # 为了保留原始 options 的顺序，通过找到索引重新赋值
                for idx in idx_list:
                    # --valuesfile 非 helm 官方 flag, 需要转换成 --values
                    self.helm_args.options[idx] = {
                        '--values': self._write_values(temp_dir, self.helm_args.options[idx][VALUESFILE_KEY])
                    }
                yield

    def _find_valuesfile(self) -> List[int]:
        """查找 options 中 --valuesfile 的配置位置"""
        idx_list = []
        for idx, flag in enumerate(self.helm_args.options):
            if isinstance(flag, dict) and VALUESFILE_KEY in flag.keys():
                idx_list.append(idx)
        return idx_list

    def _write_values(self, temp_dir: str, valuesfile: Dict[str, str]) -> str:
        """将 valuesfile 中的内容写入文件

        :param valuesfile: {'file_name': '', 'file_content': ''}
        :return 写入的文件位置
        """
        path = Path(temp_dir) / valuesfile['file_name']
        path.write_text(valuesfile['file_content'])
        return str(path.resolve())


def make_valuesfile_flag(values: str, file_name: Optional[str] = None) -> RawFlag:
    return RawFlag(
        {VALUESFILE_KEY: {'file_content': values, 'file_name': file_name or f'{get_random_string(12)}.yaml'}}
    )


@shared_task
def helm_install(access_token: str, release_args: ArgsDict) -> Optional[str]:
    try:
        deployer = HelmDeployer(access_token=access_token, helm_args=ReleaseArgs.from_dict(release_args))
        deployer.install()
    except Exception as e:
        return _generate_err_msg(HelmOperation.INSTALL, release_args, e)


@shared_task
def helm_upgrade(access_token: str, release_args: ArgsDict) -> Optional[str]:
    try:
        deployer = HelmDeployer(access_token=access_token, helm_args=ReleaseArgs.from_dict(release_args))
        deployer.upgrade()
    except Exception as e:
        return _generate_err_msg(HelmOperation.UPGRADE, release_args, e)


@shared_task
def helm_uninstall(access_token: str, release_args: ArgsDict) -> Optional[str]:
    try:
        deployer = HelmDeployer(access_token=access_token, helm_args=ReleaseArgs.from_dict(release_args))
        deployer.uninstall()
    except Exception as e:
        err_msg = _generate_err_msg(HelmOperation.UNINSTALL, release_args, e)
        # 忽略 release: not found 的错误
        if 'release: not found' in err_msg:
            return
        return err_msg


def _generate_err_msg(op: str, release_args: ArgsDict, e: Exception) -> str:
    err_msg = (
        f"helm {op} {release_args['name']} -n {release_args['namespace']} failed "
        f"in cluster({release_args['cluster_id']}): {e}"
    )
    logger.error(err_msg)
    return err_msg
