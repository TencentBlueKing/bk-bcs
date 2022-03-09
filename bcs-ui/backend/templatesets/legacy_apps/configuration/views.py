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
import time

from django.db import transaction
from django.utils import timezone
from django.utils.translation import ugettext_lazy as _
from rest_framework import generics
from rest_framework.exceptions import ValidationError
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response
from rest_framework.views import APIView

from backend.bcs_web.audit_log.audit.decorators import log_audit, log_audit_on_view
from backend.bcs_web.audit_log.constants import ActivityType
from backend.bcs_web.viewsets import SystemViewSet
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.templateset import (
    TemplatesetAction,
    TemplatesetCreatorAction,
    TemplatesetPermCtx,
    TemplatesetPermission,
    TemplatesetRequest,
)
from backend.templatesets.legacy_apps.instance.utils import check_template_available, validate_template_id
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse

from . import serializers_new
from .auditor import TemplatesetAuditor
from .mixins import TemplatePermission
from .models import (
    Template,
    VersionedEntity,
    get_default_version,
    get_model_class_by_resource_name,
    get_template_by_project_and_id,
)
from .serializers import ResourceRequstSLZ, ResourceSLZ, TemplateCreateSLZ, get_template_info, is_tempalte_instance
from .utils import create_template_with_perm_check, validate_resource_name


def is_create_template(template_id):
    # template_id 为 0 时，创建模板集
    if template_id == "0":
        return True
    return False


class TemplatesView(APIView):
    queryset = Template.objects.all()
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def filter_queryset_with_params(self, project_id, name):
        params = {"project_id": project_id}
        if name:
            params["name__icontains"] = name.strip()
        return self.queryset.filter(**params)

    @response_perms(
        action_ids=[
            TemplatesetAction.VIEW,
            TemplatesetAction.UPDATE,
            TemplatesetAction.DELETE,
            TemplatesetAction.INSTANTIATE,
            TemplatesetAction.COPY,
        ],
        permission_cls=TemplatesetPermission,
        resource_id_key='id',
    )
    def get(self, request, project_id):
        serializer = serializers_new.SearchTemplateSLZ(data=request.query_params)
        serializer.is_valid(raise_exception=True)
        data = serializer.validated_data

        templates = self.filter_queryset_with_params(project_id, data["search"])
        num_of_templates = templates.count()

        # 获取项目类型 backend.utils.permissions做了处理
        kind = request.project.kind
        # 添加分页信息
        limit, offset = data["limit"], data["offset"]
        templates = templates[offset : limit + offset]

        serializer = serializers_new.ListTemplateSLZ(templates, many=True, context={"kind": kind})
        template_list = serializer.data

        return PermsResponse(
            data={
                'count': num_of_templates,
                'has_previous': True if offset != 0 else False,
                'has_next': True if (offset + limit) < num_of_templates else False,
                'results': template_list,
            },
            resource_data=template_list,
            resource_request=TemplatesetRequest(project_id=project_id),
        )


class CreateTemplateDraftView(SystemViewSet, TemplatePermission):
    @log_audit_on_view(TemplatesetAuditor, activity_type=ActivityType.Add)
    def create_draft(self, request, project_id, template_id):
        if is_create_template(template_id):
            # template dict like {'desc': '', 'name': ''}
            tpl_args = request.data.get("template", {})
            template = create_template_with_perm_check(request, project_id, tpl_args)
        else:
            template = get_template_by_project_and_id(project_id, template_id)

        self.can_edit_template(request, template)

        serializer = serializers_new.TemplateDraftSLZ(
            template, data=request.data, context={"template_id": template.id}
        )
        serializer.is_valid(raise_exception=True)
        serializer.save(draft_updator=request.user.username)

        validated_data = serializer.validated_data

        request.audit_ctx.update_fields(
            resource=template.name, resource_id=template.id, extra=validated_data, description=_("保存草稿")
        )

        return Response(
            {"template_id": template.id, "show_version_id": -1, "real_version_id": validated_data["real_version_id"]}
        )


