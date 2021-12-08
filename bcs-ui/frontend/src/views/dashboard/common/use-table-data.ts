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
        if (!type || !category || !namespaceId) return
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
        if (!crd || !category) return
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
