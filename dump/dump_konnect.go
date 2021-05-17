package dump

import (
	"context"
	"fmt"
	"sync"

	"github.com/kong/deck/konnect"
	"github.com/kong/deck/utils"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type KonnectConfig struct {
	// ID of the Kong Control Plane being managed.
	ControlPlaneID string
}

func GetFromKonnect(ctx context.Context, konnectClient *konnect.Client,
	config KonnectConfig) (*utils.KonnectRawState, error) {
	var res utils.KonnectRawState
	var servicePackages []*konnect.ServicePackage
	var relations []*konnect.ControlPlaneServiceRelation

	group, ctx := errgroup.WithContext(ctx)
	// group1 fetches service packages and their versions
	group.Go(func() error {
		var err error
		// fetch service packages
		servicePackages, err = konnectClient.ServicePackages.ListAll(ctx)
		if err != nil {
			return err
		}

		// versions of service packages are fetched concurrently
		errChan := make(chan error)
		var err2 error

		m := &sync.Mutex{}
		m.Lock()
		go func() {
			defer m.Unlock()
			// only the last error matters
			for err := range errChan {
				err2 = err
			}
		}()

		semaphore := semaphore.NewWeighted(10)
		for i := 0; i < len(servicePackages); i++ {
			// control the number of outstanding go routines, also controlling
			// the number of parallel requests
			err := semaphore.Acquire(ctx, 2)
			if err != nil {
				return fmt.Errorf("acquire semaphore: %v", err)
			}
			go func(i int) {
				defer semaphore.Release(1)
				versions, err := konnectClient.ServiceVersions.ListForPackage(ctx, servicePackages[i].ID)
				if err != nil {
					errChan <- err
					return
				}
				servicePackages[i].Versions = versions
			}(i)
			go func(i int) {
				defer semaphore.Release(1)
				documents, err := konnectClient.Documents.ListAllForParent(ctx, servicePackages[i])
				if err != nil {
					errChan <- err
					return
				}
				res.Documents = append(res.Documents, documents...)
			}(i)
		}
		for i := 0; i < len(servicePackages); i++ {
			for _, version := range servicePackages[i].Versions {
				err := semaphore.Acquire(ctx, 1)
				if err != nil {
					return fmt.Errorf("acquire semaphore: %v", err)
				}
				go func(version konnect.ServiceVersion) {
					defer semaphore.Release(1)
					documents, err := konnectClient.Documents.ListAllForParent(ctx, &version)
					if err != nil {
						errChan <- err
						return
					}
					res.Documents = append(res.Documents, documents...)
				}(version)
			}
		}
		err = semaphore.Acquire(ctx, 10)
		if err != nil {
			return fmt.Errorf("acquire semaphore: %v", err)
		}
		close(errChan)
		semaphore.Release(10)
		m.Lock()
		defer m.Unlock()
		if err2 != nil {
			return err2
		}
		return nil
	})

	// group2 fetches CP-service relations
	group.Go(func() error {
		var err error
		relations, err = konnectClient.ControlPlaneRelations.ListAll(ctx)
		return err
	})

	err := group.Wait()
	if err != nil {
		return nil, err
	}

	res.ServicePackages = filterNonKongPackages(config.ControlPlaneID,
		servicePackages, relations)
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
