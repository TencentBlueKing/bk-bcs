<template>
  <section class="search-option">
    <bk-checkbox v-model="value"> 仅看上线操作 </bk-checkbox>
    <bk-checkbox v-model="value"> 仅看失败操作 </bk-checkbox>
    <div class="search-input__wrap"></div>
  </section>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  const value = ref(false);
</script>

<style scoped></style>
