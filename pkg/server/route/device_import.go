package route

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/saikey0379/imp-server/pkg/known"
	"github.com/saikey0379/imp-server/pkg/utils"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
)

func UploadDevice(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	w.Header().Add("Content-type", "text/html; charset=utf-8")
	r.ParseForm()
	file, handle, err := r.FormFile("file")
	if err != nil {
		w.Write([]byte("{\"Message\":\"" + err.Error() + "\",\"Status\":\"error\"}"))
		return
	}

	dir := known.RootTmp
	if !utils.FileExist(dir) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			w.Write([]byte("{\"Message\":\"" + err.Error() + "\",\"Status\":\"error\"}"))
			return
		}
	}

	list := strings.Split(handle.Filename, ".")
	fix := list[len(list)-1]

	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%s", time.Now().UnixNano()) + handle.Filename))
	cipherStr := h.Sum(nil)
	md5 := fmt.Sprintf("%s", hex.EncodeToString(cipherStr))
	filename := "device-upload-" + md5 + "." + fix

	result := make(map[string]interface{})
	result["result"] = filename

	if utils.FileExist(dir + filename) {
		os.Remove(dir + filename)
	}

	f, err := os.OpenFile(dir+filename, os.O_WRONLY|os.O_CREATE, 0666)
	io.Copy(f, file)
	if err != nil {
		w.Write([]byte("{\"Message\":\"" + err.Error() + "\",\"Status\":\"error\"}"))
		return
	}
	defer f.Close()
	defer file.Close()

	data := map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result}
	json, err := json.Marshal(data)
	if err != nil {
		w.Write([]byte("{\"Message\":\"" + err.Error() + "\",\"Status\":\"error\"}"))
		return
	}
	w.Write([]byte(json))
	return
}

