import { IListTemplateMetadataItem, ITemplateSpaceData } from '@/@types/cluster-resource-patch';

export interface IFileTreeNode {
  name: string;              // 当前节点名称（最后一段，如 "prod"）
  path: string;              // 完整路径（如 "bkenv/prod"）
  isFolder: boolean;         // 是否为文件夹节点
  children: IFileTreeNode[]; // 子节点列表
  file?: IListTemplateMetadataItem; // 叶子文件节点对应的文件数据
  expanded?: boolean;        // 是否展开
  isTemp?: boolean;          // 是否为临时创建的文件夹
  isSpace?: boolean;         // 是否为空间节点
  spaceData?: ITemplateSpaceData; // 空间节点对应的原始数据
  loading?: boolean;         // 是否加载中（空间节点使用）
  spaceID?: string;          // 所属空间ID（用于统一树中子节点标识所属空间）
}

/**
 * 将扁平文件列表 + 临时文件夹路径构建为树形结构
 * @param files 文件列表（name 字段可能包含 "/" 路径分隔符）
 * @param tempFolders 临时文件夹路径数组（如 ["bkenv", "bkenv/prod"]）
 * @returns 树形节点数组
 */
export function buildFileTree(
  files: IListTemplateMetadataItem[] = [],
  tempFolders: string[] = [],
  spaceID?: string,
): IFileTreeNode[] {
  const root: IFileTreeNode[] = [];
  const nodeMap = new Map<string, IFileTreeNode>();

  // 确保路径上的所有中间文件夹节点存在
  function ensureFolderNode(path: string, isTemp = false): IFileTreeNode {
    if (nodeMap.has(path)) return nodeMap.get(path)!;

    const segments = path.split('/');
    const name = segments[segments.length - 1];
    const node: IFileTreeNode = {
      name,
      path,
      isFolder: true,
      children: [],
      expanded: false,
      isTemp,
    };
    nodeMap.set(path, node);

    // 确保父节点存在
    if (segments.length > 1) {
      const parentPath = segments.slice(0, -1).join('/');
      const parentNode = ensureFolderNode(parentPath, isTemp);
      if (!parentNode.children.some(c => c.path === path)) {
        parentNode.children.push(node);
      }
    } else {
      // 顶层节点
      if (!root.some(c => c.path === path)) {
        root.push(node);
      }
    }
    return node;
  }

  // 先处理临时文件夹
  tempFolders.forEach((folderPath) => {
    ensureFolderNode(folderPath, true);
  });

  // 处理文件列表
  files.forEach((file) => {
    const fileName = file.name || '';
    const segments = fileName.split('/');

    if (segments.length <= 1) {
      // 无路径分隔符，直接作为顶层文件节点
      const fileNode: IFileTreeNode = {
        name: fileName,
        path: fileName,
        isFolder: false,
        children: [],
        file,
        spaceID,
      };
      root.push(fileNode);
    } else {
      // 含路径分隔符，需要创建中间文件夹节点
      const folderPath = segments.slice(0, -1).join('/');
      const parentNode = ensureFolderNode(folderPath);

      // 在父节点下添加文件节点（仅取文件名最后一段显示）
      const fileNode: IFileTreeNode = {
        name: segments[segments.length - 1],
        path: fileName,
        isFolder: false,
        children: [],
        file,
        spaceID,
      };
      // 避免重复添加
      if (!parentNode.children.some(c => c.path === fileName && !c.isFolder)) {
        parentNode.children.push(fileNode);
      }
    }
  });

  // 排序：文件夹在前，文件在后，同类型按名称排序
  function sortNodes(nodes: IFileTreeNode[]): IFileTreeNode[] {
    nodes.sort((a, b) => {
      if (a.isFolder !== b.isFolder) return a.isFolder ? -1 : 1;
      return a.name.localeCompare(b.name);
    });
    nodes.forEach(node => sortNodes(node.children));
    return nodes;
  }

  return sortNodes(root);
}

/**
 * 按搜索关键字过滤树节点
 * 当 searchValue 包含 "/" 时，按路径前缀匹配
 * 当 searchValue 不包含 "/" 时，按名称模糊匹配
 * @param tree 树形节点数组
 * @param searchValue 搜索关键字
 * @returns 过滤后的树形节点数组
 */
export function filterTreeBySearch(
  tree: IFileTreeNode[],
  searchValue: string,
): IFileTreeNode[] {
  if (!searchValue) return tree;

  const lowerSearch = searchValue.toLowerCase();

  function matchNode(node: IFileTreeNode): IFileTreeNode | null {
    if (searchValue.includes('/')) {
      // 路径搜索：按路径前缀匹配
      const nodePath = node.path.toLowerCase();
      if (nodePath.startsWith(lowerSearch) || lowerSearch.startsWith(nodePath)) {
        return { ...node };
      }
      // 在子节点中查找
      const matchedChildren = node.children
        .map(child => matchNode(child))
        .filter(Boolean) as IFileTreeNode[];
      if (matchedChildren.length > 0) {
        return { ...node, children: matchedChildren };
      }
      return null;
    }

    // 普通搜索：名称模糊匹配
    const nameMatch = node.name.toLowerCase().includes(lowerSearch);
    const pathMatch = node.path.toLowerCase().includes(lowerSearch);

    const matchedChildren = node.children
      .map(child => matchNode(child))
      .filter(Boolean) as IFileTreeNode[];

    if (nameMatch || pathMatch || matchedChildren.length > 0) {
      return { ...node, children: matchedChildren };
    }
    return null;
  }

  return tree
    .map(node => matchNode(node))
    .filter(Boolean) as IFileTreeNode[];
}

/**
 * 从文件名解析出中间路径段
 * 如 "bkenv/prod/ingress.yaml" -> ["bkenv", "bkenv/prod"]
 * @param fileName 文件名
 * @returns 路径段数组
 */
export function parseFolderPaths(fileName: string): string[] {
  const segments = fileName.split('/');
  if (segments.length <= 1) return [];

  const paths: string[] = [];
  for (let i = 1; i < segments.length; i++) {
    paths.push(segments.slice(0, i).join('/'));
  }
  return paths;
}

/**
 * 在树中查找指定路径的节点
 * @param tree 树形节点数组
 * @param path 目标路径
 * @returns 找到的节点或 null
 */
export function findNodeByPath(tree: IFileTreeNode[], path: string): IFileTreeNode | null {
  for (const node of tree) {
    if (node.path === path) return node;
    const found = findNodeByPath(node.children, path);
    if (found) return found;
  }
  return null;
}

/**
 * 切换节点展开状态
 * @param tree 树形节点数组
 * @param path 目标路径
 * @returns 是否找到并切换
 */
export function toggleNodeExpanded(tree: IFileTreeNode[], path: string): boolean {
  const node = findNodeByPath(tree, path);
  if (node && node.isFolder) {
    node.expanded = !node.expanded;
    return true;
  }
  return false;
}
