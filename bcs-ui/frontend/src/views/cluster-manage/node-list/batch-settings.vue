<template>
  <BcsContent :title="titleMap[type]" :cluster-id="clusterId">
    <Row>
      <template #left>
        <bcs-button icon="plus" @click="showAddDialog = true">{{ addBtnTextMap[type] }}</bcs-button>
      </template>
      <template #right>
        <PopoverSelector placement="bottom" offset="0,8" :on-hide="onSearchPopoverHide" ref="searchPopoverRef">
          <bcs-input
            class="min-w-[360px]"
            :placeholder="placeholderMap[type]"
            clearable
            right-icon="bk-icon icon-search"
            v-model.trim="searchKey" />
          <template #content>
            <ul class="bg-[#fff] min-w-[360px] max-h-[420px] overflow-auto" v-show="filterCheckedColList.length">
              <li
                class="bcs-dropdown-item"
                v-for="item in filterCheckedColList"
                :key="item"
                @click="handleSearchKeyChange(item)">
                {{ item }}
              </li>
            </ul>
            <!-- 无匹配搜索项的空白提示页 -->
            <div
              class="bg-[#fff] min-w-[360px] h-[64px] text-[#979BA5] text-[12px] leading-[64px] text-center"
              v-show="!filterCheckedColList.length">
              {{ $t('generic.msg.empty.noMatchLabel') }}
            </div>
          </template>
        </PopoverSelector>
      </template>
    </Row>
    <!-- 表格列配置 -->
    <div class="relative">
      <bcs-popover
        trigger="click"
        placement="bottom-end"
        theme="light custom-padding"
        class="absolute right-0 top-[16px] z-40"
        ref="settingRef">
        <div
          :class="[
            'bcs-border',
            'flex items-center justify-center w-[42px] h-[42px] bg-[#F0F1F5]',
            'text-[12px] text-[#C4C6CC] cursor-pointer'
          ]"
          @click="resetTmpCheck">
          <i
            class="bcs-icon bcs-icon-cog-shape hover:text-[#63656e]"
            v-bk-tooltips="{
              content: $t('cluster.nodeList.title.tableSetting'),
            }"
          ></i>
        </div>
        <template #content>
          <div class="pt-[24px] min-w-[420px]">
            <div class="text-[20px] leading-[26px] text-[#313238] mb-[20px] px-[24px]">
              {{ $t('cluster.nodeList.title.tableSetting') }}
            </div>
            <div
              class="flex justify-between items-center px-[24px] w-full text-[14px] text-[#63656E] h-[20px] mb-[12px]">
              <div>{{ $t('cluster.nodeList.title.fieldDisplaySettings') }}</div>
              <div class="flex items-center h-[100px] overflow-auto">
                <bcs-checkbox
                  :value="isAllKeysChecked"
                  @change="toggleCheckedKeys">
                  {{ $t('generic.button.selectAll') }}
                </bcs-checkbox>
              </div>
            </div>
            <bcs-checkbox-group
              v-model="tmpCheckedColList"
              class="!grid grid-cols-2 gap-[12px] px-[24px] overflow-y-auto max-h-[360px]">
              <bk-checkbox
                v-for="key in tableCol"
                :value="key"
                :key="key"
                class="!flex items-center">
                <div class="flex items-center">
                  <span
                    :class="[
                      'bcs-ellipsis flex-1',
                      colStatusMap[key] === 'remove'? 'text-[#C4C6CC] line-through' : ''
                    ]"
                    v-bk-overflow-tips>
                    {{ key }}
                  </span>
                  <template v-if="colStatusMap[key] === 'add'">
                    <span
                      class="text-[#2DCB56] ml-[8px]">
                      {{`[${$t('cluster.nodeList.label.thisAddition')}]`}}
                    </span>
                  </template>
                  <template v-else-if="colStatusMap[key] === 'remove'">
                    <span
                      class="text-[#FF9C01] ml-[8px]">
                      {{`[${$t('cluster.nodeList.label.thisRemoval')}]`}}
                    </span>
                  </template>
                </div>
              </bk-checkbox>
            </bcs-checkbox-group>
            <div class="bcs-border-top mt-[12px] flex items-center justify-end h-[50px] px-[24px] bg-[#FAFBFD]">
              <bcs-button theme="primary" @click="confirmSetting">
                {{ $t('generic.button.confirm') }}
              </bcs-button>
              <bcs-button class="ml10" @click="cancelSetting">{{ $t('generic.button.cancel') }}</bcs-button>
            </div>
          </div>
        </template>
      </bcs-popover>
    </div>
    <!-- 表格 -->
    <div
      class="batch-setting-table mt-[16px] w-full overflow-x-auto max-h-[calc(100%-96px)]"
      v-bkloading="{ isLoading }"
      ref="tableWrapperRef">
      <!-- 批量修改值 -->
      <div v-show="false" ref="popoverContentRef">
        <div class="leading-[22px] text-[14px] px-[16px] mt-[10px]">
          {{ $t('generic.label.batchSetting', [curEditKey]) }}
        </div>
        <div class="mt-[8px] px-[16px]">
          <bcs-input v-model="batchValue" />
        </div>
        <div class="flex items-center justify-end gap-[10px] px-[16px] mt-[12px] mb-[10px]">
          <bcs-button :disabled="!batchValue" text @click="setKeyValue">
            <span class="text-[12px]">{{ $t('generic.button.confirm') }}</span>
          </bcs-button>
          <bcs-button text @click="hidePopover">
            <span class="text-[12px]">{{ $t('generic.button.cancel') }}</span>
          </bcs-button>
        </div>
      </div>
      <table
        :class="[
          'setting-table-border',
          'text-left text-[12px]',
          'w-full'
        ]">
        <thead>
          <tr class="h-[42px] sticky top-0 z-30">
            <!-- 节点名 -->
            <th class="min-w-[240px] px-[16px] bg-[#F0F1F5] text-[#313238] sticky left-0 top-0">
              {{ $t('cluster.nodeList.label.name') }}
            </th>
            <!-- key 列 -->
            <th
              v-for="key, index in checkedColList"
              :key="key"
              :class="[
                'px-[16px] text-[#313238]',
                colStatusMap[key] === 'add' ? 'bg-[#D6F7DB]' : 'bg-[#F0F1F5]',
                highlightKey === key ? '!bg-[#E1ECFF]' : '',
                type === 'taints' ? 'w-[480px]' : 'w-[240px]'
              ]"
              :id="key">
              <div class="flex items-center">
                <template v-if="colStatusMap[key] === 'remove'">
                  <span
                    class="bcs-ellipsis flex-1 text-left text-[#C4C6CC] line-through"
                    v-bk-tooltips="{
                      disabled: formatKey(key) === key,
                      content: key,
                    }"
                  >{{ formatKey(key) }}</span>
                  <span
                    :class="[
                      'text-[12px] text-[#3A84FF] cursor-pointer',
                      {
                        'mr-[42px]': index === (checkedColList.length - 1)
                      }
                    ]" @click="undo(key)">
                    <i
                      class="bcs-icon bcs-icon-undo_line text-[16px]"
                      v-bk-tooltips="{
                        content: $t('cluster.nodeList.tips.batchUndo'),
                      }"
                    > </i>
                  </span>
                </template>
                <template v-else>
                  <span
                    class="bcs-ellipsis flex-1 text-left"
                    v-bk-tooltips="{
                      disabled: formatKey(key) === key,
                      content: key,
                    }"
                  >{{ formatKey(key) }}</span>
                  <span
                    class="text-[12px] text-[#3A84FF] cursor-pointer"
                    @mouseenter="initPopover"
                    @click="showPopover(key)">
                    <i
                      class="bcs-icon bcs-icon-batch-edit text-[16px]"
                      v-bk-tooltips="{
                        content: $t('cluster.nodeList.tips.batchEdit'),
                      }"
                    ></i>
                  </span>
                  <span
                    :class="[
                      'text-[16px] text-[#3A84FF] cursor-pointer ml-[10px] leading-none',
                      {
                        'mr-[42px]': index === (checkedColList.length - 1)
                      }
                    ]" @click="handleRemoveCol(key)">
                    <i
                      class="bk-icon icon-delete"
                      v-bk-tooltips="{
                        content: $t('cluster.nodeList.tips.batchRemove'),
                      }"
                    ></i>
                  </span>
                </template>
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data" :key="item.nodeName" class="h-[42px]">
            <td class="px-[16px] bg-[#FAFBFD] sticky left-0 z-10">{{ item.nodeName || '--' }}</td>
            <td v-for="key in checkedColList" :key="key" class="bg-[#fff] group relative">
              <!-- 删除状态 -->
              <span
                class="flex items-center h-[42px] bg-[#FAFBFD]"
                v-if="colStatusMap[key] === 'remove' || cellStatusMap[`${item.nodeName}_${key}`] === 'remove'">
                <div class="w-[240px] bcs-ellipsis flex-1 px-[16px] text-[#C4C6CC] line-through">
                  {{ item[type][key] }}
                </div>
                <!-- taintsEffect， width:241px, 多出来的1px是boreder-left的宽度 -->
                <div
                  v-if="type === 'taints'"
                  class="w-[241px] leading-[42px] bcs-ellipsis px-[16px] text-[#C4C6CC] line-through border-left">
                  {{ item.taintsEffect[key] }}
                </div>
                <span
                  class="text-[12px] text-[#3A84FF] cursor-pointer mr-[10px] hidden group-hover:inline absolute right-0"
                  @click="undoValue(item.nodeName, key)"
                  v-if="cellStatusMap[`${item.nodeName}_${key}`] === 'remove'">
                  <i
                    class="bcs-icon bcs-icon-undo_line text-[16px]"
                    v-bk-tooltips="{
                      content: $t('cluster.nodeList.tips.undo'),
                      interactive: false
                    }"
                  ></i>
                </span>
              </span>
              <!-- 其他 -->
              <span
                :class="{
                  'batch-setting-cell': true,
                  'taints': type === 'taints',
                  // value 修改状态
                  'batch-setting-cell-input-modify': cellStatusMap[`${item.nodeName}_${key}`] === 'modify',
                  // effect修改状态
                  'batch-setting-cell-select-modify': effectStatusMap[`${item.nodeName}_${key}`] === 'modify',
                  // 新增状态
                  'batch-setting-cell-add': colStatusMap[key] === 'add'
                }"
                v-else>
                <bcs-input
                  v-bk-overflow-tips
                  class="flex-1 w-[240px]"
                  v-model="item[type][key]"
                  @change="(v) => setValue(item.nodeName, key, v)"
                  @focus="curFocusCellKey = `${item.nodeName}_${key}`"
                  @blur="curFocusCellKey = ''" />
                <div class="w-[240px] overflow-hidden" v-if="type === 'taints'">
                  <bcs-select
                    class="w-[239px]"
                    v-model="item.taintsEffect[key]"
                    :clearable="false"
                    @change="(effect) => setTaintEffect(item.nodeName, key, effect)">
                    <bcs-option v-for="effect in effectList" :key="effect" :id="effect" :name="effect" />
                  </bcs-select>
                </div>
                <span
                  :class="[
                    'text-[16px] text-[#3A84FF] cursor-pointer hidden group-hover:inline',
                    'absolute right-[6px] z-10'
                  ]"
                  v-if="curFocusCellKey !== `${item.nodeName}_${key}` && (key in item[type])"
                  @click="removeValue(item.nodeName, key)">
                  <i
                    :class="['bk-icon icon-delete', colStatusMap[key] === 'add' ? 'bg-[#F2FFF4]' : 'bg-[#fff]']"
                    v-bk-tooltips="{
                      content: $t('cluster.nodeList.tips.remove'),
                      interactive: false
                    }"
                  ></i>
                </span>
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <!-- 保存操作 -->
    <div class="flex items-center sticky bottom-0 mt-[16px]">
      <bcs-button
        class="min-w-[88px]"
        theme="primary"
        :disabled="!diffTableCol.length"
        @click="showDiffDialog">{{ $t('cluster.nodeList.button.saveSetting') }}</bcs-button>
      <bcs-button @click="back">{{ $t('generic.button.cancel') }}</bcs-button>
    </div>
    <!-- 添加标签或污点 -->
    <bcs-dialog
      v-model="showAddDialog"
      :title="addBtnTextMap[type]"
      header-position="left"
      :mask-close="false"
      :width="640"
      :ok-text="$t('generic.button.add')"
      :cancel-text="$t('generic.button.cancel')"
      @confirm="confirmAdd"
      @cancel="cancelAddDialog"
      :auto-close="false">
      <!-- v-if解决keyvalue组件目前只watch一次问题 -->
      <template v-if="showAddDialog">
        <KeyValue
          v-if="kvTypes.includes(type)"
          v-model="newLabelData"
          ref="labelRef"
          :required="true"
          :value-require="false"
          :key-placeholder="$t('cluster.nodeList.placeholder.inputKey')"
          :value-placeholder="$t('cluster.nodeList.placeholder.inputValue')"
          :key-rules="[
            {
              message: $t('generic.validate.SpeciaLabelKey'),
              validator: LABEL_KEY_MAXL,
            },
            {
              message: $t('generic.validate.SpeciaLabelKey'),
              validator: LABEL_KEY_DOMAIN,
            },
            {
              message: $t('generic.validate.SpeciaLabelKey'),
              validator: LABEL_KEY_PATH,
            }
          ]"
          :value-rules="[
            {
              message: $t('generic.validate.label'),
              validator: LABEL_VALUE,
            }
          ]" />
        <Taint
          :min-item="1"
          :required="true"
          v-model="newTaintData"
          v-else-if="type === 'taints'"
          ref="taintRef"
          :key-placeholder="$t('cluster.nodeList.placeholder.inputKey')"
          :value-placeholder="$t('cluster.nodeList.placeholder.inputValue')" />
      </template>
    </bcs-dialog>
    <!-- 保存设置 -->
    <bcs-dialog
      v-model="isDiffDialogShow"
      render-directive="if"
      :title="$t('cluster.nodeList.button.saveSetting')"
      header-position="left"
      :width="1000"
      :draggable="true"
      :auto-close="false">
      <bk-alert
        v-if="isAlertShow"
        class="mb-[20px]"
        type="warning"
        :title="$t('cluster.nodeList.tips.note')">
      </bk-alert>
      <div class="text-left text-[#313238] text-[12px] mb-[16px]">
        <i18n path="cluster.nodeList.label.nodeNum" tag="span" class="text-[#979BA5] ml-[8px]">
          <span place="nodeNum" class="text-[#5E9AFE]"> {{ data.length }} </span>
        </i18n>
      </div>
      <div class="overflow-y-auto overflow-x-auto max-h-[484px] w-full">
        <table
          :class="[
            'setting-table-border',
            'table-fixed text-left w-full text-[12px]',
          ]">
          <thead>
            <tr class="h-[42px]">
              <!-- 节点名 -->
              <th class="w-[240px] px-[16px] bg-[#F0F1F5] text-[#313238] sticky left-0 top-0 z-20">
                {{ $t('cluster.nodeList.label.name') }}
              </th>
              <!-- key 列 -->
              <th
                v-for="key in diffTableCol"
                :key="key"
                class="w-[240px] px-[16px] text-[#313238] bg-[#F0F1F5] top-0 sticky">
                <div class="flex items-center">
                  <template v-if="colStatusMap[key] === 'remove'">
                    <span
                      class="bcs-ellipsis flex-1 text-left text-[#C4C6CC] line-through"
                      v-bk-tooltips="{
                        disabled: formatKey(key, true) === key,
                        content: key,
                      }"
                    >{{ formatKey(key, true) }}</span>
                  </template>
                  <template v-else>
                    <span
                      class="bcs-ellipsis flex-1 text-left"
                      v-bk-tooltips="{
                        disabled: formatKey(key, true) === key,
                        content: key,
                      }"
                    >{{ formatKey(key, true) }}</span>
                  </template>
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in data" :key="item.nodeName" class="h-[42px]">
              <td class="w-[240px] px-[16px] bg-[#FAFBFD] sticky left-0 z-10">{{ item.nodeName || '--' }}</td>
              <td v-for="key in diffTableCol" :key="key" class="bg-[#fff] group">
                <span class="bcs-ellipsis flex items-center px-[16px]">
                  <div class="w-[240px]">
                    <!-- 旧值 -->
                    <span
                      :class="{
                        'text-[#C4C6CC]': cellStatusMap[`${item.nodeName}_${key}`] === 'modify'
                          || effectStatusMap[`${item.nodeName}_${key}`] === 'modify'
                          || colStatusMap[key] === 'remove'
                          || cellStatusMap[`${item.nodeName}_${key}`] === 'remove',
                        'line-through': (colStatusMap[key] === 'remove'
                          || cellStatusMap[`${item.nodeName}_${key}`] === 'remove')
                          && originDataMap[item.nodeName]?.[type]?.[key]
                      }">
                      {{ originDataMap[item.nodeName]?.[type]?.[key] || '--' }}
                      <span v-if="type === 'taints' && originDataMap[item.nodeName].taintsEffect?.[key]">
                        ({{ originDataMap[item.nodeName].taintsEffect?.[key] }})
                      </span>
                    </span>
                    <span
                      v-if="colStatusMap[key] !== 'remove'"
                      :class="{
                        'px-[4px]': true,
                        'font-bold invisible': true,
                        'text-[#2DCB56] !visible': colStatusMap[key] === 'add'
                          && cellStatusMap[`${item.nodeName}_${key}`] !== 'remove',
                        'text-[#FF9C01] !visible': cellStatusMap[`${item.nodeName}_${key}`] === 'modify'
                          || effectStatusMap[`${item.nodeName}_${key}`] === 'modify'
                      }">
                      <i class="bcs-icon bcs-icon-arrows-right"></i>
                    </span>
                    <!-- 新值 -->
                    <span
                      v-if="
                        colStatusMap[key] !== 'remove'
                          && cellStatusMap[`${item.nodeName}_${key}`] !== 'remove'
                          && (colStatusMap[key]
                            || cellStatusMap[`${item.nodeName}_${key}`]
                            || effectStatusMap[`${item.nodeName}_${key}`])">
                      {{ item[type][key] || '--' }}
                      <span v-if="type === 'taints' && item.taintsEffect?.[key]">
                        ({{ item.taintsEffect?.[key] }})
                      </span>
                    </span>
                  </div>
                </span>

              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <template #footer>
        <div class="flex items-center justify-end">
          <bcs-button
            theme="primary"
            :loading="saving"
            :disabled="!diffTableCol.length"
            @click="handleSaveData">{{ $t('generic.button.save') }}</bcs-button>
          <bcs-button
            :disabled="saving"
            @click="isDiffDialogShow = false">{{ $t('generic.button.cancel') }}</bcs-button>
        </div>
      </template>
    </bcs-dialog>
    <bcs-dialog
      v-model="errorInfoDialogShow"
      :width="480"
      :show-footer="false"
      :auto-close="false">
      <div class="w-full px-[8px]">
        <div class="header">
          <div class="mb-[19px] mx-auto text w-[42px] h-[42px] leading-[42px] text-center bg-[#FFDDDD] rounded-[50%]">
            <i class="bcs-icon bcs-icon-close text-[15px] text-[#EA3636] font-[900]"></i>
          </div>
          <p class="text-[20px] h-[32px] leading-[32px] w-full mb-[24px]">
            {{ $t('cluster.nodeList.title.nodeSaveError') }}
          </p>
        </div>
        <div class="content">
          <div class="bg-[#F5F6FA] pl-[16px] leading-[46px] text-[#63656E] text-[14px] mb-[16px]">
            {{ $t('cluster.nodeList.title.tabelTitle') }}
          </div>
          <ul>
            <li class="bg-[#F0F1F5] h-[32px] leading-[32px] text-[14px] pl-[16px]">
              <span>{{ $t('cluster.nodeList.label.name') }}</span>
              <i
                class="text-[11px] bcs-icon bcs-icon-copy text-[#3A84FF] ml-[8px]"
                v-bk-tooltips="{
                  content: $t('cluster.nodeList.tips.copyNodeIP'),
                }"
                v-bk-copy="errNodeNameList.join('\n')"
              ></i>
            </li>
            <li
              v-for="item, index in errNodeNameList"
              class="h-[32px] leading-[32px] text-[12px] pl-[16px]"
              :key="index">
              {{ item }}
            </li>
          </ul>
        </div>
        <div class="mt-[24px] mb-[6px] text-center">
          <bcs-button class="w-[88px] h-[32px]" theme="primary" @click="gotIt">
            {{ $t('cluster.nodeList.button.gotIt') }}
          </bcs-button>
        </div>
      </div>
    </bcs-dialog>
  </BcsContent>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import Vue, { computed, onBeforeMount, ref } from 'vue';

