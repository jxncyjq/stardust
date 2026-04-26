package databases

import (
	"gorm.io/gorm"
)

// CheckWhereInfo 检查where条件信息并构建查询
func CheckWhereInfo[T any](w string, v *T, tp string, db *gorm.DB) *gorm.DB {
	if v != nil {
		switch tp {
		case "and":
			return db.Where(w, *v)
		case "in":
			return db.Where(w+" IN ?", *v)
		case "or":
			return db.Or(w, *v)
		case "notin":
			return db.Where(w+" NOT IN ?", *v)
		case "where":
			return db.Where(w, *v)
		default:
			return db
		}
	}
	return db
}
