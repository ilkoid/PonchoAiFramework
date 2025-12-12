package core

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
	Logger               = interfaces.Logger
	Validator            = interfaces.Validator
	ConfigLoader         = interfaces.ConfigLoader
	MetricsCollector     = interfaces.MetricsCollector
)
