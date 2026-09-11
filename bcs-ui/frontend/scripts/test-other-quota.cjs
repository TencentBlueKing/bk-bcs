// Run with: node scripts/test-other-quota.cjs
// Use the existing TypeScript compiler; no additional test dependencies are needed.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const Module = require('node:module');
const path = require('node:path');
const { test } = require('node:test');
const ts = require('typescript');
const { parseComponent } = require('vue-template-compiler');

const namespaceDir = path.resolve(__dirname, '../src/views/cluster-manage/namespace');

function loadScript(filename, mocks = {}) {
  const file = fs.readFileSync(filename, 'utf8');
  const source = filename.endsWith('.vue') ? parseComponent(file).script.content : file;
  const { outputText } = ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020, esModuleInterop: true },
  });
  const loaded = new Module(filename, module);
  loaded.filename = filename;
  loaded.paths = Module._nodeModulePaths(path.dirname(filename));
  const originalRequire = loaded.require.bind(loaded);
  loaded.require = id => (Object.hasOwn(mocks, id) ? mocks[id] : originalRequire(id));
  loaded._compile(outputText, filename);
  return loaded.exports;
}

const helpers = loadScript(path.join(namespaceDir, 'other-quota.ts'));
const original = { cpuRequests: '1m', cpuLimits: '15m', memoryRequests: '1Mi', memoryLimits: '400M' };
const converted = {
  cpuRequests: '0.001',
  cpuLimits: '0.015',
  memoryRequests: '0.0009765625',
  memoryLimits: '0.37252902984619140625',
};

function createDetail() {
  const requests = [];
  const messages = [];
  const events = [];
  const component = loadScript(path.join(namespaceDir, 'detail.vue'), {
    './other-quota': helpers,
    '@/api/modules/project': {
      createOtherQuota: async params => requests.push({ action: 'create', params }),
      updateOtherQuota: async params => requests.push({ action: 'update', params }),
    },
    '@/common/bkmagic': message => messages.push(message),
    '@/common/util': { timeZoneTransForm: value => value },
    '@/components/bk-magic-2.0/bk-info': () => {},
    '@/i18n/i18n-setup': { t: key => key },
  }).default;
  const state = component.setup({ data: { name: 'test-ns' }, clusterId: 'BCS-K8S-1', editable: true }, {
    emit: event => events.push(event),
  });
  return { state, requests, messages, events };
}

test('edit conversion is lossless and unchanged values preserve their original units', () => {
  assert.deepEqual(helpers.quotaToFormValues(original), converted);
  assert.deepEqual(helpers.serializeQuotaFormValues(converted, original), original);
  assert.deepEqual(helpers.serializeQuotaFormValues({ ...converted, cpuLimits: '0.0150' }, original), original);
});

test('decimal SI, binary SI, raw bytes and exponents round trip without truncation', () => {
  for (const value of ['400M', '400000000', '4e8', '4E+8', '100Mi', '1Gi', '1Ki', '1m', '0']) {
    const quota = { ...original, memoryLimits: value };
    const form = helpers.quotaToFormValues(quota);
    assert.ok(helpers.isQuotaFormValueValid(form.memoryLimits), value);
    assert.deepEqual(helpers.serializeQuotaFormValues(form, quota), quota);
  }
  assert.equal(helpers.quotaToFormValues({ memoryLimits: '4e8' }).memoryLimits, converted.memoryLimits);
  assert.equal(helpers.formatQuotaQuantity('400M', 'mem'), '0.37');
  assert.equal(helpers.formatQuotaQuantity('15m', 'cpu'), '0.02');
});

test('opening edit and immediately saving sends the original quantities', async () => {
  const { state, requests, events } = createDetail();
  state.handleEditQuota({ name: 'extra-quota', quota: original });
  assert.deepEqual(state.quotaDialog.value.form.quota, converted);
  await state.handleSaveQuota();
  assert.deepEqual(requests, [{ action: 'update', params: {
    $clusterId: 'BCS-K8S-1', $namespace: 'test-ns', $quotaName: 'extra-quota', quota: original,
  } }]);
  assert.deepEqual(events, ['refresh']);
});

test('editing one field does not change other quantities', async () => {
  const { state, requests } = createDetail();
  state.handleEditQuota({ name: 'extra-quota', quota: original });
  state.quotaDialog.value.form.quota.memoryLimits = '0.001953125';
  await state.handleSaveQuota();
  assert.deepEqual(requests[0].params.quota, { ...original, memoryLimits: '0.001953125Gi' });
});

test('create after edit resets original values and preserves entered precision', async () => {
  const { state, requests } = createDetail();
  state.handleEditQuota({ name: 'extra-quota', quota: original });
  state.handleCreateQuota();
  state.quotaDialog.value.form.name = 'new-quota';
  Object.assign(state.quotaDialog.value.form.quota, converted);
  await state.handleSaveQuota();
  assert.equal(requests[0].action, 'create');
  assert.deepEqual(requests[0].params.quota, {
    ...converted, memoryRequests: `${converted.memoryRequests}Gi`, memoryLimits: `${converted.memoryLimits}Gi`,
  });
});

test('invalid input is rejected before making a request', async () => {
  for (const value of ['', ' ', '-1', 'NaN', 'Infinity', '1Mi']) {
    const { state, requests, messages } = createDetail();
    state.handleEditQuota({ name: 'extra-quota', quota: original });
    state.quotaDialog.value.form.quota.cpuRequests = value;
    await state.handleSaveQuota();
    assert.equal(requests.length, 0, value);
    assert.equal(messages[0].theme, 'error', value);
  }
  assert.ok(helpers.isQuotaFormValueValid('0'));
});
