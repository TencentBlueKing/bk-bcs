import Decimal from 'decimal.js';

import { isQuotaFormValueValid, quantityInFormUnit } from './other-quota';

export interface LimitRangeResourceValues {
  cpu: string
  memory: string
}

export interface PodLimitRangeForm {
  name: string
  min: LimitRangeResourceValues
  max: LimitRangeResourceValues
}

export interface PodLimitRangeValue {
  name: string
  min?: Partial<LimitRangeResourceValues>
  max?: Partial<LimitRangeResourceValues>
}

export const createPodLimitRangeForm = (): PodLimitRangeForm => ({
  name: '',
  min: { cpu: '', memory: '' },
  max: { cpu: '', memory: '' },
});

function resourceToFormValues(resource: Partial<LimitRangeResourceValues> = {}): LimitRangeResourceValues {
  return {
    cpu: resource.cpu ? quantityInFormUnit(resource.cpu, false)?.toFixed() ?? resource.cpu : '',
    memory: resource.memory ? quantityInFormUnit(resource.memory, true)?.toFixed() ?? resource.memory : '',
  };
}

export function podLimitRangeToFormValues(limitRange: PodLimitRangeValue): PodLimitRangeForm {
  return {
    name: limitRange.name,
    min: resourceToFormValues(limitRange.min),
    max: resourceToFormValues(limitRange.max),
  };
}

export function isPodLimitRangeFormValid(form: PodLimitRangeForm): boolean {
  const values = [form.min.cpu, form.min.memory, form.max.cpu, form.max.memory];
  if (!values.some(value => value.trim() !== '')) return false;
  if (!values.every(value => value.trim() === '' || isQuotaFormValueValid(value))) return false;

  const minCPUValue = form.min.cpu.trim();
  const maxCPUValue = form.max.cpu.trim();
  const minMemoryValue = form.min.memory.trim();
  const maxMemoryValue = form.max.memory.trim();
  const minCPU = minCPUValue ? new Decimal(minCPUValue) : undefined;
  const maxCPU = maxCPUValue ? new Decimal(maxCPUValue) : undefined;
  const minMemory = minMemoryValue ? new Decimal(minMemoryValue) : undefined;
  const maxMemory = maxMemoryValue ? new Decimal(maxMemoryValue) : undefined;
  return !(minCPU && maxCPU && minCPU.gt(maxCPU))
    && !(minMemory && maxMemory && minMemory.gt(maxMemory));
}

function serializeResource(
  values: LimitRangeResourceValues,
  original: Partial<LimitRangeResourceValues> | undefined,
): Partial<LimitRangeResourceValues> {
  const originalValues = resourceToFormValues(original);
  return (['cpu', 'memory'] as const).reduce((result, field) => {
    const value = values[field].trim();
    if (!value) return result;
    if (original?.[field] && new Decimal(value).eq(originalValues[field])) {
      result[field] = original[field];
    } else {
      result[field] = new Decimal(value).toFixed() + (field === 'memory' ? 'Gi' : '');
    }
    return result;
  }, {} as Partial<LimitRangeResourceValues>);
}

export function serializePodLimitRangeForm(
  form: PodLimitRangeForm,
  original?: PodLimitRangeValue,
): Pick<PodLimitRangeValue, 'min' | 'max'> {
  return {
    min: serializeResource(form.min, original?.min),
    max: serializeResource(form.max, original?.max),
  };
}
