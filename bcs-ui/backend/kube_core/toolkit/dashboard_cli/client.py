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
import os
import shutil
import subprocess
import tempfile
import time
from pathlib import Path

from .exceptions import DashboardError, DashboardExecutionError

logger = logging.getLogger(__name__)


class DashboardClient:
    """
    call kubernetes dashboard by command
    """

    def __init__(self, dashboard_ctl_bin="dashboard-ctl", kubeconfig=""):
        self.dashboard_ctl_bin = dashboard_ctl_bin
        self.kubeconfig = kubeconfig or Path.home() / ".kube/config"

    def workload_status(self, kind: str, name: str, namespace: str, parameters: dict, api_log_level=3):
        """
        /api/v1/deployment/default/rumpetroll-openresty
        dashboard-ctl --kubeconfig {kubeconfig} --api-log-level {api_log_level} \
        --output-file output.json --url-path /api/v1/{kind}/{namespace}/{name}
        """
        namespace = namespace.lower()
        name = name.lower()
        kind = kind.lower()

        url_path = "/api/v1/{kind}/{namespace}/{name}".format(namespace=namespace, name=name, kind=kind)
        return self.run(
            namespace=namespace,
            url_path=url_path,
            parameters=parameters,
            api_log_level=api_log_level,
        )

    def overview(self, namespace: str, parameters: dict, api_log_level=3):
        """
        dashboard-ctl --kubeconfig {kubeconfig} --api-log-level {api_log_level} \
        --output-file output.json --url-path /api/v1/overview/{namespace}
        """
        url_path = "/api/v1/overview/{namespace}".format(namespace=namespace)
        return self.run(
            namespace=namespace,
            url_path=url_path,
            parameters=parameters,
            api_log_level=api_log_level,
        )

    def run(self, namespace: str, parameters: dict, url_path: str, api_log_level=3):
        """
        dashboard-ctl --kubeconfig {kubeconfig} --api-log-level {api_log_level} \
        --output-file output.json --url-path /api/v1/overview/{namespace}
        """

        temp_dir = tempfile.mkdtemp()
        try:
            output_file = os.path.join(temp_dir, "output.json")

            # construct cmd and run
            cmd_args = [
                self.dashboard_ctl_bin,
                "--kubeconfig",
                self.kubeconfig,
                "--output-file",
                output_file,
                "--url-path",
                url_path,
                "--api-log-level",
                str(api_log_level),
            ]
            if parameters:
                import base64

                b64parameters = base64.standard_b64encode(json.dumps(parameters))
                cmd_args += ["--parameters-base64-encode", b64parameters]

            template_out, _ = self._run_command_with_retry(max_retries=0, cmd_args=cmd_args)
            with open(output_file, "r") as f:
                output_content = f.read()
                output = json.loads(output_content)
        except (DashboardExecutionError, DashboardError):
            raise
        except Exception as e:
            logger.exception(
                ("run dashboard ctl fail: namespace={namespace}, " "parameters={parameters}, error:{error}").format(
                    namespace=namespace, parameters=parameters, error=e
                )
            )
            raise
        finally:
            shutil.rmtree(temp_dir)

        return output

    def _run_command_with_retry(self, max_retries=3, *args, **kwargs):
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
            else:
                break

        raise ValueError(max_retries)

    def _run_command(self, cmd_args):
        """Run the dashboard ctl command with wrapped exceptions"""
        try:
            logger.info("Calling dashboard ctl cmd, cmd: (%s)", " ".join(cmd_args))

            output = subprocess.check_output(cmd_args)

        except subprocess.CalledProcessError as e:
            logger.info("Calling dashboard ctl cmd result: output: %s\n" % e.output.decode("utf-8"))
            logger.exception(
                "Unable to run dashboard ctl command, return code: %s, output: %s", e.returncode, e.output
            )
            raise DashboardExecutionError(e.returncode, e.output)
        except Exception as err:
            logger.exception("Unable to run dashboard ctl command, %s", err)
            raise DashboardError("run dashboard ctl command failed: {}".format(err))
        else:
            return output, ""
