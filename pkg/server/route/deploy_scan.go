package route

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/utils"
)

func GetScanDeviceList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Limit               uint
		Offset              uint
		Keyword             string
		Company             string
		Product             string
		ModelName           string
		CpuRule             string
		Cpu                 string
		MemoryRule          string
		Memory              string
		DiskRule            string
		Disk                string
		Gpu                 string
		UserID              uint
		IsShowEnteredDevice string
		IsShowActiveDevice  string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	info.Keyword = strings.TrimSpace(info.Keyword)
	info.Company = strings.TrimSpace(info.Company)
	info.Product = strings.TrimSpace(info.Product)
	info.ModelName = strings.TrimSpace(info.ModelName)
	info.CpuRule = strings.TrimSpace(info.CpuRule)
	info.Cpu = strings.TrimSpace(info.Cpu)
	info.MemoryRule = strings.TrimSpace(info.MemoryRule)
	info.Memory = strings.TrimSpace(info.Memory)
	info.DiskRule = strings.TrimSpace(info.DiskRule)
	info.Disk = strings.TrimSpace(info.Disk)
	info.Gpu = strings.TrimSpace(info.Gpu)
	info.IsShowEnteredDevice = strings.TrimSpace(info.IsShowEnteredDevice)
	var where string
	if info.IsShowEnteredDevice == "No" || info.IsShowEnteredDevice == "" {
		where = " and t1.is_show_in_scan_list = 'Yes' "
	}

	if info.IsShowActiveDevice == "Yes" {
		where += " and (UNIX_TIMESTAMP(now()) - UNIX_TIMESTAMP(t1.last_active_time)) <= 1800 "
	}

	if info.UserID > uint(0) {
		where += " and t1.user_id = '" + fmt.Sprintf("%d", info.UserID) + "'"
	}

	if info.Company != "" {
		where += " and t1.company = '" + info.Company + "'"
	}
	if info.Product != "" {
		where += " and t1.product = '" + info.Product + "'"
	}
	if info.ModelName != "" {
		where += " and t1.model_name = '" + info.ModelName + "'"
	}
	if info.CpuRule != "" && info.Cpu != "" {
		where += " and t1.cpu_sum " + info.CpuRule + info.Cpu
	}
	if info.MemoryRule != "" && info.Memory != "" {
		where += " and t1.memory_sum " + info.MemoryRule + info.Memory
	}
	if info.DiskRule != "" && info.Disk != "" {
		where += " and t1.disk_sum " + info.DiskRule + info.Disk
	}

	if info.Keyword != "" {
		where += " and ( "
		info.Keyword = strings.Replace(info.Keyword, "\n", ",", -1)
		info.Keyword = strings.Replace(info.Keyword, " ", ",", -1)
		info.Keyword = strings.Replace(info.Keyword, ";", ",", -1)
		list := strings.Split(info.Keyword, ",")
		for k, v := range list {
			var str string
			v = strings.TrimSpace(v)
			if k == 0 {
				str = ""
			} else {
				str = " or "
			}
			where += str + " t1.sn = '" + v + "' or t1.ip = '" + v + "' or t1.company = '" + v + "' or t1.product = '" + v + "' or t1.model_name = '" + v + "'"
		}
		isValidate, _ := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.Keyword)
		if isValidate {
			where += " or t1.nic like '%%\"" + info.Keyword + "\"%%' "
		}
		where += " ) "
	}

	mods, err := repo.GetManufacturerListWithPage(info.Limit, info.Offset, where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result := make(map[string]interface{})
	result["list"] = mods

	var noDataKeywords []string
	if info.Keyword != "" {
		list := strings.Split(info.Keyword, ",")
		if len(list) > 1 {
			for _, v := range list {
				v = strings.TrimSpace(v)
				var isFind = false
				for _, device := range mods {
					if device.Sn == v {
						isFind = true
					}
				}
				if isFind == false {
					noDataKeywords = append(noDataKeywords, v)
				}
			}
		}
	}
	result["NoDataKeyword"] = noDataKeywords

	//总条数
	count, err := repo.CountManufacturerByWhere(where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func ExportScanDeviceList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Keyword    string
		Company    string
		Product    string
		ModelName  string
		CpuRule    string
		Cpu        string
		MemoryRule string
		Memory     string
		DiskRule   string
		Disk       string
		Gpu        string
		UserID     string
	}

	info.Keyword = r.FormValue("Keyword")
	info.UserID = r.FormValue("UserID")
	info.Company = r.FormValue("Company")
	info.Product = r.FormValue("Product")
	info.ModelName = r.FormValue("ModelName")
	info.CpuRule = r.FormValue("CpuRule")
	info.Cpu = r.FormValue("Cpu")
	info.MemoryRule = r.FormValue("MemoryRule")
	info.Memory = r.FormValue("Memory")
	info.DiskRule = r.FormValue("DiskRule")
	info.Disk = r.FormValue("Disk")
	info.Gpu = r.FormValue("Gpu")

	info.UserID = strings.TrimSpace(info.UserID)
	info.Keyword = strings.TrimSpace(info.Keyword)
	info.Company = strings.TrimSpace(info.Company)
	info.Product = strings.TrimSpace(info.Product)
	info.ModelName = strings.TrimSpace(info.ModelName)
	info.CpuRule = strings.TrimSpace(info.CpuRule)
	info.Cpu = strings.TrimSpace(info.Cpu)
	info.MemoryRule = strings.TrimSpace(info.MemoryRule)
	info.Memory = strings.TrimSpace(info.Memory)
	info.DiskRule = strings.TrimSpace(info.DiskRule)
	info.Disk = strings.TrimSpace(info.Disk)
	info.Gpu = strings.TrimSpace(info.Gpu)

	var where string
	where = " and t1.is_show_in_scan_list = 'Yes' "

	if info.UserID != "" {
		var userID int
		userID, _ = strconv.Atoi(info.UserID)
		where += " and t1.user_id = '" + fmt.Sprintf("%d", userID) + "'"
	}

	idsParam := r.FormValue("ids")
	if idsParam != "" {
		ids := strings.Split(idsParam, ",")
		if len(ids) > 0 {
			where += " and t1.id in (" + strings.Join(ids, ",") + ")"
		}
	}

	if info.Company != "" {
		where += " and t1.company = '" + info.Company + "'"
	}
	if info.Product != "" {
		where += " and t1.product = '" + info.Product + "'"
	}
	if info.ModelName != "" {
		where += " and t1.model_name = '" + info.ModelName + "'"
	}
	if info.CpuRule != "" && info.Cpu != "" {
		where += " and t1.cpu_sum " + info.CpuRule + info.Cpu
	}
	if info.MemoryRule != "" && info.Memory != "" {
		where += " and t1.memory_sum " + info.MemoryRule + info.Memory
	}
	if info.DiskRule != "" && info.Disk != "" {
		where += " and t1.disk_sum " + info.DiskRule + info.Disk
	}

	if info.Keyword != "" {
		where += " and ( "
		info.Keyword = strings.Replace(info.Keyword, "\n", ",", -1)
		info.Keyword = strings.Replace(info.Keyword, ";", ",", -1)
		list := strings.Split(info.Keyword, ",")
		for k, v := range list {
			var str string
			v = strings.TrimSpace(v)
			if k == 0 {
				str = ""
			} else {
				str = " or "
			}
			where += str + " t1.sn = '" + v + "' or t1.ip = '" + v + "' or t1.company = '" + v + "' or t1.product = '" + v + "' or t1.model_name = '" + v + "'"
		}
		isValidate, _ := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.Keyword)
		if isValidate {
			where += " or t1.nic like '%%\"" + info.Keyword + "\"%%' "
		}
		where += " ) "
	}

	mods, err := repo.GetManufacturerListWithPage(1000000, 0, where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var str string
	var strTitle string
	strTitle = "SN(必填),主机名(必填),IP(必填),操作系统(必填),硬件配置模板,系统安装模板(必填),位置(必填),财编,管理IP,是否支持安装虚拟机(Yes或No)\n"
	for _, device := range mods {
		str += device.Sn + ","
		str += ","
		str += ","
		str += ","
		str += ","
		str += ","
		str += ","
		str += ","
		str += ","
		str += "\n"
	}

	// cd, err := iconv.Open("gbk", "utf-8") // convert utf-8 to gbk
	// if err != nil {
	// 	w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
	// 	return
	// }
	// defer cd.Close()
	gbkStr := utils.UTF82GBK(strTitle)

	bytes := []byte(gbkStr + str)

	filename := "idcos-osinstall-scan-device.csv"
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename='%s';filename*=utf-8''%s", filename, filename))
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Write(bytes)
}

