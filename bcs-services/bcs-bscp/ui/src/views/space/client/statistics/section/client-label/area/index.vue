<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div ref="containerRef" :class="{ fullscreen: isOpenFullScreen }">
      <Card :title="`按 ${label} 统计`" :height="368">
        <template #operation>
          <OperationBtn
            :is-open-full-screen="isOpenFullScreen"
            @refresh="loadChartData"
            @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen" />
        </template>
        <template #head-suffix>
          <bk-tag theme="info" type="stroke" style="margin-left: 8px"> 标签 </bk-tag>
          <TriggerBtn v-model:currentType="currentType" style="margin-left: 8px" />
        </template>
        <bk-loading class="loading-wrap" :loading="loading">
          <component
            v-if="selectedLabelData?.length"
            :bk-biz-id="bkBizId"
            :app-id="appId"
            :is="currentComponent"
            :data="selectedLabelData"
            :label="label"
            @jump="jumpToSearch" />
          <bk-exception
            v-else
            class="exception-wrap-item exception-part"
            type="empty"
            scene="part"
            description="暂无数据">
            <template #type>
              <span class="bk-bscp-icon icon-bar-chart exception-icon" />
            </template>
          </bk-exception>
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
  import { IClientLabelItem, IClinetCommonQuery } from '../../../../../../../../types/client';
  import { getClientLabelData } from '../../../../../../../api/client';
  import useClientStore from '../../../../../../../store/client';
  import { storeToRefs } from 'pinia';
  import { useRouter } from 'vue-router';

  const router = useRouter();

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    selectedLabel: string;
  }>();

  const currentType = ref('column');
  const componentMap = {
    pie: Pie,
    column: Column,
    table: Table,
  };
  const allLabeldata = ref<{ [key: string]: IClientLabelItem[] }>();
  const loading = ref(false);
  const isOpenFullScreen = ref(false);
  const containerRef = ref();
  const initialWidth = ref(0);

  const currentComponent = computed(() => componentMap[currentType.value as keyof typeof componentMap]);
  const selectedLabelData = computed(() => allLabeldata.value?.[props.selectedLabel]);
  const label = computed(() => props.selectedLabel.charAt(0).toUpperCase() + props.selectedLabel.slice(1));

  onMounted(() => {
    loadChartData();
    initialWidth.value = containerRef.value.offsetWidth;
  });

  watch(
    () => props.appId,
    () => {
      loadChartData();
    },
  );

  watch(
    () => searchQuery.value,
    () => {
      loadChartData();
    },
    { deep: true },
  );

  watch(
    () => isOpenFullScreen.value,
    (val) => {
      containerRef.value!.style.width = val ? '100%' : `${initialWidth.value}px`;
    },
  );

  const loadChartData = async () => {
    const params: IClinetCommonQuery = {
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: searchQuery.value.search,
    };
    try {
      loading.value = true;
      const res = await getClientLabelData(props.bkBizId, props.appId, params);
      allLabeldata.value = res;
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const jumpToSearch = () => {
    router.push({
      name: 'client-search',
      params: { appId: props.appId, bizId: props.bkBizId },
    });
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
  .loading-wrap {
    height: 100%;
  }
</style>
