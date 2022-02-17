import http from '@/api'
import { stdLogsSession } from '@/api/base'

export default {
    namespaced: true,
    actions: {
        async getLogList (context, { projectId, clusterId, namespaceId, podId, containerName, previous }) {
            const data = await http.get(
                `${DEVOPS_BCS_API_URL}/api/logs/projects/${projectId}/clusters/${clusterId}/namespaces/${namespaceId}/pods/${podId}/stdlogs/?container_name=${containerName}&previous=${!!previous}`,
                {}
            ).catch(_ => {
                return {
                    data: {
                        logs: []
                    }
                }
            })
            return data
        },

        async previousLogList (context, url) {
            const data = await http.get(`${DEVOPS_BCS_API_URL}/${url}`).catch(_ => ({ data: { logs: [] } }))
            return data
        },

        async downloadLog (context, { projectId, clusterId, namespaceId, podId, containerName, previous }) {
            window.open(`${DEVOPS_BCS_API_URL}/api/logs/projects/${projectId}/clusters/${clusterId}/namespaces/${namespaceId}/pods/${podId}/stdlogs/download/?container_name=${containerName}&previous=${!!previous}`)
        },

        async stdLogsSession (context, params) {
            const data = await stdLogsSession(params).catch(() => ({}))
            return data
        }
    }
}