class CreateAppResourceView(APIView, TemplatePermission):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def _compose_create_data(self, data):
        del data["version_id"]
        if "resource_name" in data:
            del data["resource_name"]
        del data["project_id"]
        del data["resource_id"]
        data["creator"] = self.request.user.username
        return data

    @transaction.atomic
    def _create_resource_entity(self, resource_name, template_id, create_data):
        resource_class = get_model_class_by_resource_name(resource_name)
        resource = resource_class.perform_create(**create_data)
        # 创建新的模板集版本, 只保存各类资源的第一个
        resource_entity = {resource_name: str(resource.id)}
        ventity = VersionedEntity.objects.create(
            template_id=template_id,
            version=get_default_version(),
            entity=resource_entity,
            creator=self.request.user.username,
        )
        return {
            "id": resource.id,
            "version": ventity.id,
            "template_id": template_id,
            "resource_data": resource.get_res_config(is_simple=True),
        }

    @transaction.atomic
    def post(self, request, project_id, template_id, resource_name):
        validate_resource_name(resource_name)

        if is_create_template(template_id):
            # template dict like {'desc': '', 'name': ''}
            tpl_args = request.data.get("template", {})
            template = create_template_with_perm_check(request, project_id, tpl_args)
        else:
            template = get_template_by_project_and_id(project_id, template_id)

        self.can_edit_template(request, template)

        data = request.data or {}
        data.update({"version_id": 0, "resource_id": 0, "project_id": project_id})
        serializer_class = serializers_new.get_slz_class_by_resource_name(resource_name)
        serializer = serializer_class(data=data)
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data
        create_data = self._compose_create_data(validated_data)
        ret_data = self._create_resource_entity(resource_name, template.id, create_data)
        return Response(ret_data)


class UpdateDestroyAppResourceView(APIView, TemplatePermission):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def _compose_update_data(self, data):
        del data["version_id"]
        if "resource_name" in data:
            del data["resource_name"]
        del data["project_id"]
        del data["resource_id"]
        data["creator"] = self.request.user.username
        return data

    def put(self, request, project_id, version_id, resource_name, resource_id):
        """模板集中有 version 信息时，从 version 版本创建新的版本信息"""
        validate_resource_name(resource_name)

        serializer = serializers_new.VentityWithTemplateSLZ(data=self.kwargs)
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        ventity = validated_data["ventity"]
        template = validated_data["template"]
        self.can_edit_template(request, template)

        data = request.data or {}
        data.update({"version_id": version_id, "resource_id": resource_id, "project_id": project_id})

        serializer_class = serializers_new.get_slz_class_by_resource_name(resource_name)
        serializer = serializer_class(data=data)
        serializer.is_valid(raise_exception=True)

        update_data = self._compose_update_data(serializer.validated_data)
        resource_class = get_model_class_by_resource_name(resource_name)
        if int(resource_id):
            robj = resource_class.perform_update(resource_id, **update_data)
        else:
            robj = resource_class.perform_create(**update_data)

        new_ventity = VersionedEntity.update_for_new_ventity(
            ventity.id, resource_name, resource_id, str(robj.id), **{"creator": self.request.user.username}
        )

        # model Template updated field need change when update resource
        template.save(update_fields=["updated"])

        return Response(
            {"id": robj.id, "version": new_ventity.id, "resource_data": robj.get_res_config(is_simple=True)}
        )

    def delete(self, request, project_id, version_id, resource_name, resource_id):
        validate_resource_name(resource_name)

        serializer = serializers_new.VentityWithTemplateSLZ(data=self.kwargs)
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        ventity = validated_data["ventity"]
        template = validated_data["template"]
        self.can_edit_template(request, template)

        # 关联关系检查
        serializers_new.can_delete_resource(ventity, resource_name, resource_id)

        new_ventity = VersionedEntity.update_for_delete_ventity(
            ventity.id, resource_name, resource_id, **{"creator": self.request.user.username}
        )
        return Response({"id": resource_id, "version": new_ventity.id})


