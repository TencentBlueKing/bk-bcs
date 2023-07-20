<template>
  <div>
    <div class="flex justify-end mb15">
      <div class="flex items-center">
        <bcs-input
          v-model="searchValue"
          :placeholder="$t('请输入参数名称')"
          class="min-w-[278px]"
          right-icon="bk-icon icon-search"
          clearable>
        </bcs-input>
        <template v-if="!readonly">
          <i
            class="bcs-icon bcs-icon-zhongzhishuju ml15 text-[14px] cursor-pointer hover:text-[#3a84ff]"
            v-bk-tooltips.top="$t('重置参数')"
            @click="handleReset"></i>
          <i
            class="bcs-icon bcs-icon-yulan ml15 text-[14px] cursor-pointer hover:text-[#3a84ff]"
            v-bk-tooltips.top="$t('预览修改值')"
            @click="handlePreview"></i>
        </template>
      </div>
    </div>
    <bcs-table
      :data="curPageData"
      :pagination="pagination"
      v-bkloading="{ isLoading: loading }"
      @row-mouse-enter="handlekubeletMouseEnter"
      @page-change="pageChange"
      @page-limit-change="pageSizeChange">
      <bcs-table-column :label="$t('参数名称')" prop="flagName"></bcs-table-column>
      <bcs-table-column
        :label="$t('参数说明')"
        prop="flagDesc"
        show-overflow-tooltip>
      </bcs-table-column>
      <bcs-table-column :label="$t('默认值')" prop="defaultValue" v-if="!readonly"></bcs-table-column>
      <bcs-table-column :label="$t('当前值')">
        <template #default="{ row }">
          <div class="kubelet-value">
            <InputType
              v-if="editKey === row.flagName"
              :type="row.flagType"
              :options="row.flagValueList"
              :range="row.range"
              ref="editInputRef"
              v-model="kubeletParams[row.flagName]"
              @blur="handleEditBlur"
              @enter="handleEditBlur"
            ></InputType>
            <template v-else>
              <span>{{kubeletParams[row.flagName] || '--'}}</span>
              <i
                class="bcs-icon bcs-icon-edit2 ml5"
                v-show="activeKubeletFlagName === row.flagName"
                @click="handleEditkubelet(row)"></i>
            </template>
            <span
              class="error-tips" v-if="row.regex
                && kubeletParams[row.flagName]
                && !new RegExp(row.regex.validator).test(kubeletParams[row.flagName])">
              <i
                v-bk-tooltips="row.regex ? row.regex.message : ''"
                class="bk-icon icon-exclamation-circle-shape"></i>
            </span>
          </div>
        </template>
      </bcs-table-column>
      <template #empty>
        <BcsEmptyTableStatus :type="searchValue ? 'search-empty' : 'empty'" @clear="searchValue = ''" />
      </template>
    </bcs-table>
    <bcs-dialog
      :title="$t('预览修改值')"
      :show-footer="false"
      header-position="left"
      width="640"
      render-directive="if"
      v-model="showPreview">
      <bcs-table :data="kubeletDiffData">
        <bcs-table-column :label="$t('组件名称')" prop="moduleID"></bcs-table-column>
        <bcs-table-column :label="$t('组件参数')" prop="flagName"></bcs-table-column>
        <bcs-table-column :label="$t('修改前值')" prop="origin">
          <template #default="{ row }">
            {{row.origin || getDefaultValue(row)}}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('当前值')" prop="value"></bcs-table-column>
      </bcs-table>
    </bcs-dialog>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, computed, onMounted, toRefs, watch, getCurrentInstance } from 'vue';
import $store from '@/store/index';
import usePage from '@/composables/use-page';
import useSearch from '@/composables/use-search';
import InputType from '@/views/cluster-manage/components/input-type.vue';

