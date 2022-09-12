/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/
import Vue from 'vue';
import VueI18n from 'vue-i18n';
import { locale, lang } from 'bk-magic-vue';
import cookie from 'cookie';
import langMap from './lang';

Vue.use(VueI18n);

const en = {};
const cn = {};
Object.keys(langMap).forEach((key) => {
  en[key] = langMap[key][0];
  cn[key] = langMap[key][1] || key;
});

const messages = {
  'en-US': Object.assign(lang.enUS, en),
  'zh-CN': Object.assign(lang.zhCN, cn),
};

let curLang = cookie.parse(document.cookie).blueking_language || 'zh-cn';
if (['zh-CN', 'zh-cn', 'cn', 'zhCN', 'zhcn'].indexOf(curLang) > -1) {
  curLang = 'zh-CN';
} else {
  curLang = 'en-US';
}

// 代码中获取当前语言 this.$i18n.locale
const i18n = new VueI18n({
  locale: curLang,
  fallbackLocale: 'zh-CN',
  messages,
});

locale.i18n((key, value) => i18n.t(key, value));
locale.getCurLang().bk.lang = curLang;

global.i18n = i18n;

export default i18n;
