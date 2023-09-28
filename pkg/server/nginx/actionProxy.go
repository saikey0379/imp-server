package nginx

import (
	"github.com/saikey0379/imp-server/pkg/model"
	"path"

	"github.com/saikey0379/imp-server/pkg/known"
	"github.com/saikey0379/imp-server/pkg/utils"
)

type ActionProxy struct {
	PCluster model.Cluster
}

// ConfJobIPExecParam 作业执行参数
type ConfJobIPExecParam struct {
	ExecHosts   []ExecHost `json:"execHosts" validate:"required"`
	ExecPort    uint       `json:"hostPort"`
	ExecParam   ExecParam  `json:"execParam" validate:"required"`
	Provider    string     `json:"provider" validate:"required"` // provider可以为salt|puppet|openssh
	Callback    string     `json:"callback"`
	ExecuteID   string     `json:"executeId" validate:"required"`
	JobRecordID string     `json:"jobRecordId"`
	DestFile    string     `json:"destFile"`
}

// ExecHost
type ExecHost struct {
	HostIP    string `json:"hostIp"`
	EntityID  string `json:"entityId"`
	HostID    string `json:"hostId"`
	IdcName   string `json:"idcName"`
	OsType    string `json:"osType,omitempty"`
	Encoding  string `json:"encoding,omitempty"` // 系统默认的编码，如果为空，则默认以utf-8值进行处理
	ProxyID   string `json:"proxyId"`
	SrcIp     string `json:"srcIp"`
	RateLimit uint   `json:"rateLimit"`
}

// ExecParam 执行参数
type ExecParam struct {
	// 模块名称，支持 script：脚本执行, salt.state：状态应用, file：文件下发
	Pattern string `json:"pattern" validate:"required"`

	// 依据模块名称进行解释
	// pattern为script时，script为脚本内容
	// pattern为salt.state时，script为salt的state内容
	// pattern为file时，script为文件内容或url数组列表
	Script string `json:"script"`
	// 依据pattern进行解释
	// pattern为script时，scriptType为shell, bat, python
	// pattern为file时，scriptType为url或者text
	ScriptType string                 `json:"scriptType" validate:"required"`
	Params     map[string]interface{} `json:"params"`
	RunAs      string                 `json:"runas,omitempty"`
	Password   string                 `json:"password"`
	Timeout    int                    `json:"timeout" validate:"required"`
	Env        map[string]string      `json:"env"`
	ExtendData interface{}            `json:"extendData"`
	// 是否实时输出，像巡检任务、定时任务则不需要实时输出
	RealTimeOutput bool `json:"realTimeOutput"`
}

// Task Exec Result CallBack
type JobCallbackParam struct {
	JobRecordID   string             `json:"jobRecordId"`
	ExecuteID     string             `json:"executeId"`
	ExecuteStatus string             `json:"executeStatus"`
	ResultStatus  string             `json:"resultStatus"`
	HostResult    HostResultCallback `json:"hostResult"`
}

type HostResultCallback struct {
	EntityID string `json:"entityId"`
	HostIP   string `json:"hostIp"`
	IdcName  string `json:"idcName"`
	Status   string `json:"status"`
	Result   string `json:"message"`
	Time     string `json:"time"`
}

type Act2Resp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Content string `json:"content"`
}

func (pa *ActionProxy) ConfTest(host string) (output string, err error) {
	execscript := pa.PCluster.ExecTest

	var sc = &utils.SSHClient{
		Address:    host,
		Port:       pa.PCluster.SSHPort,
		User:       pa.PCluster.SSHUser,
		PrivateKey: path.Join(known.RootProxy, ".ssh/"+pa.PCluster.SSHKey),
	}
	err = sc.CreateClient()
	if err != nil {
		return "", err
	}
	defer sc.Client.Close()

	output, err = sc.RunShell(execscript)
	if err != nil {
		return output, err
	}
	return output, nil

}

func (pa *ActionProxy) ConfLoad(host string) (output string, err error) {
	execscript := pa.PCluster.ExecLoad

	var sc = &utils.SSHClient{
		Address:    host,
		Port:       pa.PCluster.SSHPort,
		User:       pa.PCluster.SSHUser,
		PrivateKey: path.Join(known.RootProxy, ".ssh/"+pa.PCluster.SSHKey),
	}

	err = sc.CreateClient()
	if err != nil {
		return "", err
	}
	defer sc.Client.Close()

	output, err = sc.RunShell(execscript)
	if err != nil {
		return output, err
	}
	return output, nil
}
