<template>
    <div>
        <div class="key-value" v-for="(item, index) in labels" :key="index">
            <bcs-input v-model="item.key"
                :placeholder="keyPlaceholder"
                @change="handleLabelKeyChange">
            </bcs-input>
            <span class="ml8 mr8">=</span>
            <bcs-input v-model="item.value"
                :placeholder="valuePlaceholder"
                @change="handleLabelValueChange"
            ></bcs-input>
            <i class="bk-icon icon-plus-circle ml15" @click="handleAddLabel(index)"></i>
            <i :class="['bk-icon icon-minus-circle ml10', { disabled: disabledDelete }]"
                @click="handleDeleteLabel(index)">
            </i>
        </div>
    </div>
</template>
<script lang="ts">
    import { computed, defineComponent, ref, watch, toRefs } from '@vue/composition-api'

    interface ILabel {
        key: string;
        value: string;
    }
    export default defineComponent({
        name: 'KeyValue',
        props: {
            value: {
                type: Object,
                default: () => ({})
            },
            keyPlaceholder: {
                type: String,
                default: ''
            },
            valuePlaceholder: {
                type: String,
                default: ''
            },
            minItem: {
                type: Number,
                default: 1
            }
        },
        model: {
            prop: 'value',
            event: 'change'
        },
        setup (props, ctx) {
            const labels = ref<ILabel[]>([])
            const { value } = toRefs(props)
            const unWatchValue = watch(value, () => {
                labels.value = Object.keys(props.value).map(key => ({
                    key,
                    value: props.value[key]
                }))
                if (labels.value.length < (props.minItem || 0)) {
                    const reset: ILabel[] = new Array(props.minItem - labels.value.length).fill({ key: '', value: '' })
                    labels.value.push(...reset)
                }
            }, { immediate: true, deep: true })
            watch(labels, () => {
                unWatchValue()
                emitChange()
            }, { deep: true })
            const disabledDelete = computed(() => {
                return labels.value.length <= props.minItem
            })
            const emitChange = () => {
                const keyValues = labels.value.reduce((pre, item) => {
                    if (!item.key) return pre

                    pre[item.key] = item.value
                    return pre
                }, {})
                ctx.emit('change', keyValues)
            }
            const handleLabelKeyChange = (newValue, oldValue) => {
                ctx.emit('key-change', newValue, oldValue)
            }
            const handleLabelValueChange = (newValue, oldValue) => {
                ctx.emit('value-change', newValue, oldValue)
            }
            const handleAddLabel = (index) => {
                labels.value.splice(index + 1, 0, { key: '', value: '' })
            }
            const handleDeleteLabel = (index) => {
                if (disabledDelete.value) return
                labels.value.splice(index, 1)
            }
            return {
                labels,
                disabledDelete,
                handleLabelKeyChange,
                handleLabelValueChange,
                handleAddLabel,
                handleDeleteLabel
            }
        }
    })
</script>
<style lang="postcss" scoped>
.key-value {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
    .bk-icon {
        font-size: 20px;
        color: #979bA5;
        cursor: pointer;
        &.disabled {
            color: #DCDEE5;
            cursor: not-allowed;
        }
    }
    .ml8 {
        margin-left: 8px;
    }
    .mr8 {
        margin-right: 8px;
    }
}
</style>
