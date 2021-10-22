<template>
    <div class="bk-keyer">
        <div class="biz-keys-list mb10">
            <div class="biz-key-item" v-for="(keyItem, index) in list" :key="index">
                <template v-if="varList.length">
                    <bkbcs-input
                        type="text"
                        :placeholder="keyPlaceholder || $t('键')"
                        style="flex-basis: 40%;"
                        :value.sync="keyItem.key"
                        :list="varList"
                        :disabled="keyItem.disabled && !keyItem.linkMessage"
                        @input="valueChange"
                        @paste="pasteKey(keyItem, $event)">
                    </bkbcs-input>
                </template>
                <template v-else>
                    <input
                        type="text"
                        class="bk-form-input key"
                        :placeholder="keyPlaceholder || $t('键')"
                        v-model="keyItem.key"
                        :disabled="keyItem.disabled && !keyItem.linkMessage"
                        @paste="pasteKey(keyItem, $event)"
                        @input="valueChange"
                    />
                </template>

                <span class="operator">=</span>

                <template v-if="varList.length">
                    <bkbcs-input
                        type="text"
                        :placeholder="valuePlaceholder || $t('值')"
                        style="flex-basis: 60%;"
                        :value.sync="keyItem.value"
                        :list="varList"
                        :disabled="keyItem.disabled && !keyItem.linkMessage"
                        @input="valueChange"
                    >
                    </bkbcs-input>
                </template>
                <template v-else>
                    <input
                        type="text"
                        class="bk-form-input value"
                        :placeholder="valuePlaceholder || $t('值')"
                        v-model="keyItem.value"
                        @input="valueChange"
                        :disabled="keyItem.disabled && !keyItem.linkMessage"
                    />
                </template>

                <bk-button class="action-btn" @click.stop.prevent="addKey">
                    <i class="bcs-icon bcs-icon-plus"></i>
                </bk-button>
                <bk-button class="action-btn" v-if="list.length > 1" @click.stop.prevent="removeKey(keyItem, index)">
                    <i class="bcs-icon bcs-icon-minus"></i>
                </bk-button>
                <bk-checkbox class="ml20" v-if="isLinkToSelector" v-model="keyItem.isSelector" @change="valueChange">
                    {{addToSelectorStr || $t('添加至选择器')}}
                </bk-checkbox>
                <div v-if="keyItem.linkMessage" class="biz-tip mt5 f12">{{keyItem.linkMessage}}</div>
            </div>
        </div>
    </div>
</template>

<script>
    export default {
        name: 'metric-keyer',
        props: {
            keyList: {
                type: Array,
                default: () => []
            },
            tip: {
                type: String,
                default: ''
            },
            isTipChange: {
                type: Boolean,
                default: false
            },
            isLinkToSelector: {
                type: Boolean,
                default: false
            },
            varList: {
                type: Array,
                default: () => []
            },
            keyPlaceholder: {
                type: String,
                default: ''
            },
            valuePlaceholder: {
                type: String,
                default: ''
            },
            addToSelectorStr: {
                type: String,
                default: ''
            }
        },
        data () {
            return {
                list: this.keyList
            }
        },
        watch: {
            'keyList' (val) {
                if (this.keyList && this.keyList.length) {
                    this.list = this.keyList
                } else {
                    this.list = [{
                        key: '',
                        value: ''
                    }]
                }
            }
        },
        methods: {
            addKey () {
                const params = {
                    key: '',
                    value: ''
                }
                if (this.isLinkToSelector) {
                    params.isSelector = false
                }
                this.list.push(params)
                const obj = this.getKeyObject(true)
                this.$emit('change', this.list, obj)
            },
            removeKey (item, index) {
                this.list.splice(index, 1)
                const obj = this.getKeyObject(true)
                this.$emit('change', this.list, obj)
            },
            valueChange () {
                this.$nextTick(() => {
                    const obj = this.getKeyObject(true)
                    this.$emit('change', this.list, obj)
                })
            },
            pasteKey (item, event) {
                const cache = item.key
                const clipboard = event.clipboardData
                const text = clipboard.getData('Text')

                if (text && text.indexOf('=') > -1) {
                    this.paste(event)
                    item.key = cache
                    setTimeout(() => {
                        item.key = cache
                    }, 0)
                }
            },
            paste (event) {
                const clipboard = event.clipboardData
                const text = clipboard.getData('Text')
                const items = text.split('\n')
                items.forEach(item => {
                    if (item.indexOf('=') > -1) {
                        const arr = item.split('=')
                        this.list.push({
                            key: arr[0],
                            value: arr[1]
                        })
                    }
                })
                setTimeout(() => {
                    this.formatData()
                }, 10)

                return false
            },
            formatData () {
                // 去掉空值
                if (this.list.length) {
                    const results = []
                    const keyObj = {}
                    const length = this.list.length
                    this.list.forEach((item, i) => {
                        if (item.key || item.value) {
                            if (!keyObj[item.key]) {
                                results.push(item)
                                keyObj[item.key] = true
                            }
                        }
                    })
                    const patchLength = results.length - length
                    if (patchLength > 0) {
                        for (let i = 0; i < patchLength; i++) {
                            results.push({
                                key: '',
                                value: ''
                            })
                        }
                    }
                    this.list.splice(0, this.list.length, ...results)
                    this.$emit('change', this.list)
                }
            },
            getKeyList (isAll) {
                let results = []
                const list = this.list
                if (isAll) {
                    return this.list
                } else {
                    results = list.filter(item => {
                        return item.key && item.value
                    })
                }

                return results
            },
            getKeyObject (isAll) {
                const results = this.getKeyList(isAll)
                if (results.length === 0) {
                    return {}
                } else {
                    const obj = {}
                    results.forEach(item => {
                        if (isAll) {
                            obj[item.key] = item.value
                        } else if (item.key && item.value) {
                            obj[item.key] = item.value
                        }
                    })
                    return obj
                }
            }
        }
    }
</script>

<style scoped lang="postcss">
    .biz-keys-list {
        .biz-key-item {
            margin-bottom: 10px;
            display: flex;
        }

        .bk-form-input {
            /* width: 150px; */
            &.key {
                flex-basis: 40%;
            }
            &.value {
                flex-basis: 60%;
            }
        }

        .bk-dropdown-box {
            margin-right: 0;
        }

        .operator {
            height: 36px;
            line-height: 36px;
            text-align: center;
            display: inline-block;
            font-size: 18px;
            padding: 0 10px;
        }

        .text {
            height: 36px;
            line-height: 36px;
            text-align: center;
            display: inline-block;
            font-size: 14px;
            padding: 0 10px;
        }
    }

    .action-btn {
        height: 36px;
        text-align: center;
        display: inline-block;
        border: none;
        background: transparent;
        outline: none;

        .bcs-icon {
            width: 24px;
            height: 24px;
            line-height: 24px;
            border-radius: 50%;
            vertical-align: middle;
            border: 1px solid #dde4eb;
            color: #737987;
            font-size: 14px;
            display: inline-block;

            &.icon-minus {
                font-size: 15px;
            }
        }
    }

    .biz-keys-list .action-btn {
        width: auto;
        padding: 0;
        margin-left: 5px;
        &.disabled {
            cursor: default;
            color: #ddd !important;
            border-color: #ddd !important;
            .bcs-icon {
                color: #ddd !important;
                border-color: #ddd !important;
            }
        }
        &:hover {
            color: #3a84ff;
            border-color: #3a84ff;
            .bcs-icon {
                color: #3a84ff;
                border-color: #3a84ff;
            }
        }
    }
    .is-danger {
        color: #ff5656;
    }
</style>
