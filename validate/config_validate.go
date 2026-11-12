package validate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kong/go-kong/kong"
)

const (
	// DefaultConfigChunkSize is the default maximum size in bytes of one config validation request.
	DefaultConfigChunkSize = 5 << 20

	configValidatePath       = "/config/validate"
	konnectCoreEntitiesPath  = "/core-entities"
	pluginSchemasPageSize    = 1000
	maxRateLimitRetries      = 10
	minRateLimitRetryBackoff = time.Second
	maxRateLimitRetryBackoff = time.Minute
)

// sleep is a variable so that tests do not wait.
var sleep = func(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// configValidateURL returns an empty string if the client does not point to the core entities of a Konnect control
// plane.
func (v *Validator) configValidateURL() string {
	if !v.client.IsKonnectMode() {
		return ""
	}
	base := strings.TrimSuffix(v.client.BaseRootURL(), "/")
	if !strings.HasSuffix(base, konnectCoreEntitiesPath) {
		return ""
	}
	return strings.TrimSuffix(base, konnectCoreEntitiesPath) + configValidatePath
}

// send sends a request and waits for each 429 response.
func (v *Validator) send(method, rawURL string, body []byte) (int, []byte, error) {
	return v.sendWith(v.client, method, rawURL, body)
}

func (v *Validator) sendWith(client *kong.Client, method, rawURL string, body []byte) (int, []byte, error) {
	for attempt := 0; ; attempt++ {
		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(v.ctx, method, rawURL, reader)
		if err != nil {
			return 0, nil, err
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if client.UserAgent != "" {
			req.Header.Set("User-Agent", client.UserAgent)
		}
		resp, err := client.DoRAW(v.ctx, req)
		if err != nil {
			return 0, nil, err
		}
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return 0, nil, fmt.Errorf("reading response body: %w", err)
		}
		if resp.StatusCode != http.StatusTooManyRequests || attempt >= maxRateLimitRetries {
			return resp.StatusCode, respBody, nil
		}
		if err := sleep(v.ctx, retryAfter(resp, attempt)); err != nil {
			return 0, nil, err
		}
	}
}

func retryAfter(resp *http.Response, attempt int) time.Duration {
	if s := resp.Header.Get("Retry-After"); s != "" {
		if seconds, err := strconv.Atoi(s); err == nil && seconds >= 0 {
			return time.Duration(seconds) * time.Second
		}
		if t, err := http.ParseTime(s); err == nil {
			return max(time.Until(t), 0)
		}
	}
	backoff := minRateLimitRetryBackoff << attempt
	if backoff <= 0 || backoff > maxRateLimitRetryBackoff {
		return maxRateLimitRetryBackoff
	}
	return backoff
}

// customPluginNames returns the names of the custom plugins of the control plane and of the state.
// A cloned plugin of a custom plugin is also a custom plugin.
func (v *Validator) customPluginNames() (map[string]bool, error) {
	names := map[string]bool{}

	defs, err := v.state.CustomPluginDefinitions.GetAll()
	if err != nil {
		return nil, err
	}
	for _, d := range defs {
		if d.Name != nil {
			names[*d.Name] = true
		}
	}

	streamed, err := v.client.CustomPlugins.ListAll(v.ctx)
	if err != nil && !kong.IsNotFoundErr(err) {
		return nil, fmt.Errorf("listing custom plugins: %w", err)
	}
	for _, d := range streamed {
		if d.Name != nil {
			names[*d.Name] = true
		}
	}

	if err := v.addPluginSchemaNames(names); err != nil {
		return nil, err
	}

	cloned, err := v.state.ClonedPluginDefinitions.GetAll()
	if err != nil {
		return nil, err
	}
	for _, c := range cloned {
		if c.Name != nil && c.Ref != nil && names[*c.Ref] {
			names[*c.Name] = true
		}
	}
	return names, nil
}

// addPluginSchemaNames adds the names of the custom plugin schemas of the control plane.
func (v *Validator) addPluginSchemaNames(names map[string]bool) error {
	base := strings.TrimSuffix(v.client.BaseRootURL(), "/") + "/plugin-schemas"
	cursor := ""
	seen := map[string]bool{}
	for {
		q := url.Values{}
		q.Set("page[size]", strconv.Itoa(pluginSchemasPageSize))
		if cursor != "" {
			q.Set("page[after]", cursor)
		}
		status, body, err := v.send(http.MethodGet, base+"?"+q.Encode(), nil)
		if err != nil {
			return fmt.Errorf("listing custom plugin schemas: %w", err)
		}
		if status == http.StatusNotFound {
			return nil
		}
		if status != http.StatusOK {
			return fmt.Errorf("listing custom plugin schemas: HTTP status %d", status)
		}
		var page struct {
			Items []struct {
				Name string `json:"name"`
			} `json:"items"`
			Page struct {
				NextCursor string `json:"next_cursor"`
			} `json:"page"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return fmt.Errorf("decoding custom plugin schemas: %w", err)
		}
		for _, item := range page.Items {
			names[item.Name] = true
		}
		cursor = page.Page.NextCursor
		if len(page.Items) < pluginSchemasPageSize || cursor == "" || seen[cursor] {
			return nil
		}
		seen[cursor] = true
	}
}

type configIssue struct {
	Description string `json:"description"`
	Resource    struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"resource"`
}

type configErrorDetail struct {
	Type     string   `json:"type,omitempty"`
	Field    string   `json:"field,omitempty"`
	Messages []string `json:"messages,omitempty"`
	Entity   *struct {
		Type string `json:"type"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"entity,omitempty"`
}

type configValidationResponse struct {
	Items   []configIssue       `json:"items"`
	Message string              `json:"message"`
	Details []configErrorDetail `json:"details"`
}

// errConfigValidationUnavailable means that the control plane does not have the config validation endpoint.
var errConfigValidationUnavailable = errors.New("config validation endpoint is not available")

func (v *Validator) postConfig(rawURL string, body []byte) (*configValidationResponse, int, error) {
	status, respBody, err := v.sendWith(v.configClient, http.MethodPost, rawURL, body)
	if err != nil {
		return nil, 0, fmt.Errorf("validate configuration: %w", err)
	}
	if status == http.StatusNotFound {
		return nil, status, errConfigValidationUnavailable
	}
	res := &configValidationResponse{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, res); err != nil && status < http.StatusBadRequest {
			return nil, status, fmt.Errorf("validate configuration: decoding response: %w", err)
		}
	}
	if status >= http.StatusBadRequest && res.Message == "" && len(res.Items) == 0 {
		res.Message = strings.TrimSpace(string(respBody))
	}
	return res, status, nil
}

// validateWithConfigEndpoint validates the entities of fields with the config validation endpoint.
// It returns false if the control plane does not have the endpoint.
// The caller then uses the per-entity validation for all entities.
func (v *Validator) validateWithConfigEndpoint(fields map[string]string) ([]error, bool) {
	endpoint := v.configValidateURL()
	if endpoint == "" {
		return nil, false
	}

	kinds := v.selectedConfigKinds(fields)
	if len(kinds) == 0 {
		return nil, false
	}

	_, status, err := v.postConfig(endpoint, []byte("{}"))
	if errors.Is(err, errConfigValidationUnavailable) {
		return nil, false
	}
	if err != nil {
		return []error{err}, true
	}
	if status != http.StatusOK {
		return []error{fmt.Errorf("validate configuration: HTTP status %d", status)}, true
	}

	customPlugins, err := v.customPluginNames()
	if err != nil {
		return []error{err}, true
	}

	plan, err := buildConfigPlan(v.state, kinds, customPlugins)
	if err != nil {
		return []error{err}, true
	}

	errs := &errorSet{seen: map[string]bool{}}
	errs.add(conflictErrors(plan.conflictEntities(true))...)
	chunks := chunkConfig(plan.entities, v.configChunkSize)
	for i, chunk := range chunks {
		res, status, err := v.postConfig(endpoint, chunk.body())
		if err != nil {
			errs.add(fmt.Errorf("chunk %d of %d: %w", i+1, len(chunks), err))
			continue
		}
		switch {
		case status == http.StatusOK:
		case len(res.Items) > 0:
			errs.add(res.issueErrors()...)
		case status == http.StatusBadRequest && len(res.Details) > 0:
			errs.add(res.detailErrors()...)
		default:
			errs.add(fmt.Errorf("validate configuration: %s", res.describe(status)))
		}
	}

	errs.add(v.validateEntities("plugins", plan.perEntityPlugins)...)
	for field, entityType := range fields {
		if kind, ok := configKindOfField(field); ok && kinds[kind.name] {
			continue
		}
		errs.add(v.validateStateField(field, entityType)...)
	}
	return errs.errs, true
}

func (r *configValidationResponse) issueErrors() []error {
	errs := make([]error, 0, len(r.Items))
	for _, issue := range r.Items {
		errs = append(errs, fmt.Errorf("validate entity '%s (%s)': %s",
			issue.Resource.Type, issue.Resource.ID, issue.Description))
	}
	return errs
}

// errorCount matches the count of errors in the message of an error response.
var errorCount = regexp.MustCompile(`has (\d+) errors?`)

// detailErrors returns one error for each detail of an error response.
// The response has a limited number of details. A last error tells if there are more.
func (r *configValidationResponse) detailErrors() []error {
	errs := make([]error, 0, len(r.Details)+1)
	for _, d := range r.Details {
		text := strings.Join(d.Messages, ", ")
		if d.Field != "" {
			text = d.Field + ": " + text
		}
		if d.Entity == nil {
			errs = append(errs, fmt.Errorf("validate configuration: %s", text))
			continue
		}
		label := d.Entity.Name
		if label == "" {
			label = d.Entity.ID
		}
		errs = append(errs, fmt.Errorf("validate entity '%s (%s)': %s", d.Entity.Type, label, text))
	}
	if m := errorCount.FindStringSubmatch(r.Message); m != nil {
		if total, err := strconv.Atoi(m[1]); err == nil && total > len(r.Details) {
			errs = append(errs, fmt.Errorf("validate configuration: %d more errors are not shown. "+
				"Fix the errors above and validate again", total-len(r.Details)))
		}
	}
	return errs
}

func (r *configValidationResponse) describe(status int) string {
	if len(r.Details) == 0 {
		return fmt.Sprintf("HTTP status %d (message: %q)", status, r.Message)
	}
	details, err := json.Marshal(r.Details)
	if err != nil {
		return fmt.Sprintf("HTTP status %d (message: %q)", status, r.Message)
	}
	return fmt.Sprintf("HTTP status %d (message: %q, details: %s)", status, r.Message, details)
}

// selectedConfigKinds returns the names of the entity lists that fields selects.
// Without an entity filter, it selects all entity lists.
func (v *Validator) selectedConfigKinds(fields map[string]string) map[string]bool {
	kinds := map[string]bool{}
	for _, kind := range configKinds {
		if _, ok := fields[kind.stateField]; ok || len(v.onlineEntitiesFilter) == 0 {
			kinds[kind.name] = true
		}
	}
	return kinds
}

// localConflictErrors finds the conflicts that the per-entity validation cannot find.
func (v *Validator) localConflictErrors(fields map[string]string) []error {
	plan, err := buildConfigPlan(v.state, v.selectedConfigKinds(fields), nil)
	if err != nil {
		return []error{err}
	}
	return conflictErrors(plan.conflictEntities(false))
}

// errorSet keeps the first error for each message.
// An entity can be in more than one chunk. Its errors then come back more than once.
type errorSet struct {
	errs []error
	seen map[string]bool
}

func (s *errorSet) add(errs ...error) {
	for _, err := range errs {
		if s.seen[err.Error()] {
			continue
		}
		s.seen[err.Error()] = true
		s.errs = append(s.errs, err)
	}
}
