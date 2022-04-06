<template>
    <section class="create-import-cluster">
        <bk-form :label-width="130" :model="importClusterInfo" :rules="formRules" class="import-form" ref="importFormRef">
            <bk-form-item :label="$t('集群名称')" property="clusterName" error-display-type="normal" required>
                <bk-input v-model="importClusterInfo.clusterName"></bk-input>
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
            <bk-form-item :label="$t('模板名称')" property="provider" error-display-type="normal" required>
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
            <bk-form-item :label="$t('集群描述')">
                <bk-input v-model="importClusterInfo.description" type="textarea"></bk-input>
            </bk-form-item>
            <bk-form-item :label="$t('集群kubeconfig')" property="yaml" error-display-type="normal" required>
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
                <bk-button class="btn" theme="primary" @click="handleImport">{{$t('导入')}}</bk-button>
                <bk-button class="btn" @click="handleCancel">{{$t('取消')}}</bk-button>
            </bk-form-item>
        </bk-form>
        <bcs-dialog v-model="showImportDialog"
            theme="primary"
            header-position="left"
            :title="$t('导入集群')"
            width="640">
            <div class="import-cluster-dialog">
                <div class="title">{{$t('导入前提，请确认')}}:</div>
                <bcs-checkbox-group v-model="importConfirm">
                    <bcs-checkbox value="1">{{$t('蓝鲸部署服务器到被导入集群APIServer网络连通正常')}}</bcs-checkbox>
                    <bcs-checkbox value="2" class="mt10">{{$t('kubeconfig用户是cluster-admin角色')}}</bcs-checkbox>
                </bcs-checkbox-group>
            </div>
            <template #footer>
                <div>
                    <bcs-button :disabled="importConfirm.length < 2"
                        theme="primary"
                        :loading="loading"
                        @click="confirmDialog"
                    >{{$t('确定，导入集群')}}</bcs-button>
                    <bcs-button @click="cancelDialog">{{$t('我再想想')}}</bcs-button>
                </div>
            </template>
        </bcs-dialog>
    </section>
</template>
<script lang="ts">
    import { defineComponent, ref, computed, onMounted } from '@vue/composition-api'
    import FormGroup from './form-group.vue'
    import Ace from '@/components/ace-editor'
    import useGoHome from '@/common/use-gohome'
    import useConfig from '@/common/use-config'

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
                clusterName: '',
                environment: 'prod',
                description: '',
                yaml: '',
                provider: ''
            })
            const loading = ref(false)
            const showImportDialog = ref(false)
            const importConfirm = ref([])
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

            const handleImport = async () => {
                const result = await importFormRef.value.validate()
                if (!result) return
                showImportDialog.value = true
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
            const confirmDialog = async () => {
                loading.value = true
                const params = {
                    clusterName: importClusterInfo.value.clusterName,
                    description: importClusterInfo.value.description,
                    projectID: curProject.value.project_id,
                    businessID: String(curProject.value.cc_app_id),
                    provider: importClusterInfo.value.provider,
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
            const cancelDialog = () => {
                showImportDialog.value = false
            }
            const templateList = ref<any>([])
            const templateLoading = ref(false)
            const handleGetTemplateList = async () => {
                templateLoading.value = true
                templateList.value = await $store.dispatch('clustermanager/fetchCloudList')
                importClusterInfo.value.provider = templateList.value[0]?.cloudID || ''
                templateLoading.value = false
            }
            onMounted(() => {
                handleGetTemplateList()
            })
            return {
                templateList,
                importClusterInfo,
                importFormRef,
                loading,
                showImportDialog,
                importConfirm,
                formRules,
                handleFileChange,
                handleImport,
                handleCancel,
                confirmDialog,
                cancelDialog,
                $INTERNAL
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
