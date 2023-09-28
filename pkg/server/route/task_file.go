package route

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/saikey0379/imp-server/pkg/known"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
	"github.com/saikey0379/imp-server/pkg/utils"
)

type FileListPageReq struct {
	ID          uint   `json:"id"`
	AccessToken string `json:"AccessToken"`
	Keyword     string `json:"keyword"`
	FileType    string `json:"FileType"`
	Limit       uint
	Offset      uint
}

func getFileConditions(req FileListPageReq) string {
	var where []string

	if req.Keyword = strings.TrimSpace(req.Keyword); req.Keyword != "" {
		where = append(where, fmt.Sprintf("( task_file.id like %s or task_file.name like %s or task_file.description like %s or task_file.md5 like %s or task_file.content like %s or task_file.manager like %s or task_file.interpreter like %s )", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'"))
	}

	if req.FileType != "" && req.FileType != "all" {
		where = append(where, fmt.Sprintf("( task_file.file_type = %s )", "'"+req.FileType+"'"))
	}

	if len(where) > 0 {
		return " where " + strings.Join(where, " and ")
	} else {
		return ""
	}
}

func GetFileList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info FileListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	type File struct {
		ID          uint
		Name        string
		Manager     string
		Description string
		FileType    string
		Status      string
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}

	mods, err := repo.GetFileListWithPage(info.Limit, info.Offset, getFileConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var file File
	var files []File
	for _, i := range mods {
		file.ID = i.ID
		file.Name = i.Name
		file.Manager = i.Manager
		file.Description = i.Description
		file.FileType = i.FileType
		file.CreatedAt = utils.ISOTime(i.CreatedAt)
		file.UpdatedAt = utils.ISOTime(i.UpdatedAt)
		count, err := repo.CountTaskByFileId(file.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		if count > 0 {
			file.Status = "online"
		} else {
			file.Status = "offline"
		}

		files = append(files, file)
	}

	result := make(map[string]interface{})
	result["list"] = files
	result["FileType"] = "all"

	if info.FileType != "" {
		result["FileType"] = info.FileType
	}

	//总条数
	count, err := repo.CountFile(getFileConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetFileById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	mod, err := repo.GetFileById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	type Task struct {
		ID   uint
		Name string
	}

	type File struct {
		ID          uint
		Name        string
		Manager     string
		Description string
		FileType    string
		FileLink    string
		Interpreter string
		Content     string
		Md5         string
		Tasks       []Task
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}

	var file File
	file.ID = info.ID
	file.Name = mod.Name
	file.Manager = mod.Manager
	file.Description = mod.Description
	file.FileType = mod.FileType
	file.FileLink = mod.FileLink
	file.Interpreter = mod.Interpreter
	file.Content = mod.Content
	file.Md5 = mod.Md5
	file.CreatedAt = utils.ISOTime(mod.CreatedAt)
	file.UpdatedAt = utils.ISOTime(mod.UpdatedAt)

	var tasks []Task
	mods, err := repo.GetTaskListByFileId(file.ID)
	for _, v := range mods {
		var task Task
		task.ID = v.ID
		task.Name = v.Name
		tasks = append(tasks, task)
	}
	file.Tasks = tasks

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": file})
}

// 添加
func AddFile(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	type File struct {
		Name        string
		Manager     string
		Description string
		FileType    string
		Interpreter string
		Content     string
		AccessToken string
	}
	var info File

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
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请填写服务组名称!"})
		return
	}

	var mod model.File
	mod.ID = uint(utils.GenSnowFlakeID())
	mod.Name = strings.TrimSpace(info.Name)
	mod.Manager = strings.TrimSpace(info.Manager)
	mod.Description = strings.TrimSpace(info.Description)
	mod.FileType = info.FileType

	if info.FileType == "script" || info.FileType == "config" {
		mod.Interpreter = info.Interpreter
		mod.Content = info.Content
		byte_content := []byte(info.Content)
		md5sum_content := md5.Sum(byte_content)
		mod.Md5 = hex.EncodeToString(md5sum_content[:])
		mod.FileLink = "/api/task/file/getContent?name=" + mod.Name
	} else {
		rootDir := known.RootDir
		filename := path.Join(rootDir, info.Name)
		f, err := os.Open(filename)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		defer f.Close()

		md5hash := md5.New()
		if _, err := io.Copy(md5hash, f); err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		mod.Md5 = hex.EncodeToString(md5hash.Sum(nil))
		mod.FileLink = "/www/upload/" + mod.Name
	}

	modfile, errAdd := repo.AddFile(mod)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}

	var message string
	contentfile, err := json.Marshal(modfile)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}

	var journal model.Journal
	journal.Title = modfile.Name
	journal.Operation = "add"
	journal.Resource = "file"
	journal.Content = "[add File:" + string(contentfile) + "]"
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

func DeleteFileById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	count, _ := repo.CountTaskByFileId(info.ID)
	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该文件已被任务调用,请删除相关任务配置"})
		return
	}

	mod, err := repo.GetFileById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if mod.FileType == "execution" || mod.FileType == "document" {
		rootDir := known.RootDir
		filename := path.Join(rootDir, mod.Name)
		err = os.Remove(filename)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}

	_, err = repo.DeleteFileById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var message string
	contentfile, err := json.Marshal(mod)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "delete"
	journal.Resource = "file"
	journal.Content = "[delete File:" + string(contentfile) + "]"
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
func GetFileContentByName(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	var info struct {
		Name string
		Type string
	}

	info.Name = r.FormValue("name")
	info.Type = r.FormValue("type")
	info.Name = strings.TrimSpace(info.Name)
	info.Type = strings.TrimSpace(info.Type)

	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		if info.Type == "raw" {
			w.Write([]byte(""))
		} else {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误", "Content": ""})
		}
		return
	}

	if info.Type == "" {
		info.Type = "raw"
	}

	if info.Name == "" {
		if info.Type == "raw" {
			w.Write([]byte(""))
		} else {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN参数不能为空"})
		}
		return
	}

	mod, err := repo.GetFileByName(info.Name)
	if err != nil {
		if info.Type == "raw" {
			w.Write([]byte(""))
		} else {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": ""})
		}

		return
	}
	mod.Content = strings.Replace(mod.Content, "\r\n", "\n", -1)
	if info.Type == "raw" {
		w.Header().Add("Content-type", "text/html; charset=utf-8")
		w.Write([]byte(mod.Content))
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "成功获取system信息", "Content": mod})
	}
}

