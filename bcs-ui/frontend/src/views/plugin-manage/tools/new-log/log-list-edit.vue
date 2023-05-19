<template>
  <section v-bkloading="{ isLoading }">
    <BkForm :model="detailData" :rules="rules" form-type="vertical" ref="formRef">
      <BkFormItem
        :label="$t('名称')"
        required
        property="name"
        error-display-type="normal">
        <bcs-input class="log-name" :disabled="isEdit" v-model="detailData.name"></bcs-input>
      </BkFormItem>
      <div class="form-row">
        <BkFormItem :label="$t('所属集群')" required class="mr25">
          <bcs-select :value="clusterId" disabled>
            <bcs-option
              v-for="item in clusterList"
              :key="item.clusterID"
              :id="item.clusterID"
              :name="item.clusterName">
            </bcs-option>
          </bcs-select>
        </BkFormItem>
        <BkFormItem
          :label="$t('命名空间')" required
          property="namespace"
          error-display-type="normal">
          <NamespaceSelect :cluster-id="clusterId" v-model="detailData.namespace" :disabled="isEdit"></NamespaceSelect>
        </BkFormItem>
      </div>
      <BkFormItem :label="$t('日志源')">
        <bk-radio-group class="mb10" v-model="detailData.config_selected">
          <bk-radio :disabled="isEdit" value="SelectedContainers">{{$t('指定容器')}}</bk-radio>
          <bk-radio :disabled="isEdit" value="AllContainers">{{$t('所有容器')}}</bk-radio>
          <bk-radio :disabled="isEdit" value="SelectedLabels">{{$t('指定标签')}}</bk-radio>
        </bk-radio-group>
        <section class="log-wrapper">
          <!--指定容器-->
          <section v-if="detailData.config_selected === 'SelectedContainers'">
            <div class="form-row">
              <BkFormItem
                class="mr25"
                :label="$t('应用类型')"
                property="config.workload.kind"
                error-display-type="normal">
                <bcs-select
                  v-model="detailData.config.workload.kind"
                  :placeholder="$t('应用类型')"
                  :disabled="isEdit">
                  <bcs-option
                    v-for="kind in kinds"
                    :key="kind"
                    :id="kind"
                    :name="kind">
                  </bcs-option>
                </bcs-select>
              </BkFormItem>
              <BkFormItem
                :label="$t('应用名称')"
                property="config.workload.name"
                error-display-type="normal">
                <bcs-input
                  v-model="detailData.config.workload.name"
                  :disabled="isEdit"
                  :placeholder="$t('请输入应用名称，支持正则匹配')">
                </bcs-input>
              </BkFormItem>
            </div>
            <BkFormItem
              :label="$t('采集路径')"
              property="config.workload.containers"
              error-display-type="normal">
              <div
                v-for="(item, index) in detailData.config.workload.containers"
                :key="index"
                class="container-wrapper mb10">
                <div style="display: flex;">
                  <div class="container-item mr25">
                    <label>{{$t('容器名')}}:</label>
                    <bcs-input v-model="item.container_name"></bcs-input>
                  </div>
                  <div class="container-item">
                    <label>{{$t('标准输出')}}:</label>
                    <bcs-checkbox v-model="item.enable_stdout">
                      {{$t('是否采集')}}
                      <i
                        class="bcs-icon bcs-icon-question-circle"
                        v-bk-tooltips.top="$t('如果不勾选，将不采集此容器的标准输出')"></i>
                    </bcs-checkbox>
                  </div>
                </div>
                <div class="container-item mt15">
                  <label>{{$t('文件路径')}}:</label>
                  <bcs-input
                    type="textarea"
                    :placeholder="$t('多个以;分隔')"
                    :value="item.paths.join('\n')"
                    class="log-path"
                    @change="(value) =>
                      item.paths = value.replace(/;/g, '\n').split('\n').filter(path => !!path)"
                  ></bcs-input>
                  <bcs-popover class="ml5" placement="top" :delay="500">
                    <i class="path-tip bcs-icon bcs-icon-question-circle"></i>
                    <div slot="content">
                      <p>1. {{$t('请填写文件的绝对路径，不支持目录')}}</p>
                      <p>2. {{$t('支持通配符，但通配符仅支持文件级别的')}}</p>
                      <p>{{$t('有效的示例: /data/log/*/*.log; /data/test.log; /data/log/log.*')}}</p>
                      <p>{{$t('无效的示例: /data/log/*; /data/log')}}</p>
                    </div>
                  </bcs-popover>
                </div>
                <span
                  class="panel-delete"
                  v-if="detailData.config.workload.containers.length > 1"
                  @click="handleDeleteConfig(index)">
                  <i class="bk-icon icon-close3-shape"></i>
                </span>
              </div>
              <div class="add-panel-btn mt10 mb10" @click="handleAddContainerConfig">
                <i class="bk-icon left-icon icon-plus"></i>
                <span>{{$t('点击增加')}}</span>
              </div>
            </BkFormItem>
          </section>
          <!--所有容器-->
          <section v-else-if="detailData.config_selected === 'AllContainers'">
            <div class="container-item">
              <label>{{$t('标准输出')}}:</label>
              <bcs-checkbox v-model="detailData.config.all_containers.enable_stdout">
                {{$t('是否采集')}}
                <i
                  class="bcs-icon bcs-icon-question-circle"
                  v-bk-tooltips.top="$t('如果不勾选，将不采集此容器的标准输出')"></i>
              </bcs-checkbox>
            </div>
            <div class="container-item mt15">
              <label>{{$t('文件路径')}}:</label>
              <bcs-input
                type="textarea"
                :value="detailData.config.all_containers.paths.join('\n')"
                @change="(value) =>
                  detailData.config.all_containers.paths = value
                    .replace(/;/g, '\n').split('\n').filter(path => !!path)">
              </bcs-input>
              <bcs-popover class="ml5" placement="top" :delay="500">
                <i class="path-tip bcs-icon bcs-icon-question-circle"></i>
                <div slot="content">
                  <p>1. {{$t('请填写文件的绝对路径，不支持目录')}}</p>
                  <p>2. {{$t('支持通配符，但通配符仅支持文件级别的')}}</p>
                  <p>{{$t('有效的示例: /data/log/*/*.log; /data/test.log; /data/log/log.*')}}</p>
                  <p>{{$t('无效的示例: /data/log/*; /data/log')}}</p>
                </div>
              </bcs-popover>
            </div>
          </section>
          <!--指定标签-->
          <section v-else-if="detailData.config_selected === 'SelectedLabels'">
            <BkFormItem
              :label="`${$t('匹配标签')}(labels)`"
              property="config.label_selector.match_labels"
              error-display-type="normal">
              <KeyValue v-model="detailData.config.label_selector.match_labels"></KeyValue>
            </BkFormItem>
            <BkFormItem
              class="mt15"
              :label-width="200"
              :label="`${$t('匹配表达式')}(expressions)`"
              property="config.label_selector.match_expressions"
              error-display-type="normal">
              <span
                class="add-express-btn" v-if="!detailData.config.label_selector.match_expressions.length"
                @click="handleAddExpression">
                <i class="bk-icon icon-plus-circle-shape mr5"></i>
                {{$t('添加')}}
              </span>
              <template v-else>
                <div
                  v-for="(item, index) in detailData.config.label_selector.match_expressions"
                  :key="index"
                  class="key-value-item mb15">
                  <bcs-input v-model="item.key" :placeholder="$t('键')"></bcs-input>
                  <bcs-select
                    class="ml15 mr15"
                    style="background: #fff;min-width:132px"
                    v-model="item.operator">
                    <bcs-option
                      v-for="operate in operateList"
                      :key="operate"
                      :id="operate"
                      :name="operate">
                    </bcs-option>
                  </bcs-select>
                  <bcs-input v-model="item.value" :placeholder="$t('值')"></bcs-input>
                  <i class="bk-icon icon-plus-circle ml10 mr5" @click="handleAddExpression"></i>
                  <i class="bk-icon icon-minus-circle" @click="handleDeleteExpressItem(index)"></i>
                </div>
              </template>
            </BkFormItem>
            <BkFormItem class="mt15" :label="$t('采集路径')">
              <div class="container-wrapper">
                <div class="container-item">
                  <label>{{$t('标准输出')}}:</label>
                  <bcs-checkbox v-model="detailData.config.label_selector.enable_stdout">
                    {{$t('是否采集')}}
                    <i
                      class="bcs-icon bcs-icon-question-circle"
                      v-bk-tooltips.top="$t('如果不勾选，将不采集此容器的标准输出')"></i>
                  </bcs-checkbox>
                </div>
                <div class="container-item mt15">
                  <label>{{$t('文件路径')}}:</label>
                  <bcs-input
                    type="textarea"
                    :value="detailData.config.label_selector.paths.join('\n')"
                    @change="(value) =>
                      detailData.config.label_selector.paths = value
                        .replace(/;/g, '\n').split('\n').filter(path => !!path)">
                  </bcs-input>
                  <bcs-popover class="ml5" placement="top" :delay="500">
                    <i class="path-tip bcs-icon bcs-icon-question-circle"></i>
                    <div slot="content">
                      <p>1. {{$t('请填写文件的绝对路径，不支持目录')}}</p>
                      <p>2. {{$t('支持通配符，但通配符仅支持文件级别的')}}</p>
                      <p>{{$t('有效的示例: /data/log/*/*.log; /data/test.log; /data/log/log.*')}}</p>
                      <p>{{$t('无效的示例: /data/log/*; /data/log')}}</p>
                    </div>
                  </bcs-popover>
                </div>
              </div>
            </BkFormItem>
          </section>
        </section>
      </BkFormItem>
      <BkFormItem :label="$t('附加日志标签')">
        <section class="log-wrapper">
          <KeyValue v-model="detailData.extra_labels"></KeyValue>
          <bcs-checkbox v-model="detailData.add_pod_label">
            {{$t('是否自动添加Pod中的labels')}}
          </bcs-checkbox>
        </section>
      </BkFormItem>
      <div class="mt15">
        <bcs-button
          theme="primary"
          :loading="btnLoading"
          @click="handleUpdateOrCreate">
          {{name ? $t('更新') : $t('创建')}}
        </bcs-button>
        <bcs-button @click="handleCancel">{{$t('取消')}}</bcs-button>
      </div>
    </BkForm>
  </section>
