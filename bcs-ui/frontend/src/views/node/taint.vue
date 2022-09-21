<template>
    <div class="taint-wrapper" v-bkloading="{ isLoading, opacity: 1 }">
        <template v-if="values.length">
            <div class="labels">
                <span>{{$t('键')}}：</span>
                <span>{{$t('值')}}：</span>
                <span>{{$t('影响')}}：</span>
            </div>
            <BcsTaints
                class="taints"
                :effect-options="effectList"
                :min-items="0"
                ref="taintRef"
                v-model="values">
            </BcsTaints>
        </template>
        <span
            class="add-btn mb15"
            v-else
            @click="handleAddTaint">
            <i class="bk-icon icon-plus-circle-shape mr5"></i>
            {{$t('添加')}}
        </span>
        <div class="footer">
            <bk-button theme="primary" :loading="isSubmitting" @click="handleSubmit">{{$t('确定')}}</bk-button>
            <bk-button theme="default" @click="handleCancel">{{$t('取消')}}</bk-button>
        </div>
    </div>
</template>

<script lang="ts">
    import { defineComponent, onMounted, ref, toRefs } from '@vue/composition-api'
    import BcsTaints from './new-taints.vue'

    interface IValueItem {
        key: string;
        value: string;
        effect: string;
    }

    export default defineComponent({
        components: { BcsTaints },
        props: {
            clusterId: {
                type: String,
                required: true
            },
            nodes: {
                type: Array,
                default: () => []
            }
        },
        setup (props, ctx) {
            const { $store } = ctx.root
            const { nodes, clusterId } = toRefs(props)
            const isLoading = ref<boolean>(false)
            const isSubmitting = ref<boolean>(false)
            const effectList = ref(['PreferNoSchedule', 'NoExecute', 'NoSchedule'])
            const values = ref<IValueItem[]>([])
            const taintRef = ref<any>(null)
            // 提交数据
            const handleSubmit = async () => {
                const result = taintRef.value?.validate()
                if (!result && values.value.length) return

                isSubmitting.value = true
                try {
                    // data是单个节点设置污点的结果，多个节点需要另外处理
                    const data: any[] = []
                    for (const item of values.value) {
                        // 只提交填了key的行
                        item.key && data.push(item)
                    }
                    await $store.dispatch('cluster/setNodeTaints', {
                        $clusterId: clusterId.value,
                        node_taint_list: nodes.value.map(node => {
                            return {
                                node_name: node.name,
                                taints: data
                            }
                        })
                    })
                    ctx.emit('confirm')
                } catch (e) {
                    console.log(e)
                } finally {
                    isSubmitting.value = false
                }
            }
            // 关闭弹窗
            const handleCancel = (refetch: boolean = false) => {
                ctx.emit('cancel', refetch)
            }
            const handleAddTaint = () => {
                values.value.push({ key: '', value: '', effect: 'PreferNoSchedule' })
            }
            onMounted(async () => {
                isLoading.value = true
                const data = await $store.dispatch('cluster/getNodeTaints', {
                    $clusterId: clusterId.value,
                    node_name_list: nodes.value.map(node => node.name)
                })
                isLoading.value = false
                // 单个节点取值
                /* eslint-disable */
                const curValues = data?.[nodes.value[0]?.inner_ip] || []
                if (curValues.length) {
                    values.value = curValues
                }
            })
            return {
                taintRef,
                isLoading,
                isSubmitting,
                values,
                effectList,
                handleSubmit,
                handleCancel,
                handleAddTaint
            }
        }
    })
</script>

<style lang="postcss" scoped>
@define-mixin flex-layout {
    display: flex;
    align-items: center;
}
.add-btn {
    cursor: pointer;
    background: #fff;
    border: 1px dashed #c4c6cc;
    border-radius: 2px;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 32px;
    font-size: 14px;
    &:hover {
        border-color: #3a84ff;
        color: #3a84ff;
    }
}
.taint-wrapper {
    padding: 20px;
    .labels {
        @mixin flex-layout;
        font-size: 14px;
        margin-bottom: 20px;
        > span {
            flex: 0 0 calc(100% / 3 - 20px);
        }
    }
    >>> .taints {
        .key {
            width: 190px;
        }
        .value {
            width: 200px;
        }
    }
    .footer {
        button {
            width: 86px;
        }
    }
}
</style>
