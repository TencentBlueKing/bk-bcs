<template>
    <div :class="[{ 'bcs-validate': isError }, type]"
        @focusin="handleFocus"
        @focusout="handleBlur">
        <slot></slot>
        <span class="error-tip"
            v-if="isError"
            v-bk-tooltips="errorMsg">
            <i class="bk-icon icon-exclamation-circle-shape"></i>
        </span>
    </div>
</template>
<script lang="ts">
    import { computed, defineComponent, ref, toRefs, watch } from '@vue/composition-api'

    export interface IValidate {
      validator: Function | RegExp;
      message: string;
    }
    export default defineComponent({
        name: 'BCSValidate',
        props: {
            type: {
                type: String,
                default: 'input'
            },
            disabled: {
                type: Boolean,
                default: false
            },
            message: {
                type: String,
                default: ''
            },
            rules: {
                type: Array,
                default: () => []
            },
            trigger: {
                type: String,
                default: 'change'
            },
            value: {
                type: [String, Array, Object, Number],
                default: ''
            },
            meta: {
                type: [String, Array, Object, Number]
            }
        },
        emits: ['validate'],
        setup (props, ctx) {
            const { disabled, message, rules, value, meta } = toRefs(props)
            const focus = ref(false)
            function handleFocus () {
                focus.value = true
            }
            function handleBlur () {
                focus.value = false
                validate()
            }

            async function validate () {
                curErrMsg.value = ''
                if (!rules.value.length || !value.value) return true
                
                const allPromise: Array<Promise<any>> = [];
                (rules.value as IValidate[]).forEach(item => {
                    const promise = new Promise(async (resolve, reject) => {
                        let result = false
                        if (typeof item.validator === 'function') {
                            result = await item.validator(value.value, meta?.value)
                        } else {
                            result = new RegExp(item.validator).test(String(value.value))
                        }
                        if (result) {
                            resolve(item)
                        } else {
                            reject(new Error(item.message))
                        }
                    })
                    allPromise.push(promise)
                })

                return Promise.all(allPromise)
                    .then(() => {
                        curErrMsg.value = ''
                        ctx.emit('validate', true)
                    })
                    .catch((err) => {
                        curErrMsg.value = err.message
                        ctx.emit('validate', false)
                    })
            }
            const curErrMsg = ref('')
            watch(value, () => {
                validate()
            }, { deep: true, immediate: true })

            const errorMsg = computed(() => {
                return curErrMsg.value || message.value
            })
            const isError = computed(() => {
                return !focus.value && !disabled.value && errorMsg.value
            })
            return {
                errorMsg,
                focus,
                isError,
                handleFocus,
                handleBlur
            }
        }
    })
</script>
<style lang="postcss" scoped>
.bcs-validate {
  position: relative;
  .error-tip {
    font-size: 16px;
    position: absolute;
    right: 8px;
    top: 8px;
    line-height: 1;
    i {
      color: #ea3636 !important;
    }
  }

  &.input {
    >>> input {
      border-color: #ff5656 !important;
      color: #ff5656 !important;
    }
  }
}
</style>
