<template>
    <div class="taint-wrapper" v-bkloading="{ isLoading, opacity: 1 }">
        <div class="labels">
            <span>{{$t('键')}}：</span>
            <span>{{$t('值')}}：</span>
            <span>{{$t('影响')}}：</span>
        </div>
        <div class="values">
            <div v-for="(item, index) of values"
                :key="index"
                class="value-item">
                <div class="key">
                    <bk-input v-model="item.key" />
                    <span class="symbol">=</span>
                </div>
                <div class="value">
                    <bk-input v-model="item.value" />
                    <span class="symbol" style="padding-left: 13px;">:</span>
                </div>
                <div class="effect">
                    <bcs-select v-model="item.effect"
                        :clearable="false">
                        <bcs-option v-for="effect of effectList"
                            :key="effect"
                            :id="effect"
                            :name="effect" />
                    </bcs-select>
                </div>
                <div class="btns">
                    <bk-button v-if="showAddBtn(index)"
                        text
                        icon="plus-circle"
                        @click.stop="addItem">
                    </bk-button>
                    <bk-button text
                        icon="minus-circle"
                        :disabled="length"
                        @click.stop="deleteItem(index)">
                    </bk-button>
                </div>
            </div>
        </div>
        <div class="footer">
            <bk-button theme="primary" :loading="isSubmitting" @click="handleSubmit">{{$t('确定')}}</bk-button>
            <bk-button theme="default" @click="handleCancel">{{$t('取消')}}</bk-button>
        </div>
    </div>
</template>

<script lang="ts">
    import { defineComponent, onMounted, ref, toRefs } from '@vue/composition-api'

    interface IValueItem {
        key: string;
        value: string;
        effect: string;
    }

    export default defineComponent({
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
            const getDefaultValue = () => ({
                key: '',
                value: '',
                effect: 'PreferNoSchedule'
            })
            const { $store } = ctx.root
            const { nodes, clusterId } = toRefs(props)
            const isLoading = ref<boolean>(false)
            const isSubmitting = ref<boolean>(false)
            const effectList = ref(['PreferNoSchedule', 'NoExecute', 'NoSchedule'])
            const values = ref<IValueItem[]>([getDefaultValue()])
            const showAddBtn = (index) => values.value.length - 1 === index
            const addItem = () => {
                values.value.push(getDefaultValue())
            }
            const deleteItem = (index: number) => {
                values.value.splice(index, 1)
                // 如果最后一行被删除则添加个空行
                !values.value.length && addItem()
            }
            // 提交数据
            const handleSubmit = async () => {
                isSubmitting.value = true
                try {
                    // data是单个节点设置污点的结果，多个节点需要另外处理
                    const data = []
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
                isLoading,
                isSubmitting,
                values,
                effectList,
                addItem,
                deleteItem,
                showAddBtn,
                handleSubmit,
                handleCancel
            }
        }
    })
</script>

<style lang="postcss" scoped>
@define-mixin flex-layout {
    display: flex;
    align-items: center;
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
    .values {
        .value-item {
            @mixin flex-layout;
            margin-bottom: 20px;
            .value, .key {
                @mixin flex-layout;
                flex: 0 0 calc(100% / 3 - 20px);
                .bk-form-control {
                    width: 184px;
                }
            }
            .effect {
                flex: 0 0 184px;
            }
            .symbol {
                color: #c3cdd7;
                padding: 0 10px;
            }
            .btns {
                @mixin flex-layout;
                button {
                    font-size: 0;
                    padding: 0;
                    color: #999999;
                    margin-left: 10px;
                    &:hover {
                        color: #3a84ff;
                    }
                    /deep/ .bk-icon {
                        width: 22px;
                        height: 22px;
                        line-height: 22px;
                        font-size: 22px;
                    }
                }
            }
        }
    }
    .footer {
        button {
            width: 86px;
        }
    }
}
</style>
