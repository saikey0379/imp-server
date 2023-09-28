package mysqlrepo

import (
	"fmt"
	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) AddTaskHost(mod model.TaskHost) error {
	err := repo.db.Create(&mod).Error
	return err
}

func (repo *MySQLRepo) GetTaskHostListByTaskId(id uint) (results []model.TaskHost, err error) {
	var sql = "SELECT * FROM task_host where task_id='" + fmt.Sprintf("%d", id) + "' order by task_file.name DESC"
	err = repo.db.Raw(sql).Scan(&results).Error
	return results, err
}

func (repo *MySQLRepo) DeleteTaskHostByTaskId(id uint) error {
	mod := model.TaskHost{}
	err := repo.db.Unscoped().Where("task_id = ?", id).Delete(&mod).Error
	return err
}
