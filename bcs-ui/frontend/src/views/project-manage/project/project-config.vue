<template>
  <bcs-dialog
    :value="value"
    theme="primary"
    :mask-close="false"
    header-position="left"
    :title="title"
    width="500"
    @value-change="handleDialogValueChange">
    <bk-form>
      <bk-form-item :label="$t('英文缩写')" required>
        <span>{{ curProject.projectCode }}</span>
      </bk-form-item>
      <bk-form-item :label="$t('编排类型')" required>
        <bk-radio-group v-model="kind">
          <bk-radio value="k8s" disabled>K8S</bk-radio>
          <!-- <bk-radio value="mesos" disabled v-if="$INTERNAL">Mesos</bk-radio> -->
        </bk-radio-group>
      </bk-form-item>
      <bk-form-item :label="$t('关联CMDB业务')" required>
        <div class="config-cmdb">
          <bcs-select
            v-if="ccList.length && !isHasCluster"
            v-model="ccKey"
            :loading="loading"
            :clearable="false"
            style="flex:1;"
            searchable>
            <bcs-option
              v-for="item in ccList"
              :key="item.businessID"
              :id="String(item.businessID)"
              :name="item.name">
            </bcs-option>
          </bcs-select>
          <bcs-input :value="curProject.cc_app_name" disabled v-else></bcs-input>
          <span class="ml5" v-bk-tooltips="$t('关联业务后，您可以从对应的业务下选择机器，搭建容器集群')">
            <i class="bcs-icon bcs-icon-info-circle"></i>
          </span>
        </div>
      </bk-form-item>
    </bk-form>
    <template #footer>
      <div class="dialog-footer">
        <span v-bk-tooltips="{ content: $t('该项目下已有集群信息，如需更改绑定业务信息，请先删除已有集群'), disabled: !isHasCluster }">
          <bk-button
            theme="primary"
            :disabled="isHasCluster || !ccList.length"
            :loading="saveLoading"
            @click="handleConfirm">
            {{ $t('保存') }}
          </bk-button>
        </span>
        <bk-button class="ml10" @click="handleCancel">{{ $t('取消') }}</bk-button>
      </div>
    </template>
  </bcs-dialog>
</template>
<script lang="ts">
/* eslint-disable camelcase */
import { computed, defineComponent, ref, toRefs, watch } from 'vue';
import useProject from '@/views/project-manage/project/use-project';
import $store from '@/store';
import $i18n from '@/i18n/i18n-setup';

export default defineComponent({
  name: 'ProjectConfig',
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { updateProject, getBusinessList } = useProject();
    const curProject = computed(() => $store.state.curProject);
    const title = computed(() => `${$i18n.t('项目')}【${curProject.value.project_name}】`);
    const isHasCluster = computed(() => $store.state.cluster.clusterList.length > 0);

    const loading = ref(false);
    const ccList = ref<any[]>([]);
    const fetchCCList = async () => {
      loading.value = true;
      ccList.value = await getBusinessList();
      loading.value = false;
    };

    const { value } = toRefs(props);
    watch(value, async () => {
      if (value.value) {
        kind.value = curProject.value.kind;
        await fetchCCList();
      }
    });

    const handleDialogValueChange = (value) => {
      ctx.emit('change', value);
    };

    const ccKey = ref(curProject.value.cc_app_id);
    const kind = ref(curProject.value.kind);

    const saveLoading = ref(false);

    const handleConfirm = async () => {
      saveLoading.value = true;
      const result = await updateProject(Object.assign({}, curProject.value, {
        // deploy_type 值固定，就是原来页面上的：部署类型：容器部署
        deployType: 2,
        // kind 业务编排类型
        kind: kind.value,
        // use_bk 值固定，就是原来页面上的：使用蓝鲸部署服务
        useBKRes: true,
        businessID: String(ccKey.value),
      }));
      saveLoading.value = false;
      handleCancel();
      result && window.location.reload();
    };
    const handleCancel = () => {
      handleDialogValueChange(false);
    };

    return {
      loading,
      title,
      ccList,
      kind,
      ccKey,
      curProject,
      isHasCluster,
      saveLoading,
      handleDialogValueChange,
      handleConfirm,
      handleCancel,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .config-cmdb {
  display: flex;
  align-items: center;
}
.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}
</style>
