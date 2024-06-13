<!-- eslint-disable max-len -->
<template>
  <BcsContent>
    <template #header>
      <biz-header
        ref="commonHeader"
        @exception="exceptionHandler"
        @saveStatefulsetSuccess="saveStatefulsetSuccess"
        @switchVersion="initResource"
        @exmportToYaml="exportToYaml">
      </biz-header>
    </template>
    <div class="biz-confignation-wrapper" v-bkloading="{ isLoading: isTemplateSaving }">
      <div class="biz-tab-box" v-show="!isDataLoading">
        <biz-tabs @tab-change="tabResource" ref="commonTab"></biz-tabs>
        <div class="biz-tab-content" v-bkloading="{ isLoading: isTabChanging }">
          <bk-alert type="info" class="mb20">
            <div slot="title">
              {{$t('deploy.templateset.statefulSetDescription')}}，
              <a class="bk-text-button" :href="PROJECT_CONFIG.k8sStatefulset" target="_blank">{{$t('plugin.tools.docs')}}</a>
            </div>
          </bk-alert>
          <template v-if="!statefulsets.length">
            <div class="biz-guide-box mt0">
              <bk-button icon="plus" type="primary" @click.stop.prevent="addLocalApplication">
                <span style="margin-left: 0;">{{$t('generic.button.add')}}StatefulSet</span>
              </bk-button>
            </div>
          </template>
          <template v-else>
            <div class="biz-configuration-topbar">
              <div class="biz-list-operation">
                <div class="item" v-for="(application, index) in statefulsets" :key="application.id">
                  <bk-button :class="['bk-button', { 'bk-primary': curApplication.id === application.id }]" @click.stop="setCurApplication(application, index)">
                    {{(application && application.config.metadata.name) || $t('deploy.templateset.unnamed')}}
                    <span class="biz-update-dot" v-show="application.isEdited"></span>
                  </bk-button>
                  <span class="bcs-icon bcs-icon-close" @click.stop="removeApplication(application, index)"></span>
                </div>

                <bcs-popover ref="applicationTooltip" :content="$t('deploy.templateset.addStatefulSet')" placement="top">
                  <bk-button class="bk-button bk-default is-outline is-icon" @click.stop="addLocalApplication">
                    <i class="bcs-icon bcs-icon-plus"></i>
                  </bk-button>
                </bcs-popover>
              </div>
            </div>
            <div class="biz-configuration-content" style="position: relative;">
              <!-- part1 start -->
              <div class="bk-form biz-configuration-form">
                <a href="javascript:void(0);" class="bk-text-button from-json-btn" @click.stop.prevent="showJsonPanel">{{$t('deploy.templateset.importYAML')}}</a>

                <bk-sideslider
                  :is-show.sync="toJsonDialogConf.isShow"
                  :title="toJsonDialogConf.title"
                  :width="toJsonDialogConf.width"
                  :quick-close="false"
                  class="biz-app-container-tojson-sideslider"
                  @hidden="closeToJson">
                  <div slot="content" style="position: relative;">
                    <div class="biz-log-box" :style="{ height: `${winHeight - 60}px` }" v-bkloading="{ isLoading: toJsonDialogConf.loading }">
                      <bk-button class="bk-button bk-primary save-json-btn" @click.stop.prevent="saveApplicationJson">{{$t('generic.button.import')}}</bk-button>
                      <bk-button class="bk-button bk-default hide-json-btn" @click.stop.prevent="hideApplicationJson">{{$t('generic.button.cancel')}}</bk-button>
                      <ace
                        :value="editorConfig.value"
                        :width="editorConfig.width"
                        :height="editorConfig.height"
                        :lang="editorConfig.lang"
                        :read-only="editorConfig.readOnly"
                        :full-screen="editorConfig.fullScreen"
                        @init="editorInitAfter">
                      </ace>
                    </div>
                  </div>
                </bk-sideslider>

                <div class="bk-form-item">
                  <div class="bk-form-item">
                    <div class="bk-form-content" style="margin-left: 0;">
                      <div class="bk-form-inline-item is-required">
                        <label class="bk-label" style="width: 140px;">{{$t('plugin.tools.appName')}}：</label>
                        <div class="bk-form-content" style="margin-left: 140px;">
                          <div class="bk-form-input-group">
                            <input type="text" :class="['bk-form-input',{ 'is-danger': errors.has('applicationName') }]" :placeholder="$t('deploy.templateset.enterCharacterLimit64')" style="width: 310px;" v-model="curApplication.config.metadata.name" maxlength="64" name="applicationName" v-validate="{ required: true, regex: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/ }">
                          </div>
                        </div>
                      </div>
                      <div class="bk-form-inline-item is-required">
                        <label class="bk-label" style="width: 140px;">{{$t('dashboard.workload.label.scaleNum')}}：</label>
                        <div class="bk-form-content" style="margin-left: 140px;">
                          <div class="bk-form-input-group">
                            <bkbcs-input
                              type="number"
                              :placeholder="$t('generic.placeholder.input')"
                              style="width: 310px;"
                              :min="0"
                              :value.sync="curApplication.config.spec.replicas"
                              :list="varList"
                            >
                            </bkbcs-input>
                          </div>
                        </div>
                      </div>
                      <div class="bk-form-tip is-danger" style="margin-left: 140px;" v-if="errors.has('applicationName')">
                        <p class="bk-tip-text">{{$t('deploy.templateset.nameIsRequired')}}</p>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="bk-form-item" v-if="curApplication.service_tag">
                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.service')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <div class="bk-dropdown-box" style="width: 310px;" @click="reloadServices">
                      <!-- <input type="text" class="bk-form-input" :value="linkServiceName" disabled> -->
                      <bk-selector
                        :placeholder="$t('deploy.templateset.selectAssociatedService')"
                        :setting-key="'service_tag'"
                        :display-key="'service_name'"
                        :selected.sync="curApplication.service_tag"
                        :list="serviceList"
                        :prevent-init-trigger="'true'"
                        :is-loading="isLoadingServices"
                        :disabled="true"
                      >
                      </bk-selector>
                    </div>
                  </div>
                </div>

                <div class="bk-form-item">
                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.importanceLevel')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <bk-radio-group v-model="curApplication.config.monitorLevel">
                      <bk-radio class="mr20" :value="'important'">{{$t('deploy.templateset.important')}}</bk-radio>
                      <bk-radio class="mr20" :value="'general'">{{$t('deploy.templateset.general')}}</bk-radio>
                      <bk-radio :value="'unimportant'">{{$t('deploy.templateset.notImportant')}}</bk-radio>
                    </bk-radio-group>
                  </div>
                </div>

                <div class="bk-form-item">
                  <label class="bk-label" style="width: 140px;">{{$t('cluster.create.label.desc')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <textarea class="bk-form-textarea" :placeholder="$t('deploy.templateset.enterCharacterLimit256')" v-model="curApplication.desc" maxlength="256"></textarea>
                  </div>
                </div>

                <div class="bk-form-item is-required">
                  <label class="bk-label" style="width: 140px;">{{$t('k8s.label')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <bk-keyer
                      :key-list.sync="curLabelList"
                      :var-list="varList"
                      ref="labelKeyer"
                      @change="updateApplicationLabel"
                      :is-link-to-selector="true"
                      :is-tip-change="isSelectorChange"
                      :tip="$t('deploy.templateset.tips2')"
                      :can-disabled="true">
                      <slot>
                        <p class="biz-tip" style="line-height: 1;">{{$t('deploy.templateset.tips1')}}</p>
                      </slot>
                    </bk-keyer>
                  </div>
                </div>

                <div class="bk-form-item">
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isMorePanelShow }]" @click.stop.prevent="toggleMore">
                      {{$t('deploy.templateset.moreSettings')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                    </button>

                    <button :class="['bk-text-button f12 mb10 pl0', { 'rotate': isPodPanelShow }]" @click.stop.prevent="togglePod">
                      {{$t('deploy.templateset.podTemplateSettings')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                    </button>
                  </div>
                  <bk-tab :type="'fill'" :active-name="'tab1'" :size="'small'" v-show="isMorePanelShow" style="margin-left: 140px;">

                    <bk-tab-panel name="tab1" :title="$t('deploy.templateset.updateStrategy')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 155px;">{{$t('generic.label.kind')}}：</label>
                          <div class="bk-form-content" style="margin-left: 155px;">
                            <bk-radio-group v-model="curApplication.config.spec.updateStrategy.type">
                              <bk-radio :value="'OnDelete'">
                                OnDelete
                              </bk-radio>
                              <bk-radio :value="'RollingUpdate'">
                                RollingUpdate
                              </bk-radio>
                            </bk-radio-group>
                          </div>
                        </div>

                        <div class="bk-form-item" v-show="curApplication.config.spec.updateStrategy.type === 'RollingUpdate'">
                          <label class="bk-label" style="width: 155px;">Partition：</label>
                          <div class="bk-form-content" style="margin-left: 155px;">
                            <bkbcs-input
                              type="number"
                              :placeholder="$t('generic.placeholder.input')"
                              style="width: 250px;"
                              :min="0"
                              :max="curApplication.config.spec.replicas - 1"
                              :value.sync="curApplication.config.spec.updateStrategy.rollingUpdate.partition"
                              :list="varList"
                            >
                            </bkbcs-input>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>
                    <bk-tab-panel name="tab2" :title="$t('dashboard.workload.pods.podManagementPolicy')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 210px;">PodManagementPolicy：</label>
                          <div class="bk-form-content" style="margin-left: 210px;">
                            <bk-radio-group v-model="curApplication.config.spec.podManagementPolicy">
                              <bk-radio :value="'OrderedReady'">OrderedReady</bk-radio>
                              <bk-radio :value="'Parallel'">Parallel</bk-radio>
                            </bk-radio-group>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>
                    <bk-tab-panel name="ta5" :title="'HostAliases'">
                      <div class="bk-form m20">
                        <table class="biz-simple-table" style="width: 720px;" v-if="curApplication.config.webCache.hostAliasesCache && curApplication.config.webCache.hostAliasesCache.length">
                          <thead>
                            <tr>
                              <th style="width: 200px;">IP</th>
                              <th style="width: 400px;">HostNames</th>
                              <th></th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr v-for="(hostAlias, index) in curApplication.config.webCache.hostAliasesCache" :key="index">
                              <td>
                                <bkbcs-input
                                  type="text"
                                  :placeholder="$t('generic.placeholder.input')"
                                  :value.sync="hostAlias.ip"
                                  :list="varList">
                                </bkbcs-input>
                              </td>
                              <td>
                                <bkbcs-input
                                  type="text"
                                  :placeholder="$t('deploy.templateset.enterHostnamesSemicolon')"
                                  :value.sync="hostAlias.hostnames"
                                  :list="varList">
                                </bkbcs-input>
                              </td>
                              <td>
                                <div class="action-box">
                                  <bk-button class="action-btn ml5" @click.stop.prevent="addHostAlias">
                                    <i class="bcs-icon bcs-icon-plus"></i>
                                  </bk-button>
                                  <bk-button class="action-btn" @click.stop.prevent="removeHostAlias(hostAlias, index)">
                                    <i class="bcs-icon bcs-icon-minus"></i>
                                  </bk-button>
                                </div>
                              </td>
                            </tr>
                          </tbody>
                        </table>
                        <div class="tc p40" v-else>
                          <bk-button type="primary" @click="addHostAlias">
                            <i class="bcs-icon bcs-icon-plus f12" style="top: -2px;"></i>
                            {{$t('deploy.templateset.addHostAlias')}}
                          </bk-button>
                        </div>
                      </div>
                    </bk-tab-panel>
                  </bk-tab>

                  <bk-tab :type="'fill'" :active-name="'tab2'" :size="'small'" v-show="isPodPanelShow" style="margin-left: 105px;">
                    <bk-tab-panel name="tab2" :title="$t('k8s.annotation')">
                      <div class="bk-form m20">
                        <bk-keyer :key-list.sync="curRemarkList" :var-list="varList" ref="remarkKeyer" @change="updateApplicationRemark"></bk-keyer>
                      </div>
                    </bk-tab-panel>
                    <bk-tab-panel name="tab3" :title="$t('deploy.templateset.restartStrategy')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 105px;">{{$t('deploy.templateset.restartPolicy')}}：</label>
                          <div class="bk-form-content" style="margin-left: 105px;">
                            <bk-radio-group v-model="curApplication.config.spec.template.spec.restartPolicy">
                              <bk-radio :value="policy" v-for="(policy, index) in restartPolicy" :key="index">
                                {{policy}}
                              </bk-radio>
                            </bk-radio-group>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>
                    <bk-tab-panel name="tab4" :title="$t('deploy.templateset.killPolicy')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 105px;">{{$t('deploy.templateset.terminationGracePeriod')}}：</label>
                          <div class="bk-form-content" style="margin-left: 105px;">
                            <div class="bk-form-input-group">
                              <bkbcs-input
                                type="number"
                                :placeholder="$t('generic.placeholder.input')"
                                style="width: 80px;"
                                :min="0"
                                :value.sync="curApplication.config.spec.template.spec.terminationGracePeriodSeconds"
                                :list="varList"
                              >
                              </bkbcs-input>
                              <span class="input-group-addon">
                                {{$t('units.suffix.seconds')}}
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab5" :title="$t('deploy.templateset.schedulingConstraints')">
                      <div class="bk-form m20">
                        <p class="title mb5">NodeSelector</p>
                        <bk-keyer :key-list.sync="curConstraintLabelList" :var-list="varList" ref="nodeSelectorKeyer" @change="updateNodeSelectorList"></bk-keyer>
                        <div class="mb5 mt10">
                          <span class="title">{{$t('deploy.templateset.affinityConstraints')}}</span>
                          <bk-checkbox class="ml10" name="image-get" v-model="curApplication.config.webCache.isUserConstraint">{{$t('logCollector.action.enable')}}</bk-checkbox>
                        </div>

                        <div style="height: 300px;" v-if="curApplication.config.webCache.isUserConstraint">
                          <ace
                            :value="curApplication.config.webCache.affinityYaml"
                            :width="yamlEditorConfig.width"
                            :height="yamlEditorConfig.height"
                            :lang="yamlEditorConfig.lang"
                            :read-only="yamlEditorConfig.readOnly"
                            :full-screen="yamlEditorConfig.fullScreen"
                            @init="yamlEditorInitAfter"
                            @input="yamlEditorInput"
                            @blur="yamlEditorBlur">
                          </ace>
                        </div>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab6" :title="$t('k8s.networking')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 105px;">{{$t('deploy.templateset.networkPolicy')}}：</label>
                          <div class="bk-form-content" style="margin-left: 105px;">
                            <bk-selector
                              style="width: 300px;"
                              :placeholder="$t('generic.placeholder.select')"
                              :setting-key="'id'"
                              :display-key="'name'"
                              :selected.sync="curApplication.config.spec.template.spec.hostNetwork"
                              :list="netStrategyList"
                              @item-selected="changeDNSPolicy">
                            </bk-selector>
                          </div>
                        </div>
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 105px;">{{$t('deploy.templateset.dnsPolicy')}}：</label>
                          <div class="bk-form-content" style="margin-left: 105px;">
                            <bk-selector
                              style="width: 300px;"
                              :placeholder="$t('generic.placeholder.select')"
                              :setting-key="'id'"
                              :display-key="'name'"
                              :selected.sync="curApplication.config.spec.template.spec.dnsPolicy"
                              :list="dnsStrategyList">
                            </bk-selector>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="ta7" :title="$t('deploy.templateset.volume')">
                      <div class="bk-form m20">
                        <table class="biz-simple-table" v-if="curApplication.config.webCache.volumes.length">
                          <thead>
                            <tr>
                              <th style="width: 200px;">{{$t('generic.label.kind')}}</th>
                              <th style="width: 220px;">{{$t('deploy.templateset.mountName')}}</th>
                              <th>{{$t('deploy.templateset.mountSource')}}</th>
                              <th style="width: 100px;"></th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr v-for="(volume, index) in curApplication.config.webCache.volumes" :key="index">
                              <td>
                                <bk-selector
                                  :placeholder="$t('generic.label.kind')"
                                  :setting-key="'id'"
                                  :selected.sync="volume.type"
                                  :list="volumeTypeList">
                                </bk-selector>
                              </td>
                              <td>
                                <bkbcs-input
                                  type="text"
                                  :placeholder="$t('generic.placeholder.input')"
                                  :value.sync="volume.name"
                                  :list="varList"
                                >
                                </bkbcs-input>
                              </td>
                              <td>
                                <template v-if="volume.type === 'emptyDir'">
                                  <bkbcs-input value="{}" :disabled="true" />
                                </template>
                                <template v-if="volume.type === 'emptyDir(Memory)'">
                                  <div class="source-flex-box">
                                    <bkbcs-input value="Memory" :disabled="true" style="width: 75px;" />
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        placeholder="sizeLimit"
                                        :min="0"
                                        :value.sync="volume.source"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        Gi
                                      </span>
                                    </div>
                                  </div>
                                </template>
                                <template v-else-if="volume.type === 'persistentVolumeClaim'">
                                  <bk-selector
                                    placeholder="PVC List"
                                    :setting-key="'id'"
                                    :searchable="true"
                                    :selected.sync="volume.source"
                                    :list="[]">
                                  </bk-selector>
                                </template>
                                <template v-else-if="volume.type === 'hostPath'">
                                  <bkbcs-input v-model="volume.source" :placeholder="$t('generic.placeholder.input')" />
                                </template>
                                <template v-else-if="volume.type === 'configMap'">
                                  <bk-selector
                                    placeholder="Configmap List"
                                    :setting-key="'id'"
                                    :display-key="'name'"
                                    :searchable="true"
                                    :selected.sync="volume.source"
                                    :list="volumeConfigmapAllList">
                                  </bk-selector>
                                </template>
                                <template v-else-if="volume.type === 'secret'">
                                  <bk-selector
                                    placeholder="Secret List"
                                    :setting-key="'name'"
                                    :display-key="'name'"
                                    :searchable="true"
                                    :selected.sync="volume.source"
                                    :list="volumeSecretList">
                                  </bk-selector>
                                </template>
                              </td>
                              <td>
                                <div class="action-box">
                                  <bk-button class="action-btn ml5" @click.stop.prevent="addVolumn()">
                                    <i class="bcs-icon bcs-icon-plus"></i>
                                  </bk-button>
                                  <bk-button class="action-btn" @click.stop.prevent="removeVolumn(volume, index)">
                                    <i class="bcs-icon bcs-icon-minus"></i>
                                  </bk-button>
                                </div>
                              </td>
                            </tr>
                          </tbody>
                        </table>
                        <div class="tc p40" v-else>
                          <bk-button type="primary" @click="addVolumn">
                            <i class="bcs-icon bcs-icon-plus f12" style="top: -2px;"></i>
                            {{$t('deploy.templateset.addVolume')}}
                          </bk-button>
                        </div>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab8" :title="$t('logCollector.text')">
                      <div class="bk-form p20">
                        <div class="biz-expand-panel">
                          <div class="panel">
                            <div class="header">
                              <span class="topic">{{$t('plugin.tools.standardLog')}}</span>
                            </div>
                            <div class="bk-form-item content">
                              <ul>
                                <li>
                                  <bk-checkbox name="type" :value="true" :disabled="true">{{$t('deploy.templateset.standardOutput')}}</bk-checkbox>
                                </li>
                              </ul>
                            </div>
                          </div>
                          <div class="panel mt0">
                            <div class="header" style="border-top: 1px solid #dfe0e5;">
                              <div class="topic">
                                {{$t('logCollector.label.extraLabels')}}
                                <bcs-popover :content="$t('deploy.templateset.additionalLogTags')" placement="top">
                                  <span class="bk-badge">
                                    <i class="bcs-icon bcs-icon-question-circle"></i>
                                  </span>
                                </bcs-popover>
                              </div>
                            </div>
                            <div class="bk-form-item content">
                              <bk-keyer
                                :key-list.sync="curLogLabelList"
                                :var-list="varList"
                                @change="updateApplicationLogLabel">
                              </bk-keyer>
                            </div>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab9" :title="$t('deploy.templateset.imageCredential')">
                      <div class="bk-form m20">
                        <bk-keyer
                          :data-key="'name'"
                          :key-list.sync="curImageSecretList"
                          :var-list="varList"
                          :key-input-width="170"
                          :value-input-width="450"
                          :tip="$t('deploy.templateset.imagePullSecretsNote')"
                          @change="updateApplicationImageSecrets">
                        </bk-keyer>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab10" :title="$t('deploy.templateset.serviceAccount')">
                      <div class="bk-form m20">
                        <div class="biz-equal-inputer">
                          <div class="inputer-content">
                            <bkbcs-input
                              type="text"
                              style="width: 170px;"
                              value="serviceAccountName"
                              :disabled="true">
                            </bkbcs-input>
                            <span class="operator">=</span>
                            <bkbcs-input
                              type="text"
                              style="width: 450px;"
                              :placeholder="$t('generic.label.value')"
                              :value.sync="curApplication.config.spec.template.spec.serviceAccountName">
                            </bkbcs-input>
                          </div>
                          <p class="biz-tip mt5">{{$t('deploy.templateset.createPodServiceAccountReminder')}}</p>
                        </div>
                      </div>
                    </bk-tab-panel>
                  </bk-tab>
                </div>
              </div>
              <!-- part1 end -->

              <!-- part2 start -->
              <div class="biz-part-header">
                <div class="bk-button-group">
                  <div class="item" v-for="(container, index) in curApplication.config.spec.template.spec.allContainers" :key="index">
                    <bk-button :class="['bk-button bk-default is-outline', { 'is-selected': curContainerIndex === index }]" @click.stop="setCurContainer(container, index)">
                      {{container.name || $t('deploy.templateset.unnamed')}}
                    </bk-button>
                    <span class="bcs-icon bcs-icon-close-circle" @click.stop="removeContainer(index)" v-if="curApplication.config.spec.template.spec.allContainers.length > 1"></span>
                  </div>
                  <bcs-popover ref="containerTooltip" :content="$t('deploy.templateset.addContainer')" placement="top">
                    <bk-button type="button" class="bk-button bk-default is-outline is-icon" @click.stop.prevent="addLocalContainer">
                      <i class="bcs-icon bcs-icon-plus"></i>
                    </bk-button>
                  </bcs-popover>
                </div>
              </div>

              <div class="bk-form biz-configuration-form pb15">
                <div class="biz-span">
                  <span class="title">{{$t('generic.title.basicInfo')}}</span>
                </div>
                <div class="bk-form-item is-required">
                  <div class="bk-form-content" style="margin-left: 0">
                    <div class="bk-form-inline-item is-required">
                      <label class="bk-label" style="width: 140px;">{{$t('dashboard.workload.container.name')}}：</label>
                      <div class="bk-form-content" style="margin-left: 140px;">
                        <input type="text" :class="['bk-form-input', { 'is-danger': errors.has('containerName') }]" :placeholder="$t('deploy.templateset.enterCharacterLimit64')" style="width: 310px;" v-model="curContainer.name" maxlength="64" name="containerName" v-validate="{ required: true, regex: /^[a-z]{1}[a-z0-9-]{0,63}$/ }">
                      </div>
                    </div>

                    <div class="bk-form-inline-item">
                      <label class="bk-label" style="width: 105px;">{{$t('generic.label.kind')}}：</label>
                      <div class="bk-form-content" style="margin-left: 105px;">
                        <bk-radio-group v-model="curContainer.webCache.containerType">
                          <bk-radio :value="'container'">
                            Container
                            <i class="bcs-icon bcs-icon-question-circle ml5" v-bk-tooltips="$t('deploy.templateset.appContainer')"></i>
                          </bk-radio>
                          <bk-radio :value="'initContainer'">
                            InitContainer
                            <i class="bcs-icon bcs-icon-question-circle ml5" v-bk-tooltips="$t('deploy.templateset.preAppContainer')"></i>
                          </bk-radio>
                        </bk-radio-group>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="bk-form-item">
                  <label class="bk-label" style="width: 140px;">{{$t('cluster.create.label.desc')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <textarea name="" id="" cols="30" rows="10" class="bk-form-textarea" :placeholder="$t('deploy.templateset.enterCharacterLimit256')" v-model="curContainer.webCache.desc" maxlength="256"></textarea>
                  </div>
                </div>
                <div class="bk-form-item is-required">
                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.imageAndVersion')}}：</label>
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <!-- <div class="mb10">
                        <span @click="handleChangeImageMode">
                          <bk-switcher
                            :selected="curContainer.webCache.isImageCustomed"
                            size="small"
                            :key="curContainer.name">
                          </bk-switcher>
                        </span>
                        <span class="vm">{{$t('deploy.templateset.useCustomImage')}}</span>
                        <span class="biz-tip vm">({{$t('deploy.templateset.enableDirectImageInput')}})</span>
                      </div> -->
                    <template>
                      <bkbcs-input
                        type="text"
                        style="width: 325px;"
                        :placeholder="$t('k8s.image')"
                        :value.sync="curContainer.webCache.imageName"
                        @change="handleImageCustom">
                      </bkbcs-input>
                      <bkbcs-input
                        type="text"
                        style="width: 250px;"
                        :placeholder="$t('deploy.templateset.versionNumber')"
                        class="ml5"
                        :value.sync="curContainer.imageVersion"
                        @change="handleImageCustom">
                      </bkbcs-input>
                    </template>

                    <bk-checkbox
                      class="ml10"
                      name="image-get"
                      :true-value="'Always'"
                      :false-value="'IfNotPresent'"
                      v-model="curContainer.imagePullPolicy">
                      {{$t('deploy.templateset.alwaysPullImageBeforeCreating')}}
                    </bk-checkbox>

                    <!-- <p class="biz-tip mt5" v-if="!isLoadingImageList && !imageList.length">{{$t('deploy.templateset.imageNotFoundError')}}
                        <router-link class="bk-text-button" :to="{ name: 'projectImage', params: { projectCode, projectId } }">{{$t('deploy.templateset.goCreate')}}</router-link>
                      </p> -->
                  </div>
                </div>

                <div class="biz-span">
                  <span class="title">{{$t('dashboard.network.portmapping')}}</span>
                </div>

                <div class="bk-form-item">
                  <div class="bk-form-content" style="margin-left: 140px;">
                    <table class="biz-simple-table">
                      <thead>
                        <tr>
                          <th style="width: 330px;">{{$t('generic.label.name')}}</th>
                          <th style="width: 135px;">{{$t('deploy.templateset.protocol')}}</th>
                          <th style="width: 135px;">{{$t('deploy.templateset.containerPort')}}</th>
                          <th></th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(port, index) in curContainer.ports" :key="index">
                          <td>
                            <bkbcs-input
                              type="text"
                              :placeholder="$t('generic.label.name')"
                              style="width: 325px;"
                              maxlength="255"
                              :value.sync="port.name"
                              :list="varList"
                            >
                            </bkbcs-input>
                          </td>
                          <td>
                            <bk-selector
                              :placeholder="$t('deploy.templateset.protocol')"
                              :setting-key="'id'"
                              :selected.sync="port.protocol"
                              :list="protocolList">
                            </bk-selector>
                          </td>
                          <td>
                            <bkbcs-input
                              type="number"
                              placeholder="1-65535"
                              style="width: 135px;"
                              :value.sync="port.containerPort"
                              :min="1"
                              :max="65535"
                              :list="varList"
                            >
                            </bkbcs-input>
                          </td>
                          <td>
                            <bk-button class="action-btn ml5" @click.stop.prevent="addPort">
                              <i class="bcs-icon bcs-icon-plus"></i>
                            </bk-button>
                            <bk-button class="action-btn" v-if="curContainer.ports.length > 1" @click.stop.prevent="removePort(port, index)">
                              <i class="bcs-icon bcs-icon-minus"></i>
                            </bk-button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                    <p class="biz-tip">{{$t('deploy.templateset.exposeServiceTargetPort')}}</p>
                  </div>
                </div>

                <div class="biz-span">
                  <div class="title">
                    <button :class="['bk-text-button', { 'rotate': isPartBShow }]" @click.stop.prevent="togglePartB">
                      {{$t('deploy.templateset.moreSettings')}}<i class="bcs-icon bcs-icon-angle-double-down f12 ml5 mb10 fb"></i>
                    </button>
                  </div>
                </div>

                <div style="margin-left: 140px;" v-show="isPartBShow">
                  <bk-tab :type="'fill'" :active-name="'tab1'" :size="'small'">
                    <bk-tab-panel name="tab1" :title="$t('dashboard.workload.container.command')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.startupCommand')}}：</label>
                          <div class="bk-form-content" style="margin-left: 140px;">
                            <bkbcs-input
                              type="text"
                              :placeholder="$t('deploy.templateset.exampleBash')"
                              :value.sync="curContainer.command"
                              :list="varList"
                            >
                            </bkbcs-input>
                          </div>
                        </div>
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.commandParams')}}：</label>
                          <div class="bk-form-content" style="margin-left: 140px;">
                            <bkbcs-input
                              type="text"
                              :placeholder="$t('generic.placeholder.command')"
                              :value.sync="curContainer.args"
                              :list="varList"
                            >
                            </bkbcs-input>
                          </div>
                        </div>
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.workingDirectory')}}：</label>
                          <div class="bk-form-content" style="margin-left: 140px;">
                            <bkbcs-input
                              type="text"
                              :placeholder="$t('deploy.templateset.examplePathParam', { path: '/mywork' })"
                              :value.sync="curContainer.workingDir"
                              :list="varList"
                            >
                            </bkbcs-input>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab2" :title="$t('k8s.volume')">
                      <div class="bk-form m20">
                        <template v-if="curMountVolumes.length">
                          <template v-if="curContainer.volumeMounts.length">
                            <p class="biz-tip mb10">
                              {{$t('deploy.templateset.msg.mountVolumes')}}</p>
                            <table class="biz-simple-table">
                              <thead>
                                <tr>
                                  <th style="width: 200px;">{{$t('deploy.templateset.volume')}}</th>
                                  <th style="width: 300px;">{{$t('dashboard.workload.container.dataDir')}}</th>
                                  <th style="width: 200px;">{{$t('deploy.templateset.subDirectory')}}</th>
                                  <th style="width: 70px;"></th>
                                  <th></th>
                                </tr>
                              </thead>
                              <tbody>
                                <tr v-for="(volumeItem, index) in curContainer.volumeMounts" :key="index">
                                  <td>
                                    <bk-selector
                                      :placeholder="$t('generic.placeholder.select')"
                                      :setting-key="'name'"
                                      :display-key="'name'"
                                      :allow-clear="true"
                                      :selected.sync="volumeItem.name"
                                      :list="curMountVolumes"
                                      @item-selected="selectVolumeType(volumeItem)">
                                    </bk-selector>
                                  </td>
                                  <td>
                                    <bkbcs-input
                                      type="text"
                                      placeholder="MountPath"
                                      maxlength="512"
                                      :value.sync="volumeItem.mountPath"
                                      :list="varList"
                                    >
                                    </bkbcs-input>
                                  </td>
                                  <td>
                                    <bkbcs-input
                                      type="text"
                                      placeholder="SubPath"
                                      maxlength="200"
                                      :value.sync="volumeItem.subPath"
                                      :list="varList">
                                    </bkbcs-input>
                                  </td>
                                  <td>
                                    <div class="biz-input-wrapper">
                                      <bk-checkbox v-model="volumeItem.readOnly">{{$t('deploy.templateset.readOnly')}}</bk-checkbox>
                                    </div>
                                  </td>
                                  <div class="action-box">
                                    <bk-button class="action-btn ml5" @click.stop.prevent="addMountVolumn()">
                                      <i class="bcs-icon bcs-icon-plus"></i>
                                    </bk-button>
                                    <bk-button class="action-btn" @click.stop.prevent="removeMountVolumn(volumeItem, index)">
                                      <i class="bcs-icon bcs-icon-minus"></i>
                                    </bk-button>
                                  </div>
                                </tr>
                              </tbody>
                            </table>
                          </template>
                          <div class="tc p40" v-else>
                            <bk-button type="primary" @click="addMountVolumn">
                              <i class="bcs-icon bcs-icon-plus f12" style="top: -2px;"></i>
                              {{$t('deploy.templateset.addMountVolume')}}
                            </bk-button>
                          </div>
                        </template>
                        <div v-else class="tc p30">
                          {{$t('deploy.templateset.msg.mountVolumes')}}
                        </div>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab3" :title="$t('dashboard.workload.container.env')">
                      <div class="bk-form m20">
                        <table class="biz-simple-table" style="width: 690px;">
                          <thead>
                            <tr>
                              <th style="width: 160px;">{{$t('generic.label.kind')}}</th>
                              <th style="width: 220px;">{{$t('deploy.templateset.variableKey')}}</th>
                              <th style="width: 220px;">{{$t('deploy.templateset.variableValue')}}</th>
                              <th></th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr v-for="(env, index) in curContainer.webCache.env_list" :key="index">
                              <td>
                                <bk-selector
                                  :placeholder="$t('generic.label.kind')"
                                  :setting-key="'id'"
                                  :selected.sync="env.type"
                                  :list="mountTypeList">
                                </bk-selector>
                              </td>
                              <td v-if="['valueFrom', 'custom', 'configmapKey', 'secretKey'].includes(env.type)">
                                <bkbcs-input
                                  type="text"
                                  :placeholder="$t('generic.placeholder.input')"
                                  :value.sync="env.key"
                                  :list="varList"
                                >
                                </bkbcs-input>
                              </td>
                              <td :colspan="['valueFrom', 'custom', 'configmapKey', 'secretKey'].includes(env.type) ? 1 : 2">
                                <template v-if="['valueFrom', 'custom'].includes(env.type)">
                                  <bkbcs-input
                                    type="text"
                                    :placeholder="$t('deploy.templateset.examplePathParam', { path: '/metadata/name' })"
                                    :value.sync="env.value"
                                    :list="varList"
                                  >
                                  </bkbcs-input>
                                </template>
                                <template v-else-if="['configmapKey'].includes(env.type)">
                                  <bk-selector
                                    :placeholder="$t('generic.placeholder.select')"
                                    :setting-key="'id'"
                                    :selected.sync="env.value"
                                    :list="configmapKeyList"
                                    @item-selected="updateEnvItem(...arguments, env)">
                                  </bk-selector>
                                </template>
                                <template v-else-if="['secretKey'].includes(env.type)">
                                  <bk-selector
                                    :placeholder="$t('generic.placeholder.select')"
                                    :setting-key="'id'"
                                    :selected.sync="env.value"
                                    :list="secretKeyList"
                                    @item-selected="updateEnvItem(...arguments, env)">
                                  </bk-selector>
                                </template>
                                <template v-else-if="['configmapFile'].includes(env.type)">
                                  <bk-selector
                                    :placeholder="$t('deploy.templateset.configMapList')"
                                    :setting-key="'name'"
                                    :selected.sync="env.value"
                                    :list="volumeConfigmapList">
                                  </bk-selector>
                                </template>
                                <template v-else-if="['secretFile'].includes(env.type)">
                                  <bk-selector
                                    :placeholder="$t('deploy.templateset.secretList')"
                                    :setting-key="'name'"
                                    :selected.sync="env.value"
                                    :list="volumeSecretList">
                                  </bk-selector>
                                </template>
                              </td>
                              <td>
                                <div class="action-box">
                                  <bk-button class="action-btn ml5" @click.stop.prevent="addEnv()">
                                    <i class="bcs-icon bcs-icon-plus"></i>
                                  </bk-button>
                                  <bk-button class="action-btn" @click.stop.prevent="removeEnv(env, index)" v-show="curContainer.webCache.env_list.length > 1">
                                    <i class="bcs-icon bcs-icon-minus"></i>
                                  </bk-button>
                                </div>
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab4" :title="$t('dashboard.workload.container.limits')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 105px;">{{$t('deploy.templateset.privileged')}}：</label>
                          <div class="bk-form-content" style="margin-left: 105px;">
                            <bk-checkbox v-model="curContainer.securityContext.privileged">{{$t('deploy.templateset.fullAccessHostResources')}}</bk-checkbox>
                          </div>
                        </div>
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 105px;">CPU：</label>
                          <div class="bk-form-content" style="margin-left: 105px;">
                            <div class="bk-form-input-group mr5">
                              <span class="input-group-addon is-left">
                                requests
                              </span>
                              <bkbcs-input
                                type="number"
                                style="width: 100px;"
                                :min="0"
                                :max="curContainer.resources.limits.cpu ? curContainer.resources.limits.cpu : Number.MAX_VALUE"
                                :placeholder="$t('generic.placeholder.input')"
                                :value.sync="curContainer.resources.requests.cpu"
                                :list="varList"
                              >
                              </bkbcs-input>
                              <span class="input-group-addon">
                                m
                              </span>
                            </div>
                            <bcs-popover :content="$t('deploy.templateset.cpuRequests')" placement="top">
                              <span class="bk-badge">
                                <i class="bcs-icon bcs-icon-question-circle"></i>
                              </span>
                            </bcs-popover>

                            <div class="bk-form-input-group ml20 mr5">
                              <span class="input-group-addon is-left">
                                limits
                              </span>
                              <bkbcs-input
                                type="number"
                                style="width: 100px;"
                                :min="0"
                                :placeholder="$t('generic.placeholder.input')"
                                :value.sync="curContainer.resources.limits.cpu"
                                :list="varList">
                              </bkbcs-input>
                              <span class="input-group-addon">
                                m
                              </span>
                            </div>
                            <bcs-popover :content="$t('deploy.templateset.cpuLimits')" placement="top">
                              <span class="bk-badge">
                                <i class="bcs-icon bcs-icon-question-circle"></i>
                              </span>
                            </bcs-popover>
                          </div>
                        </div>
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 105px;">{{$t('generic.label.mem')}}：</label>
                          <div class="bk-form-content" style="margin-left: 105px;">
                            <div class="bk-form-input-group mr5">
                              <span class="input-group-addon is-left">
                                requests
                              </span>
                              <bkbcs-input
                                type="number"
                                style="width: 100px;"
                                :min="0"
                                :max="curContainer.resources.limits.memory ? curContainer.resources.limits.memory : Number.MAX_VALUE"
                                :placeholder="$t('generic.placeholder.input')"
                                :value.sync="curContainer.resources.requests.memory"
                                :list="varList"
                              >
                              </bkbcs-input>
                              <span class="input-group-addon">
                                Mi
                              </span>
                            </div>
                            <bcs-popover :content="$t('deploy.templateset.memoryRequests')" placement="top">
                              <span class="bk-badge">
                                <i class="bcs-icon bcs-icon-question-circle"></i>
                              </span>
                            </bcs-popover>

                            <div class="bk-form-input-group ml20 mr5">
                              <span class="input-group-addon is-left">
                                limits
                              </span>
                              <bkbcs-input
                                type="number"
                                style="width: 100px;"
                                :min="0"
                                :placeholder="$t('generic.placeholder.input')"
                                :value.sync="curContainer.resources.limits.memory"
                                :list="varList">
                              </bkbcs-input>
                              <span class="input-group-addon">
                                Mi
                              </span>
                            </div>
                            <bcs-popover :content="$t('deploy.templateset.memoryLimits')" placement="top">
                              <span class="bk-badge">
                                <i class="bcs-icon bcs-icon-question-circle"></i>
                              </span>
                            </bcs-popover>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab5" :title="$t('deploy.templateset.healthCheck')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 120px;">{{$t('generic.label.kind')}}：</label>
                          <div class="bk-form-content" style="margin-left: 120px">
                            <div class="bk-dropdown-box" style="width: 250px;">
                              <bk-selector
                                :placeholder="$t('generic.placeholder.select')"
                                :setting-key="'id'"
                                :display-key="'name'"
                                :selected.sync="curContainer.webCache.livenessProbeType"
                                :list="healthCheckTypes">
                              </bk-selector>
                            </div>
                          </div>
                        </div>

                        <div class="bk-form-item" v-show="curContainer.webCache.livenessProbeType && curContainer.webCache.livenessProbeType !== 'EXEC'">
                          <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.portName')}}：</label>
                          <div class="bk-form-content" style="margin-left: 120px;">
                            <div class="bk-dropdown-box" style="width: 250px;">
                              <bk-selector
                                :placeholder="$t('generic.placeholder.select')"
                                :setting-key="'name'"
                                :display-key="'name'"
                                :selected="livenessProbePortName"
                                :list="portList"
                                @item-selected="livenessProbePortNameSelect">
                              </bk-selector>
                            </div>
                            <bcs-popover placement="right">
                              <i class="bcs-icon bcs-icon-question-circle ml5" style="vertical-align: middle; cursor: pointer;"></i>
                              <div slot="content">
                                {{$t('deploy.templateset.associatePortMappingSettings')}}
                              </div>
                            </bcs-popover>
                            <p class="biz-guard-tip bk-default mt5" v-if="!portList.length">{{$t('deploy.templateset.completePortMapping')}}</p>
                          </div>
                        </div>

                        <div class="bk-form-item" v-show="curContainer.webCache.livenessProbeType && (curContainer.webCache.livenessProbeType === 'HTTP')">
                          <div class="bk-form-content" style="margin-left: 0">
                            <div class="bk-form-inline-item">
                              <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.requestPath')}}：</label>
                              <div class="bk-form-content" style="margin-left: 120px;">
                                <bkbcs-input
                                  type="text"
                                  style="width: 521px;"
                                  :placeholder="$t('deploy.templateset.examplePathParam', { path: '/healthcheck' })"
                                  :value.sync="curContainer.livenessProbe.httpGet.path"
                                  :list="varList"
                                >
                                </bkbcs-input>
                              </div>
                            </div>
                          </div>
                        </div>

                        <div class="bk-form-item" v-show="curContainer.webCache.livenessProbeType && curContainer.webCache.livenessProbeType === 'EXEC'">
                          <div class="bk-form-content" style="margin-left: 0">
                            <div class="bk-form-inline-item">
                              <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.checkCommand')}}：</label>
                              <div class="bk-form-content" style="margin-left: 120px;">
                                <bkbcs-input
                                  type="text"
                                  style="width: 521px;"
                                  :placeholder="$t('deploy.templateset.exampleCheckCommandWithSpace')"
                                  :value.sync="curContainer.livenessProbe.exec.command"
                                  :list="varList"
                                >
                                </bkbcs-input>
                              </div>
                            </div>
                          </div>
                        </div>

                        <div class="bk-form-item" v-show="curContainer.webCache.livenessProbeType && curContainer.webCache.livenessProbeType === 'HTTP'">
                          <div class="bk-form-content" style="margin-left: 0">
                            <div class="bk-form-inline-item">
                              <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.setHeader')}}：</label>
                              <div class="bk-form-content" style="margin-left: 120px;">
                                <bk-keyer ref="livenessProbeHeaderKeyer" :key-list.sync="livenessProbeHeaders" :var-list="varList" @change="updateLivenessHeader"></bk-keyer>
                              </div>
                            </div>
                          </div>
                        </div>

                        <template>
                          <bk-button :class="['bk-text-button mt10 f12 mb10', { 'rotate': isPartCShow }]" style="margin-left: 114px;" @click.stop.prevent="togglePartC">
                            {{$t('deploy.templateset.advancedSettings')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                          </bk-button>
                          <div v-show="isPartCShow">
                            <div class="bk-form-item">
                              <div class="bk-form-content" style="margin-left: 0">
                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.initialTimeout')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 120px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :min="1"
                                        :value.sync="curContainer.livenessProbe.initialDelaySeconds"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('units.suffix.seconds')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>

                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.checkInterval')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 140px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :min="1"
                                        :value.sync="curContainer.livenessProbe.periodSeconds"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('units.suffix.seconds')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>
                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.checkTimeout')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 140px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :min="1"
                                        :value.sync="curContainer.livenessProbe.timeoutSeconds"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('units.suffix.seconds')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </div>

                            <div class="bk-form-item">
                              <div class="bk-form-content" style="margin-left: 0">
                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.unhealthyThreshold')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 120px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :min="1"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :value.sync="curContainer.livenessProbe.failureThreshold"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('deploy.templateset.failureTimes')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>

                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.healthyThreshold')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 140px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :min="1"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :value.sync="curContainer.livenessProbe.successThreshold"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('deploy.templateset.successTimes')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>

                              </div>
                            </div>
                          </div>
                        </template>

                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab5-1" :title="$t('deploy.templateset.readinessCheck')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <label class="bk-label" style="width: 120px;">{{$t('generic.label.kind')}}：</label>
                          <div class="bk-form-content" style="margin-left: 120px">
                            <div class="bk-dropdown-box" style="width: 250px;">
                              <bk-selector
                                :placeholder="$t('generic.placeholder.select')"
                                :setting-key="'id'"
                                :display-key="'name'"
                                :selected.sync="curContainer.webCache.readinessProbeType"
                                :list="healthCheckTypes">
                              </bk-selector>
                            </div>
                          </div>
                        </div>

                        <div class="bk-form-item" v-show="curContainer.webCache.readinessProbeType && curContainer.webCache.readinessProbeType !== 'EXEC'">
                          <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.portName')}}：</label>
                          <div class="bk-form-content" style="margin-left: 120px;">
                            <div class="bk-dropdown-box" style="width: 250px;">
                              <bk-selector
                                :placeholder="$t('generic.placeholder.select')"
                                :setting-key="'name'"
                                :display-key="'name'"
                                :selected="readinessProbePortName"
                                :list="portList"
                                @item-selected="readinessProbePortNameSelect">
                              </bk-selector>
                            </div>
                            <bcs-popover placement="right">
                              <i class="bcs-icon bcs-icon-question-circle ml5" style="vertical-align: middle; cursor: pointer;"></i>
                              <div slot="content">
                                {{$t('deploy.templateset.associatePortMappingSettings')}}
                              </div>
                            </bcs-popover>
                            <p class="biz-guard-tip bk-default mt5" v-if="!portList.length">{{$t('deploy.templateset.completePortMapping')}}</p>
                          </div>
                        </div>

                        <div class="bk-form-item" v-show="curContainer.webCache.readinessProbeType && (curContainer.webCache.readinessProbeType === 'HTTP')">
                          <div class="bk-form-content" style="margin-left: 0">
                            <div class="bk-form-inline-item">
                              <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.requestPath')}}：</label>
                              <div class="bk-form-content" style="margin-left: 120px;">
                                <bkbcs-input
                                  type="text"
                                  style="width: 521px;"
                                  :placeholder="$t('deploy.templateset.examplePathParam', { path: '/healthcheck' })"
                                  :value.sync="curContainer.readinessProbe.httpGet.path"
                                  :list="varList"
                                >
                                </bkbcs-input>
                              </div>
                            </div>
                          </div>
                        </div>

                        <div class="bk-form-item" v-show="curContainer.webCache.readinessProbeType && curContainer.webCache.readinessProbeType === 'EXEC'">
                          <div class="bk-form-content" style="margin-left: 0">
                            <div class="bk-form-inline-item">
                              <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.checkCommand')}}：</label>
                              <div class="bk-form-content" style="margin-left: 120px;">
                                <bkbcs-input
                                  type="text"
                                  style="width: 521px;"
                                  :placeholder="$t('deploy.templateset.examplePathParam', { path: '/tmp/check.sh' })"
                                  :value.sync="curContainer.readinessProbe.exec.command"
                                  :list="varList"
                                >
                                </bkbcs-input>
                              </div>
                            </div>
                          </div>
                        </div>

                        <div class="bk-form-item" v-show="curContainer.webCache.readinessProbeType && curContainer.webCache.readinessProbeType === 'HTTP'">
                          <div class="bk-form-content" style="margin-left: 0">
                            <div class="bk-form-inline-item">
                              <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.setHeader')}}：</label>
                              <div class="bk-form-content" style="margin-left: 120px;">
                                <bk-keyer ref="readinessProbeHeaderKeyer" :key-list.sync="readinessProbeHeaders" :var-list="varList" @change="updateReadinessHeader"></bk-keyer>
                              </div>
                            </div>
                          </div>
                        </div>

                        <template>
                          <bk-button :class="['bk-text-button mt10 f12 mb10', { 'rotate': isPartCShow }]" style="margin-left: 114px;" @click.stop.prevent="togglePartC">
                            {{$t('deploy.templateset.advancedSettings')}}<i class="bcs-icon bcs-icon-angle-double-down ml5"></i>
                          </bk-button>
                          <div v-show="isPartCShow">
                            <div class="bk-form-item">
                              <div class="bk-form-content" style="margin-left: 0">
                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.initialTimeout')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 120px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :min="1"
                                        :value.sync="curContainer.readinessProbe.initialDelaySeconds"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('units.suffix.seconds')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>

                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.checkInterval')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 140px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :min="1"
                                        :value.sync="curContainer.readinessProbe.periodSeconds"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('units.suffix.seconds')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>
                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.checkTimeout')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 140px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :min="1"
                                        :value.sync="curContainer.readinessProbe.timeoutSeconds"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('units.suffix.seconds')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </div>

                            <div class="bk-form-item">
                              <div class="bk-form-content" style="margin-left: 0">
                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 120px;">{{$t('deploy.templateset.unhealthyThreshold')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 120px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :min="1"
                                        :value.sync="curContainer.readinessProbe.failureThreshold"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('deploy.templateset.failureTimes')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>

                                <div class="bk-form-inline-item">
                                  <label class="bk-label" style="width: 140px;">{{$t('deploy.templateset.healthyThreshold')}}：</label>
                                  <div class="bk-form-content" style="margin-left: 140px;">
                                    <div class="bk-form-input-group">
                                      <bkbcs-input
                                        type="number"
                                        style="width: 70px;"
                                        :placeholder="$t('generic.placeholder.input')"
                                        :min="1"
                                        :value.sync="curContainer.readinessProbe.successThreshold"
                                        :list="varList"
                                      >
                                      </bkbcs-input>
                                      <span class="input-group-addon">
                                        {{$t('deploy.templateset.successTimes')}}
                                      </span>
                                    </div>
                                  </div>
                                </div>

                              </div>
                            </div>
                          </div>
                        </template>

                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab6" :title="$t('deploy.templateset.nonStandardLogCollection')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <div class="bk-form-content" style="margin-left: 20px">
                            <div class="bk-keyer">
                              <div class="biz-keys-list mb10">
                                <div class="biz-key-item" v-for="(logItem, index) in curContainer.webCache.logListCache" :key="index">
                                  <bkbcs-input
                                    type="text"
                                    style="width: 360px;"
                                    :placeholder="$t('deploy.templateset.enterCustomLogPath')"
                                    :value.sync="logItem.value"
                                    :list="varList"
                                  >
                                  </bkbcs-input>

                                  <bk-button class="action-btn ml5" @click.stop.prevent="addLog">
                                    <i class="bcs-icon bcs-icon-plus"></i>
                                  </bk-button>
                                  <bk-button class="action-btn" v-if="curContainer.webCache.logListCache.length > 1" @click.stop.prevent="removeLog(logItem, index)">
                                    <i class="bcs-icon bcs-icon-minus"></i>
                                  </bk-button>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>

                    <bk-tab-panel name="tab7" :title="$t('deploy.templateset.lifecycle')">
                      <div class="bk-form m20">
                        <div class="bk-form-item">
                          <div class="bk-form-content" style="margin-left: 20px">
                            <div class="bk-form-item">
                              <label class="bk-label" style="width: 108px;">{{$t('deploy.templateset.preStop')}}：</label>
                              <div class="bk-form-content" style="margin-left: 108px;">
                                <bkbcs-input
                                  type="text"
                                  :placeholder="$t('generic.placeholder.command')"
                                  :value.sync="curContainer.lifecycle.preStop.exec.command"
                                  :list="varList"
                                >
                                </bkbcs-input>
                              </div>
                            </div>
                            <div class="bk-form-item">
                              <label class="bk-label" style="width: 108px;">{{$t('deploy.templateset.postStart')}}：</label>
                              <div class="bk-form-content" style="margin-left: 108px;">
                                <bkbcs-input
                                  type="text"
                                  :placeholder="$t('generic.placeholder.command')"
                                  :value.sync="curContainer.lifecycle.postStart.exec.command"
                                  :list="varList"
                                >
                                </bkbcs-input>
                                <!-- <bcs-popover content="多个命令用空格分隔" placement="top">
                                                                        <span class="bk-badge ml5">
                                                                            <i class="bcs-icon bcs-icon-question-circle"></i>
                                                                        </span>
                                                                    </bcs-popover> -->
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </bk-tab-panel>
                  </bk-tab>
                </div>

              </div>
              <div class="operation-area mt30 mb50" style="margin-left: 105px;">
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </BcsContent>
</template>

<script>
/* eslint-disable @typescript-eslint/prefer-optional-chain */
/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable no-prototype-builtins */
/* eslint-disable no-multi-assign */
/* eslint-disable no-case-declarations */
/* eslint-disable @typescript-eslint/no-this-alias */
/* eslint-disable @typescript-eslint/no-require-imports */
import yamljs from 'js-yaml';
import _ from 'lodash';

import header from './header.vue';
import tabs from './tabs.vue';

import ace from '@/components/ace-editor';
import bkKeyer from '@/components/keyer';
import BcsContent from '@/components/layout/Content.vue';
import containerParams from '@/json/k8s-container.json';
import applicationParams from '@/json/k8s-statefulset.json';
import k8sBase from '@/mixins/configuration/k8s-base';
import mixinBase from '@/mixins/configuration/mixin-base';

export default {
  name: 'StatefulSet',
  components: {
    ace,
    'bk-keyer': bkKeyer,
    'biz-header': header,
    'biz-tabs': tabs,
    BcsContent,
  },
  mixins: [mixinBase, k8sBase],
  data() {
    return {
      isTabChanging: false,
      renderVersionIndex: 0,
      renderImageIndex: 0,
      curDesc: '',
      curImageData: {},
      winHeight: 0,
      exceptionCode: null,
      isDataLoading: true,
      isDataSaveing: false,
      isPartAShow: false, // 第一个更多设置
      isPartBShow: false, // 第二个更多设置
      isPartCShow: false, // 第三个更多设置
      imageIndex: -1,
      versionIndex: -1,
      isEditName: false,
      isEditDesc: false,
      appParamKeys: [],
      keyList: [],
      yamlContainerWebcache: [],
      curApplicationLinkLabels: [],
      isLoadingImageList: false,
      isLoadingServices: true,
      toJsonDialogConf: {
        isShow: false,
        title: '',
        timer: null,
        width: 800,
        loading: false,
      },
      editorConfig: {
        width: '100%',
        height: '100%',
        lang: 'yaml',
        readOnly: false,
        fullScreen: false,
        value: '',
        editor: null,
      },
      yamlEditorConfig: {
        width: '100%',
        height: '100%',
        lang: 'yaml',
        readOnly: false,
        fullScreen: false,
        value: '',
        editor: null,
      },
      volVisitList: [
        {
          id: 'ReadWriteOnce',
          name: 'ReadWriteOnce',
        },
        {
          id: 'ReadOnlyMany',
          name: 'ReadOnlyMany',
        },
        {
          id: 'ReadWriteMany',
          name: 'ReadWriteMany',
        },
      ],
      netList: [
        {
          id: 'HOST',
          name: 'HOST',
        },
        {
          id: 'BRIDGE',
          name: 'BRIDGE',
        },
        {
          id: 'NONE',
          name: 'NONE',
        },
        {
          id: 'USER',
          name: 'USER',
        },
        {
          id: 'CUSTOM',
          name: this.$t('generic.label.custom'),
        },
      ],
      constraintNameList: [
        {
          id: 'hostname',
          name: 'Hostname',
        },
        {
          id: 'InnerIP',
          name: 'InnerIP',
        },
      ],
      operatorList: [
        {
          id: 'CLUSTER',
          name: 'CLUSTER',
        },
        {
          id: 'GROUPBY',
          name: 'GROUPBY',
        },
        {
          id: 'LIKE',
          name: 'LIKE',
        },
        {
          id: 'UNLIKE',
          name: 'UNLIKE',
        },
        {
          id: 'UNIQUE',
          name: 'UNIQUE',
        },
        {
          id: 'MAXPER',
          name: 'MAXPER',
        },
      ],
      livenessProbeHeaders: [],
      readinessProbeHeaders: [],
      protocolList: [
        {
          id: 'TCP',
          name: 'TCP',
        },
        {
          id: 'UDP',
          name: 'UDP',
        },
      ],
      mountTypeList: [
        {
          id: 'custom',
          name: this.$t('generic.label.custom'),
        },
        {
          id: 'valueFrom',
          name: 'ValueFrom',
        },
        {
          id: 'configmapKey',
          name: this.$t('deploy.templateset.configMapSingleKey'),
        },
        {
          id: 'configmapFile',
          name: this.$t('deploy.templateset.configMapFile'),
        },
        {
          id: 'secretKey',
          name: this.$t('deploy.templateset.secretSingleKey'),
        },
        {
          id: 'secretFile',
          name: this.$t('deploy.templateset.secretFile'),
        },
      ],
      volumeTypeList: [
        {
          id: 'emptyDir',
          name: 'emptyDir',
        },
        {
          id: 'emptyDir(Memory)',
          name: 'emptyDir(Memory)',
        },
        // {
        //     id: 'persistentVolumeClaim',
        //     name: 'persistentVolumeClaim'
        // },
        {
          id: 'hostPath',
          name: 'hostPath',
        },
        {
          id: 'configMap',
          name: 'configMap',
        },
        {
          id: 'secret',
          name: 'secret',
        },
      ],
      metricIndex: [],
      configmapList: [],
      configmapKeyList: [],
      secretKeyList: [],
      secretList: [],
      volumeConfigmapList: [],
      volumeSecretList: [],
      curApplicationId: 0,
      curApplication: applicationParams,
      curContainerIndex: 0,
      curContainer: applicationParams.config.spec.template.spec.allContainers[0],
      isAlwayCheckImage: false,
      editTemplate: {
        name: '',
        desc: '',
      },
      imageList: [],
      imageVersionList: [],
      restartPolicy: ['Always', 'OnFailure', 'Never'],
      healthCheckTypes: [
        {
          id: 'HTTP',
          name: 'HTTP',
        },
        {
          id: 'TCP',
          name: 'TCP',
        },
        {
          id: 'EXEC',
          name: 'EXEC',
        },
      ],
      logList: [
        {
          value: '',
        },
      ],
      isMorePanelShow: false,
      isPodPanelShow: false,
      strategy: 'Cluster',
      netStrategyList: [
        {
          id: 0,
          name: 'Cluster',
        },
        {
          id: 1,
          name: 'Host',
        },
      ],
      curApplicationCache: null,
    };
  },
  computed: {
    volumeConfigmapAllList() {
      const list = [...this.volumeConfigmapList];
      this.existConfigmapList.forEach((item) => {
        list.push({
          id: `${item.name}:${item.cluster_id}:${item.namespace}`,
          name: `${item.name} (${item.cluster_name}-${item.namespace})`,
          cluster_name: item.cluster_name,
          cluster_id: item.cluster_id,
          namespace: item.namespace,
        });
      });
      return list;
    },
    existConfigmapList() {
      return this.$store.state.k8sTemplate.existConfigmapList;
    },
    isSelectorChange() {
      const curSelector = {};
      const list = this.curApplication.config.webCache.labelListCache;
      list.forEach((item) => {
        if (item.isSelector) {
          curSelector[item.key] = item.value;
        }
      });

      if (!this.curApplication.cache) {
        return false;
      }
      if (JSON.stringify(this.curApplication.cache.config.spec.selector.matchLabels) === '{}') {
        return false;
      }

      const selectorCache = this.curApplication.cache.config.spec.selector.matchLabels;
      if (JSON.stringify(selectorCache) !== JSON.stringify(curSelector)) {
        return true;
      }
      return false;
    },
    varList() {
      const list = this.$store.state.variable.varList;
      list.forEach((item) => {
        item._id = item.key;
        item._name = item.key;
      });
      return list;
    },
    dnsStrategyList() {
      const netType = this.curApplication.config.spec.template.spec.hostNetwork;
      let list;

      if (netType === 0) {
        list = [
          {
            id: 'ClusterFirst',
            name: 'ClusterFirst',
          },
          {
            id: 'Default',
            name: 'Default',
          },
          {
            id: 'None',
            name: 'None',
          },
        ];
      } else {
        list = [
          {
            id: 'ClusterFirstWithHostNet',
            name: 'ClusterFirstWithHostNet',
          },
          {
            id: 'Default',
            name: 'Default',
          },
          {
            id: 'None',
            name: 'None',
          },
        ];
      }
      return list;
    },
    serviceList() {
      return this.$store.state.k8sTemplate.linkServices;
    },
    linkServiceName() {
      if (this.curApplication.service_tag) {
        const service = this.serviceList.find(item => item.service_tag === this.curApplication.service_tag);
        return service ? service.service_name : '';
      }
      return '';
    },
    curMountVolumes() {
      const results = this.curApplication.config.webCache.volumes.filter(item => item.name);
      return results;
    },
    metricList() {
      return this.$store.state.k8sTemplate.metricList;
    },
    versionList() {
      const list = this.$store.state.k8sTemplate.versionList;
      return list;
    },
    isTemplateSaving() {
      return this.$store.state.k8sTemplate.isTemplateSaving;
    },
    curTemplate() {
      return this.$store.state.k8sTemplate.curTemplate;
    },
    deployments() {
      return this.$store.state.k8sTemplate.deployments;
    },
    services() {
      return this.$store.state.k8sTemplate.services;
    },
    configmaps() {
      return this.$store.state.k8sTemplate.configmaps;
    },
    secrets() {
      return this.$store.state.k8sTemplate.secrets;
    },
    daemonsets() {
      return this.$store.state.k8sTemplate.daemonsets;
    },
    jobs() {
      return this.$store.state.k8sTemplate.jobs;
    },
    statefulsets() {
      return this.$store.state.k8sTemplate.statefulsets;
    },
    livenessProbePortName() {
      const healthParams = this.curContainer.livenessProbe;
      const type = this.curContainer.webCache.livenessProbeType;
      if (type === 'HTTP') {
        return healthParams.httpGet.port;
      } if (type === 'TCP') {
        return healthParams.tcpSocket.port;
      }
      return '';
    },
    readinessProbePortName() {
      const healthParams = this.curContainer.readinessProbe;
      const type = this.curContainer.webCache.readinessProbeType;
      if (type === 'HTTP') {
        return healthParams.httpGet.port;
      } if (type === 'TCP') {
        return healthParams.tcpSocket.port;
      }
      return '';
    },
    curVersion() {
      return this.$store.state.k8sTemplate.curVersion;
    },
    templateId() {
      return this.$route.params.templateId;
    },
    projectId() {
      return this.$route.params.projectId;
    },
    projectCode() {
      return this.$route.params.projectCode;
    },
    portList() {
      let results = [];
      const { ports } = this.curContainer;

      if (ports && ports.length) {
        results = ports.filter(port => port.name && port.containerPort);
        return results;
      }
      return [];
    },
    curConstraintLabelList() {
      const keyList = [];
      const nodes = this.curApplication.config.spec.template.spec.nodeSelector;
      // 如果有缓存直接使用
      if (this.curApplication.config.webCache && this.curApplication.config.webCache.nodeSelectorList) {
        return this.curApplication.config.webCache.nodeSelectorList;
      }
      for (const [key, value] of Object.entries(nodes)) {
        keyList.push({
          key,
          value,
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
    curLabelList() {
      const keyList = [];
      const { labels } = this.curApplication.config.spec.template.metadata;
      const selector = this.curApplication.config.spec.selector.matchLabels;
      const linkLabels = this.curApplicationLinkLabels;

      // 如果有缓存直接使用
      if (this.curApplication.config.webCache && this.curApplication.config.webCache.labelListCache) {
        this.curApplication.config.webCache.labelListCache.forEach((item) => {
          const params = {
            key: item.key,
            value: item.value,
            isSelector: item.isSelector,
            disabled: item.disabled,
          };
          keyList.push(params);
        });
        for (const params of keyList) {
          const { key } = params;
          const { value } = params;
          for (const label of linkLabels) {
            if (label.key === key && label.value === value) {
              params.disabled = true;
              params.linkMessage = label.linkMessage;
            }
          }
        }
        return keyList;
      }
      for (const [key, value] of Object.entries(labels)) {
        const params = {
          key,
          value,
          isSelector: false,
          disabled: false,
          linkMessage: '',
        };
        keyList.push(params);
      }

      for (const params of keyList) {
        const { key } = params;
        const { value } = params;

        for (const [mKey, mValue] of Object.entries(selector)) {
          if (mKey === key && mValue === value) {
            params.isSelector = true;
          }
        }
        for (const label of linkLabels) {
          if (label.key === key && label.value === value) {
            params.disabled = true;
            params.linkMessage = label.linkMessage;
          }
        }
      }
      if (!keyList.length) {
        keyList.push({
          key: '',
          value: '',
          isSelector: false,
          disabled: false,
        });
      }
      return keyList;
    },
    curLogLabelList() {
      const keyList = [];
      const labels = this.curApplication.config.customLogLabel;
      // 如果有缓存直接使用
      if (this.curApplication.config.webCache && this.curApplication.config.webCache.logLabelListCache) {
        return this.curApplication.config.webCache.logLabelListCache;
      }
      for (const [key, value] of Object.entries(labels)) {
        keyList.push({
          key,
          value,
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
    curImageSecretList() {
      const list = [];
      if (this.curApplication.config.spec.template.spec.imagePullSecrets) {
        const secrets = this.curApplication.config.spec.template.spec.imagePullSecrets;
        secrets.forEach((item) => {
          list.push({
            key: 'name',
            value: item.name,
          });
        });
      }
      if (!list.length) {
        list.push({
          key: 'name',
          value: '',
        });
      }
      return list;
    },
    curRemarkList() {
      const list = [];
      // 如果有缓存直接使用
      if (this.curApplication.config.webCache && this.curApplication.config.webCache.remarkListCache) {
        return this.curApplication.config.webCache.remarkListCache;
      }
      const { annotations } = this.curApplication.config.spec.template.metadata;
      for (const [key, value] of Object.entries(annotations)) {
        list.push({
          key,
          value,
        });
      }
      if (!list.length) {
        list.push({
          key: '',
          value: '',
        });
      }
      return list;
    },
    curEnvList() {
      const list = [];
      const envs = this.curContainer.env;
      envs.forEach((env) => {
        for (const [key, value] of Object.entries(env)) {
          list.push({
            key,
            value,
          });
        }
      });
      return list;
    },
  },
  watch: {
    'curApplication.config.metadata.name'(val) {
      console.log('va', val);
      this.curApplication.config.webCache.labelListCache.forEach((item) => {
        if (item.key === 'k8s-app') {
          item.value = val;
        }
      });
    },
    'services'() {
      if (this.curVersion) {
        this.initServices(this.curVersion);
      }
    },
    'curContainer'() {
      if (this.curContainer.imagePullPolicy === 'Always') {
        this.isAlwayCheckImage = true;
      } else {
        this.isAlwayCheckImage = false;
      }

      if (!this.curContainer.ports.length) {
        this.addPort();
      }
      // else {
      //     this.curContainer.ports.forEach(item => {
      //         const projectId = this.projectId
      //         const version = this.curVersion
      //         const portId = item.id
      //         if (portId) {
      //             this.$store.dispatch('k8sTemplate/checkPortIsLink', { projectId, version, portId }).then(res => {
      //                 item.isLink = ''
      //             }, res => {
      //                 const message = res.message || res.data.data || ''
      //                 const msg = message.split(',')[0]
      //                 if (msg) {
      //                     item.isLink = msg + this.$t('deploy.templateset.cannotModifyProtocol')
      //                 } else {
      //                     item.isLink = ''
      //                 }
      //             })
      //         } else {
      //             item.isLink = ''
      //         }
      //     })
      // }

      // if (!this.curContainer.volumeMounts.length) {
      //     const volumes = this.curContainer.volumeMounts
      //     volumes.push({
      //         'name': '',
      //         'mountPath': '',
      //         'subPath': '',
      //         'readOnly': false
      //     })
      // }
    },
    'curApplication'() {
      this.curContainerIndex = 0;
      const container = this.curApplication.config.spec.template.spec.allContainers[0];
      this.setCurContainer(container, 0);
    },
    'curVersion'(val) {
      this.initVolumeConfigmaps();
      this.initVloumeSelectets();
    },
  },
  // async beforeRouteLeave (to, form, next) {
  //     // 修改模板集信息
  //     await this.$refs.commonHeader.saveTemplate()
  //     next()
  // },
  mounted() {
    this.isDataLoading = true;
    this.$refs.commonHeader.initTemplate((data) => {
      this.initResource(data);
      this.isDataLoading = false;
    });
    this.winHeight = window.innerHeight;
    this.initImageList();
    this.initVolumeConfigmaps();
    this.initVloumeSelectets();
    // this.initMetricList()
  },
  methods: {
    updateEnvItem(index, data, env) {
      env.keyCache = data.keyCache;
      env.nameCache = data.nameCache;
    },
    reloadServices() {
      if (this.curVersion) {
        this.isLoadingServices = true;
        this.initServices(this.curVersion);
      }
    },
    initServices(version) {
      const { projectId } = this;
      this.linkAppVersion = version;
      this.$store.dispatch('k8sTemplate/getServicesByVersion', { projectId, version }).then((res) => {
        this.isLoadingServices = false;
      }, (res) => {
        const { message } = res;
        this.$bkMessage({
          theme: 'error',
          message,
          hasCloseIcon: true,
          delay: '10000',
        });
      });
    },
    removeVolTpl(item, index) {
      const volumes = this.curApplication.config.spec.volumeClaimTemplates;
      volumes.splice(index, 1);
    },
    addVolTpl() {
      const volumes = this.curApplication.config.spec.volumeClaimTemplates;
      volumes.push({
        metadata: {
          name: '',
        },
        spec: {
          accessModes: [],
          storageClassName: '',
          resources: {
            requests: {
              storage: 1,
            },
          },

        },
      });
    },
    updateLivenessHeader(list, data) {
      const result = [];
      list.forEach((item) => {
        const params = {
          name: item.key,
          value: item.value,
        };
        result.push(params);
      });
      this.curContainer.livenessProbe.httpGet.httpHeaders = result;
    },
    updateReadinessHeader(list, data) {
      const result = [];
      list.forEach((item) => {
        const params = {
          name: item.key,
          value: item.value,
        };
        result.push(params);
      });
      this.curContainer.readinessProbe.httpGet.httpHeaders = result;
    },
    changeDNSPolicy(item, data) {
      if (item === 0) {
        this.curApplication.config.spec.template.spec.dnsPolicy = 'ClusterFirst';
      } else {
        this.curApplication.config.spec.template.spec.dnsPolicy = 'ClusterFirstWithHostNet';
      }
    },
    getAppParamsKeys(obj, result) {
      for (const key in obj) {
        if (key === 'nodeSelector') continue;
        if (key === 'annotations') continue;
        if (key === 'volumes') continue;
        if (key === 'affinity') continue;
        if (key === 'labels') continue;
        if (key === 'selector') continue;
        if (Object.prototype.toString.call(obj) === '[object Array]') {
          this.getAppParamsKeys(obj[key], result);
        } else if (Object.prototype.toString.call(obj) === '[object Object]') {
          if (!result.includes(key)) {
            result.push(key);
          }
          this.getAppParamsKeys(obj[key], result);
        }
      }
    },
    checkJson(jsonObj) {
      const { editor } = this.editorConfig;
      const appParams = applicationParams.config;
      const appParamKeys = [
        'id',
        'containerPort',
        'hostPort',
        'name',
        'protocol',
        'isLink',
        'isDisabled',
        'env',
        'secrets',
        'configmaps',
        'logPathList',
        'valueFrom',
        'configMapRef',
        'configMapKeyRef',
        'secretRef',
        'secretKeyRef',
        'serviceAccountName',
        'fieldRef',
        'fieldPath',
        'envFrom',
      ];
      const jsonParamKeys = [];
      this.getAppParamsKeys(appParams, appParamKeys);
      this.getAppParamsKeys(jsonObj, jsonParamKeys);
      // application查看无效字段
      for (const key of jsonParamKeys) {
        if (!appParamKeys.includes(key)) {
          this.$bkMessage({
            theme: 'error',
            message: `${key}${this.$t('deploy.templateset.invalidField')}`,
          });
          const match = editor.find(`${key}`);
          if (match) {
            editor.moveCursorTo(match.end.row, match.end.column);
          }
          return false;
        }
      }

      return true;
    },
    formatJson(jsonObj) {
      // 标签
      const keyList = [];
      const { labels } = jsonObj.spec.template.metadata;
      const selector = jsonObj.spec.selector.matchLabels;
      const linkLabels = this.curApplicationLinkLabels;
      const { hostAliases } = jsonObj.spec.template.spec;

      for (const [key, value] of Object.entries(labels)) {
        const params = {
          key,
          value,
          isSelector: false,
          disabled: false,
        };
        keyList.push(params);
      }
      for (const params of keyList) {
        const { key } = params;
        const { value } = params;

        for (const [mKey, mValue] of Object.entries(selector)) {
          if (mKey === key && mValue === value) {
            params.isSelector = true;
          }
        }
        for (const label of linkLabels) {
          if (label.key === key && label.value === value) {
            params.disabled = true;
          }
        }
      }
      if (!keyList.length) {
        keyList.push({
          key: '',
          value: '',
          isSelector: false,
          disabled: false,
        });
      }
      jsonObj.webCache.labelListCache = keyList;

      // hostAliases
      const hostAliasesCache = [];
      if (hostAliases) {
        for (const hostAlias of hostAliases) {
          hostAliasesCache.push({
            ip: hostAlias.ip,
            hostnames: hostAlias.hostnames.join(';'),
          });
        }
        jsonObj.webCache.hostAliasesCache = hostAliasesCache;
      }

      // 日志标签
      const logLabels = jsonObj.customLogLabel;
      const logLabelList = [];
      for (const [key, value] of Object.entries(logLabels)) {
        logLabelList.push({
          key,
          value,
        });
      }
      if (!logLabelList.length) {
        logLabelList.push({
          key: '',
          value: '',
        });
      }
      jsonObj.webCache.logLabelListCache = logLabelList;

      // 注解
      const remarkList = [];
      const { annotations } = jsonObj.spec.template.metadata;
      for (const [key, value] of Object.entries(annotations)) {
        remarkList.push({
          key,
          value,
        });
      }
      if (!remarkList.length) {
        remarkList.push({
          key: '',
          value: '',
        });
      }

      jsonObj.webCache.remarkListCache = remarkList;

      // 亲和性约束
      const { affinity } = jsonObj.spec.template.spec;
      if (affinity && JSON.stringify(affinity) !== '{}') {
        const yamlStr = yamljs.dump(jsonObj.spec.template.spec.affinity, { indent: 2 });
        jsonObj.webCache.affinityYaml = yamlStr;
        jsonObj.webCache.isUserConstraint = true;
      } else {
        jsonObj.spec.template.spec.affinity = {};
        jsonObj.webCache.affinityYaml = '';
        jsonObj.webCache.isUserConstraint = false;
      }

      // 调度约束
      const { nodeSelector } = jsonObj.spec.template.spec;
      const nodeSelectorList = jsonObj.webCache.nodeSelectorList = [];
      for (const [key, value] of Object.entries(nodeSelector)) {
        nodeSelectorList.push({
          key,
          value,
        });
      }
      if (!nodeSelectorList.length) {
        nodeSelectorList.push({
          key: '',
          value: '',
        });
      }

      // Metric信息 (合并原数据)
      jsonObj.webCache.isMetric = this.curApplicationCache.webCache.isMetric;
      jsonObj.webCache.metricIdList = this.curApplicationCache.webCache.metricIdList;

      // 挂载卷
      const { volumes } = jsonObj.spec.template.spec;
      let volumesCache = jsonObj.webCache.volumes;

      if (volumes && volumes.length) {
        volumesCache = [];
        volumes.forEach((volume) => {
          if (volume.hasOwnProperty('emptyDir')) {
            if (volume.emptyDir.medium) {
              volumesCache.push({
                type: 'emptyDir(Memory)',
                name: volume.name,
                source: volume.emptyDir.sizeLimit.replace('Gi', ''),
              });
            } else {
              volumesCache.push({
                type: 'emptyDir',
                name: volume.name,
                source: '',
              });
            }
          } else if (volume.hasOwnProperty('persistentVolumeClaim')) {
            volumesCache.push({
              type: 'persistentVolumeClaim',
              name: volume.name,
              source: volume.persistentVolumeClaim.claimName,
            });
          } else if (volume.hasOwnProperty('hostPath')) {
            const volumeItem = {
              type: 'hostPath',
              name: volume.name,
              source: volume.hostPath.path,
            };
            if (volume.hostPath.type) {
              volumeItem.hostType = volume.hostPath.type;
            }
            volumesCache.push(volumeItem);
          } else if (volume.hasOwnProperty('configMap')) {
            volumesCache.push({
              type: 'configMap',
              name: volume.name,
              source: volume.configMap.name,
            });
          } else if (volume.hasOwnProperty('secret')) {
            volumesCache.push({
              type: 'secret',
              name: volume.name,
              source: volume.secret.secretName,
            });
          }
        });
      }
      jsonObj.webCache.volumes = JSON.parse(JSON.stringify(volumesCache));
      // container env
      jsonObj.spec.template.spec.allContainers = [];
      const { containers } = jsonObj.spec.template.spec;
      const { initContainers } = jsonObj.spec.template.spec;
      containers.forEach((container, index) => {
        this.formatContainerJosn(container, index);
        container.webCache.containerType = 'container';
        jsonObj.spec.template.spec.allContainers.push(container);
      });
      initContainers.forEach((container, index) => {
        this.formatContainerJosn(container, index);
        container.webCache.containerType = 'initContainer';
        jsonObj.spec.template.spec.allContainers.push(container);
      });
      return jsonObj;
    },
    formatContainerJosn(container, index) {
      if (!container.webCache) {
        container.webCache = {};
      }

      // 兼容原数据webcache
      container.webCache.imageName = container.imageName;
      delete container.imageName;
      container.webCache.env_list = [];

      this.curApplicationCache.spec.template.spec.allContainers.forEach((containerCache) => {
        if (containerCache.name === container.name) {
          // 描述
          container.webCache.desc = containerCache.webCache.desc;
          // 合并非标准日志采集
          container.webCache.logListCache = containerCache.webCache.logListCache;
        }
      });

      // 环境变量
      if ((container.env && container.env.length) || (container.envFrom && container.envFrom.length)) {
        const envs = container.env || [];
        const envFroms = container.envFrom || [];
        envs.forEach((item) => {
          // valuefrom
          if (item.valueFrom && item.valueFrom.fieldRef) {
            container.webCache.env_list.push({
              type: 'valueFrom',
              key: item.name,
              value: item.valueFrom.fieldRef.fieldPath,
            });
            return false;
          }

          // configMap单键
          if (item.valueFrom && item.valueFrom.configMapKeyRef) {
            container.webCache.env_list.push({
              type: 'configmapKey',
              key: item.name,
              nameCache: item.valueFrom.configMapKeyRef.name,
              keyCache: item.valueFrom.configMapKeyRef.key,
              value: `${item.valueFrom.configMapKeyRef.name}.${item.valueFrom.configMapKeyRef.key}`,
            });
            return false;
          }

          // secret单键
          if (item.valueFrom && item.valueFrom.secretKeyRef) {
            container.webCache.env_list.push({
              type: 'secretKey',
              key: item.name,
              nameCache: item.valueFrom.secretKeyRef.name,
              keyCache: item.valueFrom.secretKeyRef.key,
              value: `${item.valueFrom.secretKeyRef.name}.${item.valueFrom.secretKeyRef.key}`,
            });
            return false;
          }

          // 自定义
          container.webCache.env_list.push({
            type: 'custom',
            key: item.name,
            value: item.value,
          });
        });

        envFroms.forEach((item) => {
          // configMap文件
          if (item.configMapRef) {
            container.webCache.env_list.push({
              type: 'configmapFile',
              key: '',
              value: item.configMapRef.name,
            });
            return false;
          }

          // secret文件
          if (item.secretRef) {
            container.webCache.env_list.push({
              type: 'secretFile',
              key: '',
              value: item.secretRef.name,
            });
            return false;
          }
        });
      }

      if (!container.webCache.env_list.length) {
        container.webCache.env_list.push({
          type: 'custom',
          key: '',
          value: '',
        });
      }

      // volumeMounts
      if (container.volumeMounts.length) {
        container.volumeMounts.forEach((volume) => {
          volume.readOnly = false;
        });
      }

      // 镜像自定义
      container.webCache.isImageCustomed = !container.image.startsWith(`${DEVOPS_ARTIFACTORY_HOST}`);

      // volumeMounts
      if (Array.isArray(container.args)) {
        container.args = container.args.join(' ');
      }

      // 端口
      if (container.ports) {
        const { ports } = container;
        ports.forEach((item, index) => {
          item.isLink = false;
          if (!item.id) {
            item.id = +new Date() + index;
          }
        });
      }

      // 资源限制
      const { resources } = container;
      if (resources.limits.cpu && resources.limits.cpu.replace) {
        resources.limits.cpu = Number(resources.limits.cpu.replace('m', ''));
      }
      if (resources.limits.memory && resources.limits.memory.replace) {
        resources.limits.memory = Number(resources.limits.memory.replace('Mi', ''));
      }
      if (resources.requests.cpu && resources.requests.cpu.replace) {
        resources.requests.cpu = Number(resources.requests.cpu.replace('m', ''));
      }
      if (resources.requests.memory && resources.requests.memory.replace) {
        resources.requests.memory = Number(resources.requests.memory.replace('Mi', ''));
      }
    },
    hideApplicationJson() {
      this.toJsonDialogConf.isShow = false;
    },
    saveApplicationJson() {
      const { editor } = this.editorConfig;
      const yaml = editor.getValue();
      const cParams = containerParams;
      let appObj = null;
      if (!yaml) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.enterYAML'),
        });
        return false;
      }

      try {
        appObj = yamljs.load(yaml);
      } catch (err) {
        this.$bkMessage({
          theme: 'error',
          message: this.$t('deploy.templateset.enterValidYAML'),
        });
        return false;
      }

      const annot = editor.getSession().getAnnotations();
      if (annot && annot.length) {
        editor.gotoLine(annot[0].row, annot[0].column, true);
        return false;
      }

      if (appObj.spec.template.spec.containers) {
        const { containers } = appObj.spec.template.spec;
        const containerCopys = [];
        containers.forEach((container) => {
          const copy = _.merge({}, cParams, container);
          containerCopys.push(copy);
        });
        containers.splice(0, containers.length, ...containerCopys);
      }

      if (appObj.spec.template.spec.initContainers) {
        const { initContainers } = appObj.spec.template.spec;
        const containerCopys = [];
        initContainers.forEach((container) => {
          const copy = _.merge({}, cParams, container);
          containerCopys.push(copy);
        });
        initContainers.splice(0, initContainers.length, ...containerCopys);
      }

      const newConfObj = _.merge({}, applicationParams.config, appObj);
      const jsonFromat = this.formatJson(newConfObj);
      this.curApplication.config = jsonFromat;
      this.curApplication.desc = this.curDesc;
      this.toJsonDialogConf.isShow = false;
      if (this.curApplication.config.spec.template.spec.allContainers.length) {
        const container = this.curApplication.config.spec.template.spec.allContainers[0];
        this.setCurContainer(container, 0);
      }
    },
    showJsonPanel() {
      this.toJsonDialogConf.title = `${this.curApplication.config.metadata.name}.yaml`;
      const appConfig = JSON.parse(JSON.stringify(this.curApplication.config));
      const { webCache } = appConfig;

      this.curDesc = this.curApplication.desc;
      // 在处理yaml导入时，保存一份原数据，方便对导入的数据进行合并处理
      this.curApplicationCache = JSON.parse(JSON.stringify(this.curApplication.config));

      // 标签
      if (webCache && webCache.labelListCache) {
        const labelKeyList = this.tranListToObject(webCache.labelListCache);
        appConfig.spec.template.metadata.labels = labelKeyList;
        appConfig.spec.selector.matchLabels = {};
        webCache.labelListCache.forEach((item) => {
          if (item.isSelector && item.key && item.value) {
            appConfig.spec.selector.matchLabels[item.key] = item.value;
          }
        });
      }

      // 日志标签
      if (webCache && webCache.logLabelListCache) {
        const labelKeyList = this.tranListToObject(webCache.logLabelListCache);
        appConfig.customLogLabel = labelKeyList;
      }

      // HostAliases
      if (webCache && webCache.hostAliasesCache) {
        appConfig.spec.template.spec.hostAliases = [];
        webCache.hostAliasesCache.forEach((item) => {
          appConfig.spec.template.spec.hostAliases.push({
            ip: item.ip,
            hostnames: item.hostnames.replace(/ /g, '').split(';'),
          });
        });
      }

      // 注解
      if (webCache && webCache.remarkListCache) {
        const remarkKeyList = this.tranListToObject(webCache.remarkListCache);
        appConfig.spec.template.metadata.annotations = remarkKeyList;
      }

      // 调度约束
      if (webCache.nodeSelectorList) {
        const nodeSelector = appConfig.spec.template.spec.nodeSelector = {};
        const { nodeSelectorList } = webCache;
        nodeSelectorList.forEach((item) => {
          nodeSelector[item.key] = item.value;
        });
      }

      // 亲和性约束
      if (webCache.isUserConstraint) {
        try {
          const yamlCode = webCache.affinityYamlCache || webCache.affinityYaml;
          webCache.affinityYaml = yamlCode;
          const json = yamljs.load(yamlCode);
          if (json) {
            appConfig.spec.template.spec.affinity = json;
          } else {
            appConfig.spec.template.spec.affinity = {};
          }
        } catch (err) {
          // error
        }
      } else {
        appConfig.spec.template.spec.affinity = {};
      }

      if (webCache && webCache.volumes) {
        const cacheColumes = webCache.volumes;
        const volumes = [];
        cacheColumes.forEach((volume) => {
          if ((volume.name && volume.source) || (volume.name && volume.type === 'emptyDir')) {
            switch (volume.type) {
              case 'emptyDir':
                volumes.push({
                  name: volume.name,
                  emptyDir: {},
                });
                break;

              case 'persistentVolumeClaim':
                volumes.push({
                  name: volume.name,
                  persistentVolumeClaim: {
                    claimName: volume.source,
                  },
                });
                break;

              case 'hostPath':
                const item = {
                  name: volume.name,
                  hostPath: {
                    path: volume.source,
                  },
                };
                if (volume.hostType) {
                  item.hostPath.type = volume.hostType;
                }
                volumes.push(item);
                break;

              case 'configMap':
                // 针对已经存的configmap处理
                let volumeSource = volume.source;
                if (volume.is_exist) {
                  volumeSource = volume.source.split(':')[0];
                }
                volumes.push({
                  name: volume.name,
                  configMap: {
                    name: volumeSource,
                  },
                });
                break;

              case 'secret':
                volumes.push({
                  name: volume.name,
                  secret: {
                    secretName: volume.source,
                  },
                });
                break;

              case 'emptyDir(Memory)':
                volumes.push({
                  name: volume.name,
                  emptyDir: {
                    medium: 'Memory',
                    sizeLimit: `${volume.source}Gi`,
                  },
                });
                break;
            }
          }
        });

        appConfig.spec.template.spec.volumes = volumes;
      }
      delete appConfig.webCache;

      // container
      appConfig.spec.template.spec.containers = [];
      appConfig.spec.template.spec.initContainers = [];

      appConfig.spec.template.spec.allContainers.forEach((container) => {
        container.imageName = container.webCache.imageName;
        this.yamlContainerWebcache.push(JSON.parse(JSON.stringify(container.webCache)));

        container.env = [];
        container.envFrom = [];
        // 环境变量
        if (container.webCache && container.webCache.env_list) {
          const envs = container.webCache.env_list;
          envs.forEach((env) => {
            // valuefrom
            if (env.type === 'valueFrom') {
              container.env.push({
                name: env.key,
                valueFrom: {
                  fieldRef: {
                    fieldPath: env.value,
                  },
                },
              });
              return false;
            }

            // configMap单键
            if (env.type === 'configmapKey') {
              container.env.push({
                name: env.key,
                valueFrom: {
                  configMapKeyRef: {
                    name: env.nameCache,
                    key: env.keyCache,
                  },
                },
              });
              return false;
            }

            // configMap文件
            if (env.type === 'configmapFile') {
              container.envFrom.push({
                configMapRef: {
                  name: env.value,
                },
              });
              return false;
            }

            // secret单键
            if (env.type === 'secretKey') {
              container.env.push({
                name: env.key,
                valueFrom: {
                  secretKeyRef: {
                    name: env.nameCache,
                    key: env.keyCache,
                  },
                },
              });
              return false;
            }

            // secret文件
            if (env.type === 'secretFile') {
              container.envFrom.push({
                secretRef: {
                  name: env.value,
                },
              });
              return false;
            }

            // 自定义
            if (env.key) {
              container.env.push({
                name: env.key,
                value: env.value,
              });
            }
          });
        }

        if (!container.webCache.env_list.length) {
          container.webCache.env_list.push({
            type: 'custom',
            key: '',
            value: '',
          });
        }
        if (container.webCache.containerType === 'initContainer') {
          appConfig.spec.template.spec.initContainers.push(container);
        } else {
          appConfig.spec.template.spec.containers.push(container);
        }
        delete container.webCache;
      });
      delete appConfig.spec.template.spec.allContainers;

      const yamlStr = yamljs.dump(appConfig, { indent: 2 });
      this.editorConfig.value = yamlStr;
      this.toJsonDialogConf.isShow = true;
    },
    editorInitAfter(editor) {
      this.editorConfig.editor = editor;
      this.editorConfig.editor.setStyle('biz-app-container-tojson-ace');
    },
    yamlEditorInitAfter(editor) {
      this.yamlEditorConfig.editor = editor;
      if (this.curApplication.config.webCache.affinityYaml) {
        editor.setValue(this.curApplication.config.webCache.affinityYaml);
      }
    },
    yamlEditorInput(val) {
      this.curApplication.config.webCache.affinityYamlCache = val;
    },
    yamlEditorBlur(val) {
      this.curApplication.config.webCache.affinityYaml = val;
    },
    setFullScreen() {
      this.editorConfig.fullScreen = !this.editorConfig.fullScreen;
    },
    cancelFullScreen() {
      this.editorConfig.fullScreen = false;
    },
    closeToJson() {
      this.toJsonDialogConf.isShow = false;
      this.toJsonDialogConf.title = '';
      this.editorConfig.value = '';
      this.copyContent = '';
    },
    initResource(data) {
      const version = data.latest_version_id || data.version;
      if (data.statefulsets && data.statefulsets.length) {
        this.setCurApplication(data.statefulsets[0], 0);
      } else if (data.statefulset && data.statefulset.length) {
        this.setCurApplication(data.statefulset[0], 0);
      }
      if (version) {
        this.initServices(version);
      } else {
        this.isLoadingServices = false;
      }
    },
    exportToYaml(data) {
      this.$router.push({
        name: 'K8sYamlTemplateset',
        params: {
          projectId: this.projectId,
          projectCode: this.projectCode,
          templateId: 0,
        },
        query: {
          action: 'export',
        },
      });
    },
    async tabResource(type, target) {
      this.isTabChanging = true;
      await this.$refs.commonHeader.saveTemplate();
      await this.$refs.commonHeader.autoSaveResource(type);
      this.$refs.commonTab.goResource(target);
    },
    exceptionHandler(exceptionCode) {
      this.isDataLoading = false;
      this.exceptionCode = exceptionCode;
    },
    livenessProbePortNameSelect(selected, data) {
      const healthParams = this.curContainer.livenessProbe;
      const type = this.curContainer.webCache.livenessProbeType;
      if (type === 'HTTP') {
        healthParams.httpGet.port = selected;
      } else if (type === 'TCP') {
        healthParams.tcpSocket.port = selected;
      }
    },
    readinessProbePortNameSelect(selected, data) {
      const healthParams = this.curContainer.readinessProbe;
      const type = this.curContainer.webCache.readinessProbeType;
      if (type === 'HTTP') {
        healthParams.httpGet.port = selected;
      } else if (type === 'TCP') {
        healthParams.tcpSocket.port = selected;
      }
    },
    toggleRouter(target) {
      this.$router.push({
        name: target,
        params: {
          projectId: this.projectId,
          templateId: this.templateId,
        },
      });
    },
    changeImagePullPolicy() {
      // 判断改变前的状态
      if (!this.isAlwayCheckImage) {
        this.curContainer.imagePullPolicy = 'Always';
      } else {
        this.curContainer.imagePullPolicy = 'IfNotPresent';
      }
    },
    addLocalApplication() {
      const application = JSON.parse(JSON.stringify(applicationParams));
      const index = this.statefulsets.length;
      const now = +new Date();
      const applicationName = `statefulset-${index + 1}`;
      const containerName = 'container-1';

      application.id = `local_${now}`;
      application.isEdited = true;

      application.config.metadata.name = applicationName;
      application.config.spec.template.spec.allContainers[0].name = containerName;
      this.statefulsets.push(application);
      this.setCurApplication(application, index);

      // 标签添加默认选择器
      const defaultLabels = [{
        disabled: true,
        isSelector: true,
        key: 'k8s-app',
        value: applicationName,
      }];
      const defaultLabelObject = {
        APP: applicationName,
      };
      this.updateApplicationLabel(defaultLabels, defaultLabelObject);
    },
    setCurApplication(application, index) {
      // eslint-disable-next-line no-plusplus
      this.renderImageIndex++;
      this.curApplication = application;
      this.curApplicationId = application.id;
      this.initLinkLabels();

      clearInterval(this.compareTimer);
      clearTimeout(this.setTimer);
      this.setTimer = setTimeout(() => {
        if (!this.curApplication.cache) {
          this.curApplication.cache = JSON.parse(JSON.stringify(application));
        }
        this.watchChange();
      }, 500);
    },
    watchChange() {
      this.compareTimer = setInterval(() => {
        const appCopy = JSON.parse(JSON.stringify(this.curApplication));
        const cacheCopy = JSON.parse(JSON.stringify(this.curApplication.cache));
        // 删除无用属性
        delete appCopy.isEdited;
        delete appCopy.cache;
        delete appCopy.id;
        delete appCopy.config.spec.template.spec.containers;
        delete appCopy.config.spec.template.spec.initContainers;
        appCopy.config.spec.template.spec.allContainers.forEach((item) => {
          if (item.ports.length === 1) {
            const port = item.ports[0];
            if (port.containerPort === '' && port.name === '') {
              item.ports = [];
            }
          }

          item.ports.forEach((port) => {
            delete port.isLink;
          });

          if (item.volumeMounts.length === 1) {
            const volumn = item.volumeMounts[0];
            if (volumn.name === '' && volumn.mountPath === '' && volumn.readOnly === false) {
              item.volumeMounts = [];
            }
          }
        });

        delete cacheCopy.isEdited;
        delete cacheCopy.cache;
        delete cacheCopy.id;
        delete cacheCopy.config.spec.template.spec.containers;
        delete cacheCopy.config.spec.template.spec.initContainers;
        cacheCopy.config.spec.template.spec.allContainers.forEach((item) => {
          if (item.ports.length === 1) {
            const port = item.ports[0];
            if (port.containerPort === '' && port.name === '') {
              item.ports = [];
            }
          }

          item.ports.forEach((port) => {
            delete port.isLink;
          });

          if (item.volumeMounts.length === 1) {
            const volumn = item.volumeMounts[0];
            if (volumn.name === '' && volumn.mountPath === '' && volumn.readOnly === false) {
              item.volumeMounts = [];
            }
          }
        });

        const appStr = JSON.stringify(appCopy);
        const cacheStr = JSON.stringify(cacheCopy);

        if (String(this.curApplication.id).indexOf('local_') > -1) {
          this.curApplication.isEdited = true;
        } else if (appStr !== cacheStr) {
          this.curApplication.isEdited = true;
        } else {
          this.curApplication.isEdited = false;
        }
      }, 1000);
    },
    getProbeHeaderList(headers) {
      const list = [];
      if (headers.forEach) {
        headers.forEach((item) => {
          list.push({
            key: item.name || '',
            value: item.value || '',
          });
        });
      }
      if (!list.length) {
        list.push({
          key: '',
          value: '',
        });
      }
      return list;
    },
    /**
             * 把上一个容器的参数重置
             */
    resetPreContainerParams() {
      this.imageVersionList = [];
    },
    /**
             * 切换container
             * @param {object} container container
             */
    setCurContainer(container, index) {
      // 利用setTimeout事件来先让当前容器的blur事件执行完才切换
      setTimeout(() => {
        // 切换container
        // this.resetPreContainerParams()
        container.ports.forEach((port) => {
          if (!port.protocol) {
            port.protocol = 'TCP';
          }
        });
        // eslint-disable-next-line no-plusplus
        this.renderImageIndex++;
        this.curContainer = container;
        this.curContainerIndex = index;

        this.livenessProbeHeaders = this.getProbeHeaderList(this.curContainer.livenessProbe.httpGet.httpHeaders);
        this.readinessProbeHeaders = this.getProbeHeaderList(this.curContainer.readinessProbe.httpGet.httpHeaders);

        const volumesNames = this.curApplication.config.webCache.volumes.map(item => item.name);
        const tmp = this.curContainer.volumeMounts.filter(item => volumesNames.includes(item.name));
        this.curContainer.volumeMounts = tmp;
      }, 300);
    },
    removeContainer(index) {
      const containers = this.curApplication.config.spec.template.spec.allContainers;
      containers.splice(index, 1);
      if (this.curContainerIndex === index) {
        this.curContainerIndex = 0;
      } else if (this.curContainerIndex > index) {
        this.curContainerIndex = this.curContainerIndex - 1;
      }
      this.curContainer = containers[this.curContainerIndex];
    },
    addLocalContainer() {
      // let container = Object.assign({}, containerParams)
      const container = JSON.parse(JSON.stringify(containerParams));
      const containers = this.curApplication.config.spec.template.spec.allContainers;
      const index = containers.length;
      container.name = `container-${index + 1}`;
      containers.push(container);
      this.setCurContainer(container, index);
      this.$refs.containerTooltip.visible = false;
    },
    removeLocalApplication(application, index) {
      // 是否删除当前项
      if (this.curApplication.id === application.id) {
        if (index === 0 && this.statefulsets[index + 1]) {
          this.setCurApplication(this.statefulsets[index + 1]);
        } else if (this.statefulsets[0]) {
          this.setCurApplication(this.statefulsets[0]);
        }
      }
      this.statefulsets.splice(index, 1);
    },
    removeApplication(application, index) {
      const self = this;
      const { projectId } = this;
      const version = this.curVersion;
      const { id } = application;
      this.$bkInfo({
        title: this.$t('generic.title.confirmDelete'),
        content: this.$createElement('p', { style: { 'text-align': 'left' } }, `${this.$t('deploy.templateset.deleteStatefulSet')}：${application.config.metadata.name || this.$t('deploy.templateset.unnamed')}`),
        confirmFn() {
          if (id.indexOf && id.indexOf('local_') > -1) {
            self.removeLocalApplication(application, index);
          } else {
            self.$store.dispatch('k8sTemplate/removeStatefulset', { id, version, projectId }).then((res) => {
              const { data } = res;
              self.removeLocalApplication(application, index);

              if (data.version) {
                self.$store.commit('k8sTemplate/updateCurVersion', data.version);
                self.$store.commit('k8sTemplate/updateBindVersion', true);
              }
            }, (res) => {
              const { message } = res;
              self.$bkMessage({
                theme: 'error',
                message,
              });
            });
          }
        },
      });
    },
    togglePartA() {
      this.isPartAShow = !this.isPartAShow;
    },
    togglePartB() {
      this.isPartBShow = !this.isPartBShow;
    },
    togglePartC() {
      this.isPartCShow = !this.isPartCShow;
    },
    toggleMore() {
      this.isMorePanelShow = !this.isMorePanelShow;
      this.isPodPanelShow = false;
    },
    togglePod() {
      this.isPodPanelShow = !this.isPodPanelShow;
      this.isMorePanelShow = false;
    },
    saveStatefulsetSuccess(params) {
      this.statefulsets.forEach((item) => {
        if (params.responseData.id === item.id || params.preId === item.id) {
          item.cache = JSON.parse(JSON.stringify(item));
        }
      });
      if (params.responseData.id === this.curApplication.id || params.preId === this.curApplication.config.id) {
        this.updateLocalData(params.resource);
      }
    },
    updateLocalData(data) {
      if (data.id) {
        this.curApplication.config.id = data.id;
        this.curApplicationId = data.id;
      }
      if (data.version) {
        this.$store.commit('k8sTemplate/updateCurVersion', data.version);
      }

      this.$store.commit('k8sTemplate/updateStatefulsets', this.statefulsets);
      setTimeout(() => {
        this.statefulsets.forEach((item) => {
          if (item.id === data.id) {
            this.setCurApplication(item);
          }
        });
      }, 500);
    },
    createFirstApplication(data) {
      const { templateId } = this;
      const { projectId } = this;
      this.$store.dispatch('k8sTemplate/addFirstApplication', { projectId, templateId, data }).then((res) => {
        const { data } = res;
        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.save1'),
        });
        this.updateLocalData(data);
        this.isDataSaveing = false;
        if (templateId === 0 || templateId === '0') {
          this.$router.push({
            name: 'mesosTemplatesetApplication',
            params: {
              projectId: this.projectId,
              templateId: data.template_id,
            },
          });
        }
      }, (res) => {
        const { message } = res;
        this.$bkMessage({
          theme: 'error',
          message,
          hasCloseIcon: true,
          delay: '10000',
        });
        this.isDataSaveing = false;
      });
    },
    updateApplication(data) {
      const version = this.curVersion;
      const { projectId } = this;
      const applicationId = this.curApplicationId;
      this.$store.dispatch('k8sTemplate/updateApplication', { projectId, version, data, applicationId }).then((res) => {
        const { data } = res;
        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.save1'),
        });
        this.updateLocalData(data);
        this.isDataSaveing = false;
      }, (res) => {
        const { message } = res;
        this.$bkMessage({
          theme: 'error',
          message,
        });
        this.isDataSaveing = false;
      });
    },
    createApplication(data) {
      const version = this.curVersion;
      const { projectId } = this;
      this.$store.dispatch('k8sTemplate/addApplication', { projectId, version, data }).then((res) => {
        const { data } = res;
        this.$bkMessage({
          theme: 'success',
          message: this.$t('generic.msg.success.save1'),
        });

        this.updateLocalData(data);
        this.isDataSaveing = false;
      }, (res) => {
        const { message } = res;
        this.$bkMessage({
          theme: 'error',
          message,
        });
        this.isDataSaveing = false;
      });
    },
    removeVolumn(item, index) {
      const { allContainers } = this.curApplication.config.spec.template.spec;
      const { volumes } = this.curApplication.config.webCache;

      let matchItem;
      for (const container of allContainers) {
        matchItem = container.volumeMounts.find(volumeMount => volumeMount.name && (volumeMount.name === item.name));
        if (matchItem) {
          this.$bkMessage({
            theme: 'error',
            message: this.$t('deploy.templateset.deleteMountVolumeWarning', { name: container.name || this.$('容器') }),
          });
          return false;
        }
      }

      if (!matchItem) {
        volumes.splice(index, 1);
      }
    },
    addVolumn() {
      const { volumes } = this.curApplication.config.webCache;
      volumes.push({
        type: 'emptyDir',
        name: '',
        source: '',
      });
    },
    removeMountVolumn(item, index) {
      const volumes = this.curContainer.volumeMounts;
      volumes.splice(index, 1);
    },
    addMountVolumn() {
      const volumes = this.curContainer.volumeMounts;
      volumes.push({
        name: '',
        mountPath: '',
        subPath: '',
        readOnly: false,
      });
    },
    removeEnv(item, index) {
      const envList = this.curContainer.webCache.env_list;
      envList.splice(index, 1);
    },
    addEnv() {
      const envList = this.curContainer.webCache.env_list;
      envList.push({
        type: 'custom',
        key: '',
        value: '',
      });
    },
    pasteKey(item, event) {
      const cache = item.key;
      this.paste(event);
      item.key = cache;
      setTimeout(() => {
        item.key = cache;
      }, 0);
    },
    paste(event) {
      const clipboard = event.clipboardData;
      const text = clipboard.getData('Text');
      const envList = this.curContainer.webCache.env_list;
      if (text) {
        const items = text.split('\n');
        items.forEach((item) => {
          const arr = item.split('=');
          envList.push({
            type: 'custom',
            key: arr[0],
            value: arr[1],
          });
        });
      }
      setTimeout(() => {
        this.formatEnvListData();
      }, 10);

      return false;
    },
    formatEnvListData() {
      // 去掉空值
      if (this.curContainer.webCache.env_list.length) {
        const results = [];
        const keyObj = {};
        const { length } = this.curContainer.webCache.env_list;
        this.curContainer.webCache.env_list.forEach((item, i) => {
          if (item.key || item.value) {
            if (!keyObj[item.key]) {
              results.push(item);
              keyObj[item.key] = true;
            }
          }
        });
        const patchLength = results.length - length;
        if (patchLength > 0) {
          for (let i = 0; i < patchLength; i++) {
            results.push({
              type: 'custom',
              key: '',
              value: '',
            });
          }
        }
        this.curContainer.webCache.env_list.splice(0, this.curContainer.webCache.env_list.length, ...results);
      }
    },
    getVolumeNameList(type) {
      if (type === 'configmap') {
        return this.configmapList;
      } if (type === 'secret') {
        return this.secretList;
      }
    },
    getVolumeSourceList(type, name) {
      if (!name) return [];
      if (type === 'configmap') {
        const list = this.configmapList;
        for (const item of list) {
          if (item.name === name) {
            return item.childList;
          }
        }
        return [];
      } if (type === 'secret') {
        const list = this.secretList;
        for (const item of list) {
          if (item.name === name) {
            return item.childList;
          }
        }
        return [];
      }
      return [];
    },
    selectOperate(data) {
      const { operate } = data;
      if (operate === 'UNIQUE') {
        data.type = 0;
        data.arg_value = '';
      }
    },
    updateNodeSelectorList(list, data) {
      if (!this.curApplication.config.webCache) {
        this.curApplication.config.webCache = {};
      }
      this.curApplication.config.webCache.nodeSelectorList = list;
    },
    updateApplicationRemark(list, data) {
      if (!this.curApplication.config.webCache) {
        this.curApplication.config.webCache = {};
      }
      this.curApplication.config.webCache.remarkListCache = list;
    },
    updateApplicationImageSecrets(list, data) {
      const secrets = [];
      list.forEach((item) => {
        secrets.push({
          name: item.value,
        });
      });
      this.curApplication.config.spec.template.spec.imagePullSecrets = secrets;
    },
    updateApplicationLabel(list, data) {
      if (!this.curApplication.config.webCache) {
        this.curApplication.config.webCache = {};
      }
      this.curApplication.config.webCache.labelListCache = list;
    },
    updateApplicationLogLabel(list, data) {
      if (!this.curApplication.config.webCache) {
        this.curApplication.config.webCache = {};
      }
      this.curApplication.config.customLogLabel = data;
      this.curApplication.config.webCache.logLabelListCache = list;
    },
    formatData() {
      const params = JSON.parse(JSON.stringify(this.curApplication));
      params.template = {
        name: this.curTemplate.name,
        desc: this.curTemplate.desc,
      };
      delete params.isEdited;
      // 键值转换
      const remarkKeyList = this.$refs.remarkKeyer.getKeyObject();
      const labelKeyList = this.$refs.labelKeyer.getKeyObject();

      params.metadata.labels = labelKeyList;
      params.metadata.annotations = remarkKeyList;

      // 转换调度约束
      const constraint = params.constraint.intersectionItem;
      constraint.forEach((item) => {
        const data = item.unionData[0];
        const { operate } = data;
        switch (operate) {
          case 'UNIQUE':
            delete data.type;
            delete data.set;
            delete data.text;
            break;
          case 'MAXPER':
            data.type = 3;
            data.text = {
              value: data.arg_value,
            };
            delete data.set;
            break;
          case 'CLUSTER':
            data.type = 4;
            if (data.arg_value.trim().length) {
              data.set = {
                item: data.arg_value.split('|'),
              };
            } else {
              data.set = {
                item: [],
              };
            }

            delete data.text;
            break;
          case 'GROUPBY':
            data.type = 4;
            if (data.arg_value.trim().length) {
              data.set = {
                item: data.arg_value.split('|'),
              };
            } else {
              data.set = {
                item: [],
              };
            }
            delete data.text;
            break;
          case 'LIKE':
            if (data.arg_value.indexOf('|') > -1) {
              data.type = 4;
              if (data.arg_value.trim().length) {
                data.set = {
                  item: data.arg_value.split('|'),
                };
              } else {
                data.set = {
                  item: [],
                };
              }
              delete data.text;
            } else {
              data.type = 3;
              data.text = {
                value: data.arg_value,
              };
              delete data.set;
            }
            break;
          case 'UNLIKE':
            if (data.arg_value.indexOf('|') > -1) {
              data.type = 4;
              if (data.arg_value.trim().length) {
                data.set = {
                  item: data.arg_value.split('|'),
                };
              } else {
                data.set = {
                  item: [],
                };
              }
              delete data.text;
            } else {
              data.type = 3;
              data.text = {
                value: data.arg_value,
              };
              delete data.set;
            }
            break;
        }
      });

      // 转换命令参数和环境变量
      const containers = params.spec.template.spec.allContainers;
      containers.forEach((container) => {
        if (container.args_text.trim().length) {
          container.args = container.args_text.split(' ');
        } else {
          container.args = [];
        }

        container.resources.limits.cpu = parseFloat(container.resources.limits.cpu);

        // docker参数
        const parameterList = container.parameter_list;
        container.parameters = [];
        parameterList.forEach((param) => {
          if (param.key && param.value) {
            container.parameters.push(param);
          }
        });

        // 端口
        const { ports } = container;
        const validatePorts = [];
        ports.forEach((item) => {
          if (item.containerPort && (item.hostPort !== undefined) && item.name && item.protocol) {
            validatePorts.push({
              id: item.id,
              containerPort: item.containerPort,
              hostPort: item.hostPort,
              protocol: item.protocol,
              name: item.name,
            });
          }
        });
        container.ports = validatePorts;

        // volumes
        const { volumes } = container;
        let validateVolumes = [];
        validateVolumes = volumes.filter(item => item.volume.hostPath && item.volume.mountPath && item.name);
        container.volumes = validateVolumes;

        // logpath
        const paths = [];
        const logList = container.logListCache;
        logList.forEach((item) => {
          if (item.value) {
            paths.push(item.value);
          }
        });
        container.logPathList = paths;
      });
      return params;
    },
    saveApplication() {
      if (!this.checkData()) {
        return false;
      }
      if (this.isDataSaveing) {
        return false;
      }
      this.isDataSaveing = true;

      const data = this.formatData();
      if (this.curVersion) {
        if (this.curApplicationId.indexOf && this.curApplicationId.indexOf('local') > -1) {
          this.createApplication(data);
        } else {
          this.updateApplication(data);
        }
      } else {
        this.createFirstApplication(data);
      }
    },
    initImageList() {
      if (this.isLoadingImageList) return false;
      this.isLoadingImageList = true;
      const { projectId } = this;
      this.$store.dispatch('k8sTemplate/getImageList', { projectId }).then((res) => {
        const { data } = res;
        setTimeout(() => {
          data.forEach((item) => {
            item._id = item.value;
            item._name = item.name;
          });
          this.imageList.splice(0, this.imageList.length, ...data);
          this.$store.commit('k8sTemplate/updateImageList', this.imageList);
          this.isLoadingImageList = false;
        }, 1000);
      }, (res) => {
        const { message } = res;
        this.$bkMessage({
          theme: 'error',
          message,
          delay: 10000,
        });
        this.isLoadingImageList = false;
      });
    },
    handleImageCustom() {
      setTimeout(() => {
        const { imageName } = this.curContainer.webCache;
        const { imageVersion } = this.curContainer;
        if (imageName && imageVersion) {
          this.curContainer.image = `${imageName}:${imageVersion}`;
        } else {
          this.curContainer.image = '';
        }
      }, 100);
    },

    handleVersionCustom() {
      this.$nextTick(() => {
        const versionName = this.curContainer.imageVersion;
        const matcher = this.imageVersionList.find(version => version._name === versionName);
        if (matcher) {
          this.setImageVersion(matcher._id, matcher);
        } else {
          const { imageName } = this.curContainer.webCache;
          const version = this.curContainer.imageVersion;

          // curImageData有值，表示是通过选择
          if (JSON.stringify(this.curImageData) !== '{}') {
            if (this.curImageData.is_pub !== undefined) {
              this.curContainer.image = `${DEVOPS_ARTIFACTORY_HOST}/${imageName}:${version}`;
              console.log('镜像是下拉，版本是自定义', this.curContainer.image);
            } else {
              this.curContainer.image = `${DEVOPS_ARTIFACTORY_HOST}/paas/${this.projectCode}/${imageName}:${version}`;
              console.log('镜像是变量，版本是自定义', this.curContainer.image);
            }
          } else {
            this.curContainer.image = `${imageName}:${version}`;
            console.log('镜像和版本都是自定义', this.curContainer.image);
          }
        }
      });
    },

    handleChangeImageMode() {
      this.curContainer.webCache.isImageCustomed = !this.curContainer.webCache.isImageCustomed;
      // 清空原来值
      this.curContainer.webCache.imageName = '';
      this.curContainer.image = '';
      this.curContainer.imageName = '';
      this.curContainer.imageVersion = '';
    },

    changeImage(value, data, isInitTrigger) {
      const { projectId } = this;
      const imageId = data.value;
      const isPub = data.is_pub;
      this.curImageData = data;
      // 如果不是输入变量
      if (isPub !== undefined) {
        this.$store.dispatch('k8sTemplate/getImageVertionList', { projectId, imageId, isPub }).then((res) => {
          const { data } = res;
          data.forEach((item) => {
            item._id = item.text;
            item._name = item.text;
          });

          this.imageVersionList.splice(0, this.imageVersionList.length, ...data);

          // 非首次关联触发，默认选择第一项或清空
          if (isInitTrigger) return;

          if (this.imageVersionList.length) {
            const imageInfo = this.imageVersionList[0];

            this.curContainer.image = imageInfo.value;
            this.curContainer.imageVersion = imageInfo.text;
          } else {
            this.curContainer.image = '';
            this.curContainer.imageVersion = '';
          }
        }, (res) => {
          this.curContainer.image = '';
          this.curContainer.imageVersion = '';
          const { message } = res;
          this.$bkMessage({
            theme: 'error',
            message,
          });
        });
      } else if (!isInitTrigger) {
        this.imageVersionList = [];
        this.curContainer.image = '';
        this.curContainer.imageVersion = '';
      }
    },

    setImageVersion(value, data) {
      // 镜像和版本都是通过下拉选择
      const { projectCode } = this;
      // curImageData不是空对象
      if (JSON.stringify(this.curImageData) !== '{}') {
        if (data.text && data.value) {
          this.curContainer.imageVersion = data.text;
          this.curContainer.image = data.value;
        } else if (this.curImageData.is_pub !== undefined) {
          // 镜像是下拉，版本是变量
          // image = imageBase + imageName + ':' + imageVersion
          const { imageName } = this.curContainer.webCache;
          this.curContainer.imageVersion = value;
          this.curContainer.image = `${DEVOPS_ARTIFACTORY_HOST}/${imageName}:${value}`;
        } else {
          // 镜像和版本是变量
          // image = imageBase +  'paas/' + projectCode + '/' + imageName + ':' + imageVersion
          const { imageName } = this.curContainer.webCache;
          this.curContainer.imageVersion = value;
          this.curContainer.image = `${DEVOPS_ARTIFACTORY_HOST}/paas/${projectCode}/${imageName}:${value}`;
        }
      }
    },
    addPort() {
      const id = +new Date();
      const params = {
        id,
        containerPort: '',
        protocol: 'TCP',
        name: '',
        isLink: false,
      };

      this.curContainer.ports.push(params);
    },
    addLog() {
      this.curContainer.webCache.logListCache.push({
        value: '',
      });
    },
    removeLog(log, index) {
      this.curContainer.webCache.logListCache.splice(index, 1);
    },
    changeProtocol(item) {
      const { projectId } = this;
      const version = this.curVersion;
      const portId = item.id;
      this.$store.dispatch('k8sTemplate/checkPortIsLink', { projectId, version, portId }).then((res) => {
      }, (res) => {
        const message = res.message || res.data.data;
        const msg = message.split(',')[0];
        this.$bkMessage({
          theme: 'error',
          message: msg + this.$t('deploy.templateset.cannotModifyProtocol'),
        });
      });
    },
    removePort(item, index) {
      const { projectId } = this;
      const version = this.curVersion;
      const portId = item.id;
      this.$store.dispatch('k8sTemplate/checkPortIsLink', { projectId, version, portId }).then((res) => {
        this.curContainer.ports.splice(index, 1);
      });
    },
    selectVolumeType(volumeItem) {
      // volumeItem.name = ''
      // volumeItem.mountPath = ''
      // let data = Object.assign([], this.curContainer.volumeMounts)
      // this.curContainer.volumeMounts.splice(0, this.curContainer.volumeMounts.length, ...data)
    },
    setVolumeName(volumeItem) {
      volumeItem.volume.hostPath = '';
    },
    initVolumeConfigmaps() {
      const version = this.curVersion;
      if (!version) {
        return false;
      }
      const { projectId } = this;

      this.$store.dispatch('k8sTemplate/getConfigmaps', { projectId, version }).then((res) => {
        const { data } = res;
        const keyList = [];
        data.forEach((item) => {
          const list = [];
          const { name } = item;
          const { keys } = item;
          item.id = item.name;
          keys.forEach((key) => {
            const params = {
              id: `${name}.${key}`,
              name: `${name}.${key}`,
              nameCache: name,
              keyCache: key,
            };
            list.push(params);
            keyList.push(params);
          });
          item.childList = list;
        });
        this.volumeConfigmapList = data;
        this.configmapKeyList.splice(0, this.configmapKeyList.length, ...keyList);
        this.configmapList.splice(0, this.configmapList.length, ...data);
      }, (res) => {
        const { message } = res;
        this.$bkMessage({
          theme: 'error',
          message,
        });
      });
    },
    initVloumeSelectets() {
      const version = this.curVersion;
      if (!version) {
        return false;
      }
      const { projectId } = this;

      this.$store.dispatch('k8sTemplate/getSecrets', { projectId, version }).then((res) => {
        const { data } = res;
        const keyList = [];
        data.forEach((item) => {
          const list = [];
          const { name } = item;
          const { keys } = item;
          keys.forEach((key) => {
            const params = {
              id: `${name}.${key}`,
              name: `${name}.${key}`,
              nameCache: name,
              keyCache: key,
            };
            list.push(params);
            keyList.push(params);
          });

          item.childList = list;
        });
        this.volumeSecretList = data;
        this.secretKeyList.splice(0, this.secretKeyList.length, ...keyList);
        this.secretList.splice(0, this.secretList.length, ...data);
      }, (res) => {
        const { message } = res;
        this.$bkMessage({
          theme: 'error',
          message,
        });
      });
    },
    initMetricList() {
      const { projectId } = this;
      this.$store.dispatch('k8sTemplate/getMetricList', projectId);
    },
    initLinkLabels() {
      const { projectId } = this;
      const versionId = this.curVersion;
      this.$store.dispatch('k8sTemplate/getApplicationLinkLabels', { projectId, versionId }).then((res) => {
        const { data } = res;
        for (const key in data) {
          const keys = key.split(':');
          if (keys.length >= 2) {
            this.curApplicationLinkLabels = [
              {
                key: keys[0],
                value: keys[1],
                linkMessage: this.$t('deploy.templateset.labelServiceWarning', {
                  key,
                  service: data[key].join('；'),
                }),
              },
            ];
          }
        }
      }, (res) => {
        this.curApplicationLinkLabels = [];
      });
    },
    addHostAlias() {
      this.curApplication.config.webCache.hostAliasesCache.push({
        ip: '',
        hostnames: '',
      });
    },

    removeHostAlias(item, index) {
      this.curApplication.config.webCache.hostAliasesCache.splice(index, 1);
    },
  },
};
</script>

<style scoped>
    @import './statefulset.css';
</style>
