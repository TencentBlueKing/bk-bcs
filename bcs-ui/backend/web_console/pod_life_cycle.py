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
import base64
import json
import logging
import shlex
import time
from concurrent.futures import ThreadPoolExecutor
from typing import Optional

import yaml
from django.conf import settings
from django.template.loader import render_to_string
from django.utils.encoding import smart_str
from django.utils.translation import ugettext_lazy as _
from kubernetes import client
from kubernetes.client.rest import ApiException
from kubernetes.stream import stream
from tornado.concurrent import run_on_executor
from tornado.ioloop import PeriodicCallback

from backend.utils.cache import rd_client
from backend.utils.lock import redis_lock
from backend.web_console import constants

logger = logging.getLogger(__name__)


class PodLifeError(Exception):
    pass


class UserTokenNotFound(Exception):
    pass


class PodLifeCycle:
    executor = ThreadPoolExecutor()

    @classmethod
    def heartbeat(cls, name):
        """定时上报存活, 清理时需要使用"""
        logger.debug("heartbeat: %s", name)

        result = rd_client.zadd(constants.WEB_CONSOLE_HEARTBEAT_KEY, {name: time.time()})

        return result

    def get_active_user_pod(self):
        """获取存活节点"""
        now = time.time()
        start = now - constants.USER_POD_EXPIRE_TIME
        expired_pods = rd_client.zremrangebyscore(constants.WEB_CONSOLE_HEARTBEAT_KEY, "-inf", start)
        actived_pods = rd_client.zrange(constants.WEB_CONSOLE_HEARTBEAT_KEY, 0, -1, withscores=True)
        actived_pods = [(smart_str(i[0]), now - i[1]) for i in actived_pods]

        pods = [i[0] for i in actived_pods]

        idle_pods_msg = ", ".join("{pod[0]}={pod[1]:.2f}s".format(pod=pod) for pod in actived_pods)
        logger.info("remove expired_pods, %s, get actived pods: %s", expired_pods, idle_pods_msg)

        return pods

    @redis_lock("pod_life_cycle.clean_user_pod", constants.CLEAN_USER_POD_INTERVAL, shift=constants.LOCK_SHIFT)
    @run_on_executor
    def clean_user_pod(self):
        logger.debug("start clean user pod")

        try:
            self._clean_user_pod()
            logger.debug("clean user pod success")
        except Exception as error:
            logger.error("clean user pod error: %s", error)

    def _clean_user_pod(self):
        """单个集群清理"""
        alive_pods = self.get_active_user_pod()

        for v1, cluster_id in K8SClient.iter_client():
            try:
                pod_list = v1.list_namespaced_pod(namespace=constants.NAMESPACE)
                self._clean_user_pod_by_cluster(v1, pod_list, alive_pods)
            except Exception as error:
                logger.info("clean %s pod not success, %s", cluster_id, error)

    def _clean_user_pod_by_cluster(self, v1, pod_list, alive_pods):
        min_expire_time = time.time() - constants.USER_POD_EXPIRE_TIME

        for pod in pod_list.items:
            if pod.status.phase == "Pending":
                continue

            # 小于一个周期的pod不清理
            if pod.metadata.labels and pod.metadata.labels.get(constants.LABEL_WEB_CONSOLE_CREATE_TIMESTAMP):
                pod_create_time = int(pod.metadata.labels[constants.LABEL_WEB_CONSOLE_CREATE_TIMESTAMP])
            else:
                pod_create_time = None

            if pod_create_time and pod_create_time > min_expire_time:
                logger.info(
                    "pod %s exist time %s > %s, just ignore", pod.metadata.name, pod_create_time, min_expire_time
                )
                continue

            # 有心跳上报的pod不清理
            if pod.metadata.name in alive_pods:
                continue

            v1.delete_namespaced_pod(
                name=pod.metadata.name, namespace=constants.NAMESPACE, body=client.V1DeleteOptions()
            )
            logger.info("delete pod %s", pod.metadata.name)

            for volume in pod.spec.volumes:
                cm = getattr(volume, "config_map", None)
                if not cm:
                    continue

                cm_name = cm.name

                v1.delete_namespaced_config_map(
                    name=cm_name, namespace=constants.NAMESPACE, body=client.V1DeleteOptions()
                )
                logger.info("delete configmap %s", cm_name)

    def start(self):
        self.scheduler = PeriodicCallback(self.clean_user_pod, constants.CLEAN_USER_POD_INTERVAL * 1000)
        self.scheduler.start()


