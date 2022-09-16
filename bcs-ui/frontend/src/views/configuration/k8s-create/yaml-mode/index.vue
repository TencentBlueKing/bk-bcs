<template>
    <div class="biz-content">
        <section class="biz-top-bar" :style="{ marginBottom: isNewTemplate ? '0px' : '70px' }">
            <i class="biz-back bcs-icon bcs-icon-arrows-left" @click="handleBeforeLeave"></i>
            <div class="biz-templateset-title">
                <span v-show="!isEditName">{{curTemplate.name}}</span>
                <input
                    type="text"
                    :placeholder="$t('30个以内的字符')"
                    maxlength="30"
                    class="bk-form-input"
                    v-model="editTemplate.name"
                    v-if="isEditName"
                    ref="templateNameInput"
                    @blur="saveTemplate"
                    @keyup.enter="saveTemplate" />
                <a href="javascript:void(0)" class="bk-text-button bk-default" v-show="!isEditName" @click="editTemplateName">
                    <i class="bcs-icon bcs-icon-edit"></i>
                </a>
            </div>
            <div class="biz-templateset-desc">
                <span v-show="!isEditDesc">{{curTemplate.desc}}</span>
                <input
                    type="text"
                    :placeholder="$t('50个以内的字符')"
                    maxlength="50"
                    class="bk-form-input"
                    v-model="editTemplate.desc"
                    v-if="isEditDesc"
                    ref="templateDescInput"
                    @blur="saveTemplate"
                    @keyup.enter="saveTemplate" />
                <a href="javascript:void(0)" class="bk-text-button bk-default" v-show="!isEditDesc" @click="editTemplateDesc" @keyup.enter="saveTemplate">
                    <i class="bcs-icon bcs-icon-edit"></i>
                </a>
            </div>
            <div class="biz-templateset-action">
                <!-- 如果不是新增状态的模板集并且有权限编辑才可查看加锁状态 -->
                <template v-if="String(templateId) !== '0'">
                    <template v-if="templateLockStatus.isLocked">
                        <template v-if="templateLockStatus.isCurLocker">
                            <div class="biz-lock-box">
                                <div class="lock-wrapper warning">
                                    <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                                    <strong class="desc">
                                        {{$t('您已经对此模板集加锁，只有解锁后，其他用户才可操作此模板集。')}}
                                        <span v-if="lateShowVersionName">
                                            （{{$t('当前版本号')}}：{{lateShowVersionName}}
                                            <bcs-popover
                                                :delay="300"
                                                :content="displayVersionNotes || '--'"
                                                style="padding-left: 6px;"
                                                placement="bottom">
                                                <span style="color: #3a84ff;">{{$t('版本说明')}}</span>
                                            </bcs-popover>）
                                        </span>
                                    </strong>
                                    <div class="action" @click="updateTemplateLockStatus">
                                        <bk-switcher
                                            :selected="templateLockStatus.isLocked"
                                            size="small">
                                        </bk-switcher>
                                    </div>
                                </div>
                            </div>
                        </template>
                        <template v-else>
                            <div class="biz-lock-box">
                                <div class="lock-wrapper warning">
                                    <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                                    <strong class="desc">
                                        {{$t('{locker}正在操作，您如需编辑请联系{locker}解锁。', templateLockStatus)}}
                                        <span v-if="lateShowVersionName">
                                            （{{$t('当前版本号')}}：{{lateShowVersionName}}
                                            <bcs-popover
                                                :delay="300"
                                                :content="displayVersionNotes || '--'"
                                                style="padding-left: 6px;"
                                                placement="bottom">
                                                <span style="color: #3a84ff;">{{$t('版本说明')}}</span>
                                            </bcs-popover>）
                                        </span>
                                    </strong>
                                    <div class="action">
                                        <a href="javascript: void(0);" class="bk-text-button" @click="reloadTemplateLockStatus">{{$t('点击刷新')}}</a>
                                    </div>
                                </div>
                            </div>
                        </template>
                    </template>
                    <template v-else>
                        <div class="biz-lock-box">
                            <div class="lock-wrapper">
                                <i class="bcs-icon bcs-icon-info-circle-shape"></i>
                                <strong class="desc">
                                    {{$t('为避免多成员同时编辑，引起内容或版本冲突，建议在编辑时，开启保护功能。')}}
                                    <span v-if="lateShowVersionName">
                                        （{{$t('当前版本号')}}：{{lateShowVersionName}}
                                        <bcs-popover
                                            :delay="300"
                                            :content="displayVersionNotes || '--'"
                                            style="padding-left: 6px;"
                                            placement="bottom">
                                            <span style="color: #3a84ff;">{{$t('版本说明')}}</span>
                                        </bcs-popover>）
                                    </span>
                                </strong>
                                <div class="action" @click="updateTemplateLockStatus">
                                    <bk-switcher
                                        :selected="templateLockStatus.isLocked"
                                        size="small">
                                    </bk-switcher>
                                </div>
                            </div>
                        </div>
                    </template>
                </template>

                <!-- 如果模板集没有加锁或者当前用户是加锁者才可以操作 -->
                <template v-if="templateLockStatus.isLocked && !templateLockStatus.isCurLocker">
                    <bk-button type="primary" disabled>{{$t('保存')}}</bk-button>
                </template>
                <template v-else>
                    <bk-button type="primary"
                        v-authority="{
                            actionId: isNewTemplate ? 'templateset_create' : 'templateset_update',
                            resourceName: curTemplate.name,
                            permCtx: {
                                resource_type: isNewTemplate ? 'project' : 'templateset',
                                project_id: projectId,
                                template_id: isNewTemplate ? undefined : Number(curTemplate.id)
                            }
                        }"
                        @click="handleSaveTemplate"
                    >{{$t('保存')}}</bk-button>
                </template>

                <bk-button :disabled="templateId === 0"
                    v-authority="{
                        clickable: isNewTemplate ? true : getAuthority('templateset_instantiate', Number(curTemplate.id)),
                        actionId: 'templateset_instantiate',
                        resourceName: curTemplate.name,
                        disablePerms: true,
                        permCtx: {
                            project_id: projectId,
                            template_id: Number(curTemplate.id)
                        }
                    }"
                    @click="createInstance">
                    {{$t('实例化')}}
                </bk-button>

                <bk-button :disabled="!allVersionList.length"
                    v-authority="{
                        clickable: isNewTemplate ? true : getAuthority('templateset_view', Number(curTemplate.id)),
                        actionId: 'templateset_view',
                        resourceName: curTemplate.name,
                        disablePerms: true,
                        permCtx: {
                            project_id: projectId,
                            template_id: Number(curTemplate.id)
                        }
                    }"
                    @click="showVersionPanel"
                >{{$t('版本列表')}}</bk-button>
            </div>
        </section>

        <p class="biz-tip m20 mb15">
            {{$t('YAML中资源所属的命名空间不需要用户指定，由平台根据用户实例化时的选择自动生成')}}
        </p>

        <section class="biz-yaml-content" v-bkloading="{ isLoading: isYamlTemplateLoading || isTemplateLocking, opacity: 0.3 }">
            <resizer :class="['resize-layout fl']"
                direction="right"
                :handler-offset="3"
                :min="300"
                :max="500"
                @resize="reRenderEditor++">
                <div class="tree-box">
                    <div class="biz-yaml-resources">
                        <ul class="yaml-tab">
                            <li :class="{ 'active': tabName === 'default' }" @click="handleToggleTab('default')">
                                {{$t('常用Manifest')}}
                            </li>
                            <li :class="{ 'active': tabName === 'custom' }" @click="handleToggleTab('custom')">
                                {{$t('自定义Manifest')}}
                            </li>
                        </ul>

                        <div class="tree-box">
                            <div>
                                <bcs-tree
                                    class="default-tree mt20"
                                    ref="defaultTree"
                                    v-show="tabName === 'default'"
                                    :data="defaultTreeData"
                                    :node-key="'id'"
                                    :tpl="renderDefaultTree"
                                    :has-border="true">
                                </bcs-tree>

                                <bcs-tree
                                    class="custom-tree mt20"
                                    ref="customTree"
                                    v-show="tabName === 'custom'"
                                    :data="customTreeData"
                                    :node-key="'id'"
                                    :tpl="renderCustomTree"
                                    :has-border="true">
                                </bcs-tree>
                            </div>
                        </div>
                    </div>
                </div>
            </resizer>

            <div class="biz-yaml-editor">
                <div class="yaml-header">
                    <strong class="title" v-bk-tooltips="curTreeNode.value.fullName">
                        {{curTreeNode.value.fullName}}
                    </strong>
                    <div class="yaml-header-action">
                        <button
                            v-if="curTreeNode.value.content !== curTreeNode.value.originContent || useEditorDiff"
                            class="biz-template-btn"
                            @click="toggleCompare">
                            {{ useEditorDiff ? $t('返回编辑') : $t('修改对比') }}
                        </button>
                        <button class="biz-template-btn primary" :key="fileImportIndex" v-bk-tooltips="{ width: 400, content: zipTooltipText }">
                            <i class="bcs-icon bcs-icon-upload"></i>
                            {{$t('导入')}}
                            <input ref="fileInput" type="file" name="upload" class="file-input" accept="application/zip,application/x-zip,application/x-zip-compressed" @change="handleFileInput(false)">
                        </button>
                        <template v-if="canTemplateExport">
                            <button class="biz-template-btn" v-bk-tooltips="$t('请先保存模板集版本再导出')" @click.stop.prevent="handleExport(curTemplate)"><i class="bcs-icon bcs-icon-download"></i>{{$t('导出')}}</button>
                        </template>
                        <template v-else>
                            <button class="biz-template-btn disabled" v-bk-tooltips="$t('请先保存模板集版本再导出')"><i class="bcs-icon bcs-icon-download"></i>{{$t('导出')}}</button>
                        </template>
                        <button class="biz-template-btn" @click.stop.prevent="handleToggleVarPanel">{{$t('变量列表')}}</button>
                        <button class="biz-template-btn" @click.stop.prevent="handleToggleImagePanel">{{$t('镜像查询')}}</button>
                    </div>
                </div>
                <div class="yaml-content">
                    <template v-if="curTreeNode.value.id">
                        <monaco-editor
                            ref="yamlEditor"
                            class="editor"
                            theme="monokai"
                            language="yaml"
                            :style="{ height: `${editorHeight}px`, width: '100%' }"
                            v-model="curTreeNode.value.content"
                            :diff-editor="useEditorDiff"
                            :options="yamlEditorOptions"
                            :key="reRenderEditor"
                            :original="curTreeNode.value.originContent"
                            @mounted="handleEditorMount">
                        </monaco-editor>
                    </template>
                    <template v-else>
                        <div class="biz-editor-tip" v-if="!isYamlTemplateLoading">
                            <i class="bcs-icon bcs-icon-edit2"></i>
                            <p>
                                {{$t('你可以通过+号新建K8S资源yaml文件，也可以通过上方的')}}
                                <a :key="fileImportIndex" href="javascript: void(0);" style="color: #aaa; text-decoration: underline;">
                                    {{$t('导入按钮')}}
                                    <input ref="fileInputClone" type="file" name="upload" class="file-input" accept="application/zip,application/x-zip,application/x-zip-compressed" @change="handleFileInput(true)">
                                </a>
                                {{$t('导入zip包')}}
                            </p>
                        </div>
                    </template>

                    <div :class="['biz-var-panel', { 'show': isVarPanelShow }]" v-clickoutside="hideVarPanel">
                        <div class="var-panel-header">
                            <strong class="var-panel-title">{{$t('可用变量')}}<span class="f12">（{{$t('模板集中引入方式')}}：{{varUserWay}}）</span></strong>
                        </div>
                        <div class="var-panel-list">
                            <table class="bk-table biz-var-table">
                                <thead>
                                    <tr>
                                        <th>{{$t('变量名')}}</th>
                                        <th style="width: 230px;">KEY</th>
                                        <th style="width: 43px;"></th>
                                    </tr>
                                </thead>
                            </table>
                            <div class="var-list">
                                <table class="bk-table biz-var-table">
                                    <tbody>
                                        <template v-if="varList.length">
                                            <tr v-for="item of varList" :key="item.name">
                                                <td>
                                                    <bcs-popover :content="item.name" placement="right">
                                                        <span class="var-name">{{item.name}}</span>
                                                    </bcs-popover>
                                                </td>
                                                <td style="width: 230px;">
                                                    <bcs-popover :content="item.key" placement="right">
                                                        <span class="var-key">{{item.key}}</span>
                                                    </bcs-popover>
                                                </td>
                                                <td style="width: 43px;">
                                                    <bk-button class="var-copy-btn" :data-clipboard-text="`{{${item.key}}}`" type="default">
                                                        <i class="bcs-icon bcs-icon-clipboard"></i>
                                                    </bk-button>
                                                </td>
                                            </tr>
                                        </template>
                                        <template v-else>
                                            <tr>
                                                <td colspan="3">
                                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                                </td>
                                            </tr>
                                        </template>
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>

                    <div :class="['biz-image-content', { 'show': isImagePanelShow }]" v-clickoutside="hideImagePanel">
                        <div class="biz-image-list" style="width: 600px;">
                            <div class="bk-dropdown-box ml20 mb20" style="width: 300px;">
                                <bkbcs-input
                                    style="width: 240px;"
                                    type="text"
                                    :placeholder="$t('选择镜像')"
                                    :display-key="'_name'"
                                    :setting-key="'_id'"
                                    :search-key="'_name'"
                                    :value.sync="imageName"
                                    :list="varList"
                                    :is-link="true"
                                    :is-select-mode="true"
                                    :default-list="imageList"
                                    @item-selected="changeImage(...arguments)">
                                </bkbcs-input>
                                <bk-button class="bk-button bk-default is-outline is-icon" @click="initImageList">
                                    <div class="bk-spin-loading bk-spin-loading-mini bk-spin-loading-default" style="margin-top: -3px;" v-if="isLoadingImageList">
                                        <div class="rotate rotate1"></div>
                                        <div class="rotate rotate2"></div>
                                        <div class="rotate rotate3"></div>
                                        <div class="rotate rotate4"></div>
                                        <div class="rotate rotate5"></div>
                                        <div class="rotate rotate6"></div>
                                        <div class="rotate rotate7"></div>
                                        <div class="rotate rotate8"></div>
                                    </div>
                                    <i class="bcs-icon bcs-icon-refresh" v-else></i>
                                </bk-button>
                            </div>

                            <div class="bk-dropdown-box mb20" style="width: 4;">
                                <bkbcs-input
                                    type="text"
                                    :placeholder="$t('版本号1')"
                                    :display-key="'_name'"
                                    :setting-key="'_id'"
                                    :search-key="'_name'"
                                    :value.sync="imageVersion"
                                    :list="varList"
                                    :is-select-mode="true"
                                    :default-list="imageVersionList"
                                    :disabled="!imageName"
                                    @item-selected="setImageVersion"
                                >
                                </bkbcs-input>
                            </div>

                            <div class="image-box" v-show="image">
                                <input type="text" class="bk-form-input" readonly :value="image" style="width: 482px;">
                                <bk-button class="bk-button bk-primary image-copy-btn" :data-clipboard-text="`${image}`">{{$t('复制')}}</bk-button>
                            </div>
                        </div>
                        <p class="biz-tip">
                            {{$t('使用指南：请将镜像复制后填入所使用的YAML中')}}
                        </p>
                    </div>
                </div>
            </div>
        </section>
        <bk-dialog
            :is-show.sync="versionDialogConf.isShow"
            :width="600"
            :has-header="false"
            :quick-close="false"
            :ext-cls="'create-project-dialog'"
            :content="versionDialogConf.content"
            @cancel="hideVersionBox">
            <template slot="content">
                <div class="version-box">
                    <p class="title">{{$t('保存修改到')}}：</p>
                    <ul :class="['version-list', { 'is-en': isEn }]">
                        <template v-if="!isNewVersion">
                            <li class="item mb10">
                                <label class="bk-form-radio label-item">
                                    <input type="radio" name="save-version-way" :class="{ 'is-checked': saveVersionWay === 'cur' }" value="cur" v-model="saveVersionWay">
                                    <i class="bk-radio-text" style="display: inline-block; min-width: 70px;">{{$t('当前版本号')}}：{{lateShowVersionName}}</i>
                                </label>
                            </li>

                            <li class="item mb10">
                                <label class="bk-form-radio label-item" style="margin-right: 0;">
                                    <input type="radio" name="save-version-way" :class="{ 'is-checked': saveVersionWay === 'new' }" value="new" v-model="saveVersionWay">
                                    <i class="bk-radio-text" style="display: inline-block; min-width: 70px;">{{$t('新版本')}}：</i>
                                    <bkbcs-input :placeholder="$t('请输入版本号')" @focus="saveVersionWay = 'new'" style="width: 176px; flex: 1;" v-model="versionKeyword" />
                                </label>
                            </li>

                            <li class="item" v-if="withoutCurVersionList.length">
                                <label class="bk-form-radio label-item" style="margin-right: 0;">
                                    <input type="radio" name="save-version-way" :class="{ 'is-checked': saveVersionWay === 'old' }" value="old" v-model="saveVersionWay">
                                    <i class="bk-radio-text" style="display: inline-block; min-width: 70px; letter-spacing: 0;">{{$t('其它版本')}}：</i>
                                    <bk-selector
                                        style="width: 176px;"
                                        :placeholder="$t('请选择版本号')"
                                        :setting-key="'show_version_id'"
                                        :selected.sync="selectedVersion"
                                        :list="withoutCurVersionList"
                                        @item-selected="selectVersion">
                                    </bk-selector>
                                </label>
                            </li>
                        </template>
                        <template v-else>
                            <li class="item">
                                <label class="bk-form-radio label-item" style="margin-right: 0;">
                                    <i class="bk-radio-text" style="display: inline-block; width: 70px; letter-spacing: 0;">{{$t('新版本')}}：</i>
                                    <bkbcs-input :placeholder="$t('请输入版本号')" @focus="saveVersionWay = 'new'" style="width: 203px; flex: 1;" v-model="versionKeyword" />
                                </label>
                            </li>
                        </template>
                        <li class="item">
                            <label :class="['notes', 'label-item', { 'new-item': isNewVersion }]" style="margin-right: 0;">
                                <i :class="['notes-text', { 'is-en-text': isEn, 'is-new': isNewVersion }]" :style="{ 'letter-spacing': 0, 'padding-left': isNewVersion ? 0 : '26px' }">{{$t('版本说明')}}：</i>
                                <bk-textarea class="notes-input" :style="{ width: isNewVersion ? '203px' : '176px' }" :placeholder="$t('请输入版本说明')" :value.sync="curVersionNotes" />
                            </label>
                        </li>
                    </ul>
                </div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template v-if="!canVersionSave">
                        <bk-button type="primary" disabled>
                            {{$t('确定')}}
                        </bk-button>
                    </template>
                    <template v-else>
                        <bk-button type="primary" :loading="isTemplateSaving" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary" @click="saveYamlTemplate">
                            {{$t('确定')}}
                        </bk-button>
                    </template>
                    <bk-button type="button" :disabled="isTemplateSaving" class="bk-dialog-btn bk-dialog-btn-cancel" @click="hideVersionBox">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </bk-dialog>

        <bk-sideslider
            :quick-close="true"
            :is-show.sync="versionSidePanel.isShow"
            :title="versionSidePanel.title"
            :width="'840'">
            <div class="p30" slot="content" v-bkloading="{ isLoading: isVersionListLoading }">
                <table class="bk-table biz-data-table has-table-bordered">
                    <thead>
                        <tr>
                            <th>{{$t('版本号')}}</th>
                            <th>{{$t('更新时间')}}</th>
                            <th>{{$t('最后更新人')}}</th>
                            <th style="width: 138px;">{{$t('操作')}}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <template v-if="allVersionList.length">
                            <tr v-for="(versionData, index) in allVersionList" :key="index">
                                <td>
                                    <p>
                                        <span>{{versionData.name}}</span>
                                        <span v-if="versionData.show_version_id === curShowVersionId">{{$t('(当前)')}}</span>
                                    </p>

                                    <bcs-popover
                                        v-if="versionData.comment"
                                        :delay="300"
                                        :content="versionData.comment"
                                        placement="right">
                                        <span style="color: #3a84ff; font-size: 12px;">{{$t('版本说明')}}</span>
                                    </bcs-popover>
                                </td>
                                <td>{{versionData.updated}}</td>
                                <td>{{versionData.updator}}</td>
                                <td>
                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop.prevent="getTemplateByVersion(versionData.show_version_id)">{{$t('加载')}}</a>
                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop.prevent="exportTemplateByVersion(versionData.show_version_id)">{{$t('导出')}}</a>
                                </td>
                            </tr>
                        </template>
                        <template v-else>
                            <tr>
                                <td colspan="4">
                                    <div class="biz-app-list">
                                        <div class="bk-message-box" style="min-height: auto;">
                                            <bcs-exception type="empty" scene="part"></bcs-exception>
                                        </div>
                                    </div>
                                </td>
                            </tr>
                        </template>
                    </tbody>
                </table>
            </div>
        </bk-sideslider>

        <bk-dialog
            width="400"
            :title="nodeDialogConf.title"
            :quick-close="false"
            :has-header="false"
            :is-show.sync="nodeDialogConf.isShow"
            @confirm="addNode"
            @cancel="hideNodeDialog">
            <template slot="content">
                <div class="bk-form bk-form-vertical">
                    <div class="bk-form-item">
                        <label class="bk-label">
                            <span>{{nodeDialogConf.action === 'addCatalog' ? $t('目录') : $t('文件')}}{{$t('名称')}}：</span>
                        </label>
                        <div class="bk-form-content mb10">
                            <bkbcs-input v-model="nodeDialogConf.name" :placeholder="$t('请输入名称')" ref="newFileInputer" @enter="addNode"></bkbcs-input>
                        </div>
                    </div>
                </div>
            </template>
        </bk-dialog>

        <bk-dialog
            width="400"
            :title="fileDialogConf.title"
            :quick-close="false"
            :has-header="false"
            :is-show.sync="fileDialogConf.isShow"
            @confirm="updateNode"
            @cancel="hideFileDialog">
            <template slot="content">
                <div class="bk-form bk-form-vertical">
                    <div class="bk-form-item">
                        <label class="bk-label">
                            <span>{{$t('文件名称')}}：</span>
                        </label>
                        <div class="bk-form-content mb10">
                            <bkbcs-input v-model="fileDialogConf.fileName" ref="renameInputer" @enter="updateNode"></bkbcs-input>
                        </div>
                    </div>
                </div>
            </template>
        </bk-dialog>
    </div>
