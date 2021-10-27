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

    const fetchList = async (type: string, category: string) => {
        const res = await $store.dispatch('dashboard/getTableData', {
            $type: type,
            $category: category
        })
        return res
    }
    const handleFetchList = async (type: string, category: string): Promise<ISubscribeData> => {
        isLoading.value = true
        const res = await fetchList(type, category)
        data.value = res.data
        webAnnotations.value = res.web_annotations || {}
        isLoading.value = false
        return res.data
    }

    const fetchCustomResourceList = async (crd?: string, category?: string) => {
        const res = await $store.dispatch('dashboard/customResourceList', {
            $crd: crd,
            $category: category
        })
        return res
    }
    const handleFetchCustomResourceList = async (crd?: string, category?: string): Promise<ISubscribeData> => {
        // crd和category必须同时存在，否则返回空数据
        if ((!crd && category) || (crd && !category)) {
            return {
                manifest_ext: {},
                manifest: {}
            }
        }
        isLoading.value = true
        const res = await fetchCustomResourceList(crd, category)
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
        fetchCustomResourceList,
        handleFetchCustomResourceList
    }
}
