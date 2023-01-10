<script setup lang="ts">
  import { ref, Ref } from 'vue'

  type IServingItem = {
    id: number
  }
  const servingList = ref([]) as Ref<IServingItem[]>
</script>
<template>
  <bk-select>
    <bk-option v-for="item in servingList" :key="item.id"></bk-option>
  </bk-select>
</template>
