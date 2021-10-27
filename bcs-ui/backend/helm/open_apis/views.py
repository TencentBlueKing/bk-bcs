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
from django.utils.translation import ugettext_lazy as _
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from backend.bcs_web.apis.views import NoAccessTokenBaseAPIViewSet
from backend.helm.app import views as app_views
from backend.helm.helm import views as chart_views
from backend.helm.helm.models.chart import Chart, ChartVersion
from backend.helm.helm.models.repo import Repository
from backend.utils.error_codes import error_codes


class ChartsApiView(NoAccessTokenBaseAPIViewSet, chart_views.ChartViewSet):
    def list_charts(self, request, project_id_or_code):
        return self.list(request, request.project.project_id)

    def retrieve_chart(self, request, project_id_or_code, chart_id):
        return self.retrieve(request, request.project.project_id, chart_id)


class ChartVersionApiView(NoAccessTokenBaseAPIViewSet, chart_views.ChartVersionView):
    def chart_versions(self, request, project_id_or_code, chart_id):
        resp = self.list(request, request.project.project_id, chart_id)
        results = resp.data['results']
        versions = [{'id': info['id'], 'version': info['version']} for info in results]
        return Response(versions)

    def retrieve_chart_version(self, request, project_id_or_code, chart_id, version_id):
        resp = self.retrieve(request, request.project.project_id, chart_id, version_id)
        data = resp.data.get('data') or {}
        version_info = {
            'id': data['id'],
            'name': data['name'],
            'version': data['version'],
            'chart': data['chart'],
            'files': data['files'],
        }

        return Response(version_info)

    def retrieve_valuefile(self, request, project_id_or_code, chart_id, version_id):
        resp = self.retrieve(request, request.project.project_id, chart_id, version_id)
        data = resp.data.get('data') or {}
        values_content = ''
        for f in data['files']:
            # only support one values file
            if '/values.yaml' in f:
                values_content = data['files'][f]
                break
        return Response({'valuesfile': values_content})


class ChartAppNamespaceApiView(NoAccessTokenBaseAPIViewSet, app_views.AppNamespaceView):
    def list_available_namespaces(self, request, project_id_or_code):
        resp = self.list(request, request.project.project_id)
        namespaces = []
        # 适配前端请求，返回对象格式
        for info in resp.data:
            # item = {'cluster': info['name'], 'children': []}
            for ns_info in info.get('children') or []:
                namespaces.append(
                    {
                        'id': ns_info['id'],
                        'name': ns_info['name'],
                        'cluster_ns_name': f'{info["name"]}/{ns_info["name"]}',
                        'project_id': ns_info['project_id'],
                        'cluster_id': ns_info['cluster_id'],
                        'has_initialized': ns_info.get('has_initialized', False),
                    }
                )

        return Response(namespaces)


class ChartsAppApiView(NoAccessTokenBaseAPIViewSet, app_views.AppView):
    def list_app(self, request, project_id_or_code):
        return self.list(request, request.project.project_id)

    def create_app(self, request, project_id_or_code):
        return self.create(request, request.project.project_id)

    def update(self, request, project_id_or_code, app_id):
        return super(ChartsAppApiView, self).update(request, request.project.project_id, app_id)

    def delete_app(self, request, project_id_or_code, app_id):
        return self.destroy(request, request.project.project_id, app_id)

    def retrieve(self, request, project_id_or_code, app_id):
        return super(ChartsAppApiView, self).retrieve(request, request.project.project_id, app_id)


class ChartsAppTransitionApiView(NoAccessTokenBaseAPIViewSet, app_views.AppTransiningView):
    def retrieve_app(self, request, project_id_or_code, app_id):
        return self.retrieve(request, request.project.project_id, app_id)


class AppUpgradeVersionView(NoAccessTokenBaseAPIViewSet, app_views.AppUpgradeVersionsView):
    def list_app_versions(self, request, project_id_or_code, app_id):
        return self.list(request, request.project.project_id, app_id)


class SyncRepoView(NoAccessTokenBaseAPIViewSet, chart_views.RepositorySyncView):
    def sync_repo(self, request, project_id_or_code):
        project_id = request.project.project_id
        repos = Repository.objects.filter(project_id=project_id)
        # 同步项目仓库
        for info in repos:
            if info.name != 'public-repo':
                return self.create(request, project_id, info.id)
        # 查询不到项目仓库时，返回异常提示
        return error_codes.ResNotFoundError(_("没有查询到项目仓库"))


class DeleteChartOrVersion(NoAccessTokenBaseAPIViewSet, chart_views.ChartVersionViewSet):
    def get_repo(self, project_id):
        try:
            project_repo = Repository.objects.exclude(name="public-repo").get(project_id=project_id)
        except Exception as e:
            raise ValidationError(_("项目:{}下没有查询到项目仓库, 错误消息: {}").format(project_id, e))
        return project_repo

    def get_chart(self, repo, chart_name):
        charts = Chart.objects.filter(repository=repo, name=chart_name)
        chart = charts.first()
        if not chart:
            raise ValidationError(_("仓库下没有查找到chart:{}").format(chart_name))
        return chart

    def get_chart_version(self, chart, version_name):
        if not version_name:
            return None
        versions = ChartVersion.objects.filter(chart=chart, version=version_name)
        version = versions.first()
        if not version:
            raise ValidationError(_("Chart:{}下没有查找到版本:{}").format(chart.name, version_name))
        return version

    def delete(self, request, project_id_or_code, chart_name):
        project_id = request.project.project_id
        project_repo = self.get_repo(project_id)
        chart = self.get_chart(project_repo, chart_name)
        version_name = request.query_params.get("version_name")
        chart_version = self.get_chart_version(chart, version_name)
        if chart_version:
            request.query_params._mutable = True
            request.query_params["version_id"] = chart_version.id
        return super(DeleteChartOrVersion, self).delete(request, project_id, chart.id)
