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

工具：mozilla/bleach
白名单参考：内部版 xss filter
测试用例：
    https://www.owasp.org/index.php/XSS_Filter_Evasion_Cheat_Sheet
参考文档：
    [方案对比](https://stackoverflow.com/questions/699468/python-html-sanitizer-scrubber-filter)
    [Bleach](http://bleach.readthedocs.io/en/latest/clean.html)
    [a标签_blank漏洞](https://blog.yongyuan.us/articles/2016-08-16-target-blank/)
    [a标签_blank漏洞](http://www.webhek.com/post/the-targetblank-vulnerability-by-example.html)
前端过滤（可选方案）：
    [JS-XSS](https://github.com/leizongmin/js-xss)
"""
import copy

from bleach.encoding import force_unicode
from bleach.sanitizer import BleachSanitizerFilter, Cleaner

allow_tags = [
    'a',
    'img',
    'br',
    'strong',
    'b',
    'code',
    'pre',
    'p',
    'div',
    'em',
    'span',
    'h1',
    'h2',
    'h3',
    'h4',
    'h5',
    'h6',
    'blockquote',
    'ul',
    'ol',
    'tr',
    'th',
    'td',
    'hr',
    'li',
    'u',
    'embed',
    's',
    'table',
    'thead',
    'tbody',
    'caption',
    'small',
    'q',
    'sup',
    'sub',
]
common_attrs = ["id", "style", "class", "name"]
nonend_tags = ["img", "hr", "br", "embed"]
tags_own_attrs = {
    "img": ["src", "width", "height", "alt", "align"],
    "a": ["href", "target", "rel", "title"],
    "embed": ["src", "width", "height", "type", "allowfullscreen", "loop", "play", "wmode", "menu"],
    "table": ["border", "cellpadding", "cellspacing"],
}


class BkBleachSanitizerFilter(BleachSanitizerFilter):
    @staticmethod
    def limit_attribute_value(data, attribute, limit=None):
        limit = limit if limit is not None else []
        if attribute not in data:
            return data
        elif data.get(attribute) not in limit:
            data.pop(attribute)
        return data

    def allow_token_a(self, token):
        if u"data" in token and (None, u'href') in token[u'data']:
            token[u"data"].setdefault((None, u"target"), u"_blank")
            token[u"data"][(None, u"rel")] = u"noopener noreferrer"
            self.limit_attribute_value(token[u'data'], (None, u'target'), [u'_blank', u'_self'])
        return token

    def allow_token(self, token):
        result = super(BkBleachSanitizerFilter, self).allow_token(token)
        allow_token_hook = getattr(self, "allow_token_%s" % token["name"], None)
        if allow_token_hook is not None:
            return allow_token_hook(result)
        return result


class BkCleaner(Cleaner):
    def clean(self, text):
        """Cleans text and returns sanitized result as unicode

        :arg str text: text to be cleaned

        :returns: sanitized text as unicode

        """
        if not text:
            return u''

        text = force_unicode(text)

        dom = self.parser.parseFragment(text)
        filtered = BkBleachSanitizerFilter(
            source=self.walker(dom),
            # Bleach-sanitizer-specific things
            attributes=self.attributes,
            strip_disallowed_elements=self.strip,
            strip_html_comments=self.strip_comments,
            # html5lib-sanitizer things
            allowed_elements=self.tags,
            allowed_css_properties=self.styles,
            allowed_protocols=self.protocols,
            allowed_svg_properties=[],
        )

        # Apply any filters after the BleachSanitizerFilter
        for filter_class in self.filters:
            filtered = filter_class(source=filtered)

        return self.serializer.render(filtered)


def clean_html(text):
    """采用bleach+内部版白名单方式"""
    attributes = {tag: copy.deepcopy(common_attrs) for tag in allow_tags}
    for tag, attr in tags_own_attrs.iteritems():
        attributes[tag] += attr
    cleaner = BkCleaner(tags=allow_tags, attributes=attributes)
    return cleaner.clean(text)


def test():
    text = """
    <html>
      <body>
        <div>
          <style>/* deleted */</style>
          <a href="">a link</a>
          <a href="#">another link</a>
          <p>a paragraph</p>
          <div>secret EVIL!</div>
          of EVIL!
          Password:
          annoying EVIL!
          <a href="evil-site">spam spam SPAM!</a>
          <img src="evil!">
          <<script></script>script> alert("Haha, I hacked your page."); <<script></script>/script>
          <div on<script></script>load="alert("Haha, I hacked your page.");
          <script>bad</script><script>bad</script><script>bad</script>
          <a href="good" onload="bad" onclick="bad" alt="good">1</div>
        </div>
      </body>
    </html>
    """
    result = clean_html(text)
    return result
