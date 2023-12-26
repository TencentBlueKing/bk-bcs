<template>
  <bk-popover
    ext-cls="default-group-rules-popover"
    theme="light"
    placement="top-start">
    <InfoLine class="default-group-tips-icon" />
    <template #content>
      <div class="title">除以下分组之外的所有实例</div>
      <div class="exclude-groups">
        <div v-for="excludeItem in props.excludedGroups" class="exclude-item">
          <span class="group-name">{{ excludeItem.name }}</span>
          <span v-if="excludeItem.rules && excludeItem.rules.length > 0" class="split-line">|</span>
          <div class="rules">
            <div v-for="(rule, index) in excludeItem.rules" :key="index">
              <template v-if="index > 0"> & </template>
              <rule-tag class="tag-item" :rule="rule" />
            </div>
          </div>
        </div>
      </div>
    </template>
  </bk-popover>
</template>
<script setup lang="ts">
  import { InfoLine } from 'bkui-vue/lib/icon';
  import { IGroupToPublish } from '../../../../../../../types/group';
  import RuleTag from '../../../../groups/components/rule-tag.vue';

  const props = defineProps<{
    excludedGroups: IGroupToPublish[],
  }>();
</script>
<style scoped lang="scss">
  .default-group-tips-icon {
    margin-left: 8px;
    font-size: 14px;
    color: #939ba5;
    cursor: pointer;
    &:hover {
      color: #3a84ff;
    }
  }
</style>
<style lang="scss">
  .default-group-rules-popover {
    .title {
      margin: 8px 0;
      color: #979ba5;
    }
    .exclude-groups {
      line-height: 16px;
      font-size: 12px;
      color: #979ba5;
      .exclude-item {
        display: flex;
        align-items: center;
      }
      .group-name {
        line-height: 20px;
        color: #63656e;
      }
      .split-line {
        margin: 0 4px 0 16px;
      }
      .rules {
        display: flex;
        align-items: center;
        line-height: 16px;
        white-space: nowrap;
      }
    }
  }
</style>