import KeyValue from '../components/key-value.vue';
import Taint, { ITaint } from '../components/new-taints.vue';

import useProxyOperate, { SettingType } from './proxy-operate';
import useNode, { IAnnotationsItem, ILabelsItem } from './use-node';

import $bkMessage from '@/common/bkmagic';
import { LABEL_KEY_DOMAIN, LABEL_KEY_MAXL, LABEL_KEY_PATH, LABEL_VALUE, TAINT_VALUE } from '@/common/constant';
import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

interface Props {
  clusterId: string
  type: SettingType
  nodeNameList?: string
}
const props = defineProps<Props>();

const { getNodeList, batchSetNodeLabels, batchSetNodeTaints, batchSetNodeAnnotations } = useNode();
const {
  originDataMap,
  data,
  colStatusMap,
  cellStatusMap,
  effectStatusMap,
  tableCol,
  initData,
  add,
  remove,
  undo,
  undoValue,
  removeValue,
  setKey,
  setValue,
  setTaintEffect,
} = useProxyOperate(props.type);// 数据操作

// 界面标题
const titleMap: Record<SettingType, string> = {
  labels: $i18n.t('cluster.nodeList.title.batchSetLabel.text'),
  taints: $i18n.t('cluster.nodeList.title.batchSetTaint.text'),
  annotations: $i18n.t('cluster.nodeList.title.batchSetAnnotation.text'),
};
// 搜索文案
const placeholderMap: Record<SettingType, string> = {
  labels: $i18n.t('cluster.nodeList.placeholder.locationTag'),
  taints: $i18n.t('cluster.nodeList.placeholder.locationTaint'),
  annotations: $i18n.t('cluster.nodeList.placeholder.locationAnnotation'),
};
// 创建按钮文案
const addBtnTextMap: Record<SettingType, string> = {
  labels: $i18n.t('cluster.nodeList.button.addLabel'),
  taints: $i18n.t('cluster.nodeList.button.addTaint'),
  annotations: $i18n.t('cluster.nodeList.button.addAnnotation'),
};

