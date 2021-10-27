import { SetupContext, computed, onMounted, ComputedRef } from '@vue/composition-api'

export interface IUseClusterListResult {
    curClusterList: ComputedRef<any[]>;
    getClusterList: () => Promise<void>;
}

/**
 * 获取集群列表（入口文件统一获取，此处不再调用）
 * @param ctx
 * @returns
 */
export default function useClusterList (ctx: SetupContext): IUseClusterListResult {
    const { $route, $store } = ctx.root
    const projectId = computed(() => $route.params.projectId)
    const curClusterList = computed(() => $store.state.cluster.clusterList)

    const getClusterList = async () => {
        const res = await $store.dispatch('cluster/getClusterList', projectId.value).catch(() => ({ data: {} }))
        if (res.data.results && res.data.results.length) {
            $store.commit('cluster/forceUpdateClusterList', res.data.results)
        }
    }

    onMounted(() => {
        if (!curClusterList.value.length) {
            getClusterList()
        }
    })

    return {
        curClusterList,
        getClusterList
    }
}
