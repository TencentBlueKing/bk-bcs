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
            @refresh="loadChartData"
            @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen"
            @toggle-show="isOpenPopover = $event"
            @select-dimension="selectedDimension = $event" />
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
            :label="primaryDimension"
            :data="data"
            @jump="jumpToSearch($event as string)" />
        </bk-loading>
      </Card>
    </div>
  </Teleport>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed, watch } from 'vue';
  import { getClientLabelData } from '../../../../../../../api/client';
  import { storeToRefs } from 'pinia';
  import useClientStore from '../../../../../../../store/client';
  import Card from '../../../components/card.vue';
  import TriggerBtn from '../../../components/trigger-btn.vue';
  import Pie from './pie.vue';
  import Column from './column.vue';
  import Table from './table.vue';
  import OperationBtn from '../../../components/operation-btn.vue';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  import { useRouter } from 'vue-router';

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const router = useRouter();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    primaryDimension: string;
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
  const loading = ref(false);
  const data = ref<IClientLabelItem[]>([]);
  const selectedDimension = ref<string[]>([]);

  const isShowOperationBtn = computed(() => isMouseEnter.value || isOpenPopover.value);

  const currentComponent = computed(() => componentMap[currentType.value as keyof typeof componentMap]);

  onMounted(() => {
    initialWidth.value = containerRef.value.offsetWidth;
    loadChartData();
  });

  watch(
    () => isOpenFullScreen.value,
    (val) => {
      containerRef.value!.style.width = val ? '100%' : `${initialWidth.value}px`;
    },
  );

  const loadChartData = async () => {
    const params = {
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: searchQuery.value.search,
      primary_key: props.primaryDimension,
    };
    const allDimension: { [key: string]: string } = {};
    selectedDimension.value.forEach((item: string) => {
      allDimension[item] = '';
    });
    params.search = Object.assign({}, searchQuery.value.search, {
      label: {
        ...searchQuery.value.search.label,
        ...allDimension,
      },
    });
    try {
      loading.value = true;
      const res = await getClientLabelData(props.bkBizId, props.appId, params);
      data.value = res[props.primaryDimension];
      // allLabelData.value = res;
      // selectedChart.value = [];
      // if (Object.keys(res).length) {
      //   selectedLabel.value.forEach((item) => {
      //     const data = allLabelData.value?.[item];
      //     if (data) {
      //       selectedChart.value.push({
      //         label: item,
      //         data,
      //       });
      //     }
      //   });
      // }
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

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
