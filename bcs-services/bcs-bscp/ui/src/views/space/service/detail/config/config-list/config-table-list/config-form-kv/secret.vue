<template>
  <bk-form-item :label="t('配置项值')" property="value" required>
    <div class="secret-list">
      <div
        :class="['secret-item', { active: selectType === item.value }]"
        v-for="item in secretType"
        :key="item.value"
        @click="handleChangeCurrentType(item.value)">
        {{ item.label }}
      </div>
    </div>
    <div v-if="selectType === 'certificate' || selectType === 'customize'" class="certificate-content">
      <SecretContentEditor
        :content="secretValue"
        :is-credential="selectType === 'certificate'"
        :height="400"
        @change="handlSecretChange" />
    </div>
    <div v-else class="value-content">
      <bk-popover
        :is-show="isShowValidateInfo"
        ext-cls="secret-validate-info-popover"
        trigger="manual"
        placement="right"
        theme="light">
        <div class="secret-content">
          <bk-input
            :model-value="valueVisible ? secretValue : encryptValue"
            class="value-input"
            :placeholder="selectTypeContent?.placeholder"
            @input="handleValueChange"
            @focus="handleInputFocus"
            @blur="isShowValidateInfo = false">
            <template #suffix>
              <info-line
                v-if="selectTypeContent!.infoList.some((item) => item.status === 'warn') && secretValue"
                class="warn-icon" />
              <Eye v-if="valueVisible" class="view-icon" @click="valueVisible = false" />
              <Unvisible v-else class="view-icon" @click="valueVisible = true" />
            </template>
          </bk-input>
        </div>
        <template #content>
          <div class="info-list">
            <div v-for="item in selectTypeContent?.infoList" :key="item.text" class="info-item">
              <span :class="['dot', item.status]"></span> {{ item.text }}
            </div>
          </div>
        </template>
      </bk-popover>
    </div>
  </bk-form-item>
  <bk-checkbox class="visible-checkbox" :model-value="secretUnVisible" :before-change="handleChangeSecretUnVisible">
    {{ t('敏感信息不可见') }}
  </bk-checkbox>
  <bk-dialog
    :is-show="isShowVisibleDialog"
    :title="t('「敏感信息不可见」启用提示')"
    class="confirm-dialog"
    :width="480"
    @close="isShowVisibleDialog = false">
    <div class="dialog-content">
      <div>{{ t('启用后，可降低数据泄露风险。尽管客户端拉去配置不受影响，但仍存在以下注意事项：') }}</div>
      <div>{{ t('⒈ 用户无法直接查看敏感数据，将导致查看和对比操作不便') }}</div>
      <div>{{ t('⒉ 编辑敏感信息时，需重新输入完整数据 ') }}</div>
      <div>{{ t('⒊ 若忘记或丢失敏感信息，可能需要通过其他途径找回或重置') }}</div>
    </div>
    <div class="dialog-footer">
      <bk-button theme="primary" @click="confirmEnable">{{ t('确定启用') }}</bk-button>
      <bk-button @click="isShowVisibleDialog = false">{{ t('取消') }}</bk-button>
    </div>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Eye, Unvisible, InfoLine } from 'bkui-vue/lib/icon';
  import SecretContentEditor from './secret-content-editor.vue';

  const { t } = useI18n();

  const selectType = ref('password');

  const valueVisible = ref(false); // 密文展示敏感信息
  const secretUnVisible = ref(false); // 敏感信息不可见
  const isShowVisibleDialog = ref(false);
  const secretValue = ref('');
  const isShowValidateInfo = ref(false);

  const selectTypeContent = computed(() => {
    return secretType.find((item) => item.value === selectType.value);
  });

  const encryptValue = computed(() => {
    return '*'.repeat(secretValue.value.length);
  });

  const secretType = [
    {
      label: t('密码'),
      value: 'password',
      infoList: [
        { status: 'warn', text: t('建议长度 {n} 字符', { n: '8~64' }) },
        { status: 'warn', text: t('至少包含大写字母、小写字母、数字和特殊字符中的 3 种类型') },
      ],
      placeholder: t('请输入密码'),
    },
    {
      label: t('证书'),
      value: 'certificate',
      infoList: [
        { status: 'warn', text: t('建议长度 {n} 字符', { n: '8~64' }) },
        { status: 'warn', text: t('至少包含大写字母、小写字母、数字和特殊字符中的 3 种类型') },
      ],
    },
    {
      label: t('API密钥'),
      value: 'secret_key',
      infoList: [
        { status: 'warn', text: t('建议长度 {n} 字符', { n: '16~64' }) },
        { status: 'warn', text: t('包含大写字母、小写字母和数字') },
      ],
      placeholder: t('请输入 API 密钥'),
    },
    {
      label: t('访问令牌'),
      value: 'token',
      infoList: [
        { status: 'warn', text: t('建议长度 {n} 字符', { n: '32~512' }) },
        { status: 'warn', text: t('包含大写字母、小写字母和数字') },
      ],
      placeholder: t('请输入访问令牌，目前只支持 OAuth2.0 与 JWT 类型的访问令牌'),
    },
    {
      label: t('自定义'),
      value: 'customize',
      infoList: [
        { status: 'warn', text: t('建议长度 {n} 字符', { n: '8~64' }) },
        { status: 'warn', text: t('至少包含大写字母、小写字母、数字和特殊字符中的 3 种类型') },
      ],
    },
  ];

  const handleChangeCurrentType = (type: string) => {
    selectType.value = type;
    secretValue.value = '';
  };

  const handleValueChange = (value: string, event: any) => {
    if (value.length > secretValue.value.length) {
      // 添加的内容长度
      const addLength = value.length - secretValue.value.length;
      // 添加索引
      const addIndex = event.target.selectionStart - addLength;
      // 添加内容
      const addContent = value.slice(addIndex, event.target.selectionStart);
      secretValue.value = secretValue.value.slice(0, addIndex) + addContent + secretValue.value.slice(addIndex);
    } else if (value.length < secretValue.value.length) {
      // 删除的内容长度
      const deleteLength = secretValue.value.length - value.length;
      // 删除索引
      const deleteIndex = event.target.selectionStart;
      secretValue.value = secretValue.value.slice(0, deleteIndex) + secretValue.value.slice(deleteIndex + deleteLength);
    }
    validateSecretValue();
  };

  const handleChangeSecretUnVisible = (val: boolean) => {
    if (val) {
      isShowVisibleDialog.value = true;
      return false;
    }
    secretUnVisible.value = false;
    return true;
  };

  const confirmEnable = () => {
    secretUnVisible.value = true;
    isShowVisibleDialog.value = false;
  };

  const handleInputFocus = () => {
    isShowValidateInfo.value = true;
    validateSecretValue();
  };

  // 校验密钥内容
  const validateSecretValue = () => {
    if (selectType.value === 'password') {
      if (secretValue.value.length < 8 || secretValue.value.length > 64) {
        selectTypeContent.value!.infoList[0].status = 'warn';
      } else {
        selectTypeContent.value!.infoList[0].status = 'success';
      }
      if (checkSecretFormat()) {
        selectTypeContent.value!.infoList[1].status = 'success';
      } else {
        selectTypeContent.value!.infoList[1].status = 'warn';
      }
    } else if (selectType.value === 'secret_key') {
      if (secretValue.value.length < 16 || secretValue.value.length > 64) {
        selectTypeContent.value!.infoList[0].status = 'warn';
      } else {
        selectTypeContent.value!.infoList[0].status = 'success';
      }
      if (checkSecretFormat(false)) {
        selectTypeContent.value!.infoList[1].status = 'success';
      } else {
        selectTypeContent.value!.infoList[1].status = 'warn';
      }
    } else if (selectType.value === 'token') {
      if (secretValue.value.length < 32 || secretValue.value.length > 512) {
        selectTypeContent.value!.infoList[0].status = 'warn';
      } else {
        selectTypeContent.value!.infoList[0].status = 'success';
      }
      if (checkSecretFormat(false)) {
        selectTypeContent.value!.infoList[1].status = 'success';
      } else {
        selectTypeContent.value!.infoList[1].status = 'warn';
      }
    }
  };

  // 判断input框内容格式
  const checkSecretFormat = (isPassword = true) => {
    if (secretValue.value.length < 3) return false;
    let hasUppercase = false; // 是否包含大写字母
    let hasLowercase = false; // 是否包含小写字母
    let hasDigit = false; // 是否包含数字
    let hasSpecialChar = false; // 是否包含特殊字符
    // 将字符串转成字符数组
    Array.from(secretValue.value).forEach((char) => {
      if (/[A-Z]/.test(char)) {
        hasUppercase = true;
      } else if (/[a-z]/.test(char)) {
        hasLowercase = true;
      } else if (/\d/.test(char)) {
        hasDigit = true;
      } else if (/[^A-Za-z\d]/.test(char)) {
        hasSpecialChar = true;
      }
    });
    if (isPassword) {
      // 计算满足的类型数量
      return [hasUppercase, hasLowercase, hasDigit, hasSpecialChar].filter(Boolean).length >= 3;
    }
    return hasUppercase && hasLowercase && hasDigit;
  };

  const handlSecretChange = (val: string) => {
    secretValue.value = val;
  };
