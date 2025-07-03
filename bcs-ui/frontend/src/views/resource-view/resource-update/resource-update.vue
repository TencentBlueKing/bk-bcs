<template>
  <BcsContent :title="title" :cluster-id="clusterId" :namespace="isEdit ? namespace : ''" class="resource-content">
    <bcs-popconfirm
      class="switch-button-pop"
      :title="$t('dashboard.workload.editor.formMode.confirm')"
      :content="$t('dashboard.workload.editor.formMode.warnning')"
      width="280"
      trigger="click"
      v-if="formUpdate"
      @confirm="handleChangeMode">
      <FixedButton position="unset" :title="$t('deploy.variable.toForm')" />
    </bcs-popconfirm>
    <div :class="['resource-update', { 'full-screen': fullScreen }]">
      <template v-if="!showDiff">
        <div class="code-editor" ref="editorWrapperRef">
          <div class="top-operate">
            <span class="title bcs-ellipsis">{{ subTitle }}</span>
            <span class="tools">
              <span
                v-if="isEdit"
                v-bk-tooltips.top="$t('dashboard.workload.editor.reset')"
                @click="handleReset">
                <i class="bcs-icon bcs-icon-reset"></i>
              </span>
              <span class="upload" v-bk-tooltips.top="$t('dashboard.workload.editor.uploadYAML')">
                <input type="file" ref="fileRef" tabindex="-1" accept=".yaml,.yml" @change="handleFileChange">
                <i class="bcs-icon bcs-icon-upload"></i>
              </span>
              <span
                :class="{ active: showExample }"
                v-bk-tooltips.top="$t('dashboard.workload.editor.example')"
                @click="handleToggleExample">
                <i class="bcs-icon bcs-icon-code-example"></i>
              </span>
              <span
                v-bk-tooltips.top="fullScreen
                  ? $t('dashboard.workload.editor.zoomOut') : $t('dashboard.workload.editor.zoomIn')"
                @click="handleFullScreen">
                <i :class="['bcs-icon', fullScreen ? 'bcs-icon-zoom-out' : 'bcs-icon-enlarge']"></i>
              </span>
            </span>
          </div>
          <bcs-resize-layout
            placement="bottom"
            :auto-minimize="true"
            :initial-divide="editorErr.message ? 100 : 0"
            :max="300"
            :min="100"
            :disabled="!editorErr.message"
            :class="['custom-layout-cls', { 'hide-help': !editorErr.message }]"
            :style="{ 'height': fullScreen ? clientHeight + 'px' : height + 'px' }">
            <div slot="aside">
              <EditorStatus
                class="status-wrapper" :message="editorErr.message" v-show="!!editorErr.message"></EditorStatus>
            </div>
            <div slot="main">
              <CodeEditor
                v-model="detail"
                :height="fullScreen ? clientHeight : height"
                ref="editorRef"
                key="editor"
                v-bkloading="{ isLoading, opacity: 1, color: '#1a1a1a' }"
                @error="handleEditorErr">
              </CodeEditor>
            </div>
          </bcs-resize-layout>
        </div>
        <div class="code-example" ref="exampleWrapperRef" v-if="showExample">
          <div class="top-operate">
            <bk-dropdown-menu
              trigger="click" @show="isDropdownShow = true" @hide="isDropdownShow = false"
              v-if="examples.items && examples.items.length">
              <div class="dropdown-trigger-text" slot="dropdown-trigger">
                <span class="title">{{ activeExample.alias }}</span>
                <i :class="['bk-icon icon-angle-down',{ 'icon-flip': isDropdownShow }]"></i>
                <span
                  :class="['desc-icon',{ active: showDesc }]"
                  v-bk-tooltips.top="$t('generic.msg.info.tips')"
                  @click.stop="showDesc = !showDesc">
                  <i class="bcs-icon bcs-icon-prompt"></i>
                </span>
              </div>
              <ul class="bk-dropdown-list" slot="dropdown-content">
                <li v-for="(item, index) in (examples.items || [])" :key="index" @click="handleChangeExample(item)">
                  {{ item.alias }}
                </li>
              </ul>
            </bk-dropdown-menu>
            <span v-else><!-- 空元素为了flex布局 --></span>
            <span class="tools">
              <span
                v-bk-tooltips.top="$t('dashboard.workload.editor.copy')"
                @click="handleCopy">
                <i class="bcs-icon bcs-icon-copy"></i>
              </span>
              <span
                v-bk-tooltips.top="$t('blueking.help')"
                @click="handleHelp"><i :class="['bcs-icon bcs-icon-help-2', { active: showHelp }]"></i></span>
              <span
                v-bk-tooltips.top="$t('generic.button.close')"
                @click="showExample = false"><i class="bcs-icon bcs-icon-close-5"></i></span>
            </span>
          </div>
          <div class="example-desc" v-if="showDesc" ref="descWrapperRef">{{ activeExample.description }}</div>
          <bcs-resize-layout
            :class="['custom-layout-cls', { 'hide-help': !showHelp }]"
            :initial-divide="initialDivide"
            :disabled="!showHelp"
            :min="100"
            :max="600"
            :style="{ height: fullScreen ? '100%' : 'auto' }">
            <CodeEditor
              slot="aside"
              :value="activeExample.manifest"
              :height="fullScreen ? '100%' : exampleEditorHeight"
              :options="{
                renderLineHighlight: 'none'
              }"
              key="example"
              readonly
              v-bkloading="{ isLoading: exampleLoading, opacity: 1, color: '#1a1a1a' }">
            </CodeEditor>
            <bcs-md
              v-show="showHelp"
              slot="main"
              theme="dark"
              class="references"
              :style="{ height: fullScreen ? '100%' : exampleEditorHeight - 2 + 'px' }"
              :code="examples.references" />
          </bcs-resize-layout>
        </div>
      </template>
      <div class="code-diff" v-else>
        <div class="top-operate">
          <span class="title">
            {{ subTitle }}
            <span class="insert ml15">+{{ diffStat.insert }}</span>
            <span class="delete ml15">-{{ diffStat.delete }}</span>
          </span>
          <span class="diff-tools">
            <span @click="nextDiffChange"><i class="bcs-icon bcs-icon-arrows-down"></i></span>
            <span class="ml5" @click="previousDiffChange"><i class="bcs-icon bcs-icon-arrows-up"></i></span>
          </span>
        </div>
        <CodeEditor
          key="diff"
          :value="detail"
          :original="original"
          :height="fullScreen ? '100%' : height"
          :options="{
            renderLineHighlight: 'none'
          }"
          diff-editor
          readonly
          ref="diffEditorRef"
          @diff-stat="handleDiffStatChange">
        </CodeEditor>
        <EditorStatus
          class="status-wrapper diff"
          :message="editorErr.message" v-show="!!editorErr.message"></EditorStatus>
      </div>
    </div>
    <div class="resource-btn-group">
      <div
        v-bk-tooltips.top="{
          disabled: !disabledResourceUpdate,
          content: $t('dashboard.workload.editor.tips.contentUnchangedOrInvalidFormat')
        }">
        <bk-button
          theme="primary"
          class="main-btn"
          :loading="updateLoading"
          :disabled="disabledResourceUpdate"
          @click="handleCreateOrUpdate">
          {{ btnText }}
        </bk-button>
      </div>
      <bk-button
        class="ml10"
        v-if="isEdit"
        :disabled="!showDiff && disabledResourceUpdate"
        @click="toggleDiffEditor"
      >
        {{ showDiff
          ? $t('dashboard.workload.editor.continueEditing')
          : $t('dashboard.workload.editor.showDifference') }}
      </bk-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </BcsContent>
