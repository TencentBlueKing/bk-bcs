<!-- eslint-disable max-len -->
<template>
  <div class="h-full flex flex-col">
    <div class="p-[24px] node-config-wrapper flex-1" ref="nodeConfigRef">
      <FormGroup :title="$t('cluster.ca.button.createNodePool')" :allow-toggle="false" class="mb-[16px]">
        <bk-form :model="nodePoolConfig" :rules="basicFormRules" ref="basicFormRef">
          <bk-form-item class="max-w-[674px]" :label="$t('cluster.ca.nodePool.label.name')" property="name" required>
            <bk-input v-model="nodePoolConfig.name"></bk-input>
            <p class="text-[#979BA5] leading-4 mt-[4px]">{{ $t('cluster.ca.nodePool.validate.name') }}</p>
          </bk-form-item>
        </bk-form>
      </FormGroup>
      <FormGroup :title="$t('cluster.ca.nodePool.title.nodeConfig')" :allow-toggle="false">
        <bk-form class="node-config" :model="nodePoolConfig" :rules="nodePoolConfigRules" ref="formRef">
          <!-- <bk-form-item :label="$t('cluster.ca.nodePool.create.imageProvider.title')" :desc="$t('cluster.ca.nodePool.create.imageProvider.desc')">
            <bk-radio-group :value="extraInfo.IMAGE_PROVIDER">
              <bk-radio value="PUBLIC_IMAGE" disabled>{{$t('deploy.image.publicImage')}}</bk-radio>
              <bk-radio value="PRIVATE_IMAGE" disabled>{{$t('cluster.ca.nodePool.create.imageProvider.private_image')}}</bk-radio>
            </bk-radio-group>
          </bk-form-item> -->
          <bk-form-item :label="$t('cluster.ca.nodePool.label.system')">
            <bcs-input disabled :value="clusterOS"></bcs-input>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.ca.nodePool.create.instanceTypeConfig.title')"
            :desc="instanceDesc">
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
                      <span class="bcs-ellipsis">{{row.typeName || row.nodeType}}</span>
                    </bcs-radio>
                  </span>
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('generic.label.specifications')" min-width="160" show-overflow-tooltip prop="nodeType"></bcs-table-column>
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
            </bcs-table>
            <p
              class="text-[12px] text-[#ea3636] error-tips"
              v-if="!nodePoolConfig.launchTemplate.instanceType">
              {{ $t('generic.validate.required') }}
            </p>
          </bk-form-item>
          <bk-form-item :label="$t('dashboard.workload.container.dataDir')" :desc="$t('cluster.ca.nodePool.create.dockerGraphPathDefault.desc')">
            <bcs-input disabled v-model="nodePoolConfig.nodeTemplate.dockerGraphPath"></bcs-input>
          </bk-form-item>
          <bk-form-item :label="$t('cluster.ca.nodePool.create.containerRuntime.title')" :desc="$t('cluster.ca.nodePool.create.containerRuntimeDefault.desc')">
            <bk-radio-group v-model="clusterAdvanceSettings.containerRuntime">
              <bk-radio value="docker" disabled>docker</bk-radio>
              <bk-radio value="containerd" disabled>containerd</bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item :label="$t('cluster.ca.nodePool.create.runtimeVersion.title')" :desc="$t('cluster.ca.nodePool.create.runtimeVersionDefault.desc')">
            <bcs-input disabled v-model="clusterAdvanceSettings.runtimeVersion"></bcs-input>
          </bk-form-item>
        </bk-form>
      </FormGroup>
    </div>
    <div class="bcs-border-top z-10 flex items-center sticky bottom-0 bg-[#fff] h-[60px] px-[24px]" v-if="!isEdit">
      <bcs-button theme="primary" @click="handleNext">{{$t('generic.button.next')}}</bcs-button>
      <bcs-button class="ml10" @click="handleCancel">{{$t('generic.button.cancel')}}</bcs-button>
    </div>
  </div>
</template>
<script lang="ts">
import { sortBy } from 'lodash';
import { computed, defineComponent, onMounted, ref, toRefs, watch } from 'vue';

