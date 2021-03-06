package core

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	kl "github.com/accuknox/KubeArmor/KubeArmor/common"
	kg "github.com/accuknox/KubeArmor/KubeArmor/log"
	tp "github.com/accuknox/KubeArmor/KubeArmor/types"
)

// ================= //
// == K8s Handler == //
// ================= //

// K8s Handler
var K8s *K8sHandler

// init Function
func init() {
	K8s = NewK8sHandler()
}

// K8sHandler Structure
type K8sHandler struct {
	K8sClient   *kubernetes.Clientset
	HTTPClient  *http.Client
	WatchClient *http.Client

	K8sToken string
	K8sHost  string
	K8sPort  string
}

// NewK8sHandler Function
func NewK8sHandler() *K8sHandler {
	kh := &K8sHandler{}

	if val, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST"); ok {
		kh.K8sHost = val
	} else {
		kh.K8sHost = "127.0.0.1"
	}

	if val, ok := os.LookupEnv("KUBERNETES_PORT_443_TCP_PORT"); ok {
		kh.K8sPort = val
	} else {
		kh.K8sPort = "8001" // kube-proxy
	}

	kh.HTTPClient = &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	kh.WatchClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return kh
}

// ================ //
// == K8s Client == //
// ================ //

// InitK8sClient Function
func (kh *K8sHandler) InitK8sClient() bool {
	if !kl.IsK8sEnv() { // not Kubernetes
		return false
	}

	if kh.K8sClient == nil {
		if kl.IsInK8sCluster() {
			return kh.InitInclusterAPIClient()
		}
		return kh.InitLocalAPIClient()
	}

	return true
}

// InitLocalAPIClient Function
func (kh *K8sHandler) InitLocalAPIClient() bool {
	var kubeconfig *string
	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		kg.Err(err.Error())
		return false
	}

	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		kg.Err(err.Error())
		return false
	}
	kh.K8sClient = client

	return true
}

// InitInclusterAPIClient Function
func (kh *K8sHandler) InitInclusterAPIClient() bool {
	read, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		kg.Err(err.Error())
		return false
	}
	kh.K8sToken = string(read)

	// create the configuration by token
	kubeConfig := &rest.Config{
		Host:        "https://" + kh.K8sHost + ":" + kh.K8sPort,
		BearerToken: kh.K8sToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		kg.Err(err.Error())
		return false
	}
	kh.K8sClient = client

	return true
}

// ============== //
// == API Call == //
// ============== //

// DoRequest Function
func (kh *K8sHandler) DoRequest(cmd string, data interface{}, path string) ([]byte, error) {
	URL := ""

	if kl.IsInK8sCluster() {
		URL = "https://" + kh.K8sHost + ":" + kh.K8sPort
	} else {
		URL = "http://" + kh.K8sHost + ":" + kh.K8sPort
	}

	pbytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(cmd, URL+path, bytes.NewBuffer(pbytes))
	if err != nil {
		return nil, err
	}

	if kl.IsInK8sCluster() {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", kh.K8sToken))
	}

	resp, err := kh.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()
	return resBody, nil
}

// ========== //
// == Node == //
// ========== //

// GetContainerRuntime Function
func (kh *K8sHandler) GetContainerRuntime() string {
	if !kl.IsK8sEnv() { // not Kubernetes
		return ""
	}

	// get a host name
	hostName := kl.GetHostName()

	// get a node from k8s api client
	node, err := kh.K8sClient.CoreV1().Nodes().Get(context.Background(), hostName, metav1.GetOptions{})
	if err != nil {
		return "Unknown"
	}

	return node.Status.NodeInfo.ContainerRuntimeVersion
}

// ========== //
// == Pods == //
// ========== //

// GetK8sPod Function
func (kh *K8sHandler) GetK8sPod(K8sPods []tp.K8sPod, namespaceName, containerGroupName string) tp.K8sPod {
	for _, pod := range K8sPods {
		if pod.Metadata["namespaceName"] == namespaceName && pod.Metadata["podName"] == containerGroupName {
			return pod
		}
	}

	return tp.K8sPod{}
}

// GetK8sPods Function
func (kh *K8sHandler) GetK8sPods() []tp.K8sPod {
	if !kl.IsK8sEnv() { // not Kubernetes
		return []tp.K8sPod{}
	}

	// get pods from k8s api client
	pods, err := kh.K8sClient.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []tp.K8sPod{}
	}

	newPods := []tp.K8sPod{}

	for _, pod := range pods.Items {
		metadata := pod.ObjectMeta

		k8spod := tp.K8sPod{}

		k8spod.Metadata = map[string]string{}
		k8spod.Metadata["podName"] = metadata.Name
		k8spod.Metadata["namespaceName"] = metadata.Namespace
		k8spod.Metadata["generation"] = strconv.FormatInt(metadata.Generation, 10)

		kl.Clone(metadata.Annotations, &k8spod.Annotations)
		kl.Clone(metadata.Labels, &k8spod.Labels)

		newPods = append(newPods, k8spod)
	}

	return newPods
}

