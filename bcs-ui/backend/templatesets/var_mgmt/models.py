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

变量管理功能
"""
import json
import logging

from django.db import models
from django.utils.translation import ugettext_lazy as _

from backend.templatesets.legacy_apps.configuration.models import BaseModel

from .constants import ALL_PROJECTS, VariableCategory, VariableScope

logger = logging.getLogger(__name__)


class VariableManager(models.Manager):
    """Manager for Variable"""

    def get_queryset(self):
        return super().get_queryset().filter(is_deleted=False)

    def filter_by_projects(self, project_id):
        return self.filter(project_id__in=[project_id, ALL_PROJECTS])

    def get_by_id_with_projects(self, project_id, id):
        return self.filter_by_projects(project_id).get(id=id)


class Variable(BaseModel):
    project_id = models.CharField(_("项目ID"), max_length=32, help_text=_("0:表示对所有项目生效"))
    key = models.CharField("KEY", max_length=255)
    name = models.CharField(_("名称"), max_length=255)
    # TODO mark refactor default使用JSONField会更合适，暂时保留TextField
    default = models.TextField(_("默认值"), help_text=_("以{'value':'默认值'}格式存储默认值,可以存储字符串和数字"))
    desc = models.TextField(_("说明"), blank=True, null=True)
    category = models.CharField(
        _("类型"), max_length=16, choices=VariableCategory.get_choices(), default=VariableCategory.CUSTOM.value
    )
    scope = models.CharField(_("作用范围"), max_length=16, choices=VariableScope.get_choices())

    objects = VariableManager()
    default_objects = models.Manager()

    class Meta:
        ordering = ('category', '-id')

    def save(self, *args, **kwargs):
        if isinstance(self.default, dict):
            self.default = json.dumps(self.default)
        super().save(*args, **kwargs)

    def get_default_data(self) -> dict:
        if not self.default:
            return {}
        try:
            default = json.loads(self.default)
        except Exception:
            logger.exception("变量默认值格式错误")
            return {}
        return default

    @property
    def default_value(self):
        return self.get_default_data().get('value', '')

    def get_show_value(self, cluster_id, ns_id):
        # TODO refactor
        if self.scope == VariableScope.CLUSTER.value:
            # if self.scope == 'cluster':
            cluster_vars = ClusterVariable.objects.filter(cluster_id=cluster_id, var_id=self.id)
            if cluster_vars.exists():
                _var = cluster_vars.first()
                return _var.get_value
        if self.scope == 'namespace':
            ns_vars = NameSpaceVariable.objects.filter(ns_id=ns_id, var_id=self.id)
            if ns_vars.exists():
                _var = ns_vars.first()
                return _var.get_value
        return self.default_value


class ClusterVariable(BaseModel):
    """集群变量"""

    var_id = models.IntegerField(_("变量ID"))
    cluster_id = models.CharField(_("集群ID"), max_length=64)
    data = models.TextField(_("值"), help_text=_("以{'value':'值'}格式存储默认值,可以存储字符串和数字"))

    class Meta:
        ordering = ('-id',)
        unique_together = ("var_id", "cluster_id")

    @property
    def get_variable(self):
        try:
            variable = Variable.objects.get(id=self.var_id)
        except Exception:
            return None
        return variable

    @property
    def get_value(self):
        try:
            data = json.loads(self.data)
        except Exception:
            logger.exception("变量值格式错误")
            return None
        return data.get('value')

    @classmethod
    def batch_save(cls, cluster_id, cluster_vars):
        """批量保存变量值"""
        not_exist_vars = []
        for _v in cluster_vars:
            try:
                Variable.objects.get(id=_v['id'])
            except Exception:
                not_exist_vars.append(_v)
            else:
                defaults = {'data': json.dumps({'value': _v.get('value')})}
                cls.objects.update_or_create(cluster_id=cluster_id, var_id=_v['id'], defaults=defaults)
        res = False if not_exist_vars else True
        return res, not_exist_vars

    @classmethod
    def get_cluster_vars(cls, cluster_id, project_id):
        """获取命名空间下的变量列表"""
        # 从变量模板中获取所有的变量及其默认值
        ns_vars = Variable.objects.filter(project_id=project_id, scope='cluster')

        var_data = []
        for _v in ns_vars:
            # 查询变量是否又在 cluster 在赋值
            ns_vars = ClusterVariable.objects.filter(var_id=_v.id, cluster_id=cluster_id)
            _ns_value = None
            if ns_vars.exists():
                _ns_var = ns_vars.first()
                _ns_value = _ns_var.get_value
            # 没有在集群中单独对变量赋值，则使用默认值
            else:
                _ns_value = _v.default_value
            var_data.append({'id': _v.id, 'key': _v.key, 'name': _v.name, 'value': _ns_value if _ns_value else ''})
        return var_data

    # 支持针对单个变量批量编辑在所有集群的值
    @classmethod
    def get_project_cluster_vars_by_var(cls, project_id, var_id):
        """支持针对单个变量批量编辑在所有集群的值"""
        cluster_vars = cls.objects.filter(var_id=var_id)
        cluster_values = {}
        for _n in cluster_vars:
            cluster_values[_n.cluster_id] = _n.get_value
        return cluster_values

    @classmethod
    def batch_save_by_var_id(cls, var_obj, var_dict):
        """批量保存
        针对一个变量，保存所有命名空间上的值
        """
        var_id = var_obj.id
        for cluster_id in var_dict:
            defaults = {'data': json.dumps({'value': var_dict.get(cluster_id)})}
            cls.objects.update_or_create(cluster_id=cluster_id, var_id=var_id, defaults=defaults)


class NameSpaceVariable(BaseModel):
    """命名空间变量"""

    var_id = models.IntegerField(_("变量ID"))
    ns_id = models.IntegerField(_("命名空间ID"))
    data = models.TextField(_("值"), help_text=_("以{'value':'值'}格式存储默认值,可以存储字符串和数字"))

    class Meta:
        ordering = ('-id',)
        unique_together = ("var_id", "ns_id")

    @property
    def get_variable(self):
        try:
            variable = Variable.objects.get(id=self.var_id)
        except Exception:
            return None
        return variable

    @property
    def get_value(self):
        try:
            data = json.loads(self.data)
        except Exception:
            logger.exception("变量值格式错误")
            return None
        return data.get('value')

    @classmethod
    def batch_save(cls, ns_id, ns_vars):
        """批量保存变量值
        针对一个命名空间，保存所有的变量
        """
        not_exist_vars = []
        for _v in ns_vars:
            try:
                Variable.objects.get(id=_v['id'])
            except Exception:
                not_exist_vars.append(_v)
            else:
                defaults = {'data': json.dumps({'value': _v.get('value')})}
                cls.objects.update_or_create(ns_id=ns_id, var_id=_v['id'], defaults=defaults)
        res = False if not_exist_vars else True
        return res, not_exist_vars

    @classmethod
    def get_ns_vars(cls, ns_id, project_id):
        """获取命名空间下的变量列表"""
        # 从变量模板中获取所有的变量及其默认值
        ns_vars = Variable.objects.filter(project_id=project_id, scope='namespace')

        var_data = []
        for _v in ns_vars:
            # 查询变量是否又在 ns 在赋值
            ns_vars = NameSpaceVariable.objects.filter(var_id=_v.id, ns_id=ns_id)
            _ns_value = None
            if ns_vars.exists():
                _ns_var = ns_vars.first()
                _ns_value = _ns_var.get_value
            # 没有在命名空间中单独对变量赋值，则使用默认值
            else:
                _ns_value = _v.default_value
            var_data.append({'id': _v.id, 'key': _v.key, 'name': _v.name, 'value': _ns_value if _ns_value else ''})
        return var_data

    @classmethod
    def get_project_ns_vars(cls, project_id):
        """获取命名空间下的变量列表"""
        # 从变量模板中获取所有的变量及其默认值
        ns_vars = Variable.objects.filter(project_id=project_id, scope='namespace')

        var_data = []
        for _v in ns_vars:
            # 查询在命名空间重新赋值的变量
            ns_vars = NameSpaceVariable.objects.filter(var_id=_v.id)
            ns_values = {}
            for _ns in ns_vars:
                ns_values[_ns.ns_id] = _ns.get_value

            var_data.append(
                {
                    'id': _v.id,
                    'key': _v.key,
                    'name': _v.name,
                    'default_value': _v.default_value,
                    'ns_values': ns_values,
                }
            )
        return var_data

    # 支持针对单个变量批量编辑在所有命名空间的值
    @classmethod
    def get_project_ns_vars_by_var(cls, project_id, var_id):
        """支持针对单个变量批量编辑在所有命名空间的值"""
        ns_vars = cls.objects.filter(var_id=var_id)
        ns_values = {}
        for _ns in ns_vars:
            ns_values[_ns.ns_id] = _ns.get_value
        return ns_values

    @classmethod
    def batch_save_by_var_id(cls, var_obj, var_dict):
        """批量保存
        针对一个变量，保存所有命名空间上的值
        """
        var_id = var_obj.id
        for ns_id in var_dict:
            defaults = {'data': json.dumps({'value': var_dict.get(ns_id)})}
            cls.objects.update_or_create(ns_id=ns_id, var_id=var_id, defaults=defaults)
