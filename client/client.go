package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
)

const (
	baseURL    = "https://truthsocial.com"
	apiBaseURL = "https://truthsocial.com/api"

	// OAuth client credentials extracted from Truth Social's JavaScript
	// These are the same credentials used by Stanford Truthbrush
	clientID     = "9X1Fdd-pxNsAgEDNi_SfhJWi8T-vLuV2WVzKIbkTCw4"
	clientSecret = "ozF8jzI4968oTKFkEnsBC-UbLPCdrSv0MkXGQu2o_-M"

	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
)

type Client struct {
	client      cycletls.CycleTLS
	accessToken string
	user        string
	pass        string
}

func NewClient(ctx context.Context, user, pass string) (*Client, error) {
	// Initialize CycleTLS client
	client := cycletls.Init()

	c := &Client{
		client: client,
		user:   user,
		pass:   pass,
	}

	// Authenticate using the same method as Truthbrush
	if err := c.authenticate(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return c, nil
}

func (c *Client) authenticate(_ context.Context) error {
	// Use the same authentication approach as Stanford Truthbrush
	authURL := baseURL + "/oauth/token"

	payload := map[string]any{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"grant_type":    "password",
		"username":      c.user,
		"password":      c.pass,
		"redirect_uri":  "urn:ietf:wg:oauth:2.0:oob",
		"scope":         "read",
	}

	// Convert payload to JSON string
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := c.client.Do(authURL, cycletls.Options{
		Body:      string(payloadBytes),
		Method:    "POST",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Content-Type":    "application/json",
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "POST")
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}

	if resp.Status != 200 {
		// Check if this is a Cloudflare block
		if resp.Status == 403 && strings.Contains(resp.Body, "Cloudflare") {
			return fmt.Errorf("blocked by Cloudflare (status 403) - try using a VPN or different IP address")
		}

		// For other errors, show a truncated response
		body := resp.Body
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		return fmt.Errorf("authentication failed: status %d - %s", resp.Status, body)
	}

	var authResp AuthResponse
	if err := json.Unmarshal([]byte(resp.Body), &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if authResp.AccessToken == "" {
		return fmt.Errorf("no access token received")
	}

	c.accessToken = authResp.AccessToken
	return nil
}

func (c *Client) Lookup(ctx context.Context, username string) (*Account, error) {
	username = strings.TrimPrefix(username, "@")
	lookupURL := fmt.Sprintf("%s/v1/accounts/lookup?acct=%s", apiBaseURL, username)

	resp, err := c.client.Do(lookupURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("account lookup request failed: %w", err)
	}

	if resp.Status != 200 {
		// Check if this is a Cloudflare block
		if resp.Status == 403 && strings.Contains(resp.Body, "Cloudflare") {
			return nil, fmt.Errorf("blocked by Cloudflare (status 403) - try using a VPN or different IP address")
		}

		// For other errors, show a truncated response
		body := resp.Body
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		return nil, fmt.Errorf("account lookup failed: status %d - %s", resp.Status, body)
	}

	var account Account
	if err := json.Unmarshal([]byte(resp.Body), &account); err != nil {
		return nil, fmt.Errorf("failed to parse account data: %w", err)
	}

	return &account, nil
}

// PullStatuses implements the same method as Stanford Truthbrush
// Returns posts in reverse chronological order (recent first)
func (c *Client) PullStatuses(ctx context.Context, username string, excludeReplies bool, limit int) ([]Status, error) {
	// First lookup the user to get their ID
	account, err := c.Lookup(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user %s: %w", username, err)
	}

	var allStatuses []Status
	var maxID string
	pageCounter := 0

	for {
		// Build URL with parameters
		statusURL := fmt.Sprintf("%s/v1/accounts/%s/statuses", apiBaseURL, account.ID)
		params := url.Values{}

		if excludeReplies {
			params.Set("exclude_replies", "true")
		}

		if maxID != "" {
			params.Set("max_id", maxID)
		}

		// Set a reasonable page size
		params.Set("limit", "40")

		if len(params) > 0 {
			statusURL += "?" + params.Encode()
		}

		resp, err := c.client.Do(statusURL, cycletls.Options{
			Method:    "GET",
			Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			UserAgent: userAgent,
			Headers: map[string]string{
				"Authorization":   "Bearer " + c.accessToken,
				"Accept":          "application/json",
				"Accept-Language": "en-US,en;q=0.9",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT":             "1",
				"Connection":      "keep-alive",
				"Sec-Fetch-Dest":  "empty",
				"Sec-Fetch-Mode":  "cors",
				"Sec-Fetch-Site":  "same-origin",
			},
		}, "GET")
		if err != nil {
			return nil, fmt.Errorf("statuses request failed: %w", err)
		}

		if resp.Status != 200 {
			// Check if this is a Cloudflare block
			if resp.Status == 403 && strings.Contains(resp.Body, "Cloudflare") {
				return nil, fmt.Errorf("blocked by Cloudflare (status 403) - try using a VPN or different IP address")
			}

			// For other errors, show a truncated response
			body := resp.Body
			if len(body) > 200 {
				body = body[:200] + "..."
			}
			return nil, fmt.Errorf("statuses request failed: status %d - %s", resp.Status, body)
		}

		var statuses []Status
		if err := json.Unmarshal([]byte(resp.Body), &statuses); err != nil {
			return nil, fmt.Errorf("failed to parse statuses data: %w", err)
		}

		// If no statuses returned, we've reached the end
		if len(statuses) == 0 {
			break
		}

		// Add statuses to our collection
		allStatuses = append(allStatuses, statuses...)
		pageCounter++

		// Check if we've reached our limit
		if limit > 0 && len(allStatuses) >= limit {
			// Trim to exact limit
			if len(allStatuses) > limit {
				allStatuses = allStatuses[:limit]
			}
			break
		}

		// Set maxID for next page (oldest post ID from current page)
		if len(statuses) > 0 {
			lastStatus := statuses[len(statuses)-1]
			maxID = lastStatus.ID
		}

		// Safety check to prevent infinite loops
		if pageCounter > 50 {
			break
		}

		// Add a small delay between requests to be respectful
		time.Sleep(500 * time.Millisecond)
	}

	return allStatuses, nil
}

// GetStatuses is a simpler method for getting recent statuses
func (c *Client) GetStatuses(ctx context.Context, accountID string, limit int) ([]Status, error) {
	statusURL := fmt.Sprintf("%s/v1/accounts/%s/statuses?limit=%d&exclude_replies=true&exclude_reblogs=true", apiBaseURL, accountID, limit)

	resp, err := c.client.Do(statusURL, cycletls.Options{
		Method:    "GET",
		Ja3:       "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
		UserAgent: userAgent,
		Headers: map[string]string{
			"Authorization":   "Bearer " + c.accessToken,
			"Accept":          "application/json",
			"Accept-Language": "en-US,en;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}, "GET")
	if err != nil {
		return nil, fmt.Errorf("statuses request failed: %w", err)
	}

	if resp.Status != 200 {
		// Check if this is a Cloudflare block
		if resp.Status == 403 && strings.Contains(resp.Body, "Cloudflare") {
			return nil, fmt.Errorf("blocked by Cloudflare (status 403) - try using a VPN or different IP address")
		}

		// For other errors, show a truncated response
		body := resp.Body
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		return nil, fmt.Errorf("statuses request failed: status %d - %s", resp.Status, body)
	}

	var statuses []Status
	if err := json.Unmarshal([]byte(resp.Body), &statuses); err != nil {
		return nil, fmt.Errorf("failed to parse statuses data: %w", err)
	}

	return statuses, nil
}

func (c *Client) Close() {
	// Safely close the CycleTLS client
	// Use recover to prevent panic from nil channel in CycleTLS
	defer func() {
		if r := recover(); r != nil {
			// Silently handle any panic during close
			// This prevents the nil channel panic from CycleTLS
		}
	}()

	c.client.Close()
}
