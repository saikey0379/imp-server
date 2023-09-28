package route

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
	"github.com/saikey0379/imp-server/pkg/utils"
)

// 重装
func BatchReInstall(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	logger, ok := middleware.LoggerFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	conf, ok := middleware.ConfigFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
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

	var message string
	for _, info := range infos {
		accessTokenUser, errAccessToken := VerifyAccessToken(info.AccessToken, ctx, false)
		if errAccessToken != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAccessToken.Error()})
			return
		}
		info.UserID = accessTokenUser.ID
		session.ID = accessTokenUser.ID
		session.Role = accessTokenUser.Role

		//log
		device, errDevice := repo.GetDeviceById(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errDevice.Error()})
			return
		}

		if session.Role != "Administrator" && device.UserID != info.UserID {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "您无权操作其他人的设备!"})
			return
		}

		_, errUpdate := repo.ReInstallDeviceById(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
			return
		}

		var key = fmt.Sprintf("IMP_DEVICE_INSTALL_%s", device.Sn)
		redis, okRedis := middleware.RedisFromContext(ctx)
		if okRedis {
			_, err := redis.Del(key)
			if err != nil {
				logger.Errorf(fmt.Sprintf("ERROR: REDIS DELETE[%s]", key))
			}
		}

		content, err := json.Marshal(device)
		if err != nil {
			message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
		}

		var journal model.Journal
		journal.Title = device.Hostname
		journal.Operation = "reinstall"
		journal.Resource = "device"
		journal.Content = "[reinstall Device:" + string(content) + "]"
		journal.User = accessTokenUser.Username
		journal.UpdatedAt = time.Now()
		err = repo.AddJournal(journal)
		if err != nil {
			message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
		}

		//删除PXE配置文件
		macs, errMac := repo.GetMacListByDeviceID(device.ID)
		if errMac != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errMac.Error()})
			return
		}
		for _, mac := range macs {
			pxeFileName := utils.GetPxeFileNameByMac(mac.Mac)
			confDir := conf.OsInstall.PxeConfigDir
			if utils.FileExist(confDir + "/" + pxeFileName) {
				err := os.Remove(confDir + "/" + pxeFileName)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}
			}
		}
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}

// 取消安装
func BatchCancelInstall(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	conf, ok := middleware.ConfigFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
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
		device, err := repo.GetDeviceById(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

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

		if session.Role != "Administrator" && device.UserID != info.UserID {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "您无权操作其他人的设备!"})
			return
		}

		//安装成功的设备不允许取消安装
		if device.Status == "success" {
			continue
		}

		_, errCancel := repo.CancelInstallDeviceById(info.ID)
		if errCancel != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCancel.Error()})
			return
		}

		//删除PXE配置文件
		macs, errMac := repo.GetMacListByDeviceID(device.ID)
		if errMac != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errMac.Error()})
			return
		}
		for _, mac := range macs {
			pxeFileName := utils.GetPxeFileNameByMac(mac.Mac)
			confDir := conf.OsInstall.PxeConfigDir
			if utils.FileExist(confDir + "/" + pxeFileName) {
				err := os.Remove(confDir + "/" + pxeFileName)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}
			}
		}

		logContent := make(map[string]interface{})
		logContent["data"] = device
		json, err := json.Marshal(logContent)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + err.Error()})
			return
		}

		_, errAddLog := repo.AddDeviceLog(info.ID, "取消安装设备", "operate", string(json))
		if errAddLog != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAddLog.Error()})
			return
		}

		_, errLog := repo.UpdateDeviceLogTypeByDeviceIdAndType(info.ID, "install", "install_history")
		if errLog != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errLog.Error()})
			return
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}

// 设备上线
func BatchOnline(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	_, ok = middleware.ConfigFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
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

		//log
		device, errDevice := repo.GetDeviceById(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errDevice.Error()})
			return
		}

		if session.Role != "Administrator" && device.UserID != info.UserID {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "您无权操作其他人的设备!"})
			return
		}

		_, errUpdate := repo.OnlineDeviceById(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
			return
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}

// 设备下线
func BatchOffline(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	_, ok = middleware.ConfigFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
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

		//log
		device, errDevice := repo.GetDeviceById(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errDevice.Error()})
			return
		}

		if session.Role != "Administrator" && device.UserID != info.UserID {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "您无权操作其他人的设备!"})
			return
		}

		_, errUpdate := repo.OfflineDeviceById(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
			return
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}

func BatchDelete(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	conf, ok := middleware.ConfigFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
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

	var message string

	for _, info := range infos {
		accessTokenUser, errAccessToken := VerifyAccessToken(info.AccessToken, ctx, false)
		if errAccessToken != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAccessToken.Error()})
			return
		}
		info.UserID = accessTokenUser.ID
		session.ID = accessTokenUser.ID
		session.Role = accessTokenUser.Role

		device, errInfo := repo.GetDeviceById(info.ID)
		if errInfo != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errInfo.Error()})
			return
		}

		if session.Role != "Administrator" && device.UserID != info.UserID {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "您无权操作其他人的设备!"})
			return
		}

		//删除PXE配置文件
		macs, errMac := repo.GetMacListByDeviceID(device.ID)
		if errMac != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "GetMac" + errMac.Error()})
			return
		}
		for _, mac := range macs {
			pxeFileName := utils.GetPxeFileNameByMac(mac.Mac)
			confDir := conf.OsInstall.PxeConfigDir
			if utils.FileExist(confDir + "/" + pxeFileName) {
				err := os.Remove(confDir + "/" + pxeFileName)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "PxeFile" + err.Error()})
					return
				}
			}
		}

		//删除mac
		_, err := repo.DeleteMacByDeviceId(info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "DeleteMac" + err.Error()})
			return
		}

		//删除设备关联的硬件信息
		_, errManufacturer := repo.DeleteManufacturerBySn(device.Sn)
		if errManufacturer != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "DeleteManufacturer" + errManufacturer.Error()})
			return
		}

		errCopy := repo.CopyDeviceToHistory(info.ID)
		if errCopy != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "History" + errCopy.Error()})
			return
		}
		_, errUpdate := repo.UpdateHistoryDeviceStatusById(info.ID, "delete")
		if errUpdate != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "UpdateHistory" + errUpdate.Error()})
			return
		}

		_, errDevice := repo.DeleteDeviceById(info.ID)
		if errDevice != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "DeleteDevice" + errDevice.Error()})
			return
		}

		content, err := json.Marshal(device)
		if err != nil {
			message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
		}

		var journal model.Journal
		journal.Title = device.Hostname
		journal.Operation = "delete"
		journal.Resource = "device"
		journal.Content = "[delete Device:" + string(content) + "]"
		journal.User = accessTokenUser.Username
		journal.UpdatedAt = time.Now()
		err = repo.AddJournal(journal)
		if err != nil {
			message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
		}

		//log
		logContent := make(map[string]interface{})
		logContent["data"] = device
		json, err := json.Marshal(logContent)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + err.Error()})
			return
		}

		_, errAddLog := repo.AddDeviceLog(device.ID, "删除设备信息", "operate", string(json))
		if errAddLog != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "AddDevice" + errAddLog.Error()})
			return
		}

		//callback
		_, errCallback := repo.DeleteDeviceInstallCallbackByDeviceID(info.ID)
		if errCallback != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Callback" + errCallback.Error()})
			return
		}
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}

func GetDeviceById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	mod, err := repo.GetDeviceById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": mod})
}

func GetDeviceBySn(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var info struct {
		Sn string
	}
	info.Sn = r.FormValue("sn")
	info.Sn = strings.TrimSpace(info.Sn)

	count, err := repo.CountDeviceBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	if count <= 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "设备不存在!"})
		return
	}

	mod, err := repo.GetDeviceBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": mod})
}

