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
)

// 上报厂商信息
func ReportSysInfo(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "application/json; charset=utf-8")
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	type NicInfo struct {
		Name string
		Mac  string
		Ip   string
	}
	type NicDevice struct {
		Id    string
		Model string
	}
	type CpuInfo struct {
		Id    string
		Model string
		Core  string
	}
	type DiskInfo struct {
		Name string
		Size string
	}
	type MemoryInfo struct {
		Name string
		Size string
		Type string
	}
	type GpuInfo struct {
		Id     string
		Model  string
		Memory string
	}
	type MotherboardInfo struct {
		Name  string
		Model string
	}

	var infoFull struct {
		Sn               string
		Company          string
		Product          string
		ModelName        string
		Ip               string
		Mac              string
		Nic              []NicInfo
		Cpu              []CpuInfo
		CpuSum           uint
		Memory           []MemoryInfo
		MemorySum        uint
		Disk             []DiskInfo
		DiskSum          uint
		Gpu              []GpuInfo
		Motherboard      MotherboardInfo
		Raid             string
		Oob              string
		DeviceID         uint
		IsVm             string
		NicDevice        []NicDevice
		VersionAgt       string
		IsShowInScanList string
	}

	var info struct {
		Sn               string
		Company          string
		Product          string
		ModelName        string
		Ip               string
		Mac              string
		Nic              string
		Cpu              string
		CpuSum           uint
		Memory           string
		MemorySum        uint
		Disk             string
		DiskSum          uint
		Gpu              string
		Motherboard      string
		Raid             string
		Oob              string
		DeviceID         uint
		IsVm             string
		NicDevice        string
		VersionAgt       string
		IsShowInScanList string
	}

	if err := r.DecodeJSONPayload(&infoFull); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	infoFull.Sn = strings.TrimSpace(infoFull.Sn)
	infoFull.Company = strings.TrimSpace(infoFull.Company)
	infoFull.Product = strings.TrimSpace(infoFull.Product)
	infoFull.ModelName = strings.TrimSpace(infoFull.ModelName)
	infoFull.IsVm = strings.TrimSpace(infoFull.IsVm)

	info.Sn = infoFull.Sn
	info.Company = infoFull.Company
	info.Product = infoFull.Product
	info.ModelName = infoFull.ModelName
	info.Ip = infoFull.Ip
	info.Mac = infoFull.Mac
	info.Raid = infoFull.Raid
	info.Oob = infoFull.Oob
	info.DeviceID = infoFull.DeviceID
	info.CpuSum = infoFull.CpuSum
	info.MemorySum = infoFull.MemorySum
	info.DiskSum = infoFull.DiskSum
	info.IsVm = infoFull.IsVm
	info.VersionAgt = infoFull.VersionAgt
	countDevice, err := repo.CountDeviceBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	//已经录入过的设备，不要再次显示到发现新设备列表 from chenli@20161117
	if countDevice > 0 {
		info.IsShowInScanList = "No"
	} else {
		info.IsShowInScanList = "Yes"
	}

	if info.IsVm != "Yes" {
		info.IsVm = "No"
	}

	//nic
	nic, err := json.Marshal(infoFull.Nic)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	info.Nic = string(nic)

	//nicDevice
	nicDevice, err := json.Marshal(infoFull.NicDevice)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	info.NicDevice = string(nicDevice)

	//bootos ip
	for _, nicInfo := range infoFull.Nic {
		nicInfo.Ip = strings.TrimSpace(nicInfo.Ip)
		if nicInfo.Ip != "" {
			info.Ip = nicInfo.Ip
			break
		}
	}

	//cpu
	cpu, err := json.Marshal(infoFull.Cpu)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	info.Cpu = string(cpu)

	//memory
	memory, err := json.Marshal(infoFull.Memory)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	info.Memory = string(memory)

	//disk
	disk, err := json.Marshal(infoFull.Disk)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	info.Disk = string(disk)

	//gpu
	gpu, err := json.Marshal(infoFull.Gpu)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	info.Gpu = string(gpu)

	//motherboard
	motherboard, err := json.Marshal(infoFull.Motherboard)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	info.Motherboard = string(motherboard)

	if info.Sn == "" || info.Company == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN和厂商名称不能为空!"})
		return
	}

	count, err := repo.CountManufacturerBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var lastActiveTime = time.Now().Format("2006-01-02 15:04:05")
	if count > 0 {
		_, errUpdate := repo.UpdateManufacturerBySn(info.Company, info.Product, info.ModelName, info.Sn, info.Ip, info.Mac, info.Nic, info.Cpu, info.CpuSum, info.Memory, info.MemorySum, info.Disk, info.DiskSum, info.Gpu, info.Motherboard, info.Raid, info.Oob, info.NicDevice, info.VersionAgt, info.IsShowInScanList, lastActiveTime)
		if errUpdate != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
			return
		}
	} else {
		_, err := repo.AddManufacturer(info.DeviceID, info.Company, info.Product, info.ModelName, info.Sn, info.Ip, info.Mac, info.Nic, info.Cpu, info.CpuSum, info.Memory, info.MemorySum, info.Disk, info.DiskSum, info.Gpu, info.Motherboard, info.Raid, info.Oob, info.NicDevice, info.VersionAgt, info.IsShowInScanList, lastActiveTime)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}

