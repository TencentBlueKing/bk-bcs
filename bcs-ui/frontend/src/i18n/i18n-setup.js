// 国际化处理
import Vue from 'vue'
import VueI18n from 'vue-i18n'
import { locale, lang } from 'bk-magic-vue'
import cookie from 'cookie'
import langMap from './lang'

Vue.use(VueI18n)

const en = {}
const cn = {}
Object.keys(langMap).forEach(key => {
    en[key] = langMap[key][0]
    cn[key] = langMap[key][1] || key
})

const messages = {
    'en-US': Object.assign(lang.enUS, en),
    'zh-CN': Object.assign(lang.zhCN, cn)
}

let curLang = cookie.parse(document.cookie).blueking_language || 'zh-cn'
if (['zh-CN', 'zh-cn', 'cn', 'zhCN', 'zhcn'].indexOf(curLang) > -1) {
    curLang = 'zh-CN'
} else {
    curLang = 'en-US'
}

// 代码中获取当前语言 this.$i18n.locale
const i18n = new VueI18n({
    locale: curLang,
    fallbackLocale: 'zh-CN',
    messages
})

locale.i18n((key, value) => i18n.t(key, value))
locale.getCurLang().bk.lang = curLang

global.i18n = i18n

export default i18n
