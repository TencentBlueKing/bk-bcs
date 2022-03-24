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
import logging
import math

from backend.templatesets.legacy_apps.instance.constants import EventType, InsState
from backend.templatesets.legacy_apps.instance.models import (
    InstanceConfig,
    InstanceEvent,
    MetricConfig,
    VersionInstance,
)
from backend.utils.exceptions import ConfigError, Rollback

logger = logging.getLogger(__name__)


class ClusterNotReady(Exception):
    pass


class BCSRollback(Rollback):
    pass


class SchedulerBase(object):
    INIT_ORDERING = {
        "secret": 1,
        "configmap": 2,
        "service": 3,
        "metric": 4,
        "ingress": 5,
        "deployment": 10,
        "application": 11,
        # k8s 相关资源
        "K8sIngress": 100,
        "K8sSecret": 101,
        "K8sConfigMap": 102,
        "K8sService": 103,
        "K8sDaemonSet": 104,
        "K8sJob": 105,
        "K8sStatefulSet": 106,
        "K8sDeployment": 107,
    }

    def __init__(self, access_token, project_id, configuration, kind, is_rollback):
        self.access_token = access_token
        self.project_id = project_id
        self.configuration = configuration
        self.plugin_client = SchedulerPluginCC(access_token, project_id)
        self.rollback_stack = {}
        self.kind = kind
        # 所有的操作都不回滚 2018-08-22
        self.is_rollback = False

    def instantiation_ns(self, ns_id, config, is_update):
        """单个命名空间实例化"""
        # 创建必须按顺序执行
        config = sorted(config.items(), key=lambda x: self.INIT_ORDERING.get(x[0], math.inf))
        for res, specs in config:
            if is_update:
                handler = getattr(self, "handler_update_%s" % res.lower(), None)
            else:
                handler = getattr(self, "handler_%s" % res.lower(), None)
            # plugin_handler = getattr(self.plugin_client, 'handler_%s' % res, None)

            if not handler:
                raise NotImplementedError("%s not have handler" % res)

            # if not plugin_handler:
            #    raise NotImplementedError('plugin %s not have handler' % res)

            for spec in specs:
                # 只获取需要使用的字段
                cluster_id = spec["context"]["SYS_CLUSTER_ID"]
                ns = spec["context"]["SYS_NAMESPACE"]
                self.rollback_stack.setdefault(ns_id, [])
                # application deployment 最后一步，失败没有创建成功，不需要回滚
                if res not in ["application", "deployment"]:
                    self.rollback_stack[ns_id].append([res, ns, cluster_id, spec["config"]])

                # 获取状态信息
                if res == "metric":
                    queryset = MetricConfig.objects.filter(pk=spec["instance_config_id"])
                else:
                    queryset = InstanceConfig.objects.filter(pk=spec["instance_config_id"])
                # 需要更新的参数
                is_update_save_kwargs = spec.get("is_update_save_kwargs", False)
                if is_update_save_kwargs:
                    save_kwargs = spec.get("save_kwargs", {})
                else:
                    save_kwargs = {}

                # ref = queryset.first()
                # # 已经成功的，且不是更新操作, 不需要再下发
                # if ref.ins_state == InsState.INS_SUCCESS.value and not is_update:
                #     continue
                try:
                    handler(ns, cluster_id, spec["config"])
                    if is_update:
                        ins_state = InsState.UPDATE_SUCCESS.value
                        # queryset.update(ins_state=InsState.UPDATE_SUCCESS.value, is_bcs_success=True)
                    else:
                        ins_state = InsState.INS_SUCCESS.value
                        # queryset.update(ins_state=InsState.INS_SUCCESS.value, is_bcs_success=True)
                    save_kwargs["ins_state"] = ins_state
                    save_kwargs["is_bcs_success"] = True
                    queryset.update(**save_kwargs)

                except Rollback as error:
                    # 捕获错误消息
                    result = error.args[0]
                    InstanceEvent.log(
                        spec["instance_config_id"], res, EventType.REQ_FAILED.value, result, spec["context"]
                    )

                    # 修改资源对应状态
                    if is_update:
                        ins_state = InsState.UPDATE_FAILED.value
                        # queryset.update(ins_state=InsState.UPDATE_FAILED.value, is_bcs_success=False)
                    else:
                        ins_state = InsState.INS_FAILED.value
                        # queryset.update(ins_state=InsState.INS_FAILED.value, is_bcs_success=False)
                        save_kwargs["ins_state"] = ins_state
                        save_kwargs["is_bcs_success"] = False
                        queryset.update(**save_kwargs)
                    # 需要抛出异常到上层进行回滚, 错误消息需要提示到前台
                    result["res_type"] = res
                    raise BCSRollback(result)

    def instantiation(self, is_update=False):
        """实例化"""
        instantiation_result = {"success": [], "failed": []}
        for ns_id, config in self.configuration.items():
            cluster_id = [i for i in config.values()][0][0]["context"]["SYS_CLUSTER_ID"]
            ns_name = [i for i in config.values()][0][0]["context"]["SYS_NAMESPACE"]

        for ns_id, config in self.configuration.items():
            instance_id = [i for i in config.values()][0][0]["context"]["SYS_INSTANCE_ID"]
            ns_name = [i for i in config.values()][0][0]["context"]["SYS_NAMESPACE"]
            ns = {"ns_id": ns_id, "ns_name": ns_name, "instance_id": instance_id, "res_type": "", "err_msg": ""}
            bcs_success = True
            try:
                self.instantiation_ns(ns_id, config, is_update)
            except Rollback as error:
                if self.is_rollback and (not is_update):
                    self.handler_rollback(ns_id)
                ns["res_type"] = error.args[0]["res_type"]
                ns["err_msg"] = error.args[0].get("message", "")
                bcs_success = False
                logger.warning("bcs_api: error, %s, add failed to result %s", ns, instantiation_result)
            except ConfigError as error:
                if self.is_rollback and (not is_update):
                    self.handler_rollback(ns_id)
                bcs_success = False
                ns["err_msg"] = str(error)
                ns["show_err_msg"] = True
                logger.exception("bcs_api: %s, instantiation error, %s", ns, error)
                logger.warning("bcs_api: exception, %s, add failed to result %s", ns, instantiation_result)
            except Exception as error:
                if self.is_rollback and (not is_update):
                    self.handler_rollback(ns_id)
                bcs_success = False
                ns["err_msg"] = str(error)
                logger.exception("bcs_api: %s, instantiation error, %s", ns, error)
                logger.warning("bcs_api: exception, %s, add failed to result %s", ns, instantiation_result)

            # 统一修改状态
            try:
                VersionInstance.objects.filter(pk=instance_id).update(is_bcs_success=bcs_success)
                # InstanceConfig.objects.filter(instance_id=instance_id).update(
                #     is_bcs_success=bcs_success)
                # MetricConfig.objects.filter(instance_id=instance_id).update(
                #     is_bcs_success=bcs_success)
            except Exception:
                logging.exception("save is_bcs_success error")

            if bcs_success is False:
                instantiation_result["failed"].append(ns)
            else:
                instantiation_result["success"].append(ns)
            logger.info("bcs_api: instantiation_result, %s", instantiation_result)

        return instantiation_result

    def handler_rollback(self, ns_id):
        """"""
        roll_back_list = self.rollback_stack[ns_id]
        roll_back_list = roll_back_list[:-1]
        for s in roll_back_list:
            handler = getattr(self, "rollback_%s" % s[0].lower(), None)
            if handler:
                logger.warning("try to rollback, %s, %s", s, s[1:])
                handler(*s[1:])
            else:
                logging.warning("have not rollback handler, %s, %s, will ignore", s, s[1:])


class SchedulerPlugin(object):
    def __init__(self, access_token, project_id):
        self.access_token = access_token
        self.project_id = project_id


class SchedulerPluginCC(SchedulerPlugin):
    def handler_application(self, ns, cluster_id, spec):
        pass

    def handler_deployment(self, ns, cluster_id, spec):
        pass

    def handler_service(self, ns, cluster_id, spec):
        pass

    def handler_lb(self, ns, cluster_id, spec):
        """负载均衡"""

    def handler_configmap(self, ns, cluster_id, spec):
        pass

    def handler_secret(self, ns, cluster_id, spec):
        pass