class LockTemplateView(SystemViewSet, TemplatePermission):
    @log_audit_on_view(TemplatesetAuditor, activity_type=ActivityType.Add)
    def lock_template(self, request, project_id, template_id):
        template = get_template_by_project_and_id(project_id, template_id)
        self.validate_template_locked(request, template)

        request.audit_ctx.update_fields(resource=template.name, resource_id=template_id, description=_("加锁模板集"))

        template.is_locked = True
        template.locker = request.user.username
        template.save()

        return Response(data={})

    @log_audit_on_view(TemplatesetAuditor, activity_type=ActivityType.Delete)
    def unlock_template(self, request, project_id, template_id):
        template = get_template_by_project_and_id(project_id, template_id)
        self.validate_template_locked(request, template)

        request.audit_ctx.update_fields(resource=template.name, resource_id=template_id, description=_("解锁模板集"))

        template.is_locked = False
        template.locker = ""
        template.save()

        return Response(data={})


# TODO refactor
class TemplateResourceView(generics.RetrieveAPIView):
    serializer_class = ResourceSLZ
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_queryset(self):
        return VersionedEntity.objects.filter(id=self.pk)

    def get_serializer_context(self):
        context = super(TemplateResourceView, self).get_serializer_context()
        self.slz = ResourceRequstSLZ(data=self.request.GET, context={"project_kind": self.project_kind})
        self.slz.is_valid(raise_exception=True)

        context.update(self.slz.data)
        return context

    def get(self, request, project_id, pk):
        self.pk = pk
        self.project_kind = request.project.kind
        return super(TemplateResourceView, self).get(self, request)


