<template>
  <bk-form form-type="vertical" ref="formRef" :model="localData" :rules="rules">
    <bk-form-item :label="t('form_服务名称')" property="name" required>
      <bk-input
        v-model="localData.name"
        :placeholder="t('请输入2-32字符，只允许英文、数字、下划线、中划线且必须以英文、数字开头和结尾')"
        :disabled="editable"
        @input="handleChange"
        v-bk-tooltips="{
          content: t('请输入2-32字符，只允许英文、数字、下划线、中划线且必须以英文、数字开头和结尾'),
          disabled: locale === 'zh-cn',
        }" />
    </bk-form-item>
    <bk-form-item :label="t('form_服务别名')" property="alias" required>
      <bk-input
        v-model="localData.alias"
        :placeholder="t('请输入2-128字符，只允许中文、英文、数字、下划线、中划线且必须以中文、英文、数字开头和结尾')"
        @input="handleChange"
        v-bk-tooltips="{
          content: t('请输入2-128字符，只允许中文、英文、数字、下划线、中划线且必须以中文、英文、数字开头和结尾'),
          disabled: locale === 'zh-cn',
        }" />
    </bk-form-item>
    <bk-form-item :label="t('服务描述')" property="memo">
      <bk-input
        v-model="localData.memo"
        :placeholder="t('服务描述限制200字符')"
        type="textarea"
        :autosize="true"
        :resize="false"
        :maxlength="200"
        @input="handleChange" />
    </bk-form-item>
    <bk-form-item :label="t('数据格式')" :description="t('tips.config')">
      <bk-radio-group v-model="localData.config_type" :disabled="editable" @change="handleConfigTypeChange">
        <bk-radio label="file">{{ t('文件型') }}</bk-radio>
        <bk-radio label="kv">{{ t('键值型') }}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item
      v-if="localData.config_type === 'kv'"
      :label="t('数据类型')"
      property="data_type"
      :description="t('tips.type')"
      required>
      <bk-select v-model="localData.data_type" class="type-select" :clearable="false" @select="handleChange">
        <bk-option id="any" :name="t('任意类型')" />
        <bk-option v-for="kvType in CONFIG_KV_TYPE" :key="kvType.id" :id="kvType.id" :name="kvType.name" />
      </bk-select>
    </bk-form-item>
    <!-- 上线审批 -->
    <bk-form-item>
      <template #label>
        <div class="label-wrap">
          上线审批
          <help
            v-bk-tooltips="{
              content: $t(
                '建议在生产环境中开启审批流程，以保证系统稳定性。测试环境中可以考虑关闭审批流程以提升操作效率',
              ),
              placement: 'top',
            }"
            class="label-help" />
          <div class="label-switch">
            <bk-switcher v-model="localData.is_approve" theme="primary" size="small" @change="handleApproveSwitch" />
          </div>
        </div>
      </template>
      <div v-if="localData.is_approve" class="approval-content">
        <bk-form-item label="指定审批人" property="approver" required>
          <bk-member-selector
            v-model="selectionsApprover"
            :api="approverApi"
            :is-error="selValidationError"
            @change="changeApprover" />
        </bk-form-item>
        <bk-form-item property="approve_type">
          <template #label>
            <div class="label-wrap">
              审批方式
              <help
                v-bk-tooltips="{
                  content: $t('或签：多人同时审批，一人同意即可通过n会签：审批人依次审批，每人都需同意才能通过'),
                  placement: 'top',
                }"
                class="label-help" />
            </div>
          </template>
          <bk-radio-group v-model="localData.approve_type" @change="handleChange">
            <bk-radio label="OrSign">{{ t('或签') }}</bk-radio>
            <bk-radio label="CountSign">{{ t('会签') }}</bk-radio>
          </bk-radio-group>
        </bk-form-item>
      </div>
    </bk-form-item>
    <!-- <bk-form-item property="encryptionSwtich">
      <template #label>
        <div class="label-key">
          数据加密公钥 <help /><bk-switcher v-model="localData.encryptionSwtich" theme="primary" size="small" />
          <div class="key-management"><help /><a href="http://www.baidu.com" target="_blank">密钥管理</a></div>
        </div>
      </template>
    </bk-form-item> -->
  </bk-form>
  <bk-dialog
    v-model:is-show="approvalDialogShow"
    ref="dialog"
    ext-cls="confirm-dialog"
    footer-align="center"
    confirm-text="再想想"
    cancel-text="仍要关闭"
    :close-icon="true"
    :show-mask="true"
    :quick-close="false"
    :multi-instance="false"
    @confirm="
      localData.is_approve = true;
      approvalDialogShow = false;
    "
    @cancel="
      localData.is_approve = false;
      approvalDialogShow = true;
    ">
    <template #header>
      <div class="tip-icon__wrap">
        <exclamation-circle-shape class="tip-icon" />
      </div>
      <div class="headline">关闭上线审批存在风险</div>
    </template>
    <div class="content-info">
      <div>生产环境不建议关闭审批</div>
      <div>审批流程可以提高配置更改的准确性和安全性</div>
    </div>
  </bk-dialog>
