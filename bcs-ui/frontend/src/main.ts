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
import VeeValidate from 'vee-validate';
import Vue from 'vue';

import BkTrace from '@blueking/bk-trace';

import '@/common/bkmagic';
import '@/fonts/svg-icon/iconcool';
import '@/views/app/performance';
import App from '@/App.vue';
import { bus } from '@/common/bus';
import { chainable } from '@/common/util';
import bkCombox from '@/components/bk-input/combox.vue';
import bkbcsInput from '@/components/bk-input/index.vue';
import bkmagic2 from '@/components/bk-magic-2.0';
import BcsEmptyTableStatus from '@/components/empty-table-status.vue';
import bkSelector from '@/components/selector/index.vue';
import Authority from '@/directives/authority';
import focus from '@/directives/focus/index';
import i18n from '@/i18n/i18n-setup';
import config from '@/mixins/config';
import router from '@/router';
import store from '@/store';
import BcsErrorPlugin from '@/views/app/bcs-error';
import k8sIngress from '@/views/deploy-manage/templateset/ingress/k8s-ingress.vue';

window.BkTrace = new BkTrace({
  url: '/bcsapi/v4/ui/report',
  struct: {
    module: '',
    operation: '',
    desc: '',
    username: '',
    projectCode: '',
    to: '',
    from: '',
    error: {},
    performance: {},
    navID: () => {
      let menu = store.state.curNav;
      while (menu?.parent) {
        menu = menu?.parent;
      }
      return menu?.id || '';
    },
  },
});

Vue.config.devtools = process.env.NODE_ENV === 'development';
Vue.prototype.$chainable = chainable;

Vue.use(BcsErrorPlugin);
Vue.use(Authority);
Vue.use(focus);
Vue.use(bkmagic2);
Vue.use(VeeValidate, {
  fieldsBagName: '_veeFields',
});
Vue.mixin(config);
Vue.component('BkbcsInput', bkbcsInput);
Vue.component('BkCombox', bkCombox);
Vue.component('BkSelector', bkSelector);
Vue.component('K8sIngress', k8sIngress);
Vue.component('BcsEmptyTableStatus', BcsEmptyTableStatus);

window.bus = bus;
window.mainComponent = new Vue({
  el: '#app',
  router,
  store,
  components: { App },
  i18n,
  template: '<App/>',
});

console.log(
  `%c${BK_BCS_WELCOME} \n %c版本信息%c${BK_BCS_VERSION}%c>> ${new Date().toString()
    .slice(0, 16)}<<`,
  'color: #2DCB56',
  'padding: 2px 5px; background: #ea3636; color: #fff; border-radius: 3px 0 0 3px;',
  'padding: 2px 5px; background: #42c02e; color: #fff; border-radius: 0 3px 3px 0; font-weight: bold;',
  'background-color: #3A84FF; color: #fff; padding: 2px 5px; border-radius: 3px; font-weight: bold;margin-left: 16px;',
);
