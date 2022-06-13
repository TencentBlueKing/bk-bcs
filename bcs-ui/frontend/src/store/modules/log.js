import { podContainersList, podLogs, podLogsDownloadURL, podLogsStreamURL } from '@/api/base'
import request, { parseUrl } from '@/api/request'

export default {
    namespaced: true,
    actions: {
        async podContainersList (context, params) {
            const data = await podContainersList(params).catch(() => [])
            return data
        },
        // 获取POD日志
        async podLogs (context, params) {
            const data = await podLogs(params).catch(() => ({ logs: [] }))
            return data
        },
        // 下载日志
        async downloadLogs (context, params) {
            window.open(`${parseUrl(podLogsDownloadURL, params)}`)
        },
        // 实时日志
        async realTimeLogStream (context, params) {
            const source = new EventSource(parseUrl(podLogsStreamURL, params), { withCredentials: true })
            return source
        },
        async previousLogList (context, url) {
            const data = await request('get', `${url}`)().catch(() => ({ logs: [] }))
            return data
        }
    }
}
