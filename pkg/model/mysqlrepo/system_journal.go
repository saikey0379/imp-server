package mysqlrepo

import (
	"fmt"

	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) GetJournalById(id uint) (*model.Journal, error) {
	var mod model.Journal
	err := repo.db.Where("id = ?", id).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) AddJournal(mod model.Journal) error {
	err := repo.db.Create(&mod).Error
	return err
}

func (repo *MySQLRepo) GetJournalListWithPage(limit uint, offset uint, where string) (result []model.Journal, err error) {
	sql := "SELECT * FROM system_journal " + where + " order by system_journal.updated_at DESC"
	if offset > 0 {
		sql += " limit " + fmt.Sprintf("%d", offset) + "," + fmt.Sprintf("%d", limit)
	} else {
		sql += " limit " + fmt.Sprintf("%d", limit)
	}
	err = repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) CountJournal(where string) (count uint, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM system_journal " + where)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
