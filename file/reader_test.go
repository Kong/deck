package file

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func Test_ensureJSON(t *testing.T) {
	type args struct {
		m map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			"empty array is kept as is",
			args{map[string]interface{}{
				"foo": []interface{}{},
			}},
			map[string]interface{}{
				"foo": []interface{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ensureJSON(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ensureJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadKongStateFromStdinFailsToParseText(t *testing.T) {
	filenames := []string{"-"}
	assert := assert.New(t)
	assert.Equal("-", filenames[0])

	var content bytes.Buffer
	content.Write([]byte("hunter2\n"))

	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content.Bytes()); err != nil {
		panic(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		panic(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile

	c, err := GetContentFromFiles(filenames)
	assert.NotNil(err)
	assert.Nil(c)
}

func TestReadKongStateFromStdin(t *testing.T) {
	filenames := []string{"-"}
	assert := assert.New(t)
	assert.Equal("-", filenames[0])

	var content bytes.Buffer
	content.Write([]byte("services:\n- host: test.com\n  name: test service\n"))

	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content.Bytes()); err != nil {
		panic(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		panic(err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	os.Stdin = tmpfile

	c, err := GetContentFromFiles(filenames)
	assert.NotNil(c)
	assert.Nil(err)

	assert.Equal(kong.Service{
		Name: kong.String("test service"),
		Host: kong.String("test.com"),
	},
		c.Services[0].Service)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	rand.Seed(42)
	type fields struct {
		currentState  *state.KongState
		targetContent *Content
		kongVersion   semver.Version
	}
	tests := []struct {
		name   string
		fields fields
		want   *utils.KongRawState
	}{
		{
			name: "generates ID for a non-existing consumer",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
						},
					},
					Info: &Info{
						SelectorTags: []string{"tag1"},
					},
				},
				currentState: emptyState(),
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("538c7f96-b164-4f1b-97bb-9f4bb472e89f"),
						Username: kong.String("foo"),
						Tags:     kong.StringSlice("tag1"),
					},
				},
			},
		},
		{
			name: "generates ID for a non-existing credential",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
							KeyAuths: []*kong.KeyAuth{
								{
									Key: kong.String("foo-key"),
								},
							},
							BasicAuths: []*kong.BasicAuth{
								{
									Username: kong.String("basic-username"),
									Password: kong.String("basic-password"),
								},
							},
							HMACAuths: []*kong.HMACAuth{
								{
									Username: kong.String("hmac-username"),
									Secret:   kong.String("hmac-secret"),
								},
							},
							JWTAuths: []*kong.JWTAuth{
								{
									Key:    kong.String("jwt-key"),
									Secret: kong.String("jwt-secret"),
								},
							},
							Oauth2Creds: []*kong.Oauth2Credential{
								{
									ClientID: kong.String("oauth2-clientid"),
									Name:     kong.String("oauth2-name"),
								},
							},
							ACLGroups: []*kong.ACLGroup{
								{
									Group: kong.String("foo-group"),
								},
							},
						},
					},
					Info: &Info{
						SelectorTags: []string{"tag1"},
					},
				},
				currentState: emptyState(),
				kongVersion:  kong140Version,
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						Username: kong.String("foo"),
						Tags:     kong.StringSlice("tag1"),
					},
				},
				KeyAuths: []*kong.KeyAuth{
					{
						ID:  kong.String("dfd79b4d-7642-4b61-ba0c-9f9f0d3ba55b"),
						Key: kong.String("foo-key"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				BasicAuths: []*kong.BasicAuth{
					{
						ID:       kong.String("0cc0d614-4c88-4535-841a-cbe0709b0758"),
						Username: kong.String("basic-username"),
						Password: kong.String("basic-password"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				HMACAuths: []*kong.HMACAuth{
					{
						ID:       kong.String("083f61d3-75bc-42b4-9df4-f91929e18fda"),
						Username: kong.String("hmac-username"),
						Secret:   kong.String("hmac-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				JWTAuths: []*kong.JWTAuth{
					{
						ID:     kong.String("9e6f82e5-4e74-4e81-a79e-4bbd6fe34cdc"),
						Key:    kong.String("jwt-key"),
						Secret: kong.String("jwt-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				Oauth2Creds: []*kong.Oauth2Credential{
					{
						ID:       kong.String("ba843ee8-d63e-4c4f-be1c-ebea546d8fac"),
						ClientID: kong.String("oauth2-clientid"),
						Name:     kong.String("oauth2-name"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				ACLGroups: []*kong.ACLGroup{
					{
						ID:    kong.String("13dd1aac-04ce-4ea2-877c-5579cfa2c78e"),
						Group: kong.String("foo-group"),
						Consumer: &kong.Consumer{
							ID: kong.String("5b1484f2-5209-49d9-b43e-92ba09dd9d52"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				MTLSAuths: nil,
			},
		},
		{
			name: "matches ID of an existing consumer",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
						},
					},
				},
				currentState: existingConsumerCredState(),
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Username: kong.String("foo"),
					},
				},
			},
		},
		{
			name: "matches ID of an existing credential",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
							KeyAuths: []*kong.KeyAuth{
								{
									Key: kong.String("foo-apikey"),
								},
							},
							BasicAuths: []*kong.BasicAuth{
								{
									Username: kong.String("basic-username"),
									Password: kong.String("basic-password"),
								},
							},
							HMACAuths: []*kong.HMACAuth{
								{
									Username: kong.String("hmac-username"),
									Secret:   kong.String("hmac-secret"),
								},
							},
							JWTAuths: []*kong.JWTAuth{
								{
									Key:    kong.String("jwt-key"),
									Secret: kong.String("jwt-secret"),
								},
							},
							Oauth2Creds: []*kong.Oauth2Credential{
								{
									ClientID: kong.String("oauth2-clientid"),
									Name:     kong.String("oauth2-name"),
								},
							},
							ACLGroups: []*kong.ACLGroup{
								{
									Group: kong.String("foo-group"),
								},
							},
							MTLSAuths: []*kong.MTLSAuth{
								{
									ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
									SubjectName: kong.String("test@example.com"),
								},
							},
						},
					},
					Info: &Info{
						SelectorTags: []string{"tag1"},
					},
				},
				currentState: existingConsumerCredState(),
				kongVersion:  kong140Version,
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Username: kong.String("foo"),
						Tags:     kong.StringSlice("tag1"),
					},
				},
				KeyAuths: []*kong.KeyAuth{
					{
						ID:  kong.String("5f1ef1ea-a2a5-4a1b-adbb-b0d3434013e5"),
						Key: kong.String("foo-apikey"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				BasicAuths: []*kong.BasicAuth{
					{
						ID:       kong.String("92f4c849-960b-43af-aad3-f307051408d3"),
						Username: kong.String("basic-username"),
						Password: kong.String("basic-password"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				HMACAuths: []*kong.HMACAuth{
					{
						ID:       kong.String("e5d81b73-bf9e-42b0-9d68-30a1d791b9c9"),
						Username: kong.String("hmac-username"),
						Secret:   kong.String("hmac-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				JWTAuths: []*kong.JWTAuth{
					{
						ID:     kong.String("917b9402-1be0-49d2-b482-ca4dccc2054e"),
						Key:    kong.String("jwt-key"),
						Secret: kong.String("jwt-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				Oauth2Creds: []*kong.Oauth2Credential{
					{
						ID:       kong.String("4eef5285-3d6a-4f6b-b659-8957a940e2ca"),
						ClientID: kong.String("oauth2-clientid"),
						Name:     kong.String("oauth2-name"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				ACLGroups: []*kong.ACLGroup{
					{
						ID:    kong.String("b7c9352a-775a-4ba5-9869-98e926a3e6cb"),
						Group: kong.String("foo-group"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				MTLSAuths: []*kong.MTLSAuth{
					{
						ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
						SubjectName: kong.String("test@example.com"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
			},
		},
		{
			name: "does not inject tags if Kong version is older than 1.4",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
							KeyAuths: []*kong.KeyAuth{
								{
									Key: kong.String("foo-apikey"),
								},
							},
							BasicAuths: []*kong.BasicAuth{
								{
									Username: kong.String("basic-username"),
									Password: kong.String("basic-password"),
								},
							},
							HMACAuths: []*kong.HMACAuth{
								{
									Username: kong.String("hmac-username"),
									Secret:   kong.String("hmac-secret"),
								},
							},
							JWTAuths: []*kong.JWTAuth{
								{
									Key:    kong.String("jwt-key"),
									Secret: kong.String("jwt-secret"),
								},
							},
							Oauth2Creds: []*kong.Oauth2Credential{
								{
									ClientID: kong.String("oauth2-clientid"),
									Name:     kong.String("oauth2-name"),
								},
							},
							ACLGroups: []*kong.ACLGroup{
								{
									Group: kong.String("foo-group"),
								},
							},
							MTLSAuths: []*kong.MTLSAuth{
								{
									ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
									SubjectName: kong.String("test@example.com"),
								},
							},
						},
					},
					Info: &Info{
						SelectorTags: []string{"tag1"},
					},
				},
				currentState: existingConsumerCredState(),
				kongVersion:  kong130Version,
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Username: kong.String("foo"),
						Tags:     kong.StringSlice("tag1"),
					},
				},
				KeyAuths: []*kong.KeyAuth{
					{
						ID:  kong.String("5f1ef1ea-a2a5-4a1b-adbb-b0d3434013e5"),
						Key: kong.String("foo-apikey"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
				BasicAuths: []*kong.BasicAuth{
					{
						ID:       kong.String("92f4c849-960b-43af-aad3-f307051408d3"),
						Username: kong.String("basic-username"),
						Password: kong.String("basic-password"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
				HMACAuths: []*kong.HMACAuth{
					{
						ID:       kong.String("e5d81b73-bf9e-42b0-9d68-30a1d791b9c9"),
						Username: kong.String("hmac-username"),
						Secret:   kong.String("hmac-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
				JWTAuths: []*kong.JWTAuth{
					{
						ID:     kong.String("917b9402-1be0-49d2-b482-ca4dccc2054e"),
						Key:    kong.String("jwt-key"),
						Secret: kong.String("jwt-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
				Oauth2Creds: []*kong.Oauth2Credential{
					{
						ID:       kong.String("4eef5285-3d6a-4f6b-b659-8957a940e2ca"),
						ClientID: kong.String("oauth2-clientid"),
						Name:     kong.String("oauth2-name"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
				ACLGroups: []*kong.ACLGroup{
					{
						ID:    kong.String("b7c9352a-775a-4ba5-9869-98e926a3e6cb"),
						Group: kong.String("foo-group"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
				MTLSAuths: []*kong.MTLSAuth{
					{
						ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
						SubjectName: kong.String("test@example.com"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
					},
				},
			},
		},
		{
			name: "inject tags if Kong version is newer than 1.4",
			fields: fields{
				targetContent: &Content{
					Consumers: []FConsumer{
						{
							Consumer: kong.Consumer{
								Username: kong.String("foo"),
							},
							KeyAuths: []*kong.KeyAuth{
								{
									Key: kong.String("foo-apikey"),
								},
							},
							BasicAuths: []*kong.BasicAuth{
								{
									Username: kong.String("basic-username"),
									Password: kong.String("basic-password"),
								},
							},
							HMACAuths: []*kong.HMACAuth{
								{
									Username: kong.String("hmac-username"),
									Secret:   kong.String("hmac-secret"),
								},
							},
							JWTAuths: []*kong.JWTAuth{
								{
									Key:    kong.String("jwt-key"),
									Secret: kong.String("jwt-secret"),
								},
							},
							Oauth2Creds: []*kong.Oauth2Credential{
								{
									ClientID: kong.String("oauth2-clientid"),
									Name:     kong.String("oauth2-name"),
								},
							},
							ACLGroups: []*kong.ACLGroup{
								{
									Group: kong.String("foo-group"),
								},
							},
							MTLSAuths: []*kong.MTLSAuth{
								{
									ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
									SubjectName: kong.String("test@example.com"),
								},
							},
						},
					},
					Info: &Info{
						SelectorTags: []string{"tag1"},
					},
				},
				currentState: existingConsumerCredState(),
				kongVersion:  kong230Version,
			},
			want: &utils.KongRawState{
				Consumers: []*kong.Consumer{
					{
						ID:       kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						Username: kong.String("foo"),
						Tags:     kong.StringSlice("tag1"),
					},
				},
				KeyAuths: []*kong.KeyAuth{
					{
						ID:  kong.String("5f1ef1ea-a2a5-4a1b-adbb-b0d3434013e5"),
						Key: kong.String("foo-apikey"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				BasicAuths: []*kong.BasicAuth{
					{
						ID:       kong.String("92f4c849-960b-43af-aad3-f307051408d3"),
						Username: kong.String("basic-username"),
						Password: kong.String("basic-password"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				HMACAuths: []*kong.HMACAuth{
					{
						ID:       kong.String("e5d81b73-bf9e-42b0-9d68-30a1d791b9c9"),
						Username: kong.String("hmac-username"),
						Secret:   kong.String("hmac-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				JWTAuths: []*kong.JWTAuth{
					{
						ID:     kong.String("917b9402-1be0-49d2-b482-ca4dccc2054e"),
						Key:    kong.String("jwt-key"),
						Secret: kong.String("jwt-secret"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				Oauth2Creds: []*kong.Oauth2Credential{
					{
						ID:       kong.String("4eef5285-3d6a-4f6b-b659-8957a940e2ca"),
						ClientID: kong.String("oauth2-clientid"),
						Name:     kong.String("oauth2-name"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				ACLGroups: []*kong.ACLGroup{
					{
						ID:    kong.String("b7c9352a-775a-4ba5-9869-98e926a3e6cb"),
						Group: kong.String("foo-group"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
				MTLSAuths: []*kong.MTLSAuth{
					{
						ID:          kong.String("533c259e-bf71-4d77-99d2-97944c70a6a4"),
						SubjectName: kong.String("test@example.com"),
						Consumer: &kong.Consumer{
							ID: kong.String("4bfcb11f-c962-4817-83e5-9433cf20b663"),
						},
						Tags: kong.StringSlice("tag1"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, err := Get(tt.fields.targetContent, RenderConfig{
				CurrentState: tt.fields.currentState,
				KongVersion:  tt.fields.kongVersion,
			})
			if err != nil {
				panic(err)
			}
			assert.Equal(tt.want, state)
		})
	}
}
