import { defineComponent, PropType, reactive, toRefs, computed, watch } from '@vue/composition-api'
import './check-mode.css'

// 表格跨页全选
export default defineComponent({
    name: 'CheckMode',
    model: {
        prop: 'value',
        event: 'change'
    },
    props: {
        disabled: {
            type: Boolean,
            default: false
        },
        value: {
            // 0: 未选择 1 半选 2 全选
            type: Number as PropType<0 | 1 | 2>,
            default: false
        },
        mode: {
            // 全选类型（跨页或者本页）
            type: String as PropType<'all'|'page'|''>,
            default: 'page'
        },
        onlyPageDisabled: {
            type: Boolean,
            default: true
        }
    },
    setup (props, ctx) {
        const state = reactive({
            checked: props.value === 2,
            indeterminate: props.value === 1
        })

        const menuList = computed(() => {
            return [
                {
                    id: 'page',
                    name: ctx.root.$i18n.t('全选当页'),
                    disabled: props.disabled
                },
                {
                    id: 'all',
                    name: ctx.root.$i18n.t('全选所有'),
                    disabled: props.onlyPageDisabled ? false : props.disabled
                }
            ]
        })

        const { value } = toRefs(props)
        watch(value, (v) => {
            state.checked = v === 2
            state.indeterminate = v === 1
        })

        const handleValueChange = (v: boolean) => {
            state.checked = v
            ctx.emit('change', v ? 2 : 0)
        }

        const handleModeChange = (item: any) => {
            if (item.disabled) return

            if (props.mode === item.id) {
                state.checked = !state.checked
            } else {
                state.checked = true
                ctx.emit('mode-change', item.id, props.mode)
            }
            handleValueChange(state.checked)
        }

        return {
            ...toRefs(state),
            menuList,
            handleValueChange,
            handleModeChange
        }
    },
    render () {
        return (
            <div class="check-mode">
                <bcs-checkbox
                    class={this.mode === 'all' ? 'all-checked' : ''}
                    indeterminate={this.indeterminate}
                    value={this.checked}
                    disabled={this.disabled}
                    onChange={this.handleValueChange}>
                </bcs-checkbox>
                <bcs-popover class="ml5"
                    theme="create-node-selector light"
                    arrow={false}
                    offset="15"
                    distance="0"
                    trigger="click"
                    scopedSlots={
                        {
                            default: () => (<span class="check-mode-angle ml5"><i class="bcs-icon bcs-icon-angle-down"></i></span>),
                            content: () => (
                                <ul class="menu-list">
                                    {
                                        this.menuList.map(item => (
                                            <li class={
                                                [
                                                    'menu-list-item',
                                                    this.mode === item.id ? 'active' : '',
                                                    item.disabled ? 'disabled' : ''
                                                ]}
                                            onClick={() => this.handleModeChange(item)}>{item.name}</li>
                                        ))
                                    }
                                </ul>
                            )
                        }
                    }>
                </bcs-popover>
            </div>
        )
    }
})
