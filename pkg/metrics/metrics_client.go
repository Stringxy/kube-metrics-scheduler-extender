package metrics

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type NodeMetricsClient struct {
	KubernetesClient *kubernetes.Clientset
	MetricsClient    *versioned.Clientset
}

func NewMetricsClient(kubeconfig string) (*NodeMetricsClient, error) {
	// 使用 kubeconfig 文件创建配置
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
	}

	// 创建标准的 Kubernetes 客户端
	kubernetesClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// 创建 metrics 客户端
	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %v", err)
	}

	return &NodeMetricsClient{
		KubernetesClient: kubernetesClient,
		MetricsClient:    metricsClient,
	}, nil
}

// HandlerMetricsNodeUsage 获取 Node 使用情况
func (c *NodeMetricsClient) HandlerMetricsNodeUsage() (*v1beta1.NodeMetricsList, error) {
	// 获取所有节点的指标
	nodeMetricsList, err := c.MetricsClient.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	}
	return nodeMetricsList, nil
}
