package passBuildingService

import (
	"context"
)

type PassBuildingService struct {
	passEditor PassEditor
}

type PassEditor interface {
	EditPassBuildingComment(
		ctx context.Context,
		passID string,
		buildingID string,
		newComment string,
	) error

	DeletePassBuilding(
		ctx context.Context,
		passID string,
		buildingID string,
	) error
}

func New() {

}

func (s *PassBuildingService) Edit(
	ctx context.Context,
	passID string,
	buildingID string,
	newComment string,
) error {
	if err := s.passEditor.EditPassBuildingComment(
		ctx,
		passID,
		buildingID,
		newComment,
	); err != nil {
		return err
	}

	return nil
}

func (s *PassBuildingService) Delete(
	ctx context.Context,
	passID string,
	buildingID string,
) error {
	if err := s.passEditor.DeletePassBuilding(
		ctx,
		passID,
		buildingID,
	); err != nil {
		return err
	}

	return nil
}
