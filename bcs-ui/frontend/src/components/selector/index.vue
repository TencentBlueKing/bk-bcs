<template>
    <div class="bk-selector"
        :class="[extCls, { 'open': open }]"
        @click="openFn($event)"
        :key="componentKey"
        v-clickoutside="close">
        <div class="bk-selector-wrapper">
            <input class="bk-selector-input" readonly="readonly"
                :class="{ placeholder: selectedText === placeholder, active: open }"
                :value="selectedText"
                :placeholder="placeholder"
                :disabled="disabled"
                @mouseover="showClearFn"
                @mouseleave="showClear = false">
            <i class="bcs-icon bcs-icon-angle-down bk-selector-icon" v-show="!isLoading && !showClear"></i>
            <i class="bcs-icon bcs-icon-close bk-selector-icon clear-icon"
                v-show="!isLoading && showClear"
                @mouseover="showClearFn"
                @mouseleave="showClear = false"
                @click="clearSelected($event)">
            </i>
            <div class="bk-spin-loading bk-spin-loading-mini bk-spin-loading-primary selector-loading-icon" v-show="isLoading">
                <div class="rotate rotate1"></div>
                <div class="rotate rotate2"></div>
                <div class="rotate rotate3"></div>
                <div class="rotate rotate4"></div>
                <div class="rotate rotate5"></div>
                <div class="rotate rotate6"></div>
                <div class="rotate rotate7"></div>
                <div class="rotate rotate8"></div>
            </div>
        </div>

        <transition :name="listSlideName">
            <div class="bk-selector-list" v-show="!isLoading && open" :style="panelStyle">
                <!-- 搜索栏 -->
                <div class="bk-selector-search-item"
                    @click="$event.stopPropagation()"
                    v-if="searchable">
                    <i class="bcs-icon bcs-icon-search"></i>
                    <input type="text" v-model="condition" @input="inputFn" ref="searchNode" :placeholder="searchPlaceholder">
                </div>
                <ul>
                    <template v-if="localList.length !== 0">
                        <li :class="['bk-selector-list-item', item.children && item.children.length ? 'bk-selector-group-list-item' : '']"
                            v-for="(item, parentIndex) in localList" :key="parentIndex">
                            <!-- 有分组 start -->
                            <template v-if="item.children && item.children.length">
                                <div class="bk-selector-group-name">{{item[displayKey]}}</div>
                                <ul class="bk-selector-group-list">
                                    <li v-for="(child, index) in item.children" :key="index" class="bk-selector-list-item">
                                        <div class="bk-selector-node bk-selector-sub-node"
                                            :class="{ 'bk-selector-selected': !multiSelect && child[settingKey] === selected,'is-disabled': child.isDisabled }">
                                            <div class="text" @click.stop="selectItem(child, $event)" :title="child[displayKey]">
                                                <label class="bk-form-checkbox bk-checkbox-small mr0 bk-selector-multi-label" v-if="multiSelect" style="line-height: 42px;">
                                                    <input type="checkbox"
                                                        :name="'multiSelect' + +new Date()"
                                                        :value="child[settingKey]"
                                                        v-model="localSelected">
                                                    <span class="select-text">{{ child[displayKey]}}</span>
                                                </label>
                                                <template v-else>
                                                    <span class="select-text">{{ child[displayKey]}}</span>
                                                </template>
                                            </div>
                                            <div class="bk-selector-tools" v-if="tools !== false">
                                                <i class="bcs-icon bcs-icon-edit2 bk-selector-list-icon"
                                                    v-if="tools.edit !== false"
                                                    @click.stop="editFn(index)"></i>
                                                <i class="bcs-icon bcs-icon-close bk-selector-list-icon"
                                                    v-if="tools.del !== false"
                                                    @click.stop="delFn(index)"></i>
                                            </div>
                                        </div>
                                    </li>
                                </ul>
                            </template>
                            <!-- 有分组 end -->

                            <!-- 没分组 start -->
                            <template v-else>
                                <div class="bk-selector-node" :class="{ 'bk-selector-selected': !multiSelect && item[settingKey] === selected, 'is-disabled': item.isDisabled }">
                                    <div class="text" @click.stop="selectItem(item, $event)" :title="item[displayKey]">
                                        <label class="bk-form-checkbox bk-checkbox-small mr0 bk-selector-multi-label" v-if="multiSelect" style="line-height: 42px;">
                                            <input type="checkbox"
                                                :name="'multiSelect' + +new Date()"
                                                :value="item[settingKey]"
                                                v-model="localSelected">
                                            <span class="select-text">{{ item[displayKey] }}</span>
                                        </label>
                                        <template v-else>
                                            <span class="select-text">{{ item[displayKey] }}</span>
                                        </template>
                                    </div>
                                    <div class="bk-selector-tools" v-if="tools !== false">
                                        <i class="bcs-icon bcs-icon-edit2 bk-selector-list-icon"
                                            v-if="tools.edit !== false"
                                            @click.stop="editFn(parentIndex)"></i>
                                        <i class="bcs-icon bcs-icon-close bk-selector-list-icon"
                                            v-if="tools.del !== false"
                                            @click.stop="delFn(parentIndex)"></i>
                                    </div>
                                </div>
                            </template>
                            <!-- 没分组 end -->
                        </li>
                    </template>
                    <li class="bk-selector-list-item" v-if="!isLoading && localList.length === 0">
                        <div class="text no-search-result">
                            {{ list.length ? (searchEmptyText || $t('无匹配数据')) : (emptyText || $t('暂无数据'))}}
                        </div>
                    </li>
                </ul>
                <!-- 新增项 start -->
                <slot></slot>
                <slot name="newItem"></slot>
                <template v-if="fieldType === 'cluster'">
                    <div class="bk-selector-create-item" @click.stop.prevent="goClusterList">
                        <i class="bcs-icon bcs-icon-apps"></i>
                        <i class="text">{{$t('集群列表')}}</i>
                    </div>
                </template>
                <template v-if="fieldType === 'namespace'">
                    <div class="bk-selector-create-item" @click.stop.prevent="goNamespaceList">
                        <i class="bcs-icon bcs-icon-apps"></i>
                        <i class="text">{{$t('命名空间列表')}}</i>
                    </div>
                </template>
                <template v-if="fieldType === 'metric'">
                    <div class="bk-selector-create-item" @click.stop.prevent="goMetricList">
                        <i class="bcs-icon bcs-icon-apps"></i>
                        <i class="text">{{$t('新建Metric')}}</i>
                    </div>
                </template>
            </div>
        </transition>
    </div>
