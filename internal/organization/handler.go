package organization

import (
	"fmt"
	"net/http"
	"spsyncpro_api/pkg/domain"
	"spsyncpro_api/pkg/msgraphapi"
	"spsyncpro_api/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type OrganizationHandler struct {
	organizationService    domain.OrganizationService
	organizationRepository domain.OrganizationRepository
	tracer                 trace.Tracer
}

func NewOrganizationHandler(
	organizationService domain.OrganizationService,
	organizationRepository domain.OrganizationRepository,
) *OrganizationHandler {
	tracer := otel.Tracer("organizationHandler")
	return &OrganizationHandler{
		organizationService:    organizationService,
		organizationRepository: organizationRepository,
		tracer:                 tracer,
	}
}

type UpsertOrganizationRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ClientID     string `json:"client_id"`
	TenantID     string `json:"tenant_id"`
	ClientSecret string `json:"client_secret"`
}

type UpsertOrganizationResponse struct {
	ID           uint   `json:"id"`
	IsAuthorized bool   `json:"is_authorized"`
	AuthorizeURL string `json:"authorize_url"`
}

// @Summary		Upsert an organization
// @Description	Upsert an organization
// @Tags			organization
// @Accept			json
// @Produce		json
// @Param			organization	body		UpsertOrganizationRequest	true	"Organization"
// @Success		200		{object}	UpsertOrganizationResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/organization/upsert [post]
func (h *OrganizationHandler) UpsertOrganization(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "UpsertOrganization")
	defer span.End()

	var req UpsertOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accountID := c.GetUint(utils.AccountIdContextKey)
	if accountID == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	clientSecret, err := h.organizationService.EncryptClientSecret(ctx, req.ClientSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newOrg := &domain.Organization{
		Name:         req.Name,
		Description:  req.Description,
		OwnerID:      accountID,
		ClientID:     req.ClientID,
		TenantID:     req.TenantID,
		ClientSecret: clientSecret,
	}

	newOrg, err = h.organizationRepository.UpsertOrganization(ctx, newOrg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	msGraphApiService := msgraphapi.NewMsGraphApiService(msgraphapi.MsGraphApiConfig{
		ClientID:     newOrg.ClientID,
		TenantID:     newOrg.TenantID,
		ClientSecret: clientSecret,
	})

	ok, err := msGraphApiService.CheckAuthorized(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, UpsertOrganizationResponse{
		ID:           newOrg.ID,
		IsAuthorized: ok,
		AuthorizeURL: fmt.Sprintf("https://login.microsoftonline.com/%s/adminconsent?client_id=%s", newOrg.TenantID, newOrg.ClientID),
	})
}

type GetOrganizationResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ClientID     string `json:"client_id"`
	TenantID     string `json:"tenant_id"`
	IsAuthorized bool   `json:"is_authorized"`
}

// @Summary		Get an organization
// @Description	Get an organization
// @Tags			organization
// @Accept			json
// @Produce		json
// @Success		200		{object}	GetOrganizationResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/organization/get [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "GetOrganization")
	defer span.End()

	accountID := c.GetUint(utils.AccountIdContextKey)
	if accountID == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	organization, err := h.organizationRepository.GetOrganizationByOwnerID(ctx, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetOrganizationResponse{
		ID:           organization.ID,
		Name:         organization.Name,
		Description:  organization.Description,
		ClientID:     organization.ClientID,
		TenantID:     organization.TenantID,
		IsAuthorized: organization.IsAuthorized,
	})
}

type DeleteOrganizationResponse struct {
	Message string `json:"message"`
}

// @Summary		Delete an organization
// @Description	Delete an organization
// @Tags			organization
// @Accept			json
// @Produce		json
// @Success		200		{object}	DeleteOrganizationResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/organization/delete [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "DeleteOrganization")
	defer span.End()

	accountID := c.GetUint(utils.AccountIdContextKey)
	if accountID == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err := h.organizationRepository.DeleteOrganizationByOwnerID(ctx, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DeleteOrganizationResponse{
		Message: "organization deleted successfully",
	})
}

type CheckAuthorizationResponse struct {
	Message string `json:"message"`
	// https://login.microsoftonline.com/${tenantId}/adminconsent?client_id=${clientId}
	AuthorizeURL string `json:"authorize_url"`
}

// @Summary Check Authorization
// @Description Check Authorization
// @Tags			organization
// @Accept			json
// @Produce		json
// @Success		200		{object}	CheckAuthorizationResponse
// @Failure		400		{object}	map[string]string
// @Failure		500		{object}	map[string]string
// @Router			/api/v1/organization/check-authorization [get]
func (h *OrganizationHandler) CheckAuthorization(c *gin.Context) {
	ctx := c.Request.Context()
	ctx, span := h.tracer.Start(ctx, "CheckAuthorization")
	defer span.End()

	accountID := c.GetUint(utils.AccountIdContextKey)
	if accountID == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	organization, err := h.organizationRepository.GetOrganizationByOwnerID(ctx, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clientSecret, err := h.organizationService.DecryptClientSecret(ctx, organization.ClientSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	organization.ClientSecret = clientSecret

	msGraphApiService := msgraphapi.NewMsGraphApiService(msgraphapi.MsGraphApiConfig{
		ClientID:     organization.ClientID,
		TenantID:     organization.TenantID,
		ClientSecret: clientSecret,
	})

	ok, err := msGraphApiService.CheckAuthorized(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if ok {
		c.JSON(http.StatusOK, CheckAuthorizationResponse{
			Message:      "organization authorized",
			AuthorizeURL: fmt.Sprintf("https://login.microsoftonline.com/%s/adminconsent?client_id=%s", organization.TenantID, organization.ClientID),
		})
	} else {
		c.JSON(http.StatusOK, CheckAuthorizationResponse{
			Message:      "organization not authorized",
			AuthorizeURL: fmt.Sprintf("https://login.microsoftonline.com/%s/adminconsent?client_id=%s", organization.TenantID, organization.ClientID),
		})
	}

}