class K8SClient(object):
    CACHE_KEY_PREFIX = "K8S:USER_TOKEN"

    def __init__(self, ctx=None):
        self.ctx = ctx

    @classmethod
    def get_api_client(cls, server_address, identifier, user_token, ctx):
        aConfiguration = client.Configuration()
        aConfiguration.verify_ssl = False
        aConfiguration.host = server_address
        aConfiguration.api_key = {"authorization": f"Bearer {user_token}"}
        aApiClient = client.ApiClient(aConfiguration)
        logger.info("use %s client, %s, %s", ctx["mode"], aConfiguration.host, aConfiguration.api_key)
        return aApiClient

    @classmethod
    def iter_client(cls):
        for key in rd_client.scan_iter(f"{cls.CACHE_KEY_PREFIX}:*"):
            key = smart_str(key)
            data = rd_client.get(key)
            try:
                ctx = json.loads(data)
            except Exception as error:
                logger.info("get k8s context error, %s", error)
                continue

            k8s_client = cls.get_api_client(
                ctx["admin_server_address"], ctx["admin_cluster_identifier"], ctx["admin_user_token"], ctx
            )
            v1 = client.CoreV1Api(k8s_client)
            yield (v1, ctx["cluster_id"])

    def get_client(self, ctx):
        # 每次保存最新的user_token， 12小时过期
        source_cluster_id = ctx["source_cluster_id"].lower()
        data = {
            "admin_user_token": ctx["admin_user_token"],
            "admin_cluster_identifier": ctx["admin_cluster_identifier"],
            "cluster_id": source_cluster_id,
            "admin_server_address": ctx["admin_server_address"],
            "mode": ctx["mode"],
        }

        if ctx.get("should_cache_ctx") is True:
            name = f"{source_cluster_id}-u{ctx['username_slug']}"
            CACHE_KEY = f"{self.CACHE_KEY_PREFIX}:{name}"
            rd_client.set(CACHE_KEY, json.dumps(data), ex=constants.USER_CTX_EXPIRE_TIME)
        k8s_client = self.get_api_client(
            ctx["admin_server_address"], ctx["admin_cluster_identifier"], ctx["admin_user_token"], ctx
        )
        return k8s_client

    @property
    def v1(self):
        k8s_client = self.get_client(self.ctx)
        v1_client = client.CoreV1Api(k8s_client)
        return v1_client

    @property
    def rbac(self):
        k8s_client = self.get_client(self.ctx)
        rbac = client.RbacAuthorizationV1Api(k8s_client)
        return rbac

    def exec_command(self, command: str) -> str:
        """执行命令，返回输出结果"""
        command = shlex.split(command)
        resp = stream(
            self.v1.connect_get_namespaced_pod_exec,
            self.ctx["user_pod_name"],
            self.ctx["namespace"],
            command=command,
            container=self.ctx["container_name"],
            stderr=True,
            stdin=False,
            stdout=True,
            tty=False,
            _preload_content=True,
        )
        return resp


def wait_user_pod_ready(ctx, name):
    sleep_time = 1
    total_sleep = 0
    # 最多等待1分钟
    wait_timeout = 60
    error_wait_timeout = 7

    k8s_client = K8SClient(ctx)

    while total_sleep < wait_timeout:
        try:
            pod = k8s_client.v1.read_namespaced_pod(name, namespace=constants.NAMESPACE)
            if pod.status.phase == "Running":
                return pod
        except Exception as error:
            # 错误一次返回
            logger.error("get user pod name error: %s", error)
            # 异常情况最多等待7秒
            if total_sleep > error_wait_timeout:
                raise PodLifeError(_("申请pod资源失败，请稍后再试{}").format(settings.COMMON_EXCEPTION_MSG))

        time.sleep(sleep_time)
        total_sleep += sleep_time
        logger.info("wait pod ready, %s, sleep, %s, total_sleep, %s", name, sleep_time, total_sleep)

    raise PodLifeError(_("申请pod资源超时，请稍后再试{}").format(settings.COMMON_EXCEPTION_MSG))


