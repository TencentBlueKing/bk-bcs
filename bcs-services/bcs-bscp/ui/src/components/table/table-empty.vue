<template>
   <bk-exception
      v-if="!props.isSearchEmpty"
      class="exception-part"
      type="empty"
      scene="part"
      :title="emptyTitle || t('暂无数据')"
    />
    <bk-exception
      v-else
      class="exception-part"
      type="search-empty"
      scene="part"
      :title="t('搜索结果为空')"
    >
      <div>{{ t('可以尝试 调整关键词 或') }} <span class="clear-button" @click="emits('clear')">{{ t('清空筛选条件') }}</span></div>
    </bk-exception>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const props = withDefaults(
  defineProps<{
      isSearchEmpty: boolean;
      emptyTitle?: string;
    }>(),
  {
    isSearchEmpty: false,
  },
);
const emits = defineEmits(['clear']);
</script>

<style scoped lang="scss">
.clear-button {
  color: #3a84ff;
  cursor: pointer;
}
.exception-part {
  height: 280px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