const isLoading = ref(false);
const effectList = ['PreferNoSchedule', 'NoExecute', 'NoSchedule'];
const tableWrapperRef = ref<HTMLElement>();
const curFocusCellKey = ref('');// 当前聚焦的单元格
const kvTypes = ref(['labels', 'annotations']);

// 格式化key值，key长度超出宽度时，中间省略，只展示key首尾字符
const formatKey = (key: string, isDialog = false) => {
  let tmpKey = key;
  const displayLen = (kvTypes.value.includes(props.type) || isDialog) ? 20 : 50; // defaultLength(目前240px) * 90% / 12
  if (key.length > displayLen) {
    const start = key.substring(0, displayLen / 2 + 1);
    const end = key.substring(key.length - (displayLen / 2 - 1));
    tmpKey = `${start}...${end}`;
  }

  return tmpKey;
};

// 标签搜索
const highlightKey = ref('');
const searchPopoverRef = ref<InstanceType<typeof PopoverSelector>>();
const searchKey = ref<string>('');
const filterCheckedColList = computed(() => checkedColList.value.filter(key => key.includes(searchKey.value)));
const onSearchPopoverHide = () => {
  const exist = checkedColList.value.find(key => key === searchKey.value);
  if (!exist) {
    searchKey.value = '';
  }
};
const handleSearchKeyChange = (key: string) => {
  searchKey.value = key;
  searchPopoverRef.value?.hide();

  // 滚动到可视区域
  document.getElementById(searchKey.value)?.scrollIntoView({ behavior: 'smooth', inline: 'center' });
  // 高亮表头
  highlightKey.value = searchKey.value;
  setTimeout(() => {
    highlightKey.value = '';
  }, 3000);
};

