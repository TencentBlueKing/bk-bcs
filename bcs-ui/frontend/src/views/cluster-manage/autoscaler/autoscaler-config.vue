<!-- eslint-disable max-len -->
<template>
  <BcsContent>
    <template #header>
      <HeaderNav :list="navList">
        <div>
          <bcs-button theme="primary" :loading="loading" @click="handleSaveConfig">{{$t('cluster.ca.button.save')}}</bcs-button>
          <bcs-button class="ml10" @click="handleCancel">{{$t('generic.button.cancel')}}</bcs-button>
        </div>
      </HeaderNav>
    </template>
    <bk-form
      :label-width="240"
      class="config-content"
      v-bkloading="{ isLoading: configLoading }"
      :rules="rules"
      :model="autoscalerData"
      ref="formRef">
      <div class="group-header mb20">
        <span class="group-header-title">{{$t('cluster.ca.title.caConfig')}}</span>
        <!-- <span class="switch-autoscaler">
          {{$t('cluster.ca.name')}}
          <bcs-switcher
            size="small"
            v-model="autoscalerData.enableAutoscale"
            :pre-check="handleToggleAutoScaler"
          ></bcs-switcher>
        </span> -->
      </div>
      <LayoutGroup :title="$t('cluster.ca.basic.title')">
        <bk-form-item
          :label="$t('cluster.ca.basic.scanInterval.label')"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('cluster.ca.basic.scanInterval.desc')">
          <bk-input int type="number" :min="5" :max="86400" v-model="autoscalerData.scanInterval">
            <template slot="append">
              <div class="group-text">{{$t('units.suffix.seconds')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
      </LayoutGroup>
      <!-- form-item放在Group内会导致form组件无法注册 -->
      <!-- <bk-form-item
        :label="$t('cluster.ca.basic.module.label')"
        :desc="$t('cluster.ca.basic.module.desc')"
        error-display-type="normal"
        property="scaleOutModuleID"
        required
        class="ml-[28px] mb20 mt10">
        <TopoSelectTree
          v-model="autoscalerData.module.scaleOutModuleID"
          :placeholder="$t('cluster.ca.basic.module.placeholder')"
          :cluster-id="clusterId"
          class="w-[500px]"
          @node-data-change="handleScaleOutDataChange" />
      </bk-form-item> -->
      <LayoutGroup :title="$t('cluster.ca.unreadyConfig.title')" class="mb10">
        <div class="flex">
          <i18n
            path="cluster.ca.unreadyConfig.path"
            class="text-[14px] text-[#63656e] flex items-center">
            <span place="0" class="px-[5px]">
              <bk-input int type="number" :min="0" :max="320000" v-model="autoscalerData.okTotalUnreadyCount">
                <template slot="append">
                  <div class="group-text">{{$t('units.suffix.units')}}</div>
                </template>
              </bk-input>
            </span>
            <span place="1" class="px-[5px] flex">
              <bk-input class="flex-1" int type="number" :min="0" :max="100" v-model="autoscalerData.maxTotalUnreadyPercentage">
              </bk-input>
              <div class="w-[42px] flex items-center justify-center bcs-border ml-[-1px] border-[#c4c6cc]">%</div>
            </span>
          </i18n>
          <span
            class="ml-[5px] text-[16px] text-[#979ba5] leading-[30px]"
            v-bk-tooltips="$t('cluster.ca.unreadyConfig.desc')">
            <i class="bk-icon icon-info-circle"></i>
          </span>
        </div>
      </LayoutGroup>
      <LayoutGroup :title="$t('cluster.ca.autoScalerConfig.title')" class="mb10">
        <bk-form-item
          :label="$t('cluster.ca.autoScalerConfig.expander.title')" :model="autoscalerData"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('cluster.ca.autoScalerConfig.expander.desc')">
          <bk-radio-group v-model="autoscalerData.expander">
            <bk-radio value="random">Random</bk-radio>
            <bk-radio value="least-waste">Least Waste</bk-radio>
            <bk-radio value="most-pods">Most Pods</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item
          desc-icon="bk-icon icon-info-circle"
          :label="$t('cluster.ca.autoScalerConfig.bufferResourceCpuRatio.title')"
          :desc="$t('cluster.ca.autoScalerConfig.bufferResourceCpuRatio.desc')">
          <bk-input int type="number" :min="0" :max="100" v-model="autoscalerData.bufferResourceCpuRatio">
            <template slot="append">
              <div class="group-text">%</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          desc-icon="bk-icon icon-info-circle"
          :label="$t('cluster.ca.autoScalerConfig.bufferResourceMemRatio.title')"
          :desc="$t('cluster.ca.autoScalerConfig.bufferResourceMemRatio.desc')">
          <bk-input int type="number" :min="0" :max="100" v-model="autoscalerData.bufferResourceMemRatio">
            <template slot="append">
              <div class="group-text">%</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          desc-icon="bk-icon icon-info-circle"
          :label="$t('cluster.ca.autoScalerConfig.bufferResourceRatio.title')"
          :desc="$t('cluster.ca.autoScalerConfig.bufferResourceRatio.desc')">
          <bk-input int type="number" :min="0" :max="100" v-model="autoscalerData.bufferResourceRatio">
            <template slot="append">
              <div class="group-text">%</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.autoScalerConfig.maxNodeProvisionTime.title')"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('cluster.ca.autoScalerConfig.maxNodeProvisionTime.desc1')">
          <bk-input int type="number" :min="900" :max="86400" v-model="autoscalerData.maxNodeProvisionTime">
            <template slot="append">
              <div class="group-text">{{$t('units.suffix.seconds')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
      </LayoutGroup>
      <LayoutGroup collapsible class="mb10" :expanded="!!autoscalerData.isScaleDownEnable">
        <template #title>
          <span>{{$t('cluster.ca.autoScalerDownConfig.title')}}</span>
          <span class="switch-autoscaler">
            {{$t('cluster.ca.autoScalerDownConfig.label')}}
            <bcs-switcher
              size="small"
              v-model="autoscalerData.isScaleDownEnable"
              @click.native.stop>
            </bcs-switcher>
          </span>
        </template>
        <bk-form-item
          desc-icon="bk-icon icon-info-circle"
          :label="$t('cluster.ca.autoScalerDownConfig.scaleDownUtilizationThreahold.title')"
          :desc="$t('cluster.ca.autoScalerDownConfig.scaleDownUtilizationThreahold.desc')">
          <bk-input int type="number" :min="0" :max="80" v-model="autoscalerData.scaleDownUtilizationThreahold">
            <template slot="append">
              <div class="group-text">%</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.autoScalerDownConfig.scaleDownUnneededTime.title')"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('cluster.ca.autoScalerDownConfig.scaleDownUnneededTime.desc')">
          <div class="flex">
            <bk-input int type="number" :min="60" :max="86400" v-model="autoscalerData.scaleDownUnneededTime">
            </bk-input>
            <div class="w-[42px] flex items-center justify-center bcs-border ml-[-1px] border-[#c4c6cc]">{{$t('units.suffix.seconds')}}</div>
            <span class="ml8">{{$t('cluster.ca.autoScalerDownConfig.scaleDownUnneededTime.suffix')}}</span>
          </div>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.autoScalerDownConfig.maxGracefulTerminationSec.title')"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('cluster.ca.autoScalerDownConfig.maxGracefulTerminationSec.desc1')">
          <bk-input int type="number" :min="60" :max="86400" v-model="autoscalerData.maxGracefulTerminationSec">
            <template slot="append">
              <div class="group-text">{{$t('units.suffix.seconds')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.autoScalerDownConfig.scaleDownDelayAfterAdd.title')"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('cluster.ca.autoScalerDownConfig.scaleDownDelayAfterAdd.desc1')">
          <bk-input int type="number" :min="1200" :max="86400" v-model="autoscalerData.scaleDownDelayAfterAdd">
            <template slot="append">
              <div class="group-text">{{$t('units.suffix.seconds')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.autoScalerDownConfig.scaleDownDelayAfterDelete.title')"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('cluster.ca.autoScalerDownConfig.scaleDownDelayAfterDelete.desc1')">
          <bk-input int type="number" :min="0" :max="86400" v-model="autoscalerData.scaleDownDelayAfterDelete">
            <template slot="append">
              <div class="group-text">{{$t('units.suffix.seconds')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.autoScalerDownConfig.scaleDownUnreadyTime.title')" :desc="$t('cluster.ca.autoScalerDownConfig.scaleDownUnreadyTime.range')">
          <bk-input int type="number" :min="1200" :max="86400" v-model="autoscalerData.scaleDownUnreadyTime">
            <template slot="append">
              <div class="group-text">{{$t('units.suffix.seconds')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.ca.autoScalerDownConfig.maxEmptyBulkDelete.title')">
          <bk-input int type="number" :min="1" :max="100" v-model="autoscalerData.maxEmptyBulkDelete">
            <template slot="append">
              <div class="group-text">{{$t('units.suffix.units')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.autoScalerDownConfig.skipNodesWithLocalStorage.title')"
          :desc="$t('cluster.ca.autoScalerDownConfig.skipNodesWithLocalStorage.desc')">
          <bcs-checkbox
            v-model="autoscalerData.skipNodesWithLocalStorage"
            :true-value="false"
            :false-value="true">
            {{ autoscalerData.skipNodesWithLocalStorage ? $t('units.boolean.false') : $t('units.boolean.true') }}
          </bcs-checkbox>
        </bk-form-item>
      </LayoutGroup>
      <LayoutGroup collapsible :title="$t('cluster.ca.podsPriorityConfig.title')">
        <i18n path="cluster.ca.podsPriorityConfig.path" class="text-[14px]">
          <span>{{ $t('units.op.le') }}</span>
          <bcs-input
            class="w-[100px] px-[5px]"
            type="number"
            :min="-2147483648"
            :max="-1"
            v-model="autoscalerData.expendablePodsPriorityCutoff">
          </bcs-input>
        </i18n>
      </LayoutGroup>
    </bk-form>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';

import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import { useAppData } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';
import HeaderNav from '@/views/cluster-manage/components/header-nav.vue';
import LayoutGroup from '@/views/cluster-manage/components/layout-group.vue';

export default defineComponent({
  components: {
    BcsContent,
    HeaderNav,
    LayoutGroup,
  },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const formRef = ref<any>(null);
    const rules = ref({
      // 转移模块放在基本配置里面了
      // scaleOutModuleID: [
      //   {
      //     validator: () => !!autoscalerData.value.module?.scaleOutModuleID,
      //     message: $i18n.t('generic.validate.required'),
      //     trigger: 'blur',
      //   },
      // ],
    });
    const { clusterList } = useClusterList();
    const navList = computed(() => [
      {
        title: clusterList.value.find(item => item.clusterID === props.clusterId)?.clusterName,
        link: {
          name: 'clusterMain',
        },
      },
      {
        title: 'Cluster Autoscaler',
        link: {
          name: 'clusterMain',
          query: {
            clusterId: props.clusterId,
            active: 'autoscaler',
          },
        },
      },
      {
        title: $i18n.t('cluster.ca.button.edit'),
        link: null,
      },
    ]);
    const configLoading = ref(false);
    const autoscalerData = ref<Record<string, any>>({
      module: {},
    });
    const handleGetAutoScalerConfig = async () => {
      configLoading.value = true;
      autoscalerData.value = await $store.dispatch('clustermanager/clusterAutoScaling', {
        $clusterId: props.clusterId,
      });
      autoscalerData.value.module = autoscalerData.value.module || {};
      configLoading.value = false;
    };

    const loading = ref(false);
    const user = computed(() => $store.state.user);
    const { _INTERNAL_ } = useAppData();
    const handleSaveConfig = async () => {
      const validate = await formRef.value?.validate();
      if (!validate) return;
      loading.value = true;
      const result = await $store.dispatch('clustermanager/updateClusterAutoScaling', {
        ...autoscalerData.value,
        provider: _INTERNAL_.value ? 'selfProvisionCloud' : '',
        updater: user.value.username,
        $clusterId: props.clusterId,
      });
      loading.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.update'),
        });
        $router.push({
          name: 'clusterMain',
          query: {
            active: 'autoscaler',
            clusterId: props.clusterId,
          },
        });
      }
    };
    const handleCancel = () => {
      $router.back();
    };
    // eslint-disable-next-line @typescript-eslint/no-misused-promises
    const handleToggleAutoScaler = async value => new Promise(async (resolve, reject) => {
      const nodepoolList = await $store.dispatch('clustermanager/nodeGroup', {
        clusterID: props.clusterId,
      });
      if (!autoscalerData.value.enableAutoscale
                        && (!nodepoolList.length || nodepoolList.every(item => !item.enableAutoscale))) {
        // 开启时前置判断是否存在节点规格 或 节点规格都是未开启状态时，要提示至少开启一个
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          subTitle: !nodepoolList.length
            ? $i18n.t('cluster.ca.tips.emptyNodePool')
            : $i18n.t('cluster.ca.tips.notEnableAnyNodePool'),
          defaultInfo: true,
          okText: $i18n.t('cluster.ca.button.create'),
          confirmFn: () => {
            $router.push({
              name: 'nodePool',
              params: {
                clusterId: props.clusterId,
              },
            }).catch((err) => {
              console.warn(err);
            });
          },
          cancelFn: () => {
            // eslint-disable-next-line prefer-promise-reject-errors
            reject(false);
          },
        });
      } else {
        // 开启或关闭扩缩容
        const result = await $store.dispatch('clustermanager/toggleClusterAutoScalingStatus', {
          enable: value,
          provider: _INTERNAL_.value ? 'selfProvisionCloud' : '',
          $clusterId: props.clusterId,
          updater: user.value.username,
        });
        if (result) {
          $bkMessage({
            theme: 'success',
            message: $i18n.t('generic.msg.success.ok'),
          });
          handleGetAutoScalerConfig();
          resolve(true);
        } else {
          // eslint-disable-next-line prefer-promise-reject-errors
          reject(false);
        }
      }
    });

    const handleScaleOutDataChange = (data) => {
      autoscalerData.value.module.scaleOutModuleName = data?.path || '';
    };

    onMounted(() => {
      handleGetAutoScalerConfig();
    });
    return {
      rules,
      formRef,
      navList,
      loading,
      configLoading,
      autoscalerData,
      handleCancel,
      handleSaveConfig,
      handleToggleAutoScaler,
      handleScaleOutDataChange,
    };
  },
});
</script>
<style lang="postcss" scoped>
.ml8 {
    margin-left: 8px;
}
.config-content {
    background: #fff;
    border: 1px solid #DDE4EB;
    border-radius: 2px;
    padding: 32px 52px;
    user-select: none;
    >>> .group-header {
        display: flex;
        align-items: center;
        font-size: 12px;
        &-title {
            font-size: 14px;
            font-weight: bold;
            color: #63656E;
        }
    }
}
>>> .bk-form-content {
    font-size: 14px;
    .bk-form-control {
        width: auto;
    }
}
>>> .bk-input-number {
    max-width: 88px;
}
.switch-autoscaler {
    margin-left: 16px;
    padding-left: 16px;
    border-left: 1px solid #DCDEE5;
    color: #63656E;
}
</style>
