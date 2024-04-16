<template>
  <div
    class="flex items-center justify-center relative text-[12px] h-[40px] bg-[#FAFBFD] cursor-pointer mx-[-16px]"
    @click="openLink">
    <i class="bcs-icon bcs-icon-fenxiang mr-[4px] !text-[12px]"></i>
    {{ linkText }}
    <bcs-divider direction="vertical" class="absolute right-[30px]"></bcs-divider>
    <span
      class="absolute right-0 h-[32px] w-[40px] flex items-center justify-center"
      v-bk-tooltips="$t('generic.button.refresh')"
      @click.stop="refresh">
      <i class="bcs-icon bcs-icon-reset !text-[12px] text-[#979BA5]"></i>
    </span>
  </div>
</template>
<script setup lang="ts">
const props = defineProps({
  linkText: {
    type: String,
  },
  link: {
    type: String,
  },
});

const emits = defineEmits(['link', 'refresh']);

const openLink = () => {
  emits('link');
  if (!props.link) return;
  window.open(props.link);
};

const refresh = () => {
  emits('refresh');
};
</script>
