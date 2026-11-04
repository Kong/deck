package validate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/blang/semver/v4"
	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCPPath           = "/v2/control-planes/cp-1"
	testConfigPath       = testCPPath + "/config/validate"
	testCoreEntitiesPath = testCPPath + "/core-entities"
)

// fakeKonnect is a test server for the control plane endpoints that the validator uses.
type fakeKonnect struct {
	t *testing.T

	mu sync.Mutex
	// configEnabled false makes the config validation endpoint return 404.
	configEnabled bool
	configBodies  []map[string][]map[string]any
	// configResponses holds the responses to send in order, after the probe. A nil entry sends 200.
	configResponses []*fakeResponse
	entityRequests  map[string][]map[string]any
	// invalidEntities holds the names of the entities that the per-entity endpoint rejects.
	invalidEntities map[string]bool
	// rejectCredentials makes the config validation endpoint act as a version without credential support.
	// That version ignores the credential lists of the probe.
	rejectCredentials bool
	inFlight          atomic.Int32
	maxInFlight       atomic.Int32

	pluginSchemas []string
	customPlugins []string
}

type fakeResponse struct {
	status  int
	headers map[string]string
	body    string
}

func newFakeKonnect(t *testing.T) *fakeKonnect {
	return &fakeKonnect{
		t:               t,
		configEnabled:   true,
		entityRequests:  map[string][]map[string]any{},
		invalidEntities: map[string]bool{},
	}
}