func GetScanDeviceById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	mod, err := repo.GetManufacturerById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	type DeviceWithTime struct {
		ID          uint
		DeviceID    uint
		Company     string
		Product     string
		ModelName   string
		Sn          string
		Ip          string
		Mac         string
		Nic         string
		Cpu         string
		CpuSum      uint
		Memory      string
		MemorySum   uint
		Disk        string
		DiskSum     uint
		Gpu         string
		Motherboard string
		Raid        string
		Oob         string
		IsVm        string
		NicDevice   string
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
	}

	var device DeviceWithTime
	device.ID = mod.ID
	device.DeviceID = mod.DeviceID
	device.Company = mod.Company
	device.Product = mod.Product
	device.ModelName = mod.ModelName
	device.Sn = mod.Sn
	device.Ip = mod.Ip
	device.Mac = mod.Mac
	device.Nic = mod.Nic
	device.Cpu = mod.Cpu
	device.CpuSum = mod.CpuSum
	device.Memory = mod.Memory
	device.MemorySum = mod.MemorySum
	device.Disk = mod.Disk
	device.DiskSum = mod.DiskSum
	device.Gpu = mod.Gpu
	device.Motherboard = mod.Motherboard
	device.Raid = strings.Replace(mod.Raid, "\n", "<br>", -1)
	device.Oob = mod.Oob
	device.NicDevice = strings.Replace(mod.NicDevice, "\n", "<br>", -1)

	device.CreatedAt = utils.ISOTime(mod.CreatedAt)
	device.UpdatedAt = utils.ISOTime(mod.UpdatedAt)

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": device})
}

