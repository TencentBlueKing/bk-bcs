<script setup lang="ts">
  import LayoutTopBar from "../../../components/layout-top-bar.vue";
  import DetailHeader from "./components/detail-header.vue"
</script>
<template>
  <div class="serving-detail-page">
    <LayoutTopBar>
      <template #head>
        <detail-header></detail-header>
      </template>
      <router-view></router-view>
    </LayoutTopBar>
  </div>
</template>