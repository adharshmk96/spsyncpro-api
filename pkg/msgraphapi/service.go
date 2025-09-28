package msgraphapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type MsGraphApiConfig struct {
	ClientID     string `json:"client_id"`
	TenantID     string `json:"tenant_id"`
	ClientSecret string `json:"client_secret"`
}

type MsGraphApiService struct {
	Config      MsGraphApiConfig
	httpClient  *http.Client
	accessToken string
}

func NewMsGraphApiService(config MsGraphApiConfig) *MsGraphApiService {
	return &MsGraphApiService{
		Config:     config,
		httpClient: &http.Client{},
	}
}

const GRAPH_API_URL = "https://graph.microsoft.com/v1.0"

func (s *MsGraphApiService) CheckAuthorized(ctx context.Context) (bool, error) {
	accessToken, err := s.GetAccessToken(ctx)
	if err != nil {
		return false, err
	}

	return s.ValidateToken(ctx, accessToken)
}

func (s *MsGraphApiService) GetAccessToken(ctx context.Context) (string, error) {
	tokenUrl := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", s.Config.TenantID)

	formData := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {s.Config.ClientID},
		"client_secret": {s.Config.ClientSecret},
		"scope":         {"https://graph.microsoft.com/.default"},
	}

	response, err := http.PostForm(tokenUrl, formData)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	s.accessToken = result.AccessToken

	return s.accessToken, nil
}

func (s *MsGraphApiService) ValidateToken(ctx context.Context, token string) (bool, error) {
	siteUrl := fmt.Sprintf("%s/sites/root", GRAPH_API_URL)

	request, err := http.NewRequestWithContext(ctx, "GET", siteUrl, nil)
	if err != nil {
		return false, err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))

	response, err := s.httpClient.Do(request)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	return response.StatusCode == http.StatusOK, nil
}

type MsGraphResponse[T any] struct {
	Context string `json:"@odata.context"`
	Value   []T    `json:"value"`
	Next    string `json:"@odata.nextLink"`
}
