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

/**
 * 函数柯里化
 *
 * @example
 *     function add (a, b) {return a + b}
 *     curry(add)(1)(2)
 *
 * @param {Function} fn 要柯里化的函数
 *
 * @return {Function} 柯里化后的函数
 */
export function curry (fn) {
    const judge = (...args) => {
        return args.length === fn.length
            ? fn(...args)
            : arg => judge(...args, arg)
    }
    return judge
}

/**
 * 判断是否是对象
 *
 * @param {Object} obj 待判断的
 *
 * @return {boolean} 判断结果
 */
export function isObject (obj) {
    return obj !== null && typeof obj === 'object'
}

/**
 * 规范化参数
 *
 * @param {Object|string} type vuex type
 * @param {Object} payload vuex payload
 * @param {Object} options vuex options
 *
 * @return {Object} 规范化后的参数
 */
export function unifyObjectStyle (type, payload, options) {
    if (isObject(type) && type.type) {
        options = payload
        payload = type
        type = type.type
    }

    if (NODE_ENV !== 'production') {
        if (typeof type !== 'string') {
            console.warn(`expects string as the type, but found ${typeof type}.`)
        }
    }

    return { type, payload, options }
}

/**
 * 以 baseColor 为基础生成随机颜色
 *
 * @param {string} baseColor 基础颜色
 * @param {number} count 随机颜色个数
 *
 * @return {Array} 颜色数组
 */
export function randomColor (baseColor, count) {
    const segments = baseColor.match(/[\da-z]{2}/g)
    // 转换成 rgb 数字
    for (let i = 0; i < segments.length; i++) {
        segments[i] = parseInt(segments[i], 16)
    }
    const ret = []
    // 生成 count 组颜色，色差 20 * Math.random
    for (let i = 0; i < count; i++) {
        ret[i] = '#'
            + Math.floor(segments[0] + (Math.random() < 0.5 ? -1 : 1) * Math.random() * 20).toString(16)
            + Math.floor(segments[1] + (Math.random() < 0.5 ? -1 : 1) * Math.random() * 20).toString(16)
            + Math.floor(segments[2] + (Math.random() < 0.5 ? -1 : 1) * Math.random() * 20).toString(16)
    }
    return ret
}

/**
 * min max 之间的随机整数
 *
 * @param {number} min 最小值
 * @param {number} max 最大值
 *
 * @return {number} 随机数
 */
export function randomInt (min, max) {
    return Math.floor(Math.random() * (max - min + 1) + min)
}

/**
 * 异常处理
 *
 * @param {Object} err 错误对象
 * @param {Object} ctx 上下文对象，这里主要指当前的 Vue 组件
 */
export function catchErrorHandler (err, ctx) {
    const data = err.data
    if (data) {
        if (!err.code || err.code === 404) {
            ctx.exceptionCode = {
                code: '404',
                msg: window.i18n.t('当前访问的页面不存在')
            }
        } else if (err.code === 403) {
            ctx.exceptionCode = {
                code: '403',
                msg: window.i18n.t('Sorry，您的权限不足!')
            }
        } else {
            console.error(err)
        }
    } else {
        // 其它像语法之类的错误不展示
        console.error(err)
    }
}

/**
 * 获取字符串长度，中文算两个，英文算一个
 *
 * @param {string} str 字符串
 *
 * @return {number} 结果
 */
export function getStringLen (str) {
    let len = 0
    for (let i = 0; i < str.length; i++) {
        if (str.charCodeAt(i) > 127 || str.charCodeAt(i) === 94) {
            len += 2
        } else {
            len++
        }
    }
    return len
}

/**
 * 转义特殊字符
 *
 * @param {string} str 待转义字符串
 *
 * @return {string} 结果
 */
export const escape = str => String(str).replace(/([.*+?^=!:${}()|[\]\/\\])/g, '\\$1')

/**
 * 对象转为 url query 字符串
 *
 * @param {*} param 要转的参数
 * @param {string} key key
 *
 * @return {string} url query 字符串
 */
export function json2Query (param, key) {
    const mappingOperator = '='
    const separator = '&'
    let paramStr = ''

    if (param instanceof String || typeof param === 'string'
            || param instanceof Number || typeof param === 'number'
            || param instanceof Boolean || typeof param === 'boolean'
    ) {
        paramStr += separator + key + mappingOperator + encodeURIComponent(param)
    } else if (typeof param === 'object') {
        Object.keys(param).forEach(p => {
            const value = param[p]
            const k = (key === null || key === '' || key === undefined)
                ? p
                : key + (param instanceof Array ? '[' + p + ']' : '.' + p)
            paramStr += separator + json2Query(value, k)
        })
    }
    return paramStr.substr(1)
}

/**
 * 字符串转换为驼峰写法
 *
 * @param {string} str 待转换字符串
 *
 * @return {string} 转换后字符串
 */
export function camelize (str) {
    return str.replace(/-(\w)/g, (strMatch, p1) => p1.toUpperCase())
}

