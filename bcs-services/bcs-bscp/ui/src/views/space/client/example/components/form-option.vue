<template>
  <div class="headline">{{ $t('示例参数') }}</div>
  <bk-form class="form-example-wrap" :model="formData" :rules="rules" form-type="vertical" ref="formRef">
    <bk-form-item property="clientKey" required>
      <template #label>
        {{ $t('客户端密钥') }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: $t('用于客户端拉取配置时身份验证，下拉列表只会展示关联过此服务且状态为启用的密钥'),
            placement: 'top',
          }" />
      </template>
      <KeySelect @current-key="setCredential" />
    </bk-form-item>
    <bk-form-item v-if="props.directoryShow" property="tempDir" :required="props.directoryShow">
      <template #label>
        {{ $t('临时目录') }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: $t('用于客户端拉取文件型配置后的临时存储目录'),
            placement: 'top',
          }" />
      </template>
      <bk-input v-model="formData.tempDir" :placeholder="$t('请输入')" clearable />
    </bk-form-item>
  </bk-form>
  <!-- 添加标签1 -->
  <AddLabel
    @send-label="
      (obj) => {
        formData.labelArr = obj;
      }
    " />
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import KeySelect from './key-selector.vue';
  import { Info } from 'bkui-vue/lib/icon';
  import AddLabel from './add-label.vue';
  import { IExampleFormData } from '../../../../../../types/client';
  import { useI18n } from 'vue-i18n';

  const props = defineProps({
    directoryShow: {
      type: Boolean,
      default: true,
    },
  });

  const emits = defineEmits(['update-option-data']);

  const { t } = useI18n();
  const sysDirectories: string[] = [
    '/bin/',
    '/boot/',
    '/dev/',
    '/etc/',
    '/lib/',
    '/lib64/',
    '/proc/',
    '/run/',
    '/sbin/',
    '/sys/',
    '/tmp/',
    '/usr/',
    '/var/',
  ];
  const rules = {
    clientKey: [
      {
        required: true,
        message: t('请先选择客户端密钥，替换下方示例代码后，再尝试复制示例'),
        validator: (value: string) => value.length,
      },
    ],
    tempDir: [
      {
        required: props.directoryShow,
        message: t('请输入路径地址，替换下方示例代码后，再尝试复制示例'),
        validator: (value: string) => value.length,
        trigger: 'change',
      },
      {
        required: props.directoryShow,
        validator: (value: string) => {
          // 必须为绝对路径, 且不能以/结尾
          if (!value.startsWith('/') || value.endsWith('/')) {
            return false;
          }
          const parts = value.split('/').slice(1);
          let isValid = true;
          // 文件路径校验
          parts.some((part) => {
            if (part.startsWith('.') || !/^[\u4e00-\u9fa5A-Za-z0-9.\-_#%,@^+=\\[\]{}]+$/.test(part)) {
              isValid = false;
              return true;
            }
            return false;
          });
          return isValid;
        },
        trigger: 'change',
        message: t('无效的路径,路径不符合Unix文件路径格式规范'),
      },
      {
        required: props.directoryShow,
        message: t('禁止使用系统目录'),
        validator: (value: string) => !sysDirectories.some((dir) => value.startsWith(dir)),
        trigger: 'change',
      },
    ],
  };

  const formRef = ref();
  const formData = ref<IExampleFormData>({
    clientKey: '', // 客户端密钥
    privacyCredential: '', // 脱敏的密钥
    tempDir: '/data/bscp', // 临时目录
    labelArr: [], // 添加的标签
  });

  watch(formData.value, () => {
    sendAll();
  });

  const setCredential = (key: string, privacyKey: string) => {
    if (key.length && privacyKey.length) {
      formRef.value.clearValidate();
    }
    formData.value.clientKey = key;
    formData.value.privacyCredential = privacyKey;
  };
  const sendAll = () => {
    if (!props.directoryShow) {
      // 不显示临时目录的菜单，删除对应值
      delete formData.value.tempDir;
    }
    emits('update-option-data', formData.value);
  };

  defineExpose({
    formRef,
  });
</script>

<style scoped lang="scss">
  .headline {
    font-size: 14px;
    font-weight: 700;
    line-height: 22px;
    color: #63656e;
  }
  .form-example-wrap {
    margin-top: 16px;
    width: 537px;
    :deep(.bk-form-label) {
      font-size: 12px;
      & > span {
        position: relative;
      }
    }
  }
  .icon-info {
    position: absolute;
    right: -33px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 14px;
    color: #979ba5;
    cursor: pointer;
  }
</style>
