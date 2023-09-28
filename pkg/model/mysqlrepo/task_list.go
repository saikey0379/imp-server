package mysqlrepo

import (
	"fmt"
	"github.com/saikey0379/imp-server/pkg/model"
	"strconv"
)

func (repo *MySQLRepo) GetTaskByNo(taskNo string) (res []model.Task, err error) {
	err = repo.db.Model(model.Task{}).Where("task_no = ?", taskNo).Find(&res).Error
	return
}

func (repo *MySQLRepo) GetTaskById(id uint) (*model.Task, error) {
	var mod model.Task
	err := repo.db.Where("id = ?", id).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) CountTaskBySn(sn string) (count uint, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM task_host where sn = '" + sn + "'")
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) GetTaskFullListBySn(sn string) ([]model.TaskFull, error) {
	var result []model.TaskFull
	sql := "select t1.id,t1.task_type,t1.task_policy,t1.file_id,t1.status,t2.name,t2.file_type,t2.file_link,t1.file_mod,t2.interpreter,t1.parameter,t1.dest_path,t2.md5 from task_list t1,task_file t2,task_host t3 where t1.file_id=t2.id and t3.task_id=t1.id and t3.sn='" + sn + "'"

	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetTaskListBySn(sn string) ([]model.Task, error) {
	var result []model.Task
	sql := "select t1.id,t1.name from task_list t1,task_host t2 where t1.id=t2.task_id and t2.sn='" + sn + "'"
	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetTaskListByFileId(id uint) ([]model.Task, error) {
	var result []model.Task
	sql := "select id,name from task_list where file_id='" + fmt.Sprintf("%d", id) + "'"
	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetTaskStatusByID(taskID uint) (status string, err error) {
	row := repo.db.DB().QueryRow("SELECT status FROM task_info where id = " + strconv.Itoa(int(taskID)))
	if err := row.Scan(&status); err != nil {
		return "", err
	}
	return status, nil
}

func (repo *MySQLRepo) AddTask(mod model.Task) (*model.Task, error) {
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateTaskById(id uint, name string, manager string, description string, match_hosts string, task_type string, task_policy string, file_id uint, file_type string, file_mod string, parameter string, destpath string, status string) (*model.Task, error) {
	mod := model.Task{Name: name, Manager: manager, Description: description, MatchHosts: match_hosts, TaskType: task_type, TaskPolicy: task_policy, FileId: file_id, FileType: file_type, FileMod: file_mod, Parameter: parameter, DestPath: destpath, Status: status}
	//err := repo.db.Model(&mod).Where("id=?",id).Update(&mod).Error
	err := repo.db.First(&mod, id).Update("name", name).Update("manager", manager).Update("description", description).Update("match_hosts", match_hosts).Update("task_type", task_type).Update("task_policy", task_policy).Update("file_id", file_id).Update("file_type", file_type).Update("file_mod", file_mod).Update("parameter", parameter).Update("dest_path", destpath).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateTaskStatusById(id uint, status string) error {
	mod := model.Task{ID: id, Status: status}
	err := repo.db.First(&mod, id).Update("status", status).Error
	return err
}

func (repo *MySQLRepo) DeleteTaskById(id uint) (*model.Task, error) {
	mod := model.Task{}
	err := repo.db.Unscoped().Where("id = ?", id).Delete(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetTaskListWithPage(limit uint, offset uint, where string) (result []model.Task, err error) {
	sql := "SELECT * FROM task_list " + where + " order by task_list.updated_at DESC"
	if offset > 0 {
		sql += " limit " + fmt.Sprintf("%d", offset) + "," + fmt.Sprintf("%d", limit)
	} else {
		sql += " limit " + fmt.Sprintf("%d", limit)
	}
	err = repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) CountTask(where string) (count uint, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM task_list " + where)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) CountTaskByFileId(id uint) (count uint, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM task_list where file_id = " + fmt.Sprintf("%d", id))
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