func (f *fakeKonnect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if !assert.NoError(f.t, err) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch {
	case r.Method == http.MethodPost && r.URL.Path == testConfigPath:
		f.serveConfig(w, body)
	case r.Method == http.MethodGet && r.URL.Path == testCoreEntitiesPath+"/plugin-schemas":
		items := make([]map[string]string, 0, len(f.pluginSchemas))
		for _, name := range f.pluginSchemas {
			items = append(items, map[string]string{"name": name})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "page": map[string]any{"total_count": len(items)}})
	case r.Method == http.MethodGet && r.URL.Path == testCoreEntitiesPath+"/custom-plugins":
		data := make([]map[string]string, 0, len(f.customPlugins))
		for _, name := range f.customPlugins {
			data = append(data, map[string]string{"name": name})
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": data})
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, testCoreEntitiesPath+"/schemas/"):
		entityType := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, testCoreEntitiesPath+"/schemas/"), "/validate")
		var entity map[string]any
		if !assert.NoError(f.t, json.Unmarshal(body, &entity)) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		f.mu.Lock()
		f.entityRequests[entityType] = append(f.entityRequests[entityType], entity)
		f.mu.Unlock()
		if name, _ := entity["name"].(string); f.invalidEntities[name] {
			writeJSON(w, http.StatusBadRequest, map[string]any{"message": "schema violation (" + name + ")"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{})
	default:
		f.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (f *fakeKonnect) serveConfig(w http.ResponseWriter, body []byte) {
	n := f.inFlight.Add(1)
	defer f.inFlight.Add(-1)
	for {
		m := f.maxInFlight.Load()
		if n <= m || f.maxInFlight.CompareAndSwap(m, n) {
			break
		}
	}
	// A short wait makes parallel requests overlap.
	time.Sleep(time.Millisecond)

	if !f.configEnabled {
		writeJSON(w, http.StatusNotFound, map[string]any{"message": "Not Found"})
		return
	}
	var config map[string][]map[string]any
	if err := json.Unmarshal(body, &config); err != nil {
		f.t.Errorf("decoding config: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if string(body) == credentialProbeBody {
		if f.rejectCredentials {
			writeJSON(w, http.StatusOK, map[string]any{})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"items": []map[string]any{{
			"description": "could not find consumer",
			"resource":    map[string]any{"id": "00000000-0000-4000-8000-000000000002", "type": "consumer"},
		}}})
		return
	}
	for list := range config {
		if f.rejectCredentials && credentialKinds[list] {
			writeJSON(w, http.StatusBadRequest, map[string]any{"message": "unknown field " + list})
			return
		}
	}

	f.mu.Lock()
	var res *fakeResponse
	if len(config) > 0 {
		f.configBodies = append(f.configBodies, config)
		if len(f.configResponses) > 0 {
			res = f.configResponses[0]
			f.configResponses = f.configResponses[1:]
		}
	}
	f.mu.Unlock()

	if res == nil {
		writeJSON(w, http.StatusOK, map[string]any{})
		return
	}
	for k, v := range res.headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(res.status)
	_, _ = w.Write([]byte(res.body))
}

var credentialKinds = map[string]bool{
	"keyauth_credentials": true, "basicauth_credentials": true, "hmacauth_credentials": true,
	"jwt_secrets": true, "acls": true, "mtls_auth_credentials": true,
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (f *fakeKonnect) validator(t *testing.T, ks *state.KongState, filter ...string) *Validator {
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	client, err := kong.NewClient(kong.String(srv.URL+testCoreEntitiesPath), srv.Client())
	require.NoError(t, err)
	client.SetKonnectFlag(true)
	return NewValidator(ValidatorOpts{
		Ctx:                  context.Background(),
		State:                ks,
		Client:               client,
		Parallelism:          4,
		OnlineEntitiesFilter: filter,
	})
}

func noSleep(t *testing.T) *[]time.Duration {
	var waits []time.Duration
	orig := sleep
	sleep = func(_ context.Context, d time.Duration) error {
		waits = append(waits, d)
		return nil
	}
	t.Cleanup(func() { sleep = orig })
	return &waits
}

var format30 = semver.MustParse("3.0.0")

func refID(entity map[string]any, field string) string {
	ref, ok := entity[field].(map[string]any)
	if !ok {
		return ""
	}
	id, _ := ref["id"].(string)
	return id
}

// requireReferencesInChunk fails if an entity of config refers to an entity that is not in config.
func requireReferencesInChunk(t *testing.T, config map[string][]map[string]any) {
	t.Helper()
	ids := map[string]map[string]bool{}
	clonedNames := map[string]bool{}
	for kind, entities := range config {
		ids[kind] = map[string]bool{}
		for _, e := range entities {
			ids[kind][e["id"].(string)] = true
			if kind == "cloned_plugins" {
				clonedNames[e["name"].(string)] = true
			}
		}
	}
	refs := map[string]map[string]string{
		"services":              {"client_certificate": "certificates"},
		"routes":                {"service": "services"},
		"upstreams":             {"client_certificate": "certificates"},
		"targets":               {"upstream": "upstreams"},
		"snis":                  {"certificate": "certificates"},
		"keys":                  {"set": "key_sets"},
		"keyauth_credentials":   {"consumer": "consumers"},
		"basicauth_credentials": {"consumer": "consumers"},
		"hmacauth_credentials":  {"consumer": "consumers"},
		"jwt_secrets":           {"consumer": "consumers"},
		"acls":                  {"consumer": "consumers"},
		"mtls_auth_credentials": {"consumer": "consumers", "ca_certificate": "ca_certificates"},
		"plugins": {
			"service": "services", "route": "routes",
			"consumer": "consumers", "consumer_group": "consumer_groups",
		},
	}
	for kind, fields := range refs {
		for _, e := range config[kind] {
			for field, target := range fields {
				if id := refID(e, field); id != "" {
					require.Truef(t, ids[target][id], "%s %s refers to %s %s that is not in the chunk",
						kind, e["id"], target, id)
				}
			}
			if kind == "plugins" && strings.HasPrefix(e["name"].(string), "clone-") {
				require.Truef(t, clonedNames[e["name"].(string)], "cloned plugin %s is not in the chunk", e["name"])
			}
		}
	}
}

func newTestState(t *testing.T) *state.KongState {
	t.Helper()
	ks, err := state.NewKongState()
	require.NoError(t, err)
	return ks
}

// addServiceTree adds a service with one route and one plugin on each of them.
func addServiceTree(t *testing.T, ks *state.KongState, i int, pluginConfig kong.Configuration) {
	t.Helper()
	svcID := fmt.Sprintf("00000000-0000-0000-0000-%012d", i)
	routeID := fmt.Sprintf("00000000-0000-0000-0001-%012d", i)
	require.NoError(t, ks.Services.Add(state.Service{Service: kong.Service{
		ID: kong.String(svcID), Name: kong.String(fmt.Sprintf("svc-%d", i)), Host: kong.String("example.com"),
	}}))
	require.NoError(t, ks.Routes.Add(state.Route{Route: kong.Route{
		ID: kong.String(routeID), Name: kong.String(fmt.Sprintf("route-%d", i)),
		Paths: kong.StringSlice(fmt.Sprintf("/r%d", i)), Service: &kong.Service{ID: kong.String(svcID)},
	}}))
	require.NoError(t, ks.Plugins.Add(state.Plugin{Plugin: kong.Plugin{
		ID: kong.String(fmt.Sprintf("00000000-0000-0000-0002-%012d", i)), Name: kong.String("key-auth"),
		Route: &kong.Route{ID: kong.String(routeID)}, Config: pluginConfig,
	}}))
	require.NoError(t, ks.Plugins.Add(state.Plugin{Plugin: kong.Plugin{
		ID: kong.String(fmt.Sprintf("00000000-0000-0000-0003-%012d", i)), Name: kong.String("rate-limiting"),
		Service: &kong.Service{ID: kong.String(svcID)}, Config: pluginConfig,
	}}))
}

func TestValidate_ConfigEndpointNotFoundUsesPerEntityValidation(t *testing.T) {
	f := newFakeKonnect(t)
	f.configEnabled = false
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)

	errs := f.validator(t, ks).Validate(format30)

	require.Empty(t, errs)
	assert.Empty(t, f.configBodies)
	assert.Len(t, f.entityRequests["services"], 1)
	assert.Len(t, f.entityRequests["routes"], 1)
	assert.Len(t, f.entityRequests["plugins"], 2)
}

func TestValidate_SmallConfigUsesOneRequest(t *testing.T) {
	f := newFakeKonnect(t)
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)
	require.NoError(t, ks.Upstreams.Add(state.Upstream{Upstream: kong.Upstream{
		ID: kong.String("10000000-0000-0000-0000-000000000001"), Name: kong.String("upstream-1"),
	}}))
	require.NoError(t, ks.Targets.Add(state.Target{Target: kong.Target{
		ID: kong.String("10000000-0000-0000-0000-000000000002"), Target: kong.String("example.com:80"),
		Upstream: &kong.Upstream{ID: kong.String("10000000-0000-0000-0000-000000000001")},
	}}))
	require.NoError(t, ks.Consumers.Add(state.Consumer{Consumer: kong.Consumer{
		ID: kong.String("10000000-0000-0000-0000-000000000003"), Username: kong.String("consumer-1"),
	}}))
	require.NoError(t, ks.KeyAuths.Add(state.KeyAuth{KeyAuth: kong.KeyAuth{
		ID: kong.String("10000000-0000-0000-0000-000000000004"), Key: kong.String("example-key"),
		Consumer: &kong.Consumer{ID: kong.String("10000000-0000-0000-0000-000000000003")},
	}}))

	errs := f.validator(t, ks).Validate(format30)

	require.Empty(t, errs)
	require.Len(t, f.configBodies, 1)
	config := f.configBodies[0]
	requireReferencesInChunk(t, config)
	assert.Len(t, config["services"], 1)
	assert.Len(t, config["routes"], 1)
	assert.Len(t, config["plugins"], 2)
	assert.Len(t, config["upstreams"], 1)
	assert.Len(t, config["targets"], 1)
	assert.Len(t, config["consumers"], 1)
	assert.Len(t, config["keyauth_credentials"], 1)
	for _, entityType := range []string{
		"services", "routes", "plugins", "upstreams", "targets", "consumers", "keyauth_credentials",
	} {
		assert.Emptyf(t, f.entityRequests[entityType], "unexpected per-entity request for %s", entityType)
	}
}

func TestValidate_LargeConfigIsSplitIntoChunks(t *testing.T) {
	f := newFakeKonnect(t)
	ks := newTestState(t)
	// Each plugin has a config of about 2 KB. 1000 services then give about 4 MB.
	config := kong.Configuration{"padding": strings.Repeat("x", 2048)}
	const services = 1000
	for i := 0; i < services; i++ {
		addServiceTree(t, ks, i, config)
	}
	require.NoError(t, ks.ConsumerGroups.Add(state.ConsumerGroup{ConsumerGroup: kong.ConsumerGroup{
		ID: kong.String("20000000-0000-0000-0000-000000000001"), Name: kong.String("group-1"),
	}}))
	// A plugin on each route and on the consumer group refers to entities in many chunks.
	for i := 0; i < services; i += 50 {
		require.NoError(t, ks.Plugins.Add(state.Plugin{Plugin: kong.Plugin{
			ID: kong.String(fmt.Sprintf("00000000-0000-0000-0004-%012d", i)), Name: kong.String("rate-limiting"),
			Route:         &kong.Route{ID: kong.String(fmt.Sprintf("00000000-0000-0000-0001-%012d", i))},
			ConsumerGroup: &kong.ConsumerGroup{ID: kong.String("20000000-0000-0000-0000-000000000001")},
		}}))
	}

	errs := f.validator(t, ks).Validate(format30)

	require.Empty(t, errs)
	require.Greater(t, len(f.configBodies), 1)
	assert.Equal(t, int32(1), f.maxInFlight.Load(), "chunks must be sent one at a time")

	validated := map[string]bool{}
	for _, config := range f.configBodies {
		requireReferencesInChunk(t, config)
		body, err := json.Marshal(config)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(body), maxConfigChunkBytes)
		for _, entities := range config {
			for _, e := range entities {
				validated[e["id"].(string)] = true
			}
		}
	}
	assert.Len(t, validated, services*4+services/50+1)
	assert.Empty(t, f.entityRequests)
}

