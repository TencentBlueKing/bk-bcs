<template>
    <div class="key-value">
        <div class="key-value-item">
            <span class="key">{{$t('键')}}:</span>
            <span class="value desc">
                {{$t('值')}}:
                <i v-bk-tooltips="valueDesc"
                    class="ml10 bcs-icon bcs-icon-question-circle"
                    v-if="valueDesc"
                ></i>
            </span>
        </div>
        <div v-for="(item, index) in keyValueData"
            :key="index"
            class="key-value-item"
        >
            <bcs-input :placeholder="$t('键')" :disabled="item.disabled" class="key" v-model="item.key"></bcs-input>
            <span class="equals-sign">=</span>
            <bcs-input :placeholder="item.placeholder || $t('值')" class="value" v-model="item.value"></bcs-input>
            <i class="bk-icon icon-plus-circle ml10 mr5" @click="handleAddKeyValue(index)"></i>
            <i class="bk-icon icon-minus-circle" v-if="keyValueData.length > 1" @click="handleDeleteKeyValue(index)"></i>
        </div>
        <div class="mt30">
            <bcs-button class="bcs-btn"
                theme="primary"
                :loading="loading"
                @click="confirmSetLabel"
            >
                {{$t('保存')}}
            </bcs-button>
            <bcs-button class="bcs-btn" :disalbed="loading" @click="hideSetLabel">
                {{$t('取消')}}
            </bcs-button>
        </div>
    </div>
</template>
<script>
    import { defineComponent, toRefs, watch, ref } from '@vue/composition-api'

    export default defineComponent({
        props: {
            modelValue: {
                type: [Object, Array],
                default: []
            },
            valueDesc: {
                type: String,
                default: ''
            },
            loading: {
                type: Boolean,
                default: false
            }
        },
        model: {
            prop: 'modelValue',
            event: 'change'
        },
        setup (props, ctx) {
            const { modelValue } = toRefs(props)
            const keyValueData = ref([])
            watch(modelValue, () => {
                if (Array.isArray(modelValue.value)) {
                    keyValueData.value = modelValue.value.map(item => ({
                        ...item,
                        disabled: true
                    }))
                } else {
                    keyValueData.value = Object.keys(modelValue.value).map(key => {
                        return {
                            key,
                            value: modelValue.value[key],
                            disabled: true
                        }
                    })
                }
                // 添加一组空值
                keyValueData.value.push({
                    key: '',
                    value: ''
                })
            }, { immediate: true })

            const handleAddKeyValue = (index) => {
                keyValueData.value.splice(index + 1, 0, {
                    key: '',
                    value: ''
                })
            }
            const handleDeleteKeyValue = (index) => {
                keyValueData.value.splice(index, 1)
            }
            const confirmSetLabel = () => {
                ctx.emit('confirm', keyValueData.value.filter(item => !!item.key))
            }
            const hideSetLabel = () => {
                ctx.emit('cancel', keyValueData.value.filter(item => !!item.key))
            }
            return {
                keyValueData,
                confirmSetLabel,
                hideSetLabel,
                handleAddKeyValue,
                handleDeleteKeyValue
            }
        }
    })
</script>
<style lang="postcss" scoped>
.key-value-item {
    display: flex;
    align-items: center;
    height: 32px;
    line-height: 32px;
    margin-bottom: 10px;
    font-size: 14px;
    .key {
        flex: 1;
    }
    .value {
        flex: 1;
    }
    .desc {
        display: flex;
        align-items: center;
    }
    .bk-icon {
        font-size: 24px;
        color: #979bA5;
        cursor: pointer;
    }
    .equals-sign {
        color: #c3cdd7;
        margin: 0 15px;
    }
}
.bcs-btn {
    width: 86px;
}
</style>
