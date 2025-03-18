package graph

import (
	buildingService "rip/internal/service/building"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	buildingService *buildingService.BuildingService
}

func NewResolver(buildingService *buildingService.BuildingService) *Resolver {
	return &Resolver{
		buildingService: buildingService,
	}
}
