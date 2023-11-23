<template>
  <template v-if="!loading">
    <whitelist-apply-page v-if="!spaceFeatureFlags.BIZ_VIEW"></whitelist-apply-page>
    <apply-perm-page v-else-if="showPermApplyPage"></apply-perm-page>
    <router-view v-else></router-view>
  </template>
</template>
<script setup lang="ts">
import { watch, ref } from 'vue';
import { useRoute } from 'vue-router';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../../store/global';
import { getSpaceFeatureFlag } from '../../api';
import whitelistApplyPage from './whitelist-apply-page.vue';
import applyPermPage from './apply-perm-page.vue';

const { spaceId, spaceFeatureFlags, showPermApplyPage } = storeToRefs(useGlobalStore());

const route = useRoute();

const loading = ref(true);

const getFeatureFlagsData = async() => {
  loading.value = true;
  const res = await getSpaceFeatureFlag(spaceId.value);
  spaceFeatureFlags.value = res;
  loading.value = false;
}

watch(
  () => route.params.spaceId,
  (val) => {
    spaceId.value = val as string;
    getFeatureFlagsData();
  },
  { immediate: true },
);
</script>
