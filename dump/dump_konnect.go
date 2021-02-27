package dump

import (
	"context"

	"github.com/kong/deck/konnect"
	"github.com/kong/deck/utils"
)

type KonnectConfig struct {
	// ID of the Kong Control Plane being managed.
	ControlPlaneID string
}

func GetFromKonnect(ctx context.Context, konnectClient *konnect.Client,
	config KonnectConfig) (*utils.KonnectRawState, error) {
	var res utils.KonnectRawState
	servicePackages, err := konnectClient.ServicePackages.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, sp := range servicePackages {
		versions, err := konnectClient.ServiceVersions.ListForPackage(ctx, sp.ID)
		if err != nil {
			return nil, err
		}
		sp.Versions = versions
	}

	relations, err := konnectClient.ControlPlaneRelations.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	res.ServicePackages = filterNonKongPackages(config.ControlPlaneID, servicePackages, relations)
	return &res, nil
}

func filterNonKongPackages(controlPlaneID string, packages []*konnect.ServicePackage,
	relations []*konnect.ControlPlaneServiceRelation) []*konnect.ServicePackage {
	kongServiceIDs := kongServiceIDs(controlPlaneID, relations)
	var res []*konnect.ServicePackage
	for _, p := range packages {
		// if a package has no versions, decK will manage it
		switch len(p.Versions) {
		case 0:
			res = append(res, p)
		default:
			// decK will manage two types of versions:
			// - either versions that don't have any implementation
			// - versions which have a Kong Service as an implementation
			pCopy := p.DeepCopy()
			pCopy.Versions = nil
			for _, v := range p.Versions {
				if v.ControlPlaneServiceRelation == nil {
					pCopy.Versions = append(pCopy.Versions, v)
				} else if !utils.Empty(v.ControlPlaneServiceRelation.ControlPlaneEntityID) &&
					kongServiceIDs[*v.ControlPlaneServiceRelation.ControlPlaneEntityID] {
					pCopy.Versions = append(pCopy.Versions, v)
				}
			}
			// manage only if at least one version satisfies the above criteria
			if len(pCopy.Versions) >= 1 {
				res = append(res, pCopy)
			}
		}
	}
	return res
}

func kongServiceIDs(cpID string,
	relations []*konnect.ControlPlaneServiceRelation) map[string]bool {
	res := map[string]bool{}
	for _, relation := range relations {
		if !utils.Empty(relation.ControlPlaneEntityID) &&
			relation.ControlPlane != nil &&
			!utils.Empty(relation.ControlPlane.ID) &&
			cpID == *relation.ControlPlane.ID {
			res[*relation.ControlPlaneEntityID] = true
		}
	}
	return res
}
