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
from typing import List, Tuple

from django.utils.translation import ugettext_lazy as _
from rest_framework import status, viewsets
from rest_framework.exceptions import ValidationError
from rest_framework.pagination import PageNumberPagination
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.apps.whitelist import enabled_force_sync_chart_repo
from backend.bcs_web.viewsets import SystemViewSet
from backend.components import bk_repo
from backend.helm.app.models import App
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer
from backend.utils.views import ActionSerializerMixin, FilterByProjectMixin, with_code_wrapper

from . import serializers
from .constants import DEFAULT_CHART_REPO_PROJECT_NAME
from .models.chart import Chart, ChartVersion, ChartVersionSnapshot
from .models.repo import Repository
from .providers.repo_provider import add_plain_repo, add_repo
from .serializers import (
    ChartDetailSLZ,
    ChartSLZ,
    ChartVersionSLZ,
    ChartVersionTinySLZ,
    ChartWithVersionRepoSLZ,
    CreateRepoSLZ,
    MinimalRepoSLZ,
    RepositorySyncSLZ,
    RepoSLZ,
)
from .tasks import sync_helm_repo
from .utils import chart as chart_utils
from .utils import chart_versions
from .utils.chart_versions import update_and_delete_chart_versions

logger = logging.getLogger(__name__)


# for debug purpose
# from rest_framework.authentication import SessionAuthentication
# class CsrfExemptSessionAuthentication(SessionAuthentication):
# def enforce_csrf(self, request):
# return  # To not perform the csrf check previously happening


class StandardResultsSetPagination(PageNumberPagination):
    page_size = 10
    page_size_query_param = 'page_size'
    max_page_size = 100


class LargeResultsSetPagination(PageNumberPagination):
    # 用于不需要分页的场景
    page_size = 10000000
    page_size_query_param = 'page_size'
    max_page_size = 1000000


class ChartViewSet(SystemViewSet):
    def list(self, request, project_id):
        # NOTE: 因为传递的project id可能是project code的值，而容器服务内部是通过project id流转，
        # 因此，需要通过request的project中获取project id
        project_id = request.project.project_id
        data = chart_utils.ChartList(project_id).get_chart_data()
        return Response(data)

    def retrieve(self, request, project_id, chart_id):
        project_id = request.project.project_id
        try:
            chart = Chart.objects.get(id=chart_id)
        except Chart.DoesNotExist:
            logger.error("chart: [%s] not found", chart_id)
            return Response()
        slz = ChartWithVersionRepoSLZ(chart)
        return Response(slz.data)


@with_code_wrapper
class ChartVersionView(ActionSerializerMixin, viewsets.ModelViewSet):
    queryset = ChartVersion.objects.all().order_by("-created")
    serializer_class = ChartVersionSLZ
    pagination_class = LargeResultsSetPagination
    lookup_field = "pk"
    lookup_url_kwarg = "version_id"

    action_serializers = {
        'list': ChartVersionTinySLZ,
    }

    def get_queryset(self):
        queryset = self.queryset
        chart_id = self.request.parser_context["kwargs"].get("chart_id")
        if chart_id is not None:
            queryset = queryset.filter(chart__id=chart_id)
        else:
            repo_id = self.request.parser_context["kwargs"].get("repo_id")
            if repo_id is not None:
                queryset = queryset.filter(chart__repo__id=repo_id)
        return queryset


@with_code_wrapper
class RepositoryView(FilterByProjectMixin, viewsets.ModelViewSet):
    """Viewset for helm chart repository management"""

    serializer_class = RepoSLZ
    queryset = Repository.objects.all()

    def get_queryset(self):
        project_id = self.request.parser_context["kwargs"]["project_id"]
        queryset = super(RepositoryView, self).get_queryset()
        queryset = queryset.filter(project_id=project_id)
        return queryset

    def list_detailed(self, request, *args, **kwargs):
        """List all repositories"""
        serializer = RepoSLZ(self.get_queryset(), many=True)
        return Response({'count': self.get_queryset().count(), 'results': serializer.data})

    def list_minimal(self, request, *args, **kwargs):
        """List all repositories minimally"""
        serializer = MinimalRepoSLZ(self.get_queryset(), many=True)
        return Response({'count': self.get_queryset().count(), 'results': serializer.data})

    def retrieve(self, request, project_id, *args, **kwargs):
        """Retrieve certain Chart Repository"""
        repo_id = kwargs.get('repo_id')
        serializer = RepoSLZ(self.queryset.get(project_id=project_id, id=repo_id))
        return Response(data=serializer.data)

    def destroy(self, request, project_id, *args, **kwargs):
        """Destroy Chart Repository"""
        repo_id = kwargs.get('repo_id')
        try:
            self.queryset.get(project_id=project_id, id=repo_id).delete()
        except Exception as e:
            raise error_codes.CheckFailed.f("Delete Chart Repo failed: {}".format(e))
        return Response(status=status.HTTP_204_NO_CONTENT)


