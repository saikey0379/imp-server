package mysqlrepo

import (
	"fmt"
	"github.com/saikey0379/imp-server/pkg/model"
)

func (repo *MySQLRepo) CountDeviceByStatus(where string) (uint, error) {
	var condition string
	if where != "" {
		condition = " where " + where
	}

	row := repo.db.DB().QueryRow("SELECT count(t1.id) as count FROM devices t1 " + condition)
	var count uint
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *MySQLRepo) GetDeviceCabinetByLocationId(id uint) ([]model.DeviceCabinet, error) {
	var mods []model.DeviceCabinet
	where := fmt.Sprintf(" where t1.location_id = %s and t2.is_vm = 'No'", id)
	sql := "SELECT t1.id,t1.sn,t1.ip,t1.hostname,t1.location_u,t1.status,t2.updated_at from devices t1 left join manufacturers t2 on t1.sn=t2.sn " + where + " order by -location_u"
	//	sql := "SELECT t1.ip,t1.location_u,t1.status,t2.updated_at from devices t1 left join manufacturers t2 on t1.sn=t2.sn " + condition + " order by -location_u"
	err := repo.db.Raw(sql).Scan(&mods).Error
	return mods, err
}
