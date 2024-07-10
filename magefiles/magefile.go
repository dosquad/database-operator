//go:build mage

package main

import (
	"context"

	"github.com/magefile/mage/mg"

	//mage:import
	"github.com/dosquad/mage"
)

// TestLocal update, protoc, format, tidy, lint & test.
func TestLocal(ctx context.Context) {
	mg.CtxDeps(ctx, mage.Kubebuilder.Manifests)
	mg.CtxDeps(ctx, mage.Kubebuilder.Generate)
	mg.CtxDeps(ctx, mage.Golang.Fmt)
	mg.CtxDeps(ctx, mage.Golang.Vet)
	mg.CtxDeps(ctx, mage.Test)
}

var Default = TestLocal
