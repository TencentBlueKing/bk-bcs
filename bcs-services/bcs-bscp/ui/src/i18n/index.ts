import { createI18n } from 'vue-i18n';
import zhCn from './zh-cn';
import enUs from './en-us';
const i18n = createI18n({
  locale: 'zh-CN',
  legacy: false,
  messages: {
    'zh-CN': zhCn,
    'en-US': enUs,
  },
});


export const localT = i18n.global.t;
export default i18n;
