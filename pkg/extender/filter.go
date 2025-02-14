package extender

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// Filter 过滤掉不满足label条件的节点
func (ex *Extender) Filter(args extenderv1.ExtenderArgs) *extenderv1.ExtenderFilterResult {
	//klog.Info("Filter received Nodes: %s", args.Nodes)
	//klog.Info("Filter received Pod: %s", args.Pod)
	//klog.Info("Filter received NodeNames: %s", args.NodeNames)
	// 过滤掉不满足条件的节点
	nodes := make([]v1.Node, 0)
	nodeNames := make([]string, 0)

	for _, node := range args.Nodes.Items {
		_, ok := node.Labels[Label]
		klog.Infof("Node[ %s ] Special Label Value: %t", node.Name, ok)
		if !ok { // 排除掉不带指定标签的节点
			continue
		}
		nodes = append(nodes, node)
		nodeNames = append(nodeNames, node.Name)
	}

	// 没有满足条件的节点就报错
	if len(nodes) == 0 {
		return &extenderv1.ExtenderFilterResult{Error: fmt.Errorf("all node do not have label %s", Label).Error()}
	}

	args.Nodes.Items = nodes
	return &extenderv1.ExtenderFilterResult{
		Nodes:     args.Nodes, // 当 NodeCacheCapable 设置为 false 时会使用这个值
		NodeNames: &nodeNames, // 当 NodeCacheCapable 设置为 true 时会使用这个值
	}
}
