package crd

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"log"
)

type Crd struct {
	Dr        dynamic.ResourceInterface
	Name      string
	Namespace string
	Obj       *unstructured.Unstructured
}


func (c *Crd) Create() {
	//创建
	objCreate, err := c.Dr.Create(context.TODO(), c.Obj, metav1.CreateOptions{})
	if err != nil {
		//panic(fmt.Errorf("Create resource ERROR: %v", err))
		log.Print(err)
	}
	log.Print("Create: : ", objCreate)
}

func (c *Crd) Get() {
	// 查询
	objGET, err := c.Dr.Get(context.TODO(), "test", metav1.GetOptions{})
	if err != nil {
		panic(fmt.Errorf("select resource ERROR: %v", err))
	}
	jsonBytes, err := json.Marshal(objGET)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(jsonBytes))
}

func (c *Crd) Update() {
	//更新
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() (err error) {
		// 查询resource是否存在
		result, getErr := c.Dr.Get(context.TODO(), c.Obj.GetName(), metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("failed to get latest version of : %v", getErr))
		}

		// 提取obj 的 spec 期望值
		spec, found, err := unstructured.NestedMap(c.Obj.Object, "spec")
		fmt.Print("get spec:", spec)
		if err != nil || !found || spec == nil {
			panic(fmt.Errorf(" not found or error in spec: %v", err))
		}
		// 更新 存在资源的spec
		if err := unstructured.SetNestedMap(result.Object, spec, "spec"); err != nil {
			panic(err)
		}
		// 更新资源
		objUpdate, err := c.Dr.Update(context.TODO(), result, metav1.UpdateOptions{})
		log.Print("update : ", objUpdate)
		return err
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	} else {
		log.Print("更新成功")
	}

}

func (c *Crd) Delete() {
	//删除
	err := c.Dr.Delete(context.TODO(), c.Obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		panic(fmt.Errorf("delete resource ERROR : %v", err))
	} else {
		log.Print("删除成功")
	}
}