func GetScanDeviceCompany(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var where string
	where = "device_id = 0"
	mod, err := repo.GetManufacturerCompanyByGroup(where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": mod})
}

func GetScanDeviceProduct(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		Company string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}
	info.Company = strings.TrimSpace(info.Company)

	var where string
	where = "device_id = 0 and company = '" + info.Company + "'"
	mod, err := repo.GetManufacturerProductByGroup(where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": mod})
}

func GetScanDeviceModelName(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		Company string
		Product string
		UserID  uint
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}
	info.Company = strings.TrimSpace(info.Company)
	info.Product = strings.TrimSpace(info.Product)

	var where string
	where = "device_id = 0 and company = '" + info.Company + "'"
	if info.Product != "" {
		where += " and product = '" + info.Product + "'"
	}
	if info.UserID > uint(0) {
		where += " and user_id = '" + fmt.Sprintf("%d", info.UserID) + "'"
	}
	mod, err := repo.GetManufacturerModelNameByGroup(where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": mod})
}

func BatchDeleteScanDevice(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	session, err := GetSession(w, r)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	var infos []struct {
		ID          uint
		AccessToken string
		UserID      uint
	}

	if err := r.DecodeJSONPayload(&infos); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}

	for _, info := range infos {
		if session.ID <= uint(0) {
			accessTokenUser, errAccessToken := VerifyAccessToken(info.AccessToken, ctx, false)
			if errAccessToken != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAccessToken.Error()})
				return
			}
			info.UserID = accessTokenUser.ID
			session.ID = accessTokenUser.ID
			session.Role = accessTokenUser.Role
		} else {
			info.UserID = session.ID
		}

		device, errInfo := repo.GetManufacturerById(info.ID)
		if errInfo != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errInfo.Error()})
			return
		}

		if session.Role != "Administrator" && device.UserID != info.UserID {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "您无权操作其他人的设备!"})
			return
		}

		_, errDevice := repo.UpdateManufacturerIsShowInScanListById(info.ID, "No")
		if errDevice != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errDevice.Error()})
			return
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}
