<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-crd-instance-title">
                <a href="javascript:void(0);" class="bcs-icon bcs-icon-arrows-left back" @click="goBack"></a>
                {{$t('日志采集规则')}}
                <span class="biz-tip ml10">({{$t('集群名称')}}：{{clusterName}})</span>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
            <app-exception
                v-if="exceptionCode && !isInitLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>

            <template v-if="!exceptionCode && !isInitLoading">
                <div class="biz-panel-header">
                    <div class="left">
                        <bk-button type="primary" @click.stop.prevent="createLoadBlance">
                            <i class="bcs-icon bcs-icon-plus" style="top: -1px;"></i>
                            <span>{{$t('新建规则')}}</span>
                        </bk-button>
                    </div>
                    <div class="right search-wrapper">
                        <div class="left">
                            <bk-selector
                                style="width: 135px;"
                                :searchable="true"
                                :placeholder="$t('命名空间')"
                                :selected.sync="searchParams.namespace"
                                :list="nameSpaceList"
                                :setting-key="'name'"
                                :display-key="'name'"
                                :allow-clear="true"
                                @clear="clusterClear">
                            </bk-selector>
                        </div>
                        <div class="left">
                            <bk-selector
                                style="width: 135px;"
                                :placeholder="$t('应用类型')"
                                :selected.sync="searchParams.workload_type"
                                :list="appTypes"
                                :setting-key="'id'"
                                :display-key="'name'"
                                :allow-clear="true"
                                @clear="clusterClear">
                            </bk-selector>
                        </div>
                        <div class="left">
                            <bkbcs-input
                                style="width: 135px;"
                                :placeholder="$t('应用名')"
                                :value.sync="searchParams.workload_name">
                            </bkbcs-input>
                        </div>
                        <div class="left">
                            <bk-button type="primary" :title="$t('查询')" icon="search" @click="handleSearch">
                                {{$t('查询')}}
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
                            <bk-table-column :label="$t('名称')" prop="name" :show-overflow-tooltip="true" min-width="250">
                                <template slot-scope="{ row }">
                                    <a href="javascript: void(0)" class="bk-text-button biz-table-title biz-resource-title" @click.stop.prevent="editCrdInstance(row, true)">{{row.name || '--'}}</a>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="`${$t('集群')} / ${$t('命名空间')}`" min-width="220">
                                <template slot-scope="{ row }">
                                    <p>{{$t('所属集群')}}：{{clusterName}}</p>
                                    <p>{{$t('命名空间')}}：{{(row.namespace === 'default' && row.config_type === 'default' && row.log_source_type === 'all_containers') ? $t('所有') : row.namespace}}</p>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('日志源')" min-width="100">
                                <template slot-scope="{ row }">
                                    {{logSource[row.log_source_type]}}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('选择器')" min-width="250">
                                <template slot-scope="{ row }">
                                    <template v-if="row.log_source_type === 'selected_containers'">
                                        <p>{{$t('类型')}}：{{row.crd_data.workload.type}}</p>
                                        <p>{{$t('名称')}}：{{row.crd_data.workload.name || '--'}}</p>
                                    </template>
                                    <template v-else-if="row.log_source_type === 'selected_labels'">
                                        <div class="data-item mt5" v-if="Object.keys(row.selector.match_labels).length">
                                            <p class="key mb5">{{$t('匹配标签')}}：</p>
                                            <p class="value">
                                                <ul class="key-list">
                                                    <li class="mb5" v-for="(value, key, labelIndex) in row.selector.match_labels" :key="labelIndex" v-if="labelIndex < 2">
                                                        <span class="key f12 m0" style="cursor: default;">{{key || '--'}}</span>
                                                        <span class="value f12 m0" style="cursor: default;">{{value || '--'}}</span>
                                                    </li>
                                                </ul>
                                            </p>
                                        </div>
                                        <div class="data-item" v-if="row.selector.match_expressions.length">
                                            <p class="key mb5">{{$t('匹配表达式')}}：</p>
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
                            <bk-table-column :label="$t('日志信息')" min-width="200">
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
                            <bk-table-column :label="$t('操作记录')" width="260">
                                <template slot-scope="{ row }">
                                    <p>{{$t('更新人')}}：{{row.operator || '--'}}</p>
                                    <p>{{$t('更新时间')}}：{{row.updated || '--'}}</p>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('状态')" min-width="100">
                                <template slot-scope="{ row }">
                                    {{row.bind_success ? $t('正常') : $t('异常')}}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('操作')" width="150">
                                <template slot-scope="{ row }">
                                    <a href="javascript:void(0);" class="bk-text-button" @click="editCrdInstance(row)">{{$t('更新')}}</a>
                                    <a href="javascript:void(0);" class="bk-text-button" @click="removeCrdInstance(row)">{{$t('删除')}}</a>
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
                                    <label class="bk-label">{{$t('名称')}}：</label>
                                    <div class="bk-form-content">
                                        <bkbcs-input
                                            :placeholder="$t('请输入')"
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
                                    <label class="bk-label">{{$t('所属集群')}}：</label>
                                    <div class="bk-form-content">
                                        <bk-selector
                                            :placeholder="$t('请输入')"
                                            :setting-key="'cluster_id'"
                                            :display-key="'name'"
                                            :selected.sync="clusterId"
                                            :list="clusterList"
                                            :disabled="true">
                                        </bk-selector>
                                    </div>
                                </div>

                                <div :class="['bk-form-inline-item', { 'is-required': curCrdInstance.log_source_type !== 'all_containers' }]" style="width: 352px; margin-left: 25px;">
                                    <label class="bk-label">{{$t('命名空间')}}：<span class="biz-tip fn" v-if="curCrdInstance.log_source_type === 'all_containers'">({{$t('不选择表示所有命名空间')}})</span></label>
                                    <div class="bk-form-content">
                                        <bkbcs-input
                                            v-if="curCrdInstance.id && curCrdInstance.namespace === 'default' && curCrdInstance.config_type === 'default' && curCrdInstance.log_source_type === 'all_containers'"
                                            :value="$t('所有')"
                                            disabled />
                                        <bk-selector
                                            v-else
                                            :searchable="true"
                                            :placeholder="$t('请选择')"
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
                            <label class="bk-label">{{$t('日志源')}}：</label>
                            <div class="bk-form-content">
                                <bk-radio-group v-model="curCrdInstance.log_source_type">
                                    <bk-radio :value="'selected_containers'" :disabled="curCrdInstance.id">{{$t('指定容器')}}</bk-radio>
                                    <bk-radio :value="'all_containers'" :disabled="curCrdInstance.id">{{$t('所有容器')}}</bk-radio>
                                    <bk-radio :value="'selected_labels'" :disabled="curCrdInstance.id">{{$t('指定标签')}}</bk-radio>
                                </bk-radio-group>
                            </div>
                        </div>

                        <section class="log-wrapper" v-if="curCrdInstance.log_source_type === 'all_containers'">
                            <div class="bk-form-content log-flex mb15">
                                <div class="bk-form-inline-item">
                                    <div class="log-form">
                                        <div class="label">{{$t('标准输出')}}：</div>
                                        <div class="content" style="width: 223px; margin-right: 32px;">
                                            <bk-checkbox name="cluster-classify-checkbox" v-model="curCrdInstance.default_conf.stdout" true-value="true" false-value="false">
                                                {{$t('是否采集')}}
                                                <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.right="$t('如果不勾选，将不采集此容器的标准输出')"></i>
                                            </bk-checkbox>
                                        </div>
                                    </div>
                                </div>

                                <div class="bk-form-inline-item">
                                    <div class="log-form">
                                        <div class="label" style="width: 108px;">{{$t('标准采集ID')}}：</div>
                                        <div class="content">
                                            <bkbcs-input
                                                style="width: 80px;"
                                                type="number"
                                                :min="0"
                                                :placeholder="$t('请输入')"
                                                :value.sync="curCrdInstance.default_conf.std_data_id"
                                                :disabled="!curCrdInstance.default_conf.is_std_custom">
                                            </bkbcs-input>
                                            <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="curCrdInstance.default_conf.is_std_custom">
                                                {{$t('是否自定义')}}
                                                <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('采集id对应数据平台的data id，不勾选平台将分配默认的data id进行日志的清洗和入库。如果有特别的清洗和计算要求，用户可以填写自己的data id') }"></i>
                                            </bk-checkbox>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="bk-form-content log-flex">
                                <div class="bk-form-inline-item">
                                    <div class="log-form" style="width: 345px;">
                                        <div class="label">{{$t('文件路径')}}：</div>
                                        <div class="content log-path-wrapper" style="width: 223px;">
                                            <textarea class="bk-form-textarea" v-model="curCrdInstance.default_conf.log_paths_str" style="width: 223px;" :placeholder="$t('多个以;分隔')"></textarea>

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
                                        <div class="label" style="width: 110px;">{{$t('文件采集ID')}}：</div>
                                        <div class="content">
                                            <bkbcs-input
                                                style="width: 80px;"
                                                type="number"
                                                :min="0"
                                                :placeholder="$t('请输入')"
                                                :value.sync="curCrdInstance.default_conf.file_data_id"
                                                :disabled="!curCrdInstance.default_conf.is_file_custom">
                                            </bkbcs-input>
                                            <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="curCrdInstance.default_conf.is_file_custom">
                                                {{$t('是否自定义')}}
                                                <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('采集id对应数据平台的data id，不勾选平台将分配默认的data id进行日志的清洗和入库。如果有特别的清洗和计算要求，用户可以填写自己的data id') }"></i>
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
                                            <div class="label">{{$t('应用类型')}}：</div>
                                            <div class="content">
                                                <bk-selector
                                                    style="width: 330px;"
                                                    :placeholder="$t('应用类型')"
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
                                            <div class="label">{{$t('应用名称')}}：</div>
                                            <div class="content">
                                                <bkbcs-input
                                                    style="width: 333px;"
                                                    :placeholder="$t('请输入应用名称，支持正则匹配')"
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
                                        <div class="label">{{$t('采集路径')}}：</div>
                                        <div class="content">
                                            <section class="log-inner-wrapper mb10" v-for="(containerConf, index) of curCrdInstance.workload.container_confs" :key="index">
                                                <div class="bk-form-content log-flex mb10">
                                                    <div class="bk-form-inline-item">
                                                        <div class="log-form">
                                                            <div class="label">{{$t('容器名')}}：</div>
                                                            <div class="content">
                                                                <bkbcs-input
                                                                    style="width: 223px;"
                                                                    :placeholder="$t('请输入')"
                                                                    :value.sync="containerConf.name">
                                                                </bkbcs-input>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>

                                                <div class="bk-form-content log-flex mb10">
                                                    <div class="bk-form-inline-item">
                                                        <div class="log-form">
                                                            <div class="label">{{$t('标准输出')}}：</div>
                                                            <div class="content" style="width: 223px; margin-right: 32px;">
                                                                <bk-checkbox name="cluster-classify-checkbox" v-model="containerConf.stdout" true-value="true" false-value="false">
                                                                    {{$t('是否采集')}}
                                                                    <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.right="$t('如果不勾选，将不采集此容器的标准输出')"></i>
                                                                </bk-checkbox>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-inline-item">
                                                        <div class="log-form">
                                                            <div class="label" style="width: 108px;">{{$t('标准采集ID')}}：</div>
                                                            <div class="content">
                                                                <bkbcs-input
                                                                    style="width: 80px;"
                                                                    type="number"
                                                                    :min="0"
                                                                    :placeholder="$t('请输入')"
                                                                    :value.sync="containerConf.std_data_id"
                                                                    :disabled="!containerConf.is_std_custom">
                                                                </bkbcs-input>
                                                                <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="containerConf.is_std_custom">
                                                                    {{$t('是否自定义')}}
                                                                    <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('采集id对应数据平台的data id，不勾选平台将分配默认的data id进行日志的清洗和入库。如果有特别的清洗和计算要求，用户可以填写自己的data id') }"></i>
                                                                </bk-checkbox>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>

                                                <div class="bk-form-content log-flex">
                                                    <div class="bk-form-inline-item">
                                                        <div class="log-form" style="width: 345px;">
                                                            <div class="label">{{$t('文件路径')}}：</div>
                                                            <div class="content log-path-wrapper" style="width: 223px;">
                                                                <textarea class="bk-form-textarea" v-model="containerConf.log_paths_str" style="width: 223px;" :placeholder="$t('多个以;分隔')"></textarea>

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
                                                            <div class="label" style="width: 110px;">{{$t('文件采集ID')}}：</div>
                                                            <div class="content">
                                                                <bkbcs-input
                                                                    style="width: 80px;"
                                                                    type="number"
                                                                    :min="0"
                                                                    :placeholder="$t('请输入')"
                                                                    :value.sync="containerConf.file_data_id"
                                                                    :disabled="!containerConf.is_file_custom">
                                                                </bkbcs-input>
                                                                <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="containerConf.is_file_custom">
                                                                    {{$t('是否自定义')}}
                                                                    <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('采集id对应数据平台的data id，不勾选平台将分配默认的data id进行日志的清洗和入库。如果有特别的清洗和计算要求，用户可以填写自己的data id') }"></i>
                                                                </bk-checkbox>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>

                                                <i class="bcs-icon bcs-icon-close log-close" @click="removeContainerConf(containerConf, index)" v-if="curCrdInstance.workload.container_confs.length > 1"></i>
                                            </section>

                                            <bk-button class="log-block-btn mt10" @click="addContainerConf">
                                                <i class="bcs-icon bcs-icon-plus"></i>
                                                {{$t('点击增加')}}
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
                                        <div class="label tl" style="width: 300px;">{{$t('匹配标签(labels)')}}：</div>
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
                                        <div class="label tl" style="width: 300px;">{{$t('匹配表达式(expressions)')}}：</div>
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
                                        <div class="label">{{$t('采集路径')}}：</div>
                                        <div class="content">
                                            <section class="log-inner-wrapper mb10">
                                                <div class="bk-form-content log-flex mb10">
                                                    <div class="bk-form-inline-item">
                                                        <div class="log-form">
                                                            <div class="label">{{$t('标准输出')}}：</div>
                                                            <div class="content" style="width: 223px; margin-right: 32px;">
                                                                <bk-checkbox name="cluster-classify-checkbox" v-model="curCrdInstance.selector.stdout" true-value="true" false-value="false">
                                                                    {{$t('是否采集')}}
                                                                    <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.right="$t('如果不勾选，将不采集此容器的标准输出')"></i>
                                                                </bk-checkbox>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div class="bk-form-inline-item">
                                                        <div class="log-form">
                                                            <div class="label" style="width: 108px;">{{$t('标准采集ID')}}：</div>
                                                            <div class="content">
                                                                <bkbcs-input
                                                                    style="width: 80px;"
                                                                    type="number"
                                                                    :min="0"
                                                                    :placeholder="$t('请输入')"
                                                                    :value.sync="curCrdInstance.selector.std_data_id"
                                                                    :disabled="!curCrdInstance.selector.is_std_custom">
                                                                </bkbcs-input>
                                                                <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="curCrdInstance.selector.is_std_custom">
                                                                    {{$t('是否自定义')}}
                                                                    <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('采集id对应数据平台的data id，不勾选平台将分配默认的data id进行日志的清洗和入库。如果有特别的清洗和计算要求，用户可以填写自己的data id') }"></i>
                                                                </bk-checkbox>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>

                                                <div class="bk-form-content log-flex">
                                                    <div class="bk-form-inline-item">
                                                        <div class="log-form" style="width: 345px;">
                                                            <div class="label">{{$t('文件路径')}}：</div>
                                                            <div class="content log-path-wrapper" style="width: 223px;">
                                                                <textarea class="bk-form-textarea" v-model="curCrdInstance.selector.log_paths_str" style="width: 223px;" :placeholder="$t('多个以;分隔')"></textarea>

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
                                                            <div class="label" style="width: 110px;">{{$t('文件采集ID')}}：</div>
                                                            <div class="content">
                                                                <bkbcs-input
                                                                    style="width: 80px;"
                                                                    type="number"
                                                                    :min="0"
                                                                    :placeholder="$t('请输入')"
                                                                    :value.sync="curCrdInstance.selector.file_data_id"
                                                                    :disabled="!curCrdInstance.selector.is_file_custom">
                                                                </bkbcs-input>
                                                                <bk-checkbox class="ml5" name="cluster-classify-checkbox" v-model="curCrdInstance.selector.is_file_custom">
                                                                    {{$t('是否自定义')}}
                                                                    <i class="bcs-icon bcs-icon-question-circle" v-bk-tooltips.left="{ width: 400, content: $t('采集id对应数据平台的data id，不勾选平台将分配默认的data id进行日志的清洗和入库。如果有特别的清洗和计算要求，用户可以填写自己的data id') }"></i>
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
                                <label class="bk-label">{{$t('附加日志标签')}}：</label>
                            </div>

                            <div class="log-wrapper">
                                <bk-keyer
                                    :key-list.sync="curLogLabelList"
                                    :var-list="varList"
                                    @change="updateLogLabels">
                                </bk-keyer>
                                <div class="mt10 mb10">
                                    <bk-checkbox class="mr20" v-model="curCrdInstance.auto_add_pod_labels" name="cluster-classify-checkbox" true-value="true" false-value="false">
                                        {{$t('是否自动添加Pod中的labels')}}
                                    </bk-checkbox>
                                </div>

                                <div>
                                    <bk-checkbox v-model="curCrdInstance.package_collection" name="cluster-classify-checkbox" true-value="true" false-value="false">
                                        {{$t('是否打包上报')}}
                                        <i class="path-tip bcs-icon bcs-icon-question-circle" v-bk-tooltips.right="{ width: 400, content: $t('若单日志文件打印速度超过10条/秒，可以考虑开启日志打包上报功能以节约带宽并在一定程度上降低日志采集组件的资源占用') }"></i>
                                    </bk-checkbox>
                                </div>
                            </div>
                        </template>

                        <div class="bk-form-item mt15">
                            <bk-button type="primary" :loading="isDataSaveing" @click.stop.prevent="saveCrdInstance">{{curCrdInstance.id ? $t('更新') : $t('创建')}}</bk-button>
                            <bk-button @click.stop.prevent="hideCrdInstanceSlider" :disabled="isDataSaveing">{{$t('取消')}}</bk-button>
                        </div>
                    </div>
                </div>
            </bk-sideslider>

            <bk-sideslider
                :quick-close="true"
                :is-show.sync="detailSliderConf.isShow"
                :title="detailSliderConf.title"
                :width="'700'">
                <div class="p30" slot="content">
                    <p class="data-title">
                        {{$t('基础信息')}}
                    </p>
                    <div class="biz-metadata-box vertical mb20">
                        <div class="data-item">
                            <p class="key">{{$t('所属集群')}}：</p>
                            <p class="value">{{clusterName || '--'}}</p>
                        </div>
                        <div class="data-item">
                            <p class="key">{{$t('命名空间')}}：</p>
                            <p class="value">{{curCrdInstance.namespace === 'default' ? $t('所有') : curCrdInstance.namespace}}</p>
                        </div>
                        <div class="data-item">
                            <p class="key">{{$t('规则名称')}}：</p>
                            <p class="value">{{curCrdInstance.name || '--'}}</p>
                        </div>
                    </div>
                    <p class="data-title">
                        {{$t('日志源信息')}}
                    </p>

                    <div class="biz-metadata-box vertical mb0">
                        <div class="data-item">
                            <p class="key">{{$t('日志源类型')}}：</p>
                            <p class="value">{{logSource[curCrdInstance.log_source_type]}}</p>
                        </div>
                        <template v-if="curCrdInstance.log_source_type === 'selected_containers'">
                            <div class="data-item">
                                <p class="key">{{$t('应用类型')}}：</p>
                                <p class="value">{{curCrdInstance.workload.type || '--'}}</p>
                            </div>
                            <div class="data-item">
                                <p class="key">{{$t('应用名称')}}：</p>
                                <p class="value">{{curCrdInstance.workload.name || '--'}}</p>
                            </div>

                            <div class="data-item">
                                <p class="key">{{$t('采集路径')}}：</p>
                                <div class="value">
                                </div>
                            </div>
                        </template>
                        <template v-else-if="curCrdInstance.log_source_type === 'all_containers'">
                            <div class="data-item">
                                <p class="key">{{$t('是否采集')}}：</p>
                                <p class="value">{{curCrdInstance.default_conf.stdout === 'true' ? $t('是') : $t('否')}}</p>
                            </div>
                            <div class="data-item">
                                <p class="key">{{$t('标准采集ID')}}：</p>
                                <p class="value">{{curCrdInstance.default_conf.std_data_id || '--'}} ({{curCrdInstance.default_conf.is_std_custom ? '自定义' : '默认'}})</p>
                            </div>
                            <div class="data-item">
                                <p class="key">{{$t('文件采集ID')}}：</p>
                                <p class="value">{{curCrdInstance.default_conf.file_data_id || '--'}} ({{curCrdInstance.default_conf.is_file_custom ? '自定义' : '默认'}})</p>
                            </div>
                            <div class="data-item">
                                <p class="key">{{$t('文件路径')}}：</p>
                                <p class="value">{{curCrdInstance.default_conf.log_paths_str || '--'}}</p>
                            </div>
                        </template>
                        <template v-else-if="curCrdInstance.log_source_type === 'selected_labels'">
                            <div class="data-item">
                                <p class="key">{{$t('是否采集')}}：</p>
                                <p class="value">{{curCrdInstance.selector.stdout === 'true' ? $t('是') : $t('否')}}</p>
                            </div>
                            <div class="data-item">
                                <p class="key">{{$t('标准采集ID')}}：</p>
                                <p class="value">{{curCrdInstance.selector.std_data_id || '--'}} ({{curCrdInstance.selector.is_std_custom ? '自定义' : '默认'}})</p>
                            </div>
                            <div class="data-item">
                                <p class="key">{{$t('文件采集ID')}}：</p>
                                <p class="value">{{curCrdInstance.selector.file_data_id || '--'}} ({{curCrdInstance.selector.is_file_custom ? '自定义' : '默认'}})</p>
                            </div>
                            <div class="data-item">
                                <p class="key">{{$t('匹配标签')}}：</p>
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
                                <p class="key">{{$t('匹配表达式')}}：</p>
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
                                <p class="key">{{$t('文件路径')}}：</p>
                                <p class="value">{{curCrdInstance.selector.log_paths_str || '--'}}</p>
                            </div>
                        </template>
                    </div>

                    <div class="biz-metadata-box mb0 mt5" style="border: none;" v-if="curCrdInstance.log_source_type === 'selected_containers'">
                        <table class="bk-table bk-log-table">
                            <thead>
                                <tr>
                                    <th style="width: 130px;">{{$t('容器名')}}</th>
                                    <th style="width: 200px;">{{$t('标准输出')}}</th>
                                    <th>{{$t('文件路径')}}</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="(containerConf, index) of curCrdInstance.workload.container_confs" :key="index">
                                    <td>
                                        {{containerConf.name || '--'}}
                                    </td>
                                    <td>
                                        {{$t('采集ID')}}：{{containerConf.std_data_id || '--'}} ({{containerConf.is_std_custom ? $t('自定义') : $t('默认')}})<br />
                                        {{$t('是否采集')}}：{{containerConf.stdout === 'true' ? $t('是') : $t('否')}}
                                    </td>
                                    <td>
                                        <p>{{$t('采集ID')}}：{{containerConf.file_data_id || '--'}} ({{containerConf.is_file_custom ? $t('自定义') : $t('默认')}})</p>
                                        <div class="log-key-value">
                                            <div style="width: 38px;">{{$t('路径')}}：</div>
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
                </div>
            </bk-sideslider>
        </div>
    </div>