// UpdateSysTimestamp
func ReportSysTimestamp(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "application/json; charset=utf-8")
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	logger, ok := middleware.LoggerFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	redis, okRedis := middleware.RedisFromContext(ctx)
	if !okRedis {
		logger.Errorf("ERROR: REDIS Unavaiablle[%s]", okRedis)
	}

	var info struct {
		Sn string
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	info.Sn = strings.TrimSpace(info.Sn)

	var repoBool bool
	if okRedis {
		keyMatch := fmt.Sprintf("IMP_CABINET_*_%s", info.Sn)
		keys, err := redis.Keys(keyMatch)
		if err != nil {
			logger.Errorf("ERROR: REDIS GET KEY(%s) [%s]", keyMatch, err.Error())
		}
		if len(keys) == 1 {
			_, err = redis.Set(keys[0], time.Now().Format("2006-01-02 15:04:05"))
			if err != nil {
				logger.Errorf("ERROR: REDIS SET KEY(%s) [%s]", keys[0], err.Error())
				repoBool = true
			}
		} else {
			if len(keys) > 1 {
				for _, key := range keys {
					_, err = redis.Del(key)
					if err != nil {
						logger.Errorf("ERROR: REDIS DELETE KEY(%s) [%s]", key, err.Error())
					}
				}
			}
			repoBool = true
		}
	}

	if repoBool {
		var lastActiveTime = time.Now().Format("2006-01-02 15:04:05")
		_, errUpdate := repo.UpdateManufacturerLastActiveTimeBySn(info.Sn, lastActiveTime)
		if errUpdate != nil {
			logger.Errorf(fmt.Sprintf("ERROR: UpdateManufacturerLastActiveTimeBySn %s", errUpdate.Error()))
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "设备时间更新失败"})
			return
		}
		if okRedis {
			modDevice, err := repo.GetDeviceBySn(info.Sn)
			if err != nil {
				logger.Errorf(fmt.Sprintf("ERROR: GetDeviceBySn %s", err.Error()))
			}

			modLocation, err := repo.GetLocationById(modDevice.LocationID)
			if err != nil {
				logger.Errorf(fmt.Sprintf("ERROR: GetDeviceBySn %s", err.Error()))
			}

			locaAry := strings.Split(modDevice.LocationU, ",")
			locaUF, err := strconv.Atoi(locaAry[0])
			if err != nil {
				logger.Errorf(fmt.Sprintf("ERROR: device.LocationU(%s)[%s]", modDevice.LocationU, err.Error()))
			}

			var key = fmt.Sprintf("IMP_CABINET_%s_%s_%s_%s_%s_%s_%s", modLocation.Pid, modLocation.ID, (locaUF-1)*12+13, len(locaAry), modDevice.Ip, modDevice.Hostname, modDevice.Sn)
			_, err = redis.Set(key, lastActiveTime)
			if err != nil {
				logger.Errorf(fmt.Sprintf("ERROR: REDIS SET K:V(%s：%s)[%s]", key, lastActiveTime, err.Error()))
			}
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}

// 查询安装信息
func GetDevicePrepareInstallInfo(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "application/json; charset=utf-8")
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		Sn        string
		Company   string
		Product   string
		ModelName string
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	info.Sn = strings.TrimSpace(info.Sn)
	info.Company = strings.TrimSpace(info.Company)
	info.Product = strings.TrimSpace(info.Product)
	info.ModelName = strings.TrimSpace(info.ModelName)

	if info.Sn == "" || info.Company == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN及厂商信息不能为空!"})
		return
	}

	result := make(map[string]string)
	//校验是否在配置库
	isValidate, err := repo.ValidateHardwareProductModel(info.Company, info.Product, info.ModelName)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": result})
		return
	}
	if isValidate == true {
		result["IsVerify"] = "true"
	} else {
		result["IsVerify"] = "false"
	}

	result["IsSkipHardwareConfig"] = "false"
	//是否跳过硬件配置(用户是否配置硬件配置模板)
	if info.Sn != "" {
		count, err := repo.CountDeviceBySn(info.Sn)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": result})
			return
		}

		if count > 0 {
			device, err := repo.GetDeviceBySn(info.Sn)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": result})
				return
			}
			if device.HardwareID <= uint(0) {
				result["IsSkipHardwareConfig"] = "true"
			}
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func BatchAssignManufacturerOnwer(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var infos []struct {
		ID          uint
		UserID      uint
		AccessToken string
	}

	session, errSession := GetSession(w, r)
	if errSession != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + errSession.Error()})
		return
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
			session.ID = accessTokenUser.ID
			session.Role = accessTokenUser.Role
		}

		if session.Role != "Administrator" {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "权限不足!"})
			return
		}

		manufacturer, err := repo.GetManufacturerById(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		_, errUpdate := repo.AssignManufacturerOnwer(manufacturer.ID, info.UserID)
		if errUpdate != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
			return
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}
