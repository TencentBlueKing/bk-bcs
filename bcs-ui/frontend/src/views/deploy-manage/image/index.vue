<template>
  <BcsContent
    :title="$t('镜像')"
    :tabs="tabs"
    :active-tab="$route.name || ''"
    hide-back
    @tab-change="handleTabChange">
    <RouterView :key="$route.path" />
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, ref } from 'vue';
import BcsContent from '@/components/layout/Content.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

export default defineComponent({
  name: 'ImageRepo',
  components: { BcsContent },
  setup() {
    const tabs = ref([
      {
        name: 'projectImage',
        displayName: $i18n.t('项目镜像'),
      },
      {
        name: 'imageLibrary',
        displayName: $i18n.t('公共镜像'),
      },
    ]);
    const handleTabChange = (item) => {
      $router.push({ name: item.name });
    };
    return {
      tabs,
      handleTabChange,
    };
  },
});
</script>
