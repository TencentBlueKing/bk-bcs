<template>
  <section class="rule-view">
    <p class="title">
      {{ t('共有') }}
      <span :class="['num', { zero: props.rules.length === 0 }]">{{ props.rules.length }}</span>
      {{ t('项关联规则') }}
    </p>
    <div v-if="props.rules.length > 0" class="rule-list">
      <div
        v-for="(rule, index) in rules"
        :key="rule.id"
        :class="['rule-item', { 'current-rule-item': previewRule?.id === rule.id }]"
        @click="handlePreviewRule(rule, index)">
        {{ rule.spec.app + rule.spec.scope }}
        <div v-if="previewRule?.id === rule.id" class="arrow-icon"></div>
      </div>
    </div>
    <bk-exception v-else scene="part" type="empty">
      <p class="empty-tips">{{ t('暂未设置关联规则') }}</p>
      <bk-button
        v-if="!props.hasManagePerm"
        class="edit-rule-btn"
        text
        theme="primary"
        size="small"
        @click="emits('edit')">
        {{ t('编辑规则') }}
      </bk-button>
    </bk-exception>
  </section>
</template>

<script setup lang="ts">
  import { ICredentialRule, IPreviewRule } from '../../../../../types/credential';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();
  const props = defineProps<{
    rules: ICredentialRule[];
    previewRule: IPreviewRule | null;
    hasManagePerm: boolean;
  }>();

  const emits = defineEmits(['edit', 'update:previewRule']);

  const handlePreviewRule = (rule: ICredentialRule, index: number) => {
    const previewRule = {
      id: rule.id,
      appName: rule.spec.app,
      scopeContent: rule.spec.scope,
      index,
    };
    emits('update:previewRule', previewRule);
  };
</script>
<style lang="scss" scoped>
  .title {
    margin: 0 0 16px;
    line-height: 20px;
    font-size: 12px;
    color: #63656e;
    .num {
      font-weight: 700;
      color: #3a84ff;
      &.zero {
        color: #63656e;
      }
    }
  }
  .rule-list {
    .rule-item {
      padding: 10px 16px;
      line-height: 20px;
      font-size: 12px;
      background: #f5f7fa;
      color: #63656e;
      border-radius: 2px;
      &:not(:last-child) {
        margin-bottom: 8px;
      }
    }
    .current-rule-item {
      position: relative;
      background: #e1ecff;
      border: 1px solid #699df4;
      border-radius: 2px;
      color: #3a84ff;
      & .arrow-icon {
        position: absolute;
        font-size: 14px;
        right: -8px;
        top: 50%;
        transform: translateY(-50%);
        color: #699df4;
        width: 0;
        height: 0;
        border-top: 8px solid transparent;
        border-bottom: 8px solid transparent;
        border-left: 8px solid #699df4;
      }
    }
  }
  .empty-tips {
    margin: 0 0 8px;
    line-height: 20px;
    color: #63656e;
  }
  .edit-rule-btn {
    font-size: 12px;
  }
</style>