// 表格列设置
const settingRef = ref();
const checkedColList = ref<string[]>([]);// 当前勾选项
const tmpCheckedColList = ref<string[]>([]);// 临时勾选项，确定后才生效
const isAllKeysChecked = computed(() => tmpCheckedColList.value.length === tableCol.value.length);
const toggleCheckedKeys = () => {
  if (isAllKeysChecked.value) {
    tmpCheckedColList.value = [];
  } else {
    tmpCheckedColList.value = cloneDeep(tableCol.value);
  }
};
const resetTmpCheck = () => {
  // 重置临时勾选项
  tmpCheckedColList.value = checkedColList.value;
};
const confirmSetting = () => {
  // 按照顺序添加
  checkedColList.value = tableCol.value.filter(key => tmpCheckedColList.value.includes(key));
  settingRef.value?.hideHandler();
};
const cancelSetting = () => {
  settingRef.value?.hideHandler();
};

// 添加标签或污点
const showAddDialog = ref(false);
const labelRef = ref();
const newLabelData = ref({});
const taintRef = ref();
const newTaintData = ref<ITaint[]>([]);
const addLabel = async () => {
  add(newLabelData.value);
  const keys = [...Object.keys(newLabelData.value), ...checkedColList.value];
  checkedColList.value = [...new Set(keys)]; // 添加后默认勾选上这些列,且去除重复key

  return true;
};
const addTaint = async () => {
  const data = newTaintData.value.filter(item => !!item.key);
  if (!data.length) return false;
  add(data);
  const keys = [...data.map(item => item.key), ...checkedColList.value];
  checkedColList.value = [...new Set(keys)]; // 添加后默认勾选上这些列,且去除重复key
  return true;
};
const confirmAdd = async () => {
  if (!await validateData()) return;
  let success: Boolean = false;
  if (kvTypes.value.includes(props.type)) {
    success = await addLabel();
  } else if (props.type === 'taints') {
    success = await addTaint();
  }
  // 校验不通过，不关闭弹框
  if (!success) return;

  cancelAddDialog();
  // 滚动到第一行
  if (tableWrapperRef.value) {
    tableWrapperRef.value.scrollLeft = 0;
  }
};
const validateData = async () => {
  let el: any;
  if (kvTypes.value.includes(props.type)) {
    el = labelRef.value;
  } else if (props.type === 'taints') {
    el = taintRef.value;
  }

  // 获取validate组件的校验方法，可以触发err图标提示
  const result = await Promise.all((el?.$children || [])
    .filter(el => el?.validate)
    .map(el => el.validate('blur')));

  if (result.some(item => !item)) return false;

  return true;
};


