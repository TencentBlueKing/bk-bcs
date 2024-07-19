<template>
  <div class="separator-wrap">
    <div class="title">{{ t('分隔符') }}</div>
    <div class="select">
      <div
        v-for="item in allSelect"
        :key="item.id"
        class="item"
        @click="selectSeparatorId = item.id"
        :class="{ 'select-item': selectSeparatorId === item.id }">
        {{ item.text }}
      </div>
    </div>
    <bk-form ref="formRef" v-if="selectSeparatorId === 4" :rules="rules" :model="formData">
      <bk-form-item required property="separator">
        <bk-input
          class="custom-input"
          v-model="formData.separator"
          :placeholder="t('请输入分隔符，限制为10个字符')" />
      </bk-form-item>
    </bk-form>
    <div class="footer">
      <bk-button theme="primary" size="small" @click="handleConfirm">{{ t('确定') }}</bk-button>
      <bk-button size="small" @click="emits('closed')">{{ t('取消') }}</bk-button>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();
  const emits = defineEmits(['closed', 'confirm']);
  const allSelect = computed(() => [
    { id: 0, text: t('空字符'), value: ' ' },
    { id: 1, text: ',', value: ',' },
    { id: 2, text: ';', value: ';' },
    { id: 3, text: '|', value: '|' },
    { id: 4, text: t('自定义') },
  ]);
  const selectSeparatorId = ref(0);
  const formData = ref({
    separator: '',
  });
  const formRef = ref();
  const regex = /^[\x20-\x7E]*$/;
  const rules = {
    separator: [
      {
        validator: (val: string) => regex.test(val),
        message: t('您输入的分隔符错误,请重新输入'),
      },
      {
        validator: (val: string) => val.length < 11,
        message: t('您输入的分隔符过长,请重新输入'),
      },
    ],
  };

  const handleConfirm = async () => {
    if (selectSeparatorId.value === 4) {
      await formRef.value.validate();
      emits('confirm', formData.value.separator);
    } else {
      emits('confirm', allSelect.value[selectSeparatorId.value].value);
    }
    emits('closed');
  };
</script>

<style scoped lang="scss">
  .separator-wrap {
    min-width: 276px;
    background: #2e2e2e;
    border: 1px solid #63656e;
    box-shadow: 0 2px 6px 0 #0000001a;
    border-radius: 2px;
    padding: 16px;
    padding-bottom: 24px;
    font-size: 12px;
    .title {
      color: #c4c6cc;
      line-height: 20px;
    }
    .select {
      display: flex;
      margin: 12px 0 20px;
      .item {
        min-width: 48px;
        padding: 0 8px;
        height: 26px;
        border: 1px solid #8e8e8e;
        color: #e0e0e0;
        text-align: center;
        line-height: 26px;
        background-color: #2f2f2f;
        cursor: pointer;
        &:hover {
          color: #488eff;
          border-color: #488eff;
        }
        &:nth-child(1) {
          border-radius: 2px 0 0 2px;
        }
        &:nth-child(5) {
          border-radius: 0 2px 2px 0;
        }
      }

      .select-item {
        color: #488eff;
        border-color: #488eff;
      }
    }
    :deep(.custom-input) {
      .bk-input--text {
        background-color: rgba($color: #000000, $alpha: 0);
        color: #c4c6cc;
      }
    }
    :deep(.bk-form-content) {
      margin-left: 0 !important;
    }
    .footer {
      display: flex;
      justify-content: flex-end;
    }
  }
</style>

<style lang="scss"></style>