func GetFullDeviceBySn(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Sn string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	mod, err := repo.GetFullDeviceBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	type Task struct {
		ID   uint
		Name string
	}

	type DeviceWithTime struct {
		ID                uint
		BatchNumber       string
		Sn                string
		Hostname          string
		Ip                string
		ManageIp          string
		NetworkID         uint
		ManageNetworkID   uint
		OsID              uint
		HardwareID        uint
		SystemID          uint
		LocationID        uint
		LocationU         string
		AssetNumber       string
		OverDate          string
		Status            string
		InstallProgress   float64
		InstallLog        string
		NetworkName       string
		ManageNetworkName string
		OsName            string
		HardwareName      string
		SystemName        string
		LocationName      string
		IsVm              string
		IsSupportVm       string
		Callback          string
		UserID            uint
		DevManager        string
		OpsManager        string
		DeviceLabel       string
		DeviceDescribe    string
		Tasks             []Task
		CreatedAt         utils.ISOTime
		UpdatedAt         utils.ISOTime
	}

	var device DeviceWithTime
	device.ID = mod.ID
	device.BatchNumber = mod.BatchNumber
	device.Sn = mod.Sn
	device.Hostname = mod.Hostname
	device.Ip = mod.Ip
	device.ManageIp = mod.ManageIp
	device.NetworkID = mod.NetworkID
	device.ManageNetworkID = mod.ManageNetworkID
	device.OsID = mod.OsID
	device.HardwareID = mod.HardwareID
	device.SystemID = mod.SystemID
	device.LocationID = mod.LocationID
	device.LocationU = mod.LocationU
	device.AssetNumber = mod.AssetNumber
	device.OverDate = mod.OverDate
	device.Status = mod.Status
	device.InstallProgress = mod.InstallProgress
	device.InstallLog = mod.InstallLog
	device.NetworkName = mod.NetworkName
	device.ManageNetworkName = mod.ManageNetworkName
	device.OsName = mod.OsName
	device.HardwareName = mod.HardwareName
	device.SystemName = mod.SystemName
	device.IsSupportVm = mod.IsSupportVm
	device.UserID = mod.UserID
	device.DevManager = mod.DevManager
	device.OpsManager = mod.OpsManager
	device.DeviceLabel = mod.DeviceLabel
	device.DeviceDescribe = mod.DeviceDescribe

	device.LocationName, err = repo.FormatLocationNameById(mod.LocationID, "", "-")
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var tasks []Task
	mods, err := repo.GetTaskListBySn(info.Sn)
	for _, v := range mods {
		var task Task
		task.ID = v.ID
		task.Name = v.Name
		tasks = append(tasks, task)
	}
	device.Tasks = tasks

	countCallback, errCount := repo.CountDeviceInstallCallbackByDeviceIDAndType(device.ID, "after_install")
	if errCount != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCount.Error()})
		return
	}
	if countCallback > 0 {
		callback, err := repo.GetDeviceInstallCallbackByDeviceIDAndType(device.ID, "after_install")
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		device.Callback = callback.Content
	}

	device.CreatedAt = utils.ISOTime(mod.CreatedAt)
	device.UpdatedAt = utils.ISOTime(mod.UpdatedAt)

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": device})
}

func GetFullDeviceById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	mod, err := repo.GetFullDeviceById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	type DeviceWithTime struct {
		ID                uint
		BatchNumber       string
		Sn                string
		Hostname          string
		Ip                string
		ManageIp          string
		NetworkID         uint
		ManageNetworkID   uint
		OsID              uint
		HardwareID        uint
		SystemID          uint
		LocationID        uint
		LocationU         string
		AssetNumber       string
		Status            string
		InstallProgress   float64
		InstallLog        string
		NetworkName       string
		ManageNetworkName string
		OsName            string
		HardwareName      string
		SystemName        string
		LocationName      string
		IsVm              string
		IsSupportVm       string
		Callback          string
		UserID            uint
		DevManager        string
		OpsManager        string
		DeviceLabel       string
		DeviceDescribe    string
		CreatedAt         utils.ISOTime
		UpdatedAt         utils.ISOTime
	}

	var device DeviceWithTime
	device.ID = mod.ID
	device.BatchNumber = mod.BatchNumber
	device.Sn = mod.Sn
	device.Hostname = mod.Hostname
	device.Ip = mod.Ip
	device.ManageIp = mod.ManageIp
	device.NetworkID = mod.NetworkID
	device.ManageNetworkID = mod.ManageNetworkID
	device.OsID = mod.OsID
	device.HardwareID = mod.HardwareID
	device.SystemID = mod.SystemID
	device.LocationID = mod.LocationID
	device.LocationU = mod.LocationU
	device.AssetNumber = mod.AssetNumber
	device.Status = mod.Status
	device.InstallProgress = mod.InstallProgress
	device.InstallLog = mod.InstallLog
	device.NetworkName = mod.NetworkName
	device.ManageNetworkName = mod.ManageNetworkName
	device.OsName = mod.OsName
	device.HardwareName = mod.HardwareName
	device.SystemName = mod.SystemName
	device.IsSupportVm = mod.IsSupportVm
	device.UserID = mod.UserID
	device.DevManager = mod.DevManager
	device.OpsManager = mod.OpsManager
	device.DeviceLabel = mod.DeviceLabel
	device.DeviceDescribe = mod.DeviceDescribe

	device.LocationName, err = repo.FormatLocationNameById(mod.LocationID, "", "-")
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	countCallback, errCount := repo.CountDeviceInstallCallbackByDeviceIDAndType(info.ID, "after_install")
	if errCount != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCount.Error()})
		return
	}
	if countCallback > 0 {
		callback, err := repo.GetDeviceInstallCallbackByDeviceIDAndType(info.ID, "after_install")
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		device.Callback = callback.Content
	}

	device.CreatedAt = utils.ISOTime(mod.CreatedAt)
	device.UpdatedAt = utils.ISOTime(mod.UpdatedAt)

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": device})
}

func GetFullDeviceHWBySn(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误", "Content": nil})
		return
	}
	var info struct {
		Sn string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error(), "Content": nil})
		return
	}

	mod, err := repo.GetManufacturerBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": nil})
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
		VersionAgt  string
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
	device.VersionAgt = mod.VersionAgt
	device.CreatedAt = utils.ISOTime(mod.CreatedAt)
	device.UpdatedAt = utils.ISOTime(mod.UpdatedAt)

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": device})
}

