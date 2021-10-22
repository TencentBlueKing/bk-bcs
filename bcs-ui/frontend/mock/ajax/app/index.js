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

 const chalk = require('chalk')

 const { sleep } = require('../util')

 module.exports = async function response (getArgs, postArgs, req) {
     console.log(chalk.cyan('req', req.method))
     console.log(chalk.cyan('getArgs', JSON.stringify(getArgs, null, 0)))
     console.log(chalk.cyan('postArgs', JSON.stringify(postArgs, null, 0)))
     console.log()
     const invoke = getArgs.invoke
     if (invoke === 'userInfo') {
         return {
             // statusCode: 401,
             code: 2000,
             data: {
                 baseInfo: {
                     a: 1,
                     b: 2
                 }
             },
             message: 'ddddd'
         }
     } else if (invoke === 'baseInfo') {
         return {
             code: 0,
             data: {
                 baseInfo: {
                     a: 1,
                     b: 2
                 }
             },
             message: 'ok'
         }
     } else if (invoke === 'enterExample1') {
         const delay = getArgs.delay
         await sleep(delay)
         return {
             // http status code, 后端返回的数据没有这个字段，这里模拟这个字段是为了在 mock 时更灵活的自定义 http status code，
             // 同时热更新即改变 http status code 后无需重启服务，这个字段的处理参见 build/ajax-middleware.js
             // statusCode: 401,
             code: 0,
             data: {
                 msg: `我是 enterExample1 请求返回的数据。本请求需耗时 ${delay} ms`
             },
             message: 'ok'
         }
     } else if (invoke === 'enterExample2') {
         const delay = postArgs.delay
         await sleep(delay)
         return {
             // http status code, 后端返回的数据没有这个字段，这里模拟这个字段是为了在 mock 时更灵活的自定义 http status code，
             // 同时热更新即改变 http status code 后无需重启服务，这个字段的处理参见 build/ajax-middleware.js
             // statusCode: 401,
             code: 0,
             data: {
                 msg: `我是 enterExample2 请求返回的数据。本请求需耗时 ${delay} ms`
             },
             message: 'ok'
         }
     } else if (invoke === 'btn1') {
         await sleep(1000)
         return {
             // http status code, 后端返回的数据没有这个字段，这里模拟这个字段是为了在 mock 时更灵活的自定义 http status code，
             // 同时热更新即改变 http status code 后无需重启服务，这个字段的处理参见 build/ajax-middleware.js
             // statusCode: 401,
             code: 0,
             data: {
                 msg: `我是 btn1 请求返回的数据。本请求需耗时 3000 ms. ${+new Date()}`
             },
             message: 'ok'
         }
     } else if (invoke === 'btn2') {
         await sleep(3000)
         return {
             // http status code, 后端返回的数据没有这个字段，这里模拟这个字段是为了在 mock 时更灵活的自定义 http status code，
             // 同时热更新即改变 http status code 后无需重启服务，这个字段的处理参见 build/ajax-middleware.js
             // statusCode: 401,
             code: 0,
             data: {
                 msg: `我是 btn2 请求返回的数据。本请求需耗时 3000 ms. ${+new Date()}`
             },
             message: 'ok'
         }
     } else if (invoke === 'del') {
         return {
             code: 0,
             data: {
                 msg: `我是 del 请求返回的数据。请求参数为 ${postArgs.time}`
             },
             message: 'ok'
         }
     } else if (invoke === 'get') {
         // await sleep(1000)
         return {
             // http status code, 后端返回的数据没有这个字段，这里模拟这个字段是为了在 mock 时更灵活的自定义 http status code，
             // 同时热更新即改变 http status code 后无需重启服务，这个字段的处理参见 build/ajax-middleware.js
             // statusCode: 401,
             code: 0,
             data: {
                 reqTime: getArgs.time,
                 resTime: +new Date()
             },
             message: 'ok'
         }
     } else if (invoke === 'post') {
         // await sleep(1000)
         return {
             code: 0,
             data: {
                 reqTime: postArgs.time,
                 resTime: +new Date()
             },
             message: 'ok'
         }
     } else if (invoke === 'long') {
         await sleep(5000)
         return {
             code: 0,
             data: {
                 reqTime: getArgs.time,
                 resTime: +new Date()
             },
             message: 'ok'
         }
     } else if (invoke === 'long1') {
         await sleep(2000)
         return {
             code: 0,
             data: {
                 reqTime: getArgs.time,
                 resTime: +new Date()
             },
             message: 'ok'
         }
     } else if (invoke === 'same') {
         await sleep(5000)
         return {
             code: 0,
             data: {
                 reqTime: getArgs.time,
                 resTime: +new Date()
             },
             message: 'ok'
         }
     } else if (invoke === 'postSame') {
         await sleep(5000)
         return {
             code: 0,
             // statusCode: 401,
             data: {
                 reqTime: postArgs.time,
                 resTime: +new Date()
             },
             message: 'ok'
         }
     } else if (invoke === 'go') {
         await sleep(2000)
         return {
             code: 0,
             statusCode: 400,
             data: {
                 reqTime: postArgs.time,
                 resTime: +new Date()
             },
             message: 'ok'
         }
     } else if (invoke === 'user') {
         return {
             code: 0,
             data: {
                 name: 'name'
             },
             message: 'ok'
         }
     }
     return {
         code: 0,
         data: {}
     }
 }
