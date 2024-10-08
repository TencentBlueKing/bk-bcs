<template>
  <div>
    <template v-if="clusterConnectSetting && clusterConnectSetting.isExtranet">
      <div class="flex h-[42px] max-w-[360px] bcs-border">
        <div class="flex items-center pl-[16px] w-[80px] bg-[#FAFBFD] bcs-border-right">
          {{ $t('tke.label.securityGroup.text') }}
        </div>
        <div class="flex items-center flex-1 px-[16px] bcs-ellipsis">
          <span class="bcs-ellipsis" v-bk-overflow-tips>{{ securityGroupName
            ?`${securityGroupName}(${clusterConnectSetting.securityGroup})`
            : clusterConnectSetting.securityGroup }}</span>
        </div>
      </div>
      <div class="flex h-[42px] max-w-[360px] bcs-border mt-[-1px]">
        <div class="flex items-center pl-[16px] w-[80px] bg-[#FAFBFD] bcs-border-right">
          {{ $t('tke.label.ipAndPort') }}
        </div>
        <div class="flex items-center flex-1 px-[16px] bcs-ellipsis">{{ apiServer || '--' }}</div>
      </div>
    </template>
    <span v-else>--</span>
  </div>
</template>
<script setup lang="ts">
defineProps({
  clusterConnectSetting: {
    type: Object,
    default: () => ({}),
  },
  securityGroupName: {
    type: String,
    default: '',
  },
  apiServer: {
    type: String,
    default: '',
  },
});
</script>
