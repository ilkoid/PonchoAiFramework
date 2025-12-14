package core

// This file contains interface aliases for the main PonchoFramework interfaces.
// It provides convenient access to core interfaces from the interfaces package.
// It includes aliases for PonchoFramework, PonchoModel, PonchoTool, and PonchoFlow.
// It serves as a bridge between the core implementation and interface definitions.
// It simplifies imports and access to framework interfaces.
// It maintains consistency with the overall architecture.
//
// These aliases allow the core package to reference framework interfaces
// without circular dependencies, providing clean separation between
// implementation details and interface definitions.
//
// The aliases include:
// - Core component interfaces (Model, Tool, Flow, Framework)
// - Registry interfaces for component management
// - Supporting interfaces (Logger, Validator, ConfigLoader, MetricsCollector)

import (
	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// Type aliases for interfaces from interfaces package
type (
	PonchoModel          = interfaces.PonchoModel
	PonchoTool           = interfaces.PonchoTool
	PonchoFlow           = interfaces.PonchoFlow
	PonchoStreamCallback = interfaces.PonchoStreamCallback
	PonchoModelRegistry  = interfaces.PonchoModelRegistry
	PonchoToolRegistry   = interfaces.PonchoToolRegistry
	PonchoFlowRegistry   = interfaces.PonchoFlowRegistry
	PonchoFramework      = interfaces.PonchoFramework
	Validator            = interfaces.Validator
	ConfigLoader         = interfaces.ConfigLoader
	MetricsCollector     = interfaces.MetricsCollector
)
