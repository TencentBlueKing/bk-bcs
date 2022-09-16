<template>
    <div class="bk-form-item">
        <label class="bk-label" style="width:300px;">{{configData.label}}：</label>
        <div class="bk-form-content" style="margin-left:300px;">
            <input type="password" :class="['bk-form-input', { 'is-danger': status.invalid }]" name="validation_name" :placeholder="$t('请输入')" v-model="configData.default" @input="checkValue">
            <div class="bk-form-tip">
                <template v-if="status.invalid">
                    <p class="bk-tip-text">{{status.errorMsg}}</p>
                </template>
                <template v-else>
                    <p v-if="configData.description" class="bk-tip-text">
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
                        'type': 'password',
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
                if (this.configData.required && !value) {
                    this.status = {
                        invalid: true,
                        errorMsg: this.$t('必填项，不能为空')
                    }
                } else {
                    this.status = {
                        invalid: false,
                        errorMsg: ''
                    }
                }
                this.status.label = this.configData.label
                return this.status
            }
        }
    }
</script>