const cancelAddDialog = () => {
  showAddDialog.value = false;
  newLabelData.value = {};
  newTaintData.value = [];
};

// 删除列
const handleRemoveCol = (key: string) => {
  // 新增的列先删除勾选项
  const index = checkedColList.value.findIndex(k => k === key);
  if (colStatusMap.value[key] === 'add' && index > -1) {
    checkedColList.value.splice(index, 1);
  }
  // 移除状态
  remove(key);
  // 移除Popover
  if (instance.value) {
    instance.value?.destroy(true);
  };
};

// 批量编辑
const batchValue = ref('');
const curEditKey = ref('');
const instance = ref();
const popoverContentRef = ref();
const initPopover = (e) => {
  if (instance.value) {
    instance.value?.destroy(true);
  };
  instance.value = Vue.prototype.$bkPopover(e.target, {
    theme: 'light custom-padding',
    trigger: 'click',
    interactive: true,
    followCursor: false,
    boundary: window,
    placement: 'bottom-end',
  });
};
const hidePopover = () => {
  instance.value?.hide(100);
  batchValue.value = '';
};
const showPopover = (key: string) => {
  curEditKey.value = key;
  popoverContentRef.value.style.display = 'unset';
  instance.value.setContent(popoverContentRef.value);
  instance.value?.show(500);
};
const setKeyValue = () => {
  let validate: Boolean = true;
  batchValue.value = batchValue.value?.trim();
  if (kvTypes.value.includes(props.type)) {
    validate = new RegExp(LABEL_VALUE).test(batchValue.value);
  } else {
    validate = new RegExp(TAINT_VALUE).test(batchValue.value);
  }

  if (!validate) return $bkMessage({
    theme: 'warning',
    message: $i18n.t('generic.msg.warning.invalid'),
  });
  setKey(curEditKey.value, batchValue.value);
  hidePopover();
};