export default defineComponent({
  name: 'KubeletParams',
  components: { InputType },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: String,
      default: '',
    },
    readonly: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { value } = toRefs(props);
    function handleTransformKubeletToParams(kubelet = '') {
      if (!kubelet) return {};

      return kubelet.split(';').reduce((pre, current) => {
        const index = current.indexOf('=');
        const key = current.slice(0, index);
        const value = current.slice(index + 1, current.length);
        if (key) {
          pre[key] = value;
        }
        return pre;
      }, {}) || {};
    };
    function handleTransformParamsToKubelet(params = {}) {
      return  Object.keys(params || {})
        .filter(key => params[key] !== '')
        .reduce<string[]>((pre, key) => {
        pre.push(`${key}=${params[key]}`);
        return pre;
      }, [])
        .join(';');
    }
    // kubelet 组件参数
    const loading = ref(false);
    const editKey = ref('');
    const showPreview = ref(false);
    const kubeletParams = ref(handleTransformKubeletToParams(value.value));

    watch(value, (newValue, oldValue) => {
      if (newValue === oldValue) return;
      kubeletParams.value = handleTransformKubeletToParams(value?.value);
    });
    watch(kubeletParams, () => {
      ctx.emit('change', handleTransformParamsToKubelet(kubeletParams.value));
    }, { deep: true });

    const originKubeletParams = ref<any>({});
    const kubeletDiffData = computed(() => Object.keys(kubeletParams.value).reduce<any[]>((pre, key) => {
      if (kubeletParams.value[key] !== ''
        && kubeletParams.value[key] !== originKubeletParams.value[key]) {
        pre.push({
          moduleID: 'kubelet',
          flagName: key,
          origin: originKubeletParams.value[key],
          value: kubeletParams.value[key],
        });
      }
      return pre;
    }, []));
    const kubeletList = ref<any[]>([]);
    const handleGetkubeletData = async () => {
      loading.value = true;
      kubeletList.value = await $store.dispatch('clustermanager/cloudModulesParamsList', {
        $cloudID: 'tencentCloud',
        $version: '1.20.6',
        $moduleID: 'kubelet',
      });
      loading.value = false;
    };
    const keys = ref(['flagName']);
    const { searchValue, tableDataMatchSearch } = useSearch(kubeletList, keys);
    const {
      pagination,
      curPageData,
      pageChange,
      pageSizeChange,
    } = usePage(tableDataMatchSearch);
    const editInputRef = ref<any>(null);
    const activeKubeletFlagName = ref('');
    const handlekubeletMouseEnter = (index, event, row) => {
      if (props.readonly) return;

      activeKubeletFlagName.value = row.flagName;
    };
    const { proxy } = getCurrentInstance() || { proxy: null };
    const handleEditkubelet = (row) => {
      editKey.value = row.flagName;
      const $refs = proxy?.$refs || {};
      setTimeout(() => {
        ($refs.editInputRef as any)?.focus();
      }, 0);
    };
    const handleEditBlur = () => {
      editKey.value = '';
    };
    const handleReset = () => {
      kubeletParams.value = JSON.parse(JSON.stringify(originKubeletParams.value));
    };
    const handlePreview = () => {
      showPreview.value = true;
    };
    // 校验kubelet参数
    const validateKubeletParams = () => kubeletList.value.every((item) => {
      if (!kubeletParams.value[item.flagName] || !item.regex?.validator) return true;

      const regx = new RegExp(item.regex.validator);
      return regx.test(kubeletParams.value[item.flagName]);
    });
    const getDefaultValue = row => kubeletList.value.find(item => item.flagName === row.flagName)?.defaultValue || '--';

    onMounted(() => {
      // kubelet原始数据（用于diff）
      originKubeletParams.value = JSON.parse(JSON.stringify(kubeletParams.value));
      handleGetkubeletData();
    });

    return {
      editInputRef,
      loading,
      searchValue,
      pagination,
      curPageData,
      editKey,
      showPreview,
      kubeletParams,
      pageChange,
      pageSizeChange,
      handleEditkubelet,
      handlePreview,
      handleReset,
      handleEditBlur,
      handlekubeletMouseEnter,
      activeKubeletFlagName,
      kubeletDiffData,
      validateKubeletParams,
      getDefaultValue,
    };
  },
});
</script>
<style lang="postcss" scoped>
.kubelet-value {
  position: relative;
  height: 32px;
  display: flex;
  align-items: center;
  .bcs-icon-edit2 {
      cursor: pointer;
      &:hover {
          color: #3a84ff;
      }
  }
  .error-tips {
      position: absolute;
      z-index: 10;
      right: 8px;
      top: 8px;
      color: #ea3636;
      cursor: pointer;
      font-size: 16px;
      display: flex;
      background-color: #fff;
  }
}
</style>
