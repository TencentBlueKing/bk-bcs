# -*- coding: utf-8 -*-
from rest_framework import serializers


class FilterNamespacesSLZ(serializers.Serializer):
    filter_use_perm = serializers.BooleanField(default=True)
    cluster_id = serializers.CharField(required=False)
    chart_id = serializers.IntegerField(required=False)
