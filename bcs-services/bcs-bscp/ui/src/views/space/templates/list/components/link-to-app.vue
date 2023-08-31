<script lang="ts" setup>
  import { useRouter } from 'vue-router'
  import { Share } from 'bkui-vue/lib/icon';

  const router = useRouter()

  const props = defineProps<{
    id: number;
  }>()

  const goToApp = () => {
    const { href } = router.resolve({ name: 'service-config', params: { appId: props.id } })
    window.open(href, '_blank')
  }

</script>
<template>
  <Share @click="goToApp" />
</template>
<style lang="scss" scoped>
</style>
