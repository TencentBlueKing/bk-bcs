<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useStore } from 'vuex'
import isCrossOriginIFrame from './utils/isCrossOriginIFrame'

import Header from "./components/head.vue";

const store = useStore()

const showLoginModal = ref(store.state.showLoginModal)

watch(() => store.state.showLoginModal, () => {
  const topWindow = isCrossOriginIFrame() ? window : window.top
  // @ts-ignore
  topWindow.BLUEKING.corefunc.open_login_dialog(store.state.loginUrl)
})

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
