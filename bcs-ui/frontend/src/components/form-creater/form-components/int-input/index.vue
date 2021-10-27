<template>
    <div class="bk-form-item">
        <label class="bk-label" style="width:300px;">{{configData.label}}：</label>
        <div class="bk-form-content" style="margin-left:300px;">
            <input type="number" :class="['bk-form-input', { 'is-danger': status.invalid }]" name="validation_name" :placeholder="$t('请输入')" v-model="configData.default" @input="checkValue">
            <div class="bk-form-tip">
                <template v-if="status.invalid">
                    <p class="bk-tip-text">{{status.errorMsg}}</p>
                </template>
                <template v-else>
                    <p class="bk-tip-text" v-if="configData.description">
                        <span class="bk-tip-variable">{{$t('值来源')}}：Values.{{configData.variable}}</span>
                        {{configData.description}}
                    </p>
                </template>
            </div>
        </div>
    </div>
</template>

<script>
    export default {
        props: {
            configData: {
                type: Object,
                default () {
                    return {
                        'variable': '',
                        'default': '',
                        'description': '',
                        'type': 'int',
                        'min': 30000,
                        'max': 32767,
                        'label': ''
                    }
                }
            }
        },
        data () {
            return {
                status: {
                    invalid: false,
                    errorMsg: ''
                }
            }
        },
        mounted () {
        },
        methods: {
            checkValue () {
                const value = this.configData.default
                const intReg = /^-?\d+$/
                if (this.configData.required && !value) {
                    this.status = {
                        invalid: true,
                        errorMsg: this.$t('必填项，不能为空')
                    }
                } else {
                    if (value && !intReg.test(value)) {
                        this.status = {
                            invalid: true,
                            errorMsg: this.$t('请输入整数')
                        }
                    } else {
                        this.status = {
                            invalid: false,
                            errorMsg: ''
                        }
                    }
                }
                this.status.label = this.configData.label
                return this.status
            }
        }
    }
</script>