func UpdateFileById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	type File struct {
		ID          uint
		Name        string
		Manager     string
		Description string
		FileType    string
		Interpreter string
		Content     string
		AccessToken string
	}
	var info File

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

	info.Name = strings.TrimSpace(info.Name)
	info.Manager = strings.TrimSpace(info.Manager)
	info.Description = strings.TrimSpace(info.Description)

	if info.Name == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请填写文件名!"})
		return
	}

	mod, err := repo.GetFileById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var md5_sum string
	if info.FileType == "script" || info.FileType == "config" {
		byte_content := []byte(info.Content)
		md5sum_content := md5.Sum(byte_content)
		md5_sum = hex.EncodeToString(md5sum_content[:])
		mod.FileLink = "/api/task/file/getContent?name=" + info.Name
	} else {
		rootDir := known.RootDir
		filename := path.Join(rootDir, info.Name)
		f, err := os.Open(filename)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		defer f.Close()

		md5hash := md5.New()
		if _, err := io.Copy(md5hash, f); err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		md5_sum = hex.EncodeToString(md5hash.Sum(nil))
		mod.FileLink = "/www/upload/" + info.Name
	}

	updatebool := false
	var content string
	var message string

	if info.Name != mod.Name {
		updatebool = true
		content = content + "[update Name:\"" + mod.Name + "\" to \"" + info.Name + "\"]"
	}
	if info.Manager != mod.Manager {
		updatebool = true
		content = content + "[update Manager:\"" + mod.Manager + "\" to \"" + info.Manager + "\"]"
	}
	if info.Description != mod.Description {
		updatebool = true
		content = content + "[update Description:\"" + mod.Description + "\" to \"" + info.Description + "\"]"
	}
	if info.FileType != mod.FileType {
		updatebool = true
		content = content + "[update FileType:\"" + mod.FileType + "\" to \"" + info.FileType + "\"]"
	}
	if md5_sum != mod.Md5 {
		updatebool = true
		content = content + "[update Content:\"" + mod.Content + "\" to \"" + info.Content + "\"]"
	}

	if info.FileType == "script" {
		if info.Interpreter != mod.Interpreter {
			updatebool = true
			content = content + "[update Interpreter:\"" + mod.Interpreter + "\" to \"" + info.Interpreter + "\"]"
		}
	}

	if !updatebool {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "数据未更改"})
		return
	}
	_, err = repo.UpdateFileById(info.ID, info.Name, info.Manager, info.Description, info.FileType, md5_sum, mod.FileLink, info.Interpreter, info.Content)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "update"
	journal.Resource = "file"
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

func UploadFile(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "text/html; charset=utf-8")
	r.ParseForm()

	file, handle, err := r.FormFile("file")
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	destfile := r.FormValue("filename")

	if destfile == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "文件名不能为空"})
		return
	}

	rootDir := known.RootDir
	if !utils.FileExist(rootDir) {
		err := os.MkdirAll(rootDir, 0777)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}

	filename := path.Join(rootDir, handle.Filename)
	dfilename := path.Join(rootDir, destfile)

	result := make(map[string]interface{})
	result["result"] = dfilename

	if utils.FileExist(filename) {
		os.Remove(filename)
	}

	if utils.FileExist(dfilename) {
		os.Remove(dfilename)
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	io.Copy(f, file)

	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	defer f.Close()
	defer file.Close()

	os.Rename(filename, dfilename)

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
	return
}
