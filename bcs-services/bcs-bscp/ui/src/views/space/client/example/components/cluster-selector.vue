<template>
  <bk-select
    v-model="currentValue"
    v-bkloading="loading"
    ref="selectorRef"
    :class="['p2p-selector', { 'select-error': isError }]"
    :popover-options="{ theme: 'light bk-select-popover' }"
    :filterable="true"
    :input-search="false"
    :clearable="false"
    :search-placeholder="$t('搜索视图')"
    :no-data-text="$t('暂无数据')"
    :no-match-text="$t('搜索结果为空')"
    @change="handleSelectChange">
    <template #trigger>
      <div class="selector-trigger">
        <bk-overflow-title v-if="currentValue" class="app-name" type="tips">
          {{ currentValue }}
        </bk-overflow-title>
        <span v-else class="no-app">{{ $t('请选择') }}</span>
        <AngleUpFill class="arrow-icon arrow-fill" />
        <span v-show="isError" class="error-msg">
          {{ $t('请先选择BCS 集群 ID，替换下方示例代码后，再尝试复制示例') }}
        </span>
      </div>
    </template>
    <bk-option-group
      v-for="groupItem in projectGroup"
      :key="groupItem.groupName"
      :label="groupItem.groupName"
      collapsible>
      <bk-option
        v-if="groupItem.children.length"
        v-for="item in groupItem.children"
        :key="item.name"
        :id="`${item.name}：${item.desc}`"
        :label="item.name + item.desc">
        <div class="cluster-option-item">
          <div class="item-name">{{ item.name }}</div>
          <div class="item-desc">{{ item.desc }}</div>
        </div>
      </bk-option>
      <div v-else class="bk-select-option no-data">{{ $t('暂无数据') }}</div>
    </bk-option-group>
  </bk-select>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  // import { useRoute } from 'vue-router';
  // import { ICredentialItem } from '../../../../../../types/client';
  // import { getClusterList } from '../../../../../api/client';
  import { AngleUpFill } from 'bkui-vue/lib/icon';

  interface IgroupItem {
    groupName: string;
    children: IchildrenInfo[];
  }
  interface IchildrenInfo {
    name: string;
    desc: string;
  }

  const emits = defineEmits(['current-cluster']);

  // const route = useRoute();
  // const router = useRouter();

  // const bizId = ref(String(route.params.spaceId));
  const isError = ref(false);
  const loading = ref(true);
  const currentValue = ref('');
  const projectGroup = ref<IgroupItem[]>([]);
  // const projectGroup = ref<IgroupItem[]>([
  //   {
  //     groupName: '项目A',
  //     children: [
  //       { name: 'bcs测试集群', desc: 'BCS-K8S-00000' },
  //       { name: '测试00000', desc: 'BCS-K8S-88888' },
  //     ],
  //   },
  //   {
  //     groupName: '项目B',
  //     children: [{ name: '蓝鲸 2.0 集群', desc: 'BCS-K8S-15336' }],
  //   },
  // ]);

  onMounted(() => {
    // loadClusterList();
  });

  // 表单校验失败检查集群ID是否为空
  const validateCluster = () => {
    isError.value = !currentValue.value;
  };
  // 获取集群ID
  // const loadClusterList = async () => {
  //   loading.value = true;
  //   try {
  //     // const query = {
  //     //   start: 0,
  //     //   all: true,
  //     // };
  //     const res = await getClusterList(bizId.value, 1);
  //     console.log(res);
  //     // credentialList.value = res.details;
  //   } catch (e) {
  //     console.error(e);
  //   } finally {
  //     loading.value = false;
  //   }
  // };

  // 下拉列表操作
  const handleSelectChange = (val: string) => {
    validateCluster();
    const parts = val.split('：');
    emits('current-cluster', { name: parts[0], value: parts[1] });
  };

  defineExpose({
    validateCluster,
  });
</script>

<style scoped lang="scss">
  .p2p-selector {
    &.select-error .selector-trigger {
      border-color: #ea3636;
    }
    &.popover-show .selector-trigger {
      border-color: #3a84ff;
      box-shadow: 0 0 3px #a3c5fd;
      .arrow-icon {
        transform: rotate(-180deg);
      }
    }
    &.is-focus {
      .selector-trigger {
        outline: 0;
      }
    }
    .selector-trigger {
      position: relative;
      padding: 0 10px 0;
      height: 32px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-radius: 2px;
      transition: all 0.3s;
      background: #ffffff;
      font-size: 14px;
      border: 1px solid #c4c6cc;
      .app-name {
        max-width: 220px;
        color: #313238;
      }
      .no-app {
        font-size: 12px;
        color: #c4c6cc;
      }
      .arrow-icon {
        margin-left: 13.5px;
        color: #979ba5;
        transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      }
      .error-msg {
        position: absolute;
        left: 0;
        bottom: -17px;
        font-size: 12px;
        line-height: 1;
        white-space: nowrap;
        color: #ea3636;
        animation: form-error-appear-animation 0.15s;
      }
    }
  }
  .cluster-option-item {
    padding: 5px 0;
    .item-name {
      color: #63656e;
    }
    .item-desc {
      color: #979ba5;
    }
  }
  .bk-popover.bk-pop2-content.bk-select-popover .bk-select-content-wrapper .bk-select-option {
    height: auto;
    line-height: 20px;
  }
  @keyframes form-error-appear-animation {
    0% {
      opacity: 0;
      transform: translateY(-30%);
    }
    100% {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
