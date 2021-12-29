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
from django.conf import settings
from django.conf.urls import include, url
from django.contrib import admin
from django.urls import path, re_path
from django.views.decorators.cache import never_cache

from backend.utils import healthz
from backend.utils.views import LoginSuccessView, VueTemplateView

urlpatterns = [
    url(r"^admin/", admin.site.urls),
    url(r"^api/healthz/", healthz.healthz_view),
    url(r"^api/test/sentry/", healthz.test_sentry),
    url(r"^", include("backend.accounts.urls")),
    # 项目管理, namespace 名称 SKIP_REQUEST_NAMESPACE 配置中, 不能省略
    re_path(
        r"^",
        include(
            ("backend.container_service.projects.urls", "backend.container_service.projects"), namespace="projects"
        ),
    ),
    url(r"^api/iam/", include("backend.iam.urls")),
    # 仓库管理
    url(r"^", include("backend.container_service.misc.depot.urls")),
    # 集群管理
    url(r"^", include("backend.container_service.clusters.urls")),
    # web_console
    url(r"^", include("backend.web_console.rest_api.urls")),
    # 网络管理
    url(r"^", include("backend.uniapps.network.urls")),
    # Resource管理
    url(r"^", include("backend.uniapps.resource.urls")),
    # 配置管理(旧模板集)
    url(r"^", include("backend.templatesets.legacy_apps.configuration.urls")),
    # TODO 新模板集url入口，后续替换上面的configuration
    url(r"^api/templatesets/projects/(?P<project_id>\w{32})/", include("backend.templatesets.urls")),
    # 变量管理
    url(r"^", include("backend.templatesets.var_mgmt.urls")),
    # 应用管理
    url(r"^", include("backend.uniapps.application.urls")),
    url(r"^", include("backend.bcs_web.audit_log.urls")),
    # 权限验证
    url(r"^", include("backend.bcs_web.legacy_verify.urls")),
    url(r"^api-auth/", include("rest_framework.urls")),
    # BCS K8S special urls
    url(r"^", include("backend.helm.helm.urls")),
    url(r"^", include("backend.helm.app.urls")),
    # Ticket凭证管理
    url(r"^", include("backend.apps.ticket.urls")),
    url(
        r"^api/hpa/projects/(?P<project_id>\w{32})/",
        include(
            "backend.kube_core.hpa.urls",
        ),
    ),
    # cd部分api
    url(r"^cd_api/", include("backend.uniapps.apis.urls")),
    url(r"^apis/", include("backend.api_urls")),
    # dashboard 相关 URL
    url(
        r"^api/dashboard/projects/(?P<project_id>\w{32})/clusters/(?P<cluster_id>[\w\-]+)/",
        include("backend.dashboard.urls"),
    ),
    # k8s Metric 相关 URL
    url(
        r"^api/metrics/projects/(?P<project_id>\w{32})/clusters/(?P<cluster_id>[\w\-]+)/",
        include("backend.container_service.observability.metric.urls"),
    ),
    # 标准日志输出
    path(
        "api/logs/projects/<slug:project_id>/clusters/<slug:cluster_id>/",
        include("backend.container_service.observability.log_stream.urls"),
    ),
    re_path(r"^api/helm/projects/(?P<project_id>\w{32})/", include("backend.helm.urls")),
    path(r"change_log/", include("backend.change_log.urls")),
    # cluster manager的代理请求
    url(
        r"^{}".format(settings.CLUSTER_MANAGER_PROXY["PREFIX_PATH"]),
        include("backend.container_service.clusters.mgr.proxy.urls"),
    ),
]

# 导入版本特定的urls
try:
    from backend.urls_ext import urlpatterns as urlpatterns_ext

    urlpatterns += urlpatterns_ext
except ImportError:
    pass


# vue urls 需要放到最后面
urlpatterns_vue = [
    # fallback to vue view
    url(r"^login_success.html", never_cache(LoginSuccessView.as_view())),
    url(r"^(?P<project_code>[\w\-]+)", never_cache(VueTemplateView.as_view(container_orchestration="k8s"))),
]
urlpatterns += urlpatterns_vue
