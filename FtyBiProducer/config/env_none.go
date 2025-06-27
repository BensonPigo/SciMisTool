//go:build !dev && !prod
// +build !dev,!prod

package config

// This file exists to force the build to fail when no build tag is provided.
// The undefined identifier below will cause a compilation error.
var _ = ThisProjectRequiresADevOrProdBuildTag
