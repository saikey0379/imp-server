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

type DomainListPageReq struct {
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

func getDomainConditions(req DomainListPageReq) string {
	var where []string
	if req.ID > 0 {
		where = append(where, fmt.Sprintf("proxy_domain.id = %d", req.ID))
	}
	if req.ClusterId > 0 {
		where = append(where, fmt.Sprintf("proxy_domain.cluster_ids like %s%d%s", "'%", req.ClusterId, "%'"))
	}
	if req.Keyword = strings.TrimSpace(req.Keyword); req.Keyword != "" {
		where = append(where, fmt.Sprintf("( proxy_domain.name like %s or proxy_domain.description like %s )", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'"))
	}
	if len(where) > 0 {
		return " where " + strings.Join(where, " and ")
	} else {
		return ""
	}
}

func GetDomainSelect(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info DomainListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	type Domain struct {
		ID   int
		Name string
	}
	mods, err := repo.GetDomainListWithPage(0, 0, getDomainConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var domain Domain
	var domains []Domain
	for _, i := range mods {
		domain.ID = i.ID
		domain.Name = i.Name
		if i.PortHttp != "" {
			domain.Name = domain.Name + ":" + i.PortHttp
		}
		if i.PortHttps != "" {
			domain.Name = domain.Name + ":" + i.PortHttps
		}
		domains = append(domains, domain)
	}
	result := make(map[string]interface{})
	result["list"] = domains
	//总条数
	count := len(domains)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetRouteSelectByDomainId(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info DomainListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	type Route struct {
		ID    int
		Route string
	}
	mods, err := repo.GetRouteListByDomainId(info.DomainId)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var route Route
	var routes []Route
	for _, i := range mods {
		route.ID = i.ID
		route.Route = i.Route
		routes = append(routes, route)
	}
	result := make(map[string]interface{})
	result["list"] = routes
	//总条数
	count := len(routes)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

// 添加
func AddDomain(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	type ClusterId struct {
		ID int
	}
	type DomainRoute struct {
		ID          int
		Route       string
		Manager     string
		MatchType   string
		Customize   string
		Description string
		UpstreamId  int
	}
	type DomainDetail struct {
		Name        string
		Manager     string
		Description string
		ClusterIds  []ClusterId
		ProxyType   int
		PortHttp    string
		PortHttps   string
		Http2       bool
		CertId      int
		Customize   string
		Routes      []DomainRoute
		AccessToken string
	}
	var info DomainDetail
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
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请填写域名!"})
		return
	}
	var domainmod model.Domain
	domainid := int(time.Now().Unix())
	domainmod.ID = domainid
	domainmod.Name = strings.TrimSpace(info.Name)
	domainmod.Manager = strings.TrimSpace(info.Manager)
	domainmod.Description = strings.TrimSpace(info.Description)
	domainmod.ProxyType = info.ProxyType
	domainmod.Http2 = info.Http2
	domainmod.CertId = info.CertId
	domainmod.UpdatedAt = time.Now()
	var clusterids string
	for i, v := range info.ClusterIds {
		clusterid := strconv.Itoa(v.ID)
		if i == 0 {
			clusterids = clusterid
		} else {
			clusterids = clusterids + "," + clusterid
		}
	}
	domainmod.ClusterIds = clusterids
	customize_domain := strings.TrimSpace(info.Customize)
	for k, i := range strings.Split(customize_domain, "\n") {
		if k == len(strings.Split(customize_domain, "\n"))-1 {
			domainmod.Customize += strings.TrimSpace(i)
		} else {
			domainmod.Customize += strings.TrimSpace(i) + "\n"
		}
	}
	var port int
	var err error
	var porthttp = strings.TrimSpace(info.PortHttp)
	var porthttps = strings.TrimSpace(info.PortHttps)
	if porthttp == "" && porthttps == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "监听端口未指定，请设置HTTP或HTTPS端口"})
		return
	}
	if porthttp != "" {
		for j, i := range strings.Split(porthttp, ",") {
			port, err = strconv.Atoi(i)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			} else if port <= 0 || port >= 65535 {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Http端口数值异常"})
				return
			}
			if j == 0 {
				domainmod.PortHttp += i
			} else {
				domainmod.PortHttp += "," + i
			}
		}
	}

	if porthttps != "" {
		if info.CertId <= 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SSL证书未指定"})
			return
		}
		for j, i := range strings.Split(porthttps, ",") {
			port, err = strconv.Atoi(i)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			} else if port <= 0 || port >= 65535 {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Https端口数值异常"})
				return
			}
			if j == 0 {
				domainmod.PortHttps += i
			} else {
				domainmod.PortHttps += "," + i
			}
		}
	}
	moddomain, errAdd := repo.AddDomain(domainmod)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}

	var message string
	var content string
	contentdomain, err := json.Marshal(moddomain)
	if err != nil {
		message = fmt.Sprintf("[Umarshal failed:%s]", err)
	}
	content = "[add Domain:" + string(contentdomain) + "]"

	for _, i := range info.Routes {
		var route model.Route
		route.DomainId = domainid
		route.Route = strings.Replace(i.Route, " ", "", -1)
		route.Manager = i.Manager
		route.MatchType = i.MatchType
		route.Description = strings.TrimSpace(i.Description)
		route.UpstreamId = i.UpstreamId
		customize_route := strings.TrimSpace(i.Customize)
		for k, i := range strings.Split(customize_route, "\n") {
			if k == len(strings.Split(customize_route, "\n"))-1 {
				route.Customize += strings.TrimSpace(i)
			} else {
				route.Customize += strings.TrimSpace(i) + "\n"
			}
		}
		modroute, err := repo.AddRoute(route)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		contentroute, err := json.Marshal(modroute)
		if err != nil {
			message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
		}
		content = content + "[add Route:" + string(contentroute) + "]"
	}
	for _, v := range info.ClusterIds {
		err = ConfCreateByClusterId(repo, v.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}

	var journal model.Journal
	journal.Title = domainmod.Name
	journal.Operation = "add"
	journal.Resource = "domain"
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

func DeleteDomainById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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
	count, err := repo.CountRouteByDomainId(info.ID)
	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该域名已设置路由,请删除相关路由配置"})
		return
	} else if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	count, err = repo.CountAccesslistByDomainId(info.ID)
	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该域名已设置访问控制,请删除相关白名单配置"})
		return
	} else if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	mod, err := repo.GetDomainById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	_, err = repo.DeleteDomainById(info.ID)
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
	content, err := json.Marshal(mod)
	if err != nil {
		message = fmt.Sprintf("[Umarshal failed:%s]", err)
	}
	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "delete"
	journal.Resource = "domain"
	journal.Content = "[delete Domain:" + string(content) + "]"
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

