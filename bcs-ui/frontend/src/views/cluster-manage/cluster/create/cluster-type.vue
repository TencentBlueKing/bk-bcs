<template>
  <BcsContent>
    <div class="text-[#333C48] text-[16px] mb-[16px]">{{ $t('新建集群') }}</div>
    <div class="flex flex-wrap">
      <div
        v-for="item in createList"
        :key="item.type"
        :class="['cluster-type-card', { disabled: item.disabled }]"
        @click="handleAddCluster(item)">
        <div class="cluster-type-card-top">
          <div class="['w-[48px] h-[48px]" :style="{ filter: item.disabled ? 'grayscale(1)' : '' }">
            <img :src="item.icon" />
          </div>
          <div class="text-[14px] font-bold mt-[10px]">{{ item.title }}</div>
          <div class="text-[14px] mt-[10px]">{{ item.subTitle }}</div>
        </div>
        <div class="flex-1 flex items-center justify-center text-[#979BA5] text-[12px]">
          {{ item.desc }}
        </div>
      </div>
    </div>
    <div class="text-[#333C48] text-[16px] mb-[16px] mt-[16px]">{{ $t('导入外部集群') }}</div>
    <div class="flex flex-wrap">
      <div
        v-for="item in importList"
        :key="item.type"
        :class="['cluster-type-card', { disabled: item.disabled }]"
        @click="handleAddCluster(item)">
        <div class="cluster-type-card-top">
          <div class="['w-[48px] h-[48px]" :style="{ filter: item.disabled ? 'grayscale(1)' : '' }">
            <img :src="item.icon" />
          </div>
          <div class="text-[14px] font-bold mt-[10px]">{{ item.title }}</div>
          <div class="text-[14px] mt-[10px]">{{ item.subTitle }}</div>
        </div>
        <div class="flex-1 flex items-center justify-center text-[#979BA5] text-[12px]">
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
import tkePng from '@/images/tke.png';
import cloudPng from '@/images/cloud.png';
import k8sPng from '@/images/k8s.png';
import kubeConfigPng from '@/images/kubeconfig.png';
import { useConfig } from '@/composables/use-app';

export default defineComponent({
  name: 'CreateCluster',
  components: { BcsContent },
  setup() {
    const createList = ref([
      {
        icon: tkePng,
        title: $i18n.t('腾讯云自研云集群'),
        subTitle: $i18n.t('为腾讯内部创建 TKE 集群提供解决方案'),
        desc: $i18n.t('提供独立集群与托管集群两种集群模式，独立集群需自行维护，托管集群由腾讯云代为维护控制面。'),
        type: 'tke',
      },
      {
        icon: cloudPng,
        title: $i18n.t('公有云集群'),
        subTitle: $i18n.t('实现多云集群统一管理，降低业务管理成本'),
        desc: $i18n.t('目前支持腾讯云、亚马逊云、谷歌云、微软云四种云服务商集群创建。'),
        type: 'cloud',
        disabled: true,
      },
      {
        icon: k8sPng,
        title: $i18n.t('K8S原生集群'),
        subTitle: $i18n.t('为私有环境搭建集群提供解决方案'),
        desc: $i18n.t('如果业务环境无法使用任何云服务商产品，该功能可在私有化环境搭建K8S原生集群。'),
        type: 'k8s',
        disabled: true,
      },
    ]);
    const { _INTERNAL_ } = useConfig();
    const importList = computed(() => [
      {
        icon: kubeConfigPng,
        title: $i18n.t('kubeconfig 集群导入'),
        subTitle: $i18n.t('可以通过 kubeconfig 导入任意外部集群'),
        desc: $i18n.t('使用具备 kube-admin 角色权限的 kubeconfig 即可使用蓝鲸容器管理平台纳管外部集群。'),
        type: 'kubeconfig',
        disabled: _INTERNAL_.value,
      },
      {
        icon: cloudPng,
        title: $i18n.t('公有云集群'),
        subTitle: $i18n.t('实现多云集群统一管理，降低业务管理成本'),
        desc: $i18n.t('目前支持腾讯云、亚马逊云、谷歌云、微软云四种云服务商集群创建。'),
        type: 'cloud',
        disabled: true,
      },
    ]);

    const handleAddCluster = (item) => {
      if (item.disabled) return;

      switch (item.type) {
        case 'tke':
          $router.push({ name: 'createFormCluster' });
          break;
        case 'kubeconfig':
          $router.push({ name: 'createImportCluster' });
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
