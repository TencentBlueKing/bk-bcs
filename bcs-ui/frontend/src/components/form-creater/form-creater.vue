<template>
    <div id="app" class="bk-form-creater">
        <!-- bk-form-header start -->
        <div class="bk-form-header" v-if="title">
            <h3 class="bk-title">{{title}}</h3>
            <div class="actions">
                <slot name="actions"></slot>
            </div>
        </div>
        <!-- bk-form-header end -->
        <bk-collapse v-model="collapseName" class="biz-var-collapse" accordion>
            <bcs-collapse-item :name="groupName" v-for="(groupName, index) of Object.keys(groups)" :key="index">
                {{groupName || ''}}
                <div slot="content" class="p10">
                    <div class="bk-form" style="width: 700px;">
                        <div class="group" v-for="(question, questionIndex) of groups[groupName]" :key="questionIndex">
                            <component
                                v-bind:is="question.type"
                                :config-data="question"
                                :group-name="groupName"
                                :ref="question.label">
                            </component>
                            <template v-for="(subQuestion, subquestionIndex) of question.subquestions">
                                <template v-if="(question.default === question.show_subquestion_if) || (question.default === String(question.show_subquestion_if))">
                                    <component
                                        v-bind:is="subQuestion.type"
                                        :group-name="groupName"
                                        :config-data="subQuestion"
                                        :ref="subQuestion.label"
                                        :key="subquestionIndex">
                                    </component>
                                </template>
                            </template>
                        </div>
                    </div>
                </div>
            </bcs-collapse-item>
        </bk-collapse>
    </div>
</template>

<script>
    import Vue from 'vue'
    import TextInput from './form-components/text-input/index.vue'
    import IntInput from './form-components/int-input/index.vue'
    import PasswordInput from './form-components/password-input/index.vue'
    import Radio from './form-components/radio/index.vue'
    import Selector from './form-components/selector/index.vue'

    // 注册组件
    Vue.component('string', TextInput)
    Vue.component('int', IntInput)
    Vue.component('password', PasswordInput)
    Vue.component('boolean', Radio)
    Vue.component('enum', Selector)

    export default {
        name: 'bk-form-creater',
        props: {
            title: {
                type: String,
                default: ''
            },
            collapseName: {
                type: Array,
                default () {
                    return []
                }
            },
            width: {
                type: Number,
                default: 700
            },
            labelWidth: {
                type: Number,
                default: 300
            },
            formData: {
                type: Object,
                default () {
                    return {}
                },
                validator (value) {
                    return value.questions
                }
            }
        },
        data () {
            return {
            }
        },
        computed: {
            groups () {
                const questions = this.formData.questions
                const groups = {}
                if (!questions) {
                    return []
                }
                questions.forEach(question => {
                    question.defaultValue = question.default
                    const groupName = question.group

                    if (question.options) {
                        question.optionList = []
                        question.options.forEach(item => {
                            question.optionList.push({
                                id: item,
                                name: item
                            })
                        })
                    }

                    if (!groups[groupName]) {
                        groups[groupName] = [question]
                    } else {
                        groups[groupName].push(question)
                    }

                    if (question.subquestions) {
                        question.subquestions.forEach(subQuestion => {
                            subQuestion.defaultValue = subQuestion.default
                            if (subQuestion.options) {
                                subQuestion.optionList = []
                                subQuestion.options.forEach(item => {
                                    subQuestion.optionList.push({
                                        id: item,
                                        name: item
                                    })
                                })
                            }
                        })
                    }
                })
                return groups
            }
        },
        mounted () {
            this.$emit('ready')
        },
        methods: {
            // 获取配置信息
            getFormConfigData () {
                return this.questions
            },
            checkValid () {
                const components = this.$refs
                for (const key in components) {
                    if (components[key] && components[key][0] && components[key][0].checkValue) {
                        if (components[key][0].checkValue) {
                            const component = components[key][0]
                            const status = components[key][0].checkValue()

                            if (status.invalid) {
                                const groupName = component.$attrs['group-name']
                                this.$bkMessage({
                                    theme: 'error',
                                    message: `${status.label}：${status.errorMsg}`
                                })
                                this.collapseName = [groupName]
                                return false
                            }
                        }
                    }
                }
                return true
            },
            // 获取表单数据
            getFormData () {
                const data = []
                for (const key in this.groups) {
                    const questions = this.groups[key]

                    questions.forEach(question => {
                        data.push({
                            name: question.variable,
                            type: question.type,
                            value: question.default
                        })
                        if (question.subquestions && (question.default === String(question.show_subquestion_if) || question.default === question.show_subquestion_if)) {
                            question.subquestions.forEach(subQuestion => {
                                data.push({
                                    name: subQuestion.variable,
                                    type: subQuestion.type,
                                    value: subQuestion.default
                                })
                            })
                        }
                    })
                }

                const components = this.$refs
                for (const key in components) {
                    if (components[key] && components[key][0] && components[key][0].checkValue) {
                        if (components[key][0].checkValue) {
                            const status = components[key][0].checkValue()
                            if (status.invalid) {
                                this.$bkMessage({
                                    theme: 'error',
                                    message: `${status.label}：${status.errorMsg}`
                                })
                                data.isError = true
                            }
                        }
                    }
                }
                return data
            }
        }
    }
</script>
<style>
    .bk-form-creater {
        .bk-form-header {
            overflow: hidden;
            padding: 10px 0;
            line-height: 36px;
        }
        .bk-form-item {
            margin-bottom: 15px;
        }
        .bk-title {
            float: left;
            margin: 0;
            color: #333;
        }
        .actions {
            float: right;
        }
        .group {
            margin-bottom: 10px;
        }
        .group-title {
            background: #fafafa;
            border-bottom: 1px solid #eee;
            padding: 10px;
            margin: 0 0 20px 0;
            font-size: 16px;
            text-align: left;
            padding-left: 300px;
            color: #333;
        }
        .bk-sideslider-closer {
            background-color: #3a84ff;
        }
        .question-json {
            height: 100%;
        }

        .bk-tip-variable {
            line-height: 20px;
            color: #737987;
            margin-right: 5px;
            font-size: 14px;
            display: block;
        }
        .bk-tip-text {
            color: #979ba5;
            line-height: 20px;
            white-space: normal;
            font-size: 13px;
        }
    }
</style>
