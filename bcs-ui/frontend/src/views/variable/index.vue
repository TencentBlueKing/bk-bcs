<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-var-title">
                {{$t('变量管理')}}
            </div>
            <bk-guide>
                <a class="bk-text-button" href="javascript: void(0);" @click="handleShowVarExample" v-if="!isSharedCluster">{{$t('如何从文件导入变量？')}}</a>
            </bk-guide>
        </div>
        <div class="biz-content-wrapper" style="margin: 0; padding: 0;" v-bkloading="{ isLoading: isLoading, opacity: 0.1 }">
            <template v-if="!isLoading">
                <div class="biz-panel-header">
                    <div class="left">
                        <bk-button type="primary" @click.stop.prevent="addVar">
                            <i class="bcs-icon bcs-icon-plus"></i>
                            <span>{{$t('新增变量')}}</span>
                        </bk-button>

                        <bk-button class="bk-button bk-default import-btn" v-if="!isSharedCluster">
                            <span @click="handleFileImport">{{$t('文件导入')}}</span>
                        </bk-button>

                        <bk-button @click.stop.prevent="removeVars">
                            <span>{{$t('批量删除')}}</span>
                        </bk-button>
                        <input ref="fileInput" type="file" style="opacity: 0; position: absolute; top: -111px;" name="upload" class="file-input" accept="application/json" @change="handleFileInput">
                    </div>
                    <div class="right">
                        <bk-select class="select-scope" data-placeholder="" v-model="searchScope" @change="searchVar">
                            <bk-option v-for="item in searchScopeList" :key="item.id" :name="item.name" :id="item.id"></bk-option>
                        </bk-select>
                        <bk-data-searcher
                            :search-key.sync="searchKeyword"
                            @search="searchVar"
                            @refresh="refresh">
                        </bk-data-searcher>
                    </div>
                </div>
                <div class="biz-table-wrapper">
                    <bk-table
                        v-bkloading="{ isLoading: isPageLoading && !isLoading }"
                        :data="curPageData"
                        :page-params="pageConf"
                        @selection-change="handleSelectionChange"
                        @page-change="pageChange"
                        @page-limit-change="changePageSize">
                        <bk-table-column type="selection" width="60" :selectable="(row, index) => row.category !== 'sys'"></bk-table-column>
                        <bk-table-column :label="$t('变量名称')" prop="name" width="200"></bk-table-column>
                        <bk-table-column label="KEY" prop="key"></bk-table-column>
                        <bk-table-column :label="$t('默认值')" prop="default_value"></bk-table-column>
                        <bk-table-column :label="$t('类型')" prop="category_name" width="120"></bk-table-column>
                        <bk-table-column :label="$t('作用范围')" prop="scope_name" width="140"></bk-table-column>
                        <bk-table-column :label="$t('操作')" width="250">
                            <template slot-scope="{ row }">
                                <a href="javascript:void(0);" class="bk-text-button" @click="getQuoteDetail(row)">{{$t('查看引用')}}</a>
                                <a href="javascript:void(0);" class="ml10 bk-text-button" @click="batchUpdate(row)" v-show="row.category !== 'sys' && (row.scope === 'namespace' || row.scope === 'cluster')">{{$t('批量更新')}}</a>

                                <template v-if="row.category === 'sys'">
                                    <a href="javascript:void(0);" class="bk-text-button is-disabled ml10" v-bk-tooltips.left="$t('系统内置变量，不能编辑')">{{$t('编辑')}}</a>
                                </template>
                                <template v-else>
                                    <a href="javascript:void(0);" class=" ml10 bk-text-button" @click="editVar(row)">{{$t('编辑')}}</a>
                                </template>

                                <template v-if="row.category === 'sys'">
                                    <a href="javascript:void(0);" class="ml10 bk-text-button is-disabled" v-bk-tooltips.left="row.category === 'sys' ? $t('系统内置变量，不能删除') : $t('已经被引用，不能删除')">{{$t('删除')}}</a>
                                </template>
                                <template v-else>
                                    <a href="javascript:void(0);" class="ml10 bk-text-button" @click="removeVar(row)">{{$t('删除')}}</a>
                                </template>
                            </template>
                        </bk-table-column>
                    </bk-table>
                </div>
            </template>
        </div>

        <bk-sideslider
            :is-show.sync="exampleConf.isShow"
            :title="exampleConf.title"
            :width="exampleConf.width"
            :quick-close="true">
            <div slot="content" style="position: relative;">
                <div class="biz-log-box" :style="{ height: `${winHeight - 200}px` }">
                    <ace
                        :value="editorConfig.content"
                        :width="editorConfig.width"
                        :height="editorConfig.height"
                        :lang="editorConfig.lang"
                        :read-only="editorConfig.readOnly"
                        :full-screen="editorConfig.fullScreen">
                    </ace>
                </div>
                <div class="example-desc">
                    <p>. {{$t('按上面的模板创建你的json文件，选择“文件导入”操作')}}</p>
                    <p>. {{$t('scope值含义，global表示全局变量，cluster表示集群变量，namespace表示命名空间变量')}}</p>
                    <p>. {{$t('cluster和namespace变量需要提供vars关键字，cluster变量的vars需要包含集群ID cluster_id和变量值 value')}}</p>
                    <p>. {{$t('namespace变量的vars需要包含集群ID cluster_id、命名空间名称 namespace 和变量值 value')}}</p>
                </div>
            </div>
        </bk-sideslider>

        <bk-sideslider
            :is-show.sync="batchUpdateConf.isShow"
            :title="batchUpdateConf.title"
            :width="batchUpdateConf.width">
            <div style="padding: 20px 20px 10px 20px;" slot="content" v-bkloading="{ isLoading: isBatchVarLoading }">
                <div style="height: 60px; margin-top: -60px;">
                    <bk-button class="fr bk-text-button f13" @click="toggleBatchMode">{{batchUpdateConf.mode === 'form' ? $t('切换为文本模式') : $t('切换为表单模式')}}</bk-button>
                </div>
                <table class="bk-table biz-data-table has-table-bordered" v-if="batchUpdateConf.mode === 'form'" style="border-bottom: none;">
                    <thead>
                        <tr>
                            <th v-if="curBatchVar && curBatchVar.scope === 'namespace'" style="width: 250px;">{{$t('所属')}}{{$t('集群')}}</th>
                            <th style="min-width: 200px;">{{$t('所属')}}{{curBatchVar && curBatchVar.scope === 'namespace' ? $t('命名空间') : $t('集群')}}</th>
                            <th>值</th>
                        </tr>
                    </thead>
                    <tbody>
                        <template v-if="batchVarList.length">
                            <tr v-for="(variable, index) in batchVarList" :key="index">
                                <td v-if="curBatchVar && curBatchVar.scope === 'namespace'">{{variable.cluster_name}}</td>
                                <td><p style="max-width: 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;" :title="variable.name">{{variable.name}}</p></td>
                                <td><bk-input v-model="variable.variable_value" /></td>
                            </tr>
                        </template>
                        <template v-else>
                            <tr>
                                <td colspan="2"><bcs-exception type="empty" scene="part"></bcs-exception></td>
                            </tr>
                        </template>
                    </tbody>
                </table>
                <div class="biz-log-box mb20" :style="{ height: `${winHeight - 200}px` }" v-else>
                    <ace
                        :value="batchEditorConfig.content"
                        :width="batchEditorConfig.width"
                        :height="batchEditorConfig.height"
                        :lang="batchEditorConfig.lang"
                        :read-only="batchEditorConfig.readOnly"
                        :full-screen="batchEditorConfig.fullScreen"
                        @init="editorInitAfter">
                    </ace>
                </div>
                <div v-if="batchVarList.length">
                    <bk-button type="primary" @click="saveBatchVar">{{$t('保存')}}</bk-button>
                    <bk-button @click="cancelBatchVar">{{$t('取消')}}</bk-button>
                </div>
            </div>
        </bk-sideslider>

        <bk-dialog
            :is-show="batchDialogConfig.isShow"
            :width="400"
            :has-header="false"
            :quick-close="false"
            :title="$t('确认删除')"
            @confirm="deleteVar(batchDialogConfig.removeIds)"
            @cancel="batchDialogConfig.isShow = false">
            <template slot="content">
                <div class="biz-batch-wrapper">
                    <p class="batch-title">{{$t('确定要删除以下变量？')}}</p>
                    <ul class="batch-list">
                        <li v-for="(item, index) of batchDialogConfig.list" :key="index">{{item}}</li>
                    </ul>
                </div>
            </template>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="varDialogConfig.isShow"
            :width="varDialogConfig.width"
            :title="varDialogConfig.title"
            :quick-close="false"
            @cancel="varDialogConfig.isShow = false">
            <template slot="content" v-bkloading="{ isLoading: isSaving }">
                <div class="content-inner">
                    <div class="bk-form" style="margin-bottom: 20px; margin-right: 10px;">
                        <div class="bk-form-item">
                            <label class="bk-label" style="width: 95px;">{{$t('作用范围')}}：</label>
                            <div class="bk-form-content" style="margin-left: 95px;">
                                <bk-radio-group v-model="curVar.scope">
                                    <bk-radio value="global"
                                        style="margin-right: 15px;"
                                        :disabled="curVar.quote_num !== undefined && curVar.quote_num > 0"
                                        v-if="!isSharedCluster">{{$t('全局变量')}}</bk-radio>
                                    <bk-radio value="cluster"
                                        style="margin-right: 15px;"
                                        :disabled="curVar.quote_num !== undefined && curVar.quote_num > 0"
                                        v-if="!isSharedCluster">{{$t('集群变量')}}</bk-radio>
                                    <bk-radio value="namespace" :disabled="curVar.quote_num !== undefined && curVar.quote_num > 0">{{$t('命名空间变量')}}</bk-radio>
                                </bk-radio-group>
                            </div>
                        </div>
                        <div class="bk-form-item is-required">
                            <label class="bk-label" style="width: 95px;">{{$t('名称')}}：</label>
                            <div class="bk-form-content" style="margin-left: 95px;">
                                <bk-input maxlength="32"
                                    :placeholder="$t('请输入32个字符以内的名称')"
                                    v-model="curVar.name" />
                            </div>
                        </div>
                        <div class="bk-form-item is-required">
                            <label class="bk-label" style="width: 95px;">KEY：</label>
                            <div class="bk-form-content" style="margin-left: 95px;">
                                <bk-input :disabled="curVar.quote_num !== undefined && curVar.quote_num > 0"
                                    maxlength="64"
                                    :placeholder="$t('请输入')"
                                    v-model="curVar.key" />
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label" style="width: 95px;">{{$t('默认值')}}：</label>
                            <div class="bk-form-content" style="margin-left: 95px;">
                                <bk-input :placeholder="$t('请输入')" v-model="curVar.default.value" />
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <label class="bk-label" style="width: 95px;">{{$t('说明')}}：</label>
                            <div class="bk-form-content" style="margin-left: 95px;">
                                <textarea maxlength="100" :class="['bk-form-textarea']" :placeholder="$t('请输入')" v-model="curVar.desc" style="height: 60px;"></textarea>
                                <p class="biz-tip" style="text-align: left;">{{$t('您可以在模板集中使用')}} {{curVarKeyText}} {{$t('来引用该变量')}}</p>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="isSaving">
                        <bk-button type="primary" disabled>{{$t('提交')}}</bk-button>
                        <bk-button disabled>{{$t('取消')}}</bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" @click="saveVar">{{$t('提交')}}</bk-button>
                        <bk-button @click="cancelVar">{{$t('取消')}}</bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :title="curVar.name"
            :is-show.sync="quoteDialogConf.isShow"
            :width="quoteDialogConf.width"
            :content="quoteDialogConf.content"
            :has-header="quoteDialogConf.hasHeader"
            :has-footer="false"
            :close-icon="true"
            @cancel="hideQuoteDialog"
            :ext-cls="'biz-var-quote-dialog'">
            <template slot="content">
                <div style="margin: -20px;">
                    <div style="min-height: 100px;">
                        <bk-table
                            v-bkloading="{ isLoading: isQuoteLoading }"
                            :data="curQuotePageData"
                            :page-params="quotePageConf"
                            :border="false"
                            :outer-border="false"
                            @page-change="quotePageChange">
                            <bk-table-column :label="$t('被引用位置')" prop="quote_location" key="quote_location" width="400" :show-overflow-tooltip="true" />
                            <bk-table-column :label="$t('上下文')" prop="context" width="150" key="context" />
                            <bk-table-column :label="$t('操作')" key="action">
                                <template slot-scope="{ row }">
                                    <a href="javascript:void(0)" class="bk-text-button" @click="checkVarQuote(row)">{{$t('查看详情')}}</a>
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
                    <div class="biz-page-box" v-if="!isQuoteLoading && quotePageConf.show && curQuotePageData.length">
                        <bk-pagination
                            :show-limit="false"
                            :current.sync="quotePageConf.curPage"
                            :count.sync="quotePageConf.count"
                            :limit="quotePageConf.pageSize"
                            @change="quotePageChange">
                        </bk-pagination>
                    </div>
                </div>
            </template>
        </bk-dialog>
    </div>
