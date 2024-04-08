<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div :class="{ fullscreen: isOpenFullScreen }">
      <Card title="按 Cluster_id 统计" :height="368">
        <template #operation>
          <OperationBtn
            :is-open-full-screen="isOpenFullScreen"
            @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen" />
        </template>
        <bk-exception class="exception-wrap-item exception-part" type="empty" scene="part" description="暂无数据">
          <template #type>
            <span class="bk-bscp-icon icon-bar-chart exception-icon" />
          </template>
        </bk-exception>
      </Card>
    </div>
  </Teleport>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import Card from '../../components/card.vue';
  import OperationBtn from '../../components/operation-btn.vue';

  const isOpenFullScreen = ref(false);
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
  .fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 5000;
    background-color: rgba(0, 0, 0, 0.6);
    .card {
      position: absolute;
      width: 100%;
      height: 80vh !important;
      top: 50%;
      transform: translateY(-50%);
      .loading-wrap {
        height: 100%;
      }
    }
  }
</style>