func ImportPriview(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	log, _ := middleware.LoggerFromContext(ctx)

	var info struct {
		Filename string
		Limit    uint
		Offset   uint
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	file := "/tmp/" + info.Filename

	// cd, err := iconv.Open("utf-8", "gbk") // convert gbk to utf8
	// if err != nil {
	// 	w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
	// 	return
	// }
	// defer cd.Close()

	input, err := os.Open(file)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	defer input.Close()

	// bufSize := 1024 * 1024
	// read := iconv.NewReader(cd, input, bufSize)
	// r2 := csv.NewReader(read)
	r2 := csv.NewReader(mahonia.NewDecoder("utf8").NewReader(input))

	ra, err := r2.ReadAll()
	if err != nil {
		log.Errorf("csvReader.ReadAll err: %v", err)
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	log.Infof("csv: %v", ra)

	length := len(ra)

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
		CpuSum          string
		MemorySum       string
		ModelName       string
		Location        string
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
		Content         string
		//		IsSupportVm     string
		UserID         uint
		DeviceType     string
		DevManager     string
		OpsManager     string
		DeviceLabel    string
		DeviceDescribe string
		VersionAgt     string
		ImportStatus   string
	}

	var data []Device
	//var result []string
	for i := 1; i < length; i++ {
		//result = append(result, ra[i][0])
		var row Device
		if len(ra[i]) != 19 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "导入文件格式错误!"
			data = append(data, row)
			continue
		}

		row.Sn = strings.TrimSpace(ra[i][0])
		row.Hostname = strings.TrimSpace(ra[i][1])
		row.Ip = strings.TrimSpace(ra[i][2])
		row.CpuSum = strings.TrimSpace(ra[i][3])
		row.MemorySum = strings.TrimSpace(ra[i][4])
		row.SystemName = strings.TrimSpace(ra[i][5])
		row.ModelName = strings.TrimSpace(ra[i][6])
		row.Location = strings.TrimSpace(ra[i][7])
		row.LocationU = strings.TrimSpace(ra[i][8])
		row.ManageIp = strings.TrimSpace(ra[i][9])
		row.AssetNumber = strings.TrimSpace(ra[i][10])
		row.BatchNumber = strings.TrimSpace(ra[i][11])
		row.DeviceType = strings.TrimSpace(ra[i][12])
		row.VersionAgt = strings.TrimSpace(ra[i][13])
		row.DevManager = strings.TrimSpace(ra[i][14])
		row.OpsManager = strings.TrimSpace(ra[i][15])
		row.DeviceLabel = strings.TrimSpace(ra[i][16])
		row.DeviceDescribe = strings.TrimSpace(ra[i][17])

		if len(row.Sn) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN长度超过255限制!"
		}

		if len(row.Hostname) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "主机名长度超过255限制!"
		}

		if len(row.Location) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "位置长度超过255限制!"
		}

		if len(row.AssetNumber) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "财编长度超过255限制!"
		}

		if row.Sn == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN不能为空!"
		}

		if row.Hostname == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "主机名不能为空!"
		}

		if row.Ip == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "IP不能为空!"
		}

		if row.SystemName == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "系统安装模板不能为空!"
		}

		if row.Location == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "位置不能为空!"
		}

		if row.LocationU == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "U位不能为空!"
		}

		//match manufacturer
		countManufacturer, errCountManufacturer := repo.CountManufacturerBySn(row.Sn)
		if errCountManufacturer != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCountManufacturer.Error()})
			return
		}
		if countManufacturer > 0 {
			//validate user from manufacturer
			manufacturer, err := repo.GetManufacturerBySn(row.Sn)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
				return
			}
			if session.Role != "Administrator" && manufacturer.UserID != session.ID {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "您无权操作其他人的设备!"
			}
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

			device, err := repo.GetDeviceBySn(row.Sn)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
				return
			}

			if session.Role != "Administrator" && device.UserID != session.ID {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该设备已被其他人录入，不能重复录入!"
			}

			//hostname
			countHostname, err := repo.CountDeviceByHostnameAndId(row.Hostname, ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误:" + err.Error()})
				return
			}
			if countHostname > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该主机名已存在!"
			}

			//IP
			countIp, err := repo.CountDeviceByIpAndId(row.Ip, ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countIp > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该IP已存在!"
			}

			if row.ManageIp != "" {
				//IP
				countManageIp, err := repo.CountDeviceByManageIpAndId(row.ManageIp, ID)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}

				if countManageIp > 0 {
					var br string
					if row.Content != "" {
						br = "<br />"
					}
					row.Content += br + "该管理IP已存在!"
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
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该主机名已存在!"
			}

			//IP
			countIp, err := repo.CountDeviceByIp(row.Ip)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countIp > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该IP已存在!"
			}

			if row.ManageIp != "" {
				//IP
				countManageIp, err := repo.CountDeviceByManageIp(row.ManageIp)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}

				if countManageIp > 0 {
					var br string
					if row.Content != "" {
						br = "<br />"
					}
					row.Content += br + "该管理IP已存在!"
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
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "IP格式不正确!"
		}

		modelIp, err := repo.GetIpByIp(row.Ip)
		if err != nil {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "未匹配到网段!"
		} else {
			network, errNetwork := repo.GetNetworkById(modelIp.NetworkID)
			if errNetwork != nil {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "未匹配到网段!"
			}
			row.NetworkName = network.Network
		}

		if row.ManageIp != "" {
			//匹配网络
			isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", row.ManageIp)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
				return
			}

			if !isValidate {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "管理IP格式不正确!"
			}

			modelIp, err := repo.GetManageIpByIp(row.ManageIp)
			if err != nil {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "未匹配到管理网段!"
			} else {
				network, errNetwork := repo.GetManageNetworkById(modelIp.NetworkID)
				if errNetwork != nil {
					var br string
					if row.Content != "" {
						br = "<br />"
					}
					row.Content += br + "未匹配到管理网段!"
				}
				row.ManageNetworkID = network.ID
			}
		}

		//SystemName
		countSystem, err := repo.CountSystemConfigByName(row.SystemName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countSystem <= 0 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "未匹配到系统安装模板!"
		}

		if row.HardwareName != "" {
			//HardwareName
			countHardware, err := repo.CountHardwareWithSeparator(row.HardwareName)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countHardware <= 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "未匹配到硬件配置模板!"
			} else {
				hardware, err := repo.GetHardwareBySeaprator(row.HardwareName)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}
				row.HardwareID = hardware.ID
			}
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
						var br string
						if row.Content != "" {
							br = "<br />"
						}
						row.Content += br + "硬件配置模板的OOB网络类型为静态IP的方式，请填写管理IP!"
					}
				}
			}
		}

		if row.Content != "" {
			row.ImportStatus = "Error"
		} else {
			row.ImportStatus = "Normal"
		}
		if countDevice > 0 {
			device, err := repo.GetDeviceBySn(row.Sn)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			if device.Status == "success" {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该设备已安装完成，重装会覆盖数据，确定安装？"
				if row.ImportStatus != "Error" {
					row.ImportStatus = "Notice"
				}
			} else if device.Status == "installing" {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该设备正在安装中，重装会覆盖数据，确定安装？"
				if row.ImportStatus != "Error" {
					row.ImportStatus = "Notice"
				}
			}
		}

		data = append(data, row)
	}

	var result []Device
	for i := 0; i < len(data); i++ {
		if uint(i) >= info.Offset && uint(i) < (info.Offset+info.Limit) {
			result = append(result, data[i])
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "recordCount": len(data), "Content": result})
}

