package convert

import (
	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
)

// updateRoutesFor314 updates all routes in the content to explicitly set the
// old default protocols ["http", "https"] if no protocols are specified.
// In Kong Gateway 3.14, the default protocols for routes changed from
// ["http", "https"] to ["https"] only.
func updateRoutesFor314(content *file.Content) {
	for idx := range content.Routes {
		setRouteProtocolDefaultsFor314(&content.Routes[idx])
	}

	for _, service := range content.Services {
		for _, route := range service.Routes {
			setRouteProtocolDefaultsFor314(route)
		}
	}
}

// setRouteProtocolDefaultsFor314 sets the old default protocols on a route
// if the protocols field is not explicitly set.
func setRouteProtocolDefaultsFor314(route *file.FRoute) {
	if route == nil {
		return
	}
	if len(route.Protocols) == 0 {
		route.Protocols = []*string{
			kong.String("http"),
			kong.String("https"),
		}
		routeName := "<unnamed>"
		if route.Name != nil {
			routeName = *route.Name
		} else if route.ID != nil {
			routeName = *route.ID
		}
		cprint.UpdatePrintf(
			"Route '%s': setting protocols to [\"http\", \"https\"] "+
				"(old default, 3.14 defaults to [\"https\"] only)\n", routeName)
	}
}
