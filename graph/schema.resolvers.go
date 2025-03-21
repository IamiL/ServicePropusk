package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"context"
	"fmt"
	"service-propusk-backend/graph/generated"
	"service-propusk-backend/graph/model"
)

// CreateBuilding is the resolver for the createBuilding field.
func (r *mutationResolver) CreateBuilding(
	ctx context.Context,
	input model.CreateBuildingInput,
) (*model.Building, error) {
	panic(fmt.Errorf("not implemented: CreateBuilding - createBuilding"))
}

// UpdateBuilding is the resolver for the updateBuilding field.
func (r *mutationResolver) UpdateBuilding(
	ctx context.Context,
	id string,
	input model.UpdateBuildingInput,
) (*model.Building, error) {
	panic(fmt.Errorf("not implemented: UpdateBuilding - updateBuilding"))
}

// DeleteBuilding is the resolver for the deleteBuilding field.
func (r *mutationResolver) DeleteBuilding(ctx context.Context, id string) (
	bool,
	error,
) {
	panic(fmt.Errorf("not implemented: DeleteBuilding - deleteBuilding"))
}

// Building is the resolver for the building field.
func (r *queryResolver) Building(
	ctx context.Context,
	id string,
) (*model.Building, error) {
	building, err := r.buildingService.GetBuilding(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get building: %w", err)
	}
	return &model.Building{
		ID:          building.Id,
		Name:        building.Name,
		Description: building.Description,
		ImgURL:      building.ImgUrl,
	}, nil
}

// Buildings is the resolver for the buildings field.
func (r *queryResolver) Buildings(
	ctx context.Context,
	name *string,
) ([]*model.Building, error) {
	panic(fmt.Errorf("not implemented: Buildings - buildings"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
