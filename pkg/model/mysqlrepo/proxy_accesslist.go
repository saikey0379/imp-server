package mysqlrepo

import (
	"fmt"
	"strconv"

	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) GetAccesslistListWithPage(limit uint, offset uint, where string) (mods []model.Accesslist, err error) {
	sql := "select * from proxy_accesslist " + where + " group by cluster_id,domain_id,route_id,access_type order by cluster_id,domain_id,route_id,access_type"
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

func (repo *MySQLRepo) CountAccesslist(where string) (count int, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM (select id from  proxy_accesslist " + where + " group by proxy_accesslist.domain_id,proxy_accesslist.route_id) list")
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) CountAccesslistByDomainId(domain_id int) (count int, err error) {
	mod := model.Accesslist{DomainId: domain_id}
	err = repo.db.Model(mod).Where("domain_id = ?", domain_id).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountAccesslistByDomainIdAndRouteId(domain_id int, route_id int) (count int, err error) {
	mod := model.Accesslist{DomainId: domain_id, RouteId: route_id}
	err = repo.db.Model(mod).Where("domain_id = ? and route_id = ?", domain_id, route_id).Count(&count).Error
	return count, err
}
func (repo *MySQLRepo) GetAccesslistListByDomainIdAndRouteId(domain_id int, route_id int) ([]model.Accesslist, error) {
	var mods []model.Accesslist
	sql := "SELECT * from proxy_accesslist where proxy_accesslist.domain_id='" + fmt.Sprintf("%d", domain_id) + "' and proxy_accesslist.route_id='" + strconv.Itoa(int(route_id)) + "'"
	err := repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}
func (repo *MySQLRepo) GetAccesslistByClusterIdAndDomainIdAndRouteIdAndAccessType(cluster_id int, domain_id int, route_id int, access_type string) ([]model.Accesslist, error) {
	var mods []model.Accesslist
	sql := "SELECT * from proxy_accesslist where proxy_accesslist.cluster_id='" + fmt.Sprintf("%d", cluster_id) + "' and proxy_accesslist.domain_id='" + strconv.Itoa(int(domain_id)) + "' and proxy_accesslist.route_id='" + strconv.Itoa(int(route_id)) + "' and proxy_accesslist.access_type='" + access_type + "'"
	err := repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) AddAccesslist(mod model.Accesslist) (*model.Accesslist, error) {
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateAccesslistById(id int, host string, description string, manager string) (*model.Accesslist, error) {
	mod := model.Accesslist{ID: id, Host: host, Description: description, Manager: manager}
	err := repo.db.First(&mod, id).Update("host", host).Update("description", description).Update("manager", manager).Error
	return &mod, err
}

func (repo *MySQLRepo) DeleteAccesslistById(id int) (*model.Accesslist, error) {
	var mod model.Accesslist
	err := repo.db.Unscoped().Where("id = ?", id).Delete(&mod).Error
	return &mod, err
}