</template>
<script lang="ts">
/* eslint-disable no-unused-expressions */
import yamljs from 'js-yaml';
import { computed, defineComponent, onBeforeUnmount, onMounted, PropType, ref, toRefs, watch } from 'vue';

import EditorStatus from './editor-status.vue';
import FixedButton from './fixed-button.vue';

import { createCustomResource, customResourceDetail, updateCustomResource } from '@/api/modules/cluster-resource';
import $bkMessage from '@/common/bkmagic';
import { copyText } from '@/common/util';
import BcsMd from '@/components/bcs-md/index.vue';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import BcsContent from '@/components/layout/Content.vue';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';

export default defineComponent({
  name: 'ResourceUpdate',
  components: {
    CodeEditor,
    EditorStatus,
    BcsMd,
    FixedButton,
    BcsContent,
  },
  props: {
    // 命名空间（更新的时候需要--crd类型编辑是可能没有，创建的时候为空）
    namespace: {
      type: String,
      default: '',
    },
    // 父分类，eg: workloads、networks（注意复数）
    type: {
      type: String,
      default: '',
      required: true,
    },
    // 子分类，eg: deployments、ingresses
    category: {
      type: String,
      default: '',
    },
    // 名称（更新的时候需要，创建的时候为空）
    name: {
      type: String,
      default: '',
    },
    kind: {
      type: String,
      default: '',
    },
    // type 为crd时，必传
    crd: {
      type: String,
      default: '',
    },
    defaultShowExample: {
      type: Boolean,
      default: false,
    },
    formUpdate: {
      type: [Boolean, String],
      default: false,
    },
    // 从表单模式切换过来的数据
    formData: {
      type: Object,
      default: () => ({}),
    },
    // 从表单模式切换过来的原始数据
    defaultOriginal: {
      type: Object,
      default: () => null,
    },
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
    // CRD资源分两种，普通和定制，customized 用来区分普通和定制
    customized: {
      type: Boolean,
      default: false,
    },
    // CRD资源的版本
    version: {
      type: String,
      default: '',
    },
    // CRD资源的分组
    group: {
      type: String,
      default: '',
    },
    resource: {
      type: String,
      default: '',
    },
    // CRD资源的作用域
    scope: {
      type: String as PropType<'Namespaced'|'Cluster'>,
      default: '',
    },
  },
  setup(props, ctx) {
    const {
      namespace,
      type,
      category,
      name,
      kind,
      crd,
      defaultShowExample,
      formData,
      formUpdate,
      clusterId,
      defaultOriginal,
      customized,
      scope,
      version,
      group,
      resource,
    } = toRefs(props);
    const { clientHeight } = document.body;

    onMounted(() => {
      document.addEventListener('keyup', handleExitFullScreen);
      if (formData.value && Object.keys(formData.value).length) {
        // 从表单模式跳转过来
        handleGetFormDetail();
      } else {
        // manifest模式
        handleGetDetail();
      }

      handleGetExample();
      handleSetHeight();
    });

    const isEdit = computed(() =>  // 编辑态
      !!name.value);
    const title = computed(() => { // 导航title
      const prefix = isEdit.value ? $i18n.t('generic.button.update') : $i18n.t('generic.button.create');
      return `${prefix} ${kind.value}`;
    });

    // ====1.代码编辑器相关逻辑====
    const editorWrapperRef = ref<Element|null>(null);
    const editorRef = ref<any>(null);
    const fileRef = ref<any>(null);
    const isLoading = ref(false);
    const original = ref<any>({});
    const detail = ref<any>({});
    const showExample = ref(defaultShowExample.value);
    const fullScreen = ref(false);
    const height = ref(600);
    const editorErr = ref({
      type: '',
      message: '',
    });
    const webAnnotations = ref<any>({});
    const subTitle = computed(() =>  // 代码编辑器title
      detail.value?.metadata?.name || $i18n.t('dashboard.workload.editor.title.cr'));
    watch(fullScreen, (value) => {
      // 退出全屏后隐藏侧栏帮助文档（防止位置错乱）
      if (!value) {
        showHelp.value = false;
      }
    });
    const disabledResourceUpdate = computed(() => { // 禁用当前更新或者创建操作
      if (editorErr.value.message && editorErr.value.type === 'content') { // 编辑器格式错误
        return true;
      }
      if (isEdit.value) {
        return showDiff.value && !Object.keys(diffStat.value).some(key => diffStat.value[key]);
      }
      return !Object.keys(detail.value).length;
    });
    const setDetail = (data: any = {}) => { // 设置代码编辑器初始值
      // 特殊处理-> apiVersion、kind、metadata强制排序在前三位
      // const newManifest = {
      //   apiVersion: data.apiVersion,
      //   kind: data.kind,
      //   metadata: data.metadata,
      //   ...data,
      // };
      detail.value = {
        ...data,
      };
      editorRef.value?.setValue(Object.keys(detail.value).length ? detail.value : '');
    };
    const handleGetDetail = async () => { // 获取manifest详情
      if (!isEdit.value) return null;
      isLoading.value = true;
      let res: any = null;
      if (type.value === 'crd') {
        res = await customResourceDetail({
          format: 'manifest',
          $clusterId: clusterId.value,
          $name: name.value,
          namespace: namespace.value,
          group: group.value,
          version: version.value,
          resource: resource.value,
        }, { needRes: true }).catch(() => ({
          data: {
            manifest: {},
            manifestExt: {},
          },
        }));
      } else {
        res = await $store.dispatch('dashboard/getResourceDetail', {
          $namespaceId: namespace.value,
          $category: category.value,
          $name: name.value,
          $type: type.value,
          $clusterId: clusterId.value,
        });
      }
      original.value = JSON.parse(JSON.stringify(res.data?.manifest || {})); // 缓存原始值

      setDetail(res.data?.manifest);
      webAnnotations.value = res.webAnnotations;
      isLoading.value = false;
      return detail.value;
    };
    const handleGetFormDetail = async () => { // 获取表单模式详情
      isLoading.value = true;
      const data = await $store.dispatch('dashboard/renderManifestPreview', {
        kind: kind.value,
        formData: formData.value,
        $clusterId: clusterId.value,
      });
      if (defaultOriginal.value) {
        original.value = JSON.parse(JSON.stringify(defaultOriginal.value || {})); // 缓存原始值
      } else {
        original.value = JSON.parse(JSON.stringify(data || {})); // 缓存原始值
      }

      setDetail(data);
      isLoading.value = false;
    };
    const handleEditorChange = (code) => {
      ctx.emit('change', code);
      ctx.emit('input', code);
    };
    const handleResetEditorErr = () => {
      editorErr.value = {
        type: '',
        message: '',
      };
    };
    const handleReset = async () => { // 重置代码编辑器
      if (isLoading.value || !isEdit.value) return;

      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('dashboard.workload.editor.msg.confirmResetEditStatus'),
        subTitle: $i18n.t('dashboard.workload.editor.dialog.resetContentLoss'),
        defaultInfo: true,
        confirmFn: () => {
          handleResetEditorErr();
          setDetail(JSON.parse(JSON.stringify(original.value)));
        },
      });
    };
    const handleFileChange = (event) => { // 文件上传
      const [file] = event.target?.files || [];
      if (!file) return;

      const reader = new FileReader();
      reader.readAsText(file);
      reader.onload = () => {
        setDetail(yamljs.load(reader.result));
        fileRef.value && (fileRef.value.value = '');
      };
      reader.onerror = () => {
        $bkMessage({
          theme: 'error',
          message: reader.error,
        });
        fileRef.value && (fileRef.value.value = '');
      };
    };
    const handleToggleExample = () => { // 显示隐藏代码示例编辑器
      showExample.value = !showExample.value;
    };
    const handleFullScreen = () => { // 全屏
      fullScreen.value = !fullScreen.value;
      fullScreen.value && $bkMessage({
        theme: 'primary',
        message: $i18n.t('generic.button.fullScreen.msg'),
      });
    };
    const handleExitFullScreen = (event: KeyboardEvent) => { // esc退出全屏
      if (event.code === 'Escape') {
        fullScreen.value = false;
      }
    };
    const handleEditorErr = (err: string) => { // 捕获编辑器错误提示
      editorErr.value.type = 'content'; // 编辑内容错误
      editorErr.value.message = err;
    };
    const handleSetHeight = () => {
      const bounding = editorWrapperRef.value?.getBoundingClientRect();
      height.value = bounding ? bounding.height - 40 : 600; // 40: 编辑器顶部操作栏高度
    };

    // ====2.代码示例相关逻辑====
    const isDropdownShow = ref(false);
    const activeExample = ref<any>({});
    const exampleLoading = ref(false);
    const examples = ref<any>({});
    const showDesc = ref(true);
    const showHelp = ref(false);
    const exampleWrapperRef = ref<Element|null>(null);
    const descWrapperHeight = ref(0);
    const descWrapperRef = ref<Element|null>(null);
    const exampleEditorHeight = computed(() =>  // 代码示例高度
      height.value - descWrapperHeight.value);
    const initialDivide = computed(() => (showHelp.value ? '50%' : '100%'));
    watch(showDesc, () => {
      setTimeout(() => { // dom更新后获取描述文字的高度
        descWrapperHeight.value = showDesc.value ? descWrapperRef.value?.getBoundingClientRect()?.height || 0 : 0;
      }, 0);
    });

    const handleGetExample = async () => { // 获取示例模板
      // if (!showExample.value) return

      exampleLoading.value = true;
      examples.value = await $store.dispatch('dashboard/exampleManifests', {
        kind: kind.value, // crd类型的模板kind固定为CustomObject
        namespace: namespace.value,
        $clusterId: clusterId.value,
      });

      // 特殊处理-> apiVersion、kind、metadata强制排序在前三位
      // examples.value.items.forEach((example) => {
      //   const newManifestMap = {
      //     apiVersion: example.manifest.apiVersion,
      //     kind: example.manifest.kind,
      //     metadata: example.manifest.metadata,
      //     ...example.manifest,
      //   };
      //   example.manifest = newManifestMap;
      // });

      activeExample.value = examples.value?.items?.[0] || {};
      exampleLoading.value = false;
      return examples.value;
    };
    const handleChangeExample = (item) => { // 示例模板切换
      activeExample.value = item;
    };
    const handleCopy = () => { // 复制例子
      copyText(yamljs.dump(activeExample.value?.manifest));
      $bkMessage({
        theme: 'success',
        message: $i18n.t('dashboard.workload.editor.msg.copyExampleSuccess'),
      });
    };
    const handleHelp = () => {
      // 帮助文档
      showHelp.value = !showHelp.value;
    };

    // 3.====diff编辑器相关逻辑====
    const showDiff = ref(false);
    const updateLoading = ref(false);
    const diffStat = ref({
      insert: 0,
      delete: 0,
    });
    const diffEditorRef = ref<any>(null);
    const handleDiffStatChange = (stat) => {
      diffStat.value = stat;
    };
    const nextDiffChange = () => {
      diffEditorRef.value?.nextDiffChange();
    };
    const previousDiffChange = () => {
      diffEditorRef.value?.previousDiffChange();
    };

    // 4.====创建、更新、取消、显示差异====
    const btnText = computed(() => {
      if (!isEdit.value) return $i18n.t('generic.button.create');

      return showDiff.value ? $i18n.t('generic.button.update') : $i18n.t('generic.button.next');
    });
    const toggleDiffEditor = () => { // 显示diff
      showDiff.value = !showDiff.value;
      if (!showDiff.value) {
        handleResetEditorErr();
      }
    };
    const handleCreateResource = async () => {
      let result = false;
      if (customized.value) { // 创建普通crd资源
        result = await createCustomResource({
          $clusterId: clusterId.value,
          format: 'manifest',
          group: group.value,
          version: version.value,
          resource: resource.value,
          namespaced: scope.value === 'Namespaced',
          rawData: detail.value,
        }).catch((err) => {
          editorErr.value.type = 'http';
          editorErr.value.message = err?.response?.data?.message || err?.message;
          return false;
        });
      } else if (type.value === 'crd') { // 创建定制crd资源 bscpConfig、gameDeployment、gameStatefulSet、hookTemplate
        result = await $store.dispatch('dashboard/customResourceCreate', {
          $crd: crd.value,
          $category: category.value,
          $clusterId: clusterId.value,
          format: 'manifest',
          rawData: detail.value,
        }).catch((err) => {
          editorErr.value.type = 'http';
          editorErr.value.message = err?.response?.data?.message || err?.message;
          return false;
        });
      } else {
        result = await $store.dispatch('dashboard/resourceCreate', {
          $type: type.value,
          $category: category.value,
          $clusterId: clusterId.value,
          format: 'manifest',
          rawData: detail.value,
        }).catch((err) => {
          editorErr.value.type = 'http';
          editorErr.value.message = err?.response?.data?.message || err?.message;
          return false;
        });
      }

      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.create'),
        });
        $store.commit('updateCurNamespace', detail.value.metadata?.namespace);
        // 跳转回列表页
        $router.back();
      }
    };
    const handleUpdateResource = () => {
      if (!showDiff.value) {
        showDiff.value = true;
        return;
      }

      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('dashboard.workload.editor.dialog.confirmResourceUpdate'),
        subTitle: $i18n.t('dashboard.workload.editor.msg.replaceConflictWarning'),
        defaultInfo: true,
        confirmFn: async () => {
          let result = false;
          if (customized.value) {
            result = await updateCustomResource({
              $clusterId: clusterId.value,
              group: group.value,
              version: version.value,
              resource: resource.value,
              format: 'manifest',
              rawData: detail.value,
            }).catch((err) => {
              editorErr.value.type = 'http';
              editorErr.value.message = err?.response?.data?.message || err?.message;
              return false;
            });
          } else if (type.value === 'crd') {
            result = await $store.dispatch('dashboard/customResourceUpdate', {
              $crd: crd.value,
              $category: category.value,
              $clusterId: clusterId.value,
              $name: name.value,
              rawData: detail.value,
              format: 'manifest',
              namespace: namespace.value,
            }).catch((err) => {
              editorErr.value.type = 'http';
              editorErr.value.message = err?.response?.data?.message || err?.message;
              return false;
            });
          } else {
            result = await $store.dispatch('dashboard/resourceUpdate', {
              $namespaceId: namespace.value,
              $type: type.value,
              $category: category.value,
              $clusterId: clusterId.value,
              $name: name.value,
              format: 'manifest',
              rawData: detail.value,
            }).catch((err) => {
              editorErr.value.type = 'http';
              editorErr.value.message = err?.response?.data?.message || err?.message;
              return false;
            });
          }

          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.update'),
            });
            // 跳转回列表页
            $router.back();
          }
        },
      });
    };
    const handleCreateOrUpdate = async () => { // 更新或创建
      updateLoading.value = true;
      if (isEdit.value) {
        await handleUpdateResource();
      } else {
        await handleCreateResource();
      }
      updateLoading.value = false;
    };
    const handleCancel = () => { // 取消
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('generic.msg.info.exitEdit.text'),
        subTitle: $i18n.t('generic.msg.info.exitEdit.subTitle'),
        defaultInfo: true,
        confirmFn: () => {
          // 跳转回列表页
          $router.back();
        },
      });
    };
    // 切换到表单模式
    const handleChangeMode = () => {
      const crdData =  props.crd ? { crd: props.crd } : {};
      $router.replace({
        name: 'dashboardFormResourceUpdate',
        params: {
          ...(isEdit.value ? { name: name.value } : {}),
          formData: formData.value as any,
          namespace: namespace.value,
        },
        query: {
          type: type.value,
          category: category.value,
          kind: kind.value,
          formUpdate: formUpdate.value as any,
          ...crdData,
        },
      });
    };

    onBeforeUnmount(() => {
      document.removeEventListener('keyup', handleExitFullScreen);
    });

    return {
      showDiff,
      isEdit,
      title,
      subTitle,
      original,
      detail,
      editorRef,
      isLoading,
      exampleLoading,
      isDropdownShow,
      activeExample,
      examples,
      showExample,
      showDesc,
      showHelp,
      initialDivide,
      fullScreen,
      height,
      disabledResourceUpdate,
      exampleEditorHeight,
      editorErr,
      updateLoading,
      diffStat,
      editorWrapperRef,
      exampleWrapperRef,
      descWrapperRef,
      fileRef,
      btnText,
      handleChangeExample,
      handleGetDetail,
      handleEditorChange,
      handleReset,
      handleFileChange,
      handleToggleExample,
      handleFullScreen,
      handleCopy,
      handleHelp,
      handleGetExample,
      handleEditorErr,
      toggleDiffEditor,
      handleCreateOrUpdate,
      handleCancel,
      handleDiffStatChange,
      clientHeight,
      handleChangeMode,
      diffEditorRef,
      nextDiffChange,
      previousDiffChange,
    };
  },
});
</script>
<style lang="postcss" scoped>
.resource-content {
    padding-bottom: 0;
    height: 100%;
    .switch-button-pop {
        position: absolute;
        right: 32px;
        top: 130px;
        z-index: 1;
    }
    .icon-back {
        font-size: 16px;
        font-weight: bold;
        color: #3A84FF;
        margin-left: 20px;
        cursor: pointer;
    }
    .dashboard-top-title {
        display: inline-block;
        height: 60px;
        line-height: 60px;
        font-size: 16px;
        margin-left: 0px;
    }
    .resource-update {
        width: 100%;
        height: calc(100% - 44px);
        border-radius: 2px;
        display: flex;
        &.full-screen {
            position: fixed;
            top: 0;
            right: 0;
            bottom: 0;
            left: 0;
            height: 100% !important;
            width: 100% !important;
            z-index: 999;
            padding: 0;
        }
        .custom-layout-cls {
            border: none;
            /deep/ {
                .bk-resize-layout-aside {
                    border-color: #292929;
                    &:after {
                        right: -6px;
                    }
                }
            }
            &.hide-help {
                /deep/ .bk-resize-layout-aside:after {
                    display: none;
                }
            }
        }
        .top-operate {
            display: flex;
            align-items: center;
            justify-content: space-between;
            background: #2e2e2e;
            height: 40px;
            padding: 0 10px 0 16px;
            color: #c4c6cc;
            i {
                &:hover, &.active {
                    color: #699df4;
                }
            }
            .title {
                font-size: 14px;
            }
            .tools {
                display: flex;
                font-size: 16px;
                span {
                    width: 26px;
                    height: 26px;
                    margin-left: 5px;
                    cursor: pointer;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    &.active {
                        color: #699df4;
                        background: #242424;
                    }
                    &:hover {
                        color: #699df4;
                    }
                }
            }

            .diff-tools {
              font-size: 18px;
              i {
                cursor: pointer;
              }
            }
        }
        .code-editor {
            flex: 1;
            width: 0;
            position: relative;
            .upload {
                position: relative;
                input {
                    width: 100%;
                    height: 100%;
                    position: absolute;
                    left: 0;
                    top: 0;
                    cursor: pointer;
                    opacity: 0;
                }
            }
        }
        .code-example {
            flex: 1;
            width: 0;
            margin-left: 2px;
            overflow: hidden;
            /deep/ .dropdown-trigger-text {
                display: flex;
                align-items: center;
                justify-content: center;
                cursor: pointer;
                font-size: 14px;
                .icon-angle-down {
                    font-size: 20px;
                }
            }
            /deep/ .bk-dropdown-list {
                li {
                    height: 32px;
                    line-height: 32px;
                    padding: 0 16px;
                    color: #63656e;
                    font-size: 12px;
                    white-space: nowrap;
                    cursor: pointer;
                    &:hover {
                        background-color: #eaf3ff;
                        color: #3a84ff;
                    }
                }
            }
            .desc-icon {
                width: 26px;
                height: 26px;
                display: flex;
                align-items: center;
                justify-content: center;
                &.active {
                    color: #699df4;
                    background: #242424;
                }
            }
            .example-desc {
                background: #292929;
                border: 1px solid #141414;
                font-size: 12px;
                color: #b0b2b8;
                padding: 15px;
            }
            /deep/ .bk-resize-layout-main {
                background-color: #1a1a1a;
            }
            .bcs-md-preview {
                background-color: #2e2e2e !important;
            }
            .references {
                margin: 1px;
            }
        }
        .code-diff {
            width: 100%;
            position: relative;
            .status-wrapper.diff {
                height: 20%;
            }
            .insert {
                color: #5e8a48;
            }
            .delete {
                color: #e66565;
            }
        }
    }
    .resource-btn-group {
        margin-top: 12px;
        display: flex;
        align-items: center;
        button {
            min-width: 80px;
        }
        .main-btn {
            min-width: 100px;
        }
    }
    /deep/ .bk-resize-layout {
        border: none;
    }
}
</style>
