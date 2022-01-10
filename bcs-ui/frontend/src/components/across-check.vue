<template>
    <div class="check">
        <bcs-button
            ext-cls="check-btn-loading"
            size="small"
            :loading="loading"
            v-if="loading"
        ></bcs-button>
        <template v-else>
            <bcs-checkbox
                :checked="isChecked"
                :indeterminate="indeterminate"
                :class="{
                    'all-check': allChecked,
                    'indeterminate': localValue === CheckType.HalfAcrossChecked
                }"
                :disabled="disabled"
                :true-value="CheckType.Checked"
                :false-value="CheckType.Uncheck"
                @change="handleCheckChange">
            </bcs-checkbox>
            <bcs-popover
                ref="popover"
                theme="light dropdown"
                trigger="click"
                placement="bottom"
                :arrow="false"
                offset="10, 0"
                :on-show="() => isDropDownShow = true"
                :on-hide="() => isDropDownShow = false"
                :disabled="disabled">
                <i
                    :class="[
                        'check-icon bk-icon',
                        `icon-angle-${ isDropDownShow ? 'up' : 'down'}`,
                        { disabled: disabled }
                    ]">
                </i>
                <template #content>
                    <ul class="dropdown-list">
                        <li v-for="item in checkTypeList"
                            :key="item.id"
                            :class="{ active: localValue === item.id }"
                            @click="handleCheckAll(item)">
                            {{ item.name }}
                        </li>
                    </ul>
                </template>
            </bcs-popover>
        </template>
    </div>
</template>
<script lang="ts">
    import { computed, defineComponent, ref, toRefs, watch } from '@vue/composition-api'
    export enum CheckType {
        Uncheck, // 0 未选
        HalfChecked, // 1 当前页半选
        HalfAcrossChecked, // 2 跨页半选
        Checked, // 3 当前页全选
        AcrossChecked // 4 跨页全选
    }
    export default defineComponent({
        name: 'AcrossCheck',
        model: {
            prop: 'value',
            event: 'change'
        },
        props: {
            value: {
                type: Number,
                default: CheckType.Uncheck
            },
            disabled: {
                type: Boolean,
                default: false
            },
            loading: {
                type: Boolean,
                default: false
            }
        },
        setup (props, ctx) {
            const { $i18n } = ctx.root
            const isDropDownShow = ref(false)
            const { value } = toRefs(props)
            watch(value, () => {
                localValue.value = value.value
            })
            const localValue = ref(value.value)
            const allChecked = computed(() => {
                return [CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(localValue.value)
            })
            const indeterminate = computed(() => {
                return [CheckType.HalfChecked, CheckType.HalfAcrossChecked].includes(localValue.value)
            })
            const isChecked = computed(() => {
                return [CheckType.Checked, CheckType.AcrossChecked].includes(localValue.value)
            })
            const checkTypeList = ref([
                {
                    id: CheckType.Checked,
                    name: $i18n.t('本页全选')
                },
                {
                    id: CheckType.AcrossChecked,
                    name: $i18n.t('跨页全选')
                }
            ])
            const handleCheckChange = (value) => {
                localValue.value = value
                ctx.emit('change', value)
            }
            const popover = ref<any>(null)
            const handleCheckAll = (item) => {
                handleCheckChange(item.id)
                popover.value && popover.value.hideHandler()
            }
            return {
                allChecked,
                indeterminate,
                popover,
                localValue,
                isChecked,
                isDropDownShow,
                checkTypeList,
                CheckType,
                handleCheckChange,
                handleCheckAll
            }
        }
    })
</script>
<style lang="postcss" scoped>
.check {
  text-align: left;
  .all-check {
    >>> .bk-checkbox {
      background-color: #fff;
      &::after {
        border-color: #3a84ff;
      }
    }
  }
  .indeterminate {
    >>> .bk-checkbox {
      &::after {
        background: #3a84ff;
      }
    }
  }
  &-icon {
    position: relative;
    top: 3px;
    font-size: 20px;
    cursor: pointer;
    color: #63656e;
    &.disabled {
      color: #c4c6cc;
    }
  }
  .check-btn-loading {
    padding: 0;
    min-width: auto;
    border: 0;
    text-align: left;
    background: transparent;
    >>> .bk-button-loading {
      position: static;
      transform: translateX(0);
      .bounce4 {
        display: none;
      }
    }
  }
}
</style>
