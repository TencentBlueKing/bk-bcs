<template>
  <apply-perm-page v-if="showPermApplyPage"></apply-perm-page>
  <router-view v-else></router-view>
</template>
<script setup lang="ts">
import { watch } from 'vue';
import { useRoute } from 'vue-router';
import { storeToRefs } from 'pinia';
import { useGlobalStore } from '../../store/global';
import applyPermPage from './apply-perm-page.vue';

const { spaceId, showPermApplyPage } = storeToRefs(useGlobalStore());

const route = useRoute();

watch(() => route.params.spaceId, (val) => {
  spaceId.value = val as string;
}, { immediate: true });

</script>
