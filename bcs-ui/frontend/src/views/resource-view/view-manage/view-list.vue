<template>
  <div>
    <div class="px-[8px]">
      <bcs-input
        behavior="simplicity"
        clearable
        right-icon="bk-icon icon-search"
        class="mb-[8px]"
        :placeholder="$t('view.placeholder.searchView')"
        v-model.trim="searchValue">
      </bcs-input>
    </div>
    <bcs-exception
      type="empty"
      scene="part"
      class="!w-[260px]"
      v-if="!filterViewList.length" />
    <div class="max-h-[360px] overflow-auto">
      <li
        v-for="item in filterViewList"
        :key="item.id"
        :class="[
          'bcs-dropdown-item group flex items-center justify-between',
          { active: item.id === dashboardViewID }
        ]"
        v-bk-trace.click="{
          ct: 'view',
          act: 'change',
          d: '视图切换操作',
        }"
        @click="changeView(item.id)">
        <span class="flex items-center flex-1">
          <span class="bcs-ellipsis" v-bk-overflow-tips="{ placement: 'right' }">{{ item.name }}</span>
          <i
            class="bcs-icon bcs-icon-alarm-insufficient text-[14px] text-[#FFB848] ml-[5px]"
            v-if="invalidateViewMap[item.id]"
            v-bk-tooltips="$t('view.tips.invalidate')">
          </i>
        </span>

        <span class="inline-flex items-center justify-center w-[32px] h-[32px]" @click.stop="editView(item.id)">
          <i
            :class="[
              'bk-icon icon-edit-line',
              'text-[#979BA5] opacity-0 group-hover:opacity-100 hover:!text-[#3a84ff]',
            ]">
          </i>
        </span>
      </li>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue';

import useViewConfig from './use-view-config';

import { useCluster } from '@/composables/use-app';

const emits = defineEmits(['change', 'edit']);
// 视图管理
const { viewList, dashboardViewID, getViewConfigList } = useViewConfig();
const { clusterNameMap } = useCluster();
const invalidateViewMap = computed(() => viewList.value.reduce((pre, item) => {
  const unknownCluster = item.clusterNamespaces?.some(item => !clusterNameMap.value[item.clusterID]);
  if (unknownCluster) {
    pre[item.id] = true;
  }
  return pre;
}, {}));
const searchValue = ref('');
const filterViewList = computed(() => viewList.value.filter(item => item?.name?.includes(searchValue.value)));

const changeView = (id = '') => {
  emits('change', id);
};

const editView = (id = '') => {
  emits('edit', id);
};

onBeforeMount(() => {
  getViewConfigList();
});
</script>
