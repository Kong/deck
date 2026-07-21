//nolint:deadcode
package integration

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kong/deck/cmd"
	deckDump "github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/require"
)

// managedByAIDeckTag is the selector tag `deck ai sync` stamps on every entity it
// manages, mirroring the scope of `deck ai dump`.
const managedByAIDeckTag = "managed_by:deck-ai"

func getKongAddress() string {
	address := os.Getenv("DECK_KONG_ADDR")
	if address != "" {
		return address
	}
	return "http://localhost:8001"
}

func getTestClient() (*kong.Client, error) {
	ctx := context.Background()
	controlPlaneName := os.Getenv("DECK_KONNECT_RUNTIME_GROUP_NAME")
	if controlPlaneName == "" {
		controlPlaneName = os.Getenv("DECK_KONNECT_CONTROL_PLANE_NAME")
	}
	konnectConfig := utils.KonnectConfig{
		Address:          os.Getenv("DECK_KONNECT_ADDR"),
		Token:            os.Getenv("DECK_KONNECT_TOKEN"),
		ControlPlaneName: controlPlaneName,
	}
	if konnectConfig.Token != "" {
		return cmd.GetKongClientForKonnectMode(ctx, &konnectConfig)
	}
	return utils.GetKongClient(utils.KongClientConfig{
		Address:   getKongAddress(),
		Retryable: true,
	})
}

func setDefaultKonnectControlPlane(t *testing.T) {
	t.Helper()
	if os.Getenv("DECK_KONNECT_CONTROL_PLANE_NAME") == "" ||
		os.Getenv("DECK_KONNECT_RUNTIME_GROUP_NAME") == "" {
		t.Setenv("DECK_KONNECT_CONTROL_PLANE_NAME", "default")
	}
}

func runWhenKonnect(t *testing.T) {
	t.Helper()

	if os.Getenv("DECK_KONNECT_TOKEN") == "" {
		t.Skip("non-Konnect test instance, skipping")
	}
}

func skipWhenKonnect(t *testing.T) {
	t.Helper()

	if os.Getenv("DECK_KONNECT_TOKEN") != "" {
		t.Skip("non-Kong test instance, skipping")
	}
}

func runWhenKongOrKonnect(t *testing.T, kongSemverRange string) {
	t.Helper()

	if os.Getenv("DECK_KONNECT_TOKEN") != "" {
		setDefaultKonnectControlPlane(t)
		return
	}
	kong.RunWhenKong(t, kongSemverRange)
	kong.SkipWhenKongRouterFlavor(t, "expressions")
}

func runWhenEnterpriseOrKonnect(t *testing.T, kongSemverRange string) {
	t.Helper()

	if os.Getenv("DECK_KONNECT_TOKEN") != "" {
		setDefaultKonnectControlPlane(t)
		return
	}
	kong.RunWhenEnterprise(t, kongSemverRange, kong.RequiredFeatures{})
	kong.SkipWhenKongRouterFlavor(t, "expressions")
}

func runWhen(t *testing.T, mode string, semverRange string) {
	t.Helper()

	switch mode {
	case "kong":
		skipWhenKonnect(t)
		kong.RunWhenKong(t, semverRange)
		kong.SkipWhenKongRouterFlavor(t, "expressions")
	case "enterprise":
		skipWhenKonnect(t)
		kong.RunWhenEnterprise(t, semverRange, kong.RequiredFeatures{})
		kong.SkipWhenKongRouterFlavor(t, "expressions")
	case "konnect":
		runWhenKonnect(t)
	}
}

func runWhenExpressions(t *testing.T, semverRange string) {
	t.Helper()
	skipWhenKonnect(t)

	// limiting to enterprise for now
	kong.RunWhenEnterprise(t, semverRange, kong.RequiredFeatures{})
	kong.RunWhenKongRouterFlavor(t, "expressions")
}

func runWhenRBAC(t *testing.T, semverRange string) {
	t.Helper()
	skipWhenKonnect(t)
	kong.RunWhenEnterprise(t, semverRange, kong.RequiredFeatures{RBAC: true})
}

// runWhenAIGateway skips the test unless it is running against a self-hosted
// AI Gateway instance within the given semver range. AI Gateway is detected via
// the "ai-gateway" marker in the Admin API Server header.
func runWhenAIGateway(t *testing.T, semverRange string) {
	t.Helper()
	skipWhenKonnect(t)
	kong.RunWhenAIGateway(t, semverRange)
}

