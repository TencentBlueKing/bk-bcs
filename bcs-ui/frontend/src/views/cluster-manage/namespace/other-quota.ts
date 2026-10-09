import Decimal from 'decimal.js';

const QUOTA_FIELDS = ['cpuRequests', 'cpuLimits', 'memoryRequests', 'memoryLimits'] as const;
export type QuotaField = typeof QUOTA_FIELDS[number];
export type QuotaValues = Record<QuotaField, string>;

// Keep conversion precision separate from the two-decimal table presentation.
const QuantityDecimal = Decimal.clone({ precision: 80 });
const decimalExponents: Record<string, number> = { n: -9, u: -6, m: -3, '': 0, k: 3, M: 6, G: 9, T: 12, P: 15, E: 18 };
const binaryExponents: Record<string, number> = { Ki: 10, Mi: 20, Gi: 30, Ti: 40, Pi: 50, Ei: 60 };

function quantityInFormUnit(value: string, memory: boolean): Decimal | undefined {
  const match = value.match(/^([+-]?(?:\d+\.?\d*|\.\d+))([eE][+-]?\d+|[numkMGTPE]|[KMGTPE]i)?$/);
  if (!match) return undefined;

  const suffix = match[2] || '';
  let quantity = new QuantityDecimal(match[1]);
  if (/^[eE][+-]?\d+$/.test(suffix)) {
    quantity = quantity.mul(new QuantityDecimal(10).pow(Number(suffix.slice(1))));
  } else if (suffix in binaryExponents) {
    quantity = quantity.mul(new QuantityDecimal(2).pow(binaryExponents[suffix]));
  } else {
    quantity = quantity.mul(new QuantityDecimal(10).pow(decimalExponents[suffix]));
  }
  return memory ? quantity.div(new QuantityDecimal(2).pow(30)) : quantity;
}

export function quotaToFormValues(quota: Partial<QuotaValues>): QuotaValues {
  const values: QuotaValues = { cpuRequests: '', cpuLimits: '', memoryRequests: '', memoryLimits: '' };
  QUOTA_FIELDS.forEach((field) => {
    const raw = quota[field] || '0';
    values[field] = quantityInFormUnit(raw, field.startsWith('memory'))?.toFixed() ?? raw;
  });
  return values;
}

export function isQuotaFormValueValid(value: string): boolean {
  return /^\+?(?:\d+\.?\d*|\.\d+)$/.test(value.trim());
}

export function serializeQuotaFormValues(values: QuotaValues, original?: Partial<QuotaValues>): QuotaValues {
  const initialValues = original ? quotaToFormValues(original) : undefined;
  const quota: QuotaValues = { cpuRequests: '', cpuLimits: '', memoryRequests: '', memoryLimits: '' };
  QUOTA_FIELDS.forEach((field) => {
    if (!isQuotaFormValueValid(values[field])) throw new Error(`Invalid quota value: ${field}`);
    const value = new QuantityDecimal(values[field].trim());
    // Preserve both units and precision for unchanged fields, including when another field is edited.
    if (original && initialValues && value.eq(initialValues[field])) {
      quota[field] = original[field] ?? '';
    } else {
      quota[field] = `${value.toFixed()}${field.startsWith('memory') ? 'Gi' : ''}`;
    }
  });
  return quota;
}

export function formatQuotaQuantity(value: string, type: 'cpu' | 'mem'): string {
  return quantityInFormUnit(value ?? '0', type === 'mem')?.toFixed(2) ?? '--';
}
