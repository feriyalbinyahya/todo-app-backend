package models

import "gorm.io/gorm"

type BlacklistedToken struct {
	gorm.Model
	Token string `json:"token" gorm:"unique;not null"`
}
