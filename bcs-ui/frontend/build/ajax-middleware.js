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

const path = require('path')
const fs = require('fs')
const url = require('url')
const queryString = require('querystring')
const chalk = require('chalk')

const requestHandler = req => {
    const pathName = req.path || ''

    const mockFilePath = path.join(__dirname, '../mock/ajax', pathName) + '.js'
    if (!fs.existsSync(mockFilePath)) {
        return false
    }

    console.log(chalk.magenta('Ajax Request Path: ', pathName))

    // 删除 require.cache[require.resolve(mockFilePath)] 缓存即每次 mock 请求都会重新执行
    // delete require.cache[require.resolve(mockFilePath)]
    return require(mockFilePath)
}

module.exports = async function ajaxMiddleWare (req, res, next) {
    let query = url.parse(req.url).query

    if (!query) {
        return next()
    }

    query = queryString.parse(query)

    if (!query.isAjax) {
        return next()
    }

    const postData = req.body || ''
    const mockDataHandler = requestHandler(req)
    let data = await mockDataHandler.response(query, postData, req)

    if (data.statusCode) {
        res.status(data.statusCode).end()
        return
    }

    let contentType = req.headers['Content-Type']

    // 返回值未指定内容类型，默认按 JSON 格式处理返回
    if (!contentType) {
        contentType = 'application/json;charset=UTF-8'
        req.headers['Content-Type'] = contentType
        res.setHeader('Content-Type', contentType)
        data = JSON.stringify(data || {})
    }

    res.end(data)

    return next()
}
