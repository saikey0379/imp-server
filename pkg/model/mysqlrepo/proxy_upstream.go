package mysqlrepo

import (
	"fmt"
	"strconv"

	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) AddUpstream(mod model.Upstream) (*model.Upstream, error) {
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateUpstreamById(id int, name string, manager string, description string, customize string, cluster_ids string, backends string) (*model.Upstream, error) {
	mod := model.Upstream{Name: name, Manager: manager, Description: description, Customize: customize, ClusterIds: cluster_ids, Backends: backends}
	err := repo.db.First(&mod, id).Update("name", name).Update("manager", manager).Update("description", description).Update("customize", customize).Update("cluster_ids", cluster_ids).Update("backends", backends).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateUpstreamUsedById(id int, used int) error {
	mod := model.Upstream{Used: used}
	err := repo.db.First(&mod, id).Update("used", used).Error
	return err
}

func (repo *MySQLRepo) DeleteUpstreamById(id int) (*model.Upstream, error) {
	mod := model.Upstream{}
	err := repo.db.Unscoped().Where("id = ?", id).Delete(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) CountUpstreamByName(name string) (uint, error) {
	mod := model.Upstream{Name: name}
	var count uint
	err := repo.db.Model(mod).Where("name = ?", name).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountUpstream(where string) (count uint, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM proxy_upstream proxy_upstream " + where)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) GetUpstreamListWithPage(limit uint, offset uint, where string) (mods []model.Upstream, err error) {
	sql := "SELECT * FROM  proxy_upstream" + where + " order by proxy_upstream.name ASC"
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

func (repo *MySQLRepo) GetUpstreamListByClusterId(id int) (mods []model.Upstream, err error) {
	sql := "SELECT * FROM  proxy_upstream where cluster_ids like'%" + fmt.Sprintf("%d", id) + "%' order by proxy_upstream.name ASC"
	err = repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) GetUpstreamIdList() ([]string, error) {
	var mods []model.Upstream
	var ids []string
	err := repo.db.Find(&mods).Error
	for _, v := range mods {
		ids = append(ids, strconv.Itoa(int(v.ID)))
	}
	return ids, err
}

func (repo *MySQLRepo) FormatUpstreamNameById(id int, content string, separator string) (string, error) {
	var mod model.Upstream
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
		parentContent, _ := repo.FormatUpstreamNameById(mod.ID, "", separator)
		content = parentContent + separator + content
	}
	return content, nil
}

func (repo *MySQLRepo) GetUpstreamNameById(id uint, content string, separator string) (string, error) {
	var mod model.Upstream
	if id <= uint(0) {
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

	return content, nil
}

func (repo *MySQLRepo) GetUpstreamById(id int) (*model.Upstream, error) {
	var mod model.Upstream
	err := repo.db.Where("id = ?", id).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetUpstreamIdByName(name string) (int, error) {
	mod := model.Upstream{Name: name}
	err := repo.db.Where("name = ?", name).Find(&mod).Error
	return mod.ID, err
}

func (repo *MySQLRepo) GetUpstreamUsedById(id int) (int, error) {
	var mod model.Upstream
	err := repo.db.Where("id = ?", id).Find(&mod).Error
	return mod.Used, err
}

func (repo *MySQLRepo) GetUpstreamNameIdListByCid(cid uint) ([]model.Upstream, error) {
	var mods []model.Upstream
	err := repo.db.Unscoped().Where("cluster_ids like '''%?%'''", cid).Find(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) GetUpstreamBackendsById(Id uint) (model.Upstream, error) {
	var mod model.Upstream
	err := repo.db.Unscoped().Where("id = ?", Id).Find(&mod).Error
	return mod, err
}