func TestValidate_CustomPluginsAndPartialsUsePerEntityValidation(t *testing.T) {
	f := newFakeKonnect(t)
	f.pluginSchemas = []string{"example-uploaded-plugin"}
	f.customPlugins = []string{"example-streamed-plugin"}
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)
	svc := &kong.Service{ID: kong.String("00000000-0000-0000-0000-000000000001")}
	customNames := []string{
		"example-uploaded-plugin", "example-streamed-plugin", "example-file-plugin", "clone-of-custom",
	}
	for i, name := range customNames {
		require.NoError(t, ks.Plugins.Add(state.Plugin{Plugin: kong.Plugin{
			ID: kong.String(fmt.Sprintf("30000000-0000-0000-0000-%012d", i)), Name: kong.String(name), Service: svc,
		}}))
	}
	require.NoError(t, ks.CustomPluginDefinitions.Add(state.CustomPluginDefinition{
		CustomPluginDefinition: kong.CustomPluginDefinition{
			ID: kong.String("30000000-0000-0000-0001-000000000001"), Name: kong.String("example-file-plugin"),
		},
	}))
	require.NoError(t, ks.ClonedPluginDefinitions.Add(state.ClonedPluginDefinition{
		ClonedPluginDefinition: kong.ClonedPluginDefinition{
			ID: kong.String("30000000-0000-0000-0002-000000000001"), Name: kong.String("clone-of-custom"),
			Ref: kong.String("example-uploaded-plugin"),
		},
	}))
	require.NoError(t, ks.Plugins.Add(state.Plugin{Plugin: kong.Plugin{
		ID: kong.String("30000000-0000-0000-0000-000000000010"), Name: kong.String("rate-limiting-advanced"),
		Service:  svc,
		Partials: []*kong.PartialLink{{Partial: &kong.Partial{ID: kong.String("30000000-0000-0000-0003-000000000001")}}},
	}}))

	errs := f.validator(t, ks).Validate(format30)

	require.Empty(t, errs)
	require.Len(t, f.configBodies, 1)
	requireReferencesInChunk(t, f.configBodies[0])
	bulkPlugins := make([]string, 0, len(f.configBodies[0]["plugins"]))
	for _, p := range f.configBodies[0]["plugins"] {
		bulkPlugins = append(bulkPlugins, p["name"].(string))
	}
	assert.ElementsMatch(t, []string{"key-auth", "rate-limiting"}, bulkPlugins)
	perEntityPlugins := make([]string, 0, len(f.entityRequests["plugins"]))
	for _, p := range f.entityRequests["plugins"] {
		perEntityPlugins = append(perEntityPlugins, p["name"].(string))
	}
	assert.ElementsMatch(t, []string{
		"example-uploaded-plugin", "example-streamed-plugin", "example-file-plugin", "clone-of-custom",
		"rate-limiting-advanced",
	}, perEntityPlugins)
}

