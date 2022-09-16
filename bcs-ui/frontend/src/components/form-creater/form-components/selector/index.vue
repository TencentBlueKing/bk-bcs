<template>
    <div class="bk-form-item">
        <label class="bk-label" style="width:300px;">{{configData.label}}：</label>
        <div class="bk-form-content" style="margin-left:300px;">
            <bk-selector
                :list="configData.optionList"
                :selected.sync="configData.default">
            </bk-selector>
            <div class="bk-form-tip">
                <p v-if="configData.description" class="bk-tip-text">
                    <span class="bk-tip-variable">{{$t('值来源')}}：Values.{{configData.variable}}</span>
                    {{configData.description}}
                </p>
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
                        'type': 'enum',
                        'group': '',
                        'options': [],
                        'required': true,
                        'label': '',
                        'show_subquestion_if': ''
                    }
                }
            }
        },
        data () {
            return {}
        },
        methods: {
            selected (id, data) {
                // alert(id, JSON.stringify(data))
            }
        }
    }
</script>
