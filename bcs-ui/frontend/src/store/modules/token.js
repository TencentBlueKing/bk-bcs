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

import {
    getTokens,
    createToken,
    updateToken,
    deleteToken
} from '@/api/base'

export default {
    namespaced: true,
    actions: {
        // 获取Token列表
        async getTokens (ctx, params) {
            const data = await getTokens(params).catch(() => [])
            return data || []
        },
        // 创建token
        async createToken (ctx, params) {
            const data = await createToken(params).catch(() => null)
            return data
        },
        // 更新Token
        async updateToken (ctx, params) {
            const data = await updateToken(params).catch(() => null)
            return data
        },
        // 删除Token
        async deleteToken (ctx, params) {
            const data = await deleteToken(params).then(() => true).catch(() => null)
            return data
        }
    }
}