func TestValidate_ClonedPluginIsInEachChunkOfItsPlugins(t *testing.T) {
	f := newFakeKonnect(t)
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)
	require.NoError(t, ks.ClonedPluginDefinitions.Add(state.ClonedPluginDefinition{
		ClonedPluginDefinition: kong.ClonedPluginDefinition{
			ID: kong.String("40000000-0000-0000-0000-000000000001"), Name: kong.String("clone-rate-limiting"),
			Ref: kong.String("rate-limiting"),
		},
	}))
	require.NoError(t, ks.Plugins.Add(state.Plugin{Plugin: kong.Plugin{
		ID: kong.String("40000000-0000-0000-0001-000000000001"), Name: kong.String("clone-rate-limiting"),
		Route: &kong.Route{ID: kong.String("00000000-0000-0000-0001-000000000001")},
	}}))

	errs := f.validator(t, ks, "Plugins").Validate(format30)

	require.Empty(t, errs)
	require.Len(t, f.configBodies, 1)
	requireReferencesInChunk(t, f.configBodies[0])
	assert.Len(t, f.configBodies[0]["plugins"], 3)
	assert.Len(t, f.configBodies[0]["cloned_plugins"], 1)
}

func TestValidate_ConfigEndpointRateLimit(t *testing.T) {
	waits := noSleep(t)
	f := newFakeKonnect(t)
	f.configResponses = []*fakeResponse{
		{status: http.StatusTooManyRequests, headers: map[string]string{"Retry-After": "7"}},
		{status: http.StatusTooManyRequests},
	}
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)

	errs := f.validator(t, ks).Validate(format30)

	require.Empty(t, errs)
	assert.Equal(t, []time.Duration{7 * time.Second, 2 * time.Second}, *waits)
	assert.Len(t, f.configBodies, 3)
}

