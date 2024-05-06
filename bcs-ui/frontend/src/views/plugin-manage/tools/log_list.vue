<!-- eslint-disable max-len -->
<template>
  <div class="biz-content">
    <Header :title="$t('plugin.tools.bklogconfig')" :desc="$t('plugin.tools.cluster', { name: clusterName })"></Header>
    <div class="biz-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
      <template v-if="!isInitLoading">
        <div class="biz-panel-header">
          <div class="left">
            <bk-button icon="plus" type="primary" @click.stop.prevent="createLoadBlance">
              <span>{{$t('plugin.tools.create')}}</span>
            </bk-button>
          </div>
          <div class="right search-wrapper">
            <div class="left">
              <bk-selector
                style="width: 135px;"
                :searchable="true"
                :placeholder="$t('k8s.namespace')"
                :selected.sync="searchParams.namespace"
                :list="nameSpaceList"
                :setting-key="'name'"
                :display-key="'name'"
                :allow-clear="true">
              </bk-selector>
            </div>
            <div class="left">
              <bk-selector
                style="width: 135px;"
                :placeholder="$t('plugin.tools.appType')"
                :selected.sync="searchParams.workload_type"
                :list="appTypes"
                :setting-key="'id'"
                :display-key="'name'"
                :allow-clear="true">
              </bk-selector>
            </div>
            <div class="left">
              <bkbcs-input
                style="width: 135px;"
                :placeholder="$t('plugin.tools._appName')"
                :value.sync="searchParams.workload_name">
              </bkbcs-input>
            </div>
            <div class="left">
              <bk-button type="primary" :title="$t('generic.button.query')" icon="search" @click="handleSearch">
                {{$t('generic.button.query')}}
              </bk-button>
            </div>
          </div>
        </div>

        <div class="biz-crd-instance">
          <div class="biz-table-wrapper">
            <bk-table
              class="biz-namespace-table"
              v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
              :size="'medium'"
              :data="curPageData"
              :pagination="pageConf"
              @page-change="handlePageChange"
              @page-limit-change="handlePageSizeChange">
              <bk-table-column :label="$t('generic.label.name')" prop="name" :show-overflow-tooltip="true" min-width="250">
                <template slot-scope="{ row }">
                  <a href="javascript: void(0)" class="bk-text-button biz-table-title biz-resource-title" @click.stop.prevent="editCrdInstance(row, true)">{{row.name || '--'}}</a>
                </template>
              </bk-table-column>
              <bk-table-column :label="`${$t('generic.label.cluster')} / ${$t('k8s.namespace')}`" min-width="220">
                <template slot-scope="{ row }">
                  <p>{{$t('generic.label.cluster1')}}：{{clusterName}}</p>
                  <p>{{$t('k8s.namespace')}}：{{(row.namespace === 'default' && row.config_type === 'default' && row.log_source_type === 'all_containers') ? $t('plugin.tools.all') : row.namespace}}</p>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('plugin.tools.logSRC')" min-width="100">
                <template slot-scope="{ row }">
                  {{logSource[row.log_source_type]}}
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('k8s.selector')" min-width="250">
                <template slot-scope="{ row }">
                  <template v-if="row.log_source_type === 'selected_containers'">
                    <p>{{$t('generic.label.kind')}}：{{row.crd_data.workload.type}}</p>
                    <p>{{$t('generic.label.name')}}：{{row.crd_data.workload.name || '--'}}</p>
                  </template>
                  <template v-else-if="row.log_source_type === 'selected_labels'">
                    <div class="data-item mt5" v-if="Object.keys(row.selector.match_labels).length">
                      <p class="key mb5">{{$t('plugin.tools.matchLables')}}：</p>
                      <p class="value">
                        <ul class="key-list">
                          <!-- eslint-disable vue/no-use-v-if-with-v-for -->
                          <li class="mb5" v-for="(value, key, labelIndex) in row.selector.match_labels" :key="labelIndex" v-if="labelIndex < 2">
                            <span class="key f12 m0" style="cursor: default;">{{key || '--'}}</span>
                            <span class="value f12 m0" style="cursor: default;">{{value || '--'}}</span>
                          </li>
                        </ul>
                      </p>
                    </div>
                    <div class="data-item" v-if="row.selector.match_expressions.length">
                      <p class="key mb5">{{$t('plugin.tools.matchPattern')}}：</p>
                      <p class="value">
                        <ul class="key-list">
                          <li class="mb5" v-for="(expression, expressIndex) of row.selector.match_expressions" :key="expressIndex" v-if="expressIndex <= 3">
                            <span class="key f12 m0">{{expression.key || '--'}}</span>
                            <span class="value f12 m0">{{expression.operator || '--'}}</span>
                            <span class="value f12 m0" v-if="expression.values">{{expression.values || '--'}}</span>
                          </li>
                        </ul>
                      </p>
                    </div>
                  </template>
                  <template v-else-if="row.log_source_type === 'all_containers'">
                    --
                  </template>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('plugin.tools.log')" min-width="200">
                <template slot-scope="{ row }">
                  <template v-if="row.log_source_type === 'selected_containers'">
                    <bcs-popover placement="top" :delay="500">
                      <p class="path-text" v-for="(conf, confIndex) of row.crd_data.workload.container_confs" :key="confIndex" style="display: block;">
                        {{conf.name}}：{{conf.log_paths[0] || '--'}}
                        <template v-if="conf.log_paths.length > 1">...</template>
                      </p>
                      <div slot="content">
                        <p v-for="(conf, confIndex) of row.crd_data.workload.container_confs" :key="confIndex">
                          {{conf.name}}：{{conf.log_paths.join(';') || '--'}}
                        </p>
                      </div>
                    </bcs-popover>
                  </template>
                  <template v-else-if="row.log_source_type === 'all_containers'">
                    --
                  </template>
                  <template v-if="row.log_source_type === 'selected_labels'">
                    <p>
                      {{row.selector.log_paths.join(';') || '--'}}
                    </p>
                  </template>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('projects.operateAudit.record')" width="260">
                <template slot-scope="{ row }">
                  <p>{{$t('generic.label.updator')}}：{{row.operator || '--'}}</p>
                  <p>{{$t('cluster.labels.updatedAt')}}：{{row.updated || '--'}}</p>
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('generic.label.status')" min-width="100">
                <template slot-scope="{ row }">
                  {{row.bind_success ? $t('generic.status.ready') : $t('generic.status.error')}}
                </template>
              </bk-table-column>
              <bk-table-column :label="$t('generic.label.action')" width="150">
                <template slot-scope="{ row }">
                  <a href="javascript:void(0);" class="bk-text-button" @click="editCrdInstance(row)">{{$t('generic.button.update')}}</a>
                  <a href="javascript:void(0);" class="bk-text-button" @click="removeCrdInstance(row)">{{$t('generic.button.delete')}}</a>
                </template>
              </bk-table-column>
            </bk-table>
          </div>
        </div>
      </template>

      <bk-sideslider
        :quick-close="false"
        :is-show.sync="crdInstanceSlider.isShow"
        :title="crdInstanceSlider.title"
        :width="800">
        <div class="p30" slot="content">
          <div class="bk-form bk-form-vertical">
            <div class="bk-form-item">
              <div class="bk-form-content">
                <div class="bk-form-inline-item is-required" style="width: 346px;">
                  <label class="bk-label">{{$t('generic.label.name')}}：</label>
                  <div class="bk-form-content">
                    <bkbcs-input
                      :placeholder="$t('generic.placeholder.input')"
                      :value.sync="curCrdInstanceName"
                      :disabled="true">
                    </bkbcs-input>
                  </div>
                </div>
              </div>
            </div>

            <div class="bk-form-item">
              <div class="bk-form-content">
                <div class="bk-form-inline-item is-required" style="width: 352px;">
                  <label class="bk-label">{{$t('generic.label.cluster1')}}：</label>
                  <div class="bk-form-content">
                    <bk-selector
                      :placeholder="$t('generic.placeholder.input')"
                      :setting-key="'cluster_id'"
                      :display-key="'name'"
                      :selected.sync="clusterId"
                      :list="clusterList"
                      :disabled="true">
                    </bk-selector>
                  </div>
                </div>

                <div :class="['bk-form-inline-item', { 'is-required': curCrdInstance.log_source_type !== 'all_containers' }]" style="width: 352px; margin-left: 25px;">
                  <label class="bk-label">{{$t('k8s.namespace')}}：<span class="biz-tip fn" v-if="curCrdInstance.log_source_type === 'all_containers'">({{$t('plugin.tools.allNS')}})</span></label>
                  <div class="bk-form-content">
                    <bkbcs-input
                      v-if="curCrdInstance.id && curCrdInstance.namespace === 'default' && curCrdInstance.config_type === 'default' && curCrdInstance.log_source_type === 'all_containers'"
                      :value="$t('plugin.tools.all')"
                      disabled />
                    <bk-selector
                      v-else
                      :searchable="true"
                      :placeholder="$t('generic.placeholder.select')"
                      :selected.sync="curCrdInstance.namespace"
                      :setting-key="'name'"
                      :list="nameSpaceList"
                      :disabled="curCrdInstance.id"
                      :allow-clear="curCrdInstance.log_source_type === 'all_containers'">
                    </bk-selector>
                  </div>
                </div>
              </div>
            </div>

            <div class="bk-form-item is-required mb5">
              <label class="bk-label">{{$t('plugin.tools.logSRC')}}：</label>
              <div class="bk-form-content">
                <bk-radio-group v-model="curCrdInstance.log_source_type">
                  <bk-radio :value="'selected_containers'" :disabled="curCrdInstance.id">{{$t('plugin.tools.container')}}</bk-radio>
                  <bk-radio :value="'all_containers'" :disabled="curCrdInstance.id">{{$t('plugin.tools.allContainers')}}</bk-radio>
                  <bk-radio :value="'selected_labels'" :disabled="curCrdInstance.id">{{$t('plugin.tools.label')}}</bk-radio>
                </bk-radio-group>
              </div>
            </div>

            <section class="log-wrapper" v-if="curCrdInstance.log_source_type === 'all_containers'">
              <div class="bk-form-content log-flex mb15">
                <div class="bk-form-inline-item">
                  <div class="log-form">
                    <div class="label">{{$t('logCollector.label.collectorType.stdout')}}：</div>
                    <div class="content" style="width: 223px; margin-right: 32px;">
                      <bk-checkbox name="cluster-classify-checkbox" v-model="curCrdInstance.default_conf.stdout" true-value="true" false-value="false">
                        {{$t('plugin.tools.enabled')}}
                        <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.right="$t('plugin.tools.stdoutTips')"></i>
                      </bk-checkbox>
                    </div>
                  </div>
                </div>

                <div class="bk-form-inline-item">
                  <div class="log-form">
                    <div class="label" style="width: 108px;">{{$t('plugin.tools.stdoutDataid')}}：</div>
                    <div class="content">
                      <bkbcs-input
                        style="width: 80px;"
                        type="number"
                        :min="0"
                        :placeholder="$t('generic.placeholder.input')"
                        :value.sync="curCrdInstance.default_conf.std_data_id"
                        :disabled="!curCrdInstance.default_conf.is_std_custom">
                      </bkbcs-input>
                      <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="curCrdInstance.default_conf.is_std_custom">
                        {{$t('plugin.tools.custom')}}
                        <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('plugin.tools.dataIdDesc') }"></i>
                      </bk-checkbox>
                    </div>
                  </div>
                </div>
              </div>

              <div class="bk-form-content log-flex">
                <div class="bk-form-inline-item">
                  <div class="log-form" style="width: 345px;">
                    <div class="label">{{$t('plugin.tools.path')}}：</div>
                    <div class="content log-path-wrapper" style="width: 223px;">
                      <textarea class="bk-form-textarea" v-model="curCrdInstance.default_conf.log_paths_str" style="width: 223px;" :placeholder="$t('plugin.tools.delimiter')"></textarea>

                      <bcs-popover placement="top" :delay="500">
                        <i class="path-tip bcs-icon bcs-icon-question-circle"></i>
                        <div slot="content">
                          <p>1. 请填写文件的绝对路径，不支持目录</p>
                          <p>2. 支持通配符，但通配符仅支持文件级别的</p>
                          <p>有效的示例: /data/log/*/*.log; /data/test.log; /data/log/log.*</p>
                          <p>无效的示例: /data/log/*; /data/log</p>
                        </div>
                      </bcs-popover>
                    </div>
                  </div>
                </div>

                <div class="bk-form-inline-item">
                  <div class="log-form">
                    <div class="label" style="width: 110px;">{{$t('plugin.tools.fileDataid')}}：</div>
                    <div class="content">
                      <bkbcs-input
                        style="width: 80px;"
                        type="number"
                        :min="0"
                        :placeholder="$t('generic.placeholder.input')"
                        :value.sync="curCrdInstance.default_conf.file_data_id"
                        :disabled="!curCrdInstance.default_conf.is_file_custom">
                      </bkbcs-input>
                      <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="curCrdInstance.default_conf.is_file_custom">
                        {{$t('plugin.tools.custom')}}
                        <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('plugin.tools.dataIdDesc') }"></i>
                      </bk-checkbox>
                    </div>
                  </div>
                </div>
              </div>
            </section>

            <section class="log-wrapper" v-if="curCrdInstance.log_source_type === 'selected_containers'">
              <div class="bk-form-item">
                <div class="bk-form-content log-flex mb15">
                  <div class="bk-form-inline-item" style="margin-right: 32px;">
                    <div class="log-form no-flex">
                      <div class="label">{{$t('plugin.tools.appType')}}：</div>
                      <div class="content">
                        <bk-selector
                          style="width: 330px;"
                          :placeholder="$t('plugin.tools.appType')"
                          :selected.sync="curCrdInstance.workload.type"
                          :list="appTypes"
                          :setting-key="'id'"
                          :display-key="'name'"
                          :allow-clear="true"
                          :disabled="curCrdInstance.id">
                        </bk-selector>
                      </div>
                    </div>
                  </div>

                  <div class="bk-form-inline-item">
                    <div class="log-form no-flex">
                      <div class="label">{{$t('plugin.tools.appName')}}：</div>
                      <div class="content">
                        <bkbcs-input
                          style="width: 333px;"
                          :placeholder="$t('plugin.tools.inputName')"
                          :value.sync="curCrdInstance.workload.name"
                          :disabled="curCrdInstance.id">
                        </bkbcs-input>
                      </div>
                    </div>
                  </div>

                </div>
              </div>

              <div class="bk-form-item">
                <div class="bk-form-content log-flex mb10">
                  <div class="log-form no-flex">
                    <div class="label">{{$t('plugin.tools.collectionPath')}}：</div>
                    <div class="content">
                      <section class="log-inner-wrapper mb10" v-for="(containerConf, index) of curCrdInstance.workload.container_confs" :key="index">
                        <div class="bk-form-content log-flex mb10">
                          <div class="bk-form-inline-item">
                            <div class="log-form">
                              <div class="label">{{$t('plugin.tools.containerName')}}：</div>
                              <div class="content">
                                <bkbcs-input
                                  style="width: 223px;"
                                  :placeholder="$t('generic.placeholder.input')"
                                  :value.sync="containerConf.name">
                                </bkbcs-input>
                              </div>
                            </div>
                          </div>
                        </div>

                        <div class="bk-form-content log-flex mb10">
                          <div class="bk-form-inline-item">
                            <div class="log-form">
                              <div class="label">{{$t('logCollector.label.collectorType.stdout')}}：</div>
                              <div class="content" style="width: 223px; margin-right: 32px;">
                                <bk-checkbox name="cluster-classify-checkbox" v-model="containerConf.stdout" true-value="true" false-value="false">
                                  {{$t('plugin.tools.enabled')}}
                                  <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.right="$t('plugin.tools.stdoutTips')"></i>
                                </bk-checkbox>
                              </div>
                            </div>
                          </div>

                          <div class="bk-form-inline-item">
                            <div class="log-form">
                              <div class="label" style="width: 108px;">{{$t('plugin.tools.stdoutDataid')}}：</div>
                              <div class="content">
                                <bkbcs-input
                                  style="width: 80px;"
                                  type="number"
                                  :min="0"
                                  :placeholder="$t('generic.placeholder.input')"
                                  :value.sync="containerConf.std_data_id"
                                  :disabled="!containerConf.is_std_custom">
                                </bkbcs-input>
                                <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="containerConf.is_std_custom">
                                  {{$t('plugin.tools.custom')}}
                                  <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('plugin.tools.dataIdDesc') }"></i>
                                </bk-checkbox>
                              </div>
                            </div>
                          </div>
                        </div>

                        <div class="bk-form-content log-flex">
                          <div class="bk-form-inline-item">
                            <div class="log-form" style="width: 345px;">
                              <div class="label">{{$t('plugin.tools.path')}}：</div>
                              <div class="content log-path-wrapper" style="width: 223px;">
                                <textarea class="bk-form-textarea" v-model="containerConf.log_paths_str" style="width: 223px;" :placeholder="$t('plugin.tools.delimiter')"></textarea>

                                <bcs-popover placement="top" :delay="500">
                                  <i class="path-tip bcs-icon bcs-icon-question-circle"></i>
                                  <div slot="content">
                                    <p>1. 请填写文件的绝对路径，不支持目录</p>
                                    <p>2. 支持通配符，但通配符仅支持文件级别的</p>
                                    <p>有效的示例: /data/log/*/*.log; /data/test.log; /data/log/log.*</p>
                                    <p>无效的示例: /data/log/*; /data/log</p>
                                  </div>
                                </bcs-popover>
                              </div>
                            </div>
                          </div>

                          <div class="bk-form-inline-item">
                            <div class="log-form">
                              <div class="label" style="width: 110px;">{{$t('plugin.tools.fileDataid')}}：</div>
                              <div class="content">
                                <bkbcs-input
                                  style="width: 80px;"
                                  type="number"
                                  :min="0"
                                  :placeholder="$t('generic.placeholder.input')"
                                  :value.sync="containerConf.file_data_id"
                                  :disabled="!containerConf.is_file_custom">
                                </bkbcs-input>
                                <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="containerConf.is_file_custom">
                                  {{$t('plugin.tools.custom')}}
                                  <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('plugin.tools.dataIdDesc') }"></i>
                                </bk-checkbox>
                              </div>
                            </div>
                          </div>
                        </div>

                        <i class="bcs-icon bcs-icon-close log-close" @click="removeContainerConf(containerConf, index)" v-if="curCrdInstance.workload.container_confs.length > 1"></i>
                      </section>

                      <bk-button class="log-block-btn mt10" @click="addContainerConf">
                        <i class="bcs-icon bcs-icon-plus"></i>
                        {{$t('plugin.tools.clickToAdd')}}
                      </bk-button>
                    </div>
                  </div>
                </div>
              </div>
            </section>

            <section class="log-wrapper" v-if="curCrdInstance.log_source_type === 'selected_labels'">
              <div class="bk-form-item">
                <div class="bk-form-content log-flex mb10">
                  <div class="log-form no-flex">
                    <div class="label tl" style="width: 300px;">{{$t('plugin.tools.labels')}}：</div>
                    <div class="content">
                      <bk-keyer
                        :key-input-width="265"
                        :value-input-width="265"
                        :key-list.sync="curCrdInstance.selector.match_labels_list"
                        :var-list="varList"
                        ref="matchLabelsKeyer"
                        @change="updateMatchLabels">
                        <p></p>
                      </bk-keyer>
                    </div>
                  </div>
                </div>
              </div>

              <div class="bk-form-item">
                <div class="bk-form-content log-flex mb10">
                  <div class="log-form no-flex">
                    <div class="label tl" style="width: 300px;">{{$t('plugin.tools.expressions')}}：</div>
                    <div class="content">
                      <bk-expression
                        :key-list.sync="curCrdInstance.selector.match_expressions_list"
                        :var-list="varList"
                        ref="expressions"
                        @change="updateExpressions">
                      </bk-expression>
                    </div>
                  </div>
                </div>
              </div>

              <div class="bk-form-item">
                <div class="bk-form-content log-flex mb10">
                  <div class="log-form no-flex">
                    <div class="label">{{$t('plugin.tools.collectionPath')}}：</div>
                    <div class="content">
                      <section class="log-inner-wrapper mb10">
                        <div class="bk-form-content log-flex mb10">
                          <div class="bk-form-inline-item">
                            <div class="log-form">
                              <div class="label">{{$t('logCollector.label.collectorType.stdout')}}：</div>
                              <div class="content" style="width: 223px; margin-right: 32px;">
                                <bk-checkbox name="cluster-classify-checkbox" v-model="curCrdInstance.selector.stdout" true-value="true" false-value="false">
                                  {{$t('plugin.tools.enabled')}}
                                  <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.right="$t('plugin.tools.stdoutTips')"></i>
                                </bk-checkbox>
                              </div>
                            </div>
                          </div>

                          <div class="bk-form-inline-item">
                            <div class="log-form">
                              <div class="label" style="width: 108px;">{{$t('plugin.tools.stdoutDataid')}}：</div>
                              <div class="content">
                                <bkbcs-input
                                  style="width: 80px;"
                                  type="number"
                                  :min="0"
                                  :placeholder="$t('generic.placeholder.input')"
                                  :value.sync="curCrdInstance.selector.std_data_id"
                                  :disabled="!curCrdInstance.selector.is_std_custom">
                                </bkbcs-input>
                                <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="curCrdInstance.selector.is_std_custom">
                                  {{$t('plugin.tools.custom')}}
                                  <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('plugin.tools.dataIdDesc') }"></i>
                                </bk-checkbox>
                              </div>
                            </div>
                          </div>
                        </div>

                        <div class="bk-form-content log-flex">
                          <div class="bk-form-inline-item">
                            <div class="log-form" style="width: 345px;">
                              <div class="label">{{$t('plugin.tools.path')}}：</div>
                              <div class="content log-path-wrapper" style="width: 223px;">
                                <textarea class="bk-form-textarea" v-model="curCrdInstance.selector.log_paths_str" style="width: 223px;" :placeholder="$t('plugin.tools.delimiter')"></textarea>

                                <bcs-popover placement="top" :delay="500">
                                  <i class="path-tip bcs-icon bcs-icon-question-circle"></i>
                                  <div slot="content">
                                    <p>1. 请填写文件的绝对路径，不支持目录</p>
                                    <p>2. 支持通配符，但通配符仅支持文件级别的</p>
                                    <p>有效的示例: /data/log/*/*.log; /data/test.log; /data/log/log.*</p>
                                    <p>无效的示例: /data/log/*; /data/log</p>
                                  </div>
                                </bcs-popover>
                              </div>
                            </div>
                          </div>

                          <div class="bk-form-inline-item">
                            <div class="log-form">
                              <div class="label" style="width: 110px;">{{$t('plugin.tools.fileDataid')}}：</div>
                              <div class="content">
                                <bkbcs-input
                                  style="width: 80px;"
                                  type="number"
                                  :min="0"
                                  :placeholder="$t('generic.placeholder.input')"
                                  :value.sync="curCrdInstance.selector.file_data_id"
                                  :disabled="!curCrdInstance.selector.is_file_custom">
                                </bkbcs-input>
                                <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="curCrdInstance.selector.is_file_custom">
                                  {{$t('plugin.tools.custom')}}
                                  <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('plugin.tools.dataIdDesc') }"></i>
                                </bk-checkbox>
                              </div>
                            </div>
                          </div>
                        </div>
                      </section>
                    </div>
                  </div>
                </div>
              </div>
            </section>

            <template>
              <div class="bk-form-item mt5" style="overflow: hidden;">
                <label class="bk-label">{{$t('logCollector.label.extraLabels')}}：</label>
              </div>

              <div class="log-wrapper">
                <bk-keyer
                  :key-list.sync="curLogLabelList"
                  :var-list="varList"
                  @change="updateLogLabels">
                </bk-keyer>
                <div class="mt10 mb10">
                  <bk-checkbox class="mr20" v-model="curCrdInstance.auto_add_pod_labels" name="cluster-classify-checkbox" true-value="true" false-value="false">
                    {{$t('plugin.tools.podLabels')}}
                  </bk-checkbox>
                </div>

                <div>
                  <bk-checkbox v-model="curCrdInstance.package_collection" name="cluster-classify-checkbox" true-value="true" false-value="false">
                    {{$t('plugin.tools.pack')}}
                    <i class="path-tip bcs-icon bcs-icon-question-circle" v-bk-tooltips.right="{ width: 400, content: $t('plugin.tools.logDesc') }"></i>
                  </bk-checkbox>
                </div>
              </div>
            </template>

            <div class="bk-form-item mt15">
              <bk-button type="primary" :loading="isDataSaveing" @click.stop.prevent="saveCrdInstance">{{curCrdInstance.id ? $t('generic.button.update') : $t('generic.button.create')}}</bk-button>
              <bk-button @click.stop.prevent="hideCrdInstanceSlider" :disabled="isDataSaveing">{{$t('generic.button.cancel')}}</bk-button>
            </div>
          </div>
        </div>
      </bk-sideslider>

      <bk-sideslider
        :quick-close="true"
        :is-show.sync="detailSliderConf.isShow"
        :title="detailSliderConf.title"
        :width="700">
        <div class="p30" slot="content">
          <p class="data-title">
            {{$t('generic.title.basicInfo')}}
          </p>
          <div class="biz-metadata-box vertical mb20">
            <div class="data-item">
              <p class="key">{{$t('generic.label.cluster1')}}：</p>
              <p class="value">{{clusterName || '--'}}</p>
            </div>
            <div class="data-item">
              <p class="key">{{$t('k8s.namespace')}}：</p>
              <p class="value">{{curCrdInstance.namespace === 'default' ? $t('plugin.tools.all') : curCrdInstance.namespace}}</p>
            </div>
            <div class="data-item">
              <p class="key">{{$t('plugin.tools.ruleName')}}：</p>
              <p class="value">{{curCrdInstance.name || '--'}}</p>
            </div>
          </div>
          <p class="data-title">
            {{$t('plugin.tools.dsInfo')}}
          </p>

          <div class="biz-metadata-box vertical mb0">
            <div class="data-item">
              <p class="key">{{$t('plugin.tools.dsType')}}：</p>
              <p class="value">{{logSource[curCrdInstance.log_source_type]}}</p>
            </div>
            <template v-if="curCrdInstance.log_source_type === 'selected_containers'">
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.appType')}}：</p>
                <p class="value">{{curCrdInstance.workload.type || '--'}}</p>
              </div>
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.appName')}}：</p>
                <p class="value">{{curCrdInstance.workload.name || '--'}}</p>
              </div>

              <div class="data-item">
                <p class="key">{{$t('plugin.tools.collectionPath')}}：</p>
                <div class="value">
                </div>
              </div>
            </template>
            <!-- <template v-else-if="curCrdInstance.log_source_type === 'all_containers'">
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.enabled')}}：</p>
                <p class="value">{{curCrdInstance.default_conf.stdout === 'true' ? $t('units.boolean.true') : $t('units.boolean.false')}}</p>
              </div>
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.stdoutDataid')}}：</p>
                <p class="value">{{curCrdInstance.default_conf.std_data_id || '--'}} ({{curCrdInstance.default_conf.is_std_custom ? 'generic.label.custom' : 'plugin.tools.default'}})</p>
              </div>
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.fileDataid')}}：</p>
                <p class="value">{{curCrdInstance.default_conf.file_data_id || '--'}} ({{curCrdInstance.default_conf.is_file_custom ? 'generic.label.custom' : 'plugin.tools.default'}})</p>
              </div>
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.path')}}：</p>
                <p class="value">{{curCrdInstance.default_conf.log_paths_str || '--'}}</p>
              </div>
            </template> -->
            <template v-else-if="curCrdInstance.log_source_type === 'selected_labels'">
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.enabled')}}：</p>
                <p class="value">{{curCrdInstance.selector.stdout === 'true' ? $t('units.boolean.true') : $t('units.boolean.false')}}</p>
              </div>
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.stdoutDataid')}}：</p>
                <p class="value">{{curCrdInstance.selector.std_data_id || '--'}} ({{curCrdInstance.selector.is_std_custom ? 'generic.label.custom' : 'plugin.tools.default'}})</p>
              </div>
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.fileDataid')}}：</p>
                <p class="value">{{curCrdInstance.selector.file_data_id || '--'}} ({{curCrdInstance.selector.is_file_custom ? 'generic.label.custom' : 'plugin.tools.default'}})</p>
              </div>
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.matchLables')}}：</p>
                <p class="value">
                  <ul class="key-list" v-if="Object.keys(curCrdInstance.selector.match_labels).length">
                    <li v-for="(label, index) of curCrdInstance.selector.match_labels_list" :key="index">
                      <span class="key f12 m0" style="cursor: default;">{{label.key || '--'}}</span>
                      <span class="value f12 m0" style="cursor: default;">{{label.value || '--'}}</span>
                    </li>
                  </ul>
                  <span v-else>--</span>
                </p>
              </div>
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.matchPattern')}}：</p>
                <p class="value">
                  <ul class="key-list" v-if="curCrdInstance.selector.match_expressions.length">
                    <li v-for="(expression, index) of curCrdInstance.selector.match_expressions" :key="index">
                      <span class="key f12 m0">{{expression.key || '--'}}</span>
                      <span class="value f12 m0">{{expression.operator || '--'}}</span>
                      <span class="value f12 m0" v-if="expression.values">{{expression.values || '--'}}</span>
                    </li>
                  </ul>
                  <span v-else>--</span>
                </p>
              </div>
              <div class="data-item">
                <p class="key">{{$t('plugin.tools.path')}}：</p>
                <p class="value">{{curCrdInstance.selector.log_paths_str || '--'}}</p>
              </div>
            </template>
          </div>

          <div class="biz-metadata-box mb0 mt5" style="border: none;" v-if="curCrdInstance.log_source_type === 'selected_containers'">
            <table class="bk-table bk-log-table">
              <thead>
                <tr>
                  <th style="width: 130px;">{{$t('plugin.tools.containerName')}}</th>
                  <th style="width: 200px;">{{$t('logCollector.label.collectorType.stdout')}}</th>
                  <th>{{$t('plugin.tools.path')}}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(containerConf, index) of curCrdInstance.workload.container_confs" :key="index">
                  <td>
                    {{containerConf.name || '--'}}
                  </td>
                  <td>
                    {{$t('plugin.tools.dataID')}}：{{containerConf.std_data_id || '--'}} ({{containerConf.is_std_custom ? $t('generic.label.custom') : $t('plugin.tools.default')}})<br />
                    {{$t('plugin.tools.enabled')}}：{{containerConf.stdout === 'true' ? $t('units.boolean.true') : $t('units.boolean.false')}}
                  </td>
                  <td>
                    <p>{{$t('plugin.tools.dataID')}}：{{containerConf.file_data_id || '--'}} ({{containerConf.is_file_custom ? $t('generic.label.custom') : $t('plugin.tools.default')}})</p>
                    <div class="log-key-value">
                      <div style="width: 38px;">{{$t('deploy.templateset.path')}}：</div>
                      <ul class="log-path-list" v-if="containerConf.log_paths.length">
                        <li v-for="path of containerConf.log_paths" :key="path" v-if="path">{{path}}</li>
                      </ul>
                      <span v-else>--</span>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <bk-table :data="[{}]" size="medium" v-else-if="curCrdInstance.log_source_type === 'all_containers'">
            <bk-table-column width="100" :label="$t('plugin.tools.containerName')">
              <template #default>{{ $t('plugin.tools.allContainers') }}</template>
            </bk-table-column>
            <bk-table-column :label="$t('logCollector.label.collectorType.stdout')">
              <template #default>
                <div class="flex flex-col">
                  <div>{{ `${$t('plugin.tools.dataID')}: ${curCrdInstance.default_conf.std_data_id || '--'}(${curCrdInstance.default_conf.is_std_custom ? $t('generic.label.custom') : $t('plugin.tools.default')})` }}</div>
                  <div>{{ `${$t('plugin.tools.enabled')}: ${curCrdInstance.default_conf.stdout === 'true' ? $t('units.boolean.true') : $t('units.boolean.false')}` }}</div>
                </div>
              </template>
            </bk-table-column>
            <bk-table-column :label="$t('plugin.tools.path')">
              <template #default>
                <div class="flex flex-col">
                  <div>{{ `${$t('plugin.tools.dataID')}: ${curCrdInstance.default_conf.file_data_id || '--'}(${curCrdInstance.default_conf.is_file_custom ? $t('generic.label.custom') : $t('plugin.tools.default')})` }}</div>
                  <div>{{ `${$t('deploy.templateset.path')}: ${curCrdInstance.default_conf.log_paths_str || '--'}` }}</div>
                </div>
              </template>
            </bk-table-column>
          </bk-table>
        </div>
      </bk-sideslider>
    </div>
  </div>
</template>

<script>
/* eslint-disable no-prototype-builtins */
/* eslint-disable @typescript-eslint/prefer-optional-chain */
/* eslint-disable @typescript-eslint/no-this-alias */
import bkExpression from './expression';