func UpdateDomainById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	type ClusterId struct {
		ID int
	}
	type DomainRoute struct {
		ID           int
		Index        int
		Route        string
		Manager      string
		MatchType    string
		Customize    string
		Description  string
		UpstreamId   int
		UpstreamUsed bool
	}
	type DomainDetail struct {
		ID          int
		Manager     string
		Description string
		ClusterIds  []ClusterId
		ProxyType   int
		PortHttp    string
		PortHttps   string
		Http2       bool
		CertId      int
		Customize   string
		Routes      []DomainRoute
		AccessToken string
	}
	var info DomainDetail
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
	var clusterids string
	for i, v := range info.ClusterIds {
		clusterid := strconv.Itoa(v.ID)
		if i == 0 {
			clusterids = clusterid
		} else {
			clusterids = clusterids + "," + clusterid
		}
	}
	updatebool := false
	mod, err := repo.GetDomainById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var domainmod model.Domain
	domainmod.Manager = strings.TrimSpace(info.Manager)
	domainmod.Description = strings.TrimSpace(info.Description)
	domainmod.ClusterIds = clusterids
	domainmod.ProxyType = info.ProxyType
	domainmod.Http2 = info.Http2
	domainmod.CertId = info.CertId
	customize_domain := strings.TrimSpace(info.Customize)
	for k, i := range strings.Split(customize_domain, "\n") {
		if k == len(strings.Split(customize_domain, "\n"))-1 {
			domainmod.Customize += strings.TrimSpace(i)
		} else {
			domainmod.Customize += strings.TrimSpace(i) + "\n"
		}
	}

	domainmod.PortHttp = strings.TrimSpace(info.PortHttp)
	domainmod.PortHttps = strings.TrimSpace(info.PortHttps)
	if domainmod.PortHttp == "" && domainmod.PortHttps == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "监听端口未指定，请设置HTTP或HTTPS端口"})
		return
	}

	var content string
	var message string
	if domainmod.Description != mod.Description {
		updatebool = true
		content = content + "[update Description:\"" + mod.Description + "\" to \"" + domainmod.Description + "\"]"
	}

	if domainmod.Manager != mod.Manager {
		updatebool = true
		content = content + "[update Manager:\"" + mod.Manager + "\" to \"" + domainmod.Manager + "\"]"
	}

	if domainmod.ClusterIds != mod.ClusterIds {
		updatebool = true
		content = content + "[update ClusterIds:\"" + mod.ClusterIds + "\" to \"" + domainmod.ClusterIds + "\"]"
	}

	if domainmod.ProxyType != mod.ProxyType {
		updatebool = true
		content = content + "[update ProxyType:\"" + strconv.Itoa(mod.ProxyType) + "\" to \"" + strconv.Itoa(domainmod.ProxyType) + "\"]"
	}

	if domainmod.Customize != mod.Customize {
		updatebool = true
		content = content + "[update Customize:\"" + mod.Customize + "\" to \"" + domainmod.Customize + "\"]"
	}

	if domainmod.PortHttp != mod.PortHttp && domainmod.PortHttp != "" {
		updatebool = true
		var port int
		var porthttp string
		for j, i := range strings.Split(domainmod.PortHttp, ",") {
			port, err = strconv.Atoi(i)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			} else if port <= 0 || port >= 65535 {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Http端口数值异常"})
				return
			}
			if j == 0 {
				porthttp += i
			} else {
				porthttp += "," + i
			}
		}
		domainmod.PortHttp = porthttp
		content = content + "[update PortHttp:\"" + mod.PortHttp + "\" to \"" + domainmod.PortHttp + "\"]"
	}

	if domainmod.PortHttps != mod.PortHttps {
		updatebool = true
		if domainmod.PortHttps != "" {
			var port int
			var porthttps string
			for j, i := range strings.Split(domainmod.PortHttps, ",") {
				port, err = strconv.Atoi(i)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				} else if port <= 0 || port >= 65535 {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Https端口数值异常"})
					return
				}
				if j == 0 {
					porthttps += i
				} else {
					porthttps += "," + i
				}
			}
			domainmod.PortHttps = porthttps
			content = content + "[update PortHttps:\"" + mod.PortHttps + "\" to \"" + domainmod.PortHttps + "\"]"
		} else {
			content = content + "[update PortHttps:\"" + mod.PortHttps + "\" to \"\"]"
			if mod.Http2 {
				domainmod.Http2 = false
				content = content + "[update Http2:\"" + strconv.FormatBool(mod.Http2) + "\" to \"false\"]"
			}
			if mod.CertId != 0 {
				domainmod.CertId = 0
				content = content + "[update CertId:\"" + strconv.Itoa(mod.CertId) + "\" to \"0\"]"
			}
		}
	}

	if domainmod.PortHttps != "" && domainmod.Http2 != mod.Http2 {
		updatebool = true
		content = content + "[update Http2:\"" + strconv.FormatBool(mod.Http2) + "\" to \"" + strconv.FormatBool(domainmod.Http2) + "\"]"
	}

	if domainmod.PortHttps != "" && domainmod.CertId != mod.CertId {
		if info.CertId <= 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SSL证书未指定"})
			return
		}
		updatebool = true
		content = content + "[update CertId:\"" + strconv.Itoa(mod.CertId) + "\" to \"" + strconv.Itoa(domainmod.CertId) + "\"]"
	}

	if updatebool {
		_, err := repo.UpdateDomainById(info.ID, domainmod.Manager, domainmod.Description, domainmod.ClusterIds, domainmod.ProxyType, domainmod.PortHttp, domainmod.PortHttps, domainmod.Http2, domainmod.CertId, domainmod.Customize)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	route_all, err := repo.GetRouteFullListByDomainId(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var routelist_used []int
	for k, i := range info.Routes {
		var route model.Route
		route.DomainId = info.ID
		route.Index = k
		route.Route = strings.Replace(i.Route, " ", "", -1)
		route.Manager = i.Manager
		route.MatchType = i.MatchType
		route.Description = strings.TrimSpace(i.Description)
		route.UpstreamId = i.UpstreamId
		customize_route := strings.TrimSpace(i.Customize)
		for k, i := range strings.Split(customize_route, "\n") {
			if k == len(strings.Split(customize_route, "\n"))-1 {
				route.Customize += strings.TrimSpace(i)
			} else {
				route.Customize += strings.TrimSpace(i) + "\n"
			}
		}
		if i.ID > 0 {
			routelist_used = append(routelist_used, i.ID)
		} else {
			updatebool = true
			modroute, err := repo.AddRoute(route)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			contentroute, err := json.Marshal(modroute)
			if err != nil {
				message = fmt.Sprintf("[Umarshal failed:%s]", err)
			}
			content = content + "[add Route:\"" + string(contentroute) + "\"]"
			//newupstream++
			used, err := repo.CountRouteByUpstreamId(i.UpstreamId)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			err = repo.UpdateUpstreamUsedById(i.UpstreamId, used+1)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
		for _, j := range route_all {
			if i.ID == j.ID {
				if route.Index != j.Index {
					updatebool = true
					content = content + "[update Route \"" + j.Route + "\" Index:\"" + strconv.Itoa(j.Index) + "\" to \"" + strconv.Itoa(route.Index) + "\"]"
				}
				if route.Route != j.Route {
					updatebool = true
					content = content + "[update Route \"" + j.Route + "\" Route:\"" + j.Route + "\" to \"" + route.Route + "\"]"
				}
				if route.MatchType != j.MatchType {
					updatebool = true
					content = content + "[update Route \"" + j.Route + "\" MatchType:\"" + j.MatchType + "\" to \"" + route.MatchType + "\"]"
				}
				if route.Manager != j.Manager {
					updatebool = true
					content = content + "[update Route \"" + j.Route + "\" Manager:\"" + j.Manager + "\" to \"" + route.Manager + "\"]"
				}
				if route.Customize != j.Customize {
					updatebool = true
					content = content + "[update Route \"" + j.Route + "\" Customize:\"" + j.Customize + "\" to \"" + route.Customize + "\"]"
				}
				if route.Description != j.Description {
					updatebool = true
					content = content + "[update Route \"" + j.Route + "\" Description:\"" + j.Description + "\" to \"" + route.Description + "\"]"
				}
				if route.UpstreamId != j.UpstreamId {
					updatebool = true
					content = content + "[update Route \"" + j.Route + "\" UpstreamId:\"" + strconv.Itoa(j.UpstreamId) + "\" to \"" + strconv.Itoa(route.UpstreamId) + "\"]"
				}

				if updatebool {
					_, err := repo.UpdateRouteById(i.ID, route.Index, route.Route, route.Manager, route.Description, route.DomainId, route.MatchType, route.Customize, route.UpstreamId)
					if err != nil {
						w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
						return
					}
					if i.UpstreamId != j.UpstreamId {
						//newupstream++
						used, err := repo.CountRouteByUpstreamId(i.UpstreamId)
						if err != nil {
							w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
							return
						}
						err = repo.UpdateUpstreamUsedById(i.UpstreamId, used)
						if err != nil {
							w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
							return
						}
						//oldupstream--
						used, err = repo.CountRouteByUpstreamId(j.UpstreamId)
						if err != nil {
							w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
							return
						}
						err = repo.UpdateUpstreamUsedById(j.UpstreamId, used)
						if err != nil {
							w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
							return
						}
					}
				}
			}
		}
	}

	for _, i := range routelist_used {
		for idx, j := range route_all {
			if i == j.ID {
				route_all = append(route_all[:idx], route_all[idx+1:]...)
			}
		}
	}
	route_del := route_all
	for _, i := range route_del {
		updatebool = true
		count, err := repo.CountAccesslistByDomainIdAndRouteId(info.ID, i.ID)
		if count > 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "欲删除路由已设置访问控制,请删除相关白名单配置"})
			return
		} else if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		err = repo.DeleteRouteById(i.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		contentroute, err := json.Marshal(i)
		if err != nil {
			message = fmt.Sprintf("[Umarshal failed:%s]", err)
		}
		content = content + "[delete Route:" + string(contentroute) + "]"
		//oldupstream--
		used, err := repo.GetUpstreamUsedById(i.UpstreamId)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		err = repo.UpdateUpstreamUsedById(i.UpstreamId, used-1)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	if updatebool {
		for _, v := range info.ClusterIds {
			err := ConfCreateByClusterId(repo, v.ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
		var journal model.Journal
		journal.Title = mod.Name
		journal.Operation = "update"
		journal.Resource = "domain"
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

func GetDomainById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		ID int
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	mod, err := repo.GetDomainById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	mod_routes, err := repo.GetRouteFullListByDomainId(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	type ClusterId struct {
		ID int
	}
	type AccessList struct {
		ClusterId int
		Type      string
	}
	type Backend struct {
		Ip           string
		Port         string
		BakCustomize string
	}
	type DomainRoute struct {
		ID               int
		Route            string
		Manager          string
		MatchType        string
		Customize        string
		Description      string
		AccessLists      []AccessList
		UpstreamId       int
		UpstreamName     string
		UpstreamBackends []Backend
		BackendsCount    int
		UpdatedAt        utils.ISOTime
	}
	type DomainDetail struct {
		ID          int
		Name        string
		Manager     string
		Description string
		AccessLists []AccessList
		ClusterIds  []ClusterId
		ProxyType   int
		PortHttp    string
		PortHttps   string
		Http2       bool
		CertName    string
		CertId      int
		Customize   string
		Routes      []DomainRoute
		RoutesCount int
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}

	var domaindetail DomainDetail
	domaindetail.ID = mod.ID
	domaindetail.Name = mod.Name
	domaindetail.Manager = mod.Manager
	domaindetail.Description = mod.Description
	domaindetail.PortHttp = mod.PortHttp
	domaindetail.PortHttps = mod.PortHttps
	domaindetail.Http2 = mod.Http2
	domaindetail.Customize = mod.Customize
	domaindetail.CreatedAt = utils.ISOTime(mod.CreatedAt)
	domaindetail.UpdatedAt = utils.ISOTime(mod.UpdatedAt)

	var clusterid ClusterId
	var clusterids []ClusterId
	var req AccessListPageReq
	req.DomainId = domaindetail.ID
	for _, i := range strings.Split(mod.ClusterIds, ",") {
		clusterid.ID, err = strconv.Atoi(i)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		clusterids = append(clusterids, clusterid)
		req.ClusterId = clusterid.ID
		req.RouteId = 1
		mods, _ := repo.GetAccesslistListWithPage(0, 0, getAccesslistConditions(req))
		if len(mods) > 0 {
			var accesslists []AccessList
			for _, v := range mods {
				var accesslist AccessList
				accesslist.ClusterId = v.ClusterId
				accesslist.Type = v.AccessType
				accesslists = append(accesslists, accesslist)
			}
			domaindetail.AccessLists = accesslists
		}
	}
	domaindetail.ClusterIds = clusterids
	if domaindetail.PortHttps != "" {
		certname, err := repo.GetCertNameById(mod.CertId)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		domaindetail.CertName = certname
		domaindetail.CertId = mod.CertId
	}
	var routes []DomainRoute
	for _, i := range mod_routes {
		var route DomainRoute
		route.ID = i.ID
		route.Route = i.Route
		route.Manager = i.Manager
		route.MatchType = i.MatchType
		route.Customize = i.Customize
		route.Description = i.Description
		route.UpstreamId = i.UpstreamId
		route.UpstreamName = i.UpstreamName
		route.UpdatedAt = utils.ISOTime(i.UpdatedAt)
		req.RouteId = route.ID
		for _, i := range strings.Split(mod.ClusterIds, ",") {
			clusterid, err := strconv.Atoi(i)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			req.ClusterId = clusterid
			mods, _ := repo.GetAccesslistListWithPage(0, 0, getAccesslistConditions(req))
			if len(mods) > 0 {
				var accesslists []AccessList
				for _, v := range mods {
					var accesslist AccessList
					accesslist.ClusterId = v.ClusterId
					accesslist.Type = v.AccessType
					accesslists = append(accesslists, accesslist)
				}
				route.AccessLists = accesslists
			}
		}
		var backend Backend
		var backends []Backend
		for _, j := range strings.Split(i.UpstreamBackends, ",") {
			for index, k := range strings.Split(j, ":") {
				if index == 0 {
					backend.Ip = k
				} else if index == 1 {
					backend.Port = k
				} else if index == 2 {
					backend.BakCustomize = k
				}
			}
			backends = append(backends, backend)
		}
		backendscount := len(backends)
		route.UpstreamBackends = backends
		route.BackendsCount = backendscount
		routes = append(routes, route)
	}
	routescount := len(routes)
	domaindetail.Routes = routes
	domaindetail.RoutesCount = routescount
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": domaindetail})
}

func GetDomainList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info DomainListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	mods, err := repo.GetDomainListWithPage(info.Limit, info.Offset, getDomainConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	type ClusterId struct {
		ID int
	}
	type Domaininfo struct {
		ID          int
		Name        string
		Manager     string
		Description string
		ClusterIds  []ClusterId
		PortHttp    string
		PortHttps   string
		ProxyType   int
		UpdatedAt   utils.ISOTime
		CreatedAt   utils.ISOTime
	}
	var domaininfo Domaininfo
	var domainlist []Domaininfo
	for _, i := range mods {
		domaininfo.ID = i.ID
		domaininfo.Name = i.Name
		domaininfo.Manager = i.Manager
		domaininfo.Description = i.Description
		domaininfo.PortHttp = i.PortHttp
		domaininfo.PortHttps = i.PortHttps
		domaininfo.ProxyType = i.ProxyType
		domaininfo.CreatedAt = utils.ISOTime(i.CreatedAt)
		domaininfo.UpdatedAt = utils.ISOTime(i.UpdatedAt)
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
		domaininfo.ClusterIds = clusterids
		domainlist = append(domainlist, domaininfo)
	}
	result := make(map[string]interface{})
	result["list"] = domainlist
	result["ClusterId"] = info.ClusterId
	//总条数
	count, err := repo.CountDomain(getDomainConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}