func TestValidate_ConfigEndpointIssues(t *testing.T) {
	f := newFakeKonnect(t)
	f.configResponses = []*fakeResponse{{
		status: http.StatusBadRequest,
		body: `{"items":[{"description":"conflicting entity","resource":` +
			`{"id":"00000000-0000-0000-0001-000000000001","type":"route"}}]}`,
	}}
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)

	errs := f.validator(t, ks).Validate(format30)

	require.Len(t, errs, 1)
	require.EqualError(t, errs[0],
		"validate entity 'route (00000000-0000-0000-0001-000000000001)': conflicting entity")
	assert.Empty(t, f.entityRequests)
}

func TestValidate_ConfigEndpointEntityErrorFindsEntities(t *testing.T) {
	f := newFakeKonnect(t)
	f.configResponses = []*fakeResponse{{
		status: http.StatusBadRequest,
		body:   `{"code":3,"message":"validation error","details":[{"field":"port"}]}`,
	}}
	f.invalidEntities["svc-2"] = true
	f.invalidEntities["svc-3"] = true
	ks := newTestState(t)
	for i := 1; i <= 3; i++ {
		addServiceTree(t, ks, i, nil)
	}

	errs := f.validator(t, ks).Validate(format30)

	require.Len(t, errs, 2)
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	assert.ElementsMatch(t, []string{
		`validate entity 'services (svc-2)': HTTP status 400 (message: "schema violation (svc-2)")`,
		`validate entity 'services (svc-3)': HTTP status 400 (message: "schema violation (svc-3)")`,
	}, messages)
	assert.Len(t, f.entityRequests["services"], 3)
	assert.Len(t, f.entityRequests["routes"], 3)
	assert.Len(t, f.entityRequests["plugins"], 6)
}

