package main

import (
	"flag"
	"fmt"
	"time"
	"path/filepath"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var err error
	var config *rest.Config

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "[可选] kubeconfig 绝对路径")
	}
	// 初始化 rest.Config 对象
	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			panic(err.Error())
		}
	}
	// 创建 Clientset 对象
	clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
		panic(err.Error())
	}
	// 初始化一个 SharedInformerFactory 设置resync为60秒一次，会触发UpdateFunc
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*60)
	// 对 Deployment 监听
	//这里如果获取v1betav1的deployment的资源
	// informerFactory.Apps().V1beta1().Deployments()
	deployInformer := informerFactory.Apps().V1().Deployments()
	// 创建 Informer（相当于注册到工厂中去，这样下面启动的时候就会去 List & Watch 对应的资源）
	informer := deployInformer.Informer()
	// 创建 deployment的 Lister
	deployLister := deployInformer.Lister()
	// 注册事件处理程序 处理事件数据
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})

	stopper := make(chan struct{})
	defer close(stopper)

	informerFactory.Start(stopper)
	informerFactory.WaitForCacheSync(stopper)

	// 从本地缓存中获取 default 命名空间中的所有 deployment 列表
	deployments, err := deployLister.Deployments("default").List(labels.Everything())
	if err != nil {
		panic(err)
	}
	for idx, deploy := range deployments {
		fmt.Printf("%d -> %v/%s\n", idx+1, deploy.Namespace, deploy.Name)
	}

	<-stopper
}

func onAdd(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	fmt.Println("add a deployment:", deploy.Name)
}

func onUpdate(old, new interface{}) {
	oldDeploy := old.(*v1.Deployment)
	newDeploy := new.(*v1.Deployment)
	fmt.Println("update deployment:", oldDeploy.Name, newDeploy.Name)
}

func onDelete(obj interface{}) {
	deploy := obj.(*v1.Deployment)
	fmt.Println("delete a deployment:", deploy.Name)
}
