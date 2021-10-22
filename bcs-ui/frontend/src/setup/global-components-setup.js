// 全局组件相关引入
import Vue from 'vue'
import appHeader from '@open/components/app-header.vue'
import Exception from '@open/components/exception'
import bkSelector from '@open/components/selector'
import bkDataSearcher from '@open/components/data-searcher'
import bkPageCounter from '@open/components/page-counter'
import bkNumber from '@open/components/number'
import bkbcsInput from '@open/components/bk-input'
import bkCombox from '@open/components/bk-input/combox'
import bkTextarea from '@open/components/bk-textarea'
import AuthComponent from '@/components/auth'
import ApplyPerm from '@/components/apply-perm'
import bkGuide from '@/components/guide'
import bkFileUpload from '@/components/file-upload'
import k8sIngress from '@open/views/ingress/k8s-ingress.vue'

Vue.component('app-header', appHeader)
Vue.component('app-exception', Exception)
Vue.component('app-auth', AuthComponent)
Vue.component('app-apply-perm', ApplyPerm)
Vue.component('bk-number-input', bkNumber)
Vue.component('bkbcs-input', bkbcsInput)
Vue.component('bk-combox', bkCombox)
Vue.component('bk-textarea', bkTextarea)
Vue.component('bk-file-upload', bkFileUpload)
Vue.component('bk-selector', bkSelector)
Vue.component('bk-guide', bkGuide)
Vue.component('bk-data-searcher', bkDataSearcher)
Vue.component('bk-page-counter', bkPageCounter)
Vue.component('k8s-ingress', k8sIngress)
