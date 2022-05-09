<template>
    <div class="biz-content form-resource" v-bkloading="{ isLoading }">
        <SwitchButton
            class="switch-yaml"
            :title="$t('切换为 YAML 模式')"
            @click="handleSwitchMode" />
        <div class="biz-top-bar">
            <span class="icon-wrapper" @click="handleCancel">
                <i class="bcs-icon bcs-icon-arrows-left icon-back"></i>
            </span>
            <div class="dashboard-top-title">
                {{ title }}
            </div>
            <DashboardTopActions />
        </div>
        <BKForm v-model="schemaFormData"
            ref="bkuiFormRef"
            class="form-resource-content"
            :schema="formSchema.schema"
            :layout="formSchema.layout"
            :context="context"
            :http-adapter="{
                request
            }"
            form-type="vertical"
        ></BKForm>
        <div class="footer">
            <bk-button class="btn"
                theme="primary"
                :loading="loading"
                @click="handleSaveFormData">
                {{isEdit ? $t('更新') : $t('创建')}}
            </bk-button>
            <bk-button class="btn ml15" @click="handleCancel">{{$t('取消')}}</bk-button>
        </div>
    </div>
</template>
<script>
    import createForm from '@/components/bkui-form-umd'
    import request from '@/api/request'
    import DashboardTopActions from '../common/dashboard-top-actions'
    import SwitchButton from './switch-mode.vue'
    import { CUR_SELECT_NAMESPACE } from '@/common/constant'

    const BKForm = createForm({
        namespace: 'bcs',
        baseWidgets: {
            'radio': 'bk-radio',
            'radio-group': 'bk-radio-group'
        }
    })
    export default {
        components: {
            BKForm,
            DashboardTopActions,
            SwitchButton
        },
        props: {
            // 命名空间（更新的时候需要--crd类型编辑是可能没有，创建的时候为空）
            namespace: {
                type: String,
                default: ''
            },
            // 父分类，eg: workloads、networks（注意复数）
            type: {
                type: String,
                default: '',
                required: true
            },
            // 子分类，eg: deployments、ingresses
            category: {
                type: String,
                default: ''
            },
            // 名称（更新的时候需要，创建的时候为空）
            name: {
                type: String,
                default: ''
            },
            kind: {
                type: String,
                default: ''
            },
            // type 为crd时，必传
            crd: {
                type: String,
                default: ''
            },
            clusterId: {
                type: String,
                default: ''
            },
            formData: {
                type: Object,
                default: () => ({})
            }
        },
        data () {
            return {
                schemaFormData: this.formData,
                formSchema: {},
                isLoading: false,
                loading: false
            }
        },
        computed: {
            curProject () {
                return this.$store.state.curProject
            },
            context () {
                return Object.assign({
                    clusterID: this.clusterId,
                    projectID: this.curProject.project_id
                }, {
                    ...this.formSchema.context,
                    baseUrl: '/bcsapi/v4/clusterresources/v1' // todo
                })
            },
            isEdit () {
                return !!this.name
            },
            title () {
                const prefix = this.isEdit ? this.$t('更新') : this.$t('创建')
                return `${prefix} ${this.kind}`
            }
        },
        created () {
            this.handleGetFormSchemaData()
            this.handleGetDetail()
        },
        methods: {
            async handleGetDetail () {
                if (!this.isEdit || (this.formData && Object.keys(this.formData).length)) return

                let res = null
                if (this.type === 'crd') {
                    res = await this.$store.dispatch('dashboard/retrieveCustomResourceDetail', {
                        $crd: this.crd,
                        $category: this.category,
                        $name: this.name,
                        namespace: this.namespace,
                        format: 'formData'
                    })
                } else {
                    res = await this.$store.dispatch('dashboard/getResourceDetail', {
                        $namespaceId: this.namespace,
                        $category: this.category,
                        $name: this.name,
                        $type: this.type,
                        format: 'formData'
                    })
                }
                this.schemaFormData = res.data.formData
            },
            async handleGetFormSchemaData () {
                this.isLoading = true
                this.formSchema = await this.$store.dispatch('dashboard/getFormSchema', {
                    kind: this.kind
                })
                this.isLoading = false
            },
            async request (url, config) {
                const requestMethods = request(config.method || 'get', url)
                const data = await requestMethods(config.params)
                return data?.selectItems || []
            },
            handleCancel () { // 取消
                this.$router.push({ name: this.$store.getters.curNavName })
            },
            // 切换Yaml模式
            async handleSwitchMode () {
                let params = {}
                if (this.isEdit) {
                    params = {
                        name: this.name,
                        namespace: this.namespace
                    }
                } else {
                    params = {
                        defaultShowExample: true
                    }
                }
                this.$router.push({
                    name: 'dashboardResourceUpdate',
                    params: {
                        ...params,
                        formData: this.schemaFormData,
                        editMode: 'form'
                    },
                    query: {
                        type: this.type,
                        category: this.category,
                        kind: this.kind,
                        crd: this.crd
                    }
                })
            },
            // 保存数据
            async handleSaveFormData () {
                const valid = this.$refs.bkuiFormRef.validateForm()
                console.log(valid)
                this.loading = true
                if (this.isEdit) {
                    await this.handleUpdateFormResource()
                } else {
                    await this.handleCreateFormResource()
                }
                this.loading = false
            },
            // 更新表单资源
            async handleUpdateFormResource () {
                this.$bkInfo({
                    type: 'warning',
                    clsName: 'custom-info-confirm',
                    title: this.$t('确认资源更新'),
                    subTitle: this.$t('将执行 Replace 操作，若多人同时编辑可能存在冲突'),
                    defaultInfo: true,
                    confirmFn: async () => {
                        let result = false
                        if (this.type === 'crd') {
                            result = await this.$store.dispatch('dashboard/customResourceUpdate', {
                                $crd: this.crd,
                                $category: this.category,
                                $name: this.name,
                                format: 'formData',
                                rawData: this.schemaFormData,
                                namespace: this.namespace
                            }).catch(err => {
                                console.log(err)
                                return false
                            })
                        } else {
                            result = await this.$store.dispatch('dashboard/resourceUpdate', {
                                $namespaceId: this.namespace,
                                $type: this.type,
                                $category: this.category,
                                $name: this.name,
                                format: 'formData',
                                rawData: this.schemaFormData
                            }).catch(err => {
                                console.log(err)
                                return false
                            })
                        }

                        if (result) {
                            this.$bkMessage({
                                theme: 'success',
                                message: this.$t('更新成功')
                            })
                            this.$router.push({ name: this.$store.getters.curNavName })
                        }
                    }
                })
            },
            // 创建表单资源
            async handleCreateFormResource () {
                let result = false
                if (this.type === 'crd') {
                    result = await this.$store.dispatch('dashboard/customResourceCreate', {
                        $crd: this.crd,
                        $category: this.category,
                        format: 'manifest',
                        rawData: this.schemaFormData
                    }).catch(err => {
                        console.error(err)
                        return false
                    })
                } else {
                    result = await this.$store.dispatch('dashboard/resourceCreate', {
                        $type: this.type,
                        $category: this.category,
                        format: 'formData',
                        rawData: this.schemaFormData
                    }).catch(err => {
                        console.error(err)
                        return false
                    })
                }

                if (result) {
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('创建成功')
                    })
                    sessionStorage.setItem(CUR_SELECT_NAMESPACE, this.schemaFormData.metadata?.namespace)
                    this.$router.push({ name: this.$store.getters.curNavName })
                }
            }
        }
    }
</script>
<style lang="postcss">
@import '@/components/bkui-form.css'
</style>
<style lang="postcss" scoped>
.form-resource {
    padding-bottom: 0;
    height: 100%;
    .switch-yaml {
        position: absolute;
        right: 16px;
        top: 72px;
        z-index: 1;
    }
    .icon-back {
        font-size: 16px;
        font-weight: bold;
        color: #3A84FF;
        margin-left: 20px;
        cursor: pointer;
    }
    .dashboard-top-title {
        display: inline-block;
        height: 60px;
        line-height: 60px;
        font-size: 16px;
        margin-left: 0px;
    }
    .form-resource-content {
        padding: 20px;
        max-height: calc(100vh - 172px);
        overflow: auto;
    }
    .footer {
        position: fixed;
        bottom: 0px;
        height: 60px;
        display: flex;
        align-items: center;
        justify-content: center;
        background-color: #fff;
        border-top: 1px solid #dcdee5;
        box-shadow: 0 -2px 4px 0 rgb(0 0 0 / 5%);
        z-index: 200;
        right: 0;
        width: calc(100% - 261px);
        .btn {
            width: 88px;
        }
    }
}
</style>
