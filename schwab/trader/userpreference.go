package trader

import (
	"context"
)

// UserPreference contains user account preferences and streaming configuration.
type UserPreference struct {
	Accounts     []PreferenceAccount `json:"accounts"`
	StreamerInfo []StreamerInfo      `json:"streamerInfo"`
	Offers       []Offer             `json:"offers"`
}

// PreferenceAccount contains preferences for a linked account.
type PreferenceAccount struct {
	AccountNumber      string `json:"accountNumber"`
	PrimaryAccount     bool   `json:"primaryAccount"`
	Type               string `json:"type"`
	NickName           string `json:"nickName"`
	AccountColor       string `json:"accountColor"`
	DisplayAcctID      string `json:"displayAcctId"`
	AutoPositionEffect bool   `json:"autoPositionEffect"`
}

// StreamerInfo contains streaming connection configuration.
type StreamerInfo struct {
	StreamerSocketURL      string `json:"streamerSocketUrl"`
	SchwabClientCustomerID string `json:"schwabClientCustomerId"`
	SchwabClientCorrelID   string `json:"schwabClientCorrelId"`
	SchwabClientChannel    string `json:"schwabClientChannel"`
	SchwabClientFunctionID string `json:"schwabClientFunctionId"`
}

// Offer represents an available service offer.
type Offer struct {
	Level2Permissions bool   `json:"level2Permissions"`
	MktDataPermission string `json:"mktDataPermission"`
}

// GetUserPreference retrieves the authenticated user's account preferences and streaming configuration.
// The API returns an array of UserPreference objects.
func (c *Client) GetUserPreference(ctx context.Context) ([]UserPreference, error) {
	req, err := c.newRequest(ctx, "GET", "/userPreference", nil)
	if err != nil {
		return nil, err
	}

	var result []UserPreference
	if doErr := c.do(req, &result); doErr != nil {
		return nil, doErr
	}
	return result, nil
}
