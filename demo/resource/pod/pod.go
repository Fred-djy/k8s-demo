package pod

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Pod struct {
	ClientSet *kubernetes.Clientset
	PodName string
	Namespace string
}

var name = "zhang"

func (p *Pod) Delete(){
	fmt.Println("删除pod: " + p.PodName)
	err := p.ClientSet.CoreV1().Pods(p.Namespace).Delete(context.TODO(), p.PodName, metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("pod: " + p.PodName + "已删除")
}

func (p *Pod) GetList() {
	fmt.Println("查询pod列表")
	podList, err  := p.ClientSet.CoreV1().Pods(p.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\t %v\t %v\n", "namespace", "status", "name")
	for _, pod := range podList.Items {
		fmt.Printf("%v\t %v\t %v\n",
			pod.Namespace,
			pod.Status.Phase,
			pod.Name)
	}
}

func (p *Pod) Create(){
	fmt.Println("创建pod: " + p.PodName)
	// 创建pod
	newPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.PodName,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "nginx", Image: "nginx:latest", Ports: []corev1.ContainerPort{{ContainerPort: 80}}},
			},
		},
	}

	obj, err := p.ClientSet.CoreV1().Pods(p.Namespace).Create(context.Background(), &newPod, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("pod: " + obj.GetName() + "已经创建")
}
