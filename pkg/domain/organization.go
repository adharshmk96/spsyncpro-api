package domain

import (
	"context"

	"gorm.io/gorm"
)

type Organization struct {
	gorm.Model
	OwnerID      uint    `json:"owner_id"`
	Owner        Account `json:"owner" gorm:"foreignKey:OwnerID"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	IsAuthorized bool    `json:"is_authorized"`
	ClientID     string  `json:"client_id"`
	TenantID     string  `json:"tenant_id"`
	ClientSecret string  `json:"client_secret"`
}

type OrganizationRepository interface {
	UpsertOrganization(ctx context.Context, organization *Organization) (*Organization, error)
	GetOrganizationByOwnerID(ctx context.Context, ownerID uint) (*Organization, error)
	DeleteOrganizationByOwnerID(ctx context.Context, ownerID uint) error
}

type OrganizationService interface {
	EncryptClientSecret(ctx context.Context, clientSecret string) (string, error)
	DecryptClientSecret(ctx context.Context, clientSecret string) (string, error)
}
