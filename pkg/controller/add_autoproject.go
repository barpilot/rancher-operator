package controller

import (
	"github.com/barpilot/rancher-operator/pkg/controller/autoproject"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, autoproject.Add)
}
