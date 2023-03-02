<template>
  <section class="cluster bcs-content-wrapper">
    <div class="mode-wrapper mt15">
      <!-- 自建集群 -->
      <div class="mode-panel" @click="handleCreateCluster">
        <span class="mode-panel-icon"><i class="bcs-icon bcs-icon-sitemap"></i></span>
        <span class="mode-panel-title">{{ $t('自建集群') }}</span>
        <span class="mode-panel-desc">{{ $t('可自定义集群基本信息和集群版本') }}</span>
      </div>
      <!-- 导入集群 -->
      <div
        :class="['mode-panel', { disabled: $INTERNAL }]"
        v-bk-tooltips="{ disabled: !$INTERNAL, content: $t('功能建设中') }"
        @click="handleImportCluster">
        <span class="mode-panel-icon"><i class="bcs-icon bcs-icon-upload"></i></span>
        <span class="mode-panel-title">{{ $t('导入集群') }}</span>
        <span class="mode-panel-desc">{{ $t('支持快速导入已存在的集群') }}</span>
      </div>
    </div>
  </section>
</template>
<script lang="ts">
import { computed, defineComponent, ref } from '@vue/composition-api';
import fullScreen from '@/directives/full-screen';
import yamljs from 'js-yaml';
import { useConfig } from '@/common/use-app';

export default defineComponent({
  name: 'CreateCluster',
  directives: {
    'full-screen': fullScreen,
  },
  setup(props, ctx) {
    const { $router } = ctx.root;
    const { _INTERNAL_ } = useConfig();

    // 自建集群
    const handleCreateCluster = () => {
      $router.push({ name: 'createFormCluster' });
    };
    // 导入集群
    const handleImportCluster = () => {
      if (_INTERNAL_.value) return;
      $router.push({ name: 'createImportCluster' });
    };

    // 展示模板详情
    const showDetail = ref(false);
    const curCloud = ref<any>({});
    const detailTitle = computed(() => `${curCloud.value.name}${curCloud.value.description ? `( ${curCloud.value.description} )` : ''}`);
    const yaml = computed(() => yamljs.dump(curCloud.value));
    const handleShowDetail = (row) => {
      showDetail.value = true;
      curCloud.value = row;
    };
    return {
      showDetail,
      curCloud,
      detailTitle,
      yaml,
      handleCreateCluster,
      handleShowDetail,
      handleImportCluster,
    };
  },
});
</script>
<style lang="postcss" scoped>
/deep/ .bk-sideslider-content {
    height: 100%;
}
.cluster {
    padding: 20px 24px;
    .title {
        font-size: 14px;
        font-weight: 700;
        text-align: left;
        color: #63656e;
        line-height: 22px;
    }
    .mode-wrapper {
        display: flex;
        align-items: center;
    }
    .mode-panel {
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        margin-right: 24px;
        flex: 1;
        background: #fff;
        border-radius: 1px;
        box-shadow: 0px 2px 4px 0px rgba(25,25,41,0.05);
        height: 238px;
        cursor: pointer;
        &:hover {
            &:not(.disabled) {
                border: 1px solid #1768ef;
                .mode-panel-icon {
                    background: #e1ecff;
                }
                .mode-panel-title {
                    color: #3a84ff;
                }
            }
        }
        &:last-child {
            margin-right: 0;
        }
        &-icon {
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 40px;
            color: #979ba5;
            width: 80px;
            height: 80px;
            border-radius: 50%;
            background: #f5f7fa;
        }
        &-title {
            margin-top: 20px;
            font-size: 20px;
            font-weight: 400;
            color: #63656e;
            line-height: 28px;
        }
        &-desc {
            margin-top: 8px;
            font-size: 14px;
            font-weight: 400;
            text-align: center;
            color: #979ba5;
            line-height: 22px;
        }
        &.disabled {
            cursor: not-allowed;
        }
    }
    .cluster-template-title {
        display: flex;
        justify-content: space-between;
        margin-top: 40px;
    }
}
</style>
