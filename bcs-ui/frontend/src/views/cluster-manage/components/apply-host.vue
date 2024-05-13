<!-- eslint-disable vue/no-v-html -->
<template>
  <div class="apply-host-wrapper" v-if="$INTERNAL">
    <div class="apply-host-btn flex items-center">
      <bk-popover placement="bottom" v-if="!hasAuth" transfer>
        <span class="bk-default bk-button-normal bk-button is-disabled">{{title}}</span>
        <div slot="content" style="width: 220px;">
          {{ authTips.content }}
        </div>
      </bk-popover>
      <template v-else>
        <bk-button
          :theme="theme"
          @click="handleOpenApplyHost">
          {{title}}
        </bk-button>
        <span
          class="bcs-icon-btn apply-record flex items-center justify-center w-[32px] h-[32px] ml-[-1px] bg-[#fff]"
          v-bk-tooltips="$t('cluster.button.applyRecord')"
          @click="handleOpenLink">
          <i class="bcs-icon bcs-icon-wenjian"></i>
        </span>
      </template>
    </div>
    <bcs-dialog
      :position="{ top: 80 }"
      v-model="applyDialogShow"
      :close-icon="false"
      :width="1000"
      :title="title"
      render-directive="if"
      header-position="left"
      ext-cls="apply-host-dialog">
      <bk-alert type="info" class="mb20">
        <template #title>
          <a
            :href="PROJECT_CONFIG.applyrecords"
            target="_blank"
            class="bk-button-text bk-primary"
            style="font-size: 12px;"
          >{{$t('cluster.button.applyRecord')}}</a>
        </template>
      </bk-alert>
      <bk-form
        ext-cls="apply-form"
        ref="applyForm"
        :label-width="100"
        :model="formdata"
        :rules="rules">
        <bk-form-item
          property="region"
          :label="$t('cluster.labels.region')"
          :required="true" :desc="defaultInfo.areaDesc">
          <bcs-select
            :placeholder="$t('cluster.placeholder.region')"
            v-model="formdata.region"
            searchable
            id-key="areaName"
            display-key="showName"
            :loading="isAreaLoading"
            :disabled="defaultInfo.disabled || isHostLoading">
            <bcs-option
              v-for="item in areaList"
              :key="item.region"
              :id="item.region"
              :name="item.regionName" />
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          property="networkKey"
          :label="$t('cluster.labels.networkType')"
          :desc="defaultInfo.netWorkDesc" :required="true">
          <div class="bk-button-group">
            <bcs-button
              :disabled="defaultInfo.networkKey && defaultInfo.networkKey !== 'overlay'"
              :class="{
                'active': formdata.networkKey === 'overlay',
                'network-btn': true,
                'network-zIndex': defaultInfo.networkKey === 'overlay'
              }"
              @click="formdata.networkKey = 'overlay'">overlay</bcs-button>
            <bcs-button
              :disabled="defaultInfo.networkKey && defaultInfo.networkKey !== 'underlay'"
              :class="{
                'active': formdata.networkKey === 'underlay',
                'network-btn': true,
                'network-zIndex': defaultInfo.networkKey === 'underlay'
              }"
              @click="formdata.networkKey = 'underlay'">underlay</bcs-button>
          </div>
        </bk-form-item>
        <bk-form-item property="zone_id" :label="$t('generic.applyHost.label.zone')" :required="true">
          <bcs-select
            :placeholder="$t('generic.applyHost.placeholder.zone')"
            v-model="formdata.zone_id"
            searchable
            :clearable="false"
            :disabled="isHostLoading"
            :loading="zoneLoading">
            <bcs-option
              v-for="item in zoneList"
              :key="item.value"
              :id="item.value"
              :name="item.label" />
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          property="vpc_name"
          :label="$t('cluster.create.label.vpc.text')"
          :required="true" :desc="defaultInfo.vpcDesc">
          <bcs-select
            :placeholder="$t('generic.applyHost.placeholder.vpc')"
            v-model="formdata.vpc_name"
            :list="vpcList"
            searchable
            :clearable="false"
            :disabled="defaultInfo.disabled"
            :loading="vpcLoading">
            <bcs-option v-for="item in vpcList" :key="item.vpcId" :id="item.vpcId" :name="item.vpcName" />
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          ext-cls="has-append-item"
          :label="$t('cluster.ca.nodePool.create.instanceTypeConfig.disk.data')"
          property="disk_size">
          <div class="disk-inner">
            <bcs-select class="w200" v-model="formdata.disk_type" :clearable="false">
              <bcs-option
                v-for="item in diskTypeList"
                :key="item.value"
                :id="item.value"
                :name="item.label"
              >
              </bcs-option>
            </bcs-select>
            <bk-input
              class="ml-[-1px]"
              v-model="formdata.disk_size" type="number"
              :min="50"
              :placeholder="$t('generic.applyHost.validate.diskSize')">
              <div class="group-text" slot="append">GB</div>
            </bk-input>
          </div>
        </bk-form-item>
        <bk-form-item :label="$t('generic.applyHost.label.replicas')">
          <bk-input
            v-model="formdata.replicas"
            type="number"
            :min="1"
            :max="50"
            :style="{ 'width': '100%' }"
            :placeholder="$t('generic.placeholder.input')">
          </bk-input>
        </bk-form-item>
        <bk-form-item class="custom-item" :label="$t('generic.ipSelector.label.serverModel')">
          <div class="form-item-inner">
            <label class="inner-label">CPU</label>
            <div class="inner-content">
              <bcs-select
                v-model="hostData.cpu"
                searchable
                :clearable="false">
                <bcs-option v-for="item in cpuList" :key="item.id" :id="item.id" :name="item.name" />
              </bcs-select>
            </div>
          </div>
          <div :class="['form-item-inner', !isEn && 'ml40']" :style="{ width: isEn ? '310px' : '286px' }">
            <label class="inner-label">{{$t('generic.label.mem')}}</label>
            <div class="inner-content">
              <bcs-select
                v-model="hostData.mem"
                searchable
                :clearable="false">
                <bcs-option v-for="item in memList" :key="item.id" :id="item.id" :name="item.name" />
              </bcs-select>
            </div>
          </div>
          <bk-button
            theme="primary"
            :disabled="isHostLoading"
            @click.stop="hanldeReloadHosts">
            {{$t('generic.button.query')}}
          </bk-button>
        </bk-form-item>
        <bk-form-item
          ref="hostItem"
          style="flex: 0 0 100%;margin-bottom: 0px"
          label=""
          :label-width="1"
          v-bkloading="{ isLoading: isHostLoading }"
          :required="true"
          property="cvm_type"
          class="host-item">
          <bk-radio-group v-model="formdata.cvm_type">
            <bk-table :data="hostTableData" :max-height="320" style="overflow-y: hidden;">
              <bk-table-column label="" width="40" :resizable="false">
                <template slot-scope="{ row }">
                  <span
                    v-bk-tooltips="{
                      content: $t('generic.applyHost.tips.insufficientResource'),
                      disabled: !((cvmData[row.specifications] || 0) < formdata.replicas) || isCvmLoading
                    }">
                    <bk-radio
                      name="host"
                      :value="row.specifications"
                      :disabled="((cvmData[row.specifications] || 0) < formdata.replicas) || isCvmLoading"
                      @change="handleRadioChange(row)">
                    </bk-radio>
                  </span>
                </template>
              </bk-table-column>
              <bk-table-column
                :label="$t('generic.ipSelector.label.serverModel')"
                prop="model" show-overflow-tooltip></bk-table-column>
              <bk-table-column
                :label="$t('generic.label.specifications')"
                prop="specifications" show-overflow-tooltip></bk-table-column>
              <bk-table-column :label="$t('generic.applyHost.label.zone')" prop="zone" key="zone">
                {{ zoneName }}
              </bk-table-column>
              <bk-table-column label="CPU" prop="cpu" width="80"></bk-table-column>
              <bk-table-column :label="$t('generic.label.mem')" prop="mem" width="80"></bk-table-column>
              <bk-table-column :label="$t('generic.applyHost.label.specifications')">
                <template #default="{ row }">
                  <LoadingIcon v-if="isCvmLoading">{{ $t('generic.status.loading') }}...</LoadingIcon>
                  <span v-else>{{ cvmData[row.specifications] }}</span>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('generic.label.memo')" prop="description" show-overflow-tooltip>
                <template #default="{ row }">
                  {{ row.description || '--' }}
                </template>
              </bk-table-column>
            </bk-table>
          </bk-radio-group>
          <span class="checked-host-tips" style="height: 30px;">
            {{formdata.cvm_type ? $t('generic.applyHost.title.selected') + '：' + getHostInfoString : ' '}}
          </span>
        </bk-form-item>
      </bk-form>
      <template slot="footer">
        <i18n v-show="isShowFooterTips" class="tips" target path="generic.applyHost.msg.path">
          <a href="wxwork://message/?username=dommyzhang" style="color: #3A84FF;" place="name">dommyzhang</a>
        </i18n>
        <bk-button
          theme="primary"
          :loading="isSubmitLoading"
          @click.stop="handleSubmitApply">
          {{$t('generic.button.confirm')}}
        </bk-button>
        <bk-button
          theme="default"
          :disabled="isSubmitLoading"
          @click.stop="handleApplyHostClose">{{$t('generic.button.cancel')}}</bk-button>
      </template>
    </bcs-dialog>
  </div>