func TestValidate_ConfigEndpointEntityErrorWithoutPerEntityError(t *testing.T) {
	f := newFakeKonnect(t)
	f.configResponses = []*fakeResponse{{
		status: http.StatusBadRequest,
		body:   `{"code":3,"message":"validation error","details":[{"field":"ref"}]}`,
	}}
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)

	errs := f.validator(t, ks).Validate(format30)

	require.Len(t, errs, 1)
	require.EqualError(t, errs[0],
		`validate configuration: HTTP status 400 (message: "validation error", details: [{"field":"ref"}])`)
}

func TestValidate_ConfigEndpointEntityErrorFindsAllBulkEntityKinds(t *testing.T) {
	f := newFakeKonnect(t)
	f.configResponses = []*fakeResponse{
		{status: http.StatusBadRequest, body: `{"message":"validation error"}`},
		{status: http.StatusBadRequest, body: `{"message":"invalid cloned plugin"}`},
	}
	f.invalidEntities["bad-consumer-group"] = true
	f.invalidEntities["bad-key"] = true
	f.invalidEntities["bad-key-set"] = true
	ks := newTestState(t)
	require.NoError(t, ks.ConsumerGroups.Add(state.ConsumerGroup{ConsumerGroup: kong.ConsumerGroup{
		ID: kong.String("70000000-0000-0000-0000-000000000001"), Name: kong.String("bad-consumer-group"),
	}}))
	require.NoError(t, ks.KeySets.Add(state.KeySet{KeySet: kong.KeySet{
		ID: kong.String("70000000-0000-0000-0000-000000000002"), Name: kong.String("bad-key-set"),
	}}))
	require.NoError(t, ks.Keys.Add(state.Key{Key: kong.Key{
		ID: kong.String("70000000-0000-0000-0000-000000000004"), Name: kong.String("bad-key"),
	}}))
	require.NoError(t, ks.ClonedPluginDefinitions.Add(state.ClonedPluginDefinition{
		ClonedPluginDefinition: kong.ClonedPluginDefinition{
			ID: kong.String("70000000-0000-0000-0000-000000000003"), Name: kong.String("bad-clone"),
			Ref: kong.String("rate-limiting"),
		},
	}))

	errs := f.validator(t, ks).Validate(format30)

	require.Len(t, errs, 4)
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	assert.ElementsMatch(t, []string{
		`validate entity 'consumer_groups (bad-consumer-group)': HTTP status 400 (message: "schema violation (bad-consumer-group)")`,
		`validate entity 'keys (bad-key)': HTTP status 400 (message: "schema violation (bad-key)")`,
		`validate entity 'key_sets (bad-key-set)': HTTP status 400 (message: "schema violation (bad-key-set)")`,
		`validate entity 'cloned_plugins (70000000-0000-0000-0000-000000000003)': ` +
			`HTTP status 400 (message: "invalid cloned plugin")`,
	}, messages)
}

func TestValidate_NotKonnectUsesPerEntityValidation(t *testing.T) {
	f := newFakeKonnect(t)
	ks := newTestState(t)
	v := f.validator(t, ks)
	v.client.SetKonnectFlag(false)
	assert.Empty(t, v.configValidateURL())
}

