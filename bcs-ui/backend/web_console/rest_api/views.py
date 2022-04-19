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
from urllib.parse import urlparse

from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from rest_framework import views
from rest_framework.renderers import BrowsableAPIRenderer

from backend.accounts import bcs_perm
from backend.components import paas_auth, paas_cc
from backend.components.bcs.k8s import K8SClient
from backend.container_service.clusters.base.utils import get_cluster
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.funutils import remove_url_domain
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import BKAPIResponse
from backend.web_console import constants, pod_life_cycle
from backend.web_console.bcs_client import k8s
from backend.web_console.utils import get_kubectld_version

from ..session import session_mgr
from . import utils
from .serializers import K8SWebConsoleOpenSLZ, K8SWebConsoleSLZ

logger = logging.getLogger(__name__)


class WebConsoleSession(views.APIView):
    # 缓存的key
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_k8s_container_context(self, request, project_id, cluster_id, client, bcs_context):
        """获取容器上下文"""
        slz = K8SWebConsoleSLZ(data=request.query_params, context={"client": client})
        slz.is_valid(raise_exception=True)

        bcs_context["mode"] = k8s.ContainerDirectClient.MODE
        bcs_context["user_pod_name"] = slz.validated_data["pod_name"]
        bcs_context.update(slz.validated_data)
        return bcs_context

    def get_k8s_cluster_context(self, request, project_id, cluster_id, client, bcs_context):
        """获取集群模式(kubectl)上下文"""
        # kubectl版本区别
        kubectld_version = get_kubectld_version(client.version)

        bcs_context = utils.get_k8s_admin_context(client, bcs_context, settings.WEB_CONSOLE_MODE)

        ctx = {
            "username": self.request.user.username,
            "settings": settings,
            "kubectld_version": kubectld_version,
            "namespace": constants.NAMESPACE,
            "pod_spec": utils.get_k8s_pod_spec(client),
            "username_slug": utils.get_username_slug(self.request.user.username),
            # 缓存ctx， 清理使用
            "should_cache_ctx": True,
        }
        ctx.update(bcs_context)
        try:
            pod_life_cycle.ensure_namespace(ctx)
            configmap = pod_life_cycle.ensure_configmap(ctx)
            logger.debug("get configmap %s", configmap)
            pod = pod_life_cycle.ensure_pod(ctx)
            logger.debug("get pod %s", pod)
        except pod_life_cycle.PodLifeError as error:
            logger.error("apply error: %s", error)
            utils.activity_log(project_id, cluster_id, self.cluster_name, request.user.username, False, "%s" % error)
            raise error_codes.APIError("%s" % error)
        except Exception as error:
            logger.exception("apply error: %s", error)
            utils.activity_log(project_id, cluster_id, self.cluster_name, request.user.username, False, "申请pod资源失败")
            raise error_codes.APIError(_("申请pod资源失败，请稍后再试{}").format(settings.COMMON_EXCEPTION_MSG))

        bcs_context["user_pod_name"] = pod.metadata.name

        return bcs_context

    def get_k8s_context(self, request, project_id, cluster_id):
        """获取k8s的上下文信息"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        try:
            bcs_context = utils.get_k8s_cluster_context(client, project_id, cluster_id)
        except Exception as error:
            logger.exception("get access cluster context failed: %s", error)
            message = _("获取集群{}【{}】WebConsole 信息失败").format(self.cluster_name, cluster_id)
            # 记录操作日志
            utils.activity_log(project_id, cluster_id, self.cluster_name, request.user.username, False, message)
            # 返回前端消息
            raise error_codes.APIError(
                _("{}，请检查 Deployment【kube-system/bcs-agent】是否正常{}").format(message, settings.COMMON_EXCEPTION_MSG)
            )

        if request.GET.get("container_id"):
            bcs_context = self.get_k8s_container_context(request, project_id, cluster_id, client, bcs_context)
        else:
            bcs_context = self.get_k8s_cluster_context(request, project_id, cluster_id, client, bcs_context)

        return bcs_context

    def get(self, request, project_id, cluster_id):
        """获取session信息"""
        cluster_data = get_cluster(request.user.token.access_token, project_id, cluster_id)
        self.cluster_name = cluster_data.get("name", "")[:32]

        # 检查白名单, 不在名单中再通过权限中心校验
        if not request.user.is_superuser:
            perm = bcs_perm.Cluster(request, project_id, cluster_id)
            try:
                perm.can_use(raise_exception=True)
            except Exception as error:
                utils.activity_log(
                    project_id, cluster_id, self.cluster_name, request.user.username, False, _("集群不正确或没有集群使用权限")
                )
                raise error

        context = self.get_k8s_context(request, project_id, cluster_id)

        context["username"] = request.user.username
        context.setdefault("namespace", constants.NAMESPACE)
        logger.info(context)

        session = session_mgr.create(project_id, cluster_id)
        session_id = session.set(context)

        # 替换http为ws地址
        bcs_api_url = urlparse(settings.DEVOPS_BCS_API_URL)
        if bcs_api_url.scheme == "https":
            scheme = "wss"
        else:
            scheme = "ws"
        bcs_api_url = bcs_api_url._replace(scheme=scheme)
        # 连接ws的来源方式, 有容器直连(direct)和多tab管理(mgr)
        source = request.query_params.get("source", "direct")

        ws_url = f"{bcs_api_url.geturl()}/web_console/projects/{project_id}/clusters/{cluster_id}/ws/?session_id={session_id}&source={source}"  # noqa

        data = {"session_id": session_id, "ws_url": remove_url_domain(ws_url)}
        utils.activity_log(project_id, cluster_id, self.cluster_name, request.user.username, True)

        return BKAPIResponse(data, message=_("获取session成功"))


class OpenSession(views.APIView):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    permission_classes = ()

    def get(self, request):
        """校验session_id，同时换取ws_url"""
        session_id = request.GET.get("session_id")
        if not session_id:
            raise error_codes.APIError(_("session_id不能为空"))

        session = session_mgr.create("", "")
        context = session.get(session_id)
        if not context:
            raise error_codes.APIError(_("session_id不合法或已经过期"))

        if context.get("operator") != request.user.username:
            raise error_codes.APIError(_("不是合法用户"))

        ws_session = session_mgr.create(context["project_id"], context["cluster_id"])
        ws_session_id = ws_session.set(context)

        bcs_api_url = urlparse(settings.DEVOPS_BCS_API_URL)
        if bcs_api_url.scheme == "https":
            scheme = "wss"
        else:
            scheme = "ws"
        bcs_api_url = bcs_api_url._replace(scheme=scheme)

        ws_url = f"{bcs_api_url.geturl()}/web_console/projects/{context['project_id']}/clusters/{context['cluster_id']}/ws/?session_id={ws_session_id}&source=direct"  # noqa

        data = {"session_id": session_id, "ws_url": remove_url_domain(ws_url)}
        return BKAPIResponse(data, message=_("获取ws session成功"))


class CreateOpenSession(views.APIView):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_k8s_context(self, request, project_id_or_code, cluster_id):
        """获取docker监控信息"""
        access_token = paas_auth.get_access_token().get("access_token")

        result = paas_cc.get_project(access_token, project_id_or_code)

        if result.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(_("项目Code或者ID不正确: {}").format(result.get("message", "")))

        project_id = result["data"]["project_id"]

        client = K8SClient(access_token, project_id, cluster_id, None)
        slz = K8SWebConsoleOpenSLZ(data=request.data, context={"client": client})
        slz.is_valid(raise_exception=True)

        try:
            bcs_context = utils.get_k8s_cluster_context(client, project_id, cluster_id)
        except Exception as error:
            logger.exception("get access cluster context failed: %s", error)
            message = _("获取集群{}【{}】WebConsole session 信息失败").format(cluster_id, cluster_id)
            # 返回前端消息
            raise error_codes.APIError(message)

        bcs_context["mode"] = k8s.ContainerDirectClient.MODE
        bcs_context["user_pod_name"] = slz.validated_data["pod_name"]
        bcs_context["project_id"] = project_id
        bcs_context.update(slz.validated_data)

        return bcs_context

    def post(self, request, project_id_or_code, cluster_id):
        """创建session_id"""
        context = self.get_k8s_context(request, project_id_or_code, cluster_id)

        context["username"] = context.get("operator", "")
        context.setdefault("namespace", constants.NAMESPACE)

        session = session_mgr.create("", "")
        context["project_id_or_code"] = project_id_or_code
        context["cluster_id"] = cluster_id
        session_id = session.set(context)
        container_name = context.get("container_name", "")

        web_console_url = (
            f"{settings.DEVOPS_BCS_API_URL}/web_console/?session_id={session_id}&container_name={container_name}"
        )

        data = {
            "session_id": session_id,
            "web_console_url": web_console_url,
        }

        return BKAPIResponse(data, message=_("创建session成功"))