// diff数据
const saving = ref(false);
const isDiffDialogShow = ref(false);
const diffTableCol = computed(() => tableCol.value.filter((key) => {
  const isColChange = !!colStatusMap.value[key];
  const isCellChange = data.value.some(row => !!cellStatusMap.value[`${row.nodeName}_${key}`] || !!effectStatusMap.value[`${row.nodeName}_${key}`]);
  return  isColChange || isCellChange;
}));
// 预览警告，触发条件：有noExecute的taint
const isAlertShow = ref(false);
// 只diff修改的地方
const showDiffDialog = () => {
  isAlertShow.value = data.value.some(node => Object.keys(node.taintsEffect).some(key => diffTableCol.value.includes(key) && node.taintsEffect[key] === 'NoExecute'));
  isDiffDialogShow.value = true;
};
const back = () => {
  $router.back();
};
// 批量设置保存标签
const saveLabels = async (type: 'labels' | 'annotations') => {
  const nodes = data.value.map<ILabelsItem | IAnnotationsItem>((row) => {
    // 过滤标记删除的key
    const filterDeletedLabels = Object.keys(row[type])
      .reduce((pre, key) => {
        if (colStatusMap.value[key] !== 'remove') {
          if (cellStatusMap.value[`${row.nodeName}_${key}`] !== 'remove') {
            pre[key] = row[type][key];
          }
        }
        return pre;
      }, {});
    return {
      nodeName: row.nodeName,
      [type]: filterDeletedLabels,
    } as unknown as ILabelsItem | IAnnotationsItem;
  });
  let result = {
    fail: [],
    success: [],
  };
  if (type === 'labels') {
    result = await batchSetNodeLabels({
      clusterID: props.clusterId,
      nodes: nodes as ILabelsItem[],
    });
  } else if (type === 'annotations') {
    result = await batchSetNodeAnnotations({
      clusterID: props.clusterId,
      nodes: nodes as IAnnotationsItem[],
    });
  }
  return result;
};
// 批量设置保存污点
const saveTaints = async () => {
  const nodes = data.value.map((row) => {
    const filterDeletedTaints = Object.keys(row.taints)
      .filter(key => colStatusMap.value[key] !== 'remove' && cellStatusMap.value[`${row.nodeName}_${key}`] !== 'remove')
      .map(key => ({
        key,
        value: row.taints[key],
        effect: row.taintsEffect[key],
      }));
    return {
      nodeName: row.nodeName,
      taints: filterDeletedTaints,
    };
  });
  const result = await batchSetNodeTaints({
    clusterID: props.clusterId,
    nodes,
  });
  return result;
};
const errNodeNameList = ref<string[]>([]);
const errorInfoDialogShow = ref(false);
const handleSaveData = async () => {
  let result = { fail: [], success: [] };
  saving.value = true;
  if (kvTypes.value.includes(props.type)) {
    result = await saveLabels(props.type as 'labels' | 'annotations');
  } else if (props.type === 'taints') {
    result = await saveTaints();
  }
  saving.value = false;
  if (result.success?.length === data.value.length) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.save'),
    });
    $router.push({
      name: 'clusterMain',
      query: {
        active: 'node',
        clusterId: props.clusterId,
      },
    });
  } else {
    const nodeNames = result.fail?.reduce<string[]>((acc, cur: { nodeName: string }) => {
      acc.push(cur.nodeName);
      return acc;
    }, []);
    errNodeNameList.value = nodeNames;
    isDiffDialogShow.value = false;
    errorInfoDialogShow.value = true;
  }
};

