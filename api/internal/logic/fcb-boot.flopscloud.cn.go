package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

const NS_PREFIX = "yinlf"

// Request/Response Types
type Request struct {
	Name string `path:"name,options=you|me"`
}

type Response struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Info    string `json:"info,omitempty"`
}

type Storage struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	Storage      uint64 `json:"storage"`
	StorageClass string `json:"storageClass"`
}

type ListStorageResponse struct {
	Code      int64     `json:"code"`
	Message   string    `json:"message"`
	Namespace string    `json:"namespace"`
	Storages  []Storage `json:"storages"`
}

type FcbPod struct {
	Name      string `json:"name"`
	Fullname  string `json:"fullname"`
	Port      int32  `json:"port"`
	Cpu       uint64 `json:"cpu"`
	Memory    uint64 `json:"memory"`
	Storage   uint64 `json:"storage"`
	Gpu       uint64 `json:"gpu"`
	Image     string `json:"image"`
	Mount     string `json:"mount"`
	Status    string `json:"status"`
	SvcUrl    string `json:"svc_url"`
	FileUrl   string `json:"file_url"`
	Namespace string `json:"namespace"`
	Command   string `json:"command"`
}

type ListVhostsResponse struct {
	Code      int64    `json:"code"`
	Message   string   `json:"message"`
	Namespace string   `json:"namespace"`
	FcbPods   []FcbPod `json:"fcbPods"`
}

type AdminListVhostsResponseData struct {
	Namespace string   `json:"namespace"`
	FcbPods   []FcbPod `json:"fcbPods"`
}

type AdminListVhostsResponse struct {
	Code    int64                         `json:"code"`
	Message string                        `json:"message"`
	Data    []AdminListVhostsResponseData `json:"data"`
}

type DelVhostRequest struct {
	Name string `json:"name"`
}

type AdminDelVhostRequest struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type VhostRequest struct {
	Host       string `json:"host"`
	ServerName string `json:"server_name"`
	GpuType    string `json:"gpu_type"`
	FcbPod     FcbPod `json:"fcbpod"`
}

type CreateStorageRequest struct {
	Name    string `json:"name"`
	Storage uint64 `json:"storage"`
}

type DelStorageRequest struct {
	Name string `json:"name"`
}

type ExecVhostSshRequest struct {
	Namespace     string `form:"namespace"`
	PodName       string `form:"pod_name"`
	ContainerName string `form:"container_name"`
}

type ServerInfo struct {
	Name        string `json:"name"`
	TotalCPU    string `json:"total_cpu"`
	UsedCPU     string `json:"used_cpu"`
	TotalMemory string `json:"total_memory"`
	UsedMemory  string `json:"used_memory"`
	TotalGPU    string `json:"total_gpu"`
	UsedGPU     string `json:"used_gpu"`
	GPUType     string `json:"gpu_type"`
	Status      string `json:"status"`
}

type ReadVhostRequest struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

type ListVhostsRequest struct {
	Host string `json:"host"`
}

type ReadVhostsResponse struct {
	Code      int64  `json:"code"`
	Message   string `json:"message"`
	Namespace string `json:"namespace"`
	FcbPod    FcbPod `json:"fcbpod"`
}

type ListServersResponse struct {
	Code    int64        `json:"code"`
	Message string       `json:"message"`
	Servers []ServerInfo `json:"servers"`
}

// API Call Functions
func ExecVhostSsh(ctx context.Context, req *ExecVhostSshRequest) (*Response, error) {
	client := &http.Client{}
	url := fmt.Sprintf("/vhost/ssh?namespace=%s&pod_name=%s&container_name=%s",
		req.Namespace, req.PodName, req.ContainerName)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

func ListServers(ctx context.Context, JWTToken string) (*ListServersResponse, error) {
	client := &http.Client{}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", BOOTURL+"/server/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Add("Authorization", JWTToken)
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	logc.Info(ctx, string(bodyBytes))

	var result ListServersResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

func GetInfo(ctx context.Context) (*Response, error) {
	client := &http.Client{}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", "/sys/info", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

// JWT Protected Endpoints
func GetFcb(ctx context.Context, req *Request) (*Response, error) {
	client := &http.Client{}
	url := fmt.Sprintf("/from/%s", req.Name)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result Response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v %s", err, string(bodyBytes))
	}

	return &result, nil
}

var BOOTURL string

func init() {
	BOOTURL = "http://boot.xxx.cn/boot/v1"

}

func TempleteCall(fcburl string, ctx context.Context, URL string, JWTToken string, req string) (respon string, err error) {
	client := &http.Client{}

	if fcburl == "" {
		fcburl = BOOTURL
	}

	url := fcburl + URL
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(req)))
	if err != nil {
		return "", fmt.Errorf("failed to create request:%s, %v", url, err)
	}

	httpReq.Header.Add("Authorization", JWTToken)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute request:%s, %v", url, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	return string(bodyBytes), nil

}

func RestartVhost(fcburl string, ctx context.Context, JWTToken string, req *DelVhostRequest) (*Response, error) {

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	logc.Info(ctx, "RestartVhost:"+string(reqBody))

	resp, err := TempleteCall(fcburl, ctx, "/vhost/restart", JWTToken, string(reqBody))
	if err != nil {
		return nil, err
	}

	logc.Info(ctx, "RestartVhost:"+resp)

	var result Response
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v, %s", err, resp)
	}

	return &result, nil
}

