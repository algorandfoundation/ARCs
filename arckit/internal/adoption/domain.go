package adoption

import "slices"

const (
	CategoryWallets        = "wallets"
	CategoryExplorers      = "explorers"
	CategoryTooling        = "tooling"
	CategoryInfra          = "infra"
	CategoryDappsProtocols = "dapps-protocols"

	ActorStatusPlanned    = "planned"
	ActorStatusInProgress = "in_progress"
	ActorStatusShipped    = "shipped"
	ActorStatusDeclined   = "declined"
	ActorStatusUnknown    = "unknown"

	ReadinessLow    = "low"
	ReadinessMedium = "medium"
	ReadinessHigh   = "high"

	ReferenceImplementationStatusPlanned  = "planned"
	ReferenceImplementationStatusWIP      = "wip"
	ReferenceImplementationStatusShipped  = "shipped"
	ReferenceImplementationStatusArchived = "archived"
)

var (
	categoryNames = []string{
		CategoryWallets,
		CategoryExplorers,
		CategoryTooling,
		CategoryInfra,
		CategoryDappsProtocols,
	}
	actorStatuses = []string{
		ActorStatusPlanned,
		ActorStatusInProgress,
		ActorStatusShipped,
		ActorStatusDeclined,
		ActorStatusUnknown,
	}
	readinessLevels = []string{
		ReadinessLow,
		ReadinessMedium,
		ReadinessHigh,
	}
	referenceImplementationStatuses = []string{
		ReferenceImplementationStatusPlanned,
		ReferenceImplementationStatusWIP,
		ReferenceImplementationStatusShipped,
		ReferenceImplementationStatusArchived,
	}
)

type ActorCategory struct {
	Name   string
	Actors []Actor
}

type RegistryCategory struct {
	Name    string
	Entries []string
}

func CategoryNames() []string {
	return slices.Clone(categoryNames)
}

func ActorStatuses() []string {
	return slices.Clone(actorStatuses)
}

func ReadinessLevels() []string {
	return slices.Clone(readinessLevels)
}

func ReferenceImplementationStatuses() []string {
	return slices.Clone(referenceImplementationStatuses)
}

func IsKnownCategory(name string) bool {
	return slices.Contains(categoryNames, name)
}

func IsKnownActorStatus(status string) bool {
	return slices.Contains(actorStatuses, status)
}

func IsKnownReadiness(readiness string) bool {
	return slices.Contains(readinessLevels, readiness)
}

func IsKnownReferenceImplementationStatus(status string) bool {
	return slices.Contains(referenceImplementationStatuses, status)
}

func (summary *Summary) ActorCategories() []ActorCategory {
	if summary == nil {
		return nil
	}
	return []ActorCategory{
		{Name: CategoryWallets, Actors: summary.Adoption.Wallets},
		{Name: CategoryExplorers, Actors: summary.Adoption.Explorers},
		{Name: CategoryTooling, Actors: summary.Adoption.Tooling},
		{Name: CategoryInfra, Actors: summary.Adoption.Infra},
		{Name: CategoryDappsProtocols, Actors: summary.Adoption.DappsProtocols},
	}
}

func (summary *Summary) Actors(category string) []Actor {
	if summary == nil {
		return nil
	}
	for _, group := range summary.ActorCategories() {
		if group.Name == category {
			return group.Actors
		}
	}
	return nil
}

func (registry *VettedAdopters) Categories() []RegistryCategory {
	if registry == nil {
		return nil
	}
	return []RegistryCategory{
		{Name: CategoryWallets, Entries: registry.Wallets},
		{Name: CategoryExplorers, Entries: registry.Explorers},
		{Name: CategoryTooling, Entries: registry.Tooling},
		{Name: CategoryInfra, Entries: registry.Infra},
		{Name: CategoryDappsProtocols, Entries: registry.DappsProtocols},
	}
}

func (registry *VettedAdopters) Entries(category string) []string {
	if registry == nil {
		return nil
	}
	for _, entry := range registry.Categories() {
		if entry.Name == category {
			return entry.Entries
		}
	}
	return nil
}
