package k8s

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"os"
)

var KubeClient *KubeClientContext

type KubeClientContext struct {
	Config        *restclient.Config
	Client        *clientset.Clientset
	MetricsClient *versioned.Clientset
}

// NewKubeClient 创建 kube client, bind 接口需要用此客户端建立 Pod 与 Node 的绑定关系.
func NewKubeClient() *KubeClientContext {
	// 先尝试从 ~/.kube 目录下获取配置, 如果没有, 则尝试寻找 Pod 内置的认证配置
	var kubeconfig string
	kubeconfig = "/etc/kubernetes/admin.conf"
	if _, err := os.Stat(kubeconfig); err != nil {
		klog.Warningf("kube config %s doesn't exist, buid config from InCluster", kubeconfig)
		kubeconfig = ""
	}
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Errorf("Error building kubeconfig: %s", err.Error())
	} else {
		klog.Info("building kubeconfig: %s", kubeConfig.String())
	}

	// kubeClient 用于集群内资源操作, crdClient 用于操作 crd 资源本身.
	// 具体区别目前还不清楚, 不过示例中大多都是这么做的.
	kubeClient, err := clientset.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("Error building kubernetes clientset: %s", err.Error())
	}
	metricsClient, err := versioned.NewForConfig(kubeConfig)
	if err != nil {
		klog.Errorf("failed to create metrics client: %v", err)
	}

	return &KubeClientContext{
		Config:        kubeConfig,
		Client:        kubeClient,
		MetricsClient: metricsClient,
	}
}

func (c *KubeClientContext) HandlerMetricsNodeUsage() (*v1beta1.NodeMetricsList, error) {
	// 获取所有节点的指标
	nodeMetricsList, err := c.MetricsClient.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	}
	klog.Info("HandlerMetricsNodeUsage: %s", nodeMetricsList.String())
	return nodeMetricsList, nil
}
