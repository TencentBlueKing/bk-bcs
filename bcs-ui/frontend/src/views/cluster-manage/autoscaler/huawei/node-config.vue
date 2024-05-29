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
        <bk-form-item :label="$t('cluster.ca.nodePool.create.imageProvider.title')">
          <bk-radio-group v-model="nodePoolConfig.launchTemplate.imageInfo.imageType">
            <bk-radio value="PUBLIC_IMAGE" disabled>{{$t('deploy.image.publicImage')}}</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <!-- 操作系统，暂时不需要desc -->
        <bk-form-item
          :label="$t('cluster.ca.nodePool.label.system')"
          property="nodePoolConfig.nodeOS"
          error-display-type="normal"
          required>
          <bcs-select :clearable="false" searchable v-model="nodePoolConfig.nodeOS" :loading="osImageLoading">
            <bcs-option v-for="item in osImageList" :key="item.imageID" :id="item.imageID" :name="item.osName"></bcs-option>
          </bcs-select>
        </bk-form-item>
        <!-- 计费模式 -->
        <bk-form-item
          :label="$t('tke.label.chargeType.text')"
          :desc="{
            allowHTML: true,
            content: '#chargeDesc',
          }"
          property="nodePoolConfig.launchTemplate.instanceChargeType"
          error-display-type="normal"
          required>
          <bk-radio-group
            class="inline-flex items-center h-[32px]"
            v-model="nodePoolConfig.launchTemplate.instanceChargeType"
            :disabled="isEdit"
            @change="handleInstanceChargeTypeChange">
            <!-- 只保留按量计费-->
            <bk-radio value="POSTPAID_BY_HOUR" :disabled="isEdit">
              {{ $t('tke.label.chargeType.postpaid_by_hour') }}
            </bk-radio>
            <!-- 模式对比不再需要 -->
          </bk-radio-group>
          <p
            class="text-[#979BA5] leading-4 mt-[4px] text-[12px]"
            v-if="nodePoolConfig.launchTemplate.instanceChargeType === 'PREPAID'">
            {{ $t('tke.tips.prepaidOfCA') }}
          </p>
          <div id="chargeDesc">
            <!-- 只保留按量计费的描述 -->
            <div>{{ $t('tke.label.chargeType.postpaid_by_hour_desc', [$t('tke.label.chargeType.postpaid_by_hour')]) }}</div>
          </div>
        </bk-form-item>
        <template v-if="nodePoolConfig.launchTemplate.instanceChargeType === 'PREPAID' && nodePoolConfig.launchTemplate.charge">
          <bk-form-item :label="$t('tke.label.period')">
            <bcs-select :disabled="isEdit" :clearable="false" searchable v-model="nodePoolConfig.launchTemplate.charge.period">
              <bcs-option v-for="item in periodList" :key="item.id" :id="item.id" :name="item.name"></bcs-option>
            </bcs-select>
          </bk-form-item>
          <bk-form-item :label="$t('tke.label.autoRenewal.text')">
            <bcs-checkbox
              true-value="NOTIFY_AND_AUTO_RENEW"
              false-value="NOTIFY_AND_MANUAL_RENEW"
              v-model="nodePoolConfig.launchTemplate.charge.renewFlag">
              {{ $t('tke.label.autoRenewal.desc') }}
            </bcs-checkbox>
          </bk-form-item>
        </template>
        <!-- 可用区 -->
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.az.title')"
          :desc="$t('cluster.ca.nodePool.create.az.desc')"
          property="nodePoolConfig.autoScaling.zones">
          <div class="flex items-center h-[32px]">
            <bk-radio-group
              :value="isSpecifiedZoneList"
              class="w-[auto]"
              @change="handleZoneTypeChange">
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
              v-model="nodePoolConfig.autoScaling.zones[0]"
              searchable
              :loading="zoneListLoading"
              :disabled="isEdit"
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
        <!-- 机型 -->
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
            @row-click="handleCheckInstanceType"
            @filter-change="handleFilterChange"
            @sort-change="handleSortChange">
            <bcs-table-column
              :label="$t('generic.ipSelector.label.serverModel')"
              :filters="nodeTypeFilters"
              :key="nodeTypeFilters.length"
              column-key="typeName"
              filter-multiple
              prop="typeName"
              show-overflow-tooltip>
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
            <bcs-table-column
              :label="$t('generic.label.specifications')"
              min-width="120"
              show-overflow-tooltip
              prop="nodeType">
            </bcs-table-column>
            <bcs-table-column label="CPU" prop="cpu" width="80" align="right" sortable>
              <template #default="{ row }">
                <span>{{ `${row.cpu}${$t('units.suffix.cores')}` }}</span>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.mem')" prop="memory" width="80" align="right" sortable>
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
          <div class="mt25 flex items-center">
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
            <bk-input
              class="w-[88px] bg-[#fff] ml10"
              type="number"
              :disabled="isEdit"
              :min="40"
              :max="1024"
              v-model="nodePoolConfig.launchTemplate.systemDisk.diskSize">
            </bk-input>
            <span :class="['company', { disabled: isEdit }]">GB</span>
            <p class="error-tips bcs-ellipsis ml-[6px] flex-1" v-if="nodePoolConfig.launchTemplate.systemDisk.diskSize % 10 !== 0">
              {{$t('cluster.ca.nodePool.create.instanceTypeConfig.validate.systemDisk')}}
            </p>
          </div>
          <div class="mt20">
            {{$t('cluster.ca.nodePool.create.instanceTypeConfig.label.purchaseDataDisk')}}
          </div>
          <template>
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
                  :min="20"
                  :max="32768"
                  v-model="disk.diskSize">
                </bk-input>
                <span :class="['company', { disabled: isEdit }]">GB</span>
              </div>
              <p class="error-tips" v-if="disk.diskSize % 10 !== 0">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.validate.dataDisks')}}</p>
              <!-- 挂载于, 第一块盘无挂载 -->
              <div class="panel-item mt10" v-if="index >= 1">
                <bk-checkbox
                  :disabled="isEdit"
                  v-model="disk.autoMount">
                  {{$t('cluster.ca.nodePool.create.instanceTypeConfig.label.huaweiMountPath')}}
                </bk-checkbox>
                <template v-if="disk.autoMount">
                  <bcs-input class="ml10" :disabled="isEdit" v-model="disk.mountTarget"></bcs-input>
                </template>
              </div>
              <p class="error-tips" v-if="showRepeatMountTarget(index)">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.validate.repeatPath')}}</p>
              <span :class="['panel-delete', { disabled: isEdit }]" @click="handleDeleteDiskData(index)">
                <i class="bk-icon icon-close3-shape"></i>
              </span>
            </div>
            <!-- 添加数据盘 -->
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
              :disabled="true"
              v-model="nodePoolConfig.launchTemplate.internetAccess.publicIPAssigned">
              {{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.text')}}
            </bk-checkbox>
          </span>
        </bk-form-item>
        <!-- 登录方式 -->
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
                :label="$t('cluster.ca.nodePool.create.password.set')"
                property="nodePoolConfig.launchTemplate.initLoginPassword"
                error-display-type="normal"
                :label-width="120">
                <bcs-input
                  type="password"
                  :disabled="isEdit"
                  v-model="nodePoolConfig.launchTemplate.initLoginPassword" />
              </bk-form-item>
              <bk-form-item
                :label="$t('cluster.ca.nodePool.create.password.confirm')"
                property="confirmPassword"
                error-display-type="normal"
                :label-width="120">
                <bcs-input
                  type="password"
                  :disabled="isEdit"
                  v-model="confirmPassword" />
              </bk-form-item>
            </template>
            <template v-else-if="loginType === 'ssh'">
              <bk-form-item
                :label="$t('cluster.ca.nodePool.create.loginType.ssh.label.publicKey.text')"
                :label-width="100"
                :desc="$t('cluster.ca.nodePool.create.loginType.ssh.label.publicKey.desc')"
                property="nodePoolConfig.launchTemplate.keyPair.keyID"
                error-display-type="normal">
                <bcs-select
                  class="bg-[#fff]"
                  :loading="keyPairsLoading"
                  searchable
                  clearable
                  :disabled="isEdit"
                  v-model="nodePoolConfig.launchTemplate.keyPair.keyID">
                  <bcs-option
                    v-for="item in keyPairsList"
                    :key="item.KeyID"
                    :id="item.KeyID"
                    :name="item.KeyName" />
                  <template #extension>
                    <div slot="extension" class="flex items-center">
                      <div class="flex-1 flex items-center justify-center cursor-pointer" @click="handleCreateKeyPair">
                        <i class="bk-icon icon-plus-circle text-[14px] mr-[5px]"></i>{{ $t('cluster.ca.nodePool.create.loginType.ssh.button.create') }}
                      </div>
                      <span
                        class="w-[48px] h-[16px] flex items-center justify-center cursor-pointer"
                        style="border-left: 1px solid #DCDEE5;"
                        @click="handleGetKeyPairs">
                        <i class="bcs-icon bcs-icon-reset"></i>
                      </span>
                    </div>
                  </template>
                </bcs-select>
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
        <!-- 安全组 -->
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
              :id="securityGroup.securityGroupID"
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
        <!-- 子网 -->
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
                <bcs-radio
                  :value="nodePoolConfig.autoScaling.subnetIDs.includes(row.subnetID)">
                  <span class="bcs-ellipsis">{{row.subnetID}}</span>
                </bcs-radio>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('cluster.ca.nodePool.create.subnet.label.subnetName')" prop="subnetName" show-overflow-tooltip></bcs-table-column>
            <bcs-table-column :label="$t('cluster.ca.nodePool.create.subnet.label.remainIp')" prop="availableIPAddressCount" align="right"></bcs-table-column>
          </bcs-table>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.containerRuntime.title')">
          <bk-radio-group v-model="nodePoolConfig.nodeTemplate.runtime.containerRuntime" :loading="runtimeLoading">
            <bk-radio value="docker" disabled>docker</bk-radio>
            <bk-radio value="containerd" disabled>containerd</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.nodePool.create.runtimeVersion.title')">
          <bcs-input disabled v-model="nodePoolConfig.nodeTemplate.runtime.runtimeVersion"></bcs-input>
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

import { cloudsRuntimeInfo, cloudsZones } from '@/api/modules/cluster-manager';
import TextTips from '@/components/layout/TextTips.vue';
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import FormGroup from '@/views/cluster-manage/add/common/form-group.vue';
import Schema from '@/views/cluster-manage/autoscaler/resolve-schema';
import { useCloud, useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';

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
        id: 'SATA',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.SATA'),
      },
      {
        id: 'SAS',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.SAS'),
      },
      {
        id: 'GPSSD',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.GPSSD'),
      },
      {
        id: 'SSD',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.SSD'),
      },
      {
        id: 'ESSD',
        name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.ESSD'),
      },
    ]);
    const confirmPassword = ref(''); // 确认密码
    const nodePoolConfig = ref({
      nodeGroupID: defaultValues.value.nodeGroupID, // 编辑时
      name: defaultValues.value.name || '', // 节点名称
      nodeOS: defaultValues.value.nodeOS || '', // 节点操作系统
      autoScaling: {
        vpcID: '', // todo 放在basic-pool-info组件比较合适
        zones: defaultValues.value.autoScaling?.zones || [],
        subnetIDs: defaultValues.value.autoScaling.subnetIDs || [], // 支持子网
      },
      launchTemplate: {
        imageInfo: {
          imageID: defaultValues.value.launchTemplate?.imageInfo?.imageID, // 镜像ID
          imageName: defaultValues.value.launchTemplate?.imageInfo?.imageName, // 镜像名称
          imageType: defaultValues.value.launchTemplate?.imageInfo?.imageType || 'PUBLIC_IMAGE', // 镜像类型
        },
        CPU: '',
        Mem: '',
        instanceType: defaultValues.value.launchTemplate?.instanceType, // 机型信息
        systemDisk: {
          diskType: diskEnum.value.some(item => item.id === defaultValues.value.launchTemplate?.systemDisk?.diskType)
            ? defaultValues.value.launchTemplate?.systemDisk?.diskType
            : 'SSD', // 系统盘类型
          diskSize: defaultValues.value.launchTemplate?.systemDisk?.diskSize, // 系统盘大小
        },
        internetAccess: {
          internetChargeType: defaultValues.value.launchTemplate?.internetAccess?.internetChargeType, // 计费方式
          internetMaxBandwidth: defaultValues.value.launchTemplate?.internetAccess?.internetMaxBandwidth, // 购买带宽
          publicIPAssigned: defaultValues.value.launchTemplate?.internetAccess?.publicIPAssigned, // 分配免费公网IP
          bandwidthPackageId: defaultValues.value.launchTemplate?.internetAccess?.bandwidthPackageId, // 带宽包ID
        },
        // 密钥信息
        keyPair: {
          keyID: defaultValues.value.launchTemplate?.keyPair?.keyID || '',
          keySecret: defaultValues.value.launchTemplate?.keyPair?.keySecret || '',
        },
        initLoginPassword: defaultValues.value.launchTemplate?.initLoginPassword, // 密码
        securityGroupIDs: defaultValues.value.launchTemplate?.securityGroupIDs || [], // 安全组
        dataDisks: defaultValues.value.launchTemplate?.dataDisks || [], // 数据盘
        // 默认值
        isSecurityService: defaultValues.value.launchTemplate?.isSecurityService || true,
        isMonitorService: defaultValues.value.launchTemplate?.isMonitorService || true,
        instanceChargeType: defaultValues.value.launchTemplate?.instanceChargeType || 'POSTPAID_BY_HOUR', // 计费模式
        charge: defaultValues.value.launchTemplate?.charge,
      },
      nodeTemplate: {
        dataDisks: defaultValues.value.nodeTemplate?.dataDisks || [],
        dockerGraphPath: defaultValues.value.nodeTemplate?.dockerGraphPath, // 容器目录
        runtime: {
          containerRuntime: defaultValues.value.nodeTemplate.runtime?.containerRuntime || '', // 运行时容器组件
          runtimeVersion: defaultValues.value.nodeTemplate.runtime?.runtimeVersion || '', // 运行时版本
        },
      },
      extra: {
        provider: '', // 机型provider信息
      },
    });

    const nodePoolConfigRules = ref({
      // 操作系统
      'nodePoolConfig.nodeOS': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => nodePoolConfig.value.nodeOS,
        },
      ],
      // 计费模式
      'nodePoolConfig.launchTemplate.instanceChargeType': [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
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
      // 密钥校验
      'nodePoolConfig.launchTemplate.keyPair.keyID': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => isEdit.value || nodePoolConfig.value.launchTemplate.keyPair.keyID.length > 0,
        },
      ],
      'nodePoolConfig.launchTemplate.keyPair.keySecret': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => isEdit.value || nodePoolConfig.value.launchTemplate.keyPair.keySecret.length > 0,
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
      'nodePoolConfig.autoScaling.zones': [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => !isSpecifiedZoneList.value || !!nodePoolConfig.value.autoScaling?.zones?.length,
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

    // 计费模式
    const handleInstanceChargeTypeChange = (value) => {
      if (value === 'PREPAID') {
        nodePoolConfig.value.launchTemplate.charge = {
          period: 1,
          renewFlag: 'NOTIFY_AND_AUTO_RENEW',
        };
      } else if (value === 'POSTPAID_BY_HOUR') {
        nodePoolConfig.value.launchTemplate.charge = null;
      }
    };
    const periodList = ref([
      ...[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11].map(month => ({
        id: month,
        name: $i18n.t('units.time.nMonths', [month]),
      })),
      ...[1, 2, 3].map(year => ({
        id: year * 12,
        name: $i18n.t('units.time.nYears', [year]),
      })),
    ]);
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
    const handleZoneTypeChange = (v: boolean) => {
      showZoneList.value = v;
      nodePoolConfig.value.autoScaling.zones = [];
    };

    const handleSetZoneData = (value) => {
      nodePoolConfig.value.autoScaling.zones = [value];
    };

    // 镜像
    watch(() => nodePoolConfig.value.launchTemplate.imageInfo.imageType, () => {
      // 重置镜像ID
      nodePoolConfig.value.launchTemplate.imageInfo.imageID = '';
      // 获取镜像列表
      handleGetOsImage();
    });
    const osImageLoading = ref(false);
    const osImageList = ref<Array<{
      imageID: string
      alias: string
      osName: string
    }>>([]);
    const handleGetOsImage = async () => {
      osImageLoading.value = true;
      osImageList.value = await $store.dispatch('clustermanager/cloudOsImage', {
        $cloudID: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
        provider: nodePoolConfig.value.launchTemplate.imageInfo.imageType,
      });
      osImageLoading.value = false;
    };

    // 机型
    const filters = ref<Record<string, string[]>>({});
    const sortData = ref({
      prop: '',
      order: '',
    });
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

    // 分页后重置当前页码
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
        // cpu: nodePoolConfig.value.launchTemplate.CPU,
        // memory: nodePoolConfig.value.launchTemplate.Mem,
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
    const filterInstanceList = computed(() => instanceTypesList.value
      .sort((a, b) => {
        // 排序
        if (sortData.value.prop === 'cpu') {
          return sortData.value.order === 'ascending' ? a.cpu - b.cpu : b.cpu - a.cpu;
        }
        if (sortData.value.prop === 'memory') {
          return sortData.value.order === 'ascending' ? a.memory - b.memory : b.memory - a.memory;
        }
        if (sortData.value.prop === 'unitPrice') {
          return sortData.value.order === 'ascending' ? a.unitPrice - b.unitPrice : b.unitPrice - a.unitPrice;
        }
        return 0;
      }));
    const nodeTypeFilters = computed(() => filterInstanceList.value
      .reduce<Array<{text: string, value: string}>>((pre, item) => {
      const exist = pre.find(data => data.value === item.typeName);
      if (!exist) {
        pre.push({
          text: item.typeName,
          value: item.typeName,
        });
      }
      return pre;
    }, []));

    const filterTableData = computed(() => filterInstanceList.value.filter(item => Object.keys(filters.value)
      .every(key => !filters.value[key]?.length || filters.value[key]?.includes(item[key]))));

    const {
      pagination,
      curPageData: instanceList,
      pageChange,
      pageSizeChange,
      handleResetPage,
    } = usePage(filterTableData);

    const handleFilterChange = (data) => {
      pageChange(1);
      filters.value = data;
    };
    const handleSortChange = ({ prop, order  }) => {
      sortData.value = {
        prop,
        order,
      };
    };

    // 数据盘
    const defaultDiskItem = {
      diskType: 'SSD',
      diskSize: '100',
      autoMount: true,
      mountTarget: '/data',
    };
    const handleShowDataDisksChange = (show) => {
      // huawei第一块数据盘不需要挂载
      const { diskType, diskSize } = defaultDiskItem;
      nodePoolConfig.value.nodeTemplate.dataDisks = show
        ? [{ diskType, diskSize }] : [];
    };
    const handleDeleteDiskData = (index) => {
      if (isEdit.value) return;
      nodePoolConfig.value.nodeTemplate.dataDisks.splice(index, 1);
    };
    const handleAddDiskData = () => {
      if (isEdit.value || nodePoolConfig.value.nodeTemplate.dataDisks.length > 4) return;
      nodePoolConfig.value.nodeTemplate.dataDisks.push(JSON.parse(JSON.stringify(defaultDiskItem)));
    };

    // 登录方式
    const loginType = ref<'ssh'|'password'>(nodePoolConfig.value.launchTemplate.keyPair.keyID
      ? 'ssh'
      : 'password');
    const handleLoginTypeChange = (type) => {
      if (type === 'ssh') {
        nodePoolConfig.value.launchTemplate.initLoginPassword = '';
      } else {
        nodePoolConfig.value.launchTemplate.keyPair.keyID = '';
        nodePoolConfig.value.launchTemplate.keyPair.keySecret = '';
      }
    };
    // 密钥对
    const keyPairsLoading = ref(false);
    const keyPairsList = ref<Array<{
      KeyID: string
      KeyName: string
    }>>([]);
    const handleGetKeyPairs = async () => {
      keyPairsLoading.value = true;
      keyPairsList.value = await $store.dispatch('clustermanager/cloudKeyPairs', {
        $cloudId: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
      });
      keyPairsLoading.value = false;
    };
    const handleCreateKeyPair = () => {
      window.open('https://console.cloud.tencent.com/cvm/sshkey');
    };

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
      securityGroupsLoading.value = false;
    };

    // VPC子网
    const subnetsLoading = ref(false);
    const subnetsList = ref<Array<{
      zone: string
      zoneName: string
      subnetID: string
    }>>([]);
    const filterSubnetsList = computed(() => subnetsList.value.sort((pre, current) => {
      const isPreDisabled = curInstanceItem.value.zones?.includes(pre.zone);
      const isCurrentDisabled = curInstanceItem.value.zones?.includes(current.zone);
      if (isPreDisabled && !isCurrentDisabled) return -1;

      if (!isPreDisabled && isCurrentDisabled) return 1;

      return 0;
    }));
    const handleGetSubnets = async () => {
      subnetsLoading.value = true;
      subnetsList.value = await $store.dispatch('clustermanager/cloudSubnets', {
        $cloudID: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
        vpcID: cluster.value.vpcID,
      });
      subnetsLoading.value = false;
    };
    const handleCheckSubnets = (row) => {
      const index = nodePoolConfig.value.autoScaling.subnetIDs.indexOf(row.subnetID);
      if (index > -1) {
        nodePoolConfig.value.autoScaling.subnetIDs.splice(index, 1);
      } else {
        nodePoolConfig.value.autoScaling.subnetIDs.push(row.subnetID);
      }
    };

    const runtimeList = ref<any[]>([]);
    const runtimeLoading = ref(false);
    // 运行时组件
    const handleGetRuntimeinfo = async () => {
      runtimeLoading.value = true;
      runtimeList.value = await cloudsRuntimeInfo({
        $cloudId: cluster.value.provider,
        clusterID: cluster.value?.clusterID,
      });
      runtimeLoading.value = false;
      // containerd: {version: ["xxx", "xxx"]}
      const containerRuntime = Object.keys(runtimeList.value)[0] || '';
      nodePoolConfig.value.nodeTemplate.runtime.containerRuntime = containerRuntime;
      nodePoolConfig.value.nodeTemplate.runtime.runtimeVersion = runtimeList.value[containerRuntime]?.version?.join(',');
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
      // nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth = String(nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth);
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
      const result = await formRef.value?.validate().catch(() => false);
      // eslint-disable-next-line max-len
      const validateDataDiskSize = nodePoolConfig.value.nodeTemplate.dataDisks.every(item => item.diskSize % 10 === 0);
      const mountTargetList = nodePoolConfig.value.nodeTemplate.dataDisks.map(item => item.mountTarget);
      const validateDataDiskMountTarget = new Set(mountTargetList).size === mountTargetList.length;
      const validateSystemDisk = nodePoolConfig.value.launchTemplate.systemDisk.diskSize % 10 === 0;
      if (!basicFormValidate
        || !nodePoolConfig.value.launchTemplate.instanceType
        || !result
        || !validateDataDiskSize
        || !validateDataDiskMountTarget
        || !validateSystemDisk) return false;

      return true;
    };
    const handleNext = async () => {
      const result = await validate();
      if (!result) {
        // 自动滚动到第一个错误的位置
        const errDom = document.getElementsByClassName('form-error-tip');
        const bcsErrDom = document.getElementsByClassName('error-tips');
        const isErrDom = document.getElementsByClassName('is-error');
        const firstErrDom = isErrDom[0] || errDom[0] || bcsErrDom[0];
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

    const showRepeatMountTarget = (index) => {
      const disk = nodePoolConfig.value.nodeTemplate.dataDisks[index];
      return disk.autoMount
            && disk.mountTarget
            && nodePoolConfig.value.nodeTemplate.dataDisks
              .filter((item, i) => i !== index && item.autoMount)
              .some(item => item.mountTarget === disk.mountTarget);
    };

    // 集群详情
    const clusterDetailLoading = ref(false);
    const { clusterData, getClusterDetail } = useClusterInfo();
    const handleGetClusterDetail = async () => {
      clusterDetailLoading.value = true;
      await getClusterDetail(cluster.value?.clusterID, true);
      nodePoolConfig.value.autoScaling.vpcID = clusterData.value.vpcID;
      clusterDetailLoading.value = false;
    };

    onMounted(async () => {
      // 显示购买数据盘
      handleShowDataDisksChange(true);
      await handleGetClusterDetail(); // 优选获取集群详情信息
      // 获取镜像，操作系统类型
      handleGetOsImage();
      handleGetInstanceTypes();
      handleGetCloudSecurityGroups();
      handleGetZoneList();
      handleGetSubnets();
      handleGetKeyPairs();
      handleGetRuntimeinfo();
    });

    return {
      isSpecifiedZoneList,
      zoneListLoading,
      zoneList,
      runtimeLoading,
      handleZoneTypeChange,
      handleSetZoneData,
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
      handleGetRuntimeinfo,
      getNodePoolData,
      CPU,
      Mem,
      cpuList,
      memList,
      clusterDetailLoading,
      instanceTypesList,
      osImageLoading,
      osImageList,
      filterSubnetsList,
      subnetsLoading,
      handleCheckSubnets,
      loginType,
      keyPairsLoading,
      keyPairsList,
      handleLoginTypeChange,
      handleGetKeyPairs,
      handleCreateKeyPair,
      nodeTypeFilters,
      handleFilterChange,
      handleSortChange,
      periodList,
      handleInstanceChargeTypeChange,
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
