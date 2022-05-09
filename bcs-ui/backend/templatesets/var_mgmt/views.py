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
import json
import re
import time

from django.db import transaction
from django.db.models import Q
from django.utils import timezone
from django.utils.translation import ugettext_lazy as _
from rest_framework import generics, views, viewsets
from rest_framework.exceptions import ValidationError
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.accounts import bcs_perm
from backend.bcs_web.audit_log.audit.decorators import log_audit, log_audit_on_view
from backend.bcs_web.audit_log.constants import ActivityType
from backend.components import paas_cc
from backend.container_service.clusters.constants import ClusterType
from backend.container_service.projects.base.constants import LIMIT_FOR_ALL_DATA
from backend.templatesets.legacy_apps.configuration.models import MODULE_DICT
from backend.templatesets.legacy_apps.configuration.utils import check_var_by_config, get_all_template_info_by_project
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer
from backend.utils.views import FinalizeResponseMixin

from ..legacy_apps.instance.serializers import VariableNamespaceSLZ
from . import serializers
from .auditor import VariableAuditor
from .import_vars import import_vars
from .models import ClusterVariable, NameSpaceVariable, Variable


class ListCreateVariableView(generics.ListCreateAPIView):
    serializer_class = serializers.CreateVariableSLZ
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_queryset(self):
        return Variable.objects.filter_by_projects(self.kwargs['project_id'])

    def get_variables_by_search_params(self, search_params):
        variables = self.get_queryset()

        scope = search_params['scope']
        if scope:
            variables = variables.filter(scope=scope)

        search_key = search_params['search_key']
        if search_key:
            variables = variables.filter(key__contains=search_key)

        return variables

    @log_audit(VariableAuditor, activity_type=ActivityType.Add)
    def perform_create(self, serializer):
        self.audit_ctx.update_fields(
            project_id=self.kwargs['project_id'], user=self.request.user.username, description=_("新增变量")
        )
        instance = serializer.save(creator=self.request.user.username)
        # 记录操作日志
        self.audit_ctx.update_fields(resource=instance.name, resource_id=instance.id, extra=serializer.data)

    def post(self, request, project_id):
        """创建变量"""
        if (
            request.data.get("cluster_type") == ClusterType.SHARED
            and request.data["scope"] != serializers.NAMESPACE_SCOPE
        ):
            raise ValidationError(_("共享集群仅允许创建命名空间变量"))
        request.data['project_id'] = project_id
        return super().create(request)

    def list(self, request, project_id):
        """"""
        serializer = serializers.SearchVariableSLZ(data=request.query_params)
        serializer.is_valid(raise_exception=True)
        data = serializer.validated_data

        offset, limit = data['offset'], data['limit']
        variables = self.get_variables_by_search_params(data)
        serializer = serializers.ListVariableSLZ(
            variables[offset : limit + offset],
            many=True,
            context={'search_type': data['type'], 'project_id': project_id},
        )
        num_of_variables = variables.count()
        return Response(
            {
                'count': num_of_variables,
                'has_previous': True if offset != 0 else False,
                'has_next': True if (offset + limit) < num_of_variables else False,
                'results': serializer.data,
            }
        )


class RetrieveUpdateVariableView(FinalizeResponseMixin, generics.RetrieveUpdateDestroyAPIView):
    serializer_class = serializers.UpdateVariableSLZ

    def get_object(self):
        return Variable.objects.get_by_id_with_projects(project_id=self.kwargs['project_id'], id=self.kwargs['pk'])

    @log_audit(VariableAuditor, activity_type=ActivityType.Modify)
    def perform_update(self, serializer):
        self.audit_ctx.update_fields(
            project_id=self.kwargs['project_id'], user=self.request.user.username, description=_("更新变量")
        )
        instance = serializer.save(
            updator=self.request.user.username,
        )
        self.audit_ctx.update_fields(resource=instance.name, resource_id=instance.id, extra=serializer.data)

    # TODO mark refactor 改成put方法, 需要前端同步调整
    def post(self, request, project_id, pk):
        """更新变量"""
        request.data.update({'project_id': project_id, 'id': pk})
        return super().update(request, project_id, pk)


