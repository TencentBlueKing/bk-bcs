/* eslint-disable camelcase */
import { ref, computed, SetupContext } from '@vue/composition-api'
import yamljs from 'js-yaml'

export interface IWorkloadDetail {
    manifest: any;
    manifest_ext: any;
    web_annotations?: any;
}

export interface IDetailOptions {
    category: string;
    name: string;
    namespace: string;
    type: string;
    defaultActivePanel: string;
}

export default function useDetail (ctx: SetupContext, options: IDetailOptions) {
    const { $store, $router, $bkInfo, $bkMessage, $i18n } = ctx.root
    const isLoading = ref(false)
    const detail = ref<IWorkloadDetail|null>(null)
    const activePanel = ref(options.defaultActivePanel)
    const showYamlPanel = ref(false)

    // 标签数据
    const labels = computed(() => {
        const obj = detail.value?.manifest?.metadata?.labels || {}
        return Object.keys(obj).map(key => ({
            key,
            value: obj[key]
        }))
    })
    // 注解数据
    const annotations = computed(() => {
        const obj = detail.value?.manifest?.metadata?.annotations || {}
        return Object.keys(obj).map(key => ({
            key,
            value: obj[key]
        }))
    })
    // metadata 数据
    const metadata = computed(() => detail.value?.manifest?.metadata || {})
    // manifestExt 数据
    const manifestExt = computed(() => detail.value?.manifest_ext || {})
    // yaml数据
    const yaml = computed(() => {
        return yamljs.dump(detail.value?.manifest || {})
    })
    const webAnnotations = ref<any>({})
    // 界面权限
    const pagePerms = computed(() => {
        return {
            create: webAnnotations.value?.perms?.page?.create_btn || {},
            delete: webAnnotations.value?.perms?.page?.delete_btn || {},
            update: webAnnotations.value?.perms?.page?.update_btn || {}
        }
    })

    const handleTabChange = (item) => {
        activePanel.value = item.name
    }
    // 获取workload详情
    const handleGetDetail = async () => {
        const { namespace, category, name, type } = options
        // workload详情
        isLoading.value = true
        const res = await $store.dispatch('dashboard/getResourceDetail', {
            $namespaceId: namespace,
            $category: category,
            $name: name,
            $type: type
        })
        detail.value = res.data
        webAnnotations.value = res.web_annotations
        isLoading.value = false
        return detail.value
    }

    const handleShowYamlPanel = () => {
        showYamlPanel.value = true
    }

    // 更新资源
    const handleUpdateResource = () => {
        const kind = detail.value?.manifest?.kind
        const { namespace, category, name, type } = options
        $router.push({
            name: 'dashboardResourceUpdate',
            params: {
                namespace,
                name
            },
            query: {
                type,
                category,
                kind
            }
        })
    }

    // 删除资源
    const handleDeleteResource = () => {
        const kind = detail.value?.manifest?.kind
        const { namespace, category, name, type } = options
        $bkInfo({
            type: 'warning',
            clsName: 'custom-info-confirm',
            title: $i18n.t('确认删除当前资源'),
            subTitle: $i18n.t('确认删除资源 {kind}: {name}', { kind, name }),
            defaultInfo: true,
            confirmFn: async (vm) => {
                const result = await $store.dispatch('dashboard/resourceDelete', {
                    $namespaceId: namespace,
                    $type: type,
                    $category: category,
                    $name: name
                })
                result && $bkMessage({
                    theme: 'success',
                    message: $i18n.t('删除成功')
                })
                $router.push({ name: $store.getters.curNavName })
            }
        })
    }

    return {
        isLoading,
        detail,
        activePanel,
        labels,
        annotations,
        metadata,
        manifestExt,
        yaml,
        showYamlPanel,
        pagePerms,
        handleShowYamlPanel,
        handleTabChange,
        handleGetDetail,
        handleUpdateResource,
        handleDeleteResource
    }
}
