<template>
  <bk-tag class="commonly-item" @click="emits('click')">
    <span class="name">{{ commonlySearchItem.spec.search_name }}</span>
    <bk-dropdown :is-show="isShow" placement="bottom">
      <Ellipsis
        v-if="commonlySearchItem.spec.creator !== 'system'"
        class="action-more-icon"
        @mouseenter="isShow = true" />
      <template #content>
        <bk-dropdown-menu>
          <bk-dropdown-item @click="handleMenuClick('update')">{{ $t('重命名') }}</bk-dropdown-item>
          <bk-dropdown-item @click="handleMenuClick('delete')">{{ $t('删除') }}</bk-dropdown-item>
        </bk-dropdown-menu>
      </template>
    </bk-dropdown>
  </bk-tag>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { ICommonlyUsedItem } from '../../../../../types/client';
  import { Ellipsis } from 'bkui-vue/lib/icon';
  defineProps<{
    commonlySearchItem: ICommonlyUsedItem;
  }>();
  const emits = defineEmits(['update', 'delete', 'click']);
  const isShow = ref(false);
  const handleMenuClick = (item: string) => {
    item === 'update' ? emits('update') : emits('delete');
    isShow.value = false;
  };
</script>

<style scoped lang="scss">
  .commonly-item {
    margin-right: 8px;
    line-height: 22px;
    .action-more-icon {
      transform: rotate(90deg);
      color: #63656e;
      cursor: pointer;
    }
  }
</style>
