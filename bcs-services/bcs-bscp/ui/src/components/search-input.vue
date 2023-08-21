<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Search } from 'bkui-vue/lib/icon';

  const props = withDefaults(defineProps<{
    keyword?: string;
    placeholder?: string;
    width?: number;
  }>(), {
    keyword: '',
    placeholder: '请输入',
    width: 320
  })

  const emits = defineEmits(['update:keyword', 'search'])

  const inputVal = ref('')

  watch(() => props.keyword, val => {
    inputVal.value = val
  })

  const handleInputChange = () => {
    if (inputVal.value === '') {
      triggerSearch()
    }
  }

  const triggerSearch = () => {
    emits('update:keyword', inputVal.value)
    emits('search', inputVal.value)
  }

</script>
<template>
  <div class="search-input" :style="`width: ${props.width}px`">
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
