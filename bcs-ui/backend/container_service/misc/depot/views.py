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

from django.conf import settings
from rest_framework import response, views, viewsets

from backend.utils.funutils import convert_mappings
from backend.utils.views import FinalizeResponseMixin

from . import api
from .serializers import AvailableTagSLZ, ImageDetailSLZ, ImageQuerySLZ

logger = logging.getLogger(__name__)


class BaseImage:
    ResultMappings = {
        "search": "searchKey",
        "filters": "repoType",
        "project_id": "projectId",
        "offset": "start",
        "limit": "limit",
    }
    project_id = None

    def parse_images(self, images, username):
        if not images:
            return []

        data_list = []
        if self.project_id:
            pro_name = self.request.project.project_code
        else:
            pro_name = ''
        if settings.DEPOT_PREFIX:
            repo_prefix = f"{settings.DEPOT_PREFIX}/{pro_name}/"
        else:
            repo_prefix = f'{pro_name}/'
        for i in images:
            data_list.append(
                {
                    'name': i.get('repo').split(repo_prefix)[-1] if pro_name else i.get('repo'),
                    "repo": i.get("repo", ""),
                    "deployBy": i.get("createdBy", ""),
                    "type": i.get("type", ""),
                    "desc": i.get("desc", ""),
                    "repoType": i.get("repoType", ""),
                    "modified": i.get("modified", ""),
                    "modifiedBy": i.get("modifiedBy", "") or i.get("createdBy", ""),
                    "imagePath": i.get("imagePath", ""),
                    "downloadCount": i.get("downloadCount", ""),
                }
            )
        data_list = sorted(data_list, key=lambda x: x['repo'])
        return data_list

    def handle_response(self, result, username):
        message = result.get("message", "")

        data = result.get("data", {}) or {}
        images = data.get('imageList', [])
        count = data.get("total")
        return response.Response(
            {
                "code": 0,
                "message": message,
                "data": {
                    "count": count,
                    "next": None,
                    "previous": None,
                    "results": self.parse_images(images, username),
                },
            }
        )


class PublicImages(FinalizeResponseMixin, views.APIView, BaseImage):
    def get_images(self, params):
        query = convert_mappings(self.ResultMappings, params, reversed=True)
        query["access_token"] = self.request.user.token.access_token
        query['project_code'] = ''
        return api.get_public_image_list(query)

    def get(self, request):
        """
        GET /api/depot/images/public/?limit=5&projId=28aa9eda67644a6eb254d694d944307e&offset=0&search=

        HTTP 200 OK
        Content-Type: application/json
        Vary: Accept

        {
            "code": 0,
            "message": "success",
            "data": {
                "count": 2,
                "next": null,
                "previous": null,
                "results": [
                    {
                        "repo": "paas/public/jdk1.8_maven",
                        "deployBy": null,
                        "type": "public",
                        "desc": "description1",
                        "repoType": "",
                    },
                    {
                        "repo": "paas/public/jdk1.8_maven2",
                        "deployBy": null,
                        "type": "public",
                        "desc": "description2",
                        "repoType": "",
                    }
                ]
            }
        }
        """
        self.slz = ImageQuerySLZ(data=request.GET)
        self.slz.is_valid(raise_exception=True)
        result = self.get_images(self.slz.data)
        username = request.user.username

        return self.handle_response(result, username)


class ProjectImage(FinalizeResponseMixin, views.APIView, BaseImage):
    def get_images(self, project_id, params):
        params['project_id'] = project_id
        query = convert_mappings(self.ResultMappings, params, reversed=True)
        query["access_token"] = self.request.user.token.access_token
        query['project_code'] = self.request.project.project_code
        return api.get_project_image_list(query)

    def get(self, request, project_id):
        """
        GET /api/depot/images/project/28aa9eda67644a6eb254d694d944307e/
        HTTP 200 OK
        Content-Type: application/json
        Vary: Accept

        {
            "code": 0,
            "message": "success",
            "data": {
                "count": 2,
                "next": null,
                "previous": null,
                "results": [
                    {
                        "repo": "paas/public/jdk1.8_maven2",
                        "deployBy": null,
                        "type": "public",
                        "desc": "description2",
                        "repoType": "",
                    },
                    {
                        "repo": "paas/public/jdk1.8_maven",
                        "deployBy": null,
                        "type": "public",
                        "desc": "description1",
                        "repoType": "",
                    }
                ]
            }
        }
        """
        self.project_id = project_id
        self.slz = ImageQuerySLZ(data=request.GET)
        self.slz.is_valid(raise_exception=True)
        result = self.get_images(project_id, self.slz.data)
        username = request.user.username

        return self.handle_response(result, username)


