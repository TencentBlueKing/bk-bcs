<script setup lang="ts">
  import { ref, reactive } from 'vue'

  const servingList = ref([])
</script>
<template>
  <bk-select>
    <bk-option v-for="item in servingList" :key="item.id"></bk-option>
  </bk-select>
</template>