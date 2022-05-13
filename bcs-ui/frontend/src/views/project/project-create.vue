<template>
    <bcs-dialog :value="value"
        theme="primary"
        :mask-close="false"
        :title="isEdit ? $t('编辑项目') : $t('新建项目')"
        width="860"
        :loading="loading"
        @value-change="handleChange"
        @confirm="handleConfirm">
        <bk-form :label-width="labelWidth" v-model="formData" ref="bkFormRef">
            <bk-form-item :label="$t('项目名称')" property="project_name" required>
                <bk-input class="create-input" :placeholder="$t('请输入4-12字符的项目名称')" v-model="formData.project_name"></bk-input>
            </bk-form-item>
            <bk-form-item :label="$t('项目英文名')" property="english_name" required>
                <bk-input class="create-input" :placeholder="$t('请输入2-32字符的小写字母+数字，以小写字母开头')" :disabled="isEdit"
                    v-model="formData.english_name"></bk-input>
            </bk-form-item>
            <bk-form-item :label="$t('项目说明')" property="description" required>
                <bk-input
                    class="create-input"
                    :placeholder="$t('请输入项目描述')"
                    type="textarea"
                    :rows="3"
                    :maxlength="100"
                    v-model="formData.description">
                </bk-input>
            </bk-form-item>
        </bk-form>
    </bcs-dialog>
</template>
<script lang="ts">
    /* eslint-disable camelcase */
    import { computed, defineComponent, ref, toRefs, watch } from '@vue/composition-api'
    import { createProject, editProject } from '@/api/base'
    import useFormLabel from '@/common/use-form-label'
    export default defineComponent({
        name: "ProjectCreate",
        model: {
            prop: 'value',
            event: 'change'
        },
        props: {
            value: {
                type: Boolean,
                default: false
            },
            projectData: {
                type: Object,
                default: () => ({})
            }
        },
        setup: (props, ctx) => {
            const { projectData, value } = toRefs(props)
            const { emit } = ctx
            const { $bkMessage, $i18n, $store } = ctx.root
            const formData = ref({
                project_name: projectData?.value?.project_name,
                english_name: projectData?.value?.english_name,
                description: projectData?.value?.description
            })
            watch(value, (isShow) => {
                if (isShow) {
                    formData.value = {
                        project_name: projectData?.value?.project_name,
                        english_name: projectData?.value?.english_name,
                        description: projectData?.value?.description
                    }
                    setTimeout(() => {
                        initFormLabelWidth(bkFormRef.value)
                    }, 0)
                }
            })
            const loading = ref(false)
            const isEdit = computed(() => {
                return projectData?.value && Object.keys(projectData.value).length
            })
            const handleChange = (value) => {
                emit('change', value)
            }
            const handleCreateProject = async () => {
                const result = await createProject({
                    bg_id: "",
                    bg_name: "",
                    center_id: "",
                    center_name: "",
                    deploy_type: [],
                    dept_id: "",
                    dept_name: "",
                    description: formData.value.description,
                    english_name: formData.value.english_name,
                    is_secrecy: false,
                    kind: "0",
                    project_name: formData.value.project_name,
                    project_type: ""
                }).catch(() => false)

                return result
            }
            const handleEditProject = async () => {
                const result = await editProject({
                    description: formData.value.description,
                    project_name: formData.value.project_name,
                    $projectId: projectData.value.project_id
                }).catch(() => false)

                return result
            }
            const handleConfirm = async () => {
                let result = false
                loading.value = true
                if (isEdit.value) {
                    result = await handleEditProject()
                } else {
                    result = await handleCreateProject()
                }
                loading.value = false
                if (result) {
                    // 更新集群列表
                    await $store.dispatch('getProjectList')
                    $bkMessage({
                        message: isEdit.value ? $i18n.t('编辑成功') : $i18n.t('创建成功'),
                        theme: 'success'
                    })
                    handleChange(false)
                }
                return result
            }
            const { initFormLabelWidth, labelWidth } = useFormLabel()
            const bkFormRef = ref(null)
            return {
                labelWidth,
                bkFormRef,
                isEdit,
                loading,
                formData,
                handleChange,
                handleCreateProject,
                handleEditProject,
                handleConfirm
            }
        }
    })
</script>
<style lang="postcss" scoped>
</style>