func TestValidate_ConflictingEntitiesShareAChunk(t *testing.T) {
	f := newFakeKonnect(t)
	ks := newTestState(t)
	config := kong.Configuration{"padding": strings.Repeat("x", 2048)}
	const services = 1000
	for i := 0; i < services; i++ {
		addServiceTree(t, ks, i, config)
	}
	// The plugins are far apart in the order. Only the conflict puts them into one chunk.
	for _, i := range []int{1, services - 1} {
		require.NoError(t, ks.Plugins.Add(state.Plugin{Plugin: kong.Plugin{
			ID: kong.String(fmt.Sprintf("00000000-0000-0000-0005-%012d", i)), Name: kong.String("cors"),
			InstanceName: kong.String("same-instance"),
			Route:        &kong.Route{ID: kong.String(fmt.Sprintf("00000000-0000-0000-0001-%012d", i))},
		}}))
	}

	errs := f.validator(t, ks).Validate(format30)

	require.Empty(t, errs)
	require.Greater(t, len(f.configBodies), 1)
	chunksWithConflict := 0
	for _, config := range f.configBodies {
		requireReferencesInChunk(t, config)
		count := 0
		for _, p := range config["plugins"] {
			if p["instance_name"] == "same-instance" {
				count++
			}
		}
		if count > 0 {
			assert.Equal(t, 2, count)
			chunksWithConflict++
		}
	}
	assert.Equal(t, 1, chunksWithConflict)
}

func TestValidate_ConflictWithPerEntityPlugin(t *testing.T) {
	f := newFakeKonnect(t)
	f.pluginSchemas = []string{"example-custom-plugin"}
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)
	for i, name := range []string{"cors", "example-custom-plugin"} {
		require.NoError(t, ks.Plugins.Add(state.Plugin{Plugin: kong.Plugin{
			ID:           kong.String(fmt.Sprintf("50000000-0000-0000-0000-%012d", i)),
			Name:         kong.String(name),
			InstanceName: kong.String("same-instance"),
			Service:      &kong.Service{ID: kong.String("00000000-0000-0000-0000-000000000001")},
		}}))
	}

	errs := f.validator(t, ks).Validate(format30)

	require.Len(t, errs, 1)
	require.EqualError(t, errs[0], `validate entity 'plugins (50000000-0000-0000-0000-000000000001)': `+
		`instance_name "same-instance" is also used by plugins 50000000-0000-0000-0000-000000000000`)
}

func TestValidate_PerEntityValidationFindsConflicts(t *testing.T) {
	f := newFakeKonnect(t)
	f.configEnabled = false
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)
	addServiceTree(t, ks, 2, nil)
	for i := 1; i <= 2; i++ {
		require.NoError(t, ks.Plugins.Add(state.Plugin{Plugin: kong.Plugin{
			ID: kong.String(fmt.Sprintf("60000000-0000-0000-0000-%012d", i)), Name: kong.String("cors"),
			InstanceName: kong.String("same-instance"),
			Service:      &kong.Service{ID: kong.String(fmt.Sprintf("00000000-0000-0000-0000-%012d", i))},
		}}))
	}

	errs := f.validator(t, ks).Validate(format30)

	require.Len(t, errs, 1)
	require.EqualError(t, errs[0], `validate entity 'plugins (60000000-0000-0000-0000-000000000002)': `+
		`instance_name "same-instance" is also used by plugins 60000000-0000-0000-0000-000000000001`)
}

func TestValidate_ConfigClientSendsConfigRequests(t *testing.T) {
	f := newFakeKonnect(t)
	var mu sync.Mutex
	agents := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		agents[r.URL.Path] = r.Header.Get("User-Agent")
		mu.Unlock()
		f.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	newClient := func(agent string) *kong.Client {
		c, err := kong.NewClient(kong.String(srv.URL+testCoreEntitiesPath), srv.Client())
		require.NoError(t, err)
		c.SetKonnectFlag(true)
		c.UserAgent = agent
		return c
	}
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)

	errs := NewValidator(ValidatorOpts{
		Ctx: context.Background(), State: ks, Parallelism: 1,
		Client: newClient("default-client"), ConfigClient: newClient("config-client"),
	}).Validate(format30)

	require.Empty(t, errs)
	assert.Equal(t, "config-client", agents[testConfigPath])
	assert.Equal(t, "default-client", agents[testCoreEntitiesPath+"/plugin-schemas"])
}

