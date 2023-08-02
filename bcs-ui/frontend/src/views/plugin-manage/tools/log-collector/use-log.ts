import {
  addonsDetail,
  updateOns as aliasUpdateOns,
} from '@/api/modules/helm';
import {
  logCollectorRules as aliasLogCollectorRules,
  retryLogCollectorRule as aliasRetryLogCollectorRule,
  enableLogCollector as aliasEnableLogCollector,
  disableLogCollector as aliasDisableLogCollector,
  deleteLogCollectorRule as aliasDeleteLogCollectorRule,
  logCollectorDetail as aliasLogCollectorDetail,
  createLogCollectorRule as aliasCreateLogCollectorRule,
  modifyLogCollectorRule as aliasModifyLogCollectorRule,
} from '@/api/modules/monitor';
import { dashbordList } from '@/api/modules/cluster-resource';
import { ref } from 'vue';

interface IAddOnsParams {
  $clusterId: string
  $name: string
}

interface IAddOnsData {
  name: string
  chartName: string
  description: string
  logo: string
  docsLink: string
  version: string
  currentVersion: string
  namespace: string
  defaultValues: string
  currentValues: string
  status: string
  message: string
  supportedActions: string[]
}

interface IUpdateOnsParams {
  $name: string
  $clusterId: string
  version: string
  values: string
}

export interface ISeparatorFilter {
  'fieldindex': string
  'word': string
  'op': '='
  'logic_op': 'and' | 'or'
}
export interface IRuleData {
  id?: string
  status?: string
  updator?: string
  updated_at?: string
  old?: boolean
  new_rule_id?: boolean
  new_rule_name?: string// 前端拼接数据
  new_rule_created_at?: string // 前端拼接数据
  name: string
  description: string
  entrypoint?: {
    file_log_url: string
    file_bk_base_url: string
    std_log_url: string
    std_bk_base_url: string
  }
  rule: {
    add_pod_label: boolean
    extra_labels: {key: string, value: string}[]
    data_info?: {
      file_bkdata_data_id: number
      std_bkdata_data_id: number
    }
    config: {
      namespaces: string[]
      paths: string[]
      data_encoding: string
      enable_stdout: boolean
      label_selector: {
        match_labels: {key: string, value: string}[]
      }
      conditions: {
        type: string
        match_type: string
        match_content: string
        separator: string
        separator_filters: ISeparatorFilter[]
      }
      container: {
        workload_type: string
        workload_name: string
        container_name: string
      }
    }
  }
  from_rule: string
}

export default function useLog() {
  const onsData = ref<Partial<IAddOnsData>>({});
  async function getOnsDetail(params: IAddOnsParams) {
    onsData.value = await addonsDetail(params).catch(() => ({}));

    return onsData.value;
  }

  const updateLoading = ref(false);
  async function updateOns(params: IUpdateOnsParams) {
    updateLoading.value = true;
    const result = await aliasUpdateOns(params).then(() => true)
      .catch(() => false);
    updateLoading.value = false;
    return result;
  }

  const ruleList = ref<IRuleData[]>([]);
  async function logCollectorRules(params: { $clusterId: string }) {
    const data = await aliasLogCollectorRules(params).catch(() => []);
    const ruleMap = data.reduce((pre, item) => {
      pre[item.id] = item;
      return pre;
    }, {});

    ruleList.value = data.map(item => ({
      ...item,
      new_rule_name: ruleMap[item.new_rule_id]?.name,
      new_rule_created_at: ruleMap[item.new_rule_id]?.created_at,
    }));
    return ruleList.value;
  }

  async function retryLogCollectorRule(params: { $clusterId: string, $ID: string }) {
    const result = await aliasRetryLogCollectorRule(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function enableLogCollector(params: { $clusterId: string, $ID: string }) {
    const result = aliasEnableLogCollector(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function disableLogCollector(params: { $clusterId: string, $ID: string }) {
    const result = aliasDisableLogCollector(params).then(() => true)
      .catch(() => false);
    return result;
  }

  async function deleteLogCollectorRule(params: { $clusterId: string, $ID: string }) {
    const result = aliasDeleteLogCollectorRule(params).then(() => true)
      .catch(() => false);
    return result;
  }

  const logDetailLoading = ref(false);
  async function logCollectorDetail(params: {$clusterId: string, $ID: string}) {
    if (!params.$clusterId || !params.$ID) return;
    logDetailLoading.value = false;
    const data = await aliasLogCollectorDetail(params).catch(() => {});
    logDetailLoading.value = true;
    return data;
  }

  async function createLogCollectorRule(data: IRuleData & { $clusterId: string }) {
    const result = await aliasCreateLogCollectorRule(data).catch(() => false);

    return result;
  }

  async function modifyLogCollectorRule(data: IRuleData & { $ID: string, $clusterId: string }) {
    const result = await aliasModifyLogCollectorRule(data).then(() => true)
      .catch(() => false);
    return result;
  }

  async function getWorkloadList(params: {
    $namespaceId: string
    $clusterId: string
    $category: string
  }) {
    const data = await dashbordList({
      ...params,
      $type: 'workloads',
    }).catch(() => []);
    return data;
  }

  return {
    onsData,
    getOnsDetail,
    updateLoading,
    updateOns,
    ruleList,
    logCollectorRules,
    retryLogCollectorRule,
    enableLogCollector,
    disableLogCollector,
    deleteLogCollectorRule,
    logDetailLoading,
    logCollectorDetail,
    createLogCollectorRule,
    modifyLogCollectorRule,
    getWorkloadList,
  };
}
