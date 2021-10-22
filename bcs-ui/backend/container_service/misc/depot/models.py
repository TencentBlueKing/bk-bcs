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

from django.db import models
from django.db.models import Count
from django.utils.translation import ugettext_lazy as _


class ImageCollection(models.Model):
    """用户收藏的镜像信息"""

    user = models.CharField(help_text=_("用户"), max_length=64, db_index=True)
    image_repo = models.CharField(help_text=_("镜像标识"), max_length=512)
    image_project = models.CharField(
        help_text=_("镜像所属仓库,为空则表示为公共仓库"), max_length=32, default='', null=True, blank=True
    )
    create_time = models.DateTimeField(help_text=_("收藏时间"))

    @classmethod
    def get_collect_nums(cls, image_repo_list):
        """镜像的收藏数量"""
        collections = cls.objects.filter(image_repo__in=image_repo_list)
        collect_counts = collections.values('image_repo').annotate(Count('user'))
        # 将数据转化为字典格式
        data_dict = {}
        for _c in collect_counts:
            data_dict[_c['image_repo']] = _c['user__count']
        return data_dict

    @classmethod
    def get_collects_by_user(cls, image_repo_list, username):
        """用户收藏的镜像列表"""
        collections = cls.objects.filter(image_repo__in=image_repo_list, user=username).values_list(
            'image_repo', flat=True
        )
        return collections
