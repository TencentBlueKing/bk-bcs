import { createI18n } from 'vue-i18n';
import { getCookie } from '../utils';
import zhCn from './zh-cn';
import enUs from './en-us';
const i18n = createI18n({
  locale: getCookie('blueking_language') || 'zh-cn',
  legacy: false,
  messages: {
    'zh-cn': zhCn,
    en: enUs,
  },
});

export const localT = i18n.global.t;
export default i18n;
