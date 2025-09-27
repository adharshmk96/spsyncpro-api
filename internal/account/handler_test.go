package account_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"go_starter_api/internal/account"
	"go_starter_api/pkg/domain"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"
	"gorm.io/gorm"
)

// HTTPTestHelper provides reusable functions for testing HTTP handlers
//
// Usage Examples:
//
//  1. Basic setup and request:
//     httpHelper := NewHTTPTestHelper()
//     httpHelper.SetupHandler("POST", "/account/register", handler.RegisterAccount)
//     w := httpHelper.MakeRequest("POST", "/account/register", reqBody, nil)
//
//  2. Authenticated request:
//     w := httpHelper.MakeAuthenticatedRequest("GET", "/account/profile", nil, "bearer_token")
//
//  3. Error response assertion:
//     httpHelper.AssertErrorResponse(t, w, http.StatusBadRequest, "account already exists")
//
//  4. JSON response assertion:
//     var response RegisterAccountResponse
//     httpHelper.AssertJSONResponse(t, w, &response)
//
//  5. Multiple handlers setup:
//     handlers := map[string]map[string]gin.HandlerFunc{
//     "POST": {
//     "/account/register": handler.RegisterAccount,
//     "/account/login": handler.LoginAccount,
//     },
//     "GET": {
//     "/account/profile": handler.GetProfile,
//     },
//     }
//     httpHelper.SetupMultipleHandlers(handlers)
type HTTPTestHelper struct {
	router *gin.Engine
}

// NewHTTPTestHelper creates a new HTTP test helper with Gin router in test mode
func NewHTTPTestHelper() *HTTPTestHelper {
	gin.SetMode(gin.TestMode)
	return &HTTPTestHelper{
		router: gin.Default(),
	}
}

// SetupHandler registers a handler with the router
func (h *HTTPTestHelper) SetupHandler(method, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		h.router.GET(path, handler)
	case "POST":
		h.router.POST(path, handler)
	case "PUT":
		h.router.PUT(path, handler)
	case "DELETE":
		h.router.DELETE(path, handler)
	case "PATCH":
		h.router.PATCH(path, handler)
	default:
		panic("unsupported HTTP method: " + method)
	}
}

// MakeRequest performs an HTTP request and returns the response
func (h *HTTPTestHelper) MakeRequest(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var jsonBody []byte
	var err error

	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			panic("failed to marshal request body: " + err.Error())
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	if err != nil {
		panic("failed to create request: " + err.Error())
	}

	// Set default content type for JSON requests
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	return w
}

// AssertJSONResponse asserts that the response is valid JSON and unmarshals it into the target
func (h *HTTPTestHelper) AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	assert.NotEmpty(t, w.Body.String())
	err := json.Unmarshal(w.Body.Bytes(), target)
	assert.NoError(t, err)
}

// MakeAuthenticatedRequest makes a request with Authorization header
func (h *HTTPTestHelper) MakeAuthenticatedRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	headers := map[string]string{
		"Authorization": token,
	}
	return h.MakeRequest(method, path, body, headers)
}

// SetupMultipleHandlers allows setting up multiple handlers at once
func (h *HTTPTestHelper) SetupMultipleHandlers(handlers map[string]map[string]gin.HandlerFunc) {
	for method, routes := range handlers {
		for path, handler := range routes {
			h.SetupHandler(method, path, handler)
		}
	}
}

func TestAccountHandler_RegisterAccount(t *testing.T) {

	anyContext := mock.MatchedBy(func(ctx context.Context) bool { return true })

	otel.SetTracerProvider(noop.NewTracerProvider())

	t.Run("should register account successfully", func(t *testing.T) {
		logger := logrus.New()
		service := domain.NewMockAccountService(t)
		repository := domain.NewMockAccountRepository(t)

		// Mock repository methods
		repository.On("GetAccountByEmail", anyContext, "test@example.com").Return(nil, gorm.ErrRecordNotFound)
		repository.On("CreateAccount", anyContext, mock.AnythingOfType("*domain.Account")).Return(&domain.Account{ID: 1, Email: "test@example.com"}, nil)
		repository.On("LogAccountActivity", anyContext, uint(1), domain.ActivityRegister).Return(nil)

		// Mock service methods
		service.On("HashPassword", anyContext, "password").Return("hashed_password", nil)
		service.On("GenerateAuthToken", anyContext, mock.AnythingOfType("*domain.Account")).Return("auth_token", nil)

		handler := account.NewAccountHandler(logger, service, repository)

		// Setup HTTP test helper
		httpHelper := NewHTTPTestHelper()
		httpHelper.SetupHandler("POST", "/account/register", handler.RegisterAccount)

		// Make request
		reqBody := account.RegisterAccountRequest{
			Email:    "test@example.com",
			Password: "password",
		}
		w := httpHelper.MakeRequest("POST", "/account/register", reqBody, nil)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response account.RegisterAccountResponse
		httpHelper.AssertJSONResponse(t, w, &response)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test@example.com", response.Email)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "auth_token", response.Token)
	})

	t.Run("should return error when account already exists", func(t *testing.T) {
		logger := logrus.New()
		service := domain.NewMockAccountService(t)
		repository := domain.NewMockAccountRepository(t)

		// Mock repository to return existing account
		existingAccount := &domain.Account{ID: 1, Email: "test@example.com"}
		repository.On("GetAccountByEmail", anyContext, "test@example.com").Return(existingAccount, nil)

		handler := account.NewAccountHandler(logger, service, repository)

		// Setup HTTP test helper
		httpHelper := NewHTTPTestHelper()
		httpHelper.SetupHandler("POST", "/account/register", handler.RegisterAccount)

		// Make request
		reqBody := account.RegisterAccountRequest{
			Email:    "test@example.com",
			Password: "password",
		}
		w := httpHelper.MakeRequest("POST", "/account/register", reqBody, nil)

		// Assertions
		var response map[string]string
		httpHelper.AssertJSONResponse(t, w, &response)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "account already exists", response["error"])
	})

}
