package mysqlrepo

import (
	"fmt"

	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) AddCluster(mod model.Cluster) (*model.Cluster, error) {
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) DeleteClusterById(id int) (*model.Cluster, error) {
	mod := model.Cluster{}
	err := repo.db.Unscoped().Where("id = ?", id).Delete(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetClusterListWithPage(limit uint, offset uint, where string) (mods []model.Cluster, err error) {
	sql := "SELECT * FROM  proxy_cluster" + where + " order by proxy_cluster.name ASC"

	if limit > 0 {
		if offset > 0 {
			sql += " limit " + fmt.Sprintf("%d", offset) + "," + fmt.Sprintf("%d", limit)
		} else {
			sql += " limit " + fmt.Sprintf("%d", limit)
		}
	}
	err = repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) GetClusterNameById(id int) (name string, err error) {
	var mod model.Cluster
	err = repo.db.Unscoped().Where("id = ?", id).Find(&mod).Error
	name = mod.Name
	return name, err
}

func (repo *MySQLRepo) GetClusterById(id int) (*model.Cluster, error) {
	var mod model.Cluster
	err := repo.db.Where("id = ?", id).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateClusterById(id int, name string, description string, ssh_user string, ssh_port int, ssh_key string, backends string, path_conf string, path_key string, exec_test string, exec_load string) (*model.Cluster, error) {
	mod := model.Cluster{Name: name, Description: description, SSHUser: ssh_user, SSHPort: ssh_port, SSHKey: ssh_key, Backends: backends, PathConf: path_conf, PathKey: path_key, ExecTest: exec_test, ExecLoad: exec_load}
	err := repo.db.First(&mod, id).Update("name", name).Update("description", description).Update("ssh_user", ssh_user).Update("ssh_port", ssh_port).Update("ssh_key", ssh_key).Update("backends", backends).Update("path_conf", path_conf).Update("path_key", path_key).Update("exec_test", exec_test).Update("exec_load", exec_load).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateClusterStatusById(id int, status string) (err error) {
	mod := model.Cluster{ID: id, Status: status}
	errUpdate := repo.db.Model(&mod).Where("id=?", id).Update("status", status).Error
	return errUpdate
}

func (repo *MySQLRepo) CountCluster(where string) (count int, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM proxy_cluster proxy_cluster " + where)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
