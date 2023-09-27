package controller

import (
	"github.com/epmd-edp/cd-pipeline-operator/v2/pkg/controller/cdpipeline"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, cdpipeline.Add)
}