</template>

<script>
    /**
     *  bk-dropdown
     *  @module components/dropdown
     *  @desc 下拉选框组件，模拟原生select
     *  @param extCls {String} - 自定义的样式
     *  @param hasCreateItem {Boolean} - 下拉菜单中是否有新增项，默认为true
     *  @param createText {String} - 下拉菜单中新增项的文字
     *  @param tools {Object, Boolean} - 待选项右侧的工具按钮，有两个可配置的key：edit和del，默认为两者都不显示。
     *  @param list {Array} - 必选，下拉菜单所需的数据列表
     *  @param filterList {Array} - 过滤列表
     *  @param selected {Number} - 必选，选中的项的index值，支持.sync修饰符
     *  @param placeholder {String, Boolean} - 是否显示占位行，默认为显示，且文字为“请选择”
     *  @param displayKey {String} - 循环list时，显示字段的key值，默认为name
     *  @param disabled {Boolean} - 是否禁用组件，默认为false
     *  @param multiSelect {Boolean} - 是否支持多选，默认为false
     *  @param searchable {Boolean} - 是否支持筛选，默认为false
     *  @param searchKey {Boolean} - 筛选时，搜索的key值，默认为'name'
     *  @param allowClear {Boolean} - 是否可以清除单选时选中的项
     *  @param settingKey {String} - 根据配置这个字段，自定义在单选时，选中某项之后的回调函数的第一个返回值的内容
     *  @example
        <bk-dropdown
            :list="list"
            :tools="tools"
            :selected.sync="selected"
            :placeholder="placeholder"
            :displayKey="displayKey"
            :has-create-item="hasCreateItem"
            :create-text="createText"
            :ext-cls="extCls"></bk-dropdown>
    */

    import clickoutside from '@/directives/clickoutside'
    import { getActualTop } from '@/common/util'

    export default {
        name: 'bk-selector',
        directives: {
            clickoutside
        },
        props: {
            fieldType: {
                type: String,
                default: ''
            },
            extCls: {
                type: String
            },
            isLoading: {
                type: Boolean,
                default: false
            },
            hasCreateItem: {
                type: Boolean,
                default: false
            },
            createText: {
                type: String,
                default: window.i18n.t('新增数据源')
            },
            tools: {
                type: [Object, Boolean],
                default: false
            },
            list: {
                type: Array,
                required: true
            },
            filterList: {
                type: Array,
                default () {
                    return []
                }
            },
            selected: {
                type: [Number, Array, String, Boolean],
                required: true
            },
            placeholder: {
                type: [String, Boolean],
                default: window.i18n.t('请选择')
            },
            // 是否联动
            isLink: {
                type: [String, Boolean],
                default: false
            },
            displayKey: {
                type: String,
                default: 'name'
            },
            disabled: {
                type: [String, Boolean, Number],
                default: false
            },
            multiSelect: {
                type: Boolean,
                default: false
            },
            searchable: {
                type: Boolean,
                default: false
            },
            searchKey: {
                type: String,
                default: 'name'
            },
            allowClear: {
                type: Boolean,
                default: false
            },
            settingKey: {
                type: String,
                default: 'id'
            },
            initPreventTrigger: {
                type: Boolean,
                default: false
            },
            emptyText: {
                type: String,
                default: ''
            },
            searchEmptyText: {
                type: String,
                default: ''
            },
            searchPlaceholder: {
                type: String,
                default: ''
            }
        },
        data () {
            return {
                open: false,
                componentKey: 0,
                selectedList: this.calcSelected(this.selected),
                condition: '',
                // localList: this.list,
                localSelected: this.selected,
                // emptyText: this.list.length ? '无匹配数据' : '暂无数据',
                showClear: false,
                panelStyle: {},
                listSlideName: 'toggle-slide'
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            localList () {
                const list = JSON.parse(JSON.stringify(this.list))
                if (!this.multiSelect) {
                    list.forEach(item => {
                        if (this.filterList.includes(item[this.settingKey])) {
                            item.isDisabled = true
                        } else {
                            item.isDisabled = false
                        }
                        if (item.children) {
                            item.children.forEach(item => {
                                if (this.filterList.includes(item[this.settingKey])) {
                                    item.isDisabled = true
                                } else {
                                    item.isDisabled = false
                                }
                                // this.componentKey++
                            })
                        }
                    })
                }

                if (this.searchable && this.condition) {
                    const arr = []
                    const key = this.searchKey

                    for (const item of list) {
                        if (item.children) {
                            const results = []
                            for (const child of item.children) {
                                if (child[key].toLowerCase().includes(this.condition.toLowerCase())) {
                                    results.push(child)
                                }
                            }
                            if (results.length) {
                                const cloneItem = Object.assign({}, item)
                                cloneItem.children = results
                                arr.push(cloneItem)
                            }
                        } else {
                            if (item[key].toLowerCase().includes(this.condition.toLowerCase())) {
                                arr.push(item)
                            }
                        }
                    }

                    return arr
                }
                return list
            },
            currentItem () {
                return this.list[this.localSelected]
            },
            selectedText () {
                const textArr = []
                if (Array.isArray(this.selectedList) && this.selectedList.length) {
                    this.selectedList.forEach(item => {
                        textArr.push(item[this.displayKey])
                    })
                } else if (this.selectedList) {
                    this.selectedList[this.displayKey] && textArr.push(this.selectedList[this.displayKey])
                }
                return textArr.length ? textArr.join(',') : this.placeholder
            }
        },
        watch: {
            selected (newVal) {
                // 重新生成选择列表
                if (this.list.length) {
                    this.selectedList = this.calcSelected(this.selected, this.isLink)
                }

                this.localSelected = this.selected
            },
            list (newVal) {
                // 重新生成选择列表
                // this.localList = this.list
                if (this.selected) {
                    this.selectedList = this.calcSelected(this.selected, this.isLink)
                } else {
                    this.selectedList = []
                }
            },
            localSelected (val) {
                // 重新生成选择列表
                if (this.list.length) {
                    this.selectedList = this.calcSelected(this.localSelected, this.isLink)
                }
            },
            open (newVal) {
                const searchNode = this.$refs.searchNode
                if (searchNode) {
                    if (newVal) {
                        this.$nextTick(() => {
                            searchNode.focus()
                        })
                    }
                }
            }
        },
        mounted () {
            this.popup = this.$el
            if (this.isLink) {
                if (this.list.length && this.selected) {
                    this.calcSelected(this.selected, this.isLink)
                }
            }
        },

        methods: {
            goClusterList () {
                this.$router.push({
                    name: 'clusterMain',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
            },
            goNamespaceList () {
                this.$router.push({
                    name: 'namespace',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
            },
            goMetricList () {
                this.$router.push({
                    name: 'metricManage',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
            },
            getItem (key) {
                let data = null
                const list = this.list

                list.forEach((item) => {
                    if (!item.children) {
                        if (item[this.settingKey] === String(key) || item[this.settingKey] === key) {
                            data = item
                        }
                    } else {
                        const list = item.children
                        list.forEach((item) => {
                            if (item[this.settingKey] === key) {
                                data = item
                            }
                        })
                    }
                })
                return data
            },
            calcSelected (selected, isTrigger) {
                let data = null

                if (Array.isArray(selected)) {
                    data = []
                    for (const key of selected) {
                        const item = this.getItem(key)
                        if (item) {
                            data.push(item)
                        }
                    }
                    if (data.length && isTrigger && !this.initPreventTrigger) {
                        this.$emit('item-selected', selected, data, isTrigger)
                    }
                } else if (selected !== undefined) {
                    const item = this.getItem(selected)
                    if (item) {
                        data = item
                    }
                    if (data && isTrigger && !this.initPreventTrigger) {
                        this.$emit('item-selected', selected, data, isTrigger)
                    }
                }
                return data
            },
            close () {
                this.open = false
                this.$emit('visible-toggle', this.open)
            },
            initSelectorPosition (currentTarget) {
                if (currentTarget) {
                    const distanceTop = getActualTop(currentTarget)
                    const winHeight = document.body.clientHeight
                    let ySet = {}
                    let listHeight = this.list.length * 42
                    if (listHeight > 160) {
                        listHeight = 160
                    }
                    const scrollTop = document.documentElement.scrollTop || document.body.scrollTop

                    if ((distanceTop + listHeight + 42 - scrollTop) < winHeight) {
                        ySet = {
                            top: '34px',
                            bottom: 'auto'
                        }

                        this.listSlideName = 'toggle-slide'
                    } else {
                        ySet = {
                            top: 'auto',
                            bottom: '34px'
                        }

                        this.listSlideName = 'toggle-slide2'
                    }

                    this.panelStyle = { ...ySet }
                }
            },
            openFn (event) {
                if (this.disabled) {
                    return
                }
                // 如果是loadin，禁止点击事件冒泡
                if (event && this.isLoading) {
                    event.stopPropagation()
                }

                if (!event) {
                    event = {}
                }

                if (!this.disabled) {
                    if (!this.open) {
                        this.initSelectorPosition(event.currentTarget)
                    }
                    this.open = !this.open
                    this.$emit('visible-toggle', this.open)
                }
            },
            /**
             *  计算返回渲染的数组
             */
            calcList () {
                if (this.searchable) {
                    const arr = []
                    const key = this.searchKey

                    for (const item of this.list) {
                        if (item.children) {
                            const results = []
                            for (const child of item.children) {
                                if (child[key].toLowerCase().includes(this.condition.toLowerCase())) {
                                    results.push(child)
                                }
                            }
                            if (results.length) {
                                const cloneItem = Object.assign({}, item)
                                cloneItem.children = results
                                arr.push(cloneItem)
                            }
                        } else {
                            if (item[key].toLowerCase().includes(this.condition.toLowerCase())) {
                                arr.push(item)
                            }
                        }
                    }

                    this.localList = arr
                } else {
                    this.localList = this.list
                }
            },
            /**
             *  是否显示清除当前选择的icon
             */
            showClearFn () {
                if (this.allowClear && !this.multiSelect && this.localSelected !== -1 && this.localSelected !== '') {
                    this.showClear = true
                }
            },
            /**
             *  清除选择
             */
            clearSelected (e) {
                this.$emit('clear', this.localSelected)
                this.localSelected = -1
                this.showClear = false
                e.stopPropagation()
                this.$emit('update:selected', '')
            },
            /**
             *  选中列表中的项
             */
            selectItem (data, event) {
                if (data.isDisabled) return
                setTimeout(() => {
                    this.toggleSelect(data, event)
                }, 10)
            },
            toggleSelect (data, event) {
                // label嵌input，触发两次click
                let $selected
                let $selectedList
                const settingKey = this.settingKey
                const isMultiSelect = this.multiSelect
                const index = (data && data[settingKey] !== undefined) ? data[settingKey] : undefined

                if (isMultiSelect && event.target.tagName.toLowerCase() === 'label') {
                    return
                }
                if (index !== undefined) {
                    if (!isMultiSelect) {
                        $selected = index
                    } else {
                        $selected = this.localSelected
                    }

                    $selectedList = this.calcSelected($selected)
                    if (isMultiSelect) {
                        $selected = $selected.filter(item => {
                            for (const selectItem of $selectedList) {
                                if (selectItem[settingKey] === item) {
                                    return true
                                }
                            }
                            return false
                        })
                    }
                    this.$emit('update:selected', $selected)
                } else {
                    this.$emit('update:selected', -1)
                }

                // 单选时，返回的两个参数是选中项的id（或通过settingKey配置的值）和选中项的数据
                // 多选时，返回的是选中项的索引数组和选中项的数据数组

                this.$emit('item-selected', $selected, $selectedList)

                if (!isMultiSelect) {
                    this.openFn()
                }

                // 点击搜索出来后的列表，不应该把搜索条件清空，清空条件会导致 calcList 方法里计算 localList 的时候计算成所有的
                // this.condition = ''
            },
            editFn (index) {
                this.$emit('edit', index)
                this.openFn()
            },
            delFn (index) {
                this.$emit('del', index)
                this.openFn()
            },
            createFn (e) {
                this.$emit('create')
                this.openFn()
                e.stopPropagation()
            },
            inputFn () {
                this.$emit('typing', this.condition)
            }
        }
    }
</script>

<style scoped>
    @import url(./index.css);
    .bk-form-checkbox {
        width: 100%;

        input[type=checkbox] {
            margin-right: 10px;
            vertical-align: middle;
        }
    }
    .select-text {
        display: inline-block;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        vertical-align: middle;
        width: calc(100% - 15px);
        font-size: 12px;
    }
    .bk-selector-input {
        padding-right: 20px;
    }
</style>
