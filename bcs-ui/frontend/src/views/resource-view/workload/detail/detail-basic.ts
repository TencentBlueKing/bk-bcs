import { computed } from 'vue';

export interface IDetailBasicOptions {
  category: string;
  detail: any;
}

export default function detailBasicList(options: IDetailBasicOptions) {
  const basicInfoMap = {
    deployments: [
      {
        label: 'Ready',
        valueKeys: ['status.readyReplicas', 'spec.replicas'],
        delimiter: ' / ',
        defaultValue: 0,
      }, {
        label: 'Up-to-date',
        valueKeys: ['status.updatedReplicas'],
        defaultValue: 0,
      }, {
        label: 'Available',
        valueKeys: ['status.availableReplicas'],
        defaultValue: 0,
      },
    ],
    daemonsets: [
      {
        label: 'Desired',
        valueKeys: ['status.desiredNumberScheduled'],
        defaultValue: 0,
      }, {
        label: 'Current',
        valueKeys: ['status.currentNumberScheduled'],
        defaultValue: 0,
      }, {
        label: 'Ready',
        valueKeys: ['status.numberReady'],
        defaultValue: 0,
      }, {
        label: 'Up-to-date',
        valueKeys: ['status.updatedNumberScheduled'],
        defaultValue: 0,
      }, {
        label: 'Available',
        valueKeys: ['status.numberAvailable'],
        defaultValue: 0,
      },
    ],
    statefulsets: [
      {
        label: 'Ready',
        valueKeys: ['status.readyReplicas', 'spec.replicas'],
        delimiter: ' / ',
        defaultValue: 0,
      }, {
        label: 'Up-to-date',
        valueKeys: ['status.updatedReplicas'],
        delimiter: ' / ',
        defaultValue: 0,
      },
    ],
    cronjobs: [
      {
        label: 'Schedule',
        valueKeys: ['spec.schedule'],
        defaultValue: '--',
      }, {
        label: 'Suspend',
        valueKeys: ['spec.suspend'],
      }, {
        label: 'Active',
        valueKeys: ['active'],
        dataBasicKey: 'manifestExt',
      }, {
        label: 'Last Schedule',
        valueKeys: ['lastSchedule'],
        dataBasicKey: 'manifestExt',
        delimiter: ' / ',
        defaultValue: 0,
      },
    ],
    jobs: [
      {
        label: 'Completions',
        valueKeys: ['status.succeeded', 'spec.completions'],
        defaultValue: '--',
      }, {
        label: 'Duration',
        valueKeys: ['duration'],
        dataBasicKey: 'manifestExt',
      },
    ],
  };
  const curBasicInfo = basicInfoMap[options.category] || [];
  const hasOwnProperty = (data, key) => Object.prototype.hasOwnProperty.call(data, key);
  const basicInfoList = computed(() => curBasicInfo.map((item) => {
    const dataKey = hasOwnProperty(item, 'dataBasicKey') ? item.dataBasicKey : 'manifest';
    const data = options.detail.value?.[dataKey];
    const hasDefaultValue = hasOwnProperty(item, 'defaultValue');
    const hasDelimiter = hasOwnProperty(item, 'delimiter');
    const values = item.valueKeys.map((key) => {
      const childrenKey = key.split('.');
      let value = data?.[childrenKey[0]];
      if (childrenKey.length > 1) {
        value = childrenKey.reduce((acc, item) => acc?.[item], data);
      }
      return hasDefaultValue ? value || item.defaultValue : value;
    });
    return {
      label: item.label,
      value: hasDelimiter ? values.join(item.delimiter) : values[0],
    };
  }));

  return basicInfoList;
}
