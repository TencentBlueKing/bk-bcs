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
import appHeader from '@open/components/app-header.vue';
import Exception from '@open/components/exception';
import bkSelector from '@open/components/selector';
import bkDataSearcher from '@open/components/data-searcher';
import bkPageCounter from '@open/components/page-counter';
import bkNumber from '@open/components/number';
import bkbcsInput from '@open/components/bk-input';
import bkCombox from '@open/components/bk-input/combox';
import bkTextarea from '@open/components/bk-textarea';
import AuthComponent from '@/components/auth';
import ApplyPerm from '@/components/apply-perm';
import bkGuide from '@/components/guide';
import bkFileUpload from '@/components/file-upload';
import k8sIngress from '@open/views/ingress/k8s-ingress.vue';

Vue.component('AppHeader', appHeader);
Vue.component('AppException', Exception);
Vue.component('AppAuth', AuthComponent);
Vue.component('AppApplyPerm', ApplyPerm);
Vue.component('BkNumberInput', bkNumber);
Vue.component('BkbcsInput', bkbcsInput);
Vue.component('BkCombox', bkCombox);
Vue.component('BkTextarea', bkTextarea);
Vue.component('BkFileUpload', bkFileUpload);
Vue.component('BkSelector', bkSelector);
Vue.component('BkGuide', bkGuide);
Vue.component('BkDataSearcher', bkDataSearcher);
Vue.component('BkPageCounter', bkPageCounter);
Vue.component('K8sIngress', k8sIngress);