class AvailableImage(FinalizeResponseMixin, views.APIView):
    """"""

    def get(self, request, project_id):
        image_list = []
        # 获取公共镜像
        pub_query = {'access_token': self.request.user.token.access_token}
        pub_resp = api.get_public_image_list(pub_query)

        pub_image_data = pub_resp.get('data', {}) or {}
        pub_image_list = pub_image_data.get('imageList', [])
        for _pub in pub_image_list:
            _repo = _pub.get('repo')
            image_list.append(
                {
                    'name': _repo.split(settings.DEPOT_PREFIX)[-1] if settings.DEPOT_PREFIX else _repo,
                    'value': _repo,
                    'is_pub': True,
                }
            )

        # 获取项目镜像
        access_token = self.request.user.token.access_token
        pro_query = {'repoType': 'all', 'projectId': project_id, 'access_token': access_token}
        pro_query['project_code'] = self.request.project.english_name if 'english_name' in self.request.project else ''
        pro_resp = api.get_project_image_list(pro_query)

        pro_image_data = pro_resp.get('data', {}) or {}
        pro_image_list = pro_image_data.get('imageList', [])

        pro_name = request.project.project_code
        if settings.DEPOT_PREFIX:
            repo_prefix = f"{settings.DEPOT_PREFIX}/{pro_name}/"
        else:
            repo_prefix = f'{pro_name}/'
        for _pub in pro_image_list:
            image_list.append(
                {
                    'name': _pub.get('repo').split(repo_prefix)[-1] if pro_name else _pub.get('repo'),
                    'value': _pub.get('repo'),
                    'is_pub': False,
                }
            )

        return response.Response({"code": 0, "message": "success", "data": image_list})


class AvailableTag(FinalizeResponseMixin, views.APIView):
    """"""

    def get(self, request, project_id):
        self.slz = AvailableTagSLZ(data=request.GET)
        self.slz.is_valid(raise_exception=True)

        slz_data = self.slz.data
        repo = slz_data.get('repo')
        is_pub = slz_data.get('is_pub')
        req_project_id = project_id
        params = {
            "projectId": '' if is_pub else req_project_id,
            "repoList": [repo],
            "imageRepo": repo,
            "includeTags": True,
        }
        params["access_token"] = self.request.user.token.access_token
        params['project_code'] = self.request.project.english_name if 'english_name' in self.request.project else ''
        if is_pub:
            tag_resp = api.get_pub_image_info(params)
        else:
            tag_resp = api.get_project_image_info(params)

        try:
            tag_data = tag_resp.get('data', [])[0].get('tags', [])
            tag_data.sort(key=lambda item: item['modified'], reverse=True)
            image_list = [{'value': tag.get('image'), 'text': tag.get('tag')} for tag in tag_data if tag.get('tag')]
        except Exception:
            image_list = []
            logger.exception(u"解析镜像(repo:%s)的tag出错" % repo)

        return response.Response({"code": 0, "message": "success", "data": image_list})


class ImagesInfo(FinalizeResponseMixin, viewsets.ViewSet):
    def get_image_detail(self, request, project_id):
        """查看镜像详情
        分页查询，tag大小一块返回
                        'has_previous': extra_data['page'] != 1,
             'has_next': total > (extra_data['from_pos'] + extra_data['page_size']),
        """
        access_token = request.user.token.access_token
        project_code = request.project.english_name if 'english_name' in self.request.project else ''

        self.slz = ImageDetailSLZ(data=request.query_params)
        self.slz.is_valid(raise_exception=True)
        offset = self.slz.data['offset']
        limit = self.slz.data['limit']
        query_params = {
            'imageRepo': self.slz.data['image_repo'],
            'tagStart': offset,
            'tagLimit': limit,
        }
        resp = api.get_image_tags(access_token, project_id, project_code, offset, limit, **query_params)
        return response.Response(resp)
