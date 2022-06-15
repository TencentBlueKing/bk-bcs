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
import re
import tempfile
from typing import Dict

import dpath
import yaml
import yaml.reader
from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from ruamel.yaml import YAML
from ruamel.yaml.compat import StringIO, ordereddict

from backend.components import paas_cc
from backend.helm.helm.utils.util import EmptyVaue, fix_rancher_value_by_type
from backend.helm.toolkit.diff import parser
from backend.resources.utils.kube_client import get_dynamic_client
from backend.utils.basic import get_bcs_component_version
from backend.utils.client import make_dashboard_ctl_client

from .constants import DASHBOARD_CTL_VERSION, DEFAULT_DASHBOARD_CTL_VERSION

yaml.reader.Reader.NON_PRINTABLE = re.compile(
    '[^\x09\x0A\x0D\x20-\x7E\x85\xA0-\uD7FF\uE000-\uFFFD\U00010000-\U0010FFFF]'
)

logger = logging.getLogger(__name__)


def represent_none(self, _):
    return self.represent_scalar('tag:yaml.org,2002:null', '')


yaml.add_representer(type(None), represent_none)


def ruamel_yaml_load(content):
    # be carefule, ruamel.yaml doesn't work well with dpath
    yaml = YAML()
    # 添加 preserve_quotes=True, 避免 json 转换 yaml 时，丢掉双引号
    yaml.preserve_quotes = True
    return yaml.load(content)


def ruamel_yaml_dump(yaml_obj):
    # be carefule, ruamel.yaml doesn't work well with dpath
    yaml = YAML()
    # 添加 preserve_quotes=True, 避免 json 转换 yaml 时，丢掉双引号
    yaml.preserve_quotes = True
    stream = StringIO()
    yaml.dump(yaml_obj, stream=stream)
    content = stream.getvalue()
    return content


def yaml_load(content):
    return yaml.load(content)


def yaml_dump(obj):
    # 添加 presenter, 避免 json 转换 yaml 时，丢掉双引号
    def literal_presenter(dumper, data):
        if isinstance(data, str) and "\n" in data:
            return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='|')
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='"')

    noalias_dumper = yaml.dumper.SafeDumper
    noalias_dumper.ignore_aliases = lambda self, data: True
    noalias_dumper.add_representer(str, literal_presenter)

    return yaml.dump(obj, default_flow_style=False, Dumper=noalias_dumper)


def sync_dict2yaml(obj_list, yaml_content):
    """
    根据 obj_list 的内容更新 yaml_content
    note: 使用 ruamel.yaml 可以保证 yaml_content 可以在load与dump之后还保持注释内容不丢失
    example
    parameters: obj_list
    [
        {
            "name": "aa",
            "type": "int",
            "value": "1"
        },
        {
            "name": "b.c.e",
            "type": "str",
            "value": "3"
        },
        {
            "name": "dd",
            "type": "int",
            "value": 0
        }
    ]
    parameters: yaml_content
    aa: 2
    xx: 1
    b:
      c:
        e: 4
    result:
    aa: 1
    xx: 1
    dd: 0
    b:
      c:
        e: "3"
    """
    yaml_obj = ruamel_yaml_load(yaml_content)
    for item in obj_list:
        try:
            value = fix_rancher_value_by_type(item["value"], item["type"])
        except EmptyVaue:
            continue
        else:
            update = dict()
            dpath.util.new(update, item["name"], value, separator=".")
            dpath.util.merge(
                dst=yaml_obj, src=update, separator=".", flags=dpath.util.MERGE_REPLACE | dpath.util.MERGE_ADDITIVE
            )

    content = ruamel_yaml_dump(yaml_obj)
    return content


