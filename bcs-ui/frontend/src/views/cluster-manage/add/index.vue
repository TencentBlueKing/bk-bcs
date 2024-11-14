<template>
  <BcsContent :title="$t('cluster.button.addCluster')">
    <div class="flex flex-col items-center justify-center">
      <div
        v-for="card, index in cardGroupList"
        :key="index"
        :class="[
          'px-[56px] max-w-[1318px] w-full',
          {
            'mt-[32px]': index !== 0
          }
        ]">
        <div
          :class="[
            'flex items-center justify-center rounded-[16px] min-w-[720px]',
            'text-[#313238] text-[16px] mb-[16px] h-[32px] bg-[#EAEBF0]'
          ]">
          {{ card.name }}
        </div>
        <div class="grid gap-[16px] grid-cols-[repeat(auto-fill,minmax(350px,1fr))] min-w-[720px]">
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
            :class="[
              'grid gap-[16px] grid-cols-[repeat(2,minmax(350px,1fr))]',
              'p-[16px] bg-[#F0F1F5] row-start-2 col-span-full text-[12px]'
            ]"
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
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, ref } from 'vue';

import BcsContent from '@/components/layout/Content.vue';
import { useAppData } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import amazonLogo from '@/images/amazon.png';
import azureLogo from '@/images/azure.png';
import googleLogo from '@/images/google.png';
import huaweiLogo from '@/images/huawei.png';
import $router from '@/router';

type AddClusterType = '_tke'
| 'createPublicCloud'
| 'createTencentCloud'
| 'createAWSCloud'
| 'createGoogleCloud'
| 'createAzureCloud'
| 'createHuaweiCloud'
| 'createK8S'
| 'createVCluster'
| 'importPublicCloud'
| 'importTencentCloud'
| 'importAmazonCloud'
| 'importGoogleCloud'
| 'importAzureCloud'
| 'importHuaweiCloud'
| 'importKubeConfig'
| 'importBKSops';

interface ICard {
  icon?: any
  title?: string
  subTitle?: string
  desc?: string
  type?: AddClusterType
  disabled?: boolean
  children?: ICard[]
}
interface IGroup {
  id: string
  name: string
  data: ICard[]
}

