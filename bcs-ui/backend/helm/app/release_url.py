# -*- coding: utf-8 -*-

from django.conf.urls import url

from . import views

urlpatterns = [
    url(
        r"namespaces/(?P<namespace>[\w-]+)/releases/(?P<release_name>[\w\-]+)/versions/$",
        views.ReleaseVersionViewSet.as_view({"get": "list_versions", "post": "update_or_create_version"}),
    ),
    url(
        r"namespaces/(?P<namespace>[\w-]+)/releases/(?P<release_name>[\w\-]+)/detail/$",
        views.ReleaseVersionViewSet.as_view({"get": "query_release_version_detail"}),
    ),
]
