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

    const clusterList = computed(() => {
        return $store.state.cluster.clusterList
    })
    const curProjectId = computed(() => {
        return $store.state.curProjectId
    })
    const clusterExtraInfo = ref({})
    const webAnnotations = ref({ perms: {} })
    // 获取集群列表
    const getClusterList = async () => {
        const res = await $store.dispatch('cluster/getClusterList', curProjectId.value)
        clusterExtraInfo.value = res.clusterExtraInfo || {}
        webAnnotations.value = res.web_annotations || { perms: {} }
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
        } else {
            start()
        }
    })

    return {
        webAnnotations,
        clusterList,
        curProjectId,
        clusterExtraInfo,
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
        newClusterList.forEach(item => {
            fetchClusterOverview(item)
        })
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
    const { $store } = ctx.root
    // 集群删除
    const deleteCluster = async (cluster): Promise<boolean> => {
        const result = await $store.dispatch('clustermanager/deleteCluster', {
            $clusterId: cluster.cluster_id
        }).catch(() => false)
        return result
    }
    // 集群重试重试
    const retryCluster = async (cluster): Promise<boolean> => {
        const result = await $store.dispatch('clustermanager/retryCluster', {
            $clusterId: cluster.cluster_id
        }).catch(() => false)
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
