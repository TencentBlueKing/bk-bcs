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

import 'bk-magic-vue/dist/bk-magic-vue.min.css';
import bcsMagic, { bkLink } from 'bk-magic-vue';

Vue.use(bkLink);
Vue.use(bcsMagic, {
  namespace: 'bcs',
});
Vue.use(bcsMagic.bkDialog, {
  headerPosition: 'left',
});

const tmpMessage = Vue.prototype.$bkMessage;
// 错误信息默认显示3行，内容超出时开启复制功能
Vue.prototype.$bkMessage = (config) => {
  const cfg = config?.theme === 'error'
    ? Object.assign({ ellipsisLine: 3, ellipsisCopy: true }, config)
    : config;

  tmpMessage(cfg);
};
const Message = Vue.prototype.$bkMessage;

let messageInstance = null;
export const messageError = (message, delay = 3000) => {
  messageInstance?.close();
  messageInstance = Message({
    message,
    delay,
    theme: 'error',
  });
};

export const messageSuccess = (message, delay = 3000) => {
  messageInstance?.close();
  messageInstance = Message({
    message,
    delay,
    theme: 'success',
  });
};

export const messageInfo = (message, delay = 3000) => {
  messageInstance?.close();
  messageInstance = Message({
    message,
    delay,
    theme: 'primary',
  });
};

export const messageWarn = (message, delay = 3000) => {
  messageInstance?.close();
  messageInstance = Message({
    message,
    delay,
    theme: 'warning',
    hasCloseIcon: true,
  });
};

Vue.prototype.messageError = messageError;
Vue.prototype.messageSuccess = messageSuccess;
Vue.prototype.messageInfo = messageInfo;
Vue.prototype.messageWarn = messageWarn;

export default Message;
