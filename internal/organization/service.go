package organization

import (
	"context"
	"spsyncpro_api/pkg/domain"
	"spsyncpro_api/pkg/utils"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type OrganizationService struct {
	tracer    trace.Tracer
	encryptor *utils.Encryptor
}

func NewOrganizationService() domain.OrganizationService {
	encryptor, err := utils.NewEncryptor([]byte(viper.GetString("ENCRYPTION_KEY")))
	if err != nil {
		panic(err)
	}
	tracer := otel.Tracer("organizationService")
	return &OrganizationService{
		tracer:    tracer,
		encryptor: encryptor,
	}
}

func (s *OrganizationService) EncryptClientSecret(ctx context.Context, clientSecret string) (string, error) {
	_, span := s.tracer.Start(ctx, "EncryptClientSecret")
	defer span.End()
	return clientSecret, nil
}

func (s *OrganizationService) DecryptClientSecret(ctx context.Context, clientSecret string) (string, error) {
	_, span := s.tracer.Start(ctx, "DecryptClientSecret")
	defer span.End()
	return clientSecret, nil
}
