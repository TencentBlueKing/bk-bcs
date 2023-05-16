# -*- coding: utf-8 -*-

from django.conf.urls import url

from . import views

urlpatterns = [
    url(r"charts/(?P<chart_name>[\w\-\.]+)/versions/$", views.ChartVersionsViewSet.as_view({"get": "list_versions"})),
    url(
        r"charts/(?P<chart_name>[\w\-\.]+)/versions/(?P<version>[\w\-\.]+)/$",
        views.ChartVersionsViewSet.as_view({"post": "update_or_create_version"}),
    ),
]
