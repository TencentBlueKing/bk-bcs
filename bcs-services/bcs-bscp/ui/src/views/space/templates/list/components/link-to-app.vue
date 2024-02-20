<template>
  <Share class="share-icon" @click="handleClick" />
</template>
<script lang="ts" setup>
  import { useRouter } from 'vue-router';
  import { Share } from 'bkui-vue/lib/icon';

  const router = useRouter();

  const props = defineProps<{
    id: number;
    autoJump?: boolean;
  }>();

  const emits = defineEmits(['custom-click']);

  const handleClick = () => {
    if (props.autoJump) {
      const { href } = router.resolve({ name: 'service-config', params: { appId: props.id } });
      window.open(href, '_blank');
    } else {
      emits('custom-click');
    }
  };
</script>
<style lang="scss" scoped>
  .share-icon {
    font-size: 12px;
    color: #3a84ff;
    cursor: pointer;
  }
</style>
