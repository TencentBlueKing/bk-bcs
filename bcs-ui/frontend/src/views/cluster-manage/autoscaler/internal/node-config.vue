<!-- eslint-disable max-len -->
<template>
  <div class="p-[24px] node-config-wrapper" ref="nodeConfigRef">
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
        <bk-form-item :label="$t('cluster.ca.nodePool.create.imageProvider.title')" :desc="$t('cluster.ca.nodePool.create.imageProvider.desc')">
          <bk-radio-group :value="extraInfo.IMAGE_PROVIDER">
            <bk-radio value="PUBLIC_IMAGE" disabled>{{$t('deploy.image.publicImage')}}</bk-radio>
            <bk-radio value="PRIVATE_IMAGE" disabled>{{$t('cluster.ca.nodePool.create.imageProvider.private_image')}}</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.label.system')">
          <bcs-input disabled :value="clusterOS"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.az.title')"
          :desc="$t('cluster.ca.nodePool.create.az.desc')"
          property="nodePoolConfig.autoScaling.zones">
          <div class="flex items-center h-[32px]">
            <bk-radio-group
              :value="isSpecifiedZoneList"
              class="w-[auto]"
              @change="handleZoneChange">
              <bk-radio :value="false" :disabled="isEdit">
                <TextTips :text="$t('cluster.ca.nodePool.create.az.random')" />
              </bk-radio>
              <bk-radio :value="true" :disabled="isEdit">
                <TextTips :text="$t('cluster.ca.nodePool.create.az.selected')" />
              </bk-radio>
            </bk-radio-group>
            <bcs-select
              class="flex-1 ml-[10px]"
              v-if="isSpecifiedZoneList"
              v-model="nodePoolConfig.autoScaling.zones"
              multiple
              searchable
              :loading="zoneListLoading"
              :disabled="isEdit"
              :clearable="false"
              selected-style="checkbox"
              @change="handleSetDefaultInstance">
              <bcs-option
                v-for="zone in zoneList"
                :key="zone.zoneID"
                :id="zone.zone"
                :name="zone.zoneName" />
            </bcs-select>
          </div>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.instanceTypeConfig.title')"
          :desc="isEdit ? $t('cluster.ca.nodePool.create.instanceTypeConfig.desc') : ''">
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
              disabled
              :value="true"
              @change="handleShowDataDisksChange">
              {{$t('cluster.ca.nodePool.create.instanceTypeConfig.label.purchaseDataDisk')}}
            </bk-checkbox>
          </div>
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
            <div class="panel-item mt10">
              <bk-checkbox
                :disabled="(isEdit || index === 0)"
                v-model="disk.autoFormatAndMount">
                {{$t('cluster.ca.nodePool.create.instanceTypeConfig.label.mountPath')}}
              </bk-checkbox>
              <template v-if="disk.autoFormatAndMount">
                <bcs-select
                  class="min-width-80 ml10 bg-[#fff]"
                  :disabled="(isEdit || index === 0)"
                  v-model="disk.fileSystem"
                  :clearable="false">
                  <bcs-option
                    v-for="fileSystem in getSchemaByProp('launchTemplate.dataDisks.fileSystem').enum"
                    :key="fileSystem"
                    :id="fileSystem"
                    :name="fileSystem">
                  </bcs-option>
                </bcs-select>
                <bcs-input class="ml10" :disabled="(isEdit || index === 0)" v-model="disk.mountTarget"></bcs-input>
              </template>
            </div>
            <p class="error-tips" v-if="showRepeatMountTarget(index)">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.validate.repeatPath')}}</p>
            <span :class="['panel-delete', { disabled: isEdit || index === 0 }]" @click="handleDeleteDiskData(index)">
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
          <span class="inline-flex mt15" v-bk-tooltips="$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.tips')">
            <bk-checkbox
              disabled
              v-model="nodePoolConfig.launchTemplate.internetAccess.publicIPAssigned">
              {{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.text')}}
            </bk-checkbox>
          </span>
          <div class="panel" v-if="nodePoolConfig.launchTemplate.internetAccess.publicIPAssigned">
            <div class="panel-item">
              <label class="label">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.chargeMode.text')}}</label>
              <bk-radio-group v-model="nodePoolConfig.launchTemplate.internetAccess.internetChargeType">
                <bk-radio :disabled="isEdit" value="TRAFFIC_POSTPAID_BY_HOUR">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.chargeMode.traffic_postpaid_by_hour')}}</bk-radio>
                <bk-radio :disabled="isEdit" value="BANDWIDTH_PREPAID">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.chargeMode.bandwidth_prepaid')}}</bk-radio>
              </bk-radio-group>
            </div>
            <div class="panel-item mt10">
              <label class="label">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.maxBandWidth')}}</label>
              <bk-input
                class="max-width-150"
                type="number"
                :disabled="isEdit"
                v-model="nodePoolConfig.launchTemplate.internetAccess.internetMaxBandwidth">
              </bk-input>
              <span :class="['company', { disabled: isEdit }]">Mbps</span>
            </div>
          </div>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.password.set')"
          property="nodePoolConfig.launchTemplate.initLoginPassword"
          error-display-type="normal">
          <bcs-input
            type="password"
            disabled
            :placeholder="$t('cluster.ca.nodePool.create.password.placeholder')"
            v-model="nodePoolConfig.launchTemplate.initLoginPassword"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.password.confirm')"
          property="confirmPassword"
          error-display-type="normal">
          <bcs-input
            type="password"
            :placeholder="$t('cluster.ca.nodePool.create.password.placeholder')"
            disabled
            v-model="confirmPassword"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.securityGroup')"
          property="nodePoolConfig.launchTemplate.securityGroupIDs"
          error-display-type="normal"
          required>
          <bcs-select
            :loading="securityGroupsLoading"
            v-model="nodePoolConfig.launchTemplate.securityGroupIDs"
            multiple
            searchable
            class="bg-[#fff]"
            selected-style="checkbox"
            disabled>
            <bcs-option
              v-for="securityGroup in securityGroupsList"
              :key="securityGroup.securityGroupID"
              :id="securityGroup.securityGroupID"
              :name="securityGroup.securityGroupName">
            </bcs-option>
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.subnet.title')"
          :desc="$t('cluster.ca.nodePool.create.subnet.desc')">
          <bcs-input
            :placeholder="$t('cluster.ca.nodePool.create.subnet.placeholder')"
            disabled></bcs-input>
        </bk-form-item>
        <bk-form-item :label="$t('dashboard.workload.container.dataDir')" :desc="$t('cluster.ca.nodePool.create.dockerGraphPath.desc')">
          <bcs-input disabled v-model="nodePoolConfig.nodeTemplate.dockerGraphPath"></bcs-input>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.containerRuntime.title')" :desc="$t('cluster.ca.nodePool.create.containerRuntime.desc')">
          <bk-radio-group v-model="clusterAdvanceSettings.containerRuntime">
            <bk-radio value="docker" disabled>docker</bk-radio>
            <bk-radio value="containerd" disabled>containerd</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.runtimeVersion.title')" :desc="$t('cluster.ca.nodePool.create.runtimeVersion.desc')">
          <bcs-input disabled v-model="clusterAdvanceSettings.runtimeVersion"></bcs-input>
        </bk-form-item>
        <div class="bcs-fixed-footer" v-if="!isEdit">
          <bcs-button theme="primary" @click="handleNext">{{$t('generic.button.next')}}</bcs-button>
          <bcs-button class="ml10" @click="handleCancel">{{$t('generic.button.cancel')}}</bcs-button>
        </div>
      </bk-form>
    </FormGroup>
  </div>
</template>
<script lang="ts">
import { sortBy } from 'lodash';
import { computed, defineComponent, onMounted, ref, toRefs, watch } from 'vue';

import { cloudsZones } from '@/api/modules/cluster-manager';
import TextTips from '@/components/layout/TextTips.vue';
import { useProject } from '@/composables/use-app';
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import FormGroup from '@/views/cluster-manage/add/common/form-group.vue';
import Schema from '@/views/cluster-manage/autoscaler/resolve-schema';
import { useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';

export default defineComponent({
  components: { FormGroup, TextTips },
  props: {
    schema: {
      type: Object,
      default: () => ({}),
    },
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
    const { defaultValues, cluster, isEdit, schema } = toRefs(props);
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
    const confirmPassword = ref(''); // 确认密码
    const nodePoolConfig = ref({
      nodeGroupID: defaultValues.value.nodeGroupID, // 编辑时
      name: defaultValues.value.name || '', // 节点名称
      autoScaling: {
        vpcID: '', // todo 放在basic-pool-info组件比较合适
        zones: defaultValues.value.autoScaling?.zones || [],
      },
      launchTemplate: {
        imageInfo: {
          imageID: defaultValues.value.launchTemplate?.imageInfo?.imageID, // 镜像ID
          imageName: defaultValues.value.launchTemplate?.imageInfo?.imageName, // 镜像名称
        },
        CPU: '',
        Mem: '',
        instanceType: defaultValues.value.launchTemplate?.instanceType, // 机型信息
        systemDisk: {
          diskType: defaultValues.value.launchTemplate?.systemDisk?.diskType, // 系统盘类型
          diskSize: defaultValues.value.launchTemplate?.systemDisk?.diskSize, // 系统盘大小
        },
        internetAccess: {
          internetChargeType: defaultValues.value.launchTemplate?.internetAccess?.internetChargeType, // 计费方式
          internetMaxBandwidth: defaultValues.value.launchTemplate?.internetAccess?.internetMaxBandwidth, // 购买带宽
          publicIPAssigned: defaultValues.value.launchTemplate?.internetAccess?.publicIPAssigned, // 分配免费公网IP
        },
        initLoginPassword: defaultValues.value.launchTemplate?.initLoginPassword, // 密码
        securityGroupIDs: defaultValues.value.launchTemplate?.securityGroupIDs || [], // 安全组
        dataDisks: defaultValues.value.launchTemplate?.dataDisks || [], // 数据盘
        // 默认值
        isSecurityService: defaultValues.value.launchTemplate?.isSecurityService || true,
        isMonitorService: defaultValues.value.launchTemplate?.isMonitorService || true,
      },
      nodeTemplate: {
        dataDisks: defaultValues.value.nodeTemplate?.dataDisks || [],
        dockerGraphPath: defaultValues.value.nodeTemplate?.dockerGraphPath, // 容器目录
        runtime: {
          containerRuntime: defaultValues.value.nodeTemplate?.runtime?.containerRuntime || 'docker', // 运行时容器组件
          runtimeVersion: defaultValues.value.nodeTemplate?.runtime?.runtimeVersion || '19.3', // 运行时版本
        },
      },
      extra: {
        provider: '', // 机型provider信息
      },
    });

    const nodePoolConfigRules = ref({
      // 安全组校验
      'nodePoolConfig.launchTemplate.securityGroupIDs': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => !!nodePoolConfig.value.launchTemplate.securityGroupIDs.length,
        },
      ],
      'nodePoolConfig.autoScaling.zones': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => !isSpecifiedZoneList.value || !!nodePoolConfig.value.autoScaling?.zones?.length,
        },
      ],
    });
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

    // 可用区
    const showZoneList = ref(false);
    const zoneList = ref<any[]>([]);
    const zoneListLoading = ref(false);
    const isSpecifiedZoneList = computed(() => (!!nodePoolConfig.value.autoScaling?.zones?.length)
    || showZoneList.value);
    const handleGetZoneList = async () => {
      zoneListLoading.value = true;
      zoneList.value = await cloudsZones({
        $cloudId: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
      });
      zoneListLoading.value = false;
    };
    const handleZoneChange = (v: boolean) => {
      showZoneList.value = v;
      nodePoolConfig.value.autoScaling.zones = [];
    };

    // 机型
    const instanceTypesLoading = ref(false);
    const instanceData = ref<any[]>([]);
    const instanceTypesList = computed(() => {
      const zoneList: string[] = nodePoolConfig.value.autoScaling?.zones || [];
      const cacheInstanceMap = {};
      if (!zoneList.length) return instanceData.value
        .filter((instance) => {
        // todo 简单过滤同类型机型
          if (!cacheInstanceMap[instance.nodeType]) {
            cacheInstanceMap[instance.nodeType] = true;
            return true;
          }
          return false;
        })
        .filter(instance => (!CPU.value || instance.cpu === CPU.value)
        && (!Mem.value || instance.memory === Mem.value));;
      // 先过滤可用区, 再过滤同类型机型
      return instanceData.value
        .filter(instance => instance?.zones?.some(zone => zoneList.includes(zone)))
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
      // // 获取机型
      // handleGetInstanceTypes();
    });
    watch(curInstanceItem, () => {
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
    const { projectID } = useProject();
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
      instanceTypesLoading.value = true;
      // const cpu = nodePoolConfig.value.launchTemplate.CPU || undefined;
      // const memory =  nodePoolConfig.value.launchTemplate.Mem || undefined;
      const data = await $store.dispatch('clustermanager/cloudInstanceTypes', {
        $cloudID: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
        provider: 'yunti', // todo self
        // cpu,
        // memory,
        projectID: projectID.value,
        version: 'v2',
        // bizID: curProject.value?.businessID,
      });
      instanceData.value = data.sort((pre, current) => pre.cpu - current.cpu);
      handleSetDefaultInstance();
      instanceTypesLoading.value = false;
    };
    const {
      pagination,
      curPageData: instanceList,
      pageChange,
      pageSizeChange,
    } = usePage(instanceTypesList);

    // 数据盘
    const defaultDiskItem = {
      diskType: 'CLOUD_PREMIUM',
      diskSize: '100',
      fileSystem: 'ext4',
      autoFormatAndMount: true,
      mountTarget: '/data',
    };
    const handleShowDataDisksChange = (show) => {
      nodePoolConfig.value.nodeTemplate.dataDisks = show
        ? [JSON.parse(JSON.stringify(defaultDiskItem))] : [];
    };
    const handleDeleteDiskData = (index) => {
      if (isEdit.value || index === 0) return;
      nodePoolConfig.value.nodeTemplate.dataDisks.splice(index, 1);
    };
    const handleAddDiskData = () => {
      if (isEdit.value || nodePoolConfig.value.nodeTemplate.dataDisks.length > 4) return;
      nodePoolConfig.value.nodeTemplate.dataDisks.push(JSON.parse(JSON.stringify(defaultDiskItem)));
    };

    // 免费分配公网IP
    watch(() => nodePoolConfig.value.launchTemplate.internetAccess.publicIPAssigned, (publicIPAssigned) => {
      nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth = publicIPAssigned ? '1' : '0';
    });

    // 安全组
    const securityGroupsLoading = ref(false);
    const securityGroupsList = ref<any[]>([]);
    const handleGetCloudSecurityGroups = async () => {
      securityGroupsLoading.value = true;
      const data = await $store.dispatch('clustermanager/cloudSecurityGroups', {
        $cloudID: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
      });
      securityGroupsList.value = data.filter(item => item.securityGroupName !== 'default');
      // 默认安全组先写死（没有唯一ID标识，只能通过名称，很low）
      nodePoolConfig.value.launchTemplate.securityGroupIDs = [
        securityGroupsList.value.find(item => item.securityGroupName === '云梯默认安全组')?.securityGroupID,
      ];
      securityGroupsLoading.value = false;
    };

    // 运行版本
    watch(() => nodePoolConfig.value.nodeTemplate.runtime.containerRuntime, (runtime) => {
      if (runtime === 'docker') {
        nodePoolConfig.value.nodeTemplate.runtime.runtimeVersion = '19.3';
      } else {
        nodePoolConfig.value.nodeTemplate.runtime.runtimeVersion = '1.4.3';
      }
    });
    const versionList = computed(() => (nodePoolConfig.value.nodeTemplate.runtime.containerRuntime === 'docker'
      ? ['18.6', '19.3']
      : ['1.3.4', '1.4.3']));

    // 操作
    const getNodePoolData = () => {
      // 系统盘、数据盘、宽度大小要转换为字符串类型
      // eslint-disable-next-line max-len
      nodePoolConfig.value.launchTemplate.systemDisk.diskSize = String(nodePoolConfig.value.launchTemplate.systemDisk.diskSize);
      nodePoolConfig.value.nodeTemplate.dataDisks = nodePoolConfig.value.nodeTemplate.dataDisks.map(item => ({
        ...item,
        diskSize: String(item.diskSize),
      }));
      // eslint-disable-next-line max-len
      nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth = String(nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth);
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
      const basicFormValidate = await basicFormRef.value?.validate().catch(() => false);
      if (!basicFormValidate && nodeConfigRef.value) {
        nodeConfigRef.value.scrollTop = 0;
        return false;
      }
      // 校验机型
      if (!nodePoolConfig.value.launchTemplate.instanceType) {
        nodeConfigRef.value.scrollTop = 20;
        return false;
      }
      const result = await formRef.value?.validate().catch(() => false);
      if (!result && nodeConfigRef.value) {
        if (isSpecifiedZoneList.value && !nodePoolConfig.value.autoScaling?.zones?.length) {
          nodeConfigRef.value.scrollTop = 0;
        } else {
          nodeConfigRef.value.scrollTop = nodeConfigRef.value.offsetHeight;
        }
      }
      // eslint-disable-next-line max-len
      const validateDataDiskSize = nodePoolConfig.value.nodeTemplate.dataDisks.every(item => item.diskSize % 10 === 0);
      const mountTargetList = nodePoolConfig.value.nodeTemplate.dataDisks.map(item => item.mountTarget);
      const validateDataDiskMountTarget = new Set(mountTargetList).size === mountTargetList.length;
      if (!basicFormValidate || !result || !validateDataDiskSize || !validateDataDiskMountTarget) return false;

      return true;
    };
    const handleNext = async () => {
      if (!await validate()) return;

      ctx.emit('next', getNodePoolData());
    };

    const handleCancel = () => {
      $router.back();
    };

    const getSchemaByProp = props => Schema.getSchemaByProp(schema.value, props);

    const showRepeatMountTarget = (index) => {
      const disk = nodePoolConfig.value.nodeTemplate.dataDisks[index];
      return disk.autoFormatAndMount
                    && disk.mountTarget
                    && nodePoolConfig.value.nodeTemplate.dataDisks
                      .filter((item, i) => i !== index && item.autoFormatAndMount)
                      .some(item => item.mountTarget === disk.mountTarget);
    };

    // 集群详情
    const clusterDetailLoading = ref(false);
    const { clusterData, clusterOS, clusterAdvanceSettings, extraInfo, getClusterDetail } = useClusterInfo();
    const handleGetClusterDetail = async () => {
      clusterDetailLoading.value = true;
      await getClusterDetail(cluster.value?.clusterID, true);
      nodePoolConfig.value.autoScaling.vpcID = clusterData.value.vpcID;
      clusterDetailLoading.value = false;
    };

    onMounted(async () => {
      if (!nodePoolConfig.value.nodeTemplate.dataDisks.length && !isEdit.value) {
        // 必须有一块数据盘
        handleAddDiskData();
      }
      await handleGetClusterDetail();
      handleGetInstanceTypes();
      handleGetCloudSecurityGroups();
      handleGetZoneList();
    });

    return {
      isSpecifiedZoneList,
      zoneListLoading,
      zoneList,
      handleZoneChange,
      handleSetDefaultInstance,
      extraInfo,
      clusterAdvanceSettings,
      clusterOS,
      versionList,
      nodeConfigRef,
      formRef,
      basicFormRef,
      diskEnum,
      confirmPassword,
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
      handleShowDataDisksChange,
      handleDeleteDiskData,
      handleAddDiskData,
      securityGroupsLoading,
      securityGroupsList,
      handleNext,
      handleCancel,
      getSchemaByProp,
      showRepeatMountTarget,
      validate,
      getNodePoolData,
      CPU,
      Mem,
      cpuList,
      memList,
      clusterDetailLoading,
      instanceTypesList,
    };
  },
});
</script>
<style lang="postcss" scoped>
.node-config-wrapper {
  max-height: calc(100vh - 164px);
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