</template>

<script>
    import { catchErrorHandler } from '@/common/util'
    import ace from '@/components/ace-editor'
    import exampleData from './variable.json'
    import { mapGetters } from 'vuex'

    export default {
        components: {
            ace
        },
        data () {
            return {
                varScoped: {
                    global: this.$t('全局变量'),
                    namespace: this.$t('命名空间变量'),
                    cluster: this.$t('集群变量')
                },
                exampleConf: {
                    width: 800,
                    isShow: false,
                    title: this.$t('如何从文件导入变量？')
                },
                batchEditorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'json',
                    readOnly: false,
                    fullScreen: false,
                    content: '',
                    editor: null
                },
                editorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'json',
                    readOnly: false,
                    fullScreen: false,
                    content: JSON.stringify(exampleData, null, 4),
                    editor: null
                },
                winHeight: 0,
                curProjectData: null,
                isQuoteLoading: true,
                importContent: '',
                curAllSelectedData: [],
                curQuotePageData: [],
                quoteList: [],
                batchVarList: [],
                pageConf: {
                    total: 0,
                    totalPage: 1,
                    pageSize: 10,
                    curPage: 1
                },
                batchDialogConfig: {
                    isShow: false,
                    list: [],
                    removeIds: []
                },
                quotePageConf: {
                    totalPage: 1,
                    pageSize: 5,
                    curPage: 1,
                    count: 1,
                    show: true
                },
                curPageData: [],
                searchKeyword: '',
                searchScope: '',
                varDialogConfig: {
                    isShow: false,
                    width: 640,
                    title: this.$t('新增变量')
                },
                isBatchVarLoading: false,
                batchUpdateConf: {
                    mode: 'form',
                    isShow: false,
                    title: '',
                    width: 690
                },
                quoteDialogConf: {
                    isShow: false,
                    width: 690,
                    hasHeader: false,
                    closeIcon: false
                },
                curBatchVar: null,
                curVar: {
                    name: '',
                    key: '',
                    default: {
                        value: ''
                    },
                    desc: '',
                    scope: 'global'
                },
                isLoading: true,
                isPageLoading: false,
                isSaving: false,
                searchScopeList: [],
                alreadySelectedNums: 0
            }
        },
        computed: {
            isEn () {
                return this.$store.state.isEn
            },
            projectId () {
                return this.$route.params.projectId
            },
            varList () {
                return JSON.parse(JSON.stringify(this.$store.state.variable.varList))
            },
            curVarKeyText () {
                return `{{${this.curVar.key || this.$t('变量KEY')}}}`
            },
            ...mapGetters('cluster', ['isSharedCluster'])
        },
        created () {
            if (this.isSharedCluster) {
                this.curVar.scope = 'namespace'
                this.searchScope = 'namespace'
                this.searchScopeList = [
                    {
                        id: 'namespace',
                        name: this.$t('命名空间变量')
                    }
                ]
            } else {
                this.curVar.scope = 'global'
                this.searchScopeList = [
                    {
                        id: '',
                        name: this.$t('全部作用范围')
                    },
                    {
                        id: 'global',
                        name: this.$t('全局变量')
                    },
                    {
                        id: 'cluster',
                        name: this.$t('集群变量')
                    },
                    {
                        id: 'namespace',
                        name: this.$t('命名空间变量')
                    }
                ]
            }
        },
        mounted () {
            this.winHeight = window.innerHeight
            this.getDataByPage()
        },
        methods: {
            /**
             * 刷新列表
             */
            refresh () {
                this.pageConf.curPage = 1
                this.isPageLoading = true
                this.searchKeyword = ''
                this.searchScope = ''
                this.getDataByPage()
            },

            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            changePageSize (pageSize) {
                this.pageConf.pageSize = pageSize
                this.pageConf.curPage = 1
                this.getDataByPage()
            },

            /**
             * 显示新增变量窗口
             */
            addVar () {
                this.isSaving = false
                this.clearInput()
                this.varDialogConfig.title = this.$t('新增变量')
                this.varDialogConfig.isShow = true
            },

            /**
             * 取消批量删除
             */
            cancelBatchVar () {
                this.batchUpdateConf.isShow = false
            },

            /**
             * 获取对应变量所有命名空间变量列表
             *
             * @param  {Object} data 变量
             */
            async batchUpdate (data) {
                this.curBatchVar = data
                const projectId = this.projectId
                const variableId = data.id
                this.batchUpdateConf.mode = 'form'
                this.batchUpdateConf.isShow = true
                this.batchUpdateConf.title = data.name
                this.isBatchVarLoading = true

                const url = data.scope === 'namespace' ? 'variable/getNamespaceBatchVarList' : 'variable/getClusterBatchVarList'
                try {
                    const res = await this.$store.dispatch(url, { projectId, variableId })
                    this.batchVarList = res.data
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isBatchVarLoading = false
                }
            },

            /**
             * 保存所有命名空间下变量
             */
            async saveBatchVar () {
                const projectId = this.projectId
                const varId = this.curBatchVar.id
                let url = 'variable/updateNamespaceBatchVar'
                let data = {}

                // 处于文本编辑模式
                if (this.batchUpdateConf.mode === 'text') {
                    try {
                        const content = this.batchEditorConfig.editor.getValue()
                        if (!content) {
                            this.$bkMessage({
                                theme: 'error',
                                message: this.$t('请输入内容')
                            })
                            return false
                        }

                        const varList = JSON.parse(content)
                        if (!Array.isArray(varList)) {
                            this.$bkMessage({
                                theme: 'error',
                                message: this.$t('请输入合法的JSON格式')
                            })
                            return false
                        }

                        if (this.curBatchVar && this.curBatchVar.scope === 'namespace') {
                            varList.forEach(varItem => {
                                this.batchVarList.forEach(matchItem => {
                                    if (varItem.cluster_name === matchItem.cluster_name && varItem.namespace === matchItem.name) {
                                        matchItem.variable_value = varItem.value
                                    }
                                })
                            })
                        } else {
                            varList.forEach(varItem => {
                                this.batchVarList.forEach(matchItem => {
                                    if (varItem.cluster_name === matchItem.name) {
                                        matchItem.variable_value = varItem.value
                                    }
                                })
                            })
                        }
                    } catch (e) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请输入合法的JSON格式')
                        })
                        return false
                    }
                }

                if (this.curBatchVar.scope === 'namespace') {
                    data = {
                        ns_vars: {}
                    }
                    this.batchVarList.forEach(item => {
                        data.ns_vars[item.namespace_id] = item.variable_value
                    })
                } else {
                    data = {
                        cluster_vars: {}
                    }
                    this.batchVarList.forEach(item => {
                        data.cluster_vars[item.cluster_id] = item.variable_value
                    })
                    url = 'variable/updateClusterBatchVar'
                }

                try {
                    await this.$store.dispatch(url, { projectId, varId, data })
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('变量更新成功')
                    })
                    this.batchUpdateConf.isShow = false
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 编辑变量
             * @param  {Object} data 变量
             */
            editVar (data) {
                this.isSaving = false
                this.curVar = JSON.parse(JSON.stringify(data))
                this.varDialogConfig.title = this.$t('编辑变量')
                this.varDialogConfig.isShow = true
            },

            /**
             * 批量删除变量
             */
            removeVars () {
                const names = []
                const ids = []

                if (this.curAllSelectedData.length) {
                    this.curAllSelectedData.forEach(item => {
                        names.push(item.name)
                        ids.push(item.id)
                    })
                } else {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择要删除的变量')
                    })
                    return false
                }

                this.batchDialogConfig.list = names
                this.batchDialogConfig.removeIds = ids
                this.batchDialogConfig.isShow = true
            },

            /**
             * 删除变量
             *
             * @param {Object} data 变量
             */
            removeVar (data) {
                const self = this
                this.$bkInfo({
                    title: this.$t('确认删除'),
                    clsName: 'biz-remove-dialog',
                    content: this.$createElement('p', {
                        class: 'biz-confirm-desc'
                    }, `${this.$t('确定要删除变量')}【${data.name}】?`),
                    confirmFn () {
                        self.deleteVar([data.id])
                    }
                })
            },

            /**
             * 取消提交
             */
            cancelVar () {
                this.clearInput()
                this.varDialogConfig.isShow = false
            },

            /**
             * 重置当前变量数据
             */
            clearInput () {
                this.curVar = {
                    'name': '',
                    'key': '',
                    'default': {
                        'value': ''
                    },
                    'desc': '',
                    'scope': this.isSharedCluster ? 'namespace' : 'global'
                }
            },

            /**
             * 表格选中事件
             */
            handleSelectionChange (selection) {
                this.curAllSelectedData.splice(0, this.curAllSelectedData.length, ...selection)
                this.alreadySelectedNums = this.curAllSelectedData.length
            },

            /**
             * 取消选择
             */
            clearSelectedVarList () {
                this.curPageData.forEach((item) => {
                    item.isChecked = false
                })
            },

            /**
             * 初始化变量列表
             */
            async getVarList (offset, limit) {
                const projectId = this.projectId
                const keyword = this.searchKeyword
                const scope = this.searchScope
                try {
                    const res = await this.$store.dispatch('variable/getVarListByPage', { projectId, offset, limit, keyword, scope })

                    this.searchKeyWord = ''
                    this.pageConf.total = res.count
                    this.curPageData = res.results

                    const checkVariableIdList = this.curAllSelectedData.map(variable => variable.id)
                    this.curPageData.forEach(item => {
                        if (item.category !== 'sys') {
                            item.isChecked = checkVariableIdList.indexOf(item.id) > -1
                        }
                    })

                    // 当前页选中的
                    const checkedCurPageList = this.curPageData.filter(item => item.isChecked === true)
                    // 当前页合法的
                    const validList = this.curPageData.filter(item => item.category !== 'sys')
                    this.isCheckCurPageAll = validList.length === 0
                        ? false
                        : checkedCurPageList.length === validList.length

                    this.initPageConf()
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isLoading = false
                        this.isPageLoading = false
                    }, 200)
                }
            },

            /**
             * 搜索变量
             */
            searchVar () {
                this.getDataByPage()
            },

            /**
             * 初始化分页配置
             */
            initPageConf () {
                this.pageConf.totalPage = Math.ceil(this.pageConf.total / this.pageConf.pageSize)
                if (this.pageConf.curPage > this.pageConf.totalPage) {
                    this.pageConf.curPage = this.pageConf.totalPage
                }
                if (this.pageConf.curPage === 0) {
                    this.pageConf.curPage = 1
                }
            },

            /**
             * 加载当前前页数据
             */
            reloadCurPage () {
                this.getDataByPage(this.pageConf.curPage)
            },

            /**
             * 获取相应页变量数据
             *
             * @param {Number} page 页
             */
            getDataByPage (page = 1) {
                const offset = (page - 1) * this.pageConf.pageSize
                const limit = this.pageConf.pageSize
                this.isPageLoading = true
                this.getVarList(offset, limit)
            },

            /**
             * 页改变
             */
            pageChange (page = 1) {
                this.pageConf.curPage = page
                this.getDataByPage(page)
            },

            /**
             * 初始化变量引用分页配置
             */
            initQuotePageConf () {
                const total = this.quoteList.length
                this.quotePageConf.totalPage = Math.ceil(total / this.quotePageConf.pageSize)
                this.quotePageConf.count = total
            },

            /**
             * 加载变量引用当前页数据
             */
            reloadQuoteCurPage () {
                this.initQuotePageConf()
                if (this.quotePageConf.curPage > this.quotePageConf.totalPage) {
                    this.quotePageConf.curPage = this.quotePageConf.totalPage
                }
                this.curQuotePageData = this.getDataByPage(this.quotePageConf.curPage)
            },

            /**
             * 获取相应页的变量引用数据
             *
             * @param {Number} page 页
             */
            getQuoteDataByPage (page) {
                let startIndex = (page - 1) * this.quotePageConf.pageSize
                let endIndex = page * this.quotePageConf.pageSize
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.quoteList.length) {
                    endIndex = this.quoteList.length
                }
                const data = this.quoteList.slice(startIndex, endIndex)
                return data
            },

            quotePageChange (page) {
                this.quotePageConf.curPage = page
                const data = this.getQuoteDataByPage(page)
                this.curQuotePageData = JSON.parse(JSON.stringify(data))
            },

            /**
             * 检查提交的变量数据
             */
            checkData () {
                if (!this.curVar.name) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入变量名称')
                    })
                    return false
                }

                if (this.curVar.name.length > 32) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入32个字符以内的变量名称')
                    })
                    return false
                }

                if (!this.curVar.key) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入变量KEY')
                    })
                    return false
                }

                const keyReg = /^[A-Za-z][A-Za-z0-9_]{0,63}$/
                if (!keyReg.test(this.curVar.key)) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('KEY 只能包含字母、数字和下划线，且以字母开头，最大长度为64个字符')
                    })
                    return false
                }
                return true
            },

            /**
             * 提交变量数据
             */
            saveVar () {
                if (!this.checkData()) {
                    return false
                }
                if (this.curVar.id) {
                    this.updateVar()
                } else {
                    this.newVar()
                }
            },

            /**
             * 提交新增的变量
             */
            async newVar () {
                this.isSaving = true
                const projectId = this.projectId
                const data = JSON.parse(JSON.stringify(this.curVar))
                if (this.isSharedCluster) {
                    data.cluster_type = 'SHARED'
                }

                try {
                    const res = await this.$store.dispatch('variable/addVar', { projectId, data })
                    if (res.code === 0) {
                        this.$bkMessage({
                            theme: 'success',
                            message: this.$t('变量创建成功')
                        })
                        this.clearInput()
                        this.varDialogConfig.isShow = false
                        this.getDataByPage()

                        this.pageConf.curPage = 1
                        this.isCheckCurPageAll = false
                        this.curAllSelectedData.splice(0, this.curAllSelectedData.length, ...[])
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                    this.isSaving = false
                }
            },

            /**
             * 提交更新的变量
             */
            async updateVar () {
                this.isSaving = true
                const projectId = this.projectId
                const data = JSON.parse(JSON.stringify(this.curVar))
                const varId = data.id

                try {
                    const res = await this.$store.dispatch('variable/updateVar', { projectId, varId, data })
                    if (res.code === 0) {
                        this.$bkMessage({
                            theme: 'success',
                            message: this.$t('变量更新成功')
                        })
                        this.clearInput()
                        this.varDialogConfig.isShow = false
                        this.getDataByPage()

                        this.pageConf.curPage = 1
                        this.isCheckCurPageAll = false
                        this.curAllSelectedData.splice(0, this.curAllSelectedData.length, ...[])
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                    this.isSaving = false
                }
            },

            /**
             * 删除变量
             * @param {Array} ids 变量id列表
             */
            async deleteVar (ids) {
                this.isSaving = true
                this.batchDialogConfig.isShow = false
                const projectId = this.projectId
                const data = {
                    id_list: JSON.stringify(ids)
                }
                this.isPageLoading = true

                try {
                    const res = await this.$store.dispatch('variable/deleteVar', { projectId, data })
                    if (res.code === 0) {
                        this.$bkMessage({
                            theme: 'success',
                            message: this.$t('变量删除成功')
                        })
                        this.getDataByPage()

                        this.pageConf.curPage = 1
                        this.isCheckCurPageAll = false
                        this.curAllSelectedData.splice(0, this.curAllSelectedData.length, ...[])
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                    this.isPageLoading = false
                }
            },

            /**
             * 获取相应变量的引用列表
             * @param {Object} variable 变量
             */
            async getQuoteDetail (variable) {
                const projectId = this.projectId
                const varId = variable.id
                this.curVar = variable
                this.isQuoteLoading = true
                this.quoteDialogConf.isShow = true
                this.quoteList = []
                this.curQuotePageData = []

                try {
                    const res = await this.$store.dispatch('variable/getQuoteDetail', { projectId, varId })
                    this.quoteList = res.data.quote_list
                    this.curProjectData = {
                        projectId: res.data.project_id,
                        projectCode: res.data.project_code,
                        projectKind: res.data.project_kind
                    }
                    this.quotePageConf.curPage = 1
                    this.initQuotePageConf()
                    this.curQuotePageData = this.getQuoteDataByPage(this.quotePageConf.curPage)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isQuoteLoading = false
                }
            },

            /**
             * 查看变量引用详情
             *
             * @param {Object} quote 引用
             */
            async checkVarQuote (quote) {
                if (this.curProjectData) {
                    let routeName = ''
                    const type = quote.category
                    if (this.curProjectData.projectKind === 1) {
                        const k8sRoutes = {
                            'K8sDeployment': 'k8sTemplatesetDeployment',
                            'K8sService': 'k8sTemplatesetService',
                            'K8sConfigMap': 'k8sTemplatesetConfigmap',
                            'K8sSecret': 'k8sTemplatesetSecret',
                            'K8sDaemonSet': 'k8sTemplatesetDaemonset',
                            'K8sStatefulSet': 'k8sTemplatesetStatefulset',
                            'K8sJob': 'k8sTemplatesetJob',
                            'K8sIngress': 'k8sTemplatesetIngress'
                        }

                        routeName = k8sRoutes[type]
                    }
                    if (routeName) {
                        this.$router.push({
                            name: routeName,
                            params: {
                                projectId: this.curProjectData.projectId,
                                projectCode: this.curProjectData.projectCode,
                                templateId: quote.template_id
                            }
                        })
                    }
                }
            },

            hideQuoteDialog () {
                this.quoteDialogConf.isShow = false
            },

            handleFileImport () {
                this.$refs.fileInput.click()
            },

            handleFileInput (e) {
                const fileInput = this.$refs.fileInput
                const self = this
                if (fileInput.files && fileInput.files.length) {
                    const file = fileInput.files[0]
                    if (window.FileReader) {
                        const reader = new FileReader()
                        reader.onloadend = function (event) {
                            if (event.target.readyState === FileReader.DONE) {
                                console.log(event.target.result)
                                self.importContent = event.target.result
                                self.$store.dispatch('variable/importVars', {
                                    projectId: self.projectId,
                                    data: {
                                        variables: event.target.result
                                    }
                                }).then(() => {
                                    self.$bkMessage({
                                        theme: 'success',
                                        message: self.$t('批量导入成功')
                                    })
                                    self.refresh()
                                }).catch((e) => {
                                })
                            }
                        }
                        reader.readAsText(file)
                    }
                }
                e.target.value = ''
            },

            handleShowVarExample () {
                this.exampleConf.isShow = true
            },

            editorInitAfter (editor) {
                this.batchEditorConfig.editor = editor
            },

            syncBatchVarList () {
                try {
                    const content = this.batchEditorConfig.editor.getValue()
                    const varList = JSON.parse(content)
                    if (this.curBatchVar && this.curBatchVar.scope === 'namespace') {
                        varList.forEach(varItem => {
                            this.batchVarList.forEach(matchItem => {
                                if (varItem.cluster_name === matchItem.cluster_name && varItem.namespace === matchItem.name) {
                                    matchItem.variable_value = varItem.value
                                }
                            })
                        })
                    } else {
                        varList.forEach(varItem => {
                            this.batchVarList.forEach(matchItem => {
                                if (varItem.cluster_name === matchItem.name) {
                                    matchItem.variable_value = varItem.value
                                }
                            })
                        })
                    }
                } catch (e) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入合法的JSON格式')
                    })
                }
            },

            toggleBatchMode () {
                if (this.batchUpdateConf.mode === 'form') {
                    this.batchUpdateConf.mode = 'text'
                    const varList = this.batchVarList.map(item => {
                        if (this.curBatchVar && this.curBatchVar.scope === 'namespace') {
                            return {
                                cluster_name: item.cluster_name,
                                namespace: item.name,
                                value: item.variable_value
                            }
                        } else {
                            return {
                                cluster_name: item.name,
                                value: item.variable_value
                            }
                        }
                    })
                    this.batchEditorConfig.content = JSON.stringify(varList, null, 2)
                } else {
                    try {
                        // 校验数据格式
                        const content = this.batchEditorConfig.editor.getValue()
                        const varList = JSON.parse(content)
                        if (!Array.isArray(varList)) {
                            this.$bkMessage({
                                theme: 'error',
                                message: this.$t('请输入合法的JSON格式')
                            })
                            return false
                        }
                        this.batchUpdateConf.mode = 'form'
                        this.syncBatchVarList()
                    } catch (e) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请输入合法的JSON格式')
                        })
                    }
                }
            }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
