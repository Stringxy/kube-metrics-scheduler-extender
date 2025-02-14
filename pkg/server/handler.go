package server

import (
	"bytes"
	"encoding/json"
	"github.com/emicklei/go-restful/v3"
	"github.com/string/kube-metrics-scheduler-extender/pkg/extender"
	"io"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"net/http"
)

type Handler struct {
	ex *extender.Extender
}

func NewHandler(ex *extender.Extender) *Handler {
	return &Handler{ex: ex}
}
func (h *Handler) Preempt(req *restful.Request, resp *restful.Response) {
	klog.Infof("preempt request...")

	var buf bytes.Buffer
	body := io.TeeReader(req.Request.Body, &buf)

	preemptionArgs := &extenderv1.ExtenderPreemptionArgs{}
	var preemptionResult *extenderv1.ExtenderPreemptionResult

	err := json.NewDecoder(body).Decode(&preemptionArgs)
	if err != nil {
		preemptionResult = &extenderv1.ExtenderPreemptionResult{}
	} else {
		preemptionResult = h.ex.Preempt(preemptionArgs)
	}
	klog.Infof("preempt response: %+v", preemptionResult)
	err = resp.WriteAsJson(preemptionResult)
	if err != nil {
		klog.Errorf("[Preempt] failed to encode result: %v", err)
	}
}
func (h *Handler) Filter(req *restful.Request, resp *restful.Response) {
	if req.Request.Method != http.MethodPost {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var args extenderv1.ExtenderArgs
	var result *extenderv1.ExtenderFilterResult
	err := json.NewDecoder(req.Request.Body).Decode(&args)
	if err != nil {
		result = &extenderv1.ExtenderFilterResult{
			Error: err.Error(),
		}
	} else {
		result = h.ex.Filter(args)
	}
	klog.Infof("Filter response: %+v", result)
	err = resp.WriteAsJson(result)
	if err != nil {
		klog.Errorf("[Filter] failed to encode result: %v", err)
	}
}

func (h *Handler) Prioritize(req *restful.Request, resp *restful.Response) {
	if req.Request.Method != http.MethodPost {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	args := &extenderv1.ExtenderArgs{}
	var result *extenderv1.HostPriorityList

	err := json.NewDecoder(req.Request.Body).Decode(&args)
	if err != nil {
		result = &extenderv1.HostPriorityList{}
	} else {
		result = h.ex.Prioritize(args)
	}
	err = resp.WriteAsJson(result)
	if err != nil {
		klog.Errorf("[Prioritize] failed to encode result: %v", err)
	}
}

func (h *Handler) Bind(req *restful.Request, resp *restful.Response) {
	if req.Request.Method != http.MethodPost {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	args := &extenderv1.ExtenderBindingArgs{}
	var result *extenderv1.ExtenderBindingResult

	err := json.NewDecoder(req.Request.Body).Decode(&args)
	if err != nil {
		result = &extenderv1.ExtenderBindingResult{
			Error: err.Error(),
		}
	} else {
		result = h.ex.Bind(args)
	}

	err = resp.WriteAsJson(result)
	if err != nil {
		klog.Errorf("[Bind] failed to encode result: %v", err)
	}
}
