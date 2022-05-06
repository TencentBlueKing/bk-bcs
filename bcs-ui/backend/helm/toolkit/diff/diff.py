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
import io
import logging
import sys
from difflib import Differ

from . import parser

logger = logging.getLogger(__name__)

"""
reference <https://github.com/databus23/helm-diff/blob/master/diff/diff.go>
"""


def diff_manifests(old_index, new_index, suppressed_kinds, context, to):
    for key, old_content in old_index.items():
        if key in new_index:
            new_content = new_index[key]
            if old_content.content == new_content.content:
                continue

            # modified
            to.write("%s has changed:\n" % key)
            print_diff(suppressed_kinds, new_content.kind, context, old_content.content, new_content.content, to)
            to.write("\n")
        else:
            # removed
            to.write("%s has been removed:\n" % key)
            print_diff(suppressed_kinds, old_content.kind, context, old_content.content, b"", to)
            to.write("\n")

    for key, new_content in new_index.items():
        if key in old_index:
            continue

        # added
        to.write("%s has been added:\n" % key)
        print_diff(suppressed_kinds, new_content.kind, context, b"", new_content.content, to)
        to.write("\n")


def print_diff(suppressed_kinds, kind, context, before, after, to):
    d = Differ()
    diffs = "".join(d.compare(before.decode("utf8").splitlines(True), after.decode("utf8").splitlines(True)))
    diffs = diffs.splitlines(True)
    logger.debug("".join(diffs))

    if kind in suppressed_kinds:
        string = "+ Changes suppressed on sensitive content of type %s\n" % kind
        to.write(string)
        return

    if context >= 0:
        distances = calculate_distances(diffs)
        omitting = False
        for i, diff in enumerate(diffs):
            if distances[i] > context:
                if not omitting:
                    to.write("...")
                    omitting = True
            else:
                omitting = False
                print_diff_record(diff, to)
        return

    for diff in diffs:
        print_diff_record(diff, to)


def print_diff_record(diff, to):
    if isinstance(diff, bytes):
        diff = diff.decode("utf8")
    to.write(diff)


def calculate_distances(diffs):
    """Calculate distance of every diff-line to the closest change"""
    distances = dict()

    # Iterate forwards through diffs, set 'distance' based on closest 'change' before this line
    change = -1
    for i, diff in enumerate(diffs):
        if diff and diff[0] != " ":
            change = i
        distance = sys.maxsize
        if change != -1:
            distance = i - change
        distances[i] = distance

    # Iterate backwards through diffs, reduce 'distance' based on closest 'change' after this line
    change = -1
    for i in range(len(diffs) - 1, 0, -1):
        diff = diffs[i]
        if diff and diff[0] != " ":
            change = i

        if change != -1:
            distance = change - i
            if distance < distances[i]:
                distances[i] = distance

    return distances


def simple_diff(content_old, content_new, namespace):
    output = io.StringIO()
    diff_manifests(
        old_index=parser.parse(content_old, namespace),
        new_index=parser.parse(content_new, namespace),
        suppressed_kinds=[-1],
        context=-1,
        to=output,
    )

    differnece = output.getvalue()
    return differnece
