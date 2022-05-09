<template>
    <section class="create-import-cluster">
        <bk-form :label-width="labelWidth" :model="importClusterInfo" :rules="formRules" class="import-form" ref="importFormRef">
            <bk-form-item :label="$t('集群名称')" property="clusterName" error-display-type="normal" required>
                <bk-input v-model="importClusterInfo.clusterName"></bk-input>
            </bk-form-item>
            <bk-form-item :label="$t('集群描述')">
                <bk-input v-model="importClusterInfo.description" type="textarea"></bk-input>
            </bk-form-item>
            <bk-form-item :label="$t('集群环境')" required v-if="$INTERNAL">
                <bk-radio-group v-model="importClusterInfo.environment">
                    <bk-radio value="debug">
                        {{ $t('测试环境') }}
                    </bk-radio>
                    <bk-radio value="prod">
                        {{ $t('正式环境') }}
                    </bk-radio>
                </bk-radio-group>
            </bk-form-item>
            <!-- <bk-form-item :label="$t('导入方式')">
                <bk-radio-group v-model="importClusterInfo.importType">
                    <bk-radio value="kubeconfig">{{$t('kubeconfig')}}</bk-radio>
                    <bk-radio value="provider">{{$t('云服务商')}}</bk-radio>
                </bk-radio-group>
            </bk-form-item> -->
            <bk-form-item :label="$t('云服务商')"
                property="provider"
                error-display-type="normal"
                required
                v-if="importClusterInfo.importType === 'provider'">
                <bcs-select :loading="templateLoading" class="w640"
                    v-model="importClusterInfo.provider"
                    :clearable="false">
                    <bcs-option v-for="item in templateList"
                        :key="item.cloudID"
                        :id="item.cloudID"
                        :name="item.name">
                    </bcs-option>
                </bcs-select>
            </bk-form-item>
            <bk-form-item :label="$t('集群kubeconfig')"
                property="yaml"
                error-display-type="normal"
                required>
                <bk-button class="mb10">
                    <input class="file-input"
                        accept=".yaml,yml"
                        type="file"
                        @change="handleFileChange">
                    {{$t('文件导入')}}
                </bk-button>
                <Ace class="cube-config"
                    lang="yaml"
                    width="100%"
                    :height="480"
                    v-model="importClusterInfo.yaml"
                    :show-gutter="false"
                ></Ace>
            </bk-form-item>
            <bk-form-item>
                <bk-button theme="primary"
                    :loading="testLoading"
                    @click="handleTest"
                >{{$t('kubeconfig可用性测试')}}</bk-button>
                <bk-button class="btn"
                    theme="primary"
                    :loading="loading"
                    :disabled="!isTestSuccess"
                    @click="handleImport"
                >{{$t('导入')}}</bk-button>
                <bk-button class="btn"
                    @click="handleCancel"
                >{{$t('取消')}}</bk-button>
            </bk-form-item>
        </bk-form>
    </section>
</template>
<script lang="ts">
    import { defineComponent, ref, computed, onMounted } from '@vue/composition-api'
    import FormGroup from './form-group.vue'
    import Ace from '@/components/ace-editor'
    import useGoHome from '@/common/use-gohome'
    import useConfig from '@/common/use-config'
    import useFormLabel from '@/common/use-form-label'

    export default defineComponent({
        name: 'CreateImportCluster',
        components: {
            FormGroup,
            Ace
        },
        setup (props, ctx) {
            const { $router, $bkMessage, $i18n, $route, $store } = ctx.root
            const { goHome } = useGoHome()
            const { $INTERNAL } = useConfig()
            const importClusterInfo = ref({
                importType: 'kubeconfig',
                clusterName: '',
                environment: 'prod',
                description: '',
                yaml: '',
                provider: ''
            })
            const isTestSuccess = ref(false)
            const testLoading = ref(false)
            const loading = ref(false)
            const importFormRef = ref<any>(null)
            const formRules = ref({
                provider: [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                yaml: [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                clusterName: [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ]
            })

            const handleFileChange = (event) => {
                const [file] = event.target.files
                if (!file) return
                const reader = new FileReader()
                reader.readAsText(file, 'UTF-8')
                reader.onload = (e) => {
                    importClusterInfo.value.yaml = e?.target?.result as string || ''
                }
            }

            const handleCancel = () => {
                $router.back()
            }
            
            const curProject = computed(() => {
                return $store.state.curProject
            })
            const user = computed(() => {
                return $store.state.user
            })
            const handleTest = async () => {
                const validate = await importFormRef.value.validate()
                if (!validate) return
                testLoading.value = true
                const result = await $store.dispatch('clustermanager/kubeConfig', {
                    kubeConfig: importClusterInfo.value.yaml
                })
                if (result) {
                    isTestSuccess.value = true
                }
                testLoading.value = false
            }
            const handleImport = async () => {
                const validate = await importFormRef.value.validate()
                if (!validate) return

                loading.value = true
                const params = {
                    clusterName: importClusterInfo.value.clusterName,
                    description: importClusterInfo.value.description,
                    projectID: curProject.value.project_id,
                    businessID: String(curProject.value.cc_app_id),
                    provider: 'bluekingCloud', // importClusterInfo.value.provider,
                    region: 'default',
                    environment: "prod",
                    engineType: "k8s",
                    isExclusive: true,
                    clusterType: "single",
                    manageType: 'INDEPENDENT_CLUSTER',
                    creator: user.value.username,
                    cloudMode: {
                        cloudID: "",
                        kubeConfig: importClusterInfo.value.yaml
                    },
                    version: '',
                    networkType: "overlay"
                }
                const result = await $store.dispatch('clustermanager/importCluster', params)
                loading.value = false
                if (result) {
                    $bkMessage({
                        theme: 'success',
                        message: $i18n.t('导入成功')
                    })
                    goHome($route)
                }
            }
            const templateList = ref<any>([])
            const templateLoading = ref(false)
            const handleGetTemplateList = async () => {
                templateLoading.value = true
                templateList.value = await $store.dispatch('clustermanager/fetchCloudList')
                importClusterInfo.value.provider = templateList.value[0]?.cloudID || ''
                templateLoading.value = false
            }
            const { labelWidth, initFormLabelWidth } = useFormLabel()
            onMounted(() => {
                handleGetTemplateList()
                initFormLabelWidth(importFormRef.value)
            })
            return {
                isTestSuccess,
                testLoading,
                labelWidth,
                templateList,
                importClusterInfo,
                importFormRef,
                loading,
                formRules,
                handleFileChange,
                handleImport,
                handleCancel,
                $INTERNAL,
                handleTest
            }
        }
    })
</script>
<style lang="postcss" scoped>
.create-import-cluster {
    padding: 24px;
    /deep/ .w640 {
        width: 640px;
        background: #fff;
    }
    /deep/ .bk-input-text {
        width: 640px;
    }
    /deep/ .bk-textarea-wrapper {
        width: 640px;
        height: 80px;
    }
    /deep/ .cube-config {
        max-width: 1000px;
    }
    /deep/ .btn {
        width: 88px;
    }
    /deep/ .file-input {
        position: absolute;
        opacity: 0;
        left: 0;
        top: 0;
        width: 100%;
        height: 100%;
    }
}
.import-cluster-dialog {
    padding: 0 4px 24px 4px;
    .title {
        font-size: 14px;
        font-weight: 700;
        margin-bottom: 14px
    }
}
</style>
