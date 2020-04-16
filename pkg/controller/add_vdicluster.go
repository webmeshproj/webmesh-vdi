package controller

import (
	"github.com/tinyzimmer/kvdi/pkg/controller/vdicluster"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, vdicluster.Add)
}
