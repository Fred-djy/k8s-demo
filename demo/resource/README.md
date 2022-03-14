### 实现Pod资源 、Deployment、Crd 资源的增删改查

使用方法
```
# go run main.go -kubeconfig=/root/.kube/config -kind=pod -name=demo -namespace=default
```

自定义参数

```
-kind string 资源类型，例如：pod、deployment、daemonSet、job、crd (default "Pod")
-kubeconfig string kubeconfig file (default "/home/.kube/config")
-name string 资源名字 (default "demo-pod")
-namespace string 命名空间 (default "default")
```

