package middleware

import (
	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/model"
)

// ctxRedisClientKey 注入的model.RedisClient对应的查询Key
var ctxRedisClientKey uint8

// RedisClientFromContext 从ctx中获取model.RedisClient
func RedisFromContext(ctx context.Context) (model.Redis, bool) {
	redis, ok := ctx.Value(&ctxRedisClientKey).(model.Redis)
	return redis, ok
}

// InjectRedisClient 注入model.RedisClient
func InjectRedisClient(redis model.Redis) rest.Middleware {
	fn := func(h rest.HandlerFunc) rest.HandlerFunc {
		return func(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
			ctx = context.WithValue(ctx, &ctxRedisClientKey, redis)
			h(ctx, w, r)
		}
	}
	return rest.MiddlewareSimple(fn)
}
