<template>
    <section>
        <FormGroup :title="$t('基本信息')">
            <bcs-form :label-width="100">
                <bcs-form-item :label="$t('集群名称')" required>
                    <bcs-input></bcs-input>
                </bcs-form-item>
                <bcs-form-item :label="$t('集群环境')" required>
                    <bcs-radio-group>
                        <bcs-radio>
                            {{ $t('测试环境') }}
                        </bcs-radio>
                        <bcs-radio>
                            {{ $t('正式环境') }}
                        </bcs-radio>
                    </bcs-radio-group>
                </bcs-form-item>
                <bcs-form-item :label="$t('云服务商')" required>
                    <div class="template-name">
                        <bcs-select></bcs-select>
                        <bcs-button text class="ml10">{{ $t('注册云服务商') }}</bcs-button>
                    </div>
                </bcs-form-item>
                <bcs-form-item :label="$t('描述')">
                    <bcs-input></bcs-input>
                </bcs-form-item>
            </bcs-form>
        </FormGroup>
        <FormGroup :title="$t('选择Master')" :desc="$t('仅支持数量为3,5和7个')"></FormGroup>
    </section>
</template>
<script lang="ts">
    import { defineComponent } from '@vue/composition-api'
    import FormGroup from './form-group.vue'

    export default defineComponent({
        name: 'CreateImportCluster',
        components: {
            FormGroup
        },
        setup () {

        }
    })
</script>
<style lang="postcss" scoped>

</style>
