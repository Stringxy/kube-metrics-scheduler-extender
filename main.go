package main

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/string/kube-metrics-scheduler-extender/pkg/extender"
	"github.com/string/kube-metrics-scheduler-extender/pkg/server"
	"k8s.io/klog/v2"
	"net/http"
)

var h *server.Handler

const (
	apiPrefix      = "/scheduler"
	filterPath     = "/filter"
	prioritizePath = "/prioritize"
	preemptionPath = "/preemption"
	bindPath       = "/bind"
)

func main() {
	klog.Infof("regist pprof handler")
	server.RegistPPROF()

	klog.Infof("regist schedule extender handler")
	ws := &restful.WebService{}
	ws.Path(apiPrefix)
	// 预选过滤接口
	ws.Route(ws.POST(filterPath).To(h.Filter))
	// 优选打分接口
	ws.Route(ws.POST(prioritizePath).To(h.Prioritize))
	ws.Route(ws.POST(preemptionPath).To(h.Preempt))
	// 当核心 scheduler 调度器确定 Node 与 Pod 时的回调接口,
	// 核心调度器会把即将绑定的一对 Pod 与 Node 发送到这个接口.
	ws.Route(ws.POST(bindPath).To(h.Bind))
	restful.Add(ws)

	klog.Infof("start listening")
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		return
	}
}
func init() {
	h = server.NewHandler(extender.NewExtender())
}
