<template>
  <BcsContent :title="$t('projects.project.info')" hide-back>
    <bk-form class="project-info">
      <bk-form-item :label="$t('projects.project.name')">
        <span class="text-[#313238] text-[12px]">{{ curProject.name }}</span>
      </bk-form-item>
      <bk-form-item class="!mt-[0px]" :label="$t('projects.project.engName')">
        <span class="text-[#313238] text-[12px]">{{ curProject.projectCode }}</span>
      </bk-form-item>
      <bk-form-item class="!mt-[0px]" :label="$t('projects.project.intro')">
        <span class="text-[#313238] text-[12px]">{{ curProject.description || '--' }}</span>
      </bk-form-item>
      <bk-form-item class="!mt-[0px]" :label="$t('projects.project.businessID')">
        <span class="text-[#313238] text-[12px]">
          {{ `${curProject.businessName}(${curProject.businessID})` }}
        </span>
      </bk-form-item>
    </bk-form>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent } from 'vue';

import BcsContent from '@/components/layout/Content.vue';
import { useProject } from '@/composables/use-app';

export default defineComponent({
  name: 'ProjectInfo',
  components: { BcsContent },
  setup() {
    const { curProject } = useProject();

    return {
      curProject,
    };
  },
});
</script>
<style lang="postcss" scoped>
.project-info {
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  padding: 32px;
}
</style>
