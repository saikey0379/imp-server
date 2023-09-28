package server

import (
	"time"

	"github.com/saikey0379/go-json-rest/rest"

	"github.com/saikey0379/imp-server/pkg/config"
	"github.com/saikey0379/imp-server/pkg/logger"
	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
)

type PipelineSetupFunc func(conf *config.Config, log logger.Logger, repo model.Repo, rds model.Redis) []rest.Middleware

// TestPipeline 测试用中间件栈
func TestPipeline(conf *config.Config, log logger.Logger, repo model.Repo, rds model.Redis) []rest.Middleware {
	var pipe []rest.Middleware

	pipe = append(pipe, middleware.NewLimiterMiddleware(log, 10, time.Second))
	// pipe = append(pipe, middleware.NewCloseMiddleware(log))
	pipe = append(pipe, middleware.NewTimeoutMiddleware(60*time.Second))
	pipe = append(pipe, &rest.RecoverMiddleware{EnableResponseStackTrace: true})
	pipe = append(pipe, &rest.JSONIndentMiddleware{})
	pipe = append(pipe, &rest.ContentTypeCheckerMiddleware{})
	pipe = append(pipe, middleware.InjectConfig(conf))
	pipe = append(pipe, middleware.InjectLogger(log))
	pipe = append(pipe, middleware.InjectRepo(repo))
	pipe = append(pipe, middleware.InjectRedisClient(rds))

	return pipe
}

// DevPipeline 开发用中间件栈
func DevPipeline(conf *config.Config, log logger.Logger, repo model.Repo, redis model.Redis) []rest.Middleware {
	var pipe []rest.Middleware
	pipe = append(pipe, &rest.RecoverMiddleware{EnableResponseStackTrace: true})
	// pipe = append(pipe, middleware.NewLimiterMiddleware(log, 10, time.Second))
	// pipe = append(pipe, middleware.NewCloseMiddleware(log))
	// pipe = append(pipe, middleware.NewTimeoutMiddleware(60*time.Second))
	pipe = append(pipe, &rest.JSONIndentMiddleware{})
	//pipe = append(pipe, &rest.ContentTypeCheckerMiddleware{})
	pipe = append(pipe, middleware.InjectConfig(conf))
	pipe = append(pipe, middleware.InjectLogger(log))
	pipe = append(pipe, middleware.InjectRepo(repo))
	pipe = append(pipe, middleware.InjectRedisClient(redis))

	return pipe
}
