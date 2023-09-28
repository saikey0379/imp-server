package route

import (
	"encoding/json"
	"fmt"
	"github.com/saikey0379/imp-server/pkg/utils"
	"regexp"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
)

func ValidateManageIp(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Ip string
	}
	info.Ip = strings.TrimSpace(info.Ip)

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.Ip)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	if !isValidate {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "IP格式不正确!", "Content": ""})
		return
	}

	modelIp, err := repo.GetManageIpByIp(info.Ip)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "未匹配到网段!"})
		return
	}

	network, err := repo.GetManageNetworkById(modelIp.NetworkID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "未匹配到网段!"})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "匹配成功", "Content": network})
}

func DeleteManageNetworkById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	mod, err := repo.GetManageNetworkById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	_, err = repo.DeleteManageNetworkById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	_, errDelete := repo.DeleteManageIpByNetworkId(info.ID)
	if errDelete != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errDelete.Error()})
		return
	}

	var message string
	contentManageNetwork, err := json.Marshal(mod)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}

	var journal model.Journal
	journal.Title = mod.Network
	journal.Operation = "delete"
	journal.Resource = "manageNetwork"
	journal.Content = "[delete ManageNetwork:" + string(contentManageNetwork) + "]"
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

func UpdateManageNetworkById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		ID          uint
		Network     string
		Netmask     string
		Gateway     string
		Vlan        string
		Trunk       string
		Bonding     string
		AccessToken string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	info.Network = strings.TrimSpace(info.Network)
	info.Netmask = strings.TrimSpace(info.Netmask)
	info.Gateway = strings.TrimSpace(info.Gateway)
	info.Vlan = strings.TrimSpace(info.Vlan)
	info.Trunk = strings.TrimSpace(info.Trunk)
	info.Bonding = strings.TrimSpace(info.Bonding)
	info.AccessToken = strings.TrimSpace(info.AccessToken)

	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}

	if info.Network == "" || info.Netmask == "" || info.Gateway == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请将信息填写完整!"})
		return
	}

	isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.Netmask)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	if !isValidate {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "掩码格式不正确!", "Content": ""})
		return
	}

	isValidateGageway, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.Gateway)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	if !isValidateGageway {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "网关格式不正确!", "Content": ""})
		return
	}

	count, err := repo.CountManageNetworkByNetworkAndId(info.Network, info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该网段已存在!"})
		return
	}

	mod, err := repo.GetManageNetworkById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	updatebool := false
	var content string
	var message string

	info.Netmask = strings.TrimSpace(info.Netmask)
	info.Gateway = strings.TrimSpace(info.Gateway)
	info.Vlan = strings.TrimSpace(info.Vlan)
	info.Trunk = strings.TrimSpace(info.Trunk)
	info.Bonding = strings.TrimSpace(info.Bonding)

	if info.Netmask != mod.Netmask {
		updatebool = true
		content = content + "[update Netmask:\"" + mod.Netmask + "\" to \"" + info.Netmask + "\"]"
	}

	if info.Gateway != mod.Gateway {
		updatebool = true
		content = content + "[update Gateway:\"" + mod.Gateway + "\" to \"" + info.Gateway + "\"]"
	}

	if info.Vlan != mod.Vlan {
		updatebool = true
		content = content + "[update Vlan:\"" + mod.Vlan + "\" to \"" + info.Vlan + "\"]"
	}

	if info.Trunk != mod.Trunk {
		updatebool = true
		content = content + "[update Trunk:\"" + mod.Trunk + "\" to \"" + info.Trunk + "\"]"
	}

	if info.Bonding != mod.Bonding {
		updatebool = true
		content = content + "[update Bonding:\"" + mod.Bonding + "\" to \"" + info.Bonding + "\"]"
	}

	if !updatebool {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "数据未更改"})
		return
	}

	_, err = repo.UpdateManageNetworkById(info.ID, info.Network, info.Netmask, info.Gateway, info.Vlan, info.Trunk, info.Bonding)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	//网段发生更改的情况下，重新分配IP
	if mod.Network != info.Network {
		//处理网段
		network, err := utils.GetCidrInfo(info.Network)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		ipList, err := utils.GetIPListByMinAndMaxIP(network["MinIP"], network["MaxIP"])
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		_, errDelete := repo.DeleteManageIpByNetworkId(info.ID)
		if errDelete != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errDelete.Error()})
			return
		}
		for _, ip := range ipList {
			_, err := repo.AddManageIp(info.ID, ip)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
	}

	var journal model.Journal
	journal.Title = mod.Network
	journal.Operation = "update"
	journal.Resource = "manageNetwork"
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

func GetManageNetworkById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	mod, err := repo.GetManageNetworkById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": mod})
}

func GetManageNetworkList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	mods, err := repo.GetManageNetworkListWithPage(info.Limit, info.Offset)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result := make(map[string]interface{})
	result["list"] = mods

	//总条数
	count, err := repo.CountManageNetwork()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

// 添加
func AddManageNetwork(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		Network     string
		Netmask     string
		Gateway     string
		Vlan        string
		Trunk       string
		Bonding     string
		AccessToken string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}

	info.Network = strings.TrimSpace(info.Network)
	info.Netmask = strings.TrimSpace(info.Netmask)
	info.Gateway = strings.TrimSpace(info.Gateway)
	info.Vlan = strings.TrimSpace(info.Vlan)
	info.Trunk = strings.TrimSpace(info.Trunk)
	info.Bonding = strings.TrimSpace(info.Bonding)
	info.AccessToken = strings.TrimSpace(info.AccessToken)

	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}

	if info.Network == "" || info.Netmask == "" || info.Gateway == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请将信息填写完整!"})
		return
	}

	isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.Netmask)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	if !isValidate {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "掩码格式不正确!", "Content": ""})
		return
	}

	isValidateGageway, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.Gateway)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	if !isValidateGageway {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "网关格式不正确!", "Content": ""})
		return
	}

	//处理网段
	network, err := utils.GetCidrInfo(info.Network)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	ipList, err := utils.GetIPListByMinAndMaxIP(network["MinIP"], network["MaxIP"])
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	count, err := repo.CountManageNetworkByNetwork(info.Network)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该网段已存在!"})
		return
	}

	mod, errAdd := repo.AddManageNetwork(info.Network, info.Netmask, info.Gateway, info.Vlan, info.Trunk, info.Bonding)

	_, errDelete := repo.DeleteManageIpByNetworkId(mod.ID)
	if errDelete != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errDelete.Error()})
		return
	}
	for _, ip := range ipList {
		_, err := repo.AddManageIp(mod.ID, ip)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}

	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}

	var message string
	contentManageNetwork, err := json.Marshal(mod)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}

	var journal model.Journal
	journal.Title = mod.Network
	journal.Operation = "add"
	journal.Resource = "manageNetwork"
	journal.Content = "[add ManageNetwork:" + string(contentManageNetwork) + "]"
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