class ResourceVariableView(FinalizeResponseMixin, views.APIView):
    def post(self, request, project_id, version_id):

        project_kind = request.project.kind
        self.slz = VariableNamespaceSLZ(data=request.data, context={'project_kind': project_kind})
        self.slz.is_valid(raise_exception=True)
        slz_data = self.slz.data

        if 'instance_entity' not in slz_data:
            raise ValidationError(_("请选择要实例化的模板"))
        instance_entity = slz_data['instance_entity']

        lb_services = []
        key_list = []
        for cate in instance_entity:
            cate_data = instance_entity[cate]
            cate_id_list = [i.get('id') for i in cate_data if i.get('id')]
            # 查询这些配置文件的变量名
            for _id in cate_id_list:
                if cate == 'metric':
                    continue
                try:
                    resource = MODULE_DICT.get(cate).objects.get(id=_id)
                except Exception:
                    continue
                config = resource.config
                search_list = check_var_by_config(config)
                key_list.extend(search_list)

        key_list = list(set(key_list))
        variable_dict = {}
        if key_list:
            # 验证变量名是否符合规范，不符合抛出异常，否则后续用 django 模板渲染变量也会抛出异常

            var_objects = Variable.objects.filter(Q(project_id=project_id) | Q(project_id=0))

            access_token = request.user.token.access_token
            namespace_res = paas_cc.get_namespace_list(access_token, project_id, limit=LIMIT_FOR_ALL_DATA)
            namespace_data = namespace_res.get('data', {}).get('results') or []
            namespace_dict = {str(i['id']): i['cluster_id'] for i in namespace_data}

            ns_list = slz_data['namespaces'].split(',') if slz_data['namespaces'] else []
            for ns_id in ns_list:
                _v_list = []
                for _key in key_list:
                    key_obj = var_objects.filter(key=_key)
                    if key_obj.exists():
                        _obj = key_obj.first()
                        # 只显示自定义变量
                        if _obj.category == 'custom':
                            cluster_id = namespace_dict.get(ns_id, 0)
                            _v_list.append(
                                {"key": _obj.key, "name": _obj.name, "value": _obj.get_show_value(cluster_id, ns_id)}
                            )
                    else:
                        _v_list.append({"key": _key, "name": _key, "value": ""})
                variable_dict[ns_id] = _v_list
        return Response(
            {"code": 0, "message": "OK", "data": {"lb_services": lb_services, "variable_dict": variable_dict}}
        )