func GetDeviceList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	logger, _ := middleware.LoggerFromContext(ctx)

	var info struct {
		Limit          uint
		Offset         uint
		Keyword        string
		LocationID     int
		CompanyHwc     string
		HardwareID     int
		SystemID       int
		Status         string
		BatchNumber    string
		StartUpdatedAt string
		EndUpdatedAt   string
		UserID         int
		ID             int
		CpuSum         int
		MemorySum      int
		Company        string
		ModelName      string
		VersionAgt     string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	info.Keyword = strings.TrimSpace(info.Keyword)
	info.Status = strings.TrimSpace(info.Status)
	info.BatchNumber = strings.TrimSpace(info.BatchNumber)

	var where string
	where = " where t1.id > 0 "
	if info.ID > 0 {
		where += " and t1.id = " + strconv.Itoa(info.ID)
	}
	if info.LocationID > 0 {
		where += " and t1.location_id = " + strconv.Itoa(info.LocationID)
	}
	if info.CompanyHwc != "" && info.HardwareID > 0 {
		where += " and t1.hardware_id = " + strconv.Itoa(info.HardwareID)
	}
	if info.SystemID > 0 {
		where += " and t1.system_id = " + strconv.Itoa(info.SystemID)
	}
	if info.Status == "install" {
		where += " and t1.status not like '" + "%line%" + "'"
	} else if info.Status != "" {
		where += " and t1.status = '" + info.Status + "'"
	}
	if info.BatchNumber != "" {
		where += " and t1.batch_number = '" + info.BatchNumber + "'"
	}

	if info.StartUpdatedAt != "" {
		where += " and t1.updated_at >= '" + info.StartUpdatedAt + "'"
	}

	if info.EndUpdatedAt != "" {
		where += " and t1.updated_at <= '" + info.EndUpdatedAt + "'"
	}

	if info.UserID > 0 {
		where += " and t1.user_id = '" + strconv.Itoa(info.UserID) + "'"
	}

	if info.CpuSum > 0 {
		where += " and t8.`cpu_sum` = '" + strconv.Itoa(info.CpuSum) + "'"
	}

	if info.MemorySum > 0 {
		where += " and t8.memory_sum = '" + strconv.Itoa(info.MemorySum) + "'"
	}

	if info.VersionAgt != "" {
		where += " and t8.version_agt = '" + info.VersionAgt + "'"
	}

	if info.Company != "" {
		where += " and t8.company = '" + info.Company + "'"

		if info.ModelName != "" {
			where += " and t8.model_name = '" + info.ModelName + "'"
		}
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
			where += str + " t1.sn like '" + "%" + v + "%" + "' or t1.batch_number like '" + "%" + v + "%" + "' or t1.hostname like '" + "%" + v + "%" + "' or t1.ip like '" + "%" + v + "%" + "' or t1.device_label like '" + "%" + v + "%" + "' or t1.device_describe like '" + "%" + v + "%'"
		}
		where += " ) "
	}

	mods, err := repo.GetDeviceListWithPage(info.Limit, info.Offset, where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	type GPU struct {
		Id     string `json:"Id"`
		Model  string `json:"Model"`
		Memory string `json:"Memory"`
	}

	type DeviceWithTime struct {
		ID              uint
		BatchNumber     string
		Sn              string
		Hostname        string
		Ip              string
		NetworkID       uint
		OsID            uint
		CpuSum          uint
		MemorySum       uint
		GpuSum          uint
		Gpu             string
		DevManager      string
		OpsManager      string
		DeviceLabel     string
		HardwareID      uint
		SystemID        uint
		ModelName       string
		LocationID      uint
		LocationU       string
		AssetNumber     string
		Status          string
		InstallProgress float64
		InstallLog      string
		NetworkName     string
		OsName          string
		HardwareName    string
		SystemName      string
		LocationName    string
		IsSupportVm     string
		UserID          uint
		OwnerName       string
		BootosIp        string
		OobIp           string
		CreatedAt       utils.ISOTime
		UpdatedAt       utils.ISOTime
	}
	var rows []DeviceWithTime
	for _, v := range mods {
		var device DeviceWithTime
		device.ID = v.ID
		device.BatchNumber = v.BatchNumber
		device.Sn = v.Sn
		device.Hostname = v.Hostname
		device.Ip = v.Ip
		device.NetworkID = v.NetworkID
		device.OsID = v.OsID
		device.CpuSum = v.CpuSum
		device.MemorySum = v.MemorySum
		device.DevManager = v.DevManager
		device.OpsManager = v.OpsManager
		device.DeviceLabel = v.DeviceLabel
		device.HardwareID = v.HardwareID
		device.SystemID = v.SystemID
		device.ModelName = v.ModelName
		device.LocationID = v.LocationID
		device.LocationU = v.LocationU
		device.AssetNumber = v.AssetNumber
		device.Status = v.Status
		device.InstallProgress = v.InstallProgress
		device.InstallLog = v.InstallLog
		device.NetworkName = v.NetworkName
		device.OsName = v.OsName
		device.HardwareName = v.HardwareName
		device.SystemName = v.SystemName
		device.IsSupportVm = v.IsSupportVm
		device.UserID = v.UserID
		device.OwnerName = v.OwnerName
		device.BootosIp = v.BootosIp
		device.OobIp = v.OobIp

		if v.Gpu != "" {
			var gpus []GPU
			if err := json.Unmarshal([]byte(strings.ReplaceAll(v.Gpu, "\\", "")), &gpus); err != nil {
				result := fmt.Sprintf("FAILURE: Unmarshal InvokeAddress[%s]", err.Error())
				logger.Errorf(result)

				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": result})
				return
			}

			var gpu string
			for _, i := range gpus {
				gpu = fmt.Sprintf("%s\n%s [%s] %s", gpu, i.Id, i.Model, i.Memory)
			}

			device.GpuSum = uint(len(gpus))
			device.Gpu = gpu
		}

		device.LocationName, err = repo.FormatLocationNameById(v.LocationID, "", "-")
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		device.CreatedAt = utils.ISOTime(v.CreatedAt)
		device.UpdatedAt = utils.ISOTime(v.UpdatedAt)

		deviceLog, _ := repo.GetLastDeviceLogByDeviceID(v.ID)
		device.InstallLog = deviceLog.Title
		rows = append(rows, device)
	}

	result := make(map[string]interface{})
	result["list"] = rows

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
	result["Status"] = info.Status

	//总条数
	count, err := repo.CountDevice(where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetDeviceNumByStatus(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Status string
		UserID int
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	info.Status = strings.TrimSpace(info.Status)

	var where string
	where = " where t1.id > 0 "
	where += " and t1.status = '" + info.Status + "'"
	if info.UserID > 0 {
		where += " and t1.user_id = " + strconv.Itoa(info.UserID)
	}

	//总条数
	count, err := repo.CountDevice(where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result := make(map[string]interface{})
	result["count"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

// 添加
func AddDevice(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	logger, _ := middleware.LoggerFromContext(ctx)

	var info struct {
		BatchNumber     string
		Sn              string
		Hostname        string
		Ip              string
		ManageIp        string
		NetworkID       uint
		ManageNetworkID uint
		OsID            uint
		HardwareID      uint
		SystemID        uint
		LocationID      uint
		LocationU       string
		AssetNumber     string
		OverDate        string
		IsSupportVm     string
		Status          string
		UserID          uint
		AccessToken     string
		Callback        string
		DevManager      string
		OpsManager      string
		DeviceLabel     string
		DeviceDescribe  string
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}

	session, err := GetSession(w, r)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	//先检测传入数据是否有问题
	info.BatchNumber = strings.TrimSpace(info.BatchNumber)
	info.Sn = strings.TrimSpace(info.Sn)
	info.Sn = strings.Replace(info.Sn, "	", "", -1)
	info.Sn = strings.Replace(info.Sn, " ", "", -1)
	info.Hostname = strings.TrimSpace(info.Hostname)
	info.Ip = strings.TrimSpace(info.Ip)
	info.ManageIp = strings.TrimSpace(info.ManageIp)
	info.AssetNumber = strings.TrimSpace(info.AssetNumber)
	info.OverDate = time.Now().Format(info.OverDate)
	info.Status = strings.TrimSpace(info.Status)
	info.AccessToken = strings.TrimSpace(info.AccessToken)
	info.Callback = strings.TrimSpace(info.Callback)
	info.UserID = session.ID

	accessTokenUser, errAccessToken := VerifyAccessToken(info.AccessToken, ctx, false)
	if errAccessToken != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAccessToken.Error()})
		return
	}
	info.UserID = accessTokenUser.ID
	session.ID = accessTokenUser.ID
	session.Role = accessTokenUser.Role
	info.IsSupportVm = strings.TrimSpace(info.IsSupportVm)
	if info.IsSupportVm == "" {
		info.IsSupportVm = "No"
	}

	if info.NetworkID == uint(0) {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "无匹配的网络信息，请确认!"})
		return
	}
	if info.Sn == "" || info.Hostname == "" || info.Ip == "" || info.NetworkID == uint(0) || info.OsID == uint(0) {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请将信息填写完整!"})
		return
	}
	//match manufacturer
	countManufacturer, errCountManufacturer := repo.CountManufacturerBySn(info.Sn)
	if errCountManufacturer != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCountManufacturer.Error()})
		return
	}
	if countManufacturer <= 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "未在【资源池管理】里匹配到该SN，请先将该设备加电并进入BootOS!"})
		return
	}
	//validate user from manufacturer
	count, err := repo.CountDeviceBySn(info.Sn)
	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "设备已存在！SN：" + info.Sn})
		return
	} else {
		count, err := repo.CountDeviceByHostname(info.Hostname)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		if count > 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Hostname + " 该主机名已存在!"})
			return
		}
		countIp, err := repo.CountDeviceByIp(info.Ip)
		if err != nil {
			logger.Errorf(fmt.Sprintf("ERROR: CountDeviceByIp[%s]%s", info.Ip, countIp, err.Error()))
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		if countIp > 0 {
			logger.Errorf(fmt.Sprintf("ERROR: CountDeviceByIp[%s:%v]", info.Ip, countIp))
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Ip + " 该IP已存在!"})
			return
		}
		if info.ManageIp != "" {
			countManageIp, err := repo.CountDeviceByManageIp(info.ManageIp)
			if err != nil {
				logger.Errorf(fmt.Sprintf("ERROR: CountDeviceByManageIp[%s]%s", info.ManageIp, err.Error()))
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			if countManageIp > 0 {
				logger.Errorf(fmt.Sprintf("ERROR: CountDeviceByManageIp[%s:%v]", info.ManageIp, countManageIp))
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + " 该管理IP已存在!"})
				return
			}
		}
	}
	//匹配网络
	isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.Ip)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	if !isValidate {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Ip + "IP格式不正确!"})
		return
	}
	modelIp, err := repo.GetIpByIp(info.Ip)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Ip + "未匹配到网段!"})
		return
	}
	_, errNetwork := repo.GetNetworkById(modelIp.NetworkID)
	if errNetwork != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Ip + "未匹配到网段!"})
		return
	}
	if info.ManageIp != "" {
		//匹配网络
		isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.ManageIp)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
			return
		}
		if !isValidate {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + "IP格式不正确!"})
			return
		}
		modelIp, err := repo.GetManageIpByIp(info.ManageIp)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + "未匹配到网段!"})
			return
		}
		_, errNetwork := repo.GetManageNetworkById(modelIp.NetworkID)
		if errNetwork != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Ip + "未匹配到网段!"})
			return
		}
	}
	//校验是否使用OOB静态IP及管理IP是否填写
	if info.HardwareID > uint(0) {
		hardware, err := repo.GetHardwareById(info.HardwareID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		if hardware.Data != "" {
			if strings.Contains(hardware.Data, "<{manage_ip}>") || strings.Contains(hardware.Data, "<{manage_netmask}>") || strings.Contains(hardware.Data, "<{manage_gateway}>") {
				if info.ManageIp == "" || info.ManageNetworkID <= uint(0) {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该设备(SN:" + info.Sn + ")使用的硬件配置模板的OOB网络类型为静态IP的方式，请填写管理IP!"})
					return
				}
			}
		}
	}

	var status = "offline"
	var message string
	device, err := repo.AddDevice("", info.Sn, info.Hostname, info.Ip, info.ManageIp, info.NetworkID, info.ManageNetworkID, info.OsID, info.HardwareID, info.SystemID, info.LocationID, info.LocationU, info.AssetNumber, status, info.IsSupportVm, info.UserID, info.DevManager, info.OpsManager, info.DeviceLabel, info.DeviceDescribe)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + err.Error()})
		return
	}
	contentDevice, err := json.Marshal(device)
	if err != nil {
		message = message + fmt.Sprintf("[Umarshal failed:%s]", err)
	}
	var journal model.Journal
	journal.Title = device.Hostname
	journal.Operation = "add"
	journal.Resource = "device"
	journal.Content = "[add Device:" + string(contentDevice) + "]"
	journal.User = accessTokenUser.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
	}
	//log
	logContent := make(map[string]interface{})
	logContent["data"] = device
	json, err := json.Marshal(logContent)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + err.Error()})
		return
	}
	_, errAddLog := repo.AddDeviceLog(device.ID, "录入新设备", "operate", string(json))
	if errAddLog != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAddLog.Error()})
		return
	}
	//init manufactures device_id
	countManufacturer, errCountManufacturer = repo.CountManufacturerBySn(info.Sn)
	if errCountManufacturer != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCountManufacturer.Error()})
		return
	}
	if countManufacturer > 0 {
		manufacturerId, errGetManufacturerBySn := repo.GetManufacturerIdBySn(info.Sn)
		if errGetManufacturerBySn != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errGetManufacturerBySn.Error()})
			return
		}
		_, errUpdate := repo.UpdateManufacturerDeviceIdById(manufacturerId, device.ID)
		if errUpdate != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
			return
		}
	}
	//callback script
	errCallback := SaveDeviceInstallCallback(ctx, device.ID, "after_install", info.Callback)
	if errCallback != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCallback.Error()})
		return
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}

