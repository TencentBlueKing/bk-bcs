<script setup lang="ts">
  import { computed } from 'vue';
  import { IGroupRuleItem } from '../../../../types/group';
  import { GROUP_RULE_OPS } from '../../../constants';

  const props = defineProps<{
    rule: IGroupRuleItem
  }>()

  const opName = computed(() => {
    const op = GROUP_RULE_OPS.find(item => item.id === props.rule.op)
    return op?.name
  })

</script>
<template>
  <span class="rule-tag">{{ `${props.rule.key} ${opName} ${props.rule.value}` }}</span>
</template>
<style lang="scss" scoped>
  .rule-tag {
    display: inline-block;
    padding: 0 10px;
    line-height: 22px;
    font-size: 12px;
    background: #f0f1f5;
    border-radius: 2px;
  }
</style>