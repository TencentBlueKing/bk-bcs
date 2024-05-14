<template>
  <div class="search-input" :style="`width: ${props.width ? props.width + 'px' : '100%'}`">
    <bk-input
      v-model="inputVal"
      :placeholder="props.placeholder"
      :clearable="true"
      @clear="triggerSearch"
      @input="triggerSearch">
      <template #suffix>
        <Search class="search-input-icon" />
      </template>
    </bk-input>
  </div>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Search } from 'bkui-vue/lib/icon';
  import { debounce } from 'lodash';
  import { localT } from '../i18n';

  const props = withDefaults(
    defineProps<{
      modelValue: string;
      placeholder?: string;
      width?: number;
    }>(),
    {
      placeholder: localT('请输入'),
    },
  );

  const emits = defineEmits(['update:modelValue', 'search']);

  const inputVal = ref('');

  watch(
    () => props.modelValue,
    (val) => {
      inputVal.value = val;
    },
  );

  const triggerSearch = debounce(() => {
    emits('update:modelValue', inputVal.value);
    emits('search', inputVal.value);
  }, 300);
</script>
<style lang="scss" scoped>
  .search-input-icon {
    padding-right: 10px;
    font-size: 16px;
    color: #979ba5;
    background: #ffffff;
  }
</style>
