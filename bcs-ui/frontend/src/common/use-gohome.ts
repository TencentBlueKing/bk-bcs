import { computed } from '@vue/composition-api'
import store from '@/store'
import router from '@/router'

type ViewModeType = 'dashboard' | 'cluster'
// 调整首页逻辑
export default function useGoHome () {
    // 视图类型
    const viewMode = computed<ViewModeType>(() => {
        return store.state.viewMode as ViewModeType
    })
    // 当前集群（执行调整首页之前，需要先更新当前集群信息）
    const cluster = computed(() => {
        return store.state.cluster.curCluster
    })
    const goHome = () => {
        if (viewMode.value === 'dashboard') {
            // 资源视图首页
            router.replace({
                name: 'dashboard',
                params: {
                    clusterId: cluster.value.cluster_id
                }
            })
        } else if (!cluster.value.cluster_id) {
            // 全部集群首页
            router.replace({ name: 'clusterMain' })
        } else if (cluster.value.is_shared) {
            // 公共集群首页
            router.replace({ name: 'namespace' })
        } else {
            // 单集群
            router.replace({
                name: 'clusterOverview',
                params: {
                    clusterId: cluster.value.cluster_id
                }
            })
        }
    }

    return {
        goHome
    }
}
