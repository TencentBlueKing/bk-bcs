import { createI18n } from 'vue-i18n';
import zhCn from './zh-cn';
import enUs from './en-us';
export default createI18n({
  locale: 'zh-cn',
  legacy: false,
  messages: {
    'zh-cn': zhCn,
    'en-us': enUs
  }
});
