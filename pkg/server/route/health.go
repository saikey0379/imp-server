package route

import (
	"net/http"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
)

func Health(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	_, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	w.WriteHeader(http.StatusOK)
	w.WriteJSON(map[string]interface{}{"Status": "up"})
}
