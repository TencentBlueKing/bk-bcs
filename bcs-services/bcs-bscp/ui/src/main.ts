import { createApp } from 'vue';
import './style.css';
import App from './App.vue';
import bkui, { bkTooltips, bkEllipsis } from 'bkui-vue';
import 'bkui-vue/dist/style.css'
import router from './router';
import i18n from './i18n/index';
import '@tencent/bk-icon-bk_bscp/src/index.css';

const app = createApp(App)

app.directive('bkTooltips', bkTooltips)
app.directive('bkEllipsis', bkEllipsis)

app.use(i18n)
.use(router)
.use(bkui)
.mount('#app')
