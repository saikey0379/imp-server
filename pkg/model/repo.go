package model

// Repo 数据仓库
type Repo interface {
	Close() error
	DropDB() error // 测试时使用

	//装机相关
	IQueryTerm
	IDevice
	INetwork
	IOsConfig
	ISystemConfig
	IHardware
	ILocation
	IIp
	IManageNetwork
	IManageIp
	IMac
	IManufacturer
	IDeviceLog
	IDeviceHistory
	IUser
	IUserAccessToken
	IDeviceInstallReport
	IDeviceInstallCallback
	IPlatformConfig

	//负载均衡相关
	IDomain
	IRoute
	IUpstream
	ICert
	IAccesslist
	ICluster

	//任务管理相关
	ITask
	ITaskHost
	ITaskResult
	IFile

	//日志审计
	IJournal
}
