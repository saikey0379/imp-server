package route

import (
	"encoding/json"
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

type UpstreamListPageReq struct {
	ID          uint        `json:"id"`
	ClusterId   uint        `json:"ClusterId"`
	ClusterIds  []ClusterId `json:"ClusterIds"`
	DomainId    uint        `json:"DomainId"`
	DomainName  string      `json:"DomainName"`
	RouteId     uint        `json:"RouteId"`
	AccessToken string      `json:"AccessToken"`
	Keyword     string      `json:"keyword"`
	AccessType  string      `json:"AccessType"`
	Limit       uint
	Offset      uint
}

func getUpstreamConditions(req UpstreamListPageReq) string {
	var where []string
	if req.ID > 0 {
		where = append(where, fmt.Sprintf("proxy_upstream.id = %d", req.ID))
	}

	if req.ClusterId > 0 {
		where = append(where, fmt.Sprintf("proxy_upstream.cluster_ids like %s%d%s", "'%", req.ClusterId, "%'"))
	}

	for _, v := range req.ClusterIds {
		if v.ID > 0 {
			where = append(where, fmt.Sprintf("proxy_upstream.cluster_ids like %s%d%s", "'%", v.ID, "%'"))
		}
	}

	if req.Keyword = strings.TrimSpace(req.Keyword); req.Keyword != "" {
		where = append(where, fmt.Sprintf("( proxy_upstream.name like %s or proxy_upstream.description like %s )", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'"))
	}

	if len(where) > 0 {
		return " where " + strings.Join(where, " and ")
	} else {
		return ""
	}
}

func GetUpstreamSelectByClusterIds(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info UpstreamListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	type Upstream struct {
		ID   int
		Name string
	}

	mods, err := repo.GetUpstreamListWithPage(0, 0, getUpstreamConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var upstream Upstream
	var upstreams []Upstream
	for _, i := range mods {
		upstream.ID = i.ID
		upstream.Name = i.Name
		upstreams = append(upstreams, upstream)
	}

	result := make(map[string]interface{})
	result["list"] = upstreams
	result["ClusterId"] = info.ClusterId

	//总条数
	count := len(upstreams)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetUpstreamList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info UpstreamListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	type ClusterId struct {
		ID int
	}

	type Upstream struct {
		ID          int
		Name        string
		Manager     string
		ClusterIds  []ClusterId
		Description string
		Used        int
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}

	mods, err := repo.GetUpstreamListWithPage(info.Limit, info.Offset, getUpstreamConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var upstream Upstream
	var upstreams []Upstream
	for _, i := range mods {
		upstream.ID = i.ID
		upstream.Name = i.Name
		upstream.Manager = i.Manager
		upstream.Description = i.Description
		upstream.Used = i.Used
		upstream.CreatedAt = utils.ISOTime(i.CreatedAt)
		upstream.UpdatedAt = utils.ISOTime(i.UpdatedAt)

		var clusterid ClusterId
		var clusterids []ClusterId
		for _, i := range strings.Split(i.ClusterIds, ",") {
			clusterid.ID, err = strconv.Atoi(i)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			clusterids = append(clusterids, clusterid)
		}
		upstream.ClusterIds = clusterids

		upstreams = append(upstreams, upstream)
	}

	result := make(map[string]interface{})
	result["list"] = upstreams
	result["ClusterId"] = info.ClusterId

	//总条数
	count, err := repo.CountUpstream(getUpstreamConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetUpstreamBackendsById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Id uint
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	mod, err := repo.GetUpstreamBackendsById(info.Id)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	type Server struct {
		Ip           string
		Port         string
		BakCustomize string
	}

	var server Server
	var servers []Server

	for _, i := range strings.Split(mod.Backends, ",") {
		for index, j := range strings.Split(i, ":") {
			if index == 0 {
				server.Ip = j
			} else if index == 1 {
				server.Port = j
			} else if index == 2 {
				server.BakCustomize = j
			}
		}
		servers = append(servers, server)
	}

	result := make(map[string]interface{})
	result["list"] = servers

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "C ntent": result})
}

func GetUpstreamById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		ID          int
		AccessToken string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	mod, err := repo.GetUpstreamById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	type ClusterId struct {
		ID int
	}

	type Backend struct {
		Ip           string
		Port         string
		BakCustomize string
	}

	type Upstream struct {
		ID          int
		Name        string
		Manager     string
		Backends    []Backend
		ClusterIds  []ClusterId
		Customize   string
		Description string
		Routes      []model.RouteUs
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}

	var upstream Upstream

	upstream.ID = info.ID
	upstream.Name = mod.Name
	upstream.Manager = mod.Manager
	upstream.Customize = mod.Customize
	upstream.Description = mod.Description
	upstream.CreatedAt = utils.ISOTime(mod.CreatedAt)
	upstream.UpdatedAt = utils.ISOTime(mod.UpdatedAt)

	var clusterid ClusterId
	var clusterids []ClusterId
	for _, i := range strings.Split(mod.ClusterIds, ",") {
		clusterid.ID, err = strconv.Atoi(i)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		clusterids = append(clusterids, clusterid)
	}
	upstream.ClusterIds = clusterids

	var backend Backend
	var backends []Backend

	for _, i := range strings.Split(mod.Backends, ",") {
		for index, j := range strings.Split(i, ":") {
			if index == 0 {
				backend.Ip = j
			} else if index == 1 {
				backend.Port = j
			} else if index == 2 {
				backend.BakCustomize = j
			}
		}
		backends = append(backends, backend)
	}
	upstream.Backends = backends

	routeUs, err := repo.GetRouteListByUpstreamId(info.ID)
	if len(routeUs) > 0 && err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	upstream.Routes = routeUs

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": upstream})
}

// 添加
func AddUpstream(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	type ClusterId struct {
		ID int
	}

	type Backend struct {
		Ip           string
		Port         string
		BakCustomize string
	}

	type Upstream struct {
		Name        string
		Manager     string
		Backends    []Backend
		ClusterIds  []ClusterId
		Customize   string
		Description string
		AccessToken string
	}
	var info Upstream

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

	var mod model.Upstream
	mod.ID = int(time.Now().Unix())
	mod.Name = strings.TrimSpace(info.Name)
	mod.Manager = strings.TrimSpace(info.Manager)
	mod.Customize = info.Customize
	mod.Description = strings.TrimSpace(info.Description)
	mod.UpdatedAt = time.Now()

	var clusterids string
	for i, v := range info.ClusterIds {
		clusterid := strconv.Itoa(v.ID)
		if i == 0 {
			clusterids = clusterid
		} else {
			clusterids = clusterids + "," + clusterid
		}
	}
	mod.ClusterIds = clusterids

	var backends string
	for k, i := range info.Backends {
		if !utils.IsIpAddress(i.Ip) {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "IP地址异常,请确认"})
			return
		}
		port, err := strconv.Atoi(i.Port)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		} else if port <= 0 || port >= 65535 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Http端口数值异常"})
			return
		}

		if k == 0 {
			backends += i.Ip + ":" + i.Port + ":" + i.BakCustomize
		} else {
			backends += "," + i.Ip + ":" + i.Port + ":" + i.BakCustomize
		}
	}

	mod.Backends = backends
	modupstream, errAdd := repo.AddUpstream(mod)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}

	var message string
	var content string
	contentupstream, err := json.Marshal(modupstream)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}
	content = content + "[add Upstream:" + string(contentupstream) + "]"

	for _, v := range info.ClusterIds {
		err := ConfCreateByClusterId(repo, v.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	var journal model.Journal
	journal.Title = modupstream.Name
	journal.Operation = "add"
	journal.Resource = "upstream"
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

func DeleteUpstreamById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		ID          int
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

	count, err := repo.CountRouteByUpstreamId(info.ID)
	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该服务组已被调用,请删除相关域名配置"})
		return
	} else if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	mod, err := repo.GetUpstreamById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	_, err = repo.DeleteUpstreamById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	for _, v := range strings.Split(mod.ClusterIds, ",") {
		id, err := strconv.Atoi(v)
		err = ConfCreateByClusterId(repo, id)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}

	var message string
	contentupstream, err := json.Marshal(mod)
	if err != nil {
		message = fmt.Sprintf("[Umarshal failed:%s]", err)
	}
	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "delete"
	journal.Resource = "upstream"
	journal.Content = "[delete Upstream:" + string(contentupstream) + "]"
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

func UpdateUpstreamById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	type ClusterId struct {
		ID int
	}

	type Backend struct {
		Ip           string
		Port         string
		BakCustomize string
	}

	type Upstream struct {
		ID          int
		Name        string
		Manager     string
		Backends    []Backend
		ClusterIds  []ClusterId
		Customize   string
		Description string
		AccessToken string
	}
	var info Upstream

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

	mod, err := repo.GetUpstreamById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var clusterids string
	for i, v := range info.ClusterIds {
		clusterid := strconv.Itoa(v.ID)
		if i == 0 {
			clusterids = clusterid
		} else {
			clusterids = clusterids + "," + clusterid
		}
	}

	var backends string
	for k, i := range info.Backends {
		if !utils.IsIpAddress(i.Ip) {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "IP地址异常,请确认"})
			return
		}

		port, err := strconv.Atoi(i.Port)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		} else if port <= 0 || port >= 65535 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Http端口数值异常"})
			return
		}

		if k == 0 {
			backends += i.Ip + ":" + i.Port + ":" + i.BakCustomize
		} else {
			backends += "," + i.Ip + ":" + i.Port + ":" + i.BakCustomize
		}
	}
	var mod_update model.Upstream
	mod_update.Name = strings.TrimSpace(info.Name)
	mod_update.Manager = strings.TrimSpace(info.Manager)
	mod_update.Customize = strings.TrimSpace(info.Customize)
	mod_update.Description = strings.TrimSpace(info.Description)
	mod_update.ClusterIds = clusterids
	mod_update.Backends = backends
	mod_update.UpdatedAt = time.Now()

	var updatebool = false
	var message string
	var content string

	if mod_update.Name != mod.Name {
		updatebool = true
		content = content + "[update Name:\"" + mod.Name + "\" to \"" + mod_update.Name + "\"]"
	}
	if mod_update.Manager != mod.Manager {
		updatebool = true
		content = content + "[update Manager:\"" + mod.Manager + "\" to \"" + mod_update.Manager + "\"]"
	}
	if mod_update.Customize != mod.Customize {
		updatebool = true
		content = content + "[update Customize:\"" + mod.Customize + "\" to \"" + mod_update.Customize + "\"]"
	}
	if mod_update.Description != mod.Description {
		updatebool = true
		content = content + "[update Description:\"" + mod.Description + "\" to \"" + mod_update.Description + "\"]"
	}
	if mod_update.ClusterIds != mod.ClusterIds {
		updatebool = true
		content = content + "[update ClusterIds:\"" + mod.ClusterIds + "\" to \"" + mod_update.ClusterIds + "\"]"
	}
	if mod_update.Backends != mod.Backends {
		updatebool = true
		content = content + "[update Backends:\"" + mod.Backends + "\" to \"" + mod_update.Backends + "\"]"
	}

	if !updatebool {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "数据未更改"})
		return
	}

	for _, v := range strings.Split(mod.ClusterIds, ",") {
		if !strings.ContainsAny(clusterids, v) {
			count, err := repo.CountRouteByUpstreamIdAndClusterId(info.ID, v)
			if count > 0 {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该服务组已被其他集群域名调用,请删除相关域名路由配置"})
				return
			} else if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
	}

	_, errAdd := repo.UpdateUpstreamById(info.ID, mod_update.Name, mod_update.Manager, mod_update.Description, mod_update.Customize, mod_update.ClusterIds, mod_update.Backends)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}

	for _, v := range info.ClusterIds {
		err = ConfCreateByClusterId(repo, v.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "update"
	journal.Resource = "upstream"
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
