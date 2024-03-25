<template>
  <div>
    <SectionTitle :title="'客户端组件信息统计'" />
    <div class="content-wrap">
      <div class="left">
        <VersionRelease />
      </div>
      <div class="right">
        <Card v-for="item in 6" :key="item" :title="'平均拉取耗时'" :width="207" :height="128" />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import SectionTitle from '../../components/section-title.vue';
  import VersionRelease from './version-release.vue';
  import Card from '../../components/card.vue';
</script>

<style scoped lang="scss">
  .content-wrap {
    display: flex;
    .left {
      flex: 1;
      margin-right: 16px;
    }
    .right {
      display: flex;
      justify-content: space-between;
      flex-wrap: wrap;
      width: 430px;
    }
  }
</style>