import FormGroup from '@/components/form-group.vue';
import { useProject } from '@/composables/use-app';
import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import { useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';

export default defineComponent({
  components: { FormGroup },
  props: {
    defaultValues: {
      type: Object,
      default: () => ({}),
    },
    cluster: {
      type: Object,
      default: () => ({}),
      required: true,
    },
    isEdit: {
      type: Boolean,
      default: false,
    },
    saveLoading: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { defaultValues, cluster, isEdit } = toRefs(props);
    const nodeConfigRef = ref<any>(null);
    const formRef = ref<any>(null);
    const basicFormRef = ref<any>(null);
    // 磁盘类型
    const diskEnum = ref([
      {
        id: 'CLOUD_PREMIUM',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.premium'),
      },
      {
        id: 'CLOUD_SSD',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.ssd'),
      },
      {
        id: 'CLOUD_HSSD',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.hssd'),
      },
    ]);

    const nodePoolConfig = ref({
      nodeGroupID: defaultValues.value.nodeGroupID, // 编辑时
      name: defaultValues.value.name || '', // 节点名称
      autoScaling: {
        vpcID: '', // todo 放在basic-pool-info组件比较合适
        // zones: defaultValues.value.autoScaling?.zones || [],
      },
      launchTemplate: {
        imageInfo: {},
        CPU: '',
        Mem: '',
        instanceType: defaultValues.value.launchTemplate?.instanceType, // 机型信息
        systemDisk: {},
        internetAccess: {},
        // initLoginPassword: defaultValues.value.launchTemplate?.initLoginPassword, // 密码
        // securityGroupIDs: defaultValues.value.launchTemplate?.securityGroupIDs || [], // 安全组
        dataDisks: [], // 数据盘
        // 默认值
        isSecurityService: defaultValues.value.launchTemplate?.isSecurityService || true,
        isMonitorService: defaultValues.value.launchTemplate?.isMonitorService || true,
      },
      nodeTemplate: {
        dataDisks: [],
        dockerGraphPath: defaultValues.value.nodeTemplate?.dockerGraphPath, // 容器目录
        runtime: {
          containerRuntime: '', // 运行时容器组件
          runtimeVersion: '', // 运行时版本
        },
      },
      extra: {
        poolID: '',
        provider: '', // 机型provider信息
      },
      region: defaultValues.value.region, // 地域
    });

    const nodePoolConfigRules = ref({});
    const basicFormRules = ref({
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

    // 机型
    const instanceDesc = computed(() => $i18n.t('tkeCa.tips.selfPoolInstanceDesc'));
    const instanceTypesLoading = ref(false);
    const instanceData = ref<any[]>([]);
    const instanceTypesList = computed(() => instanceData.value.reduce<any[]>((pre, instance) => {
      const exitIndex = pre.findIndex(item => item.nodeType === instance.nodeType);
      if (exitIndex > -1) { // 过滤同类型机型
        if (pre[exitIndex]?.status === 'SOLD_OUT') {
          // 告罄
          pre.splice(exitIndex, 1, {
            ...instance,
            resourcePoolID: [instance.resourcePoolID],
          });
        } else if (!pre[exitIndex]?.resourcePoolID?.includes(instance.resourcePoolID)) {
          // 合并 resourcePoolID 字段（选择同类型机型时需要给后端）
          pre[exitIndex]?.resourcePoolID?.push(instance.resourcePoolID);
        }
      } else { // 不存在同类型
        pre.push({
          ...instance,
          resourcePoolID: [instance.resourcePoolID], // 改变 resourcePoolID 数据结构，同类型机型需要把每个 resourcePoolID 给后端
        });
      }
      return pre;
    }, [])
      .filter(instance => (!CPU.value || instance.cpu === CPU.value)
      && (!Mem.value || instance.memory === Mem.value)));
    // eslint-disable-next-line max-len
    const curInstanceItem = computed(() => instanceTypesList.value.find(instance => instance.nodeType === nodePoolConfig.value.launchTemplate.instanceType) || {});
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
      nodePoolConfig.value.launchTemplate.instanceType = instanceTypesList.value
        .find(instance => instance.status === 'SELL')?.nodeType;
      // // 获取机型
      // handleGetInstanceTypes();
    });
    watch(curInstanceItem, () => {
      // 资源池
      nodePoolConfig.value.extra.provider = curInstanceItem.value.provider;
      // nodePoolConfig.value.autoScaling.zones = [];
    });
    const instanceRowClass = ({ row }) => {
      // SELL 表示售卖，SOLD_OUT 表示售罄
      if (row.status === 'SELL' || isEdit.value) {
        return 'table-row-enable';
      }
      return 'table-row-disabled';
    };
    const handleCheckInstanceType = (row) => {
      if (row.status === 'SOLD_OUT' || isEdit.value) return;
      nodePoolConfig.value.launchTemplate.instanceType = row.nodeType;
    };
    const { projectID, curProject } = useProject();
    const handleGetInstanceTypes = async () => {
      instanceTypesLoading.value = true;
      const data = await $store.dispatch('clustermanager/cloudInstanceTypes', {
        $cloudID: cluster.value.provider,
        region: '',
        accountID: cluster.value.cloudAccountID,
        projectID: projectID.value,
        version: 'v2',
        provider: 'self',
        bizID: curProject.value?.businessID,
      });
      instanceData.value = data.sort((pre, current) => pre.cpu - current.cpu);
      instanceTypesLoading.value = false;
    };
    const {
      pagination,
      curPageData: instanceList,
      pageChange,
      pageSizeChange,
    } = usePage(instanceTypesList);

    // 运行版本
    // watch(() => nodePoolConfig.value.nodeTemplate.runtime.containerRuntime, (runtime) => {
    //   if (runtime === 'docker') {
    //     nodePoolConfig.value.nodeTemplate.runtime.runtimeVersion = '19.3';
    //   } else {
    //     nodePoolConfig.value.nodeTemplate.runtime.runtimeVersion = '1.4.3';
    //   }
    // });
    const versionList = computed(() => (nodePoolConfig.value.nodeTemplate.runtime.containerRuntime === 'docker'
      ? ['18.6', '19.3']
      : ['1.3.4', '1.4.3']));

    // 操作
    const getNodePoolData = () => {
      const resourcePoolIDList = curInstanceItem.value?.resourcePoolID;
      nodePoolConfig.value.extra.poolID = Array.isArray(resourcePoolIDList) ? resourcePoolIDList?.join(',') : resourcePoolIDList;
      nodePoolConfig.value.region = curInstanceItem.value.region;

      // CPU和mem信息从机型获取
      nodePoolConfig.value.launchTemplate.CPU = curInstanceItem.value.cpu;
      nodePoolConfig.value.launchTemplate.Mem = curInstanceItem.value.memory;
      return nodePoolConfig.value;
    };
    const validate = async () => {
      const basicFormValidate = await basicFormRef.value?.validate().catch(() => false);

      const result = await formRef.value?.validate().catch(() => false);
      // eslint-disable-next-line max-len
      if (!basicFormValidate
        || !nodePoolConfig.value.launchTemplate.instanceType // 校验机型
        || !result) return false;

      return true;
    };
    const { focusOnErrorField } = useFocusOnErrorField();
    const handleNext = async () => {
      // 校验错误滚动到第一个错误的位置
      const result = await validate();
      if (!result) {
        focusOnErrorField();
        return;
      }
      ctx.emit('next', getNodePoolData());
    };

    const handleCancel = () => {
      $router.back();
    };

    // 集群详情
    const clusterDetailLoading = ref(false);
    const { clusterOS, clusterAdvanceSettings, extraInfo, getClusterDetail } = useClusterInfo();
    const handleGetClusterDetail = async () => {
      clusterDetailLoading.value = true;
      await getClusterDetail(cluster.value?.clusterID, true);
      clusterDetailLoading.value = false;
    };
    const runtimeConf = computed(() => clusterAdvanceSettings.value);
    watch(runtimeConf, () => {
      nodePoolConfig.value.nodeTemplate.runtime.containerRuntime = clusterAdvanceSettings.value.containerRuntime;
      nodePoolConfig.value.nodeTemplate.runtime.runtimeVersion = clusterAdvanceSettings.value.runtimeVersion;
    });

    onMounted(async () => {
      await handleGetClusterDetail();
      handleGetInstanceTypes();
    });

    return {
      extraInfo,
      clusterAdvanceSettings,
      clusterOS,
      versionList,
      nodeConfigRef,
      formRef,
      basicFormRef,
      diskEnum,
      nodePoolConfig,
      nodePoolConfigRules,
      basicFormRules,
      curInstanceItem,
      instanceTypesLoading,
      instanceList,
      pagination,
      pageChange,
      pageSizeChange,
      instanceRowClass,
      handleCheckInstanceType,
      handleNext,
      handleCancel,
      validate,
      focusOnErrorField,
      getNodePoolData,
      CPU,
      Mem,
      cpuList,
      memList,
      clusterDetailLoading,
      instanceTypesList,
      instanceDesc,
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
