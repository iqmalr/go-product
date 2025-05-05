package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string 		`json:"name"`
	Description string 		`json:"description"`
	Price       int    		`json:"price"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
    p.ID = uuid.New()
    return
}