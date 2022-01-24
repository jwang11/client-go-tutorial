package main

import (
	"flag"
        "fmt"
	"context"
        "path/filepath"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
        "k8s.io/client-go/util/homedir"

)

func RestClient(kubeconfig *string) {
	var err error
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	restClient, _ := rest.RESTClientFor(config)

	fmt.Printf("\n--------------RestClient-----------\n")
	res_nodes := &corev1.NodeList{}
	if err = restClient.Get().Namespace("").Resource("nodes").
                 VersionedParams(&metav1.ListOptions{Limit: 100}, scheme.ParameterCodec).
                 Do(context.TODO()).Into(res_nodes); err != nil {
		panic(err)
	}

	fmt.Printf("Node List:\n-----------\n")
	for _, d := range res_nodes.Items {
		fmt.Printf("%v\n", d.Name)
	}

	res_pods := &corev1.PodList{}
	if err = restClient.Get().Namespace(corev1.NamespaceDefault).Resource("pods").
              VersionedParams(&metav1.ListOptions{Limit: 100}, scheme.ParameterCodec).
              Do(context.TODO()).Into(res_pods); err != nil {
		panic(err)
	}
	fmt.Printf("\nPod List:\n-----------\n")
	for _, d := range res_pods.Items {
		fmt.Printf("%v/%v %v\n", d.Namespace, d.Name, d.Status.Phase)
	}
}

func ClientSet(kubeconfig *string) {
	var err error

	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	clientset, _ := kubernetes.NewForConfig(config)


	fmt.Printf("\n--------------ClientSet-----------\n")
	res_nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{Limit: 500})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Node List:\n-----------\n")
	for _, d := range res_nodes.Items {
		fmt.Printf("%v\n", d.Name)
	}


	res_pods, err := clientset.CoreV1().Pods(corev1.NamespaceDefault).List(context.TODO(), metav1.ListOptions{Limit: 500})
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nPod List:\n-----------\n")
	for _, d := range res_pods.Items {
		fmt.Printf("%v/%v %v\n", d.Namespace, d.Name, d.Status.Phase)
	}

}

func DynamicClient(kubeconfig *string) {
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	dynamicClient, _ := dynamic.NewForConfig(config)

	fmt.Printf("\n--------------DynamicClient-----------\n")
	gvr_nodes := schema.GroupVersionResource{Version: "v1", Resource: "nodes"}
	unstructNodes, _ := dynamicClient.
			Resource(gvr_nodes).
			Namespace("").
			List(context.TODO(), metav1.ListOptions{Limit: 100})
	res_nodes := &corev1.NodeList{}
	runtime.DefaultUnstructuredConverter.FromUnstructured(unstructNodes.UnstructuredContent(), res_nodes)
	fmt.Printf("Node List:\n-----------\n")
	for _, d := range res_nodes.Items {
		fmt.Printf("%v\n", d.Name)
	}

	gvr_pods := schema.GroupVersionResource{Version: "v1", Resource: "pods"}
	unstructPods, _ := dynamicClient.
			Resource(gvr_pods).
			Namespace(corev1.NamespaceDefault).
			List(context.TODO(), metav1.ListOptions{Limit: 100})
	res_pods := &corev1.PodList{}
	runtime.DefaultUnstructuredConverter.FromUnstructured(unstructPods.UnstructuredContent(), res_pods)

	fmt.Printf("\nPod List:\n-----------\n")
	for _, d := range res_pods.Items {
		fmt.Printf("%v/%v %v\n", d.Namespace, d.Name, d.Status.Phase)
	}
}

func DiscoverClient(kubeconfig *string) {
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	discoveryClient, _ := discovery.NewDiscoveryClientForConfig(config)

	fmt.Printf("\n--------------DynamicClient-----------\n")
	// 获取所有分组和资源数据
	APIGroup, APIResourceListSlice, _ := discoveryClient.ServerGroupsAndResources()
	// 先看Group信息
	fmt.Printf("APIGroup :\n %v\n\n",APIGroup)

	// APIResourceListSlice是个切片，里面的每个元素代表一个GroupVersion及其资源
	for _, singleAPIResourceList := range APIResourceListSlice {

		// GroupVersion是个字符串，例如"apps/v1"
		groupVerionStr := singleAPIResourceList.GroupVersion

		// ParseGroupVersion方法将字符串转成数据结构
		gv, _ := schema.ParseGroupVersion(groupVerionStr)

		fmt.Printf("\n*********************%v********************************\n", groupVerionStr)
		fmt.Printf("GV struct [%#v]\n", gv)

		fmt.Printf("resources:\n")
		// APIResources字段是个切片，里面是当前GroupVersion下的所有资源
		for _, singleAPIResource := range singleAPIResourceList.APIResources {
			fmt.Printf("\t%v\n", singleAPIResource.Name)
		}
	}
}


func main() {
        var kubeconfig *string

        if home := homedir.HomeDir(); home != "" {
                kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "[可选] kubeconfig 绝对路径")
        }
	flag.Parse()
	RestClient(kubeconfig)
	ClientSet(kubeconfig)
	DynamicClient(kubeconfig)
	DiscoverClient(kubeconfig)
}
