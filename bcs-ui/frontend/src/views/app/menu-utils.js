/**
 * @typedef {object} Menu
 * @property {string} id
 * @property {Menu[]} [children]
 */

/**
 * @param {Record<string, boolean> | null | undefined} featureFlags
 * @param {Menu[]} menus
 * @returns {Menu[]}
 */
function filterMenu(featureFlags, menus) {
  return menus.reduce((filteredMenus, menu) => {
    if (featureFlags && Object.prototype.hasOwnProperty.call(featureFlags, menu.id) && !featureFlags[menu.id]) {
      return filteredMenus;
    }
    filteredMenus.push(menu);
    if (menu.children?.length) {
      menu.children = filterMenu(featureFlags, menu.children);
    }
    return filteredMenus;
  }, []);
}

/**
 * @param {Record<string, boolean> | null | undefined} featureFlags
 * @param {Menu[]} menus
 * @returns {Menu[]}
 */
function getVisibleMenus(featureFlags, menus) {
  if (!featureFlags || !Object.keys(featureFlags).length) return [];
  return menus;
}

module.exports = {
  filterMenu,
  getVisibleMenus,
};
