package route

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
	"github.com/saikey0379/imp-server/pkg/utils"
)

const (
	TaskStatusStart   = "waiting"
	TaskStatusFailure = "failure"
)

type TaskListPageReq struct {
	ID          uint   `json:"id"`
	TaskType    string `json:"TaskType"`
	FileType    string `json:"FileType"`
	AccessToken string `json:"AccessToken"`
	Keyword     string `json:"keyword"`
	Limit       uint
	Offset      uint
}

func getTaskConditions(req TaskListPageReq) string {
	var where []string

	if req.Keyword = strings.TrimSpace(req.Keyword); req.Keyword != "" {
		where = append(where, fmt.Sprintf("( task_list.id like %s or task_list.name like %s or task_list.description like %s or task_list.manager like %s or task_list.match_hosts like %s )", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'"))
	}

	if req.TaskType != "" && req.TaskType != "all" {
		where = append(where, fmt.Sprintf("( task_list.task_type = %s )", "'"+req.TaskType+"'"))
	}

	if len(where) > 0 {
		return " where " + strings.Join(where, " and ")
	} else {
		return ""
	}
}

func GetFileSelect(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info TaskListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	type File struct {
		ID   uint
		Name string
	}

	mods, err := repo.GetFileListByFileType(info.FileType)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var file File
	var files []File
	for _, i := range mods {
		file.ID = i.ID
		file.Name = i.Name
		files = append(files, file)
	}

	result := make(map[string]interface{})
	result["list"] = files

	//总条数
	count := len(files)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetTaskList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info TaskListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	type Task struct {
		ID          uint
		Name        string
		Manager     string
		Description string
		TaskType    string
		FileType    string
		Status      string
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}

	mods, err := repo.GetTaskListWithPage(info.Limit, info.Offset, getTaskConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var task Task
	var tasks []Task
	for _, i := range mods {
		task.ID = i.ID
		task.Name = i.Name
		task.Manager = i.Manager
		task.Description = i.Description
		task.TaskType = i.TaskType
		task.Status = i.Status
		task.CreatedAt = utils.ISOTime(i.CreatedAt)
		task.UpdatedAt = utils.ISOTime(i.UpdatedAt)
		task.FileType = i.FileType

		tasks = append(tasks, task)
	}

	result := make(map[string]interface{})
	result["list"] = tasks
	result["TaskType"] = "all"

	if info.TaskType != "" {
		result["TaskType"] = info.TaskType
	}

	//总条数
	count, err := repo.CountTask(getTaskConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetTaskById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		ID          uint
		AccessToken string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	info.AccessToken = strings.TrimSpace(info.AccessToken)
	_, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}

	mod, err := repo.GetTaskById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	type Cron struct {
		Minute string
		Hour   string
		DMonth string
		Month  string
		DWeek  string
	}

	type Task struct {
		ID          uint
		Name        string
		Manager     string
		Description string
		MatchHosts  string
		TaskType    string
		ExecTime    string
		PolicyCron  Cron
		FileId      uint
		FileName    string
		FileType    string
		FileMod     string
		Parameter   string
		DestPath    string
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}

	var task Task

	task.ID = info.ID
	task.Name = mod.Name
	task.Manager = mod.Manager
	task.Description = mod.Description
	task.MatchHosts = mod.MatchHosts
	task.TaskType = mod.TaskType
	task.FileId = mod.FileId
	task.FileType = mod.FileType
	task.FileMod = mod.FileMod
	task.Parameter = mod.Parameter
	task.DestPath = mod.DestPath
	task.CreatedAt = utils.ISOTime(mod.CreatedAt)
	task.UpdatedAt = utils.ISOTime(mod.UpdatedAt)
	switch mod.TaskType {
	case "cron":
		for i, v := range strings.Split(mod.TaskPolicy, ",") {
			switch i {
			case 0:
				task.PolicyCron.Minute = v
			case 1:
				task.PolicyCron.Hour = v
			case 2:
				task.PolicyCron.DMonth = v
			case 3:
				task.PolicyCron.Month = v
			case 4:
				task.PolicyCron.DWeek = v
			}
		}
	default:
		task.ExecTime = mod.TaskPolicy
	}

	mod_file, err := repo.GetFileById(task.FileId)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	task.FileName = mod_file.Name

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": task})
}

// 添加
func AddTask(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	type Cron struct {
		Minute string
		Hour   string
		DMonth string
		Month  string
		DWeek  string
	}

	type Task struct {
		Name        string
		Manager     string
		Description string
		MatchHosts  string
		TaskType    string
		ExecTime    string
		PolicyCron  Cron
		FileId      uint
		FileType    string
		FileMod     string
		Parameter   string
		DestPath    string
		AccessToken string
	}
	var info Task

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}

	info.AccessToken = strings.TrimSpace(info.AccessToken)
	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}

	if info.Name == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请填写任务名称!"})
		return
	}

	var mod model.Task
	mod.ID = uint(utils.GenSnowFlakeID())
	mod.Name = strings.TrimSpace(info.Name)
	mod.Manager = strings.TrimSpace(info.Manager)
	mod.Description = strings.TrimSpace(info.Description)
	mod.MatchHosts = strings.TrimSpace(info.MatchHosts)

	var matchhosts []string
	for _, k := range strings.Split(info.MatchHosts, "\n") {
		strs := utils.StrSplitAny(k, `|,， 、/;；:：\`)
		matchhosts = append(matchhosts, strs...)
	}
	mod.MatchHosts = strings.Join(matchhosts, ",")

	mod.TaskType = info.TaskType
	mod.FileId = info.FileId
	mod.FileType = info.FileType

	if utils.IsFileMode(info.FileMod) {
		mod.FileMod = info.FileMod
	} else {
		mod.FileMod = "0750"
	}

	switch mod.FileType {
	case "config", "document":
		mod.DestPath = strings.TrimSpace(info.DestPath)
	case "script", "execution":
		mod.Parameter = utils.Delete_extra_space(strings.TrimSpace(info.Parameter))
	}
	mod.UpdatedAt = time.Now()

	var parseId uint64
	var errParse error
	switch info.TaskType {
	case "cron":
		mod.TaskPolicy = info.PolicyCron.Minute + "," + info.PolicyCron.Hour + "," + info.PolicyCron.DMonth + "," + info.PolicyCron.Month + "," + info.PolicyCron.DWeek
	case "fixed":
		etime, _ := time.Parse("2006-01-02T15:04", info.ExecTime)
		mod.TaskPolicy = etime.Format("2006-01-02 15:04")
		parseId, errParse = strconv.ParseUint(etime.Format("200601021504"), 10, 64)
	case "immed":
		m, _ := time.ParseDuration("1m")
		mod.TaskPolicy = time.Now().Add(m).Format("2006-01-02 15:04")
		parseId, errParse = strconv.ParseUint(time.Now().Add(m).Format("200601021504"), 10, 64)
	case "trigger":
		mod.TaskPolicy = time.Now().Format("2006-01-02 15:04")
		parseId, errParse = strconv.ParseUint(time.Now().Format("200601021504"), 10, 64)
	}

	if errParse != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "BatchID Parse Error"})
		return
	}
	batchId := uint(parseId)

	var snList []string
	var hosts []model.TaskHost
	for _, v := range matchhosts {
		modDevice, err := repo.GetDeviceByHostname(v)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "主机名" + v + "查询失败"})
			return
		}

		var host model.TaskHost
		host.TaskID = mod.ID
		host.Hostname = v
		host.Sn = modDevice.Sn
		snList = append(snList, host.Sn)

		host.Ip = modDevice.Ip
		hosts = append(hosts, host)
	}

	if utils.HasDuplicate(snList) {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "任务节点重复"})
		return
	}

	modtask, errAdd := repo.AddTask(mod)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}

	result := make(map[string]interface{})
	var success string
	var failure string
	var message string

	var jsonReq struct {
		ID          uint
		BatchId     uint
		FileId      uint
		FileMod     string
		Parameter   string
		DestPath    string
		FileName    string
		FileType    string
		FileLink    string
		Interpreter string
		Md5         string
	}
	var byte_req []byte
	if mod.TaskType == "trigger" {
		mod_file, err := repo.GetFileById(mod.FileId)
		if err != nil {
			message = err.Error()
		}
		jsonReq.ID = mod.ID
		jsonReq.BatchId = batchId
		jsonReq.FileId = mod.FileId
		jsonReq.FileMod = mod.FileMod
		jsonReq.Parameter = mod.Parameter
		jsonReq.DestPath = mod.DestPath
		jsonReq.FileName = mod_file.Name
		jsonReq.FileType = mod_file.FileType
		jsonReq.FileLink = mod_file.FileLink
		jsonReq.Interpreter = mod_file.Interpreter
		jsonReq.Md5 = mod_file.Md5
		byte_req, err = json.Marshal(jsonReq)
		if err != nil {
			if len(message) == 0 {
				message = err.Error()
			} else {
				message = message + ";[" + err.Error() + "]"
			}
		}
	}

	var modResult model.TaskResult
	modResult.TaskId = mod.ID
	modResult.BatchId = batchId
	modResult.Status = TaskStatusStart
	modResult.StartTime = time.Now()
	modResult.EndTime = time.Now()
	modResult.CreatedAt = time.Now()

	for _, v := range hosts {
		if mod.TaskType != "cron" {
			modResult.Hostname = v.Hostname
			_, errAdd := repo.AddTaskResult(modResult)
			if errAdd != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
				return
			}
		}
		var err error
		if mod.TaskType == "trigger" {
			err = TriggerTask(v.Ip, byte_req)
		} else {
			err = repo.AddTaskHost(v)
		}
		if err != nil {
			if failure == "" {
				failure = v.Hostname
			} else {
				failure = failure + "," + v.Hostname
			}
			if len(message) == 0 {
				message = v.Hostname + ":[" + err.Error() + "]"
			} else {
				message = message + ";" + v.Hostname + ":[" + err.Error() + "]"
			}
			continue
		}
		if success == "" {
			success = v.Hostname
		} else {
			success = success + "," + v.Hostname
		}
	}

	result["success"] = success

	if failure == "" {
		message = ""
	} else {
		result["failure"] = failure
	}

	contenttask, err := json.Marshal(modtask)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}

	var journal model.Journal
	journal.Title = modtask.Name
	journal.Operation = "add"
	journal.Resource = "task"
	journal.Content = "[add Task:" + string(contenttask) + "]"
	journal.User = user.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message, "Content": result})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}

func DeleteTaskById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	logger, ok := middleware.LoggerFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		ID          uint
		AccessToken string
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	info.AccessToken = strings.TrimSpace(info.AccessToken)
	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}

	modtask, err := repo.GetTaskById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	redis, okRedis := middleware.RedisFromContext(ctx)
	if okRedis {
		taskHosts, err := repo.GetTaskHostListByTaskId(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		for _, taskHost := range taskHosts {
			key := fmt.Sprintf("IMP_TASK_LIST_%s", taskHost.Sn)
			_, err := redis.Del(key)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
	}

	err = repo.DeleteTaskHostByTaskId(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	_, err = repo.DeleteTaskById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	contenttask, err := json.Marshal(modtask)
	if err != nil {
		logger.Errorf(fmt.Sprintf("ERROR: TASK Umarshal [%s]", err.Error()))
	}

	var journal model.Journal
	journal.Title = modtask.Name
	journal.Operation = "delete"
	journal.Resource = "task"
	journal.Content = "[delete Task:" + string(contenttask) + "]"
	journal.User = user.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		logger.Errorf(fmt.Sprintf("ERROR: AddJournal [%s]", err.Error()))
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}

func UpdateTaskById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	logger, ok := middleware.LoggerFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	type Cron struct {
		Minute string
		Hour   string
		DMonth string
		Month  string
		DWeek  string
	}

	type Task struct {
		ID          uint
		Name        string
		Manager     string
		Description string
		MatchHosts  string
		TaskType    string
		ExecTime    string
		PolicyCron  Cron
		FileId      uint
		FileType    string
		FileMod     string
		Parameter   string
		DestPath    string
		AccessToken string
	}

	var info Task

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}

	info.AccessToken = strings.TrimSpace(info.AccessToken)
	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}

	if info.Name == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "任务名不可为空!"})
		return
	}

	mod, err := repo.GetTaskById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var taskPolicy string
	var parseId uint64
	var errParse error

	switch info.TaskType {
	case "cron":
		taskPolicy = info.PolicyCron.Minute + "," + info.PolicyCron.Hour + "," + info.PolicyCron.DMonth + "," + info.PolicyCron.Month + "," + info.PolicyCron.DWeek
	case "fixed":
		etime, _ := time.Parse("2006-01-02T15:04", info.ExecTime)
		mod.TaskPolicy = etime.Format("2006-01-02 15:04")
		parseId, errParse = strconv.ParseUint(etime.Format("200601021504"), 10, 64)
	case "immed":
		m, _ := time.ParseDuration("1m")
		mod.TaskPolicy = time.Now().Add(m).Format("2006-01-02 15:04")
		parseId, errParse = strconv.ParseUint(time.Now().Add(m).Format("200601021504"), 10, 64)
	case "trigger":
		mod.TaskPolicy = time.Now().Format("2006-01-02 15:04")
		parseId, errParse = strconv.ParseUint(time.Now().Format("200601021504"), 10, 64)
	}

	if errParse != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "BatchID Parse Error"})
		return
	}
	batchId := uint(parseId)

	var mod_update model.Task
	mod_update.ID = info.ID
	mod_update.Name = strings.TrimSpace(info.Name)
	mod_update.Manager = strings.TrimSpace(info.Manager)
	mod_update.Description = strings.TrimSpace(info.Description)

	var matchhosts []string
	for _, k := range strings.Split(info.MatchHosts, "\n") {
		strs := utils.StrSplitAny(k, `|,， 、/;；:：\`)
		matchhosts = append(matchhosts, strs...)
	}
	mod_update.MatchHosts = strings.Join(matchhosts, ",")

	mod_update.TaskType = info.TaskType
	mod_update.TaskPolicy = taskPolicy
	mod_update.FileId = info.FileId
	mod_update.FileType = info.FileType

	switch mod_update.FileType {
	case "config", "document":
		mod_update.DestPath = strings.TrimSpace(info.DestPath)
	case "script", "execution":
		mod_update.Parameter = utils.Delete_extra_space(strings.TrimSpace(info.Parameter))
	}

	if utils.IsFileMode(info.FileMod) {
		mod_update.FileMod = info.FileMod
	} else {
		mod_update.FileMod = "0750"
	}

	mod_update.UpdatedAt = time.Now()
	mod_update.Status = TaskStatusStart

	updatebool := false
	var content string

	if mod_update.Name != mod.Name {
		updatebool = true
		content = content + "[update Name:\"" + mod.Name + "\" to \"" + mod_update.Name + "\"]"
	}
	if mod_update.Manager != mod.Manager {
		updatebool = true
		content = content + "[update Manager:\"" + mod.Manager + "\" to \"" + mod_update.Manager + "\"]"
	}
	if mod_update.Description != mod.Description {
		updatebool = true
		content = content + "[update Description:\"" + mod.Description + "\" to \"" + mod_update.Description + "\"]"
	}
	if mod_update.MatchHosts != mod.MatchHosts {
		updatebool = true
		content = content + "[update MatchHosts:\"" + mod.MatchHosts + "\" to \"" + mod_update.MatchHosts + "\"]"
	}
	if mod_update.TaskType != mod.TaskType {
		updatebool = true
		content = content + "[update TaskType:\"" + mod.TaskType + "\" to \"" + mod_update.TaskType + "\"]"
	}
	if mod_update.TaskPolicy != mod.TaskPolicy {
		updatebool = true
		content = content + "[update TaskPolicy:\"" + mod.TaskPolicy + "\" to \"" + mod_update.TaskPolicy + "\"]"
	}
	if mod_update.FileType != mod.FileType {
		updatebool = true
		content = content + "[update FileType:\"" + mod.FileType + "\" to \"" + mod_update.FileType + "\"]"
	}
	if mod_update.FileId != mod.FileId {
		updatebool = true
		content = content + "[update FileId:\"" + strconv.Itoa(int(mod.FileId)) + "\" to \"" + strconv.Itoa(int(mod_update.FileId)) + "\"]"
	}
	if mod_update.FileMod != mod.FileMod {
		updatebool = true
		content = content + "[update FileMod:\"" + mod.FileMod + "\" to \"" + mod_update.FileMod + "\"]"
	}

	if mod_update.FileType != mod.FileType {
		switch info.FileType {
		case "script", "execution":
			if mod_update.Parameter != mod.Parameter {
				updatebool = true
				content = content + "[update Parameter:\"" + mod.Parameter + "\" to \"" + info.Parameter + "\"]"
			}
		case "config", "document":
			if mod_update.DestPath != mod.DestPath {
				updatebool = true
				content = content + "[update DestPath:\"" + mod.DestPath + "\" to \"" + info.DestPath + "\"]"
			}
		}
	}

	if !updatebool {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "数据未更改"})
		return
	}

	var snList []string
	var hosts []model.TaskHost
	for _, v := range matchhosts {
		mod_d, err := repo.GetDeviceByHostname(v)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "主机名" + v + "查询失败"})
			return
		}
		var host model.TaskHost

		host.TaskID = mod_update.ID
		host.Hostname = v
		host.Sn = mod_d.Sn
		snList = append(snList, host.Sn)

		host.Ip = mod_d.Ip
		hosts = append(hosts, host)
	}

	if utils.HasDuplicate(snList) {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "任务节点重复"})
		return
	}

	_, err = repo.UpdateTaskById(info.ID, mod_update.Name, mod_update.Manager, mod_update.Description, mod_update.MatchHosts, mod_update.TaskType, mod_update.TaskPolicy, mod_update.FileId, mod_update.FileType, mod_update.FileMod, mod_update.Parameter, mod_update.DestPath, mod_update.Status)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	taskHosts, err := repo.GetTaskHostListByTaskId(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	err = repo.DeleteTaskHostByTaskId(mod_update.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	redis, okRedis := middleware.RedisFromContext(ctx)
	if okRedis {
		for _, taskHost := range taskHosts {
			key := fmt.Sprintf("IMP_TASK_LIST_%s", taskHost.Sn)
			_, err := redis.Del(key)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
	}

	var jsonReq struct {
		ID          uint
		BatchId     uint
		FileId      uint
		FileMod     string
		Parameter   string
		DestPath    string
		FileName    string
		FileType    string
		FileLink    string
		Interpreter string
		Md5         string
	}
	if mod_update.TaskType == "trigger" {
		mod_file, err := repo.GetFileById(mod_update.FileId)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		jsonReq.ID = mod_update.ID
		jsonReq.BatchId = batchId
		jsonReq.FileId = mod_update.FileId
		jsonReq.FileMod = mod_update.FileMod
		jsonReq.Parameter = mod_update.Parameter
		jsonReq.DestPath = mod_update.DestPath
		jsonReq.FileName = mod_file.Name
		jsonReq.FileType = mod_file.FileType
		jsonReq.FileLink = mod_file.FileLink
		jsonReq.Interpreter = mod_file.Interpreter
		jsonReq.Md5 = mod_file.Md5
		byte_req, err := json.Marshal(jsonReq)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		for _, v := range hosts {
			err = TriggerTask(v.Ip, byte_req)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
	} else {
		var modResult model.TaskResult
		modResult.TaskId = mod_update.ID
		modResult.BatchId = batchId
		modResult.Status = "waiting"
		modResult.StartTime = time.Now()
		modResult.EndTime = time.Now()
		modResult.CreatedAt = time.Now()

		for _, v := range hosts {
			if mod_update.TaskType != "cron" {
				modResult.Hostname = v.Hostname
				_, errAdd := repo.AddTaskResult(modResult)
				if errAdd != nil {
					logger.Errorf("ERROR: AddTaskResult[%s]", errAdd.Error())
				}
			}

			err = repo.AddTaskHost(v)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
	}

	var journal model.Journal
	journal.Title = mod_update.Name
	journal.Operation = "update"
	journal.Resource = "task"
	journal.Content = "[update Task:" + content + "]"
	journal.User = user.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		logger.Errorf(fmt.Sprintf("ERROR: AddJournal[%s]", err.Error()))
	}
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}

func TriggerTask(host string, request []byte) (err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":10079")
	if err != nil {
		fmt.Println("Fatal error: %s", err.Error())
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Fatal error: %s", err.Error())
		return err
	}

	defer conn.Close()

	_, err = conn.Write(request)
	if err != nil {
		fmt.Println("Fatal error: %s", err.Error())
		return err
	}

	var message bytes.Buffer

	buffer := make([]byte, 1024)
	recvLen, err := conn.Read(buffer)
	if err != nil {
		return err
	}
	message.Write(buffer[:recvLen])

	if message.String() == "success" {
		return nil
	} else {
		return errors.New(message.String())
	}
}

func GetTaskFullListBySn(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	logger, ok := middleware.LoggerFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		Sn string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	var key = fmt.Sprintf("IMP_TASK_LIST_%s", info.Sn)
	redis, okRedis := middleware.RedisFromContext(ctx)
	if ok {
		v, err := redis.Get(key)
		if err == nil && v != "" {
			w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": v})
			return
		}
	}

	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	type FileInfo struct {
		FileId      uint
		FileName    string
		FileType    string
		FileLink    string
		FileMod     string
		Interpreter string
		Parameter   string
		DestPath    string
		Md5         string
	}

	type TaskInfo struct {
		ID         uint
		BatchId    uint
		TaskType   string
		TaskPolicy string
		ReportURL  string
		File       FileInfo
	}

	//总条数
	count, err := repo.CountTaskBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if count == 0 {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "Null"})
		return
	}

	mods, err := repo.GetTaskFullListBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var taskList []TaskInfo
	for _, v := range mods {
		switch v.TaskType {
		case "fixed", "immed", "cron":
			taskInfo := TaskInfo{
				ID:         v.ID,
				TaskType:   v.TaskType,
				TaskPolicy: v.TaskPolicy,
				File: FileInfo{
					FileId:      v.FileId,
					FileName:    v.FileName,
					FileType:    v.FileType,
					FileLink:    v.FileLink,
					FileMod:     v.FileMod,
					Interpreter: v.Interpreter,
					Parameter:   v.Parameter,
					DestPath:    v.DestPath,
					Md5:         v.Md5,
				},
			}

			if v.TaskType == "cron" {
				taskList = append(taskList, taskInfo)
			} else {
				stamp, _ := time.ParseInLocation("2006-01-02 15:04", v.TaskPolicy, time.Local)
				if stamp.Unix()+60 > time.Now().Unix() {
					taskList = append(taskList, taskInfo)
				}
			}
		}
	}

	if okRedis {
		redisValue, err := json.Marshal(taskList)
		if err != nil {
			logger.Errorf(fmt.Sprintf("ERROR: JSON Marshal[%s]", err.Error()))
		} else {
			_, err := redis.SetEx(key, string(redisValue), 3600)
			if err != nil {
				logger.Errorf(fmt.Sprintf("ERROR: REDIS SetEX[%s]", err.Error()))
			}
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": taskList})
}
