package mysqlrepo

import (
	"fmt"
	"strconv"

	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) AddDomain(mod model.Domain) (*model.Domain, error) {
	err := repo.db.Create(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) UpdateDomainById(id int, manager string, description string, cluster_ids string, proxy_type int, port_http string, port_https string, http2 bool, cert_id int, customize string) (*model.Domain, error) {
	mod := model.Domain{Manager: manager, Description: description, ClusterIds: cluster_ids, ProxyType: proxy_type, PortHttp: port_http, PortHttps: port_https, Http2: http2, CertId: cert_id, Customize: customize}
	//err := repo.db.Model(&mod).Where("id=?",id).Update(&mod).Error
	err := repo.db.First(&mod, id).Update("manager", manager).Update("description", description).Update("cluster_ids", cluster_ids).Update("proxy_type", proxy_type).Update("port_http", port_http).Update("port_https", port_https).Update("http2", http2).Update("cert_id", cert_id).Update("customize", customize).Error
	return &mod, err
}

func (repo *MySQLRepo) DeleteDomainById(id int) (*model.Domain, error) {
	mod := model.Domain{}
	err := repo.db.Unscoped().Where("id = ?", id).Delete(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) CountDomainByName(name string) (int, error) {
	mod := model.Domain{Name: name}
	var count int
	err := repo.db.Model(mod).Where("name = ?", name).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountDomain(where string) (count int, err error) {
	row := repo.db.DB().QueryRow("SELECT count(1) FROM proxy_domain proxy_domain " + where)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) GetDomainListWithPage(limit uint, offset uint, where string) (mods []model.Domain, err error) {
	sql := "SELECT * FROM  proxy_domain" + where + " order by name"
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

func (repo *MySQLRepo) GetDomainIDList() ([]string, error) {
	var mods []model.Domain
	var ids []string
	err := repo.db.Find(&mods).Error
	for _, v := range mods {
		ids = append(ids, strconv.Itoa(v.ID))
	}
	return ids, err
}

func (repo *MySQLRepo) FormatDomainNameById(id int, content string, separator string) (string, error) {
	var mod model.Domain
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
		parentContent, _ := repo.FormatDomainNameById(mod.ID, "", separator)
		content = parentContent + separator + content
	}
	return content, nil
}

func (repo *MySQLRepo) GetDomainNameById(id int) (string, error) {
	var mod model.Domain
	err := repo.db.Unscoped().Where("id = ?", id).Find(&mod).Error
	return mod.Name, err
}

func (repo *MySQLRepo) GetDomainById(id int) (*model.Domain, error) {
	var mod model.Domain
	err := repo.db.Where("id = ?", id).Find(&mod).Error
	return &mod, err
}

func (repo *MySQLRepo) GetDomainIdByName(name string) (int, error) {
	mod := model.Domain{Name: name}
	err := repo.db.Where("name = ?", name).Find(&mod).Error
	return mod.ID, err
}

func (repo *MySQLRepo) GetDomainListByClusterId(id int) (mods []model.Domain, err error) {
	countsql := "select max(length(replace(name,'.','..'))-length(name)) from proxy_domain where cluster_ids like'%" + fmt.Sprintf("%d", id) + "%';"
	var count int
	row := repo.db.DB().QueryRow(countsql)
	err = row.Scan(&count)

	sql := "SELECT * FROM  proxy_domain where cluster_ids like'%" + fmt.Sprintf("%d", id) + "%' order by LPAD(SUBSTRING_INDEX(name,\".\",-1),32,\" \")"
	for i := 0; i < int(count); i++ {
		sql += ",LPAD(SUBSTRING_INDEX(name,\".\"," + fmt.Sprintf("-%d", i+2) + "),32,\" \")"
	}

	err = repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) GetDomainListByCertId(id int) (mods []model.DomainUs, err error) {
	countsql := "select max(length(replace(name,'.','..'))-length(name)) from proxy_domain where cert_id='%" + fmt.Sprintf("%d", id) + "%';"
	var count int
	row := repo.db.DB().QueryRow(countsql)
	err = row.Scan(&count)

	sql := "select t1.id,t1.name,t1.cluster_ids,t1.manager,t2.name from proxy_domain t1 left join proxy_cluster t2 on t1.cluster_ids like t2.id where t1.cert_id='" + fmt.Sprintf("%d", id) + "' order by LPAD(SUBSTRING_INDEX(t1.name,\".\",-1),32,\" \")"
	for i := 0; i < int(count); i++ {
		sql += ",LPAD(SUBSTRING_INDEX(t1.name,\".\"," + fmt.Sprintf("-%d", i+2) + "),32,\" \")"
	}

	err = repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}

func (repo *MySQLRepo) CountDomainByCertId(id int) (int, error) {
	mod := model.Domain{CertId: id}
	var count int
	err := repo.db.Model(mod).Where("cert_id = ?", id).Count(&count).Error
	return count, err
}

func (repo *MySQLRepo) CountDomainByCertIdAndClusterId(cert_id int, clusterid string) (int, error) {
	sql := "select count(*) from proxy_domain where cert_id='" + strconv.Itoa(cert_id) + "' and cluster_ids like'%" + clusterid + "%'"
	row := repo.db.DB().QueryRow(sql)
	var count int
	err := row.Scan(&count)
	return count, err
}
