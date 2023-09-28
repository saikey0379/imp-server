package route

import (
	"encoding/json"
	"fmt"
	"github.com/saikey0379/imp-server/pkg/server/nginx"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/known"
	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
	"github.com/saikey0379/imp-server/pkg/utils"
)

type ClusterListPageReq struct {
	ID          int    `json:"id"`
	AccessToken string `json:"AccessToken"`
	Keyword     string `json:"keyword"`
	Limit       uint
	Offset      uint
}

func GetClusterSelect(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	type Cluster struct {
		ID   int
		Name string
	}

	var info ClusterListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	mods, err := repo.GetClusterListWithPage(0, 0, "")
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var cluster Cluster
	var clusters []Cluster
	for _, i := range mods {
		cluster.ID = i.ID
		cluster.Name = i.Name
		clusters = append(clusters, cluster)
	}
	result := make(map[string]interface{})
	result["list"] = clusters
	result["recordCount"] = len(mods)
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetClusterList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info ClusterListPageReq
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

	type Cluster struct {
		ID          int
		Name        string
		ConfMd5Curr string
		Description string
		CStatus     string
		Status      string
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}
	mods, err := repo.GetClusterListWithPage(info.Limit, info.Offset, getClusterConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var cluster Cluster
	var clusters []Cluster
	for _, i := range mods {
		cluster.ID = i.ID
		cluster.Name = i.Name
		cluster.Description = i.Description
		cluster.Status = i.Status
		cluster.CreatedAt = utils.ISOTime(i.CreatedAt)
		cluster.UpdatedAt = utils.ISOTime(i.UpdatedAt)
		confile_tmp := path.Join(known.RootProxy, "conf/"+strconv.Itoa(int(i.ID))+"/nginx.conf.tmp")
		confile := path.Join(known.RootProxy, "conf/"+strconv.Itoa(int(i.ID))+"/nginx.conf")

		md5_tmp, _ := utils.GetMd5ByFile(confile_tmp)
		md5, _ := utils.GetMd5ByFile(confile)
		if md5_tmp == md5 {
			cluster.CStatus = "正常"
		} else if !utils.FileExist(confile_tmp) {
			cluster.CStatus = "待生成"
		} else if !utils.FileExist(confile) || md5_tmp != md5 {
			cluster.CStatus = "待同步"
		}
		clusters = append(clusters, cluster)
	}
	result := make(map[string]interface{})
	result["list"] = clusters
	count, err := repo.CountCluster(getClusterConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func getClusterConditions(req ClusterListPageReq) string {
	var where []string
	if req.Keyword = strings.TrimSpace(req.Keyword); req.Keyword != "" {
		where = append(where, fmt.Sprintf("( proxy_cluster.name like %s or proxy_cluster.description like %s )", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'"))
	}
	if len(where) > 0 {
		return " where " + strings.Join(where, " and ")
	} else {
		return ""
	}
}

func GetClusterById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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
	_, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}
	mod, err := repo.GetClusterById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	type Backend struct {
		Ip string
	}
	type Cluster struct {
		ID          int
		Name        string
		SSHUser     string
		SSHPort     int
		SSHKey      string
		ExecTest    string
		ExecLoad    string
		PathConf    string
		PathKey     string
		Description string
		Backends    []Backend
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}
	var cluster Cluster
	cluster.ID = info.ID
	cluster.Name = mod.Name
	cluster.SSHUser = mod.SSHUser
	cluster.SSHPort = mod.SSHPort
	cluster.SSHKey = mod.SSHKey
	cluster.ExecTest = mod.ExecTest
	cluster.ExecLoad = mod.ExecLoad
	cluster.PathConf = mod.PathConf
	cluster.PathKey = mod.PathKey
	cluster.Description = mod.Description
	cluster.CreatedAt = utils.ISOTime(mod.CreatedAt)
	cluster.UpdatedAt = utils.ISOTime(mod.UpdatedAt)
	var backend Backend
	var backends []Backend
	for _, i := range strings.Split(mod.Backends, ",") {
		backend.Ip = i
		backends = append(backends, backend)
	}
	cluster.Backends = backends
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": cluster})
}

func UpdateClusterById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	type Backend struct {
		Ip string
	}
	type Cluster struct {
		ID          int
		Name        string
		Backends    []Backend
		SSHUser     string
		SSHPort     string
		SSHKey      string
		Description string
		ExecTest    string
		ExecLoad    string
		PathConf    string
		PathKey     string
		AccessToken string
	}
	var info Cluster
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
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
	mod, err := repo.GetClusterById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var backends string
	for k, i := range info.Backends {
		if !utils.IsIpAddress(i.Ip) {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "IP地址异常,请确认"})
			return
		}
		if k == 0 {
			backends += i.Ip
		} else {
			backends += "," + i.Ip
		}
	}

	var mod_update model.Cluster
	mod_update.Name = strings.TrimSpace(info.Name)
	mod_update.SSHUser = strings.TrimSpace(info.SSHUser)
	mod.SSHPort, err = strconv.Atoi(info.SSHPort)
	if err != nil || utils.IsValidPort(mod.SSHPort) {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "端口号错误!"})
		return
	}
	mod_update.SSHKey = strings.TrimSpace(info.SSHKey)
	mod_update.ExecTest = strings.TrimSpace(info.ExecTest)
	mod_update.ExecLoad = strings.TrimSpace(info.ExecLoad)
	mod_update.PathConf = strings.TrimSpace(info.PathConf)
	mod_update.PathKey = strings.TrimSpace(info.PathKey)
	mod_update.Description = strings.TrimSpace(info.Description)
	mod_update.Backends = backends
	mod_update.UpdatedAt = time.Now()

	var updatebool = false
	var message string
	var content string

	if mod_update.Name != mod.Name {
		updatebool = true
		content = content + "[update Name:\"" + mod.Name + "\" to \"" + mod_update.Name + "\"]"
	}
	if mod_update.SSHUser != mod.SSHUser {
		updatebool = true
		content = content + "[update SSHUser:\"" + mod.SSHUser + "\" to \"" + mod_update.SSHUser + "\"]"
	}
	if mod_update.SSHPort != mod.SSHPort {
		updatebool = true
		content = content + fmt.Sprintf("[update SSHPort:\"%d\" to \"%d\"]", mod.SSHPort, mod_update.SSHPort)
	}
	if mod_update.SSHKey != mod.SSHKey {
		updatebool = true
		content = content + "[update SSHKey:\"" + mod.SSHKey + "\" to \"" + mod_update.SSHKey + "\"]"
	}
	if mod_update.ExecTest != mod.ExecTest {
		updatebool = true
		content = content + "[update ExecTest:\"" + mod.ExecTest + "\" to \"" + mod_update.ExecTest + "\"]"
	}
	if mod_update.ExecLoad != mod.ExecLoad {
		updatebool = true
		content = content + "[update ExecLoad:\"" + mod.ExecLoad + "\" to \"" + mod_update.ExecLoad + "\"]"
	}
	if mod_update.PathConf != mod.PathConf {
		updatebool = true
		content = content + "[update PathConf:\"" + mod.PathConf + "\" to \"" + mod_update.PathConf + "\"]"
	}
	if mod_update.PathKey != mod.PathKey {
		updatebool = true
		content = content + "[update PathKey:\"" + mod.PathKey + "\" to \"" + mod_update.PathKey + "\"]"
	}
	if mod_update.PathKey != mod.PathKey {
		updatebool = true
		content = content + "[update PathKey:\"" + mod.PathKey + "\" to \"" + mod_update.PathKey + "\"]"
	}
	if mod_update.Description != mod.Description {
		updatebool = true
		content = content + "[update Description:\"" + mod.Description + "\" to \"" + mod_update.Description + "\"]"
	}
	if mod_update.Backends != mod.Backends {
		updatebool = true
		content = content + "[update Backends:\"" + mod.Backends + "\" to \"" + mod_update.Backends + "\"]"
	}
	if !updatebool {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "数据未更改"})
		return
	}

	_, errAdd := repo.UpdateClusterById(info.ID, mod_update.Name, mod_update.Description, mod_update.SSHUser, mod_update.SSHPort, mod_update.SSHKey, mod_update.Backends, mod_update.PathConf, mod_update.PathKey, mod_update.ExecTest, mod_update.ExecLoad)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "update"
	journal.Resource = "cluster"
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

