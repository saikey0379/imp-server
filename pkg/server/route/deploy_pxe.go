package route

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
)

// OsConfig
func DeleteOsConfigById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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
	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}

	mod, err := repo.GetOsConfigById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	_, err = repo.DeleteOsConfigById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var message string
	contentosconfig, err := json.Marshal(mod)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "delete"
	journal.Resource = "pxe"
	journal.Content = "[delete OsConfig:" + string(contentosconfig) + "]"
	journal.User = user.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}

func UpdateOsConfigById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		ID          uint
		Name        string
		Pxe         string
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

	info.Name = strings.TrimSpace(info.Name)
	info.Pxe = strings.TrimSpace(info.Pxe)
	info.Pxe = strings.Replace(info.Pxe, "\r\n", "\n", -1)

	if info.Name == "" || info.Pxe == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请将信息填写完整!"})
		return
	}

	count, err := repo.CountOsConfigByNameAndId(info.Name, info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该操作系统版本已存在!"})
		return
	}

	mod, err := repo.GetOsConfigById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	updatebool := false
	var content string
	var message string

	if info.Name != mod.Name {
		updatebool = true
		content = content + "[update Name:\"" + mod.Name + "\" to \"" + info.Name + "\"]"
	}

	if info.Pxe != mod.Pxe {
		updatebool = true
		content = content + "[update Pxe:\"" + mod.Pxe + "\" to \"" + info.Pxe + "\"]"
	}

	if !updatebool {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "数据未更改"})
		return
	}

	_, err = repo.UpdateOsConfigById(info.ID, info.Name, info.Pxe)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "update"
	journal.Resource = "pxe"
	journal.Content = content
	journal.User = user.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}

func GetOsConfigById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		ID uint
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	pxe, err := repo.GetOsConfigById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": pxe})
}

func GetOsConfigList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Limit  uint
		Offset uint
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	pxes, err := repo.GetOsConfigListWithPage(info.Limit, info.Offset)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result := make(map[string]interface{})
	result["list"] = pxes

	//总条数
	count, err := repo.CountOsConfig()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

// 添加OS操作系统
func AddOsConfig(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Name        string
		Pxe         string
		AccessToken string
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}

	info.Name = strings.TrimSpace(info.Name)
	info.Pxe = strings.TrimSpace(info.Pxe)
	info.Pxe = strings.Replace(info.Pxe, "\r\n", "\n", -1)

	info.AccessToken = strings.TrimSpace(info.AccessToken)
	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}

	if info.Name == "" || info.Pxe == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请将信息填写完整!"})
		return
	}

	count, err := repo.CountOsConfigByName(info.Name)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该操作系统版本已存在!"})
		return
	}

	mod, errAdd := repo.AddOsConfig(info.Name, info.Pxe)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}

	var message string
	contentosconfig, err := json.Marshal(mod)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "add"
	journal.Resource = "pxe"
	journal.Content = "[add OsConfig:" + string(contentosconfig) + "]"
	journal.User = user.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}
