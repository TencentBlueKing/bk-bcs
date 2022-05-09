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

from rest_framework.exceptions import ParseError

from backend.kube_core.toolkit.injectors import InjectManager

from ..toolkit.diff.parser import split_manifest
from .bcs_info_provider import BcsInfoProvider
from .utils import yaml_dump, yaml_load

yaml_seperator = b"\n---\n"
logger = logging.getLogger(__name__)


def make_matcher_configs(matcher_cls, kinds):
    return [{"type": matcher_cls, "parameters": {"kind": kind}} for kind in kinds]


def make_kind_matcher_configs(kinds):
    return make_matcher_configs("KindMatcher", kinds)


def make_re_kind_matcher_configs(kinds):
    return make_matcher_configs("ReKindMatcher", kinds)


def parse_manifest(manifest):
    if not isinstance(manifest, bytes):
        if isinstance(manifest, str):
            manifest = manifest.encode("utf-8")
        else:
            manifest = bytes(manifest, "utf-8")

    result = list()
    contents = split_manifest(manifest)
    for content in contents:
        content = content.replace(b'\t\n', b'\n')
        content = content.strip(b'\t')
        try:
            resource = yaml_load(content)
        # except yaml.composer.ComposerError as e:
        except Exception as e:
            message = "Parse manifest failed: \n{error}\n\nManifest content:\n{content}".format(
                error=e, content=content.decode("utf-8")
            )
            logger.exception(message)
            raise ParseError(message)

        if not resource:
            continue

        result.append(resource)
    return result


def join_manifest(resources_list):
    resources_list = [yaml_dump(resource) for resource in resources_list]
    return yaml_seperator.decode().join(resources_list)


def inject_configs(
    access_token,
    project_id,
    cluster_id,
    namespace_id,
    namespace,
    creator,
    updator,
    created_at,
    updated_at,
    version,
    ignore_empty_access_token=False,
    extra_inject_source=None,
    source_type='helm',
):
    if extra_inject_source is None:
        extra_inject_source = dict()

    context = {"creator": creator, "updator": updator, "version": version}
    context.update(extra_inject_source)

    provider = BcsInfoProvider(
        access_token=access_token,
        project_id=project_id,
        cluster_id=cluster_id,
        namespace_id=namespace_id,
        namespace=namespace,
        context=context,
        ignore_empty_access_token=True,
    )

    bcs_annotations = provider.provide_annotations(source_type)
    # resouce may not have annotations field
    bcs_annotations = {"annotations": bcs_annotations}

    bcs_labels = provider.provide_labels(source_type)
    bcs_labels = {"labels": bcs_labels}

    bcs_pod_labels = {"labels": provider.provide_pod_labels(source_type)}
    # Some pod may has no env config, so we shouldn't add `env` to path,
    # Add it to be injected data, make sure it will merge to pod's env anyway.
    bcs_env = {"env": provider.provide_container_env()}

    configs = [
        {
            # annotations
            "matchers": make_re_kind_matcher_configs([".+"]),
            "paths": ["/metadata"],
            "data": bcs_annotations,
            "force_str": True,
        },
        {
            # pod labels
            "matchers": make_re_kind_matcher_configs([".+"]),
            "paths": ["/metadata"],
            "data": bcs_labels,
            "force_str": True,
        },
        {
            # pod labels
            "matchers": make_kind_matcher_configs(["Deployment", "StatefulSet", "DaemonSet", "ReplicaSet", "Job"]),
            "paths": ["/spec/template/metadata"],
            "data": bcs_pod_labels,
            "force_str": True,
        },
        {
            # pod env
            "matchers": make_kind_matcher_configs(["Pod"]),
            "paths": ["/spec/containers/*"],
            "data": bcs_env,
            "force_str": True,
        },
        {
            # pod env
            "matchers": make_kind_matcher_configs(["Deployment", "StatefulSet", "DaemonSet", "ReplicaSet", "Job"]),
            "paths": ["/spec/template/spec/containers/*"],
            "data": bcs_env,
            "force_str": True,
        },
    ]

    return configs


def inject_bcs_info(
    access_token,
    project_id,
    cluster_id,
    namespace_id,
    namespace,
    creator,
    updator,
    created_at,
    updated_at,
    resources,
    version,
    ignore_empty_access_token=False,
    extra_inject_source=None,
):
    configs = inject_configs(
        access_token=access_token,
        project_id=project_id,
        cluster_id=cluster_id,
        namespace_id=namespace_id,
        namespace=namespace,
        creator=creator,
        updator=updator,
        created_at=created_at,
        updated_at=updated_at,
        version=version,
        ignore_empty_access_token=ignore_empty_access_token,
        extra_inject_source=extra_inject_source,
    )
    resources_list = parse_manifest(resources)
    context = {"creator": creator, "updator": updator, "version": version}
    manager = InjectManager(configs=configs, resources=resources_list, context=context)
    resources_list = manager.do_inject()
    content = join_manifest(resources_list)
    return content