func UpdateDevice(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	var info struct {
		ID              uint
		Hostname        string
		Ip              string
		ManageIp        string
		NetworkID       uint
		ManageNetworkID uint
		OsID            uint
		HardwareID      uint
		SystemID        uint
		LocationID      uint
		LocationU       string
		IsSupportVm     string
		AssetNumber     string
		OverDate        string
		UserID          uint
		AccessToken     string
		Callback        string
		DevManager      string
		OpsManager      string
		DeviceLabel     string
		DeviceDescribe  string
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}

	//先检测传入数据是否有问题
	info.Hostname = strings.TrimSpace(info.Hostname)
	info.Ip = strings.TrimSpace(info.Ip)
	info.ManageIp = strings.TrimSpace(info.ManageIp)
	info.AssetNumber = strings.TrimSpace(info.AssetNumber)
	info.AccessToken = strings.TrimSpace(info.AccessToken)
	info.UserID = session.ID
	info.LocationU = strings.TrimSpace(info.LocationU)
	info.DeviceDescribe = strings.TrimSpace(info.DeviceDescribe)

	accessTokenUser, errAccessToken := VerifyAccessToken(info.AccessToken, ctx, false)
	if errAccessToken != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAccessToken.Error()})
		return
	}
	info.UserID = accessTokenUser.ID
	session.ID = accessTokenUser.ID
	session.Role = accessTokenUser.Role

	if info.Hostname == "" || info.Ip == "" || info.NetworkID == uint(0) || info.OsID == uint(0) {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请将信息填写完整!"})
		return
	}

	mod, err := repo.GetDeviceById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	if session.Role != "Administrator" && mod.UserID != session.ID {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "该设备已被录入，不能重复录入!"})
		return
	}

	//validate host server info
	count, err := repo.CountDeviceByHostnameAndId(info.Hostname, info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Hostname + " 该主机名已存在!"})
		return
	}

	//validate ip from device
	countIp, err := repo.CountDeviceByIpAndId(info.Ip, info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if countIp > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Ip + " 该IP已存在!"})
		return
	}

	if info.ManageIp != "" {
		countManageIp, err := repo.CountDeviceByManageIpAndId(info.ManageIp, info.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countManageIp > 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + " 该管理IP已存在!"})
			return
		}
	}

	//匹配网络
	isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.Ip)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	if !isValidate {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Ip + "IP格式不正确!"})
		return
	}

	modelIp, err := repo.GetIpByIp(info.Ip)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Ip + "未匹配到网段!"})
		return
	}

	_, errNetwork := repo.GetNetworkById(modelIp.NetworkID)
	if errNetwork != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.Ip + "未匹配到网段!"})
		return
	}

	if info.ManageIp != "" {
		//匹配网络
		isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", info.ManageIp)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
			return
		}

		if !isValidate {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + "IP格式不正确!"})
			return
		}

		modelIp, err := repo.GetManageIpByIp(info.ManageIp)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + "未匹配到网段!"})
			return
		}

		_, errNetwork := repo.GetManageNetworkById(modelIp.NetworkID)
		if errNetwork != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + "未匹配到网段!"})
			return
		}
	}

	//校验是否使用OOB静态IP及管理IP是否填写
	if info.HardwareID > uint(0) {
		hardware, err := repo.GetHardwareById(info.HardwareID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		if hardware.Data != "" {
			if strings.Contains(hardware.Data, "<{manage_ip}>") || strings.Contains(hardware.Data, "<{manage_netmask}>") || strings.Contains(hardware.Data, "<{manage_gateway}>") {
				if info.ManageIp == "" || info.ManageNetworkID <= uint(0) {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该设备(SN:" + mod.Sn + ")使用的硬件配置模板的OOB网络类型为静态IP的方式，请填写管理IP!"})
					return
				}
			}
		}
	}
	if info.LocationU != "" {
		sublocation := strings.Split(info.LocationU, ",")
		min := 0
		for i := 0; i < len(sublocation); i++ {
			intlocate, err := strconv.Atoi(sublocation[i])
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + "U位填写错误，非整数"})
				return
			}

			if intlocate < 0 || intlocate >= 50 {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + "U位填写错误，非常规位置区间【1-50】"})
			}

			if min != 0 {
				if (intlocate - min) != 1 {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": info.ManageIp + "U位填写错误，非连续!"})
					return
				}
			}

			if intlocate > min {
				min = intlocate
			}
		}
	}

	info.IsSupportVm = strings.TrimSpace(info.IsSupportVm)
	if info.IsSupportVm == "" {
		info.IsSupportVm = "No"
	}

	//log
	logContent := make(map[string]interface{})
	logContent["data_old"] = mod

	_, errLog := repo.UpdateDeviceLogTypeByDeviceIdAndType(info.ID, "install", "install_history")
	if errLog != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errLog.Error()})
		return
	}

	updatebool := false
	var content string
	var message string

	if info.Hostname != mod.Hostname {
		updatebool = true
		content = content + "[update Hostname:\"" + mod.Hostname + "\" to \"" + info.Hostname + "\"]"
	}

	if info.Ip != mod.Ip {
		updatebool = true
		content = content + "[update Ip:\"" + mod.Ip + "\" to \"" + info.Ip + "\"]"
	}

	if info.ManageIp != mod.ManageIp {
		updatebool = true
		content = content + "[update ManageIp:\"" + mod.ManageIp + "\" to \"" + info.ManageIp + "\"]"
	}

	if info.NetworkID != mod.NetworkID {
		updatebool = true
		content = content + "[update NetworkID:\"" + strconv.Itoa(int(mod.NetworkID)) + "\" to \"" + strconv.Itoa(int(info.NetworkID)) + "\"]"
	}

	if info.ManageNetworkID != mod.ManageNetworkID {
		updatebool = true
		content = content + "[update ManageNetworkID:\"" + strconv.Itoa(int(mod.ManageNetworkID)) + "\" to \"" + strconv.Itoa(int(info.ManageNetworkID)) + "\"]"
	}

	if info.OsID != mod.OsID {
		updatebool = true
		content = content + "[update OsID:\"" + strconv.Itoa(int(mod.OsID)) + "\" to \"" + strconv.Itoa(int(info.OsID)) + "\"]"
	}

	if info.HardwareID != mod.HardwareID {
		updatebool = true
		content = content + "[update HardwareID:\"" + strconv.Itoa(int(mod.HardwareID)) + "\" to \"" + strconv.Itoa(int(info.HardwareID)) + "\"]"
	}

	if info.SystemID != mod.SystemID {
		updatebool = true
		content = content + "[update SystemID:\"" + strconv.Itoa(int(mod.SystemID)) + "\" to \"" + strconv.Itoa(int(info.SystemID)) + "\"]"
	}

	if info.LocationID != mod.LocationID {
		updatebool = true
		content = content + "[update LocationID:\"" + strconv.Itoa(int(mod.LocationID)) + "\" to \"" + strconv.Itoa(int(info.LocationID)) + "\"]"
	}

	if info.LocationU != mod.LocationU {
		updatebool = true
		content = content + "[update LocationU:\"" + mod.LocationU + "\" to \"" + info.LocationU + "\"]"
	}

	if info.AssetNumber != mod.AssetNumber {
		updatebool = true
		content = content + "[update AssetNumber:\"" + mod.AssetNumber + "\" to \"" + info.AssetNumber + "\"]"
	}

	if info.OverDate != mod.OverDate {
		updatebool = true
		content = content + "[update OverDate:\"" + mod.OverDate + "\" to \"" + info.OverDate + "\"]"
	}

	if info.IsSupportVm != mod.IsSupportVm {
		updatebool = true
		content = content + "[update IsSupportVm:\"" + mod.IsSupportVm + "\" to \"" + info.IsSupportVm + "\"]"
	}

	if info.UserID != mod.UserID {
		updatebool = true
		content = content + "[update UserID:\"" + strconv.Itoa(int(mod.UserID)) + "\" to \"" + strconv.Itoa(int(info.UserID)) + "\"]"
	}

	if info.DevManager != mod.DevManager {
		updatebool = true
		content = content + "[update DevManager:\"" + mod.DevManager + "\" to \"" + info.DevManager + "\"]"
	}

	if info.OpsManager != mod.OpsManager {
		updatebool = true
		content = content + "[update OpsManager:\"" + mod.OpsManager + "\" to \"" + info.OpsManager + "\"]"
	}

	if info.DeviceLabel != mod.DeviceLabel {
		updatebool = true
		content = content + "[update DeviceLabel:\"" + mod.DeviceLabel + "\" to \"" + info.DeviceLabel + "\"]"
	}

	if info.DeviceDescribe != mod.DeviceDescribe {
		updatebool = true
		content = content + "[update DeviceDescribe:\"" + mod.DeviceDescribe + "\" to \"" + info.DeviceDescribe + "\"]"
	}

	if !updatebool {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "数据未更改"})
		return
	}

	deviceNew, errUpdate := repo.UpdateDeviceById(info.ID, mod.BatchNumber, mod.Sn, info.Hostname, info.Ip, info.ManageIp, info.NetworkID, info.ManageNetworkID, info.OsID, info.HardwareID, info.SystemID, info.LocationID, info.LocationU, info.AssetNumber, info.OverDate, mod.Status, info.IsSupportVm, info.UserID, info.DevManager, info.OpsManager, info.DeviceLabel, info.DeviceDescribe)
	if errUpdate != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + errUpdate.Error()})
		return
	}

	logContent["data"] = deviceNew

	json, err := json.Marshal(logContent)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + err.Error()})
		return
	}

	_, errAddLog := repo.AddDeviceLog(mod.ID, "修改设备信息", "operate", string(json))
	if errAddLog != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAddLog.Error()})
		return
	}

	//callback script
	errCallback := SaveDeviceInstallCallback(ctx, mod.ID, "after_install", info.Callback)
	if errCallback != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCallback.Error()})
		return
	}

	var journal model.Journal
	journal.Title = mod.Hostname
	journal.Operation = "update"
	journal.Resource = "device"
	journal.Content = content
	journal.User = accessTokenUser.Username
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

