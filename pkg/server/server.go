package server

import (
	"github.com/saikey0379/go-json-rest/rest"
	"github.com/saikey0379/imp-server/pkg/config"
	"github.com/saikey0379/imp-server/pkg/logger"
	"github.com/saikey0379/imp-server/pkg/model"
	"github.com/saikey0379/imp-server/pkg/model/mysqlrepo"
	"net/http"
)

type Server struct {
	Conf    *config.Config
	Log     logger.Logger
	Repo    model.Repo
	Redis   model.Redis
	handler http.Handler
}

// NewServer 实例化http server
func NewServer(log logger.Logger, conf *config.Config, setup PipelineSetupFunc) (*Server, error) {
	repo, err := mysqlrepo.NewRepo(conf, log)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	api := rest.NewAPI()

	redis, err := model.NewRedis(conf, log)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	api.Use(setup(conf, log, repo, redis)...)

	// routes a global
	router, err := rest.MakeRouter(routes...)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	api.SetApp(router)

	return &Server{
		Conf:    conf,
		Log:     log,
		Repo:    repo,
		Redis:   redis,
		handler: api.MakeHandler(),
	}, nil
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.handler.ServeHTTP(w, r)
}
