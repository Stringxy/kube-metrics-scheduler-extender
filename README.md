# k8s 自定义调度器逻辑之 Scheduler Extender


实现一个读取Metrics-Server指标打分的简单Scheduler Extender。

功能如下：

* 1）过滤阶段：仅调度到带有 Label `filter.xy.com` 的节点上
* 2）打分阶段：读取Metrics-Server指标，CPU利用率最低的节点为优
<!-- more -->

## Scheduler Extender 规范

Scheduler Extender 通过 HTTP 请求的方式，将调度框架阶段中的调度决策委托给外部的调度器，然后将调度结果返回给调度框架。

我们只需要实现一个 HTTP 服务，然后通过配置文件将其注册到调度器中，就可以实现自定义调度器。

通过 Scheduler Extender 扩展原有调度器一般分为以下两步：

* 1）创建一个 HTTP 服务，实现对应接口
* 2）修改调度器配置 KubeSchedulerConfiguration，增加 extenders 相关配置



外置调度器可以影响到三个阶段：

* Filter：调度框架将调用 Filter 函数，过滤掉不适合被调度的节点。

* Priority：调度框架将调用 Priority 函数，为每个节点计算一个优先级，优先级越高，节点越适合被调度。

* Bind：调度框架将调用 Bind 函数，将 Pod 绑定到一个节点上。

