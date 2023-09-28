package route

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
)

type AccessListPageReq struct {
	ID          int    `json:"id"`
	ClusterId   int    `json:"ClusterId"`
	DomainId    int    `json:"DomainId"`
	DomainName  string `json:"DomainName"`
	RouteId     int    `json:"RouteId"`
	AccessToken string `json:"AccessToken"`
	Keyword     string `json:"keyword"`
	AccessType  string `json:"AccessType"`
	Limit       uint
	Offset      uint
}

func getAccesslistConditions(req AccessListPageReq) string {
	var where []string
	if req.Keyword = strings.TrimSpace(req.Keyword); req.Keyword != "" {
		where = append(where, fmt.Sprintf("proxy_accesslist.domain_name like %s or proxy_accesslist.route_name like %s or proxy_accesslist.host like %s", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'"))
	}
	if req.ClusterId > 0 {
		where = append(where, fmt.Sprintf("proxy_accesslist.cluster_id = %d", req.ClusterId))
	}
	if req.DomainId > 0 {
		where = append(where, fmt.Sprintf("proxy_accesslist.domain_id = %d", req.DomainId))
	}
	if req.RouteId > 0 {
		where = append(where, fmt.Sprintf("proxy_accesslist.route_id = %d", req.RouteId))
	}
	if req.AccessType != "" {
		where = append(where, fmt.Sprintf("proxy_accesslist.access_type = %s", "'"+req.AccessType+"'"))
	}
	if len(where) > 0 {
		return " where " + strings.Join(where, " and ")
	} else {
		return ""
	}
}

func GetAccesslistList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info AccessListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	type Accesslist struct {
		ClusterId   int
		ClusterName string
		DomainId    int
		DomainName  string
		PortHttp    string
		PortHttps   string
		RouteId     int
		RouteName   string
		AccessType  string
	}
	mods, err := repo.GetAccesslistListWithPage(info.Limit, info.Offset, getAccesslistConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var accesslist Accesslist
	var accesslists []Accesslist
	for _, i := range mods {
		accesslist.ClusterId = i.ClusterId
		if i.ClusterId > 0 {
			accesslist.ClusterName, err = repo.GetClusterNameById(i.ClusterId)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
		accesslist.DomainId = i.DomainId
		mod, err := repo.GetDomainById(i.DomainId)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		accesslist.DomainName = mod.Name
		accesslist.PortHttp = mod.PortHttp
		accesslist.PortHttps = mod.PortHttps
		accesslist.RouteId = i.RouteId
		if i.RouteId == 1 {
			accesslist.RouteName = "全局"
		} else {
			accesslist.RouteName, err = repo.GetRouteNameById(i.RouteId)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
		accesslist.AccessType = i.AccessType
		accesslists = append(accesslists, accesslist)
	}
	result := make(map[string]interface{})
	result["ClusterId"] = info.ClusterId
	result["list"] = accesslists
	//总条数
	count, err := repo.CountAccesslist(getAccesslistConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		ClusterId   int
		DomainId    int
		RouteId     int
		AccessType  string
		AccessToken string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	mods, err := repo.GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType(info.ClusterId, info.DomainId, info.RouteId, info.AccessType)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	type Host struct {
		ID          int
		Host        string
		Description string
		Manager     string
	}
	type Accesslist struct {
		ClusterId   int
		ClusterName string
		DomainId    int
		DomainName  string
		RouteId     int
		RouteName   string
		AccessType  string
		Hosts       []Host
	}
	var accesslist Accesslist
	accesslist.ClusterId = info.ClusterId
	accesslist.ClusterName, err = repo.GetClusterNameById(info.ClusterId)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	accesslist.DomainId = info.DomainId
	accesslist.DomainName, err = repo.GetDomainNameById(info.DomainId)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	accesslist.RouteId = info.RouteId
	if accesslist.RouteId == 1 {
		accesslist.RouteName = "全局"
	} else {
		accesslist.RouteName, err = repo.GetRouteNameById(info.RouteId)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	accesslist.AccessType = info.AccessType
	var host Host
	var hosts []Host
	for _, i := range mods {
		host.ID = i.ID
		host.Host = i.Host
		host.Description = i.Description
		host.Manager = i.Manager
		hosts = append(hosts, host)
	}
	accesslist.Hosts = hosts
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": accesslist})
}

// 添加
func AddAccesslist(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	type Host struct {
		Host        string
		Description string
		Manager     string
	}
	type Accesslist struct {
		ClusterId   int
		DomainId    int
		RouteId     int
		AccessType  string
		Hosts       []Host
		AccessToken string
	}
	var info Accesslist
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

	var message string
	var content string

	for _, i := range info.Hosts {
		var accesslist model.Accesslist
		accesslist.ClusterId = info.ClusterId
		accesslist.DomainId = info.DomainId
		if info.RouteId == 0 {
			accesslist.RouteId = 1
		} else {
			accesslist.RouteId = info.RouteId
		}
		accesslist.AccessType = info.AccessType
		list := strings.Split(i.Host, "/")
		if accesslist.AccessType != "white" && accesslist.AccessType != "black" {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "访问控制类型异常!"})
			return
		}
		if len(list) <= 2 {
			isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", list[0])
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			if !isValidate {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "IP格式不正确"})
				return
			}
			if len(list) == 2 {
				prefix, err := strconv.Atoi(list[1])
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "掩码位格式不正确!"})
				}
				if prefix > 32 || prefix < 1 {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "掩码数值不正确!"})
					return
				}
			}
			accesslist.Host = i.Host
		} else {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "网段格式不正确"})
			return
		}
		accesslist.Description = i.Description
		accesslist.Manager = i.Manager
		modaccesslist, errAdd := repo.AddAccesslist(accesslist)
		if errAdd != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
			return
		}
		contentaccesslist, err := json.Marshal(modaccesslist)
		if err != nil {
			message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
		}
		content = content + "[add Accesslist:" + string(contentaccesslist) + "]"
	}
	err := ConfCreateByClusterId(repo, info.ClusterId)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var journal model.Journal
	journal.Title = strconv.Itoa(info.ClusterId) + "/" + strconv.Itoa(info.DomainId) + "/" + strconv.Itoa(info.RouteId) + "/" + info.AccessType
	journal.Operation = "add"
	journal.Resource = "accesslist"
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

