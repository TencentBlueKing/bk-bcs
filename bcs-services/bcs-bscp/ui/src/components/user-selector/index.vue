<template>
  <BkMemberSelector v-bind="$attrs" :class="{ 'input-error': isError }">
    <template v-for="(obj, name) in $slots" #[name]="data">
      <slot :name="name" v-bind="data" />
    </template>
  </BkMemberSelector>
</template>
<script setup lang="ts">
  import BkMemberSelector from './user-selector-origin/index';

  withDefaults(
    defineProps<{
      isError: boolean;
    }>(),
    {},
  );
</script>
<style lang="scss" scoped>
  .input-error {
    :deep(.user-selector-layout) {
      .user-selector-container {
        transition: all 0.3s;
        border-color: #ea3636;
      }
    }
  }
</style>
