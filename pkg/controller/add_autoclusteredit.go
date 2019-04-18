package controller

import (
	"github.com/barpilot/rancher-operator/pkg/controller/autoclusteredit"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, autoclusteredit.Add)
}