func sortSlices(x, y interface{}) bool {
	var xName, yName string
	switch xEntity := x.(type) {
	case *kong.Service:
		yEntity := y.(*kong.Service)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case *kong.Route:
		yEntity := y.(*kong.Route)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case *kong.Vault:
		yEntity := y.(*kong.Vault)
		xName = *xEntity.Prefix
		yName = *yEntity.Prefix
	case *kong.Consumer:
		yEntity := y.(*kong.Consumer)
		if xEntity.Username != nil {
			xName = *xEntity.Username
		} else {
			xName = *xEntity.ID
		}
		if yEntity.Username != nil {
			yName = *yEntity.Username
		} else {
			yName = *yEntity.ID
		}
	case *kong.ConsumerGroup:
		yEntity := y.(*kong.ConsumerGroup)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case *kong.ConsumerGroupObject:
		yEntity := y.(*kong.ConsumerGroupObject)
		xName = *xEntity.ConsumerGroup.Name
		yName = *yEntity.ConsumerGroup.Name
	case *kong.ConsumerGroupPlugin:
		yEntity := y.(*kong.ConsumerGroupPlugin)
		xName = *xEntity.ConsumerGroup.ID
		yName = *yEntity.ConsumerGroup.ID
	case *kong.KeyAuth:
		yEntity := y.(*kong.KeyAuth)
		xName = *xEntity.Key
		yName = *yEntity.Key
	case *kong.Plugin:
		yEntity := y.(*kong.Plugin)
		xName = *xEntity.Name
		yName = *yEntity.Name
		if xEntity.Route != nil {
			xName += *xEntity.Route.ID
		}
		if xEntity.Service != nil {
			xName += *xEntity.Service.ID
		}
		if xEntity.Consumer != nil {
			xName += *xEntity.Consumer.ID
		}
		if xEntity.ConsumerGroup != nil {
			xName += *xEntity.ConsumerGroup.ID
		}
		if xEntity.Model != nil && xEntity.Model.ID != nil {
			xName += *xEntity.Model.ID
		}
		if yEntity.Route != nil {
			yName += *yEntity.Route.ID
		}
		if yEntity.Service != nil {
			yName += *yEntity.Service.ID
		}
		if yEntity.Consumer != nil {
			yName += *yEntity.Consumer.ID
		}
		if yEntity.ConsumerGroup != nil {
			yName += *yEntity.ConsumerGroup.ID
		}
		if yEntity.Model != nil && yEntity.Model.ID != nil {
			yName += *yEntity.Model.ID
		}
	case *kong.Key:
		yEntity := y.(*kong.Key)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case *kong.KeySet:
		yEntity := y.(*kong.KeySet)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case *kong.ClonedPluginDefinition:
		yEntity := y.(*kong.ClonedPluginDefinition)
		xName = *xEntity.Name
		yName = *yEntity.Name
	case *kong.CustomPluginDefinition:
		yEntity := y.(*kong.CustomPluginDefinition)
		xName = *xEntity.Name
		yName = *yEntity.Name
	}

	return xName < yName
}

