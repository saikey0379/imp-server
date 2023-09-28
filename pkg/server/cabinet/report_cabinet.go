package cabinet

import (
	"fmt"
	"github.com/saikey0379/imp-server/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
)

type Location struct {
	ID        uint
	Pid       uint
	Locations []Location
}

type CabinetTitle struct {
	ID   uint
	Name string
}

type DeviceLoca struct {
	ID       string
	Sn       string
	Ip       string
	Hostname string
	LocaUF   int
	SumU     int
	Color    string
}

type DeviceCabinetInfo struct {
	LocationID         uint
	LocationName       string
	DeviceLocas        []DeviceLoca
	DeviceCabinetInfos []DeviceCabinetInfo
}

func GetDeviceCabinetTitle(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	locationList, err := repo.GetLocationListOrderByPId()
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "获取机柜信息失败"})
		return
	}

	var titleIDList []uint

	var rootLocation *Location
	var unmatchLocationList []model.Location
	for _, i := range locationList {
		for x, j := range unmatchLocationList {
			if j.Pid == rootLocation.Pid && j.ID != rootLocation.ID {
				//对应父Location下元素大于1，提级
				var currLocation = Location{
					ID:  j.ID,
					Pid: j.Pid,
				}

				var parent = *rootLocation
				var tmp = &Location{
					ID:        j.Pid,
					Locations: []Location{parent, currLocation},
				}
				rootLocation = tmp

				//查看当前Location下是否存在设备，若存在且父Location不存在，则增加父Location至Title
				countLocationDevice, err := repo.CountDeviceByLocationId(j.ID)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "获取位置内设备信息错误"})
					return
				}

				if countLocationDevice > 0 && !utils.IsInArrayUint(j.Pid, titleIDList) {
					titleIDList = append([]uint{j.Pid}, titleIDList...)
				}
				unmatchLocationList = append(unmatchLocationList[x:], unmatchLocationList[:x]...)
			}
		}

		var currLocation = Location{
			ID:  i.ID,
			Pid: i.Pid,
		}

		if i.Pid == rootLocation.Pid && i.ID != rootLocation.ID {
			//对应父Location下元素大于1，提级
			var parent = *rootLocation
			var tmp = &Location{
				ID:        i.Pid,
				Locations: []Location{parent, currLocation},
			}
			rootLocation = tmp

			//查看当前Location下是否存在设备，若存在且父Location不存在，则增加父Location至Title
			countLocationDevice, err := repo.CountDeviceByLocationId(i.ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "获取位置内设备信息错误"})
				return
			}

			if countLocationDevice > 0 && !utils.IsInArrayUint(i.Pid, titleIDList) {
				titleIDList = append([]uint{i.Pid}, titleIDList...)
			}
		} else if currLocation.ID == rootLocation.ID {
			//若LocationID匹配。则补充Location的Pid
			rootLocation.Pid = currLocation.Pid
		} else {
			//若均未匹配，则添加至前置匹配队列中
			unmatchLocationList = append(unmatchLocationList, i)
		}
	}

	var cabinetTitleList []CabinetTitle

	for _, i := range titleIDList {
		var cabinetTitle CabinetTitle
		cabinetTitle.ID = i
		cabinetTitle.Name, err = repo.FormatLocationNameById(i, "", "-")
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		cabinetTitleList = append([]CabinetTitle{cabinetTitle}, cabinetTitleList...)
	}

	result := make(map[string]interface{})
	result["title"] = cabinetTitleList
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetDeviceCabinetInfo(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
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

	var info struct {
		ID uint
	}

	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	//获取当前机房的机柜列表
	localtionList, err := repo.GetLocationListByPid(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "获取当前机房机柜列表错误"})
		return
	}

	redis, okRedis := middleware.RedisFromContext(ctx)
	if !okRedis {
		logger.Errorf("ERROR: REDIS Unavaiablle[%s]", okRedis)
	}

	var deviceCabinetInfos []DeviceCabinetInfo

	// 查询机柜下设备列表
	for _, i := range localtionList {
		var repoBool bool
		var deviceCabinetInfo DeviceCabinetInfo
		deviceCabinetInfo.LocationID = i.ID
		deviceCabinetInfo.LocationName = i.Name

		var deviceLocas []DeviceLoca
		if okRedis {
			count, err := repo.CountDeviceByLocationId(i.ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "获取机柜设备数错误"})
				logger.Errorf(fmt.Sprintf("ERROR: CountDeviceByLocationId(%s)[%s]", i, err.Error()))
				return
			}

			keyMatch := fmt.Sprintf("IMP_CABINET_%s_%s_*", info.ID, i.ID)
			keys, err := redis.Keys(keyMatch)
			if err != nil {
				logger.Errorf("ERROR: REDIS KEYS[%s]%s", keyMatch, err.Error())
			}

			if int(count) != len(keys) {
				logger.Errorf(fmt.Sprintf("ERROR: CountDeviceByLocationId(%s)[%s]", i, err.Error()))
				repoBool = true
				if int(count) < len(keys) {
					for _, key := range keys {
						_, err = redis.Del(key)
						if err != nil {
							logger.Errorf(fmt.Sprintf("ERROR: REDIS DELETE KEY(%s)[%s]", key, err.Error()))
						}
					}
				}
			} else {
				for _, key := range keys {
					keyArray := strings.Split(key, "_")

					var deviceLoca DeviceLoca
					deviceLoca.ID = keyArray[3]

					localUF, err := strconv.Atoi(keyArray[4])
					if err != nil {
						logger.Errorf(fmt.Sprintf("ERROR: REDIS LocalUF Atoi(%s)[%s]", keyArray[4], err.Error()))
					}
					deviceLoca.LocaUF = localUF

					sumU, err := strconv.Atoi(keyArray[5])
					if err != nil {
						logger.Errorf(fmt.Sprintf("ERROR: REDIS SumU Atoi(%s)[%s]", keyArray[5], err.Error()))
					}
					deviceLoca.SumU = sumU

					deviceLoca.Ip = keyArray[6]
					deviceLoca.Hostname = keyArray[7]
					deviceLoca.Sn = keyArray[8]

					value, err := redis.Get(key)
					if err != nil {
						logger.Errorf(fmt.Sprintf("ERROR: REDIS GET KEY(%s)[%s]", key, err.Error()))
						repoBool = true
						break
					}

					t, err := time.Parse("2006-01-02 15:04:05", value)
					if err != nil {
						logger.Errorf(fmt.Sprintf("ERROR: REDIS VALUE Parse(%s)[%s]", value, err.Error()))
						repoBool = true
						break
					}

					if time.Now().Unix()-t.Unix() <= 600 {
						deviceLoca.Color = "lime"
					} else {
						deviceLoca.Color = "red"
					}

					deviceLocas = append(deviceLocas, deviceLoca)
				}
			}
		}
		if repoBool {
			mods, err := repo.GetDeviceCabinetByLocationId(i.ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "获取机柜设备信息错误"})
				return
			}
			for v, i := range mods {
				var deviceLoca DeviceLoca

				if v < 9 {
					deviceLoca.ID = "0" + strconv.Itoa(v+1)
				} else {
					deviceLoca.ID = strconv.Itoa(v + 1)
				}
				deviceLoca.Sn = i.Sn
				deviceLoca.Ip = i.Ip
				deviceLoca.Hostname = i.Hostname
				locaAry := strings.Split(i.LocationU, ",")
				locaUF, _ := strconv.Atoi(locaAry[0])
				deviceLoca.LocaUF = (locaUF-1)*12 + 13
				deviceLoca.SumU = len(locaAry)

				if time.Now().Unix()-i.UpdatedAt.Unix() <= 600 {
					deviceLoca.Color = "lime"
				} else {
					deviceLoca.Color = "red"
				}

				deviceLocas = append(deviceLocas, deviceLoca)
			}
		}
		deviceCabinetInfo.DeviceLocas = deviceLocas
		deviceCabinetInfos = append(deviceCabinetInfos, deviceCabinetInfo)
	}

	var rootDeviceCabinetInfo DeviceCabinetInfo
	rootDeviceCabinetInfo.LocationID = info.ID
	rootDeviceCabinetInfo.DeviceCabinetInfos = deviceCabinetInfos

	result := make(map[string]interface{})
	result["list"] = rootDeviceCabinetInfo
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}