</template>

<script>
    import yamljs from 'js-yaml'
    import MonacoEditor from './editor'
    // import CollapseTransition from '@/components/menu/collapse-transition'
    import { catchErrorHandler, uuid } from '@/common/util'
    import Clipboard from 'clipboard'
    import clickoutside from '@/directives/clickoutside'
    import JSZip from 'jszip'
    import { saveAs } from 'file-saver'
    import { Archive } from 'libarchive.js/main.js'
    import path2tree from '@/common/path2tree'
    import resizer from '@/components/resize'

    Archive.init({
        workerUrl: `${window.STATIC_URL}${window.VERSION_STATIC_URL}/archive-worker/worker-bundle.js`
    })
    export default {
        components: {
            MonacoEditor,
            resizer
        },
        directives: {
            clickoutside
        },
        data () {
            return {
                isEditName: false,
                isEditDesc: false,
                isTemplateSaving: false,
                isVersionListLoading: true,
                isLoadingImageList: false,
                isTemplateLocking: false,
                isVarPanelShow: false,
                isImagePanelShow: false,
                isEditFileName: false,
                pathList: [],
                fileImportIndex: 0,
                useEditorDiff: false,
                winHeight: 500,
                isYamlTemplateLoading: true,
                reRenderEditor: 0,
                updatedTimestamp: 0,
                curTemplate: {
                    name: '',
                    desc: '',
                    show_version: {
                        name: ''
                    },
                    template_files: []
                },
                saveVersionWay: 'cur',
                selectedVersion: '',
                versionSidePanel: {
                    isShow: false,
                    title: this.$t('版本列表')
                },
                fileNameReg: /^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$/,
                renderTimer: 0,
                imageList: [],
                imageVersionList: [],
                curImageData: {},
                imageName: '',
                imageVersion: '',
                image: '',
                varUserWay: this.$t('{{变量KEY}}'),
                editorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'yaml',
                    readOnly: false,
                    fullScreen: false,
                    value: '',
                    editor: null
                },
                addFileNameTmp: '',
                editFileNameTmp: '',
                curResource: {
                    files: []
                },
                curResourceFile: {
                    id: 0,
                    name: '',
                    content: '',
                    originContent: ''
                },
                curTreeNode: {
                    id: 0,
                    name: '',
                    value: {
                        id: 0,
                        fullName: '',
                        content: '',
                        originContent: ''
                    }
                },
                editTemplate: {
                    name: '',
                    desc: ''
                },
                versionDialogConf: {
                    isShow: false,
                    closeIcon: false
                },
                versionKeyword: '',
                yamlEditorOptions: {
                    readOnly: false,
                    fontSize: 14
                },
                yamlResourceConf: {
                    initial_templates: {},
                    resource_names: []
                },
                yamlTemplateJson: null,
                tabName: 'default',
                defaultTreeData: [],
                customTreeData: [],
                yamlFileList: [],
                nodeDialogConf: {
                    title: '',
                    isShow: false,
                    name: ''
                },
                fileDialogConf: {
                    title: '',
                    isShow: false,
                    fileName: ''
                },
                zipTooltipText: this.$t('请选择zip压缩包导入，包中的文件名以.yaml结尾。其中的yaml文件(非"_常用Manifest"目录下的文件)将会统一导入到自定义Manifest分类下。注意：同名文件会被覆盖'),
                curVersionNotes: '',
                displayVersionNotes: '--',
                webAnnotations: { perms: {} }
            }
        },
        computed: {
            isEn () {
                return this.$store.state.isEn
            },
            curProject () {
                return this.$store.state.curProject
            },
            templateId () {
                return Number(this.$route.params.templateId)
            },
            isNewTemplate () {
                return this.templateId === 0
            },
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            editorHeight () {
                // 由于则导航的高度最小为630，导致整个页面的高度最不为630
                const height = this.winHeight - 260
                return height < 630 ? 630 : height
            },
            varList () {
                const list = this.$store.state.variable.varList
                list.forEach(item => {
                    item._id = item.key
                    item._name = item.key
                })
                return list
            },
            lateShowVersionName () {
                let name = ''
                this.allVersionList.forEach(item => {
                    if (item.show_version_id === this.curShowVersionId) {
                        name = item.name
                    }
                })
                return name
            },
            canVersionSave () {
                if (this.saveVersionWay === 'cur' && this.curShowVersionId) {
                    return true
                } else if (this.saveVersionWay === 'old' && this.selectedVersion) {
                    return true
                } else if (this.saveVersionWay === 'new' && this.versionKeyword) {
                    return true
                }
                return false
            },
            curShowVersionId () {
                return this.curTemplate.show_version.show_version_id
            },
            curShowVersionName () {
                return this.curTemplate.show_version.name
            },
            allVersionList () {
                return this.$store.state.k8sTemplate.versionList
            },
            userInfo () {
                return this.$store.state.user
            },
            templateLockStatus () {
                const status = {
                    isLocked: false,
                    isCurLocker: false,
                    locker: ''
                }
                // 模块集已经加锁
                if (this.curTemplate && this.curTemplate.is_locked) {
                    status.isLocked = true
                    status.locker = this.curTemplate.locker
                    // 如果是当前用户加锁
                    if (this.curTemplate.locker && this.curTemplate.locker === this.userInfo.username) {
                        status.isCurLocker = true
                    } else {
                        status.isCurLocker = false
                    }
                }
                return status
            },
            withoutCurVersionList () {
                // 去掉当前版本
                return this.$store.state.k8sTemplate.versionList.filter(item => {
                    return item.show_version_id !== this.curShowVersionId
                })
            },
            defaultTemplateFiles () {
                return this.curTemplate.template_files.filter(item => {
                    return item.resource_name !== 'CustomManifest'
                })
            },
            customTemplateFiles () {
                return this.curTemplate.template_files.filter(item => {
                    return item.resource_name === 'CustomManifest'
                })
            },
            canTemplateSave () {
                if (!this.curTemplate.id) {
                    const resource = this.curTemplate.template_files.find(item => {
                        return item.files.length
                    })
                    return resource
                }
                return true
            },
            canTemplateExport () {
                return !!this.templateId
            },
            curTemplateFiles () {
                if (this.tabName === 'default') {
                    return this.defaultTemplateFiles
                } else {
                    return this.customTemplateFiles
                }
            },
            isNewVersion () {
                return !(this.allVersionList.length && this.curShowVersionId !== -1)
            }
        },

        watch: {
            useEditorDiff () {
                this.reRenderEditor++
            },
            editorHeight () {
                this.reRenderEditor++
            },
            'curResourceFile.id' () {
                this.reRenderEditor++
            },
            varList () {
                if (this.clipboardVarInstance && this.clipboardVarInstance.off) {
                    this.clipboardVarInstance.off('success')
                }
                this.clipboardVarInstance = new Clipboard('.var-copy-btn')
                this.clipboardVarInstance.on('success', e => {
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('复制成功')
                    })
                    this.isVarPanelShow = false
                })
            },
            image () {
                if (this.clipboardImageInstance && this.clipboardImageInstance.off) {
                    this.clipboardImageInstance.off('success')
                }

                this.clipboardImageInstance = new Clipboard('.image-copy-btn')
                this.clipboardImageInstance.on('success', e => {
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('复制成功')
                    })
                    this.isImagePanelShow = false
                })
            },
            'curTemplate.template_files': {
                deep: true,
                handler (val) {
                    clearTimeout(this.renderTimer)
                    // 防止频繁刷新
                    this.renderTimer = setTimeout(() => {
                        const defaultTree = this.getTreeNodes('default')
                        const customTree = this.getTreeNodes('CustomManifest', { hideRoot: true })
                        this.defaultTreeData = [defaultTree]
                        this.customTreeData = [customTree]
                    }, 500)
                }
            },
            'curShowVersionId' () {
                this.allVersionList.forEach(item => {
                    if (item.show_version_id === this.curShowVersionId) {
                        this.curVersionNotes = item.comment
                        this.displayVersionNotes = item.comment
                    }
                })
            },
            'saveVersionWay' (val, old) {
                if (val && val === old) return
                if (val === 'new') {
                    this.curVersionNotes = ''
                    return
                }
                let item = null
                if (val === 'cur') {
                    item = this.allVersionList.find(item => item.show_version_id === this.curShowVersionId)
                } else if (val === 'old' && this.selectedVersion) {
                    item = this.allVersionList.find(item => item.show_version_id === this.selectedVersion)
                }
                item && (this.curVersionNotes = item.comment)
            }
        },

        async created () {
            this.isYamlTemplateLoading = true
            await this.initYamlResources()
            this.initVarList()
            this.initTemplate()
            this.initImageList()
        },

        mounted () {
            this.winHeight = window.innerHeight

            const debounce = this.debounce(() => {
                this.winHeight = window.innerHeight
                this.reRenderEditor++
            }, 200)

            window.addEventListener('resize', () => {
                debounce()
            })
        },

        beforeRouteLeave (to, from, next) {
            let isEdited = false
            const changeActions = ['create', 'delete', 'update']
            const curTemplate = this.getYamlParams()
            curTemplate.template_files.forEach(resource => {
                resource.files.forEach(file => {
                    if (changeActions.includes(file.action)) {
                        isEdited = true
                    } else if (file.content !== file.originContent) {
                        isEdited = true
                    }
                })
            })

            if (isEdited) {
                this.$bkInfo({
                    title: this.$t('确认离开'),
                    content: this.$createElement('p', {
                        style: {
                            textAlign: 'left'
                        }
                    }, this.$t('模板编辑的内容未保存，确认要离开？')),
                    confirmFn () {
                        next(true)
                    }
                })
            } else {
                next(true)
            }
        },

        beforeDestroy () {
            if (this.clipboardVarInstance && this.clipboardVarInstance.off) {
                this.clipboardVarInstance.off('success')
            }

            if (this.clipboardImageInstance && this.clipboardImageInstance.off) {
                this.clipboardImageInstance.off('success')
            }
        },

        methods: {
            getAuthority (actionId, templateId) {
                return !!this.webAnnotations?.perms[templateId]?.[actionId]
            },
            /**
             * 初始化入口
             */
            initTemplate () {
                if (this.templateId === 0) {
                    this.createNewTemplate()
                } else {
                    this.getYamlTemplateDetail()
                }
                this.getVersionList()
            },

            /**
             * 初始化资源列表
             */
            async initYamlResources () {
                this.yamlTemplateJson = {
                    name: '',
                    desc: '',
                    show_version: {
                        name: ''
                    },
                    template_files: []
                }

                const projectId = this.projectId
                try {
                    const res = await this.$store.dispatch('k8sTemplate/getYamlResources', { projectId })
                    this.yamlResourceConf = res.data
                    this.yamlResourceConf.resource_names.forEach(resource => {
                        this.yamlTemplateJson.template_files.push({
                            resource_name: resource,
                            actived: true,
                            files: []
                        })
                    })
                } catch (e) {
                    catchErrorHandler(e, this)
                    this.isYamlTemplateLoading = false
                }
            },

            /**
             * 镜像列表初始化
             */
            async initImageList () {
                if (this.isLoadingImageList) return false

                this.isLoadingImageList = true

                const projectId = this.projectId
                try {
                    const res = await this.$store.dispatch('k8sTemplate/getImageList', { projectId })
                    const data = res.data
                    setTimeout(() => {
                        data.forEach(item => {
                            item._id = item.value
                            item._name = item.name
                        })
                        this.imageList.splice(0, this.imageList.length, ...data)
                        this.$store.commit('k8sTemplate/updateImageList', this.imageList)
                        this.isLoadingImageList = false
                    }, 1500)
                } catch (e) {
                    this.isLoadingImageList = false
                    catchErrorHandler(e, this)
                }
            },

            getTreeNodes (type, conf) {
                let templateFiles = []
                const filterType = type || this.tabName
                if (filterType === 'default') {
                    templateFiles = this.curTemplate.template_files.filter(item => {
                        return item.resource_name !== 'CustomManifest'
                    })
                } else {
                    templateFiles = this.curTemplate.template_files.filter(item => {
                        return item.resource_name === 'CustomManifest'
                    })
                }
                const yamlFileList = []
                templateFiles.forEach(resource => {
                    if (resource.files.length) {
                        resource.files.forEach(file => {
                            file.originContent = file.content
                            const fullName = `${resource.resource_name}/${file.name}`
                            const item = {
                                id: uuid(),
                                type: 'file',
                                status: 'normal',
                                name: fullName,
                                value: {
                                    ...file,
                                    fullName: fullName,
                                    resourceName: resource.resource_name
                                }
                            }
                            yamlFileList.push(item)
                        })
                    } else {
                        const parent = {
                            id: uuid(),
                            type: 'catalog',
                            status: 'normal',
                            name: `${resource.resource_name}`
                        }
                        yamlFileList.push(parent)
                    }
                })
                const tree = path2tree(yamlFileList, conf)
                return tree
            },

            changeImage (value, data) {
                const projectId = this.projectId
                const imageId = data.value
                const isPub = data.is_pub

                this.curImageData = data
                // 如果不是输入变量
                if (isPub !== undefined) {
                    this.$store.dispatch('k8sTemplate/getImageVertionList', { projectId, imageId, isPub }).then(res => {
                        const data = res.data
                        data.forEach(item => {
                            item._id = item.text
                            item._name = item.text
                        })

                        this.imageVersionList.splice(0, this.imageVersionList.length, ...data)
                        if (this.imageVersionList.length) {
                            const imageInfo = this.imageVersionList[0]

                            this.imageVersion = imageInfo.text
                            this.setImageVersion(imageInfo.value, imageInfo)
                        } else {
                            this.image = ''
                        }
                    }, res => {
                        const message = res.message
                        this.$bkMessage({
                            theme: 'error',
                            message: message
                        })
                    })
                } else {
                    this.image = ''
                    this.imageVersion = ''
                    this.imageVersionList = []
                }
            },

            /**
             * 设置镜像版本
             */
            setImageVersion (value, data) {
                // 镜像和版本都是通过下拉选择
                const projectCode = this.projectCode
                // curImageData不是空对象
                if (JSON.stringify(this.curImageData) !== '{}') {
                    if (data.text && data.value) {
                        this.imageVersion = data.text
                        const items = data.value.split('/')
                        items.splice(0, 1, '{{SYS_JFROG_DOMAIN}}')
                        this.image = `'${items.join('/')}'`
                    } else if (this.curImageData.is_pub !== undefined) {
                        // 镜像是下拉，版本是变量
                        // image = imageBase + imageName + ':' + imageVersion
                        const imageName = this.imageName
                        this.imageVersion = value
                        this.image = `'{{SYS_JFROG_DOMAIN}}/${imageName}:${value}'`
                    } else {
                        // 镜像和版本是变量
                        // image = imageBase +  'paas/' + projectCode + '/' + imageName + ':' + imageVersion
                        const imageName = this.imageName
                        this.imageVersion = value
                        this.image = `'{{SYS_JFROG_DOMAIN}}/paas/${projectCode}/${imageName}:${value}'`
                    }
                }
            },

            /**
             * 创建新的模板
             */
            createNewTemplate () {
                const params = JSON.parse(JSON.stringify(this.yamlTemplateJson))
                params.name = this.$t('模板集_') + (+new Date())
                params.desc = this.$t('模板集描述')

                // 如果是从表单模板集导出过来
                if (this.$route.query.action === 'export' && localStorage['cloneTemplateSet']) {
                    const templateset = JSON.parse(localStorage['cloneTemplateSet'])
                    for (const key in templateset) {
                        const resource = params.template_files.find(item => item.resource_name === key)
                        if (resource) {
                            resource.files = templateset[key].map((content, index) => {
                                const application = yamljs.load(content)
                                const defaultName = `${key.toLowerCase()}-${index} + 1`
                                const name = `${application.metadata.name || defaultName}.yaml`
                                const id = `local_${+new Date()}`
                                return {
                                    id: id,
                                    name: name,
                                    content: content,
                                    originContent: content,
                                    isEdited: false,
                                    action: 'create'
                                }
                            })
                        }
                    }
                }
                this.curTemplate = params
                this.isYamlTemplateLoading = false
            },

            handleEditorMount (editorInstance, monacoEditor) {
                this.monacoEditor = monacoEditor
            },

            handleSelectFile (resource, file) {
                this.setCurResourceFile(resource, file)
            },

            editTemplateName () {
                this.isEditName = true
                this.editTemplate.name = this.curTemplate.name

                this.$nextTick(() => {
                    const inputer = this.$refs.templateNameInput
                    inputer.focus()
                    inputer.select()
                })
            },

            cancelEditName () {
                setTimeout(() => {
                    this.isEditName = false
                }, 200)
            },

            editTemplateDesc () {
                this.isEditDesc = true
                this.editTemplate.desc = this.curTemplate.desc

                this.$nextTick(() => {
                    const inputer = this.$refs.templateDescInput
                    inputer.focus()
                    inputer.select()
                })
            },

            cancelEditDesc () {
                setTimeout(() => {
                    this.isEditDesc = false
                }, 200)
            },

            /**
             * 保存模板集名称和描述
             */
            async saveTemplate () {
                const data = this.editTemplate
                const projectId = this.projectId
                const templateId = this.templateId

                // 用户填空数据，用原数据
                if (!data.name) {
                    data.name = this.curTemplate.name
                }
                if (!data.desc) {
                    data.desc = this.curTemplate.desc
                }

                if (this.updatedTimestamp) {
                    data.updated_timestamp = this.updatedTimestamp
                }
                // 没有修改，不处理
                if (data.name === this.curTemplate.name && data.desc === this.curTemplate.desc) {
                    this.isEditName = false
                    this.isEditDesc = false
                    return true
                }

                if (templateId && templateId !== 0) {
                    try {
                        await this.$store.dispatch('k8sTemplate/updateYamlTemplate', {
                            projectId,
                            templateId,
                            data
                        })

                        const res = await this.$store.dispatch('k8sTemplate/getYamlTemplateDetail', {
                            projectId: this.projectId,
                            templateId: this.templateId
                        })
                        this.updatedTimestamp = res.data.updated_timestamp
                        this.curTemplate.updated_timestamp = res.data.updated_timestamp
                        this.updateTemplateBaseInfo(data)
                        this.$bkMessage({
                            theme: 'success',
                            message: this.$t('模板集基础信息保存成功')
                        })
                    } catch (e) {
                        catchErrorHandler(e, this)
                    }
                } else {
                    this.updateTemplateBaseInfo(data)
                }
            },

            updateTemplateBaseInfo (data) {
                this.isEditName = false
                this.isEditDesc = false
                this.curTemplate.name = data.name
                this.curTemplate.desc = data.desc
            },

            /**
             * 展开、折叠
             *
             * @param {Object} resource 资源
             */
            handleToggleResource (resource) {
                resource.actived = !resource.actived
            },

            focusAddNameInput () {
                this.$nextTick(() => {
                    const inputer = this.$refs.fileNameInput && this.$refs.fileNameInput[0]
                    if (inputer) {
                        inputer.focus()
                        inputer.select()
                    }
                })
            },

            /**
             * 添加相应的资源文件
             *
             * @param {Object} resource 资源
             * action 'create' 创建
             * action 'update' 更新
             * action 'delete' 删除
             * action 'unchange' 没改变
             */
            handleAddFile (resource, isImmediate, fileNamePrefix) {
                // if (this.addFileNameTmp) {
                //     // this.focusAddNameInput()
                //     return false
                // }
                const type = resource.resource_name
                const index = resource.files.length + 1
                const content = this.yamlResourceConf.initial_templates[type] || ''
                const name = `${type.toLowerCase()}-${index}.yaml`
                const id = `local_${+new Date()}`

                const file = {
                    id: id,
                    name: fileNamePrefix ? `${fileNamePrefix}/${name}` : name,
                    content: content,
                    originContent: content,
                    isEdited: true,
                    action: 'create'
                }

                resource.actived = true
                resource.files.push(file)
                this.addFileNameTmp = file.name

                if (isImmediate) {
                    file.isEdited = false
                    this.setCurResourceFile(resource, file)
                } else {
                    // this.focusAddNameInput()
                }
                this.getTreeNodes()
            },

            checkFileName (resource, file, name, action) {
                const nameReg = /^[a-z]{1}[a-z0-9-.]{0,63}$/
                if (!name) {
                    // 如果是刚新建，直接删除
                    if (file.action === 'create' && action === 'addFileBlur') {
                        // 把新建的删除
                        resource.files = resource.files.filter(item => {
                            return item.id !== file.id
                        })
                    } else {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请输入资源文件名称')
                        })
                    }
                    return false
                }

                if (!nameReg.test(name)) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('文件名称错误，只能包含：小写字母、数字、点(.)、连字符(-)，必须是字母开头，长度小于64个字符'),
                        delay: 8000
                    })
                    return false
                }

                // 判断是否已经重复命名(存在两个)
                const repeatFile = resource.files.find(resourceFile => {
                    return resourceFile.name === name && resourceFile.id !== file.id
                })

                if (repeatFile) {
                    // const deleteFile = files.find(file => file.action === 'delete')
                    // 如果新建和已经删除的重命名，把已经删除的重新启用
                    if (repeatFile.action === 'delete') {
                        repeatFile.action = 'update'
                        repeatFile.content = file.content
                        repeatFile.originContent = file.originContent

                        // 把新建的删除
                        resource.files = resource.files.filter(item => {
                            return item.id !== file.id
                        })

                        this.setCurResourceFile(resource, repeatFile)
                        return false
                    } else {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('文件名称不能重复'),
                            delay: 8000
                        })
                        return false
                    }
                } else {
                    return true
                }
            },
            /**
             * 编辑当前资源文件名称失焦点
             *
             * @param {Object} resource 资源
             * @param {Object} file 资源文件
             */
            handleFileEnter (resource, file) {
                this.isEnterTrigger = true
                const name = this.addFileNameTmp
                if (!this.checkFileName(resource, file, name)) return false

                file.isEdited = false
                file.name = name
                this.setCurResourceFile(resource, file)
                this.addFileNameTmp = ''
            },
            handleFileBlur (resource, file) {
                const name = this.addFileNameTmp

                if (this.isEnterTrigger) {
                    this.isEnterTrigger = false
                    return false
                }
                if (!this.checkFileName(resource, file, name, 'addFileBlur')) return false

                file.isEdited = false
                file.name = name
                this.setCurResourceFile(resource, file)
                this.addFileNameTmp = ''
            },

            /**
             * 设置当前要编辑的资源文件
             * @param {Object} file 资源文件
             */
            setCurResourceFile (resource, file) {
                this.addFileNameTmp = ''
                this.useEditorDiff = false
                this.curResource = resource
                this.curResourceFile = file
                this.yamlEditorOptions.readOnly = false
                if (resource.resource_name === 'CustomManifest') {
                    this.tabName = 'custom'
                }
            },

            /**
             * 清空当前资源文件
             */
            clearCurResourfeFile () {
                this.curResourceFile = {
                    id: 0,
                    name: '',
                    yamlValue: ''
                }
            },

            /**
             * 删除资源文件
             *
             * @param {Object} resource 资源
             * @param {Object} file 文件
             * @param {Number} index 索引
             */
            handleRemoveFile (resource, file, index) {
                const self = this
                this.$bkInfo({
                    title: this.$t('确认删除'),
                    content: this.$createElement('p', { style: { 'text-align': 'left' } }, `${this.$t('删除')} ${resource.resource_name}：${file.name}`),
                    confirmFn () {
                        self.removeLocalFile(resource, file, index)
                    }
                })
            },

            /**
             * 删除本地资源文件
             * @param  {object} application application
             * @param  {number} index 索引
             */
            removeLocalFile (resource, file, index) {
                // 如果是新建直接删除，否则设置action为delete
                if (String(file.id).startsWith('local_')) {
                    resource.files.splice(index, 1)
                } else {
                    file.action = 'delete'
                }
                // 如果删除的为当前编辑文件，则重新设置编辑文件
                if (this.curResourceFile.id === file.id) {
                    this.clearCurResourfeFile()

                    // 找到第一个不为delete状态的文件
                    const activeFile = resource.files.find(file => file.action !== 'delete')
                    if (activeFile) {
                        this.setCurResourceFile(resource, activeFile)
                    } else {
                        this.yamlEditorOptions.readOnly = true
                    }
                }

                this.useEditorDiff = false
            },

            toggleCompare () {
                this.useEditorDiff = !this.useEditorDiff
            },

            debounce (func, wait) {
                let timer
                const that = this
                return function () {
                    const context = that
                    const args = arguments
                    if (timer) {
                        clearTimeout(timer)
                    }

                    timer = setTimeout(() => {
                        func.apply(context, args)
                    }, wait)
                }
            },

            /**
             * 校验版本名称
             */
            checkVersionData (versionName) {
                const nameReg = /^[a-zA-Z0-9-_.]{1,45}$/

                if (!nameReg.test(versionName)) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请填写1至45个字符（由字母、数字、下划线以及 - 或 . 组成）')
                    })
                    return false
                }

                for (const item of this.allVersionList) {
                    if (item.name === versionName) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('版本{versionKeyword}已经存在', {
                                versionKeyword: this.versionKeyword
                            })
                        })
                        return false
                    }
                }
                return true
            },

            /**
             * 保存模板集
             */
            handleSaveTemplate () {
                const template = this.getYamlParams()
                // 验证模板集信息
                if (!template.name) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入模板集名称')
                    })
                    return false
                }

                if (!template.desc) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入模板集描述')
                    })
                    return false
                }

                const resources = template.template_files
                for (const resource of resources) {
                    const files = resource.files.filter(file => file.action !== 'delete')
                    for (const file of files) {
                        if (!file.name) {
                            this.$bkMessage({
                                theme: 'error',
                                message: this.$t('请输入资源文件名称')
                            })
                            return false
                        }

                        if (!file.content) {
                            this.$bkMessage({
                                theme: 'error',
                                message: `${file.name}：${this.$t('请输入资源文件内容')}`
                            })
                            return false
                        }
                        try {
                            // 一个yaml文件支持多个，以---分隔
                            const yamls = file.content.split(/[^-]---[^-]/)
                            for (let i = 0; i < yamls.length; i++) {
                                const content = yamls[i]
                                if (i !== 0 && (!content || /^[\s\t\n\r]+$/.test(content))) {
                                    this.$bkMessage({
                                        theme: 'error',
                                        message: `${file.name}：${this.$t('请输入合法的YAML')}`
                                    })
                                    return false
                                }
                                yamljs.load(content)
                            }
                        } catch (err) {
                            console.error(err)
                            this.$bkMessage({
                                theme: 'error',
                                delay: 10000,
                                message: `${file.name}：${this.$t('请输入合法的YAML')}：${err.message}`
                            })
                            return false
                        }
                    }
                }

                this.showVersionBox()
            },

            saveYamlTemplate () {
                if (this.saveVersionWay === 'new' && this.versionKeyword && !this.checkVersionData(this.versionKeyword)) {
                    return false
                }
                if (this.isTemplateSaving) {
                    return false
                }

                if (this.templateId === 0) {
                    this.createYamlTemplate()
                } else {
                    this.updateYamlTemplate()
                }
            },

            /**
             * 组装数据
             */
            getYamlFiles (node) {
                if (node.value) {
                    this.pathList = []
                    this.getAbsolutePath(node)

                    const file = JSON.parse(JSON.stringify(node.value))
                    file.fullName = this.pathList.join('/')
                    this.yamlListCache.push(file)
                } else {
                    if (node.children) {
                        node.children.forEach(child => {
                            this.getYamlFiles(child)
                        })
                    }
                }
            },
            mergeParams (file, params) {
                const paths = file.fullName.split('/')
                const resourceName = paths.shift()
                const fileName = paths.join('/')
                const matchResource = params.template_files.find(resource => resource.resource_name === resourceName)
                if (matchResource) {
                    // 增加或删除再增加重名
                    if (String(file.id).startsWith('local_')) {
                        // 防止通过导入而重名
                        const matchFile = matchResource.files.find(originFile => originFile.name === file.name)
                        if (!matchFile) {
                            file.action = 'create'
                            file.isMatch = true
                            delete file.id
                            matchResource.files.push(file)
                        } else {
                            if (String(matchFile.id).startsWith('local_')) {
                                delete matchFile.id
                                matchFile.action = 'create'
                            } else if (file.content !== matchFile.content || file.name !== fileName) {
                                matchFile.action = 'update'
                            }
                            matchFile.isMatch = true
                            matchFile.name = fileName
                            matchFile.content = file.content
                        }
                    } else if (matchResource.files.length) {
                        const matchFile = matchResource.files.find(originFile => String(originFile.id) === String(file.id))
                        // 更新
                        if (matchFile) {
                            matchFile.isMatch = true
                            if (file.content !== file.originContent) {
                                matchFile.action = 'update'
                                matchFile.content = file.content
                            }
                            if (matchFile.name !== fileName) {
                                matchFile.action = 'update'
                                matchFile.name = fileName
                            }
                        }
                    }
                } else {
                    params.template_files.push({
                        resource_name: resourceName,
                        files: [
                            {
                                action: 'create',
                                isMatch: true,
                                name: file.name,
                                content: file.content
                            }
                        ]
                    })
                }
            },
            getYamlParams () {
                const params = JSON.parse(JSON.stringify(this.curTemplate))

                params.template_files = params.template_files.filter(resource => {
                    return resource.files.length
                })
                this.yamlListCache = []
                this.getYamlFiles(this.defaultTreeData[0])
                this.getYamlFiles(this.customTreeData[0])
                this.yamlListCache.forEach(file => {
                    this.mergeParams(file, params)
                })
                params.template_files.forEach(resource => {
                    delete resource.actived

                    if (resource.files.length) {
                        resource.files.forEach((file, index) => {
                            if (!file.hasOwnProperty('isMatch')) {
                                // 经过前面合并对比，如果没匹配则表示删除
                                file.action = 'delete'

                                // 导入的时候会新建然后再删除，处理这种情况
                                if (String(file.id).startsWith('local_')) {
                                    resource.files.splice(index, 1)
                                }
                            }
                            delete file.isEdited
                        })
                    } else {
                        delete resource.files
                    }
                })

                // 选择旧版或者创建新版本
                if (this.saveVersionWay === 'old') {
                    const versionData = this.withoutCurVersionList.find(version => {
                        return version.show_version_id === this.selectedVersion
                    })
                    if (versionData) {
                        params.show_version = {
                            name: versionData.name
                        }
                    }
                } else if (this.saveVersionWay === 'new') {
                    params.show_version.name = this.versionKeyword
                }
                params.show_version.old_show_version_id = this.curShowVersionId
                delete params.show_version.show_version_id
                return params
            },

            /**
             * 创建yaml模板集
             */
            async createYamlTemplate () {
                const params = this.getYamlParams()
                params.show_version.comment = this.curVersionNotes

                this.isTemplateSaving = true
                try {
                    const res = await this.$store.dispatch('k8sTemplate/createYamlTemplate', {
                        projectId: this.projectId,
                        data: params
                    })

                    this.$router.push({
                        name: 'K8sYamlTemplateset',
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode,
                            templateId: res.data.template_id
                        }
                    })
                    this.hideVersionBox()
                } catch (e) {
                    this.isTemplateSaving = false
                    if (e.message && e.message.indexOf('file name is duplicated') > -1) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('文件名称不能重复')
                        })
                    } else {
                        catchErrorHandler(e, this)
                    }
                }
            },

            /**
             * 更新yaml模板集
             */
            async updateYamlTemplate () {
                const params = this.getYamlParams()
                params.show_version.comment = this.curVersionNotes

                this.isTemplateSaving = true
                try {
                    await this.$store.dispatch('k8sTemplate/updateYamlTemplate', {
                        projectId: this.projectId,
                        templateId: this.templateId,
                        data: params
                    })
                    // this.updateLocalYamlTemplate(res.data)
                    this.getYamlTemplateDetail()
                    this.getVersionList()
                    this.hideVersionBox()

                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('保存成功')
                    })
                } catch (e) {
                    this.isTemplateSaving = false
                    if (e.message && e.message.indexOf('file name is duplicated') > -1) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('文件名称不能重复')
                        })
                    } else {
                        catchErrorHandler(e, this)
                    }
                }
            },

            /**
             * 获取模板集详情
             */
            async getYamlTemplateDetail () {
                const originYamlParams = JSON.parse(JSON.stringify(this.yamlTemplateJson))

                try {
                    const res = await this.$store.dispatch('k8sTemplate/getYamlTemplateDetail', {
                        projectId: this.projectId,
                        templateId: this.templateId
                    })
                    this.webAnnotations = res.web_annotations || { perms: {} }
                    this.setCurTemplte(res.data)
                } catch (e) {
                    this.curTemplate = originYamlParams
                    catchErrorHandler(e, this)
                } finally {
                    this.isYamlTemplateLoading = false
                    this.isTemplateSaving = false
                }
            },

            setCurTemplte (template) {
                const originYamlParams = JSON.parse(JSON.stringify(this.yamlTemplateJson))
                const resourceList = originYamlParams.template_files
                let hasDefaultList = false
                let hasCustomList = false
                this.updatedTimestamp = template.updated_timestamp || 0
                template.template_files.forEach(resource => {
                    resource.files.forEach(file => {
                        file.action = 'unchange'
                        file.originContent = file.content
                        file.isEdited = false
                    })
                    if (resource.files.length) {
                        if (resource.resource_name === 'CustomManifest') {
                            hasCustomList = true
                        } else {
                            hasDefaultList = true
                        }
                    }
                })

                // 如果从没选过，设置默认tab
                if (!this.curTreeNode.value.id) {
                    if (hasCustomList && !hasDefaultList) {
                        this.tabName = 'custom'
                    } else {
                        this.tabName = 'default'
                    }
                }

                resourceList.forEach(resource => {
                    const targetResource = template.template_files.find(serverResource => serverResource.resource_name === resource.resource_name)
                    if (targetResource) {
                        resource.files = targetResource.files
                    }
                })

                template.template_files = resourceList
                this.curTemplate = template
                setTimeout(() => {
                    this.setDefaultEditFile()
                }, 1000)
            },

            selectNode (children) {
                const path = this.pathList.shift()
                children.forEach(node => {
                    if (node.name === path) {
                        if (this.pathList.length) {
                            this.selectNode(node.children || [])
                        } else {
                            node.parent.expanded = true
                            this.$set(node, 'selected', true)
                            this.curTreeNode = node
                        }
                    } else {
                        this.$set(node, 'selected', false)
                    }
                })
            },

            /**
             * 设置默认要编辑的资源文件
             */
            setDefaultEditFile () {
                if (this.curTreeNode.value.id && this.curTreeNode.value.fullName) {
                    this.pathList = this.curTreeNode.value.fullName.split('/')
                    let tree = this.defaultTreeData
                    if (this.curTreeNode.value.resourceName === 'CustomManifest') {
                        tree = this.customTreeData
                    }
                    this.selectNode(tree[0].children)
                }
                // 如果有上次目录记录
                // if (this.curResource.resource_name) {
                //     const activeResource = this.curTemplate.template_files.find(resource => {
                //         return resource.resource_name === this.curResource.resource_name && resource.files.length
                //     })

                //     if (activeResource) {
                //         let defaultFile = null
                //         // 如果上次文件有记录
                //         if (this.curResourceFile) {
                //             defaultFile = activeResource.files.find(file => {
                //                 return file.name === this.curResourceFile.name
                //             })
                //         }
                //         const curFile = defaultFile || activeResource.files[0]
                //         this.setCurResourceFile(activeResource, curFile)
                //     } else {
                //         this.setFirstActiveResource()
                //     }
                // } else {
                //     this.setFirstActiveResource()
                // }
            },

            setFirstActiveResource () {
                // 从其它目录有文件作为第一个编辑
                const activeResource = this.curTemplate.template_files.find(resource => {
                    return resource.files.length
                })
                if (activeResource) {
                    this.setCurResourceFile(activeResource, activeResource.files[0])
                }
            },

            showVersionBox () {
                this.versionDialogConf.isShow = true

                if (this.isNewTemplate) {
                    this.$nextTick(() => {
                        this.$refs.versionInput.focus()
                    })
                }
            },

            hideVersionBox () {
                this.saveVersionWay = 'cur'
                this.versionKeyword = ''
                this.selectedVersion = ''
                this.versionDialogConf.isShow = false
            },

            /**
             * 获取版本列表
             */
            async getVersionList () {
                const projectId = this.projectId
                const templateId = this.templateId

                if (templateId !== 0) {
                    this.isVersionListLoading = true
                    try {
                        const res = await this.$store.dispatch('k8sTemplate/getVersionList', { projectId, templateId })
                        if (res && res.data) {
                            const versionList = res.data
                            if (versionList) {
                                versionList.forEach(item => {
                                    if (item.show_version_id === Number(this.curShowVersionId) || item.show_version_id === this.curShowVersionId) {
                                        this.versionMetadata = {
                                            show_version_id: item.show_version_id,
                                            name: item.name,
                                            real_version_id: item.real_version_id
                                        }
                                        this.curVersionNotes = item.comment
                                        this.displayVersionNotes = item.comment
                                    }
                                })
                            }
                        }
                    } catch (e) {
                        this.$store.commit('k8sTemplate/updateVersionList', [])
                        catchErrorHandler(e, this)
                    } finally {
                        this.isVersionListLoading = false
                    }
                } else {
                    this.$store.commit('k8sTemplate/updateVersionList', [])
                }
            },

            /**
             * 离开当前编辑页面
             */
            handleBeforeLeave (callback) {
                this.goTemplateIndex()
            },

            goTemplateIndex () {
                // 清空数据
                this.$router.push({
                    name: 'templateset',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
            },

            /**
             * 实例化
             */
            createInstance () {
                this.curTemplate.edit_mode = 'yaml'
                this.$router.push({
                    name: 'instantiation',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        templateId: this.templateId,
                        curTemplate: this.curTemplate,
                        curShowVersionId: this.curShowVersionId
                    }
                })
            },

            showVersionPanel () {
                this.versionSidePanel.isShow = true
                this.getVersionList()
            },

            clearPrevContext () {
                this.curTreeNode = {
                    id: 0,
                    name: '',
                    value: {
                        id: 0,
                        content: '',
                        fullName: '',
                        originContent: ''
                    }
                }
            },

            /**
             * 获取相应版本的资源详情
             *
             * @param {Number} versionId 版本id
             * @param {Object} [varname] [description]
             */
            async getTemplateByVersion (versionId, isVersionRemove) {
                const projectId = this.projectId
                const templateId = this.templateId

                this.isEditFileName = false
                try {
                    const res = await this.$store.dispatch('k8sTemplate/getYamlTemplateDetailByVersion', { projectId, templateId, versionId })
                    this.clearPrevContext()
                    this.setCurTemplte(res.data)
                    // 如果不是操作删除版本，则可隐藏
                    if (!isVersionRemove) {
                        this.versionSidePanel.isShow = false
                    }
                } catch (e) {
                    this.$store.commit('k8sTemplate/updateVersionList', [])
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 导出相应版本的资源文件
             *
             * @param {Number} versionId 版本id
             */
            async exportTemplateByVersion (versionId) {
                const projectId = this.projectId
                const templateId = this.templateId

                try {
                    const res = await this.$store.dispatch('k8sTemplate/getYamlTemplateDetailByVersion', { projectId, templateId, versionId })
                    this.handleExport(res.data)
                } catch (e) {
                    this.$store.commit('k8sTemplate/updateVersionList', [])
                    catchErrorHandler(e, this)
                }
            },

            async handleFileInput (isClone) {
                const fileInput = isClone ? this.$refs.fileInputClone : this.$refs.fileInput
                if (fileInput.files && fileInput.files.length) {
                    try {
                        const file = fileInput.files[0]
                        const archive = await Archive.open(file)
                        const zipFile = await archive.extractFiles()
                        if (zipFile) {
                            this.handleToggleTab('default')
                            this.renderYamls(zipFile)
                            this.fileImportIndex++
                            this.clearPrevContext()
                        }
                    } catch (e) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('请选择合适的压缩包')
                        })
                    }
                }
            },

            renderYamls (zip, folderName = '') {
                const self = this
                for (const key in zip) {
                    const file = zip[key]
                    // 如果是文件
                    if (file.name) {
                        if (file.name.endsWith('.yaml') && !file.name.startsWith('._') && window.FileReader) {
                            const reader = new FileReader()
                            reader.onloadend = function (event) {
                                if (event.target.readyState === FileReader.DONE) {
                                    const content = event.target.result
                                    let fileName = file.name
                                    // 判断是否匹配到相应目录，否则为自定义
                                    let resource = self.curTemplate.template_files.find(resource => {
                                        // 如果是从系统导出的包来导入，目前以_开头，如_Deployment
                                        return resource.resource_name === folderName.substring(1)
                                    })
                                    // 自定义
                                    if (!resource) {
                                        resource = self.customTemplateFiles[0]
                                        fileName = `${folderName ? folderName + '/' : ''}${file.name}`
                                    }

                                    if (resource.resource_name === 'CustomManifest') {
                                        self.handleToggleTab('custom')
                                    }

                                    const matchFile = resource.files.find(item => item.name === file.name)
                                    // 查看是否有同名
                                    if (matchFile) {
                                        matchFile.content = content
                                        matchFile.action = 'update'
                                    } else {
                                        resource.files.push({
                                            id: `local_${+new Date()}`,
                                            name: fileName,
                                            content: content,
                                            originContent: content,
                                            action: 'create'
                                        })
                                    }

                                    self.$bkMessage({
                                        theme: 'success',
                                        message: self.$t('导入成功')
                                    })
                                }
                            }
                            reader.readAsText(file)
                        }
                    } else {
                        let catalogName = key
                        if (folderName && !folderName.startsWith('_')) {
                            catalogName = `${folderName}/${key}`
                        }
                        this.renderYamls(file, catalogName)
                    }
                }
            },

            handleExport (templateData) {
                const zip = new JSZip()
                const template = templateData || this.curTemplate
                template.template_files.forEach(resource => {
                    if (resource.files.length) {
                        // 用_开头表示是通过内部导出的文件
                        const folder = zip.folder(`_${resource.resource_name}`)

                        resource.files.forEach(file => {
                            if (file.action !== 'delete') {
                                folder.file(file.name, file.content, { binary: false })
                            }
                        })
                    }
                })

                zip.generateAsync({ type: 'blob' }).then((content) => {
                    saveAs(content, `${this.curTemplate.name || 'yaml'}_${this.curTemplate.show_version.name}.zip`)
                })
            },

            /**
             * 展示/隐藏变量面板
             */
            handleToggleVarPanel () {
                this.isVarPanelShow = !this.isVarPanelShow
                this.isImagePanelShow = false
            },

            /**
             * 展示/隐藏变量面板
             */
            handleToggleImagePanel () {
                this.isImagePanelShow = !this.isImagePanelShow
                this.isVarPanelShow = false
            },

            hideVarPanel () {
                this.isVarPanelShow = false
            },

            hideImagePanel () {
                this.isImagePanelShow = false
            },

            /**
             * 获取变量数据
             */
            async initVarList () {
                try {
                    await this.$store.dispatch('variable/getBaseVarList', this.projectId)
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 编辑当前资源文件名
             */
            handleEditFileName () {
                this.isEditFileName = true
                this.editFileNameTmp = this.curResourceFile.name

                this.$nextTick(() => {
                    const inputer = this.$refs.resourceFileNameInput
                    inputer.focus()
                    inputer.select()
                })
            },

            handleEditNameEnter () {
                const name = this.editFileNameTmp
                if (!this.checkFileName(this.curResource, this.curResourceFile, name)) {
                    this.$nextTick(() => {
                        const inputer = this.$refs.resourceFileNameInput
                        inputer.focus()
                        inputer.select()
                    })
                    return false
                }
                this.isEditFileName = false
                this.curResourceFile.name = name
                this.curResourceFile.action = 'update'
                this.isEnterTrigger = true
                this.editFileNameTmp = ''
            },

            handleEditNameBlur () {
                // 用setTimeout主要是考虑tab切换时，放弃编辑操作
                setTimeout(() => {
                    if (this.isEditFileName) {
                        const name = this.editFileNameTmp
                        if (this.isEnterTrigger) {
                            this.isEnterTrigger = false
                            return false
                        }
                        if (!this.checkFileName(this.curResource, this.curResourceFile, name)) {
                            this.$nextTick(() => {
                                const inputer = this.$refs.resourceFileNameInput
                                inputer.focus()
                                inputer.select()
                            })
                            return false
                        }

                        this.isEditFileName = false
                        this.curResourceFile.name = name
                        this.curResourceFile.action = 'update'
                        this.editFileNameTmp = ''
                    }
                }, 500)
            },
            handleImagePanel () {
                return false
            },
            updateTemplateLockStatus () {
                // 判断curTemplate name为空防止返回时清空当前数据解发switcher change事件
                if (this.isTemplateLocking || this.curTemplate.name === '') {
                    return false
                }
                if (this.templateLockStatus.isLocked) {
                    this.unlockTemplateset()
                } else {
                    this.lockTemplateset()
                }
            },

            async reloadTemplateLockStatus () {
                this.isTemplateLocking = true
                try {
                    const res = await this.$store.dispatch('k8sTemplate/getYamlTemplateDetail', {
                        projectId: this.projectId,
                        templateId: this.templateId
                    })
                    this.curTemplate.is_locked = res.data.is_locked
                    this.curTemplate.locker = res.data.locker
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isTemplateLocking = false
                }
            },

            async lockTemplateset () {
                const projectId = this.projectId
                const templateId = this.templateId
                this.isTemplateLocking = true
                try {
                    await this.$store.dispatch('k8sTemplate/lockTemplateset', { projectId, templateId })
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('加锁成功')
                    })
                    this.reloadTemplateLockStatus()
                } catch (res) {
                    this.$bkMessage({
                        theme: 'error',
                        message: res.message,
                        hasCloseIcon: true,
                        delay: '3000'
                    })
                } finally {
                    setTimeout(() => {
                        this.isTemplateLocking = false
                    }, 1000)
                }
            },

            async unlockTemplateset () {
                const projectId = this.projectId
                const templateId = this.templateId
                // 不是当前加锁者不能解锁
                if (!this.templateLockStatus.isCurLocker) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('{locker}正在操作，您如需编辑请联系{locker}解锁！', this.templateLockStatus)
                    })
                    return false
                }
                this.isTemplateLocking = true
                try {
                    await this.$store.dispatch('k8sTemplate/unlockTemplateset', { projectId, templateId })
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('解锁成功')
                    })
                    this.reloadTemplateLockStatus()
                } catch (res) {
                    this.$bkMessage({
                        theme: 'error',
                        message: res.message,
                        hasCloseIcon: true,
                        delay: '3000'
                    })
                } finally {
                    setTimeout(() => {
                        this.isTemplateLocking = false
                    }, 1000)
                }
            },

            /**
             * 获取文件详情
             * @param  {object} file 文件
             */
            getFileDetail (file) {
                if (file.hasOwnProperty('value')) {
                    this.curReourceFile = file
                }
            },

            handleToggleTab (name) {
                this.tabName = name
                this.isEditFileName = false
                this.$nextTick(() => {
                    this.clearCurResourfeFile()
                })
            },

            renderDefaultTree (node, ctx) {
                let titleClass = node.selected ? 'node-title node-selected' : 'node-title'

                // 有value属性表示为叶子节点（文件）
                if (node.value) {
                    const nodeId = `file_${node.value.id}`
                    return <div id={ nodeId }>
                        <span class={ titleClass } domPropsInnerHTML={ node.name } title={ node.name } onClick={() => this.handleSelectNode(node, 'default') }></span>
                        <div class="actions">
                            <span class="bcs-icon bcs-icon-more" onClick={this.preventEvent}></span>
                            <div class="menulist">
                                <div class="menu-item" onClick={() => this.handleRenameNode(node, 'file') }>{ this.$t('重命名') }</div>
                                <div class="menu-item" onClick={() => this.handleRemoveNode(node, 'default') }>{ this.$t('删除') }</div>
                            </div>
                        </div>
                    </div>
                } else if (node.parent) {
                    if (this.yamlResourceConf.resource_names.includes(node.name)) {
                        titleClass += ' node-catalog'
                    }
                    return <div>
                        <div class={ titleClass } title={ node.name }>
                            { node.name }
                            <span class="badge">({ node.children ? node.children.length : 0 })</span>
                        </div>
                        <div class="actions" onClick={this.preventEvent}>
                            <span class="bcs-icon bcs-icon-plus" onClick={() => this.handleAddNode(node, 'default') }></span>
                        </div>
                    </div>
                } else {
                    return <span></span>
                }
            },

            renderCustomTree (node, ctx) {
                let titleClass = node.selected ? 'node-title node-selected' : 'node-title'
                if (node.name === 'CustomManifest') {
                    titleClass += ' node-catalog'
                }
                if (node.value) {
                    return <div>
                        <span class={ titleClass } domPropsInnerHTML={ node.name } onClick={() => this.handleSelectNode(node, 'custom') } title={ node.name }></span>
                        <div class="actions" onClick={this.preventEvent}>
                            <span class="bcs-icon bcs-icon-more" onClick={this.preventEvent}></span>
                            <div class="menulist">
                                <div class="menu-item" onClick={() => this.handleRenameNode(node, 'file') }>{ this.$t('重命名') }</div>
                                <div class="menu-item" onClick={() => this.handleRemoveNode(node, 'custom') }>{ this.$t('删除') }</div>
                            </div>
                        </div>
                    </div>
                } else if (node.parent) {
                    if (node.name === 'CustomManifest') {
                        return <div>
                            <span class={ titleClass } domPropsInnerHTML={ node.name } title={ node.name }></span>
                            <div class="actions" onClick={this.preventEvent}>
                                <span class="bcs-icon bcs-icon-more" onClick={this.preventEvent}></span>
                                <div class="menulist">
                                    <div class="menu-item" onClick={() => this.handleAddCatalog(node, 'custom') }>{ this.$t('新增目录') }</div>
                                    <div class="menu-item" onClick={() => this.handleAddNode(node, 'custom') }>{ this.$t('新增文件') }</div>
                                </div>
                            </div>
                        </div>
                    } else {
                        return <div>
                            <span class={ titleClass } domPropsInnerHTML={ node.name } title={ node.name }></span>
                            <div class="actions" onClick={this.preventEvent}>
                                <span class="bcs-icon bcs-icon-more" onClick={this.preventEvent}></span>
                                <div class="menulist">
                                    <div class="menu-item" onClick={() => this.handleAddCatalog(node, 'custom') }>{ this.$t('新增目录') }</div>
                                    <div class="menu-item" onClick={() => this.handleAddNode(node, 'custom') }>{ this.$t('新增文件') }</div>
                                    <div class="menu-item" onClick={() => this.handleRenameNode(node, 'catalog') }>{ this.$t('重命名') }</div>
                                    <div class="menu-item" onClick={() => this.handleRemoveNode(node, 'custom') }>{ this.$t('删除') }</div>
                                </div>
                            </div>
                        </div>
                    }
                } else {
                    return <div>
                        <span class={ titleClass } domPropsInnerHTML={ node.name } title={ node.name }></span>
                        <div class="actions" onClick={this.preventEvent}>
                            <span class="bcs-icon bcs-icon-more" onClick={this.preventEvent}></span>
                            <div class="menulist">
                                <div class="menu-item" onClick={() => this.handleAddCatalog(node, 'custom') }>{ this.$t('新增目录') }</div>
                                <div class="menu-item" onClick={() => this.handleAddNode(node, 'custom') }>{ this.$t('新增文件') }</div>
                            </div>
                        </div>
                    </div>
                }
            },

            preventEvent (event) {
                event.stopPropagation()
            },

            handleAddNode (node, type) {
                const index = node.children ? node.children.length + 1 : 1
                const catalogName = node.name
                const name = `${catalogName.toLowerCase()}-${index}.yaml`

                this.curNode = node
                this.nodeDialogConf.action = 'addFile'
                this.nodeDialogConf.name = name
                this.nodeDialogConf.type = type
                this.nodeDialogConf.isShow = true
                setTimeout(() => {
                    this.$refs.newFileInputer.focus()
                }, 500)
            },

            handleAddCatalog (node, type) {
                this.curNode = node
                this.nodeDialogConf.action = 'addCatalog'
                this.nodeDialogConf.type = type
                this.nodeDialogConf.isShow = true
                setTimeout(() => {
                    this.$refs.newFileInputer.focus()
                }, 500)
            },

            addNode () {
                const { name, type, action } = this.nodeDialogConf
                const tree = type === 'default' ? this.$refs.defaultTree : this.$refs.customTree
                if (!name) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('名称不能为空')
                    })
                    return false
                }
                const node = (this.curNode.children || []).find(child => child.name === name)
                if (node) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('名称不能重复')
                    })
                    return false
                }
                if (!this.fileNameReg.test(name)) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('名称只能包含数字、字母、中划线(-)、下划线(_)、点(.)，开头结尾必须是数字或字母')
                    })
                    return false
                }

                if (action === 'addCatalog') {
                    this.curNode.openedIcon = 'icon-folder-open'
                    this.curNode.closedIcon = 'icon-folder'
                    this.curNode.expanded = true

                    tree.addNode(this.curNode, {
                        name: name,
                        title: name,
                        openedIcon: 'icon-folder-open',
                        closedIcon: 'icon-folder',
                        icon: 'icon-folder',
                        expanded: false,
                        id: uuid()
                    })
                } else {
                    if (!name.endsWith('.yaml') && !name.endsWith('.yml')) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('文件名称后缀以yaml、yml结尾')
                        })
                        return false
                    }

                    this.pathList = []
                    this.getAbsolutePath(this.curNode)
                    const resourceName = this.pathList.shift()
                    const content = this.yamlResourceConf.initial_templates[resourceName] || ''
                    const fileNamePrefix = this.pathList.join('/')
                    this.curNode.openedIcon = 'icon-folder-open'
                    this.curNode.closedIcon = 'icon-folder'
                    delete this.curNode.icon

                    const fileName = fileNamePrefix ? `${fileNamePrefix}/${name}` : name
                    const fileNode = {
                        id: uuid(),
                        name: name,
                        icon: 'icon-file',
                        selected: true,
                        value: {
                            id: `local_${+new Date()}`,
                            name: fileName,
                            fullName: `${resourceName}/${fileName}`,
                            content: content,
                            resourceName: resourceName,
                            originContent: content,
                            action: 'create'
                        }
                    }
                    tree.addNode(this.curNode, fileNode)
                    this.curTreeNode.selected = false
                    this.$set(this.curTreeNode, 'selected', false) // 把上次选择取消
                    this.curTreeNode = fileNode
                    setTimeout(() => {
                        this.setDefaultEditFile()
                    }, 1000)
                }

                this.hideNodeDialog()
            },

            hideNodeDialog () {
                this.nodeDialogConf.name = ''
                this.nodeDialogConf.action = ''
                this.nodeDialogConf.type = ''
                this.nodeDialogConf.isShow = false
            },

            handleRenameNode (node, type) {
                this.curNode = node
                this.fileDialogConf.action = 'rename'
                this.fileDialogConf.fileName = node.name
                this.fileDialogConf.isShow = true
                this.fileDialogConf.type = type
                setTimeout(() => {
                    this.$refs.renameInputer.focus()
                }, 500)
            },

            updateNode () {
                const { fileName } = this.fileDialogConf
                if (!fileName) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入文件名称')
                    })
                    return false
                }

                if (!this.fileNameReg.test(fileName)) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('文件名称只能包含数字、字母、中划线(-)、下划线(_)、点(.)，开头结尾必须是数字或字母')
                    })
                    return false
                }

                if (this.fileDialogConf.type === 'file') {
                    if (!fileName.endsWith('.yaml') && !fileName.endsWith('.yml')) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('文件名称后缀以yaml、yml结尾')
                        })
                        return false
                    }
                }

                const file = this.curNode.parent.children.find(child => child.name === fileName)
                if (file) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('名称不能重复')
                    })
                    return false
                }

                this.curNode.name = fileName

                if (this.curNode.value && this.curNode.value.fullName) {
                    const paths = this.curNode.value.fullName.split('/')
                    paths.pop()
                    paths.push(fileName)
                    this.curNode.value.fullName = paths.join('/')
                } else {
                    this.changeFileNodeName(this.curNode)
                }
                this.hideFileDialog()
            },

            changeFileNodeName (node) {
                if (node.children) {
                    node.children.forEach(child => {
                        if (child.value) {
                            this.pathList = []
                            this.getAbsolutePath(child)
                            child.value.fullName = this.pathList.join('/')
                        } else {
                            this.changeFileNodeName(child)
                        }
                    })
                }
            },

            hideFileDialog () {
                this.fileDialogConf.fileName = ''
                this.fileDialogConf.type = ''
                this.fileDialogConf.action = ''
                this.fileDialogConf.isShow = false
            },

            handleSelectNode (node, type) {
                const tree = type === 'default' ? this.$refs.defaultTree : this.$refs.customTree
                this.$set(this.curTreeNode, 'selected', false) // 把上次选择取消
                this.curTreeNode = node
                tree.nodeSelected(node)
                this.useEditorDiff = false
                this.reRenderEditor++
            },

            handleRemoveNode (node, type) {
                const self = this
                const tree = type === 'default' ? this.$refs.defaultTree : this.$refs.customTree
                this.$bkInfo({
                    title: this.$t('确认删除'),
                    content: this.$createElement('p', { style: { 'text-align': 'left' } }, `${this.$t('删除')} ${node.name}`),
                    confirmFn () {
                        self.pathList = []
                        self.getAbsolutePath(node)
                        const nodePath = self.pathList.join('/')
                        if (self.curTreeNode.value.fullName.startsWith(nodePath)) {
                            self.curTreeNode = {
                                id: 0,
                                name: '',
                                value: {
                                    id: 0,
                                    fullName: '',
                                    content: '',
                                    originContent: ''
                                }
                            }
                        }

                        tree.delNode(node.parent, node)
                        node.parent.icon = 'icon-folder'
                    }
                })
            },

            getAbsolutePath (node) {
                if (node.name && node.name !== '/') {
                    this.pathList.unshift(node.name)
                }
                if (node.parent) {
                    this.getAbsolutePath(node.parent)
                }
            },

            handleRenameBlur (node) {
                const input = document.getElementById(node.value.id)
                alert(input.value)
            },

            selectVersion (id, item) {
                this.curVersionNotes = item.comment
            }
            // removeVersion (data) {
            //     const self = this
            //     this.$bkInfo({
            //         title: `确认`,
            //         content: this.$createElement('p', { style: { 'text-align': 'center' } }, `删除版本：“${data.name}”`),
            //         confirmFn () {
            //             const projectId = self.projectId
            //             const templateId = self.templateId
            //             const versionId = data.show_version_id
            //             self.$store.dispatch('k8sTemplate/removeVersion', { projectId, templateId, versionId }).then(res => {
            //                 self.$bkMessage({
            //                     theme: 'success',
            //                     message: '操作成功！'
            //                 })

            //                 self.getVersionList().then(versionList => {
            //                     // 如果是删除当前版本
            //                     if (versionId === self.curShowVersionId || String(versionId) === self.curShowVersionId) {
            //                         // 加载第一项，优先选择非草稿
            //                         if (self.versionList.length) {
            //                             let versionData = self.versionList[0]
            //                             if (versionData.show_version_id === -1 && self.versionList.length > 1) {
            //                                 versionData = self.versionList[1]
            //                             }
            //                             self.getTemplateByVersion(versionData.show_version_id, true)
            //                         } else {
            //                             self.getTemplateByVersion(-1)
            //                         }
            //                     }
            //                 })
            //             }, res => {
            //                 this.$bkMessage({
            //                     theme: 'error',
            //                     message: res.message,
            //                     delay: '3000'
            //                 })
            //             })
            //         }
            //     })
            // }
        }
    }
</script>

<style lang="postcss" scoped>
    @import './index.css';
    @import '../header.css';
</style>
