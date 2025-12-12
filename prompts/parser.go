package prompts

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ilkoid/PonchoAiFramework/interfaces"
)

// V1PromptData represents the minimal data structure for version 1 prompt format
type V1PromptData struct {
	Config  string                    `json:"config"`
	System  string                    `json:"system"`
	User    string                    `json:"user"`
	Media   map[string]string         `json:"media"`   // media variables like { "photoUrl": "url" }
	Variables map[string]interface{}   `json:"variables"` // extracted variables
}

// V1Parser is a minimal parser for version 1 prompt format
type V1Parser struct{}

// NewV1Parser creates a new V1 parser instance
func NewV1Parser() *V1Parser {
	return &V1Parser{}
}

// V1Integration provides integration for V1 prompt format with existing prompt system
type V1Integration struct {
	v1Parser *V1Parser
	logger   interfaces.Logger
}

// NewV1Integration creates a new V1 integration instance
func NewV1Integration(logger interfaces.Logger) *V1Integration {
	return &V1Integration{
		v1Parser: NewV1Parser(),
		logger:   logger,
	}
}

// ParseAndConvert parses V1 format content and converts to standard PromptTemplate
func (v1i *V1Integration) ParseAndConvert(content string, name string) (*interfaces.PromptTemplate, error) {
	if v1i.logger != nil {
		v1i.logger.Debug("Parsing V1 format and converting to PromptTemplate", "name", name)
	}
	
	// Validate format first
	if err := v1i.v1Parser.ValidateFormat(content); err != nil {
		return nil, fmt.Errorf("invalid V1 format: %w", err)
	}
	
	// Parse using V1 parser
	v1Data, err := v1i.v1Parser.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse V1 content: %w", err)
	}
	
	// Convert to standard PromptTemplate
	template := v1i.v1Parser.ToPromptTemplate(v1Data, name)
	
	if v1i.logger != nil {
		v1i.logger.Debug("V1 content parsed and converted successfully",
			"name", template.Name,
			"parts", len(template.Parts),
			"variables", len(template.Variables))
	}
	
	return template, nil
}

// IsV1Format checks if content is in V1 format
func (v1i *V1Integration) IsV1Format(content string) bool {
	// Check for V1 format indicators
	return strings.Contains(content, "{{role") &&
		(strings.Contains(content, "{{role \"config\"") ||
		 strings.Contains(content, "{{role \"system\"") ||
		 strings.Contains(content, "{{role \"user\""))
}

// GenerateTemplateName generates template name from V1 content
func (v1i *V1Integration) GenerateTemplateName(content string) string {
	// Look for system content to generate meaningful name
	if strings.Contains(content, "sketch") || strings.Contains(content, "fashion") {
		return "sketch_description"
	} else if strings.Contains(content, "analyzer") {
		return "analyzer"
	} else if strings.Contains(content, "description") {
		return "description"
	}
	
	return "v1_template"
}

// Parse parses version 1 prompt format with {{role "...}} and {{media url=...}} syntax
func (p *V1Parser) Parse(content string) (*V1PromptData, error) {
	result := &V1PromptData{
		Media:     make(map[string]string),
		Variables: make(map[string]interface{}),
	}

	// Regular expressions for parsing
	roleRegex := regexp.MustCompile(`\{\{role\s+"([^"]+)"\}\}`)

	// Find all role delimiters with their positions
	matches := roleRegex.FindAllStringSubmatch(content, -1)
	
	// Process content by finding sections between role delimiters
	for i, match := range matches {
		if len(match) < 2 {
			continue
		}
		
		role := match[1]
		
		// Find the content between this role and the next role
		startPos := strings.Index(content, match[0]) + len(match[0])
		endPos := len(content)
		
		if i+1 < len(matches) {
			endPos = strings.Index(content, matches[i+1][0])
		}
		
		sectionContent := strings.TrimSpace(content[startPos:endPos])
		
		// Process content based on role
		switch role {
		case "config":
			result.Config = sectionContent
		case "system":
			result.System = sectionContent
		case "user":
			// Extract media variables from user content
			processedContent, mediaVars := p.extractMediaVars(sectionContent)
			result.User = processedContent
			for k, v := range mediaVars {
				result.Media[k] = v
				result.Variables[k] = v
			}
		}
	}

	return result, nil
}