// 上报安装进度
func ReportInstallInfo(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "application/json; charset=utf-8")
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	conf, ok := middleware.ConfigFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	logger, ok := middleware.LoggerFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	if conf.OsInstall.PxeConfigDir == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Pxe配置文件目录没有指定"})
		return
	}

	var info struct {
		Sn              string
		Title           string
		InstallProgress float64
		InstallLog      string
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}

	info.Sn = strings.TrimSpace(info.Sn)
	info.Title = strings.TrimSpace(info.Title)

	if info.Sn == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN参数不能为空!"})
		return
	}

	deviceId, err := repo.GetDeviceIdBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该设备不存在!"})
		return
	}

	device, err := repo.GetDeviceById(deviceId)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if device.Status != "pre_install" && device.Status != "installing" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该设备不在安装列表里!"})
		return
	}

	var status string
	var logTitle string

	if info.InstallProgress == -1 {
		status = "failure"
		info.InstallProgress = 0
		logTitle = info.Title
	} else if info.InstallProgress >= 0 && info.InstallProgress < 1 {
		status = "installing"
		logTitle = info.Title + "(" + fmt.Sprintf("安装进度 %.1f", info.InstallProgress*100) + "%)"
	} else if info.InstallProgress == 1 {
		status = "success"
		logTitle = info.Title + "(" + fmt.Sprintf("安装进度 %.1f", info.InstallProgress*100) + "%)"
		//logTitle = "安装成功"
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "安装进度参数不正确!"})
		return
	}

	/*
		if device.InstallLog != "" {
			info.InstallLog = device.InstallLog + "\n" + info.InstallLog
		}
	*/
	_, errUpdate := repo.UpdateInstallInfoById(device.ID, status, info.InstallProgress)
	if errUpdate != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
		return
	}

	if info.InstallProgress == 1 {
		//删除PXE配置文件
		macs, err := repo.GetMacListByDeviceID(device.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		for _, mac := range macs {
			pxeFileName := utils.GetPxeFileNameByMac(mac.Mac)
			confDir := conf.OsInstall.PxeConfigDir
			if utils.FileExist(confDir + "/" + pxeFileName) {
				err := os.Remove(confDir + "/" + pxeFileName)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}
			}
		}

		//删除Redis Key
		var key = fmt.Sprintf("IMP_DEVICE_INSTALL_%s", device.Sn)
		redis, okRedis := middleware.RedisFromContext(ctx)
		if okRedis {
			_, err := redis.Del(key)
			if err != nil {
				logger.Errorf(fmt.Sprintf("ERROR: REDIS DELETE[%s]", key))
			}
		}
	}

	var installLog string
	byteDecode, err := base64.StdEncoding.DecodeString(info.InstallLog)
	if err != nil {
		installLog = ""
	} else {
		installLog = string(byteDecode)
	}

	_, errAddLog := repo.AddDeviceLog(device.ID, logTitle, "install", installLog)
	if errAddLog != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAddLog.Error()})
		return
	}

	//add report
	if info.InstallProgress == 1 {
		errReportLog := repo.CopyDeviceToInstallReport(device.ID)
		if errReportLog != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errReportLog.Error()})
			return
		}
	}

	//exec callback script
	if info.InstallProgress == 1 {
		countCallback, errCountCallback := repo.CountDeviceInstallCallbackByDeviceIDAndType(device.ID, "after_install")
		if errCountCallback != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCountCallback.Error()})
			return
		}
		if countCallback > uint(0) {
			callback, errCallback := repo.GetDeviceInstallCallbackByDeviceIDAndType(device.ID, "after_install")
			if errCallback != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCallback.Error()})
				return
			}
			if callback.Content != "" {
				callback.Content = strings.Replace(callback.Content, "<{sn}>", device.Sn, -1)
				callback.Content = strings.Replace(callback.Content, "<{hostname}>", device.Hostname, -1)
				callback.Content = strings.Replace(callback.Content, "<{ip}>", device.Ip, -1)
				callback.Content = strings.Replace(callback.Content, "<{manage_ip}>", device.ManageIp, -1)
				if device.NetworkID > uint(0) {
					network, _ := repo.GetNetworkById(device.NetworkID)
					callback.Content = strings.Replace(callback.Content, "<{gateway}>", network.Gateway, -1)
					callback.Content = strings.Replace(callback.Content, "<{netmask}>", network.Netmask, -1)
				}
				if device.ManageNetworkID > uint(0) {
					manageNetwork, _ := repo.GetManageNetworkById(device.ManageNetworkID)
					callback.Content = strings.Replace(callback.Content, "<{manage_gateway}>", manageNetwork.Gateway, -1)
					callback.Content = strings.Replace(callback.Content, "<{manage_netmask}>", manageNetwork.Netmask, -1)
				}
				var runResult = "执行脚本:\n" + callback.Content
				bytes, errRunScript := utils.ExecScript(callback.Content)
				runResult += "\n\n" + "执行结果:\n" + string(bytes)
				var runStatus = "success"
				var runTime = time.Now().Format("2006-01-02 15:04:05")
				if errRunScript != nil {
					runStatus = "failure"
					runResult += "\n\n" + "错误信息:\n" + errRunScript.Error()
				}

				_, errUpdate := repo.UpdateDeviceInstallCallbackRunInfoByID(callback.ID, runTime, runResult, runStatus)
				if errUpdate != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
					return
				}
			}
		}
	}

	result := make(map[string]string)
	result["Result"] = "true"
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

