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

import Rules from './rules'

/**
 *  解析type
 */
// function parseType (type) {
//     if (typeof type !== 'string') {
//         return {
//             el: 'input',
//             type: 'text'
//         }
//     }

//     const $type = type.split(':')

//     return {
//         el: $type[0],
//         type: $type[1]
//     }
// }

/**
 *  解析rule
 */
function parseRule (rule) {
    if (typeof rule !== 'string') {
        return {
            rule: 'not_empty'
        }
    }

    const $rule = rule.split(':')

    return {
        rule: $rule[0],
        ext: $rule[1]
    }
}

/**
 *  错误控制
 *  @param {Element} el - 当前绑定了指令的DOM节点
 *  @param {Boolean} valid - 当前的值是否通过检测
 */
function ErrorHandler (el, valid) {
    if (!valid) {
        el.classList.add('has-error')
        el.setAttribute('data-bk-valid', false)
    } else {
        el.classList.remove('has-error')
        el.setAttribute('data-bk-valid', true)
    }
}

const install = Vue => {
    Vue.directive('bk-validation', {
        inserted: function (el) {
            // el.focus()
        },
        update (el, binding) {
            const {
                value,
                oldValue
            } = binding

            // 避免不必要的更新
            if (value.val === oldValue.val) return

            const parsedRule = parseRule(value.rule)
            let result

            switch (parsedRule.rule) {
                case 'not_empty':
                    result = Rules.notEmpty(value.val)
                    break
                case 'limit':
                    result = Rules.limit(value.val, parsedRule.ext)
                    break
            }

            ErrorHandler(el, result)
        }
    })
}

export default install