export default defineComponent({
  name: 'ClusterType',
  components: { BcsContent },
  setup() {
    const { flagsMap, _INTERNAL_ } = useAppData();
    const createList = ref<ICard[]>([
      // tke集群
      {
        icon: 'bcs-icon-color-tencentcloud',
        title: $i18n.t('cluster.create.type.tencentCloud.title'),
        subTitle: $i18n.t('cluster.create.type.tencentCloud.subTitle'),
        desc: $i18n.t('cluster.create.type.tencentCloud.desc'),
        type: '_tke',
        disabled: !_INTERNAL_.value,
      },
      // 创建公有云集群
      {
        icon: 'bcs-icon-color-publiccloud',
        title: $i18n.t('cluster.create.type.cloudProvider.title'),
        subTitle: $i18n.t('cluster.create.type.cloudProvider.subTitle'),
        desc: $i18n.t('cluster.create.type.cloudProvider.desc'),
        type: 'createPublicCloud',
        disabled: _INTERNAL_.value,
        children: [
          // 腾讯云
          {
            icon: 'bcs-icon-color-tencentcloud',
            title: $i18n.t('publicCloud.tencent.title'),
            desc: $i18n.t('publicCloud.tencent.desc'),
            type: 'createTencentCloud',
          },
          // AWS
          {
            icon: amazonLogo,
            title: $i18n.t('publicCloud.amazon.title'),
            desc: $i18n.t('publicCloud.amazon.desc'),
            type: 'createAWSCloud',
          },
          // google
          {
            icon: googleLogo,
            title: $i18n.t('publicCloud.google.title'),
            desc: $i18n.t('publicCloud.google.desc'),
            type: 'createGoogleCloud',
          },
          // Azure
          {
            icon: azureLogo,
            title: $i18n.t('publicCloud.azure.title'),
            desc: $i18n.t('publicCloud.azure.desc'),
            type: 'createAzureCloud',
          },
          // huawei
          {
            icon: huaweiLogo,
            title: $i18n.t('publicCloud.huawei.title'),
            desc: $i18n.t('publicCloud.huawei.desc'),
            type: 'createHuaweiCloud',
            disabled: true,
          },
        ],
      },
      // 创建原生K8S集群
      {
        icon: 'bcs-icon-color-k8s',
        title: $i18n.t('cluster.create.type.k8s.title'),
        subTitle: $i18n.t('cluster.create.type.k8s.subTitle'),
        desc: $i18n.t('cluster.create.type.k8s.desc'),
        type: 'createK8S',
        disabled: !flagsMap.value.k8s,
      },
      // vCluster集群
      {
        icon: 'bcs-icon-color-vcluster',
        title: 'vCluster',
        subTitle: $i18n.t('cluster.create.type.vCluster.subTitle'),
        desc: $i18n.t('cluster.create.type.vCluster.desc'),
        type: 'createVCluster',
        disabled: !flagsMap.value.VCLUSTER,
      },
    ]);
    // 导入集群
    const importList = ref<ICard[]>([
      // 导入公有云
      {
        icon: 'bcs-icon-color-publiccloud',
        title: $i18n.t('cluster.create.type.cloudProvider.title'),
        subTitle: $i18n.t('cluster.create.type.cloudProvider.subTitle'),
        desc: $i18n.t('cluster.create.type.cloudProvider.importDesc'),
        type: 'importPublicCloud',
        disabled: _INTERNAL_.value,
        children: [
          {
            icon: 'bcs-icon-color-tencentcloud',
            title: $i18n.t('publicCloud.tencent.title'),
            desc: $i18n.t('publicCloud.tencent.desc'),
            type: 'importTencentCloud',
            disabled: _INTERNAL_.value,
          },
          {
            icon: amazonLogo,
            title: $i18n.t('publicCloud.amazon.title'),
            desc: $i18n.t('publicCloud.amazon.desc'),
            type: 'importAmazonCloud',
            disabled: _INTERNAL_.value || !flagsMap.value.AZURECLOUD,
          },
          {
            icon: googleLogo,
            title: $i18n.t('publicCloud.google.title'),
            desc: $i18n.t('publicCloud.google.desc'),
            type: 'importGoogleCloud',
            disabled: _INTERNAL_.value,
          },
          {
            icon: azureLogo,
            title: $i18n.t('publicCloud.azure.title'),
            desc: $i18n.t('publicCloud.azure.desc'),
            type: 'importAzureCloud',
            disabled: _INTERNAL_.value,
          },
          {
            icon: huaweiLogo,
            title: $i18n.t('publicCloud.huawei.title'),
            desc: $i18n.t('publicCloud.huawei.desc'),
            type: 'importHuaweiCloud',
            disabled: _INTERNAL_.value,
          },
        ],
      },
      // 导入kubeConfig
      {
        icon: 'bcs-icon-color-k8s',
        title: $i18n.t('cluster.create.type.kubeconfig.title'),
        subTitle: $i18n.t('cluster.create.type.kubeconfig.subTitle'),
        desc: $i18n.t('cluster.create.type.kubeconfig.desc'),
        type: 'importKubeConfig',
      },
      // 导入标准运维
      {
        icon: 'bcs-icon-operations',
        title: $i18n.t('cluster.create.type.bkSops.title'),
        subTitle: $i18n.t('cluster.create.type.bkSops.subTitle'),
        desc: $i18n.t('cluster.create.type.bkSops.desc'),
        type: 'importBKSops',
        disabled: !flagsMap.value.IMPORTSOPSCLUSTER,
      },
    ]);
    // 分组列表
    const cardGroupList = computed<IGroup[]>(() => [
      {
        id: 'create',
        name: $i18n.t('cluster.create.title.newCluster'),
        data: createList.value.filter((item) => {
          if (_INTERNAL_.value) {
            return item.type !== 'createPublicCloud';
          }
          return item.type !== '_tke';
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
    // 类型 -> 路由名称map
    const typeRouteNameMap: Partial<Record<AddClusterType, string>> = {
      _tke: 'createTencentCloudCluster',
      createTencentCloud: 'createTKECluster',
      createAWSCloud: 'CreateAWSCloudCluster',
      createAzureCloud: 'CreateAzureCloudCluster',
      createGoogleCloud: 'CreateGoogleCloudCluster',
      createK8S: 'createK8SCluster',
      createVCluster: 'createVCluster',
      // importKubeConfig: 'importCluster',
      // importTencentCloud: 'importCluster',
      importGoogleCloud: 'importGoogleCluster',
      importAzureCloud: 'importAzureCluster',
      importHuaweiCloud: 'importHuaweiCluster',
      importAmazonCloud: 'importAwsCluster',
      importBKSops: 'importBkSopsCluster',
    };
    const toggleActiveItem = (item: ICard, group: IGroup) => {
      if (activeItem.value === item || !group) {
        activeItem.value = null;
        activeGroupID.value = '';
      } else {
        activeItem.value = item;
        activeGroupID.value = group.id;
      }
    };
    const handleAddCluster = (item: ICard, group: IGroup) => {
      if (item.disabled || !item.type) return;

      toggleActiveItem(item, group);
      switch (item.type) {
        case 'importKubeConfig':
          $router.push({
            name: 'importCluster',
            params: { importType: 'kubeconfig' },
          });
          break;
        case 'importTencentCloud':
          $router.push({
            name: 'importCluster',
            params: { importType: 'provider' },
          });
          break;
        default:
          if (typeRouteNameMap[item.type]) {
            $router.push({
              name: typeRouteNameMap[item.type],
            });
          }
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
