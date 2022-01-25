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

this is the operations for helm cli

required: helm 2.9.1+
"""

import contextlib
import json
import logging
import os
import shutil
import stat
import subprocess
import tempfile
import time
from dataclasses import asdict
from typing import Dict, List

from django.conf import settings
from django.template.loader import render_to_string

from backend.apps.whitelist import enable_helm_v3
from backend.container_service.clusters.base.utils import get_cluster_type
from backend.container_service.clusters.constants import ClusterType

from .exceptions import HelmError, HelmExecutionError, HelmMaxTryError
from .options import Options

logger = logging.getLogger(__name__)

YTT_RENDERER_NAME = "ytt_renderer"


def write_files(temp_dir, files):
    for name, content in files.items():
        path = os.path.join(temp_dir, name)
        base_path = os.path.dirname(path)
        if not os.path.exists(base_path):
            os.makedirs(base_path)

        with open(path, "w") as f:
            f.write(content)

    for name, _ in files.items():
        parts = name.split("/")
        if len(parts) > 0:
            return os.path.join(temp_dir, parts[0])

    return temp_dir


class KubeHelmClient:
    """
    render the templates with values.yaml / answers.yaml
    NOTE: helm3 和 helm2的命令执行参数不同
    """

    def __init__(self, helm_bin="helm", kubeconfig=""):
        self.helm_bin = helm_bin
        self.kubeconfig = kubeconfig

    def _make_answers_to_args(self, answers):
        """
        {"a": 1, "b": 2, "c": "3"} => ["--set", "a=1,b=2", "--set-string", "3"]
        """
        if not answers:
            return []
        set_values = ",".join(["{k}={v}".format(k=k, v=v) for k, v in answers.items() if not isinstance(v, str)])
        set_string_values = ",".join(["{k}={v}".format(k=k, v=v) for k, v in answers.items() if isinstance(v, str)])
        return ["--set", set_values, "--set-string", set_string_values]

    def _get_cmd_args_for_template(self, root_dir, app_name, namespace, cluster_id):
        if enable_helm_v3(cluster_id):
            return [settings.HELM3_BIN, "template", app_name, root_dir, "--namespace", namespace]

        return [
            settings.HELM_BIN,
            "template",
            root_dir,
            "--name",
            app_name,
            "--namespace",
            namespace,
        ]

    # def template(self, release, namespace: str):
    def template(self, files, name, namespace, parameters, valuefile, cluster_id=None):
        """
        helm template {dir} --name {name} --namespace {namespace} --set k1=v1,k2=v2,k3=v3 --values filename
        """
        app_name = name or "default"

        temp_dir = tempfile.mkdtemp()
        valuefile_name = None
        try:
            # 1. write template files into fp
            root_dir = write_files(temp_dir, files)

            # 2. parse answers.yaml to values
            values = self._make_answers_to_args(parameters)

            # 3. construct cmd and run
            base_cmd_args = self._get_cmd_args_for_template(root_dir, app_name, namespace, cluster_id)

            # 4.1 helm template
            template_cmd_args = base_cmd_args
            if values:
                template_cmd_args += values

            if valuefile:
                FILENAME = "__valuefile__.yaml"
                valuefile_x = {FILENAME: valuefile}
                write_files(temp_dir, valuefile_x)
                valuefile_name = os.path.join(temp_dir, FILENAME)
                template_cmd_args += ["--values", valuefile_name]

            template_out, _ = self._run_command_with_retry(max_retries=0, cmd_args=template_cmd_args)

            # 4.2 helm template --notes
            notes_out = ""
            # not be used currently, comment it for accelerate
            # notes_cmd_args = base_cmd_args + ["--notes"]
            # notes_out, _ = self._run_command_with_retry(max_retries=0, cmd_args=notes_cmd_args)

        except Exception as e:
            logger.exception(
                (
                    "do helm template fail: namespace={namespace}, name={name}\n"
                    "parameters={parameters}\nvaluefile={valuefile}\nfiles={files}"
                ).format(
                    namespace=namespace,
                    name=name,
                    parameters=parameters,
                    valuefile=valuefile,
                    files=files,
                )
            )
            raise e
        finally:
            shutil.rmtree(temp_dir)

        return template_out, notes_out

    def template_with_ytt_renderer(
        self, files, name, namespace, parameters, valuefile, cluster_id, bcs_inject_data, **kwargs
    ):
        """支持post renderer的helm template，并使用ytt(YAML Templating Tool)注入平台信息
        命令: helm template release_name chart -n namespace --post-renderer ytt-renderer
        """
        try:
            with write_chart_with_ytt(files, bcs_inject_data) as (temp_dir, ytt_config_dir):
                # 1. parse answers.yaml to values
                values = self._make_answers_to_args(parameters)

                # 2. construct cmd and run
                base_cmd_args = [settings.HELM3_BIN, "template", name, temp_dir, "--namespace", namespace]

                # 3. helm template command params
                template_cmd_args = base_cmd_args
                if values:
                    template_cmd_args += values

                # 兼容先前逻辑
                if valuefile:
                    FILENAME = "__valuefile__.yaml"
                    valuefile_x = {FILENAME: valuefile}
                    write_files(temp_dir, valuefile_x)
                    valuefile_name = os.path.join(temp_dir, FILENAME)
                    template_cmd_args += ["--values", valuefile_name]

                # 4. add post render params
                template_cmd_args += ["--post-renderer", f"{ytt_config_dir}/{YTT_RENDERER_NAME}"]

                # 添加命名行参数
                template_cmd_args = self._compose_cmd_args(template_cmd_args, cmd_flags=kwargs.get("cmd_flags"))

                template_out, _ = self._run_command_with_retry(max_retries=0, cmd_args=template_cmd_args)
                # NOTE: 现阶段不需要helm notes输出
                notes_out = ""

        except Exception as e:
            logger.exception(
                (
                    "do helm template fail: namespace={namespace}, name={name}\n"
                    "parameters={parameters}\nvaluefile={valuefile}\nfiles={files}"
                ).format(
                    namespace=namespace,
                    name=name,
                    parameters=parameters,
                    valuefile=valuefile,
                    files=files,
                )
            )
            raise e

        return template_out, notes_out

    def _install_or_upgrade(self, cmd_args, files, chart_values, bcs_inject_data, **kwargs):
        try:
            with write_chart_with_ytt(files, bcs_inject_data) as (temp_dir, ytt_config_dir):
                # NOTE: 设置用户渲染的value文件名为bcs-values.yaml；写入用户渲染的内容
                values_path = os.path.join(temp_dir, "bcs-values.yaml")
                with open(values_path, "w") as f:
                    f.write(chart_values)
                # 组装命令行参数
                cmd_args = self._compose_cmd_args(
                    cmd_args, temp_dir, values_path, ytt_config_dir, kwargs.get("cmd_flags")
                )

                cmd_out, cmd_err = self._run_command_with_retry(max_retries=0, cmd_args=cmd_args)
        except Exception as e:
            logger.exception("执行helm命令失败，命令参数: %s", json.dumps(cmd_args))
            raise e

        return cmd_out, cmd_err

    def install(self, name, namespace, files, chart_values, bcs_inject_data, content=None, **kwargs):
        """install helm chart
        NOTE: 这里需要组装chart格式，才能使用helm install
        必要条件
        - Chart.yaml
        - templates/xxx.yaml
        步骤:
        - 写临时文件, 用于组装chart结构
        - 组装命令行参数
        - 执行命令
        """
        install_cmd_args = [settings.HELM3_BIN, "install", name, "--namespace", namespace]
        return self._install_or_upgrade(install_cmd_args, files, chart_values, bcs_inject_data, **kwargs)

    def upgrade(self, name, namespace, files, chart_values, bcs_inject_data, content=None, **kwargs):
        """upgrade helm release
        NOTE: 这里需要组装chart格式，才能使用helm upgrade
        必要条件
        - Chart.yaml
        - templates/xxx.yaml
        步骤:
        - 写临时文件, 用于组装chart结构
        - 组装命令行参数
        - 执行命令
        """
        # NOTE: helm3需要升级到3.3.1版本
        upgrade_cmd_args = [settings.HELM3_BIN, "upgrade", name, "--namespace", namespace, "--install"]
        return self._install_or_upgrade(upgrade_cmd_args, files, chart_values, bcs_inject_data, **kwargs)

    def _uninstall_or_rollback(self, cmd_args):
        try:
            cmd_out, cmd_err = self._run_command_with_retry(max_retries=0, cmd_args=cmd_args)
        except Exception as e:
            logger.exception("执行helm命令失败，命令参数: %s", json.dumps(cmd_args))
            raise e

        return cmd_out, cmd_err

    def uninstall(self, name, namespace):
        """uninstall helm release"""
        uninstall_cmd_args = [settings.HELM3_BIN, "uninstall", name, "--namespace", namespace]
        return self._uninstall_or_rollback(uninstall_cmd_args)

    def rollback(self, name, namespace, revision):
        """rollback helm release by revision"""
        rollback_cmd_args = [settings.HELM3_BIN, "rollback", name, str(revision), "--namespace", namespace]
        return self._uninstall_or_rollback(rollback_cmd_args)

    def _compose_args_and_run(self, cmd_args, options):
        opts = Options(options)
        cmd_args.extend(opts.options())
        return self._run_command_with_retry(max_retries=0, cmd_args=cmd_args)

    def do_install_or_upgrade(self, operation, name, namespace, chart, options):
        # e.g. helm install mynginx https://example.com/charts/nginx-1.2.3.tgz --username xxx --password xxx
        cmd_args = [settings.HELM3_BIN, operation, name, "--namespace", namespace, chart]
        return self._compose_args_and_run(cmd_args, options)

    def _run_command_with_retry(self, max_retries=1, *args, **kwargs):
        for i in range(max_retries + 1):
            try:
                stdout, stderr = self._run_command(*args, **kwargs)
                return stdout, stderr
            except Exception:
                if i == max_retries:
                    raise

                # retry after 0.5, 1, 1.5, ... seconds
                time.sleep((i + 1) * 0.5)
                continue

        raise HelmMaxTryError(f"max retries {max_retries} fail")

    def _run_command(self, cmd_args):
        """Run the helm command with wrapped exceptions"""
        try:
            logger.info("Calling helm cmd, cmd: (%s)", " ".join(cmd_args))

            proc = subprocess.Popen(
                cmd_args,
                shell=False,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                env={"KUBECONFIG": self.kubeconfig},  # 添加连接集群信息
            )
            stdout, stderr = proc.communicate()

            if proc.returncode != 0:
                logger.exception("Unable to run helm command, return code: %s, output: %s", proc.returncode, stderr)
                raise HelmExecutionError(proc.returncode, stderr)

            return stdout, stderr
        except Exception as err:
            logger.exception("Unable to run helm command")
            raise HelmError("run helm command failed: {}".format(err))

    def _compose_cmd_args(
        self,
        cmd_args: List[str],
        chart_path: str = None,
        values_path: str = None,
        ytt_config_path: str = None,
        cmd_flags: List[Dict] = None,
    ):
        """组装下发的helm命令
        这里需要兼容已经存在的helm的操作
        NOTE: 当用户输入的选项中包含能覆盖values内容的选项，如--set、--set-string等，--values选项必须放在用户输入选项前面
        """
        # 初始的 helm 命令参数
        init_opts = Options(cmd_args)
        if chart_path:
            init_opts.add(chart_path)
        if ytt_config_path:
            init_opts.add({"--post-renderer": f"{ytt_config_path}/{YTT_RENDERER_NAME}"})

        # 用户输入的配置选项
        opts = Options(cmd_flags)
        options = opts.options()
        # 当--reuse-values不存在并且values内容存在时，添加--values选项
        if "--reuse-values" not in options and values_path:
            init_opts.add({"--values": values_path})

        cmd_args = init_opts.options()
        cmd_args.extend(options)
        return cmd_args


@contextlib.contextmanager
def write_chart_with_ytt(files, bcs_inject_data):
    """组装helm template功能需要的文件，并且使用ytt注入平台需要的信息
    主要包含以下两部分
    - chart部分
    - ytt配置部分
    """
    with tempfile.TemporaryDirectory() as temp_dir:
        for name, content in files.items():
            path = os.path.join(temp_dir, name)
            base_path = os.path.dirname(path)
            if not os.path.exists(base_path):
                os.makedirs(base_path)

            with open(path, "w") as f:
                f.write(content)

        # 获取chart配置的目录
        chart_dir = temp_dir
        for name, _ in files.items():
            parts = name.split("/")
            if len(parts) > 0:
                chart_dir = os.path.join(temp_dir, parts[0])
                break

        # 获取ytt配置的目录
        ytt_config_dir = os.path.join(temp_dir, "ytt_config")
        if not os.path.exists(ytt_config_dir):
            os.makedirs(ytt_config_dir)

        # 获取注入信息的模板文件名
        tpl_name = get_injected_tpl(bcs_inject_data.cluster_id)

        inject_values_path = os.path.join(ytt_config_dir, tpl_name)
        with open(inject_values_path, "w") as f:
            f.write(render_to_string(tpl_name, asdict(bcs_inject_data)))
        ytt_sh_path = os.path.join(ytt_config_dir, YTT_RENDERER_NAME)
        with open(ytt_sh_path, "w") as f:
            # helm post renderer依赖执行命令
            ytt_sh_content = f"#!/bin/bash\n{settings.YTT_BIN} --ignore-unknown-comments -f - -f {ytt_config_dir}/"
            f.write(ytt_sh_content)
        # 确保文件可执行
        os.chmod(ytt_sh_path, stat.S_IRWXU)

        yield chart_dir, ytt_config_dir


def get_injected_tpl(cluster_id: str) -> str:
    """获取平台注入模板

    - 共享集群时，返回 shared_cluster_injected_tpl.yaml
    - 专用集群时，返回 single_cluster_injected_tpl.yaml
    """
    return f"{get_cluster_type(cluster_id).lower()}_cluster_injected_tpl.yaml"
