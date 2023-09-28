package mysqlrepo

import (
	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) GetSystemConfigList() ([]model.Sysc, error) {
	var result []model.Sysc
	sql := "SELECT DISTINCT id,name from system_configs order by LPAD(name,255,0) ASC"
	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetHardwareList() ([]model.Hwc, error) {
	var result []model.Hwc
	sql := "SELECT DISTINCT company from hardwares order by company,LPAD(model_name,255,0) ASC"
	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetCompanyModelList() ([]model.Model, error) {
	var result []model.Model
	sql := "SELECT DISTINCT company from manufacturers order by company,LPAD(model_name,255,0) ASC"

	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetCpuSumList() ([]model.CpuSum, error) {
	var result []model.CpuSum
	sql := "SELECT DISTINCT cpu_sum from manufacturers order by LPAD(cpu_sum,11,0) ASC"

	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetMemorySumList() ([]model.MemSum, error) {
	var result []model.MemSum
	sql := "SELECT DISTINCT memory_sum from manufacturers order by LPAD(memory_sum,11,0) ASC"

	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetAgentVersionList() ([]model.VersionAgt, error) {
	var result []model.VersionAgt
	sql := "SELECT DISTINCT version_agt from manufacturers order by version_agt DESC"

	err := repo.db.Raw(sql).Scan(&result).Error
	return result, err
}

func (repo *MySQLRepo) GetManuModelNameByCompany(where string) ([]model.Model, error) {
	var result []model.Model
	err := repo.db.Raw("select model_name from manufacturers where " + where + " group by model_name order by LPAD(model_name,255,0),count(*) DESC").Scan(&result).Error
	return result, err
}