Extender是外部服务，支持Filter、Preempt、Prioritize和Bind的扩展，scheduler运行到相应阶段时，通过调用Extender注册的webhook来运行扩展的逻辑，影响调度流程中各阶段的决策结果。
分别对应四个 POST-HTTP 接口，各自的请求、响应结构定义在这里：[#kubernetes/kube-scheduler/extender/v1/types.go](https://github.com/kubernetes/kube-scheduler/blob/master/extender/v1/types.go)

在这个 HTTP 服务中，我们可以实现上述阶段中的**任意一个或多个阶段的接口**，来定制我们的调度需求。

## 代码实现
#### main.go
- 使用go-restful框架搭建一个http服务，绑定路由及端口
```golang
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
```

#### handler.go
- 处理来自 Kubernetes 调度器扩展器的 HTTP 请求，类似于Java的Controller。具体功能如下：
-  Preempt：处理抢占请求，解析请求体并调用扩展器的 Preempt 方法。
-  Filter：处理过滤请求，解析请求体并调用扩展器的 Filter 方法。
-  Prioritize：处理优先级请求，解析请求体并调用扩展器的 Prioritize 方法。
-  Bind：处理绑定请求，解析请求体并调用扩展器的 Bind 方法。

#### extender.go
- 用于创建与 Kubernetes API 服务器的连接，并初始化扩展器。具体功能如下：
- 常量和结构体： Label：定义了一个标签键，用于后续过滤节点。 Extender：包含一个 kubernetes.Interface 类型的字段 ClientSet，用于与 Kubernetes API 交互。
- NewExtender 方法：
  创建一个新的 Extender 实例。
  调用 NewClient 方法来获取 Kubernetes 客户端集（ClientSet），并将其赋值给 Extender 的 ClientSet 字段。
  如果创建客户端失败，则记录错误信息。
- NewClient 方法：
  尝试从集群内部配置或外部配置文件中加载 Kubernetes API 服务器的配置。
  优先使用环境变量 KUBECONFIG 指定的路径，如果未设置，则默认使用用户主目录下的 .kube/config 文件。
  使用 rest.InClusterConfig 尝试从集群内部获取配置，如果失败则使用 clientcmd.BuildConfigFromFlags 从外部配置文件构建配置。
  使用配置创建 Kubernetes 客户端集并返回。

#### filter.go
- 用于过滤掉不满足条件的节点，并返回满足条件的节点列表。具体功能如下：
- 接收 ExtenderArgs 参数，包含待过滤的节点和 Pod。
- 遍历所有节点，检查每个节点是否包含指定标签（Label），排除不带该标签的节点。
- 如果没有节点满足条件，则返回错误信息。
- 返回过滤后的节点列表，根据 NodeCacheCapable 的配置，使用 Nodes 或 NodeNames 字段。
```golang
func (ex *Extender) Filter(args extenderv1.ExtenderArgs) *extenderv1.ExtenderFilterResult {
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
```

#### prioritize.go
- 用于为每个节点上的 Pod 打分。具体功能如下：
- 初始化 Kubernetes 客户端
- 调用 HandlerMetricsNodeUsage 方法获取所有节点的 CPU 使用情况。
- 计算最低 CPU 使用量：
  遍历所有节点，找到 CPU 使用量最低的值 minCpuUsage，作为打分的基础。
- 计算每个节点的得分：
  再次遍历所有节点，根据每个节点的 CPU 使用量计算得分。得分规则是：CPU 使用量越低，得分越高（score = minCpuUsage - cpuUsage + 1）。

- 返回包含所有节点及其得分的 HostPriorityList。

#### 其他代码没有特殊逻辑，不做赘述
## 编译部署
192.168.112.150:5001为本地Harbor仓库

```bash
make build-image
#或者直接docker build -t 192.168.112.150:5001/library/kube-metrics-scheduler-extender:v0.0.1 . 
```
部署到集群

```bash 
kubectl apply -f deploy/manifest.yaml
```

确认服务正常运行

```bash
[root@master scheduler-dev]# kubectl -n kube-system get po|grep scheduler-extender
kube-metrics-scheduler-extender-7b7cf4f9c8-7rk9s   2/2     Running       0               116s
```

查看extender日志
```bash
kubectl logs -f --tail 500 kube-metrics-scheduler-extender-7b7cf4f9c8-jg8zs -c kube-metrics-scheduler-extender -n kube-system
```


## 测试

### 启动测试 Pod

创建一个 Deployment 并指定使用上一步中部署的 Scheduler，然后测试会调度到哪个节点上。
```bash
root@k8s-master:/home/xiaoyao/kube/xy-github/kube-metrics-scheduler-extender/deploy# kubectl apply -f deploy-test.yaml 
deployment.apps/test created
root@k8s-master:/home/xiaoyao/kube/xy-github/kube-metrics-scheduler-extender/deploy# kubectl get po
NAME                    READY   STATUS    RESTARTS   AGE
test-8666bdf76f-ktd8g   0/1     Pending   0          6s
# 查看pod Events,没有符合带有label的节点，pod将不被调度
Events:
  Type     Reason            Age    From                             Message
  ----     ------            ----   ----                             -------
  Warning  FailedScheduling  6m19s  kube-metrics-scheduler-extender  Post "http://kube-metrics-scheduler-extender.kube-system.svc:8001/scheduler/filter": EOF
  Warning  FailedScheduling  76s    kube-metrics-scheduler-extender  Post "http://kube-metrics-scheduler-extender.kube-system.svc:8001/scheduler/filter": EOF
  Warning  FailedScheduling  51s    kube-metrics-scheduler-extender  Post "http://kube-metrics-scheduler-extender.kube-system.svc:8001/scheduler/filter": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
  Warning  FailedScheduling  50s    kube-metrics-scheduler-extender  all node do not have label filter.xy.com
# 给节点添加标签，pod将能成功调度，因集群节点只有一个节点满足filter，故不走prioritize
kubectl  label node k8s-node1 filter.xy.com=true
# 若给其他节点也打上标签 会走prioritize
kubectl  label node k8s-node2 filter.xy.com=true
# 此时会发现虽然可以正常调度，但extender会报错，因为集群没有安装
E0214 06:46:25.480282       1 client.go:61] the server could not find the requested resource (get nodes.metrics.k8s.io)
E0214 06:46:25.480355       1 prioritize.go:18] Failed to get node metrics: the server could not find the requested resource (get nodes.metrics.k8s.io)
# 通过以下命令安装metrics server，具体请参考文档https://blog.csdn.net/u011837804/article/details/128487211
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
# 删除pod后重新调度，查看extender日志，发现成功获取metrics指标
I0214 07:29:54.650718       1 prioritize.go:35] Node [ k8s-master ] Cpu Usage Value: 264, Score: -196
I0214 07:29:54.650725       1 prioritize.go:35] Node [ k8s-node1 ] Cpu Usage Value: 87, Score: -19
I0214 07:29:54.650728       1 prioritize.go:35] Node [ k8s-node2 ] Cpu Usage Value: 67, Score: 1
```