func testKongState(t *testing.T, client *kong.Client, isKonnect bool,
	isConsumerGroupPolicyOverrideSet bool, expectedState utils.KongRawState,
	ignoreFields []cmp.Option,
) {
	t.Helper()

	// Get entities from Kong
	ctx := context.Background()
	dumpConfig := deckDump.Config{
		CustomEntityTypes:                []string{"degraphql_routes", "graphql_ratelimiting_cost_decorations"},
		IsConsumerGroupPolicyOverrideSet: isConsumerGroupPolicyOverrideSet,
		IncludePluginDefinitions:         true,
	}
	if expectedState.RBACEndpointPermissions != nil {
		dumpConfig.RBACResourcesOnly = true
	}
	if isKonnect {
		controlPlaneName := os.Getenv("DECK_KONNECT_CONTROL_PLANE_NAME")
		if controlPlaneName == "" {
			controlPlaneName = os.Getenv("DECK_KONNECT_CONTROL_PLANE_NAME")
		}
		if controlPlaneName != "" {
			dumpConfig.KonnectControlPlane = controlPlaneName
		} else {
			dumpConfig.KonnectControlPlane = "default"
		}
	}
	kongState, err := deckDump.Get(ctx, client, dumpConfig)
	require.NoError(t, err)

	opt := make([]cmp.Option, 0, 22+len(ignoreFields))
	opt = append(opt,
		cmpopts.IgnoreFields(kong.Service{}, "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Route{}, "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Plugin{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Upstream{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Target{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.CACertificate{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.RBACEndpointPermission{}, "Role", "CreatedAt"),
		cmpopts.IgnoreFields(kong.RBACRole{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Consumer{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.Vault{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Certificate{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.SNI{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.ConsumerGroup{}, "CreatedAt", "ID"),
		cmpopts.IgnoreFields(kong.ConsumerGroupPlugin{}, "CreatedAt", "ID"),
		cmpopts.IgnoreFields(kong.KeyAuth{}, "ID", "CreatedAt"),
		cmpopts.IgnoreFields(kong.FilterChain{}, "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Key{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.KeySet{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.Partial{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.ClonedPluginDefinition{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.CustomPluginDefinition{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.IgnoreFields(kong.AIModel{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.SortSlices(sortSlices),
		cmpopts.SortSlices(func(a, b *string) bool { return *a < *b }),
		cmpopts.EquateEmpty(),
	)
	opt = append(opt, ignoreFields...)

	if diff := cmp.Diff(kongState, &expectedState, opt...); diff != "" {
		t.Errorf("unexpected diff:\n%s", diff)
	}
}

func fetchCurrentState(ctx context.Context, client *kong.Client,
	dumpConfig deckDump.Config, t *testing.T,
) (*state.KongState, error) {
	t.Helper()
	controlPlaneName := os.Getenv("DECK_KONNECT_CONTROL_PLANE_NAME")

	if controlPlaneName != "" {
		dumpConfig.KonnectControlPlane = controlPlaneName
	}

	rawState, err := deckDump.Get(ctx, client, dumpConfig)
	if err != nil {
		return nil, err
	}

	currentState, err := state.Get(rawState)
	if err != nil {
		return nil, err
	}
	return currentState, nil
}

func reset(t *testing.T, opts ...string) {
	t.Helper()

	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "reset", "--force"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	require.NoError(t, deckCmd.Execute(), "failed to reset Kong's state")
}

func readFile(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// mustReadFile reads a file and returns its content as a string, failing the
// test immediately if the file cannot be read.
func mustReadFile(t *testing.T, filepath string) string {
	t.Helper()
	content, err := os.ReadFile(filepath)
	require.NoError(t, err)
	return string(content)
}

// setup sets deck env variable to prevent analytics in tests and registers reset
// command with t.Cleanup().
//
// NOTE: Can't be called with tests running t.Parallel() because of the usage
// of t.Setenv().
func setup(t *testing.T) {
	// disable analytics for integration tests
	t.Setenv("DECK_ANALYTICS", "off")
	t.Cleanup(func() {
		reset(t)
	})
}

func apply(ctx context.Context, kongFile string, opts ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "apply", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	return deckCmd.ExecuteContext(ctx)
}

func sync(ctx context.Context, kongFile string, opts ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "sync", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	return deckCmd.ExecuteContext(ctx)
}

// aiSync converts one or more AI Gateway source files to Kong configuration and
// syncs them directly to Kong (the equivalent of `deck ai sync <f1> <f2> ...`).
// Sources may be files, directories, or "-" for stdin. It runs
// non-interactively so it can be used in tests.
func aiSync(ctx context.Context, sourceFiles ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := append([]string{"ai", "sync"}, sourceFiles...)
	args = append(args, "--yes")
	deckCmd.SetArgs(args)
	return deckCmd.ExecuteContext(ctx)
}

// aiDump reads the AI-managed entities from Kong and writes them in AI Gateway
// format (the equivalent of `deck ai dump`), returning the generated output.
func aiDump(opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"ai", "dump"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// capture command output to be used during tests
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdErr := deckCmd.ExecuteContext(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout

	return stripansi.Strip(string(out)), cmdErr
}

func syncWithOutput(ctx context.Context, kongFile string, opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "sync", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// overwrite default standard output
	r, w, _ := os.Pipe()
	color.Output = w

	// execute decK command
	cmdErr := deckCmd.ExecuteContext(ctx)

	// read command output
	w.Close()
	out, _ := io.ReadAll(r)

	return stripansi.Strip(string(out)), cmdErr
}

func multiFileSync(ctx context.Context, kongFiles []string, opts ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "sync"}
	args = append(args, kongFiles...)
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	return deckCmd.ExecuteContext(ctx)
}

func diff(kongFile string, opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "diff", kongFile}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// overwrite default standard output
	r, w, _ := os.Pipe()
	color.Output = w

	// execute decK command
	cmdErr := deckCmd.ExecuteContext(context.Background())

	// read command output
	w.Close()
	out, _ := io.ReadAll(r)

	return stripansi.Strip(string(out)), cmdErr
}

func dump(opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "dump"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// capture command output to be used during tests
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdErr := deckCmd.ExecuteContext(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout

	return stripansi.Strip(string(out)), cmdErr
}

func fileLint(opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"file", "lint"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// capture command output to be used during tests
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdErr := deckCmd.ExecuteContext(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout

	return stripansi.Strip(string(out)), cmdErr
}

func fileLintWithStderr(opts ...string) (stdout string, stderr string, err error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"file", "lint"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// capture stdout
	rescueStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// capture stderr
	rescueStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	cmdErr := deckCmd.ExecuteContext(context.Background())

	wOut.Close()
	wErr.Close()
	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)
	os.Stdout = rescueStdout
	os.Stderr = rescueStderr

	return stripansi.Strip(string(outBytes)), stripansi.Strip(string(errBytes)), cmdErr
}

func fileFormat(opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"file", "format"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// capture command output to be used during tests
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdErr := deckCmd.ExecuteContext(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout

	return stripansi.Strip(string(out)), cmdErr
}

func fileConvert(opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"file", "convert"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// capture command output to be used during tests
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdErr := deckCmd.ExecuteContext(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout

	return stripansi.Strip(string(out)), cmdErr
}

func fileAI2Kong(opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"file", "ai2kong"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// capture command output to be used during tests
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdErr := deckCmd.ExecuteContext(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout

	return stripansi.Strip(string(out)), cmdErr
}

func ping(opts ...string) error {
	deckCmd := cmd.NewRootCmd()
	args := []string{"gateway", "ping"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)
	return deckCmd.ExecuteContext(context.Background())
}

func render(opts ...string) (string, error) {
	deckCmd := cmd.NewRootCmd()
	args := []string{"file", "render"}

	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	// capture command output to be used during tests
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdErr := deckCmd.ExecuteContext(context.Background())

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = rescueStdout

	return stripansi.Strip(string(out)), cmdErr
}

func validate(online bool, opts ...string) error {
	deckCmd := cmd.NewRootCmd()

	var args []string
	if online {
		args = []string{"gateway", "validate"}
	} else {
		args = []string{"file", "validate"}
	}

	if len(opts) > 0 {
		args = append(args, opts...)
	}
	deckCmd.SetArgs(args)

	return deckCmd.ExecuteContext(context.Background())
}

// runDualTestWithSkipDefaults runs a test twice for Konnect and Enterprise tests:
// once with default settings and once with DECK_SKIP_DEFAULTS_FILL=true
func runDualTestWithSkipDefaults(t *testing.T, testName string, testFunc func(t *testing.T)) {
	t.Run(testName+" (default fill)", func(t *testing.T) {
		testFunc(t)
	})

	t.Run(testName+" (skip defaults fill)", func(t *testing.T) {
		t.Setenv("DECK_SKIP_DEFAULTS_FILL", "true")
		testFunc(t)
	})
}

// parseAIState unmarshals a dump into file.Content so states compare
// structurally, not as text.
func parseAIState(t *testing.T, dumped string) *file.Content {
	t.Helper()
	var content file.Content
	require.NoError(t, yaml.Unmarshal([]byte(dumped), &content))
	return &content
}

// assertAIStateEqual asserts two AI-managed dumps are equivalent, ignoring
// ordering and server-side fields.
func assertAIStateEqual(t *testing.T, expected, actual string) {
	t.Helper()
	opts := []cmp.Option{
		// dump orders plugins by server ID, which differs across syncs.
		cmpopts.SortSlices(func(a, b *file.FPlugin) bool {
			return pluginSortKey(a) < pluginSortKey(b)
		}),
		// tags/paths/methods are sets; their order is not significant.
		cmpopts.SortSlices(func(a, b *string) bool { return *a < *b }),
		// KeyAuth TTL is a server-side countdown.
		cmpopts.IgnoreFields(kong.KeyAuth{}, "TTL"),
		cmpopts.EquateEmpty(),
	}
	if diff := cmp.Diff(parseAIState(t, expected), parseAIState(t, actual), opts...); diff != "" {
		t.Errorf("unexpected AI-managed state diff:\n%s", diff)
	}
}

// pluginSortKey keys a plugin by full content; json.Marshal sorts map keys, so
// equal plugins yield equal keys regardless of ID.
func pluginSortKey(p *file.FPlugin) string {
	b, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(b)
}
