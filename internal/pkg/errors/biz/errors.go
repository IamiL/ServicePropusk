package bizErrors

import "errors"

var (
	ErrorNoPermission = errors.New("There is no permission for this.")

	ErrorBuildingNotFound = errors.New("There is no such building.")

	ErrorBuildingsNotFound = errors.New("buildings not found")

	ErrorInternalServer = errors.New("There is an internal server error.")

	ErrorAuthToken = errors.New("token authentication error")

	ErrorInvalidBuilding = errors.New("invalid building")

	ErrorInvalidPass = errors.New("invalid pass")

	ErrorStatusNotFormed = errors.New("пропуск не сформирован создателем")

	ErrorCannotBeDeleted = errors.New("cannot be deleted")

	ErrorCannotBeEditing = errors.New("cannot be editing")

	ErrorPassesNotFound = errors.New("passes not found")

	ErrorStatusNotDraft = errors.New("пропуск не в статусе черновик")

	ErrorBuildingAlreadyAdded = errors.New("The building has already been added to the pass")

	ErrorCannotBeFormed = errors.New("пропуск не может быть сформирован")

	ErrorInvalidPassBuilding = errors.New("такого корпуса в пропуске нет")

	ErrorPassIsNotDraft = errors.New("пропуск не в статусе черновика")

	ErrorShortPassword = errors.New("password must be at least 8 characters")

	ErrorUserAlreadyExists = errors.New("the user already exists")

	ErrorPassNotFound = errors.New("pass not found")
)
