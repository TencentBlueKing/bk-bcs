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

Operating kubernetes with kubectl command
"""
import json
import logging
import subprocess
import tempfile
import time
from pathlib import Path

from .exceptions import KubectlError, KubectlExecutionError

logger = logging.getLogger(__name__)

KUBECONFIG_ENV = "KUBECONFIG"


class KubectlClusterClient:
    """deploy and undeploy resources to kubernetes cluster by using kubectl binary tool,
    retry on any error

    :param kubectl_bin: location of kubectl binary
    :param kubeconfig: location of kubeconfig file, if not given, use "~/.kube/config" as default
    :param options: kubectl options, like server address
    """

    def __init__(self, kubectl_bin="kubectl", kubeconfig="", **options):
        self.kubectl_bin = kubectl_bin
        self.kubeconfig = kubeconfig or Path.home() / ".kube/config"
        self.options = options

    def parse_options(self):
        options = []
        for key, value in self.options.items():
            if value:
                options.append("--{key}={value}".format(key=key, value=value))
            else:
                options.append("--{key}".format(key=key))

        return options

    @property
    def kubectl_base(self):
        return [self.kubectl_bin] + self.parse_options()

    def apply(self, template: str, namespace="default"):
        """Apply resources in kubernetes cluster

        :returns: None if succeed
        :raises: KubectlError or KubectlExecutionError
        """
        # --force",
        with tempfile.NamedTemporaryFile() as fp:
            cmd_arguments = self.kubectl_base + [
                "--namespace=%s" % namespace,
                "apply",
                "--overwrite",
                "--filename",
                fp.name,
            ]
            fp.write(template.encode())
            fp.flush()
            self._run_command_with_retry(max_retries=0, cmd_arguments=cmd_arguments)

    def delete(self, template: str, namespace="default"):
        """Delete resources in kubernetes cluster

        :returns: None if succeed
        :raises: KubectlError or KubectlExecutionError
        """
        with tempfile.NamedTemporaryFile() as fp:
            cmd_arguments = self.kubectl_base + [
                "--namespace=%s" % namespace,
                "delete",
                "--ignore-not-found=true",
                "--filename",
                fp.name,
            ]
            fp.write(template.encode())
            fp.flush()
            self._run_command_with_retry(max_retries=0, cmd_arguments=cmd_arguments)

    def delete_one_by_one(self, items, namespace="default"):
        """Delete resources in kubernetes cluster

        :returns: None if succeed
        :raises: KubectlError or KubectlExecutionError
        """
        for item in items:
            # 针对chart中以`-`开头的资源名称，进行忽略，防止出现删除不掉的情况
            # TODO: 如果后续有其它场景，再以正则进行完整校验
            if item["name"].startswith("-"):
                continue
            cmd_arguments = self.kubectl_base + [
                "--namespace=%s" % namespace,
                "delete",
                "--ignore-not-found=true",
                item["kind"],
                item["name"],
            ]
            try:
                self._run_command_with_retry(max_retries=0, cmd_arguments=cmd_arguments)
            except KubectlExecutionError as e:
                if "the server doesn't have a resource type " in e.output:
                    pass
                else:
                    raise

    def get(self, kind: str, name: str, namespace="default"):
        """get resources {name} of kind {kind} under namespace {namespace} in kubernetes cluster

        :returns: dict
        :raises: KubectlError or KubectlExecutionError
        """
        cmd_arguments = self.kubectl_base + ["get", kind.lower(), name, "--namespace=%s" % namespace, "-o", "json"]
        output = self._run_command_with_retry(max_retries=0, cmd_arguments=cmd_arguments)
        return json.loads(output)

    def get_by_file(self, filename, namespace="default"):
        cmd_arguments = self.kubectl_base + [
            "get",
            "-f",
            filename,
            "--ignore-not-found=true",
            "--namespace=%s" % namespace,
            "-o",
            "json",
        ]
        output = self._run_command_with_retry(max_retries=0, cmd_arguments=cmd_arguments)
        return json.loads(output)

    def ensure_namespace(self, namespace):
        if namespace == "default":
            return

        get_cmd_arguments = self.kubectl_base + [
            "get",
            "namespace",
            namespace,
        ]
        create_cmd_arguments = self.kubectl_base + [
            "create",
            "namespace",
            namespace,
        ]
        try:
            return self._run_command_with_retry(max_retries=0, cmd_arguments=get_cmd_arguments)
        except Exception:
            pass

        try:
            return self._run_command_with_retry(max_retries=0, cmd_arguments=create_cmd_arguments)
        except Exception as e:
            if "AlreadyExists" in str(e):
                pass
            else:
                raise

    def _run_command_with_retry(self, max_retries=1, *args, **kwargs):
        for i in range(max_retries + 1):
            try:
                return self._run_command(*args, **kwargs)
            except Exception:
                if i == max_retries:
                    raise

                # retry after 0.5, 1, 1.5, ... seconds
                time.sleep((i + 1) * 0.5)
                continue

        raise ValueError(max_retries)

    def _run_command(self, cmd_arguments):
        """Run the kubectl command with wrapped exceptions"""
        cmd_str = " ".join(cmd_arguments)
        logger.info("Calling kubectl cmd, cmd: (%s)", cmd_str)
        try:
            output = subprocess.check_output(
                cmd_arguments, stderr=subprocess.STDOUT, env={KUBECONFIG_ENV: self.kubeconfig}
            )
        except subprocess.CalledProcessError as err:
            logger.exception(
                "Unable to run kubectl command, return code: %s\ncommand output: \n%s", err.returncode, err.output
            )
            raise KubectlExecutionError(err.returncode, err.output)
        except Exception as err:
            logger.exception("Unable to run kubectl command")
            raise KubectlError("run kubectl command failed: \n{}".format(err))
        else:
            return output
