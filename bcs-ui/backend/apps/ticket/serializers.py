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
import re

from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.apps.ticket.models import TlsCert

RE_TLS_NAME = re.compile(r'^[A-Za-z0-9_.]{1,64}$')


class TlsCertModelSLZ(serializers.ModelSerializer):
    class Meta:
        model = TlsCert
        fields = ('id', 'name', 'cert', 'key', "creator", "created", "updated", "updator")
        read_only_fields = ("creator", "created", "updated", "updator")


class TlsCertSlZ(serializers.Serializer):
    name = serializers.RegexField(
        RE_TLS_NAME,
        max_length=64,
        required=True,
        error_messages={'invalid': _('证书名称只能包含：英文大小写、数字、下划线和英文句号，最大长度为64个字符')},
    )
    cert = serializers.CharField(required=True)
    key = serializers.CharField(required=True)

    def validate_name(self, name):
        groups = TlsCert.default_objects.all()
        pk = self.context.get('pk')
        if pk:
            groups = groups.exclude(id=pk)

        is_exist = groups.filter(name=name, project_id=self.context['project_id']).exists()
        if is_exist:
            raise ValidationError('{}[{}]{}'.format(_("证书名称"), name, _("已经存在")))

        return name

    def validate(self, data):
        if not settings.IS_USE_BCS_TLS:
            raise ValidationError(_('容器服务TLS服务未开放'))

        return data
