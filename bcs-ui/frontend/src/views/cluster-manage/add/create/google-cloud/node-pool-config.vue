<!-- eslint-disable max-len -->
<template>
  <div>
    <div class="p-[24px] node-config-wrapper" ref="nodeConfigRef">
      <bk-form class="node-config" :model="nodePoolConfig" :rules="nodePoolConfigRules" ref="formRef" :label-width="100">
        <bk-form-item class="max-w-[674px]" :label="$t('cluster.ca.nodePool.label.name')" property="name" required>
          <bk-input v-model="nodePoolConfig.name"></bk-input>
          <p class="text-[#979BA5] leading-4 mt-[4px]">{{ $t('cluster.ca.nodePool.validate.name') }}</p>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.label.system')"
          error-display-type="normal"
          required
          :desc="$t('googleCA.tips.system')">
          <bcs-input disabled v-model="nodePoolConfig.launchTemplate.imageInfo.imageName"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.az.title')"
          :desc="$t('googleCA.tips.zone')"
          property="nodePoolConfig.autoScaling.zones">
          <div class="flex items-center h-[32px]">
            <bcs-select
              class="flex-1"
              :value="nodePoolConfig.autoScaling.zones[0]"
              searchable
              :loading="zoneListLoading"
              :clearable="false"
              selected-style="checkbox"
              @change="handleSetZoneData">
              <bcs-option
                v-for="zone in zoneList"
                :key="zone.zoneID"
                :id="zone.zone"
                :name="zone.zoneName" />
            </bcs-select>
          </div>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.instanceTypeConfig.title')">
          <div class="mb15" style="display: flex;">
            <div class="prefix-select">
              <span :class="['prefix', { disabled: isEdit }]">CPU</span>
              <bcs-select
                v-model="CPU"
                searchable
                :clearable="false"
                :disabled="isEdit"
                :placeholder="' '"
                class="bg-[#fff]">
                <bcs-option id="" :name="$t('generic.label.total')"></bcs-option>
                <bcs-option
                  v-for="cpuItem in cpuList"
                  :key="cpuItem"
                  :id="cpuItem"
                  :name="cpuItem">
                </bcs-option>
              </bcs-select>
              <span :class="['company', { disabled: isEdit }]">{{$t('units.suffix.cores')}}</span>
            </div>
            <div class="prefix-select ml30">
              <span :class="['prefix', { disabled: isEdit }]">{{$t('generic.label.mem')}}</span>
              <bcs-select
                v-model="Mem"
                searchable
                :clearable="false"
                :disabled="isEdit"
                :placeholder="' '"
                class="bg-[#fff]">
                <bcs-option id="" :name="$t('generic.label.total')"></bcs-option>
                <bcs-option
                  v-for="memItem in memList"
                  :key="memItem"
                  :id="memItem"
                  :name="memItem">
                </bcs-option>
              </bcs-select>
              <span :class="['company', { disabled: isEdit }]">G</span>
            </div>
          </div>
          <bcs-table
            :data="instanceList"
            v-bkloading="{ isLoading: instanceTypesLoading || clusterDetailLoading }"
            :pagination="pagination"
            :row-class-name="instanceRowClass"
            @page-change="pageChange"
            @page-limit-change="pageSizeChange"
            @row-click="handleCheckInstanceType">
            <bcs-table-column :label="$t('generic.ipSelector.label.serverModel')" prop="typeName" show-overflow-tooltip>
              <template #default="{ row }">
                <span v-bk-tooltips="{ disabled: row.status !== 'SOLD_OUT', content: $t('cluster.ca.nodePool.create.instanceTypeConfig.status.soldOut') }">
                  <bcs-radio
                    :value="nodePoolConfig.launchTemplate.instanceType === row.nodeType"
                    :disabled="row.status === 'SOLD_OUT' || isEdit">
                    <span class="bcs-ellipsis">{{row.typeName}}</span>
                  </bcs-radio>
                </span>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.specifications')" min-width="120" show-overflow-tooltip prop="nodeType"></bcs-table-column>
            <bcs-table-column label="CPU" prop="cpu" width="80" align="right">
              <template #default="{ row }">
                <span>{{ `${row.cpu}${$t('units.suffix.cores')}` }}</span>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.mem')" prop="memory" width="80" align="right">
              <template #default="{ row }">
                <span>{{ row.memory }}G</span>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.status')" width="80">
              <template #default="{ row }">
                {{ row.status === 'SELL' ? $t('cluster.ca.nodePool.create.instanceTypeConfig.status.sell') : $t('cluster.ca.nodePool.create.instanceTypeConfig.status.soldOut') }}
              </template>
            </bcs-table-column>
          </bcs-table>
          <p class="text-[12px] text-[#ea3636]" v-if="!nodePoolConfig.launchTemplate.instanceType">{{ $t('generic.validate.required') }}</p>
          <div class="mt25" style="display:flex;align-items:center;">
            <div class="prefix-select">
              <span :class="['prefix', { disabled: isEdit }]">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.disk.system')}}</span>
              <bcs-select
                v-model="nodePoolConfig.launchTemplate.systemDisk.diskType"
                :disabled="isEdit"
                :clearable="false"
                class="min-width-150 bg-[#fff]">
                <bcs-option
                  v-for="diskItem in diskEnum"
                  :key="diskItem.id"
                  :id="diskItem.id"
                  :name="diskItem.name">
                </bcs-option>
              </bcs-select>
            </div>
            <bcs-select
              class="w-[88px] bg-[#fff] ml10"
              :disabled="isEdit"
              :clearable="false"
              v-model="nodePoolConfig.launchTemplate.systemDisk.diskSize">
              <bcs-option id="50" name="50"></bcs-option>
              <bcs-option id="100" name="100"></bcs-option>
            </bcs-select>
            <span :class="['company', { disabled: isEdit }]">GB</span>
          </div>
          <div class="mt20">
            <bk-checkbox
              :disabled="isEdit"
              v-model="showDataDisks"
              @change="handleShowDataDisksChange">
              <span v-bk-tooltips="$t('googleCA.tips.dataDisk')">
                {{$t('cluster.ca.nodePool.create.instanceTypeConfig.label.purchaseDataDisk')}}
              </span>
            </bk-checkbox>
          </div>
          <template v-if="showDataDisks">
            <div class="panel" v-for="(disk, index) in nodePoolConfig.nodeTemplate.dataDisks" :key="index">
              <div class="panel-item">
                <div class="prefix-select">
                  <span :class="['prefix', { disabled: isEdit }]">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.disk.data')}}</span>
                  <bcs-select
                    :disabled="isEdit"
                    v-model="disk.diskType"
                    :clearable="false"
                    class="min-width-150 bg-[#fff]">
                    <bcs-option
                      v-for="diskItem in diskEnum"
                      :key="diskItem.id"
                      :id="diskItem.id"
                      :name="diskItem.name">
                    </bcs-option>
                  </bcs-select>
                </div>
                <bk-input
                  class="max-width-130 ml10"
                  type="number"
                  :disabled="isEdit"
                  :min="50"
                  :max="16380"
                  v-model="disk.diskSize">
                </bk-input>
                <span :class="['company', { disabled: isEdit }]">GB</span>
              </div>
              <p class="error-tips" v-if="disk.diskSize % 10 !== 0">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.validate.dataDisks')}}</p>
              <span :class="['panel-delete', { disabled: isEdit }]" @click="handleDeleteDiskData(index)">
                <i class="bk-icon icon-close3-shape"></i>
              </span>
            </div>
            <div
              :class="['add-panel-btn', { disabled: isEdit || nodePoolConfig.nodeTemplate.dataDisks.length > 4 }]"
              v-bk-tooltips="{
                content: $t('cluster.ca.nodePool.create.instanceTypeConfig.validate.maxDataDisks'),
                disabled: nodePoolConfig.nodeTemplate.dataDisks.length <= 4
              }"
              @click="handleAddDiskData">
              <i class="bk-icon left-icon icon-plus"></i>
              <span>{{$t('cluster.ca.nodePool.create.instanceTypeConfig.button.addDataDisks')}}</span>
            </div>
          </template>
          <span class="inline-flex mt15">
            <bk-checkbox
              :disabled="isEdit"
              v-model="nodePoolConfig.launchTemplate.internetAccess.publicIPAssigned">
              {{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.text')}}
            </bk-checkbox>
          </span>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.securityGroup')">
          <i18n path="googleCA.tips.securityGroup">
            <a
              class="text-[#3a84ff] cursor-pointer"
              href="https://console.cloud.google.com/net-security/firewall-manager/firewall-policies/list"
              target="_blank">
              {{ $t('googleCA.button.vpcRule') }}
            </a>
          </i18n>
        </bk-form-item>
        <!-- <bk-form-item :label="$t('dashboard.workload.container.dataDir')">
          <bcs-input disabled v-model="nodePoolConfig.nodeTemplate.dockerGraphPath"></bcs-input>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.containerRuntime.title')">
          <bk-radio-group v-model="nodePoolConfig.containerRuntime">
            <bk-radio value="docker" disabled>docker</bk-radio>
            <bk-radio value="containerd" disabled>containerd</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.runtimeVersion.title')">
          <bcs-input disabled v-model="nodePoolConfig.runtimeVersion"></bcs-input>
        </bk-form-item> -->
      </bk-form>
    </div>
    <div class="bcs-border-top z-10 flex items-center sticky bottom-0 bg-[#fff] h-[60px] px-[24px]">
      <bcs-button theme="primary" @click="handleNext">{{$t('generic.button.next')}}</bcs-button>
      <bcs-button class="ml10" @click="handleCancel">{{$t('generic.button.cancel')}}</bcs-button>
    </div>
  </div>
</template>
<script lang="ts">
import { sortBy } from 'lodash';
import { computed, defineComponent, onMounted, ref, toRefs, watch } from 'vue';

import { cloudsZones } from '@/api/modules/cluster-manager';
import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';

export default defineComponent({
  props: {
    defaultValues: {
      type: Object,
      default: () => ({}),
    },
    cluster: {
      type: Object,
      default: () => ({}),
    },
    isEdit: {
      type: Boolean,
      default: false,
    },
    saveLoading: {
      type: Boolean,
      default: false,
    },
    cloudAccountID: {
      type: String,
      default: '',
    },
    cloudID: {
      type: String,
      default: '',
    },
    region: {
      type: String,
      default: '',
    },
    vpcID: {
      type: String,
      default: '',
    },
  },
  setup(props, ctx) {
    const { defaultValues, isEdit } = toRefs(props);
    const nodeConfigRef = ref<any>(null);
    const formRef = ref<any>(null);
    const basicFormRef = ref<any>(null);
    // 磁盘类型
    const diskEnum = ref([
      {
        id: 'pd-balanced',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.balanced'),
      },
      {
        id: 'pd-ssd',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.ssd'),
      },
      {
        id: 'pd-standard',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.standard'),
      },
    ]);
    const user = computed(() => $store.state.user);
    const nodePoolConfig = ref({
      $id: defaultValues.value.$id || `id_${Date.now()}`,
      status: 'CREATING',
      creator: user.value.username, // 创建人
      nodeOS: defaultValues.value.nodeOS || '', // 节点操作系统
      name: defaultValues.value.name || '', // 节点名称
      autoScaling: {
        subnetIDs: [], // 子网
        zones: defaultValues.value.autoScaling?.zones || [],
      },
      launchTemplate: {
        imageInfo: {
          imageName: defaultValues.value.launchTemplate?.imageInfo?.imageName, // 镜像名称
          imageType: defaultValues.value.launchTemplate?.imageInfo?.imageType || 'PUBLIC_IMAGE', // 镜像类型
        },
        CPU: '',
        Mem: '',
        instanceType: defaultValues.value.launchTemplate?.instanceType, // 机型信息
        systemDisk: {
          diskType: defaultValues.value.launchTemplate?.systemDisk?.diskType, // 系统盘类型
          diskSize: defaultValues.value.launchTemplate?.systemDisk?.diskSize, // 系统盘大小
        },
        internetAccess: {
          // internetChargeType: defaultValues.value.launchTemplate?.internetAccess?.internetChargeType, // 计费方式
          // internetMaxBandwidth: defaultValues.value.launchTemplate?.internetAccess?.internetMaxBandwidth, // 购买带宽
          publicIPAssigned: defaultValues.value.launchTemplate?.internetAccess?.publicIPAssigned, // 分配免费公网IP
          // bandwidthPackageId: defaultValues.value.launchTemplate?.internetAccess?.bandwidthPackageId, // 带宽包ID
        },
        initLoginUsername: defaultValues.value.launchTemplate?.initLoginUsername || '', // 登录用户
        // 密钥信息
        keyPair: {
          keyPublic: defaultValues.value.launchTemplate?.keyPair?.keyPublic || '',
          keySecret: defaultValues.value.launchTemplate?.keyPair?.keySecret || '',
        },
        initLoginPassword: defaultValues.value.launchTemplate?.initLoginPassword, // 密码
        dataDisks: defaultValues.value.launchTemplate?.dataDisks || [], // 数据盘
        // 默认值
        isSecurityService: defaultValues.value.launchTemplate?.isSecurityService || true,
        isMonitorService: defaultValues.value.launchTemplate?.isMonitorService || true,
      },
      nodeTemplate: {
        dataDisks: defaultValues.value.nodeTemplate?.dataDisks || [],
        dockerGraphPath: defaultValues.value.nodeTemplate?.dockerGraphPath, // 容器目录
        runtime: {
          containerRuntime: defaultValues.value.nodeTemplate?.runtime?.containerRuntime, // 运行时容器组件
          runtimeVersion: defaultValues.value.nodeTemplate?.runtime?.runtimeVersion, // 运行时版本
        },
        nodeOS: defaultValues.value.nodeTemplate?.nodeOS,
      },
      extra: {
        provider: '', // 机型provider信息
      },
      containerRuntime: defaultValues.value.containerRuntime, // 容器运行时
      runtimeVersion: defaultValues.value.runtimeVersion, // 运行时版本
    });

    const nodePoolConfigRules = ref({
      name: [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
        {
          message: $i18n.t('cluster.ca.nodePool.create.validate.name'),
          trigger: 'blur',
          validator: (v: string) => /^[\u4E00-\u9FA5A-Za-z0-9._-]+$/.test(v) && v.length <= 255 && v.length >= 2,
        },
      ],
    });

    // 可用区
    const zoneList = ref<any[]>([]);
    const zoneListLoading = ref(false);
    const handleGetZoneList = async () => {
      zoneListLoading.value = true;
      zoneList.value = await cloudsZones({
        $cloudId: props.cloudID,
        region: props.region,
        accountID: props.cloudAccountID,
      });
      zoneListLoading.value = false;
    };
    const handleSetZoneData = (value) => {
      nodePoolConfig.value.autoScaling.zones = [value];
    };
    watch(() => nodePoolConfig.value.autoScaling.zones, () => {
      handleGetInstanceTypes();
    });

    // 镜像
    watch(() => nodePoolConfig.value.launchTemplate.imageInfo.imageType, () => {
      // 获取镜像列表
      handleGetOsImage();
    });
    const osImageLoading = ref(false);
    const osImageList = ref<Array<{
      imageID: string
      alias: string
    }>>([]);
    const handleGetOsImage = async () => {
      osImageLoading.value = true;
      osImageList.value = await $store.dispatch('clustermanager/cloudOsImage', {
        $cloudID: props.cloudID,
        region: props.region,
        accountID: props.cloudAccountID,
        provider: nodePoolConfig.value.launchTemplate.imageInfo.imageType,
      });
      // eslint-disable-next-line max-len
      nodePoolConfig.value.launchTemplate.imageInfo.imageName = (nodePoolConfig.value.launchTemplate.imageInfo.imageName || osImageList.value[0]?.imageID);
      osImageLoading.value = false;
    };

    // 机型
    const instanceTypesLoading = ref(false);
    const instanceData = ref<any[]>([]);
    const instanceTypesList = computed(() => {
      const cacheInstanceMap = {};
      // 先过滤可用区, 再过滤同类型机型
      return instanceData.value
        .filter((instance) => {
        // todo 简单过滤同类型机型
          if (!cacheInstanceMap[instance.nodeType]) {
            cacheInstanceMap[instance.nodeType] = true;
            return true;
          }
          return false;
        })
        .filter(instance => (!CPU.value || instance.cpu === CPU.value)
        && (!Mem.value || instance.memory === Mem.value));
    });
    let timer: any = null;
    watch(() => instanceTypesList.value.length, () => {
      const index = instanceTypesList.value.findIndex(item => item.nodeType === nodePoolConfig.value.launchTemplate.instanceType);
      timer && clearTimeout(timer);
      timer = setTimeout(() => {
        pageChange(Math.ceil((index + 1) / pagination.value.limit));
      }, 100);
      handleResetPage();
    });
    // eslint-disable-next-line max-len
    const curInstanceItem = computed(() => instanceData.value.find(instance => instance.nodeType === nodePoolConfig.value.launchTemplate.instanceType) || {});
    const cpuList = computed(() => {
      const data = instanceData.value.reduce((pre, item) => {
        if (!pre.includes(item.cpu)) {
          pre.push(item.cpu);
        }
        return pre;
      }, []);
      return sortBy(data);
    });
    const memList = computed(() => {
      const data = instanceData.value.reduce((pre, item) => {
        if (!pre.includes(item.memory)) {
          pre.push(item.memory);
        }
        return pre;
      }, []);
      return sortBy(data);
    });

    const CPU = ref('');
    const Mem = ref('');
    watch(() => [
      CPU.value,
      Mem.value,
    ], () => {
      // 重置机型
      nodePoolConfig.value.launchTemplate.instanceType = '';
      handleSetDefaultInstance();
    });
    watch(curInstanceItem, () => {
      nodePoolConfig.value.extra.provider = curInstanceItem.value.provider;
    });
    const instanceRowClass = ({ row }) => {
      // SELL 表示售卖，SOLD_OUT 表示售罄
      if (row.status === 'SELL' || isEdit.value) {
        return 'table-row-enable';
      }
      return 'table-row-disabled';
    };
    const handleCheckInstanceType = (row) => {
      if (row.status === 'SOLD_OUT') return;
      nodePoolConfig.value.launchTemplate.instanceType = row.nodeType;
    };
    // 设置默认机型
    const handleSetDefaultInstance = () => {
      setTimeout(() => {
        // 默认机型配置
        if (!nodePoolConfig.value.launchTemplate.instanceType) {
          nodePoolConfig.value.launchTemplate.instanceType = instanceTypesList.value
            .find(instance => instance.status === 'SELL')?.nodeType;
        }
      });
    };
    const handleGetInstanceTypes = async () => {
      if (!nodePoolConfig.value.autoScaling.zones?.length) return;
      instanceTypesLoading.value = true;
      const data = await $store.dispatch('clustermanager/cloudInstanceTypes', {
        $cloudID: props.cloudID,
        region: nodePoolConfig.value.autoScaling.zones[0],
        accountID: props.cloudAccountID,
        provider: nodePoolConfig.value.launchTemplate.imageInfo.imageType,
      });
      instanceData.value = data.sort((pre, current) => {
        if (pre.status !== 'SELL') return 1;
        if (current.status !== 'SELL') return 0;
        return pre.cpu - current.cpu;
      });
      handleSetDefaultInstance();
      instanceTypesLoading.value = false;
    };
    const {
      pagination,
      curPageData: instanceList,
      pageChange,
      pageSizeChange,
      handleResetPage,
    } = usePage(instanceTypesList);

    // 数据盘
    const showDataDisks = ref(!!nodePoolConfig.value.nodeTemplate.dataDisks.length);
    const defaultDiskItem = defaultValues.value?.hardwareProfile?.dataDisks.length
      ? defaultValues.value.hardwareProfile.dataDisks[0]
      : {
        // 默认为pd-standard
        diskType: 'pd-standard',
        diskSize: '100',
        autoFormatAndMount: true,
      };
    const handleShowDataDisksChange = (show) => {
      nodePoolConfig.value.nodeTemplate.dataDisks = show
        ? [JSON.parse(JSON.stringify(defaultDiskItem))] : [];
    };
    const handleDeleteDiskData = (index) => {
      if (isEdit.value) return;
      nodePoolConfig.value.nodeTemplate.dataDisks.splice(index, 1);
    };
    const handleAddDiskData = () => {
      if (isEdit.value || nodePoolConfig.value.nodeTemplate.dataDisks.length > 4) return;
      nodePoolConfig.value.nodeTemplate.dataDisks.push(JSON.parse(JSON.stringify(defaultDiskItem)));
    };

    // 操作
    const getNodePoolData = () => {
      // 系统盘、数据盘、宽度大小要转换为字符串类型
      // eslint-disable-next-line max-len
      nodePoolConfig.value.launchTemplate.systemDisk.diskSize = String(nodePoolConfig.value.launchTemplate.systemDisk.diskSize);
      nodePoolConfig.value.nodeTemplate.dataDisks = nodePoolConfig.value.nodeTemplate.dataDisks.map(item => ({
        ...item,
        diskSize: String(item.diskSize),
      }));
      // 数据盘后端存了两个地方
      nodePoolConfig.value.launchTemplate.dataDisks = nodePoolConfig.value.nodeTemplate.dataDisks.map(item => ({
        diskType: item.diskType,
        diskSize: item.diskSize,
      }));
      // CPU和mem信息从机型获取
      nodePoolConfig.value.launchTemplate.CPU = curInstanceItem.value.cpu;
      nodePoolConfig.value.launchTemplate.Mem = curInstanceItem.value.memory;
      return nodePoolConfig.value;
    };
    const validate = async () => {
      // const basicFormValidate = await basicFormRef.value?.validate().catch(() => false);
      // if (!basicFormValidate && nodeConfigRef.value) {
      //   nodeConfigRef.value.scrollTop = 0;
      //   return false;
      // }
      // 校验机型
      if (!nodePoolConfig.value.launchTemplate.instanceType) {
        nodeConfigRef.value.scrollTop = 20;
        return false;
      }
      const result = await formRef.value?.validate().catch(() => false);
      if (!result && nodeConfigRef.value) {
        nodeConfigRef.value.scrollTop = nodeConfigRef.value.scrollHeight;
      }
      // eslint-disable-next-line max-len
      const validateDataDiskSize = nodePoolConfig.value.nodeTemplate.dataDisks.every(item => item.diskSize % 10 === 0);
      // if (!basicFormValidate || !result || !validateDataDiskSize) return false;
      if (!result || !validateDataDiskSize) return false;

      return true;
    };

    const { focusOnErrorField } = useFocusOnErrorField();
    const handleNext = async () => {
      nodePoolConfig.value.region = nodePoolConfig.value.autoScaling.zones?.[0];
      // 校验错误滚动到第一个错误的位置
      const result = await validate();
      if (!result) {
        // 自动滚动到第一个错误的位置
        focusOnErrorField();
        return;
      }
      ctx.emit('next', getNodePoolData());
    };

    const handleCancel = () => {
      ctx.emit('close');
    };

    // 集群详情
    const clusterDetailLoading = ref(false);

    onMounted(async () => {
      handleGetOsImage();
      // 机型
      handleGetInstanceTypes();
      handleGetZoneList();
    });

    return {
      handleSetDefaultInstance,
      nodeConfigRef,
      formRef,
      basicFormRef,
      diskEnum,
      nodePoolConfig,
      nodePoolConfigRules,
      curInstanceItem,
      instanceTypesLoading,
      instanceList,
      pagination,
      pageChange,
      handleResetPage,
      pageSizeChange,
      instanceRowClass,
      handleCheckInstanceType,
      handleShowDataDisksChange,
      handleDeleteDiskData,
      handleAddDiskData,
      handleNext,
      handleCancel,
      validate,
      getNodePoolData,
      CPU,
      Mem,
      cpuList,
      memList,
      clusterDetailLoading,
      instanceTypesList,
      osImageLoading,
      osImageList,
      showDataDisks,
      // 可用区
      zoneList,
      zoneListLoading,
      handleSetZoneData,
    };
  },
});
</script>
<style lang="postcss" scoped>
.node-config-wrapper {
  overflow: auto;
}
.node-config {
    font-size: 14px;
    overflow: auto;
    >>> .group-text {
        line-height: 30px;
        background-color: #fafbfd;
    }
    >>> .bk-form-content {
        max-width: 650px;
        .bk-form-radio {
            white-space: nowrap;
        }
        .bk-form-checkbox {
            white-space: nowrap;
        }
    }
    >>> .table-row-enable {
        cursor: pointer;
    }
    >>> .table-row-disabled {
        cursor: not-allowed;
        color: #C4C6CC;
        .bk-checkbox-text {
            color: #C4C6CC;
        }
    }
    .max-width-130 {
        max-width: 130px;
    }
    .max-width-150 {
        max-width: 150px;
    }
    .min-width-80 {
        min-width: 80px;
    }
    .min-width-150 {
        min-width: 150px;
    }
    .prefix-select {
        display: flex;
        .prefix {
            display: inline-block;
            height: 32px;
            background: #F0F1F5;
            border: 1px solid #C4C6CC;
            border-radius: 2px 0 0 2px;
            border-radius: 2px 0 0 2px;
            padding: 0 8px;
            &.disabled {
              border-color: #dcdee5;
            }
        }
        >>> .bk-select {
            min-width: 110px;
            margin-left: -1px;
        }
    }
    .company {
        display: inline-block;
        min-width: 30px;
        padding: 0 4px 0 4px;
        height: 32px;
        line-height: 30px;
        border: 1px solid #C4C6CC;
        text-align: center;
        border-left: none;
        background-color: #fafbfd;
        &.disabled {
          border-color: #dcdee5;
        }
    }
    >>> .panel {
        background: #F5F7FA;
        padding: 16px 24px;
        margin-top: 2px;
        margin-bottom: 16px;
        position: relative;
        &-item {
            display: flex;
            align-items: center;
            height: 32px;
            .label {
                display: inline-block;
                width: 80px;
            }

        }
        &-delete {
            position: absolute;
            cursor: pointer;
            color: #979ba5;
            top: 0;
            right: 8px;
            &:hover {
                color: #3a84ff;
            }
            &.disabled {
                color: #C4C6CC;
                cursor: not-allowed;
            }
        }
        .bk-form-control {
            display: flex;
            width: auto;
            align-items: center;
        }
    }
    >>> .add-panel-btn {
        cursor: pointer;
        background: #fafbfd;
        border: 1px dashed #c4c6cc;
        border-radius: 2px;
        display: flex;
        align-items: center;
        justify-content: center;
        height: 32px;
        font-size: 12px;
        .bk-icon {
            font-size: 20px;
        }
        &:hover {
            border-color: #3a84ff;
            color: #3a84ff;
        }
        &.disabled {
            color: #C4C6CC;
            cursor: not-allowed;
            border-color: #C4C6CC;
        }
    }
    .bcs-icon-info-circle-shape {
        color: #979ba5;
    }
    .error-tips {
        color: red;
        font-size: 12px;
    }
}
</style>