func ImportDevice(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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
		Filename string
		Sns      []string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	if len(info.Sns) <= 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "请选中要导入的设备!"})
		return
	}

	file := "/tmp/" + info.Filename

	input, err := os.Open(file)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	r2 := csv.NewReader(mahonia.NewDecoder("utf8").NewReader(input)) // convert gbk to utf8
	ra, err := r2.ReadAll()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	length := len(ra)
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
		CpuSum          string
		MemorySum       string
		ModelName       string
		Location        string
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
		Content         string
		UserID          uint
		DeviceType      string
		DevManager      string
		OpsManager      string
		DeviceLabel     string
		DeviceDescribe  string
		VersionAgt      string
	}

	batchNumber, err := repo.CreateBatchNumber()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	//var result []string
	for i := 1; i < length; i++ {
		//result = append(result, ra[i][0])
		var row Device

		if len(ra[i]) != 19 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "导入文件格式错误!"
			continue
		}
		row.Sn = strings.TrimSpace(ra[i][0])
		row.Hostname = strings.TrimSpace(ra[i][1])
		row.Ip = strings.TrimSpace(ra[i][2])
		row.CpuSum = strings.TrimSpace(ra[i][3])
		row.MemorySum = strings.TrimSpace(ra[i][4])
		row.SystemName = strings.TrimSpace(ra[i][5])
		row.ModelName = strings.TrimSpace(ra[i][6])
		row.Location = strings.TrimSpace(ra[i][7])
		row.LocationU = strings.TrimSpace(ra[i][8])
		row.ManageIp = strings.TrimSpace(ra[i][9])
		row.AssetNumber = strings.TrimSpace(ra[i][10])
		row.BatchNumber = strings.TrimSpace(ra[i][11])
		if row.BatchNumber == "" {
			row.BatchNumber = batchNumber
		}
		row.DeviceType = strings.TrimSpace(ra[i][12])
		row.VersionAgt = strings.TrimSpace(ra[i][13])
		row.DevManager = strings.TrimSpace(ra[i][14])
		row.OpsManager = strings.TrimSpace(ra[i][15])
		row.DeviceLabel = strings.TrimSpace(ra[i][16])
		row.DeviceDescribe = strings.TrimSpace(ra[i][17])
		row.UserID = session.ID
		if !utils.IsInArrayStr(row.Sn, info.Sns) {
			continue
		}

		if len(row.Sn) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN长度超过255限制!"
		}

		if len(row.Hostname) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "主机名长度超过255限制!"
		}

		if len(row.CpuSum) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "CPU核数长度超过255限制!"
		}

		if len(row.MemorySum) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "内存大小长度超过255长度限制!"
		}

		if len(row.ModelName) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "设备型号长度超过255限制!"
		}

		if len(row.Location) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "位置长度超过255限制!"
		}
		if len(row.LocationU) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "U位长度超过255限制!"
		}

		if len(row.DevManager) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "负责人长度超过255限制!"
		}

		if len(row.OpsManager) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "负责人长度超过255限制!"
		}

		if len(row.DeviceLabel) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "设备标签长度超过255限制!"
		}
		if len(row.DeviceDescribe) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "设备备注长度超过255限制!"
		}

		if len(row.AssetNumber) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "财编长度超过255限制!"
		}

		if row.Sn == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN不能为空!"
		}

		if row.Hostname == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "主机名不能为空!"
		}

		if row.Ip == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "IP不能为空!"
		}

		if row.SystemName == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "系统安装模板不能为空!"
		}

		if row.Location == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "位置不能为空!"
		}

		if row.LocationU == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "U位不能为空!"
		}

		//match manufacturer
		countManufacturer, errCountManufacturer := repo.CountManufacturerBySn(row.Sn)
		if errCountManufacturer != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCountManufacturer.Error()})
			return
		}
		if countManufacturer > 0 {
			//validate user from manufacturer
			manufacturer, err := repo.GetManufacturerBySn(row.Sn)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
				return
			}
			if session.Role != "Administrator" && manufacturer.UserID != session.ID {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "您无权操作其他人的设备!"
			}
		}

		countDevice, err := repo.CountDeviceBySn(row.Sn)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countDevice > 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该SN设备已存在"})
			return
		} else {
			//hostname
			countHostname, err := repo.CountDeviceByHostname(row.Hostname)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误:" + err.Error()})
				return
			}
			if countHostname > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "该主机名已存在!"
			}

			//IP
			countIp, err := repo.CountDeviceByIp(row.Ip)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countIp > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "该IP已存在!"
			}

			if row.ManageIp != "" {
				//IP
				countManageIp, err := repo.CountDeviceByManageIp(row.ManageIp)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}

				if countManageIp > 0 {
					var br string
					if row.Content != "" {
						br = "<br />"
					}
					row.Content += br + "SN:" + row.Sn + "该管理IP已存在!"
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
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN:" + row.Sn + "IP格式不正确!"
		}

		modelIp, err := repo.GetIpByIp(row.Ip)
		if err != nil {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN:" + row.Sn + "未匹配到网段!"
		}

		_, errNetwork := repo.GetNetworkById(modelIp.NetworkID)
		if errNetwork != nil {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN:" + row.Sn + "未匹配到网段!"
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
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "管理IP格式不正确!"
			}

			modelIp, err := repo.GetManageIpByIp(row.ManageIp)
			if err != nil {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "未匹配到管理网段!"
			}

			_, errNetwork := repo.GetManageNetworkById(modelIp.NetworkID)
			if errNetwork != nil {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "未匹配到管理网段!"
			}

			row.ManageNetworkID = modelIp.NetworkID
		}

		//SystemName
		countSystem, err := repo.CountSystemConfigByName(row.SystemName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countSystem <= 0 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN:" + row.Sn + "未匹配到系统安装模板!"
		}

		systemId, err := repo.GetSystemConfigIdByName(row.SystemName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		row.SystemID = systemId

		if row.Location != "" {
			locationlist := strings.SplitAfter(row.Location, "-")
			location := locationlist[len(locationlist)-1]

			countLocation, err := repo.CountLocationByName(location)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			if countLocation > 0 {
				locationId, err := repo.GetLocationIdByName(location)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}
				row.LocationID = locationId
			}
		}
		var status = "offline"
		device, err := repo.AddDevice(row.BatchNumber, row.Sn, row.Hostname, row.Ip, row.ManageIp, row.NetworkID, row.ManageNetworkID, row.OsID, row.HardwareID, row.SystemID, row.LocationID, row.LocationU, row.AssetNumber, status, "No", row.UserID, row.DevManager, row.OpsManager, row.DeviceLabel, row.DeviceDescribe)
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
		countManufacturer, errCountManufacturer = repo.CountManufacturerBySn(row.Sn)
		if errCountManufacturer != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errCountManufacturer.Error()})
			return
		}
		if countManufacturer > 0 {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "设备已存在"})
			return
		} else {
			CpuSumInt, err := strconv.Atoi(row.CpuSum)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + "CPU非整数"})
				return
			}
			CpuSumUInt := uint(CpuSumInt)

			MemorySumInt, err := strconv.Atoi(row.MemorySum)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + "Memory非整数"})
				return
			}
			MemorySumUInt := uint(MemorySumInt)

			_, err = repo.AddManufacturer(row.ID, "", "", row.ModelName, row.Sn, row.Ip, "", "", "", CpuSumUInt, "", MemorySumUInt, "", 0, "", "", "", "", "", row.VersionAgt, "No", "")
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + "设备硬件信息添加" + err.Error()})
				return
			}

		}
	}

	//删除文件
	if utils.FileExist(file) {
		err := os.Remove(file)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}
