<template>
  <bcs-dialog
    :value="value"
    theme="primary"
    :mask-close="false"
    :title="isEdit ? $t('编辑项目') : $t('新建项目')"
    width="860"
    :loading="loading"
    :auto-close="false"
    render-directive="if"
    @value-change="handleChange"
    @confirm="handleConfirm">
    <bk-form :label-width="labelWidth" :model="formData" :rules="rules" ref="bkFormRef">
      <bk-form-item :label="$t('项目名称')" property="name" error-display-type="normal" required>
        <bk-input
          :maxlength="64"
          class="create-input"
          :placeholder="$t('请输入2-64字符的项目名称')"
          v-model="formData.name">
        </bk-input>
      </bk-form-item>
      <bk-form-item :label="$t('项目英文名')" property="projectCode" error-display-type="normal" required>
        <bk-input
          class="create-input"
          :maxlength="32"
          :placeholder="$t('请输入2-32字符的小写字母、数字、中划线，以小写字母开头')"
          :disabled="isEdit"
          v-model="formData.projectCode">
        </bk-input>
      </bk-form-item>
      <bk-form-item :label="$t('项目说明')" property="description" error-display-type="normal" required>
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
import { computed, defineComponent, ref, toRefs, watch } from 'vue';
import useFormLabel from '@/composables/use-form-label';
import useProjects from '@/views/project-manage/project/use-project';
import { SPECIAL_REGEXP } from '@/common/constant';
import $store from '@/store';
import $i18n from '@/i18n/i18n-setup';
import $bkMessage from '@/common/bkmagic';

export default defineComponent({
  name: 'ProjectCreate',
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Boolean,
      default: false,
    },
    projectData: {
      type: Object,
      default: () => ({}),
    },
  },
  setup(props, ctx) {
    const { projectData, value } = toRefs(props);
    const { emit } = ctx;
    const { updateProject, createProject } = useProjects();
    const bkFormRef = ref<any>(null);
    const formData = ref({
      name: projectData?.value?.name,
      projectCode: projectData?.value?.projectCode,
      description: projectData?.value?.description,
    });
    const rules = ref({
      name: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
        {
          message: $i18n.t('项目名称校验提示语'),
          trigger: 'blur',
          validator(value) {
            return /^[\w\W]{2,64}$/g.test(value) && !SPECIAL_REGEXP.test(value);
          },
        },
      ],
      projectCode: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
        {
          message: $i18n.t('请输入2-32字符的小写字母、数字、中划线，以小写字母开头'),
          trigger: 'blur',
          validator(value) {
            return /^[a-z][a-z0-9-]{1,31}$/g.test(value);
          },
        },
      ],
      description: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
    });
    watch(value, (isShow) => {
      if (isShow) {
        formData.value = {
          name: projectData?.value?.name,
          projectCode: projectData?.value?.projectCode,
          description: projectData?.value?.description,
        };
        setTimeout(() => {
          initFormLabelWidth(bkFormRef.value);
        }, 0);
      }
    });
    const loading = ref(false);
    const isEdit = computed(() => !!(projectData?.value && Object.keys(projectData.value).length));
    const handleChange = (value) => {
      emit('change', value);
    };
    const handleCreateProject = async () => {
      const result = await createProject({
        description: formData.value.description,
        name: formData.value.name,
        projectCode: formData.value.projectCode,
      });

      return result;
    };
    const handleEditProject = async () => {
      const result = await updateProject(Object.assign(
        {},
        projectData.value,
        {
          description: formData.value.description,
          name: formData.value.name,
          $projectId: projectData.value.projectID,
        },
      ));

      return result;
    };
    const handleConfirm = async () => {
      const validate = await bkFormRef.value?.validate();
      if (!validate) return;

      let result = false;
      loading.value = true;
      if (isEdit.value) {
        result = await handleEditProject();
      } else {
        result = await handleCreateProject();
      }
      loading.value = false;
      if (result) {
        // 更新集群列表
        await $store.dispatch('getProjectList');
        $bkMessage({
          message: isEdit.value ? $i18n.t('编辑成功') : $i18n.t('创建成功'),
          theme: 'success',
        });
        handleChange(false);
        emit('finished');
      }
      return result;
    };
    const { initFormLabelWidth, labelWidth } = useFormLabel();

    return {
      labelWidth,
      bkFormRef,
      isEdit,
      loading,
      formData,
      rules,
      handleChange,
      handleCreateProject,
      handleEditProject,
      handleConfirm,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .form-error-tip {
  text-align: left;
}
</style>
