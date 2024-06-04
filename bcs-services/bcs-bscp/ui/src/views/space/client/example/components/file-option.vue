<template>
  <div class="headline">示例参数</div>
  <bk-form ref="formRef" class="form-example-wrap" :model="formData" :rules="rules" form-type="vertical">
    <bk-form-item property="clientKey" required>
      <template #label>
        {{ $t('客户端密钥') }}
        <bk-popover content="用于客户端拉取配置时身份验证" placement="top-center">
          <info width="14" height="14" class="icon-info" />
        </bk-popover>
      </template>
      <KeySelect @current-key="(val: string) => (formData.clientKey = val)" />
      <template #error>请先选择客户端密钥，替换下方示例代码后，再尝试复制示例</template>
    </bk-form-item>
    <bk-form-item property="tempContents" required>
      <template #label>
        {{ $t('临时目录') }}
        <bk-popover content="用于客户端拉取文件型配置后的临时存储目录" placement="top-center">
          <info width="14" height="14" class="icon-info" />
        </bk-popover>
      </template>
      <bk-input v-model="formData.tempContents" :placeholder="$t('请输入')" clearable />
      <template #error>请输入路径地址，替换下方示例代码后，再尝试复制示例</template>
    </bk-form-item>
  </bk-form>
  <!-- 添加标签1 -->
  <AddLabel />
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  // import BkForm, { BkFormItem } from 'bkui-vue/lib/form';
  // import BkInput from 'bkui-vue/lib/input';
  import KeySelect from './key-selector.vue';
  import { Info } from 'bkui-vue/lib/icon';
  import AddLabel from './add-label.vue';

  const formRef = ref('');
  const formData = ref({
    clientKey: '',
    tempContents: '',
  });
  const rules = {
    clientKey: [
      {
        validator: (value: string) => value.length,
        trigger: 'change',
      },
    ],
    tempContents: [
      {
        validator: (value: string) => value.length,
        trigger: 'change',
      },
    ],
  };
</script>

<style scoped lang="scss">
  .headline {
    font-size: 14px;
    font-weight: 700;
    line-height: 22px;
    color: #63656e;
  }
  .form-example-wrap {
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
    cursor: pointer;
  }
</style>
