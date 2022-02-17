<template>
    <div class="node-searcher" ref="searchWrapper" v-clickoutside="hide">
        <div class="searcher" @click="foucusSearcher">
            <ul class="search-params-wrapper" ref="searchParamsWrapper">
                <template v-if="searchParams && searchParams.length">
                    <li v-for="(sp, spIndex) in searchParams" :key="spIndex" ref="searchParamsLi">
                        <div class="selectable" @click.stop="searchParamsClickHandler($event, sp, spIndex)">
                            <div class="name">{{sp.text}}</div>
                            <template v-if="sp.id === 'ip'">
                                <div class="value-container" v-if="sp.value && !sp.isEditing">
                                    <div class="value">
                                        <span class="value-item" v-for="(valueArrItem, valueArrIndex) in sp.valueArr" :key="valueArrIndex">{{valueArrItem}}</span>
                                    </div>
                                    <div class="remove-search-params" @click.stop="removeSearchParams(sp, spIndex)"><i class="bcs-icon bcs-icon-close"></i></div>
                                </div>
                                <div class="value-container edit" v-if="sp.isEditing">
                                    <input type="text" class="input" v-model="sp.value" :ref="`editInput${spIndex}`" @blur="editInputBlurHandler($event, sp, spIndex)">
                                </div>
                            </template>
                            <template v-else-if="sp.id === 'labels'">
                                <div class="value-container" v-if="sp.key">
                                    <div class="value">{{sp.key}}:</div>
                                    <div class="value-value">
                                        <span class="value-item" v-for="(valueArrItem, valueArrIndex) in sp.valueArr" :key="valueArrIndex">{{valueArrItem}}</span>
                                    </div>
                                    <div class="remove-search-params" @click.stop="removeSearchParams(sp, spIndex)"><i class="bcs-icon bcs-icon-close"></i></div>
                                </div>
                            </template>
                            <template v-else>
                                <div class="value-container" v-if="sp.text">
                                    <div class="value">
                                        <span class="value-item">{{sp.value}}</span>
                                    </div>
                                    <div class="remove-search-params" @click.stop="removeSearchParams(sp, spIndex)"><i class="bcs-icon bcs-icon-close"></i></div>
                                </div>
                            </template>
                        </div>
                    </li>
                </template>

                <li v-if="curSearchParams && Object.keys(curSearchParams)">
                    <template v-if="curSearchParams.id === 'ip'">
                        <div class="selectable">
                            <div class="name">{{curSearchParams.text}}</div>
                        </div>
                    </template>
                    <template v-else>
                        <div class="selectable">
                            <div class="name">{{curSearchParams.text}}</div>
                            <div class="value-container" v-if="curSearchParams.key">
                                <div class="value">{{curSearchParams.key}}</div>
                            </div>
                        </div>
                    </template>
                </li>

                <li ref="searchInputParent">
                    <input type="text" class="input" ref="searchInput" v-model="curInputValue"
                        :placeholder="inputPlaceholder"
                        :style="{ maxWidth: `${maxInputWidth}px`, minWidth: `${minInputWidth}px` }"
                        :maxlength="isListeningInputKeyup ? Infinity : 0"
                        @blur="searchInputBlurHandler($event)"
                        @keyup="inputKeyup($event)"
                        @keypress="preventKeyboardEvt($event)"
                        @keydown="preventKeyboardEvt($event)"
                        @paste="handleInputPaste">
                </li>
            </ul>
        </div>

        <div class="node-searcher-dropdown-menu show-enter-tip" v-show="showEnterTip">
            <div class="node-searcher-dropdown-content" :class="showEnterTip ? 'is-show' : ''" :style="{ left: `${searcherDropdownLeft}px` }">
                <ul class="node-searcher-dropdown-list">
                    <li>
                        <i class="bcs-icon bcs-icon-search"></i>
                        <div>Press Enter to search</div>
                    </li>
                </ul>
            </div>
        </div>

        <div class="node-searcher-dropdown-menu label-list" v-show="showLabel">
            <div class="node-searcher-dropdown-content" :class="showLabel ? 'is-show' : ''" :style="{ left: `${searcherDropdownLeft}px` }">
                <ul class="node-searcher-dropdown-list">
                    <li v-for="(label, labelIndex) in labelList" :key="labelIndex">
                        <a href="javascript:void(0);" @click="selectLabel(label, labelIndex)">{{label.text}}</a>
                    </li>
                </ul>
            </div>
        </div>

        <div class="node-searcher-dropdown-menu tag-list" v-show="showStatus">
            <div class="node-searcher-dropdown-content is-show" :style="{ left: `${searcherDropdownLeft}px` }">
                <ul class="node-searcher-dropdown-list">
                    <li v-for="(s, sIndex) in statusList" :key="sIndex">
                        <a href="javascript:void(0);" @click="selectStatus(s)">{{s.text}}</a>
                    </li>
                </ul>
            </div>
        </div>

        <div class="node-searcher-dropdown-menu tag-list" v-show="showKey">
            <div class="node-searcher-dropdown-content" :class="showKey ? 'is-show' : ''" :style="{ left: `${searcherDropdownLeft}px` }">
                <ul class="node-searcher-dropdown-list" v-bkloading="{ isLoading: tagLoading, opacity: 1 }">
                    <template v-if="keyList && keyList.length">
                        <li v-for="(k, kIndex) in keyList" :key="kIndex">
                            <a href="javascript:void(0);" @click="selectKey(k)">{{k}}</a>
                        </li>
                    </template>
                    <template v-else>
                        <li>
                            <a href="javascript:void(0);">{{$t('没有数据')}}</a>
                        </li>
                    </template>
                </ul>
            </div>
        </div>

        <div class="node-searcher-dropdown-menu value-list" v-show="showValue">
            <div class="node-searcher-dropdown-content" :class="showValue ? 'is-show' : ''" :style="{ left: `${searcherDropdownLeft}px` }">
                <ul class="node-searcher-dropdown-list" v-bkloading="{ isLoading: tagLoading, opacity: 1 }">
                    <template v-if="valueList && valueList.length">
                        <li v-for="(v, vIndex) in valueList" :key="vIndex">
                            <a href="javascript:void(0);" @click="selectValue(v)">
                                {{v}}
                                <i class="bcs-icon bcs-icon-check-1" v-if="selectedValues[v]"></i>
                            </a>
                        </li>
                    </template>
                    <template v-else>
                        <li>
                            <a href="javascript:void(0);">{{$t('没有数据')}}</a>
                        </li>
                    </template>
                </ul>
                <div class="action" v-if="valueList && valueList.length && Object.keys(selectedValues).filter(v => selectedValues[v]).length">
                    <span class="btn" @click="confirmSelectValue">{{$t('确认')}}</span>
                    <span class="btn" @click="cancelSelectValue">{{$t('取消')}}</span>
                </div>
                <div class="action" v-else>
                    <span class="btn disabled">{{$t('确认')}}</span>
                    <span class="btn disabled">{{$t('取消')}}</span>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import clickoutside from '@/components/bk-searcher/clickoutside'
    // import { getActualLeft, getStringLen, insertAfter } from '@/common/util'
    import { getStringLen, catchErrorHandler } from '@/common/util'

    export default {
        name: 'node-searcher',
        directives: {
            clickoutside
        },
        props: {
            projectId: {
                type: String
            },
            clusterId: {
                type: String,
                default: 'all'
            },
            params: {
                type: Array,
                default: []
            },
            hadSearchData: {
                type: Boolean,
                default: false
            },
            searchLabelsData: {
                type: Object,
                default: () => ({})
            }
        },
        data () {
            return {
                curInputValue: '',
                labelList: [
                    { id: 'ip', text: this.$t('IP地址') },
                    { id: 'labels', text: this.$t('标签') },
                    { id: 'status_list', text: this.$t('状态') }
                ],
                // 输入框的最小宽度
                minInputWidth: 190,
                // 输入框的最大宽度
                maxInputWidth: 200,
                // 显示 label 的弹层
                showLabel: false,
                // 显示 key 的弹层
                showKey: false,
                // 显示 value 的弹层
                showValue: false,
                // 过滤项下拉框的左偏移
                searcherDropdownLeft: 0,
                // search-params-wrapper 里的 li 元素的 margin 值
                searchParamsItemMargin: 3,
                // 搜索参数
                searchParams: [],
                // 当前正在输入的那个搜索参数
                curSearchParams: null,
                // 是否监听 input keyup 的开关
                isListeningInputKeyup: false,
                // 已经存在于搜索框中搜索参数元素 li 的总宽度
                allSearchParamsWidth: 0,
                // key 的数组，页面显示的，文本框输入筛选的
                keyList: [],
                // key 的数组，ajax 数组返回的
                keyListTmp: [],
                // value 的数组，页面显示的，文本框输入筛选的
                valueList: [],
                // value 的数组，ajax 数组返回的
                valueListTmp: [],
                tagLoading: false,
                inputPlaceholder: '',
                selectedValues: {},
                curTmpLabelsValueContainerWidth: 0,
                maxLeftOffset: 514,
                showEnterTip: false,
                showStatus: false,
                statusList: [
                    { text: this.$t('初始化中'), value: ['INITIALIZATION'] },
                    { text: this.$t('正常'), value: ['RUNNING'] },
                    { text: this.$t('不正常'), value: ['NOTREADY'] },
                    { text: this.$t('不可调度'), value: ['REMOVABLE'] },
                    { text: this.$t('删除中'), value: ['DELETING'] },
                    { text: this.$t('上架失败'), value: ['ADD-FAILURE'] },
                    { text: this.$t('下架失败'), value: ['REMOVE-FAILURE'] },
                    { text: this.$t('未知状态'), value: ['UNKNOWN'] }
                ]
            }
        },
        computed: {
            curProject () {
                return this.$store.state.curProject
            }
        },
        watch: {
            searchParams (val) {
                this.$nextTick(() => {
                    this.allSearchParamsWidth = 0
                    const searchParamsLi = this.$refs.searchParamsLi
                    if (searchParamsLi) {
                        searchParamsLi.forEach(node => {
                            this.allSearchParamsWidth += node.offsetWidth + this.searchParamsItemMargin
                        })
                    }
                })
            },
            searcherDropdownLeft (val) {
                if (val > this.maxLeftOffset) {
                    this.searcherDropdownLeft = this.maxLeftOffset
                }
            },
            params (val) {
                if (this.params.length) {
                    this.searchParams.splice(0, this.searchParams.length, ...this.params)
                }
            }
        },
        mounted () {
            if (this.params.length) {
                this.searchParams.splice(0, this.searchParams.length, ...this.params)
            }
        },
        methods: {
            /**
             * 组件 click 事件
             */
            foucusSearcher () {
                if (this.isListeningInputKeyup) {
                    return
                }

                this.$nextTick(() => {
                    this.$refs.searchInput.focus()
                })

                setTimeout(() => {
                    const searchParamsWrapper = this.$refs.searchParamsWrapper
                    if (this.allSearchParamsWidth > searchParamsWrapper.offsetWidth) {
                        this.searcherDropdownLeft = this.allSearchParamsWidth - searchParamsWrapper.parentNode.scrollLeft
                    } else {
                        this.searcherDropdownLeft = this.allSearchParamsWidth
                    }
                }, 0)

                this.showLabel = true
            },

            /**
             * k8s node labels search
             */
            setSearchLabelKeys () {
                const keyList = Object.keys(this.searchLabelsData)
                this.keyList.splice(0, this.keyList.length, ...keyList)
                this.keyListTmp.splice(0, this.keyList.length, ...keyList)
                this.$nextTick(() => {
                    this.$refs.searchInput.focus()
                    this.isListeningInputKeyup = true
                })
                this.inputPlaceholder = this.$t('请输入要搜索的key')
            },

            /**
             * label 选择事件回调
             *
             * @param {Object} label 当前选择的 label
             */
            async selectLabel (label) {
                const curSearchParams = {
                    id: label.id,
                    text: label.text
                }

                // 一个字符大约是 8 px，横向 padding 10 px，左右 padding 一共是 20 px
                this.searcherDropdownLeft += getStringLen(label.text) * 8 + 10

                if (label.id === 'ip') {
                    curSearchParams.value = ''
                    this.curSearchParams = Object.assign({}, curSearchParams)
                    this.inputPlaceholder = this.$t('请输入要搜索的ip，多个ip以 | 隔开')

                    this.$nextTick(() => {
                        this.$refs.searchInput.focus()
                        this.showLabel = false
                        this.isListeningInputKeyup = true
                        this.showEnterTip = true
                    })
                } else if (label.id === 'labels') {
                    curSearchParams.key = ''
                    curSearchParams.value = ''
                    this.curSearchParams = Object.assign({}, curSearchParams)

                    this.showLabel = false
                    this.showKey = true
                    // 直接传入数据，不需要通过接口获取
                    if (this.hadSearchData) {
                        this.setSearchLabelKeys()
                        return
                    }
                    this.tagLoading = true
                    try {
                        const res = await this.$store.dispatch('cluster/getNodeKeyList', {
                            projectId: this.projectId,
                            clusterId: this.clusterId
                        })
                        const keyList = res.data || []
                        this.keyList.splice(0, this.keyList.length, ...keyList)
                        this.keyListTmp.splice(0, this.keyList.length, ...keyList)
                        this.$nextTick(() => {
                            this.$refs.searchInput.focus()
                            this.isListeningInputKeyup = true
                        })
                        this.inputPlaceholder = this.$t('请输入要搜索的key')
                    } catch (e) {
                        catchErrorHandler(e, this)
                    } finally {
                        this.tagLoading = false
                    }
                } else {
                    curSearchParams.key = ''
                    curSearchParams.value = ''
                    this.curSearchParams = Object.assign({}, curSearchParams)
                    this.showLabel = false
                    this.showStatus = true
                }
            },

            /**
             * 选择 status 回调事件
             *
             * @param {Object} v 选中的 status 对象
             */
            selectStatus (s) {
                this.$refs.searchInput.focus()
                this.curSearchParams.key = 'status'

                const searchParams = []
                searchParams.splice(0, 0, ...this.searchParams)
                searchParams.push({
                    id: this.curSearchParams.id,
                    key: this.curSearchParams.key,
                    text: this.curSearchParams.text,
                    value: s.text,
                    valueArr: s.value
                })
                this.searchParams.splice(0, this.searchParams.length, ...searchParams)

                this.curSearchParams = null
                this.showStatus = false
                this.adjustOffset()
                this.$emit('search')
            },

            /**
             * k8s node label search values
             */
            setSearchLabelValues (key) {
                this.showKey = false
                this.showValue = true
                this.curInputValue = ''
                this.inputPlaceholder = this.$t('请输入要搜索的value')
                const valueList = this.searchLabelsData[key] || []
                this.valueList.splice(0, this.valueList.length, ...valueList)
                this.valueListTmp.splice(0, this.valueList.length, ...valueList)
                this.$refs.searchInput.focus()
            },

            /**
             * 选择 key 事件
             *
             * @param {string} k 选中的 key
             */
            async selectKey (k) {
                this.curSearchParams.key = k
                if (this.hadSearchData) {
                    this.setSearchLabelValues(k)
                    return
                }
                try {
                    this.showKey = false
                    this.showValue = true
                    this.tagLoading = true
                    this.curInputValue = ''
                    this.inputPlaceholder = this.$t('请输入要搜索的value')
                    let valueList = []
                    const res = await this.$store.dispatch('cluster/getNodeValueListByKey', {
                        projectId: this.projectId,
                        clusterId: this.clusterId,
                        keyName: k
                    })
                    valueList = res.data || []
                    this.valueList.splice(0, this.valueList.length, ...valueList)
                    this.valueListTmp.splice(0, this.valueList.length, ...valueList)
                    this.$refs.searchInput.focus()
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.tagLoading = false
                }
            },

            /**
             * 选择 value 事件
             *
             * @param {string} v 选中的 value
             */
            selectValue (v) {
                const selectedValues = Object.assign({}, this.selectedValues)
                selectedValues[v] = !selectedValues[v]
                this.selectedValues = Object.assign({}, selectedValues)
                this.$refs.searchInput.focus()
            },

            /**
             * 取消选择 value
             */
            cancelSelectValue () {
                this.selectedValues = Object.assign({}, {})
                this.keyList.splice(0, this.keyList.length, ...this.keyListTmp)
                this.curSearchParams.key = ''
                this.inputPlaceholder = this.$t('请输入要搜索的key')
                this.showKey = true
                this.showValue = false
                this.$nextTick(() => {
                    this.searcherDropdownLeft -= this.curTmpLabelsValueContainerWidth
                    this.$refs.searchInput.focus()
                    this.isListeningInputKeyup = true
                    this.curTmpLabelsValueContainerWidth = 0
                    this.curInputValue = ''
                })
            },

            /**
             * 确认选择 value
             */
            confirmSelectValue () {
                const searchParams = []
                searchParams.splice(0, 0, ...this.searchParams)
                searchParams.push({
                    id: this.curSearchParams.id,
                    key: this.curSearchParams.key,
                    text: this.curSearchParams.text,
                    value: Object.keys(this.selectedValues).join('|'),
                    valueArr: Object.keys(this.selectedValues)
                })
                this.searchParams.splice(0, this.searchParams.length, ...searchParams)

                this.curSearchParams = null
                this.curInputValue = ''
                this.isListeningInputKeyup = false
                this.selectedValues = Object.assign({}, {})
                this.showValue = false
                this.inputPlaceholder = ''
                this.curTmpLabelsValueContainerWidth = 0

                this.adjustOffset()
                this.$emit('search')
            },

            /**
             * 搜索文本框失去焦点事件
             *
             * @param {Object} e 事件对象
             *
             * @return {string} returnDesc
             */
            searchInputBlurHandler (e) {
                // const value = e.target.value.trim()
                // if (!value && !this.showKey && !this.showValue) {
                //     this.curSearchParams = null
                //     this.curInputValue = ''
                //     this.isListeningInputKeyup = false
                // }
            },

            /**
             * searchParams 点击事件
             *
             * @param {Object} e 事件对象
             * @param {Object} sp 当前点击的参数对象
             * @param {Number} sp 当前点击的参数对象索引
             */
            searchParamsClickHandler (e, sp, spIndex) {
                sp.isEditing = true
                this.$set(this.searchParams, spIndex, sp)
                this.$nextTick(() => {
                    const editInput = this.$refs[`editInput${spIndex}`]
                    if (editInput && editInput[0]) {
                        editInput[0].focus()
                    }
                })
            },

            /**
             * searchParams 编辑文本框 blur 事件
             *
             * @param {Object} e 事件对象
             * @param {Object} sp 当前点击的参数对象
             * @param {Number} sp 当前点击的参数对象索引
             */
            editInputBlurHandler (e, sp, spIndex) {
                const value = e.target.value.trim()
                if (!value) {
                    this.removeSearchParams(sp, spIndex)
                    return
                }
                sp.isEditing = false
                sp.value = value.replace(/\|$/, '')
                sp.valueArr = value.split('|').filter(item => item)

                this.$set(this.searchParams, spIndex, sp)
                this.$emit('search')
            },

            /**
             * 删除 searchParams
             *
             * @param {Object} sp 当前点击的参数对象
             * @param {Number} sp 当前点击的参数对象索引
             */
            removeSearchParams (sp, spIndex) {
                const searchParams = []
                searchParams.splice(0, 0, ...this.searchParams)
                searchParams.splice(spIndex, 1)

                this.searchParams.splice(0, this.searchParams.length, ...searchParams)
                this.inputPlaceholder = ''
                this.showLabel = false
                this.showKey = false
                this.showValue = false
                this.showStatus = false
                this.$emit('search')
            },

            /**
             * 输入框 keyup 事件
             *
             * @param {Object} e 事件对象
             */
            inputKeyup (e) {
                if (!this.isListeningInputKeyup) {
                    return
                }

                const keyCode = e.keyCode

                if (this.showKey) {
                    this.searchKey()
                }

                if (this.showValue) {
                    this.searchValue()
                }

                const curInputValue = this.curInputValue.trim()
                if (!curInputValue) {
                    return
                }

                // tab || enter
                if ((keyCode === 9 || keyCode === 13) && !this.showKey && !this.showValue) {
                    const searchParams = []
                    searchParams.splice(0, 0, ...this.searchParams)
                    searchParams.push({
                        id: this.curSearchParams.id,
                        text: this.curSearchParams.text,
                        value: curInputValue.replace(/\|$/, ''),
                        valueArr: curInputValue.split('|').filter(item => item)
                    })
                    this.searchParams.splice(0, this.searchParams.length, ...searchParams)

                    this.curSearchParams = null
                    this.curInputValue = ''
                    this.isListeningInputKeyup = false

                    this.adjustOffset()
                    this.showEnterTip = false
                    this.$emit('search')
                }
            },

            /**
             * 搜索 key
             */
            searchKey () {
                const inputValue = this.curInputValue.trim()
                if (!inputValue) {
                    this.keyList.splice(0, this.keyList.length, ...this.keyListTmp)
                    return
                }
                const keyList = this.keyListTmp.filter(item => item.indexOf(inputValue) > -1)
                this.keyList.splice(0, this.keyList.length, ...keyList)
            },

            /**
             * 搜索 value
             */
            searchValue () {
                const inputValue = this.curInputValue.trim()
                if (!inputValue) {
                    this.valueList.splice(0, this.valueList.length, ...this.valueListTmp)
                    return
                }
                const valueList = this.valueListTmp.filter(item => item.indexOf(inputValue) > -1)
                this.valueList.splice(0, this.valueList.length, ...valueList)
            },

            /**
             * 阻止 input 框一些按键的默认事件
             *
             * @param {Object} e 事件对象
             */
            preventKeyboardEvt (e) {
                switch (e.keyCode) {
                    // down
                    case 40:
                        e.stopPropagation()
                        e.preventDefault()
                        break
                    // up
                    case 38:
                        e.stopPropagation()
                        e.preventDefault()
                        break
                    // left
                    case 37:
                        e.stopPropagation()
                        e.preventDefault()
                        break
                    // right
                    case 39:
                        e.stopPropagation()
                        e.preventDefault()
                        break
                    // tab
                    case 9:
                        e.stopPropagation()
                        e.preventDefault()
                        break
                    default:
                }
            },

            /**
             * 处理 input 框 paste 事件
             * 把数据中的换行符\n 转换成 |
             */
            handleInputPaste (e) {
                console.log(1, e)
                const value = e.clipboardData.getData('text')
                console.log('handleInputPaste', value)
                if (value && this.curSearchParams && this.curSearchParams.id === 'ip') {
                    this.curInputValue = value.replace(/\r/g, '').split('\n').join('|')
                }
                e.target.blur()
                setTimeout(() => {
                    e.target.focus()
                }, 10)
            },

            /**
             * 修正弹框的左偏移
             */
            adjustOffset () {
                setTimeout(() => {
                    const searchParamsWrapper = this.$refs.searchParamsWrapper
                    let leftOffset = 0
                    if (this.allSearchParamsWidth > searchParamsWrapper.offsetWidth) {
                        leftOffset = this.allSearchParamsWidth - searchParamsWrapper.parentNode.scrollLeft
                    } else {
                        leftOffset = this.allSearchParamsWidth
                    }
                    if (leftOffset + this.minInputWidth / 2 >= searchParamsWrapper.offsetWidth) {
                        searchParamsWrapper.parentNode.scrollLeft
                            += Math.max(2 * (leftOffset - searchParamsWrapper.offsetWidth), this.minInputWidth)
                    }
                    setTimeout(() => {
                        this.foucusSearcher()
                    }, 0)
                }, 0)
            },

            /**
             * 清除 searcher 搜索条件
             */
            clear () {
                this.searchParams.splice(0, this.searchParams.length, ...[])
                this.hide()
                this.$emit('search')
            },

            /**
             * 隐藏
             */
            hide () {
                if (this.tagLoading) {
                    return
                }
                const inputValue = this.curInputValue.trim()
                if (!inputValue) {
                    this.curSearchParams = null
                    this.isListeningInputKeyup = false
                    this.showLabel = false
                    this.showKey = false
                    this.showValue = false
                    this.showStatus = false
                    this.keyList.splice(0, this.keyList.length, ...[])
                    this.keyListTmp.splice(0, this.keyListTmp.length, ...[])
                    this.valueList.splice(0, this.valueList.length, ...[])
                    this.valueListTmp.splice(0, this.valueListTmp.length, ...[])
                    this.inputPlaceholder = ''
                    this.curTmpLabelsValueContainerWidth = 0
                    this.selectedValues = Object.assign({}, {})

                    this.showEnterTip = false
                }
            }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
