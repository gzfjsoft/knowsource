package utils

import (
	"context"
	"flag"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

// GetPodsByNamespace 根据指定的命名空间和筛选条件获取所有的 Pod 列表。
// 该函数使用 Kubernetes 客户端 clientset 来查询特定命名空间下的所有 Pod，并根据提供的筛选条件进行过滤。
// 函数返回获取到的 Pod 列表以及在执行过程中遇到的任何错误。
// 参数:
//   - clientset: Kubernetes 客户端集合，用于与 Kubernetes API 交互。
//   - namespace: 字符串类型，指定查询的命名空间名称。
//   - listOptions: 筛选条件，用于过滤 Pod 列表。
//
// 返回:
//   - pods: 包含查询结果的 Pod 列表。
//   - err: 在执行过程中遇到的任何错误。
func GetPodsByNamespace(clientset *kubernetes.Clientset, namespace string, listOptions metav1.ListOptions) (*apiv1.PodList, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, err
	}
	fmt.Printf("There are %d pods in the %s\n", len(pods.Items), namespace)
	return pods, nil
}

// CreateDeployment 创建一个Kubernetes部署(deployment)。
// 参数:
// - clientset: Kubernetes客户端集合，用于与Kubernetes API服务器通信。
// - name: 部署的名称。
// - replicas: 副本数量，指明维持运行的副本数。
// - labels: 标签映射，用于选择Pod。
// - containerList: 容器列表，部署中每个Pod包含的容器。
// 返回值:
// - result: 创建的部署对象。
// - success: 操作是否成功，成功时返回true，失败时返回false。
func CreateDeployment(clientset *kubernetes.Clientset, name string, replicas int32, labels map[string]string, containerList []apiv1.Container) (result *appsv1.Deployment, success bool) {
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: containerList,
				},
			},
		},
	}

	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, false
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return result, true // 成功创建后返回true
}

// UpdateDeployment 更新Kubernetes部署中的副本数和镜像名称。
// 参数:
// - clientset: Kubernetes客户端集，用于与API服务器通信。
// - name: 需要更新的部署的名称。
// - replicas: 更新后的部署的副本数量。
// - imageName: 更新后容器使用的镜像名称。
// 返回值:
// - success: 如果更新成功返回true，否则返回false。
// 该函数通过retry.RetryOnConflict函数来处理并发更新冲突，确保部署的最新版本被应用。
func UpdateDeployment(clientset *kubernetes.Clientset, name string, replicas int32, imageName string) (success bool) {
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, getErr := deploymentsClient.Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		result.Spec.Replicas = ptr.To(replicas)
		result.Spec.Template.Spec.Containers[0].Image = imageName

		_, updateErr := deploymentsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
		return updateErr
	})
	// 如果重试后仍有错误，设置success为false
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
		return false
	}

	fmt.Println("Updated deployment...")
	return true
}

// ListDeployments 列出指定命名空间下的所有 Deployment
// ListDeployments 列出指定命名空间中的部署。
// 该函数通过clientset与Kubernetes集群进行通信，获取并打印指定命名空间内的所有部署及其副本数量。
// 参数:
//
//	clientset: Kubernetes clientset，用于与API服务器通信。
//	namespaces: 字符串，表示要列出其中部署的命名空间。
//
// 返回:
//
//	list: 包含部署列表的 metav1.DeploymentList 类型对象。
//	ok: 布尔值，表示操作是否成功。
func ListDeployments(clientset *kubernetes.Clientset, namespaces string) (list *appsv1.DeploymentList, ok bool) {
	deploymentsClient := clientset.AppsV1().Deployments(namespaces)
	fmt.Printf("Listing deployments in namespace %q:\n", namespaces)
	list, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, false
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}
	return list, true
}

// DeleteDeployment 删除 Deployment
// DeleteDeployment 删除指定的 Deployment。
// 该函数通过 clientset 初始化一个部署客户端，指定删除策略，并执行删除操作。
// 参数:
//
//		clientset *kubernetes.Clientset: 用于与 Kubernetes API 交互的客户端集。
//		name string: 要删除的 Deployment 的名称。
//	    namespace string: Deployment 所在的命名空间。使用以下常量 ,如：apiv1.NamespaceDefault：
//	    - apiv1.NamespaceDefault: 表示对象位于默认命名空间，默认情况下客户端未指定时会应用此命名空间。
//	    - apiv1.NamespaceAll: 在上下文中指定此参数表示要跨所有命名空间列出或过滤资源。
//	    - apiv1.NamespaceNodeLease: 用于放置节点租约对象（用于节点心跳）的命名空间。
//
// 返回值:
//
//	bool: 如果删除成功，返回 true；否则返回 false。
func DeleteDeployment(clientset *kubernetes.Clientset, name string, namespace string) bool {
	// 为默认命名空间初始化一个部署客户端。
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	// 指定删除策略，使用前景删除策略，意味着相关资源也将被删除。
	deletePolicy := metav1.DeletePropagationForeground

	// 尝试删除指定的 Deployment。如果发生错误，返回 false。
	if err := deploymentsClient.Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return false
	}

	// 打印删除成功的消息。
	fmt.Println("Deleted deployment.")

	// 如果删除成功，返回 true。
	return true
}

// createKubernetesClient 创建一个Kubernetes客户端
// 参数:
// - kubeconfigPath: kubeconfig文件的绝对路径
// 返回值:
// - *kubernetes.Clientset: Kubernetes客户端集合
// - error: 错误信息，如果客户端创建成功，则为nil
func createKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	kubeconfig := flag.String("kubeconfig", kubeconfigPath, "(optional) absolute path to the kubeconfig file")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from flags: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return clientset, nil
}
