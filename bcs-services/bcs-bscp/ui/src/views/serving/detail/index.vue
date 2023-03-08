<script setup lang="ts">
  import { onMounted, ref, watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { useStore } from 'vuex'
  import { getAppDetail } from '../../../api';
  import LayoutTopBar from "../../../components/layout-top-bar.vue";
  import DetailHeader from "./components/detail-header.vue"

  const route = useRoute()
  const store = useStore()

  const bkBizId = ref(String(route.params.spaceId))
  const appId = ref(Number(route.params.appId))
  const appDataLoading = ref(true)

  watch(() => route.params.appId, val => {
    appId.value = Number(val)
    bkBizId.value = String(route.params.spaceId)
    getAppData()
  })

  onMounted(() => {
    getAppData()
  })

  const getAppData = async() => {
    appDataLoading.value = true
    try {
      const res = await getAppDetail(bkBizId.value, appId.value)
      store.commit('config/setAppData', res)
      appDataLoading.value = false
    } catch (e) {
      console.error(e)
    }
  }

</script>
<template>
  <div class="serving-detail-page">
    <LayoutTopBar>
      <template #head>
        <detail-header :bk-biz-id="bkBizId" :app-id="appId"></detail-header>
      </template>
      <template v-if="!appDataLoading">
        <router-view :bk-biz-id="bkBizId" :app-id="appId"></router-view>
      </template>
    </LayoutTopBar>
  </div>
</template>