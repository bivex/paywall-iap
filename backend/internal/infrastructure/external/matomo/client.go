package matomo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

const (
	// DefaultTimeout for HTTP requests
	DefaultTimeout = 30 * time.Second
	// MaxRetries for failed requests
	MaxRetries = 3
	// RetryDelay for retries
	RetryDelay = 500 * time.Millisecond
)

// Config represents Matomo configuration
type Config struct {
	BaseURL    string `json:"base_url"`
	SiteID     string `json:"site_id"`
	TokenAuth  string `json:"token_auth"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
}

// Client represents a Matomo HTTP client
type Client struct {
	config     Config
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient creates a new Matomo HTTP client
func NewClient(config Config, logger *zap.Logger) *Client {
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = MaxRetries
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// TrackEventRequest represents a standard event tracking request
type TrackEventRequest struct {
	Category        string            `json:"category"`
	Action          string            `json:"action"`
	Name            string            `json:"name,omitempty"`
	Value           float64           `json:"value,omitempty"`
	UserID          string            `json:"user_id"`
	EventTime       time.Time         `json:"event_time,omitempty"`
	CustomVariables map[string]string `json:"custom_variables,omitempty"`
}

// TrackEvent tracks a standard event in Matomo
func (c *Client) TrackEvent(ctx context.Context, req TrackEventRequest) error {
	params := url.Values{}
	params.Set("rec", "1")
	params.Set("idsite", c.config.SiteID)
	params.Set("token_auth", c.config.TokenAuth)
	params.Set("e_c", req.Category)
	params.Set("e_a", req.Action)
	if req.Name != "" {
		params.Set("e_n", req.Name)
	}
	if req.Value > 0 {
		params.Set("e_v", fmt.Sprintf("%.2f", req.Value))
	}
	if req.UserID != "" {
		params.Set("cid", req.UserID) // Use cid for user ID (Matomo uses this as visitor ID)
	}

	// Add custom variables
	i := 1
	for key, value := range req.CustomVariables {
		params.Set(fmt.Sprintf("cvar[%d][0]", i), key)
		params.Set(fmt.Sprintf("cvar[%d][1]", i), value)
		i++
	}

	// Add timestamp if provided
	if !req.EventTime.IsZero() {
		params.Set("h", fmt.Sprintf("%d", req.EventTime.Hour()))
		params.Set("m", fmt.Sprintf("%d", req.EventTime.Minute()))
		params.Set("s", fmt.Sprintf("%d", req.EventTime.Second()))
	}

	// Add random string to prevent caching
	params.Set("rand", fmt.Sprintf("%d", time.Now().UnixNano()))

	return c.doRequest(ctx, "/matomo.php", params)
}

// TrackEcommerceRequest represents an ecommerce tracking request
type TrackEcommerceRequest struct {
	UserID       string             `json:"user_id"`
	Revenue      float64            `json:"revenue"`
	OrderID      string             `json:"order_id,omitempty"`
	Items        []EcommerceItem    `json:"items,omitempty"`
	EventTime    time.Time          `json:"event_time,omitempty"`
	CustomVars   map[string]string  `json:"custom_variables,omitempty"`
}

// EcommerceItem represents an item in an ecommerce transaction
type EcommerceItem struct {
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Category  string  `json:"category,omitempty"`
}

// TrackEcommerce tracks an ecommerce event (purchase) in Matomo
func (c *Client) TrackEcommerce(ctx context.Context, req TrackEcommerceRequest) error {
	params := url.Values{}
	params.Set("rec", "1")
	params.Set("idsite", c.config.SiteID)
	params.Set("token_auth", c.config.TokenAuth)
	params.Set("e_c", "ecommerce")
	params.Set("e_a", "purchase")
	params.Set("revenue", fmt.Sprintf("%.2f", req.Revenue))
	if req.UserID != "" {
		params.Set("cid", req.UserID)
	}
	if req.OrderID != "" {
		params.Set("ec_id", req.OrderID)
	}

	// Add items
	if len(req.Items) > 0 {
		itemsJSON, err := json.Marshal(req.Items)
		if err == nil {
			params.Set("ec_items", string(itemsJSON))
		}
	}

	// Add custom variables
	i := 1
	for key, value := range req.CustomVars {
		params.Set(fmt.Sprintf("cvar[%d][0]", i), key)
		params.Set(fmt.Sprintf("cvar[%d][1]", i), value)
		i++
	}

	// Add random string
	params.Set("rand", fmt.Sprintf("%d", time.Now().UnixNano()))

	return c.doRequest(ctx, "/matomo.php", params)
}

// CohortRequest represents a cohort analysis request
type CohortRequest struct {
	Segment      string    `json:"segment,omitempty"`
	DateFrom     time.Time `json:"date_from"`
	DateTo       time.Time `json:"date_to"`
	CohortPeriod string    `json:"cohort_period"` // "day", "week", "month"
}

// CohortResponse represents the cohort analysis response
type CohortResponse struct {
	Cohorts []CohortData `json:"cohorts"`
	Meta    CohortMeta   `json:"meta"`
}

// CohortData represents cohort data for a specific time period
type CohortData struct {
	Period        string                 `json:"period"`
	Retention     map[string]int         `json:"retention"`     // day0 -> 100%, day1 -> 85%, etc.
	SampleSize    int                    `json:"sample_size"`
	Metrics       map[string]float64     `json:"metrics"`
	CustomData    map[string]interface{} `json:"custom_data,omitempty"`
}

// CohortMeta represents metadata about the cohort response
type CohortMeta struct {
	TotalUsers    int       `json:"total_users"`
	AverageRetention float64 `json:"average_retention"`
	DateFrom      time.Time `json:"date_from"`
	DateTo        time.Time `json:"date_to"`
}

// GetCohorts retrieves cohort analysis data from Matomo
func (c *Client) GetCohorts(ctx context.Context, req CohortRequest) (*CohortResponse, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "CustomReports.getCohorts")
	params.Set("format", "json")
	params.Set("idSite", c.config.SiteID)
	params.Set("token_auth", c.config.TokenAuth)
	params.Set("date", fmt.Sprintf("%s,%s", req.DateFrom.Format("2006-01-02"), req.DateTo.Format("2006-01-02")))
	params.Set("period", req.CohortPeriod)
	if req.Segment != "" {
		params.Set("segment", req.Segment)
	}

	var response CohortResponse
	if err := c.doJSONRequest(ctx, "/index.php", params, &response); err != nil {
		return nil, fmt.Errorf("failed to get cohorts: %w", err)
	}

	return &response, nil
}

// FunnelRequest represents a funnel analysis request
type FunnelRequest struct {
	FunnelID     string    `json:"funnel_id"`
	Segment      string    `json:"segment,omitempty"`
	DateFrom     time.Time `json:"date_from"`
	DateTo       time.Time `json:"date_to"`
}

// FunnelResponse represents the funnel analysis response
type FunnelResponse struct {
	FunnelID      string           `json:"funnel_id"`
	FunnelName    string           `json:"funnel_name"`
	Steps         []FunnelStep     `json:"steps"`
	TotalEntries  int              `json:"total_entries"`
	TotalExits    int              `json:"total_exits"`
	ConversionRate float64         `json:"conversion_rate"`
}

// FunnelStep represents a single step in the funnel
type FunnelStep struct {
	StepID        string    `json:"step_id"`
	StepName      string    `json:"step_name"`
	Visitors      int       `json:"visitors"`
	Dropoff       int       `json:"dropoff"`
	DropoffRate   float64   `json:"dropoff_rate"`
}

// GetFunnels retrieves funnel analysis data from Matomo
func (c *Client) GetFunnels(ctx context.Context, req FunnelRequest) (*FunnelResponse, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "Funnels.getFunnel")
	params.Set("format", "json")
	params.Set("idSite", c.config.SiteID)
	params.Set("token_auth", c.config.TokenAuth)
	params.Set("idFunnel", req.FunnelID)
	params.Set("date", fmt.Sprintf("%s,%s", req.DateFrom.Format("2006-01-02"), req.DateTo.Format("2006-01-02")))
	params.Set("period", "day")
	if req.Segment != "" {
		params.Set("segment", req.Segment)
	}

	var response FunnelResponse
	if err := c.doJSONRequest(ctx, "/index.php", params, &response); err != nil {
		return nil, fmt.Errorf("failed to get funnels: %w", err)
	}

	return &response, nil
}

// RealtimeVisitor represents a realtime visitor
type RealtimeVisitor struct {
	VisitorID       string            `json:"visitor_id"`
	UserID          string            `json:"user_id,omitempty"`
	LastActionTime  time.Time         `json:"last_action_time"`
	Pages           []string          `json:"pages"`
	CustomVariables map[string]string `json:"custom_variables,omitempty"`
}

// GetRealtimeVisitors retrieves current realtime visitors
func (c *Client) GetRealtimeVisitors(ctx context.Context, minutes int, limit int) ([]RealtimeVisitor, error) {
	params := url.Values{}
	params.Set("module", "API")
	params.Set("method", "Live.getLastVisitsDetails")
	params.Set("format", "json")
	params.Set("idSite", c.config.SiteID)
	params.Set("token_auth", c.config.TokenAuth)
	params.Set("period", "day")
	params.Set("date", "today")
	params.Set("filter_limit", fmt.Sprintf("%d", limit))
	if minutes > 0 {
		params.Set("lastMinutes", fmt.Sprintf("%d", minutes))
	}

	var response []RealtimeVisitor
	if err := c.doJSONRequest(ctx, "/index.php", params, &response); err != nil {
		return nil, fmt.Errorf("failed to get realtime visitors: %w", err)
	}

	return response, nil
}

// doRequest performs an HTTP POST request to Matomo with retries
func (c *Client) doRequest(ctx context.Context, path string, params url.Values) error {
	var lastErr error

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		// Build URL
		fullURL := c.config.BaseURL + path

		// Create request
		reqBody := params.Encode()
		httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBufferString(reqBody))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Execute request
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			c.logger.Warn("Matomo request failed, retrying",
				zap.Int("attempt", attempt+1),
				zap.Error(err),
			)
			time.Sleep(RetryDelay * time.Duration(attempt+1))
			continue
		}

		// Read and close body
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Check status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("Matomo returned status %d: %s", resp.StatusCode, string(body))
		c.logger.Warn("Matomo request returned error status, retrying",
			zap.Int("attempt", attempt+1),
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		time.Sleep(RetryDelay * time.Duration(attempt+1))
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doJSONRequest performs an HTTP GET request and parses JSON response
func (c *Client) doJSONRequest(ctx context.Context, path string, params url.Values, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		// Build URL
		fullURL := c.config.BaseURL + path + "?" + params.Encode()

		// Create request
		httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Execute request
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = err
			c.logger.Warn("Matomo JSON request failed, retrying",
				zap.Int("attempt", attempt+1),
				zap.Error(err),
			)
			time.Sleep(RetryDelay * time.Duration(attempt+1))
			continue
		}

		// Read and close body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			time.Sleep(RetryDelay * time.Duration(attempt+1))
			continue
		}

		// Check status code
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("Matomo returned status %d: %s", resp.StatusCode, string(body))
			c.logger.Warn("Matomo JSON request returned error status, retrying",
				zap.Int("attempt", attempt+1),
				zap.Int("status", resp.StatusCode),
				zap.String("response", string(body)),
			)
			time.Sleep(RetryDelay * time.Duration(attempt+1))
			continue
		}

		// Parse JSON response
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to parse JSON response: %w", err)
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// HealthCheck performs a health check on the Matomo server
func (c *Client) HealthCheck(ctx context.Context) error {
	params := url.Values{}
	params.Set("rec", "1")
	params.Set("idsite", c.config.SiteID)
	params.Set("token_auth", c.config.TokenAuth)
	params.Set("e_c", "health")
	params.Set("e_a", "check")
	params.Set("rand", fmt.Sprintf("%d", time.Now().UnixNano()))

	return c.doRequest(ctx, "/matomo.php", params)
}

// Close closes the HTTP client
func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}