// 添加
func AddCluster(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	type Backend struct {
		Ip string
	}
	type Cluster struct {
		ID          int
		Name        string
		Backends    []Backend
		SSHUser     string
		SSHPort     string
		SSHKey      string
		Description string
		ExecTest    string
		ExecLoad    string
		PathConf    string
		PathKey     string
		AccessToken string
	}
	var info Cluster
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}
	info.AccessToken = strings.TrimSpace(info.AccessToken)
	user, err := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	if info.Name == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请填写服务组名称!"})
		return
	}

	var mod model.Cluster
	mod.ID = int(time.Now().Unix())
	mod.Name = strings.TrimSpace(info.Name)
	mod.SSHUser = strings.TrimSpace(info.SSHUser)
	mod.SSHPort, err = strconv.Atoi(info.SSHPort)
	if err != nil || utils.IsValidPort(mod.SSHPort) {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "端口号错误!"})
		return
	}

	mod.SSHKey = strings.TrimSpace(info.SSHKey)
	mod.ExecTest = strings.TrimSpace(info.ExecTest)
	mod.ExecLoad = strings.TrimSpace(info.ExecLoad)
	mod.PathConf = strings.TrimSpace(info.PathConf)
	mod.PathKey = strings.TrimSpace(info.PathKey)
	mod.Description = strings.TrimSpace(info.Description)
	mod.UpdatedAt = time.Now()
	var backends string
	for k, i := range info.Backends {
		if !utils.IsIpAddress(i.Ip) {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "IP地址异常,请确认"})
			return
		}
		if k == 0 {
			backends = i.Ip
		} else {
			backends = backends + "," + i.Ip
		}
	}
	mod.Backends = backends
	_, errAdd := repo.AddCluster(mod)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}
	rootDir := known.RootProxy
	confGDir := path.Join(rootDir, "conf/global")
	if !utils.FileExist(confGDir) {
		err := os.MkdirAll(confGDir, 0700)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	confDir := path.Join(rootDir, "conf/"+strconv.Itoa(int(mod.ID)))
	if !utils.FileExist(confDir) {
		err := os.MkdirAll(confDir, 0700)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	sslGDir := path.Join(rootDir, "ssl/global")
	if !utils.FileExist(sslGDir) {
		err := os.MkdirAll(sslGDir, 0700)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	sslDir := path.Join(rootDir, "ssl/"+strconv.Itoa(int(mod.ID)))
	if !utils.FileExist(sslDir) {
		err := os.MkdirAll(sslDir, 0700)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}

	var message string
	var content string
	contentcluster, err := json.Marshal(mod)
	if err != nil {
		message = fmt.Sprintf("[Umarshal failed:%s]", err)
	}
	content = "[add Cluster:" + string(contentcluster) + "]"

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "add"
	journal.Resource = "cluster"
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

func DeleteClusterById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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
	mod, err := repo.GetClusterById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	_, err = repo.DeleteClusterById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var message string
	contentcluster, err := json.Marshal(mod)
	if err != nil {
		message = fmt.Sprintf("[Umarshal failed:%s]", err)
	}
	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "delete"
	journal.Resource = "cluster"
	journal.Content = "[delete Cluster:" + string(contentcluster) + "]"
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

func ConfCreateByClusterId(repo model.Repo, id int) (err error) {
	conf_cur, err := GetNginxConfByClusterId(repo, id)
	if err != nil {
		return err
	}
	rootDir := known.RootProxy
	confDir := path.Join(rootDir, "conf/"+strconv.Itoa(id))
	confile := path.Join(confDir, "nginx.conf.tmp")
	if utils.FileExist(confile) {
		err := os.Remove(confile)
		if err != nil {
			return err
		}
	}
	f, err := os.OpenFile(confile, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = io.WriteString(f, conf_cur)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func GetNginxConfByClusterId(repo model.Repo, id int) (conf string, err error) {
	mods_upstream, err := repo.GetUpstreamListByClusterId(id)
	if err != nil {
		return "", err
	}
	var conf_upstream, conf_server string
	for _, i := range mods_upstream {
		conf_upstream += "upstream " + i.Name + " {\n"
		for _, j := range strings.Split(i.Backends, ",") {
			for q, k := range strings.Split(j, ":") {
				if q == 0 {
					conf_upstream += "    server " + k
				} else if q == 1 {
					conf_upstream += ":" + k
				} else if q == 2 {
					if k != "" {
						conf_upstream += " " + k + "\n"
					} else {
						conf_upstream += ";\n"
					}
				}
			}
		}
		for _, cu_upstream := range strings.Split(i.Customize, "\n") {
			if cu_upstream != "" {
				conf_upstream += "    " + cu_upstream + "\n"
			}
		}
		conf_upstream += "}\n"
	}
	mods_domain, err := repo.GetDomainListByClusterId(id)
	if err != nil {
		return "", err
	}
	mod_cluster, err := repo.GetClusterById(id)
	if err != nil {
		return "", err
	}
	for _, i := range mods_domain {
		conf_server += "server {\n"
		conf_server += "    server_name " + i.Name + ";\n"
		if i.PortHttp != "" {
			for _, p := range strings.Split(i.PortHttp, ",") {
				conf_server += "    listen " + p + ";\n"
			}
		}
		if i.PortHttps != "" && i.CertId > 0 {
			for _, ps := range strings.Split(i.PortHttps, ",") {
				if i.Http2 {
					conf_server += "    listen " + ps + " ssl http2;\n"
				} else {
					conf_server += "    listen " + ps + " ssl;\n"
				}
			}
			mod_cert, err := repo.GetCertById(i.CertId)
			if err != nil {
				return "", err
			}
			conf_server += "    ssl_certificate " + path.Join(mod_cluster.PathKey, mod_cert.FileCert) + ";\n"
			conf_server += "    ssl_certificate_key " + path.Join(mod_cluster.PathKey, mod_cert.FileKey) + ";\n"
		}
		mods_whitelistGlobal, err := repo.GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType(id, i.ID, 1, "white")
		if err != nil {
			return "", err
		}
		if len(mods_whitelistGlobal) > 0 {
			for _, j := range mods_whitelistGlobal {
				conf_server += "    allow " + j.Host + ";\n"
			}
			conf_server += "    deny all;\n"
		}
		mods_blacklistGlobal, err := repo.GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType(id, i.ID, 1, "black")
		if err != nil {
			return "", err
		}
		if len(mods_blacklistGlobal) > 0 {
			for _, j := range mods_blacklistGlobal {
				conf_server += "    deny " + j.Host + ";\n"
			}
		}
		for _, cu_domain := range strings.Split(i.Customize, "\n") {
			if cu_domain != "" {
				conf_server += "    " + cu_domain + "\n"
			}
		}
		mods_route, err := repo.GetRouteFullListByDomainId(i.ID)
		if err != nil {
			return "", err
		}
		for _, j := range mods_route {
			if j.MatchType != "[default]" {
				conf_server += "    location " + j.MatchType + " " + j.Route + " {\n"
			} else {
				conf_server += "    location " + " " + j.Route + " {\n"
			}
			mods_whitelistRoute, err := repo.GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType(id, i.ID, j.ID, "white")
			if err != nil {
				return "", err
			}
			if len(mods_whitelistRoute) > 0 {
				for _, j := range mods_whitelistRoute {
					conf_server += "        allow " + j.Host + ";\n"
				}
				conf_server += "        deny all;\n"
			}
			mods_blacklistRoute, err := repo.GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType(id, i.ID, j.ID, "black")
			if err != nil {
				return "", err
			}
			if len(mods_blacklistRoute) > 0 {
				for _, j := range mods_blacklistRoute {
					conf_server += "        deny " + j.Host + ";\n"
				}
			}
			conf_server += "        proxy_pass http://" + j.UpstreamName + ";\n"
			for _, cu_route := range strings.Split(j.Customize, "\n") {
				if cu_route != "" {
					conf_server += "        " + cu_route + "\n"
				}
			}
			conf_server += "    }\n"
		}
		conf_server += "}\n"
	}
	conf = conf_upstream + conf_server
	return conf, err
}

func GetClusterConfById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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
	conf_temp_file := path.Join(known.RootProxy, "conf/"+strconv.Itoa(int(info.ID))+"/nginx.conf.tmp")
	conf_used_file := path.Join(known.RootProxy, "conf/"+strconv.Itoa(int(info.ID))+"/nginx.conf")
	conf_diff, err := utils.FileDiff(conf_used_file, conf_temp_file)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
	}
	result := make(map[string]interface{})
	result["ID"] = info.ID
	result["ConfDiff"] = conf_diff
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func UploadSSHFile(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "text/html; charset=utf-8")
	r.ParseForm()
	file, handle, err := r.FormFile("file")
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	destfile := r.FormValue("filename")
	dirSSH := path.Join(known.RootProxy, ".ssh")
	if !utils.FileExist(dirSSH) {
		err := os.MkdirAll(dirSSH, 0700)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	SSHSrc := path.Join(dirSSH, handle.Filename)
	if utils.FileExist(SSHSrc) {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "源文件已存在,请重新命名或手动清理"})
		return
	}
	SSHDst := path.Join(dirSSH, destfile)
	if utils.FileExist(SSHDst) {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "目的文件已存在,请重新命名或手动清理"})
		return
	}
	f, err := os.OpenFile(SSHSrc, os.O_WRONLY|os.O_CREATE, 0600)
	io.Copy(f, file)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	defer f.Close()
	defer file.Close()
	os.Rename(SSHSrc, SSHDst)

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	return
}

