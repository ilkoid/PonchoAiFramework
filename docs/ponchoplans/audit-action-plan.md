# PonchoAiFramework Architecture Audit Action Plan

## Overview
This document outlines the action plan to address the architectural violations identified in the PonchoAiFramework audit. The primary issue is the violation of the Dependency Rule where the Core layer imports implementation packages.

## Executive Summary

`★ Insight ─────────────────────────────────────`
1. **Critical Issues**: Core package directly imports model and tool implementations, breaking clean architecture
2. **Easy Fixes**: Interfaces are well-designed and don't need changes - just move the factories
3. **Quick Impact**: These changes will immediately improve testability and maintainability
`─────────────────────────────────────────────────`

## 🚨 Critical Violations to Fix

### 1. Core Package Importing Implementation
- **Files affected**:
  - `core/config/model_factory.go` (imports `models/deepseek` and `models/zai`)
  - `core/config/s3_tool_factory.go` (imports `tools/article_importer`)
- **Impact**: Breaks Dependency Inversion Principle
- **Priority**: Critical

### 2. Factory Pattern Misplacement
- **Issue**: Factories are in core package instead of implementation layer
- **Files to move**:
  - `core/config/model_factory.go` → `factories/models/`
  - `core/config/s3_tool_factory.go` → `factories/tools/`
- **Priority**: Critical

## 📋 Action Plan Phases

### Phase 1: Immediate Fixes (Week 1)

#### 1.1 Create Factory Package Structure
```bash
mkdir -p factories/models
mkdir -p factories/tools
```

#### 1.2 Move Model Factory
**Action**: Move `core/config/model_factory.go` to `factories/models/`

**Steps**:
1. Copy file to new location
2. Update import paths
3. Create interface definition in core
4. Update all references

**Files to create**:
- `factories/models/model_factory.go` (implementation)
- `core/interfaces/factory.go` (interface only)

#### 1.3 Move Tool Factory
**Action**: Move `core/config/s3_tool_factory.go` to `factories/tools/`

**Steps**:
1. Copy file to new location
2. Update import paths
3. Update all references

#### 1.4 Update Core to Use Factory Interface
**File**: `core/config/validator.go` (if exists) or create new core factory manager

**Changes**:
- Remove direct imports of implementations
- Use factory interfaces
- Inject dependencies

### Phase 2: Implement Service Locator (Week 2)

#### 2.1 Create Service Locator Interface
**File**: `core/service_locator.go`

```go
type ServiceLocator interface {
    GetModelFactory(provider string) (ModelFactory, error)
    GetToolFactory(name string) (ToolFactory, error)
    RegisterModelFactory(provider string, factory ModelFactory)
    RegisterToolFactory(name string, factory ToolFactory)
}
```

#### 2.2 Update Framework Initialization
**File**: `core/framework.go`

**Changes**:
- Add ServiceLocator field
- Initialize factories during startup
- Remove hardcoded factory knowledge

#### 2.3 Create Factory Registration Mechanism
**New file**: `core/registry/factory_registry.go`

**Purpose**: Dynamic factory registration based on configuration

### Phase 3: Testing and Validation (Week 3)

#### 3.1 Unit Tests
- Test factory implementations
- Test service locator
- Test dependency injection

#### 3.2 Integration Tests
- Verify models can be created through new architecture
- Verify tools can be created through new architecture
- Test configuration loading

#### 3.3 Validation
- Run existing tests
- Verify no regression
- Check dependency graph (use `go mod graph`)

## 📁 New Package Structure After Refactoring

```
/interfaces/              # Core interfaces (no changes)
├── types.go             # Data structures
├── interfaces.go        # Component interfaces
└── factory.go           # NEW: Factory interfaces

/core/                   # Business logic
├── base/                # Base classes
├── config/              # Configuration (minus factories)
├── registry/            # Component registries
├── service_locator.go   # NEW: Service locator interface
└── framework.go         # Main framework (updated)

/factories/              # NEW: Factory implementations
├── models/              # Model factories
│   ├── factory.go       # Model factory interface impl
│   ├── deepseek_factory.go
│   └── zai_factory.go
├── tools/               # Tool factories
│   ├── factory.go
│   └── s3_factory.go
└── registry.go          # Factory registry implementation

/models/                 # Model implementations (no changes)
├── deepseek/
└── zai/

/tools/                  # Tool implementations (no changes)
└── s3/
```

## 🔄 Dependency Flow After Fix

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Commands   │───▶│    Core     │───▶│ Interfaces  │
│             │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
                          │                   ▲
                          │                   │
                          ▼                   │
                   ┌─────────────┐           │
                   │ Service     │───────────┘
                   │ Locator     │
                   └─────────────┘
                          │
                          ▼
                   ┌─────────────┐
                   │  Factories  │
                   │  (Models &  │
                   │   Tools)    │
                   └─────────────┘
```

## ✅ Validation Checklist

After completing all phases, verify:

- [ ] Core package has NO imports from `models/` or `tools/`
- [ ] All factories are in `factories/` package
- [ ] Service locator pattern is implemented
- [ ] Dependency injection is used throughout
- [ ] All existing tests pass
- [ ] New tests cover factory pattern
- [ ] Documentation is updated

## 📝 Migration Commands

### 1. Move files
```bash
# Create directories
mkdir -p factories/models
mkdir -p factories/tools

# Move model factory
mv core/config/model_factory.go factories/models/

# Move tool factory
mv core/config/s3_tool_factory.go factories/tools/
```

### 2. Update imports
```bash
# Find all files that need import updates
grep -r "core/config/model_factory" .
grep -r "core/config/s3_tool_factory" .
```

### 3. Verify dependencies
```bash
# Check core package imports
go mod graph | grep "PonchoAiFramework/core" | grep -E "(models|tools)"

# Should return no results
```

## 📊 Expected Benefits

1. **Improved Testability**: Core can be tested with mock factories
2. **Better Maintainability**: Changes to implementations don't affect core
3. **Cleaner Architecture**: Proper dependency direction
4. **Enhanced Extensibility**: New models/tools can be added easily
5. **Reduced Coupling**: Core doesn't know about specific implementations

## ⚠️ Risks and Mitigations

### Risk 1: Breaking Changes
**Mitigation**:
- Maintain backward compatibility in interfaces
- Use feature flags if needed
- Thorough testing

### Risk 2: Performance Overhead
**Mitigation**:
- Use factory caching
- Lazy initialization
- Profile before and after

### Risk 3: Complexity Increase
**Mitigation**:
- Clear documentation
- Good examples
- Keep interfaces simple

## 📅 Timeline

| Week | Tasks |
|------|-------|
| 1 | Phase 1: Move factories and fix immediate violations |
| 2 | Phase 2: Implement service locator pattern |
| 3 | Phase 3: Testing, validation, and documentation |

## 🎯 Success Criteria

1. Zero dependency rule violations
2. All tests passing
3. No regression in functionality
4. Improved test coverage (>95% for core components)
5. Documentation updated

## 🔧 Quick Reference Commands

```bash
# Check for violations
go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./core/... | grep -E "(models|tools)"

# Run tests
go test ./...

# Build all
go build ./...

# Verify package structure
tree -d -L 2 .
```

---

## Next Steps

1. **Review this plan** with the team
2. **Create branches** for each phase
3. **Assign owners** for each task
4. **Schedule review meetings**
5. **Track progress** using project management tools

Remember: The key is to maintain the Dependency Rule - dependencies must point inwards, from outer layers to inner layers only.