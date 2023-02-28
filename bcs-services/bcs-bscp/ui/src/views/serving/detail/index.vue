<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import LayoutTopBar from "../../../components/layout-top-bar.vue";
  import DetailHeader from "./components/detail-header.vue"

  const route = useRoute()

  const bkBizId = ref(String(route.params.spaceId))
  const appId = ref(Number(route.params.appId))

  watch(() => route.params.appId, val => {
    appId.value = Number(val)
    bkBizId.value = String(route.params.spaceId)
  })

</script>
<template>
  <div class="serving-detail-page">
    <LayoutTopBar>
      <template #head>
        <detail-header :bk-biz-id="bkBizId" :app-id="appId"></detail-header>
      </template>
      <router-view :bk-biz-id="bkBizId" :app-id="appId"></router-view>
    </LayoutTopBar>
  </div>
</template>