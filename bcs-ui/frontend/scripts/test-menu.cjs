const assert = require('node:assert/strict');
const fs = require('node:fs');
const Module = require('node:module');
const path = require('node:path');
const { test } = require('node:test');
const ts = require('typescript');
const { ref } = require('vue');

const filename = path.resolve(__dirname, '../src/views/app/use-menu.ts');
const { outputText } = ts.transpileModule(fs.readFileSync(filename, 'utf8'), {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2020,
    esModuleInterop: true,
  },
});

function createMenu(initialFlags = {}, requestFlags = async () => initialFlags) {
  const flagsMap = ref(initialFlags);
  const requests = [];
  const appData = {
    flagsMap,
    async getFeatureFlags(params) {
      requests.push(params);
      flagsMap.value = await requestFlags(params);
      return flagsMap.value;
    },
  };
  const loaded = new Module(filename, module);
  loaded.filename = filename;
  loaded.paths = Module._nodeModulePaths(path.dirname(filename));
  const originalRequire = loaded.require.bind(loaded);
  loaded.require = (id) => {
    if (id === '@/composables/use-app') return { useAppData: () => appData };
    if (id === '@/i18n/i18n-setup') return { t: key => key };
    return originalRequire(id);
  };
  loaded._compile(outputText, filename);
  return { ...loaded.exports.default(), flagsMap, requests };
}

const projectFlags = {
  DEPLOYMENTMANAGE: true,
  DEPLOYMENTMANAGEPULL: false,
  PLUGINMANAGE: false,
  PLATFORMMANAGE: false,
};

function releaseRoute(meta = { menuId: 'DEPLOYMENTMANAGE' }) {
  return { name: 'releaseList', params: { projectCode: 'blueking' }, meta, matched: [] };
}

test('initial empty feature flags do not expose any navigation menus', () => {
  const { menus, leafMenus } = createMenu();
  assert.equal(menus.value.length, 0);
  assert.equal(leafMenus.value.length, 0);
});

test('null feature flags do not expose any navigation menus', () => {
  const { menus, leafMenus } = createMenu(null);
  assert.equal(menus.value.length, 0);
  assert.equal(leafMenus.value.length, 0);
});

test('menus remain hidden until the initial feature flag request resolves', async () => {
  let resolveFlags;
  const response = new Promise(resolve => { resolveFlags = resolve; });
  const { menus, leafMenus, requests, validateRouteEnable } = createMenu({}, () => response);
  const validation = validateRouteEnable(releaseRoute());
  assert.deepEqual(requests, [{ projectCode: 'blueking' }]);
  try {
    assert.equal(menus.value.length, 0);
    assert.equal(leafMenus.value.length, 0);
  } finally {
    resolveFlags(projectFlags);
  }
  assert.equal(await validation, true);
  assert.deepEqual(menus.value.map(menu => menu.id), [
    'CLUSTERRESOURCE', 'CLUSTERMANAGE', 'DEPLOYMENTMANAGE', 'PROJECTMANAGE',
  ]);
  assert.ok(leafMenus.value.some(menu => menu.route === 'releaseList'));
});

test('externally loaded flags reveal menus with their existing parent navigation links', () => {
  const { menus, leafMenus, flagsMap } = createMenu();
  assert.equal(menus.value.length, 0);
  flagsMap.value = projectFlags;
  const deploymentMenu = menus.value.find(menu => menu.id === 'DEPLOYMENTMANAGE');
  const helmMenu = deploymentMenu.children.find(menu => menu.id === 'HELM');
  const releaseMenu = leafMenus.value.find(menu => menu.route === 'releaseList');
  assert.equal(helmMenu.parent, deploymentMenu);
  assert.equal(releaseMenu.parent, helmMenu);
  assert.equal(releaseMenu.root, deploymentMenu);
});

test('loaded feature flags hide disabled parent menus and their descendants', () => {
  const { menus, leafMenus } = createMenu(projectFlags);
  assert.equal(menus.value.filter(menu => menu.title === 'nav.deploy').length, 1);
  assert.equal(menus.value.some(menu => ['PLUGINMANAGE', 'PLATFORMMANAGE'].includes(menu.id)), false);
  assert.equal(leafMenus.value.some(menu => ['TOOLS', 'LOG', 'PLATFORMPROJECT'].includes(menu.id)), false);
});

test('loaded defaults preserve menus whose feature flags are not explicitly disabled', () => {
  const { menus, leafMenus } = createMenu({
    CLOUDTOKEN: false,
    PROJECT_LIST: false,
    AZURECLOUD: true,
    IMPORTSOPSCLUSTER: true,
    PROJECTQUOTAS: false,
    PLATFORMMANAGE: false,
    DEPLOYMENTMANAGEPULL: false,
    ENABLEPOPUP: false,
  });
  assert.deepEqual(menus.value.map(menu => menu.id), [
    'CLUSTERRESOURCE', 'CLUSTERMANAGE', 'DEPLOYMENTMANAGE', 'PROJECTMANAGE', 'PLUGINMANAGE',
  ]);
  assert.ok(leafMenus.value.some(menu => menu.id === 'DEPLOYMENT'));
  assert.equal(leafMenus.value.some(menu => ['CLOUDTOKEN', 'PROJECTQUOTAS'].includes(menu.id)), false);
});

test('route validation reuses loaded flags and rejects explicitly disabled menus', async () => {
  const { validateRouteEnable, requests } = createMenu(projectFlags);
  assert.equal(await validateRouteEnable(releaseRoute()), true);
  assert.equal(await validateRouteEnable({
    name: 'platformProjectList',
    params: { projectCode: 'blueking' },
    meta: { menuId: 'PLATFORMMANAGE' },
    matched: [],
  }), false);
  assert.equal(await validateRouteEnable({ name: 'projectList', params: {}, meta: {}, matched: [] }), true);
  assert.deepEqual(requests, []);
});
