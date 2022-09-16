<template>
    <div class="key-value">
        <div class="key-value-item" v-if="showHeader">
            <span class="key desc">
                {{$t('键')}}:
                <i v-bk-tooltips="keyDesc"
                    class="ml10 bcs-icon bcs-icon-question-circle"
                    v-if="keyDesc"
                ></i>
            </span>
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
            <bcs-dropdown-menu class="key" trigger="click"
                v-if="keyAdvice.length > 0 && !item.disabled">
                <template #dropdown-trigger>
                    <Validate
                        :rules="rules"
                        :value="item.key"
                        :meta="index">
                        <bcs-input :placeholder="$t('键')"
                            :disabled="item.disabled"
                            v-model="item.key">
                        </bcs-input>
                    </Validate>
                </template>
                <template #dropdown-content>
                    <ul class="bk-dropdown-list">
                        <li v-for="(advice, i) in keyAdvice" :key="i"
                            @click="handleAdvice(advice, item)">
                            <a href="javascript:;"
                                v-bk-tooltips="{
                                    content: advice.desc,
                                    disabled: !advice.desc,
                                    placement: 'right',
                                    boundary: 'window'
                                }"
                            >{{advice.name}}</a>
                        </li>
                    </ul>
                </template>
            </bcs-dropdown-menu>
            <Validate
                class="key"
                :rules="rules"
                :value="item.key"
                :meta="index"
                v-else>
                <bcs-input :placeholder="$t('键')"
                    :disabled="item.disabled"
                    v-model="item.key">
                </bcs-input>
            </Validate>
            <span class="equals-sign">=</span>
            <bcs-input :placeholder="item.placeholder || $t('值')" class="value" v-model="item.value"></bcs-input>
            <i class="bk-icon icon-plus-circle ml10 mr5" @click="handleAddKeyValue(index)"></i>
            <i :class="['bk-icon icon-minus-circle', { disabled: keyValueData.length === 1 }]"
                @click="handleDeleteKeyValue(index)"
            ></i>
        </div>
        <div class="mt30" v-if="showFooter">
            <bcs-button class="bcs-btn"
                theme="primary"
                :loading="loading"
                :disalbed="loading"
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
<script lang="ts">
    import { defineComponent, toRefs, watch, ref, computed } from '@vue/composition-api'
    import Validate from './validate.vue'
    import $i18n from '@/i18n/i18n-setup'

    export interface IData {
        key: string;
        value: string;
        placeholder?: any;
        disabled?: boolean;
    }
    export default defineComponent({
        components: { Validate },
        props: {
            modelValue: {
                type: [Object, Array],
                default: []
            },
            valueDesc: {
                type: String,
                default: ''
            },
            keyDesc: {
                type: String,
                default: ''
            },
            loading: {
                type: Boolean,
                default: false
            },
            showFooter: {
                type: Boolean,
                default: true
            },
            showHeader: {
                type: Boolean,
                default: true
            },
            keyAdvice: {
                type: Array,
                default: () => []
            },
            keyRules: {
                type: Array,
                default: () => []
            }
        },
        model: {
            prop: 'modelValue',
            event: 'change'
        },
        setup (props, ctx) {
            const { modelValue, keyRules } = toRefs(props)
            const keyValueData = ref<IData[]>([])
            watch(modelValue, () => {
                if (Array.isArray(modelValue.value)) {
                    keyValueData.value = modelValue.value.map((item: any) => ({
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
                if (!keyValueData.value.length) {
                    keyValueData.value.push({
                        key: '',
                        value: ''
                    })
                }
            }, { immediate: true })

            const handleAddKeyValue = (index) => {
                keyValueData.value.splice(index + 1, 0, {
                    key: '',
                    value: ''
                })
            }
            const handleDeleteKeyValue = (index) => {
                if (keyValueData.value.length === 1) return
                keyValueData.value.splice(index, 1)
            }
            const labels = computed(() => {
                return keyValueData.value.filter(item => !!item.key).reduce((pre, curLabelItem) => {
                    pre[curLabelItem.key] = curLabelItem.value
                    return pre
                }, {})
            })
            const confirmSetLabel = async () => {
                if (!validate()) return
                ctx.emit('confirm', labels.value)
            }
            const hideSetLabel = () => {
                ctx.emit('cancel', labels.value)
            }
            // key联想功能
            const handleAdvice = (advice, item) => {
                item.key = advice.name
                item.value = advice.default
            }
            const rules = ref(
                [
                    ...keyRules.value,
                    {
                        message: $i18n.t('重复键'),
                        validator: (value, index) => {
                            return keyValueData.value.filter((_, i) => i !== index).every(d => d.key !== value)
                        }
                    }
                ]
            )
            const validate = () => {
                const data = keyValueData.value.reduce<string[]>((pre, item) => {
                    if (item.key) {
                        pre.push(item.key)
                    }
                    return pre
                }, [])
                const removeDuplicateData = new Set(data)
                if (data.length !== removeDuplicateData.size) {
                    return false
                }
                
                return data.every(key => {
                    return keyRules.value.every(rule => {
                        return new RegExp(rule.validator).test(key)
                    })
                })
            }

            return {
                rules,
                labels,
                keyValueData,
                validate,
                confirmSetLabel,
                hideSetLabel,
                handleAddKeyValue,
                handleDeleteKeyValue,
                handleAdvice
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
    .bk-icon.disabled {
        color: #DCDEE5;
        cursor: not-allowed;
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