// WatchK8sPods Function
func (kh *K8sHandler) WatchK8sPods() *http.Response {
	if !kl.IsK8sEnv() { // not Kubernetes
		return nil
	}

	if kl.IsInK8sCluster() { // kube-apiserver
		URL := "https://" + kh.K8sHost + ":" + kh.K8sPort + "/api/v1/pods?watch=true"

		req, err := http.NewRequest("GET", URL, nil)
		if err != nil {
			return nil
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", kh.K8sToken))

		resp, err := kh.WatchClient.Do(req)
		if err != nil {
			return nil
		}

		return resp
	}

	// kube-proxy (local)
	URL := "http://" + kh.K8sHost + ":" + kh.K8sPort + "/api/v1/pods?watch=true"

	if resp, err := http.Get(URL); err == nil {
		return resp
	}

	return nil
}

// ====================== //
// == Custom Resources == //
// ====================== //

// CheckCustomResourceDefinition Function
func (kh *K8sHandler) CheckCustomResourceDefinition(resourceName string) bool {
	if !kl.IsK8sEnv() { // not Kubernetes
		return false
	}

	exist := false
	apiGroup := metav1.APIGroup{}

	// check APIGroup
	if resBody, errOut := kh.DoRequest("GET", nil, "/apis"); errOut == nil {
		res := metav1.APIGroupList{}
		if errIn := json.Unmarshal(resBody, &res); errIn == nil {
			for _, group := range res.Groups {
				if group.Name == "security.accuknox.com" {
					exist = true
					apiGroup = group
					break
				}
			}
		}
	}

	// check APIResource
	if exist {
		if resBody, errOut := kh.DoRequest("GET", nil, "/apis/"+apiGroup.PreferredVersion.GroupVersion); errOut == nil {
			res := metav1.APIResourceList{}
			if errIn := json.Unmarshal(resBody, &res); errIn == nil {
				for _, resource := range res.APIResources {
					if resource.Name == resourceName {
						return true
					}
				}
			}
		}
	}

	return false
}

// GetK8sSecurityPolicies Function
func (kh *K8sHandler) GetK8sSecurityPolicies() []tp.SecurityPolicy {
	if !kl.IsK8sEnv() { // not Kubernetes
		return []tp.SecurityPolicy{}
	}

	if resBody, errOut := kh.DoRequest("GET", nil, "/apis/security.accuknox.com/v1/kubearmorpolicies"); errOut == nil {
		res := tp.K8sKubeArmorPolicies{}
		if errIn := json.Unmarshal(resBody, &res); errIn == nil {
			securityPolicies := []tp.SecurityPolicy{}

			for _, item := range res.Items {
				securityPolicy := tp.SecurityPolicy{}

				securityPolicy.Metadata = map[string]string{}
				securityPolicy.Metadata["namespaceName"] = item.Metadata.Namespace
				securityPolicy.Metadata["policyName"] = item.Metadata.Name
				securityPolicy.Metadata["generation"] = strconv.FormatInt(item.Metadata.Generation, 10)

				kl.Clone(item.Spec, &securityPolicy.Spec)

				securityPolicy.Spec.Selector.Identities = append(securityPolicy.Spec.Selector.Identities, "namespaceName="+item.Metadata.Namespace)

				for k, v := range securityPolicy.Spec.Selector.MatchNames {
					if kl.ContainsElement([]string{"containerGroupName", "containerName", "hostName", "imageName"}, k) {
						securityPolicy.Spec.Selector.Identities = append(securityPolicy.Spec.Selector.Identities, k+"="+v)
					}
				}

				for k, v := range securityPolicy.Spec.Selector.MatchLabels {
					securityPolicy.Spec.Selector.Identities = append(securityPolicy.Spec.Selector.Identities, k+"="+v)
				}

				kg.Printf("Fetched a new Security Policy (%s/%s)", securityPolicy.Metadata["namespaceName"], securityPolicy.Metadata["policyName"])

				securityPolicies = append(securityPolicies, securityPolicy)
			}

			return securityPolicies
		}
	}

	return []tp.SecurityPolicy{}
}

// WatchK8sSecurityPolicies Function
func (kh *K8sHandler) WatchK8sSecurityPolicies() *http.Response {
	if !kl.IsK8sEnv() { // not Kubernetes
		return nil
	}

	if kl.IsInK8sCluster() {
		URL := "https://" + kh.K8sHost + ":" + kh.K8sPort + "/apis/security.accuknox.com/v1/kubearmorpolicies?watch=true"

		req, err := http.NewRequest("GET", URL, nil)
		if err != nil {
			return nil
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", kh.K8sToken))

		resp, err := kh.WatchClient.Do(req)
		if err != nil {
			return nil
		}

		return resp
	}

	// kube-proxy (local)
	URL := "http://" + kh.K8sHost + ":" + kh.K8sPort + "/apis/security.accuknox.com/v1/kubearmorpolicies?watch=true"

	if resp, err := http.Get(URL); err == nil {
		return resp
	}

	return nil
}