// 上报Mac信息，生成Pxe文件
func ReportMacInfo(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "application/json; charset=utf-8")
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	conf, ok := middleware.ConfigFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	if conf.OsInstall.PxeConfigDir == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Pxe配置文件目录没有指定"})
		return
	}

	var info struct {
		Sn  string
		Mac []string
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}
	info.Sn = strings.TrimSpace(info.Sn)

	if info.Sn == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN参数不能为空!"})
		return
	}

	deviceId, err := repo.GetDeviceIdBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该设备不存在!"})
		return
	}

	device, err := repo.GetDeviceById(deviceId)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	osConfig, err := repo.GetOsConfigById(device.OsID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "PXE信息没有配置" + err.Error()})
		return
	}

	for _, i := range info.Mac {
		//mac 大写转为 小写
		mac := strings.ToLower(strings.TrimSpace(i))

		//录入Mac信息
		count, err := repo.CountMacByMacAndDeviceID(mac, device.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "MAC地址查询失败" + err.Error()})
			return
		}

		if count <= 0 {
			count, err := repo.CountMacByMac(mac)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if count > 0 {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该MAC地址已被其他设备录入"})
				return
			}

			_, errAddMac := repo.AddMac(device.ID, mac)
			if errAddMac != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAddMac.Error()})
				return
			}
		}

		//替换占位符
		osConfig.Pxe = strings.Replace(osConfig.Pxe, "{sn}", info.Sn, -1)
		osConfig.Pxe = strings.Replace(osConfig.Pxe, "\r\n", "\n", -1)

		pxeFileName := utils.GetPxeFileNameByMac(mac)
		logger, ok := middleware.LoggerFromContext(ctx)
		if !ok {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
			return
		}
		logger.Debugf("Create pxe file: %s", conf.OsInstall.PxeConfigDir+"/"+pxeFileName)

		errCreatePxeFile := utils.DeleteAndCreateFileIfExist(conf.OsInstall.PxeConfigDir, pxeFileName, osConfig.Pxe)
		if errCreatePxeFile != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "配置文件生成失败" + errCreatePxeFile.Error()})
			return
		}
	}

	result := make(map[string]string)
	result["Result"] = "true"
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func IsInPreInstallList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "application/json; charset=utf-8")
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误", "Content": ""})
		return
	}
	logger, ok := middleware.LoggerFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误", "Content": ""})
		return
	}

	var info struct {
		Sn string
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误", "Content": ""})
		return
	}

	info.Sn = strings.TrimSpace(info.Sn)

	if info.Sn == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN参数不能为空!"})
		return
	}

	result := make(map[string]string)

	var key = fmt.Sprintf("IMP_DEVICE_INSTALL_%s", info.Sn)
	redis, okRedis := middleware.RedisFromContext(ctx)
	if okRedis {
		v, err := redis.Get(key)
		if err == nil && v != "" {
			result["Result"] = v
			if v == "true" {
				w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作系统即将开始安装", "Content": result})
			} else {
				w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "未查询到安装任务", "Content": result})
			}
			return
		}
	}

	device, err := repo.GetDeviceStatusBySn(info.Sn)
	if err != nil {
		result["Result"] = "false"
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "查询失败,拒绝安装", "Content": result})
		return
	}

	if device.Status == "pre_install" || device.Status == "installing" {
		result["Result"] = "true"
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作系统即将开始安装", "Content": result})
	} else {
		result["Result"] = "false"
		if okRedis {
			_, err := redis.SetEx(key, result["Result"], 3600)
			if err != nil {
				logger.Errorf(fmt.Sprintf("ERROR: REDIS SetEX[%s]", err.Error()))
			}
		}
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "未查询到安装任务", "Content": result})
	}
}

func GetHardwareBySn(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "application/json; charset=utf-8")
	//repo := middleware.RepoFromContext(ctx)
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		//rest.Error(w, " ,", http.StatusInternalServerError)
		//w.WriteHeader(http.StatusFound)
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误", "Content": ""})
		return
	}
	var info struct {
		Sn string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		//rest.Error(w, " ", http.status)
		//w.WriteHeader(http.StatusFound)
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误", "Content": ""})
		return
	}

	info.Sn = strings.TrimSpace(info.Sn)

	if info.Sn == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN参数不能为空!"})
		return
	}

	device, err := repo.GetDeviceBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": ""})
		return
	}

	var manageNetwork model.ManageNetwork
	if device.ManageNetworkID > 0 {
		manageNetworkDetail, err := repo.GetManageNetworkById(device.ManageNetworkID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": ""})
			return
		}
		manageNetwork.Netmask = manageNetworkDetail.Netmask
		manageNetwork.Gateway = manageNetworkDetail.Gateway
	}

	hardware, err := repo.GetHardwareBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": ""})
		return
	}

	type ChildData struct {
		Name  string `json:"Name"`
		Value string `json:"Value"`
	}

	type ScriptData struct {
		Name string       `json:"Name"`
		Data []*ChildData `json:"Data"`
	}

	var data []*ScriptData
	var result2 []map[string]interface{}
	if hardware.Data != "" {
		hardware.Data = strings.Replace(hardware.Data, "\r\n", "\n", -1)
		bytes := []byte(hardware.Data)
		errData := json.Unmarshal(bytes, &data)
		if errData != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": ""})
			return
		}

		for _, v := range data {
			result3 := make(map[string]interface{})
			result3["Name"] = v.Name
			var result5 []map[string]interface{}
			for _, v2 := range v.Data {
				result4 := make(map[string]interface{})
				if strings.Contains(v2.Value, "<{manage_ip}>") {
					v2.Value = strings.Replace(v2.Value, "<{manage_ip}>", device.ManageIp, -1)
				}
				if strings.Contains(v2.Value, "<{manage_netmask}>") {
					v2.Value = strings.Replace(v2.Value, "<{manage_netmask}>", manageNetwork.Netmask, -1)
				}
				if strings.Contains(v2.Value, "<{manage_gateway}>") {
					v2.Value = strings.Replace(v2.Value, "<{manage_gateway}>", manageNetwork.Gateway, -1)
				}

				result4["Name"] = v2.Name
				result4["Script"] = base64.StdEncoding.EncodeToString([]byte(v2.Value))
				result5 = append(result5, result4)
			}
			result3["Scripts"] = result5
			result2 = append(result2, result3)
		}
	}

	result := make(map[string]interface{})
	result["Company"] = hardware.Company
	result["Product"] = hardware.Product
	result["ModelName"] = hardware.ModelName

	result["Hardware"] = result2

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "成功获取hardware信息", "Content": result})
}

func GetNetworkBySn(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	var info struct {
		Sn   string
		Type string
	}

	info.Sn = r.FormValue("sn")
	info.Type = r.FormValue("type")
	info.Sn = strings.TrimSpace(info.Sn)
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

	if info.Sn == "" {
		if info.Type == "raw" {
			w.Write([]byte(""))
		} else {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN参数不能为空"})
		}
		return
	}

	deviceId, err := repo.GetDeviceIdBySn(info.Sn)
	if err != nil {
		if info.Type == "raw" {
			w.Write([]byte(""))
		} else {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": ""})
		}
		return
	}

	device, err := repo.GetDeviceById(deviceId)
	if err != nil {
		if info.Type == "raw" {
			w.Write([]byte(""))
		} else {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": ""})
		}
		return
	}

	mod, err := repo.GetNetworkBySn(info.Sn)
	if err != nil {
		if info.Type == "raw" {
			w.Write([]byte(""))
		} else {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error(), "Content": ""})
		}
		return
	}

	mac, _ := repo.GetManufacturerMacBySn(info.Sn)

	mod.Vlan = strings.Replace(mod.Vlan, "\r\n", "\n", -1)
	mod.Trunk = strings.Replace(mod.Trunk, "\r\n", "\n", -1)
	mod.Bonding = strings.Replace(mod.Bonding, "\r\n", "\n", -1)

	result := make(map[string]interface{})
	result["Hostname"] = device.Hostname
	result["Ip"] = device.Ip
	result["Netmask"] = mod.Netmask
	result["Gateway"] = mod.Gateway
	result["Vlan"] = mod.Vlan
	result["Trunk"] = mod.Trunk
	result["Bonding"] = mod.Bonding
	result["HWADDR"] = mac
	if info.Type == "raw" {
		w.Header().Add("Content-type", "text/html; charset=utf-8")
		var str string
		if device.Hostname != "" {
			str += "HOSTNAME=\"" + device.Hostname + "\""
		}
		if device.Ip != "" {
			str += "\nIPADDR=\"" + device.Ip + "\""
		}
		if mod.Netmask != "" {
			str += "\nNETMASK=\"" + mod.Netmask + "\""
		}
		if mod.Gateway != "" {
			str += "\nGATEWAY=\"" + mod.Gateway + "\""
		}
		if mod.Vlan != "" {
			str += "\nVLAN=\"" + mod.Vlan + "\""
		}
		if mod.Trunk != "" {
			str += "\nTrunk=\"" + mod.Trunk + "\""
		}
		if mod.Bonding != "" {
			str += "\nBonding=\"" + mod.Bonding + "\""
		}
		str += "\nHWADDR=\"" + mac + "\""
		w.Write([]byte(str))
	} else {
		w.Header().Add("Content-type", "application/json; charset=utf-8")
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "成功获取network信息", "Content": result})
	}
}

func ValidateSn(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	var info struct {
		Sn string
	}
	info.Sn = strings.TrimSpace(info.Sn)

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	if info.Sn == "" {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "SN参数不能为空!", "Content": ""})
		return
	}

	count, err := repo.CountDeviceBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "参数错误!"})
		return
	}

	//match manufacturer
	countManufacturer, errCountManufacturer := repo.CountManufacturerBySn(info.Sn)
	if errCountManufacturer != nil {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": errCountManufacturer.Error()})
		return
	}
	if countManufacturer <= 0 {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "未在【资源池管理】里匹配到该SN，请先将该设备加电并进入BootOS!"})
		return
	}

	manufacturer, err := repo.GetManufacturerBySn(info.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "未在【资源池管理】里匹配到该SN，请先将该设备加电并进入BootOS!"})
		return
	}
	//validate user from manufacturer
	if session.Role != "Administrator" {
		if manufacturer.UserID != session.ID {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Content": manufacturer, "Message": "您无权操作其他人的设备!"})
			return
		}
	}

	if count > 0 {
		session, err := GetSession(w, r)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "参数错误" + err.Error()})
			return
		}

		device, err := repo.GetDeviceBySn(info.Sn)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "参数错误" + err.Error()})
			return
		}

		if session.Role != "Administrator" {
			if device.UserID != session.ID {
				w.WriteJSON(map[string]interface{}{"Status": "failure", "Content": manufacturer, "Message": "该设备已被录入，不能重复录入!"})
				return
			}
		}

		if device.Status == "success" {
			w.WriteJSON(map[string]interface{}{"Status": "failure", "Content": manufacturer, "Message": "该设备已安装成功，确定要重装？"})
			return
		}

		w.WriteJSON(map[string]interface{}{"Status": "failure", "Content": manufacturer, "Message": "该SN已存在，继续填写会覆盖旧的数据!"})
		return

	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Content": manufacturer, "Message": "SN填写正确!"})
		return
	}

}

