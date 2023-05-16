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
import logging
import os
import tempfile
from datetime import datetime, timedelta

from django.conf import settings
from django.core.cache import cache
from django.db import IntegrityError
from django.template.loader import render_to_string
from django.utils.translation import ugettext_lazy as _
from jinja2 import Template
from rest_framework import viewsets
from rest_framework.exceptions import APIException, ValidationError
from rest_framework.permissions import IsAuthenticated
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.components import bk_repo, paas_cc
from backend.components.base import ComponentAuth
from backend.components.paas_cc import PaaSCCClient
from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.misc.bke_client.client import BCSClusterCredentialsNotFound, BCSClusterNotFound
from backend.helm.app.repo import get_or_create_private_repo
from backend.helm.app.serializers import FilterNamespacesSLZ, ReleaseParamsSLZ
from backend.helm.app.utils import get_helm_dashboard_path
from backend.helm.authtoken.authentication import TokenAuthentication
from backend.helm.helm.constants import RELEASE_VERSION_PREFIX
from backend.helm.helm.models.chart import ChartVersion
from backend.helm.helm.serializers import ChartVersionDetailSLZ, ChartVersionSLZ
from backend.helm.permissions import check_cluster_perm
from backend.helm.releases.utils.release_secret import RecordReleases
from backend.helm.toolkit import chart_versions
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.namespace import NamespaceRequest, calc_iam_ns_id
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedAction, NamespaceScopedPermission
from backend.kube_core.toolkit.dashboard_cli.exceptions import DashboardError, DashboardExecutionError
from backend.resources.namespace.constants import K8S_PLAT_NAMESPACE
from backend.utils import client as bcs_utils_client
from backend.utils.errcodes import ErrorCode
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse
from backend.utils.views import AccessTokenMixin, ActionSerializerMixin, AppMixin, ProjectMixin, with_code_wrapper

from .models import App
from .serializers import (
    AppCreatePreviewSLZ,
    AppDetailSLZ,
    AppPreviewSLZ,
    AppReleaseDiffSLZ,
    AppReleasePreviewSLZ,
    AppRollbackPreviewSLZ,
    AppRollbackSelectionsSLZ,
    AppRollbackSLZ,
    AppSLZ,
    AppStateSLZ,
    AppUpgradeByAPISLZ,
    AppUpgradeSLZ,
    AppUpgradeVersionsSLZ,
    NamespaceSLZ,
    ReleaseListSLZ,
    SyncDict2YamlToolSLZ,
    SyncYaml2DictToolSLZ,
)
from .utils import collect_resource_state, collect_resource_status

logger = logging.getLogger(__name__)

# Helm 执行超时时间，设置为600s
HELM_TASK_TIMEOUT = 600


class AppViewBase(AccessTokenMixin, ProjectMixin, viewsets.ModelViewSet):
    queryset = App.objects.all()
    lookup_url_kwarg = "app_id"

    def get_queryset(self):
        queryset = super(AppViewBase, self).get_queryset().filter(project_id=self.project_id).order_by("-updated")
        return queryset.exclude()


