package server

import (
	"github.com/saikey0379/go-json-rest/rest"
	"github.com/saikey0379/imp-server/pkg/server/cabinet"

	"github.com/saikey0379/imp-server/pkg/server/route"
)

var routes []*rest.Route

func init() {
	//healthz
	routes = append(routes, rest.Get("/healthz", route.Health))
	//metrics
	routes = append(routes, rest.Get("/metrics", route.Metrics))
	//Agent
	routes = append(routes, rest.Post("/api/agent/reportSysInfo", route.ReportSysInfo))
	routes = append(routes, rest.Post("/api/agent/reportSysTimestamp", route.ReportSysTimestamp))
	routes = append(routes, rest.Post("/api/agent/reportMacInfo", route.ReportMacInfo))
	routes = append(routes, rest.Post("/api/agent/reportInstallInfo", route.ReportInstallInfo))
	routes = append(routes, rest.Post("/api/agent/reportTaskResult", route.ReportTaskResult))

	routes = append(routes, rest.Post("/api/agent/getTaskFullListBySn", route.GetTaskFullListBySn))
	routes = append(routes, rest.Post("/api/agent/getHardwareBySn", route.GetHardwareBySn))
	routes = append(routes, rest.Get("/api/agent/getNetworkBySn", route.GetNetworkBySn))

	routes = append(routes, rest.Post("/api/agent/getPrepareInstallInfo", route.GetDevicePrepareInstallInfo))

	routes = append(routes, rest.Post("/api/agent/isInInstallList", route.IsInPreInstallList))
	//CabinetInfo
	routes = append(routes, rest.Post("/api/report/getCabinetTitle", cabinet.GetDeviceCabinetTitle))
	routes = append(routes, rest.Post("/api/report/getCabinetInfo", cabinet.GetDeviceCabinetInfo))
	//DeviceInstallReport
	routes = append(routes, rest.Post("/api/report/getInstallReport", route.GetDeviceInstallReport))
	//Device
	routes = append(routes, rest.Post("/api/device/terms/list", route.GetDeviceQueryTermsList))
	routes = append(routes, rest.Post("/api/device/getManuModelNameByCompany", route.GetManuModelNameByCompany))
	routes = append(routes, rest.Post("/api/device/add", route.AddDevice))
	routes = append(routes, rest.Post("/api/device/update", route.UpdateDevice))
	routes = append(routes, rest.Post("/api/device/list", route.GetDeviceList))
	routes = append(routes, rest.Post("/api/device/view", route.GetDeviceById))
	routes = append(routes, rest.Post("/api/device/viewFullById", route.GetFullDeviceById))
	routes = append(routes, rest.Post("/api/device/viewFullBySn", route.GetFullDeviceBySn))
	routes = append(routes, rest.Post("/api/device/viewFullHWBySn", route.GetFullDeviceHWBySn))
	routes = append(routes, rest.Post("/api/device/getNumByStatus", route.GetDeviceNumByStatus))
	routes = append(routes, rest.Post("/api/device/batchReInstall", route.BatchReInstall))
	routes = append(routes, rest.Post("/api/device/batchOnline", route.BatchOnline))
	routes = append(routes, rest.Post("/api/device/batchOffline", route.BatchOffline))
	routes = append(routes, rest.Post("/api/device/batchDelete", route.BatchDelete))
	routes = append(routes, rest.Post("/api/device/validateSn", route.ValidateSn))
	routes = append(routes, rest.Post("/api/device/batchCancelInstall", route.BatchCancelInstall))
	routes = append(routes, rest.Post("/api/device/importDeviceForOpenApi", route.ImportDeviceForOpenApi))
	routes = append(routes, rest.Get("/api/device/getDeviceBySn", route.GetDeviceBySn))
	routes = append(routes, rest.Get("/api/device/export", route.ExportDevice))
	routes = append(routes, rest.Post("/api/device/cidr/get", route.GetCidrInfoByNetwork))
	//Batch
	routes = append(routes, rest.Post("/api/device/batchPowerOn", route.BatchPowerOn))
	routes = append(routes, rest.Post("/api/device/batchPowerOff", route.BatchPowerOff))
	routes = append(routes, rest.Post("/api/device/batchReStart", route.BatchReStart))
	//Import device
	routes = append(routes, rest.Post("/api/device/upload", route.UploadDevice))
	routes = append(routes, rest.Post("/api/device/importPriview", route.ImportPriview))
	routes = append(routes, rest.Post("/api/device/importDevice", route.ImportDevice))
	//Network
	routes = append(routes, rest.Post("/api/device/network/add", route.AddNetwork))
	routes = append(routes, rest.Post("/api/device/network/list", route.GetNetworkList))
	routes = append(routes, rest.Post("/api/device/network/view", route.GetNetworkById))
	routes = append(routes, rest.Post("/api/device/network/update", route.UpdateNetworkById))
	routes = append(routes, rest.Post("/api/device/network/delete", route.DeleteNetworkById))
	routes = append(routes, rest.Post("/api/device/network/validateIp", route.ValidateIp))
	routes = append(routes, rest.Post("/api/device/network/getNotUsedIPListByNetworkId", route.GetNotUsedIPListByNetworkId))
	//ManageNetwork
	routes = append(routes, rest.Post("/api/device/manageNetwork/add", route.AddManageNetwork))
	routes = append(routes, rest.Post("/api/device/manageNetwork/list", route.GetManageNetworkList))
	routes = append(routes, rest.Post("/api/device/manageNetwork/view", route.GetManageNetworkById))
	routes = append(routes, rest.Post("/api/device/manageNetwork/update", route.UpdateManageNetworkById))
	routes = append(routes, rest.Post("/api/device/manageNetwork/delete", route.DeleteManageNetworkById))
	routes = append(routes, rest.Post("/api/device/manageNetwork/validateIp", route.ValidateManageIp))
	//Location
	routes = append(routes, rest.Post("/api/device/location/add", route.AddLocation))
	routes = append(routes, rest.Post("/api/device/location/list", route.GetLocationListByPid))
	routes = append(routes, rest.Post("/api/device/location/view", route.GetLocationById))
	routes = append(routes, rest.Post("/api/device/location/update", route.UpdateLocationById))
	routes = append(routes, rest.Post("/api/device/location/delete", route.DeleteLocationById))
	routes = append(routes, rest.Post("/api/device/location/tree", route.FormatLocationToTreeByPid))
	routes = append(routes, rest.Post("/api/device/location/getLocationTreeNameById", route.GetLocationTreeNameById))
	//Deploy
	routes = append(routes, rest.Post("/api/deploy/batchStartFromPxe", route.BatchStartFromPxe))
	routes = append(routes, rest.Post("/api/deploy/callback/list", route.GetDeviceInstallCallbackList))
	routes = append(routes, rest.Post("/api/deploy/reportInstallReport", route.ReportDeviceInstallReport))
	routes = append(routes, rest.Post("/api/deploy/installog/list", route.GetDeviceLogByDeviceIdAndType))
	//Scan device
	routes = append(routes, rest.Post("/api/deploy/scan/list", route.GetScanDeviceList))
	routes = append(routes, rest.Post("/api/deploy/scan/view", route.GetScanDeviceById))
	routes = append(routes, rest.Post("/api/deploy/scan/company/list", route.GetScanDeviceCompany))
	routes = append(routes, rest.Post("/api/deploy/scan/product/list", route.GetScanDeviceProduct))
	routes = append(routes, rest.Post("/api/deploy/scan/modelName/list", route.GetScanDeviceModelName))
	routes = append(routes, rest.Get("/api/deploy/scan/export", route.ExportScanDeviceList))
	routes = append(routes, rest.Post("/api/deploy/scan/batchAssignOwner", route.BatchAssignManufacturerOnwer))
	routes = append(routes, rest.Post("/api/deploy/scan/viewBySn", route.GetFullDeviceHWBySn))
	routes = append(routes, rest.Post("/api/deploy/scan/batchDelete", route.BatchDeleteScanDevice))
	//PxeConfig
	routes = append(routes, rest.Post("/api/deploy/pxe/add", route.AddOsConfig))
	routes = append(routes, rest.Post("/api/deploy/pxe/list", route.GetOsConfigList))
	routes = append(routes, rest.Post("/api/deploy/pxe/view", route.GetOsConfigById))
	routes = append(routes, rest.Post("/api/deploy/pxe/update", route.UpdateOsConfigById))
	routes = append(routes, rest.Post("/api/deploy/pxe/delete", route.DeleteOsConfigById))
	//SystemConfig
	routes = append(routes, rest.Post("/api/deploy/system/add", route.AddSystemConfig))
	routes = append(routes, rest.Post("/api/deploy/system/list", route.GetSystemConfigList))
	routes = append(routes, rest.Post("/api/deploy/system/view", route.GetSystemConfigById))
	routes = append(routes, rest.Post("/api/deploy/system/update", route.UpdateSystemConfigById))
	routes = append(routes, rest.Post("/api/deploy/system/delete", route.DeleteSystemConfigById))
	//Hardware
	routes = append(routes, rest.Post("/api/deploy/hardware/add", route.AddHardware))
	routes = append(routes, rest.Post("/api/deploy/hardware/list", route.GetHardwareList))
	routes = append(routes, rest.Post("/api/deploy/hardware/view", route.GetHardwareById))
	routes = append(routes, rest.Post("/api/deploy/hardware/update", route.UpdateHardwareById))
	routes = append(routes, rest.Post("/api/deploy/hardware/delete", route.DeleteHardwareById))
	routes = append(routes, rest.Post("/api/deploy/hardware/getCompanyByGroup", route.GetCompanyByGroup))
	routes = append(routes, rest.Post("/api/deploy/hardware/getProductByWhereAndGroup", route.GetProductByWhereAndGroup))
	routes = append(routes, rest.Post("/api/deploy/hardware/getModelNameByWhereAndGroup", route.GetModelNameByWhereAndGroup))
	routes = append(routes, rest.Get("/api/deploy/hardware/export", route.ExportHardware))
	routes = append(routes, rest.Post("/api/deploy/hardware/uploadCompanyHardware", route.UploadCompanyHardware))
	routes = append(routes, rest.Post("/api/deploy/hardware/uploadHardware", route.UploadHardware))
	routes = append(routes, rest.Post("/api/deploy/hardware/checkOnlineUpdate", route.CheckOnlineUpdate))
	routes = append(routes, rest.Post("/api/deploy/hardware/runOnlineUpdate", route.RunOnlineUpdate))
	//SlbDomain
	routes = append(routes, rest.Post("/api/slb/domain/select", route.GetDomainSelect))
	routes = append(routes, rest.Post("/api/slb/route/select", route.GetRouteSelectByDomainId))

	routes = append(routes, rest.Post("/api/slb/domain/add", route.AddDomain))
	routes = append(routes, rest.Post("/api/slb/domain/list", route.GetDomainList))
	routes = append(routes, rest.Post("/api/slb/domain/view", route.GetDomainById))
	routes = append(routes, rest.Post("/api/slb/domain/update", route.UpdateDomainById))
	routes = append(routes, rest.Post("/api/slb/domain/delete", route.DeleteDomainById))
	//SlbUpstream
	routes = append(routes, rest.Post("/api/slb/upstream/select", route.GetUpstreamSelectByClusterIds))

	routes = append(routes, rest.Post("/api/slb/upstream/backends", route.GetUpstreamBackendsById))
	routes = append(routes, rest.Post("/api/slb/upstream/add", route.AddUpstream))
	routes = append(routes, rest.Post("/api/slb/upstream/list", route.GetUpstreamList))
	routes = append(routes, rest.Post("/api/slb/upstream/view", route.GetUpstreamById))
	routes = append(routes, rest.Post("/api/slb/upstream/update", route.UpdateUpstreamById))
	routes = append(routes, rest.Post("/api/slb/upstream/delete", route.DeleteUpstreamById))
	//SlbCert
	routes = append(routes, rest.Post("/api/slb/cert/select", route.GetCertSelectByDomainNameAndClusterIds))

	routes = append(routes, rest.Post("/api/slb/cert/add", route.AddCert))
	routes = append(routes, rest.Post("/api/slb/cert/list", route.GetCertList))
	routes = append(routes, rest.Post("/api/slb/cert/view", route.GetCertById))
	routes = append(routes, rest.Post("/api/slb/cert/update", route.UpdateCertById))
	routes = append(routes, rest.Post("/api/slb/cert/delete", route.DeleteCertById))
	//SlbAcl
	routes = append(routes, rest.Post("/api/slb/accesslist/add", route.AddAccesslist))
	routes = append(routes, rest.Post("/api/slb/accesslist/list", route.GetAccesslistList))
	routes = append(routes, rest.Post("/api/slb/accesslist/view", route.GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType))
	routes = append(routes, rest.Post("/api/slb/accesslist/update", route.UpdateAccesslistByDomainIdAndRouteId))
	//SlbCluster
	routes = append(routes, rest.Post("/api/slb/cluster/select", route.GetClusterSelect))

	routes = append(routes, rest.Post("/api/slb/cluster/add", route.AddCluster))
	routes = append(routes, rest.Post("/api/slb/cluster/list", route.GetClusterList))
	routes = append(routes, rest.Post("/api/slb/cluster/view", route.GetClusterById))
	routes = append(routes, rest.Post("/api/slb/cluster/update", route.UpdateClusterById))
	routes = append(routes, rest.Post("/api/slb/cluster/delete", route.DeleteClusterById))
	routes = append(routes, rest.Post("/api/slb/cluster/conf", route.GetClusterConfById))
	routes = append(routes, rest.Post("/api/slb/cluster/confsync", route.ConfSyncByClusterId))
	routes = append(routes, rest.Post("/api/slb/cluster/conftest", route.ConfTestByClusterId))
	routes = append(routes, rest.Post("/api/slb/cluster/confload", route.ConfLoadByClusterId))
	routes = append(routes, rest.Post("/api/slb/cluster/file/upload", route.UploadSSHFile))
	//TaskManage
	routes = append(routes, rest.Post("/api/task/list", route.GetTaskList))
	routes = append(routes, rest.Post("/api/task/add", route.AddTask))
	routes = append(routes, rest.Post("/api/task/view", route.GetTaskById))
	routes = append(routes, rest.Post("/api/task/update", route.UpdateTaskById))
	routes = append(routes, rest.Post("/api/task/delete", route.DeleteTaskById))
	routes = append(routes, rest.Post("/api/task/result", route.GetTaskResultList))
	routes = append(routes, rest.Post("/api/task/result/terms/list", route.GetTaskResultQueryTermsList))
	routes = append(routes, rest.Post("/api/task/result/detail", route.GetTaskResultByResultId))
	routes = append(routes, rest.Post("/api/task/result/clear", route.ClearTaskResult))
	//TaskFile
	routes = append(routes, rest.Post("/api/task/file/select", route.GetFileSelect))

	routes = append(routes, rest.Post("/api/task/file/list", route.GetFileList))
	routes = append(routes, rest.Post("/api/task/file/add", route.AddFile))
	routes = append(routes, rest.Post("/api/task/file/view", route.GetFileById))
	routes = append(routes, rest.Post("/api/task/file/update", route.UpdateFileById))
	routes = append(routes, rest.Post("/api/task/file/delete", route.DeleteFileById))
	routes = append(routes, rest.Post("/api/task/file/upload", route.UploadFile))
	routes = append(routes, rest.Get("/api/task/file/getContent", route.GetFileContentByName))
	//SystemJournal
	routes = append(routes, rest.Post("/api/system/journal/list", route.GetJournalList))
	routes = append(routes, rest.Post("/api/system/journal/detail", route.GetJournalById))
	//User
	routes = append(routes, rest.Post("/api/user/add", route.AddUser))
	routes = append(routes, rest.Post("/api/user/list", route.GetUserList))
	routes = append(routes, rest.Post("/api/user/view", route.GetUserById))
	routes = append(routes, rest.Post("/api/user/update", route.UpdateUserById))
	routes = append(routes, rest.Post("/api/user/updateMyInfo", route.UpdateMyInfo))
	routes = append(routes, rest.Post("/api/user/delete", route.DeleteUserById))
	routes = append(routes, rest.Post("/api/user/login", route.Login))
	routes = append(routes, rest.Post("/api/user/logout", route.LoginOut))
	//PlatformConfig
	routes = append(routes, rest.Post("/api/platform/save", route.SavePlatformConfig))
	routes = append(routes, rest.Post("/api/platform/viewByName", route.GetPlatformConfigByName))
}
