<script setup lang="ts">
  import { ref } from 'vue'
  import { RightShape, Del } from 'bkui-vue/lib/icon'
  import { ICategoryGroup } from '../../../../../types/group'

  const props = defineProps<{
    categoryGroup: ICategoryGroup
  }>()

  const folded = ref(true)

  const handleToggleFold = () => {
    folded.value = !folded.value
  }

</script>
<template>
  <section class="category-group">
    <div :class="['header-area', { 'expanded': !folded }]" @click="handleToggleFold">
      <div class="category-content">
        <RightShape class="arrow-icon" />
        <span class="name">{{ categoryGroup.config.spec.name }}</span>
      </div>
      <Del class="delete-icon" />
    </div>
    <bk-table v-if="!folded" class="group-table" :border="['outer']">
      <bk-table-column label="分组名称"></bk-table-column>
      <bk-table-column label="分组规则"></bk-table-column>
      <bk-table-column label="当前上线版本"></bk-table-column>
      <bk-table-column label="操作"></bk-table-column>
    </bk-table>
  </section>
</template>
<style lang="scss" scoped>
  .header-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 18px 8px 8px;
    background: #dcdee5;
    border-radius: 2px;
    cursor: pointer;
    &.expanded {
      .arrow-icon {
        transform: rotate(90deg);
      }
    }
    .category-content {
      display: flex;
      align-items: center;
    }
    .arrow-icon {
      display: inline-block;
      font-size: 12px;
      color: #63656e;
    }
    .name {
      margin-left: 9px;
      color: #313238;
      font-size: 12px;
      line-height: 16px;
    }
    .delete-icon {
      font-size: 13px;
      color: #979ba5;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
  .group-table {
    margin-top: 8px
  }
</style>