package config

// Loader 定义统一的配置加载接口
type Loader interface {
	Load() (*Config, error)
	Save(*Config) error
}

// Config config 数据结构体
type Config struct {
	Server struct {
		Listen        string `ini:"listen"`
		Port          int    `ini:"port"`
		RedisAddr     string `ini:"redisAddr"`
		RedisPort     int    `ini:"redisPort"`
		RedisPasswd   string `ini:"redisPasswd"`
		RedisDBNumber int    `ini:"redisDBNumber"`
	}
	Logger struct {
		Color     bool   `ini:"color"`
		Level     string `ini:"level"`
		LogFile   string `ini:"logFile"`
		Logger    Logger
		Repo      Repo
		OsInstall OsInstall
		Rsa       Rsa
		Cron      Cron
	}
	Repo struct {
		Connection          string `ini:"connection"`
		ConnectionIsCrypted string `ini:"connectionIsCrypted"`
		Addr                string
	}
	OsInstall struct {
		PxeConfigDir string
		LocalServer  string
	}
	Rsa struct {
		PublicKey  string `ini:"publicKey"`
		PrivateKey string `ini:"privateKey"`
	}
	Cron struct {
		InstallTimeout int `ini:"installTimeout"`
	}
}

type Logger struct {
	Color   bool   `ini:"color"`
	Level   string `ini:"level"`
	LogFile string `ini:"logFile"`
}

type Repo struct {
	Connection          string `ini:"connection"`
	ConnectionIsCrypted string `ini:"connectionIsCrypted"`
	Addr                string
}

type OsInstall struct {
	PxeConfigDir string `ini:"pxeConfigDir"`
}

type Rsa struct {
	PublicKey  string `ini:"publicKey"`
	PrivateKey string `ini:"privateKey"`
}

type Cron struct {
	InstallTimeout int `ini:"installTimeout"`
}