import { catchErrorHandler } from '@/common/util';
import bkKeyer from '@/components/keyer';
import Header from '@/components/layout/Header.vue';
import { useNamespace } from '@/views/cluster-manage/namespace/use-namespace';

export default {
  components: {
    bkKeyer,
    bkExpression,
    Header,
  },
  data() {
    return {
      isInitLoading: true,
      isPageLoading: false,
      curPageData: [],
      isDataSaveing: false,
      prmissions: {},
      pageConf: {
        count: 0,
        totalPage: 1,
        limit: 5,
        current: 1,
        show: true,
      },
      crdInstanceSlider: {
        title: this.$t('plugin.tools.create'),
        isShow: false,
      },
      searchParams: {
        clusterIndex: 0,
        namespace: 0,
        workload_type: 0,
        workload_name: '',
      },
      appTypes: [
        {
          id: 'Deployment',
          name: 'Deployment',
        },
        {
          id: 'DaemonSet',
          name: 'DaemonSet',
        },
        {
          id: 'Job',
          name: 'Job',
        },
        {
          id: 'StatefulSet',
          name: 'StatefulSet',
        },
        {
          id: 'GameStatefulSet',
          name: 'GameStatefulSet',
        },
      ],
      curLogLabelList1: [
        {
          key: '',
          value: '',
        },
      ],
      searchScope: '',
      nameSpaceList: [],
      curLabelList: [
        {
          key: '',
          value: '',
        },
      ],

      dbTypes: [
        {
          id: 'mysql',
          name: 'mysql',
        },
        {
          id: 'spider',
          name: 'spider',
        },
      ],

      logSource: {
        selected_containers: this.$t('plugin.tools.container'),
        selected_labels: this.$t('plugin.tools.label'),
        all_containers: this.$t('plugin.tools.allContainers'),
      },
      curCrdInstance: {
        // 'crd_kind': 'BcsLogConfig',
        // 'cluster_id': '',
        namespace: '',
        config_type: 'custom',
        log_source_type: 'selected_containers',
        app_id: '',
        labels: {},
        auto_add_pod_labels: 'false',
        package_collection: 'false',
        default_conf: {
          stdout: 'true',
          is_std_custom: false,
          is_file_custom: false,
          std_data_id: '',
          file_data_id: '',
          log_paths: [],
          log_paths_str: '',
        },
        workload: {
          name: '',
          type: '',
          container_confs: [
            {
              name: '',
              std_data_id: '',
              file_data_id: '',
              stdout: 'true',
              log_paths: [],
              log_paths_str: '',
              is_std_custom: false,
              is_file_custom: false,
            },
          ],
        },
        selector: {
          std_data_id: '',
          file_data_id: '',
          stdout: 'true',
          log_paths: [],
          log_paths_str: '',
          match_labels: {},
          is_std_custom: false,
          is_file_custom: false,
          match_labels_list: [
            {
              key: '',
              value: '',
            },
          ],
          match_expressions: [],
        },
      },
      detailSliderConf: {
        isShow: false,
        title: '',
      },
      crdKind: 'BcsLogConfig',
      defaultStdDataId: 0,
      defaultFileDataId: 0,
    };
  },
  computed: {
    isEn() {
      return this.$store.state.isEn;
    },
    varList() {
      return this.$store.state.variable.varList;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    crdInstanceList() {
      const list = Object.assign([], this.$store.state.crdInstanceList);
      const results = list.map((item) => {
        const data = {
          ...item,
          ...item.crd_data,
        };
        if (data.log_source_type === 'selected_containers') {
          data.crd_data.workload.isExpand = false;
          data.crd_data.workload.container_confs.forEach((conf) => {
            if (!conf.log_paths) {
              conf.log_paths = [];
            }
          });
        } else if (data.log_source_type === 'selected_labels') {
          // eslint-disable-next-line no-prototype-builtins
          if (!data.selector.hasOwnProperty('match_labels')) {
            data.selector.match_labels = {};
          }
          // eslint-disable-next-line no-prototype-builtins
          if (!data.selector.hasOwnProperty('match_expressions')) {
            data.selector.match_expressions = [];
          }
        }
        return data;
      });
      return results;
    },
    clusterList() {
      return this.$store.state.cluster.clusterList;
    },
    curProject() {
      return this.$store.state.curProject;
    },
    clusterId() {
      return this.$route.params.clusterId;
    },
    clusterName() {
      const cluster = this.clusterList.find(item => item.cluster_id === this.clusterId);
      return cluster ? cluster.name : '';
    },
    searchScopeList() {
      const { clusterList } = this.$store.state.cluster;
      let results = [];
      if (clusterList.length) {
        results = [];
        clusterList.forEach((item) => {
          results.push({
            id: item.cluster_id,
            name: item.name,
          });
        });
      }

      return results;
    },
    curCrdInstanceName() {
      // 编辑状态，使用bcslog_name
      if (this.curCrdInstance.id && this.curCrdInstance.bcslog_name) {
        return this.curCrdInstance.bcslog_name;
      } if (this.curCrdInstance.log_source_type === 'all_containers') {
        return 'default-std-log';
      }
      const app = this.curCrdInstance.workload;
      if (app.type && app.name) {
        return `${app.type}-${app.name}-log`.toLowerCase();
      }
      return 'log';
    },
    curLogLabelList() {
      const keyList = [];
      const labels = this.curCrdInstance.labels || {};
      for (const key in labels) {
        keyList.push({
          key,
          value: labels[key],
        });
      }
      if (!keyList.length) {
        keyList.push({
          key: '',
          value: '',
        });
      }
      return keyList;
    },
  },
  watch: {
    crdInstanceList() {
      this.curPageData = this.getDataByPage(this.pageConf.current);
    },
    curPageData() {
      this.curPageData.forEach((item) => {
        if (item.clb_status && item.clb_status !== 'Running') {
          this.getCrdInstanceStatus(item);
        }
      });
    },
    curCrdInstance: {
      deep: true,
      handler() {
        if (this.curCrdInstance.log_source_type === 'selected_containers') {
          const containerConfs = this.curCrdInstance.workload.container_confs;
          containerConfs.forEach((conf) => {
            if (!conf.is_std_custom) {
              conf.std_data_id = this.defaultStdDataId;
            }
            if (!conf.is_file_custom) {
              conf.file_data_id = this.defaultFileDataId;
            }
          });
        } else if (this.curCrdInstance.log_source_type === 'all_containers') {
          const defaultConf = this.curCrdInstance.default_conf;
          if (!defaultConf.is_std_custom) {
            defaultConf.std_data_id = this.defaultStdDataId;
          }
          if (!defaultConf.is_file_custom) {
            defaultConf.file_data_id = this.defaultFileDataId;
          }
        } else if (this.curCrdInstance.log_source_type === 'selected_labels') {
          const { selector } = this.curCrdInstance;
          if (!selector.is_std_custom) {
            selector.std_data_id = this.defaultStdDataId;
          }
          if (!selector.is_file_custom) {
            selector.file_data_id = this.defaultFileDataId;
          }
        }
      },
    },
  },
  created() {
    this.getCrdInstanceList();
    this.getNameSpaceList();
    this.getLogInfo();
  },
  methods: {
    goBack() {
      this.$router.push({
        name: 'logCrdcontroller',
        params: {
          projectId: this.projectId,
        },
      });
    },

    /**
             * 搜索列表
             */
    handleSearch() {
      this.pageConf.current = 1;
      this.isPageLoading = true;

      this.getCrdInstanceList();
    },

    /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
    handlePageSizeChange(pageSize) {
      this.pageConf.limit = pageSize;
      this.pageConf.current = 1;
      this.initPageConf();
      this.handlePageChange();
    },

    /**
             * 新建
             */
    createLoadBlance() {
      this.curCrdInstance = {
        // 'crd_kind': 'BcsLog',
        // 'cluster_id': this.clusterId,
        namespace: '',
        config_type: 'custom',
        log_source_type: 'selected_containers',
        app_id: this.curProject.cc_app_id,

        labels: {},
        auto_add_pod_labels: 'false',
        package_collection: 'false',
        default_conf: {
          stdout: 'true',
          std_data_id: this.defaultStdDataId,
          file_data_id: this.defaultFileDataId,
          log_paths: [],
          log_paths_str: '',
          is_std_custom: false,
          is_file_custom: false,
        },
        workload: {
          name: '',
          type: '',
          container_confs: [
            {
              name: '',
              std_data_id: this.defaultStdDataId,
              file_data_id: this.defaultFileDataId,
              stdout: 'true',
              log_paths: '',
              log_paths_str: '',
              is_std_custom: false,
              is_file_custom: false,
            },
          ],
        },
        selector: {
          std_data_id: this.defaultStdDataId,
          file_data_id: this.defaultFileDataId,
          stdout: 'true',
          log_paths: [],
          log_paths_str: '',
          match_labels: {},
          match_labels_list: [
            {
              key: '',
              value: '',
            },
          ],
          match_expressions: [],
          match_expressions_list: [],
        },
      };

      this.crdInstanceSlider.title = this.$t('plugin.tools.create');
      this.crdInstanceSlider.isShow = true;
    },

    updateLogLabels(list, data) {
      this.curCrdInstance.labels = data;
    },

    updateMatchLabels(list, data) {
      const obj = {};
      for (const key in data) {
        if (key) {
          obj[key] = data[key];
        }
      }
      this.curCrdInstance.selector.match_labels = obj;
    },

    updateExpressions(list) {
      this.curCrdInstance.selector.match_expressions = list;
    },

    /**
             * 编辑
             * @param  {object} crdInstance crdInstance
             * @param  {number} index 索引
             */
    async editCrdInstance(crdInstance, isReadonly) {
      if (this.isDetailLoading) {
        return false;
      }
      this.isDetailLoading = true;

      try {
        const { projectId } = this;
        const { clusterId } = this;
        const crdId = crdInstance.id;
        const { crdKind } = this;
        const res = await this.$store.dispatch('crdcontroller/getLogCrdInstanceDetail', {
          projectId,
          clusterId,
          crdId,
          crdKind,
        });
        const data = {
          ...res.data,
          ...res.data.crd_data,
        };
        data.id = res.data.id;
        data.name = res.data.name;
        if (data.log_source_type === 'selected_containers') {
          data.workload.container_confs.forEach((conf) => {
            if (!conf.log_paths) {
              conf.log_paths = [];
            }
            conf.log_paths_str = conf.log_paths.join(';');
            // 是否自定义
            conf.is_std_custom = String(conf.std_data_id) !== String(this.defaultStdDataId);
            conf.is_file_custom = String(conf.file_data_id) !== String(this.defaultFileDataId);
          });
        } else if (data.log_source_type === 'all_containers') {
          // 是否自定义
          data.default_conf.is_std_custom = String(data.default_conf.std_data_id) !== String(this.defaultStdDataId);
          data.default_conf.is_file_custom = String(data.default_conf.file_data_id) !== String(this.defaultFileDataId);
          if (data.default_conf.log_paths) {
            data.default_conf.log_paths_str = data.default_conf.log_paths.join(';');
          } else {
            data.default_conf.log_paths = [];
            data.default_conf.log_paths_str = '';
          }
        } else if (data.log_source_type === 'selected_labels') {
          const { selector } = data;
          // 是否自定义
          selector.is_std_custom = String(selector.std_data_id) !== String(this.defaultStdDataId);
          selector.is_file_custom = String(selector.file_data_id) !== String(this.defaultFileDataId);
          selector.log_paths_str = selector.log_paths.join(';');

          // 匹配标签(
          selector.match_labels_list = [];
          if (selector.match_labels && Object.keys(selector.match_labels).length) {
            for (const key in selector.match_labels) {
              selector.match_labels_list.push({
                key,
                value: selector.match_labels[key],
              });
            }
          } else {
            selector.match_labels = {};
            selector.match_labels_list.push({
              key: '',
              value: '',
            });
          }

          // 匹配表达式
          if (selector.match_expressions && selector.match_expressions.length) {
            selector.match_expressions_list = selector.match_expressions.map((item) => {
              if (!item.hasOwnProperty('values')) {
                item.values = '';
              }
              return item;
            });
          } else {
            selector.match_expressions = [];
            selector.match_expressions_list = [];
          }
        }

        if (!data.hasOwnProperty('package_collection')) {
          data.package_collection = 'false';
        }

        if (!data.hasOwnProperty('auto_add_pod_labels')) {
          data.auto_add_pod_labels = 'false';
        }

        this.curCrdInstance = data;

        if (isReadonly) {
          this.detailSliderConf.title = `${this.curCrdInstance.name}`;
          this.detailSliderConf.isShow = true;
        } else {
          this.crdInstanceSlider.title = this.$t('plugin.tools.editRule');
          this.crdInstanceSlider.isShow = true;
        }
      } catch (e) {
        // catchErrorHandler(e, this)
      } finally {
        this.isDetailLoading = false;
      }
    },

    /**
             * 删除
             * @param  {object} crdInstance crdInstance
             * @param  {number} index 索引
             */
    async removeCrdInstance(crdInstance) {
      const self = this;
      const { projectId } = this;
      const { clusterId } = this;
      const { crdKind } = this;
      const crdId = crdInstance.id;

      this.$bkInfo({
        title: this.$t('generic.title.confirmDelete'),
        clsName: 'biz-remove-dialog',
        content: this.$createElement('p', {
          class: 'biz-confirm-desc',
        }, `${this.$t('plugin.tools.confirmDelete')}【${crdInstance.name}】？`),
        async confirmFn() {
          self.isPageLoading = true;
          try {
            await self.$store.dispatch('crdcontroller/deleteCrdInstance', { projectId, clusterId, crdKind, crdId });
            self.$bkMessage({
              theme: 'success',
              message: self.$t('generic.msg.success.delete'),
            });
            self.getCrdInstanceList();
          } catch (e) {
            catchErrorHandler(e, this);
          } finally {
            self.isPageLoading = false;
          }
        },
      });
    },

    /**
             * 获取
             * @param  {number} crdInstanceId id
             * @return {object} crdInstance crdInstance
             */
    getCrdInstanceById(crdInstanceId) {
      return this.crdInstanceList.find(item => item.id === crdInstanceId);
    },

    /**
             * 初始化分页配置
             */
    initPageConf() {
      const total = this.crdInstanceList.length;
      this.pageConf.count = total;
      this.pageConf.totalPage = Math.ceil(total / this.pageConf.limit);
      if (this.pageConf.current > this.pageConf.totalPage) {
        this.pageConf.current = this.pageConf.totalPage;
      }
    },

    /**
             * 重新加载当前页
             */
    reloadCurPage() {
      this.initPageConf();
      if (this.pageConf.current > this.pageConf.totalPage) {
        this.pageConf.current = this.pageConf.totalPage;
      }
      this.curPageData = this.getDataByPage(this.pageConf.current);
    },

    /**
             * 获取页数据
             * @param  {number} page 页
             * @return {object} data lb
             */
    getDataByPage(page) {
      // 如果没有page，重置
      if (!page) {
        // eslint-disable-next-line no-multi-assign
        this.pageConf.current = page = 1;
      }
      let startIndex = (page - 1) * this.pageConf.limit;
      let endIndex = page * this.pageConf.limit;
      // this.isPageLoading = true
      if (startIndex < 0) {
        startIndex = 0;
      }
      if (endIndex > this.crdInstanceList.length) {
        endIndex = this.crdInstanceList.length;
      }
      this.isPageLoading = false;
      return this.crdInstanceList.slice(startIndex, endIndex);
    },

    /**
             * 分页改变回调
             * @param  {number} page 页
             */
    handlePageChange(page = 1) {
      this.isPageLoading = true;
      this.pageConf.current = page;
      const data = this.getDataByPage(page);
      this.curPageData = JSON.parse(JSON.stringify(data));
    },

    /**
             * 隐藏lb侧面板
             */
    hideCrdInstanceSlider() {
      this.crdInstanceSlider.isShow = false;
    },

    /**
             * 加载数据
             */
    async getCrdInstanceList() {
      try {
        const { projectId } = this;
        const { clusterId } = this;
        const { crdKind } = this;

        const params = {
          // cluster_id: this.clusterId
          // crd_kind: this.crdKind
        };

        if (this.searchParams.namespace) {
          params.namespace = this.searchParams.namespace;
        }

        if (this.searchParams.workload_type) {
          params.workload_type = this.searchParams.workload_type;
        }

        if (this.searchParams.workload_name) {
          params.workload_name = this.searchParams.workload_name;
        }

        this.isPageLoading = true;
        await this.$store.dispatch('getBcsCrdsList', {
          projectId,
          clusterId,
          crdKind,
          params,
        });

        this.initPageConf();
        this.curPageData = this.getDataByPage(this.pageConf.current);
      } catch (e) {
        catchErrorHandler(e, this);
      } finally {
        // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
        setTimeout(() => {
          this.isPageLoading = false;
          this.isInitLoading = false;
        }, 500);
      }
    },

    /**
             * 获取命名空间列表
             */
    async getNameSpaceList() {
      try {
        const { clusterId } = this;
        const { getNamespaceData } = useNamespace();

        const res = await getNamespaceData({
          $clusterId: clusterId,
        });
        const list = res;
        list.forEach((item) => {
          item.isSelected = false;
        });
        this.nameSpaceList.splice(0, this.nameSpaceList.length, ...list);
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },

    /**
             * 获取日志信息
             */
    async getLogInfo() {
      try {
        const { projectId } = this;
        const res = await this.$store.dispatch('getLogPlans', projectId);
        this.defaultStdDataId = res.data.std_data_id;
        this.defaultFileDataId = res.data.file_data_id;
      } catch (e) {
        catchErrorHandler(e, this);
      }
    },

    /**
             * 选择/取消选择命名空间
             * @param  {object} nameSpace 命名空间
             * @param  {number} index 索引
             */
    toggleSelected(nameSpace) {
      nameSpace.isSelected = !nameSpace.isSelected;
      this.nameSpaceList = JSON.parse(JSON.stringify(this.nameSpaceList));
    },

    /**
             * 检查提交的数据
             * @return {boolean} true/false 是否合法
             */
    checkData(params) {
      if (params.log_source_type !== 'all_containers' && !params.namespace) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('dashboard.ns.validate.emptyNs'),
          delay: 5000,
        });
        return false;
      }

      if (params.log_source_type === 'selected_containers') {
        if (!params.workload.type) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('plugin.tools.selectAppType'),
          });
          return false;
        }

        if (!params.workload.name) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('deploy.templateset.validate.name'),
          });
          return false;
        }

        try {
          const reg = new RegExp(params.workload.name);
          console.log(reg);
        } catch (e) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('plugin.tools.appNameError'),
          });
          return false;
        }

        for (const conf of params.workload.container_confs) {
          if (!conf.name) {
            this.$bkMessage({
              theme: 'error',
              message: this.$t('plugin.tools.enterContainerName'),
            });
            return false;
          }

          if (!conf.std_data_id) {
            this.$bkMessage({
              theme: 'error',
              message: this.$t('plugin.tools.stdinDataID'),
            });
            return false;
          }

          if (!conf.file_data_id) {
            this.$bkMessage({
              theme: 'error',
              message: this.$t('plugin.tools.fileDataID'),
            });
            return false;
          }
        }
      } else if (params.log_source_type === 'all_containers') {
        if (!params.default_conf.std_data_id) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('plugin.tools.stdinDataID'),
          });
          return false;
        }
        if (!params.default_conf.file_data_id) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('plugin.tools.fileDataID'),
          });
          return false;
        }
      } else if (params.log_source_type === 'selected_labels') {
        console.log('params', params);
        if (!params.selector.hasOwnProperty('match_labels') && !params.selector.hasOwnProperty('match_expressions')) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('plugin.tools.emptyTips'),
          });
          return false;
        }
        if (!params.selector.std_data_id) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('plugin.tools.stdinDataID'),
          });
          return false;
        }

        if (!params.selector.file_data_id) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('plugin.tools.fileDataID'),
          });
          return false;
        }

        if (!params.selector.log_paths.length) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('plugin.tools.filePath'),
          });
          return false;
        }
      }

      return true;
    },

    showCrdInstanceDetail(data) {
      data.labels = [];
      for (const key in data.pod_selector) {
        data.labels.push({
          key,
          value: data.pod_selector[key],
        });
      }
      this.curCrdInstance = data;

      this.detailSliderConf.title = `${data.name}`;
      this.detailSliderConf.isShow = true;
    },

    /**
             * 格式化数据，符合接口需要的格式
             */
    formatData() {
      const params = JSON.parse(JSON.stringify(this.curCrdInstance));
      // 附加日志标签
      const { labels } = params;
      params.labels = {};
      if (params.labels) {
        for (const key in labels) {
          if (key) {
            params.labels[key] = labels[key];
          }
        }
      }

      if (params.log_source_type === 'selected_containers') {
        params.workload.container_confs.forEach((conf) => {
          const paths = conf.log_paths_str.split(/[;|\n]/).filter(item => !!item.trim())
            .map(item => item.trim());

          if (paths.length) {
            conf.log_paths = paths;
          } else {
            delete conf.log_paths;
          }

          // 接口接受'true'、'false'字符类型
          conf.stdout = String(conf.stdout);
        });
        delete params.std_data_id;
        delete params.is_std_custom;
        delete params.stdout;
        delete params.selector;
        delete params.default_conf;
      } else if (params.log_source_type === 'all_containers') {
        if (!params.id) {
          if (!params.namespace) {
            params.config_type = 'default';
            delete params.namespace;
          } else {
            params.config_type = 'custom';
          }
        }

        const paths = params.default_conf.log_paths_str.split(/[;|\n]/).filter(item => !!item.trim())
          .map(item => item.trim());

        if (paths.length) {
          params.default_conf.log_paths = paths;
        } else {
          delete params.default_conf.log_paths;
        }
        delete params.workload;
        delete params.selector;
        delete params.default_conf.log_paths_str;
        delete params.default_conf.is_std_custom;
        delete params.default_conf.is_file_custom;
      } else if (params.log_source_type === 'selected_labels') {
        const paths = params.selector.log_paths_str.split(/[;|\n]/).filter(item => !!item.trim())
          .map(item => item.trim());

        params.selector.log_paths = paths;
        delete params.std_data_id;
        delete params.is_std_custom;
        delete params.stdout;
        delete params.workload;
        delete params.default_conf;
        // selector
        delete params.selector.match_expressions_list;
        delete params.selector.match_labels_list;
        delete params.selector.log_paths_str;

        if (Object.keys(params.selector.match_labels).length === 0) {
          delete params.selector.match_labels;
        }

        if (params.selector.match_expressions.length === 0) {
          delete params.selector.match_expressions;
        } else {
          params.selector.match_expressions.forEach((item) => {
            if (['Exists', 'DoesNotExist'].includes(item.operator)) {
              delete item.values;
            }
          });
        }
      }

      return params;
    },

    /**
             * 保存新建的
             */
    async createCrdInstance(params) {
      const { projectId } = this;
      const { clusterId } = this;
      const { crdKind } = this;
      const data = params;
      this.isDataSaveing = true;

      try {
        await this.$store.dispatch('crdcontroller/addCrdInstance', { projectId, clusterId, crdKind, data });

        this.$bkMessage({
          theme: 'success',
          message: this.$t('plugin.tools.success'),
        });
        this.getCrdInstanceList();
        this.hideCrdInstanceSlider();
      } catch (e) {
        // catchErrorHandler(e, this)
      } finally {
        this.isDataSaveing = false;
      }
    },

    /**
             * 保存更新的
             */
    async updateCrdInstance(params) {
      const { projectId } = this;
      const { clusterId } = this;
      const { crdKind } = this;
      const data = params;
      this.isDataSaveing = true;
      delete data.cluster_id;
      // data.crd_kind = this.crdKind
      try {
        await this.$store.dispatch('crdcontroller/updateCrdInstance', { projectId, clusterId, crdKind, data });

        this.$bkMessage({
          theme: 'success',
          message: this.$t('plugin.tools.upgraded'),
        });

        this.hideCrdInstanceSlider();
        // sideslider和loading层有样式冲突
        setTimeout(() => {
          this.getCrdInstanceList();
        }, 500);
      } catch (e) {
        // catchErrorHandler(e, this)
      } finally {
        this.isDataSaveing = false;
      }
    },

    /**
             * 保存
             */
    saveCrdInstance() {
      const params = this.formatData();
      if (this.checkData(params) && !this.isDataSaveing) {
        if (this.curCrdInstance.id > 0) {
          this.updateCrdInstance(params);
        } else {
          this.createCrdInstance(params);
        }
      }
    },

    handleNamespaceSelect(index, data) {
      this.curCrdInstance.namespace = data.name;
    },

    changeLabels(labels) {
      // this.curCrdInstance.pod_selector = data
      this.curCrdInstance.labels = labels;
    },

    addContainerConf() {
      this.curCrdInstance.workload.container_confs.push({
        name: '',
        std_data_id: this.defaultStdDataId,
        file_data_id: this.defaultFileDataId,
        stdout: 'true',
        log_paths: [],
        log_paths_str: '',
        is_file_custom: false,
        is_std_custom: false,

      });
    },

    removeContainerConf(data, index) {
      this.curCrdInstance.workload.container_confs.splice(index, 1);
    },

    toggleExpand(crdInstance) {
      crdInstance.crd_data.workload.isExpand = !crdInstance.crd_data.workload.isExpand;
    },
  },
};
</script>

<style scoped>
    @import './log_list.css';
</style>