// extractMediaVars extracts {{media url=variable}} patterns and returns processed content and variables
func (p *V1Parser) extractMediaVars(content string) (string, map[string]string) {
	mediaVars := make(map[string]string)
	
	// Find all media references
	mediaRegex := regexp.MustCompile(`\{\{media\s+url=([^}]+)\}\}`)
	matches := mediaRegex.FindAllStringSubmatch(content, -1)
	
	processedContent := content
	for _, match := range matches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])
			mediaVars[varName] = varName // Store variable name as placeholder
		}
	}
	
	return processedContent, mediaVars
}

// ToPromptTemplate converts V1PromptData to standard PromptTemplate
func (p *V1Parser) ToPromptTemplate(data *V1PromptData, name string) *interfaces.PromptTemplate {
	template := &interfaces.PromptTemplate{
		Name:        name,
		Description:  fmt.Sprintf("Version 1 prompt: %s", name),
		Version:     "1.0",
		Category:    "v1",
		Tags:        []string{"v1", "legacy"},
		Parts:       make([]*interfaces.PromptPart, 0),
		Variables:   make([]*interfaces.PromptVariable, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    &interfaces.PromptMetadata{},
	}

	// Add config part if exists
	if data.Config != "" {
		// Parse config as YAML for model settings
		configPart := &interfaces.PromptPart{
			Type:    interfaces.PromptPartTypeSystem,
			Content: data.Config,
		}
		template.Parts = append(template.Parts, configPart)
	}

	// Add system part if exists
	if data.System != "" {
		systemPart := &interfaces.PromptPart{
			Type:    interfaces.PromptPartTypeSystem,
			Content: data.System,
		}
		template.Parts = append(template.Parts, systemPart)
	}

	// Add user part if exists
	if data.User != "" {
		userPart := &interfaces.PromptPart{
			Type:    interfaces.PromptPartTypeUser,
			Content: data.User,
		}
		template.Parts = append(template.Parts, userPart)
	}

	// Add media parts and variables
	for varName, varValue := range data.Media {
		// Add media part
		mediaPart := &interfaces.PromptPart{
			Type:    interfaces.PromptPartTypeMedia,
			Content: varValue,
			Media: &interfaces.MediaPart{
				URL: varValue,
			},
		}
		template.Parts = append(template.Parts, mediaPart)

		// Add variable definition
		variable := &interfaces.PromptVariable{
			Name:         varName,
			Type:         "string",
			Description:  fmt.Sprintf("Media variable: %s", varName),
			Required:     true,
			DefaultValue: varValue,
		}
		template.Variables = append(template.Variables, variable)
	}

	return template
}

// ValidateFormat checks if content matches version 1 format
func (p *V1Parser) ValidateFormat(content string) error {
	// Check for required role patterns
	roleRegex := regexp.MustCompile(`\{\{role\s+"[^"]+"\}\}`)
	matches := roleRegex.FindAllString(content, -1)
	
	if len(matches) == 0 {
		return fmt.Errorf("no role delimiters found in content")
	}

	// Check for valid role types
	validRoles := map[string]bool{
		"config": true,
		"system": true,
		"user":   true,
	}

	for _, match := range matches {
		roleRegex := regexp.MustCompile(`\{\{role\s+"([^"]+)"\}\}`)
		roleMatch := roleRegex.FindStringSubmatch(match)
		if len(roleMatch) > 1 {
			role := roleMatch[1]
			if !validRoles[role] {
				return fmt.Errorf("invalid role type: %s", role)
			}
		}
	}

	return nil
}

// PromptTemplateLoaderImpl implements PromptTemplateLoader interface
type PromptTemplateLoaderImpl struct {
	config *PromptConfig
	logger interfaces.Logger
}

// NewPromptTemplateLoader creates a new PromptTemplateLoader instance
func NewPromptTemplateLoader(config *PromptConfig, logger interfaces.Logger) interfaces.PromptTemplateLoader {
	return &PromptTemplateLoaderImpl{
		config: config,
		logger: logger,
	}
}

// LoadFromFile loads template from file
func (ptl *PromptTemplateLoaderImpl) LoadFromFile(filePath string) (*interfaces.PromptTemplate, error) {
	ptl.logger.Debug("Loading template from file", "path", filePath)

	// Read file content
	content, err := readFileContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse template
	template, err := ptl.parseTemplate(content, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Set file metadata
	if template.Metadata == nil {
		template.Metadata = &interfaces.PromptMetadata{}
	}
	
	// Extract metadata from file
	if err := ptl.extractFileMetadata(template, filePath, content); err != nil {
		ptl.logger.Warn("Failed to extract file metadata", "path", filePath, "error", err)
	}

	ptl.logger.Debug("Template loaded successfully", "name", template.Name, "path", filePath)
	return template, nil
}

// LoadFromDirectory loads all templates from directory
func (ptl *PromptTemplateLoaderImpl) LoadFromDirectory(dirPath string) (map[string]*interfaces.PromptTemplate, error) {
	ptl.logger.Debug("Loading templates from directory", "path", dirPath)

	templates := make(map[string]*interfaces.PromptTemplate)
	
	// Get all template files
	files, err := listTemplateFiles(dirPath, ptl.config.Templates.Extensions)
	if err != nil {
		return nil, fmt.Errorf("failed to list template files: %w", err)
	}

	// Load each template
	for _, filePath := range files {
		template, err := ptl.LoadFromFile(filePath)
		if err != nil {
			ptl.logger.Error("Failed to load template file", "path", filePath, "error", err)
			continue
		}

		// Use filename (without extension) as template name
		name := getTemplateName(filePath)
		templates[name] = template
	}

	ptl.logger.Info("Templates loaded from directory",
		"path", dirPath,
		"count", len(templates))

	return templates, nil
}

// SaveToFile saves template to file
func (ptl *PromptTemplateLoaderImpl) SaveToFile(template *interfaces.PromptTemplate, filePath string) error {
	ptl.logger.Debug("Saving template to file", "name", template.Name, "path", filePath)

	// Serialize template to YAML
	data, err := marshalTemplate(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	// Write to file
	if err := writeFileContent(filePath, data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	ptl.logger.Debug("Template saved successfully", "name", template.Name, "path", filePath)
	return nil
}

// ValidateFile validates template file format
func (ptl *PromptTemplateLoaderImpl) ValidateFile(filePath string) error {
	ptl.logger.Debug("Validating template file", "path", filePath)

	// Try to load template
	_, err := ptl.LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("template file validation failed: %w", err)
	}

	ptl.logger.Debug("Template file validation passed", "path", filePath)
	return nil
}

// parseTemplate parses template content
func (ptl *PromptTemplateLoaderImpl) parseTemplate(content, filePath string) (*interfaces.PromptTemplate, error) {
	v1Integration := NewV1Integration(ptl.logger)
	
	// Check if this is V1 format
	if v1Integration.IsV1Format(content) {
		templateName := v1Integration.GenerateTemplateName(content)
		return v1Integration.ParseAndConvert(content, templateName)
	}
	
	// Fall back to basic parsing for non-V1 content
	return ptl.parseBasicTemplate(content)
}

// parseBasicTemplate parses simple template content
func (ptl *PromptTemplateLoaderImpl) parseBasicTemplate(content string) (*interfaces.PromptTemplate, error) {
	template := &interfaces.PromptTemplate{
		Name:        "basic_template",
		Description: "Basic prompt template",
		Version:     "1.0",
		Category:    "basic",
		Tags:        []string{"basic"},
		Parts:       make([]*interfaces.PromptPart, 0),
		Variables:   make([]*interfaces.PromptVariable, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    &interfaces.PromptMetadata{},
	}

	// Simple parsing - treat entire content as user message
	userPart := &interfaces.PromptPart{
		Type:    interfaces.PromptPartTypeUser,
		Content: strings.TrimSpace(content),
	}
	template.Parts = append(template.Parts, userPart)

	return template, nil
}

// extractFileMetadata extracts metadata from file
func (ptl *PromptTemplateLoaderImpl) extractFileMetadata(template *interfaces.PromptTemplate, filePath, content string) error {
	// Set file path in metadata
	if template.Metadata != nil {
		ptl.logger.Debug("Template file metadata", "path", filePath, "size", len(content))
	}

	return nil
}

// Helper functions

// readFileContent reads file content
func readFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(data), nil
}

// writeFileContent writes content to file
func writeFileContent(filePath string, data []byte) error {
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

// marshalTemplate marshals template to YAML
func marshalTemplate(template *interfaces.PromptTemplate) ([]byte, error) {
	// Simple YAML marshaling
	return []byte(fmt.Sprintf("# Template: %s\n%s", template.Name, template.Parts[0].Content)), nil
}

// listTemplateFiles lists all template files in directory
func listTemplateFiles(dirPath string, extensions []string) ([]string, error) {
	var files []string
	
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Check if file extension matches allowed extensions
		ext := strings.ToLower(filepath.Ext(path))
		for _, allowedExt := range extensions {
			if ext == allowedExt {
				files = append(files, path)
				break
			}
		}
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}
	
	return files, nil
}

// getTemplateName extracts template name from file path
func getTemplateName(filePath string) string {
	// Extract filename without extension
	filename := filepath.Base(filePath)
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}