const gotIt = () => {
  errorInfoDialogShow.value = false;
};

// 获取节点数据
const handleGetNodeData = async () => {
  if (!props.clusterId) return;
  isLoading.value = true;
  const list = (props.nodeNameList?.split(',') ?? []).filter(v => !!v);
  const nodeList = await getNodeList(props.clusterId);
  const filterNodeList =  list.length
    ? nodeList.filter(row => row.nodeName && list.includes(row.nodeName))
    : nodeList;
  initData(filterNodeList);
  // 初始化勾选列
  checkedColList.value = cloneDeep(tableCol.value);
  isLoading.value = false;
};

onBeforeMount(() => {
  handleGetNodeData();
});
</script>

<style lang="postcss" scoped>
/deep/.taints .bk-select .bk-select-name {
  padding: 0 36px 0 16px;
}
.taints .bk-select.is-default-trigger.is-unselected:before{
  left: 16px;
}
.setting-table-border {
  border-collapse: separate;
  border-spacing: 0;
  th {
    border-top: 1px solid #DCDEE5;
    border-bottom: 1px solid #DCDEE5;
    border-right: 1px solid #DCDEE5;
  }
  td {
    border-bottom: 1px solid #DCDEE5;
    border-right: 1px solid #DCDEE5;
  }
  th:first-child,td:first-child, .border-left {
    border-left: 1px solid #DCDEE5;
  }
}
>>> .batch-setting-table {
  .hidden {
    visibility: unset !important;
  }
}
>>> .batch-setting-cell {
  display: flex;
  align-items: center;
  position: relative;
  &-input-modify {
    &:not(.taints) {
      background-color: #FFF3E1;
    }
    .bk-form-input {
      background-color: #FFF3E1;
    }
  }
  &-select-modify {
    background-color: #FFF3E1;
  }
  &-add {
    background-color: #F2FFF4;
    .bk-form-input {
      background-color: #F2FFF4;
    }
  }
  &.taints .bk-form-input {
    &:not(:focus) {
      border-right: 1px solid #DCDEE5 !important;
    }
  }
  .bk-form-input {
    height: 42px;
    padding: 0 16px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: normal;
    word-break: break-all;
    display: -webkit-box;
    -webkit-line-clamp: 1;
    -webkit-box-orient: vertical;
    &:not(:focus) {
      border: unset;
    }
  }
  .bk-select {
    border: none;
    box-shadow: unset;
  }
}
.content {
  ul {
    border: 1px solid #EAEBF0;
    border-radius: 2px;
    li:last-child {
      background: #FAFBFD;
    }
  }
}
>>> .bk-checkbox-text {
  flex: 1;
}
</style>
<style>
.tippy-popper .custom-padding-theme {
  padding: 0;
  pointer-events: auto;
}
</style>