def sync_yaml2dict(obj_list, yaml_content):
    """
    根据 yaml_content 的内容更新 obj_list
    parameters: obj_list
    [
        {
            "name": "aa",
            "type": "int",
            "value": "1"
        },
        {
            "name": "b.c.e",
            "type": "str",
            "value": "3"
        },
        {
            "name": "dd",
            "type": "int",
            "value": 0
        }
    ]
    parameters: yaml_content
    aa: 2
    xx: 1
    b:
      c:
        e: 4
    result:
    [
        {
            "name": "aa",
            "type": "int",
            "value": 2
        },
        {
            "name": "b.c.e",
            "type": "str",
            "value": "4"
        },
        {
            "name": "dd",
            "type": "int",
            "value": 0
        }
    ]
    """
    yaml_obj = ruamel_yaml_load(yaml_content)
    for idx, item in enumerate(obj_list):
        try:
            value = dpath.util.get(yaml_obj, item["name"], separator=".")
        except KeyError:
            pass
        else:
            try:
                value = fix_rancher_value_by_type(value, item["type"])
            except EmptyVaue:
                continue
            else:
                obj_list[idx]["value"] = value

    return obj_list


def safe_get(data, key, default):
    try:
        return int(dpath.util.get(data, key, separator="."))
    except (KeyError, ValueError):
        return default


def collect_resource_state(kube_client, namespace, content):
    state_keys = ["replicas", "readyReplicas", "availableReplicas", "updatedReplicas"]

    def extract_state_info(data):
        return {key: safe_get(data, "status.%s" % key, 0) for key in state_keys}

    with tempfile.NamedTemporaryFile("w") as f:
        f.write(content)
        f.flush()

        res = kube_client.get_by_file(filename=f.name, namespace=namespace)

    result = {"summary": {}, "items": []}
    for item in res["items"]:
        state = extract_state_info(item)
        result["items"].append(state)

    for key in state_keys:
        result["summary"][key] = sum([x[key] for x in result["items"]])

    return result


def merge_valuefile(source, new):
    source = ruamel_yaml_load(source)
    if not source:
        source = ordereddict()

    new = ruamel_yaml_load(new)
    dpath.util.merge(source, new)
    return ruamel_yaml_dump(source)


def dashboard_get_overview(kubeconfig, namespace, bin_path=settings.DASHBOARD_CTL_BIN):
    dashboard_client = make_dashboard_ctl_client(kubeconfig=kubeconfig, bin_path=bin_path)
    dashboard_overview = dashboard_client.overview(namespace=namespace, parameters=dict())
    return dashboard_overview


def update_pods_status(resource: Dict) -> Dict:
    """添加pods的状态信息"""
    # TODO：优化调整为通过helm和kubectl获取状态，现阶段先兼容处理
    # 新版本的 dashboard，返回的部分资源的 pods 字段调整为了 podInfo 字段
    if ("pods" not in resource) and ("podInfo" in resource):
        resource["pods"] = resource["podInfo"]
        return resource

    return resource


def extract_state_info_from_dashboard_overview(overview_status, kind, namespace, name):
    for key in overview_status.keys():
        if key.lower() != "{kind}list".format(kind=kind).lower():
            continue

        for k in overview_status[key].keys():
            if k.lower() == "{kind}s".format(kind=kind).lower():
                obj_list = overview_status[key][k]
                break
        else:
            if "items" in overview_status[key]:
                obj_list = overview_status[key]["items"]

        for item in obj_list:
            if item["objectMeta"]["name"].lower() == name.lower():
                item = update_pods_status(item)
                return item

    return dict()


