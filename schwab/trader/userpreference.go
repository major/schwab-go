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
	StreamerURL            string `json:"streamerUrl"`
	Token                  string `json:"token"`
	TokenExpirationTime    string `json:"tokenExpTime"`
	AppID                  string `json:"appId"`
	ACL                    string `json:"acl"`
	SchwabClientCustomerID string `json:"schwabClientCustomerId"`
	SchwabClientCorrelID   string `json:"schwabClientCorrelId"`
	SchwabClientChannel    string `json:"schwabClientChannel"`
	SchwabClientFunctionID string `json:"schwabClientFunctionId"`
}

// String returns a redacted representation of StreamerInfo for safe logging.
func (s StreamerInfo) String() string {
	return s.redactedString()
}

// GoString returns a redacted representation for %#v debug formatting.
func (s StreamerInfo) GoString() string {
	return s.redactedString()
}

func (s StreamerInfo) redactedString() string {
	return "{StreamerSocketURL:" + s.StreamerSocketURL +
		" StreamerURL:" + s.StreamerURL +
		" Token:<redacted> TokenExpirationTime:" + s.TokenExpirationTime +
		" AppID:" + s.AppID + " ACL:<redacted>" +
		" SchwabClientCustomerID:" + s.SchwabClientCustomerID +
		" SchwabClientCorrelID:" + s.SchwabClientCorrelID +
		" SchwabClientChannel:" + s.SchwabClientChannel +
		" SchwabClientFunctionID:" + s.SchwabClientFunctionID + "}"
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