func UpdateAccesslistByDomainIdAndRouteId(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	type Host struct {
		ID          int
		Host        string
		Description string
		Manager     string
	}
	type Accesslist struct {
		ClusterId   int
		DomainId    int
		RouteId     int
		AccessType  string
		Hosts       []Host
		AccessToken string
	}
	var info Accesslist
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
	mods, err := repo.GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType(info.ClusterId, info.DomainId, info.RouteId, info.AccessType)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	updatebool := false
	var accesslist_used []int

	var message string
	var content string
	var title = strconv.Itoa(info.ClusterId) + "/" + strconv.Itoa(info.DomainId) + "/" + strconv.Itoa(info.RouteId) + "/" + info.AccessType

	for _, i := range info.Hosts {
		var accesslist model.Accesslist
		accesslist.ID = i.ID
		accesslist.ClusterId = info.ClusterId
		accesslist.DomainId = info.DomainId
		if info.RouteId == 0 {
			accesslist.RouteId = 1
		} else {
			accesslist.RouteId = info.RouteId
		}
		accesslist.AccessType = info.AccessType
		accesslist.Description = i.Description
		accesslist.Manager = i.Manager
		list := strings.Split(i.Host, "/")
		if len(list) <= 2 {
			isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", list[0])
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			if !isValidate {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "IP格式不正确"})
				return
			}
			if len(list) == 2 {
				prefix, err := strconv.Atoi(list[1])
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "掩码位格式不正确!"})
				}
				if prefix > 32 || prefix < 1 {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "掩码数值不正确!"})
					return
				}
			}
			accesslist.Host = i.Host
		} else {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "网段格式不正确"})
			return
		}
		if accesslist.ID > 0 {
			accesslist_used = append(accesslist_used, accesslist.ID)
		} else {
			updatebool = true
			modaccesslist, err := repo.AddAccesslist(accesslist)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			contentaccesslist, err := json.Marshal(modaccesslist)
			if err != nil {
				message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
			}
			content = content + "[add Accesslist:\"" + string(contentaccesslist) + "\"]"
		}
		for _, j := range mods {
			if accesslist.ID == j.ID {
				if accesslist.Host != j.Host {
					updatebool = true
					content = content + "[update Accesslist \"" + title + "\" Host:\"" + j.Host + "\" to \"" + accesslist.Host + "\"]"
				}

				if accesslist.Description != j.Description {
					updatebool = true
					content = content + "[update Accesslist \"" + title + "\" Description:\"" + j.Description + "\" to \"" + accesslist.Description + "\"]"
				}

				if accesslist.Manager != j.Manager {
					updatebool = true
					content = content + "[update Accesslist \"" + title + "\" Manager:\"" + j.Manager + "\" to \"" + accesslist.Manager + "\"]"
				}
				if updatebool {
					_, errAdd := repo.UpdateAccesslistById(accesslist.ID, accesslist.Host, accesslist.Description, accesslist.Manager)
					if errAdd != nil {
						w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
						return
					}
				}
			}
		}
	}
	for _, i := range accesslist_used {
		for idx, j := range mods {
			if i == j.ID {
				mods = append(mods[:idx], mods[idx+1:]...)
			}
		}
	}
	if len(mods) > 0 {
		updatebool = true
		for _, i := range mods {
			_, err := repo.DeleteAccesslistById(i.ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			contentaccesslist, err := json.Marshal(i)
			if err != nil {
				message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
			}
			content = content + "[delete Accesslist:" + string(contentaccesslist) + "]"
		}
	}
	if updatebool {
		err := ConfCreateByClusterId(repo, info.ClusterId)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		var journal model.Journal
		journal.Title = title
		journal.Operation = "update"
		journal.Resource = "accesslist"
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
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "数据未更改"})
	}
}
