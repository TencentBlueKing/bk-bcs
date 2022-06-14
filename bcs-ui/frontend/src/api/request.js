import http from '@/api'
import { json2Query } from '@/common/util'
import router from '@/router'
import store from '@/store'
// import { crPrefix } from '@/api/base'

const methodsWithoutData = ['get', 'head', 'options', 'delete']
const defaultConfig = { needRes: false }
const prefixData = []

export const parseUrl = (url, params = {}) => {
    // 全局URL变量替换
    const currentRoute = router.currentRoute
    const variableData = {
        '$projectId': currentRoute.params.projectId,
        '$clusterId': store.state.curClusterId || currentRoute.query.cluster_id || currentRoute.params.cluster_id
    }
    Object.keys(params).forEach(key => {
        // 自定义url变量
        if (key.indexOf('$') === 0) {
            variableData[key] = params[key]
        }
    })
    let newUrl = `${/(http|https):\/\/([\w.]+\/?)\S*/.test(url) || prefixData.some(prefix => url.indexOf(prefix) === 0)
        ? url : `${DEVOPS_BCS_API_URL}${url}`}`
    Object.keys(variableData).forEach(key => {
        if (!variableData[key]) {
            // console.warn(`路由变量未配置${key}`)
            // 去除后面的路径符号
            newUrl = newUrl.replace(`/${key}`, '')
        } else {
            newUrl = newUrl.replace(new RegExp(`\\${key}`, 'g'), variableData[key])
        }
        delete params[key]
    })
    return newUrl
}

export const request = (method, url) => (params = {}, config = {}) => {
    const reqMethod = method.toLowerCase()
    const reqConfig = Object.assign({}, defaultConfig, config)

    let newUrl = parseUrl(url, params)
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
