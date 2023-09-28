package route

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
)

func Metrics(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}

	var textInform strings.Builder

	modsCert, _ := repo.GetMetricListCert()
	for _, i := range modsCert {
		if i.NotAfter.Unix()-time.Now().Unix() < 3600*24*7 {
			textInform.WriteString("imp_expired_timestamp_slb_cert{domain=\"")
			textInform.WriteString(i.Name)
			textInform.WriteString("\",manager=\"")
			textInform.WriteString(i.Manager)
			textInform.WriteString("\"} ")
			textInform.WriteString(strconv.FormatInt(i.NotAfter.Unix(), 10))
			textInform.WriteString("\n")
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(textInform.String()))
}
