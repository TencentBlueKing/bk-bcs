<template>
    <textarea
        class="bk-form-textarea"
        v-model="localValue"
        :style="extStyle"
        :placeholder="placeholder"
        @keyup="handlerKeyup">
    </textarea>
</template>
<script>
    export default {
        props: {
            maxlength: {
                type: Number
            },
            extStyle: {
                type: Object
            },
            minlength: {
                type: Number
            },
            value: {
                type: [Number, Object]
            },
            placeholder: {
                type: String,
                default: window.i18n.t('请输入')
            }
        },
        data () {
            let value = this.value
            if (typeof this.value === 'object') {
                value = JSON.stringify(this.value)
            }
            return {
                localValue: value
            }
        },
        watch: {
            value (val) {
                let value = val
                if (typeof val === 'object') {
                    value = JSON.stringify(val)
                }
                this.localValue = value
                // 解决传入的数据为对象问题
                this.$emit('update:value', this.localValue)
            }
        },
        methods: {
            handlerKeyup () {
                this.$emit('update:value', this.localValue)
            }
        }
    }
</script>
<style scoped>
    @import './index.css';
</style>