func ConfSyncByClusterId(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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
	mod, err := repo.GetClusterById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	confsrc := path.Join(known.RootProxy, "conf/"+strconv.Itoa(int(mod.ID))+"/nginx.conf.tmp")
	confdst := path.Join(known.RootProxy, "conf/"+strconv.Itoa(int(mod.ID))+"/nginx.conf")
	confbak := path.Join(known.RootProxy, "conf/"+strconv.Itoa(int(mod.ID))+"/nginx.conf_"+strconv.Itoa(int(time.Now().Unix())))
	md5_src, _ := utils.GetMd5ByFile(confsrc)
	md5_dst, _ := utils.GetMd5ByFile(confdst)
	if md5_src != md5_dst {
		input_dst, _ := ioutil.ReadFile(confdst)
		err = ioutil.WriteFile(confbak, input_dst, 0600)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "集群配置备份失败"})
			return
		}
		input_src, _ := ioutil.ReadFile(confsrc)
		err = ioutil.WriteFile(confdst, input_src, 0600)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "集群配置生成失败"})
			return
		}
	}
	path_ssh := path.Join(known.RootProxy, ".ssh/"+mod.SSHKey)
	path_key_gl := path.Join(known.RootProxy, "ssl/global/")
	path_key_cl := path.Join(known.RootProxy, fmt.Sprintf("ssl/%d/", mod.ID))

	var content string

	for _, i := range strings.Split(mod.Backends, ",") {
		var sc = &utils.SSHClient{
			Address:    i,
			Port:       mod.SSHPort,
			User:       mod.SSHUser,
			PrivateKey: path_ssh,
		}
		var out string
		dir, file := utils.GetDir(mod.PathKey)
		out, err = sc.FileSync(fmt.Sprintf("%s/", path_key_gl), dir, file)
		if err != nil {
			repo.UpdateClusterStatusById(info.ID, "同步失败")
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Host:" + i + "\n" + out + err.Error()})
			return
		}
		content = content + "[sync Global SSLFile:\"" + mod.PathKey + "\"]"

		out, err = sc.FileSync(fmt.Sprintf("%s/", path_key_cl), dir, file)
		if err != nil {
			repo.UpdateClusterStatusById(info.ID, "同步失败")
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Host:" + i + "\n" + out + err.Error()})
			return
		}
		content = content + "[sync Cluster SSLFile:\"" + mod.PathKey + "\"]"

		dir, file = utils.GetDir(mod.PathConf)
		out, err = sc.FileSync(confdst, dir, file)
		if err != nil {
			repo.UpdateClusterStatusById(info.ID, "同步失败")
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Host:" + i + "\n" + out + err.Error()})
			return
		}
		content = content + "[sync Cluster ConfFile:\"" + mod.PathConf + "\"]"
	}
	repo.UpdateClusterStatusById(info.ID, "待测试")

	var message string
	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "update"
	journal.Resource = "cluster"
	journal.Content = "[update Cluster:" + content + "]"
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