</template>
<script setup lang="ts">
  import { onBeforeMount, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IServiceEditForm } from '../../../../../../types/service';
  import { CONFIG_KV_TYPE } from '../../../../../constants/config';
  import { Help, ExclamationCircleShape } from 'bkui-vue/lib/icon';
  import BkMemberSelector from '../../../../../components/user-selector/index.vue';

  const { t, locale } = useI18n();

  const emits = defineEmits(['change', 'approvalChange']);

  const props = defineProps<{
    formData: IServiceEditForm;
    editable?: boolean;
    approverApi: string;
  }>();

  const rules = {
    approver: [
      {
        required: true,
        message: t('指定审批人不能为空'),
        validator: (value: string) => value.length,
      },
    ],
    name: [
      {
        validator: (value: string) => value.length >= 2,
        message: t('最小长度2个字符'),
      },
      {
        validator: (value: string) => value.length <= 32,
        message: t('最大长度32个字符'),
      },
      {
        validator: (value: string) => /^[a-zA-Z0-9](?:[a-zA-Z0-9_-]*[a-zA-Z0-9])?$/.test(value),
        message: t('服务名称由英文、数字、下划线、中划线组成且以英文、数字开头和结尾'),
      },
    ],
    alias: [
      {
        validator: (value: string) => value.length >= 2,
        message: t('最小长度2个字符'),
      },
      {
        validator: (value: string) => value.length <= 128,
        message: t('最大长度128个字符'),
      },
      {
        validator: (value: string) =>
          /^[a-zA-Z0-9\u4e00-\u9fa5][a-zA-Z0-9_\-\u4e00-\u9fa5]*[a-zA-Z0-9\u4e00-\u9fa5]$/.test(value),
        message: t('服务别名由中文、英文、数字、下划线、中划线且必须以中文、英文、数字开头和结尾'),
      },
    ],
  };

  const localData = ref({ ...props.formData });
  const formRef = ref();
  const approvalDialogShow = ref(false);
  const selectionsApprover = ref<string[]>([]);
  const selValidationError = ref(false);

  watch(
    () => props.formData,
    (val) => {
      localData.value = { ...val };
    },
  );
  onBeforeMount(() => {
    formatApprover();
  });

  // 审批开关
  const handleApproveSwitch = (val: boolean) => {
    approvalDialogShow.value = !val;
    formatApprover();
    handleChange();
  };

  // 审批人变化
  const changeApprover = (data: string[]) => {
    if (data.length) {
      localData.value.approver = data.join(',').replace(/\s+/g, '');
    } else {
      localData.value.approver = '';
    }
    formRef.value.validate('approver'); // 验证审批人
    validateApprover();
    handleChange();
  };

  // 审批人格式转换
  const formatApprover = () => {
    if (localData.value.approver) {
      selectionsApprover.value = localData.value.approver.split(',');
    } else {
      selectionsApprover.value = [];
    }
  };

  // 验证审批人框的样式
  const validateApprover = () => {
    selValidationError.value = !localData.value.approver.length;
  };

  const handleConfigTypeChange = () => {
    if (localData.value.config_type === 'kv') {
      localData.value.data_type = 'any';
    } else {
      localData.value.data_type = '';
    }
    handleChange();
  };

  const handleChange = () => {
    emits('change', localData.value);
  };

  const validate = () => formRef.value.validate();

  defineExpose({
    validateApprover,
    validate,
  });
</script>

<style lang="scss" scoped>
  .type-select {
    width: 240px;
  }
  .label-key {
    display: flex;
    justify-content: flex-start;
    align-items: center;
  }
  .key-management {
    margin-left: auto;
  }
  .approval-content {
    padding: 12px 16px;
    background-color: #f5f7fa;
    .bk-form-item:last-child {
      margin-bottom: 0;
    }
    :deep(.bk-form-error) {
      top: 32px;
    }
  }
  .content-info {
    margin-top: 4px;
    padding: 12px 16px;
    font-size: 14px;
    line-height: 22px;
    color: #63656e;
    background-color: #f5f6fa;
  }
  .label-wrap {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    .label-help {
      margin: 0 9px;
      font-size: 16px;
      color: #979ba5;
      cursor: pointer;
    }
    .label-switch {
      position: relative;
      padding-left: 8px;
      height: 16px;
      line-height: 14px;
      &::after {
        content: '';
        position: absolute;
        left: 0;
        top: 0;
        height: 100%;
        border-left: 1px solid #dcdee5;
      }
    }
  }
  :deep(.confirm-dialog) {
    .bk-modal-body {
      padding-bottom: 0;
    }
    .bk-modal-content {
      padding: 0 32px;
      height: auto;
      max-height: none;
      min-height: auto;
      border-radius: 2px;
    }
    .bk-modal-footer {
      position: relative;
      padding: 24px 0;
      height: auto;
      border: none;
    }
    .bk-dialog-footer .bk-button {
      min-width: 88px;
    }
  }
  .tip-icon__wrap {
    margin: 0 auto;
    width: 42px;
    height: 42px;
    position: relative;
    &::after {
      content: '';
      position: absolute;
      z-index: -1;
      top: 50%;
      left: 50%;
      transform: translate3d(-50%, -50%, 0);
      width: 30px;
      height: 30px;
      border-radius: 50%;
      background-color: #ff9c01;
    }
    .tip-icon {
      font-size: 42px;
      line-height: 42px;
      vertical-align: middle;
      color: #ffe8c3;
    }
  }
  .headline {
    margin-top: 16px;
    text-align: center;
  }
  .user-selector {
    min-width: 100%;
  }
</style>
