package route

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
	"github.com/saikey0379/imp-server/pkg/utils"
)

type TaskResultListPageReq struct {
	ID          uint   `json:"id"`
	TaskId      uint   `json:"TaskId"`
	BatchId     uint   `json:"BatchId"`
	Hostname    string `json:"Hostname"`
	AccessToken string `json:"AccessToken"`
	Keyword     string `json:"keyword"`
	Limit       uint
	Offset      uint
}

func getTaskResultConditions(req TaskResultListPageReq) string {
	var where []string

	if req.Keyword = strings.TrimSpace(req.Keyword); req.Keyword != "" {
		where = append(where, fmt.Sprintf("( task_result.id like %s or task_result.result like %s or task_result.hostname like %s or task_result.status like %s )", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'"))
	}

	if req.TaskId > 0 {
		where = append(where, fmt.Sprintf("( task_result.task_id = %s%d%s )", "'", req.TaskId, "'"))
	}

	if req.BatchId > 0 {
		where = append(where, fmt.Sprintf("( task_result.batch_id = %s%d%s )", "'", req.BatchId, "'"))
	}

	if req.Hostname != "" {
		where = append(where, fmt.Sprintf("( task_result.hostname = %s )", "'"+req.Hostname+"'"))
	}

	if len(where) > 0 {
		return " where " + strings.Join(where, " and ")
	} else {
		return ""
	}
}

// 上报任务结果
func ReportTaskResult(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "application/json; charset=utf-8")
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		TaskId    uint
		BatchId   uint
		Hostname  string
		FileSync  string
		Result    string
		Status    string
		StartTime time.Time
		EndTime   time.Time
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}
	var mod model.TaskResult
	mod.TaskId = info.TaskId
	mod.BatchId = info.BatchId
	mod.Hostname = info.Hostname
	mod.FileSync = info.FileSync
	mod.Result = info.Result
	mod.Status = info.Status
	mod.StartTime = info.StartTime
	mod.EndTime = info.EndTime
	mod.CreatedAt = time.Now()

	id, err := repo.GetTaskResultId(fmt.Sprintf("where task_id=%s and batch_id=%s and hostname=%s order by created_at desc limit 1", "'"+strconv.Itoa(int(mod.TaskId))+"'", "'"+strconv.Itoa(int(mod.BatchId))+"'", "'"+mod.Hostname+"'"))
	if id > 1 || err == nil {
		_, errUpdate := repo.UpdateTaskResultById(id, mod.FileSync, mod.Result, mod.Status, mod.StartTime, mod.EndTime)
		if errUpdate != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
			return
		}
	} else {
		_, errAdd := repo.AddTaskResult(mod)
		if errAdd != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
			return
		}
	}

	modTask, err := repo.GetTaskById(mod.TaskId)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	} else {
		if modTask.Status == TaskStatusStart || mod.Status == TaskStatusFailure {
			err = repo.UpdateTaskStatusById(mod.TaskId, mod.Status)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}

func GetTaskResultQueryTermsList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info TaskResultListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	batchlist, err := repo.GetTaskResultBatchList(getTaskResultConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	hostlist, err := repo.GetTaskResultHostList(getTaskResultConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	result := make(map[string]interface{})
	result["BatchList"] = batchlist
	result["HostList"] = hostlist

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetTaskResultList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info TaskResultListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	type TaskResult struct {
		ID        uint
		BatchId   uint
		Hostname  string
		FileSync  string
		Result    string
		Status    string
		StartTime utils.ISOTime
		EndTime   utils.ISOTime
		CreatedAt utils.ISOTime
	}

	mods, err := repo.GetTaskResultListWithPage(info.Limit, info.Offset, getTaskResultConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var taskResult TaskResult
	var taskResults []TaskResult
	for _, i := range mods {
		taskResult.ID = i.ID
		taskResult.BatchId = i.BatchId
		taskResult.Hostname = i.Hostname
		taskResult.FileSync = i.FileSync
		taskResult.Result = i.Result
		taskResult.Status = i.Status
		taskResult.StartTime = utils.ISOTime(i.StartTime)
		taskResult.EndTime = utils.ISOTime(i.EndTime)
		taskResult.CreatedAt = utils.ISOTime(i.CreatedAt)
		taskResults = append(taskResults, taskResult)
	}

	result := make(map[string]interface{})
	result["list"] = taskResults

	//总条数
	count, err := repo.CountTaskResult(getTaskResultConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetTaskResultByResultId(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info TaskResultListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	mod, err := repo.GetTaskResultByResultId(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": mod})
}

func ClearTaskResult(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info TaskResultListPageReq

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

	err := repo.ClearTaskResult(getTaskResultConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": ""})
}