</template>

<script>
/* eslint-disable max-len */
import http from '@/api';
import { request } from '@/api/request';
import LoadingIcon from '@/components/loading-icon.vue';

export default {
  components: { LoadingIcon },
  props: {
    theme: {
      type: String,
      default: 'default',
    },
    isBackfill: {
      type: Boolean,
      default: false,
    },
    clusterId: {
      type: String,
      default: '',
    },
    title: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      cvmData: {},
      timer: null,
      isSubmitLoading: false,
      isAreaLoading: false,
      isHostLoading: false,
      isCvmLoading: false,
      vpcLoading: false,
      zoneLoading: false,
      applyDialogShow: false,
      areaList: [],
      vpcList: [],
      zoneList: [],
      diskTypeList: [],
      formdata: {
        region: '',
        disk_size: 50,
        replicas: 1,
        cvm_type: '',
        vpc_name: '',
        zone_id: '',
        disk_type: '',
        networkKey: 'overlay',
      },
      isQuery: false,
      rules: {
        region: [{
          required: true,
          trigger: 'blur',
          message: this.$t('cluster.placeholder.region'),
        }],
        disk_size: [{
          validator: value => value >= 50 && value % 50 === 0,
          trigger: 'blur',
          message: this.$t('generic.applyHost.validate.diskSize'),
        }],
        cvm_type: [{
          validator: value => this.isQuery || !!value,
          trigger: 'change',
          message: this.$t('generic.applyHost.tips.emptyCvm'),
        }],
      },
      hostData: {
        cpu: 0,
        mem: 0,
      },
      checkedHostInfo: {},
      hostTableData: [],
      defaultInfo: {
        areaDesc: '',
        vpcDesc: '',
        netWorkDesc: '',
        disabled: false,
      },
      clusterInfo: {},
      cpuList: [{
        id: 0,
        name: this.$t('generic.label.total'),
      }, {
        id: 4,
        name: 4 + this.$t('units.suffix.cores'),
      }, {
        id: 8,
        name: 8 + this.$t('units.suffix.cores'),
      }, {
        id: 12,
        name: 12 + this.$t('units.suffix.cores'),
      }, {
        id: 16,
        name: 16 + this.$t('units.suffix.cores'),
      }, {
        id: 20,
        name: 20 + this.$t('units.suffix.cores'),
      }, {
        id: 24,
        name: 24 + this.$t('units.suffix.cores'),
      }, {
        id: 28,
        name: 28 + this.$t('units.suffix.cores'),
      }, {
        id: 32,
        name: 32 + this.$t('units.suffix.cores'),
      }, {
        id: 48,
        name: 48 + this.$t('units.suffix.cores'),
      }, {
        id: 84,
        name: 84 + this.$t('units.suffix.cores'),
      }],
      memList: [{
        id: 0,
        name: this.$t('generic.label.total'),
      }, {
        id: 8,
        name: '8GiB',
      }, {
        id: 16,
        name: '16GiB',
      }, {
        id: 24,
        name: '24GiB',
      }, {
        id: 32,
        name: '32GiB',
      }, {
        id: 36,
        name: '36GiB',
      }, {
        id: 48,
        name: '48GiB',
      }, {
        id: 56,
        name: '56GiB',
      }, {
        id: 60,
        name: '60GiB',
      }, {
        id: 64,
        name: '64GiB',
      }, {
        id: 80,
        name: '80GiB',
      }, {
        id: 128,
        name: '128GiB',
      }, {
        id: 160,
        name: '160GiB',
      }, {
        id: 320,
        name: '320GiB',
      }],
    };
  },
  computed: {
    maintainers() {
      return this.$store.state.cluster.maintainers || [];
    },
    curProject() {
      return this.$store.state.curProject;
    },
    projectId() {
      return this.$store.getters.curProjectId;
    },
    getHostInfoString() {
      if (!this.formdata.cvm_type) return '';
      return `${this.checkedHostInfo.specifications} （${this.checkedHostInfo.model}，${this.checkedHostInfo.cpu + this.checkedHostInfo.mem}）`;
    },
    isEn() {
      return this.$store.state.isEn;
    },
    isShowFooterTips() {
      if (!this.formdata.cvm_type) return false;
      if (!this.checkedHostInfo.cpu) return false;
      return this.formdata.disk_size / parseInt(this.checkedHostInfo.cpu) > 50;
    },
    authTips() {
      return {
        content: this.$t('bcs.msg.notDevOps'),
        width: 240,
      };
    },
    userInfo() {
      return this.$store.state.user;
    },
    zoneName() {
      const zone = this.zoneList.find(item => item.value === this.formdata.zone_id) || {};
      return zone.label || '--';
    },
    hasAuth() {
      return this.maintainers.includes(this.userInfo.username);
    },
  },
  watch: {
    'formdata.networkKey'(val, old) {
      val && val !== old && this.changeNetwork();
    },
    'formdata.cvm_type'() {
      this.$refs.applyForm?.$refs.hostItem?.clearError();
    },
    'formdata.region': {
      immediate: true,
      async handler(value, old) {
        if (value !== old) {
          this.formdata.vpc_name = '';
          this.vpcList = [];
          await this.fetchZone();
          await this.fetchVPC();
        }
      },
    },
    clusterId: {
      immediate: true,
      async handler(value, old) {
        if (value && value !== old) {
          await this.fetchClusterInfo();
        }
      },
    },
  },
  async created() {
    if (!this.hasAuth) return;
    if (this.isBackfill) {
      this.defaultInfo = {
        areaDesc: this.$t('generic.applyHost.desc.regionDesc'),
        vpcDesc: this.$t('generic.applyHost.desc.vpcDesc'),
        netWorkDesc: this.$t('generic.applyHost.desc.networkDesc'),
        disabled: true,
      };
    }
    // this.getApplyHostStatus();
    // this.fetchDiskType();
  },
  beforeDestroy() {
    clearTimeout(this.timer) && (this.timer = null);
  },
  methods: {
    /**
             * 获取当前集群数据
             */
    async fetchClusterInfo() {
      if (!this.clusterId) return;

      try {
        const res = await this.$store.dispatch('clustermanager/clusterDetail', {
          $clusterId: this.clusterId,
        });
        this.clusterInfo = res.data || {};
        if (this.clusterInfo.networkType && this.isBackfill) {
          this.formdata.networkKey = this.clusterInfo.networkType;
          this.defaultInfo.networkKey = this.formdata.networkKey;
        }
      } catch (e) {
        console.error(e);
      }
    },
    /**
             * 获取所属地域
             */
    async getAreas() {
      try {
        const list = await this.$store.dispatch('clustermanager/fetchCloudRegion', {
          $cloudId: 'tencentCloud',
        });
        this.areaList = list.map(item => ({
          ...item,
          areaId: item.cloudID,
          areaName: item.region,
          showName: item.regionName,
        }));
        if (this.clusterInfo.region && this.isBackfill) {
          const area = this.areaList.find(item => item.areaName === this.clusterInfo.region);
          if (area) {
            this.formdata.region = area.areaName;
          }
        }
        // else if (this.areaList.length) {
        //   // 默认选中第一个
        //   this.formdata.region = this.areaList[0].areaName;
        // }
      } catch (e) {
        this.areaList = [];
        console.error(e);
      } finally {
        this.isAreaLoading = false;
      }
    },

    /**
             * 选择网络类型
             */
    async changeNetwork() {
      this.vpcList = [];
      await this.fetchVPC();
    },

    /**
             * 获取园区列表
             */
    async fetchZone() {
      if (!this.formdata.region) return;

      try {
        this.zoneLoading = true;
        const data = await this.$store.dispatch('cluster/getZoneList', {
          projectId: this.projectId,
          region: this.formdata.region,
        });
        this.zoneLoading = false;
        this.zoneList = data.data;
        if (this.clusterInfo.zone_id && this.isBackfill) {
          const zone = this.zoneList.find(item => item.value === this.clusterInfo.zone_id);
          if (zone) {
            this.formdata.zone_id = zone.value;
          }
        } else if (this.zoneList.length) {
          this.formdata.zone_id = this.zoneList[0].value;
        }
      } catch (e) {
        console.error(e);
      }
    },

    /**
             * 获取数据盘类型列表
             */
    async fetchDiskType() {
      try {
        const data = await this.$store.dispatch('cluster/getDiskTypeList', {
          projectId: this.projectId,
        });
        this.diskTypeList = data.data;
        if (this.clusterInfo.disk_type && this.isBackfill) {
          const diskType = this.diskTypeList.find(item => item.value === this.clusterInfo.disk_type);
          if (diskType) {
            this.formdata.disk_type = diskType.value;
          }
        } else if (this.diskTypeList.length) {
          this.formdata.disk_type = this.diskTypeList[1].value;
        }
      } catch (e) {
        console.error(e);
      }
    },

    async fetchVPC() {
      if (!this.formdata.region) {
        return;
      }
      try {
        this.vpcLoading = true;
        const data = await this.$store.dispatch('clustermanager/fetchCloudVpc', {
          cloudID: 'tencentCloud',
          region: this.formdata.region,
          networkType: this.formdata.networkKey,
          businessID: this.curProject?.businessID,
        });
        this.vpcLoading = false;
        const vpcList = data.map(item => ({
          vpcId: item.vpcID,
          vpcName: `${item.vpcName}(${item.vpcId})`,
        }));
        this.vpcList.splice(0, this.vpcList.length, ...vpcList);
        if (this.clusterInfo.vpcID && this.isBackfill) {
          const vpc = this.vpcList.find(item => item.vpcId === this.clusterInfo.vpcID);
          if (vpc) {
            this.formdata.vpc_name = vpc.vpcId;
          } else {
            // 回填不上则直接显示当前vpc id
            this.vpcList.unshift({
              vpcId: this.clusterInfo.vpcID,
              vpcName: this.clusterInfo.vpcID,
            });
            this.formdata.vpc_name = this.clusterInfo.vpcID;
          }
        } else if (this.vpcList.length) {
          // 默认选中第一个
          this.formdata.vpc_name = this.vpcList[0].vpcId;
        }
        // this.hanldeReloadHosts();
      } catch (e) {
        console.error(e);
      }
    },
    /**
             * 获取主机列表
             */
    async getHosts() {
      this.$refs.applyForm?.$refs.hostItem?.clearError();
      try {
        this.isHostLoading = true;
        const res = await this.$store.dispatch('cluster/getSCRHosts', {
          projectId: this.projectId,
          region: this.formdata.region,
          vpc_name: this.formdata.vpc_name,
          cpu_core_num: this.hostData.cpu,
          mem_size: this.hostData.mem,
          zone_id: this.formdata.zone_id,
        });
        const list = res.data || [];
        this.hostTableData = list.map((item) => {
          const getRow = item.describe.split('; ');
          return {
            model: getRow[0],
            specifications: item.value,
            cpu: getRow[1].replace('CPU:', ''),
            mem: getRow[2].replace('MEM:', ''),
            description: getRow[3].replace('备注:', ''),
          };
        });
        // 接口很慢，异步执行
        await this.getCvmCapacity();
        this.hostTableData = this.hostTableData.sort((pre, next) => (this.cvmData[next.specifications] || 0) - (this.cvmData[pre.specifications] || 0));
        this.isHostLoading = false;
      } catch (e) {
        this.hostTableData = [];
        console.error(e);
      } finally {
        this.isHostLoading = false;
      }
    },
    /**
             * 申请服务器
             */
    async handleSubmitApply() {
      this.isQuery = false;
      const validate = await this.$refs.applyForm.validate();
      if (!validate) return;
      try {
        this.isSubmitLoading = true;
        await this.$store.dispatch('cluster/applySCRHost', Object.assign({}, this.formdata, {
          projectId: this.projectId,
        }));
        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.applyHost.msg.submit'),
        });
        this.handleApplyHostClose();
      } catch (e) {
        console.error(e);
      } finally {
        this.isSubmitLoading = false;
      }
    },
    async hanldeReloadHosts() {
      this.isQuery = true;
      const result = await this.$refs.applyForm.validate();
      if (!result) return;

      this.formdata.cvm_type = '';
      this.getHosts();
    },
    handleRadioChange(row) {
      this.checkedHostInfo = row;
    },

    /**
             * 打开申请服务器 dialog
             */
    handleOpenApplyHost() {
      // reset
      this.formdata = {
        ...this.formdata,
        region: '',
        vpc_name: '',
        disk_size: 50,
        replicas: 1,
        cvm_type: '',
        zone_id: '',
      };

      this.hostData = {
        cpu: 0,
        mem: 0,
      };

      this.hostTableData = [];
      this.applyDialogShow = true;
      this.isAreaLoading = true;
      this.getAreas();
      this.fetchDiskType();
    },

    /**
             * 关闭申请服务器 dialog
             */
    handleApplyHostClose() {
      this.applyDialogShow = false;
      clearTimeout(this.timer) && (this.timer = null);
      // this.getApplyHostStatus();
    },
    async getCvmCapacity() {
      if (!this.formdata.zone_id || !this.formdata.region || !this.$INTERNAL) return;
      // 取消上一次的请求
      await http.cancel('get_cvm_capacity_request_id');
      this.isCvmLoading = true;
      const srePrefix = `${process.env.NODE_ENV === 'development' ? '' : window.BK_SRE_HOST}`;
      const cvmCapacity = request('get', `${srePrefix}/bcsadmin/cvmcapacity`);
      this.cvmData = await cvmCapacity({
        zone_id: this.formdata.zone_id,
        region_id: this.formdata.region,
        vpc_id: this.formdata.vpc_name,
      }, { requestId: 'get_cvm_capacity_request_id' }).catch(() => ({}));
      this.isCvmLoading = false;
    },
    handleOpenLink() {
      window.open(this.PROJECT_CONFIG?.applyrecords);
    },
  },
};
</script>

