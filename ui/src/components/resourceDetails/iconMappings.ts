/**
 * Icon mappings for all resource detail components
 * Using react-icons/fa to match sidebar icon style
 */

// Workloads
export const POD_ICONS = {
  info: 'FaBox',           // 📦
  containers: 'FaBox',     // 📦
  initContainers: 'FaFlask', // 🔬
  volumes: 'FaBook',       // 📖
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

export const DEPLOYMENT_ICONS = {
  info: 'FaRocket',        // 🚀
  selector: 'FaBullseye',  // 🎯
  containers: 'FaBox',     // 📦
  labels: 'FaBook',        // 📖
  relatedPods: 'FaLink',   // 🔗
  events: 'FaCalendarAlt', // 📅
};

export const STATEFULSET_ICONS = {
  info: 'FaRocket',        // 🚀
  selector: 'FaBullseye',  // 🎯
  relatedPods: 'FaLink',   // 🔗
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

export const DAEMONSET_ICONS = {
  info: 'FaServer',        // 🖥️
  selector: 'FaBullseye',  // 🎯
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

export const JOB_ICONS = {
  info: 'FaBriefcase',     // 💼
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

export const CRONJOB_ICONS = {
  info: 'FaClock',         // ⏰
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

// Network
export const SERVICE_ICONS = {
  info: 'FaNetworkWired',  // 🌐
  ports: 'FaPlug',         // 🔌
  endpoints: 'FaLink',     // 🔗
  selector: 'FaBullseye',  // 🎯
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

export const INGRESS_ICONS = {
  info: 'FaExchangeAlt',   // 🔀
  rules: 'FaGlobe',        // 🌐
  tls: 'FaLock',           // 🔒
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

// Storage
export const PVC_ICONS = {
  info: 'FaHdd',           // 💾
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

export const PV_ICONS = {
  info: 'FaDatabase',      // 🗄️
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

export const STORAGECLASS_ICONS = {
  info: 'FaHdd',           // 💾
  labels: 'FaBook',        // 📖
};

// Config
export const CONFIGMAP_ICONS = {
  info: 'FaFileAlt',       // 📄
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

export const SECRET_ICONS = {
  info: 'FaLock',          // 🔒
  labels: 'FaBook',        // 📖
  events: 'FaCalendarAlt', // 📅
};

// Cluster
export const NODE_ICONS = {
  info: 'FaServer',        // 🖥️
  system: 'FaMicrochip',   // 🔧
  resources: 'FaClock',    // ⏱️
  labels: 'FaBook',        // 📖
  annotations: 'FaBook',   // 📖
  pods: 'FaLink',          // 🔗
  events: 'FaCalendarAlt', // 📅
};

export const NAMESPACE_ICONS = {
  info: 'FaCube',          // 📦
  labels: 'FaBook',        // 📖
  annotations: 'FaBook',   // 📖
  events: 'FaCalendarAlt', // 📅
};
