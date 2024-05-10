<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div
      ref="containerRef"
      :class="{ fullscreen: isOpenFullScreen }"
      @mouseenter="isMouseEnter = true"
      @mouseleave="isMouseEnter = false">
      <Card :title="$t(`按 {n} 统计`, { n: primaryDimension })" :height="368">
        <template #operation>
          <OperationBtn
            v-show="isShowOperationBtn"
            :need-down="true"
            :is-open-full-screen="isOpenFullScreen"
            :all-label="allLabel"
            :primary-dimension="primaryDimension"
            @refresh="emits('refresh')"
            @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen"
            @toggle-show="isOpenPopover = $event" />
        </template>
        <template #head-suffix>
          <bk-tag theme="info" type="stroke" style="margin-left: 8px"> {{ $t('标签') }} </bk-tag>
          <TriggerBtn v-model:currentType="currentType" style="margin-left: 8px" />
        </template>
        <bk-loading class="loading-wrap" :loading="loading">
          <component
            :bk-biz-id="bkBizId"
            :app-id="appId"
            :is="currentComponent"
            :data="data"
            @jump="jumpToSearch($event as string)" />
        </bk-loading>
      </Card>
    </div>
  </Teleport>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed, watch } from 'vue';
  import Card from '../../../components/card.vue';
  import TriggerBtn from '../../../components/trigger-btn.vue';
  import Pie from './pie.vue';
  import Column from './column.vue';
  import Table from './table.vue';
  import OperationBtn from '../../../components/operation-btn.vue';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  import { useRouter } from 'vue-router';

  const router = useRouter();

  const emits = defineEmits(['refresh']);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    primaryDimension: string;
    data: IClientLabelItem[];
    loading: boolean;
    allLabel: string[];
  }>();

  const currentType = ref('column');
  const componentMap = {
    pie: Pie,
    column: Column,
    table: Table,
  };
  const isOpenFullScreen = ref(false);
  const containerRef = ref();
  const initialWidth = ref(0);
  const isMouseEnter = ref(false);
  const isOpenPopover = ref(false);
  const isShowOperationBtn = computed(() => isMouseEnter.value || isOpenPopover.value);

  const currentComponent = computed(() => componentMap[currentType.value as keyof typeof componentMap]);

  onMounted(() => {
    initialWidth.value = containerRef.value.offsetWidth;
  });

  watch(
    () => isOpenFullScreen.value,
    (val) => {
      containerRef.value!.style.width = val ? '100%' : `${initialWidth.value}px`;
    },
  );

  const jumpToSearch = (value: string) => {
    const routeData = router.resolve({
      name: 'client-search',
      params: { appId: props.appId, bizId: props.bkBizId },
      query: { label: `${props.primaryDimension}=${value}` },
    });
    window.open(routeData.href, '_blank');
  };
</script>

<style scoped lang="scss">
  .fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 5000;
    .card {
      width: 100%;
      height: 100vh !important;
      :deep(.operation-btn) {
        top: 0 !important;
      }
    }
  }
  .loading-wrap {
    height: 100%;
  }
</style>
