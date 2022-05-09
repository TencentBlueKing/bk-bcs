<template>
    <bcs-dialog class="selector-dialog"
        :mask-close="false"
        :close-icon="false"
        :esc-close="false"
        :value="modelValue"
        :width="dialogWidth"
        :auto-close="false"
        @value-change="handleValueChange"
        @confirm="handleConfirm">
        <Selector ref="selector"
            :key="selectorKey"
            :height="dialogHeight"
            :ip-list="ipList"
            v-if="modelValue"
            @change="handleIpSelectorChange"
        ></Selector>
    </bcs-dialog>
</template>
<script lang="ts">
    import { defineComponent, ref, toRefs, watch, onMounted } from '@vue/composition-api'
    import Selector from './ip-selector-bcs.vue'

    export default defineComponent({
        name: 'selector-dialog',
        components: {
            Selector
        },
        model: {
            prop: 'modelValue',
            event: 'change'
        },
        props: {
            modelValue: {
                type: Boolean,
                default: false
            },
            // 回显IP列表
            ipList: {
                type: Array,
                default: () => ([])
            }
        },

        setup (props, ctx) {
            const { emit } = ctx
            const { modelValue } = toRefs(props)
            const dialogWidth = ref(1200)
            const dialogHeight = ref(600)

            const selectorKey = ref(String(new Date().getTime()))
            watch(modelValue, () => {
                selectorKey.value = String(new Date().getTime())
            })
            const handleValueChange = (value: boolean) => {
                emit('change', value)
            }

            const handleIpSelectorChange = (data) => {
                emit('nodes-change', data)
            }

            const selector = ref<any>()
            const handleConfirm = () => {
                const data = selector.value?.handleGetData() || []
                if (!data.length) {
                    ctx.root.$bkMessage({
                        theme: 'error',
                        message: ctx.root.$i18n.t('请选择服务器')
                    })
                    return
                }
                emit('confirm', data)
            }

            onMounted(() => {
                dialogWidth.value = document.body.clientWidth < 1650 ? 1200 : document.body.clientWidth - 650
                dialogHeight.value = document.body.clientHeight < 1000 ? 600 : document.body.clientHeight - 320
            })

            return {
                selector,
                selectorKey,
                dialogWidth,
                dialogHeight,
                handleValueChange,
                handleConfirm,
                handleIpSelectorChange
            }
        }
    })
</script>
<style lang="postcss" scoped>
.selector-dialog {
    >>> .bk-dialog {
        top: 100px;
    }
    >>> .bk-dialog-tool {
        display: none;
    }
    >>> .bk-dialog-body {
        padding: 0;
    }
    >>> .bk-dialog-footer {
        border-top: none;
    }
}
</style>
