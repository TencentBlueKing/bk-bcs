<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Search } from 'bkui-vue/lib/icon';

  const props = withDefaults(defineProps<{
    modelValue: string;
    placeholder?: string;
    width?: number;
  }>(), {
    placeholder: '请输入'
  })

  const emits = defineEmits(['update:modelValue', 'search'])

  const inputVal = ref('')

  watch(() => props.modelValue, val => {
    inputVal.value = val
  })

  const handleInputChange = () => {
    emits('update:modelValue', inputVal.value)
    if (inputVal.value === '') {
      triggerSearch()
    }
  }

  const triggerSearch = () => {
    emits('update:modelValue', inputVal.value)
    emits('search', inputVal.value)
  }

</script>
<template>
  <div class="search-input" :style="`width: ${props.width ? props.width + 'px' : '100%'}`">
    <bk-input
      v-model="inputVal"
      :placeholder="props.placeholder"
      :clearable="true"
      @enter="triggerSearch"
      @change="handleInputChange">
        <template #suffix>
          <Search class="search-input-icon" />
        </template>
    </bk-input>
  </div>
</template>
<style lang="scss" scoped>
  .search-input-icon {
    padding-right: 10px;
    font-size: 16px;
    color: #979ba5;
    background: #ffffff;
  }
</style>
