<template>
  <div class="bcs-border text-[12px] w-[280px]">
    <div class="h-[32px] bg-[#F5F7FA] px-[16px] leading-[32px]">
      {{ $t('cluster.create.label.highAvailabilityScheme.text') }}
    </div>
    <div class="min-h-[42px] px-[16px] flex items-center">
      <span v-if="enabled">
        bcs-apiserver-proxy
        <i
          class="bk-icon icon-info-circle ml-[5px] text-[16px] text-[#979ba5]"
          v-bk-tooltips="$t('cluster.create.label.highAvailabilityScheme.bcsApiServerProxy')"></i>
      </span>
      <span class="text-[#979BA5] leading-[20px]" v-else>
        {{ $t('cluster.create.label.highAvailabilityScheme.notSupport') }}
      </span>
    </div>
  </div>
</template>
<script setup lang="ts">
defineProps({
  enabled: {
    type: Boolean,
    default: false,
  },
});
</script>