class VariableOverView(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    @log_audit_on_view(VariableAuditor, activity_type=ActivityType.Delete)
    @transaction.atomic
    def batch_delete(self, request, project_id):
        self.slz = serializers.VariableDeleteSLZ(data=request.query_params)
        self.slz.is_valid(raise_exception=True)
        id_list = self.slz.data['id_list']

        query_sets = Variable.objects.filter(project_id=project_id, category='custom', id__in=id_list)
        name_list = query_sets.values_list('name', flat=True)
        del_id_list = list(query_sets.values_list('id', flat=True))

        deled_id_list = []
        for _s in query_sets:
            # 删除后KEY添加 [deleted]前缀
            _del_prefix = f'[deleted_{int(time.time())}]'
            _s.key = f"{_del_prefix}{_s.key}"
            _s.is_deleted = True
            _s.deleted_time = timezone.now()
            _s.save()
            deled_id_list.append(_s.id)

        request.audit_ctx.update_fields(
            resource=','.join(name_list), resource_id=json.dumps(del_id_list), description=_("删除变量")
        )

        return Response({"code": 0, "message": "OK", "data": {"deled_id_list": deled_id_list}})

    def get_quote_info(self, request, project_id, pk):
        qs = Variable.objects.filter((Q(project_id=project_id) | Q(project_id=0)), id=pk)
        if not qs:
            raise ValidationError(u"not found")
        qs = qs.first()
        quote_list = []
        all_template_info = get_all_template_info_by_project(project_id)
        for tem in all_template_info:
            config = tem.get('config') or ''

            key_pattern = re.compile(r'"([^"]+)":\s*"([^"]*{{%s}}[^"]*)"' % qs.key)
            search_list = key_pattern.findall(config)

            for _q in search_list:
                quote_key = _q[0]
                context = _q[1]
                quote_list.append(
                    {
                        "context": context,
                        "quote_location": "%s/%s/%s/%s/%s"
                        % (
                            tem['template_name'],
                            tem['show_version_name'],
                            tem['category_name'],
                            tem['resource_name'],
                            quote_key,
                        ),
                        "key": qs.key,
                        'template_id': tem['template_id'],
                        'template_name': tem['template_name'],
                        'show_version_id': tem['show_version_id'],
                        'category': tem['category'],
                        'resource_id': tem['resource_id'],
                    }
                )

        return Response(
            {
                "code": 0,
                "message": "OK",
                "data": {
                    'quote_list': quote_list,
                    'project_kind': request.project.kind,
                    'project_id': request.project.project_id,
                    'project_code': request.project.english_name,
                },
            }
        )

    def batch_import(self, request, project_id):
        try:
            variables = json.loads(request.data.get('variables'))
        except Exception as e:
            raise ValidationError(str(e))

        serializer = serializers.ImportVariableSLZ(
            data={'variables': variables},
            context={'project_id': project_id, 'access_token': request.user.token.access_token},
        )
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        try:
            import_vars(request.user.username, project_id, validated_data['variables'])
        except Exception as e:
            raise error_codes.APIError(str(e))
        return Response()


class NameSpaceVariableView(viewsets.ViewSet):
    def get_variables(self, request, project_id, ns_id):
        """获取命名空间下所有的变量信息"""
        variables = NameSpaceVariable.get_ns_vars(ns_id, project_id)
        return Response({"code": 0, "message": "OK", "count": len(variables), "data": variables})

    def get_ns_list_by_user_perm(self, request, project_id):
        """获取用户所有有使用权限的命名空间"""
        access_token = request.user.token.access_token
        # 获取全部namespace，前台分页
        result = paas_cc.get_namespace_list(access_token, project_id, with_lb=0, limit=LIMIT_FOR_ALL_DATA)
        if result.get('code') != 0:
            raise error_codes.APIError.f(result.get('message', ''))

        ns_list = result['data']['results'] or []
        if not ns_list:
            return []

        # 补充cluster_name字段
        cluster_ids = [i['cluster_id'] for i in ns_list]
        cluster_list = paas_cc.get_cluster_list(access_token, project_id, cluster_ids).get('data') or []
        cluster_dict = {i['cluster_id']: i for i in cluster_list}
        # 命名空间列表补充集群信息，过来权限时需要
        for i in ns_list:
            i['namespace_id'] = i['id']
            if i['cluster_id'] in cluster_dict:
                i['cluster_name'] = cluster_dict[i['cluster_id']]['name']
                i['environment'] = cluster_dict[i['cluster_id']]['environment']
            else:
                i['cluster_name'] = i['cluster_id']
                i['environment'] = None

        return ns_list

    def get_var_obj(self, project_id, var_id):
        qs = Variable.objects.filter((Q(project_id=project_id) | Q(project_id=0)), id=var_id)
        if not qs:
            raise ValidationError(u"not found")
        qs = qs.first()
        return qs

    def get_batch_variables(self, request, project_id, var_id):
        """查询变量在所有命名空间下的值"""
        qs = self.get_var_obj(project_id, var_id)
        default_value = qs.default_value

        ns_vars = NameSpaceVariable.get_project_ns_vars_by_var(project_id, var_id)
        ns_list = self.get_ns_list_by_user_perm(request, project_id)
        for _n in ns_list:
            if _n['id'] in ns_vars:
                _n['variable_value'] = ns_vars.get(_n['id'])
            else:
                _n['variable_value'] = default_value

        return Response({"code": 0, "message": "OK", "data": ns_list})

    def save_batch_variables(self, request, project_id, var_id):
        """批量保存
        针对一个变量，保存所有命名空间上的值
        """
        self.slz = serializers.NsVariableSLZ(data=request.data)
        self.slz.is_valid(raise_exception=True)
        qs = self.get_var_obj(project_id, var_id)
        NameSpaceVariable.batch_save_by_var_id(qs, self.slz.data['ns_vars'])
        return Response({"code": 0, "message": 'ok', "data": []})


class ClusterVariableView(viewsets.ViewSet):
    def get_variables(self, request, project_id, cluster_id):
        """获取集群下所有的变量信息"""
        variables = ClusterVariable.get_cluster_vars(cluster_id, project_id)
        return Response({"code": 0, "message": "OK", "count": len(variables), "data": variables})

    def batch_save(self, request, project_id, cluster_id):
        """批量保存集群变量
        针对一个命名空间，保存所有的变量
        """
        self.slz = serializers.ClusterVariableSLZ(data=request.data)
        self.slz.is_valid(raise_exception=True)
        res, not_exist_vars = ClusterVariable.batch_save(cluster_id, self.slz.data['cluster_vars'])
        msg = 'OK'
        if not_exist_vars:
            not_exist_show_msg = ['%s[id:%s]' % (i['key'], i['id']) for i in not_exist_vars]
            msg = _('以下变量不存在:{}').format(';'.join(not_exist_show_msg))
        return Response({"code": 0, "message": msg, "data": not_exist_vars})

    def get_var_obj(self, project_id, var_id):
        qs = Variable.objects.filter((Q(project_id=project_id) | Q(project_id=0)), id=var_id)
        if not qs:
            raise ValidationError(u"not found")
        qs = qs.first()
        return qs

    def get_cluser_list_by_user_perm(self, request, project_id):
        """
        获取项目下所有集群(对接 iam v3 时去除了权限控制)
        """
        cluster_data = paas_cc.get_all_clusters(request.user.token.access_token, project_id).get('data') or {}
        return cluster_data.get('results') or []

    def get_batch_variables(self, request, project_id, var_id):
        """查询变量在所有集群下的值"""
        qs = self.get_var_obj(project_id, var_id)
        default_value = qs.default_value

        cluser_vars = ClusterVariable.get_project_cluster_vars_by_var(project_id, var_id)
        cluster_list = self.get_cluser_list_by_user_perm(request, project_id)
        for _n in cluster_list:
            if _n['cluster_id'] in cluser_vars:
                _n['variable_value'] = cluser_vars.get(_n['cluster_id'])
            else:
                _n['variable_value'] = default_value

        return Response({"code": 0, "message": "OK", "data": cluster_list})

    def save_batch_variables(self, request, project_id, var_id):
        """批量保存
        针对一个变量，保存所有命名空间上的值
        """
        self.slz = serializers.ClusterVariableSLZ(data=request.data)
        self.slz.is_valid(raise_exception=True)
        qs = self.get_var_obj(project_id, var_id)
        ClusterVariable.batch_save_by_var_id(qs, self.slz.data['cluster_vars'])
        return Response({"code": 0, "message": 'ok', "data": []})