@with_code_wrapper
class AppView(ActionSerializerMixin, AppViewBase):
    serializer_class = AppSLZ

    action_serializers = {
        'update': AppUpgradeSLZ,
        'retrieve': AppDetailSLZ,
    }

    def get_project_cluster(self, request, project_id):
        """获取项目下集群信息"""
        project_cluster = paas_cc.get_all_clusters(request.user.token.access_token, project_id, desire_all_data=True)
        if project_cluster.get('code') != ErrorCode.NoError:
            logger.error('Request cluster info error, detail: %s' % project_cluster.get('message'))
            return {}
        data = project_cluster.get('data') or {}
        results = data.get('results') or []
        return {info['cluster_id']: info for info in results} if results else {}

    def _is_transition_timeout(self, updated: str, transitioning_on: bool) -> bool:
        """判断app的transition是否超时"""
        updated_time = datetime.strptime(updated, settings.REST_FRAMEWORK["DATETIME_FORMAT"])
        if transitioning_on and (datetime.now() - updated_time) > timedelta(seconds=HELM_TASK_TIMEOUT):
            return True
        return False

    @response_perms(
        action_ids=[NamespaceScopedAction.VIEW, NamespaceScopedAction.UPDATE, NamespaceScopedAction.DELETE],
        permission_cls=NamespaceScopedPermission,
        resource_id_key='iam_ns_id',
    )
    def list(self, request, project_id, *args, **kwargs):
        """"""
        project_cluster = self.get_project_cluster(request, project_id)
        qs = self.get_queryset()
        # 获取过滤参数
        params = request.query_params
        # 集群和命名空间必须传递
        cluster_id = params.get('cluster_id')
        namespace = params.get("namespace")
        # TODO: 先写入db中，防止前端通过ID，获取数据失败；后续通过helm服务提供API
        if cluster_id and namespace:
            try:
                ctx_cluster = CtxCluster.create(
                    id=cluster_id, token=request.user.token.access_token, project_id=project_id
                )
                RecordReleases(ctx_cluster, namespace).record()
            except Exception as e:
                logger.error("获取集群内release数据失败，%s", e)

        if cluster_id:
            qs = qs.filter(cluster_id=cluster_id)
        if namespace:
            if not cluster_id:
                raise ValidationError(_("命名空间作为过滤参数时，需要提供集群ID"))
            qs = qs.filter(namespace=namespace)
        # 获取返回的数据
        slz = ReleaseListSLZ(qs, many=True)
        data = slz.data

        # do fix on the data which version is emtpy
        iam_ns_ids = []
        app_list = []
        for item in data:
            # 过滤掉k8s系统和bcs平台命名空间下的release
            if item["namespace"] in K8S_PLAT_NAMESPACE:
                continue
            cluster_info = project_cluster.get(item['cluster_id']) or {'name': item['cluster_id']}
            item['cluster_name'] = cluster_info['name']

            item['iam_ns_id'] = calc_iam_ns_id(item['cluster_id'], item['namespace'])
            iam_ns_ids.append({'iam_ns_id': item['iam_ns_id']})

            item['cluster_env'] = settings.CLUSTER_ENV_FOR_FRONT.get(cluster_info.get('environment'))
            item["current_version"] = item.pop("version")
            if not item["current_version"]:
                version = App.objects.filter(id=item["id"]).values_list(
                    "release__chartVersionSnapshot__version", flat=True
                )[0]
                App.objects.filter(id=item["id"]).update(version=version)
                item["current_version"] = version

            # 判断任务超时，并更新字段
            if self._is_transition_timeout(item["updated"], item["transitioning_on"]):
                err_msg = _("Helm操作超时，请重试!")
                App.objects.filter(id=item["id"]).update(
                    transitioning_on=False,
                    transitioning_result=False,
                    transitioning_message=err_msg,
                )
                item["transitioning_result"] = False
                item["transitioning_on"] = False
                item["transitioning_message"] = err_msg

            app_list.append(item)

        result = {"count": len(app_list), "next": None, "previous": None, "results": app_list}
        try:
            ns_request = NamespaceRequest(project_id=project_id, cluster_id=cluster_id)
        except TypeError:
            return Response(result)
        else:
            return PermsResponse(
                data=result,
                resource_request=ns_request,
                resource_data=iam_ns_ids,
            )

    def retrieve(self, request, *args, **kwargs):
        app_id = self.request.parser_context["kwargs"]["app_id"]
        if not App.objects.filter(id=app_id).exists():
            return Response({"code": 404, "detail": "app not found"})
        return super(AppView, self).retrieve(request, *args, **kwargs)

    def create(self, request, *args, **kwargs):
        try:
            return super(AppView, self).create(request, *args, **kwargs)
        except IntegrityError as e:
            logger.warning("helm app create IntegrityError, %s", e)
            return Response(
                status=400,
                data={
                    "code": 400,
                    "message": "helm app name already exists in this cluster",
                },
            )
        except BCSClusterNotFound:
            return Response(
                data={
                    "code": 40031,
                    "message": _("集群未注册"),
                }
            )
        except BCSClusterCredentialsNotFound:
            return Response(
                data={
                    "code": 40031,
                    "message": _("集群证书未上报"),
                }
            )

    def destroy(self, request, *args, **kwargs):
        """重载默认的 destroy 方法，用于实现判断是否删除成功"""
        instance = self.get_object()

        check_cluster_perm(
            user=request.user, project_id=instance.project_id, cluster_id=instance.cluster_id, request=request
        )

        self.perform_destroy(instance)

        if App.objects.filter(id=instance.id).exists():
            instance = App.objects.get(id=instance.id)
            data = {
                "transitioning_result": instance.transitioning_result,
                "transitioning_message": instance.transitioning_message,
            }
        else:
            data = {"transitioning_result": True, "transitioning_message": "success deleted"}
        return Response(data)

    def perform_destroy(self, instance):
        instance.destroy(username=self.request.user.username, access_token=self.access_token)


