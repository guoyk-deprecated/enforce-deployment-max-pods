# enforce-deployment-max-pods

## 使用方式

* 初始化 `admission-bootstrapper`
  参照此文档 https://github.com/k8s-autoops/admission-bootstrapper ，完成 `admission-bootstrapper` 的初始化步骤
* 部署以下 YAML

```yaml
# create serviceaccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: enforce-deployment-max-pods
  namespace: autoops
---
# create clusterrole
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: enforce-deployment-max-pods
rules:
  - apiGroups: [ "apps" ]
    resources: [ "deployments", "replicasets" ]
    verbs: [ "get" ]
  - apiGroups: [ "" ]
    resources: [ "pods" ]
    verbs: [ "list" ]
---
# create clusterrolebinding
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: enforce-deployment-max-pods
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: enforce-deployment-max-pods
subjects:
  - kind: ServiceAccount
    name: enforce-deployment-max-pods
    namespace: autoops
---
# create job
apiVersion: batch/v1
kind: Job
metadata:
  name: install-enforce-deployment-max-pods
  namespace: autoops
spec:
  template:
    spec:
      serviceAccount: admission-bootstrapper
      containers:
        - name: admission-bootstrapper
          image: autoops/admission-bootstrapper
          env:
            - name: ADMISSION_NAME
              value: enforce-deployment-max-pods
            - name: ADMISSION_IMAGE
              value: autoops/enforce-deployment-max-pods
            - name: ADMISSION_ENVS
              value: "MAX_PODS=50"
            - name: ADMISSION_SERVICE_ACCOUNT
              value: "enforce-deployment-max-pods"
            - name: ADMISSION_MUTATING
              value: "true"
            - name: ADMISSION_IGNORE_FAILURE
              value: "false"
            - name: ADMISSION_SIDE_EFFECT
              value: "None"
            - name: ADMISSION_RULES
              value: '[{"operations":["CREATE"],"apiGroups":[""], "apiVersions":["*"], "resources":["pods"]}]'
      restartPolicy: OnFailure
```

## Credits

Guo Y.K., MIT License
