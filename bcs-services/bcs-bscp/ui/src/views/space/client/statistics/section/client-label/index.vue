<template>
  <div>
    <SectionTitle :title="'客户端标签/附加信息分布'" />
    <div class="content-wrap">
      <div class="area">
        <Area :bk-biz-id="bkBizId" :app-id="appId" :selected-label="selectedLabel" />
      </div>
      <div class="cluster">
        <Clustre :bk-biz-id="bkBizId" :app-id="appId" />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import SectionTitle from '../../components/section-title.vue';
  import Area from './area/index.vue';
  import Clustre from './clustre.vue';

  defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const selectedLabel = ref('area');
</script>

<style scoped lang="scss">
  .content-wrap {
    display: flex;
    .area,
    .cluster {
      flex-grow: 1;
      flex-basis: 0;
    }
    .area {
      margin-right: 16px;
    }
  }
</style>