@with_code_wrapper
class RepositoryCreateView(FilterByProjectMixin, viewsets.ViewSet):
    """Viewset for creating helm chart repository management"""

    serializer_class = CreateRepoSLZ

    def create(self, request, project_id, *args, **kwargs):
        """Create Repository (support all kind of repo create)"""
        serializer = CreateRepoSLZ(data=request.data)
        serializer.is_valid(raise_exception=True)

        data = serializer.data
        new_chart_repo = add_repo(
            project_id,
            provider_name=data.get('provider'),
            user=self.request.user,
            name=data["name"],
            url=data.get("url"),
        )

        return Response(status=status.HTTP_201_CREATED, data=RepoSLZ(new_chart_repo).data)


@with_code_wrapper
class RepositorySyncView(FilterByProjectMixin, viewsets.ViewSet):
    """RepositorySyncView call sync_helm_repo directly"""

    serializer_class = RepositorySyncSLZ

    def create(self, request, project_id, repo_id, *args, **kwargs):
        """Sync Chart Repository"""
        # 默认不需要设置为强制同步
        sync_helm_repo(repo_id, request.data.get("force_sync") or False)

        data = {"code": 0, "message": "repo sync success"}

        return Response(status=status.HTTP_200_OK, data=data)


@with_code_wrapper
class RepositorySyncByProjectView(FilterByProjectMixin, viewsets.ViewSet):
    """RepositorySyncByProjectView call sync_helm_repo directly"""

    serializer_class = RepositorySyncSLZ

    def create(self, request, project_id, *args, **kwargs):
        """Sync Chart Repository"""
        id_name_list = list(Repository.objects.filter(project_id=project_id).values_list("id", "name"))
        # 白名单控制强制同步项目仓库，不强制同步公共仓库
        force_sync_repo = False
        if enabled_force_sync_chart_repo(project_id):
            force_sync_repo = True

        for repo_id, repo_name in id_name_list:
            # 如果是公共仓库，不允许强制同步
            if repo_name == 'public-repo':
                sync_helm_repo(repo_id, False)
            else:
                sync_helm_repo(repo_id, force_sync_repo)

        data = {"code": 0, "message": "success sync %s repositories" % len(id_name_list)}

        return Response(status=status.HTTP_200_OK, data=data)


class RepositorySyncByProjectAPIView(RepositorySyncByProjectView):
    authentication_classes = []
    permission_classes = []

    def create(self, request, sync_project_id, *args, **kwargs):
        project_id = sync_project_id
        return super(RepositorySyncByProjectAPIView, self).create(request, project_id, *args, **kwargs)


class ChartVersionViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_chart_versions(self, chart_id, version_id):
        if version_id:
            chart_versions = ChartVersion.objects.filter(id=version_id)
        else:
            chart_versions = ChartVersion.objects.filter(chart__id=chart_id)
        if not chart_versions:
            raise ValidationError(_("没有查询到chart对应的版本信息"))
        return chart_versions

    def get_release_queryset(self, chart_id, version_id):
        # 获取chart version
        chart_versions = self.get_chart_versions(chart_id, version_id)
        # 因为是同一个chart，所以通过第一个version信息获取
        chart = chart_versions[0].chart
        version_list = [info.version for info in chart_versions]
        # 根据version和chart_name, 过滤相应的release
        return App.objects.filter(version__in=version_list, chart=chart)

    def release_list(self, request, project_id, chart_id):
        """查询chart下的release
        如果有传递version id，则查询version id对应的信息
        """
        # version id
        version_id = request.query_params.get("version_id")
        release_qs = self.get_release_queryset(chart_id, version_id)
        data = release_qs.values("id", "name", "cluster_id", "namespace", "namespace_id")

        return Response(data)

    def _delete_version(self, username: str, pwd: str, project_code: str, name: str, version: str):
        # 兼容harbor中chart仓库项目名称
        project_name = DEFAULT_CHART_REPO_PROJECT_NAME or project_code
        try:
            client = bk_repo.BkRepoClient(username=username, password=pwd)
            client.delete_chart_version(project_name, project_code, name, version)
        except bk_repo.BkRepoDeleteVersionError as e:
            raise error_codes.APIError(f"delete chart: {name} version: {version} failed, {e}")

    def delete(self, request, project_id, chart_id):
        """删除chart或指定的chart版本"""
        version_id = request.query_params.get("version_id")
        release_qs = self.get_release_queryset(chart_id, version_id)
        # 如果release不为空，则不能进行删除
        if release_qs.exists():
            raise ValidationError(_("chart下存在release，请先删除release"))
        # 如果指定version id，则只删除指定的version，否则删除所有version及chart
        chart_versions = self.get_chart_versions(chart_id, version_id)
        project_code = request.project.project_code
        for info in chart_versions:
            repo_info = info.chart.repository
            auth = repo_info.plain_auths
            # 如果auth为空，则赋值为空
            if not auth:
                username = pwd = ""
            else:
                credentials = auth[0]["credentials"]
                username = credentials["username"]
                pwd = credentials["password"]
            # 删除repo中chart版本记录
            self._delete_version(username, pwd, project_code, info.chart.name, info.version)
            # 处理digest不变动的情况
            ChartVersionSnapshot.objects.filter(digest=info.digest).delete()
            # 删除db中记录
            info.delete()

        # 如果chart下没有版本了，则删除chart，否则如果默认版本删除了，调整chart的对应的默认版本
        chart_all_versions = ChartVersion.objects.filter(chart__id=chart_id)
        chart = Chart.objects.filter(id=chart_id)
        if not chart_all_versions:
            chart.delete()
        elif chart.first().defaultChartVersion is None:
            chart.update(defaultChartVersion=chart_all_versions.order_by("-created")[0])

        # 设置commit id为空，以防出现相同版本digest不变动的情况
        Repository.objects.filter(project_id=project_id).exclude(name="public-repo").update(commit=None)

        return Response()


