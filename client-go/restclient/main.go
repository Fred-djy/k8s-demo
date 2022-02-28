package main

import (
	"context"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func main() {
	var kubeconfig *string

	// home是家目录，如果能取得家目录的值，就可以用来做默认值
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "kubeconfig file")
	} else {
		// 如果取不到当前用户的家目录，就没办法设置kubeconfig的默认目录了，只能从入参中取
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	// kubeconfig加载失败就直接退出了
	if err != nil {
		panic(err.Error())
	}

	// 参考path : /api/v1/namespaces/{namespace}/pods
	config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs

	// 根据配置信息构建restClient实例
	restClient, err := rest.RESTClientFor(config)

	if err != nil {
		panic(err.Error())
	}

	// 保存pod结果的数据结构实例
	result := &corev1.PodList{}

	//  指定namespace
	namespace := "kube-system"
	// GET请求
	err = restClient.Get().
		// /api/v1/namespaces/{namespace}/pods
		Namespace(namespace).
		//api/v1/namespaces/{namespace}/pods
		Resource("pods").
		VersionedParams(&metav1.ListOptions{Limit: 100}, scheme.ParameterCodec).
		// 请求
		Do(context.TODO()).
		// 结果存入result
		Into(result)

	if err != nil {
		panic(err.Error())
	}

	// 打印名称
	fmt.Printf("Namespace\t Status\t\t Name\n")

	// 每个pod都打印Namespace、Status.Phase、Name三个字段
	for _, d := range result.Items {
		fmt.Printf("%v\t %v\t %v\n",
			d.Namespace,
			d.Status.Phase,
			d.Name)
	}
}
