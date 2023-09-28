package mysqlrepo

import (
	"fmt"
	"time"

	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) AddTaskResult(mod model.TaskResult) (*model.TaskResult, error) {
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateTaskResultById(id uint, filesync string, result string, status string, startTime time.Time, endTime time.Time) (*model.TaskResult, error) {
	mod := model.TaskResult{FileSync: filesync, Result: result, Status: status, StartTime: startTime, EndTime: endTime}
	err := repo.db.First(&mod, id).Update("file_sync", filesync).Update("result", result).Update("status", status).Update("start_time", startTime).Update("end_time", endTime).Error
	return &mod, err
}

func (repo *MySQLRepo) GetTaskResultId(where string) (id uint, err error) {
	row := repo.db.DB().QueryRow("SELECT id FROM task_result " + where)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *MySQLRepo) GetTaskResultListWithPage(limit uint, offset uint, where string) (result []model.TaskResult, err error) {
	sql := "SELECT * FROM task_result" + where + " order by task_result.created_at DESC"
	if limit > 0 {
		if offset > 0 {
			sql += " limit " + fmt.Sprintf("%d", offset) + "," + fmt.Sprintf("%d", limit)
		} else {
			sql += " limit " + fmt.Sprintf("%d", limit)
		}
	}

	err = repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) CountTaskResult(where string) (count uint, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM task_result " + where)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) GetTaskResultByResultId(id uint) (*model.TaskResult, error) {
	var mod model.TaskResult
	err := repo.db.Where("id = ?", id).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetTaskResultBatchList(where string) ([]model.BatchId, error) {
	var result []model.BatchId
	sql := "SELECT DISTINCT batch_id from task_result " + where + " order by batch_id DESC"

	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetTaskResultHostList(where string) ([]model.HostName, error) {
	var result []model.HostName
	sql := "SELECT DISTINCT hostname from task_result " + where + " order by hostname ASC"

	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) ClearTaskResult(where string) error {
	sql := "delete from task_result " + where
	err := repo.db.Exec(sql).Error
	return err
}
func (repo *MySQLRepo) GetTaskBatchStatusByTaskId(taskId string) ([]model.TaskResult, error) {
	var mods []model.TaskResult
	sql := fmt.Sprintf("SELECT status,end_time from task_result where task_id='%s' and batch_id=(select max(batch_id) from task_result where task_id='%s') group by status", taskId, taskId)
	err := repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}
