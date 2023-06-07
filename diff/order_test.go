package diff

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/types"
)

func Test_reverse(t *testing.T) {
	type args struct {
		src [][]types.EntityType
	}
	tests := []struct {
		name string
		args args
		want [][]types.EntityType
	}{
		{
			name: "doesn't panic on empty slice",
			args: args{
				src: nil,
			},
			want: [][]types.EntityType{},
		},
		{
			name: "doesn't panic on empty slice",
			args: args{
				src: [][]types.EntityType{
					{"foo"},
					{"bar"},
					{"baz", "fubar"},
				},
			},
			want: [][]types.EntityType{
				{"baz", "fubar"},
				{"bar"},
				{"foo"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverse(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventsInOrder(t *testing.T) {
	e := func(entityType types.EntityType) crud.Event {
		return crud.Event{Kind: crud.Kind(entityType)}
	}

	eventsOutOfOrder := []crud.Event{
		e(types.Consumer),
		e(types.Service),
		e(types.KeyAuth),
		e(types.Route),
		e(types.ServicePackage),
		e(types.ConsumerGroup),
		e(types.ServiceVersion),
		e(types.Plugin),
	}

	order := reverseOrder()
	result := eventsInOrder(eventsOutOfOrder, order)

	require.Equal(t, [][]crud.Event{
		{
			e(types.Plugin),
		},
		{
			e(types.Route),
			e(types.ServiceVersion),
		},
		{
			e(types.Service),
			e(types.KeyAuth),
			e(types.ConsumerGroup),
		},
		{
			e(types.Consumer),
			e(types.ServicePackage),
		},
	}, result)
}
