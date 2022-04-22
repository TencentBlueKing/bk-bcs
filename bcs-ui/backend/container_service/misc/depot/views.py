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
from typing import Dict, Optional

from django.conf import settings
from rest_framework import permissions
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.components.bk_repo import BkRepoClient, PageData

from .serializers import AvailableTagSLZ, ImageDetailSLZ, ImageQuerySLZ

logger = logging.getLogger(__name__)


class BaseImagesViewSet(SystemViewSet):
    def list_images(
        self, access_token: str, project_name: str, repo_name: str, page: PageData, name: Optional[str] = None
    ) -> Dict:
        client = BkRepoClient(access_token)
        return client.list_images(project_name, repo_name, page, name)

    def compose_data(self, access_token: str, project_name: str, repo_name: str, repo_type: str) -> Dict:
        params = self.params_validate(ImageQuerySLZ)
        page = PageData(pageNumber=params["offset"], pageSize=params["limit"])
        resp_data = self.list_images(access_token, project_name, repo_name, page, params.get("search"))
        # 转换为前端需要的格式
        data = [
            {
                "name": i["name"],
                "repo": f"/{project_name}/{repo_name}/{i['name']}",
                "deployBy": "system",
                "type": repo_type,
                "desc": i["description"],
                "repoType": "",
                "modified": i["lastModifiedDate"],
                "modifiedBy": i["lastModifiedBy"],
                "imagePath": "",
                "downloadCount": "",
            }
            for i in resp_data["records"]
        ]
        return {"count": resp_data["totalRecords"], "results": data}


class SharedRepoImagesViewSet(BaseImagesViewSet):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request):
        """查询共享仓库下的镜像信息"""
        return Response(
            self.compose_data(
                request.user.token.access_token,
                settings.BK_REPO_SHARED_PROJECT_NAME,
                settings.BK_REPO_SHARED_IMAGE_DEPOT_NAME,
                "public",
            )
        )


class ProjectImagesViewSet(BaseImagesViewSet):
    def get(self, request, project_id):
        """查询项目专用仓库下的镜像信息"""
        project_name = repo_name = request.project.project_code
        return Response(
            self.compose_data(request.user.token.access_token, project_name, f"{repo_name}-docker", "private")
        )


class AvailableImagesViewSet(BaseImagesViewSet):
    def get(self, request, project_id):
        """获取项目下可用的镜像列表，包含项目专用仓库和共享仓库"""
        page = PageData()
        shared_images = self.list_images(
            request.user.token.access_token,
            settings.BK_REPO_SHARED_PROJECT_NAME,
            settings.BK_REPO_SHARED_IMAGE_DEPOT_NAME,
            page,
        )
        project_name = repo_name = request.project.project_code
        dedicated_images = self.list_images(request.user.token.access_token, project_name, f"{repo_name}-docker", page)
        # 组装数据
        data = []
        for i in dedicated_images.get("records") or []:
            name = value = f"{project_name}/{repo_name}-docker/{i['name']}"
            data.append({"name": name, "value": value, "is_pub": False})
        for i in shared_images.get("records") or []:
            project_name = settings.BK_REPO_SHARED_PROJECT_NAME
            repo_name = settings.BK_REPO_SHARED_IMAGE_DEPOT_NAME
            name = value = f"/{project_name}/{repo_name}/{i['name']}"
            data.append({"name": name, "value": value, "is_pub": True})
        return Response(data)


class BaseImageTagsViewSet(SystemViewSet):
    def list_image_tags(self, access_token: str, image_path: str) -> Dict:
        page = PageData()
        # 前端传递的路径格式为/项目名/仓库名/镜像，需要转换为 镜像
        repo_list = image_path.split("/", 3)
        project_name = repo_list[1]
        repo_name = repo_list[2]
        image_name = repo_list[3]
        client = BkRepoClient(access_token)
        return client.list_image_tags(project_name, repo_name, image_name, page)


class AvailableTagsViewSet(BaseImageTagsViewSet):
    def get(self, request, project_id):
        params = self.params_validate(AvailableTagSLZ)
        resp_data = self.list_image_tags(params["repo"])
        # 转换为前端需要的数据
        data = [
            {"value": f"{settings.BK_REPO_DOMAIN}/{params['repo']}:{i['tag']}", "text": i["tag"]}
            for i in resp_data["records"]
        ]
        return Response(data)


class ImageDetailViewSet(SystemViewSet):
    def get(self, request, project_id):
        params = self.params_validate(ImageDetailSLZ)
        # 前端传递的格式为/项目名/仓库名/镜像，需要转换为 镜像
        resp_data = self.list_image_tags(params["image_repo"])
        tags = [
            {"tag": i["tag"], "size": f"{i['size']} MB", "modified": i["lastModifiedDate"]}
            for i in resp_data["records"]
        ]
        latest_tag = resp_data["records"][-1]
        data = {
            "tags": tags,
            "modified": latest_tag["lastModifiedDate"],
            "modifiedBy": latest_tag["lastModifiedBy"],
            "imageName": params["image_repo"].split("/", 3)[-1],
            "repo": params["image_repo"],
            "tagCount": resp_data["totalRecords"],
            "downloadCount": 0,  # 暂无下载次数记录
        }

        return Response(data)


try:
    from .views_ext import *  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