func ConfTestByClusterId(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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
	mod, err := repo.GetClusterById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	type Output struct {
		IP     string
		Result string
	}

	var output Output
	var outputs []Output
	var ap = &nginx.ActionProxy{
		PCluster: *mod,
	}
	for _, i := range strings.Split(mod.Backends, ",") {
		rst, err := ap.ConfTest(i)
		if err != nil {
			repo.UpdateClusterStatusById(info.ID, "测试失败")
			output.IP = i
			output.Result = rst
			outputs = append(outputs, output)
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Host:" + output.IP + ":" + output.Result + ";" + err.Error()})
			return
		}
		output.IP = i
		output.Result = rst
		outputs = append(outputs, output)
	}
	repo.UpdateClusterStatusById(info.ID, "待加载")

	var message string
	contenttest, err := json.Marshal(outputs)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "update"
	journal.Resource = "cluster"
	journal.Content = "[test Cluster:" + string(contenttest) + "]"
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

func ConfLoadByClusterId(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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
	mod, err := repo.GetClusterById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	type Output struct {
		IP     string
		Result string
	}

	var output Output
	var outputs []Output
	var ap = &nginx.ActionProxy{
		PCluster: *mod,
	}
	for _, i := range strings.Split(mod.Backends, ",") {
		rst, err := ap.ConfLoad(i)
		if err != nil {
			repo.UpdateClusterStatusById(info.ID, "加载失败")
			output.IP = i
			output.Result = rst
			outputs = append(outputs, output)
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Host:" + output.IP + ":" + output.Result + ";" + err.Error()})
			return
		}
		output.IP = i
		output.Result = rst
		outputs = append(outputs, output)
	}
	repo.UpdateClusterStatusById(info.ID, "正常")

	var message string
	contentload, err := json.Marshal(outputs)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "update"
	journal.Resource = "cluster"
	journal.Content = "[load Cluster:" + string(contentload) + "]"
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
