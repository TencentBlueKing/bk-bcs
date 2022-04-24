<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-tpl-title">
                {{$t('Helm Chart仓库')}}
            </div>
            <bk-guide>
                <a href="javascript:void(0);" class="bk-text-button" @click.stop.prevent="showGuide">{{$t('如何推送Helm Chart到项目仓库？')}}</a>
            </bk-guide>
        </div>

        <guide ref="clusterGuide"></guide>

        <bk-dialog
            :is-show.sync="helmDialog.isShow"
            :width="500"
            :has-footer="false"
            :title="$t('项目Chart仓库配置信息')"
            @cancel="hideHelmDialog">
            <template slot="content">
                <div class="helm-repos-detail" v-if="reposData">
                    <div class="repos-item" v-for="repo in reposData.privateRepos" :key="repo.url">
                        <div class="wrapper mb10">
                            <h2 class="label">{{$t('项目Chart仓库地址')}}：</h2>
                            <p class="url">
                                {{repo.url}}
                                <bcs-popover placement="top" :content="$t('复制')">
                                    <span :data-clipboard-text="repo.url" class="copy-btn bk-text-button bcs-icon bcs-icon-clipboard ml5" style="color: #999;"></span>
                                </bcs-popover>
                            </p>
                        </div>
                        <div class="auth" v-for="auth in repo.auths" :key="auth.credentials_decoded.username">
                            <div>username：{{auth.credentials_decoded.username}}</div>
                            <div>password：
                                <template v-if="repo.isPasswordShow">{{auth.credentials_decoded.password}}</template>
                                <template v-else>************</template>
                                <bcs-popover placement="top" :content="repo.isPasswordShow ? $t('隐藏') : $t('查看')">
                                    <span :class="['bk-text-button bcs-icon',{ 'bcs-icon-eye': !repo.isPasswordShow, 'bcs-icon-eye-slash': repo.isPasswordShow }]" @click="togglePassword(repo)" style="color: #999;"></span>
                                </bcs-popover>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
        </bk-dialog>

        <div class="biz-content-wrapper biz-tpl-wrapper" style="padding: 0; margin: 0;" v-bkloading="{ isLoading: showLoading, opacity: 0.1 }">
            <template v-if="!showLoading">
                <div class="biz-panel-header" style="padding: 20px;">
                    <div class="left">
                        <bk-button type="primary" @click="syncHelmTpl" :loading="isTplSynLoading">{{$t('同步仓库')}}</bk-button>
                        <span class="biz-tip ml5">{{$t('同步仓库中的Helm Chart')}}</span>
                        <a class="bk-text-button f12 ml10" href="javascript:void(0);" @click="getHelmDeops">{{$t('查看项目Chart仓库配置信息')}}</a>
                    </div>
                    <div class="right">
                        <div class="biz-search-input" style="width: 300px;">
                            <bk-input right-icon="bk-icon icon-search"
                                :placeholder="$t('输入关键字，按Enter搜索')"
                                clearable
                                v-model="searchKeyword"
                                @enter="search"
                                @clear="clearSearch" />
                        </div>
                    </div>
                </div>

                <app-exception
                    v-if="exceptionCode && !showLoading"
                    :type="exceptionCode.code"
                    :text="exceptionCode.msg">
                </app-exception>

                <template>
                    <svg style="display: none;">
                        <title>{{$t('模板集默认图标')}}</title>
                        <symbol id="biz-set-icon" viewBox="0 0 60 60">
                            <g id="图层_6">
                                <g id="图层_32_1_">
                                    <path class="st0" d="M12,8v4H8c-1.1,0-2,0.9-2,2v42c0,1.1,0.9,2,2,2h42c1.1,0,2-0.9,2-2v-4h4c1.1,0,2-0.9,2-2V8c0-1.1-0.9-2-2-2
                                        H14C12.9,6,12,6.9,12,8z M48,48v4v2H10V16h2h4h32V48z M54,48h-2V14c0-1.1-0.9-2-2-2H16v-2h38V48z" />
                                </g>
                                <path class="st1" d="M45.7,33.7h-1.8l-3.4-8.3l1.3-1.3c0.5-0.5,0.5-1.3,0-1.8l0,0c-0.5-0.5-1.3-0.5-1.8,0l-1.3,1.3l-8.4-3.5v-1.8
                                    c0-0.7-0.6-1.3-1.3-1.3l0,0c-0.7,0-1.3,0.6-1.3,1.3V20l-8.4,3.5l-1.2-1.2c-0.5-0.5-1.3-0.5-1.8,0l0,0c-0.5,0.5-0.5,1.3,0,1.8
                                    l1.2,1.2L14,33.7h-1.8c-0.7,0-1.3,0.6-1.3,1.3l0,0c0,0.7,0.6,1.3,1.3,1.3H14l3.5,8.4L16.2,46c-0.5,0.5-0.5,1.3,0,1.8l0,0
                                    c0.5,0.5,1.3,0.5,1.8,0l1.3-1.3l8.3,3.4v1.8c0,0.7,0.6,1.3,1.3,1.3l0,0c0.7,0,1.3-0.6,1.3-1.3v-1.9l8.3-3.4l1.3,1.3
                                    c0.5,0.5,1.3,0.5,1.8,0l0,0c0.5-0.5,0.5-1.3,0-1.8l-1.3-1.3l3.4-8.3h1.9c0.7,0,1.3-0.6,1.3-1.3l0,0C47,34.3,46.4,33.7,45.7,33.7z
                                     M30.3,23.4l6,2.5l-4.6,4.6c-0.4-0.2-0.9-0.4-1.3-0.6v-6.5H30.3z M27.7,23.4V30c-0.5,0.1-0.9,0.3-1.4,0.6l-4.7-4.7L27.7,23.4z
                                     M19.9,27.7l4.7,4.7c-0.2,0.4-0.4,0.9-0.5,1.3h-6.6L19.9,27.7z M17.4,36.3H24c0.1,0.5,0.3,0.9,0.6,1.3l-4.7,4.7L17.4,36.3z
                                     M27.7,46.5l-6-2.5l4.7-4.7c0.4,0.2,0.8,0.4,1.3,0.5V46.5z M29,37.5c-1.4,0-2.6-1.2-2.6-2.6c0-1.4,1.2-2.6,2.6-2.6s2.6,1.2,2.6,2.6
                                    C31.6,36.4,30.4,37.5,29,37.5z M30.3,46.5v-6.6c0.5-0.1,0.9-0.3,1.3-0.5l4.6,4.6L30.3,46.5z M38,42.2l-4.6-4.6
                                    c0.2-0.4,0.4-0.8,0.6-1.3h6.5L38,42.2z M34,33.7c-0.1-0.5-0.3-0.9-0.5-1.3l4.6-4.6l2.5,6H34V33.7z" />
                                <g class="st2">
                                    <path class="st3" d="M41,49H17c-1.1,0-2-0.9-2-2V23c0-1.1,0.9-2,2-2h24c1.1,0,2,0.9,2,2v24C43,48.1,42.1,49,41,49z" />
                                </g>
                                <g>
                                    <path class="st0" d="M42.2,25c-1.9,0-2.9,0.5-2.9,1.5v17.1c0,1,1,1.5,2.9,1.5v1.8H31.4V45c2,0,3-0.5,3-1.5v-8H23.6v8
                                        c0,1,1,1.5,3,1.5v1.8H15.8V45c1.9,0,2.8-0.5,2.8-1.5V26.4c0-1-0.9-1.5-2.8-1.5V23h10.8v2c-2,0-3,0.5-3,1.5v6.8h10.8v-6.8
                                        c0-1-1-1.5-3-1.5v-1.9h10.8V25z" />
                                </g>
                            </g>
                        </symbol>
                    </svg>

                    <div class="bk-tab2" style="border-left: none; border-right: none; font-size: 0;">
                        <div class="bk-tab2-head is-fill">
                            <div class="bk-tab2-nav" style="width: 100%;">
                                <div :title="$t('项目仓库')" :class="['tab2-nav-item', { 'active': tabActiveName === 'privateRepo' }]" @click="tabActiveName = 'privateRepo'" style="width: 180px;">
                                    {{$t('项目仓库')}}
                                </div>
                                <div :title="$t('公共仓库')" :class="['tab2-nav-item', { 'active': tabActiveName === 'publicRepo' }]" @click="tabActiveName = 'publicRepo'" style="width: 180px;">
                                    {{$t('公共仓库')}}
                                </div>
                            </div>
                        </div>
                        <div class="bk-tab2-content">
                            <div class="biz-namespace mt20">
                                <table class="bk-table biz-templateset-table mb20">
                                    <thead>
                                        <tr>
                                            <th style="width: 120px; padding-left: 0;" class="center">{{$t('图标')}}</th>
                                            <th style="width: 230px; padding-left: 20px;">{{$t('Helm Chart名称')}}</th>
                                            <th style="width: 120px; padding-left: 0;">{{$t('版本')}}</th>
                                            <th style="padding-left: 0;">{{$t('描述')}}</th>
                                            <th style="width: 170px; padding-left: 0;">{{$t('最近更新')}}</th>
                                            <th style="width: 200px; padding-left: 0;">{{$t('操作')}}</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <template v-if="tplList.length">
                                            <tr
                                                v-for="template in tplList"
                                                :key="template.id">
                                                <td colspan="7">
                                                    <table class="biz-inner-table">
                                                        <tr>
                                                            <td class="logo">
                                                                <div class="logo-wrapper" v-if="template.icon && isImage(template.icon)">
                                                                    <img :src="template.icon">
                                                                </div>
                                                                <svg class="biz-set-icon" v-else>
                                                                    <use xlink:href="#biz-set-icon"></use>
                                                                </svg>
                                                            </td>
                                                            <td class="data">
                                                                <bcs-popover placement="top" :delay="500">
                                                                    <p class="tpl-name">
                                                                        <router-link class="bk-text-button bk-primary bk-button-small" :to="{ name: 'helmTplDetail', params: { tplId: template.id } }">{{template.name}}</router-link>
                                                                    </p>
                                                                    <template slot="content">
                                                                        <p>{{template.name}}</p>
                                                                    </template>
                                                                </bcs-popover>
                                                            </td>
                                                            <td class="version">
                                                                <template v-if="template.defaultChartVersion">
                                                                    <bcs-popover placement="top" :delay="500">
                                                                        <p class="tpl-version">
                                                                            {{template.defaultChartVersion.version}}
                                                                        </p>
                                                                        <template slot="content">
                                                                            <p>{{template.defaultChartVersion.version}}</p>
                                                                        </template>
                                                                    </bcs-popover>
                                                                </template>
                                                                <template v-else>--</template>
                                                            </td>
                                                            <td class="description">
                                                                <p class="text">{{template.description}}</p>
                                                            </td>
                                                            <td class="update">
                                                                {{template.changed_at}}
                                                            </td>
                                                            <td class="action">
                                                                <span v-bk-tooltips="{
                                                                    placement: 'top',
                                                                    content: $t('仅允许平台部署，如有疑问请联系蓝鲸容器助手'),
                                                                    disabled: !template.annotations.only_for_platform || (tabActiveName === 'privateRepo')
                                                                }">
                                                                    <router-link class="bk-button bk-primary mr5"
                                                                        :to="template.annotations.only_for_platform ? {} : { name: 'helmTplInstance', params: { tplId: template.id } }"
                                                                        :disabled="template.annotations.only_for_platform">
                                                                        {{$t('部署')}}
                                                                    </router-link>
                                                                </span>
                                                                <bk-button v-if="tabActiveName === 'publicRepo'" theme="default" @click="handleDownloadChart(template)">{{$t('下载版本')}}</bk-button>
                                                                <bk-dropdown-menu class="dropdown-menu" :align="'right'" ref="dropdown" v-if="tabActiveName === 'privateRepo' && $INTERNAL">
                                                                    <bk-button class="bk-button bk-default btn" slot="dropdown-trigger" style="width: 82px; position: relative;">
                                                                        <span class="f14">{{$t('更多')}}</span>
                                                                        <i class="bcs-icon bcs-icon-angle-down dropdown-menu-angle-down ml0" style="font-size: 10px;"></i>
                                                                    </bk-button>
                                                                    <ul class="bk-dropdown-list" slot="dropdown-content">
                                                                        <li>
                                                                            <a href="javascript:void(0)" @click="handleRemoveChart(template)">{{$t('删除Chart')}}</a>
                                                                        </li>
                                                                        <li>
                                                                            <a href="javascript:void(0)" @click="showChooseDialog(template)">{{$t('删除版本')}}</a>
                                                                        </li>
                                                                    </ul>
                                                                </bk-dropdown-menu>
                                                            </td>
                                                        </tr>
                                                    </table>
                                                </td>
                                            </tr>
                                        </template>
                                        <template v-if="!tplList.length && !showLoading">
                                            <tr>
                                                <td colspan="6">
                                                    <div class="biz-empty-message" style="padding: 80px;">
                                                        <template v-if="isSearchMode">
                                                            <bcs-exception type="empty" scene="part"></bcs-exception>
                                                        </template>
                                                        <template v-else>
                                                            <span style="vertical-align: middle;">{{$t('无数据，请尝试')}}</span> <a href="javascript:void(0);" class="bk-text-button" @click="syncHelmTpl">{{$t('同步仓库')}}</a>
                                                        </template>
                                                    </div>
                                                </td>
                                            </tr>
                                        </template>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                </template>
            </template>
        </div>

        <bk-dialog
            :is-show.sync="delInstanceDialogConf.isShow"
            :width="delInstanceDialogConf.width"
            :title="delInstanceDialogConf.title"
            :quick-close="false"
            :ext-cls="'biz-config-templateset-del-instance-dialog'"
            @cancel="delInstanceDialogConf.isShow = false">
            <template slot="content" v-bkloading="{ isLoading: isDeleting }">
                <div class="content-inner">
                    <div class="bk-form bk-form-vertical" style="margin-bottom: 20px;">
                        <div class="bk-form-item">
                            <label class="bk-label fl">
                                {{$t('选择要删除的版本')}}:
                            </label>
                            <div class="bk-form-content">
                                <div class="bk-dropdown-box" style="width: 300px;">
                                    <bcs-select v-if="delInstanceDialogConf.isShow"
                                        v-model="delInstanceDialogConf.versionIds"
                                        :multi-select="true"
                                        searchable
                                        multiple
                                        show-select-all
                                        :placeholder="$t('请选择')">
                                        <bk-option v-for="item in delInstanceDialogConf.versions"
                                            :key="item.id"
                                            :name="item.version"
                                            :id="item.id">
                                        </bk-option>
                                    </bcs-select>
                                </div>
                            </div>
                        </div>
                    </div>
                    <template v-if="delInstanceDialogConf && delInstanceDialogConf.releases.length && delInstanceDialogConf.versionIds.length">
                        <p style="font-weight: bold; color: #737987; font-size: 14px; text-align: left;">
                            {{$t('当前版本含有的Release:')}} <span class="biz-tip" style="font-weight: normal;">({{$t('格式：命名空间:Release名称')}})</span>
                        </p>
                        <ul class="key-list mt10 mb10">
                            <li v-for="release of delInstanceDialogConf.releases" :key="release.name">
                                <span class="key">{{release.namespace}}</span>
                                <span class="value">{{release.name}}</span>
                            </li>
                        </ul>
                        <div>
                            {{$t('您需要先删除所有Release，再进行版本删除操作')}}
                        </div>
                    </template>
                </div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="delInstanceDialogConf.versionIds.length && !delInstanceDialogConf.releases.length">
                        <bk-button type="primary" :loading="isVersionDeleting" class="bk-dialog-btn bk-dialog-btn-confirm"
                            @click="confirmDelVersion">
                            {{$t('提交')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" class="bk-dialog-btn bk-dialog-btn-confirm" disabled>
                            {{$t('提交')}}
                        </bk-button>
                    </template>

                    <bk-button type="button" :disabled="isVersionDeleting" class="bk-dialog-btn bk-dialog-btn-cancel" @click="cancelDelVersion">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="delTemplateDialogConf.isShow"
            :width="delTemplateDialogConf.width"
            :ext-cls="'biz-config-templateset-copy-dialog'"
            :has-header="false"
            :quick-close="false"
            @cancel="delTemplateDialogConf.isShow = false">
            <template slot="content" style="padding: 0 20px;">
                <div style="color: #333; font-size: 20px">
                    Chart【{{delTemplateDialogConf.title}}】{{$t('包含以下Releases：')}}
                    <span class="biz-tip">({{$t('格式：命名空间:Release名称')}})</span>
                </div>
                <ul class="key-list mt20 mb5">
                    <li v-for="release of delTemplateDialogConf.releases" :key="release.name">
                        <span class="key">{{release.namespace}}</span>
                        <span class="value">{{release.name}}</span>
                    </li>
                </ul>
                <div style="clear: both; margin-bottom: 20px;">
                    {{$t('您需要先删除所有Release，再进行Chart删除操作')}}
                </div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <bk-button type="primary" @click="delTemplateCancel">
                        {{$t('关闭')}}
                    </bk-button>
                </div>
            </div>
        </bk-dialog>

        <bcs-dialog
            v-model="downloadDialog.isShow"
            header-position="left"
            :title="$t('下载Chart版本')"
            :width="550"
            :mask-close="false">
            <bcs-form form-type="vertical">
                <bcs-form-item :label="$t('选择要下载的版本')">
                    <bcs-select v-model="downloadDialog.downloadVersion"
                        :loading="isTplVersionLoading"
                        :clearable="false"
                        @change="handleSelectVersion">
                        <bcs-option v-for="item in downloadDialog.versions"
                            :key="item.version"
                            :id="item.version"
                            :name="item.version">
                        </bcs-option>
                    </bcs-select>
                </bcs-form-item>
            </bcs-form>
            <template slot="footer">
                <bcs-button theme="primary"
                    :disabled="!downloadDialog.downloadVersion"
                    :loading="isVersionDetailLoading"
                    @click="handleComfirmDownload">
                    {{$t('确定')}}
                </bcs-button>
                <bcs-button @click="handleCancelDownload">{{$t('取消')}}</bcs-button>
            </template>
        </bcs-dialog>

    </div>
</template>

<script>
    import Guide from './guide'
    import Clipboard from 'clipboard'
    import { catchErrorHandler } from '@/common/util'

    export default {
        components: {
            Guide
        },
        data () {
            return {
                tplList: [],
                publicTplList: [],
                privateTplList: [],
                tplListCache: [],
                showLoading: false,
                exceptionCode: null,
                searchKeyword: '',
                isSearchMode: false,
                curProjectId: '',
                isTplSynLoading: false,
                isRepoDataLoading: false,
                tabActiveName: 'privateRepo',
                delTemplateDialogConf: {
                    isShow: false,
                    width: 650,
                    title: '',
                    closeIcon: false,
                    template: {},
                    releases: []
                },
                isReleaseLoading: false,
                delInstanceDialogConf: {
                    isShow: false,
                    width: 550,
                    title: '',
                    closeIcon: false,
                    template: {},
                    versions: [],
                    releases: [],
                    versionIds: []
                },
                reposData: {
                    publicRepos: [],
                    privateRepos: []
                },
                helmDialog: {
                    isShow: false
                },
                deleteVersionTimer: null,
                downloadDialog: {
                    isShow: false,
                    downloadVersion: '',
                    chartName: '',
                    versions: []
                },
                isTplVersionLoading: false,
                isVersionDetailLoading: false
            }
        },
        computed: {
            curProject () {
                return this.$store.state.curProject
            },
            projectId () {
                this.curProjectId = this.$route.params.projectId
                return this.curProjectId
            },
            projectCode () {
                return this.$route.params.projectCode
            }
        },
        watch: {
            searchKeyword (newVal, oldVal) {
                // 如果删除，为空时触发搜索
                if (oldVal && !newVal) {
                    this.search()
                }
            },
            curProjectId () {
                // 如果不是k8s类型的项目，无法访问些页面，重定向回集群首页
                if (this.curProject && (this.curProject.kind !== PROJECT_K8S && this.curProject.kind !== PROJECT_TKE)) {
                    this.$router.push({
                        name: 'clusterMain',
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode
                        }
                    })
                }
            },
            tabActiveName (val) {
                this.setTplList()
            },
            'delInstanceDialogConf.versionIds' () {
                if (this.delInstanceDialogConf.isShow) {
                    this.delInstanceDialogConf.releases = []
                    if (this.deleteVersionTimer) {
                        clearTimeout(this.deleteVersionTimer)
                        this.deleteVersionTimer = null
                    }
                    this.deleteVersionTimer = setTimeout(this.getReleaseByVersion, 300)
                }
            }
        },
        mounted () {
            this.getTplList()
        },
        methods: {
            /**
             * 显示/隐藏模板仓库密码
             * @param  {object} repo 模板仓库
             */
            togglePassword (repo) {
                repo.isPasswordShow = !repo.isPasswordShow
            },

            /**
             * 显示引导层(如何推送Helm Chart到项目仓库？)
             */
            showGuide () {
                this.$refs.clusterGuide.show()
            },

            /**
             * 获取集群对应的helm仓库信息
             * @param  {object} cluster 集群
             */
            async getHelmDeops (cluster) {
                this.reposData = {
                    publicRepos: [],
                    privateRepos: []
                }
                this.isRepoDataLoading = true

                try {
                    const res = await this.$store.dispatch('helm/getHelmDeops', {
                        projectId: this.projectId,
                        clusterId: cluster.cluster_id
                    })

                    const repos = res.data.results
                    repos.forEach(item => {
                        item.isPasswordShow = false
                        // 区分私有和公有
                        if (item.name === 'public-repo') {
                            this.reposData.publicRepos.push(item)
                        } else {
                            this.reposData.privateRepos.push(item)
                        }
                    })
                    this.helmDialog.isShow = true

                    setTimeout(() => {
                        this.clipboardInstance = new Clipboard('.copy-btn')
                        this.clipboardInstance.on('success', e => {
                            this.$bkMessage({
                                limit: 1,
                                theme: 'success',
                                message: this.$t('复制成功')
                            })
                        })
                    }, 2000)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isRepoDataLoading = false
                }
            },

            /**
             * 隐藏helm仓库信息
             */
            hideHelmDialog () {
                this.helmDialog.isShow = false
            },

            /**
             * 同步仓库
             */
            async syncHelmTpl () {
                if (this.isTplSynLoading) {
                    return false
                }

                this.isTplSynLoading = true
                try {
                    await this.$store.dispatch('helm/syncHelmTpl', { projectId: this.projectId })

                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('同步成功')
                    })
                    this.getTplList()
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isTplSynLoading = false
                }
            },

            /**
             * 简单判断是否为图片
             * @param  {string} img 图片url
             * @return {Boolean} true/false
             */
            isImage (img) {
                if (!img) {
                    return false
                }
                if (img.startsWith('http://') || img.startsWith('https://') || img.startsWith('data:image/')) {
                    return true
                }
                return false
            },

            /**
             * 获取模板列表
             */
            async getTplList () {
                const projectId = this.projectId
                this.showLoading = true

                try {
                    const res = await this.$store.dispatch('helm/getTplList', projectId)

                    const publicRepo = []
                    const privateRepo = []

                    // 进行分类，包括项目仓库和私有仓库
                    const tplList = res.data.filter(item => {
                        return item.defaultChartVersion
                    })

                    tplList.forEach(item => {
                        if (item.repository.name === 'public-repo') {
                            publicRepo.push(item)
                        } else {
                            privateRepo.push(item)
                        }
                    })
                    this.publicTplList = publicRepo
                    this.privateTplList = privateRepo
                    this.setTplList()
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.showLoading = false
                }
            },

            /**
             * 根据当前显示公有/私有模板
             */
            setTplList () {
                this.clearSearch()
                if (this.tabActiveName === 'publicRepo') {
                    this.tplList = JSON.parse(JSON.stringify(this.publicTplList))
                } else {
                    this.tplList = JSON.parse(JSON.stringify(this.privateTplList))
                }
                this.tplListCache = JSON.parse(JSON.stringify(this.tplList))
            },

            /**
             * 搜索
             */
            search () {
                const keyword = this.searchKeyword
                if (keyword) {
                    const results = this.tplListCache.filter(item => {
                        if (item.name.indexOf(keyword) > -1) {
                            return true
                        } else {
                            return false
                        }
                    })
                    this.tplList.splice(0, this.tplList.length, ...results)
                    this.isSearchMode = true
                } else {
                    this.tplList.splice(0, this.tplList.length, ...this.tplListCache)
                    this.isSearchMode = false
                }
            },

            /**
             * 清除搜索
             */
            clearSearch () {
                this.searchKeyword = ''
                this.search()
            },

            async handleRemoveChart (template) {
                try {
                    // 先检测当前Chart是否有release
                    const res = await this.$store.dispatch('helm/getExistReleases', {
                        projectId: this.projectId,
                        chartName: template.name
                    })

                    // 如果没有release，可删除
                    if (!res.data.length) {
                        const me = this
                        me.$bkInfo({
                            title: this.$t('确认删除'),
                            clsName: 'biz-remove-dialog',
                            content: me.$createElement('p', {
                                class: 'biz-confirm-desc'
                            }, `${this.$t('确定要删除Chart')}【${template.name}】?`),
                            async confirmFn () {
                                me.deleteTemplate(template)
                            }
                        })
                    } else {
                        this.delTemplateDialogConf.isShow = true
                        this.delTemplateDialogConf.title = template.name
                        this.delTemplateDialogConf.template = Object.assign({}, template)
                        this.delTemplateDialogConf.releases = res.data
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 确认删除Chart
             * @param {Object} template 当前模板集对象
             */
            async deleteTemplate (template) {
                try {
                    await this.$store.dispatch('helm/removeTemplate', {
                        projectId: this.projectId,
                        chartName: template.name
                    })

                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('删除成功')
                    })
                    this.getTplList()
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 确认删除版本
             * @param {Object} template 当前模板集对象
             */
            async deleteTemplateVersion () {
                const versions = this.delInstanceDialogConf.versions.filter(item => this.delInstanceDialogConf.versionIds.includes(item.id))
                this.isVersionDeleting = true
                try {
                    await this.$store.dispatch('helm/removeTemplate', {
                        projectId: this.projectId,
                        chartName: this.delInstanceDialogConf.template.name,
                        versions: versions.map(item => item.version)
                    })

                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('删除成功')
                    })

                    this.getTplList()
                    this.cancelDelVersion()
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isVersionDeleting = false
                }
            },

            showTemplateVersion () {
                const template = Object.assign({}, this.delTemplateDialogConf.template)
                this.delTemplateDialogConf.isShow = false
                this.delTemplateDialogConf.title = ''
                this.delTemplateDialogConf.template = Object.assign({}, {})
                this.delTemplateDialogConf.releases = []
                this.showChooseDialog(template)
            },

            delTemplateCancel () {
                this.delTemplateDialogConf.isShow = false
                this.delTemplateDialogConf.title = ''
                this.delTemplateDialogConf.template = Object.assign({}, {})
                this.delTemplateDialogConf.releases = []
            },

            async showChooseDialog (template) {
                // 之前没选择过，那么展开第一个
                this.delInstanceDialogConf.title = `${this.$t('删除')}【${template.name}】${this.$t('Chart的版本')}`
                this.delInstanceDialogConf.isShow = true
                this.delInstanceDialogConf.template = Object.assign({}, template)
                this.delInstanceDialogConf.versions = []
                this.delInstanceDialogConf.releases = []
                this.delInstanceDialogConf.versionIds = []
                this.getTplVersions()
            },

            async getTplVersions () {
                const projectId = this.projectId

                try {
                    if (!this.$INTERNAL) {
                        const tplId = this.delInstanceDialogConf.template.name
                        const res = await this.$store.dispatch('helm/getTplVersionList', {
                            projectId,
                            tplId,
                            isPublic: this.tabActiveName === 'publicRepo'
                        })
                        this.delInstanceDialogConf.versions = res.data
                        this.delInstanceDialogConf.releases = []
                    } else {
                        const tplId = this.delInstanceDialogConf.template.id
                        const res = await this.$store.dispatch('helm/getTplVersions', { projectId, tplId })
                        this.delInstanceDialogConf.versions = res.data.results
                        this.delInstanceDialogConf.releases = []
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            async getReleaseByVersion () {
                const versions = this.delInstanceDialogConf.versions.filter(item => this.delInstanceDialogConf.versionIds.includes(item.id))
                this.isReleaseLoading = true
                try {
                    const res = await this.$store.dispatch('helm/getExistReleases', {
                        projectId: this.projectId,
                        chartName: this.delInstanceDialogConf.template.name,
                        versions: versions.map(item => item.version)
                    })

                    this.delInstanceDialogConf.releases = res.data
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isReleaseLoading = false
                }
            },

            cancelDelVersion () {
                // 之前没选择过，那么展开第一个
                this.delInstanceDialogConf.title = ''
                this.delInstanceDialogConf.isShow = false
                this.delInstanceDialogConf.template = {}
                this.delInstanceDialogConf.versions = []
                this.delInstanceDialogConf.releases = []
                this.delInstanceDialogConf.versionIds = []
            },

            /**
             * 删除命名空间弹层确认
             */
            async confirmDelVersion () {
                const me = this
                if (!this.delInstanceDialogConf.versionIds.length) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择Chart版本')
                    })
                    return
                }
                const versions = this.delInstanceDialogConf.versions.filter(item => this.delInstanceDialogConf.versionIds.includes(item.id))
                this.$bkInfo({
                    title: this.$t('确认删除以下版本'),
                    content: versions.map(item => item.version).join(', '),
                    async confirmFn () {
                        me.deleteTemplateVersion()
                    }
                })
            },

            async getChartVersionDetail (payload) {
                const { chartId, chartName, downloadVersion, downloadVersionId } = payload
                this.isVersionDetailLoading = true
                let url = ''
                try {
                    const fnPath = this.$INTERNAL ? 'helm/getChartVersionDetail' : 'helm/getChartByVersion'
                    const isPublic = this.$INTERNAL ? this.tabActiveName === 'publicRepo' : undefined
                    const res = await this.$store.dispatch(fnPath, {
                        projectId: this.projectId,
                        chartId: this.$INTERNAL ? chartName : chartId,
                        version: this.$INTERNAL ? downloadVersion : downloadVersionId,
                        isPublic
                    })
                    const data = res.data || {}
                    const urls = (data.data || {}).urls || []
                    url = urls[0] || ''
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isVersionDetailLoading = false
                }
                return url
            },

            async getTplVersionList (template) {
                this.isTplVersionLoading = true
                try {
                    if (this.$INTERNAL) {
                        // 内部版本
                        const res = await this.$store.dispatch('helm/getTplVersionList', {
                            projectId: this.projectId,
                            isPublic: this.tabActiveName === 'publicRepo',
                            tplId: template.name
                        })
                        this.downloadDialog.versions = res.data || []
                    } else {
                        // 外部版本
                        const res = await this.$store.dispatch('helm/getTplVersions', {
                            projectId: this.projectId,
                            tplId: template.id
                        })
                        this.downloadDialog.versions = res.data.results || []
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isTplVersionLoading = false
                }
            },

            async handleDownloadChart (template) {
                this.downloadDialog.downloadVersion = ''
                this.downloadDialog.versions = []
                this.downloadDialog.chartName = template.name
                this.downloadDialog.chartId = template.id
                this.downloadDialog.isShow = true
                await this.getTplVersionList(template)
            },

            handleSelectVersion (version) {
                const curVersionData = this.downloadDialog.versions.find(item => item.version === version)
                this.downloadDialog.downloadVersionId = curVersionData.id
            },

            async handleComfirmDownload () {
                const url = await this.getChartVersionDetail(this.downloadDialog)
                const a = document.createElement('a')
                a.href = url
                a.click()
                this.handleCancelDownload()
            },

            handleCancelDownload () {
                this.downloadDialog.isShow = false
                this.downloadDialog.versions = []
                this.downloadDialog.chartName = ''
                this.downloadDialog.downloadVersion = ''
            }
        }
    }
</script>

<style scoped>
    @import './tpl-list.css';
</style>
