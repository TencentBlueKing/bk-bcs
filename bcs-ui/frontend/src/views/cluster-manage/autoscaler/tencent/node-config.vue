<!-- eslint-disable max-len -->
<template>
  <div>
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
              <bk-radio value="MARKET_IMAGE" disabled>{{$t('deploy.image.marketImage')}}</bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item
            :label="$t('cluster.ca.nodePool.label.system')"
            property="launchTemplate.imageInfo.imageID"
            error-display-type="normal"
            required
            :desc="$t('cluster.ca.nodePool.create.image.desc')">
            <bcs-input disabled :value="clusterOS"></bcs-input>
          </bk-form-item>
          <!-- 计费模式 -->
          <bk-form-item
            :label="$t('tke.label.chargeType.text')"
            :desc="{
              allowHTML: true,
              content: '#chargeDesc',
            }"
            property="launchTemplate.instanceChargeType"
            error-display-type="normal"
            required>
            <bk-radio-group
              class="inline-flex items-center h-[32px]"
              v-model="nodePoolConfig.launchTemplate.instanceChargeType"
              :disabled="isEdit"
              @change="handleInstanceChargeTypeChange">
              <bk-radio value="POSTPAID_BY_HOUR" :disabled="isEdit">
                {{ $t('tke.label.chargeType.postpaid_by_hour') }}
              </bk-radio>
              <bk-radio value="PREPAID" :disabled="isEdit">
                {{ $t('tke.label.chargeType.prepaid') }}
              </bk-radio>
              <bk-link
                theme="primary"
                class="ml-[30px]"
                href="https://cloud.tencent.com/document/product/213/2180"
                target="_blank">
                <i class="bcs-icon bcs-icon-fenxiang mr-[2px]"></i>
                {{ $t('tke.button.chargeTypeDiff') }}
              </bk-link>
            </bk-radio-group>
            <p
              class="text-[#979BA5] leading-4 mt-[4px] text-[12px]"
              v-if="nodePoolConfig.launchTemplate.instanceChargeType === 'PREPAID'">
              {{ $t('tke.tips.prepaidOfCA') }}
            </p>
            <div id="chargeDesc">
              <div>{{ $t('tke.label.chargeType.prepaidDesc', [$t('tke.label.chargeType.prepaid')]) }}</div>
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
                v-model="nodePoolConfig.autoScaling.zones"
                multiple
                searchable
                :loading="zoneListLoading"
                :disabled="isEdit"
                :clearable="false"
                selected-style="checkbox">
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
              <bcs-table-column
                :label="$t('cluster.ca.nodePool.create.instanceTypeConfig.label.configurationFee.text')"
                min-width="120"
                prop="unitPrice"
                sortable>
                <template #default="{ row }">
                  {{ $t('cluster.ca.nodePool.create.instanceTypeConfig.label.configurationFee.unit', { price: row.unitPrice }) }}
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('generic.label.status')" width="80">
                <template #default="{ row }">
                  {{ row.status === 'SELL' ? $t('cluster.ca.nodePool.create.instanceTypeConfig.status.sell') : $t('cluster.ca.nodePool.create.instanceTypeConfig.status.soldOut') }}
                </template>
              </bcs-table-column>
            </bcs-table>
            <p class="text-[12px] text-[#ea3636]" v-if="!nodePoolConfig.launchTemplate.instanceType">{{ $t('generic.validate.required') }}</p>
            <!-- 系统盘 -->
            <SystemDisk
              class="mt-[24px]"
              :list="systemDisks"
              :value="nodePoolConfig.launchTemplate.systemDisk"
              :first-trigger="firstTrigger"
              :loading="isLoading"
              :ref="el => systemDiskRef = el"
              :is-edit="isEdit"
              @change="(v) => nodePoolConfig.launchTemplate.systemDisk = v"
              @validate="firstTrigger = false" />
            <!-- 数据盘 -->
            <DataDisk
              class="mt-[20px]"
              :list="dataDisks"
              :value="nodePoolConfig.nodeTemplate.dataDisks"
              :is-edit="isEdit"
              :first-trigger="firstTrigger"
              :loading="isLoading"
              :ref="el => dataDiskRef = el"
              @change="(v) => nodePoolConfig.nodeTemplate.dataDisks = v"
              @validate="firstTrigger = false" />

            <span class="inline-flex mt15">
              <bk-checkbox
                :disabled="isEdit"
                v-model="nodePoolConfig.launchTemplate.internetAccess.publicIPAssigned">
                {{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.text')}}
              </bk-checkbox>
            </span>
            <div class="panel" v-if="nodePoolConfig.launchTemplate.internetAccess.publicIPAssigned">
              <div class="panel-item">
                <label class="label">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.chargeMode.text')}}</label>
                <bk-radio-group
                  v-model="nodePoolConfig.launchTemplate.internetAccess.internetChargeType"
                  @change="handleChargeTypeChange">
                  <bk-radio :disabled="isEdit || accountType === 'LEGACY'" value="TRAFFIC_POSTPAID_BY_HOUR">
                    <span
                      v-bk-tooltips="{
                        content: $t('ca.internetAccess.tips.accountNotAvailable'),
                        disabled: accountType !== 'LEGACY'
                      }">
                      {{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.chargeMode.traffic_postpaid_by_hour')}}
                    </span>
                  </bk-radio>
                  <bk-radio :disabled="isEdit || accountType === 'LEGACY'" value="BANDWIDTH_PREPAID">
                    <span
                      v-bk-tooltips="{
                        content: $t('ca.internetAccess.tips.accountNotAvailable'),
                        disabled: accountType !== 'LEGACY'
                      }">
                      {{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.chargeMode.bandwidth_prepaid')}}
                    </span>
                  </bk-radio>
                  <bk-radio :disabled="isEdit" value="BANDWIDTH_PACKAGE">
                    {{ $t('ca.internetAccess.bandwidthPackage') }}
                  </bk-radio>
                </bk-radio-group>
              </div>
              <div
                class="panel-item mt10"
                v-if="accountType === 'STANDARD'
                  && nodePoolConfig.launchTemplate.internetAccess.internetChargeType === 'BANDWIDTH_PACKAGE'">
                <label class="label">{{ $t('ca.internetAccess.label.selectBandwidthPackage') }}</label>
                <bcs-select
                  :loading="bandwidthLoading"
                  :disabled="isEdit"
                  class="w-[200px] bg-[#fff]"
                  v-model="nodePoolConfig.launchTemplate.internetAccess.bandwidthPackageId">
                  <bcs-option
                    v-for="item in bandwidthList"
                    :key="item.id"
                    :id="item.id"
                    :name="`${item.name}(${item.id})`" />
                  <span slot="extension" class="cursor-pointer" @click="createBandWidth">
                    <i class="bcs-icon bcs-icon-fenxiang mr5 !text-[12px]"></i>
                    {{ $t('ca.internetAccess.button.createBandWidth') }}
                  </span>
                </bcs-select>
                <span
                  class="ml10 text-[12px] cursor-pointer"
                  v-bk-tooltips.top="$t('generic.button.refresh')"
                  @click="getBandwidthList">
                  <i class="bcs-icon bcs-icon-reset"></i>
                </span>
              </div>
              <div class="panel-item mt10">
                <label class="label">{{$t('cluster.ca.nodePool.create.instanceTypeConfig.publicIPAssigned.maxBandWidth')}}</label>
                <bk-radio-group v-model="maxBandwidthType">
                  <bk-radio :disabled="isEdit" value="limit">
                    <div class="flex items-center">
                      <span class="mr-[8px]">{{ $t('ca.internetAccess.label.limit') }}</span>
                      <bk-input
                        class="max-w-[80px]"
                        type="number"
                        :min="1"
                        :max="nodePoolConfig.launchTemplate.internetAccess.internetChargeType === 'BANDWIDTH_PACKAGE' ? 2000 : 100"
                        :disabled="isEdit"
                        v-model="nodePoolConfig.launchTemplate.internetAccess.internetMaxBandwidth">
                      </bk-input>
                      <span :class="['company', { disabled: isEdit }]">Mbps</span>
                    </div>
                  </bk-radio>
                  <bk-radio
                    :disabled="isEdit"
                    value="un-limit"
                    v-if="nodePoolConfig.launchTemplate.internetAccess.internetChargeType === 'BANDWIDTH_PACKAGE'">
                    <span>{{ $t('ca.internetAccess.label.unLimit') }}</span>
                  </bk-radio>
                </bk-radio-group>
              </div>
            </div>
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
          <bk-form-item
            :label="$t('cluster.ca.nodePool.create.subnet.title')"
            property="nodePoolConfig.autoScaling.subnetIDs"
            error-display-type="normal"
            required>
            <bcs-table
              :data="filterSubnetsList"
              :row-class-name="subnetsRowClass"
              v-bkloading="{ isLoading: subnetsLoading }"
              @row-click="handleCheckSubnets">
              <bcs-table-column :label="$t('cluster.ca.nodePool.create.subnet.label.subnetID')" min-width="150">
                <template #default="{ row }">
                  <bk-checkbox
                    :disabled="!(curInstanceItem.zones && curInstanceItem.zones.includes(row.zone)) || isEdit"
                    :value="nodePoolConfig.autoScaling.subnetIDs.includes(row.subnetID)">
                    <span class="bcs-ellipsis">{{row.subnetID}}</span>
                  </bk-checkbox>
                </template>
              </bcs-table-column>
              <bcs-table-column :label="$t('cluster.ca.nodePool.create.subnet.label.subnetName')" prop="subnetName" show-overflow-tooltip></bcs-table-column>
              <bcs-table-column :label="$t('cluster.ca.nodePool.create.az.title')" prop="zoneName" show-overflow-tooltip></bcs-table-column>
              <bcs-table-column :label="$t('cluster.ca.nodePool.create.subnet.label.remainIp')" prop="availableIPAddressCount" align="right"></bcs-table-column>
              <bcs-table-column :label="$t('cluster.ca.nodePool.create.subnet.label.unUsedReason.text')" show-overflow-tooltip>
                <template #default="{ row }">
                  {{ curInstanceItem.zones && curInstanceItem.zones.includes(row.zone)
                    ? '--' : $t('cluster.ca.nodePool.create.subnet.label.unUsedReason.desc') }}
                </template>
              </bcs-table-column>
            </bcs-table>
          </bk-form-item>
          <bk-form-item :label="$t('dashboard.workload.container.dataDir')">
            <bcs-input disabled v-model="nodePoolConfig.nodeTemplate.dockerGraphPath"></bcs-input>
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

import { DiskItem, useDisk } from '../../add/create/tencent-public-cloud/use-disk';

import { cloudsZones } from '@/api/modules/cluster-manager';
import FormGroup from '@/components/form-group.vue';
import TextTips from '@/components/layout/TextTips.vue';
import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import usePage from '@/composables/use-page';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import DataDisk from '@/views/cluster-manage/add/components/data-disk.vue';
import SystemDisk from '@/views/cluster-manage/add/components/system-disk.vue';
import Schema from '@/views/cluster-manage/autoscaler/resolve-schema';
import { useCloud, useClusterInfo } from '@/views/cluster-manage/cluster/use-cluster';

export default defineComponent({
  components: { FormGroup, TextTips, SystemDisk, DataDisk },
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
          imageID: defaultValues.value.launchTemplate?.imageInfo?.imageID, // 镜像ID
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
        // 密钥信息
        keyPair: {
          keyID: defaultValues.value.launchTemplate?.keyPair?.keyID || '',
          keySecret: defaultValues.value.launchTemplate?.keyPair?.keySecret || '',
        },
        initLoginPassword: defaultValues.value.launchTemplate?.initLoginPassword, // 密码
        securityGroupIDs: defaultValues.value.launchTemplate?.securityGroupIDs || [], // 安全组
        dataDisks: defaultValues.value.launchTemplate?.dataDisks || [] as DiskItem[], // 数据盘
        // 默认值
        isSecurityService: defaultValues.value.launchTemplate?.isSecurityService || true,
        isMonitorService: defaultValues.value.launchTemplate?.isMonitorService || true,
        instanceChargeType: defaultValues.value.launchTemplate?.instanceChargeType || 'POSTPAID_BY_HOUR', // 计费模式
        charge: defaultValues.value.launchTemplate?.charge,
      },
      nodeTemplate: {
        dataDisks: defaultValues.value.nodeTemplate?.dataDisks || [] as DiskItem[], // 数据盘
        dockerGraphPath: defaultValues.value.nodeTemplate?.dockerGraphPath, // 容器目录
        runtime: {
          containerRuntime: defaultValues.value.nodeTemplate.runtime?.containerRuntime, // 运行时容器组件
          runtimeVersion: defaultValues.value.nodeTemplate.runtime?.runtimeVersion, // 运行时版本
        },
      },
      extra: {
        provider: '', // 机型provider信息
      },
    });

    const nodePoolConfigRules = ref({
      // 镜像ID校验
      'launchTemplate.imageInfo.imageID': [
        {
          required: true,
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
        },
      ],
      'launchTemplate.instanceChargeType': [
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
          validator: () => isEdit.value
            || confirmPassword.value === nodePoolConfig.value.launchTemplate.initLoginPassword,
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
      instanceFamily.value = '';
      cpu.value = 0;
      memory.value = 0;
      handleSetDefaultInstance();
      // // 获取机型
      // handleGetInstanceTypes();
    });
    watch(curInstanceItem, () => {
      nodePoolConfig.value.extra.provider = curInstanceItem.value.provider;
      // nodePoolConfig.value.autoScaling.zones = [];
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
      instanceFamily.value = row.nodeFamily;
      cpu.value = row.cpu;
      memory.value = row.memory;
    };
    // 设置默认机型
    const handleSetDefaultInstance = () => {
      setTimeout(() => {
        // 默认机型配置
        let instance;
        if (!nodePoolConfig.value.launchTemplate.instanceType) {
          instance = instanceTypesList.value
            .find(instance => instance.status === 'SELL');
        } else {
          instance = instanceTypesList.value
            .find(instance => instance.nodeType === nodePoolConfig.value.launchTemplate.instanceType);
        }
        nodePoolConfig.value.launchTemplate.instanceType = instance?.nodeType;
        instanceFamily.value = instance?.nodeFamily;
        cpu.value = instance?.cpu;
        memory.value = instance?.memory;
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
        if (pre.status !== 'SELL') return 1;
        if (current.status !== 'SELL') return 0;
        return pre.cpu - current.cpu;
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


    // 免费分配公网IP
    watch(() => nodePoolConfig.value.launchTemplate.internetAccess.publicIPAssigned, (publicIPAssigned) => {
      nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth = publicIPAssigned ? '10' : '0';
    });
    const maxBandwidth = 65535;
    const maxBandwidthType = ref<'limit'|'un-limit'>(Number(nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth) === maxBandwidth
      ? 'un-limit'
      : 'limit');
    const bandwidthLoading = ref(false);
    const bandwidthList = ref<any[]>([]);
    const getBandwidthList = async () => {
      bandwidthLoading.value = true;
      bandwidthList.value = await getCloudBwps({
        $cloudId: cluster.value.provider,
        accountID: cluster.value.cloudAccountID,
        region: cluster.value.region,
      });
      bandwidthLoading.value = false;
    };
    const handleChargeTypeChange = (value: 'TRAFFIC_POSTPAID_BY_HOUR' | 'BANDWIDTH_PREPAID' | 'BANDWIDTH_PACKAGE') => {
      if (value !== 'BANDWIDTH_PACKAGE' && Number(nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth) > 100) {
        // 'TRAFFIC_POSTPAID_BY_HOUR' | 'BANDWIDTH_PREPAID' 类型最大带宽为 100
        nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth = '100';
      }
      if (value !== 'BANDWIDTH_PACKAGE' && maxBandwidthType.value === 'un-limit') {
        maxBandwidthType.value = 'limit';
      }
    };
    const createBandWidth = () => {
      window.open('https://console.cloud.tencent.com/vpc/package');
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
    const filterSubnetsList = computed(() => {
      const { zones } = nodePoolConfig.value.autoScaling;
      return subnetsList.value
        .filter(item => !zones.length || zones.includes(item.zone))
        .sort((pre, current) => {
          const isPreDisabled = curInstanceItem.value.zones?.includes(pre.zone);
          const isCurrentDisabled = curInstanceItem.value.zones?.includes(current.zone);
          if (isPreDisabled && !isCurrentDisabled) return -1;

          if (!isPreDisabled && isCurrentDisabled) return 1;

          return 0;
        });
    });
    const subnetsRowClass = ({ row }) => {
      if (curInstanceItem.value.zones?.includes(row.zone)) {
        return 'table-row-enable';
      }
      return 'table-row-disabled';
    };
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
      if (!curInstanceItem.value.zones?.includes(row.zone) || isEdit.value) return;

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
      nodePoolConfig.value.nodeTemplate.dataDisks.forEach((item) => {
        // eslint-disable-next-line no-param-reassign
        item.diskSize = String(item.diskSize);
      });
      // eslint-disable-next-line max-len
      nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth = String(nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth);
      if (maxBandwidthType.value === 'un-limit') {
        nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth = '65535';
      }
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


    // 校验
    const firstTrigger = ref(true);
    const systemDiskRef = ref();
    const dataDiskRef = ref();

    const validate = async () => {
      firstTrigger.value = false;
      const basicFormValidate = await basicFormRef.value?.validate().catch(() => false);
      const result = await formRef.value?.validate().catch(() => false);
      const validateSystemDisk = await systemDiskRef.value?.validate().catch(() => false);
      const validateDataDisk = await dataDiskRef.value?.validate().catch(() => false);

      if (!basicFormValidate
        || !nodePoolConfig.value.launchTemplate.instanceType
        || !result
        || !validateSystemDisk
        || !validateDataDisk) return false;

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
    const { clusterData, clusterOS, clusterAdvanceSettings, getClusterDetail } = useClusterInfo();
    const handleGetClusterDetail = async () => {
      clusterDetailLoading.value = true;
      await getClusterDetail(cluster.value?.clusterID, true);
      nodePoolConfig.value.autoScaling.vpcID = clusterData.value.vpcID;
      clusterDetailLoading.value = false;
    };

    // 账户类型
    const { accountType, getCloudAccountType, getCloudBwps } = useCloud();

    // 磁盘类型
    const { systemDisks, dataDisks, getDisks, isLoading } = useDisk();
    const instanceFamily = ref('');
    const cpu = ref(0);
    const memory = ref(0);
    async function getDiskEnum() {
      const params = nodePoolConfig.value.autoScaling.zones.length > 0 ? {
        zones: [...nodePoolConfig.value.autoScaling.zones],
      } : {};

      await getDisks({
        cloudID: cluster.value.provider,
        accountID: cluster.value.cloudAccountID,
        region: cluster.value.region,
        instanceFamilies: [instanceFamily.value],
        diskChargeType: nodePoolConfig.value.launchTemplate.instanceChargeType,
        cpu: cpu.value,
        memory: memory.value,
        ...params,
      });
      // 系统盘
      const flag = systemDisks.value.some(item => item.id === nodePoolConfig.value.launchTemplate.systemDisk?.diskType)
        && systemDisks.value?.length > 0;
      if (!flag) {
        // 新选中的机型没有对应磁盘类型，清空
        nodePoolConfig.value.launchTemplate.systemDisk.diskType = '';
      }
      // 数据盘
      nodePoolConfig.value.nodeTemplate.dataDisks.forEach((item) => {
        const val = item;
        const flag = dataDisks.value.some(v => v.id === val.diskType) && dataDisks.value?.length > 0;
        if (!flag) {
          // 新选中的机型没有对应磁盘类型，清空
          val.diskType = '';
        }
      });
    }
    watch(() => instanceFamily.value, async () => {
      if (!instanceFamily.value) {
        systemDisks.value = [];
        dataDisks.value = [];
        return;
      };
      await getDiskEnum();
    }, { immediate: true });

    onMounted(async () => {
      await handleGetClusterDetail(); // 优选获取集群详情信息
      handleGetOsImage();
      handleGetInstanceTypes();
      handleGetCloudSecurityGroups();
      handleGetZoneList();
      handleGetSubnets();
      handleGetKeyPairs();
      getCloudAccountType({
        $cloudId: cluster.value.provider,
        accountID: cluster.value.cloudAccountID,
      }).then(() => {
        if (accountType.value === 'LEGACY') { // 传统账户只能选择带宽包
          nodePoolConfig.value.launchTemplate.internetAccess.internetChargeType = 'BANDWIDTH_PACKAGE';
        }
      });
      getBandwidthList();
    });

    return {
      clusterOS,
      isSpecifiedZoneList,
      zoneListLoading,
      zoneList,
      handleZoneTypeChange,
      handleSetDefaultInstance,
      nodeConfigRef,
      formRef,
      basicFormRef,
      systemDisks,
      dataDisks,
      isLoading,
      firstTrigger,
      systemDiskRef,
      dataDiskRef,
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
      securityGroupsLoading,
      securityGroupsList,
      handleNext,
      handleCancel,
      getSchemaByProp,
      showRepeatMountTarget,
      validate,
      focusOnErrorField,
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
      subnetsRowClass,
      subnetsLoading,
      handleCheckSubnets,
      clusterAdvanceSettings,
      loginType,
      keyPairsLoading,
      keyPairsList,
      handleLoginTypeChange,
      handleGetKeyPairs,
      handleCreateKeyPair,
      accountType,
      maxBandwidthType,
      bandwidthList,
      bandwidthLoading,
      getBandwidthList,
      handleChargeTypeChange,
      createBandWidth,
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