<style lang="postcss" scoped>
.apply-record {
  border: 1px solid #C4C6CC;
  border-left: 1px solid #DCDEE5;
}
.apply-host-dialog {
    .bk-dialog-footer {
        .tips {
            display: inline-block;
            vertical-align: middle;
            max-width: 640px;
            font-size: 12px;
            margin-right: 8px;
            text-align: left;
        }
    }
}
.apply-form {
    display: flex;
    flex-wrap: wrap;
    /deep/ .bk-form-item {
        flex: 0 0 50%;
        margin-top: 0;
        margin-bottom: 20px;
        &.custom-item {
            flex: 0 0 100%;
            .bk-form-content {
                display: flex;
            }
        }
        &.is-error {
            .bk-selector-wrapper > input {
                border-color: #ff5656 !important;
                color: #ff5656;
            }
            .bk-selector-list input {
                border-color: #dde4eb !important;
                color: #63656e !important;
            }
            &.host-item {
                .tooltips-icon {
                    left: 16px;
                    right: unset !important;
                    top: 14px;
                }
            }
        }
        &.has-append-item {
            .tooltips-icon {
                right: 74px !important;
            }
        }
        .bk-button-group {
            width: 100%;
            display: block;
            .network-btn {
                width: 50%;
            }
            .network-zIndex {
                z-index: 1;
            }
        }
        .disk-inner {
            display: flex;
            .w200 {
                width: 200px;
            }
        }
        .form-item-inner {
            width: 326px;
            display: flex;
            margin-right: 20px;
        }
        .inner-label {
            height: 32px;
            line-height: 32px;
            margin-right: 10px;
        }
        .inner-content {
            flex: 1;
        }
        .checked-host-tips {
            display: block;
            margin-top: 10px;
        }
    }
}
</style>
