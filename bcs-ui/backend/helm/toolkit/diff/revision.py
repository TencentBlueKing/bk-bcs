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

from . import diff, parser

"""
a simple wrapper for compare two release,
reference <https://github.com/databus23/helm-diff/blob/master/cmd/revision.go>
"""


class AppRevisionDiffer:
    """get difference of two release
    suppressed_kinds:  allows suppression of the values listed in the diff output
    output_context: output NUM lines of context around changes
    """

    def __init__(self, app, revisions, suppressed_kinds: list, output_context: int = -1):
        self.app = app
        self.revisions = revisions
        self.suppressed_kinds = suppressed_kinds or []
        self.output_context = output_context

    def get_release_content(self, release):
        if release.content:
            return release.content, ""

        content, notes = release.render(namespace=self.app.namespace)
        return content, notes

    def differentiate(self):
        revisions_len = len(self.revisions)
        output = io.StringIO()
        if revisions_len == 1:
            revision1, revision2 = self.app.release, self.revisions[0]
        elif revisions_len == 2:
            # compare two release
            revision1, revision2 = self.revisions[0], self.revisions[1]
            if revision1.id > revision2.id:
                revision1, revision2 = revision2, revision1
        else:
            raise ValueError(revisions_len)

        revision1_content, _ = self.get_release_content(revision1)
        revision2_content, _ = self.get_release_content(revision2)

        diff.diff_manifests(
            parser.parse(revision1_content, self.app.namespace),
            parser.parse(revision2_content, self.app.namespace),
            self.suppressed_kinds,
            self.output_context,
            output,
        )

        return output.getvalue()
