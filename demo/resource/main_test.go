package main

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"testing"
	"time"
)

func TestPod(t *testing.T)  {
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	client, _ := kubernetes.NewForConfig(config)

	test := []struct{
		client *kubernetes.Clientset
		method string
	}{
		{client, "create"},
		{client, "update"},
		{client, "search"},
		{client, "delete"},
	}

	for _,tt := range test{
		newPod(tt.client, tt.method)
		time.Sleep(time.Second)
	}
}

func TestDeployment(t *testing.T)  {
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	client, _ := dynamic.NewForConfig(config)

	test := []struct{
		client dynamic.Interface
		method string
	}{
		{client, "create"},
		{client, "update"},
		{client, "search"},
		{client, "delete"},
	}

	for _,tt := range test{
		newDeployment(tt.client, tt.method)
		time.Sleep(time.Second)
	}
}

func TestCrd(t *testing.T)  {
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	test := []struct{
		config *rest.Config
		method string
	}{
		{config, "create"},
		{config, "update"},
		{config, "search"},
		{config, "delete"},
	}

	for _,tt := range test{
		newCrd(tt.config, tt.method)
		time.Sleep(time.Second)
	}
}