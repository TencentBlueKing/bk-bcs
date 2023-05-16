<template>
  <div class="detail p30" v-bkloading="{ isLoading: loading }">
    <div class="detail-title">
      {{ $t('基础信息') }}
    </div>
    <div class="detail-content basic-info">
      <div class="basic-info-item">
        <label>{{ $t('名称') }}</label>
        <span>{{ data.name }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{$t('描述')}}</label>
        <span
          class="bcs-ellipsis"
          v-bk-tooltips="{
            disabled: !data.desc,
            content: data.desc
          }">{{ data.desc || '--' }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('更新时间') }}</label>
        <span>{{ data.updateTime }}</span>
      </div>
    </div>
    <template v-if="kubeletData.length">
      <div class="detail-title mt20">
        {{ $t('Kubelet组件参数') }}
      </div>
      <div class="detail-content basic-info">
        <div
          class="basic-info-item" v-for="(item, index) in kubeletData"
          :key="index">
          <label>{{ item.key }}</label>
          <span>{{ item.value }}</span>
        </div>
      </div>
    </template>
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="label" :label="$t('标签')">
        <bk-table :data="handleTransformObjToArr(data.labels)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="taints" :label="$t('污点')">
        <bk-table :data="data.taints">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
          <bk-table-column label="Effect" prop="effect"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="annotations" :label="$t('注解')">
        <bk-table :data="handleTransformObjToArr(data.annotations)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
    </bcs-tab>

    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="label" :label="$t('前置初始化')">
        <pre class="bash-script" v-if="data.preStartUserScript">{{data.preStartUserScript}}</pre>
        <bcs-exception type="empty" scene="part" v-else></bcs-exception>
      </bcs-tab-panel>
      <bcs-tab-panel name="taints" :label="$t('后置初始化')">
        <pre class="bash-script" v-if="data.userScript">{{data.userScript}}</pre>
        <template v-else-if="paramsList.length && currentSops">
          <div class="mb15">{{currentSops.templateName}}</div>
          <bk-table :data="paramsList">
            <bk-table-column :label="$t('参数名')" prop="key"></bk-table-column>
            <bk-table-column :label="$t('值')" prop="value"></bk-table-column>
          </bk-table>
        </template>
        <bcs-exception type="empty" scene="part" v-else></bcs-exception>
      </bcs-tab-panel>
    </bcs-tab>
    <div class="mt15" v-if="operate">
      <bcs-button theme="primary" @click="handleEditTemplate">{{$t('编辑')}}</bcs-button>
      <bcs-button @click="handleDeleteTemplate">{{$t('删除')}}</bcs-button>
      <!-- <bcs-button @click="handleCancel">{{$t('取消')}}</bcs-button> -->
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';
import $store from '@/store/index';
import $router from '@/router';
import $i18n from '@/i18n/i18n-setup';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export default defineComponent({
  props: {
    // 当前行数据
    data: {
      type: Object,
      default: () => ({}),
    },
    operate: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const loading = ref(false);
    const curProject = computed(() => $store.state.curProject);
    const user = computed(() => $store.state.user);
    const bkSopsList = ref<any[]>([]);
    const handleGetbkSopsList = async () => {
      loading.value = true;
      bkSopsList.value = await $store.dispatch('clustermanager/bkSopsList', {
        $businessID: curProject.value.cc_app_id,
        operator: user.value.username,
        templateSource: 'business',
        scope: 'cmdb_biz',
      });
      loading.value = false;
    };

    const kubeletData = computed(() => {
      const data = props.data?.extraArgs?.kubelet?.split(';') || [];
      return data.map((item) => {
        const [key, value] = item.split('=');
        return {
          key,
          value,
        };
      }).filter(item => !!item.key);
    });
    const handleTransformObjToArr = (obj) => {
      if (!obj) return [];

      return Object.keys(obj).reduce<any[]>((data, key) => {
        data.push({
          key,
          value: obj[key],
        });
        return data;
      }, []);
    };
    // eslint-disable-next-line max-len
    const params = computed(() => props.data.scaleOutExtraAddons?.plugins?.[props.data.scaleOutExtraAddons?.postActions?.[0]]?.params || {});
    const currentSops = computed(() => bkSopsList.value.find(item => item.templateID === params.value.template_id));
    const paramsList = computed(() => Object.keys(params.value)
      .map(key => ({
        key,
        value: params.value[key],
      })));

    function handleEditTemplate() {
      $router.push({
        name: 'editNodeTemplate',
        params: {
          nodeTemplateID: props.data.nodeTemplateID,
        },
      });
    }
    function handleDeleteTemplate() {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        subTitle: props.data.name,
        title: $i18n.t('确认删除配置模版？'),
        defaultInfo: true,
        confirmFn: async () => {
          loading.value = true;
          const result = await $store.dispatch('clustermanager/deleteNodeTemplate', {
            $nodeTemplateId: props.data.nodeTemplateID,
          });
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('删除成功'),
            });
            ctx.emit('delete');
          }
          loading.value = false;
        },
      });
    }

    function handleCancel() {
      ctx.emit('cancel');
    }

    onMounted(() => {
      params.value.template_id && handleGetbkSopsList();
    });

    return {
      loading,
      handleTransformObjToArr,
      kubeletData,
      paramsList,
      currentSops,
      handleEditTemplate,
      handleDeleteTemplate,
      handleCancel,
    };
  },
});
</script>
<style lang="postcss" scoped>
.detail {
    font-size: 14px;
    /deep/ .bk-tab-label-item {
        background-color: #FAFBFD;
        border-bottom: 1px solid #dcdee5;
        line-height: 41px !important;
        height: 41px;
        &.active {
            border-bottom: none;
        }
    }
    /deep/ .bk-tab-label-wrapper {
        overflow: unset !important;
    }
    .bash-script {
        padding: 8px 16px;
        background: #F4F4F7;
        border-radius: 2px;
        font-size: 12px;
    }
    &-title {
        margin-bottom: 10px;
        color: #313238;
    }
    &-content {
        .basic-info-item {
            display: flex;
            padding: 0 15px;
            label {
                flex-shrink: 0;
                min-width: 88px;
            }
            .content {
                padding: 0 15px;
                flex: 1;
            }
        }
        &.worker-config {
            .basic-info-item {
                align-items: flex-start;
                .script {
                    background: #F4F4F7;
                    border-radius: 2px;
                    max-height: 200px;
                    overflow: auto;
                    padding: 8px 16px;
                }
            }
        }
        &.basic-info {
            border: 1px solid #dfe0e5;
            border-radius: 2px;
            .basic-info-item {
                align-items: center;
                height: 32px;
                &:nth-of-type(even) {
                    background: #F7F8FA;
                }
                label {
                    border-right: 1px solid #dfe0e5;
                    line-height: 32px;
                    width: 200px;
                }
                span {
                    padding: 0 15px;
                    flex: 1;
                }
            }
        }
    }
}

</style>
