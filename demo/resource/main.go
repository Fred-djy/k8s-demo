package main

import (
	"context"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"resource-demo/crd"
	"resource-demo/deployment"
	"resource-demo/pod"
)

var kubeconfig *string
var name string
var namespace string
var kind string
var method string

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig file")
	}

	flag.StringVar(&method, "method", "create", "增删改查：create delete update search")
	flag.StringVar(&name, "name", "demo-pod", "资源名字")
	flag.StringVar(&kind, "kind", "Pod", "资源类型，例如：pod、deployment、daemonSet、job、crd")
	flag.StringVar(&namespace, "namespace", "default", "命名空间")
}

func main() {
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	switch kind {
	case "pod":
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}
		newPod(client, method)
		break
	case "deployment":
		// 使用clientSet也行，这里使用dynamic
		client, err := dynamic.NewForConfig(config)
		if err != nil {
			panic(err)
		}
		newDeployment(client, method)
		break
	case "crd":
		newCrd(config, method)
	default:
		break
	}
}

func newCrd(config *rest.Config, method string) {
	// 创建dynamic客户端
	dynamicClient, err := dynamic.NewForConfig(config)
	// 创建discovery客户端
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return
	}

	// 自定义数据
	var metaCRD = `
apiVersion: "cs.handpay.cn/v1"
kind: Redis
metadata:
  name: test
  namespace: default
spec:
  schedule: "2022-11-17T10:12:00Z"
  command: "echo redis crd2!"
  replicas: 2
  phase: "Running"
`

	obj := &unstructured.Unstructured{}
	_, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode([]byte(metaCRD), nil, obj)

	// 获取GVK GVR 映射
	mapperGVRGVK := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))
	// 根据资源GVK 获取资源的GVR GVK映射
	resourceMapper, err := mapperGVRGVK.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return
	}

	var dr dynamic.ResourceInterface
	if resourceMapper.Scope.Name() == meta.RESTScopeNameNamespace {
		// 获取gvr对应的动态客户端
		dr = dynamicClient.Resource(resourceMapper.Resource).Namespace(namespace)
	} else {
		// 获取gvr对应的动态客户端
		dr = dynamicClient.Resource(resourceMapper.Resource)
	}

	if err != nil {
		panic(fmt.Errorf("failed to get dr: %v", err))
	}

	crd := crd.Crd{
		Dr:        dr,
		Obj:       obj,
		Name:      name,
		Namespace: namespace,
	}
	if err != nil {
		panic(fmt.Errorf("failed to get GVK: %v", err))
	}

	switch method {
	case "create":
		crd.Create()
		break
	case "delete":
		crd.Delete()
	case "update":
		// 更新
		crd.Update()
		break
	case "search":
		crd.Get()
		break
	}
}

func newPod(client *kubernetes.Clientset, method string) {
	podObject := pod.Pod{
		ClientSet: client,
		PodName:   name,
		Namespace: namespace,
	}

	switch method {
	case "create":
		podObject.Create()
		break
	case "delete":
		podObject.Delete()
	case "update":
		// TODO
		break
	case "search":
		// 查询pod
		//_, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})

		// 查询podlist
		podObject.GetList()
		break
	}
}

func newDeployment(client dynamic.Interface, method string) {
	deploymentObject := deployment.Deployment{
		Client:    client,
		Name:      name,
		Namespace: namespace,
	}

	switch method {
	case "create":
		deploymentObject.Create()
		break
	case "delete":
		// TODO
		break
	case "update":
		// TODO
		break
	case "search":
		// TODO
		break
	}

}

// 新建namespace
func createNamespace(client *kubernetes.Clientset) {
	fmt.Println("创建namespace: " + namespace)
	namespaceClient := client.CoreV1().Namespaces()

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err := namespaceClient.Create(context.TODO(), namespace, metav1.CreateOptions{})

	if err != nil {
		fmt.Println(err)
	}
}
