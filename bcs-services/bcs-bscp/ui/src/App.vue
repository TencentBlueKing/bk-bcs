<script setup lang="ts">
import Header from "./components/head.vue";
</script>

<template>
  <div>
    <Header></Header>
    <div class="content">
      <router-view></router-view>
    </div>
  </div>
</template>


<style scoped>
.content {
  height: calc(100vh - 52px);
}
</style>
