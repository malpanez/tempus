package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	scopeCalendarEvents      = "https://www.googleapis.com/auth/calendar.events"
	grantTypeDeviceCode      = "urn:ietf:params:oauth:grant-type:device_code"
	contentTypeFormEncoded   = "application/x-www-form-urlencoded"
	contentTypeJSON          = "application/json"
	defaultDevicePollSeconds = 5
	tokenExpiryLeeway        = 30 * time.Second
)

var (
	defaultDeviceEndpoint  = "https://oauth2.googleapis.com/device/code"
	defaultTokenEndpoint   = "https://oauth2.googleapis.com/token"
	defaultCalendarBaseURL = "https://www.googleapis.com/calendar/v3"
)

// Options configure the Google Calendar client.
type Options struct {
	ClientID        string
	ClientSecret    string
	TokenFile       string
	DeviceEndpoint  string
	TokenEndpoint   string
	CalendarBaseURL string
	HTTPClient      *http.Client
}

// Client synchronises ICS events with Google Calendar.
type Client struct {
	opts       Options
	httpClient *http.Client
	mu         sync.Mutex
	token      *token
}

type token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
}

func (t *token) valid(now time.Time) bool {
	if t == nil {
		return false
	}
	if strings.TrimSpace(t.AccessToken) == "" {
		return false
	}
	return t.Expiry.After(now.Add(tokenExpiryLeeway))
}

func (t *token) tokenType() string {
	if strings.TrimSpace(t.TokenType) == "" {
		return "Bearer"
	}
	return t.TokenType
}

type storedToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Expiry       string `json:"expiry"`
}

// NewClient creates a Google Calendar client using the supplied options.
func NewClient(ctx context.Context, opts Options) (*Client, error) {
	opts = normalizeOptions(opts)
	if strings.TrimSpace(opts.ClientID) == "" {
		return nil, fmt.Errorf("ClientID is required")
	}
	if strings.TrimSpace(opts.ClientSecret) == "" {
		return nil, fmt.Errorf("ClientSecret is required")
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	c := &Client{
		opts:       opts,
		httpClient: client,
	}

	if opts.TokenFile != "" {
		if tok, err := readTokenFromFile(opts.TokenFile); err == nil {
			c.token = tok
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	return c, nil
}

// ImportICS uploads all events from the ICS payload into the target calendar.
func (c *Client) ImportICS(ctx context.Context, calendarID string, ics string) error {
	if strings.TrimSpace(calendarID) == "" {
		return fmt.Errorf("calendarID cannot be empty")
	}
	if strings.TrimSpace(ics) == "" {
		return fmt.Errorf("ICS payload cannot be empty")
	}

	if err := c.ensureToken(ctx); err != nil {
		return err
	}

	events, err := parseICS(ics)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return fmt.Errorf("no VEVENT blocks found in ICS payload")
	}

	for _, ev := range events {
		if err := c.insertEvent(ctx, calendarID, ev); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if c.token != nil && c.token.valid(now) {
		return nil
	}

	// Attempt refresh using in-memory token.
	if c.token != nil && c.token.RefreshToken != "" {
		if err := c.refreshToken(ctx, c.token.RefreshToken); err == nil {
			return nil
		}
	}

	// Reload from disk (might have been updated).
	if c.opts.TokenFile != "" {
		if tok, err := readTokenFromFile(c.opts.TokenFile); err == nil {
			c.token = tok
			if c.token.valid(now) {
				return nil
			}
			if c.token.RefreshToken != "" {
				if err := c.refreshToken(ctx, c.token.RefreshToken); err == nil {
					return nil
				}
			}
		}
	}

	// Fall back to device flow.
	return c.deviceFlow(ctx)
}

func (c *Client) refreshToken(ctx context.Context, refreshToken string) error {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.opts.ClientID)
	form.Set("client_secret", c.opts.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.opts.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentTypeFormEncoded)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed: %s", strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode refresh token: %w", err)
	}
	if payload.AccessToken == "" {
		return fmt.Errorf("token refresh returned empty access_token")
	}
	if payload.RefreshToken == "" {
		payload.RefreshToken = refreshToken
	}
	if payload.TokenType == "" {
		payload.TokenType = "Bearer"
	}
	expiry := time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	if payload.ExpiresIn == 0 && c.token != nil && !c.token.Expiry.IsZero() {
		expiry = c.token.Expiry
	}

	c.token = &token{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		TokenType:    payload.TokenType,
		Expiry:       expiry,
	}

	return c.saveToken()
}

