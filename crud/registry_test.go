package crud

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testActionFixture struct {
	state string
}

func newTestActionFixture(state string) testActionFixture {
	return testActionFixture{state: state}
}

func (t testActionFixture) invoke(op string, inputs ...Arg) (Arg, error) {
	res := t.state + " " + op

	for _, input := range inputs {
		iString, ok := input.(string)
		if !ok {
			return nil, fmt.Errorf("input is not a string")
		}
		res += " " + iString
	}
	return res, nil
}

func (t testActionFixture) Create(ctx context.Context, input ...Arg) (Arg, error) {
	return t.invoke("create", input...)
}

func (t testActionFixture) Delete(ctx context.Context, input ...Arg) (Arg, error) {
	return t.invoke("delete", input...)
}

func (t testActionFixture) Update(ctx context.Context, input ...Arg) (Arg, error) {
	return t.invoke("update", input...)
}

func TestRegistryRegister(t *testing.T) {
	assert := assert.New(t)
	var r Registry
	var a Actions = newTestActionFixture("yolo")

	err := r.Register("", nil)
	assert.NotNil(err)

	err = r.Register("foo", a)
	assert.Nil(err)

	err = r.Register("foo", a)
	assert.NotNil(err)
}

func TestRegistryMustRegister(t *testing.T) {
	assert := assert.New(t)
	var r Registry
	var a Actions = newTestActionFixture("yolo")

	assert.Panics(func() {
		r.MustRegister("", nil)
	})

	assert.NotPanics(func() {
		r.MustRegister("foo", a)
	})

	assert.Panics(func() {
		r.MustRegister("foo", a)
	})
}

func TestRegistryGet(t *testing.T) {
	assert := assert.New(t)
	var r Registry
	var a Actions = newTestActionFixture("foo")

	err := r.Register("foo", a)
	assert.Nil(err)

	a, err = r.Get("foo")
	assert.Nil(err)
	assert.NotNil(a)

	a, err = r.Get("bar")
	assert.NotNil(err)
	assert.Nil(a)

	a, err = r.Get("")
	assert.NotNil(err)
	assert.Nil(a)
}

func TestRegistryCreate(t *testing.T) {
	assert := assert.New(t)
	var r Registry
	var a Actions = newTestActionFixture("foo")

	err := r.Register("foo", a)
	assert.Nil(err)

	res, err := r.Create(context.Background(), "foo", "yolo")
	assert.Nil(err)
	assert.NotNil(res)
	result, ok := res.(string)
	assert.True(ok)
	assert.Equal("foo create yolo", result)

	// make sure it takes multiple arguments
	res, err = r.Create(context.Background(), "foo", "yolo", "always")
	assert.Nil(err)
	assert.NotNil(res)
	result, ok = res.(string)
	assert.True(ok)
	assert.Equal("foo create yolo always", result)

	res, err = r.Create(context.Background(), "foo", 42)
	assert.NotNil(err)
	assert.Nil(res)

	res, err = r.Create(context.Background(), "bar", 42)
	assert.NotNil(err)
	assert.Nil(res)
}

func TestRegistryUpdate(t *testing.T) {
	assert := assert.New(t)
	var r Registry
	var a Actions = newTestActionFixture("foo")

	err := r.Register("foo", a)
	assert.Nil(err)

	res, err := r.Update(context.Background(), "foo", "yolo")
	assert.Nil(err)
	assert.NotNil(res)
	result, ok := res.(string)
	assert.True(ok)
	assert.Equal("foo update yolo", result)

	// make sure it takes multiple arguments
	res, err = r.Update(context.Background(), "foo", "yolo", "always")
	assert.Nil(err)
	assert.NotNil(res)
	result, ok = res.(string)
	assert.True(ok)
	assert.Equal("foo update yolo always", result)

	res, err = r.Update(context.Background(), "foo", 42)
	assert.NotNil(err)
	assert.Nil(res)

	res, err = r.Update(context.Background(), "bar", 42)
	assert.NotNil(err)
	assert.Nil(res)
}

func TestRegistryDelete(t *testing.T) {
	assert := assert.New(t)
	var r Registry
	var a Actions = newTestActionFixture("foo")

	err := r.Register("foo", a)
	assert.Nil(err)

	res, err := r.Delete(context.Background(), "foo", "yolo")
	assert.Nil(err)
	assert.NotNil(res)
	result, ok := res.(string)
	assert.True(ok)
	assert.Equal("foo delete yolo", result)

	// make sure it takes multiple arguments
	res, err = r.Delete(context.Background(), "foo", "yolo", "always")
	assert.Nil(err)
	assert.NotNil(res)
	result, ok = res.(string)
	assert.True(ok)
	assert.Equal("foo delete yolo always", result)

	res, err = r.Delete(context.Background(), "foo", 42)
	assert.NotNil(err)
	assert.Nil(res)

	res, err = r.Delete(context.Background(), "bar", 42)
	assert.NotNil(err)
	assert.Nil(res)
}

func TestRegistryDo(t *testing.T) {
	assert := assert.New(t)
	var r Registry
	var a Actions = newTestActionFixture("foo")

	err := r.Register("foo", a)
	assert.Nil(err)

	res, err := r.Do(context.Background(), "foo", Create, "yolo")
	assert.Nil(err)
	assert.NotNil(res)
	result, ok := res.(string)
	assert.True(ok)
	assert.Equal("foo create yolo", result)

	// make sure it takes multiple arguments
	res, err = r.Do(context.Background(), "foo", Update, "yolo", "always")
	assert.Nil(err)
	assert.NotNil(res)
	result, ok = res.(string)
	assert.True(ok)
	assert.Equal("foo update yolo always", result)

	res, err = r.Do(context.Background(), "foo", Delete, 42)
	assert.NotNil(err)
	assert.Nil(res)

	res, err = r.Do(context.Background(), "foo", Op{"unknown-op"}, 42)
	assert.NotNil(err)
	assert.Nil(res)

	res, err = r.Do(context.Background(), "bar", Create, "yolo")
	assert.NotNil(err)
	assert.Nil(res)
}
