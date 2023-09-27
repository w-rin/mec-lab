package controller

import (
	"github.com/epmd-edp/reconciler/v2/pkg/controller/stage"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, stage.Add)
}
