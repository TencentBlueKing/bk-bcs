<template>
  <Card title="按 Cluster_id 统计" :height="368">
    <bk-exception class="exception-wrap-item exception-part" type="empty" scene="part" description="暂无数据" />
  </Card>
</template>

<script lang="ts" setup>
  import Card from '../../components/card.vue';
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
</style>
