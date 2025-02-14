package extender

import (
	"github.com/string/kube-metrics-scheduler-extender/pkg/k8s"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// Prioritize 给 Pod 打分
// 注意：此处返回得分 Scheduler 会将其与其他插件打分合并后再选择节点，因此这里的逻辑不能完全控制最终的调度结果。
// 想要完全控制调度结果，只能在 Filter 接口中实现，过滤掉不满足条件的节点，并对剩余节点进行打分，最终 Filter 接口只返回得分最高的那个节点
func (ex *Extender) Prioritize(args *extenderv1.ExtenderArgs) *extenderv1.HostPriorityList {
	var result extenderv1.HostPriorityList
	klog.Infof("Initializing kube client for Prioritize")
	kubeClientContext := k8s.NewKubeClient()
	nodeMetricsList, err := kubeClientContext.HandlerMetricsNodeUsage()
	if err != nil {
		klog.Errorf("Failed to get node metrics: %v", err)
	}
	var minCpuUsage int64 = -1
	// Find the node with the minimum CPU usage
	for _, node := range nodeMetricsList.Items {
		cpuUsage := node.Usage.Cpu().MilliValue()
		if minCpuUsage == -1 || cpuUsage < minCpuUsage {
			minCpuUsage = cpuUsage
		}
	}
	// Calculate scores based on the minimum CPU usage
	for _, node := range nodeMetricsList.Items {
		cpuUsage := node.Usage.Cpu().MilliValue()
		score := int64(0)
		if minCpuUsage != -1 {
			score = minCpuUsage - cpuUsage + 1 // Higher score for lower CPU usage
		}
		klog.Infof("Node [ %s ] Cpu Usage Value: %d, Score: %d", node.Name, cpuUsage, score)
		result = append(result, extenderv1.HostPriority{
			Host:  node.Name,
			Score: score,
		})
	}
	return &result
}