def collect_resource_status(kubeconfig, app, project_code, cluster_id, bin_path=settings.DASHBOARD_CTL_BIN):
    """
    dashboard_client = make_dashboard_ctl_client(
        kubeconfig=kubeconfig
    )
    """

    def status_sumary(status, app, bin_path=settings.DASHBOARD_CTL_BIN):
        if not status and not app.transitioning_result:
            return {
                "messages": _("未找到资源，可能未部署成功，请在Helm Release列表也查看失败原因."),
                "is_normal": False,
                "desired_pods": "-",
                "ready_pods": "-",
            }

        # 暂未实现该类资源状态信息
        if "pods" not in status:
            return {
                "messages": "",
                "is_normal": True,
                "desired_pods": "-",
                "ready_pods": "-",
            }

        messages = [item["message"] for item in status["pods"]["warnings"]]
        messages = filter(lambda x: x, messages)

        desired_pods = safe_get(status, "pods.desired", None)
        ready_pods = safe_get(status, "pods.running", None)
        data = {
            "desired_pods": str(desired_pods),
            "ready_pods": str(ready_pods),
            "messages": "\n".join(messages),
            "is_normal": desired_pods == ready_pods,
        }
        return data

    namespace = app.namespace
    content = app.release.content
    resources = parser.parse(content, app.namespace)
    resources = resources.values()
    release_name = app.name

    dashboard_overview = dashboard_get_overview(kubeconfig=kubeconfig, namespace=namespace, bin_path=bin_path)

    result = {}
    structure = app.release.extract_structure(namespace)
    for item in structure:
        kind = item["kind"]
        name = item["name"]

        status = extract_state_info_from_dashboard_overview(
            overview_status=dashboard_overview, kind=kind, namespace=namespace, name=name
        )
        """
        status = {}
        if kind.lower() in ["deployment", "replicaset", "daemonset",
                            "job", "statefulset", "cronjob", "replicationcontroller"]:
            try:
                status = dashboard_client.workload_status(
                    kind=kind,
                    name=name,
                    namespace=namespace,
                    parameters=dict()
                )
            except DashboardExecutionError as e:
                if "handler returned wrong status code: got 404 want 200" in e.output:
                    pass
                else:
                    raise
        """

        key = "{kind}/{namespace}/{name}".format(
            name=name,
            namespace=namespace,
            kind=kind,
        )
        result[key] = {
            "namespace": namespace,
            "name": name,
            "kind": kind,
            "cluster_id": cluster_id,
            "status": status,
            "status_sumary": status_sumary(status, app),
        }
    return result


def resource_link(kind, project_code, name, namespace, release_name):
    kind_map = {
        "deployment": "deployments",
        "statefulset": "statefulset",
        "daemonset": "daemonset",
        "job": "job",
    }
    kind = kind.lower()
    if kind not in kind_map:
        return None

    fix_kind = kind_map[kind]
    url = (
        f"{settings.SITE_URL.rstrip('/')}/{project_code}/app/{fix_kind}/{name}/{namespace}/{kind}"
        "?name={name}&namespace={namespace}&category={kind}"
    )
    return url


def compose_url_with_scheme(url, scheme="http"):
    """组装URL"""
    url_split_info = url.split('//')
    return '{scheme}://{domain}'.format(scheme=scheme, domain=url_split_info[-1])


def get_cc_app_id(access_token, project_id):
    resp = paas_cc.get_project(access_token, project_id)
    project_info = resp.get("data") or {}
    return str(project_info.get("cc_app_id") or "")


def get_helm_dashboard_path(access_token: str, project_id: str, cluster_id: str) -> str:
    """获取dashboard的路径"""
    client = get_dynamic_client(access_token, project_id, cluster_id)
    # 获取版本
    version = get_bcs_component_version(
        client.version["kubernetes"]["gitVersion"], DASHBOARD_CTL_VERSION, DEFAULT_DASHBOARD_CTL_VERSION
    )

    bin_path_map = getattr(settings, "DASHBOARD_CTL_VERSION_MAP", {})
    return bin_path_map.get(version, settings.DASHBOARD_CTL_BIN)


def remove_updater_creator_from_manifest(manifest: str) -> str:
    """删除manifest中的添加的平台注入的updater和creator

    :param manifest: 资源的yaml内容
    :return: 返回移除updater和creator后的内容
    """
    stream = StringIO(manifest)
    refine_stream = StringIO()
    for l in stream.readlines():
        if ("io.tencent.paas.creator" in l) or ("io.tencent.paas.updator" in l):
            continue
        refine_stream.write(l)
    return refine_stream.getvalue()
