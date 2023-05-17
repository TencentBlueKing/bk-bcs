<!-- eslint-disable max-len -->
<template>
  <bk-form class="node-config" :model="nodePoolConfig" :rules="nodePoolConfigRules" ref="formRef">
    <bk-form-item :label="$t('镜像提供方')">
      <bk-radio-group v-model="imageProvider">
        <bk-radio value="PUBLIC_IMAGE" :disabled="isEdit">{{$t('公共镜像')}}</bk-radio>
        <bk-radio value="MARKET_IMAGE" :disabled="isEdit">{{$t('市场镜像')}}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item
      :label="$t('操作系统')"
      property="launchTemplate.imageInfo.imageID"
      error-display-type="normal"
      required
      :desc="$t('推荐使用TencentOS Server')">
      <bcs-select
        :loading="osImageLoading"
        v-model="nodePoolConfig.launchTemplate.imageInfo.imageID"
        searchable
        :disabled="isEdit"
        :clearable="false">
        <bcs-option
          v-for="imageItem in osImageList"
          :key="imageItem.imageID"
          :id="imageItem.imageID"
          :name="imageItem.alias"></bcs-option>
      </bcs-select>
    </bk-form-item>
    <bk-form-item>
      <div class="mb15" style="display: flex;">
        <div class="prefix-select">
          <span class="prefix">CPU</span>
          <bcs-select
            v-model="nodePoolConfig.launchTemplate.CPU"
            searchable
            :clearable="false"
            :disabled="isEdit">
            <bcs-option
              v-for="cpuItem in getSchemaByProp('launchTemplate.CPU').enum"
              :key="cpuItem"
              :id="cpuItem"
              :name="cpuItem">
            </bcs-option>
          </bcs-select>
          <span class="company">{{$t('核')}}</span>
        </div>
        <div class="prefix-select ml30">
          <span class="prefix">{{$t('内存')}}</span>
          <bcs-select
            v-model="nodePoolConfig.launchTemplate.Mem"
            searchable
            :clearable="false"
            :disabled="isEdit">
            <bcs-option
              v-for="memItem in getSchemaByProp('launchTemplate.Mem').enum"
              :key="memItem"
              :id="memItem"
              :name="memItem">
            </bcs-option>
          </bcs-select>
          <span class="company">G</span>
        </div>
      </div>
      <bcs-table
        :data="instanceList"
        v-bkloading="{ isLoading: instanceTypesLoading }"
        :pagination="pagination"
        :row-class-name="instanceRowClass"
        @page-change="pageChange"
        @page-limit-change="pageSizeChange"
        @row-click="handleCheckInstanceType">
        <bcs-table-column :label="$t('机型')" prop="typeName" show-overflow-tooltip>
          <template #default="{ row }">
            <bcs-radio
              :value="nodePoolConfig.launchTemplate.instanceType === row.nodeType"
              :disabled="row.status === 'SOLD_OUT' || isEdit">
              <span class="bcs-ellipsis">{{row.typeName}}</span>
            </bcs-radio>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('规格')" prop="nodeType"></bcs-table-column>
        <bcs-table-column label="CPU" prop="cpu" width="60" align="right">
          <template #default="{ row }">
            <span>{{ `${row.cpu}${$t('核')}` }}</span>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('内存')" prop="memory" width="60" align="right">
          <template #default="{ row }">
            <span>{{ row.memory }}G</span>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('配置费用')" prop="unitPrice">
          <template #default="{ row }">
            {{ $t('￥{price}元/小时起', { price: row.unitPrice }) }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('状态')" width="80">
          <template #default="{ row }">
            {{ row.status === 'SELL' ? $t('售卖') : $t('售罄') }}
          </template>
        </bcs-table-column>
      </bcs-table>
      <div class="mt25" style="display:flex;align-items:center;">
        <div class="prefix-select">
          <span class="prefix">{{$t('系统盘')}}</span>
          <bcs-select
            v-model="nodePoolConfig.launchTemplate.systemDisk.diskType"
            :disabled="isEdit"
            :clearable="false"
            class="min-width-150">
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
          min="50"
          max="1024"
          v-model="nodePoolConfig.launchTemplate.systemDisk.diskSize"
          :disabled="isEdit">
          <div slot="append" class="group-text">GB</div>
        </bk-input>
      </div>
      <div class="mt20">
        <bk-checkbox
          v-model="showDataDisks"
          :disabled="isEdit"
          @change="handleShowDataDisksChange">
          {{$t('购买数据盘')}}
        </bk-checkbox>
      </div>
      <template v-if="showDataDisks">
        <div class="panel" v-for="(disk, index) in nodePoolConfig.launchTemplate.dataDisks" :key="index">
          <div class="panel-item">
            <div class="prefix-select">
              <span class="prefix">{{$t('数据盘')}}</span>
              <bcs-select
                :disabled="isEdit"
                v-model="disk.diskType"
                :clearable="false"
                class="min-width-150">
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
              min="10"
              max="32000"
              v-model="disk.diskSize">
              <div slot="append" class="group-text">GB</div>
            </bk-input>
          </div>
          <p class="error-tips" v-if="disk.diskSize % 10 !== 0">{{$t('范围: 10~32000, 步长: 10')}}</p>
          <div class="panel-item mt10">
            <bk-checkbox
              :disabled="isEdit"
              v-model="disk.autoFormatAndMount">
              {{$t('格式化并挂载于')}}
            </bk-checkbox>
            <template v-if="disk.autoFormatAndMount">
              <bcs-select
                class="min-width-80 ml10"
                :disabled="isEdit"
                v-model="disk.fileSystem"
                :clearable="false">
                <bcs-option
                  v-for="fileSystem in getSchemaByProp('launchTemplate.dataDisks.fileSystem').enum"
                  :key="fileSystem"
                  :id="fileSystem"
                  :name="fileSystem">
                </bcs-option>
              </bcs-select>
              <bcs-input class="ml10" :disabled="isEdit" v-model="disk.mountTarget"></bcs-input>
            </template>
          </div>
          <p class="error-tips" v-if="showRepeatMountTarget(index)">{{$t('目录不能重复')}}</p>
          <span :class="['panel-delete', { disabled: isEdit }]" @click="handleDeleteDiskData(index)">
            <i class="bk-icon icon-close3-shape"></i>
          </span>
        </div>
        <div :class="['add-panel-btn', { disabled: isEdit }]" @click="handleAddDiskData">
          <i class="bk-icon left-icon icon-plus"></i>
          <span>{{$t('添加数据盘')}}</span>
        </div>
      </template>
      <div class="mt15">
        <bk-checkbox
          :disabled="isEdit"
          v-model="nodePoolConfig.launchTemplate.internetAccess.publicIPAssigned">
          {{$t('分配免费公网IP')}}
        </bk-checkbox>
      </div>
      <div class="panel" v-if="nodePoolConfig.launchTemplate.internetAccess.publicIPAssigned">
        <div class="panel-item">
          <label class="label">{{$t('计费方式')}}</label>
          <bk-radio-group v-model="nodePoolConfig.launchTemplate.internetAccess.internetChargeType">
            <bk-radio :disabled="isEdit" value="TRAFFIC_POSTPAID_BY_HOUR">{{$t('按使用流量计费')}}</bk-radio>
            <bk-radio :disabled="isEdit" value="BANDWIDTH_PREPAID">{{$t('按带宽计费')}}</bk-radio>
          </bk-radio-group>
        </div>
        <div class="panel-item mt10">
          <label class="label">{{$t('购买宽带')}}</label>
          <bk-input
            class="max-width-150"
            type="number"
            :disabled="isEdit"
            v-model="nodePoolConfig.launchTemplate.internetAccess.internetMaxBandwidth">
            <div slot="append" class="group-text">Mbps</div>
          </bk-input>
        </div>
      </div>
    </bk-form-item>
    <bk-form-item
      :label="$t('设置root密码')"
      property="nodePoolConfig.launchTemplate.initLoginPassword"
      error-display-type="normal"
      required>
      <bcs-input
        type="password"
        :disabled="isEdit"
        v-model="nodePoolConfig.launchTemplate.initLoginPassword"></bcs-input>
    </bk-form-item>
    <bk-form-item
      :label="$t('确认root密码')"
      property="confirmPassword"
      error-display-type="normal">
      <bcs-input type="password" :disabled="isEdit" v-model="confirmPassword"></bcs-input>
    </bk-form-item>
    <bk-form-item
      :label="$t('安全组')"
      property="nodePoolConfig.launchTemplate.securityGroupIDs"
      error-display-type="normal"
      required>
      <bcs-select
        :loading="securityGroupsLoading"
        v-model="nodePoolConfig.launchTemplate.securityGroupIDs"
        multiple
        searchable
        :disabled="isEdit">
        <bcs-option
          v-for="securityGroup in securityGroupsList"
          :key="securityGroup.securityGroupID"
          :id="securityGroup.securityGroupID"
          :name="securityGroup.securityGroupName">
        </bcs-option>
      </bcs-select>
    </bk-form-item>
    <bk-form-item
      :label="$t('支持子网')"
      property="nodePoolConfig.autoScaling.subnetIDs"
      error-display-type="normal"
      required>
      <bcs-table
        :data="subnetsList"
        :row-class-name="subnetsRowClass"
        v-bkloading="{ isLoading: subnetsLoading }"
        @row-click="handleCheckSubnets">
        <bcs-table-column :label="$t('子网ID')" min-width="150">
          <template #default="{ row }">
            <bk-checkbox
              :disabled="!(curInstanceItem.zones && curInstanceItem.zones.includes(row.zone)) || isEdit"
              :value="nodePoolConfig.autoScaling.subnetIDs.includes(row.subnetID)">
              <span class="bcs-ellipsis">{{row.subnetID}}</span>
            </bk-checkbox>
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('子网名称')" prop="subnetName"></bcs-table-column>
        <bcs-table-column :label="$t('可用区')" prop="zone" show-overflow-tooltip></bcs-table-column>
        <bcs-table-column :label="$t('剩余IP数')" prop="availableIPAddressCount" align="right"></bcs-table-column>
        <bcs-table-column :label="$t('不可用原因')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ curInstanceItem.zones && curInstanceItem.zones.includes(row.zone)
              ? '--' : $t('该可用区暂不支持该机型') }}
          </template>
        </bcs-table-column>
      </bcs-table>
    </bk-form-item>
    <bk-form-item :label="$t('容器目录')" :desc="$t('设置容器和镜像存储目录，建议存储到数据盘')">
      <bcs-input :disabled="isEdit" v-model="nodePoolConfig.nodeTemplate.dockerGraphPath"></bcs-input>
    </bk-form-item>
    <bk-form-item :label="$t('运行时组件')">
      <bk-radio-group v-model="nodePoolConfig.nodeTemplate.runtime.containerRuntime" :disabled="isEdit">
        <bk-radio value="docker" :disabled="isEdit">docker</bk-radio>
        <bk-radio value="containerd" :disabled="isEdit">containerd</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item :label="$t('运行时版本')">
      <bcs-select
        v-model="nodePoolConfig.nodeTemplate.runtime.runtimeVersion"
        :clearable="false"
        :disabled="isEdit">
        <bcs-option
          v-for="version in versionList"
          :key="version"
          :id="version"
          :name="version">
        </bcs-option>
      </bcs-select>
    </bk-form-item>
    <bk-form-item
      :label="$t('自定义脚本')"
      :desc="$t('指定自定义脚本来配置Node，即当Node启动后运行配置的脚本，需要自行保证脚本的可重入及重试逻辑, 脚本及其生成的日志文件可在节点的/usr/local/qcloud/tke/userscript路径查看')">
      <bcs-input
        type="textarea"
        :disabled="isEdit"
        :rows="6"
        v-model="nodePoolConfig.nodeTemplate.userScript">
      </bcs-input>
    </bk-form-item>
    <bk-form-item v-if="!isEdit">
      <bcs-button @click="handlePre">{{$t('上一步')}}</bcs-button>
      <bcs-button
        theme="primary"
        :loading="saveLoading"
        @click="handleSaveNodePoolData">
        {{isEdit ? $t('保存节点池') : $t('创建节点池')}}
      </bcs-button>
      <bcs-button @click="handleCancel">{{$t('取消')}}</bcs-button>
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs, watch } from 'vue';
import $router from '@/router';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';
import usePage from '@/composables/use-page';
import { mergeDeep } from '@/common/util';
import Schema from '@/views/cluster-manage/cluster/autoscaler/resolve-schema';

