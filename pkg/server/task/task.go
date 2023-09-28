package task

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/saikey0379/imp-server/pkg/utils"
)

var (
	apiTaskUpdate = "/api/task/update"
)

var (
	taskAddr *TaskAddr
	once     sync.Once
)

func init() {
	once.Do(func() {
		taskAddr = &TaskAddr{}
	})
}

type TaskAddr struct {
	address string
	port    int
	mux     sync.Mutex
}

func GetTaskAddr() *TaskAddr {
	return taskAddr
}

func (taskAddr *TaskAddr) SetTaskAddr(address string, port int) {
	taskAddr.mux.Lock()
	defer taskAddr.mux.Unlock()
	taskAddr.address = address
	taskAddr.port = port
}

type Task struct {
	ID          int64  `json:"ID"`
	Name        string `json:"Name"`
	MatchHosts  string `json:"MatchHosts"`
	TaskType    string `json:"TaskType"`
	FileId      int64  `json:"FileId"`
	FileType    string `json:"FileType"`
	AccessToken string `json:"AccessToken"`
	Parameter   string `json:"Parameter"`
}

func (task Task) UpdateTask(username string) (resp string, err error) {
	var url string
	url = fmt.Sprintf("http://%s:%d%s", taskAddr.address, taskAddr.port, apiTaskUpdate)

	task.Parameter = fmt.Sprintf("%s %s", task.Parameter, username)

	reqBody, err := json.Marshal(task)
	if err != nil {
		return resp, err
	}

	respBody, err := utils.PostRestApi(url, reqBody)
	return string(respBody), err
}