/**
 * 获取元素的样式
 *
 * @param {Object} elem dom 元素
 * @param {string} prop 样式属性
 *
 * @return {string} 样式值
 */
export function getStyle (elem, prop) {
    if (!elem || !prop) {
        return false
    }

    // 先获取是否有内联样式
    let value = elem.style[camelize(prop)]

    if (!value) {
        // 获取的所有计算样式
        let css = ''
        if (document.defaultView && document.defaultView.getComputedStyle) {
            css = document.defaultView.getComputedStyle(elem, null)
            value = css ? css.getPropertyValue(prop) : null
        }
    }

    return String(value)
}

/**
 *  获取元素相对于页面的高度
 *
 *  @param {Object} node 指定的 DOM 元素
 */
export function getActualTop (node) {
    let actualTop = node.offsetTop
    let current = node.offsetParent

    while (current !== null) {
        actualTop += current.offsetTop
        current = current.offsetParent
    }

    return actualTop
}

/**
 *  获取元素相对于页面左侧的宽度
 *
 *  @param {Object} node 指定的 DOM 元素
 */
export function getActualLeft (node) {
    let actualLeft = node.offsetLeft
    let current = node.offsetParent

    while (current !== null) {
        actualLeft += current.offsetLeft
        current = current.offsetParent
    }

    return actualLeft
}

/**
 * document 总高度
 *
 * @return {number} 总高度
 */
export function getScrollHeight () {
    let scrollHeight = 0
    let bodyScrollHeight = 0
    let documentScrollHeight = 0

    if (document.body) {
        bodyScrollHeight = document.body.scrollHeight
    }

    if (document.documentElement) {
        documentScrollHeight = document.documentElement.scrollHeight
    }

    scrollHeight = (bodyScrollHeight - documentScrollHeight > 0) ? bodyScrollHeight : documentScrollHeight

    return scrollHeight
}

/**
 * 滚动条在 y 轴上的滚动距离
 *
 * @return {number} y 轴上的滚动距离
 */
export function getScrollTop () {
    let scrollTop = 0
    let bodyScrollTop = 0
    let documentScrollTop = 0

    if (document.body) {
        bodyScrollTop = document.body.scrollTop
    }

    if (document.documentElement) {
        documentScrollTop = document.documentElement.scrollTop
    }

    scrollTop = (bodyScrollTop - documentScrollTop > 0) ? bodyScrollTop : documentScrollTop

    return scrollTop
}

/**
 * 浏览器视口的高度
 *
 * @return {number} 浏览器视口的高度
 */
export function getWindowHeight () {
    const windowHeight = document.compatMode === 'CSS1Compat'
        ? document.documentElement.clientHeight
        : document.body.clientHeight

    return windowHeight
}

/**
 * 简单的 loadScript
 *
 * @param {string} url js 地址
 * @param {Function} callback 回调函数
 */
export function loadScript (url, callback) {
    const script = document.createElement('script')
    script.async = true
    script.src = url

    script.onerror = () => {
        callback(new Error('Failed to load: ' + url))
    }

    script.onload = () => {
        callback()
    }

    document.getElementsByTagName('head')[0].appendChild(script)
}

/**
 * 在当前节点后面插入节点
 *
 * @param {Object} newElement 待插入 dom 节点
 * @param {Object} targetElement 当前节点
 */
export function insertAfter (newElement, targetElement) {
    const parent = targetElement.parentNode
    if (parent.lastChild === targetElement) {
        parent.appendChild(newElement)
    } else {
        parent.insertBefore(newElement, targetElement.nextSibling)
    }
}

/**
 * 生成UUID
 *
 * @param  {number} len 位数
 * @param  {number} radix 进制
 * @return {string} uuid
 */
export function uuid (len, radix) {
    const chars = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz'.split('')
    const uuid = []
    let i
    radix = radix || chars.length
    if (len) {
        for (i = 0; i < len; i++) {
            uuid[i] = chars[0 | Math.random() * radix]
        }
    } else {
        let r
        uuid[8] = uuid[13] = uuid[18] = uuid[23] = '-'
        uuid[14] = '4'

        for (i = 0; i < 36; i++) {
            if (!uuid[i]) {
                r = 0 | Math.random() * 16
                uuid[i] = chars[(i === 19) ? (r & 0x3) | 0x8 : r]
            }
        }
    }

    return uuid.join('')
}

/**
 * 图表颜色
 */
export const chartColors = [
    '#2ec7c9',
    '#b6a2de',
    '#5ab1ef',
    '#fdb980',
    '#d87a80',
    '#8d98b3',
    '#e5cf0f',
    '#97b552',
    '#95706d',
    '#dc69aa'
]

/* 格式化日期
 *
 * @param  {string} date 日期
 * @param  {string} formatStr 格式
 * @return {str} 格式化后的日期
 */
