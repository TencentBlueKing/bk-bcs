import { createApp } from 'vue';
import { pinia } from './store/index'
import bkui, { bkTooltips, bkEllipsis } from 'bkui-vue';
import 'bkui-vue/dist/style.css'
import './css/style.css';
import App from './App.vue';
import router from './router';
import './utils/login'
import i18n from './i18n/index';

const app = createApp(App)

app.directive('bkTooltips', bkTooltips)
app.directive('bkEllipsis', bkEllipsis)

app
.use(pinia)
.use(i18n)
.use(router)
.use(bkui)
.mount('#app')
