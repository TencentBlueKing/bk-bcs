/**
 * Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

import Vue from 'vue'

import 'bk-magic-vue/dist/bk-magic-vue.min.css'
import bcsMagic, { bkLink } from 'bk-magic-vue'

Vue.use(bkLink)
Vue.use(bcsMagic, {
    namespace: 'bcs'
})
Vue.use(bcsMagic.bkDialog, {
    headerPosition: 'left'
})

const tmpMessage = Vue.prototype.$bkMessage
// 错误信息默认显示3行，内容超出时开启复制功能
const Message = Vue.prototype.$bkMessage = (config) => {
    const cfg = config?.theme === 'error'
        ? Object.assign({ ellipsisLine: 3, ellipsisCopy: true }, config)
        : config

    tmpMessage(cfg)
}

let messageInstance = null
export const messageError = (message, delay = 3000) => {
    messageInstance && messageInstance.close()
    messageInstance = Message({
        message,
        delay,
        theme: 'error'
    })
}

export const messageSuccess = (message, delay = 3000) => {
    messageInstance && messageInstance.close()
    messageInstance = Message({
        message,
        delay,
        theme: 'success'
    })
}

export const messageInfo = (message, delay = 3000) => {
    messageInstance && messageInstance.close()
    messageInstance = Message({
        message,
        delay,
        theme: 'primary'
    })
}

export const messageWarn = (message, delay = 3000) => {
    messageInstance && messageInstance.close()
    messageInstance = Message({
        message,
        delay,
        theme: 'warning',
        hasCloseIcon: true
    })
}

Vue.prototype.messageError = messageError
Vue.prototype.messageSuccess = messageSuccess
Vue.prototype.messageInfo = messageInfo
Vue.prototype.messageWarn = messageWarn
