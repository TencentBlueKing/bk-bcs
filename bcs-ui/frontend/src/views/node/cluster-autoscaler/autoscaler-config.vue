<!-- eslint-disable max-len -->
<template>
  <BcsContent>
    <template #header>
      <HeaderNav :list="navList">
        <div>
          <bcs-button theme="primary" :loading="loading" @click="handleSaveConfig">{{$t('保存配置')}}</bcs-button>
          <bcs-button class="ml10" @click="handleCancel">{{$t('取消')}}</bcs-button>
        </div>
      </HeaderNav>
    </template>
    <bk-form :label-width="240" class="config-content" v-bkloading="{ isLoading: configLoading }">
      <div class="group-header mb20">
        <span class="group-header-title">{{$t('Cluster Autoscaler配置')}}</span>
        <span class="switch-autoscaler">
          {{$t('Cluster Autoscaler')}}
          <bcs-switcher
            size="small"
            v-model="autoscalerData.enableAutoscale"
            :pre-check="handleToggleAutoScaler"
          ></bcs-switcher>
        </span>
      </div>
      <LayoutGroup :title="$t('基本配置')" class="mb10">
        <bk-form-item
          :label="$t('扩缩容检测时间间隔')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('默认为10秒，取值范围5 ~ 86400秒')">
          <bk-input type="number" :min="5" :max="86400" v-model="autoscalerData.scanInterval">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
      </LayoutGroup>
      <LayoutGroup :title="$t('扩缩容暂停配置')" class="mb10">
        <div class="flex">
          <i18n
            path="NotReady节点数大于 {0} 个且超过集群总节点数 {1} 时暂停自动扩缩容"
            class="text-[14px] text-[#63656e] flex items-center">
            <span place="0" class="px-[5px]">
              <bk-input type="number" :min="0" :max="320000" v-model="autoscalerData.okTotalUnreadyCount">
                <template slot="append">
                  <div class="group-text">{{$t('个')}}</div>
                </template>
              </bk-input>
            </span>
            <span place="1" class="px-[5px]">
              <bcs-input type="number" :min="0" :max="100" v-model="autoscalerData.maxTotalUnreadyPercentage">
                <template slot="append">
                  <div class="group-text">{{$t('%')}}</div>
                </template>
              </bcs-input>
            </span>
          </i18n>
          <span
            class="ml-[5px] text-[16px] text-[#979ba5] leading-[30px]"
            v-bk-tooltips="$t('自动扩缩容保护机制，如果NotReady节点数量或比例过大，自动扩容上来的节点也有可能会是NotReady状态，导致业务成本增加，当NotReady节点数不符合暂停触发条件时自动恢复自动扩缩容')">
            <i class="bk-icon icon-info-circle"></i>
          </span>
        </div>
      </LayoutGroup>
      <LayoutGroup :title="$t('自动扩容配置')" class="mb10">
        <bk-form-item
          :label="$t('扩容算法')" :model="autoscalerData"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('random：在有多个节点池时，随机选择节点池<br/>least-waste：在有多个节点池时，以最小浪费原则选择，选择有最少可用资源的节点池<br/>most-pods：在有多个节点池时，选择容量最大（可以创建最多Pod）的节点池')">
          <bk-radio-group v-model="autoscalerData.expander">
            <bk-radio value="random">Random</bk-radio>
            <bk-radio value="least-waste">Least Waste</bk-radio>
            <bk-radio value="most-pods">Most Pods</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :label="$t('触发扩容资源阈值 (CPU)')"
          :desc="$t('CPU资源(Request)使用率超过该阈值触发扩容, 无论内存资源使用率是否达到阈值')">
          <bk-input type="number" :min="0" :max="100" v-model="autoscalerData.bufferResourceCpuRatio">
            <template slot="append">
              <div class="group-text">%</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :label="$t('触发扩容资源阈值 (内存)')"
          :desc="$t('内存资源(Request)使用率超过该阈值触发扩容, 无论CPU资源使用率是否达到阈值')">
          <bk-input type="number" :min="0" :max="100" v-model="autoscalerData.bufferResourceMemRatio">
            <template slot="append">
              <div class="group-text">%</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('等待节点提供最长时间')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('如果节点池在设置的时间范围内没有提供可用资源，会导致此次自动扩容失败')">
          <bk-input type="number" :min="60" :max="86400" v-model="autoscalerData.maxNodeProvisionTime">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <!-- <bk-form-item :label="$t('(没有ready节点时) 允许自动扩容')">
          <bk-checkbox v-model="autoscalerData.scaleUpFromZero"></bk-checkbox>
        </bk-form-item> -->
      </LayoutGroup>
      <LayoutGroup collapsible :expanded="!!autoscalerData.isScaleDownEnable">
        <template #title>
          <span>{{$t('自动缩容配置')}}</span>
          <span class="switch-autoscaler">
            {{$t('允许缩容节点')}}
            <bcs-switcher
              size="small"
              v-model="autoscalerData.isScaleDownEnable"
              @click.native.stop>
            </bcs-switcher>
          </span>
        </template>
        <bk-form-item
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :label="$t('触发缩容资源阈值 (CPU/内存)')"
          :desc="$t('CPU和内存资源(Request)必须同时低于设定阈值才会触发缩容')">
          <bk-input type="number" :min="0" :max="100" v-model="autoscalerData.scaleDownUtilizationThreahold">
            <template slot="append">
              <div class="group-text">%</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('节点连续空闲')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('节点从第一次被标记空闲状态到设定时间内一直处于空闲状态才会被缩容，防止集群资源使用率短时间内波动造成频繁扩缩容操作')">
          <div class="flex">
            <bcs-input type="number" :min="60" :max="86400" v-model="autoscalerData.scaleDownUnneededTime">
              <template slot="append">
                <div class="group-text">{{$t('秒')}}</div>
              </template>
            </bcs-input>
            <span class="ml8">{{$t('后执行缩容')}}</span>
          </div>
        </bk-form-item>
        <bk-form-item
          :label="$t('等待 Pod 退出最长时间')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('缩容节点时，等待 pod 停止的最长时间（不会遵守 terminationGracefulPeriodSecond，超时强杀）默认为600秒，取值范围60 ~ 86400秒')">
          <bk-input type="number" :min="60" :max="86400" v-model="autoscalerData.maxGracefulTerminationSec">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('扩容后判断缩容时间间隔')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('扩容节点后多久才继续缩容判断，如果业务自定义初始化任务所需时间比较长，需要适当上调此值，取值范围60 ~ 86400秒')">
          <bk-input type="number" :min="60" :max="86400" v-model="autoscalerData.scaleDownDelayAfterAdd">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('连续两次缩容时间间隔')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('缩容节点后多久再继续缩容节点，默认设置为0，代表与扩缩容检测时间间隔设置的值相同，取值范围0 ~ 86400秒')">
          <bk-input type="number" :min="0" :max="86400" v-model="autoscalerData.scaleDownDelayAfterDelete">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <!-- <bk-form-item :label="$t('缩容失败后重试时间间隔')">
          <bk-input type="number" :min="60" :max="86400" v-model="autoscalerData.scaleDownDelayAfterFailure">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item> -->
        <bk-form-item :label="$t('NotReady节点缩容等待时间')">
          <bk-input type="number" :min="60" :max="86400" v-model="autoscalerData.scaleDownUnreadyTime">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item :label="$t('单次缩容最大节点数')">
          <bk-input type="number" :min="1" :max="320000" v-model="autoscalerData.maxEmptyBulkDelete">
            <template slot="append">
              <div class="group-text">{{$t('个')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
      </LayoutGroup>
    </bk-form>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref } from '@vue/composition-api';
import BcsContent from '../bcs-content.vue';
import HeaderNav from '../header-nav.vue';
import { useClusterList } from '@/views/cluster/use-cluster';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router/index';
import $store from '@/store/index';
import LayoutGroup from '../LayoutGroup.vue';

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
  setup(props, ctx) {
    const { $bkMessage, $bkInfo } = ctx.root;
    const { clusterList } = useClusterList(ctx);
    const navList = computed(() => [
      {
        title: clusterList.value.find(item => item.clusterID === props.clusterId)?.clusterName,
        link: {
          name: 'clusterDetail',
        },
      },
      {
        title: 'Cluster Autoscaler',
        link: {
          name: 'clusterDetail',
          query: {
            active: 'AutoScaler',
          },
        },
      },
      {
        title: $i18n.t('编辑配置'),
        link: null,
      },
    ]);
    const configLoading = ref(false);
    const autoscalerData = ref<Record<string, string>>({});
    const handleGetAutoScalerConfig = async () => {
      configLoading.value = true;
      autoscalerData.value = await $store.dispatch('clustermanager/clusterAutoScaling', {
        $clusterId: props.clusterId,
      });
      configLoading.value = false;
    };

    const loading = ref(false);
    const user = computed(() => $store.state.user);
    const handleSaveConfig = async () => {
      loading.value = true;
      const result = await $store.dispatch('clustermanager/updateClusterAutoScaling', {
        ...autoscalerData.value,
        provider: 'selfProvisionCloud',
        updater: user.value.username,
        $clusterId: props.clusterId,
      });
      loading.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('更新成功'),
        });
        $router.push({
          name: 'clusterDetail',
          query: {
            active: 'AutoScaler',
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
        // 开启时前置判断是否存在节点池 或 节点池都是未开启状态时，要提示至少开启一个
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          subTitle: !nodepoolList.length
            ? $i18n.t('没有检测到可用节点池，请先创建节点池')
            : $i18n.t('请至少启用 1 个节点池的自动扩缩容功能或创建新的节点池'),
          defaultInfo: true,
          okText: $i18n.t('立即新建'),
          confirmFn: () => {
            $router.push({
              name: 'nodePool',
              params: {
                clusterId: props.clusterId,
              },
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
          provider: 'selfProvisionCloud',
          $clusterId: props.clusterId,
          updater: user.value.username,
        });
        if (result) {
          $bkMessage({
            theme: 'success',
            message: $i18n.t('操作成功'),
          });
          handleGetAutoScalerConfig();
          resolve(true);
        } else {
          // eslint-disable-next-line prefer-promise-reject-errors
          reject(false);
        }
      }
    });

    onMounted(() => {
      handleGetAutoScalerConfig();
    });
    return {
      navList,
      loading,
      configLoading,
      autoscalerData,
      handleCancel,
      handleSaveConfig,
      handleToggleAutoScaler,
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
    width: 88px;
}
.switch-autoscaler {
    margin-left: 16px;
    padding-left: 16px;
    border-left: 1px solid #DCDEE5;
    color: #63656E;
}
</style>