type NodeLabels struct {
	Nvidia string `json:"nvidia"`
	Region string `json:"region"`
}
type NodeLabelsRequest struct {
	Name       string     `json:"name"`
	NodeLabels NodeLabels `json:"nodeLabels"`
}

func ServerLabelsUpdate(fcburl string, ctx context.Context, JWTToken string, req *NodeLabelsRequest) (*Response, error) {

	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	logc.Info(ctx, string(reqBody))

	if fcburl == "" {
		fcburl = BOOTURL
	}

	url := fcburl + "/server/labels/update"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Add("Authorization", JWTToken)

	logc.Info(ctx, JWTToken)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	logc.Info(ctx, string(bodyBytes))
	logc.Info(ctx, url)

	var result Response
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		logx.Errorw("failed to decode response", logx.Field("url", url), logx.Field("error", err), logx.Field("response", string(bodyBytes)))

		return nil, fmt.Errorf("failed to decode response: %v", err)

	}

	return &result, nil
}

func CreateVhost(fcburl string, ctx context.Context, JWTToken string, req *VhostRequest) (*Response, error) {

	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	logc.Info(ctx, string(reqBody))

	if fcburl == "" {
		fcburl = BOOTURL
	}

	url := fcburl + "/vhost/create"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Add("Authorization", JWTToken)

	logc.Info(ctx, JWTToken)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// logc.Info(ctx, resp.Body.String())
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	logc.Info(ctx, string(bodyBytes))
	logc.Info(ctx, url)

	var result Response
	if err := json.Unmarshal(bodyBytes, &result); err != nil {

		return nil, fmt.Errorf("failed to decode response: %v", err)

	}

	return &result, nil
}

func UpdateVhost(fcburl string, ctx context.Context, JWTToken string, req *VhostRequest) (*Response, error) {
	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	logx.Infof("UpdateVhost:%s", string(reqBody))

	if fcburl == "" {
		fcburl = BOOTURL
	}

	url := fcburl + "/vhost/update"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", JWTToken)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result Response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	logx.Infof("UpdateVhost:%s", string(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v %s", err, string(bodyBytes))
	}

	return &result, nil
}

func ListVhosts(fcburl string, ctx context.Context, JWTToken string, req *ListVhostsRequest) (*ListVhostsResponse, error) {
	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", BOOTURL+"/vhost/list", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", JWTToken)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result ListVhostsResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v %s", err, string(bodyBytes))
	}

	return &result, nil
}

func ReadVhost(fcburl string, ctx context.Context, JWTToken string, req *ReadVhostRequest) (*ReadVhostsResponse, error) {
	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", BOOTURL+"/vhost/read", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", JWTToken)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result ReadVhostsResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v %s", err, string(bodyBytes))
	}

	return &result, nil
}

func DeleteVhost(fcburl string, ctx context.Context, JWTToken string, req *DelVhostRequest) (*Response, error) {

	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	logc.Info(ctx, string(reqBody))

	httpReq, err := http.NewRequest("POST", BOOTURL+"/vhost/del", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Add("Authorization", JWTToken)

	logc.Info(ctx, JWTToken)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// logc.Info(ctx, resp.Body.String())
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	logc.Info(ctx, string(bodyBytes))

	var result Response
	if err := json.Unmarshal(bodyBytes, &result); err != nil {

		return nil, fmt.Errorf("failed to decode response: %v", err)

	}

	return &result, nil

}

func CreateStorage(fcburl string, ctx context.Context, JWTToken string, req *CreateStorageRequest) (*Response, error) {
	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", BOOTURL+"/storage/create", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Add("Authorization", JWTToken)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var result Response
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v %s", err, string(bodyBytes))
	}

	return &result, nil
}

func DeleteStorage(fcburl string, ctx context.Context, req *DelStorageRequest, JWTToken string) (*Response, error) {
	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", BOOTURL+"/storage/del", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Add("Authorization", JWTToken)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var result Response
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v %s", err, string(bodyBytes))
	}

	return &result, nil
}

func ListStorage(fcburl string, ctx context.Context) (*ListStorageResponse, error) {
	client := &http.Client{}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", BOOTURL+"/storage/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result ListStorageResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v %s", err, string(bodyBytes))
	}

	return &result, nil
}

func AdminListVhosts(fcburl string, ctx context.Context, JWTToken string, req *ListVhostsRequest) (*ListVhostsResponse, error) {
	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	if fcburl == "" {
		fcburl = BOOTURL
	}

	url := fcburl + "/admin/vhost/list"

	// fmt.Println(url, string(reqBody))
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Add("Authorization", JWTToken)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result ListVhostsResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	// logc.Info(ctx, string(bodyBytes))
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v %s", err, string(bodyBytes))
	}

	return &result, nil
}

func AdminDeleteVhost(fcburl string, ctx context.Context, JWTToken string, req *AdminDelVhostRequest) (*Response, error) {
	client := &http.Client{}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	if fcburl == "" {
		fcburl = BOOTURL
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fcburl+"/admin/vhost/del", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Add("Authorization", JWTToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result Response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v %s", err, string(bodyBytes))
	}

	return &result, nil
}
