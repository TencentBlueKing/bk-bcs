import { createApp } from 'vue';
import pinia from './store/index';
import bkui, { bkTooltips, bkEllipsis, overflowTitle } from 'bkui-vue';
import 'bkui-vue/dist/style.css';
import './css/style.scss';
import App from './App.vue';
import router from './router';
import './utils/login';
import i18n from './i18n/index';
import cursor from './components/permission/cursor';
import './components/permission/cursor.css';
import auth from './common/auth';
import zhCn from 'bkui-vue/dist/locale/zh-cn.esm';
import en from 'bkui-vue/dist/locale/en.esm';
import { getCookie } from './utils';

auth().then(() => {
  const app = createApp(App);
  app.directive('bkTooltips', bkTooltips);
  app.directive('bkEllipsis', bkEllipsis);
  app.directive('overflowTitle', overflowTitle);
  app.directive('cursor', cursor);
  app.directive('clickOutside', {
    mounted(el, binding) {
      const handleClickOutside = (event: any) => {
        if (!el.contains(event.target) && el !== event.target) {
          binding.value(event);
        }
      };
      setTimeout(() => {
        document.addEventListener('click', handleClickOutside);
        el.clickOutsideHandler = handleClickOutside;
      }, 0);
    },
    unmounted(el) {
      document.removeEventListener('click', el.clickOutsideHandler);
      delete el.clickOutsideHandler;
    },
  });

  app
    .use(pinia)
    .use(i18n)
    .use(router)
    .use(bkui, {
      locale: getCookie('blueking_language') === 'zh-cn' ? zhCn : en,
    })
    .mount('#app');
});

// 监听登录成功页通过postMessage发送的消息，刷新当前页面
window.addEventListener('message', (event) => {
  if (event.data === 'login') {
    window.location.reload();
  }
});