func ImportDeviceForOpenApi(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	type Device struct {
		ID              uint
		BatchNumber     string
		Sn              string
		Hostname        string
		Ip              string
		ManageIp        string
		NetworkID       uint
		ManageNetworkID uint
		OsID            uint
		HardwareID      uint
		SystemID        uint
		Location        string
		LocationID      uint
		LocationU       string
		AssetNumber     string
		OverDate        string
		Status          string
		InstallProgress float64
		InstallLog      string
		NetworkName     string
		OsName          string
		HardwareName    string
		SystemName      string
		Content         string
		IsSupportVm     string
		UserID          uint
		DevManager      string
		OpsManager      string
		DeviceLabel     string
		DeviceDescribe  string
		AccessToken     string
	}

	var row Device
	if err := r.DecodeJSONPayload(&row); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error(), "Content": ""})
		return
	}

	accessTokenUser, errAccessToken := VerifyAccessToken(row.AccessToken, ctx, false)
	if errAccessToken != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAccessToken.Error()})
		return
	}
	row.UserID = accessTokenUser.ID

	row.Sn = strings.TrimSpace(row.Sn)
	row.Hostname = strings.TrimSpace(row.Hostname)
	row.Ip = strings.TrimSpace(row.Ip)
	row.ManageIp = strings.TrimSpace(row.ManageIp)
	row.HardwareName = strings.TrimSpace(row.HardwareName)
	row.SystemName = strings.TrimSpace(row.SystemName)
	row.OsName = strings.TrimSpace(row.OsName)
	row.AssetNumber = strings.TrimSpace(row.AssetNumber)
	row.Location = strings.TrimSpace(row.Location)
	row.LocationU = strings.TrimSpace(row.LocationU)
	row.DevManager = strings.TrimSpace(row.DevManager)
	row.OpsManager = strings.TrimSpace(row.OpsManager)
	row.DeviceLabel = strings.TrimSpace(row.DeviceLabel)
	row.DeviceDescribe = strings.TrimSpace(row.DeviceDescribe)

	batchNumber, err := repo.CreateBatchNumber()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if len(row.Sn) > 255 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN长度超过255限制"})
		return
	}

	if len(row.Hostname) > 255 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "主机名长度超过255限制"})
		return
	}

	if len(row.Location) > 255 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "位置长度超过255限制"})
		return
	}

	if len(row.LocationU) > 25 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "U位长度超过25限制"})
		return
	}

	if len(row.AssetNumber) > 255 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "财编长度超过255限制"})
		return
	}

	if len(row.DevManager) > 255 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "管理员长度超过255限制"})
		return
	}

	if len(row.OpsManager) > 255 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "管理员长度超过255限制"})
		return
	}

	if len(row.DeviceLabel) > 255 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "设备标签长度超过255限制"})
		return
	}

	if len(row.DeviceDescribe) > 255 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "设备备注长度超过255限制"})
		return
	}

	if row.Sn == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN不能为空"})
		return
	}

	if row.Hostname == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "主机名不能为空"})
		return
	}

	if row.Ip == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "IP不能为空"})
		return
	}

	if row.OsName == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作系统模板名称不能为空"})
		return
	}

	if row.SystemName == "" {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "系统安装模板不能为空"})
		return
	}

	//match manufacturer
	countManufacturer, errCountManufacturer := repo.CountManufacturerBySn(row.Sn)
	if errCountManufacturer != nil {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": errCountManufacturer.Error()})
		return
	}
	if countManufacturer <= 0 {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "未在【资源池管理】里匹配到该SN，请先将该设备加电并进入BootOS!"})
		return
	}

	//validate user from manufacturer
	manufacturer, err := repo.GetManufacturerBySn(row.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	if accessTokenUser.Role != "Administrator" && manufacturer.UserID != accessTokenUser.ID {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "您无权操作其他人的设备!"})
		return
	}

	countDevice, err := repo.CountDeviceBySn(row.Sn)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if countDevice > 0 {
		ID, err := repo.GetDeviceIdBySn(row.Sn)
		row.ID = ID
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		device, errDevice := repo.GetDeviceBySn(row.Sn)
		if errDevice != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
			return
		}

		if accessTokenUser.Role != "Administrator" && device.UserID != accessTokenUser.ID {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该设备已被其他人录入，不能重复录入"})
			return
		} else {
			if device.Status == "success" {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该设备已安装成功，请使用【单台录入】的功能重新录入并安装"})
				return
			}
		}

		//hostname
		countHostname, err := repo.CountDeviceByHostnameAndId(row.Hostname, ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误:" + err.Error()})
			return
		}
		if countHostname > 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该主机名已存在"})
			return
		}

		//IP
		countIp, err := repo.CountDeviceByIpAndId(row.Ip, ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countIp > 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该IP已存在"})
			return
		}

		if row.ManageIp != "" {
			//IP
			countManageIp, err := repo.CountDeviceByManageIpAndId(row.ManageIp, ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countManageIp > 0 {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该管理IP已存在"})
				return
			}
		}
	} else {
		//hostname
		countHostname, err := repo.CountDeviceByHostname(row.Hostname)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误:" + err.Error()})
			return
		}
		if countHostname > 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该主机名已存在"})
			return
		}

		//IP
		countIp, err := repo.CountDeviceByIp(row.Ip)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countIp > 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该IP已存在"})
			return
		}

		if row.ManageIp != "" {
			//IP
			countManageIp, err := repo.CountDeviceByManageIp(row.ManageIp)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countManageIp > 0 {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该管理IP已存在"})
				return
			}
		}
	}

	//匹配网络
	isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", row.Ip)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	if !isValidate {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "IP格式不正确"})
		return
	}

	modelIp, err := repo.GetIpByIp(row.Ip)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "未匹配到网段"})
		return
	}

	_, errNetwork := repo.GetNetworkById(modelIp.NetworkID)
	if errNetwork != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "未匹配到网段"})
		return
	}

	row.NetworkID = modelIp.NetworkID

	if row.ManageIp != "" {
		//匹配网络
		isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", row.ManageIp)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
			return
		}

		if !isValidate {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "管理IP格式不正确"})
			return
		}

		modelIp, err := repo.GetManageIpByIp(row.ManageIp)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "未匹配到管理网段"})
			return
		}

		_, errNetwork := repo.GetManageNetworkById(modelIp.NetworkID)
		if errNetwork != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "未匹配到管理网段"})
			return
		}

		row.ManageNetworkID = modelIp.NetworkID
	}

	//OSName
	countOs, err := repo.CountOsConfigByName(row.OsName)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if countOs <= 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "未匹配到操作系统"})
		return
	}
	mod, err := repo.GetOsConfigByName(row.OsName)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	row.OsID = mod.ID

	//SystemName
	countSystem, err := repo.CountSystemConfigByName(row.SystemName)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	if countSystem <= 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "未匹配到系统安装模板"})
		return
	}

	systemId, err := repo.GetSystemConfigIdByName(row.SystemName)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	row.SystemID = systemId

	if row.HardwareName != "" {
		//HardwareName
		countHardware, err := repo.CountHardwareWithSeparator(row.HardwareName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countHardware <= 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "未匹配到硬件配置模板"})
			return
		}

		hardware, err := repo.GetHardwareBySeaprator(row.HardwareName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		row.HardwareID = hardware.ID
	}

	if row.HardwareID > uint(0) {
		hardware, err := repo.GetHardwareById(row.HardwareID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		if hardware.Data != "" {
			if strings.Contains(hardware.Data, "<{manage_ip}>") || strings.Contains(hardware.Data, "<{manage_netmask}>") || strings.Contains(hardware.Data, "<{manage_gateway}>") {
				if row.ManageIp == "" || row.ManageNetworkID <= uint(0) {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "SN:" + row.Sn + "硬件配置模板的OOB网络类型为静态IP的方式，请填写管理IP!"})
					return
				}
			}
		}
	}

	if row.Location != "" {
		countLocation, err := repo.CountLocationByName(row.Location)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		if countLocation > 0 {
			locationId, err := repo.GetLocationIdByName(row.Location)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			row.LocationID = locationId
		}
		if row.LocationID <= uint(0) {
			locationId, err := repo.ImportLocation(row.Location)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if locationId <= uint(0) {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "未匹配到位置"})
				return
			}
			row.LocationID = locationId
		}
	}
	var status = "pre_install"
	row.IsSupportVm = "No"
	if countDevice > 0 {
		id, err := repo.GetDeviceIdBySn(row.Sn)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		deviceOld, err := repo.GetDeviceById(id)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		//init manufactures device_id
		countManufacturer, errCountManufacturer := repo.CountManufacturerBySn(row.Sn)
		if errCountManufacturer != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCountManufacturer.Error()})
			return
		}
		if countManufacturer > 0 {
			manufacturerId, errGetManufacturerBySn := repo.GetManufacturerIdBySn(row.Sn)
			if errGetManufacturerBySn != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errGetManufacturerBySn.Error()})
				return
			}
			_, errUpdate := repo.UpdateManufacturerDeviceIdById(manufacturerId, id)
			if errUpdate != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
				return
			}
		}

		_, errLog := repo.UpdateDeviceLogTypeByDeviceIdAndType(id, "install", "install_history")
		if errLog != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errLog.Error()})
			return
		}

		device, errUpdate := repo.UpdateDeviceById(id, batchNumber, row.Sn, row.Hostname, row.Ip, row.ManageIp, row.NetworkID, row.ManageNetworkID, row.OsID, row.HardwareID, row.SystemID, row.LocationID, row.LocationU, row.AssetNumber, row.OverDate, status, row.IsSupportVm, row.UserID, row.DevManager, row.OpsManager, row.DeviceLabel, row.DeviceDescribe)
		if errUpdate != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + errUpdate.Error()})
			return
		}

		//log
		logContent := make(map[string]interface{})
		logContent["data_old"] = deviceOld
		logContent["data"] = device

		json, err := json.Marshal(logContent)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + err.Error()})
			return
		}

		_, errAddLog := repo.AddDeviceLog(device.ID, "修改设备信息", "operate", string(json))
		if errAddLog != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAddLog.Error()})
			return
		}
	} else {
		device, err := repo.AddDevice(batchNumber, row.Sn, row.Hostname, row.Ip, row.ManageIp, row.NetworkID, row.ManageNetworkID, row.OsID, row.HardwareID, row.SystemID, row.LocationID, row.LocationU, row.AssetNumber, status, row.IsSupportVm, row.UserID, row.DevManager, row.OpsManager, row.DeviceLabel, row.DeviceDescribe)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + err.Error()})
			return
		}

		//log
		logContent := make(map[string]interface{})
		logContent["data"] = device
		json, err := json.Marshal(logContent)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + err.Error()})
			return
		}

		_, errAddLog := repo.AddDeviceLog(device.ID, "录入新设备", "operate", string(json))
		if errAddLog != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAddLog.Error()})
			return
		}

		//init manufactures device_id
		countManufacturer, errCountManufacturer := repo.CountManufacturerBySn(row.Sn)
		if errCountManufacturer != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCountManufacturer.Error()})
			return
		}
		if countManufacturer > 0 {
			manufacturerId, errGetManufacturerBySn := repo.GetManufacturerIdBySn(row.Sn)
			if errGetManufacturerBySn != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errGetManufacturerBySn.Error()})
				return
			}
			_, errUpdate := repo.UpdateManufacturerDeviceIdById(manufacturerId, device.ID)
			if errUpdate != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errUpdate.Error()})
				return
			}
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}

