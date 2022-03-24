<template>
    <bk-exception :type="type">
        <span>{{$t('该操作需要以下权限')}}</span>
        <bk-table :data="tableData" class="mt25" v-bkloading="{ isLoading }">
            <bk-table-column :label="$t('系统')" prop="system" min-width="150">
                <template>
                    {{ $t('容器管理平台') }}
                </template>
            </bk-table-column>
            <bk-table-column :label="$t('需要申请的权限')" prop="auth" min-width="220">
                <template slot-scope="{ row }">
                    {{ actionsMap[row.action_id] || '--' }}
                </template>
            </bk-table-column>
            <bk-table-column :label="$t('关联的资源实例')" prop="resource" min-width="220">
                <template slot-scope="{ row }">
                    {{ row.resource_name || '--' }}
                </template>
            </bk-table-column>
        </bk-table>
        <bk-button theme="primary"
            class="mt25"
            :disabled="!href"
            @click="handleGotoIAM"
        >{{$t('去申请')}}</bk-button>
    </bk-exception>
</template>
<script>
    import { defineComponent, onBeforeMount, ref, computed } from '@vue/composition-api'
    import { userPermsByAction } from '@/api/base'
    import actionsMap from '@/components/apply-perm/actions-map'

    export default defineComponent({
        props: {
            type: {
                type: String,
                default: '403'
            },
            actionId: {
                type: String,
                default: ''
            },
            resourceName: {
                type: String,
                default: ''
            },
            permCtx: {
                type: [Object, String],
                default: () => ({})
            },
            fromRoute: {
                type: String,
                default: ''
            }
        },
        setup (props, ctx) {
            const { $store } = ctx.root
            const tableData = ref([])
            const href = ref('')
            const isLoading = ref(false)
            const handleGotoIAM = () => {
                window.open(href.value)
            }
            const projectList = computed(() => {
                return $store.state.sideMenu.onlineProjectList
            })
            onBeforeMount(async () => {
                if (!props.actionId) return
                isLoading.value = true
                const data = await userPermsByAction({
                    $actionId: [props.actionId],
                    perm_ctx: typeof props.permCtx === 'string'
                        ? JSON.parse(props.permCtx)
                        : props.permCtx
                }).catch(() => ({}))
                isLoading.value = false
                if (data?.perms?.[props.actionId] && props.fromRoute && projectList.value.length) {
                    window.location.href = props.fromRoute
                } else {
                    // eslint-disable-next-line camelcase
                    href.value = data?.perms?.apply_url
                    tableData.value = [{
                        resource_name: props.resourceName,
                        action_id: props.actionId
                    }]
                }
            })
            return {
                handleGotoIAM,
                actionsMap,
                tableData,
                isLoading,
                href
            }
        }
    })
</script>
