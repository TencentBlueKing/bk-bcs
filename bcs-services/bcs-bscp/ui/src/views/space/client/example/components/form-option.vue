<template>
  <div class="headline">示例参数</div>
  <bk-form class="form-example-wrap" :model="formData" :rules="rules" form-type="vertical" ref="formRef">
    <bk-form-item property="clientKey" required>
      <template #label>
        {{ $t('客户端密钥') }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: '用于客户端拉取配置时身份验证',
            placement: 'top',
          }" />
      </template>
      <KeySelect @current-key="setCredential" />
    </bk-form-item>
    <bk-form-item v-if="props.contentsShow" property="tempDir" :required="props.contentsShow">
      <template #label>
        {{ $t('临时目录') }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: '用于客户端拉取文件型配置后的临时存储目录',
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

  const props = defineProps({
    contentsShow: {
      type: Boolean,
      default: true,
    },
  });

  const emits = defineEmits(['option-data']);

  const rules = {
    clientKey: [
      {
        required: true,
        message: '请先选择客户端密钥，替换下方示例代码后，再尝试复制示例',
        validator: (value: string) => value.length,
        trigger: 'change',
      },
    ],
    tempDir: [
      {
        required: props.contentsShow,
        message: '请输入路径地址，替换下方示例代码后，再尝试复制示例',
        validator: (value: string) => value.length,
        trigger: 'change',
      },
    ],
  };

  const formRef = ref('');
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
    formData.value.clientKey = key;
    formData.value.privacyCredential = privacyKey;
  };
  const sendAll = () => {
    if (!props.contentsShow) {
      // 不显示临时目录的菜单，删除对应值
      delete formData.value.tempDir;
    }
    emits('option-data', formData.value);
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
    cursor: pointer;
  }
</style>
