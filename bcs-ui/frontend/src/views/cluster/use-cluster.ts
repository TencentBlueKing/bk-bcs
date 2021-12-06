/* eslint-disable camelcase */
import { SetupContext, computed, ref, watch, Ref, set } from '@vue/composition-api'
import useInterval from '@/views/dashboard/common/use-interval'

/**
 * 获取集群列表
 * @param ctx
 * @returns
 */
export function useClusterList (ctx: SetupContext) {
    const { $store } = ctx.root

    const clusterList = ref<any[]>([])
    const clusterPerm = ref({})
    const curProjectId = computed(() => {
        return $store.state.curProjectId
    })
    // 获取集群列表
    const getClusterList = async () => {
        const res = await $store.dispatch('cluster/getClusterList', curProjectId.value)
        clusterList.value = res.data.map(item => {
            return {
                cluster_id: item.clusterID,
                name: item.clusterName,
                project_id: item.projectID,
                ...item
            }
        })
        clusterPerm.value = res.clusterPerm
        // 更新全局集群列表信息
        $store.commit('cluster/forceUpdateClusterList', res.data)
    }
    // 开启轮询
    const { start, stop } = useInterval(getClusterList, 5000)
    const runningClusterIds = computed(() => {
        return clusterList.value.filter(item => [
            'INITIALIZATION',
            'DELETING'].includes(item.status)).map(item => item.cluster_id)
    })
    watch(runningClusterIds, (newValue, oldValue) => {
        if (!newValue.length) {
            stop()
        } else if (newValue.sort().join() !== oldValue.sort().join()) {
            start()
        }
    })

    return {
        clusterList,
        clusterPerm,
        curProjectId,
        getClusterList
    }
}
export interface IOverviewMap {
    [key: string]: {
        cpu_usage: any;
        disk_usage: any;
        memory_usage: any;
    };
}
/**
 * 获取集群指标数据
 * @param ctx
 * @param clusterList
 * @returns
 */
export function useClusterOverview (ctx: SetupContext, clusterList: Ref<any[]>) {
    const { $store } = ctx.root

    const clusterOverviewMap = ref<IOverviewMap>({})
    // 获取当前集群的指标信息
    const getClusterOverview = (clusterId) => {
        if (!clusterOverviewMap.value[clusterId]) return null

        return clusterOverviewMap.value[clusterId]
    }
    // 集群指标信息
    const fetchClusterOverview = async (cluster) => {
        const res = await $store.dispatch('cluster/clusterOverview', {
            projectId: cluster.project_id,
            clusterId: cluster.cluster_id
        }).catch(() => ({ data: {} }))
        set(clusterOverviewMap.value, cluster.cluster_id, res.data)
        return res.data
    }

    watch(clusterList, (newValue, oldValue) => {
        const newClusterList = newValue.filter(item => item.status === 'RUNNING' && !clusterOverviewMap.value?.[item.cluster_id])
        const oldClusterList = oldValue.filter(item => item.status === 'RUNNING' && !clusterOverviewMap.value?.[item.cluster_id])

        const newClusterIds = newClusterList.map(item => item.cluster_id)
        const oldClusterIds = oldClusterList.map(item => item.cluster_id)

        if (newClusterIds.sort().join() !== oldClusterIds.sort().join()) {
            newClusterList.forEach(item => {
                fetchClusterOverview(item)
            })
        }
    })

    return {
        clusterOverviewMap,
        getClusterOverview
    }
}
/**
 * 集群操作
 * @param ctx
 * @returns
 */
export function useClusterOperate (ctx: SetupContext) {
    const { $store, $bkMessage, $i18n } = ctx.root
    // 集群删除
    const deleteCluster = async (cluster): Promise<boolean> => {
        const result = await $store.dispatch('clustermanager/deleteCluster', {
            $clusterId: cluster.cluster_id
        }).catch(() => false)
        result && $bkMessage({
            theme: 'success',
            message: $i18n.t('删除成功')
        })
        return result
    }
    // 集群重试重试
    const retryCluster = async (cluster): Promise<boolean> => {
        const result = await $store.dispatch('clustermanager/retryCluster', {
            $clusterId: cluster.cluster_id
        }).catch(() => false)
        result && $bkMessage({
            theme: 'success',
            message: $i18n.t('重试成功')
        })
        return result
    }
    // 获取集群下面所有的节点
    const clusterNode = async (cluster) => {
        const data = await $store.dispatch('clustermanager/clusterNode', {
            $clusterId: cluster.clusterID
        })
        return data
    }

    return {
        deleteCluster,
        retryCluster,
        clusterNode
    }
}

/**
 * 任务操作
 * @param ctx
 */
export function useTask (ctx: SetupContext) {
    const { $store } = ctx.root
    // 查询任务列表
    const taskList = async (cluster) => {
        const data = await $store.dispatch('clustermanager/taskList', {
            clusterID: cluster.clusterID,
            projectID: cluster.projectID
        })
        return data
    }
    return {
        taskList
    }
}
