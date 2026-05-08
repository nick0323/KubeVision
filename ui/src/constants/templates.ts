/**
 * 预定义资源模板
 */

export interface ResourceTemplate {
  name: string;
  label: string;
  icon: string;
  description: string;
  yaml: string;
}

// Deployment 模板
export const DEPLOYMENT_TEMPLATE: ResourceTemplate = {
  name: 'deployment',
  label: 'Deployment',
  icon: 'FaRocket',
  description: 'Stateless application deployment with rolling updates and rollbacks',
  yaml: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
  labels:
    app: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-app
        image: nginx:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 256Mi`,
};

// Service 模板
export const SERVICE_TEMPLATE: ResourceTemplate = {
  name: 'service',
  label: 'Service',
  icon: 'FaNetworkWired',
  description: 'Network service exposing applications (ClusterIP/NodePort/LoadBalancer)',
  yaml: `apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: default
  labels:
    app: my-app
spec:
  type: ClusterIP
  selector:
    app: my-app
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
    name: http`,
};

// ConfigMap 模板
export const CONFIGMAP_TEMPLATE: ResourceTemplate = {
  name: 'configmap',
  label: 'ConfigMap',
  icon: 'FaFileAlt',
  description: 'Store non-sensitive configuration data',
  yaml: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: default
data:
  key1: value1
  key2: value2
  config.yaml: |
    server:
      port: 8080
    database:
      host: localhost
      port: 5432`,
};

// Secret 模板
export const SECRET_TEMPLATE: ResourceTemplate = {
  name: 'secret',
  label: 'Secret',
  icon: 'FaLock',
  description: 'Store sensitive information (passwords, tokens, keys)',
  yaml: `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm
  api-key: c2VjcmV0LWFwaS1rZXk=`,
};

// Pod 模板
export const POD_TEMPLATE: ResourceTemplate = {
  name: 'pod',
  label: 'Pod',
  icon: 'FaCube',
  description: 'Basic Kubernetes compute unit - single container or multiple',
  yaml: `apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  namespace: default
  labels:
    app: my-app
spec:
  containers:
  - name: my-container
    image: nginx:latest
    ports:
    - containerPort: 80
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 256Mi
    volumeMounts:
    - name: config-volume
      mountPath: /etc/config
  volumes:
  - name: config-volume
    configMap:
      name: my-config`,
};

// PVC 模板
export const PVC_TEMPLATE: ResourceTemplate = {
  name: 'pvc',
  label: 'PersistentVolumeClaim',
  icon: 'FaHdd',
  description: 'Request persistent storage',
  yaml: `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  namespace: default
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: standard`,
};

// Ingress 模板
export const INGRESS_TEMPLATE: ResourceTemplate = {
  name: 'ingress',
  label: 'Ingress',
  icon: 'FaDoorOpen',
  description: 'HTTP/HTTPS routing rules for external access',
  yaml: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: myapp.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-service
            port:
              number: 80`,
};

// StatefulSet 模板
export const STATEFULSET_TEMPLATE: ResourceTemplate = {
  name: 'statefulset',
  label: 'StatefulSet',
  icon: 'FaTree',
  description: 'Stateful application deployment with ordered pods',
  yaml: `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: my-statefulset
  namespace: default
spec:
  serviceName: my-service
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-app
        image: nginx:latest
        ports:
        - containerPort: 80
        volumeMounts:
        - name: data
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi`,
};

// DaemonSet 模板
export const DAEMONSET_TEMPLATE: ResourceTemplate = {
  name: 'daemonset',
  label: 'DaemonSet',
  icon: 'FaCogs',
  description: 'Run one Pod replica on each node',
  yaml: `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: my-daemonset
  namespace: default
spec:
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-app
        image: nginx:latest
        resources:
          limits:
            cpu: 100m
            memory: 128Mi`,
};

// Job 模板
export const JOB_TEMPLATE: ResourceTemplate = {
  name: 'job',
  label: 'Job',
  icon: 'FaBriefcase',
  description: 'Run one-time tasks that complete and exit',
  yaml: `apiVersion: batch/v1
kind: Job
metadata:
  name: my-job
  namespace: default
spec:
  template:
    spec:
      containers:
      - name: my-job
        image: busybox:latest
        command: ["sh", "-c", "echo Hello from Job && sleep 10"]
      restartPolicy: Never
  backoffLimit: 4`,
};

// CronJob 模板
export const CRONJOB_TEMPLATE: ResourceTemplate = {
  name: 'cronjob',
  label: 'CronJob',
  icon: 'FaClock',
  description: 'Scheduled tasks that run periodically based on Cron expressions',
  yaml: `apiVersion: batch/v1
kind: CronJob
metadata:
  name: my-cronjob
  namespace: default
spec:
  schedule: "0 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: my-cronjob
            image: busybox:latest
            command: ["sh", "-c", "echo Hello from CronJob"]
          restartPolicy: OnFailure`,
};

// Namespace 模板
export const NAMESPACE_TEMPLATE: ResourceTemplate = {
  name: 'namespace',
  label: 'Namespace',
  icon: 'FaThLarge',
  description: 'Virtual clusters for resource isolation',
  yaml: `apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
  labels:
    name: my-namespace`,
};

// 所有模板列表
export const RESOURCE_TEMPLATES: ResourceTemplate[] = [
  POD_TEMPLATE,
  DEPLOYMENT_TEMPLATE,
  SERVICE_TEMPLATE,
  CONFIGMAP_TEMPLATE,
  SECRET_TEMPLATE,
  PVC_TEMPLATE,
  INGRESS_TEMPLATE,
  STATEFULSET_TEMPLATE,
  DAEMONSET_TEMPLATE,
  JOB_TEMPLATE,
  CRONJOB_TEMPLATE,
  NAMESPACE_TEMPLATE,
];

// 根据资源类型获取模板
export function getTemplateByName(name: string): ResourceTemplate | undefined {
  return RESOURCE_TEMPLATES.find(t => t.name === name);
}

// 根据资源类型获取模板
export function getTemplateByResourceType(resourceType: string): ResourceTemplate | undefined {
  return RESOURCE_TEMPLATES.find(t => t.name === resourceType);
}
