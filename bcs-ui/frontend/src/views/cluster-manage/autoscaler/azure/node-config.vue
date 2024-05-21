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
        <bk-form-item
          :label="$t('cluster.ca.nodePool.label.system')"
          error-display-type="normal"
          required
          :desc="$t('azureCA.tips.system')">
          <bcs-input disabled v-model="nodePoolConfig.launchTemplate.imageInfo.imageName"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.az.title')"
          :desc="$t('azureCA.tips.zone')"
          property="nodePoolConfig.autoScaling.zones"
          error-display-type="normal"
          required>
          <div class="flex items-center h-[32px]">
            <bcs-select
              class="flex-1"
              v-model="nodePoolConfig.autoScaling.zones"
              multiple
              searchable
              :loading="zoneListLoading"
              :disabled="isEdit"
              :clearable="false"
              selected-style="checkbox"
              @change="handleSpecifiedZoneChange">
              <bcs-option
                v-for="zone in zoneList"
                :key="zone.zone"
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
              <span v-bk-tooltips="$t('azureCA.tips.dataDisk')">
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
                    class="min-width-150 bg-[#fff] w-[150px]">
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
        <bk-form-item :label="$t('cluster.ca.nodePool.create.loginType.text')">
          <bk-radio-group v-model="loginType" @change="handleLoginTypeChange">
            <span class="inline-block">
              <bk-radio :disabled="isEdit" value="password">{{$t('cluster.ca.nodePool.create.loginType.password')}}</bk-radio>
              <bk-radio :disabled="isEdit" value="ssh">{{$t('cluster.ca.nodePool.create.loginType.ssh.text')}}</bk-radio>
            </span>
          </bk-radio-group>
          <div class="bg-[#F5F7FA] p-[16px] mt-[4px]">
            <template v-if="loginType === 'password'">
              <bk-form-item
                :label-width="100"
                :label="$t('azureCloud.label.loginUser')"
                property="nodePoolConfig.launchTemplate.initLoginUsername"
                error-display-type="normal">
                <bcs-input :disabled="isEdit" v-model="nodePoolConfig.launchTemplate.initLoginUsername"></bcs-input>
              </bk-form-item>
              <bk-form-item
                :label="$t('azureCA.password.set')"
                property="nodePoolConfig.launchTemplate.initLoginPassword"
                error-display-type="normal"
                :label-width="100">
                <bcs-input
                  type="password"
                  :disabled="isEdit"
                  v-model="nodePoolConfig.launchTemplate.initLoginPassword" />
              </bk-form-item>
              <bk-form-item
                :label="$t('azureCA.password.confirm')"
                property="confirmPassword"
                error-display-type="normal"
                :label-width="100">
                <bcs-input
                  type="password"
                  :disabled="isEdit"
                  v-model="confirmPassword" />
              </bk-form-item>
            </template>
            <template v-else-if="loginType === 'ssh'">
              <bk-form-item
                :label-width="100"
                :label="$t('azureCloud.label.loginUser')"
                property="nodePoolConfig.launchTemplate.initLoginUsername"
                error-display-type="normal">
                <bcs-input :disabled="isEdit" v-model="nodePoolConfig.launchTemplate.initLoginUsername"></bcs-input>
              </bk-form-item>
              <bk-form-item
                :label="$t('cluster.ca.nodePool.create.loginType.ssh.label.publicKey.text')"
                :label-width="100"
                :desc="$t('cluster.ca.nodePool.create.loginType.ssh.label.publicKey.desc')"
                property="nodePoolConfig.launchTemplate.keyPair.keyPublic"
                error-display-type="normal">
                <bcs-input
                  type="textarea"
                  :rows="4"
                  :disabled="isEdit"
                  v-model="nodePoolConfig.launchTemplate.keyPair.keyPublic">
                </bcs-input>
              </bk-form-item>
              <bk-form-item
                :label="$t('cluster.ca.nodePool.create.loginType.ssh.label.privateKey.text')"
                :label-width="100"
                :desc="$t('cluster.ca.nodePool.create.loginType.ssh.label.privateKey.desc')"
                property="nodePoolConfig.launchTemplate.keyPair.keySecret"
                error-display-type="normal">
                <bcs-input
                  type="textarea"
                  :rows="4"
                  :disabled="isEdit"
                  :placeholder="isEdit ? $t('cluster.ca.nodePool.create.loginType.ssh.placeholder.privateKey') : $t('generic.placeholder.input')"
                  v-model="nodePoolConfig.launchTemplate.keyPair.keySecret" />
              </bk-form-item>
            </template>
          </div>
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
            :disabled="isEdit"
            selected-style="checkbox">
            <bcs-option
              v-for="securityGroup in securityGroupsList"
              :key="securityGroup.securityGroupID"
              :id="securityGroup.securityGroupName"
              :name="securityGroup.securityGroupName">
              <bcs-checkbox
                :value="nodePoolConfig.launchTemplate.securityGroupIDs.includes(securityGroup.securityGroupID)">
                <div class="flex items-center text-[12px]">
                  <span>{{ securityGroup.securityGroupName }}</span>
                  <span class="text-[#C4C6CC]">({{ securityGroup.securityGroupID }})</span>
                </div>
              </bcs-checkbox>
            </bcs-option>
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.subnet.title')"
          property="nodePoolConfig.autoScaling.subnetIDs"
          error-display-type="normal"
          required>
          <bcs-table
            :data="filterSubnetsList"
            row-class-name="table-row-enable"
            v-bkloading="{ isLoading: subnetsLoading }"
            @row-click="handleCheckSubnets">
            <bcs-table-column :label="$t('cluster.ca.nodePool.create.subnet.label.subnetID')" min-width="150">
              <template #default="{ row }">
                <bk-checkbox
                  :value="nodePoolConfig.autoScaling.subnetIDs.includes(row.subnetID)">
                  <span class="bcs-ellipsis">{{row.subnetID}}</span>
                </bk-checkbox>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('cluster.ca.nodePool.create.subnet.label.subnetName')" prop="subnetName" show-overflow-tooltip></bcs-table-column>
            <bcs-table-column :label="$t('cluster.ca.nodePool.create.subnet.label.remainIp')" prop="availableIPAddressCount" align="right"></bcs-table-column>
          </bcs-table>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.containerRuntime.title')">
          <bk-radio-group v-model="clusterAdvanceSettings.containerRuntime">
            <bk-radio value="docker" disabled>docker</bk-radio>
            <bk-radio value="containerd" disabled>containerd</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.runtimeVersion.title')">
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
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import FormGroup from '@/views/cluster-manage/add/common/form-group.vue';
import Schema from '@/views/cluster-manage/autoscaler/resolve-schema';
import { useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';

export default defineComponent({
  components: { FormGroup },
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
    const defaultUser = 'azureuser';
    // 磁盘类型
    const diskEnum = ref([
      {
        id: 'Standard_LRS',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.Standard_LRS'),
      },
      {
        id: 'StandardSSD_LRS',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.StandardSSD_LRS'),
      },
      {
        id: 'StandardSSD_ZRS',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.StandardSSD_ZRS'),
      },
      {
        id: 'Premium_LRS',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.Premium_LRS'),
      },
      {
        id: 'PremiumV2_LRS',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.PremiumV2_LRS'),
      },
      {
        id: 'Premium_ZRS',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.Premium_ZRS'),
      },
    ]);
    const confirmPassword = ref(''); // 确认密码
    const nodePoolConfig = ref({
      nodeGroupID: defaultValues.value.nodeGroupID, // 编辑时
      name: defaultValues.value.name || '', // 节点名称
      autoScaling: {
        vpcID: '', // todo 放在basic-pool-info组件比较合适
        zones: defaultValues.value.autoScaling?.zones || [],
        subnetIDs: defaultValues.value.autoScaling.subnetIDs || [], // 支持子网
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
          internetChargeType: defaultValues.value.launchTemplate?.internetAccess?.internetChargeType, // 计费方式
          internetMaxBandwidth: defaultValues.value.launchTemplate?.internetAccess?.internetMaxBandwidth, // 购买带宽
          publicIPAssigned: defaultValues.value.launchTemplate?.internetAccess?.publicIPAssigned, // 分配免费公网IP
          bandwidthPackageId: defaultValues.value.launchTemplate?.internetAccess?.bandwidthPackageId, // 带宽包ID
        },
        initLoginUsername: defaultValues.value.launchTemplate?.initLoginUsername || defaultUser || '', // 登录用户
        // 密钥信息
        keyPair: {
          keyPublic: defaultValues.value.launchTemplate?.keyPair?.keyPublic || '',
          keySecret: defaultValues.value.launchTemplate?.keyPair?.keySecret || '',
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
        runtime: {
          containerRuntime: defaultValues.value.nodeTemplate?.runtime?.containerRuntime, // 运行时容器组件
          runtimeVersion: defaultValues.value.nodeTemplate?.runtime?.runtimeVersion, // 运行时版本
        },
      },
      extra: {
        provider: '', // 机型provider信息
      },
    });

    const nodePoolConfigRules = ref({
      // 密码校验
      'nodePoolConfig.launchTemplate.initLoginPassword': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => isEdit.value || nodePoolConfig.value.launchTemplate.initLoginPassword.length > 0,
        },
        {
          message: $i18n.t('cluster.ca.nodePool.create.validate.password'),
          trigger: 'blur',
          validator: () => isEdit.value
            || /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[^]{8,30}$/.test(nodePoolConfig.value.launchTemplate.initLoginPassword),
        },
      ],
      confirmPassword: [
        {
          message: $i18n.t('cluster.ca.nodePool.create.validate.passwordNotSame'),
          trigger: 'blur',
          validator: () => confirmPassword.value === nodePoolConfig.value.launchTemplate.initLoginPassword,
        },
      ],
      'nodePoolConfig.launchTemplate.initLoginUsername': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => isEdit.value || nodePoolConfig.value.launchTemplate.initLoginUsername.length > 0,
        },
        {
          message: $i18n.t('generic.validate.notRootName'),
          trigger: 'blur',
          validator: () => isEdit.value || nodePoolConfig.value.launchTemplate.initLoginUsername !== 'root',
        },
      ],
      // 密钥校验
      'nodePoolConfig.launchTemplate.keyPair.keyPublic': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => isEdit.value || nodePoolConfig.value.launchTemplate.keyPair.keyPublic.length > 0,
        },
      ],
      'nodePoolConfig.launchTemplate.keyPair.keySecret': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => isEdit.value || nodePoolConfig.value.launchTemplate.keyPair.keySecret.length > 0,
        },
      ],
      // 可用区校验
      'nodePoolConfig.autoScaling.zones': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => !!nodePoolConfig.value.autoScaling?.zones?.length,
        },
      ],
      // 安全组校验
      'nodePoolConfig.launchTemplate.securityGroupIDs': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => !!nodePoolConfig.value.launchTemplate.securityGroupIDs.length,
        },
      ],
      // 子网校验
      'nodePoolConfig.autoScaling.subnetIDs': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => !!nodePoolConfig.value.autoScaling.subnetIDs.length,
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
    const zoneList = ref<any[]>([]);
    const zoneListLoading = ref(false);
    const handleGetZoneList = async () => {
      zoneListLoading.value = true;
      zoneList.value = await cloudsZones({
        $cloudId: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
      });
      zoneListLoading.value = false;
    };
    // 不再需要切换可用区类型
    // 指定可用区
    const handleSpecifiedZoneChange = () => {
      nodePoolConfig.value.launchTemplate.instanceType = '';
    };

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
        $cloudID: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
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
      const zoneList: string[] = nodePoolConfig.value.autoScaling?.zones || [];
      const cacheInstanceMap = {};
      // 先过滤可用区, 再过滤同类型机型
      return instanceData.value
        .filter(instance => !zoneList.length || instance?.zones?.some(zone => zoneList.includes(zone)))
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
    watch(() => instanceTypesList.value.length, () => {
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
      nodePoolConfig.value.autoScaling.zones,
    ], () => {
      // 重置机型
      nodePoolConfig.value.launchTemplate.instanceType = '';
      handleSetDefaultInstance();
    });
    watch(curInstanceItem, () => {
      nodePoolConfig.value.extra.provider = curInstanceItem.value.provider;
    });
    watch(() => nodePoolConfig.value.launchTemplate.instanceType, () => {
      // 机型变更时重置子网数据
      nodePoolConfig.value.autoScaling.subnetIDs = [];
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
      const data = await $store.dispatch('clustermanager/cloudInstanceTypes', {
        $cloudID: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
        provider: nodePoolConfig.value.launchTemplate.imageInfo.imageType,
      });
      instanceData.value = data.sort((pre, current) => {
        // 排序要求 状态--SELL在前，其他在后；CPU升序，
        if (pre.status === 'SELL') {
          return current.status === 'SELL' ? pre.cpu - current.cpu : -1;
        }
        // 当 pre.status 不等于 'SELL'
        return current.status !== 'SELL' ? pre.cpu - current.cpu : 1;
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
    const defaultDiskItem = {
      // 默认为Premium_LRS(高级 SSD)
      diskType: 'Premium_LRS',
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

    // 免费分配公网IP
    watch(() => nodePoolConfig.value.launchTemplate.internetAccess.publicIPAssigned, (publicIPAssigned) => {
      nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth = publicIPAssigned ? '10' : '0';
    });

    // 登录方式
    const loginType = ref<'ssh'|'password'>(nodePoolConfig.value.launchTemplate.keyPair.keyPublic
      ? 'ssh'
      : 'password');
    const handleLoginTypeChange = (type) => {
      if (type === 'ssh') {
        nodePoolConfig.value.launchTemplate.initLoginUsername = '';
        nodePoolConfig.value.launchTemplate.initLoginPassword = '';
        confirmPassword.value = '';
      } else {
        nodePoolConfig.value.launchTemplate.initLoginUsername = defaultUser;
        nodePoolConfig.value.launchTemplate.keyPair.keyPublic = '';
        nodePoolConfig.value.launchTemplate.keyPair.keySecret = '';
      }
    };

    // 安全组
    const securityGroupsLoading = ref(false);
    const securityGroupsList = ref<any[]>([]);
    const handleGetCloudSecurityGroups = async () => {
      securityGroupsLoading.value = true;
      const data = await $store.dispatch('clustermanager/cloudSecurityGroups', {
        $cloudID: cluster.value.provider,
        resourceGroupName: cluster.value.extraInfo.nodeResourceGroup,
        accountID: cluster.value.cloudAccountID,
      });
      securityGroupsList.value = data.filter(item => item.securityGroupName !== 'default');
      securityGroupsLoading.value = false;
    };

    // VPC子网
    const subnetsLoading = ref(false);
    const subnetsList = ref<Array<{
      zone: string
      zoneName: string
      subnetID: string
    }>>([]);
    const filterSubnetsList = computed(() => subnetsList.value
      .sort((pre, current) => {
        const isPreDisabled = curInstanceItem.value.zones?.includes(pre.zone);
        const isCurrentDisabled = curInstanceItem.value.zones?.includes(current.zone);
        if (isPreDisabled && !isCurrentDisabled) return -1;

        if (!isPreDisabled && isCurrentDisabled) return 1;

        return 0;
      }));
    const subnetsRowClass = ({ row }) => {
      if (curInstanceItem.value.zones?.includes(row.zone)) {
        // return 'table-row-enable';
      }
      // return 'table-row-disabled';
      return 'table-row-enable';
    };
    const handleGetSubnets = async () => {
      subnetsLoading.value = true;
      subnetsList.value = await $store.dispatch('clustermanager/cloudSubnets', {
        $cloudID: cluster.value.provider,
        resourceGroupName: cluster.value.extraInfo.nodeResourceGroup,
        accountID: cluster.value.cloudAccountID,
        vpcID: cluster.value.vpcID,
      });
      subnetsLoading.value = false;
    };
    const handleCheckSubnets = (row) => {
      // if (!curInstanceItem.value.zones?.includes(row.zone) || isEdit.value) return;

      const index = nodePoolConfig.value.autoScaling.subnetIDs.indexOf(row.subnetID);
      if (index > -1) {
        nodePoolConfig.value.autoScaling.subnetIDs.splice(index, 1);
      } else {
        nodePoolConfig.value.autoScaling.subnetIDs.push(row.subnetID);
      }
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
      const basicFormValidate = await basicFormRef.value?.validate().catch(() => false);;
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
        if (!nodePoolConfig.value.autoScaling?.zones?.length) {
          nodeConfigRef.value.scrollTop = 0;
        } else {
          nodeConfigRef.value.scrollTop = nodeConfigRef.value.scrollHeight;
        }
      }
      // eslint-disable-next-line max-len
      const validateDataDiskSize = nodePoolConfig.value.nodeTemplate.dataDisks.every(item => item.diskSize % 10 === 0);
      if (!basicFormValidate || !result || !validateDataDiskSize) return false;

      return true;
    };
    const handleNext = async () => {
      // 校验错误滚动到第一个错误的位置
      const result = await validate();
      if (!result) {
        // 自动滚动到第一个错误的位置
        const errDom = document.getElementsByClassName('form-error-tip');
        const bcsErrDom = document.getElementsByClassName('error-tips');
        const innerErrDom = document.getElementsByClassName('is-error');
        const firstErrDom = innerErrDom[0] || errDom[0] || bcsErrDom[0];
        firstErrDom?.scrollIntoView({
          block: 'center',
          behavior: 'smooth',
        });
        return;
      }
      ctx.emit('next', getNodePoolData());
    };

    const handleCancel = () => {
      $router.back();
    };

    const getSchemaByProp = props => Schema.getSchemaByProp(schema.value, props);

    // 集群详情
    const clusterDetailLoading = ref(false);
    const { clusterData, clusterOS, clusterAdvanceSettings, getClusterDetail } = useClusterInfo();
    const handleGetClusterDetail = async () => {
      clusterDetailLoading.value = true;
      await getClusterDetail(cluster.value?.clusterID, true);
      nodePoolConfig.value.autoScaling.vpcID = clusterData.value.vpcID;
      clusterDetailLoading.value = false;
    };

    onMounted(async () => {
      await handleGetClusterDetail(); // 优选获取集群详情信息
      handleGetOsImage();
      // 机型
      handleGetInstanceTypes();
      // 安全组
      handleGetCloudSecurityGroups();
      // 可用区
      handleGetZoneList();
      // 子网
      handleGetSubnets();
    });

    return {
      securityGroupsLoading,
      securityGroupsList,
      // 子网
      subnetsLoading,
      subnetsList,
      filterSubnetsList,
      clusterOS,
      zoneListLoading,
      zoneList,
      handleSetDefaultInstance,
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
      handleResetPage,
      pageSizeChange,
      loginType,
      instanceRowClass,
      handleCheckInstanceType,
      handleShowDataDisksChange,
      handleDeleteDiskData,
      handleAddDiskData,
      handleLoginTypeChange,
      handleNext,
      handleCancel,
      getSchemaByProp,
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
      clusterAdvanceSettings,
      handleSpecifiedZoneChange,
      handleGetCloudSecurityGroups,
      // 子网
      subnetsRowClass,
      handleCheckSubnets,
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
