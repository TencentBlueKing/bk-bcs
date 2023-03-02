<template>
  <FixedButton drag position="fixed" class="bottom-[20px] left-[90%]">
    <PopoverSelector trigger="mouseenter" offset="0, 6" placement="top">
      <div class="flex items-center h-[32px]">
        <i class="bcs-icon bcs-icon-terminal mr5 cursor-default text-[14px]"></i>
        <div class="cursor-default">WebConsole</div>
        <i class="bcs-icon bcs-icon-helper ml5 cursor-default text-[14px]" @click="handleOpenLink"></i>
      </div>
      <template #content>
        <bcs-input
          behavior="simplicity"
          clearable
          right-icon="bk-icon icon-search"
          v-model="searchValue"
          :placeholder="$t('输入集群名或ID搜索')">
        </bcs-input>
        <bcs-exception
          class="w-[260px]"
          type="empty"
          scene="part"
          v-if="!clusterList.length">
        </bcs-exception>
        <ul class="bg-[#fff] max-h-[260px] w-[260px] overflow-auto text-[12px] text-[#63656e] py-[6px]" v-else>
          <li
            v-for="item in clusterList"
            :key="item.clusterID"
            class="px-[16px] cursor-pointer hover:bg-[#eaf3ff] hover:text-[#3a84ff]"
            @click="handleGotoConsole(item)">
            <div class="flex flex-col justify-center h-[46px]">
              <span class="bcs-ellipsis">{{ item.clusterName }}</span>
              <span class="mt-[5px]">{{ item.clusterID }}</span>
            </div>
          </li>
        </ul>
      </template>
    </PopoverSelector>
  </FixedButton>
</template>
<script lang="ts">
import { useCluster, useProject } from '@/common/use-app';
import { computed, defineComponent, ref } from '@vue/composition-api';
import FixedButton from '../dashboard/resource-update/fixed-button.vue';
import PopoverSelector from './popover-selector.vue';

export default defineComponent({
  name: 'BcsTerminal',
  components: { FixedButton, PopoverSelector },
  setup() {
    const { clusterList: clusterData } = useCluster();
    const searchValue = ref('');
    const clusterList = computed(() => clusterData.value
      .filter(item => !item.is_shared
      && (
        item.clusterName.includes(searchValue.value) || item.clusterID.includes(searchValue.value)
      )));
    const handleOpenLink = () => {
      window.open('https://bk.tencent.com/docs/document/7.0/173/14130');
    };

    const { projectID } = useProject();
    const terminalWins = ref<Window | null>(null);
    const handleGotoConsole = (cluster) => {
      const url = `${window.DEVOPS_BCS_API_URL}/web_console/projects/${projectID.value}/mgr/#cluster=${cluster.clusterID}`;

      // 缓存当前窗口，再次打开时重新进入
      if (terminalWins.value) {
        if (!terminalWins.value.closed) {
          terminalWins.value.postMessage({
            clusterId: cluster.clusterID,
            clusterName: cluster.name,
          }, location.origin);
          terminalWins.value.focus();
        } else {
          terminalWins.value = window.open(url, '');
        }
      } else {
        terminalWins.value = window.open(url, '');
      }
    };

    return {
      searchValue,
      clusterList,
      handleOpenLink,
      handleGotoConsole,
    };
  },
});
</script>
