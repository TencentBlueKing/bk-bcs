<template>
  <FixedButton drag position="fixed" class="bottom-[20px] left-[85%]">
    <PopoverSelector trigger="click" offset="0, 6" placement="top">
      <div class="flex items-center justify-center w-[116px] h-[32px] cursor-pointer">
        <i class="bcs-icon bcs-icon-terminal mr5 text-[14px]"></i>
        <div>WebConsole</div>
        <!-- <i class="bcs-icon bcs-icon-helper ml5 cursor-default text-[14px]" @click="handleOpenLink"></i> -->
      </div>
      <template #content>
        <ClusterSelectPopover
          :selectable="false"
          :update-store="false"
          cluster-type="all"
          @click="(clusterID) => handleGotoConsole({ clusterID })" />
      </template>
    </PopoverSelector>
  </FixedButton>
</template>
<script lang="ts">
import { defineComponent } from 'vue';

import PopoverSelector from '../../components/popover-selector.vue';
import FixedButton from '../resource-view/resource-update/fixed-button.vue';

import ClusterSelectPopover from '@/components/cluster-selector/cluster-select-popover.vue';
import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'BcsTerminal',
  components: { FixedButton, PopoverSelector, ClusterSelectPopover },
  setup() {
    const { handleGotoConsole } = useCluster();
    const handleOpenLink = () => {
      window.open('https://bk.tencent.com/docs/document/7.0/173/14130');
    };

    return {
      handleOpenLink,
      handleGotoConsole,
    };
  },
});
</script>
