const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');
const ts = require('typescript');
const { createFeatureFlagsLoader } = require('../src/composables/feature-flags-utils');

const useAppDataPath = path.resolve(__dirname, '../src/composables/use-app.ts');
const useAppDataSource = ts.createSourceFile(
  useAppDataPath,
  fs.readFileSync(useAppDataPath, 'utf8'),
  ts.ScriptTarget.Latest,
  true,
  ts.ScriptKind.TS,
);
const defaultFlags = { DEFAULT_FEATURE: false };


function createLoader(requestFlags) {
  const updates = [];
  return {
    getFeatureFlags: createFeatureFlagsLoader({
      requestFlags,
      updateFeatureFlags: flags => updates.push(flags),
      defaultFlags,
    }),
    updates,
  };
}

test('getFeatureFlags is safely hoisted for cyclic API initialization', () => {
  const getFeatureFlagsDeclaration = useAppDataSource.statements.find(statement => (
    ts.isFunctionDeclaration(statement) && statement.name?.text === 'getFeatureFlags'
  ));
  const useAppDataDeclaration = useAppDataSource.statements.find(statement => (
    ts.isFunctionDeclaration(statement) && statement.name?.text === 'useAppData'
  ));
  let returnsHoistedGetter = false;

  function findGetterReturn(node) {
    if (ts.isShorthandPropertyAssignment(node) && node.name.text === 'getFeatureFlags') {
      returnsHoistedGetter = true;
    }
    ts.forEachChild(node, findGetterReturn);
  }

  if (useAppDataDeclaration) findGetterReturn(useAppDataDeclaration);
  assert.ok(getFeatureFlagsDeclaration);
  assert.ok(returnsHoistedGetter);
});

test('only the latest concurrent request updates feature flags', async () => {
  let resolveOldRequest;
  let resolveLatestRequest;
  const oldResponse = new Promise(resolve => { resolveOldRequest = resolve; });
  const latestResponse = new Promise(resolve => { resolveLatestRequest = resolve; });
  const { getFeatureFlags, updates } = createLoader(({ projectCode }) => (
    projectCode === 'old' ? oldResponse : latestResponse
  ));

  const oldRequest = getFeatureFlags({ projectCode: 'old' });
  const latestRequest = getFeatureFlags({ projectCode: 'latest' });

  resolveLatestRequest({ LATEST_FEATURE: true });
  assert.deepEqual(await latestRequest, { LATEST_FEATURE: true });
  resolveOldRequest({ LATEST_FEATURE: false, OLD_FEATURE: true });
  assert.deepEqual(await oldRequest, { LATEST_FEATURE: false, OLD_FEATURE: true });

  assert.deepEqual(updates, [{ DEFAULT_FEATURE: false, LATEST_FEATURE: true }]);
});

test('a stale request failure cannot overwrite the latest feature flags', async () => {
  let rejectOldRequest;
  let resolveLatestRequest;
  const oldResponse = new Promise((resolve, reject) => { rejectOldRequest = reject; });
  const latestResponse = new Promise(resolve => { resolveLatestRequest = resolve; });
  const { getFeatureFlags, updates } = createLoader(({ projectCode }) => (
    projectCode === 'old' ? oldResponse : latestResponse
  ));

  const oldRequest = getFeatureFlags({ projectCode: 'old' });
  const latestRequest = getFeatureFlags({ projectCode: 'latest' });
  resolveLatestRequest({ LATEST_FEATURE: true });
  await latestRequest;
  rejectOldRequest(new Error('old request failed'));

  assert.deepEqual(await oldRequest, {});
  assert.deepEqual(updates, [{ DEFAULT_FEATURE: false, LATEST_FEATURE: true }]);
});

test('a feature flag request failure resolves with defaults instead of rejecting', async () => {
  const { getFeatureFlags, updates } = createLoader(async () => {
    throw new Error('feature flag request failed');
  });

  const request = getFeatureFlags({ projectCode: 'failed' });
  await assert.doesNotReject(request);
  assert.deepEqual(await request, {});
  assert.deepEqual(updates, [{ DEFAULT_FEATURE: false }]);
});

test('merging request results does not mutate default flags', async () => {
  const results = [
    Promise.resolve({ DEFAULT_FEATURE: true }),
    Promise.reject(new Error('feature flag request failed')),
  ];
  const { getFeatureFlags, updates } = createLoader(() => results.shift());

  assert.deepEqual(await getFeatureFlags({ projectCode: 'success' }), { DEFAULT_FEATURE: true });
  assert.deepEqual(await getFeatureFlags({ projectCode: 'failed' }), {});
  assert.deepEqual(defaultFlags, { DEFAULT_FEATURE: false });
  assert.deepEqual(updates, [
    { DEFAULT_FEATURE: true },
    { DEFAULT_FEATURE: false },
  ]);
});
