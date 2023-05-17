/* eslint-disable camelcase */
import { defineComponent, computed, ref, watch, onMounted, toRefs } from 'vue';
import { useSelectItemsNamespace } from '../namespace/use-namespace';
import usePage from '../../../composables/use-page';
import useSubscribe, { ISubscribeData, ISubscribeParams } from './use-subscribe';
import useTableData from './use-table-data';
import { padIPv6 } from '@/common/util';
import yamljs from 'js-yaml';
import './base-layout.css';
import fullScreen from '@/directives/full-screen';
import { CUR_SELECT_CRD } from '@/common/constant';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import Header from '@/components/layout/Header.vue';
import jp from 'jsonpath';
import useTableSort from '@/composables/use-table-sort';
import $store from '@/store';
import $i18n from '@/i18n/i18n-setup';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import $router from '@/router';

export default defineComponent({
  name: 'BaseLayout',
  components: { CodeEditor },
  directives: {
    'full-screen': fullScreen,
  },
  props: {
    title: {
      type: String,
      default: '',
      required: true,
    },
    // 父分类（crd类型的需要特殊处理），eg: workloads、networks（注意复数）
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
    // 轮询时类型（type为crd时，kind仅作为资源详情展示的title用），eg: Deployment、Ingress（注意首字母大写）
    kind: {
      type: String,
      default: '',
      required: true,
    },
    // 是否显示命名空间（不展示的话不会发送获取命名空间列表的请求）
    showNameSpace: {
      type: Boolean,
      default: true,
    },
    // 是否显示创建资源按钮
    showCreate: {
      type: Boolean,
      default: true,
    },
    // 默认CRD值
    defaultCrd: {
      type: String,
      default: '',
    },
    // 是否显示crd下拉菜单
    showCrd: {
      type: Boolean,
      default: false,
    },
    // 是否显示总览和yaml切换的tab
    showDetailTab: {
      type: Boolean,
      default: true,
    },
    // 默认展示详情标签
    defaultActiveDetailType: {
      type: String,
      default: 'overview',
    },
  },
  setup(props) {
    const {
      type,
      category,
      kind,
      showNameSpace,
      showCrd,
      defaultActiveDetailType,
      defaultCrd,
    } = toRefs(props);
    const defaultCustomObjectsMap = ref(['gamedeployments.tkex.tencent.com', 'gamestatefulsets.tkex.tencent.com', 'hooktemplates.tkex.tencent.com']);
    const updateStrategyMap = ref({
      RollingUpdate: $i18n.t('滚动升级'),
      InplaceUpdate: $i18n.t('原地升级'),
      OnDelete: $i18n.t('手动删除'),
      Recreate: $i18n.t('重新创建'),
    });

    // crd
    const currentCrd = ref(defaultCrd.value || localStorage.getItem(CUR_SELECT_CRD) || '');
    const crdLoading = ref(false);
    // crd 数据
    const crdData = ref<ISubscribeData|null>(null);
    // crd 列表
    const crdList = computed(() => (crdData.value?.manifest?.items)
      ?.filter(i => !defaultCustomObjectsMap.value.includes(i.metadata.name)) || []);
    // 自定义资源当前CRD是从列表里面读取的
    const currentCrdExt = computed(() => {
      const item = crdList.value.find(item => item.metadata.name === currentCrd.value);
      return crdData.value?.manifestExt?.[item?.metadata?.uid] || {};
    });
    // 未选择crd时提示
    const crdTips = computed(() => (type.value === 'crd' && !currentCrd.value ? $i18n.t('请选择CRD') : ''));
    // 自定义资源的kind类型是根据选择的crd确定的
    const crdKind = computed(() => currentCrdExt.value.kind);
    // 自定义CRD（GameStatefulSets、GameDeployments、CustomObjects）
    const customCrd = computed(() => (type.value === 'crd' && kind.value !== 'CustomResourceDefinition'));
    const clusterId = computed(() => $store.getters.curClusterId);
    const handleGetCrdData = async () => {
      crdLoading.value = true;
      const res = await fetchCRDData(clusterId.value);
      crdData.value = res.data;
      crdLoading.value = false;
      // 校验初始化的crd值是否正确
      const crd = crdData.value?.manifest?.items?.find(item => item.metadata.name === currentCrd.value);
      if (!crd) {
        currentCrd.value = crdList.value[0]?.metadata?.name;
        localStorage.removeItem(CUR_SELECT_CRD);
      }
    };
    const handleCrdChange = async (value) => {
      localStorage.setItem(CUR_SELECT_CRD, value);
      handleGetTableData();
    };
    const renderCrdHeader = (h, { column }) => {
      const additionalData = additionalColumns.value.find(item => item.name === column.label);
      return h('span', {
        directives: [
          {
            name: 'bk-tooltips',
            value: {
              content: additionalData?.description || column.label,
              placement: 'top',
              boundary: 'window',
            },
          },
        ],
      }, [column.label]);
    };
    const getJsonPathValue = (row, path: string) => jp.value(row, path.indexOf('$') === 0 ? path : `$.${path}`);
    // 状态
    const statusMap = {
      normal: $i18n.t('正常'),
      creating: $i18n.t('创建中'),
      updating: $i18n.t('更新中'),
      deleting: $i18n.t('删除中'),
    };
    const statusFilters = computed(() => Object.keys(statusMap).map(key => ({
      text: statusMap[key],
      value: key,
    })));
    const statusFilterMethod = (value, row) => handleGetExtData(row.metadata.uid, 'status') === value;

    // 命名空间
    const namespaceDisabled = computed(() => {
      const { scope } = currentCrdExt.value;
      return type.value === 'crd' && scope && scope !== 'Namespaced';
    });
    // 获取命名空间
    const { namespaceLoading, namespaceValue, namespaceList, getNamespaceData } = useSelectItemsNamespace();

    // 表格数据
    const {
      isLoading,
      data,
      webAnnotations,
      handleFetchList,
      fetchCRDData,
      handleFetchCustomResourceList,
    } = useTableData();

    // 获取表格数据
    const handleGetTableData = async (subscribe = true) => {
      // 获取表格数据
      if (type.value === 'crd') {
        // crd scope 为Namespaced时需要传递命名空间（gamedeployments、gamestatefulsets、hooktemplates三个特殊的资源）
        const customResourceNamespace = currentCrdExt.value?.scope === 'Namespaced'
                    || defaultCustomObjectsMap.value.includes(defaultCrd.value)
          ? namespaceValue.value
          : undefined;
        // crd 界面无需传当前crd参数
        const crd = customCrd.value ? currentCrd.value : '';
        await handleFetchCustomResourceList(clusterId.value, crd, category.value, customResourceNamespace);
      } else {
        await handleFetchList(type.value, category.value, namespaceValue.value, clusterId.value);
      }

      // 重新订阅（获取表格数据之后，resourceVersion可能会变更）
      subscribe && handleStartSubscribe();
    };

    // 动态表格字段
    const additionalColumns = computed(() => webAnnotations.value.additionalColumns || []);
    // 表格数据（排序后）
    const allTableData = computed(() => data.value.manifest.items || []);
    const {
      sortTableData: tableData,
      handleSortChange,
    } = useTableSort(allTableData, item => handleGetExtData(item.metadata.uid) || {});
    const resourceVersion = computed(() => data.value.manifest?.metadata?.resourceVersion || '');

    // 模糊搜索功能
    const keys = ref(kind.value === 'Pod'
      ? ['metadata.name', 'creator', 'status.hostIP', 'podIPv4', 'podIPv6']
      : ['metadata.name', 'creator']); // 模糊搜索字段
    const searchValue = ref('');
    const tableDataMatchSearch = computed(() => {
      if (!searchValue.value) return tableData.value;

      return tableData.value.filter(item => keys.value.some((key) => {
        const extData = data.value.manifestExt?.[item.metadata?.uid] || {};
        const newItem = {
          ...extData,
          ...item,
        };
        const tmpKey = String(key).split('.');
        const str = tmpKey.reduce((pre, key) => {
          if (typeof pre === 'object') {
            return pre[key];
          }
          return pre;
        }, newItem);
        if (key === 'podIPv6') {
          return padIPv6(str).includes(padIPv6(searchValue.value));
        }
        return String(str).toLowerCase()
          .includes(padIPv6(searchValue.value.toLowerCase()));
      }));
    });

    const handleNamespaceSelected = (value) => {
      $store.commit('updateCurNamespace', value);
      handleGetTableData();
    };

    // 分页
    const { pagination, curPageData, pageConf, pageChange, pageSizeChange } = usePage(tableDataMatchSearch);
    // 搜索时重置分页
    watch([searchValue, namespaceValue, currentCrd], () => {
      pageConf.current = 1;
    });

    // 订阅事件
    const { handleSubscribe } = useSubscribe(data);
    const subscribeKind = computed(() =>
    // 自定义资源（非CustomResourceDefinition类型的crd）的kind是根据选择的crd动态获取的，不能取props的kind值
      (kind.value === 'CustomObject' ? crdKind.value : kind.value));

    // GameDeployment、GameStatefulSet apiVersion前端固定
    const apiVersion = computed(() => (['GameDeployment', 'GameStatefulSet'].includes(kind.value) ? 'tkex.tencent.com/v1alpha1' : currentCrdExt.value.api_version));

    const handleStartSubscribe = () => {
      // 自定义的CRD订阅时必须传apiVersion
      if (!subscribeKind.value || !resourceVersion.value || (customCrd.value && !apiVersion.value)) return;

      const params: ISubscribeParams = {
        kind: subscribeKind.value,
        resourceVersion: resourceVersion.value,
        namespace: namespaceValue.value,
      };
      if (apiVersion.value) {
        params.apiVersion = apiVersion.value;
      }
      if (customCrd.value) {
        params.CRDName = currentCrd.value;
      }
      handleSubscribe(params);
    };

    // 获取额外字段方法
    const handleGetExtData = (uid: string, ext?: string) => {
      const extData = data.value.manifestExt[uid] || {};
      return ext ? extData[ext] : extData;
    };

    // 跳转详情界面
    const gotoDetail = (row) => {
      $router.push({
        name: 'dashboardWorkloadDetail',
        params: {
          category: category.value,
          name: row.metadata.name,
          namespace: row.metadata.namespace,
          clusterId: clusterId.value,
        },
        query: {
          kind: subscribeKind.value,
          crd: currentCrd.value,
        },
      });
    };

    // 详情侧栏
    const showDetailPanel = ref(false);
    // 当前详情行数据
    const curDetailRow = ref<any>({
      data: {},
      extData: {},
      clusterId: clusterId.value,
    });
    // 侧栏展示类型
    const detailType = ref({
      active: defaultActiveDetailType.value,
      list: [
        {
          id: 'overview',
          name: $i18n.t('总览'),
        },
        {
          id: 'yaml',
          name: 'YAML',
        },
      ],
    });
    // 显示侧栏详情
    const handleShowDetail = (row) => {
      curDetailRow.value.data = row;
      curDetailRow.value.extData = handleGetExtData(row.metadata.uid);
      showDetailPanel.value = true;
    };

    const showCapacityDialog = ref(false);
    const replicas = ref(0);
    // 显示扩缩容弹框
    const handleEnlargeCapacity = (row) => {
      curDetailRow.value.data = row;
      replicas.value = row.spec.replicas;
      showCapacityDialog.value = true;
    };
    // 确定扩缩容
    const handleConfirmChangeCapacity = async () => {
      let result = false;
      const { name, namespace } = curDetailRow.value.data?.metadata || {};
      if (type.value === 'crd') {
        result = await $store.dispatch('dashboard/crdEnlargeCapacityChange', {
          $crdName: defaultCrd.value,
          $cobjName: name,
          $clusterId: clusterId.value,
          replicas: replicas.value,
          namespace,
        });
      } else {
        result = await $store.dispatch('dashboard/enlargeCapacityChange', {
          $namespace: namespace,
          $category: category.value,
          $name: name,
          $clusterId: clusterId.value,
          replicas: replicas.value,
        });
      }
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('修改成功'),
      });
      handleGetTableData();
    };

    // 切换详情类型
    const handleChangeDetailType = (type) => {
      detailType.value.active = type;
    };
    // 重置详情类型
    watch(showDetailPanel, () => {
      handleChangeDetailType(defaultActiveDetailType.value);
    });
    // yaml内容
    const yaml = computed(() => {
      // 特殊处理-> apiVersion、kind、metadata强制排序在前三位
      const newDetailRow = {
        apiVersion: curDetailRow.value.data.apiVersion,
        kind: curDetailRow.value.data.kind,
        metadata: curDetailRow.value.data.metadata,
        ...curDetailRow.value.data,
      };
      return yamljs.dump(newDetailRow || {});
    });
    // 创建资源
    const handleCreateResource = () => {
      const curKind = currentCrdExt.value.kind || kind.value;
      $router.push({
        name: 'dashboardResourceUpdate',
        params: {
          defaultShowExample: (kind.value !== 'CustomObject') as any,
          namespace: namespaceValue.value,
        },
        query: {
          type: type.value,
          category: category.value,
          kind: curKind,
          crd: currentCrd.value,
          formUpdate: webAnnotations.value?.featureFlag?.FORM_CREATE,
          menuId: showCrd.value ? 'CUSTOMOBJECT' : '',
        },
      });
    };
    // 创建资源（表单模式）
    const handleCreateFormResource = () => {
      const curKind = currentCrdExt.value.kind || kind.value;
      $router.push({
        name: 'dashboardFormResourceUpdate',
        params: {
          namespace: namespaceValue.value,
        },
        query: {
          type: type.value,
          category: category.value,
          kind: curKind,
          crd: currentCrd.value,
          formUpdate: webAnnotations.value?.featureFlag?.FORM_CREATE,
          menuId: showCrd.value ? 'CUSTOMOBJECT' : '',
        },
      });
    };
    // 更新资源
    const handleUpdateResource = (row) => {
      const { name, namespace, uid } = row.metadata || {};
      const editMode = handleGetExtData(uid, 'editMode');
      if (editMode === 'yaml') {
        $router.push({
          name: 'dashboardResourceUpdate',
          params: {
            namespace,
            name,
          },
          query: {
            type: type.value,
            category: category.value,
            kind: type.value === 'crd' ? kind.value : row.kind,
            crd: currentCrd.value,
            menuId: showCrd.value ? 'CUSTOMOBJECT' : '',
          },
        });
      } else {
        $router.push({
          name: 'dashboardFormResourceUpdate',
          params: {
            namespace,
            name,
          },
          query: {
            type: type.value,
            category: category.value,
            kind: type.value === 'crd' ? kind.value : row.kind,
            crd: currentCrd.value,
            formUpdate: webAnnotations.value?.featureFlag?.FORM_CREATE,
            menuId: showCrd.value ? 'CUSTOMOBJECT' : '',
          },
        });
      }
    };
    // 删除资源
    const handleDeleteResource = (row) => {
      const { name, namespace } = row.metadata || {};
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('确认删除当前资源'),
        subTitle: `${row.kind} ${name}`,
        defaultInfo: true,
        confirmFn: async () => {
          let result = false;
          if (type.value === 'crd') {
            result = await $store.dispatch('dashboard/customResourceDelete', {
              namespace,
              $crd: currentCrd.value,
              $category: category.value,
              $clusterId: clusterId.value,
              $name: name,
            });
          } else {
            result = await $store.dispatch('dashboard/resourceDelete', {
              $namespaceId: namespace,
              $type: type.value,
              $category: category.value,
              $clusterId: clusterId.value,
              $name: name,
            });
          };
          result && $bkMessage({
            theme: 'success',
            message: $i18n.t('删除成功'),
          });
          handleGetTableData();
        },
      });
    };

    onMounted(async () => {
      isLoading.value = true;
      const list: Promise<any>[] = [];
      // 获取命名空间下拉列表
      if (showNameSpace.value) {
        list.push(getNamespaceData({
          clusterId: clusterId.value,
        }));
      }

      // 获取CRD下拉列表
      if (showCrd.value) {
        list.push(handleGetCrdData());
      }
      await Promise.all(list); // 等待初始数据加载完毕

      await handleGetTableData(false);// 关闭默认触发订阅的逻辑，等待CRD类型的列表初始化完后开始订阅
      isLoading.value = false;

      // 所有资源就绪后开始订阅
      handleStartSubscribe();
    });

    // 清空搜索数据
    const handleClearSearchData = () => {
      searchValue.value = '';
    };
    return {
      updateStrategyMap,
      namespaceValue,
      namespaceLoading,
      namespaceDisabled,
      showDetailPanel,
      curDetailRow,
      replicas,
      yaml,
      detailType,
      isLoading,
      pageConf: pagination,
      nameValue: searchValue,
      data,
      curPageData,
      namespaceList,
      currentCrd,
      crdLoading,
      crdList,
      currentCrdExt,
      additionalColumns,
      crdTips,
      webAnnotations,
      statusMap,
      showCapacityDialog,
      getJsonPathValue,
      renderCrdHeader,
      stop,
      handlePageChange: pageChange,
      handlePageSizeChange: pageSizeChange,
      handleGetExtData,
      handleSortChange,
      gotoDetail,
      handleShowDetail,
      handleChangeDetailType,
      handleUpdateResource,
      handleDeleteResource,
      handleCreateResource,
      handleCreateFormResource,
      handleCrdChange,
      handleNamespaceSelected,
      handleEnlargeCapacity,
      handleConfirmChangeCapacity,
      statusFilters,
      statusFilterMethod,
      handleClearSearchData,
    };
  },
  render() {
    const renderCreate = () => {
      if (this.showCreate && !this.isLoading) {
        if (this.webAnnotations?.featureFlag?.FORM_CREATE) {
          return (
            <bk-dropdown-menu trigger="click" {...{
              scopedSlots: {
                'dropdown-trigger': () => (
                        <bk-button
                            theme="primary"
                            icon-right="icon-angle-down">
                            { this.$t('创建') }
                        </bk-button>
                ),
                'dropdown-content': () => (
                        <ul class="bk-dropdown-list">
                            <li onClick={this.handleCreateFormResource}><a href="javascript:;">{this.$t('表单模式')}</a></li>
                            <li onClick={this.handleCreateResource}>
                                <a href="javascript:;">{this.$t('YAML模式')}</a>
                            </li>
                        </ul>
                ),
              },
            }}>
            </bk-dropdown-menu>
          );
        }
        return (
          <bk-button
              icon="plus"
              theme="primary"
              onClick={this.handleCreateResource}>
              { this.$t('创建') }
          </bk-button>
        );
      }
      return <div></div>;
    };
    return (
      <div class="biz-content base-layout">
        <Header hide-back title={this.title} />
        <div class="biz-content-wrapper" v-bkloading={{ isLoading: this.isLoading, opacity: 1 }}>
          <div class="base-layout-operate mb20">
              {
                renderCreate()
              }
              <div class="search-wapper">
                  {
                    this.showCrd
                      ? (
                        <div class="select-wrapper">
                            <span class="select-prefix">CRD</span>
                            <bcs-select loading={this.crdLoading}
                                class="w-[250px] mr-[5px] bg-[#fff]"
                                v-model={this.currentCrd}
                                searchable
                                clearable={false}
                                placeholder={this.$t('选择CRD')}
                                onChange={this.handleCrdChange}>
                                {
                                    this.crdList.map(option => (
                                        <bcs-option
                                            key={option.metadata.name}
                                            id={option.metadata.name}
                                            name={option.metadata.name}>
                                        </bcs-option>
                                    ))
                                }
                            </bcs-select>
                        </div>
                      )
                      : null
                  }
                  {/** Scope类型不为Namespace时，隐藏命名空间方式会跳动，暂时用假的select替换 */}
                  {
                      this.showNameSpace
                        ? (
                          <div class="select-wrapper">
                              <span class="select-prefix">{this.$t('命名空间')}</span>
                              {
                                  this.namespaceDisabled
                                    ? <bcs-select
                                        class="w-[250px] mr-[5px] bg-[#fff]"
                                        placeholder={this.$t('请选择命名空间')}
                                        disabled />
                                    : <bcs-select
                                        v-bk-tooltips={{ disabled: !this.namespaceDisabled, content: this.crdTips }}
                                        loading={this.namespaceLoading}
                                        class="w-[250px] mr-[5px] bg-[#fff]"
                                        v-model={this.namespaceValue}
                                        onSelected={this.handleNamespaceSelected}
                                        searchable
                                        clearable={false}
                                        disabled={this.namespaceDisabled}
                                        placeholder={this.$t('请选择命名空间')}>
                                        {
                                            this.namespaceList.map(option => (
                                                <bcs-option
                                                    key={option.name}
                                                    id={option.name}
                                                    name={option.name}>
                                                </bcs-option>
                                            ))
                                        }
                                      </bcs-select>
                              }
                          </div>
                        )
                        : null
                  }
                  <bk-input
                      class="search-input"
                      clearable
                      v-model={this.nameValue}
                      right-icon="bk-icon icon-search"
                      placeholder={this.kind === 'Pod' ? this.$t('输入名称、创建人、IP搜索') : this.$t('输入名称、创建人搜索')}>
                  </bk-input>
              </div>
          </div>
          {
              this.$scopedSlots.default?.({
                isLoading: this.isLoading,
                pageConf: this.pageConf,
                data: this.data,
                curPageData: this.curPageData,
                statusMap: this.statusMap,
                handlePageChange: this.handlePageChange,
                handlePageSizeChange: this.handlePageSizeChange,
                handleGetExtData: this.handleGetExtData,
                handleSortChange: this.handleSortChange,
                gotoDetail: this.gotoDetail,
                handleShowDetail: this.handleShowDetail,
                handleEnlargeCapacity: this.handleEnlargeCapacity,
                handleUpdateResource: this.handleUpdateResource,
                handleDeleteResource: this.handleDeleteResource,
                getJsonPathValue: this.getJsonPathValue,
                renderCrdHeader: this.renderCrdHeader,
                additionalColumns: this.additionalColumns,
                namespaceDisabled: this.namespaceDisabled,
                webAnnotations: this.webAnnotations,
                updateStrategyMap: this.updateStrategyMap,
                statusFilters: this.statusFilters,
                statusFilterMethod: this.statusFilterMethod,
                nameValue: this.nameValue,
                handleClearSearchData: this.handleClearSearchData,
              })
          }
        </div>
        <bcs-sideslider
            quick-close
            isShow={this.showDetailPanel}
            width={800}
            {
            ...{
              on: {
                'update:isShow': (show: boolean) => {
                  this.showDetailPanel = show;
                },
              },
              scopedSlots: {
                header: () => (
                  <div class="detail-header">
                      <span>{this.curDetailRow.data?.metadata?.name}</span>
                      {
                        this.showDetailTab
                          ? (<div class="bk-button-group">
                                {
                                  this.detailType.list.map(item => (
                                    <bk-button class={{ 'is-selected': this.detailType.active === item.id }}
                                        onClick={() => {
                                          this.handleChangeDetailType(item.id);
                                        }}>
                                        {item.name}
                                    </bk-button>
                                  ))
                                }
                            </div>)
                          : null
                      }
                  </div>
                ),
                content: () => (this.detailType.active === 'overview'
                  ? (this.$scopedSlots.detail?.({
                    ...this.curDetailRow,
                  }))
                  : <CodeEditor
                    v-full-screen={{ tools: ['fullscreen', 'copy'], content: this.yaml }}
                    options={{
                      roundedSelection: false,
                      scrollBeyondLastLine: false,
                      renderLineHighlight: false,
                    }}
                    width="100%" height="100%" lang="yaml"
                    readonly={true} value={this.yaml} />),
              },
            }
            }></bcs-sideslider>
        <bcs-dialog
          v-model={this.showCapacityDialog}
          mask-close={false}
          title={`${this.curDetailRow?.data?.metadata?.name}${this.$t('扩缩容')}`}
          on-confirm={this.handleConfirmChangeCapacity}
        >
          <span class="capacity-dialog-content">
            { this.$t('实例数量') }
            <bk-input v-model={this.replicas} type="number" class="ml10" style="flex: 1;" min={0}></bk-input>
          </span>
        </bcs-dialog>
      </div>
    );
  },
});
