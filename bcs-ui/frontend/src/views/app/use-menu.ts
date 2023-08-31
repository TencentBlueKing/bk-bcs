import menusData, { IMenu } from './menus';
import { useCluster, useConfig, useAppData } from '@/composables/use-app';
import { computed } from 'vue';
import { has } from 'lodash';
import { Route } from 'vue-router';

export default function useMenu() {
  const parseTreeMenuToMap = (menus: IMenu[], initialValue = {}, parent?: IMenu) => (
    menus.reduce<Record<string, IMenu>>((pre, item) => {
      item.parent = parent;// 父节点
      if (!item.id) {
        console.warn('menu id is null', item);
      } else if (pre[item.id]) {
        console.warn('menu id is repeat', item);
      } else {
        pre[item.id] = item;
      }
      if (item.children?.length) {
        pre = parseTreeMenuToMap(item.children, pre, item);
      }
      return pre;
    }, initialValue)
  );
  // 因为ref里面不能存有递归关闭的数据，这里缓存一份含有parent指向的map数据
  const menusDataMap = parseTreeMenuToMap(menusData);

  const { flagsMap, getFeatureFlags } = useAppData();
  // 过滤未开启feature_flag的菜单
  const filterMenu = (featureFlags: Record<string, boolean>, data: IMenu[]) => data.reduce<IMenu[]>((pre, menu) => {
    if (has(featureFlags, menu.id) && !featureFlags[menu.id]) return pre; // 未开启菜单项
    pre.push(menu);
    if (menu.children?.length) {
      menu.children = filterMenu(featureFlags, menu.children);
    }
    return pre;
  }, []);
  const menus = computed<IMenu[]>(() => filterMenu(flagsMap.value, menusData));
  // 扁平化子菜单
  const flatLeafMenus = (menus: IMenu[], root?: IMenu) => {
    const data: IMenu[] = [];
    for (const item of menus) {
      const rootMenu = root ?? item;
      if (item.children?.length) {
        data.push(...flatLeafMenus(item.children, rootMenu));
      } else {
        data.push({
          root: rootMenu,
          ...item,
        });
      }
    }
    return data;
  };
  const allLeafMenus = computed(() => flatLeafMenus(menusData));
  // 所有路由父节点只是用于分组（指向子路由），真正的菜单项是子节点
  const getCurrentMenuByRoute = (name: string, id?: string) => allLeafMenus.value
    .find(item => item.route === name || item.id === id);
  // 校验菜单是否开启
  const validateMenuID = (menu: IMenu): boolean => {
    if (has(flagsMap.value, menu?.id || '') && !flagsMap.value[menu?.id || '']) {
      return false;
    }
    // 如果父菜单没有开启，则子菜单也不能开启
    if (menu?.parent) {
      return validateMenuID(menu.parent);
    }
    return true;
  };
  const validateRouteEnable = async (route: Route) => {
    if (!route.params.projectCode) return true; // 处理根路由
    // 首次加载时获取feature_flag数据
    if (!flagsMap.value || !Object.keys(flagsMap.value)?.length) {
      await getFeatureFlags({ projectCode: route.params.projectCode });
    }
    // 路由配置上带有menuId（父菜单ID）或 ID（当前菜单ID）, 先判断配置的ID是否开启了feature_flag
    if (route.meta?.id && has(flagsMap.value, route.meta?.id) && !flagsMap.value[route.meta.id]) {
      return false;
    }
    if (route.meta?.menuId && has(flagsMap.value, route.meta.menuId) && !flagsMap.value[route.meta.menuId]) {
      return false;
    }
    // 直接返回的菜单项不包含parent信息, 需要去menusDataMap找含有parent信息的菜单项
    const menuID = getCurrentMenuByRoute(route.name || '')?.id || '';
    return menusDataMap[menuID] ? validateMenuID(menusDataMap[menuID]) : true;
  };
  // 共享集群禁用菜单
  const { _INTERNAL_ } = useConfig();
  const { isSharedCluster } = useCluster();
  const disabledMenuIDs = computed(() => (isSharedCluster.value
    ? [
      'DAEMONSET',
      'PERSISTENTVOLUME',
      'STORAGECLASS',
      'HPA',
      'CRD',
      'CUSTOMOBJECT',
    ]
    : []));

  return {
    menus,
    disabledMenuIDs,
    flatLeafMenus,
    validateRouteEnable,
  };
}
