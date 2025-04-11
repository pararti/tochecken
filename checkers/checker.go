package checkers

import "tochecken/models"

type Checker interface {
	FetchTokens(ru *RubUsd, firstStart bool) error
	GetNews() []*models.CommonModel
	GetTokens() map[string]*models.CommonModel
	GetType() string
}
