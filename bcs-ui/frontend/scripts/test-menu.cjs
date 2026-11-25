const assert = require('node:assert/strict');
const { test } = require('node:test');
const { filterMenu, getVisibleMenus } = require('../src/views/app/menu-utils');

const projectFlags = {
  DEPLOYMENTMANAGE: true,
  DEPLOYMENTMANAGEPULL: false,
  PLUGINMANAGE: false,
  PLATFORMMANAGE: false,
};

function createMenus() {
  return [
    {
      id: 'DEPLOYMENTMANAGE',
      title: '部署管理',
      children: [
        { id: 'HELM', children: [{ id: 'RELEASELIST', route: 'releaseList' }] },
        { id: 'TEMPLATE_FILE', route: 'templatefile' },
      ],
    },
    {
      id: 'DEPLOYMENTMANAGEPULL',
      title: '部署管理',
      children: [{ id: 'PULL_RELEASELIST', route: 'releaseList' }],
    },
    {
      id: 'PLUGINMANAGE',
      title: '插件管理',
      children: [{ id: 'TOOLS', route: 'dbCrdcontroller' }],
    },
    {
      id: 'PLATFORMMANAGE',
      title: '平台管理',
      children: [{ id: 'PLATFORMPROJECT', route: 'platformProjectList' }],
    },
  ];
}

test('empty feature flags keep all navigation menus hidden', () => {
  const menus = filterMenu({}, createMenus());
  assert.deepEqual(getVisibleMenus({}, menus), []);
});

test('null feature flags keep all navigation menus hidden', () => {
  const menus = filterMenu(null, createMenus());
  assert.deepEqual(getVisibleMenus(null, menus), []);
});

test('menus appear after feature flags load and the duplicate deployment menu stays hidden', () => {
  const pendingFlags = {};
  assert.deepEqual(getVisibleMenus(pendingFlags, filterMenu(pendingFlags, createMenus())), []);

  const menus = filterMenu(projectFlags, createMenus());
  const visibleMenus = getVisibleMenus(projectFlags, menus);
  assert.deepEqual(visibleMenus.map(menu => menu.id), ['DEPLOYMENTMANAGE']);
  assert.equal(visibleMenus[0].title, '部署管理');
  assert.equal(visibleMenus[0].children[0].children[0].route, 'releaseList');
});

test('disabled parent menus and their descendants are filtered out', () => {
  const menus = filterMenu(projectFlags, createMenus());
  const menuIDs = menus.flatMap(menu => [menu.id, ...(menu.children || []).map(child => child.id)]);
  assert.equal(menuIDs.some(id => ['DEPLOYMENTMANAGEPULL', 'PLUGINMANAGE', 'PLATFORMMANAGE', 'TOOLS', 'PLATFORMPROJECT'].includes(id)), false);
});

test('feature flags omitted from the response do not disable menu items', () => {
  const menus = filterMenu({ PLATFORMMANAGE: false }, createMenus());
  assert.deepEqual(menus.map(menu => menu.id), ['DEPLOYMENTMANAGE', 'DEPLOYMENTMANAGEPULL', 'PLUGINMANAGE']);
});
