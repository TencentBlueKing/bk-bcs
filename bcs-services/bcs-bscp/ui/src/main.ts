import { createApp } from 'vue';
import pinia from './store/index';
import bkui, { bkTooltips, bkEllipsis, overflowTitle } from 'bkui-vue';
import 'bkui-vue/dist/style.css';
import './css/style.css';
import App from './App.vue';
import router from './router';
import './utils/login';
import i18n from './i18n/index';
import cursor from './components/permission/cursor';
import './components/permission/cursor.css';
import auth from './common/auth';

auth().then(() => {
  const app = createApp(App);
  app.directive('bkTooltips', bkTooltips);
  app.directive('bkEllipsis', bkEllipsis);
  app.directive('overflowTitle', overflowTitle);
  app.directive('cursor', cursor);

  app.use(pinia).use(i18n).use(router).use(bkui).mount('#app');
});