class HelmChartVersionsViewSet(SystemViewSet):
    def list_releases_by_chart_versions(self, request, project_id, chart_name):
        """查询chart版本对应的release列表"""
        project_code = request.project.project_code
        repo_project_name = self._get_repo_project_name(project_code)
        username, password = self._get_repo_auth(project_code, project_id)
        chart_data = chart_versions.ChartData(
            project_name=repo_project_name, repo_name=project_code, chart_name=chart_name
        )
        repo_auth = chart_versions.RepoAuth(username=username, password=password)
        version_list = self._get_version_list(request, chart_data, repo_auth)
        # 根据版本判断部署的releases
        chart = self._get_chart(project_id, project_code, chart_name)
        release_qs = App.objects.filter(version__in=version_list, chart=chart)
        # 用于前端展示集群/命名空间:release名称
        return Response(release_qs.values("id", "name", "cluster_id", "namespace", "namespace_id"))

    def batch_delete(self, request, project_id, chart_name):
        """删除 chart 版本
        如果需要删除chart，则删除chart下的所有版本即可
        """
        project_code = request.project.project_code
        repo_project_name = self._get_repo_project_name(project_code)
        username, password = self._get_repo_auth(project_code, project_id)
        # 组装数据
        chart_data = chart_versions.ChartData(
            project_name=repo_project_name, repo_name=project_code, chart_name=chart_name
        )
        repo_auth = chart_versions.RepoAuth(username=username, password=password)
        version_list = self._get_version_list(request, chart_data, repo_auth)
        # 开始删除版本
        try:
            chart_versions.batch_delete_chart_versions(chart_data, repo_auth, version_list)
        except Exception as e:
            logger.error("删除项目:%s下chart:%s失败，详情:%s", project_id, chart_name, str(e))
            raise error_codes.APIError(_("删除chart版本失败"))
        # 处理平台记录的版本信息
        chart = self._get_chart(project_id, project_code, chart_name)
        update_and_delete_chart_versions(project_id, project_code, chart, version_list)

        return Response()

    def _get_version_list(
        self, request, chart_data: chart_versions.ChartData, repo_auth: chart_versions.RepoAuth
    ) -> List[str]:
        slz = serializers.ChartVersionParamsSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        version_list = slz.validated_data["version_list"]
        # 如果version列表为空，则需要查询chart下的所有版本
        if not version_list:
            version_list = chart_versions.get_chart_version_list(chart_data, repo_auth)
        return version_list

    def _get_chart(self, project_id: str, project_code: str, chart_name: str) -> Chart:
        try:
            # 限制仅查询项目仓库下的chart
            return Chart.objects.get(name=chart_name, repository__project_id=project_id, repository__name=project_code)
        except Chart.DoesNotExist:
            raise ValidationError(_("chart:{}不存在").format(chart_name))

    def _get_repo_auth(self, project_code: str, project_id: str) -> Tuple[str, str]:
        try:
            repo = Repository.objects.get(name=project_code, project_id=project_id)
        except Repository.DoesNotExist:
            raise ValidationError(
                _("项目【project_id:{}, project_code: {}】没有查询到Chart仓库信息").format(project_id, project_code)
            )
        return repo.username_password

    def _get_repo_project_name(self, project_code: str) -> str:
        """获取仓库的项目名称"""
        # 兼容Harbor项目地址及bk repo项目地址, 其中harbor项目地址固定，bk repo地址项目地址为bcs project code
        return DEFAULT_CHART_REPO_PROJECT_NAME or project_code
