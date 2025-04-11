package app

import (
	"tochecken/checkers"
	"tochecken/db"
)

type App struct {
	DB           *db.Db
	CheckersPoll map[string]checkers.Checker
}
