package db

import (
	"pethealth/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func NewPostgresDB(cfg config.ShardConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.MasterDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(cfg.MasterDSN)},
		Replicas: []gorm.Dialector{postgres.Open(cfg.ReplicaDSN)},
		Policy:   dbresolver.RandomPolicy{},
	}))

	return db, err
}
