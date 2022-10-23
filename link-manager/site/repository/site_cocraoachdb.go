package repository

import (
	"context"
	"errors"
	"github.com/cockroachdb/cockroach-go/v2/crdb/crdbgorm"
	"github.com/google/uuid"
	"github.com/gookit/config/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type configdb struct {
	Url string `mapstructure:"url"`
}

type SiteRepository struct {
	db *gorm.DB
}

func NewSiteRepository() *SiteRepository {
	err := config.LoadFiles("config/config.yaml")
	if err != nil {
		panic(err)
	}
	conf := configdb{}
	err = config.BindStruct("cockroach", &conf)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(postgres.Open(conf.Url+"&application_name=$ docs_simplecrud_gorm"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Site{})
	return &SiteRepository{
		db: db,
	}
}

func (r *SiteRepository) Save(url, category string, parentID *uuid.UUID) (*uuid.UUID, error) {
	newID := uuid.New()
	if err := crdbgorm.ExecuteTx(context.Background(), r.db, nil,
		func(tx *gorm.DB) error {
			return r.db.Create(&Site{ID: newID, Status: "NEW", Category: category, ParentID: parentID, Url: url}).Error
		},
	); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *SiteRepository) GetByUrl(url string) (*Site, error) {
	var site Site
	t := r.db.First(&site, Site{Url: url})
	if t.Error != nil {
		if errors.Is(t.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, t.Error
	}
	return &site, nil
}

func (r *SiteRepository) UpdateStatusByUrl(url, status string) error {
	t := r.db.Model(&Site{}).Where("url = ?", url).Update("status", status)
	if t == nil || t.Error != nil {
		return t.Error
	}
	return nil
}

func (r *SiteRepository) GetWithParentsByUrl(url string) *Site {
	//db.Preload("Created").First(&result, user1.ID)
	return nil
}
