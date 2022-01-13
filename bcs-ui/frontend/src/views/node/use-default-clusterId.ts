import { computed } from '@vue/composition-api'
import store from '@/store'
import { BCS_CLUSTER } from '@/common/constant'

export default function useDefaultClusterId () {
    const curClusterId = computed(() => {
        return store.state.curClusterId
    })
    const clusterList = computed(() => {
        return (store.state as any).cluster.clusterList || []
    })
    // 单集群ID > sessionStorage >列表第一个
    const defaultClusterId = computed<string|undefined>(() => {
        return curClusterId.value
        || sessionStorage.getItem(BCS_CLUSTER)
        || clusterList.value[0]?.clusterID
    })

    return {
        curClusterId,
        defaultClusterId,
        clusterList
    }
}
