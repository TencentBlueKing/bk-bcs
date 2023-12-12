<template>
  <BcsContent>
    <div
      v-for="card, index in cardGroupList"
      :key="index"
      :class="{
        'mt-[16px]': index !== 0
      }">
      <div class="text-[#333C48] text-[16px] mb-[16px]">{{ card.name }}</div>
      <div class="grid gap-[16px] grid-cols-[repeat(auto-fill,minmax(350px,1fr))]">
        <div
          v-for="item, i in card.data"
          :key="i"
          :class="[
            'cluster-type-card',
            { disabled: item.disabled }
          ]"
          @click="handleAddCluster(item, card)">
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
        <div
          class="row-start-2 col-span-full text-[12px] grid gap-[16px] grid-cols-[repeat(2,minmax(350px,1fr))]"
          v-if="activeItem && activeGroupID === card.id">
          <div
            v-for="child in activeItem.children"
            :key="child.type"
            :class="['cluster-type-card-2', { disabled: child.disabled }]"
            @click="handleAddCluster(child, card)">
            <div class="flex items-center">
              <div :style="{ filter: child.disabled ? 'grayscale(1)' : '' }">
                <img :src="child.icon" v-if="child.icon.indexOf('data:image/png') === 0" />
                <svg class="icon svg-icon" width="48px" height="48px" v-else>
                  <use :xlink:href="`#${child.icon}`"></use>
                </svg>
              </div>
              <span class="font-bold text-[14px] ml-[16px]">{{ child.title }}</span>
            </div>
            <div class="mt-[6px] text-[#979BA5]">
              {{ child.desc }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, ref } from 'vue';

import BcsContent from '../components/bcs-content.vue';

import { useAppData, useConfig } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import amazonLogo from '@/images/amazon.png';
import azureLogo from '@/images/azure.png';
import googleLogo from '@/images/google.png';
import $router from '@/router';

interface ICard {
  icon?: any
  title?: string
  subTitle?: string
  desc?: string
  type?: string
  disabled?: boolean
  children?: ICard[]
}
export default defineComponent({
  name: 'ClusterType',
  components: { BcsContent },
  setup() {
    const { _INTERNAL_ } = useConfig();
    const { flagsMap } = useAppData();
    const createList = ref<ICard[]>([
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
        children: [
          {
            icon: 'bcs-icon-color-tencentcloud',
            title: $i18n.t('publicCloud.tencent.title'),
            desc: $i18n.t('publicCloud.tencent.desc'),
            type: 'createTencentCloud',
            disabled: _INTERNAL_.value,
          },
          {
            icon: amazonLogo,
            title: $i18n.t('publicCloud.amazon.title'),
            desc: $i18n.t('publicCloud.amazon.desc'),
            type: 'amazonCloud',
            disabled: true,
          },
          {
            icon: googleLogo,
            title: $i18n.t('publicCloud.google.title'),
            desc: $i18n.t('publicCloud.google.desc'),
            type: 'googleCloud',
            disabled: true,
          },
          {
            icon: azureLogo,
            title: $i18n.t('publicCloud.azure.title'),
            desc: $i18n.t('publicCloud.azure.desc'),
            type: 'azureCloud',
            disabled: true,
          },
        ],
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
    const importList = ref<ICard[]>([
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
        children: [
          {
            icon: 'bcs-icon-color-tencentcloud',
            title: $i18n.t('publicCloud.tencent.title'),
            desc: $i18n.t('publicCloud.tencent.desc'),
            type: 'tencentCloud',
            disabled: _INTERNAL_.value,
          },
          {
            icon: amazonLogo,
            title: $i18n.t('publicCloud.amazon.title'),
            desc: $i18n.t('publicCloud.amazon.desc'),
            type: 'amazonCloud',
            disabled: true,
          },
          {
            icon: googleLogo,
            title: $i18n.t('publicCloud.google.title'),
            desc: $i18n.t('publicCloud.google.desc'),
            type: 'googleCloud',
            disabled: _INTERNAL_.value,
          },
          {
            icon: azureLogo,
            title: $i18n.t('publicCloud.azure.title'),
            desc: $i18n.t('publicCloud.azure.desc'),
            type: 'azureCloud',
            disabled: true,
          },
        ],
      },
    ]);
    const cardGroupList = computed(() => [
      {
        id: 'create',
        name: $i18n.t('cluster.create.title.newCluster'),
        data: createList.value.filter((item) => {
          if (_INTERNAL_.value) {
            return item.type !== 'cloud';
          }
          return item.type !== 'tke';
        }),
      },
      {
        id: 'import',
        name: $i18n.t('cluster.create.title.importCluster'),
        data: importList.value,
      },
    ]);

    const activeItem = ref<ICard|null>(null);
    const activeGroupID = ref('');
    const toggleActiveItem = (item, card) => {
      if (activeItem.value === item || !card) {
        activeItem.value = null;
        activeGroupID.value = '';
      } else {
        activeItem.value = item;
        activeGroupID.value = card.id;
      }
    };
    const handleAddCluster = (item, card) => {
      if (item.disabled) return;

      toggleActiveItem(item, card);
      switch (item.type) {
        case 'vCluster':
          $router.push({ name: 'createVCluster' });
          break;
        case 'tke':
          // 创建腾讯云集群
          $router.push({ name: 'createTencentCloudCluster' });
          break;
        case 'k8s':
          $router.push({ name: 'createK8SCluster' });
          break;
        case 'createTencentCloud':
          // 创建腾讯云公有云集群
          $router.push({ name: 'createTKECluster' });
          break;
        case 'kubeconfig':
          $router.push({
            name: 'importCluster',
            params: { importType: 'kubeconfig' },
          });
          break;
        case 'tencentCloud':
          $router.push({
            name: 'importCluster',
            params: { importType: 'provider' },
          });
          break;
        case 'googleCloud':
          $router.push({
            name: 'importGoogleCluster',
          });
          break;
      }
    };

    return {
      activeItem,
      activeGroupID,
      cardGroupList,
      handleAddCluster,
    };
  },
});
</script>
<style lang="postcss">
.cloud-cluster-popover-theme {
  min-width: 800px !important;
  padding: 16px !important;
  background-color: #F0F1F5 !important;
  .tippy-content {
    max-width: 1000px;
  }
}
</style>
<style lang="postcss" scoped>
.cluster-type-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 0 24px;
  height: 220px;
  background: #fff;
  box-shadow: 0 2px 4px 0 #0000001a, 0 2px 4px 0 #1919290d;
  border-radius: 2px;
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
.cluster-type-card-2 {
  background-color: #fff;
  padding: 24px 16px;
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
