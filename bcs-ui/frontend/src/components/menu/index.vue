<template>
    <div class="bk-menu">
        <template v-if="menuList.length">
            <ul>
                <li class="bk-menu-item" v-for="(item, itemIndex) in menuList" :key="itemIndex">
                    <template v-if="item.name === 'line'">
                        <div class="line"></div>
                    </template>
                    <template v-else-if="featureFlag[item.id]">
                        <div v-if="(!item.children || !item.children.length)" class="bk-menu-title-wrapper"
                            :class="[item.hide, item.disable, item.isSelected ? 'selected' : '']"
                            @click="(!item.disable && !item.hide) ? handleClick(item, itemIndex, $event) : () => {}">
                            <i class="bcs-icon left-icon" :class="[item.disable, item.icon, item.isSelected ? 'selected' : '']"></i>
                            <div class="bk-menu-title">{{item.name}}</div>
                            <i class="biz-badge" v-if="item.badge !== undefined">{{item.badge}}</i>
                        </div>

                        <div v-else class="bk-menu-title-wrapper"
                            :class="[item.hide, item.disable, item.isChildSelected ? 'child-selected' : '']"
                            @click="(!item.disable && !item.hide) ? openChildren(item, itemIndex, $event) : () => {}">
                            <i class="bcs-icon left-icon" :class="[item.disable, item.icon]"></i>
                            <div class="bk-menu-title">{{item.name}}</div>
                            <i class="bcs-icon right-icon bcs-icon-angle-down" :class="item.isOpen ? 'selected' : 'bcs-icon-angle-down'"></i>
                        </div>
                        <collapse-transition>
                            <ul v-show="item.isOpen">
                                <li class="bk-menu-child-item" v-for="(child, childIndex) in item.children" :key="childIndex">
                                    <div class="bk-menu-child-title-wrapper" :class="child.isSelected ? 'selected' : ''"
                                        @click="handleChildClick(item, itemIndex, child, childIndex, $event)">
                                        {{child.name}}
                                    </div>
                                </li>
                            </ul>
                        </collapse-transition>
                    </template>
                </li>
            </ul>
        </template>
        <template v-else>
            <div class="biz-no-data" style="margin-top: 100px;">
                <i class="bcs-icon bcs-icon-empty"></i>
                <p>{{$t("无数据")}}</p>
            </div>
        </template>
    </div>
</template>

<script>
    import CollapseTransition from './collapse-transition'

    export default {
        name: 'bk-menu',
        components: {
            CollapseTransition
        },
        props: {
            list: {
                type: Array,
                required: true
            },
            icon: {
                type: String,
                default: () => {
                    return 'icon-id'
                }
            },
            menuChangeHandler: {
                type: Function,
                default: null
            }
        },
        data () {
            return {
                menuList: this.list
            }
        },
        computed: {
            featureFlag () {
                return this.$store.getters.featureFlag || {}
            }
        },

        watch: {
            list () {
                this.menuList = this.list
            }
        },
        methods: {
            clearSelectCls (menuList = this.menuList) {
                menuList.forEach(item => {
                    item.isSelected = false
                    item.isOpen = false
                    if (item.children) {
                        item.isChildSelected = false
                        item.children.forEach(childItem => {
                            childItem.isSelected = false
                        })
                    }
                })
            },
            openChildren (item, itemIndex, e) {
                item.isOpen = !item.isOpen
                this.menuList.splice(itemIndex, 1, item)
            },
            handleClick (item, itemIndex) {
                // 当传入 menuChangeHandler 时，点击菜单不走 emit item-selected 的逻辑，而是需要判断 menuChangeHandler 的
                // 返回值来决定是否选中菜单
                if (this.menuChangeHandler && typeof this.menuChangeHandler === 'function') {
                    const data = {
                        isChild: true,
                        item,
                        itemIndex
                    }
                    if (item.isSaveData) {
                        sessionStorage['bcs-selected-menu-data'] = JSON.stringify(data)
                    } else {
                        sessionStorage.removeItem('bcs-selected-menu-data')
                    }
                    const ret = this.menuChangeHandler({
                        isChild: false,
                        item,
                        itemIndex
                    })
                    if (ret) {
                        this.clearSelectCls()
                        item.isSelected = !item.isSelected
                        this.menuList.splice(itemIndex, 1, item)
                    }
                    return
                }
                this.clearSelectCls()
                item.isSelected = !item.isSelected
                this.menuList.splice(itemIndex, 1, item)
                this.$emit('item-selected', {
                    isChild: false,
                    item,
                    itemIndex
                })
            },
            handleChildClick (item, itemIndex, child, childIndex, e) {
                // if (child.isSelected) {
                //     return
                // }

                // 当传入 menuChangeHandler 时，点击菜单不走 emit item-selected 的逻辑，而是需要判断 menuChangeHandler 的
                // 返回值来决定是否选中菜单
                if (this.menuChangeHandler && typeof this.menuChangeHandler === 'function') {
                    const data = {
                        isChild: true,
                        item,
                        itemIndex,
                        child,
                        childIndex
                    }
                    if (item.isSaveData) {
                        sessionStorage['bcs-selected-menu-data'] = JSON.stringify(data)
                    } else {
                        // sessionStorage['bcs-selected-menu-data'] = ''
                        sessionStorage.removeItem('bcs-selected-menu-data')
                    }
                    const ret = this.menuChangeHandler({
                        isChild: true,
                        item,
                        itemIndex,
                        child,
                        childIndex
                    })
                    if (ret) {
                        this.clearSelectCls()
                        child.isSelected = !child.isSelected
                        item.children.splice(childIndex, 1, child)
                        if (child.isSelected) {
                            item.isChildSelected = true
                            item.isOpen = true
                        }

                        this.menuList.splice(itemIndex, 1, item)
                    }
                    return
                }

                this.clearSelectCls()
                child.isSelected = !child.isSelected
                item.children.splice(childIndex, 1, child)
                if (child.isSelected) {
                    item.isChildSelected = true
                    item.isOpen = true
                }

                this.menuList.splice(itemIndex, 1, item)
                this.$emit('item-selected', {
                    isChild: true,
                    item,
                    itemIndex,
                    child,
                    childIndex
                })
            }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
