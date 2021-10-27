<template>
    <div class="bk-form-item">
        <label class="bk-label" style="width:300px;">{{configData.label}}：</label>
        <div class="bk-form-content" style="margin-left:300px;">
            <label class="bk-form-radio">
                <input type="radio" :class="{ 'is-checked': configData.default === true }" :name="configData.label" :checked="configData.default" :value="true" v-model="configData.default">
                <i class="bk-radio-text">true</i>
            </label>
            <label class="bk-form-radio">
                <input type="radio" :class="{ 'is-checked': configData.default === false }" :name="configData.label" :checked="configData.default" :value="false" v-model="configData.default">
                <i class="bk-radio-text">false</i>
            </label>
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
                        'default': true,
                        'description': '',
                        'label': '',
                        'type': 'boolean',
                        'show_subquestion_if': false,
                        'group': ''
                    }
                }
            }
        },
        mounted () {
            // console.log(this.configData)
        }
    }
</script>

<style scoped>
    .bk-form-radio input[type=radio].checked {
        background-position: -33px 0;
    }
</style>