</template>

<script>
    import { catchErrorHandler } from '@/common/util'
    import bkKeyer from '@/components/keyer'
    import bkExpression from './expression'

    export default {
        components: {
            bkKeyer,
            bkExpression
        },
        data () {
            return {
                isInitLoading: true,
                isPageLoading: false,
                exceptionCode: null,
                curPageData: [],
                isDataSaveing: false,
                prmissions: {},
                pageConf: {
                    count: 0,
                    totalPage: 1,
                    limit: 5,
                    current: 1,
                    show: true
                },
                crdInstanceSlider: {
                    title: this.$t('新建规则'),
                    isShow: false
                },
                searchParams: {
                    clusterIndex: 0,
                    namespace: 0,
                    workload_type: 0,
                    workload_name: ''
                },
                appTypes: [
                    {
                        id: 'Deployment',
                        name: 'Deployment'
                    },
                    {
                        id: 'DaemonSet',
                        name: 'DaemonSet'
                    },
                    {
                        id: 'Job',
                        name: 'Job'
                    },
                    {
                        id: 'StatefulSet',
                        name: 'StatefulSet'
                    },
                    {
                        id: 'GameStatefulSet',
                        name: 'GameStatefulSet'
                    }
                ],
                curLogLabelList1: [
                    {
                        key: '',
                        value: ''
                    }
                ],
                searchScope: '',
                nameSpaceList: [],
                curLabelList: [
                    {
                        key: '',
                        value: ''
                    }
                ],

                dbTypes: [
                    {
                        id: 'mysql',
                        name: 'mysql'
                    },
                    {
                        id: 'spider',
                        name: 'spider'
                    }
                ],

                logSource: {
                    'selected_containers': this.$t('指定容器'),
                    'selected_labels': this.$t('指定标签'),
                    'all_containers': this.$t('所有容器')
                },
                curCrdInstance: {
                    // 'crd_kind': 'BcsLogConfig',
                    // 'cluster_id': '',
                    'namespace': '',
                    'config_type': 'custom',
                    'log_source_type': 'selected_containers',
                    'app_id': '',
                    'labels': {},
                    'auto_add_pod_labels': 'false',
                    'package_collection': 'false',
                    'default_conf': {
                        'stdout': 'true',
                        'is_std_custom': false,
                        'is_file_custom': false,
                        'std_data_id': '',
                        'file_data_id': '',
                        'log_paths': [],
                        'log_paths_str': ''
                    },
                    'workload': {
                        'name': '',
                        'type': '',
                        'container_confs': [
                            {
                                'name': '',
                                'std_data_id': '',
                                'file_data_id': '',
                                'stdout': 'true',
                                'log_paths': [],
                                'log_paths_str': '',
                                'is_std_custom': false,
                                'is_file_custom': false
                            }
                        ]
                    },
                    'selector': {
                        'std_data_id': '',
                        'file_data_id': '',
                        'stdout': 'true',
                        'log_paths': [],
                        'log_paths_str': '',
                        'match_labels': {},
                        'is_std_custom': false,
                        'is_file_custom': false,
                        'match_labels_list': [
                            {
                                'key': '',
                                'value': ''
                            }
                        ],
                        'match_expressions': []
                    }
                },
                detailSliderConf: {
                    isShow: false,
                    title: ''
                },
                crdKind: 'BcsLogConfig',
                defaultStdDataId: 0,
                defaultFileDataId: 0
            }
        },
        computed: {
            isEn () {
                return this.$store.state.isEn
            },
            varList () {
                return this.$store.state.variable.varList
            },
            projectId () {
                return this.$route.params.projectId
            },
            crdInstanceList () {
                const list = Object.assign([], this.$store.state.crdInstanceList)
                const results = list.map(item => {
                    const data = {
                        ...item,
                        ...item.crd_data
                    }
                    if (data.log_source_type === 'selected_containers') {
                        data.crd_data.workload.isExpand = false
                        data.crd_data.workload.container_confs.forEach(conf => {
                            if (!conf.log_paths) {
                                conf.log_paths = []
                            }
                        })
                    } else if (data.log_source_type === 'selected_labels') {
                        if (!data.selector.hasOwnProperty('match_labels')) {
                            data.selector.match_labels = {}
                        }
                        if (!data.selector.hasOwnProperty('match_expressions')) {
                            data.selector.match_expressions = []
                        }
                    }
                    return data
                })
                return results
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            },
            curProject () {
                return this.$store.state.curProject
            },
            clusterId () {
                return this.$route.params.clusterId
            },
            clusterName () {
                const cluster = this.clusterList.find(item => {
                    return item.cluster_id === this.clusterId
                })
                return cluster ? cluster.name : ''
            },
            searchScopeList () {
                const clusterList = this.$store.state.cluster.clusterList
                let results = []
                if (clusterList.length) {
                    results = []
                    clusterList.forEach(item => {
                        results.push({
                            id: item.cluster_id,
                            name: item.name
                        })
                    })
                }

                return results
            },
            curCrdInstanceName () {
                // 编辑状态，使用bcslog_name
                if (this.curCrdInstance.id && this.curCrdInstance.bcslog_name) {
                    return this.curCrdInstance.bcslog_name
                } else if (this.curCrdInstance.log_source_type === 'all_containers') {
                    return 'default-std-log'
                } else {
                    const app = this.curCrdInstance.workload
                    if (app['type'] && app['name']) {
                        return `${app['type']}-${app['name']}-log`.toLowerCase()
                    } else {
                        return 'log'
                    }
                }
            },
            curLogLabelList () {
                const keyList = []
                const labels = this.curCrdInstance.labels || {}
                for (const key in labels) {
                    keyList.push({
                        key: key,
                        value: labels[key]
                    })
                }
                if (!keyList.length) {
                    keyList.push({
                        key: '',
                        value: ''
                    })
                }
                return keyList
            }
        },
        watch: {
            crdInstanceList () {
                this.curPageData = this.getDataByPage(this.pageConf.current)
            },
            curPageData () {
                this.curPageData.forEach(item => {
                    if (item.clb_status && item.clb_status !== 'Running') {
                        this.getCrdInstanceStatus(item)
                    }
                })
            },
            curCrdInstance: {
                deep: true,
                handler () {
                    if (this.curCrdInstance.log_source_type === 'selected_containers') {
                        const containerConfs = this.curCrdInstance.workload.container_confs
                        containerConfs.forEach(conf => {
                            if (!conf.is_std_custom) {
                                conf.std_data_id = this.defaultStdDataId
                            }
                            if (!conf.is_file_custom) {
                                conf.file_data_id = this.defaultFileDataId
                            }
                        })
                    } else if (this.curCrdInstance.log_source_type === 'all_containers') {
                        const defaultConf = this.curCrdInstance.default_conf
                        if (!defaultConf.is_std_custom) {
                            defaultConf.std_data_id = this.defaultStdDataId
                        }
                        if (!defaultConf.is_file_custom) {
                            defaultConf.file_data_id = this.defaultFileDataId
                        }
                    } else if (this.curCrdInstance.log_source_type === 'selected_labels') {
                        const selector = this.curCrdInstance.selector
                        if (!selector.is_std_custom) {
                            selector.std_data_id = this.defaultStdDataId
                        }
                        if (!selector.is_file_custom) {
                            selector.file_data_id = this.defaultFileDataId
                        }
                    }
                }
            }
        },
        created () {
            this.getCrdInstanceList()
            this.getNameSpaceList()
            this.getLogInfo()
        },
        methods: {
            goBack () {
                this.$router.push({
                    name: 'logCrdcontroller',
                    params: {
                        projectId: this.projectId
                    }
                })
            },

            /**
             * 搜索列表
             */
            handleSearch () {
                this.pageConf.current = 1
                this.isPageLoading = true

                this.getCrdInstanceList()
            },

            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            handlePageSizeChange (pageSize) {
                this.pageConf.limit = pageSize
                this.pageConf.current = 1
                this.initPageConf()
                this.handlePageChange()
            },

            /**
             * 新建
             */
            createLoadBlance () {
                this.curCrdInstance = {
                    // 'crd_kind': 'BcsLog',
                    // 'cluster_id': this.clusterId,
                    'namespace': '',
                    'config_type': 'custom',
                    'log_source_type': 'selected_containers',
                    'app_id': this.curProject.cc_app_id,

                    'labels': {},
                    'auto_add_pod_labels': 'false',
                    'package_collection': 'false',
                    'default_conf': {
                        'stdout': 'true',
                        'std_data_id': this.defaultStdDataId,
                        'file_data_id': this.defaultFileDataId,
                        'log_paths': [],
                        'log_paths_str': '',
                        'is_std_custom': false,
                        'is_file_custom': false
                    },
                    'workload': {
                        'name': '',
                        'type': '',
                        'container_confs': [
                            {
                                'name': '',
                                'std_data_id': this.defaultStdDataId,
                                'file_data_id': this.defaultFileDataId,
                                'stdout': 'true',
                                'log_paths': '',
                                'log_paths_str': '',
                                'is_std_custom': false,
                                'is_file_custom': false
                            }
                        ]
                    },
                    'selector': {
                        'std_data_id': this.defaultStdDataId,
                        'file_data_id': this.defaultFileDataId,
                        'stdout': 'true',
                        'log_paths': [],
                        'log_paths_str': '',
                        'match_labels': {},
                        'match_labels_list': [
                            {
                                'key': '',
                                'value': ''
                            }
                        ],
                        'match_expressions': [],
                        'match_expressions_list': []
                    }
                }

                this.crdInstanceSlider.title = this.$t('新建规则')
                this.crdInstanceSlider.isShow = true
            },

            updateLogLabels (list, data) {
                this.curCrdInstance.labels = data
            },

            updateMatchLabels (list, data) {
                const obj = {}
                for (const key in data) {
                    if (key) {
                        obj[key] = data[key]
                    }
                }
                this.curCrdInstance.selector.match_labels = obj
            },

            updateExpressions (list, data) {
                this.curCrdInstance.selector.match_expressions = list
            },

            /**
             * 编辑
             * @param  {object} crdInstance crdInstance
             * @param  {number} index 索引
             */
            async editCrdInstance (crdInstance, isReadonly) {
                if (this.isDetailLoading) {
                    return false
                }
                this.isDetailLoading = true

                try {
                    const projectId = this.projectId
                    const clusterId = this.clusterId
                    const crdId = crdInstance.id
                    const crdKind = this.crdKind
                    const res = await this.$store.dispatch('crdcontroller/getLogCrdInstanceDetail', {
                        projectId,
                        clusterId,
                        crdId,
                        crdKind
                    })
                    const data = {
                        ...res.data,
                        ...res.data.crd_data
                    }
                    data.id = res.data.id
                    data.name = res.data.name
                    if (data.log_source_type === 'selected_containers') {
                        data.workload.container_confs.forEach(conf => {
                            if (!conf.log_paths) {
                                conf.log_paths = []
                            }
                            conf.log_paths_str = conf.log_paths.join(';')
                            // 是否自定义
                            conf.is_std_custom = String(conf.std_data_id) !== String(this.defaultStdDataId)
                            conf.is_file_custom = String(conf.file_data_id) !== String(this.defaultFileDataId)
                        })
                    } else if (data.log_source_type === 'all_containers') {
                        // 是否自定义
                        data.default_conf.is_std_custom = String(data.default_conf.std_data_id) !== String(this.defaultStdDataId)
                        data.default_conf.is_file_custom = String(data.default_conf.file_data_id) !== String(this.defaultFileDataId)
                        if (data.default_conf.log_paths) {
                            data.default_conf.log_paths_str = data.default_conf.log_paths.join(';')
                        } else {
                            data.default_conf.log_paths = []
                            data.default_conf.log_paths_str = ''
                        }
                    } else if (data.log_source_type === 'selected_labels') {
                        const selector = data.selector
                        // 是否自定义
                        selector.is_std_custom = String(selector.std_data_id) !== String(this.defaultStdDataId)
                        selector.is_file_custom = String(selector.file_data_id) !== String(this.defaultFileDataId)
                        selector.log_paths_str = selector.log_paths.join(';')

                        // 匹配标签(
                        selector.match_labels_list = []
                        if (selector.match_labels && Object.keys(selector.match_labels).length) {
                            for (const key in selector.match_labels) {
                                selector.match_labels_list.push({
                                    key: key,
                                    value: selector.match_labels[key]
                                })
                            }
                        } else {
                            selector.match_labels = {}
                            selector.match_labels_list.push({
                                key: '',
                                value: ''
                            })
                        }

                        // 匹配表达式
                        if (selector.match_expressions && selector.match_expressions.length) {
                            selector.match_expressions_list = selector.match_expressions.map(item => {
                                if (!item.hasOwnProperty('values')) {
                                    item.values = ''
                                }
                                return item
                            })
                        } else {
                            selector.match_expressions = []
                            selector.match_expressions_list = []
                        }
                    }

                    if (!data.hasOwnProperty('package_collection')) {
                        data.package_collection = 'false'
                    }

                    if (!data.hasOwnProperty('auto_add_pod_labels')) {
                        data.auto_add_pod_labels = 'false'
                    }

                    this.curCrdInstance = data

                    if (isReadonly) {
                        this.detailSliderConf.title = `${this.curCrdInstance.name}`
                        this.detailSliderConf.isShow = true
                    } else {
                        this.crdInstanceSlider.title = this.$t('编辑规则')
                        this.crdInstanceSlider.isShow = true
                    }
                } catch (e) {
                    // catchErrorHandler(e, this)
                } finally {
                    this.isDetailLoading = false
                }
            },

            /**
             * 删除
             * @param  {object} crdInstance crdInstance
             * @param  {number} index 索引
             */
            async removeCrdInstance (crdInstance, index) {
                const self = this
                const projectId = this.projectId
                const clusterId = this.clusterId
                const crdKind = this.crdKind
                const crdId = crdInstance.id

                this.$bkInfo({
                    title: this.$t('确认删除'),
                    clsName: 'biz-remove-dialog',
                    content: this.$createElement('p', {
                        class: 'biz-confirm-desc'
                    }, `${this.$t('确定要删除')}【${crdInstance.name}】？`),
                    async confirmFn () {
                        self.isPageLoading = true
                        try {
                            await self.$store.dispatch('crdcontroller/deleteCrdInstance', { projectId, clusterId, crdKind, crdId })
                            self.$bkMessage({
                                theme: 'success',
                                message: self.$t('删除成功')
                            })
                            self.getCrdInstanceList()
                        } catch (e) {
                            catchErrorHandler(e, this)
                        } finally {
                            self.isPageLoading = false
                        }
                    }
                })
            },

            /**
             * 获取
             * @param  {number} crdInstanceId id
             * @return {object} crdInstance crdInstance
             */
            getCrdInstanceById (crdInstanceId) {
                return this.crdInstanceList.find(item => {
                    return item.id === crdInstanceId
                })
            },

            /**
             * 初始化分页配置
             */
            initPageConf () {
                const total = this.crdInstanceList.length
                this.pageConf.count = total
                this.pageConf.totalPage = Math.ceil(total / this.pageConf.limit)
                if (this.pageConf.current > this.pageConf.totalPage) {
                    this.pageConf.current = this.pageConf.totalPage
                }
            },

            /**
             * 重新加载当前页
             */
            reloadCurPage () {
                this.initPageConf()
                if (this.pageConf.current > this.pageConf.totalPage) {
                    this.pageConf.current = this.pageConf.totalPage
                }
                this.curPageData = this.getDataByPage(this.pageConf.current)
            },

            /**
             * 获取页数据
             * @param  {number} page 页
             * @return {object} data lb
             */
            getDataByPage (page) {
                // 如果没有page，重置
                if (!page) {
                    this.pageConf.current = page = 1
                }
                let startIndex = (page - 1) * this.pageConf.limit
                let endIndex = page * this.pageConf.limit
                // this.isPageLoading = true
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.crdInstanceList.length) {
                    endIndex = this.crdInstanceList.length
                }
                this.isPageLoading = false
                return this.crdInstanceList.slice(startIndex, endIndex)
            },

            /**
             * 分页改变回调
             * @param  {number} page 页
             */
            handlePageChange (page = 1) {
                this.isPageLoading = true
                this.pageConf.current = page
                const data = this.getDataByPage(page)
                this.curPageData = JSON.parse(JSON.stringify(data))
            },

            /**
             * 隐藏lb侧面板
             */
            hideCrdInstanceSlider () {
                this.crdInstanceSlider.isShow = false
            },

            /**
             * 加载数据
             */
            async getCrdInstanceList () {
                try {
                    const projectId = this.projectId
                    const clusterId = this.clusterId
                    const crdKind = this.crdKind

                    const params = {
                        // cluster_id: this.clusterId
                        // crd_kind: this.crdKind
                    }

                    if (this.searchParams.namespace) {
                        params.namespace = this.searchParams.namespace
                    }

                    if (this.searchParams.workload_type) {
                        params.workload_type = this.searchParams.workload_type
                    }

                    if (this.searchParams.workload_name) {
                        params.workload_name = this.searchParams.workload_name
                    }

                    this.isPageLoading = true
                    await this.$store.dispatch('getBcsCrdsList', {
                        projectId,
                        clusterId,
                        crdKind,
                        params
                    })

                    this.initPageConf()
                    this.curPageData = this.getDataByPage(this.pageConf.current)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isPageLoading = false
                        this.isInitLoading = false
                    }, 500)
                }
            },

            /**
             * 获取命名空间列表
             */
            async getNameSpaceList () {
                try {
                    const projectId = this.projectId
                    const clusterId = this.clusterId
                    const res = await this.$store.dispatch('crdcontroller/getNameSpaceListByCluster', { projectId, clusterId })
                    const list = res.data
                    list.forEach(item => {
                        item.isSelected = false
                    })
                    this.nameSpaceList.splice(0, this.nameSpaceList.length, ...list)
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取日志信息
             */
            async getLogInfo () {
                try {
                    const projectId = this.projectId
                    const res = await this.$store.dispatch('getLogPlans', projectId)
                    this.defaultStdDataId = res.data.std_data_id
                    this.defaultFileDataId = res.data.file_data_id
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 选择/取消选择命名空间
             * @param  {object} nameSpace 命名空间
             * @param  {number} index 索引
             */
            toggleSelected (nameSpace, index) {
                nameSpace.isSelected = !nameSpace.isSelected
                this.nameSpaceList = JSON.parse(JSON.stringify(this.nameSpaceList))
            },

            /**
             * 检查提交的数据
             * @return {boolean} true/false 是否合法
             */
            checkData (params) {
                if (params.log_source_type !== 'all_containers' && !params.namespace) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择命名空间'),
                        delay: 5000
                    })
                    return false
                }

                if (params.log_source_type === 'selected_containers') {
                    if (!params.workload.type) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请选择应用类型')
                        })
                        return false
                    }

                    if (!params.workload.name) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请输入应用名称')
                        })
                        return false
                    }

                    try {
                        const reg = new RegExp(params.workload.name)
                        console.log(reg)
                    } catch (e) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('应用名称不合法')
                        })
                        return false
                    }

                    for (const conf of params.workload.container_confs) {
                        if (!conf.name) {
                            this.$bkMessage({
                                theme: 'error',
                                message: this.$t('请输入容器名')
                            })
                            return false
                        }

                        if (!conf.std_data_id) {
                            this.$bkMessage({
                                theme: 'error',
                                message: this.$t('请输入标准采集ID')
                            })
                            return false
                        }

                        if (!conf.file_data_id) {
                            this.$bkMessage({
                                theme: 'error',
                                message: this.$t('请输入文件采集ID')
                            })
                            return false
                        }
                    }
                } else if (params.log_source_type === 'all_containers') {
                    if (!params.default_conf.std_data_id) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请输入标准采集ID')
                        })
                        return false
                    }
                    if (!params.default_conf.file_data_id) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请输入文件采集ID')
                        })
                        return false
                    }
                } else if (params.log_source_type === 'selected_labels') {
                    console.log('params', params)
                    if (!params.selector.hasOwnProperty('match_labels') && !params.selector.hasOwnProperty('match_expressions')) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('匹配标签和匹配表达式不能同时为空')
                        })
                        return false
                    }
                    if (!params.selector.std_data_id) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请输入标准采集ID')
                        })
                        return false
                    }

                    if (!params.selector.file_data_id) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请输入文件采集ID')
                        })
                        return false
                    }

                    if (!params.selector.log_paths.length) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请输入文件路径')
                        })
                        return false
                    }
                }

                return true
            },

            showCrdInstanceDetail (data) {
                data.labels = []
                for (const key in data.pod_selector) {
                    data.labels.push({
                        key: key,
                        value: data.pod_selector[key]
                    })
                }
                this.curCrdInstance = data

                this.detailSliderConf.title = `${data.name}`
                this.detailSliderConf.isShow = true
            },

            /**
             * 格式化数据，符合接口需要的格式
             */
            formatData () {
                const params = JSON.parse(JSON.stringify(this.curCrdInstance))
                // 附加日志标签
                const labels = params.labels
                params.labels = {}
                if (params.labels) {
                    for (const key in labels) {
                        if (key) {
                            params.labels[key] = labels[key]
                        }
                    }
                }

                if (params.log_source_type === 'selected_containers') {
                    params.workload.container_confs.forEach(conf => {
                        const paths = conf.log_paths_str.split(/[;|\n]/).filter(item => {
                            return !!item.trim()
                        }).map(item => {
                            return item.trim()
                        })

                        if (paths.length) {
                            conf.log_paths = paths
                        } else {
                            delete conf.log_paths
                        }

                        // 接口接受'true'、'false'字符类型
                        conf.stdout = String(conf.stdout)
                    })
                    delete params.std_data_id
                    delete params.is_std_custom
                    delete params.stdout
                    delete params.selector
                    delete params.default_conf
                } else if (params.log_source_type === 'all_containers') {
                    if (!params.id) {
                        if (!params.namespace) {
                            params.config_type = 'default'
                            delete params.namespace
                        } else {
                            params.config_type = 'custom'
                        }
                    }

                    const paths = params.default_conf.log_paths_str.split(/[;|\n]/).filter(item => {
                        return !!item.trim()
                    }).map(item => {
                        return item.trim()
                    })

                    if (paths.length) {
                        params.default_conf.log_paths = paths
                    } else {
                        delete params.default_conf.log_paths
                    }
                    delete params.workload
                    delete params.selector
                    delete params.default_conf.log_paths_str
                    delete params.default_conf.is_std_custom
                    delete params.default_conf.is_file_custom
                } else if (params.log_source_type === 'selected_labels') {
                    const paths = params.selector.log_paths_str.split(/[;|\n]/).filter(item => {
                        return !!item.trim()
                    }).map(item => {
                        return item.trim()
                    })

                    params.selector.log_paths = paths
                    delete params.std_data_id
                    delete params.is_std_custom
                    delete params.stdout
                    delete params.workload
                    delete params.default_conf
                    // selector
                    delete params.selector.match_expressions_list
                    delete params.selector.match_labels_list
                    delete params.selector.log_paths_str

                    if (Object.keys(params.selector.match_labels).length === 0) {
                        delete params.selector.match_labels
                    }

                    if (params.selector.match_expressions.length === 0) {
                        delete params.selector.match_expressions
                    } else {
                        params.selector.match_expressions.forEach(item => {
                            if (['Exists', 'DoesNotExist'].includes(item.operator)) {
                                delete item.values
                            }
                        })
                    }
                }

                return params
            },

            /**
             * 保存新建的
             */
            async createCrdInstance (params) {
                const projectId = this.projectId
                const clusterId = this.clusterId
                const crdKind = this.crdKind
                const data = params
                this.isDataSaveing = true

                try {
                    await this.$store.dispatch('crdcontroller/addCrdInstance', { projectId, clusterId, crdKind, data })

                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('规则创建成功')
                    })
                    this.getCrdInstanceList()
                    this.hideCrdInstanceSlider()
                } catch (e) {
                    // catchErrorHandler(e, this)
                } finally {
                    this.isDataSaveing = false
                }
            },

            /**
             * 保存更新的
             */
            async updateCrdInstance (params) {
                const projectId = this.projectId
                const clusterId = this.clusterId
                const crdKind = this.crdKind
                const data = params
                this.isDataSaveing = true
                delete data.cluster_id
                // data.crd_kind = this.crdKind
                try {
                    await this.$store.dispatch('crdcontroller/updateCrdInstance', { projectId, clusterId, crdKind, data })

                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('规则更新成功')
                    })

                    this.hideCrdInstanceSlider()
                    // sideslider和loading层有样式冲突
                    setTimeout(() => {
                        this.getCrdInstanceList()
                    }, 500)
                } catch (e) {
                    // catchErrorHandler(e, this)
                } finally {
                    this.isDataSaveing = false
                }
            },

            /**
             * 保存
             */
            saveCrdInstance () {
                const params = this.formatData()
                if (this.checkData(params) && !this.isDataSaveing) {
                    if (this.curCrdInstance.id > 0) {
                        this.updateCrdInstance(params)
                    } else {
                        this.createCrdInstance(params)
                    }
                }
            },

            handleNamespaceSelect (index, data) {
                this.curCrdInstance.namespace = data.name
            },

            changeLabels (labels, data) {
                // this.curCrdInstance.pod_selector = data
                this.curCrdInstance.labels = labels
            },

            addContainerConf () {
                this.curCrdInstance.workload.container_confs.push({
                    'name': '',
                    'std_data_id': this.defaultStdDataId,
                    'file_data_id': this.defaultFileDataId,
                    'stdout': 'true',
                    'log_paths': [],
                    'log_paths_str': '',
                    'is_file_custom': false,
                    'is_std_custom': false

                })
            },

            removeContainerConf (data, index) {
                this.curCrdInstance.workload.container_confs.splice(index, 1)
            },

            toggleExpand (crdInstance) {
                crdInstance.crd_data.workload.isExpand = !crdInstance.crd_data.workload.isExpand
            }
        }
    }
</script>

<style scoped>
    @import './log_list.css';
</style>