export default defineComponent({
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
    nodePoolInfo: {
      type: Object,
      default: () => ({}),
    },
  },
  setup(props, ctx) {
    const { defaultValues, cluster, isEdit, schema, nodePoolInfo } = toRefs(props);
    const formRef = ref<any>(null);
    // 磁盘类型
    const diskEnum = ref([
      {
        id: 'CLOUD_PREMIUM',
        name: $i18n.t('高性能云硬盘'),
      },
      {
        id: 'CLOUD_SSD',
        name: $i18n.t('SSD云硬盘'),
      },
      {
        id: 'CLOUD_HSSD',
        name: $i18n.t('增强型SSD云硬盘'),
      },
    ]);
    const confirmPassword = ref(''); // 确认密码
    const nodePoolConfig = ref({
      nodeGroupID: defaultValues.value.nodeGroupID, // 编辑时
      launchTemplate: {
        imageInfo: {
          imageID: defaultValues.value.launchTemplate.imageInfo.imageID, // 镜像ID
        },
        CPU: defaultValues.value.launchTemplate.CPU,
        Mem: defaultValues.value.launchTemplate.Mem,
        instanceType: defaultValues.value.launchTemplate.instanceType, // 机型信息
        systemDisk: {
          diskType: defaultValues.value.launchTemplate.systemDisk.diskType, // 系统盘类型
          diskSize: defaultValues.value.launchTemplate.systemDisk.diskSize, // 系统盘大小
        },
        internetAccess: {
          internetChargeType: defaultValues.value.launchTemplate.internetAccess.internetChargeType, // 计费方式
          internetMaxBandwidth: defaultValues.value.launchTemplate.internetAccess.internetMaxBandwidth, // 购买带宽
          publicIPAssigned: defaultValues.value.launchTemplate.internetAccess.publicIPAssigned, // 分配免费公网IP
        },
        initLoginPassword: defaultValues.value.launchTemplate.initLoginPassword, // 密码
        securityGroupIDs: defaultValues.value.launchTemplate.securityGroupIDs || [], // 安全组
        dataDisks: defaultValues.value.launchTemplate.dataDisks || [], // 数据盘
        // 默认值
        isSecurityService: defaultValues.value.launchTemplate.isSecurityService || true,
        isMonitorService: defaultValues.value.launchTemplate.isMonitorService || true,
      },
      autoScaling: {
        subnetIDs: defaultValues.value.autoScaling.subnetIDs || [], // 支持子网
      },
      nodeTemplate: {
        dataDisks: defaultValues.value.nodeTemplate.dataDisks || [],
        dockerGraphPath: defaultValues.value.nodeTemplate.dockerGraphPath, // 容器目录
        userScript: defaultValues.value.nodeTemplate.userScript, // 自定义数据
        runtime: {
          containerRuntime: defaultValues.value.nodeTemplate.runtime?.containerRuntime || 'docker', // 运行时容器组件
          runtimeVersion: defaultValues.value.nodeTemplate.runtime?.runtimeVersion || '19.3', // 运行时版本
        },
      },

    });

    const nodePoolConfigRules = ref({
      // 镜像ID校验
      'launchTemplate.imageInfo.imageID': [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      // 密码校验
      'nodePoolConfig.launchTemplate.initLoginPassword': [
        {
          message: $i18n.t('必填项'),
          trigger: 'blur',
          validator: () => isEdit.value || nodePoolConfig.value.launchTemplate.initLoginPassword.length > 0,
        },
        {
          message: $i18n.t('linux机器密码需8到30位，推荐使用12位以上的密码；且必须包含大写字母，小写字母和数字'),
          trigger: 'blur',
          validator: () => isEdit.value
                                || /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[^]{8,30}$/.test(nodePoolConfig.value.launchTemplate.initLoginPassword),
        },
      ],
      confirmPassword: [
        {
          message: $i18n.t('两次输入的密码不一致'),
          trigger: 'blur',
          validator: () => confirmPassword.value === nodePoolConfig.value.launchTemplate.initLoginPassword,
        },
      ],
      // 安全组校验
      'nodePoolConfig.launchTemplate.securityGroupIDs': [
        {
          message: $i18n.t('必填项'),
          trigger: 'blur',
          validator: () => !!nodePoolConfig.value.launchTemplate.securityGroupIDs.length,
        },
      ],
      // 子网校验
      'nodePoolConfig.autoScaling.subnetIDs': [
        {
          message: $i18n.t('必填项'),
          trigger: 'blur',
          validator: () => !!nodePoolConfig.value.autoScaling.subnetIDs.length,
        },
      ],
    });

    // 镜像
    const imageProvider = ref('PUBLIC_IMAGE');
    watch(imageProvider, () => {
      // 重置镜像ID
      nodePoolConfig.value.launchTemplate.imageInfo.imageID = '';
      // 获取镜像列表
      handleGetOsImage();
    });
    const osImageLoading = ref(false);
    const osImageList = ref<any[]>([]);
    const handleGetOsImage = async () => {
      osImageLoading.value = true;
      osImageList.value = await $store.dispatch('clustermanager/cloudOsImage', {
        $cloudID: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
        provider: imageProvider.value,
      });
      osImageLoading.value = false;
    };

    // 机型
    const instanceTypesLoading = ref(false);
    const instanceTypesList = ref<any[]>([]);
    const curInstanceItem = computed(() => instanceTypesList.value
      .find(instance => instance.nodeType === nodePoolConfig.value.launchTemplate.instanceType) || {});
    watch(() => [
      nodePoolConfig.value.launchTemplate.Mem,
      nodePoolConfig.value.launchTemplate.CPU,
    ], () => {
      // 重置机型
      nodePoolConfig.value.launchTemplate.instanceType = '';
      // 获取机型
      handleGetInstanceTypes();
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
    const handleGetInstanceTypes = async () => {
      instanceTypesLoading.value = true;
      instanceTypesList.value = await $store.dispatch('clustermanager/cloudInstanceTypes', {
        $cloudID: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
        provider: imageProvider.value,
        cpu: nodePoolConfig.value.launchTemplate.CPU,
        memory: nodePoolConfig.value.launchTemplate.Mem,
      });
      // 默认机型配置
      if (!nodePoolConfig.value.launchTemplate.instanceType) {
        nodePoolConfig.value.launchTemplate.instanceType = instanceTypesList.value
          .find(instance => instance.status === 'SELL')?.nodeType;
      }
      instanceTypesLoading.value = false;
    };
    const {
      pagination,
      curPageData: instanceList,
      pageChange,
      pageSizeChange,
    } = usePage(instanceTypesList);

    // 数据盘
    const showDataDisks = ref(!!nodePoolConfig.value.launchTemplate.dataDisks.length);
    const dataDisksSchema = Schema.getSchemaByProp(schema.value, 'launchTemplate.dataDisks')?.items || {};
    const defaultDiskItem = Schema.getSchemaDefaultValue(dataDisksSchema);
    const handleShowDataDisksChange = (show) => {
      nodePoolConfig.value.launchTemplate.dataDisks = show
        ? [JSON.parse(JSON.stringify(defaultDiskItem))] : [];
    };
    const handleDeleteDiskData = (index) => {
      if (isEdit.value) return;
      nodePoolConfig.value.launchTemplate.dataDisks.splice(index, 1);
    };
    const handleAddDiskData = () => {
      if (isEdit.value) return;
      nodePoolConfig.value.launchTemplate.dataDisks.push(JSON.parse(JSON.stringify(defaultDiskItem)));
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
      securityGroupsList.value = await $store.dispatch('clustermanager/cloudSecurityGroups', {
        $cloudID: cluster.value.provider,
        region: cluster.value.region,
        accountID: cluster.value.cloudAccountID,
      });
      securityGroupsLoading.value = false;
    };

    // VPC子网
    const subnetsLoading = ref(false);
    const subnetsList = ref([]);
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
    const user = computed(() => $store.state.user);
    const handlePre = () => {
      ctx.emit('pre');
    };
    const saveLoading = ref(false);
    const handleSaveNodePoolData = async () => {
      const result = await formRef.value?.validate();
      const validateDataDiskSize = nodePoolConfig.value.launchTemplate.dataDisks
        .every(item => item.diskSize % 10 === 0);
      const mountTargetList = nodePoolConfig.value.launchTemplate.dataDisks.map(item => item.mountTarget);
      const validateDataDiskMountTarget = new Set(mountTargetList).size === mountTargetList.length;
      if (!result || !validateDataDiskSize || !validateDataDiskMountTarget) return;

      // 系统盘、数据盘、宽度大小要转换为字符串类型
      // eslint-disable-next-line max-len
      nodePoolConfig.value.launchTemplate.systemDisk.diskSize = String(nodePoolConfig.value.launchTemplate.systemDisk.diskSize);
      nodePoolConfig.value.launchTemplate.dataDisks = nodePoolConfig.value.launchTemplate.dataDisks.map(item => ({
        ...item,
        diskSize: String(item.diskSize),
      }));
      // eslint-disable-next-line max-len
      nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth = String(nodePoolConfig.value.launchTemplate.internetAccess.internetMaxBandwidth);

      // 数据盘后端存了两个地方
      nodePoolConfig.value.nodeTemplate.dataDisks = nodePoolConfig.value.launchTemplate.dataDisks;

      saveLoading.value = true;
      if (isEdit.value) {
        await handleEditNodePool();
      } else {
        await handleCreateNodePool();
      }
      saveLoading.value = false;
    };
    const handleEditNodePool = async () => {
      const data = {
        $nodeGroupID: nodePoolConfig.value.nodeGroupID,
        ...mergeDeep(nodePoolInfo.value, nodePoolConfig.value),
        clusterID: cluster.value.clusterID,
        region: cluster.value.region,
        updater: user.value.username,
      };
      console.log(data);
      const result = await $store.dispatch('clustermanager/updateNodeGroup', data);
      if (result) {
        $router.push({
          name: 'clusterDetail',
          query: {
            active: 'autoscaler',
          },
        });
      }
    };
    const handleCreateNodePool = async () => {
      const data = {
        ...mergeDeep(nodePoolInfo.value, nodePoolConfig.value),
        clusterID: cluster.value.clusterID,
        region: cluster.value.region,
        creator: user.value.username,
      };
      console.log(data);
      const result = await $store.dispatch('clustermanager/createNodeGroup', data);
      if (result) {
        $router.push({
          name: 'clusterDetail',
          query: {
            active: 'autoscaler',
          },
        });
      }
    };
    const handleCancel = () => {
      $router.back();
    };

    const getSchemaByProp = props => Schema.getSchemaByProp(schema.value, props);

    const showRepeatMountTarget = (index) => {
      const disk = nodePoolConfig.value.launchTemplate.dataDisks[index];
      return disk.autoFormatAndMount
                    && disk.mountTarget
                    && nodePoolConfig.value.launchTemplate.dataDisks
                      .filter((item, i) => i !== index && item.autoFormatAndMount)
                      .some(item => item.mountTarget === disk.mountTarget);
    };
    onMounted(() => {
      handleGetOsImage();
      handleGetInstanceTypes();
      handleGetCloudSecurityGroups();
      handleGetSubnets();
    });
    return {
      versionList,
      formRef,
      diskEnum,
      imageProvider,
      confirmPassword,
      nodePoolConfig,
      nodePoolConfigRules,
      osImageLoading,
      osImageList,
      curInstanceItem,
      instanceTypesLoading,
      instanceList,
      pagination,
      pageChange,
      pageSizeChange,
      instanceRowClass,
      handleCheckInstanceType,
      showDataDisks,
      handleShowDataDisksChange,
      handleDeleteDiskData,
      handleAddDiskData,
      securityGroupsLoading,
      securityGroupsList,
      subnetsLoading,
      subnetsRowClass,
      subnetsList,
      handleCheckSubnets,
      handleGetOsImage,
      handlePre,
      saveLoading,
      handleSaveNodePoolData,
      handleCancel,
      getSchemaByProp,
      showRepeatMountTarget,
    };
  },
});
</script>
<style lang="postcss" scoped>
.node-config {
    font-size: 14px;
    >>> .group-text {
        line-height: 30px;
    }
    >>> .bk-form-content {
        max-width: 600px;
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
        }
        .company {
            display: inline-block;
            width: 30px;
            height: 32px;
            border: 1px solid #C4C6CC;
            text-align: center;
            border-left: none;
        }
        >>> .bk-select {
            min-width: 110px;
            margin-left: -1px;
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
        .bk-select {
            background: #fff;
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