# TODO refactor
class SingleTemplateView(generics.RetrieveUpdateDestroyAPIView):
    serializer_class = serializers_new.TemplateSLZ
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    iam_perm = TemplatesetPermission()

    def get_queryset(self):
        return Template.objects.filter(project_id=self.project_id, id=self.pk)

    def get_serializer_context(self):
        context = super(SingleTemplateView, self).get_serializer_context()
        context.update({"request": self.request})
        return context

    @log_audit(TemplatesetAuditor, activity_type=ActivityType.Modify)
    def perform_update(self, serializer):
        self.audit_ctx.update_fields(
            user=self.request.user.username, project_id=self.project_id, description=_("更新模板集")
        )
        instance = serializer.save(updator=self.request.user.username, project_id=self.project_id)
        self.audit_ctx.update_fields(resource=instance.name, resource_id=instance.id, extra=serializer.data)

    def post(self, request, project_id, pk):
        self.request = request
        self.project_id = project_id
        self.pk = pk
        template = validate_template_id(project_id, pk, is_return_tempalte=True)

        # 验证用户是否有编辑权限
        perm_ctx = TemplatesetPermCtx(username=request.user.username, project_id=project_id, template_id=pk)
        self.iam_perm.can_update(perm_ctx)

        # 检查模板集是否可操作（即未被加锁）
        check_template_available(template, request.user.username)

        # 验证模板名是否已经存在
        new_template_name = request.data.get("name")
        is_exist = (
            Template.default_objects.exclude(id=pk).filter(name=new_template_name, project_id=project_id).exists()
        )
        if is_exist:
            detail = {"field": [_("模板集名称[{}]已经存在").format(new_template_name)]}
            raise ValidationError(detail=detail)

        self.slz = serializers_new.UpdateTemplateSLZ(data=request.data)
        self.slz.is_valid(raise_exception=True)

        return super(SingleTemplateView, self).update(self.slz)

    @response_perms(
        action_ids=[TemplatesetAction.VIEW, TemplatesetAction.UPDATE, TemplatesetAction.INSTANTIATE],
        permission_cls=TemplatesetPermission,
        resource_id_key='id',
    )
    def get(self, request, project_id, pk):
        self.request = request
        self.project_id = project_id
        self.pk = pk

        validate_template_id(project_id, pk, is_return_tempalte=True)

        perm_ctx = TemplatesetPermCtx(username=request.user.username, project_id=project_id, template_id=pk)
        self.iam_perm.can_view(perm_ctx)

        # 获取项目类型
        kind = request.project.kind

        tems = self.get_queryset()
        if tems:
            tem = tems.first()
            data = get_template_info(tem, kind)
        else:
            data = {}

        return PermsResponse(data, TemplatesetRequest(project_id=project_id))

    @log_audit_on_view(TemplatesetAuditor, activity_type=ActivityType.Delete)
    def delete(self, request, project_id, pk):
        self.request = request
        self.project_id = project_id
        self.pk = pk
        # 验证用户是否删除权限
        template = validate_template_id(project_id, pk, is_return_tempalte=True)

        perm_ctx = TemplatesetPermCtx(username=request.user.username, project_id=project_id, template_id=pk)
        self.iam_perm.can_delete(perm_ctx)

        # 检查模板集是否可操作（即未被加锁）
        check_template_available(template, request.user.username)

        # 已经实例化过的版本，不能被删除
        exist_version = is_tempalte_instance(pk)
        if exist_version:
            return Response({"code": 400, "message": _("模板集已经被实例化过，不能被删除"), "data": {}})
        instance = self.get_queryset().first()

        request.audit_ctx.update_fields(resource=instance.name, resource_id=instance.id, description=_("删除模板集"))

        # 删除后名称添加 [deleted]前缀
        _del_prefix = f"[deleted_{int(time.time())}]"
        del_tem_name = f"{_del_prefix}{instance.name}"
        self.get_queryset().update(name=del_tem_name, is_deleted=True, deleted_time=timezone.now())

        return Response({"code": 0, "message": "OK", "data": {"id": pk}})

    @log_audit_on_view(TemplatesetAuditor, activity_type=ActivityType.Add)
    @transaction.atomic
    def put(self, request, project_id, pk):
        """复制模板"""
        self.request = request

        validate_template_id(project_id, pk, is_return_tempalte=True)

        perm_ctx = TemplatesetPermCtx(username=request.user.username, project_id=project_id, template_id=pk)
        self.iam_perm.can_copy(perm_ctx)

        self.project_id = project_id
        self.pk = pk
        data = request.data
        data["project_id"] = project_id
        self.slz = TemplateCreateSLZ(data=data)
        self.slz.is_valid(raise_exception=True)
        new_template_name = self.slz.data["name"]
        # 验证模板名是否已经存在
        is_exist = Template.default_objects.filter(name=new_template_name, project_id=project_id).exists()
        if is_exist:
            detail = {"field": [_("模板集名称[{}]已经存在").format(new_template_name)]}
            raise ValidationError(detail=detail)
        # 验证 old模板集id 是否正确
        old_tems = self.get_queryset()
        if not old_tems.exists():
            detail = {"field": [_("要复制的模板集不存在")]}
            raise ValidationError(detail=detail)
        old_tem = old_tems.first()

        username = request.user.username
        template_id, version_id, show_version_id = old_tem.copy_tempalte(project_id, new_template_name, username)

        self.iam_perm.grant_resource_creator_actions(
            TemplatesetCreatorAction(
                template_id=template_id, name=new_template_name, project_id=project_id, creator=username
            ),
        )

        # 记录操作日志
        request.audit_ctx.update_fields(resource=new_template_name, resource_id=template_id, description=_("复制模板集"))
        return Response(
            {
                "code": 0,
                "message": "OK",
                "data": {"template_id": template_id, "version_id": version_id, "show_version_id": show_version_id},
            }
        )
