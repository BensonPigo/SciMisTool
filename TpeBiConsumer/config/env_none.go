//go:build !dev && !prod
// +build !dev,!prod

package config

// This file exists to force the build to fail when no build tag is provided.
// ConfigFilePath is defined here so that references compile, but the build will
// still fail because of the undefined identifier below. This makes the error
// message clearer when no build tag is specified.
const ConfigFilePath = ""

// The undefined identifier below will cause a compilation error when neither
// 'dev' nor 'prod' build tags are supplied.
var _ = "ThisProjectRequiresADevOrProdBuildTag"
