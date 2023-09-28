package mysqlrepo

import (
	"strconv"

	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) AddRoute(mod model.Route) (*model.Route, error) {
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateRouteById(id int, index int, route string, manager string, description string, domain_id int, match_type string, customize string, upstream_id int) (*model.Route, error) {
	mod := model.Route{Index: index, Route: route, Manager: manager, Description: description, DomainId: domain_id, MatchType: match_type, Customize: customize, UpstreamId: upstream_id}
	err := repo.db.First(&mod, id).Update("idx", index).Update("route", route).Update("manager", manager).Update("description", description).Update("domain_id", domain_id).Update("match_type", match_type).Update("customize", customize).Update("upstream_id", upstream_id).Error
	return &mod, err
}

func (repo *MySQLRepo) DeleteRouteById(id int) error {
	var mod model.Route
	err := repo.db.Unscoped().Where("id = ?", id).Delete(&mod).Error
	return err
}

func (repo *MySQLRepo) DeleteRouteByDomainId(id int) error {
	mod := model.Route{}
	err := repo.db.Unscoped().Where("domain_id = ?", id).Delete(&mod).Error
	return err
}

func (repo *MySQLRepo) GetRouteListByDomainId(id int) ([]model.Route, error) {
	var mods []model.Route
	err := repo.db.Where("domain_id = ?", id).Find(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) GetRouteListByUpstreamId(id int) ([]model.RouteUs, error) {
	var mods []model.RouteUs
	sql := "SELECT t1.id,t1.route,t1.manager,t1.match_type,t2.id,t2.name,t2.manager,GROUP_CONCAT(t3.name SEPARATOR '\n') as name from proxy_route t1 left join proxy_domain t2 on t2.id=t1.domain_id left join proxy_cluster t3 on find_in_set(t3.id,t2.cluster_ids) where t1.upstream_id='" + strconv.Itoa(int(id)) + "' group by t1.id order by t3.name asc,t2.name asc,t1.route asc"
	err := repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) GetRouteFullListByDomainId(id int) ([]model.RouteFull, error) {
	var mods []model.RouteFull
	sql := "SELECT t1.id,t1.route,t1.manager,t1.idx,t1.match_type,t1.customize,t1.description,t1.updated_at,t2.name,t2.id,t2.backends from proxy_route t1 left join proxy_upstream t2 on t2.id=t1.upstream_id where t1.domain_id='" + strconv.Itoa(int(id)) + "' and t1.deleted_at IS NULL order by t1.idx"
	err := repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) CountRouteByUpstreamId(upstream_id int) (int, error) {
	mod := model.Route{UpstreamId: upstream_id}
	var count int
	err := repo.db.Model(mod).Where("upstream_id = ?", upstream_id).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountRouteByDomainId(id int) (count int, err error) {
	mod := model.Route{DomainId: id}
	err = repo.db.Model(mod).Where("domain_id = ?", id).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountRouteByUpstreamIdAndClusterId(upstream_id int, clusterid string) (int, error) {
	sql := "select count(*) from proxy_route join proxy_domain on proxy_route.domain_id=proxy_domain.id and upstream_id='" + strconv.Itoa(upstream_id) + "' and cluster_ids like'" + clusterid + "'"
	row := repo.db.DB().QueryRow(sql)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (repo *MySQLRepo) GetRouteNameById(id int) (string, error) {
	var mod model.Route
	err := repo.db.Unscoped().Where("id = ?", id).Find(&mod).Error
	return mod.Route, err
}
