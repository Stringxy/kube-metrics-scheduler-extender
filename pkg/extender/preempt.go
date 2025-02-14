package extender

import (
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// Preempt 处理 Kubernetes 调度器扩展程序的抢占请求...
func (ex *Extender) Preempt(args *extenderv1.ExtenderPreemptionArgs) *extenderv1.ExtenderPreemptionResult {
	klog.Infof(
		"preemption request: Pod: %+v, NodeNameToVictims: %+v, NodeNameToMetaVictims: %+v",
		args.Pod, args.NodeNameToVictims, args.NodeNameToMetaVictims,
	)
	return &extenderv1.ExtenderPreemptionResult{
		NodeNameToMetaVictims: args.NodeNameToMetaVictims,
	}
}
