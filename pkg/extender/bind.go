package extender

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// Bind 将 Pod 绑定到指定节点
func (ex *Extender) Bind(args *extenderv1.ExtenderBindingArgs) *extenderv1.ExtenderBindingResult {
	klog.Infof("bind pod: %s/%s to node:%s", args.PodNamespace, args.PodName, args.Node)

	// 创建绑定关系
	binding := &corev1.Binding{
		ObjectMeta: metav1.ObjectMeta{Name: args.PodName, Namespace: args.PodNamespace, UID: args.PodUID},
		Target:     corev1.ObjectReference{Kind: "Node", APIVersion: "v1", Name: args.Node},
	}

	result := new(extenderv1.ExtenderBindingResult)
	err := ex.ClientSet.CoreV1().Pods(args.PodNamespace).Bind(context.Background(), binding, metav1.CreateOptions{})
	if err != nil {
		klog.ErrorS(err, "Failed to bind pod", "pod", args.PodName, "namespace", args.PodNamespace, "podUID", args.PodUID, "node", args.Node)
		result.Error = err.Error()
	}

	return result
}
