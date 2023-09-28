package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DeviceFull struct {
	ID                uint
	BatchNumber       string
	Sn                string
	Hostname          string
	Ip                string
	ManageIp          string
	CpuSum            uint
	MemorySum         uint
	Gpu               string
	NetworkID         uint
	ManageNetworkID   uint
	OsID              uint
	HardwareID        uint
	SystemID          uint
	ModelName         string
	Location          string
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
	IsSupportVm       string
	UserID            uint
	OwnerName         string
	Callback          string
	BootosIp          string
	OobIp             string
	DeviceLabel       string
	DeviceDescribe    string
	DevManager        string
	OpsManager        string
	VersionAgt        string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Device struct {
	gorm.Model
	BatchNumber     string `sql:"not null;"`        //录入批次号
	Sn              string `sql:"not null;unique;"` //序列号
	Hostname        string `sql:"not null;"`        //主机名
	Ip              string `sql:"not null;unique;"` //IP
	ManageIp        string `sql:"unique;"`          //IP
	NetworkID       uint   `sql:"not null;"`        //网段模板ID
	ManageNetworkID uint   ``                       //管理网段模板ID
	OsID            uint   `sql:"not null;"`        //操作系统ID
	HardwareID      uint   ``                       //硬件配置模板ID
	SystemID        uint   `sql:"not null;"`        //系统配置模板ID
	LocationID      uint   `sql:"not null;"`
	LocationU       string
	AssetNumber     string //财编
	OverDate        string
	Status          string  `sql:"not null;"`                     //状态 'pre_run' 待安装,'running' 安装中,'success' 安装成功,'failure' 安装失败
	InstallProgress float64 `sql:"type:decimal(11,4);default:0;"` //安装进度
	InstallLog      string  `sql:"type:text;"`                    //安装日志
	IsSupportVm     string  `sql:"enum('Yes','No');NOT NULL;DEFAULT 'Yes'"`
	UserID          uint    `sql:"not null;default:0;"`
	DeviceLabel     string
	DeviceDescribe  string
	DevManager      string `sql:"not null;"`
	OpsManager      string `sql:"not null;"`
}

type DeviceCabinet struct {
	Sn        string
	Ip        string
	Hostname  string
	LocationU string
	Status    string
	UpdatedAt time.Time
}

// IDevice 设备操作接口
type IDevice interface {
	CountDeviceBySn(sn string) (uint, error)
	CountDeviceByHostname(hostname string) (uint, error)
	CountDeviceByHostnameAndId(hostname string, id uint) (uint, error)
	CountDeviceByIp(ip string) (uint, error)
	CountDeviceByLocationId(id uint) (uint, error)
	CountDeviceByManageIp(ManageIp string) (uint, error)
	CountDeviceByIpAndId(ip string, id uint) (uint, error)
	CountDeviceByManageIpAndId(ManageIp string, id uint) (uint, error)
	GetDeviceIdBySn(sn string) (uint, error)
	GetDeviceBySn(sn string) (*Device, error)
	GetDeviceByHostname(hostname string) (*Device, error)
	GetDeviceStatusBySn(sn string) (*Device, error)
	CountDevice(where string) (int, error)
	GetDeviceListWithPage(Limit uint, Offset uint, where string) ([]DeviceFull, error)
	GetDeviceById(Id uint) (*Device, error)
	DeleteDeviceById(Id uint) (*Device, error)
	ReInstallDeviceById(Id uint) (*Device, error)
	OnlineDeviceById(Id uint) (*Device, error)
	OfflineDeviceById(Id uint) (*Device, error)
	CancelInstallDeviceById(Id uint) (*Device, error)
	CreateBatchNumber() (string, error)
	AddDevice(BatchNumber string, Sn string, Hostname string, Ip string, ManageIp string, NetworkID uint, ManageNetworkID uint, OsID uint, HardwareID uint, SystemID uint, LocationID uint, ULocation string, AssetNumber string, Status string, IsSupportVm string, UserID uint, DevManager string, OpsManager string, DeviceLabel string, DeviceDescribe string) (*Device, error)
	UpdateDeviceById(ID uint, BatchNumber string, Sn string, Hostname string, Ip string, ManageIp string, NetworkID uint, ManageNetworkID uint, OsID uint, HardwareID uint, SystemID uint, LocationID uint, ULocation string, AssetNumber string, OverDate string, Status string, IsSupportVm string, UserID uint, DevManager string, OpsManager string, DeviceLabel string, DeviceDescribe string) (*Device, error)
	UpdateInstallInfoById(ID uint, status string, installProgress float64) (*Device, error)
	GetNetworkBySn(sn string) (*Network, error)
	GetFullDeviceById(id uint) (*DeviceFull, error)
	GetFullDeviceBySn(sn string) (*DeviceFull, error)
	CountDeviceByWhere(where string) (int, error)
	GetDeviceByWhere(where string) ([]Device, error)
	GetDeviceCabinetByLocationId(id uint) ([]DeviceCabinet, error)
}

type DeviceHistory struct {
	gorm.Model
	BatchNumber     string  `sql:"not null;"`        //录入批次号
	Sn              string  `sql:"not null;unique;"` //序列号
	Hostname        string  `sql:"not null;"`        //主机名
	Ip              string  `sql:"not null;unique;"` //IP
	ManageIp        string  `sql:"unique;"`          //ManageIP
	NetworkID       uint    `sql:"not null;"`        //网段模板ID
	ManageNetworkID uint    ``                       //管理网段模板ID
	OsID            uint    `sql:"not null;"`        //操作系统ID
	HardwareID      uint    ``                       //硬件配置模板ID
	SystemID        uint    `sql:"not null;"`        //系统配置模板ID
	Location        string  `sql:"not null;"`        //位置
	LocationID      uint    `sql:"not null;"`
	ULocation       string  `sql:"not null;"` //U位
	AssetNumber     string  //财编
	Status          string  `sql:"not null;"`                     //状态 'pre_run' 待安装,'running' 安装中,'success' 安装成功,'failure' 安装失败
	InstallProgress float64 `sql:"type:decimal(11,4);default:0;"` //安装进度
	InstallLog      string  `sql:"type:text;"`                    //安装日志
	IsSupportVm     string
}

// IDevice 设备操作接口
type IDeviceHistory interface {
	UpdateHistoryDeviceStatusById(ID uint, status string) (*DeviceHistory, error)
	CopyDeviceToHistory(ID uint) error
}

type DeviceInstallReport struct {
	gorm.Model
	Sn           string `sql:"not null;unique;"` //序列号
	OsName       string
	HardwareName string
	SystemName   string
	Status       string
	UserID       uint
}

type DeviceHardwareNameInstallReport struct {
	HardwareName string
	Count        uint
}

type DeviceProductNameInstallReport struct {
	ProductName string
	Count       uint
}

type DeviceOsNameInstallReport struct {
	OsName string
	Count  uint
}

type DeviceSystemNameInstallReport struct {
	SystemName string
	Count      uint
}

// IDevice 设备操作接口
type IDeviceInstallReport interface {
	CopyDeviceToInstallReport(ID uint) error
	CountDeviceInstallReportByWhere(Where string) (uint, error)
	GetDeviceHardwareNameInstallReport(Where string) ([]DeviceHardwareNameInstallReport, error)
	GetDeviceProductNameInstallReport(Where string) ([]DeviceProductNameInstallReport, error)
	GetDeviceCompanyNameInstallReport(Where string) ([]DeviceProductNameInstallReport, error)
	GetDeviceOsNameInstallReport(Where string) ([]DeviceOsNameInstallReport, error)
	GetDeviceSystemNameInstallReport(Where string) ([]DeviceSystemNameInstallReport, error)
}

// Network 网络
type Network struct {
	ID      uint   `sql:"not null;`
	Network string `sql:"not null;unique;"` //网段
	Netmask string `sql:"not null;`         //掩码
	Gateway string `sql:"not null;"`        //网关
	Vlan    string //vlan
	Trunk   string //trunk
	Bonding string //bonding
}

// INetwork 网络操作接口
type INetwork interface {
	CountNetworkByNetwork(Network string) (uint, error)
	CountNetwork() (uint, error)
	GetNetworkListWithPage(Limit uint, Offset uint) ([]Network, error)
	GetNetworkById(Id uint) (*Network, error)
	UpdateNetworkById(Id uint, Network string, Netmask string, Gateway string, Vlan string, Trunk string, Bonding string) (*Network, error)
	DeleteNetworkById(Id uint) (*Network, error)
	AddNetwork(Network string, Netmask string, Gateway string, Vlan string, Trunk string, Bonding string) (*Network, error)
}

// Network 网络
type Ip struct {
	gorm.Model
	NetworkID uint   `sql:"not null;"`
	Ip        string `sql:"not null;"`
}

// INetwork 网络操作接口
type IIp interface {
	DeleteIpByNetworkId(NetworkID uint) (*Ip, error)
	AddIp(NetworkID uint, Ip string) (*Ip, error)
	GetIpByIp(Ip string) (*Ip, error)
	GetNotUsedIPListByNetworkId(NetworkID uint) ([]Ip, error)
}

// ManageNetwork 网络
type ManageNetwork struct {
	gorm.Model
	Network string `sql:"not null;unique;"` //网段
	Netmask string `sql:"not null;`         //掩码
	Gateway string `sql:"not null;"`        //网关
	Vlan    string //vlan
	Trunk   string //trunk
	Bonding string //bonding
}

// INetwork 网络操作接口
type IManageNetwork interface {
	CountManageNetworkByNetwork(Network string) (uint, error)
	CountManageNetworkByNetworkAndId(Network string, ID uint) (uint, error)
	CountManageNetwork() (uint, error)
	GetManageNetworkListWithPage(Limit uint, Offset uint) ([]ManageNetwork, error)
	GetManageNetworkById(Id uint) (*ManageNetwork, error)
	UpdateManageNetworkById(Id uint, Network string, Netmask string, Gateway string, Vlan string, Trunk string, Bonding string) (*ManageNetwork, error)
	DeleteManageNetworkById(Id uint) (*ManageNetwork, error)
	AddManageNetwork(Network string, Netmask string, Gateway string, Vlan string, Trunk string, Bonding string) (*ManageNetwork, error)
	GetManufacturerMacBySn(Sn string) (string, error)
}

// Network 网络
type ManageIp struct {
	gorm.Model
	NetworkID uint   `sql:"not null;"`
	Ip        string `sql:"not null;"`
}

// INetwork 网络操作接口
type IManageIp interface {
	DeleteManageIpByNetworkId(NetworkID uint) (*ManageIp, error)
	AddManageIp(NetworkID uint, Ip string) (*ManageIp, error)
	GetManageIpByIp(Ip string) (*ManageIp, error)
}

// OS 操作系统
type OsConfig struct {
	gorm.Model
	Name string `sql:"not null;unique;"`    //操作系统名称
	Pxe  string `sql:"type:text;not null;"` //pxe信息
}

// IOS 操作系统操作接口
type IOsConfig interface {
	//GetOSByID(TaskID uint) (*OsConfig, error)
	CountOsConfigByName(Name string) (uint, error)
	CountOsConfigByNameAndId(Name string, ID uint) (uint, error)
	CountOsConfig() (uint, error)
	GetOsConfigListWithPage(Limit uint, Offset uint) ([]OsConfig, error)
	GetOsConfigById(Id uint) (*OsConfig, error)
	UpdateOsConfigById(Id uint, Name string, Pxe string) (*OsConfig, error)
	DeleteOsConfigById(Id uint) (*OsConfig, error)
	AddOsConfig(Name string, Pxe string) (*OsConfig, error)
	GetOsConfigByName(Name string) (*OsConfig, error)
}

// OS 操作系统
type DeviceLog struct {
	gorm.Model
	DeviceID uint   `sql:"not null;"`
	Title    string `sql:"not null;"`
	Type     string `sql:"not null;default:'install';"`
	Content  string `sql:"type:text;"` //pxe信息
}

type IDeviceLog interface {
	GetLastDeviceLogByDeviceID(DeviceID uint) (DeviceLog, error)
	GetDeviceLogListByDeviceIDAndType(DeviceID uint, Type string, Order string, MaxID uint) ([]DeviceLog, error)
	AddDeviceLog(DeviceID uint, Title string, Type string, Content string) (*DeviceLog, error)
	UpdateDeviceLogTypeByDeviceIdAndType(deviceID uint, Type string, NewType string) ([]DeviceLog, error)
}

// System 系统配置
type SystemConfig struct {
	gorm.Model
	Name    string `sql:"not null;unique;"`    //操作系统名称
	Content string `sql:"type:text;not null;"` //信息
}

// ISystemConfg 操作系统操作接口
type ISystemConfig interface {
	CountSystemConfigByName(Name string) (uint, error)
	CountSystemConfigByNameAndId(Name string, ID uint) (uint, error)
	GetSystemConfigIdByName(Name string) (uint, error)
	CountSystemConfig() (uint, error)
	GetSystemConfigListWithPage(Limit uint, Offset uint) ([]SystemConfig, error)
	GetSystemConfigById(Id uint) (*SystemConfig, error)
	UpdateSystemConfigById(Id uint, Name string, Content string) (*SystemConfig, error)
	DeleteSystemConfigById(Id uint) (*SystemConfig, error)
	AddSystemConfig(Name string, Content string) (*SystemConfig, error)
}

// Hardware 硬件配置
type Hardware struct {
	gorm.Model
	Company     string `sql:"not null;"`  //企业名称
	Product     string `sql:"not null;"`  //产品
	ModelName   string `sql:"not null;"`  //型号
	Raid        string `sql:"type:text;"` //raid配置
	Oob         string `sql:"type:text;"` //oob配置
	Bios        string `sql:"type:text;"` //bios配置
	IsSystemAdd string `sql:"enum('Yes','No');NOT NULL;DEFAULT 'Yes'"`
	Tpl         string //厂商提交的JSON信息
	Data        string //最终要执行的脚本信息
	Source      string //来源
	Version     string //版本
	Status      string `sql:"enum('Pending','Success','Failure');NOT NULL;DEFAULT 'Success'"` //状态
}

// IHardware 硬件配置操作接口
type IHardware interface {
	GetHardwareBySn(sn string) (*Hardware, error)
	CountHardwareByCompanyAndProductAndName(Company string, Product string, ModelName string) (uint, error)
	CountHardwareByCompanyAndProductAndNameAndId(Company string, Product string, ModelName string, ID uint) (uint, error)
	CountHardwareWithSeparator(Name string) (uint, error)
	GetHardwareIdByCompanyAndProductAndName(Company string, Product string, ModelName string) (uint, error)
	CountHardware(where string) (uint, error)
	GetHardwareListWithPage(Limit uint, Offset uint, where string) ([]Hardware, error)
	GetHardwareById(Id uint) (*Hardware, error)
	UpdateHardwareById(Id uint, Company string, Product string, ModelName string, Raid string, Oob string, Bios string, Tpl string, Data string, Source string, Version string, Status string) (*Hardware, error)
	DeleteHardwareById(Id uint) (*Hardware, error)
	AddHardware(Company string, Product string, ModelName string, Raid string, Oob string, Bios string, IsSystemAdd string, Tpl string, Data string, Source string, Version string, Status string) (*Hardware, error)
	GetCompanyByGroup() ([]Hardware, error)
	GetProductByWhereAndGroup(where string) ([]Hardware, error)
	GetModelNameByWhereAndGroup(where string) ([]Hardware, error)
	GetHardwareBySeaprator(Name string) (*Hardware, error)
	ValidateHardwareProductModel(Company string, Product string, ModelName string) (bool, error)
	CountHardwareByWhere(Where string) (uint, error)
	GetHardwareByWhere(Where string) (*Hardware, error)
	GetLastestVersionHardware() (Hardware, error)
	CreateHardwareBackupTable(Fix string) error
	RollbackHardwareFromBackupTable(Fix string) error
}

type Local struct {
	ID       uint   `sql:"not null;"`
	ShowName string `sql:"not null;"`
}

type Sysc struct {
	ID   uint   `sql:"not null;"`
	Name string `sql:"not null;"`
}

type Hwc struct {
	Company   string `sql:"not null;"`
	ModelName string `sql:"not null;"`
}

type Model struct {
	Company   string `sql:"not null;"`
	ModelName string `sql:"not null;"`
}

type CpuSum struct {
	CpuSum uint `sql:"not null;"`
}

type MemSum struct {
	MemorySum uint `sql:"not null;"`
}

type VersionAgt struct {
	VersionAgt string `sql:"not null;"`
}

type QueryTerm struct {
	LocaList   []Local
	SyscList   []Sysc
	HwcList    []Hwc
	ModelList  []Model
	CpuSumList []CpuSum
	MemSumList []CpuSum
	AgtList    []VersionAgt
}

type IQueryTerm interface {
	GetSystemConfigList() ([]Sysc, error)
	GetHardwareList() ([]Hwc, error)
	GetCompanyModelList() ([]Model, error)
	GetCpuSumList() ([]CpuSum, error)
	GetMemorySumList() ([]MemSum, error)
	GetManuModelNameByCompany(string) ([]Model, error)
	GetAgentVersionList() ([]VersionAgt, error)
}

// Location 位置
type Location struct {
	gorm.Model
	ID   uint   `sql:"not null;"` //ID
	Pid  uint   `sql:"not null;"` //父级ID
	Name string `sql:"not null;"` //位置名
}

// ILocation 位置操作接口
type ILocation interface {
	CountLocationByName(Name string) (uint, error)
	GetLocationIdByName(Name string) (uint, error)
	GetLocationListOrderByPId() ([]Location, error)

	FormatLocationToTreeByPid(Pid uint, Content []map[string]interface{}, Floor uint, SelectPid uint) ([]map[string]interface{}, error)
	FormatLocationNameById(id uint, content string, separator string) (string, error)
	GetLocationListByPidWithPage(Limit uint, Offset uint, pid uint) ([]Location, error)
	GetLocationListByPid(pid uint) ([]Location, error)
	CountLocationByPid(Pid uint) (uint, error)
	CountLocationByNameAndPid(Name string, Pid uint) (uint, error)
	CountLocationByNameAndPidAndId(Name string, Pid uint, ID uint) (uint, error)
	GetLocationById(Id uint) (*Location, error)
	UpdateLocationById(Id uint, Pid uint, Name string) (*Location, error)
	DeleteLocationById(Id uint) (*Location, error)
	AddLocation(Pid uint, Name string) (*Location, error)
	ImportLocation(Name string) (uint, error)
	FormatChildLocationIdById(id uint, content string, separator string) (string, error)
}

// Mac mac地址
type Mac struct {
	gorm.Model
	DeviceID uint   `sql:"not null;"`
	Mac      string `sql:"not null;unique;"` //位置名
}

type IMac interface {
	CountMacByMac(Mac string) (uint, error)
	CountMacByMacAndDeviceID(Mac string, DeviceID uint) (uint, error)
	AddMac(DeviceID uint, Mac string) (*Mac, error)
	GetMacListByDeviceID(DeviceID uint) ([]Mac, error)
	DeleteMacByDeviceId(deviceId uint) (*Mac, error)
}

type Manufacturer struct {
	gorm.Model
	DeviceID         uint   `sql:"not null;"`
	Company          string `sql:"not null;"`
	Product          string
	ModelName        string
	Sn               string
	Ip               string
	Mac              string
	Nic              string
	Cpu              string
	CpuSum           uint `sql:"type:int(11);default:0;"`
	Memory           string
	MemorySum        uint `sql:"type:int(11);default:0;"`
	Disk             string
	DiskSum          uint `sql:"type:int(11);default:0;"`
	Gpu              string
	Motherboard      string
	Raid             string
	Oob              string
	UserID           uint   `sql:"not null;default:0;"`
	IsShowInScanList string `sql:"enum('Yes','No');NOT NULL;DEFAULT 'Yes'"`
	NicDevice        string
	VersionAgt       string
	LastActiveTime   string
}

type ManufacturerFull struct {
	ID               uint
	DeviceID         uint
	Company          string
	Product          string
	ModelName        string
	Sn               string
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
	UserID           uint
	OwnerName        string
	NicDevice        string
	VersionAgt       string
	IsShowInScanList string
	LastActiveTime   string
}

type IManufacturer interface {
	GetManufacturerById(Id uint) (*Manufacturer, error)
	GetManufacturerBySn(Sn string) (*Manufacturer, error)
	DeleteManufacturerBySn(Sn string) (*Manufacturer, error)
	AddManufacturer(DeviceID uint, Company string, Product string, ModelName string, Sn string, Ip string, Mac string, Nic string, Cpu string, CpuSum uint, Memory string, MemorySum uint, Disk string, DiskSum uint, Gpu string, Motherboard string, Raid string, Oob string, NicDevice string, VersionAgt string, IsShowInScanList string, lastActiveTime string) (*Manufacturer, error)
	UpdateManufacturerBySn(Company string, Product string, ModelName string, Sn string, Ip string, Mac string, Nic string, Cpu string, CpuSum uint, Memory string, MemorySum uint, Disk string, DiskSum uint, Gpu string, Motherboard string, Raid string, Oob string, NicDevice string, VersionAgt string, IsShowInScanList string, lastActiveTime string) (*Manufacturer, error)
	UpdateManufacturerIsShowInScanListById(id uint, IsShowInScanList string) (*Manufacturer, error)
	UpdateManufacturerDeviceIdById(id uint, deviceId uint) (*Manufacturer, error)
	GetManufacturerListWithPage(Limit uint, Offset uint, Where string) ([]ManufacturerFull, error)
	CountManufacturerByWhere(Where string) (int, error)
	GetManufacturerCompanyByGroup(Where string) ([]Manufacturer, error)
	GetManufacturerProductByGroup(Where string) ([]Manufacturer, error)
	GetManufacturerModelNameByGroup(Where string) ([]Manufacturer, error)
	CountManufacturerBySn(Sn string) (uint, error)
	GetManufacturerIdBySn(Sn string) (uint, error)
	AssignManufacturerOnwer(Id uint, UserID uint) (*Manufacturer, error)
	AssignManufacturerNewOnwer(NewUserID uint, OldUserID uint) error
	UpdateManufacturerLastActiveTimeBySn(Sn string, time string) (*Manufacturer, error)
}

// Mac mac地址
type User struct {
	gorm.Model
	Username    string `sql:"not null;unique;"`
	Password    string `sql:"not null;"`
	Name        string
	PhoneNumber string
	Permission  string
	Status      string `sql:"enum('Enable','Disable');NOT NULL;DEFAULT 'Enable'"`
	Role        string `sql:"enum('Administrator','User');NOT NULL;DEFAULT 'User'"`
}

type IUser interface {
	CountUserByUsername(Username string) (uint, error)
	GetUserByUsername(Username string) (*User, error)
	GetUserById(Id uint) (*User, error)
	CountUser(Where string) (uint, error)
	DeleteUserById(Id uint) (*User, error)
	AddUser(Username string, Password string, Name string, PhoneNumber string, Permission string, Status string, Role string) (*User, error)
	UpdateUserById(Id uint, Password string, Name string, PhoneNumber string, Permission string, Status string, Role string) (*User, error)
	GetUserListWithPage(Limit uint, Offset uint, Where string) ([]User, error)
}

// Mac mac地址
type UserWithToken struct {
	ID          uint
	Username    string
	Name        string
	PhoneNumber string
	Status      string
	Role        string
	AccessToken string
}

// Mac mac地址
type UserAccessToken struct {
	gorm.Model
	UserID      uint   `sql:"not null;"`
	AccessToken string `sql:"not null;"`
}

type IUserAccessToken interface {
	CountUserAccessTokenByToken(AccessToken string) (uint, error)
	GetUserByAccessToken(AccessToken string) (*UserWithToken, error)
	DeleteUserAccessTokenByToken(AccessToken string) (*UserAccessToken, error)
	AddUserAccessToken(UserID uint, AccessToken string) (*UserAccessToken, error)
}

type DeviceInstallCallback struct {
	gorm.Model
	DeviceID     uint   `sql:"not null;"`
	DeviceSN     string `sql:"not null;"`
	CallbackType string `sql:"not null;"`
	Content      string `sql:"not null;"`
	RunTime      string
	RunResult    string
	RunStatus    string
}

type IDeviceInstallCallback interface {
	CountDeviceInstallCallbackByDeviceIDAndType(DeviceID uint, CallbackType string) (uint, error)
	GetDeviceInstallCallbackByWhere(Where string, Order string) ([]DeviceInstallCallback, error)
	GetDeviceInstallCallbackByDeviceIDAndType(DeviceID uint, CallbackType string) (*DeviceInstallCallback, error)
	DeleteDeviceInstallCallbackByDeviceID(DeviceID uint) (*DeviceInstallCallback, error)
	AddDeviceInstallCallback(DeviceID uint, CallbackType string, Content string, RunTime string, RunResult string, RunStatus string) (*DeviceInstallCallback, error)
	UpdateDeviceInstallCallbackByID(Id uint, DeviceID uint, CallbackType string, Content string, RunTime string, RunResult string, RunStatus string) (*DeviceInstallCallback, error)
	UpdateDeviceInstallCallbackRunInfoByID(Id uint, RunTime string, RunResult string, RunStatus string) (*DeviceInstallCallback, error)
}

type PlatformConfig struct {
	gorm.Model
	Name    string `sql:"not null;unique;"`
	Content string `sql:"type:longtext;"`
}

type IPlatformConfig interface {
	CountPlatformConfigByName(Name string) (uint, error)
	UpdatePlatformConfigById(Id uint, Name string, Pxe string) (*PlatformConfig, error)
	AddPlatformConfig(Name string, Content string) (*PlatformConfig, error)
	GetPlatformConfigByName(Name string) (*PlatformConfig, error)
}

const (
	Waiting   = "waiting"
	Executing = "executing"
	Rsyncd    = "rsyncd"
	//Success 成功
	Success = "success"
	//Failure 失败
	Failure = "failure"
	//Unknown 未知
	Unknown = "unknown"
	Init    = "init"
)

const (
	Shell  = "shell"
	Python = "python"
)

const (
	Script1 = "script"
	File1   = "file"
)

const (
	SSH  = "ssh"
	Salt = "salt"
)

// Domain 域名
type Domain struct {
	ID          int       `sql:"not null;"`    //ID
	Name        string    `gorm:"column:name"` //域名
	Manager     string    `gorm:"column:manager"`
	Description string    `gorm:"column:description"`
	Customize   string    `gorm:"column:customize`
	ClusterIds  string    `gorm:"column:cluster_ids"`
	ProxyType   int       `gorm:"column:proxy_type"`
	PortHttp    string    `gorm:"column:port_http"`
	PortHttps   string    `gorm:"column:port_https"`
	Http2       bool      `gorm:"column:http2`
	CertId      int       `gorm:"column:cert_id"`
	CreatedAt   time.Time `gorm:"column:created_at`
	UpdatedAt   time.Time `gorm:"column:updated_at`
}

type DomainUs struct {
	ID          int    `gorm:"primary_key"`
	Name        string `gorm:"column:name"`
	Manager     string `gorm:"column:manager"`
	ClusterName string `gorm:"column:name"`
}

// IDomain 位置操作接口
type IDomain interface {
	GetDomainNameById(id int) (name string, err error)
	CountDomain(where string) (int, error)
	GetDomainListWithPage(Limit uint, Offset uint, where string) ([]Domain, error)
	GetDomainListByClusterId(id int) ([]Domain, error)
	GetDomainById(Id int) (*Domain, error)
	UpdateDomainById(id int, manager string, description string, cluster_ids string, proxy_type int, port_http string, port_https string, http2 bool, cert_id int, customize string) (*Domain, error)
	DeleteDomainById(Id int) (*Domain, error)
	AddDomain(mod Domain) (*Domain, error)
	GetDomainListByCertId(id int) ([]DomainUs, error)
	CountDomainByCertId(id int) (int, error)
	CountDomainByCertIdAndClusterId(id int, clusterid string) (int, error)
}

func (Domain) TableName() string {
	return "proxy_domain"
}

type RouteFull struct {
	ID               int
	Index            int       `gorm:"column:idx"`
	Route            string    `gorm:"column:route"`
	Manager          string    `gorm:"column:manager"`
	MatchType        string    `gorm:"column:match_type`
	Customize        string    `gorm:"column:customize`
	Description      string    `gorm:"column:description"`
	UpstreamName     string    `gorm:"column:name"`
	UpstreamId       int       `gorm:"column:id"`
	UpstreamBackends string    `gorm:"column:backends"`
	UpdatedAt        time.Time `gorm:"column:updated_at`
}

type RouteUs struct {
	ID          int
	Route       string `gorm:"column:route"`
	RManager    string `gorm:"column:manager"`
	MatchType   string `gorm:"column:match_type`
	DomainId    string `gorm:"column:id"`
	DomainName  string `gorm:"column:name"`
	DManager    string `gorm:"column:manager"`
	ClusterName string `gorm:"column:name"`
}

type Route struct {
	ID          int    `sql:"not null;"` //ID
	Index       int    `gorm:"column:idx"`
	Route       string `gorm:"column:route"`
	Manager     string `gorm:"column:manager"`
	MatchType   string `gorm:"column:match_type`
	Customize   string `gorm:"column:customize`
	Description string `gorm:"column:description"`
	DomainId    int    `gorm:"column:domain_id"`
	UpstreamId  int    `gorm:"column:upstream_id"`
}

type IRoute interface {
	GetRouteFullListByDomainId(Id int) ([]RouteFull, error)
	GetRouteListByUpstreamId(id int) ([]RouteUs, error)
	GetRouteListByDomainId(id int) ([]Route, error)
	GetRouteNameById(id int) (string, error)
	AddRoute(mod Route) (*Route, error)
	UpdateRouteById(id int, index int, route string, manager string, description string, domain_id int, match_type string, customize string, upstream_id int) (*Route, error)
	DeleteRouteById(id int) error
	CountRouteByUpstreamId(id int) (int, error)
	CountRouteByDomainId(id int) (int, error)
	CountRouteByUpstreamIdAndClusterId(id int, clusterid string) (int, error)
}

func (Route) TableName() string {
	return "proxy_route"
}

type Upstream struct {
	ID          int       `sql:"not null;"`
	Name        string    `gorm:"column:name"`
	Manager     string    `gorm:"column:manager"`
	Description string    `gorm:"column:description"`
	Customize   string    `gorm:"column:customize"`
	ClusterIds  string    `gorm:"column:cluster_ids"`
	Backends    string    `gorm:"column:backends"`
	Used        int       `gorm:"column:used"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type IUpstream interface {
	GetUpstreamListWithPage(Limit uint, Offset uint, where string) ([]Upstream, error)
	GetUpstreamListByClusterId(Id int) ([]Upstream, error)
	GetUpstreamById(Id int) (*Upstream, error)
	CountUpstream(where string) (uint, error)
	GetUpstreamBackendsById(Id uint) (Upstream, error)
	AddUpstream(mod Upstream) (*Upstream, error)
	UpdateUpstreamById(id int, name string, manager string, description string, customize string, cluster_ids string, backends string) (*Upstream, error)
	GetUpstreamUsedById(id int) (int, error)
	UpdateUpstreamUsedById(id int, used int) error
	DeleteUpstreamById(id int) (*Upstream, error)
}

func (Upstream) TableName() string {
	return "proxy_upstream"
}

type Cert struct {
	ID          int       `sql:"not null;"`
	Name        string    `gorm:"column:name"`
	Manager     string    `gorm:"column:manager"`
	Description string    `gorm:"column:description"`
	ClusterIds  string    `gorm:"column:cluster_ids"`
	FileCert    string    `gorm:"column:file_cert"`
	ContentCert string    `gorm:"column:content_cert"`
	FileKey     string    `gorm:"column:file_key"`
	ContentKey  string    `gorm:"column:content_key"`
	NotBefore   time.Time `gorm:"column:not_before"`
	NotAfter    time.Time `gorm:"column:not_after"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type ICert interface {
	GetCertListWithPage(Limit uint, Offset uint, where string) ([]Cert, error)
	GetCertById(Id int) (*Cert, error)
	GetCertNameById(Id int) (string, error)
	CountCert(where string) (int, error)
	AddCert(mod Cert) (*Cert, error)
	UpdateCertById(id int, name string, description string, cluster_ids string, file_cert string, content_cert string, file_key string, content_key string, not_before time.Time, not_after time.Time, manager string) (*Cert, error)
	DeleteCertById(id int) (*Cert, error)
	GetMetricListCert() ([]Cert, error)
}

func (Cert) TableName() string {
	return "proxy_cert"
}

type Accesslist struct {
	ID          int       `sql:"not null;"`
	ClusterId   int       `gorm:"column:cluster_id"`
	DomainId    int       `gorm:"column:domain_id"`
	RouteId     int       `gorm:"column:route_id"`
	AccessType  string    `gorm:"column:access_type"`
	Host        string    `gorm:"column:host"`
	Manager     string    `gorm:"column:manager"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type IAccesslist interface {
	GetAccesslistListWithPage(Limit uint, Offset uint, where string) ([]Accesslist, error)
	CountAccesslist(where string) (int, error)
	CountAccesslistByDomainId(domain_id int) (int, error)
	CountAccesslistByDomainIdAndRouteId(domain_id int, route_id int) (int, error)
	GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType(cluster_id int, domain_id int, route_id int, access_type string) ([]Accesslist, error)
	AddAccesslist(mod Accesslist) (*Accesslist, error)
	UpdateAccesslistById(id int, allowed string, description string, manager string) (*Accesslist, error)
	DeleteAccesslistById(id int) (*Accesslist, error)
}

func (Accesslist) TableName() string {
	return "proxy_accesslist"
}

type Cluster struct {
	ID          int       `sql:"not null;"` //ID
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	SSHUser     string    `gorm:"column:ssh_user"`
	SSHPort     int       `gorm:"column:ssh_port"`
	SSHKey      string    `gorm:"column:ssh_key"`
	Backends    string    `gorm:"column:backends"`
	Status      string    `gorm:"column:status"`
	PathConf    string    `gorm:"column:path_conf"`
	PathKey     string    `gorm:"column:path_key"`
	ExecTest    string    `gorm:"column:exec_test"`
	ExecLoad    string    `gorm:"column:exec_load"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type ICluster interface {
	GetClusterListWithPage(Limit uint, Offset uint, where string) ([]Cluster, error)
	GetClusterNameById(id int) (name string, err error)
	GetClusterById(Id int) (*Cluster, error)
	UpdateClusterById(Id int, name string, description string, ssh_user string, ssh_port int, ssh_key string, backends string, path_conf string, path_key string, exec_test string, exec_load string) (*Cluster, error)
	UpdateClusterStatusById(Id int, status string) (err error)
	AddCluster(mod Cluster) (*Cluster, error)
	DeleteClusterById(id int) (*Cluster, error)
	CountCluster(where string) (count int, err error)
}

func (Cluster) TableName() string {
	return "proxy_cluster"
}

type Task struct {
	ID          uint
	Name        string    `gorm:"column:name"`
	Manager     string    `gorm:"column:manager"`
	Description string    `gorm:"column:description"`
	MatchHosts  string    `gorm:"column:match_hosts"`
	TaskType    string    `gorm:"column:task_type"`
	TaskPolicy  string    `gorm:"column:task_policy"`
	FileId      uint      `gorm:"column:file_id"`
	FileType    string    `gorm:"column:file_type"`
	FileMod     string    `gorm:"column:file_mod"`
	Parameter   string    `gorm:"column:parameter"`
	DestPath    string    `gorm:"column:dest_path"`
	Status      string    `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type TaskFull struct {
	ID          uint      `gorm:"column:id"`
	TaskType    string    `gorm:"column:task_type"`
	TaskPolicy  string    `gorm:"column:task_policy"`
	FileId      uint      `gorm:"column:file_id"`
	FileName    string    `gorm:"column:name"`
	FileType    string    `gorm:"column:file_type"`
	FileLink    string    `gorm:"column:file_link"`
	FileMod     string    `gorm:"column:file_mod"`
	Interpreter string    `gorm:"column:interpreter"`
	Parameter   string    `gorm:"column:parameter"`
	DestPath    string    `gorm:"column:dest_path"`
	Md5         string    `gorm:"column:md5"`
	Status      string    `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type ITask interface {
	GetTaskListWithPage(Limit uint, Offset uint, where string) ([]Task, error)
	GetTaskFullListBySn(sn string) ([]TaskFull, error)
	GetTaskListBySn(sn string) ([]Task, error)
	GetTaskListByFileId(id uint) ([]Task, error)
	GetTaskById(Id uint) (*Task, error)
	CountTask(where string) (uint, error)
	CountTaskBySn(sn string) (count uint, err error)
	CountTaskByFileId(id uint) (uint, error)
	AddTask(mod Task) (*Task, error)
	UpdateTaskById(id uint, name string, manager string, description string, match_hosts string, task_type string, task_policy string, file_id uint, file_type string, file_mod string, parameter string, destpath string, status string) (*Task, error)
	UpdateTaskStatusById(id uint, status string) error
	DeleteTaskById(id uint) (*Task, error)
}

type TaskHost struct {
	ID       uint
	Sn       string
	Hostname string
	Ip       string
	TaskID   uint
}

type ITaskHost interface {
	AddTaskHost(host TaskHost) error
	GetTaskHostListByTaskId(id uint) ([]TaskHost, error)
	DeleteTaskHostByTaskId(id uint) error
}

type TaskResult struct {
	ID        uint      `gorm:"column:id"`
	TaskId    uint      `gorm:"column:task_id"`
	BatchId   uint      `gorm:"column:batch_id"`
	Hostname  string    `gorm:"column:hostname"`
	FileSync  string    `gorm:"column:file_sync"`
	Result    string    `gorm:"column:result"`
	Status    string    `gorm:"column:status"`
	StartTime time.Time `gorm:"column:start_time"`
	EndTime   time.Time `gorm:"column:end_time"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

type BatchId struct {
	BatchId uint `sql:"not null;"`
}

type HostName struct {
	Hostname string `gorm:"column:hostname"`
}

type ITaskResult interface {
	AddTaskResult(mod TaskResult) (*TaskResult, error)
	GetTaskResultListWithPage(Limit uint, Offset uint, where string) ([]TaskResult, error)
	CountTaskResult(where string) (uint, error)
	GetTaskResultByResultId(id uint) (*TaskResult, error)
	GetTaskResultBatchList(where string) ([]BatchId, error)
	GetTaskResultHostList(where string) ([]HostName, error)
	UpdateTaskResultById(id uint, filesync string, result string, status string, startTime time.Time, endTime time.Time) (*TaskResult, error)
	GetTaskResultId(where string) (id uint, err error)
	ClearTaskResult(where string) error
}

func (Task) TableName() string {
	return "task_list"
}

func (TaskHost) TableName() string {
	return "task_host"
}

func (TaskResult) TableName() string {
	return "task_result"
}

type File struct {
	ID          uint
	Name        string    `gorm:"column:name"`
	Manager     string    `gorm:"column:manager"`
	Description string    `gorm:"column:description"`
	FileType    string    `gorm:"column:file_type"`
	FileLink    string    `gorm:"column:file_link"`
	Md5         string    `gorm:"column:md5"`
	Interpreter string    `gorm:"column:interpreter"`
	Content     string    `gorm:"column:content"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type IFile interface {
	GetFileListWithPage(Limit uint, Offset uint, where string) ([]File, error)
	GetFileListByFileType(file_type string) ([]File, error)
	GetFileById(id uint) (*File, error)
	GetFileByName(name string) (*File, error)
	CountFile(where string) (uint, error)
	AddFile(mod File) (*File, error)
	UpdateFileById(id uint, name string, manager string, description string, file_type string, md5 string, link string, interpreter string, content string) (*File, error)
	DeleteFileById(id uint) (*File, error)
}

func (File) TableName() string {
	return "task_file"
}

type Journal struct {
	ID        uint
	Title     string    `gorm:"column:title"`
	Operation string    `gorm:"column:operation"`
	Resource  string    `gorm:"column:resource"`
	Content   string    `gorm:"column:content"`
	User      string    `gorm:"column:user"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type IJournal interface {
	AddJournal(mod Journal) error
	GetJournalListWithPage(Limit uint, Offset uint, where string) ([]Journal, error)
	GetJournalById(id uint) (*Journal, error)
	CountJournal(where string) (uint, error)
}

func (Journal) TableName() string {
	return "system_journal"
}
