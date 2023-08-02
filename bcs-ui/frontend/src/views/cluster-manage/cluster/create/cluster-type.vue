<template>
  <BcsContent>
    <div class="text-[#333C48] text-[16px] mb-[16px]">{{ $t('cluster.create.title.newCluster') }}</div>
    <div class="flex flex-wrap">
      <div
        v-for="item in createList"
        :key="item.type"
        :class="['cluster-type-card', { disabled: item.disabled }]"
        @click="handleAddCluster(item)">
        <div class="cluster-type-card-top">
          <div :style="{ filter: item.disabled ? 'grayscale(1)' : '' }">
            <svg class="icon svg-icon" width="48px" height="48px">
              <use :xlink:href="`#${item.icon}`"></use>
            </svg>
          </div>
          <div class="text-[14px] font-bold mt-[10px]">{{ item.title }}</div>
          <div class="text-[14px] mt-[10px] bcs-ellipsis line-clamp-2">{{ item.subTitle }}</div>
        </div>
        <div class="flex-1 flex items-center justify-center text-[#979BA5] text-[12px] h-[78px]">
          {{ item.desc }}
        </div>
      </div>
    </div>
    <div class="text-[#333C48] text-[16px] mb-[16px] mt-[16px]">{{ $t('cluster.create.title.importCluster') }}</div>
    <div class="flex flex-wrap">
      <div
        v-for="item in importList"
        :key="item.type"
        :class="['cluster-type-card', { disabled: item.disabled }]"
        @click="handleAddCluster(item)">
        <div class="cluster-type-card-top">
          <div :style="{ filter: item.disabled ? 'grayscale(1)' : '' }">
            <svg class="icon svg-icon" width="48px" height="48px">
              <use :xlink:href="`#${item.icon}`"></use>
            </svg>
          </div>
          <div class="text-[14px] font-bold mt-[10px]">{{ item.title }}</div>
          <div class="text-[14px] mt-[10px] bcs-ellipsis line-clamp-2">{{ item.subTitle }}</div>
        </div>
        <div class="flex-1 flex items-center justify-center text-[#979BA5] text-[12px] h-[78px]">
          {{ item.desc }}
        </div>
      </div>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, ref } from 'vue';
import BcsContent from '../../components/bcs-content.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import { useConfig, useAppData } from '@/composables/use-app';

export default defineComponent({
  name: 'CreateCluster',
  components: { BcsContent },
  setup() {
    const { _INTERNAL_ } = useConfig();
    const { flagsMap } = useAppData();
    const createList = ref([
      {
        icon: 'bcs-icon-color-vcluster',
        title: 'vCluster',
        subTitle: $i18n.t('cluster.create.type.vCluster.subTitle'),
        desc: $i18n.t('cluster.create.type.vCluster.desc'),
        type: 'vCluster',
        disabled: !flagsMap.value.VCLUSTER,
      },
      {
        icon: 'bcs-icon-color-tencentcloud',
        title: $i18n.t('cluster.create.type.tencentCloud.title'),
        subTitle: $i18n.t('cluster.create.type.tencentCloud.subTitle'),
        desc: $i18n.t('cluster.create.type.tencentCloud.desc'),
        type: 'tke',
        disabled: !_INTERNAL_.value,
      },
      {
        icon: 'bcs-icon-color-publiccloud',
        title: $i18n.t('cluster.create.type.cloudProvider.title'),
        subTitle: $i18n.t('cluster.create.type.cloudProvider.subTitle'),
        desc: $i18n.t('cluster.create.type.cloudProvider.desc'),
        type: 'cloud',
        disabled: true,
      },
      {
        icon: 'bcs-icon-color-k8s',
        title: $i18n.t('cluster.create.type.k8s.title'),
        subTitle: $i18n.t('cluster.create.type.k8s.subTitle'),
        desc: $i18n.t('cluster.create.type.k8s.desc'),
        type: 'k8s',
        disabled: !flagsMap.value.k8s,
      },
    ]);
    const importList = computed(() => [
      {
        icon: 'bcs-icon-color-kubeconfig',
        title: $i18n.t('cluster.create.type.kubeconfig.title'),
        subTitle: $i18n.t('cluster.create.type.kubeconfig.subTitle'),
        desc: $i18n.t('cluster.create.type.kubeconfig.desc'),
        type: 'kubeconfig',
        disabled: _INTERNAL_.value,
      },
      {
        icon: 'bcs-icon-color-publiccloud',
        title: $i18n.t('cluster.create.type.cloudProvider.title'),
        subTitle: $i18n.t('cluster.create.type.cloudProvider.subTitle'),
        desc: $i18n.t('cluster.create.type.cloudProvider.desc'),
        type: 'import-cloud',
        disabled: _INTERNAL_.value,
      },
    ]);

    const handleAddCluster = (item) => {
      if (item.disabled) return;

      switch (item.type) {
        case 'vCluster':
          $router.push({ name: 'createVCluster' });
          break;
        case 'tke':
          $router.push({ name: 'createTencentCloudCluster' });
          break;
        case 'k8s':
          $router.push({ name: 'createCluster' });
          break;
        case 'kubeconfig':
          $router.push({
            name: 'importCluster',
            params: { importType: 'kubeconfig' },
          });
          break;
        case 'import-cloud':
          $router.push({
            name: 'importCluster',
            params: { importType: 'provider' },
          });
          break;
      }
    };

    return {
      createList,
      importList,
      handleAddCluster,
    };
  },
});
</script>
<style lang="postcss" scoped>
.cluster-type-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 0 24px;
  width: 350px;
  height: 220px;
  background: #fff;
  box-shadow: 0 2px 4px 0 #0000001a, 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  margin-right: 16px;
  margin-bottom: 16px;
  cursor: pointer;
  border: 1px solid #fff;
  &:hover {
    border:1px solid #3A84FF;
  }
  &.disabled {
    cursor: not-allowed;
    border: unset;
    div {
      color: #c4c6cc;
    }
  }
}
.cluster-type-card-top {
  height: 142px;
  border-bottom: 1px solid #DCDEE5;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
}
</style>
