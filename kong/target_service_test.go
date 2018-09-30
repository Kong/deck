package kong

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTargetsUpstream(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	target := &Target{
		Target: String("10.0.0.1"),
	}

	// upstream is required
	badTarget, err := client.Targets.Create(defaultCtx, nil, target)
	assert.NotNil(err)
	assert.Nil(badTarget)

	// create a upstream
	fixtureUpstream, err := client.Upstreams.Create(defaultCtx, &Upstream{
		Name: String("vhost.com"),
	})
	assert.Nil(err)
	assert.NotNil(fixtureUpstream)
	assert.NotNil(fixtureUpstream.ID)

	createdTarget, err := client.Targets.Create(defaultCtx, fixtureUpstream.ID, &Target{
		Target: String("10.0.0.1:80"),
	})
	assert.Nil(err)
	assert.NotNil(createdTarget)

	// client.SetDebugMode(true)
	// target, err = client.Targets.Get(defaultCtx, fixtureUpstream.ID, createdTarget.ID)
	// assert.Nil(err)
	// assert.NotNil(target)
	// client.SetDebugMode(false)

	// createdTarget.Target = String("10.0.0.2")
	// target, err = client.Targets.Update(defaultCtx, fixtureUpstream.ID, createdTarget)
	// assert.Nil(err)
	// assert.NotNil(target)
	// assert.Equal("10.0.0.2", *target.Target)

	err = client.Targets.Delete(defaultCtx, fixtureUpstream.ID, createdTarget.ID)
	assert.Nil(err)

	// PUT request is not yet supported
	// TODO uncomment this target entity is migrated over to new DAO

	// ID can be specified
	// id := uuid.NewV4().String()
	// target = &Target{
	// 	ID:         String(id),
	// 	Target:     String("10.0.0.3"),
	// 	UpstreamID: fixtureUpstream.ID,
	// }

	// createdTarget, err = client.Targets.Create(defaultCtx, target)
	// assert.Nil(err)
	// assert.NotNil(createdTarget)
	// assert.Equal(id, *createdTarget.ID)

	err = client.Upstreams.Delete(defaultCtx, fixtureUpstream.ID)
	assert.Nil(err)
}

func TestTargetListEndpoint(T *testing.T) {
	assert := assert.New(T)

	client, err := NewClient(nil, nil)
	assert.Nil(err)
	assert.NotNil(client)

	upstream := &Upstream{
		Name: String("vhost2.com"),
	}

	createdUpstream, err := client.Upstreams.Create(defaultCtx, upstream)
	assert.Nil(err)
	assert.NotNil(createdUpstream)

	// fixtures
	targets := []*Target{
		{
			Target:     String("target1"),
			UpstreamID: createdUpstream.ID,
		},
		{
			Target:     String("target2"),
			UpstreamID: createdUpstream.ID,
		},
		{
			Target:     String("target3"),
			UpstreamID: createdUpstream.ID,
		},
	}

	// create fixturs
	for i := 0; i < len(targets); i++ {
		target, err := client.Targets.Create(defaultCtx, createdUpstream.ID, targets[i])
		assert.Nil(err)
		assert.NotNil(target)
		targets[i] = target
	}

	targetsFromKong, next, err := client.Targets.List(defaultCtx, createdUpstream.ID, nil)
	assert.Nil(err)
	assert.Nil(next)
	assert.NotNil(targetsFromKong)
	assert.Equal(3, len(targetsFromKong))

	// check if we see all targets
	assert.True(compareTargets(targets, targetsFromKong))

	// TODO seems like a bug in Kong in pagination of targets, uncomment this once the bug is fixed
	// Test pagination
	// targetsFromKong = []*Target{}

	// // first page
	// page1, next, err := client.Targets.List(defaultCtx, createdUpstream.ID, &ListOpt{Size: 1})
	// assert.Nil(err)
	// assert.NotNil(next)
	// assert.NotNil(page1)
	// assert.Equal(1, len(page1))
	// targetsFromKong = append(targetsFromKong, page1...)

	// // last page
	// next.Size = 2
	// page2, next, err := client.Targets.List(defaultCtx, createdUpstream.ID, next)
	// assert.Nil(err)
	// assert.Nil(next)
	// assert.NotNil(page2)
	// assert.Equal(2, len(page2))
	// targetsFromKong = append(targetsFromKong, page2...)

	// assert.True(compareTargets(targets, targetsFromKong))

	assert.Nil(client.Upstreams.Delete(defaultCtx, createdUpstream.ID))
}

func compareTargets(expected, actual []*Target) bool {
	var expectedUsernames, actualUsernames []string
	for _, target := range expected {
		expectedUsernames = append(expectedUsernames, *target.Target)
	}

	for _, target := range actual {
		actualUsernames = append(actualUsernames, *target.Target)
	}

	return (compareSlices(expectedUsernames, actualUsernames))
}
