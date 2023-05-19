<template>
  <div class="choose-node-template bcs-content-wrapper">
    <bk-form>
      <FormGroup :allow-toggle="false" class="choose-node">
        <bk-form-item :label="$t('选择节点')">
          <bcs-button theme="primary" icon="plus" @click="handleAddNode">{{$t('添加节点')}}</bcs-button>
          <bcs-table class="mt15" :data="ipList">
            <bcs-table-column type="index" :label="$t('序号')" width="60"></bcs-table-column>
            <bcs-table-column :label="$t('内网IP')" prop="bk_host_innerip" width="120"></bcs-table-column>
            <bcs-table-column :label="$t('Agent状态')" width="100">
              <template #default="{ row }">
                <StatusIcon :status="String(row.agent_alive)" :status-color-map="statusColorMap">
                  {{row.agent_alive ? $t('正常') : $t('异常')}}
                </StatusIcon>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('机房')" prop="idc_unit_name"></bcs-table-column>
            <bcs-table-column :label="$t('机型')" prop="svr_device_class"></bcs-table-column>
            <bcs-table-column :label="$t('操作')" width="100">
              <template #default="{ row }">
                <bk-button text @click="handleRemoveIp(row)">{{$t('移除')}}</bk-button>
              </template>
            </bcs-table-column>
          </bcs-table>
        </bk-form-item>
        <bk-form-item :label="$t('节点初始化模板')" v-if="$INTERNAL">
          <TemplateSelector
            :is-tke-cluster="isTkeCluster"
            :cluster-id="clusterId"
            @template-change="handleTemplateChange" />
        </bk-form-item>
      </FormGroup>
    </bk-form>
    <div class="mt25">
      <span v-bk-tooltips="{ content: $t('请选择节点'), disabled: ipList.length }">
        <bk-button
          class="mw88"
          theme="primary"
          :disabled="!ipList.length"
          @click="handleShowConfirmDialog">
          {{$t('确定')}}
        </bk-button>
      </span>
      <bk-button class="mw88 ml10" @click="handleCancel">{{$t('取消')}}</bk-button>
    </div>
    <ConfirmDialog
      v-model="showConfirmDialog"
      :title="$t('确定添加节点')"
      :sub-title="$t('请确认以下配置:')"
      :tips="checkList"
      :ok-text="$t('确定添加')"
      :cancel-text="$t('取消')"
      :confirm="handleConfirm"
      theme="primary">
    </ConfirmDialog>
    <!-- IP选择器 -->
    <IpSelector
      :cloud-id="curCluster.provider"
      :region="curCluster.region"
      :vpc="{ vpcID: curCluster.vpcID }"
      v-model="showIpSelector"
      :ip-list="ipList"
      @confirm="chooseServer">
    </IpSelector>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, ref } from 'vue';
import FormGroup from '@/views/cluster-manage/cluster/create/form-group.vue';
import $router from '@/router';
import IpSelector from '@/components/ip-selector/selector-dialog.vue';
import StatusIcon from '@/components/status-icon';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store/index';
import useNode from './use-node';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
import TemplateSelector from '../components/template-selector.vue';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  components: { FormGroup, IpSelector, StatusIcon, ConfirmDialog, TemplateSelector },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const curCluster = computed(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === props.clusterId) || {});
    const isTkeCluster = computed(() => curCluster.value?.provider === 'tencentCloud');
    const statusColorMap = ref({
      0: 'red',
      1: 'green',
    });

    const showIpSelector = ref(false);
    const ipList = ref<any[]>([]);
    const handleAddNode = () => {
      showIpSelector.value = true;
    };
    const handleRemoveIp = (row) => {
      const index = ipList.value.findIndex(item => item.bk_host_innerip === row.bk_host_innerip);
      if (index > -1) {
        ipList.value.splice(index, 1);
      }
    };
    const chooseServer = (data) => {
      ipList.value = data;
      showIpSelector.value = false;
    };
    const { addNode } = useNode();
    const checkList = computed(() => [
      $i18n.t('请确认是否对 {ip} 等 {num} 个IP进行操作系统初始化和安装容器服务相关组件操作', {
        ip: ipList.value[0]?.bk_host_innerip,
        num: ipList.value.length,
      }),
      $i18n.t('为了保证集群环境标准化，添加节点会格式化数据盘/dev/vdb，盘内数据将被清除，请确认该数据盘内没有放置业务数据'),
    ]);
    const currentTemplate = ref<Record<string, string>>({});
    const handleTemplateChange = (item) => {
      currentTemplate.value = item;
    };
    const showConfirmDialog = ref(false);
    const handleShowConfirmDialog = () => {
      currentTemplate.value.nodeTemplateID
        ? $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: $i18n.t('确定使用节点模板 {name}', { name: currentTemplate.value.name }),
          defaultInfo: true,
          confirmFn: () => {
            showConfirmDialog.value = true;
          },
        })
        : showConfirmDialog.value = true;
    };
    const confirmLoading = ref(false);
    const handleConfirm = async () => {
      confirmLoading.value = true;
      const result = await addNode({
        clusterId: props.clusterId,
        nodeIps: ipList.value.map(item => item.bk_host_innerip),
        nodeTemplateID: currentTemplate.value.nodeTemplateID,
      });
      confirmLoading.value = false;

      if (result) {
        $router.push({
          name: 'clusterDetail',
          params: {
            clusterId: props.clusterId,
          },
          query: {
            active: 'node',
          },
        });
      }
    };
    const handleCancel = () => {
      $router.back();
    };

    return {
      curCluster,
      isTkeCluster,
      confirmLoading,
      showConfirmDialog,
      checkList,
      showIpSelector,
      ipList,
      statusColorMap,
      handleRemoveIp,
      chooseServer,
      handleCancel,
      handleShowConfirmDialog,
      handleConfirm,
      handleAddNode,
      handleTemplateChange,
    };
  },
});
</script>
<style lang="postcss" scoped>
.choose-node-template {
  padding: 24px;
  >>> .choose-node {
    .form-group-content {
      padding-top: 0;
    }
  }
}
</style>