</script>

<style scoped lang="scss">
  .secret-list {
    display: flex;
    align-items: center;
    margin-bottom: 12px;
    .secret-item {
      height: 26px;
      width: 80px;
      line-height: 26px;
      background: #f8f8f8;
      text-align: center;
      background: #ffffff;
      border: 1px solid #c4c6cc;
      color: #63656e;
      cursor: pointer;
      &.active {
        background: #e1ecff;
        border: 1px solid #3a84ff;
        color: #3a84ff;
      }
    }
  }
  .secret-content {
    .value-input {
      width: 560px;
      .view-icon {
        cursor: pointer;
        margin: 0 8px;
        font-size: 14px;
        color: #979ba5;
        &:hover {
          color: #3a84ff;
        }
      }
      .warn-icon {
        font-size: 14px;
        color: #ff9c01;
      }
    }
  }

  .confirm-dialog {
    :deep(.bk-modal-wrapper) {
      .bk-modal-body {
        padding-bottom: 24px;
      }
      .bk-modal-header {
        .bk-dialog-title {
          text-align: center !important;
        }
      }
      .dialog-content {
        padding: 12px 16px;
        width: 416px;
        background: #f5f6fa;
        border-radius: 2px;
        font-size: 14px;
        color: #63656e;
        line-height: 22px;
      }
      .dialog-footer {
        display: flex;
        justify-content: center;
        gap: 8px;
        margin-top: 24px;
      }
      .bk-modal-footer {
        display: none;
      }
    }
  }

  .secret-validate-info-popover {
    .info-list {
      .info-item {
        font-size: 12px;
        color: #63656e;
        line-height: 16px;
        &:not(:last-child) {
          margin-bottom: 8px;
        }
        .dot {
          display: inline-block;
          width: 6px;
          height: 6px;
          border-radius: 50%;
          &.warn {
            background: #ff9c01;
          }
          &.success {
            background: #2dcb56;
          }
        }
      }
    }
  }
</style>

<style>
  .secret-validate-info-popover.bk-popover.bk-pop2-content {
    padding: 8px;
  }
</style>
