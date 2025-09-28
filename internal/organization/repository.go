package organization

import (
	"context"
	"errors"
	"spsyncpro_api/pkg/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type OrganizationRepo struct {
	db    *gorm.DB
	trace trace.Tracer
}

func NewOrganizationRepository(db *gorm.DB) domain.OrganizationRepository {
	trace := otel.Tracer("organizationRepository")
	return &OrganizationRepo{
		db:    db,
		trace: trace,
	}
}

func (r *OrganizationRepo) UpsertOrganization(ctx context.Context, organization *domain.Organization) (*domain.Organization, error) {
	_, span := r.trace.Start(ctx, "UpsertOrganization")
	defer span.End()
	err := r.db.Where("owner_id = ?", organization.OwnerID).First(organization).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.db.Create(organization).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		err = r.db.Save(organization).Error
		if err != nil {
			return nil, err
		}
	}

	return organization, nil
}

func (r *OrganizationRepo) GetOrganizationByOwnerID(ctx context.Context, ownerID uint) (*domain.Organization, error) {
	_, span := r.trace.Start(ctx, "GetOrganizationByOwnerID")
	defer span.End()
	var organization domain.Organization
	err := r.db.Where("owner_id = ?", ownerID).First(&organization).Error
	if err != nil {
		return nil, err
	}
	return &organization, nil
}

func (r *OrganizationRepo) DeleteOrganizationByOwnerID(ctx context.Context, ownerID uint) error {
	_, span := r.trace.Start(ctx, "DeleteOrganizationByOwnerID")
	defer span.End()
	return r.db.Delete(&domain.Organization{}, ownerID).Error
}