@with_code_wrapper
class AppRollbackView(AppViewBase):
    serializer_class = AppRollbackSLZ


class AppNamespaceView(AccessTokenMixin, ProjectMixin, viewsets.ReadOnlyModelViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    serializer_class = NamespaceSLZ

    def filter_namespaces(self, cluster_id: str):
        paas_cc = PaaSCCClient(auth=ComponentAuth(self.access_token))
        ns_data = paas_cc.get_cluster_namespace_list(project_id=self.project_id, cluster_id=cluster_id)
        ns_list = ns_data['results'] or []

        # 过滤掉 k8s 系统和 bcs 平台使用的命名空间
        return [ns for ns in ns_list if ns['name'] not in K8S_PLAT_NAMESPACE]

    @response_perms(
        action_ids=[NamespaceScopedAction.USE], permission_cls=NamespaceScopedPermission, resource_id_key='iam_ns_id'
    )
    def list(self, request, project_id):
        slz = FilterNamespacesSLZ(data=request.query_params)
        slz.is_valid(raise_exception=True)
        params = slz.validated_data

        cluster_id = params["cluster_id"]
        ns_list = self.filter_namespaces(cluster_id)

        if not ns_list:
            return Response([])

        # check which namespace has the chart_id initialized
        namespace_ids = []
        chart_id = params.get("chart_id")
        if chart_id:
            namespace_ids = set(
                App.objects.filter(project_id=self.project_id, chart__id=chart_id).values_list(
                    "namespace_id", flat=True
                )
            )

        serializer = self.serializer_class(ns_list, many=True)
        data = serializer.data

        for item in data:
            item["has_initialized"] = item["id"] in namespace_ids
            item['iam_ns_id'] = calc_iam_ns_id(cluster_id, item['name'])

        return PermsResponse(data, NamespaceRequest(project_id=project_id, cluster_id=cluster_id))


@with_code_wrapper
class AppUpgradeVersionsView(AppMixin, viewsets.ReadOnlyModelViewSet):
    serializer_class = AppUpgradeVersionsSLZ

    def get_queryset(self):
        instance = App.objects.get(id=self.app_id)
        return instance.get_upgrade_version_selections()


@with_code_wrapper
class AppRollbackSelectionsView(AppMixin, viewsets.ReadOnlyModelViewSet):
    serializer_class = AppRollbackSelectionsSLZ

    def get_queryset(self):
        instance = App.objects.get(id=self.app_id)
        return instance.get_history_releases()


@with_code_wrapper
class AppReleaseDiffView(viewsets.ModelViewSet):
    serializer_class = AppReleaseDiffSLZ


@with_code_wrapper
class AppReleasePreviewView(AccessTokenMixin, ProjectMixin, viewsets.ModelViewSet):
    serializer_class = AppReleasePreviewSLZ


@with_code_wrapper
class AppRollbackPreviewView(AppMixin, AccessTokenMixin, ProjectMixin, viewsets.ModelViewSet):
    serializer_class = AppRollbackPreviewSLZ


@with_code_wrapper
class AppPreviewView(AppMixin, AccessTokenMixin, ProjectMixin, viewsets.ModelViewSet):
    serializer_class = AppPreviewSLZ

    def get_object(self):
        instance = App.objects.get(id=self.app_id)
        content, notes = instance.render_app(username=self.request.user.username, access_token=self.access_token)
        return {"content": content, "notes": notes, "token": self.access_token}


@with_code_wrapper
class AppCreatePreviewView(AccessTokenMixin, ProjectMixin, viewsets.ModelViewSet):
    serializer_class = AppCreatePreviewSLZ


@with_code_wrapper
class AppUpdateChartVersionView(AppMixin, viewsets.ReadOnlyModelViewSet):
    serializer_class = ChartVersionSLZ
    lookup_url_kwarg = "update_chart_version_id"

    def retrieve(self, request, *args, **kwargs):
        app = App.objects.get(id=self.app_id)
        update_chart_version_id = int(self.request.parser_context["kwargs"]["update_chart_version_id"])
        if update_chart_version_id == -1:
            chart_version_snapshot = app.release.chartVersionSnapshot
            chart_version = ChartVersion(
                id=0,
                chart=app.chart,
                keywords="mocked chart version",
                version=chart_version_snapshot.version,
                digest=chart_version_snapshot.digest,
                name=chart_version_snapshot.name,
                home=chart_version_snapshot.home,
                description=chart_version_snapshot.description,
                engine=chart_version_snapshot.engine,
                created=chart_version_snapshot.created,
                maintainers=chart_version_snapshot.maintainers,
                sources=chart_version_snapshot.sources,
                urls=chart_version_snapshot.urls,
                files=chart_version_snapshot.files,
                questions=chart_version_snapshot.questions,
            )
        else:
            chart_version = ChartVersion.objects.get(id=update_chart_version_id)
        slz = self.serializer_class(chart_version)
        return Response(slz.data)


def render_bcs_agent_template(token, bcs_cluster_id, namespace, access_token, project_id, cluster_id):
    bcs_server_host = bcs_utils_client.get_bcs_host(access_token, project_id, cluster_id)
    token = base64.b64encode(str.encode(token))

    # 除 prod 环境外，其它环境都的bcs agent名称都带 环境前缀
    prefix = ""
    if settings.HELM_REPO_ENV != "prod":
        prefix = "%s-" % settings.HELM_REPO_ENV

    render_context = {
        'prefix': prefix,
        'namespace': namespace,
        'token': str(token, encoding='utf-8'),
        'bcs_address': bcs_server_host,
        'bcs_cluster_id': str(bcs_cluster_id),
        'hub_host': settings.DEVOPS_ARTIFACTORY_HOST,
    }

    # 允许通过配置文件，调整 bcs agent YAML 文件名称
    tmpl_name = getattr(settings, 'BCS_AGENT_YAML_TEMPLTE_NAME', 'bcs_agent_tmpl.yaml')
    return render_to_string(tmpl_name, render_context)


@with_code_wrapper
class SyncDict2YamlToolView(viewsets.ModelViewSet):
    serializer_class = SyncDict2YamlToolSLZ


@with_code_wrapper
class SyncYaml2DictToolView(viewsets.ModelViewSet):
    serializer_class = SyncYaml2DictToolSLZ


@with_code_wrapper
class AppStateView(AppMixin, AccessTokenMixin, viewsets.ReadOnlyModelViewSet):
    serializer_class = AppStateSLZ
    lookup_url_kwarg = "app_id"

    def retrieve(self, request, app_id, *args, **kwargs):
        app = App.objects.get(id=self.app_id)

        check_cluster_perm(user=request.user, project_id=app.project_id, cluster_id=app.cluster_id, request=request)

        content = app.release.content
        # resources = parser.parse(content, app.namespace)
        with bcs_utils_client.make_kubectl_client(
            access_token=self.access_token, project_id=app.project_id, cluster_id=app.cluster_id
        ) as (client, err):
            if err:
                raise APIException(str(err))

            state = collect_resource_state(kube_client=client, namespace=app.namespace, content=content)

        return Response(state)


@with_code_wrapper
class AppStatusView(AppMixin, AccessTokenMixin, viewsets.ReadOnlyModelViewSet, ProjectMixin):
    serializer_class = AppStateSLZ
    lookup_url_kwarg = "app_id"

    def retrieve(self, request, app_id, *args, **kwargs):
        app = App.objects.get(id=self.app_id)

        project_code_cache_key = "helm_project_cache_key:%s" % self.project_id
        if project_code_cache_key in cache:
            resp = cache.get(project_code_cache_key)

        else:
            # get_project_name
            resp = paas_cc.get_project(self.access_token, self.project_id)
            if resp.get('code') != 0:
                logger.error(
                    "查询project的信息出错(project_id:{project_id}):{message}".format(
                        project_id=self.project_id, message=resp.get('message')
                    )
                )
                return Response({"code": 500, "message": _("后台接口异常，根据项目ID获取项目英文名失败！")})

            cache.set(project_code_cache_key, resp, 60 * 15)

        project_code = resp["data"]["english_name"]

        check_cluster_perm(user=request.user, project_id=app.project_id, cluster_id=app.cluster_id, request=request)

        kubeconfig = bcs_utils_client.get_kubectl_config_context(
            access_token=self.access_token, project_id=app.project_id, cluster_id=app.cluster_id
        )

        # 获取dashboard对应的path
        bin_path = get_helm_dashboard_path(
            access_token=self.access_token, project_id=app.project_id, cluster_id=app.cluster_id
        )
        try:
            with tempfile.NamedTemporaryFile("w") as f:
                f.write(kubeconfig)
                f.flush()

                data = collect_resource_status(
                    kubeconfig=f.name, app=app, project_code=project_code, cluster_id=app.cluster_id, bin_path=bin_path
                )
        except DashboardExecutionError as e:
            message = "get helm app status failed, error_no: {error_no}\n{output}".format(
                error_no=e.error_no, output=e.output
            )
            return Response(
                {
                    "code": 400,
                    "message": message,
                }
            )
        except DashboardError as e:
            message = "get helm app status failed, dashboard ctl error: {err}".format(err=e)
            logger.exception(message)
            return Response(
                {
                    "code": 400,
                    "message": message,
                }
            )
        except Exception as e:
            message = "get helm app status failed, {err}".format(err=e)
            logger.exception(message)
            return Response(
                {
                    "codee": 500,
                    "message": message,
                }
            )

        response = {
            "status": data,
            "app": {
                "transitioning_action": app.transitioning_action,
                "transitioning_on": app.transitioning_on,
                "transitioning_result": app.transitioning_result,
                "transitioning_message": app.transitioning_message,
            },
        }
        return Response(response)


@with_code_wrapper
class AppAPIView(viewsets.ModelViewSet):
    authentication_classes = (TokenAuthentication,)
    permission_classes = (IsAuthenticated,)
    serializer_class = AppUpgradeByAPISLZ
    lookup_url_kwarg = "app_id"
    queryset = App.objects.all()


class HowToPushHelmChartView(SystemViewSet):
    def get_private_repo_info(self, user, project):
        private_repo = get_or_create_private_repo(user, project)

        if not private_repo.plain_auths:
            return {
                "url": "",
                "username": "",
                "password": "",
            }
        repo_info = private_repo.plain_auths[0]["credentials"]
        repo_info["url"] = private_repo.url
        return repo_info

    def retrieve(self, request, project_id, *args, **kwargs):
        project_code = request.project.english_name
        repo_info = self.get_private_repo_info(user=request.user, project=request.project)

        base_url = request.build_absolute_uri()
        base_url = base_url.split("bcs/k8s")[0]
        repo_url = repo_info["url"]

        context = {
            "project_id": project_id,
            "chart_domain": settings.PLATFORM_REPO_DOMAIN,
            "helm_env": settings.HELM_REPO_ENV,
            "project_code": str(project_code),
            "username": repo_info["username"],
            "password": repo_info["password"],
            "base_url": base_url,
            "repo_url": repo_url,
            "helm_push_parameters": "",
            "rumpetroll_demo_url": settings.RUMPETROLL_DEMO_DOWNLOAD_URL,
        }

        file_prefix = 'backend/helm/app/documentation'

        if request.LANGUAGE_CODE == 'en':
            filename = f'{file_prefix}/how-to-push-chart-en.md'
        else:
            filename = f'{file_prefix}/how-to-push-chart.md'

        with open(os.path.join(settings.STATIC_ROOT.split("staticfiles")[0], filename), "r") as f:
            template = Template(f.read())

        content = template.render(**context)
        return Response({"content": content})


@with_code_wrapper
class ContainerRegistryDomainView(AccessTokenMixin, ProjectMixin, viewsets.ViewSet):
    def retrieve(self, request, project_id, *args, **kwargs):
        cluster_id = request.query_params.get("cluster_id")

        # 获取镜像地址
        jfrog_domain = paas_cc.get_jfrog_domain(
            access_token=self.access_token, project_id=self.project_id, cluster_id=cluster_id
        )

        cluster_info = paas_cc.get_cluster(request.user.token.access_token, project_id, cluster_id)["data"]

        context = dict(
            cluster_id=cluster_id,
            cluster_name=cluster_info["name"],
            jfrog_domain=jfrog_domain,
            expr="{{ .Values.__BCS__.SYS_JFROG_DOMAIN }}",
            link=settings.HELM_DOC_TRICKS,
        )
        note = _(
            '''集群: {cluster}({cluster_id})的容器仓库域名为:{dept_domain},
        可在Chart直接引用 {expr} 更加方便, [详细说明]({link})'''
        ).format(
            cluster=context['cluster_name'],
            cluster_id=context['cluster_id'],
            dept_domain=context['jfrog_domain'],
            expr=context['expr'],
            link=context['link'],
        )
        context["note"] = note
        return Response(data=context)


@with_code_wrapper
class AppTransiningView(AppViewBase):
    def retrieve(self, request, *args, **kwargs):
        app_id = self.request.parser_context["kwargs"]["app_id"]
        # 统一通过ID查询记录信息，防止多次查询时，间隙出现已经删除的情况
        try:
            app = App.objects.get(id=app_id)
        except App.DoesNotExist:
            return Response({"code": 404, "message": "app not found"})

        data = {
            "transitioning_action": app.transitioning_action,
            "transitioning_on": app.transitioning_on,
            "transitioning_result": app.transitioning_result,
            "transitioning_message": app.transitioning_message,
        }
        response = {
            "code": 0,
            "message": "ok",
            "data": data,
        }
        return Response(response)


class ReleaseVersionViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def _get_app(self, project_id: str, cluster_id: str, namespace: str, release_name: str) -> App:
        """查询 app 记录

        :param project_id: 项目 ID
        :param cluster_id: 集群 ID
        :param namespace: 命名空间名称
        :param release_name: 部署后的release的名称
        """
        try:
            return App.objects.get(
                project_id=project_id,
                cluster_id=cluster_id,
                name=release_name,
                namespace=namespace,
            )
        except App.DoesNotExist:
            raise ValidationError(
                f"App not exist, project: {project_id}, cluster_id: {cluster_id},"
                f"namespace: {namespace}, release_name: {release_name}"
            )

    def list_versions(self, request, project_id, cluster_id, namespace, release_name):
        app = self._get_app(project_id, cluster_id, namespace, release_name)
        repo = app.release.repository
        username, password = repo.username_password

        # 获取chart版本
        chart_name = app.chart.name
        client = bk_repo.BkRepoClient(username=username, password=password)
        helm_project_name, helm_repo_name = chart_versions.get_helm_project_and_repo_name(
            request.project.project_code, repo_name=repo.name
        )
        versions = client.get_chart_versions(helm_project_name, helm_repo_name, chart_name)
        versions = chart_versions.ReleaseVersionListSLZ(data=versions, many=True)
        versions.is_valid(raise_exception=True)
        versions = versions.validated_data
        # 以created逆序
        versions = chart_versions.sort_version_list(versions)

        # 因为values.yaml内容可以变更，实例化后的release的版本的values和仓库中的版本可能不一致
        # 因此，添加release的版本信息，用以前端展示relese的版本及详情，格式是: (release-version) version
        release_version = [{"version": f"{RELEASE_VERSION_PREFIX} {app.version}", "name": chart_name}]
        versions = release_version + versions

        return Response(versions)

    def update_or_create_version(self, request, project_id, cluster_id, namespace, release_name):
        slz = ReleaseParamsSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        version = slz.validated_data["version"]

        app = self._get_app(project_id, cluster_id, namespace, release_name)
        # 查询版本详情
        repo = app.release.repository
        username, password = repo.username_password
        helm_project_name, helm_repo_name = chart_versions.get_helm_project_and_repo_name(
            request.project.project_code, repo_name=repo.name
        )
        detail = chart_versions.get_chart_version(
            helm_project_name, helm_repo_name, app.chart.name, version, username, password
        )
        # 更新或创建chart版本
        chart_version, _ = chart_versions.update_or_create_chart_version(app.chart, detail)

        return Response(ChartVersionDetailSLZ(chart_version).data)

    def query_release_version_detail(self, request, project_id, cluster_id, namespace, release_name):
        """查询 release 使用的版本详情"""
        app = self._get_app(project_id, cluster_id, namespace, release_name)
        chart_version_snapshot = app.release.chartVersionSnapshot
        chart_version = chart_versions.release_snapshot_to_version(chart_version_snapshot, app.chart)

        return Response(ChartVersionDetailSLZ(chart_version).data)
