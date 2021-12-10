import { ref, SetupContext, Ref, computed } from "@vue/composition-api"
import { ISubscribeData } from './use-subscribe'
import { CUR_SELECT_NAMESPACE } from '@/common/constant'
export interface IUseNamespace {
    namespaceLoading: Ref<boolean>;
    namespaceData: Ref<ISubscribeData>;
    namespaceValue: Ref<string>;
    namespaceList: Ref<any[]>;
    getNamespaceData: () => Promise<ISubscribeData>;
}

/**
 * 获取命名空间
 * @param ctx
 * @returns
 */
export default function useNamespace (ctx: SetupContext): IUseNamespace {
    const { $store } = ctx.root

    const namespaceValue = ref('')
    const namespaceLoading = ref(false)
    const namespaceData = ref<ISubscribeData>({
        manifest: {},
        manifest_ext: {}
    })
    // 命名空间数据
    const namespaceList = computed(() => {
        return namespaceData.value.manifest.items || []
    })

    const getNamespaceData = async (): Promise<ISubscribeData> => {
        namespaceLoading.value = true
        const data = await $store.dispatch('dashboard/getNamespaceList')
        namespaceData.value = data
        // 初始化默认选中命名空间
        const defaultSelectNamespace = namespaceList.value.find(data => data.metadata.name === sessionStorage.getItem(CUR_SELECT_NAMESPACE))
        namespaceValue.value = defaultSelectNamespace?.metadata?.name || namespaceList.value[0]?.metadata?.name
        sessionStorage.setItem(CUR_SELECT_NAMESPACE, namespaceValue.value)
        namespaceLoading.value = false
        return data
    }

    // onMounted(getNamespaceData)

    return {
        namespaceLoading,
        namespaceData,
        namespaceValue,
        namespaceList,
        getNamespaceData
    }
}