</template>
<script lang="ts">
import { defineComponent, ref, computed, onMounted, toRefs, PropType } from 'vue';
import KeyValue from '@/components/key-value.vue';
import $store from '@/store/index';
import $i18n from '@/i18n/i18n-setup';
import BkForm from 'bk-magic-vue/lib/form';
import BkFormItem from 'bk-magic-vue/lib/form-item';
import useLog from './use-log';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import $bkMessage from '@/common/bkmagic';

export default defineComponent({
  components: { KeyValue, BkForm, BkFormItem, NamespaceSelect },
  props: {
    name: {
      type: String,
      default: '',
    },
    kinds: {
      type: Array as PropType<any[]>,
      default: () => [],
    },
    operateList: {
      type: Array,
      default: () => ['In', 'NotIn', 'Exists', 'DoesNotExist'],
    },
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props, ctx) {
    const { name } = toRefs(props);
    const isEdit = computed(() => !!name.value);
    const clusterList = computed(() => ($store.state as any).cluster.clusterList);

    const formRef = ref<any>(null);
    const detailData = ref<{
      config_selected: 'SelectedContainers' | 'AllContainers' | 'SelectedLabels'
      name: string
      namespace: string
      extra_labels: Record<string, string>
      add_pod_label: boolean
      config: {
        // AllContainers
        all_containers: {
          enable_stdout: boolean
          paths: string[]
        }
        // SelectedContainers
        workload: {
          name: string
          kind: string
          containers: Array<{
            container_name: string
            enable_stdout: boolean
            paths: string[]
          }>
        }
        // SelectedLabels
        label_selector: {
          enable_stdout: boolean
          paths: string[]
          match_labels: Record<string, string>
          match_expressions: any[]
        }
      }
    }>({
      config_selected: 'SelectedContainers',
      name: '',
      namespace: '',
      extra_labels: {},
      add_pod_label: false,
      config: {
        // AllContainers
        all_containers: {
          enable_stdout: true,
          paths: [],
        },
        // SelectedContainers
        workload: {
          name: '',
          kind: '',
          containers: [{
            container_name: '',
            enable_stdout: true,
            paths: [],
          }],
        },
        // SelectedLabels
        label_selector: {
          enable_stdout: true,
          paths: [],
          match_labels: {},
          match_expressions: [],
        },
      },
    });
    const rules = ref({
      name: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'custom',
        },
        {
          validator: () => /^[A-Za-z0-9_]+$/.test(detailData.value.name)
                              && detailData.value.name.length >= 5
                              && detailData.value.name.length <= 50,
          message: $i18n.t('名称只能包含数字、英文和下划线, 且长度在5 ~ 50之间'),
          trigger: 'custom',
        },
      ],
      namespace: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'custom',
        },
      ],
      'config.workload.kind': [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'custom',
        },
      ],
      'config.workload.name': [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'custom',
        },
      ],
      'config.workload.containers': [
        {
          validator: () => detailData.value.config.workload.containers.every(item => !!item.container_name),
          message: $i18n.t('输入容器名'),
          trigger: 'custom',
        },
      ],
      'config.label_selector.match_labels': [{
        validator: () => Object.keys(detailData.value.config.label_selector.match_labels).length
                          || detailData.value.config.label_selector.match_expressions.length,
        message: $i18n.t('匹配标签和匹配表达式不能同时为空'),
        trigger: 'custom',
      }],
      'config.label_selector.match_expressions': [{
        validator: () => Object.keys(detailData.value.config.label_selector.match_labels).length
                          || detailData.value.config.label_selector.match_expressions.length,
        message: $i18n.t('匹配标签和匹配表达式不能同时为空'),
        trigger: 'custom',
      }],
    });
    const handleAddContainerConfig = () => {
      detailData.value.config.workload.containers.push({
        container_name: '',
        enable_stdout: true,
        paths: [],
      });
    };
    const handleDeleteConfig = (index) => {
      detailData.value.config.workload.containers.splice(index, 1);
    };
    const handleAddExpression = () => {
      detailData.value.config.label_selector.match_expressions.push({
        key: '',
        operator: 'In',
        value: '',
      });
    };
    const handleDeleteExpressItem = (index) => {
      detailData.value.config.label_selector.match_expressions.splice(index, 1);
    };
    const btnLoading = ref(false);
    const handleUpdateOrCreate = async () => {
      const result = await formRef.value?.validate();
      if (!result) return;

      btnLoading.value = true;
      const data: any = {
        ...detailData.value,
        config: {},
      };
      if (detailData.value.config_selected === 'AllContainers') {
        data.config.all_containers = detailData.value.config.all_containers;
      }
      if (detailData.value.config_selected === 'SelectedContainers') {
        data.config.workload = detailData.value.config.workload;
      }
      if (detailData.value.config_selected === 'SelectedLabels') {
        data.config.label_selector = detailData.value.config.label_selector;
      }
      console.log(data);
      if (isEdit.value) {
        await handleEdit(data);
      } else {
        await handleCreate(data);
      }
      btnLoading.value = false;
    };

    const {
      handleUpdateLogRule,
      handleCreateLogRule,
      handleGetLogDetail,
    } = useLog();

    const handleEdit = async (data) => {
      const result = await handleUpdateLogRule({
        $clusterId: props.clusterId,
        $name: data.name,
        ...data,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('更新成功'),
        });
        ctx.emit('confirm');
      }
    };
    const handleCreate = async (data) => {
      const result = await handleCreateLogRule({
        $clusterId: props.clusterId,
        ...data,
      });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('创建成功'),
        });
        ctx.emit('confirm');
      }
    };
    const handleCancel = () => {
      ctx.emit('cancel');
    };

    const isLoading = ref(false);
    const handleGetDetail = async () => {
      if (!name.value) return;

      isLoading.value = true;
      detailData.value = await handleGetLogDetail({
        $name: name.value,
        $clusterId: props.clusterId,
      });
      isLoading.value = false;
    };

    onMounted(() => {
      handleGetDetail();
    });

    return {
      isEdit,
      formRef,
      btnLoading,
      isLoading,
      detailData,
      clusterList,
      rules,
      handleAddContainerConfig,
      handleDeleteConfig,
      handleAddExpression,
      handleDeleteExpressItem,
      handleUpdateOrCreate,
      handleCancel,
    };
  },
});
</script>
<style lang="postcss" scoped>
.log-name {
  width: 50%;
  padding-right: 15px;
}
.form-row {
  display: flex;
  align-items: flex-start;
  margin: 8px 0;
  >>> .bk-form-item {
      flex: 1;
      margin-top: 0px !important;
  }
}
.log-wrapper {
  border: 1px solid #dcdee5;
  border-radius: 2px;
  padding: 20px;
  background: #fafbfd;
  >>> .bk-label-text {
      color: #737987;
  }
}
.add-panel-btn {
  cursor: pointer;
  background: #fff;
  border: 1px dashed #c4c6cc;
  border-radius: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 42px;
  font-size: 14px;
  .bk-icon {
      font-size: 24px;
  }
  &:hover {
      border-color: #3a84ff;
      color: #3a84ff;
  }
  &.disabled {
      color: #C4C6CC;
      cursor: not-allowed;
      border-color: #C4C6CC;
  }
}
.container-wrapper {
  padding: 20px 15px;
  background: #fff;
  border-radius: 2px;
  border: 1px solid #dcdee5;
  position: relative;
  .panel-delete {
      position: absolute;
      cursor: pointer;
      color: #979ba5;
      top: 0;
      right: 8px;
      &:hover {
          color: #3a84ff;
      }
  }
}
.container-item {
  display: flex;
  align-items: center;
  flex: 1;
  font-size: 14px;
  label {
      min-width: 80px;
      color: #737987;
  }
  >>> .bk-form-control {
      flex: 1
  }
}
.add-express-btn {
  font-size: 14px;
  color: #3a84ff;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
}
.key-value-item {
  display: flex;
  align-items: center;
  .bk-icon {
      font-size: 20px;
      color: #979bA5;
      cursor: pointer;
  }
}
>>> .log-path .bk-textarea-wrapper {
  border-color: #c4c6cc;
}
</style>
