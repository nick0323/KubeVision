/**
 * 常量定义
 */

export interface MenuItem {
  key: string;
  label: string;
  icon: string;
}

export interface MenuGroup {
  group: string;
  items: MenuItem[];
}

// 菜单配置
export const MENU_LIST: MenuGroup[] = [
  {
    group: '',
    items: [
      { key: 'overview', label: 'Overview', icon: 'FaChartPie' }, // 📊 概览图表
    ],
  },
  {
    group: 'Workloads', // 工作负载
    items: [
      { key: 'pods', label: 'Pods', icon: 'FaCube' }, // 🧊 Pod 是最小单元
      { key: 'deployments', label: 'Deployments', icon: 'FaRocket' }, // 🚀 部署应用
      { key: 'statefulsets', label: 'StatefulSets', icon: 'FaTree' }, // 🌲 有状态服务（树状结构）
      { key: 'daemonsets', label: 'DaemonSets', icon: 'FaCogs' }, // ⚙️ 守护进程
      { key: 'jobs', label: 'Jobs', icon: 'FaBriefcase' }, // 💼 任务工作
      { key: 'cronjobs', label: 'CronJobs', icon: 'FaClock' }, // 🕐 定时任务
    ],
  },
  {
    group: 'Network', // 网络（更准确的命名）
    items: [
      { key: 'services', label: 'Services', icon: 'FaNetworkWired' }, // 🌐 服务发现
      { key: 'ingress', label: 'Ingress', icon: 'FaDoorOpen' }, // 🚪 入口网关
    ],
  },
  {
    group: 'Storage', // 存储
    items: [
      { key: 'pvcs', label: 'PVCs', icon: 'FaHdd' }, // 💾 存储申请
      { key: 'pvs', label: 'PVs', icon: 'FaDatabase' }, // 🗄️ 持久卷
      { key: 'storageclasses', label: 'StorageClasses', icon: 'FaListAlt' }, // 📋 存储类定义
      { key: 'configmaps', label: 'ConfigMaps', icon: 'FaFileAlt' }, // 📄 配置文件
      { key: 'secrets', label: 'Secrets', icon: 'FaLock' }, // 🔒 密钥管理
    ],
  },
  {
    group: 'Cluster', // 集群
    items: [
      { key: 'nodes', label: 'Nodes', icon: 'FaServer' }, // 🖥️ 节点服务器
      { key: 'namespaces', label: 'Namespaces', icon: 'FaThLarge' }, // ⬜ 命名空间隔离
      { key: 'events', label: 'Events', icon: 'FaBell' }, // 🔔 事件通知
    ],
  },
];

// 分页配置
export const DEFAULT_PAGE_SIZE = 10;
export const PAGE_SIZE_OPTIONS = [10, 15, 20, 50];

