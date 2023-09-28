package route

import (
	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
)

func GetDeviceQueryTermsList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Pid       uint
		SelectPid uint
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	result := make(map[string]interface{})
	//localist
	//var initContent []map[string]interface{}
	localist, err := repo.FormatLocationToTreeByPid(info.Pid, nil, 0, info.SelectPid)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	result["localist"] = localist
	//sysclist
	sysclist, err := repo.GetSystemConfigList()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["sysclist"] = sysclist

	hwclist, err := repo.GetHardwareList()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	result["hwclist"] = hwclist

	modellist, err := repo.GetCompanyModelList()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	result["modellist"] = modellist
	//cpusumlist
	cpusumlist, err := repo.GetCpuSumList()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["cpusumlist"] = cpusumlist
	//memsumlist
	memsumlist, err := repo.GetMemorySumList()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	result["memsumlist"] = memsumlist
	//agtlist
	agtlist, err := repo.GetAgentVersionList()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	result["agtlist"] = agtlist

	//result
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetManuModelNameByCompany(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		Company     string
		IsSystemAdd string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	where := " company = '" + info.Company + "'"

	if info.IsSystemAdd != "" {
		where += " and is_system_add = '" + info.IsSystemAdd + "'"
	}

	mod, err := repo.GetManuModelNameByCompany(where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": mod})
}
