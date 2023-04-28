<script setup lang="ts">
  import { watch } from 'vue'
  import { useRoute } from 'vue-router'
  import { storeToRefs } from 'pinia';
  import { useGlobalStore } from '../../store/global';
  import applySpacePerm from './apply-space-perm.vue'

  const { spaceId, showSpacePermApply } = storeToRefs(useGlobalStore())

  const route = useRoute()

  watch(() => route.params.spaceId, (val) => {
    showSpacePermApply.value = false
    spaceId.value = <string>val
  }, { immediate: true })

</script>
<template>
  <apply-space-perm v-if="showSpacePermApply"></apply-space-perm>
  <router-view v-else></router-view>
</template>
