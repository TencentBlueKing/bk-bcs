<template>
  <div class="choose-node-template">
    <bk-form>
      <FormGroup :allow-toggle="false" class="choose-node">
        <bk-form-item :label="$t('选择节点模板')">
          <div
            class="item-node-template" v-bk-tooltips="{
              disabled: isTkeCluster,
              content: $t('非TKE集群不支持节点模板')
            }">
            <bcs-select
              searchable
              :clearable="false"
              placeholder=" "
              :disabled="!isTkeCluster"
              :loading="loading"
              v-model="nodeTemplateID"
              @change="handleNodeTemplateIDChange">
              <bcs-option id="" :name="$t('不使用节点模板')"></bcs-option>
              <bcs-option
                v-for="item in templateList"
                :key="item.nodeTemplateID"
                :id="item.nodeTemplateID"
                :name="item.name">
              </bcs-option>
              <template #extension>
                <span style="cursor: pointer" @click="handleGotoNodeTemplate">
                  <i
                    class="bcs-icon bcs-icon-fenxiang mr5"
                    style="font-size: 12px"></i>
                  {{$t('节点模板配置')}}
                </span>
              </template>
            </bcs-select>
            <template v-if="isTkeCluster">
              <span
                class="icon ml10"
                v-bk-tooltips.top="$t('刷新列表')"
                @click="handleNodeTemplateList">
                <i class="bcs-icon bcs-icon-reset"></i>
              </span>
              <span class="icon" v-if="nodeTemplateID">
                <i
                  class="bcs-icon bcs-icon-yulan ml15"
                  v-bk-tooltips.top="$t('预览')"
                  @click="handlePreview"></i>
              </span>
            </template>
          </div>
        </bk-form-item>
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
      </FormGroup>
    </bk-form>
    <div class="mt25">
      <span v-bk-tooltips="{ content: $t('请选择节点'), disabled: ipList.length }">
        <bk-button
          class="mw88"
          theme="primary"
          :disabled="!ipList.length"
          @click="handleConfirm">
          {{$t('确定')}}
        </bk-button>
      </span>
      <bk-button class="mw88 ml10" @click="handleCancel">{{$t('取消')}}</bk-button>
    </div>
    <!-- IP选择器 -->
    <IpSelector v-model="showIpSelector" :ip-list="ipList" @confirm="chooseServer"></IpSelector>
    <bcs-sideslider
      :is-show.sync="showDetail"
      :title="currentRow.name"
      quick-close
      width="800">
      <div slot="content">
        <NodeTemplateDetail :data="currentRow"></NodeTemplateDetail>
      </div>
    </bcs-sideslider>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref } from '@vue/composition-api';
import FormGroup from '@/views/cluster/create-cluster/form-group.vue';
import $store from '@/store/index';
import $router from '@/router';
import IpSelector from '@/components/ip-selector/selector-dialog.vue';
import StatusIcon from '@/views/dashboard/common/status-icon';
import $i18n from '@/i18n/i18n-setup';
import useNode from './use-node';
import NodeTemplateDetail from './node-template-detail.vue';
import { NODE_TEMPLATE_ID } from '@/common/constant';

export default defineComponent({
  components: { FormGroup, IpSelector, StatusIcon, NodeTemplateDetail },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props, ctx) {
    const { $bkInfo } = ctx.root;
    const isTkeCluster = computed(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === props.clusterId)?.provider === 'tencentCloud');
    const loading = ref(false);
    const statusColorMap = ref({
      0: 'red',
      1: 'green',
    });
    const showIpSelector = ref(false);
    const ipList = ref<any[]>([]);
    const nodeTemplateID = ref(localStorage.getItem(NODE_TEMPLATE_ID) || '');
    const templateList = ref<any[]>([]);
    const handleNodeTemplateList = async () => {
      loading.value = true;
      templateList.value = await $store.dispatch('clustermanager/nodeTemplateList');
      if (!isTkeCluster.value
                    || !templateList.value.find(item => item.nodeTemplateID === nodeTemplateID.value)
      ) {
        nodeTemplateID.value = '';
      }
      loading.value = false;
    };
    const handleGotoNodeTemplate = () => {
      const location = $router.resolve({ name: 'nodeTemplate' });
      window.open(location.href);
    };
    const showDetail = ref(false);
    const currentRow = computed(() => templateList.value
      .find(item => item.nodeTemplateID === nodeTemplateID.value) || {});
    const handlePreview = () => {
      showDetail.value = true;
    };
    const handleCancel = () => {
      $router.back();
    };

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
    const handleConfirm = () => {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        subTitle: $i18n.t('请确认是否对 {ip} 等 {num} 个IP进行操作系统初始化和安装容器服务相关组件操作', {
          ip: ipList.value[0].bk_host_innerip,
          num: ipList.value.length,
        }),
        title: $i18n.t('确认添加节点'),
        defaultInfo: true,
        confirmFn: async () => {
          const result = await addNode({
            clusterId: props.clusterId,
            nodeIps: ipList.value.map(item => item.bk_host_innerip),
            nodeTemplateID: nodeTemplateID.value,
          });

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
        },
      });
    };

    // todo 框架支持数据持久化
    const handleNodeTemplateIDChange = (value) => {
      localStorage.setItem(NODE_TEMPLATE_ID, value);
    };

    onMounted(() => {
      handleNodeTemplateList();
    });
    return {
      loading,
      isTkeCluster,
      showDetail,
      currentRow,
      showIpSelector,
      ipList,
      nodeTemplateID,
      templateList,
      statusColorMap,
      handleRemoveIp,
      chooseServer,
      handleCancel,
      handleConfirm,
      handleGotoNodeTemplate,
      handlePreview,
      handleAddNode,
      handleNodeTemplateIDChange,
      handleNodeTemplateList,
    };
  },
});
</script>
<style lang="postcss" scoped>
.choose-node-template {
    padding: 24px;
    >>> .item-node-template {
        display: flex;
        max-width: 524px;
        .bk-select {
          width: 400px;
        }
        .icon:hover {
          color: #3a84ff;
          cursor: pointer;
        }
    }
    >>> .choose-node {
        .form-group-content {
            padding-top: 0;
        }
    }
}
</style>
