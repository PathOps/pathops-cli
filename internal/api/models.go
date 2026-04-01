package api

type LoginResponse struct {
	Status string          `json:"status"`
	Data   LoginResultData `json:"data"`
}

type LoginResultData struct {
	UserID               int64  `json:"userId"`
	TenantID             int64  `json:"tenantId"`
	TenantName           string `json:"tenantName"`
	TenantSlug           string `json:"tenantSlug"`
	MembershipRole       string `json:"membershipRole"`
	IdentityChanged      bool   `json:"identityChanged"`
	RequiresTokenRefresh bool   `json:"requiresTokenRefresh"`
	RequiresToolRelogin  bool   `json:"requiresToolRelogin"`
}

type WhoAmI struct {
	Profile   string `json:"profile"`
	BaseURL   string `json:"baseUrl"`
	HasTokens bool   `json:"hasTokens"`
}