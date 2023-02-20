<script setup lang="ts">
import { onMounted } from 'vue'
import { useStore } from 'vuex'

import Header from "./components/head.vue";

const store = useStore()

onMounted(() => {
  store.dispatch('getUserInfo')
});


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
