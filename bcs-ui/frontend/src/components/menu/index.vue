<template>
    <ul class="bk-menu">
        <li class="bk-menu-item" v-for="(item, itemIndex) in list" :key="itemIndex">
            <div class="line" v-if="item.type === 'line'"></div>
            <div :class="['bk-menu-title-wrapper', item.disable, { selected: selected === item.id }]"
                @click="handleItemClick(item)" v-else>
                <i :class="['bcs-icon left-icon', item.icon]"></i>
                <div class="bk-menu-title">{{item.name}}</div>
                <i :class="['bcs-icon right-icon bcs-icon-angle-down', openedMenu.includes(item.id) ? 'selected' : 'bcs-icon-angle-down']"
                    v-if="item.children && item.children.length"></i>
            </div>
            <collapse-transition>
                <ul v-show="openedMenu.includes(item.id)">
                    <li class="bk-menu-child-item" v-for="(child, childIndex) in (item.children || [])" :key="childIndex">
                        <div :class="['bk-menu-child-title-wrapper', { selected: selected === child.id }]"
                            @click="handleChildClick(child, item)">
                            {{child.name}}
                        </div>
                    </li>
                </ul>
            </collapse-transition>
        </li>
    </ul>
</template>

<script lang="ts">
    import { defineComponent, ref, toRefs, watch, PropType } from '@vue/composition-api'
    import CollapseTransition from './collapse-transition'
    import { IMenuItem } from '@/store/menu'

    export default defineComponent({
        name: 'SideMenu',
        components: { CollapseTransition },
        props: {
            // 菜单列表
            list: {
                type: Array as PropType<IMenuItem[]>,
                default: () => ([])
            },
            // 选中菜单的ID
            selected: {
                type: String,
                default: ''
            }
        },
        setup (props, ctx) {
            const { emit } = ctx

            const openedMenu = ref<string[]>([]) // 展开的菜单项
            const { selected, list } = toRefs(props)
            watch(selected, () => {
                // 如果是子菜单选中时默认展开父级
                const parent = list.value.find((item: IMenuItem) => {
                    return !!item.children?.some(child => child.id === selected.value)
                })
                if (parent) {
                    const exit = openedMenu.value.some(id => id === parent.id)
                    !exit && openedMenu.value.push(parent.id)
                }
            }, { immediate: true })

            const handleItemClick = (item) => {
                if (item.children && item.children.length) {
                    const index = openedMenu.value.findIndex(id => id === item.id)
                    if (index > -1) {
                        openedMenu.value.splice(index, 1)
                    } else {
                        openedMenu.value.push(item.id)
                    }
                } else {
                    emit('change', item)
                }
            }
            const handleChildClick = (child, item) => {
                // 点击子菜单时折叠其他菜单项
                openedMenu.value = [item.id]
                emit('change', child)
            }

            return {
                openedMenu,
                handleItemClick,
                handleChildClick
            }
        }
    })
</script>

<style scoped>
    @import './index.css';
</style>
