<template>
    <div class="tplList-card">
        <template v-if="tplList.length">
            <div
                class="card-item"
                v-for="(template, index) in tplList"
                :key="index"
            >
                <div class="card-item-layer">
                    <router-link class="bk-button bk-primary mr5"
                        :to="{ name: 'helmTplInstance', params: { chartName: template.name } }">
                        {{$t('部署')}}
                    </router-link>
                    <!-- <span v-bk-tooltips="{
                        placement: 'top',
                        content: $t('仅允许平台部署，如有疑问请联系蓝鲸容器助手'),
                        disabled: !template.annotations.only_for_platform || (selectedName === 'privateRepo')
                    }">
                        <router-link class="bk-button bk-primary mr5"
                            :to="template.annotations.only_for_platform ? {} : { name: 'helmTplInstance', params: { tplId: template.id, chartName: template.name } }"
                            :disabled="template.annotations.only_for_platform">
                            {{$t('部署')}}
                        </router-link>
                    </span> -->
                    <bk-button v-if="selectedName === 'publicRepo'" theme="default" class="ml5" @click="handleDownloadChart(template)">{{$t('下载版本')}}</bk-button>
                    <bk-dropdown-menu class="dropdown-menu" :align="'right'" style="height: 32px;" ref="dropdown" v-if="selectedName === 'privateRepo' && !$INTERNAL">
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
                </div>
                <div class="tpl-info">
                    <div class="left">
                        <div class="logo-wrapper" v-if="template.icon && isImage(template.icon)">
                            <img :src="template.icon">
                        </div>
                        <svg class="biz-set-icon" v-else>
                            <use xlink:href="#biz-set-icon"></use>
                        </svg>
                    </div>
                    <div class="right">
                        <div class="tpl-name">
                            {{ template.name }}
                        </div>
                        <div class="tpl-version">
                            Version {{ template.defaultChartVersion.version }}
                        </div>
                    </div>
                </div>
                <div class="tpl-desc" :title="template.description">
                    {{ template.description || '--' }}
                </div>
            </div>
        </template>
        <template v-else>
            <div class="biz-empty-message" style="padding: 80px; margin: auto;">
                <template v-if="isSearchMode">
                    <bcs-exception type="empty" scene="part"></bcs-exception>
                </template>
                <template v-else>
                    <span style="vertical-align: middle;">{{$t('无数据，请尝试')}}</span> <a href="javascript:void(0);" class="bk-text-button" @click="syncHelmTpl">{{$t('同步仓库')}}</a>
                </template>
            </div>
        </template>
        <bcs-dialog
            v-model="downloadDialog.isShow"
            header-position="left"
            :title="$t('下载Chart版本')"
            :width="550"
            :mask-close="false">
            <bcs-form form-type="vertical">
                <bcs-form-item :label="$t('选择要下载的版本')">
                    <bcs-select v-model="downloadDialog.downloadVersion"
                        :loading="isVersionLoading"
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
                <bcs-button
                    theme="primary"
                    :disabled="!downloadDialog.downloadVersion"
                    :loading="isVersionDetailLoading"
                    @click="handleComfirmDownload"
                >
                    {{$t('确定')}}
                </bcs-button>
                <bcs-button @click="handleCancelDownload">{{$t('取消')}}</bcs-button>
            </template>
        </bcs-dialog>

        <bk-dialog
            :is-show.sync="delTemplateDialogConf.isShow"
            :width="delTemplateDialogConf.width"
            :ext-cls="'biz-config-templateset-copy-dialog'"
            :has-header="false"
            :quick-close="false">
            <template v-if="delTemplateDialogConf.canDeleted" slot="content" style="padding: 0 20px;">
                <div style="color: #333; font-size: 20px">
                    {{ delTemplateDialogConf.title }}
                </div>
                <div style="clear: both; margin: 20px 0 10px;">
                    {{ `${$i18n.t('确认要删除')} ${delTemplateDialogConf.name}?` }}
                </div>
            </template>
            <template v-else slot="content" style="padding: 0 20px;">
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
                    <bk-button v-if="delTemplateDialogConf.canDeleted" type="primary" @click="handleDeleteTemplate">
                        {{$t('确定')}}
                    </bk-button>
                    <bk-button @click="delTemplateCancel">
                        {{$t('关闭')}}
                    </bk-button>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="delInstanceDialogConf.isShow"
            :width="delInstanceDialogConf.width"
            :title="delInstanceDialogConf.title"
            :quick-close="false"
            :ext-cls="'biz-config-templateset-del-instance-dialog'"
            @cancel="delInstanceDialogConf.isShow = false">
            <template slot="content">
                <div class="content-inner">
                    <div class="bk-form bk-form-vertical" style="margin-bottom: 20px;">
                        <div class="bk-form-item">
                            <label class="bk-label fl">
                                {{$t('选择要删除的版本')}}:
                            </label>
                            <div class="bk-form-content">
                                <div class="bk-dropdown-box" style="width: 100%;">
                                    <bcs-select
                                        v-if="delInstanceDialogConf.isShow"
                                        v-model="delInstanceDialogConf.versionIds"
                                        :loading="isVersionLoading"
                                        :multi-select="true"
                                        searchable
                                        multiple
                                        show-select-all
                                        :placeholder="$t('请选择')">
                                        <bk-option v-for="item in delInstanceDialogConf.versions"
                                            :key="item.version"
                                            :name="item.version"
                                            :id="item.version">
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
    </div>
</template>

<script>
    import tplListMixin from './tpl-list-mixin'

    export default {
        name: 'tplListCard',
        components: {
            
        },
        mixins: [tplListMixin]
    }
</script>

<style lang="postcss" scoped>
    .tplList-card {
        display: flex;
        flex-wrap: wrap;
        .card-item {
            position: relative;
            width: 254px;
            height: 144px;
            padding: 20px 20px 10px;
            background-color: #fff;
            margin-left: 20px;
            margin-bottom: 20px;
            border-radius: 4px;
            &:hover {
                .card-item-layer {
                    display: inline-flex;
                }
            }
        }
        .card-item-layer {
            display: flex;
            justify-content: center;
            display: none;
            align-items: center;
            width: 254px;
            height: 144px;
            position: absolute;
            right: 0;
            top: 0;
            background-color: rgba(0,0,0,.4);
        }
    }
    .biz-set-icon {
        width: 60px;
        height: 60px;
        border-radius: 6px;
    }
    .tpl-info {
        opacity: 1;
        display: flex;
        .logo-wrapper {
            width: 60px;
            height: 60px;
            img {
                width: 100%;
            }
        }
        .right {
            margin-left: 10px;
            .tpl-name {
                font-size: 14px;
                font-weight: 600;
                color: #222222;
            }
            .tpl-version {
                color: #666666;
                font-size: 12px;
                margin-top: 10px;
            }
        }
    }
    .tpl-desc {
        color: #666666;
        font-size: 12px;
        padding-top: 10px;
        display: -webkit-box;
        overflow: hidden;
        white-space: normal !important;
        text-overflow: ellipsis;
        word-wrap: break-word;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
    }
</style>