def get_service_account_token(k8s_client) -> Optional[str]:
    """获取web-console token"""
    if settings.EDITION != settings.COMMUNITY_EDITION:
        return

    token_prefix = f"{constants.NAMESPACE}-token"
    for item in k8s_client.v1.list_namespaced_secret(constants.NAMESPACE).items:
        if not item.metadata.name.startswith(token_prefix):
            continue

        return smart_str(base64.b64decode(item.data["token"]))


def create_service_account_rbac(k8s_client, ctx):
    """创建serviceAccount, 绑定Role"""
    if settings.EDITION != settings.COMMUNITY_EDITION:
        return

    service_account_body = yaml.load(render_to_string("conf_tpl/service_account.yaml", ctx))
    service_account_rolebind_body = yaml.load(render_to_string("conf_tpl/service_account_rolebind.yaml", ctx))

    k8s_client.v1.create_namespaced_service_account(constants.NAMESPACE, service_account_body)
    k8s_client.rbac.create_cluster_role_binding(service_account_rolebind_body)


def ensure_namespace(ctx):
    """创建命名空间"""
    k8s_client = K8SClient(ctx)

    def _add_token_to_ctx(k8s_client, ctx):
        # 获取service_account_token
        try:
            ctx["service_account_token"] = get_service_account_token(k8s_client)
        except Exception as error:
            logger.error("get service_account_token error, %s", error)
            ctx["service_account_token"] = None

    try:
        ns = k8s_client.v1.read_namespace(name=constants.NAMESPACE)
        _add_token_to_ctx(k8s_client, ctx)
        return ns
    except ApiException as error:
        if error.status == 404:
            body = yaml.load(render_to_string("conf_tpl/namespace.yaml", ctx))
            try:
                ns = k8s_client.v1.create_namespace(body=body)
                create_service_account_rbac(k8s_client, ctx)
                _add_token_to_ctx(k8s_client, ctx)
                return ns
            except ApiException as error:
                logger.exception("create namespace error: %s", error)
                raise error


def ensure_configmap(ctx):
    """创建configmap"""
    name = "kube-config-%s-u%s" % (ctx["source_cluster_id"], ctx["username_slug"])
    name = name.lower()

    k8s_client = K8SClient(ctx)
    try:
        cm = k8s_client.v1.read_namespaced_config_map(name, namespace=constants.NAMESPACE)
        return cm
    except ApiException as error:
        # 不存在，则创建
        if error.status == 404:
            body = yaml.load(render_to_string("conf_tpl/configmap.yaml", ctx))
            try:
                cm = k8s_client.v1.create_namespaced_config_map(body=body, namespace=constants.NAMESPACE)
                return cm
            except ApiException as error:
                logger.exception("create config_map error: %s", error)
                raise error
        raise error


def ensure_pod(ctx):
    """创建configmap"""
    name = "kubectld-%s-u%s" % (ctx["source_cluster_id"], ctx["username_slug"])
    name = name.lower()

    k8s_client = K8SClient(ctx)

    try:
        pod = k8s_client.v1.read_namespaced_pod(name, namespace=constants.NAMESPACE)
        if pod.status.phase != "Running":
            raise PodLifeError(_("pod不是Running状态，请稍后再试{}").format(settings.COMMON_EXCEPTION_MSG))
        return pod
    except ApiException as error:
        # 不存在，则创建
        if error.status == 404:
            # 添加时间戳
            ctx["LABEL_WEB_CONSOLE_CREATE_TIMESTAMP"] = constants.LABEL_WEB_CONSOLE_CREATE_TIMESTAMP
            ctx["create_timestamp"] = int(time.time())
            body = yaml.load(render_to_string("conf_tpl/pod.yaml", ctx))
            # 添加环境特有变量
            body["spec"].update(ctx["pod_spec"])
            try:
                k8s_client.v1.create_namespaced_pod(body=body, namespace=constants.NAMESPACE)
                pod = wait_user_pod_ready(ctx, name)
                if pod:
                    pod_life_cycle = PodLifeCycle()
                    pod_life_cycle.heartbeat(name)
                    return pod
                raise PodLifeError(_("申请pod失败或不是Running状态，请稍后再试{}").format(settings.COMMON_EXCEPTION_MSG))
            except ApiException as error:
                raise error
        raise error
