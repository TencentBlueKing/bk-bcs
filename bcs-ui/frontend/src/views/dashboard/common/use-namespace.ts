import { ref, computed, SetupContext, Ref } from "@vue/composition-api"
import { ISubscribeData } from './use-subscribe'

export interface IUseNamespace {
    namespaceLoading: Ref<boolean>;
    namespaceData: Ref<ISubscribeData>;
    getNamespaceData: () => Promise<ISubscribeData>;
}

/**
 * 获取命名空间
 * @param ctx
 * @returns
 */
export default function useNamespace (ctx: SetupContext): IUseNamespace {
    const { $store } = ctx.root

    const namespaceLoading = ref(false)
    const namespaceData = ref<ISubscribeData>({
        manifest: {},
        manifest_ext: {}
    })

    const getNamespaceData = async (): Promise<ISubscribeData> => {
        namespaceLoading.value = true
        const data = await $store.dispatch('dashboard/getNamespaceList')
        namespaceData.value = data
        namespaceLoading.value = false
        return data
    }

    // onMounted(getNamespaceData)

    return {
        namespaceLoading,
        namespaceData,
        getNamespaceData
    }
}
