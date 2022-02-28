package main

import (
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func main() {

	var kubeconfig *string
	// home是家目录，如果能取得家目录的值，就可以用来做默认值
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// 从本机加载kubeconfig配置文件，因此第一个参数为空字符串
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	// kubeconfig加载失败就直接退出了
	if err != nil {
		panic(err.Error())
	}

	// 新建discoveryClient实例
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	// 获取所有分组和资源数据
	APIGroup, APIResourceListSlice, err := discoveryClient.ServerGroupsAndResources()

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("APIGroup :\n\n %v\n\n\n\n", APIGroup)

	// APIResourceListSlice里面的每个元素代表一个GroupVersion及其资源
	for _, singleAPIResourceList := range APIResourceListSlice {
		groupVerionStr := singleAPIResourceList.GroupVersion

		// ParseGroupVersion方法将字符串转成数据结构
		gv, err := schema.ParseGroupVersion(groupVerionStr)

		if err != nil {
			panic(err.Error())
		}

		fmt.Println("*****************************************************************")
		fmt.Printf("GV string [%v]\nGV struct [%#v]\nresources :\n\n", groupVerionStr, gv)

		// 里面是当前GroupVersion下的所有资源
		for _, singleAPIResource := range singleAPIResourceList.APIResources {
			fmt.Printf("%v\n", singleAPIResource.Name)
		}
	}
}
