import { SetupContext, ref } from '@vue/composition-api'
import { ISubscribeData } from './use-subscribe'

/**
 * 加载表格数据
 * @param ctx
 * @returns
 */
export default function useTableData (ctx: SetupContext) {
    const isLoading = ref(false)
    const data = ref<ISubscribeData>({
        manifest_ext: {},
        manifest: {}
    })
    const webAnnotations = ref<any>({})

    const { $store } = ctx.root

    const fetchList = async (type: string, category: string, namespaceId: string) => {
        const action = namespaceId ? 'dashboard/getTableData' : 'dashboard/getTableDataWithoutNamespace'
        const res = await $store.dispatch(action, {
            $type: type,
            $category: category,
            $namespaceId: namespaceId
        })
        return res
    }
    const handleFetchList = async (type: string, category: string, namespaceId: string): Promise<ISubscribeData|undefined> => {
        // persistent_volumes、storage_classes资源和命名空间无关，其余资源必须传命名空间
        if (!namespaceId && !['persistent_volumes', 'storage_classes'].includes(category)) return
        isLoading.value = true
        const res = await fetchList(type, category, namespaceId)
        data.value = res.data
        webAnnotations.value = res.web_annotations || {}
        isLoading.value = false
        return res.data
    }

    const fetchCRDData = async () => {
        const res = await $store.dispatch('dashboard/crdList')
        return res
    }
    const handleFetchCustomResourceList = async (crd?: string, category?: string, namespace?: string): Promise<ISubscribeData|undefined> => {
        // crd 和 category 必须同时存在（同时不存在：crd列表，同时存在：特定类型自定义资源列表）
        if ((crd && !category) || (!crd && category)) return
        isLoading.value = true
        const res = await $store.dispatch('dashboard/customResourceList', {
            $crd: crd,
            $category: category,
            namespace
        })
        data.value = res.data
        webAnnotations.value = res.web_annotations || {}
        isLoading.value = false
        return res.data
    }

    return {
        isLoading,
        data,
        webAnnotations,
        fetchList,
        handleFetchList,
        fetchCRDData,
        handleFetchCustomResourceList
    }
}
