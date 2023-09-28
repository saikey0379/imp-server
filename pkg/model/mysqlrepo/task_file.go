package mysqlrepo

import (
	"fmt"
	"strconv"

	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) GetFileById(id uint) (*model.File, error) {
	var mod model.File
	err := repo.db.Where("id = ?", id).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetFileByName(name string) (*model.File, error) {
	var mod model.File
	err := repo.db.Where("name = ?", name).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetFileStatusByID(fileID uint) (status string, err error) {
	row := repo.db.DB().QueryRow("SELECT status FROM task_file where id = " + strconv.Itoa(int(fileID)))
	if err := row.Scan(&status); err != nil {
		return "", err
	}
	return status, nil
}

func (repo *MySQLRepo) AddFile(mod model.File) (*model.File, error) {
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateFileById(id uint, name string, manager string, description string, file_type string, md5 string, file_link string, interpreter string, content string) (*model.File, error) {
	mod := model.File{Name: name, Manager: manager, Description: description, FileType: file_type, Interpreter: interpreter, Content: content, Md5: md5, FileLink: file_link}
	//err := repo.db.Model(&mod).Where("id=?",id).Update(&mod).Error
	err := repo.db.First(&mod, id).Update("name", name).Update("manager", manager).Update("description", description).Update("file_type", file_type).Update("interpreter", interpreter).Update("content", content).Update("Md5", md5).Update("FileLink", file_link).Error
	return &mod, err
}

func (repo *MySQLRepo) DeleteFileById(id uint) (*model.File, error) {
	mod := model.File{}
	err := repo.db.Unscoped().Where("id = ?", id).Delete(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetFileListByFileType(file_type string) (result []model.File, err error) {
	var sql string
	if file_type != "" {
		sql = "SELECT * FROM task_file where file_type = '" + file_type + "' order by task_file.name DESC"
	} else {
		sql = "SELECT * FROM task_file order by task_file.name DESC"
	}
	err = repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetFileListWithPage(limit uint, offset uint, where string) (result []model.File, err error) {
	sql := "SELECT * FROM task_file " + where + " order by task_file.updated_at DESC"
	if offset > 0 {
		sql += " limit " + fmt.Sprintf("%d", offset) + "," + fmt.Sprintf("%d", limit)
	} else {
		sql += " limit " + fmt.Sprintf("%d", limit)
	}
	err = repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) CountFile(where string) (count uint, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM task_file " + where)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
