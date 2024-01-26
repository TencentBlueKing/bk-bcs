<template>
  <section class="rule-view">
    <p class="title">
      {{ t('共有') }}
      <span :class="['num', { zero: props.rules.length === 0 }]">{{ props.rules.length }}</span>
      {{ t('项关联规则') }}
    </p>
    <div v-if="props.rules.length > 0" class="rule-list">
      <div v-for="rule in rules" :key="rule.id" class="rule-item">{{ rule.spec.app + rule.spec.scope }}</div>
    </div>
    <bk-exception v-else scene="part" type="empty">
      <p class="empty-tips">{{ t('暂未设置关联规则') }}</p>
      <bk-button class="edit-rule-btn" text theme="primary" size="small" @click="emits('edit')">
        {{ t('编辑规则') }}
      </bk-button>
    </bk-exception>
  </section>
</template>

<script setup lang="ts">
  import { ICredentialRule } from '../../../../../types/credential';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();
  const props = defineProps<{
    rules: ICredentialRule[];
  }>();

  const emits = defineEmits(['edit']);
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
