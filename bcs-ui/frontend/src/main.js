/**
 * @file main entry
 */
import Vue from 'vue'
import VeeValidate from 'vee-validate'
import i18n from '@open/setup/i18n-setup'
import '@open/setup/global-components-setup'
import VueCompositionAPI from '@vue/composition-api'
import '@open/common/bkmagic'
import bkmagic2 from '@open/components/bk-magic-2.0'
import { bus } from '@open/common/bus'
import { loadScript } from '@open/common/util'
import { injectCSRFTokenToHeaders } from '@open/api'
import focus from '@/directives/focus/index'
import App from '@/App'
import router from '@/router'
import store from '@open/store'
import Authority from '@/directives/authority'
import '@icon-cool/bk-icon-bk-bcs'

Vue.config.devtools = NODE_ENV === 'development'

Vue.use(VueCompositionAPI)

Vue.use(Authority)
Vue.use(focus)
Vue.use(bkmagic2)
Vue.use(VeeValidate)

/**
 * 加载 devops-utils.js 的回调函数
 *
 * @param {string} e 错误信息
 */
function loadScriptCallback (e) {
    if (e) {
        console.error(e)
        return
    }

    injectCSRFTokenToHeaders()

    Vue.mixin({
        data () {
            return { PROJECT_CONFIG: window.BCS_CONFIG }
        },
        computed: {
            $INTERNAL  () {
                return !['ce', 'ee'].includes(window.REGION)
            }
        }
    })

    if (store.getters.lang === 'en-US') {
        document.body.style.fontFamily = 'arial,sans-serif'
    }

    window.bus = bus
    window.mainComponent = new Vue({
        el: '#app',
        router,
        store,
        components: {
            App
        },
        i18n,
        template: '<App/>'
    })
}
loadScript(DEVOPS_HOST + '/console/static/devops-utils.js', loadScriptCallback)
