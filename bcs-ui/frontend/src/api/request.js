import http from '@/api'
import { json2Query } from '@/common/util'
import router from '@/router'
import store from '@/store'

const methodsWithoutData = ['get', 'head', 'options', 'delete']
const defaultConfig = { needRes: false }

export const request = (method, url) => (params = {}, config = {}) => {
    const reqMethod = method.toLowerCase()
    const reqConfig = Object.assign({}, defaultConfig, config)

    // 全局URL变量替换
    const variableData = {
        '$projectId': router.currentRoute.params.projectId,
        '$clusterId': store.state.curClusterId || ''
        // '$namespace': ''
    }
    Object.keys(params).forEach(key => {
        // 自定义url变量
        if (key.indexOf('$') === 0) {
            variableData[key] = params[key]
        }
    })
    let newUrl = `${DEVOPS_BCS_API_URL}${url}`
    Object.keys(variableData).forEach(key => {
        if (!variableData[key]) {
            // console.warn(`路由变量未配置${key}`)
            // 去除后面的路径符号
            newUrl = newUrl.replace(new RegExp(`\\${key}/`, 'g'), '')
        } else {
            newUrl = newUrl.replace(new RegExp(`\\${key}`, 'g'), variableData[key])
        }
        delete params[key]
    })

    let req = null
    if (methodsWithoutData.includes(reqMethod)) {
        const query = json2Query(params, '')
        if (query) {
            newUrl += `?${query}`
        }
        req = http[reqMethod](newUrl, null, reqConfig)
    } else {
        req = http[reqMethod](newUrl, params, reqConfig)
    }
    return req.then((res) => {
        if (reqConfig.needRes) return Promise.resolve(res)

        return Promise.resolve(res.data)
    }).catch((err) => {
        console.error('request error', err)
        return Promise.reject(err)
    })
}

export default request