func ExportDevice(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Keyword        string
		CpuSum         string
		MemorySum      string
		ModelName      string
		LocationID     string
		HardwareID     string
		SystemID       string
		VersionAgt     string
		Status         string
		StartUpdatedAt string
		EndUpdatedAt   string
		UserID         string
	}

	info.Keyword = r.FormValue("Keyword")
	info.CpuSum = r.FormValue("CpuSum")
	info.MemorySum = r.FormValue("MemorySum")
	info.ModelName = r.FormValue("ModelName")
	info.LocationID = r.FormValue("LocationID")
	info.Status = r.FormValue("Status")
	info.StartUpdatedAt = r.FormValue("StartUpdatedAt")
	info.EndUpdatedAt = r.FormValue("EndUpdatedAt")
	info.UserID = r.FormValue("UserID")
	info.HardwareID = r.FormValue("HardwareID")
	info.SystemID = r.FormValue("SystemID")
	info.VersionAgt = r.FormValue("VersionAgt")

	info.Keyword = strings.TrimSpace(info.Keyword)
	info.CpuSum = strings.TrimSpace(info.CpuSum)
	info.MemorySum = strings.TrimSpace(info.MemorySum)
	info.ModelName = strings.TrimSpace(info.ModelName)
	info.LocationID = strings.TrimSpace(info.LocationID)
	info.Status = strings.TrimSpace(info.Status)
	info.StartUpdatedAt = strings.TrimSpace(info.StartUpdatedAt)
	info.EndUpdatedAt = strings.TrimSpace(info.EndUpdatedAt)
	info.UserID = strings.TrimSpace(info.UserID)
	info.HardwareID = strings.TrimSpace(info.HardwareID)
	info.SystemID = strings.TrimSpace(info.SystemID)
	info.VersionAgt = strings.TrimSpace(info.VersionAgt)

	var where string
	where = " where t1.id > 0 "
	if info.CpuSum != "" {
		cpuSum, _ := strconv.Atoi(info.CpuSum)
		where += " and t8.cpu_sum = " + fmt.Sprintf("%d", cpuSum)
	}
	if info.MemorySum != "" {
		memorySum, _ := strconv.Atoi(info.MemorySum)
		where += " and t8.memory_sum = " + fmt.Sprintf("%d", memorySum)
	}
	if info.ModelName != "" {
		where += " and t8.model_name = '" + info.ModelName + "'"
	}
	if info.VersionAgt != "" {
		where += " and t8.version_agt = '" + info.VersionAgt + "'"
	}
	if info.LocationID != "" {
		locationID, _ := strconv.Atoi(info.LocationID)
		where += " and t1.location_id = " + fmt.Sprintf("%d", locationID)
	}
	if info.HardwareID != "" {
		hardwareID, _ := strconv.Atoi(info.HardwareID)
		where += " and t1.hardware_id = " + fmt.Sprintf("%d", hardwareID)
	}
	if info.SystemID != "" {
		systemID, _ := strconv.Atoi(info.SystemID)
		where += " and t1.system_id = " + fmt.Sprintf("%d", systemID)
	}
	if info.Status != "" {
		where += " and t1.status = '" + info.Status + "'"
	}

	if info.StartUpdatedAt != "" {
		where += " and t1.updated_at >= '" + info.StartUpdatedAt + "'"
	}

	if info.EndUpdatedAt != "" {
		where += " and t1.updated_at <= '" + info.EndUpdatedAt + "'"
	}

	if info.UserID != "" {
		userID, _ := strconv.Atoi(info.UserID)
		where += " and t1.user_id = " + fmt.Sprintf("%d", userID)
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
			where += str + " t1.sn like '%" + v + "%' or t1.batch_number like '%" + v + "%' or t1.hostname like '%" + v + "%' or t1.ip like '%" + v + "%'"
		}
		where += " ) "
	}

	mods, err := repo.GetDeviceListWithPage(1000000, 0, where)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	// cd, err := iconv.Open("gbk", "utf-8") // convert utf-8 to gbk
	// if err != nil {
	// 	w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
	// 	return
	// }
	// defer cd.Close()

	var str string
	str = "SN,主机名,IP,CPU,Memory,系统模板,设备型号,位置,U位,管理IP,资产编号,批次号,设备类型,客户端版本,开发负责人,运维负责人,设备标签,备注,状态\n"
	str = utils.UTF82GBK(str)
	for _, device := range mods {
		var locationName string
		if device.LocationID > uint(0) {
			locationName, err = repo.FormatLocationNameById(device.LocationID, "", "-")
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}

		var statusName string
		switch device.Status {
		case "online":
			statusName = "在线"
		case "offline":
			statusName = "离线"
		case "pre_install":
			statusName = "等待安装"
		case "installing":
			statusName = "正在安装"
		case "success":
			statusName = "安装成功"
		case "failure":
			statusName = "安装失败"
		}

		str += device.Sn + ","
		str += utils.UTF82GBK(device.Hostname) + ","
		str += device.Ip + ","
		str += strconv.Itoa(int(device.CpuSum)) + ","
		str += strconv.Itoa(int(device.MemorySum)) + ","
		str += utils.UTF82GBK(device.SystemName) + ","
		str += device.ModelName + ","
		str += utils.UTF82GBK(locationName) + ","
		str += "\"" + device.LocationU + "\"" + ","
		str += device.ManageIp + ","
		str += utils.UTF82GBK(device.AssetNumber) + ","
		str += device.BatchNumber + ","
		str += "\"" + utils.UTF82GBK(device.VersionAgt) + "\"" + ","
		str += "\"" + utils.UTF82GBK(device.DevManager) + "\"" + ","
		str += "\"" + utils.UTF82GBK(device.OpsManager) + "\"" + ","
		str += "\"" + utils.UTF82GBK(device.DeviceLabel) + "\"" + ","
		str += "\"" + utils.UTF82GBK(device.DeviceDescribe) + "\"" + ","
		str += utils.UTF82GBK(statusName) + ","
		str += "\n"
	}
	bytes := []byte(str)
	filename := "device-list.csv"
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename='%s';filename*=utf-8''%s", filename, filename))
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Write(bytes)
}
