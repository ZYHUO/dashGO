package repository

import (
	"dashgo/internal/model"

	"gorm.io/gorm"
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) Get(key string) (string, error) {
	var setting model.Setting
	err := r.db.Where("`key` = ?", key).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // 设置项不存在时返回空字符串，不报�?
		}
		return "", err
	}
	return setting.Value, nil
}

func (r *SettingRepository) Set(key, value string) error {
	var setting model.Setting
	err := r.db.Where("`key` = ?", key).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		setting = model.Setting{Key: key, Value: value}
		return r.db.Create(&setting).Error
	}
	if err != nil {
		return err
	}
	setting.Value = value
	return r.db.Save(&setting).Error
}

func (r *SettingRepository) GetAll() (map[string]string, error) {
	var settings []model.Setting
	err := r.db.Find(&settings).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

func (r *SettingRepository) Delete(key string) error {
	return r.db.Where("`key` = ?", key).Delete(&model.Setting{}).Error
}