export function formatDate (date, formatStr = 'YYYY-MM-DD hh:mm:ss') {
    if (!date) return ''

    const dateObj = new Date(date)
    const o = {
        'M+': dateObj.getMonth() + 1, // 月份
        'D+': dateObj.getDate(), // 日
        'h+': dateObj.getHours(), // 小时
        'm+': dateObj.getMinutes(), // 分
        's+': dateObj.getSeconds(), // 秒
        'q+': Math.floor((dateObj.getMonth() + 3) / 3), // 季度
        'S': dateObj.getMilliseconds() // 毫秒
    }
    if (/(Y+)/.test(formatStr)) {
        formatStr = formatStr.replace(RegExp.$1, (dateObj.getFullYear() + '').substr(4 - RegExp.$1.length))
    }
    for (const k in o) {
        if (new RegExp('(' + k + ')').test(formatStr)) {
            formatStr = formatStr.replace(RegExp.$1, (RegExp.$1.length === 1) ? (o[k]) : (('00' + o[k]).substr(('' + o[k]).length)))
        }
    }

    return formatStr
}

/**
 * bytes 转换
 *
 * @param {Number} bytes 字节数
 * @param {Number} decimals 保留小数位
 *
 * @return {string} 转换后的值
 */
export function formatBytes (bytes, decimals) {
    if (parseFloat(bytes) === 0) {
        return '0 B'
    }
    const k = 1024
    const dm = decimals || 2
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    if (i === -1) {
        return bytes + ' B'
    }

    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + (sizes[i] || '')
}

/**
 * 判断是否为空
 * @param {Object} obj
 */
export function isEmpty (obj) {
    return typeof obj === 'undefined' || obj === null || obj === ''
}

/**
 * 生成随机数
 * @param {Number} n
 */
export const random = (n) => { // 生成n位长度的字符串
    const str = 'abcdefghijklmnopqrstuvwxyz0123456789' // 可以作为常量放到random外面
    let result = ''
    for (let i = 0; i < n; i++) {
        result += str[parseInt(Math.random() * str.length, 10)]
    }
    return result
}

/**
 * 清空对象的属性值
 * @param {*} obj
 */
export function clearObjValue (obj) {
    if (Object.prototype.toString.call(obj) !== '[object Object]') return

    Object.keys(obj).forEach(key => {
        if (Array.isArray(obj[key])) {
            obj[key] = []
        } else if (Object.prototype.toString.call(obj[key]) === '[object Object]') {
            clearObjValue(obj[key])
        } else {
            obj[key] = ''
        }
    })
}

/**
 * 获取对象key键下的值
 * @param {*} obj { a: { b: { c: '123' } } }
 * @param {*} key 'a.b.c'
 */
export const getObjectProps = (obj, key) => {
    if (!isObject(obj)) return obj[key]

    return String(key).split('.').reduce((pre, k) => {
        return pre && pre[k] ? pre[k] : undefined
    }, obj)
}

/**
 * 排序数组对象
 * 排序规则：1. 数字 => 2. 字母 => 3. 中文
 * @param {*} arr
 * @param {*} key
 */
export const sort = (arr, key, order = 'ascending') => {
    if (!Array.isArray(arr)) return arr
    const reg = /^[0-9a-zA-Z]/
    const data = arr.sort((pre, next) => {
        if (isObject(pre) && isObject(next) && key) {
            const preStr = String(getObjectProps(pre, key))
            const nextStr = String(getObjectProps(next, key))
            if (reg.test(preStr) && !reg.test(nextStr)) {
                return -1
            } if (!reg.test(preStr) && reg.test(nextStr)) {
                return 1
            }
            return preStr.localeCompare(nextStr)
        }
        return (`${pre}`).toString().localeCompare((`${pre}`))
    })
    return order === 'ascending' ? data : data.reverse()
}

// 格式化时间
export const formatTime = (timestamp, fmt) => {
    const time = new Date(timestamp)
    const opt = {
        "M+": time.getMonth() + 1, // 月份
        "d+": time.getDate(), // 日
        "h+": time.getHours(), // 小时
        "m+": time.getMinutes(), // 分
        "s+": time.getSeconds(), // 秒
        "q+": Math.floor((time.getMonth() + 3) / 3), // 季度
        "S": time.getMilliseconds() // 毫秒
    }
    if (/(y+)/.test(fmt)) {
        fmt = fmt.replace(RegExp.$1, (time.getFullYear() + "").substr(4 - RegExp.$1.length))
    }
    for (const k in opt) {
        if (new RegExp("(" + k + ")").test(fmt)) {
            fmt = fmt.replace(RegExp.$1, (RegExp.$1.length === 1)
                ? opt[k]
                : (`00${opt[k]}`).substr(String(opt[k]).length))
        }
    }
    return fmt
}

export const copyText = (text, errorMsg) => {
    const textarea = document.createElement('textarea')
    document.body.appendChild(textarea)
    textarea.value = text
    textarea.select()
    if (document.execCommand('copy')) {
        document.execCommand('copy')
    } else if (errorMsg) {
        errorMsg('浏览器不支持此功能，请使用谷歌浏览器。')
    }
    document.body.removeChild(textarea)
}
