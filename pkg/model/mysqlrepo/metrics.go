package mysqlrepo

import (
	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) GetMetricListCert() ([]model.Cert, error) {
	var result []model.Cert
	sql := "SELECT name,manager,not_after from proxy_cert order by name;"

	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}
