package main

import (
	"fmt"
	"github.com/saikey0379/imp-server/pkg/config/iniconf"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/saikey0379/imp-server/pkg/config"
	"github.com/saikey0379/imp-server/pkg/logger"
	"github.com/saikey0379/imp-server/pkg/server"
	"github.com/saikey0379/imp-server/pkg/server/task"
	"github.com/saikey0379/imp-server/pkg/utils"
)

const (
	DefaultCnf    = "/../conf/imp-server.conf"
	VersionNumber = "0.0.1"
)

func main() {
	app := cli.NewApp()
	app.Name = "imp-server"
	app.Description = "seed server"
	app.Usage = "seed server install tool"
	app.Version = VersionNumber
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	configFile := dir + DefaultCnf
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Value:   configFile,
			Usage:   "config file",
		},
	}
	app.Action = func(c *cli.Context) (err error) {
		configFile = c.String("c")
		if !utils.FileExist(configFile) {
			return cli.NewExitError(fmt.Sprintf("The configuration file does not exist: %s", configFile), -1)
		}
		conf, err := iniconf.New(configFile).Load()
		if err = runServer(conf); err != nil {
			return cli.NewExitError(err.Error(), -1)
		}
		return nil
	}
	app.Run(os.Args)
}

func runServer(conf *config.Config) (err error) {
	log := logger.NewBeeLogger(conf)
	srvr, err := server.NewServer(log, conf, server.DevPipeline)
	if err != nil {
		log.Error(err)
		return err
	}

	taskAddr := task.GetTaskAddr()
	taskAddr.SetTaskAddr("127.0.0.1", conf.Server.Port)

	addr := fmt.Sprintf("%s:%d", conf.Server.Listen, conf.Server.Port)
	l4, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Error(err)
		return err
	}

	utils.InitSnowFlake()
	log.Infof(fmt.Sprintf("The HTTP server is running at http://%s", addr))
	//  sql upgrade

	if err := http.Serve(l4, srvr); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
