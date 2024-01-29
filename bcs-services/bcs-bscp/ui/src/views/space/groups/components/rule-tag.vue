<template>
  <span class="rule-tag">{{ `${props.rule.key} ${opName} ${valueText}` }}</span>
</template>
<script setup lang="ts">
  import { computed } from 'vue';
  import { IGroupRuleItem } from '../../../../../types/group';
  import GROUP_RULE_OPS from '../../../../constants/group';

  const props = defineProps<{
    rule: IGroupRuleItem;
  }>();

  const opName = computed(() => {
    const op = GROUP_RULE_OPS.find((item) => item.id === props.rule.op);
    return op?.name;
  });

  const valueText = computed(() => {
    if (['in', 'nin'].includes(props.rule.op)) {
      return `(${(props.rule.value as string[]).join(', ')})`;
    }
    return props.rule.value;
  });
</script>
<style lang="scss" scoped>
  .rule-tag {
    display: inline-block;
    line-height: 22px;
    font-size: 12px;
  }
</style>
