<template>
  <div class="packages-tag">
    <template v-if="props.pkgs.length > 0">
      <div
        v-overflow-title
        class="pkg-tag tag-in-table"
        :key="props.pkgs[0].template_set_id"
        @click="goToPkg(props.pkgs[0].template_set_id)">
        {{ props.pkgs[0].template_set_name }}
      </div>
    </template>
    <div v-if="extPkgs.length > 0" class="ext-pkgs-num">
      <bk-popover theme="light" placement="top-center">
        <div class="pkg-tag">+{{ extPkgs.length }}</div>
        <template #content>
          <div v-for="pkg in extPkgs" class="pkg-tag" :key="pkg.template_set_id" @click="goToPkg(pkg.template_set_id)">
            {{ pkg.template_set_name }}
          </div>
        </template>
      </bk-popover>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import useTemplateStore from '../../../../../store/template';

const route = useRoute();
const router = useRouter();

const templateStore = useTemplateStore();

const props = defineProps<{
  pkgs: { template_set_id: number; template_set_name: string }[];
}>();

const extPkgs = computed(() => {
  if (props.pkgs.length > 0) {
    return props.pkgs.slice(1);
  }
  return [];
});

const goToPkg = (id: number) => {
  const { params } = route;
  router.push({
    name: 'templates-list',
    params: {
      ...params,
      packageId: id,
    },
  });

  templateStore.$patch((state) => {
    state.currentPkg = id;
  });
};
</script>
<style lang="scss" scoped>
.packages-tag {
  display: flex;
  align-items: center;
  width: 100%;
  height: 100%;
}
.pkg-tag {
  display: inline-block;
  padding: 0 8px;
  height: 22px;
  line-height: 22px;
  font-size: 12px;
  color: #63656e;
  background: #f0f1f5;
  border-radius: 2px;
  cursor: pointer;
  &:not(:last-of-type) {
    margin-right: 8px;
  }
}
.tag-in-table {
  max-width: 125px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