func (c *Client) deviceFlow(ctx context.Context) error {
	form := url.Values{}
	form.Set("client_id", c.opts.ClientID)
	form.Set("scope", scopeCalendarEvents)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.opts.DeviceEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentTypeFormEncoded)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("device authorization failed: %s", strings.TrimSpace(string(body)))
	}

	var device struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURL string `json:"verification_url"`
		ExpiresIn       int64  `json:"expires_in"`
		Interval        int64  `json:"interval"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return fmt.Errorf("decode device response: %w", err)
	}
	if device.DeviceCode == "" {
		return fmt.Errorf("device response missing device_code")
	}

	// Inform the user (best-effort).
	fmt.Printf("Authorize Tempus at %s with code %s\n", device.VerificationURL, device.UserCode)

	interval := time.Duration(device.Interval) * time.Second
	if interval <= 0 {
		interval = defaultDevicePollSeconds * time.Second
	}
	expiry := time.Now().Add(time.Duration(device.ExpiresIn) * time.Second)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(expiry) {
			return fmt.Errorf("device code expired before authorization completed")
		}

		token, err := c.pollDeviceToken(ctx, device.DeviceCode)
		if err == nil {
			c.token = token
			return c.saveToken()
		}

		var authPending *authorizationPendingError
		if errors.As(err, &authPending) {
			time.Sleep(interval)
			continue
		}
		return err
	}
}

type authorizationPendingError struct {
	message string
}

func (e *authorizationPendingError) Error() string {
	return e.message
}

func (c *Client) pollDeviceToken(ctx context.Context, deviceCode string) (*token, error) {
	form := url.Values{}
	form.Set("grant_type", grantTypeDeviceCode)
	form.Set("device_code", deviceCode)
	form.Set("client_id", c.opts.ClientID)
	form.Set("client_secret", c.opts.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.opts.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentTypeFormEncoded)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusBadRequest {
		var payload struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &payload); err == nil {
			if payload.Error == "authorization_pending" {
				return nil, &authorizationPendingError{message: strings.TrimSpace(payload.Error)}
			}
			return nil, fmt.Errorf("device token error: %s", strings.TrimSpace(payload.Error))
		}
		return nil, fmt.Errorf("device token error: %s", strings.TrimSpace(string(body)))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("device token request failed: %s", strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode device token: %w", err)
	}
	if payload.AccessToken == "" {
		return nil, fmt.Errorf("device token response missing access_token")
	}
	if payload.TokenType == "" {
		payload.TokenType = "Bearer"
	}

	expiry := time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	return &token{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		TokenType:    payload.TokenType,
		Expiry:       expiry,
	}, nil
}

func (c *Client) insertEvent(ctx context.Context, calendarID string, ev icsEvent) error {
	payload := map[string]interface{}{
		"summary": ev.Summary,
	}
	if strings.TrimSpace(ev.Description) != "" {
		payload["description"] = ev.Description
	}
	if strings.TrimSpace(ev.Location) != "" {
		payload["location"] = ev.Location
	}

	if ev.AllDay {
		start := map[string]interface{}{"date": ev.Start.Format("2006-01-02")}
		end := map[string]interface{}{"date": ev.End.Format("2006-01-02")}
		if ev.StartTZ != "" {
			start["timeZone"] = ev.StartTZ
		}
		if ev.EndTZ != "" {
			end["timeZone"] = ev.EndTZ
		}
		payload["start"] = start
		payload["end"] = end
	} else {
		start := map[string]interface{}{"dateTime": ev.Start.Format(time.RFC3339)}
		end := map[string]interface{}{"dateTime": ev.End.Format(time.RFC3339)}
		if ev.StartTZ != "" {
			start["timeZone"] = ev.StartTZ
		}
		if ev.EndTZ != "" {
			end["timeZone"] = ev.EndTZ
		}
		payload["start"] = start
		payload["end"] = end
	}

	if strings.TrimSpace(ev.RRule) != "" {
		payload["recurrence"] = []string{"RRULE:" + ev.RRule}
	}

	if len(ev.Reminders) > 0 {
		overrides := make([]map[string]interface{}, 0, len(ev.Reminders))
		for _, minutes := range ev.Reminders {
			if minutes <= 0 {
				continue
			}
			overrides = append(overrides, map[string]interface{}{
				"method":  "popup",
				"minutes": minutes,
			})
		}
		if len(overrides) > 0 {
			payload["reminders"] = map[string]interface{}{
				"useDefault": false,
				"overrides":  overrides,
			}
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	u := fmt.Sprintf("%s/calendars/%s/events", strings.TrimRight(c.opts.CalendarBaseURL, "/"), url.PathEscape(calendarID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.token.tokenType(), c.token.AccessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("event insert failed: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) saveToken() error {
	if strings.TrimSpace(c.opts.TokenFile) == "" || c.token == nil {
		return nil
	}

	dir := filepath.Dir(c.opts.TokenFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	st := storedToken{
		AccessToken:  c.token.AccessToken,
		RefreshToken: c.token.RefreshToken,
		TokenType:    c.token.TokenType,
		Expiry:       c.token.Expiry.UTC().Format(time.RFC3339Nano),
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.opts.TokenFile, data, 0o600)
}

func readTokenFromFile(path string) (*token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var st storedToken
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	expiry, err := time.Parse(time.RFC3339Nano, st.Expiry)
	if err != nil {
		return nil, err
	}
	return &token{
		AccessToken:  st.AccessToken,
		RefreshToken: st.RefreshToken,
		TokenType:    st.TokenType,
		Expiry:       expiry,
	}, nil
}

func normalizeOptions(opts Options) Options {
	if opts.DeviceEndpoint == "" {
		opts.DeviceEndpoint = defaultDeviceEndpoint
	}
	if opts.TokenEndpoint == "" {
		opts.TokenEndpoint = defaultTokenEndpoint
	}
	if opts.CalendarBaseURL == "" {
		opts.CalendarBaseURL = defaultCalendarBaseURL
	}
	return opts
}

// ------------- ICS parsing helpers -------------

type icsEvent struct {
	Summary           string
	Description       string
	Location          string
	Start             time.Time
	End               time.Time
	StartTZ           string
	EndTZ             string
	AllDay            bool
	RRule             string
	Reminders         []int
	AbsoluteReminders []time.Time
}

func parseICS(ics string) ([]icsEvent, error) {
	lines := unfoldICS(ics)
	var events []icsEvent

	var current *icsEvent
	var inEvent bool
	var inAlarm bool
	var defaultTZ string
	var alarmMinutes int
	var alarmAbsolute *time.Time

	for _, raw := range lines {
		prop, err := parseICSLine(raw)
		if err != nil {
			continue
		}

		switch prop.Name {
		case "X-WR-TIMEZONE":
			if defaultTZ == "" {
				defaultTZ = prop.Value
			}
		case "BEGIN":
			switch strings.ToUpper(strings.TrimSpace(prop.Value)) {
			case "VEVENT":
				inEvent = true
				current = &icsEvent{}
			case "VALARM":
				if inEvent {
					inAlarm = true
					alarmMinutes = 0
					alarmAbsolute = nil
				}
			}
		case "END":
			switch strings.ToUpper(strings.TrimSpace(prop.Value)) {
			case "VALARM":
				if inEvent && inAlarm {
					switch {
					case alarmMinutes > 0:
						current.Reminders = append(current.Reminders, alarmMinutes)
					case alarmAbsolute != nil:
						current.AbsoluteReminders = append(current.AbsoluteReminders, *alarmAbsolute)
					}
				}
				inAlarm = false
				alarmMinutes = 0
				alarmAbsolute = nil
			case "VEVENT":
				if inEvent && current != nil {
					if current.End.IsZero() {
						if current.AllDay {
							current.End = current.Start.AddDate(0, 0, 1)
						} else {
							current.End = current.Start.Add(time.Hour)
						}
					}
					if current.EndTZ == "" {
						current.EndTZ = current.StartTZ
					}
					if len(current.AbsoluteReminders) > 0 && !current.Start.IsZero() {
						for _, absTime := range current.AbsoluteReminders {
							minutes := minutesBetween(current.Start, absTime)
							if minutes > 0 {
								current.Reminders = append(current.Reminders, minutes)
							}
						}
					}
					events = append(events, *current)
				}
				inEvent = false
				current = nil
			}
		default:
			if !inEvent || current == nil {
				continue
			}

			if inAlarm {
				switch prop.Name {
				case "TRIGGER":
					if absTime, ok := parseICSTriggerAbsolute(prop, defaultTZ, current.StartTZ); ok {
						alarmAbsolute = &absTime
						alarmMinutes = 0
					} else if mins, err := parseICSDurationMinutes(prop.Value); err == nil {
						if mins < 0 {
							mins = -mins
						}
						alarmMinutes = mins
					}
				}
				continue
			}

			switch prop.Name {
			case "SUMMARY":
				current.Summary = unescapeICSValue(prop.Value)
			case "DESCRIPTION":
				current.Description = unescapeICSValue(prop.Value)
			case "LOCATION":
				current.Location = unescapeICSValue(prop.Value)
			case "RRULE":
				current.RRule = strings.TrimSpace(prop.Value)
			case "DTSTART":
				t, tz, allDay, err := parseICSTime(prop, defaultTZ)
				if err == nil {
					current.Start = t
					current.AllDay = current.AllDay || allDay
					if tz != "" {
						current.StartTZ = tz
					}
				}
			case "DTEND":
				t, tz, allDay, err := parseICSTime(prop, defaultTZ)
				if err == nil {
					current.End = t
					current.AllDay = current.AllDay || allDay
					if tz != "" {
						current.EndTZ = tz
					}
				}
			}
		}
	}

	return events, nil
}

type icsProperty struct {
	Name   string
	Params map[string]string
	Value  string
}

func parseICSLine(line string) (icsProperty, error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return icsProperty{}, fmt.Errorf("invalid ICS line")
	}
	head := parts[0]
	value := parts[1]

	segments := strings.Split(head, ";")
	name := strings.ToUpper(strings.TrimSpace(segments[0]))
	params := make(map[string]string, len(segments)-1)
	for _, seg := range segments[1:] {
		if seg == "" {
			continue
		}
		kv := strings.SplitN(seg, "=", 2)
		if len(kv) != 2 {
			continue
		}
		params[strings.ToUpper(strings.TrimSpace(kv[0]))] = strings.TrimSpace(kv[1])
	}
	return icsProperty{
		Name:   name,
		Params: params,
		Value:  strings.TrimSpace(value),
	}, nil
}

func parseICSTime(prop icsProperty, defaultTZ string) (time.Time, string, bool, error) {
	value := strings.TrimSpace(prop.Value)
	value = strings.ReplaceAll(value, "T", "T")
	tz := strings.TrimSpace(prop.Params["TZID"])
	if tz == "" {
		tz = strings.TrimSpace(defaultTZ)
	}

	if strings.EqualFold(prop.Params["VALUE"], "DATE") {
		loc := time.UTC
		if tz != "" {
			if l, err := time.LoadLocation(tz); err == nil {
				loc = l
			}
		}
		t, err := time.ParseInLocation("20060102", value, loc)
		return t, tz, true, err
	}

	if strings.HasSuffix(value, "Z") {
		t, err := time.Parse("20060102T150405Z", value)
		return t, "UTC", false, err
	}

	loc := time.Local
	if tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			loc = l
		}
	}
	t, err := time.ParseInLocation("20060102T150405", value, loc)
	return t, tz, false, err
}

func unfoldICS(ics string) []string {
	normalized := strings.ReplaceAll(ics, "\r\n", "\n")
	rawLines := strings.Split(normalized, "\n")
	lines := make([]string, 0, len(rawLines))

	var current strings.Builder
	for _, raw := range rawLines {
		if raw == "" && current.Len() == 0 {
			continue
		}
		if strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t") {
			current.WriteString(strings.TrimLeft(raw, " \t"))
			continue
		}
		if current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
		}
		current.WriteString(strings.TrimRight(raw, "\r"))
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

func unescapeICSValue(v string) string {
	v = strings.ReplaceAll(v, `\n`, "\n")
	v = strings.ReplaceAll(v, `\;`, ";")
	v = strings.ReplaceAll(v, `\,`, ",")
	v = strings.ReplaceAll(v, `\\`, `\`)
	return v
}

var icsDurationPattern = regexp.MustCompile(`(?i)^([+-])?P(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

func parseICSDurationMinutes(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	m := icsDurationPattern.FindStringSubmatch(raw)
	if m == nil {
		return 0, fmt.Errorf("invalid duration %q", raw)
	}
	sign := 1
	if m[1] == "-" {
		sign = -1
	}
	weeks := atoiSafe(m[2])
	days := atoiSafe(m[3])
	hours := atoiSafe(m[4])
	minutes := atoiSafe(m[5])
	seconds := atoiSafe(m[6])

	totalMinutes := weeks*7*24*60 + days*24*60 + hours*60 + minutes
	if seconds >= 30 {
		totalMinutes++
	}
	return sign * totalMinutes, nil
}

func atoiSafe(s string) int {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

func parseICSTriggerAbsolute(prop icsProperty, defaultTZ, eventTZ string) (time.Time, bool) {
	value := strings.TrimSpace(prop.Value)
	if value == "" {
		return time.Time{}, false
	}

	tz := strings.TrimSpace(prop.Params["TZID"])
	if tz == "" {
		tz = strings.TrimSpace(eventTZ)
	}
	if tz == "" {
		tz = strings.TrimSpace(defaultTZ)
	}

	if strings.EqualFold(prop.Params["VALUE"], "DATE") && len(value) == len("20060102") {
		loc := time.Local
		if tz != "" {
			if l, err := time.LoadLocation(tz); err == nil {
				loc = l
			}
		}
		if t, err := time.ParseInLocation("20060102", value, loc); err == nil {
			return t, true
		}
	}

	upper := strings.ToUpper(value)
	if strings.HasSuffix(upper, "Z") {
		for _, layout := range []string{"20060102T150405Z", "20060102T1504Z"} {
			if t, err := time.Parse(layout, upper); err == nil {
				return t, true
			}
		}
	}

	loc := time.Local
	if tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			loc = l
		}
	}
	for _, layout := range []string{"20060102T150405", "20060102T1504"} {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func minutesBetween(start, trigger time.Time) int {
	if start.IsZero() || trigger.IsZero() {
		return 0
	}
	diff := start.Sub(trigger)
	if diff <= 0 {
		return 0
	}
	minutes := int(diff / time.Minute)
	if minutes == 0 {
		return 0
	}
	return minutes
}
