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

from django.db import models

from ..constants import FileResourceName
from .base import BaseModel

IMAGES_PATTEN = re.compile('image:(.*)')


def find_containers(content):
    images = IMAGES_PATTEN.findall(content)
    containers = []
    for image in images:
        # remove comment
        if '#' in image:
            image = image.split('#')[0]
        image = image.strip().strip("'").split('/')[-1]
        containers.append({'name': 'container', 'image': image})
    return containers


class ResourceFile(BaseModel):
    name = models.TextField('file name')
    content = models.TextField('file content')
    resource_name = models.CharField(choices=FileResourceName.get_choices(), max_length=32)
    template_id = models.IntegerField()

    class Meta:
        ordering = ('name',)
