/**
 * @file main entry
 */
import Vue from 'vue'
import VeeValidate from 'vee-validate'
import i18n from '@/i18n/i18n-setup'
import VueCompositionAPI from '@vue/composition-api'
import '@/common/bkmagic'
import bkmagic2 from '@/components/bk-magic-2.0'
import { bus } from '@/common/bus'
import focus from '@/directives/focus/index'
import App from '@/App'
import router from '@/router'
import store from '@/store'
import Authority from '@/directives/authority'
import config from '@/mixins/config'
import Exception from '@/components/exception'
import bkSelector from '@/components/selector'
import bkDataSearcher from '@/components/data-searcher'
import bkPageCounter from '@/components/page-counter'
import bkNumber from '@/components/number'
import bkbcsInput from '@/components/bk-input'
import bkCombox from '@/components/bk-input/combox'
import bkTextarea from '@/components/bk-textarea'
import ApplyPerm from '@/components/apply-perm'
import bkGuide from '@/components/guide'
import bkFileUpload from '@/components/file-upload'
import k8sIngress from '@/views/ingress/k8s-ingress.vue'

Vue.config.devtools = NODE_ENV === 'development'

Vue.use(VueCompositionAPI)
Vue.use(Authority)
Vue.use(focus)
Vue.use(bkmagic2)
Vue.use(VeeValidate)
Vue.mixin(config)
Vue.component('app-exception', Exception)
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

window.bus = bus
window.mainComponent = new Vue({
    el: '#app',
    router,
    store,
    components: { App },
    i18n,
    template: '<App/>'
})
console.log(`%c${BK_CI_BUILD_NUM}`, 'color: #3a84ff')
