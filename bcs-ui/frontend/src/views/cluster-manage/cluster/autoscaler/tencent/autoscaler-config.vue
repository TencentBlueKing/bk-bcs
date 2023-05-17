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
            @change="handleToggleAutoScaler"
          ></bcs-switcher>
        </span>
      </div>
      <LayoutGroup :title="$t('基本配置')" class="mb10">
        <bk-form-item :label="$t('扩缩容检测时间间隔')">
          <bk-input type="number" v-model="autoscalerData.scanInterval">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('允许unready节点')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('自动扩缩容保护机制，集群中unready节点大于允许unready节点数量，且unready节点的比例大于设置的比例，会停止Cluster Autoscaler功能，否则Cluster Autoscaler功能正常运行')">
          <bk-input type="number" v-model="autoscalerData.okTotalUnreadyCount">
            <template slot="append">
              <div class="group-text">{{$t('个')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item :label="$t('unready节点超过集群总节点')">
          <div style="display: flex;">
            <bcs-input type="number" v-model="autoscalerData.maxTotalUnreadyPercentage">
              <template slot="append">
                <div class="group-text">{{$t('%')}}</div>
              </template>
            </bcs-input>
            <span class="ml8">{{$t('时停止自动扩缩容')}}</span>
          </div>
        </bk-form-item>
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
        <bk-form-item :label="$t('触发扩容资源阈值')">
          <bk-input type="number" v-model="autoscalerData.bufferResourceRatio">
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
          <bk-input type="number" v-model="autoscalerData.maxNodeProvisionTime">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item :label="$t('(没有ready节点时) 允许自动扩容')">
          <bk-checkbox v-model="autoscalerData.scaleUpFromZero"></bk-checkbox>
        </bk-form-item>
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
        <bk-form-item :label="$t('触发缩容资源阈值 (CPU/内存)')">
          <bk-input type="number" v-model="autoscalerData.scaleDownUtilizationThreahold">
            <template slot="append">
              <div class="group-text">%</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('执行缩容等待时间')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('Cluster Autocaler组件评估集群可以缩容多久后开始执行缩容，防止集群容量在短时间内或高或低于设置的缩容阈值造成频繁扩缩容操作')">
          <bk-input type="number" v-model="autoscalerData.scaleDownUnneededTime">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('等待 Pod 退出最长时间')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('缩容节点时，等待 pod 停止的最长时间（不会遵守 terminationGracefulPeriodSecond，超时强杀）')">
          <bk-input type="number" v-model="autoscalerData.maxGracefulTerminationSec">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('扩容后判断缩容时间间隔')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('扩容节点后多久才继续缩容判断，如果业务自定义初始化任务所需时间比较长，需要适当上调此值')">
          <bk-input type="number" v-model="autoscalerData.scaleDownDelayAfterAdd">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('连续两次缩容时间间隔')"
          desc-type="icon"
          desc-icon="bk-icon icon-info-circle"
          :desc="$t('缩容节点后多久再继续缩容节点，默认设置为0，代表与扩缩容检测时间间隔设置的值相同')">
          <bk-input type="number" v-model="autoscalerData.scaleDownDelayAfterDelete">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item :label="$t('缩容失败后重试时间间隔')">
          <bk-input type="number" v-model="autoscalerData.scaleDownDelayAfterFailure">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>
        <bk-form-item :label="$t('unready节点缩容等待时间')">
          <bk-input type="number" v-model="autoscalerData.scaleDownUnreadyTime">
            <template slot="append">
              <div class="group-text">{{$t('秒')}}</div>
            </template>
          </bk-input>
        </bk-form-item>

        <bk-form-item :label="$t('单次缩容最大节点数')">
          <bk-input type="number" v-model="autoscalerData.maxEmptyBulkDelete">
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
import { computed, defineComponent, onMounted, ref } from 'vue';
import BcsContent from '@/views/cluster-manage/components/bcs-content.vue';
import HeaderNav from '@/views/cluster-manage/components/header-nav.vue';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import LayoutGroup from '@/views/cluster-manage/components/layout-group.vue';
import $bkMessage from '@/common/bkmagic';

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
    const { clusterList } = useClusterList();
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
            active: 'autoscaler',
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
            active: 'autoscaler',
          },
        });
      }
    };
    const handleCancel = () => {
      $router.back();
    };
    const handleToggleAutoScaler = async (value) => {
      const result = await $store.dispatch('clustermanager/toggleClusterAutoScalingStatus', {
        enable: value,
        $clusterId: props.clusterId,
        updater: user.value.username,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('操作成功'),
        });
        handleGetAutoScalerConfig();
      }
    };

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