// addConsumerCredentials adds a consumer with one credential of each kind.
func addConsumerCredentials(t *testing.T, ks *state.KongState, i int) {
	t.Helper()
	consumer := &kong.Consumer{ID: kong.String(fmt.Sprintf("80000000-0000-0000-0000-%012d", i))}
	caCertID := fmt.Sprintf("80000000-0000-0000-0001-%012d", i)
	id := func(n int) *string { return kong.String(fmt.Sprintf("80000000-0000-0000-%04d-%012d", n, i)) }
	require.NoError(t, ks.Consumers.Add(state.Consumer{Consumer: kong.Consumer{
		ID: consumer.ID, Username: kong.String(fmt.Sprintf("consumer-%d", i)),
	}}))
	require.NoError(t, ks.CACertificates.Add(state.CACertificate{CACertificate: kong.CACertificate{
		ID: kong.String(caCertID), Cert: kong.String(fmt.Sprintf("example-cert-%d", i)),
	}}))
	require.NoError(t, ks.KeyAuths.Add(state.KeyAuth{KeyAuth: kong.KeyAuth{
		ID: id(2), Key: kong.String(fmt.Sprintf("example-key-%d", i)), Consumer: consumer,
	}}))
	require.NoError(t, ks.BasicAuths.Add(state.BasicAuth{BasicAuth: kong.BasicAuth{
		ID: id(3), Username: kong.String(fmt.Sprintf("basic-%d", i)), Password: kong.String("example"), Consumer: consumer,
	}}))
	require.NoError(t, ks.HMACAuths.Add(state.HMACAuth{HMACAuth: kong.HMACAuth{
		ID: id(4), Username: kong.String(fmt.Sprintf("hmac-%d", i)), Consumer: consumer,
	}}))
	require.NoError(t, ks.JWTAuths.Add(state.JWTAuth{JWTAuth: kong.JWTAuth{
		ID: id(5), Key: kong.String(fmt.Sprintf("jwt-%d", i)), Consumer: consumer,
	}}))
	require.NoError(t, ks.ACLGroups.Add(state.ACLGroup{ACLGroup: kong.ACLGroup{
		ID: id(6), Group: kong.String("example-group"), Consumer: consumer,
	}}))
	require.NoError(t, ks.MTLSAuths.Add(state.MTLSAuth{MTLSAuth: kong.MTLSAuth{
		ID: id(7), SubjectName: kong.String(fmt.Sprintf("client-%d", i)), Consumer: consumer,
		CACertificate: &kong.CACertificate{ID: kong.String(caCertID)},
	}}))
}

func TestValidate_CredentialsUseTheConfigEndpoint(t *testing.T) {
	f := newFakeKonnect(t)
	ks := newTestState(t)
	const consumers = 3000
	for i := 0; i < consumers; i++ {
		addConsumerCredentials(t, ks, i)
	}

	errs := f.validator(t, ks).Validate(format30)

	require.Empty(t, errs)
	require.Greater(t, len(f.configBodies), 1)
	counts := map[string]int{}
	for _, config := range f.configBodies {
		requireReferencesInChunk(t, config)
		for kind := range credentialKinds {
			counts[kind] += len(config[kind])
		}
	}
	for kind := range credentialKinds {
		assert.Equalf(t, consumers, counts[kind], "credentials of %s", kind)
	}
	assert.Empty(t, f.entityRequests)
}

func TestValidate_CredentialsUsePerEntityValidationIfTheEndpointRejectsThem(t *testing.T) {
	f := newFakeKonnect(t)
	f.rejectCredentials = true
	ks := newTestState(t)
	addServiceTree(t, ks, 1, nil)
	addConsumerCredentials(t, ks, 1)

	errs := f.validator(t, ks).Validate(format30)

	require.Empty(t, errs)
	require.Len(t, f.configBodies, 1)
	requireReferencesInChunk(t, f.configBodies[0])
	assert.Len(t, f.configBodies[0]["services"], 1)
	assert.Len(t, f.configBodies[0]["consumers"], 1)
	for kind := range credentialKinds {
		assert.Emptyf(t, f.configBodies[0][kind], "bulk request for %s", kind)
	}
	for _, kind := range []string{"keyauth_credentials", "basicauth_credentials", "hmacauth_credentials", "jwt_secrets", "acls"} {
		assert.Lenf(t, f.entityRequests[kind], 1, "per-entity requests for %s", kind)
	}
}
