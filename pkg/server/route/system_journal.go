package route

import (
	"fmt"
	"strings"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/utils"
)

type JournalListPageReq struct {
	ID          uint   `json:"id"`
	Resource    string `json:"Resource"`
	Operation   string `json:"Operation"`
	Username    string `json:"Username"`
	AccessToken string `json:"AccessToken"`
	Keyword     string `json:"keyword"`
	Limit       uint
	Offset      uint
}

func getJournalConditions(req JournalListPageReq) string {
	var where []string

	if req.Keyword = strings.TrimSpace(req.Keyword); req.Keyword != "" {
		where = append(where, fmt.Sprintf("( system_journal.id like %s or system_journal.title like %s or system_journal.resource like %s or system_journal.username like %s or system_journal.content like %s )", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'"))
	}

	if req.Resource != "" {
		where = append(where, fmt.Sprintf("( system_journal.resource = %s )", "'"+req.Resource+"'"))
	}
	if req.Operation != "" {
		where = append(where, fmt.Sprintf("( system_journal.operation = %s )", "'"+req.Operation+"'"))
	}
	if req.Username != "" {
		where = append(where, fmt.Sprintf("( system_journal.user = %s )", "'"+req.Username+"'"))
	}
	if len(where) > 0 {
		return " where " + strings.Join(where, " and ")
	} else {
		return ""
	}
}

func GetJournalList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info JournalListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	type Journal struct {
		ID        uint
		Title     string
		Operation string
		Resource  string
		Content   string
		User      string
		UpdatedAt utils.ISOTime
	}

	mods, err := repo.GetJournalListWithPage(info.Limit, info.Offset, getJournalConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	var journal Journal
	var journals []Journal
	for _, i := range mods {
		journal.ID = i.ID
		journal.Title = i.Title
		journal.Operation = i.Operation
		journal.Resource = i.Resource
		journal.Content = i.Content
		journal.User = i.User
		journal.UpdatedAt = utils.ISOTime(i.UpdatedAt)
		journals = append(journals, journal)
	}

	result := make(map[string]interface{})
	result["list"] = journals

	//总条数
	count, err := repo.CountJournal(getJournalConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetJournalById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info JournalListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	mod, err := repo.GetJournalById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": mod})
}
