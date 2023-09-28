package mysqlrepo

import (
	"fmt"
	"strconv"
	"time"

	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) AddCert(mod model.Cert) (*model.Cert, error) {
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateCertById(id int, name string, description string, cluster_ids string, file_cert string, content_cert string, file_key string, content_key string, not_before time.Time, not_after time.Time, manager string) (*model.Cert, error) {
	mod := model.Cert{Name: name, Description: description, ClusterIds: cluster_ids, FileCert: file_cert, ContentCert: content_cert, FileKey: file_key, ContentKey: content_key, NotBefore: not_before, NotAfter: not_after, Manager: manager}
	err := repo.db.First(&mod, id).Update("name", name).Update("description", description).Update("cluster_ids", cluster_ids).Update("file_cert", file_cert).Update("content_cert", content_cert).Update("file_key", file_key).Update("content_key", content_key).Update("not_before", not_before).Update("not_after", not_after).Update("manager", manager).Error
	return &mod, err
}

func (repo *MySQLRepo) DeleteCertById(id int) (*model.Cert, error) {
	mod := model.Cert{}
	err := repo.db.Unscoped().Where("id = ?", id).Delete(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) CountCertByName(name string) (int, error) {
	mod := model.Cert{Name: name}
	var count int
	err := repo.db.Model(mod).Where("name = ?", name).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountCert(where string) (count int, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM proxy_cert proxy_cert " + where)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) GetCertListWithPage(limit uint, offset uint, where string) (mods []model.Cert, err error) {
	sql := "SELECT * FROM  proxy_cert" + where + " order by proxy_cert.name ASC"
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

func (repo *MySQLRepo) GetCertListByClusterId(id int) (mods []model.Cert, err error) {
	sql := "SELECT * FROM  proxy_cert where cluster_ids like '%" + fmt.Sprintf("%d", id) + "%' order by proxy_cert.name ASC"
	err = repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) GetCertIdList() ([]string, error) {
	var mods []model.Cert
	var ids []string
	err := repo.db.Find(&mods).Error
	for _, v := range mods {
		ids = append(ids, strconv.Itoa(int(v.ID)))
	}
	return ids, err
}

func (repo *MySQLRepo) FormatCertNameById(id int, content string, separator string) (string, error) {
	var mod model.Cert
	if id <= int(0) {
		return content, nil
	}
	//result := make(map[uint]interface{})
	err := repo.db.Unscoped().Where("id = ?", id).Find(&mod).Error
	if err != nil {
		return content, err
	}

	if content == "" {
		content = mod.Name
	} else {
		content = mod.Name + separator + content
	}

	if mod.ID > 0 {
		parentContent, _ := repo.FormatCertNameById(mod.ID, "", separator)
		content = parentContent + separator + content
	}
	return content, nil
}

func (repo *MySQLRepo) GetCertNameById(id int) (name string, err error) {
	var mod model.Cert
	err = repo.db.Unscoped().Where("id = ?", id).Find(&mod).Error
	name = mod.Name
	return name, err
}

func (repo *MySQLRepo) GetCertById(id int) (*model.Cert, error) {
	var mod model.Cert
	err := repo.db.Where("id = ?", id).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetCertIdByName(name string) (int, error) {
	mod := model.Cert{Name: name}
	err := repo.db.Where("name = ?", name).Find(&mod).Error
	return mod.ID, err
}

func (repo *MySQLRepo) GetCertNameIdListByCid(cid int) ([]model.Cert, error) {
	var mods []model.Cert
	err := repo.db.Unscoped().Where("cluster_ids like '''%?%'''", cid).Find(&mods).Error
	return mods, err
}
