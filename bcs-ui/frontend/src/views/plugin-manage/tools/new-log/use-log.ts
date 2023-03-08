import {
  logEntrypoints,
  logRules,
  ruleDetail,
  createLogRule,
  updateLogRule,
  deleteLogRule,
} from '@/api/modules/monitor';

export default function useLog() {
  const handleGetEntrypoints = async ($clusterId: string) => {
    const data = await logEntrypoints({ $clusterId }).catch(() => ({}));
    return data;
  };
  const handleGetLogRules = async ($clusterId: string) => {
    const data = await logRules({ $clusterId }).catch(() => ([]));
    return data;
  };
  const handleCreateLogRule = async (params: {
    $clusterId: string
    name: string
    namespace: string
    add_pod_label: boolean
    config_selected: string
    config: any
  }) => {
    const result = await createLogRule(params).then(() => true)
      .catch(() => false);
    return result;
  };
  const handleUpdateLogRule = async (params: {
    $clusterId: string
    $name: string
    namespace: string
    add_pod_label: boolean
    config_selected: string
    config: any
  }) => {
    const result = await updateLogRule(params).then(() => true)
      .catch(() => false);
    return result;
  };
  const handleDeleteLogRule = async (params: {
    $clusterId: string
    $name: string
  }) => {
    const result = await deleteLogRule(params).then(() => true)
      .catch(() => false);
    return result;
  };
  const handleGetLogDetail = async (params: {
    $clusterId: string
    $name: string
  }) => {
    const data = await ruleDetail(params).catch(() => ({}));
    return data;
  };

  return {
    handleGetEntrypoints,
    handleGetLogRules,
    handleCreateLogRule,
    handleUpdateLogRule,
    handleDeleteLogRule,
    handleGetLogDetail,
  };
}
