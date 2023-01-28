<script setup lang="ts">
  import { ref } from 'vue'
  import LayoutTopBar from "../../../components/layout-top-bar.vue";
  import DetailHeader from "./components/detail-header.vue"

  const bkBizId = ref(2) // 缺少服务按照业务分组接口，暂时写死一个id调试
</script>
<template>
  <div class="serving-detail-page">
    <LayoutTopBar>
      <template #head>
        <detail-header :bk-biz-id="bkBizId"></detail-header>
      </template>
      <router-view :bk-biz-id="bkBizId"></router-view>
    </LayoutTopBar>
  </div>
</template>