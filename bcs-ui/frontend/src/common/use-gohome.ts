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
    const goHome = ($route) => {
        if (viewMode.value === 'dashboard' && $route.name !== 'dashboard') {
            // 资源视图首页
            router.replace({
                name: 'dashboard',
                params: {
                    clusterId: cluster.value.cluster_id
                }
            })
        } else if (!cluster.value.cluster_id && $route.name !== 'clusterMain') {
            // 全部集群首页
            router.replace({ name: 'clusterMain' })
        } else if (cluster.value.is_shared && $route.name !== 'namespace') {
            // 公共集群首页
            router.replace({ name: 'namespace' })
        } else if (cluster.value.cluster_id && !cluster.value.is_shared && $route.name !== 'clusterOverview') {
            // 单集群首页
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
