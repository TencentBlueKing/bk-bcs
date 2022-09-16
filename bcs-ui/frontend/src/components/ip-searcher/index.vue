<template>
    <div class="biz-ip-searcher-wrapper">
        <div :class="['biz-searcher', { 'active': isEdit }, { 'disable': disable }]" ref="scrollNode" @click="foucusSearcher">
            <ul class="search-keys">
                <template v-if="!isEdit && !disable && (!searchParams || !searchParams.length)">
                    <li class="placeholder">
                        {{$t('请输入IP，按Enter搜索')}}
                    </li>
                </template>
                <template v-else>
                    <li class="key-node" v-for="(param, index) in searchParams" :key="index">
                        <span class="tag">{{param.text}}</span>
                        <a href="javascript:void(0)" class="remove-key" @click.stop.prevent="removeSearchParams(param, index)">
                            <i class="bcs-icon bcs-icon-close"></i>
                        </a>
                    </li>
                </template>
                <li>
                    <input
                        type="text"
                        class="input"
                        v-model="curInputValue"
                        :style="inputStyle"
                        :disabled="disable"
                        @blur="inputBlur"
                        @paste="paste($event)"
                        @keyup="inputKeyup">
                </li>
            </ul>
        </div>
        <div class="actions">
            <bk-popover :content="$t('清空')" placement="top">
                <template v-if="disable">
                    <a href="javascript:void(0)" class="btn clear-btn disable-btn" v-show="searchParams.length">
                        <i class="bcs-icon bcs-icon-close"></i>
                    </a>
                </template>
                <template v-else>
                    <a href="javascript:void(0)" class="btn clear-btn" @click.stop.prevent="clearSearchParams" v-show="searchParams.length">
                        <i class="bcs-icon bcs-icon-close"></i>
                    </a>
                </template>
            </bk-popover>
            <bk-popover :content="$t('搜索')" placement="top">
                <template v-if="disable">
                    <a href="javascript:void(0)" class="btn search-btn disable-btn">
                        <i class="bcs-icon bcs-icon-search"></i>
                    </a>
                </template>
                <template v-else>
                    <a href="javascript:void(0)" class="btn search-btn" @click.stop.prevent="search">
                        <i class="bcs-icon bcs-icon-search"></i>
                    </a>
                </template>
            </bk-popover>
        </div>
        <div class="ip-searcher-footer" v-show="showTip && !disable">
            <p class="placeholder">{{placeholderRender}}</p>
        </div>
    </div>
</template>

<script>
    export default {
        props: {
            placeholder: {
                type: String,
                default: ''
            },
            disable: {
                type: Boolean,
                default: false
            },
            searchParams: {
                type: Array,
                default: () => []
            }
        },
        data () {
            return {
                timer: 0,
                isEdit: false,
                curParams: {},
                curInputValue: '',
                showTip: false,
                scrollNode: null,
                placeholderRender: ''
            }
        },
        computed: {
            inputStyle () {
                const inputValue = this.curInputValue
                const charLen = this.getCharLength(inputValue) + 1
                return { width: charLen * 20 + 'px' }
            }
        },
        created () {
            this.placeholderRender = this.placeholder || this.$t('多个IP以空格分开，按回车搜索')
        },
        mounted () {
            this.scrollNode = this.$refs.scrollNode
        },
        methods: {
            /**
             * 获取搜索框中字符长度
             *
             * @param {string} str 搜索框里内容
             *
             * @return {number} 长度
             */
            getCharLength (str) {
                const len = str.length
                let bitLen = 0
                for (let i = 0; i < len; i++) {
                    if ((str.charCodeAt(i) & 0xff00) !== 0) {
                        bitLen++
                    }
                    bitLen++
                }
                return bitLen
            },

            /**
             * 输入框 keyup 事件
             *
             * @param {Object} e 事件对象
             */
            inputKeyup (e) {
                switch (e.keyCode) {
                    // enter
                    case 13:
                        this.search()
                        break
                    // Backspace
                    case 8:
                        this.removeSearchLastParams()
                        break
                    // Space
                    case 32:
                        this.setCurParams()
                        break
                    default:
                }
            },

            /**
             * 将搜索框的输入设置到 params 中
             */
            setCurParams () {
                const inputVal = this.curInputValue.trim()
                if (!inputVal) {
                    return
                }
                this.curParams.text = inputVal

                const params = {
                    text: this.curParams.text
                }
                const index = this.searchParams.length
                this.searchParams.splice(index, 1, params)

                this.clearInputParams()

                this.$nextTick(() => {
                    this.scrollNode.scrollTo(this.scrollNode.offsetWidth + this.scrollNode.scrollWidth, 0)
                })
            },

            /**
             * 清除搜索框的输入
             */
            clearInputParams () {
                this.curInputValue = ''
            },

            /**
             * 清空搜索框的输入以及之前设置过的 params
             */
            clearSearchParams () {
                this.isEdit = false
                this.curParams = {}
                this.curInputValue = ''
                this.searchParams.splice(0, this.searchParams.length)
                this.$emit('clear', this)
                this.$emit('search', [])
            },

            /**
             * 触发搜索
             */
            search () {
                this.setCurParams()
                this.triggerSearch()
                this.inputBlur()
                this.$el.querySelector('.input').blur()
            },

            triggerSearch () {
                const results = {}
                this.searchParams.forEach((params) => {
                    results[params.text] = 1
                })
                this.$emit('search', Object.keys(results))
            },

            /**
             * 删除最后一个 params，用于 Backspace 按键
             */
            removeSearchLastParams () {
                if (this.curInputValue === '') {
                    this.searchParams.pop()
                }
            },

            /**
             * 删除当前点击的这个 param
             *
             * @param {Object} data 当前点击的 param 对象
             * @param {number} index 当前点击的 param 对象的索引
             */
            removeSearchParams (data, index) {
                this.searchParams.splice(index, 1)
                this.foucusSearcher()
                this.triggerSearch()
                if (this.searchParams.length === 0) {
                    this.$emit('clear', this)
                }
            },

            /**
             * 搜索框 blur 事件
             */
            inputBlur () {
                this.timer = setTimeout(() => {
                    this.curParams = {}
                    this.isEdit = false
                    this.showTip = false
                }, 100)
            },

            /**
             * 组件 click 事件
             */
            foucusSearcher () {
                clearTimeout(this.timer)
                this.isEdit = true
                if (!this.curParams) {
                    this.curParams = {
                        text: ''
                    }
                }
                this.showTip = true
                this.$nextTick(() => {
                    this.$el.querySelector('.input').focus()
                })
            },

            /**
             * 粘贴事件
             *
             * @param {Object} e 事件对象
             */
            paste (e) {
                const clipboard = event.clipboardData
                if (!clipboard) {
                    return
                }
                const text = clipboard.getData('Text')
                if (text) {
                    const items = text.split(/\s+|\n+/)
                    items.forEach(t => {
                        if (t.trim()) {
                            const index = this.searchParams.length
                            this.searchParams.splice(index, 1, {
                                text: t
                            })
                        }
                    })
                    setTimeout(() => {
                        this.clearInputParams()
                        this.$nextTick(() => {
                            this.scrollNode.scrollTo(this.scrollNode.offsetWidth + this.scrollNode.scrollWidth, 0)
                        })
                    }, 10)
                }
                return false
            }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
