# Руководство по парсингу и обработке унифицированного формата промптов

## Обзор

Документ содержит детальные рекомендации по реализации парсера и обработчика унифицированного формата промптов PonchoFramework с использованием `{{...}}` синтаксиса.

## Архитектура парсера

### Основные компоненты

```go
type PromptParser struct {
    syntaxTree    *SyntaxTree
    validator     *PromptValidator
    templateEngine *TemplateEngine
    errorHandler  *ErrorHandler
    cache         *ParseCache
}

type SyntaxTree struct {
    Root          *SyntaxNode
    Metadata      *MetadataNode
    Config        *ConfigNode
    Roles         map[string]*RoleNode
    Variables     map[string]*VariableNode
    Media         []*MediaNode
    Helpers       []*HelperNode
}

type SyntaxNode struct {
    Type        NodeType
    Value       string
    Attributes  map[string]interface{}
    Children    []*SyntaxNode
    Position    Position
    Parent      *SyntaxNode
}
```

### Алгоритм парсинга

```go
func (pp *PromptParser) Parse(content string) (*SyntaxTree, error) {
    // 1. Предварительная обработка
    normalized := pp.preprocess(content)
    
    // 2. Лексический анализ
    tokens, err := pp.tokenize(normalized)
    if err != nil {
        return nil, fmt.Errorf("tokenization failed: %w", err)
    }
    
    // 3. Синтаксический анализ
    ast, err := pp.parseTokens(tokens)
    if err != nil {
        return nil, fmt.Errorf("syntax parsing failed: %w", err)
    }
    
    // 4. Семантический анализ
    if err := pp.semanticAnalysis(ast); err != nil {
        return nil, fmt.Errorf("semantic analysis failed: %w", err)
    }
    
    // 5. Постобработка и оптимизация
    optimized := pp.optimizeAST(ast)
    
    return optimized, nil
}
```

## Лексический анализ

### Регулярные выражения для токенизации

```go
type TokenType int

const (
    TokenText TokenType = iota
    TokenMetadataStart
    TokenConfigStart
    TokenRoleStart
    TokenRoleEnd
    TokenMedia
    TokenVariable
    TokenHelperStart
    TokenHelperEnd
    TokenComment
)

type Token struct {
    Type     TokenType
    Value    string
    Position Position
    Length   int
}

var tokenPatterns = []struct {
    pattern *regexp.Regexp
    tokenType TokenType
}{
    {regexp.MustCompile(`\{\{\s*metadata\s+`), TokenMetadataStart},
    {regexp.MustCompile(`\{\{\s*config\s+`), TokenConfigStart},
    {regexp.MustCompile(`\{\{\s*(system|user|model|assistant)\s*\}\}`), TokenRoleStart},
    {regexp.MustCompile(`\{\{\s*\/(system|user|model|assistant)\s*\}\}`), TokenRoleEnd},
    {regexp.MustCompile(`\{\{\s*media\s+`), TokenMedia},
    {regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`), TokenVariable},
    {regexp.MustCompile(`\{\{\s*#(if|unless|each|with)\s+`), TokenHelperStart},
    {regexp.MustCompile(`\{\{\s*\/(if|unless|each|with)\s*\}\}`), TokenHelperEnd},
    {regexp.MustCompile(`\{\{#.*?\}\}`), TokenComment},
}
```

### Лексер

```go
type Lexer struct {
    input    string
    position int
    line     int
    column   int
}

func (l *Lexer) Tokenize() ([]Token, error) {
    var tokens []Token
    
    for l.position < len(l.input) {
        // Пропуск пробелов и комментариев
        l.skipWhitespace()
        
        if l.position >= len(l.input) {
            break
        }
        
        // Поиск совпадений с паттернами
        matched := false
        for _, pattern := range tokenPatterns {
            if matches := pattern.pattern.FindStringSubmatchIndex(l.input[l.position:]); len(matches) > 0 {
                token := Token{
                    Type:     pattern.tokenType,
                    Value:    l.input[l.position+matches[0] : l.position+matches[1]],
                    Position: Position{Line: l.line, Column: l.column},
                    Length:   matches[1] - matches[0],
                }
                
                tokens = append(tokens, token)
                l.advance(matches[1])
                matched = true
                break
            }
        }
        
        if !matched {
            // Обычный текст
            start := l.position
            for l.position < len(l.input) && !l.isTemplateStart() {
                l.advance(1)
            }
            
            tokens = append(tokens, Token{
                Type:     TokenText,
                Value:    l.input[start:l.position],
                Position: Position{Line: l.line, Column: l.column},
                Length:   l.position - start,
            })
        }
    }
    
    return tokens, nil
}
```

## Синтаксический анализ

### Грамматика языка

```bnf
<prompt> ::= <metadata>? <config>? <role_section>* <text>*

<metadata> ::= "{{metadata" <attributes> "}}"

<config> ::= "{{config" <attributes> "}}"

<role_section> ::= <role_start> <content> <role_end>

<role_start> ::= "{{" <role_name> "}}"
<role_end> ::= "{{/" <role_name> "}}"

<media> ::= "{{media" <attributes> "}}"

<variable> ::= "{{" <variable_name> "}}"

<helper> ::= <helper_start> <content> <helper_end>

<attributes> ::= <attribute>*
<attribute> ::= <key> "=" <value>

<content> ::= <text> | <media> | <variable> | <helper>*
```

### Recursive Descent Parser

```go
type Parser struct {
    tokens   []Token
    current  int
    errors   []ParseError
}

func (p *Parser) Parse() (*SyntaxTree, error) {
    tree := &SyntaxTree{
        Root: &SyntaxNode{Type: NodeRoot},
    }
    
    // Парсинг метаданных
    if p.peek().Type == TokenMetadataStart {
        metadata, err := p.parseMetadata()
        if err != nil {
            return nil, err
        }
        tree.Metadata = metadata
        tree.Root.Children = append(tree.Root.Children, metadata)
    }
    
    // Парсинг конфигурации
    if p.peek().Type == TokenConfigStart {
        config, err := p.parseConfig()
        if err != nil {
            return nil, err
        }
        tree.Config = config
        tree.Root.Children = append(tree.Root.Children, config)
    }
    
    // Парсинг ролевых секций
    for p.peek().Type == TokenRoleStart {
        role, err := p.parseRole()
        if err != nil {
            return nil, err
        }
        if tree.Roles == nil {
            tree.Roles = make(map[string]*RoleNode)
        }
        tree.Roles[role.Name] = role
        tree.Root.Children = append(tree.Root.Children, role)
    }
    
    // Парсинг оставшегося контента
    for !p.isAtEnd() {
        content, err := p.parseContent()
        if err != nil {
            return nil, err
        }
        if content != nil {
            tree.Root.Children = append(tree.Root.Children, content)
        }
    }
    
    return tree, nil
}

func (p *Parser) parseMetadata() (*MetadataNode, error) {
    token := p.advance() // TokenMetadataStart
    
    node := &MetadataNode{
        SyntaxNode: SyntaxNode{
            Type:     NodeMetadata,
            Position: token.Position,
        },
        Attributes: make(map[string]interface{}),
    }
    
    // Парсинг атрибутов до закрывающего тега
    for !p.isAtEnd() && !p.isNextTokenText() {
        attrToken := p.peek()
        if attrToken.Type == TokenText {
            break
        }
        
        key, value, err := p.parseAttribute()
        if err != nil {
            return nil, err
        }
        
        node.Attributes[key] = value
    }
    
    return node, nil
}

func (p *Parser) parseAttribute() (string, interface{}, error) {
    // Ожидаем формат: key="value" или key=value
    token := p.advance()
    
    parts := strings.SplitN(token.Value, "=", 2)
    if len(parts) != 2 {
        return "", nil, fmt.Errorf("invalid attribute format: %s", token.Value)
    }
    
    key := strings.TrimSpace(parts[0])
    valueStr := strings.TrimSpace(parts[1])
    
    // Парсинг значения
    value, err := p.parseValue(valueStr)
    if err != nil {
        return "", nil, fmt.Errorf("invalid attribute value %s: %w", valueStr, err)
    }
    
    return key, value, nil
}

func (p *Parser) parseValue(value string) (interface{}, error) {
    // Строковые значения в кавычках
    if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
        return strings.Trim(value, `"`), nil
    }
    
    // Числовые значения
    if num, err := strconv.ParseFloat(value, 64); err == nil {
        return num, nil
    }
    
    // Булевы значения
    if strings.ToLower(value) == "true" {
        return true, nil
    }
    if strings.ToLower(value) == "false" {
        return false, nil
    }
    
    // Массивы (через запятую)
    if strings.Contains(value, ",") {
        items := strings.Split(value, ",")
        result := make([]string, len(items))
        for i, item := range items {
            result[i] = strings.TrimSpace(item)
        }
        return result, nil
    }
    
    // Простая строка без кавычек
    return value, nil
}
```

## Обработка шаблонов

### Template Engine

```go
type TemplateEngine struct {
    functions map[string]TemplateFunction
    cache     *TemplateCache
    logger    Logger
}

type TemplateFunction func(ctx context.Context, args []interface{}) (interface{}, error)

func (te *TemplateEngine) Render(template string, data interface{}) (string, error) {
    // Компиляция шаблона
    compiled, err := te.compile(template)
    if err != nil {
        return "", fmt.Errorf("template compilation failed: %w", err)
    }
    
    // Исполнение шаблона
    result, err := te.execute(compiled, data)
    if err != nil {
        return "", fmt.Errorf("template execution failed: %w", err)
    }
    
    return result, nil
}

func (te *TemplateEngine) registerBuiltinFunctions() {
    te.functions["date"] = te.dateFunction
    te.functions["timestamp"] = te.timestampFunction
    te.functions["uuid"] = te.uuidFunction
    te.functions["join"] = te.joinFunction
    te.functions["fashion_category"] = te.fashionCategoryFunction
    te.functions["wildberries_category"] = te.wildberriesCategoryFunction
    te.functions["russian_case"] = te.russianCaseFunction
    te.functions["pluralize_russian"] = te.pluralizeRussianFunction
}

func (te *TemplateEngine) dateFunction(ctx context.Context, args []interface{}) (interface{}, error) {
    if len(args) == 0 {
        return time.Now().Format("2006-01-02"), nil
    }
    
    format, ok := args[0].(string)
    if !ok {
        return "", fmt.Errorf("date function expects string format argument")
    }
    
    return time.Now().Format(format), nil
}

func (te *TemplateEngine) joinFunction(ctx context.Context, args []interface{}) (interface{}, error) {
    if len(args) < 2 {
        return "", fmt.Errorf("join function requires at least 2 arguments")
    }
    
    separator, ok := args[1].(string)
    if !ok {
        return "", fmt.Errorf("join function second argument must be string separator")
    }
    
    switch v := args[0].(type) {
    case []string:
        return strings.Join(v, separator), nil
    case []interface{}:
        parts := make([]string, len(v))
        for i, item := range v {
            parts[i] = fmt.Sprintf("%v", item)
        }
        return strings.Join(parts, separator), nil
    default:
        return "", fmt.Errorf("join function first argument must be array")
    }
}
```

### Fashion-specific функции

```go
func (te *TemplateEngine) fashionCategoryFunction(ctx context.Context, args []interface{}) (interface{}, error) {
    if len(args) == 0 {
        return "", fmt.Errorf("fashion_category function requires category argument")
    }
    
    categoryID, ok := args[0].(string)
    if !ok {
        return "", fmt.Errorf("fashion_category function expects string category ID")
    }
    
    // Интеграция с базой данных категорий
    category, err := te.categoryService.GetFashionCategory(categoryID)
    if err != nil {
        return "", fmt.Errorf("failed to get fashion category: %w", err)
    }
    
    return category.Name, nil
}

func (te *TemplateEngine) wildberriesCategoryFunction(ctx context.Context, args []interface{}) (interface{}, error) {
    if len(args) == 0 {
        return "", fmt.Errorf("wildberries_category function requires category ID")
    }
    
    categoryID := fmt.Sprintf("%v", args[0])
    
    // API вызов к Wildberries
    wbClient := te.wildberriesService.GetClient()
    category, err := wbClient.GetCategory(ctx, categoryID)
    if err != nil {
        return "", fmt.Errorf("failed to get Wildberries category: %w", err)
    }
    
    return category.Name, nil
}

func (te *TemplateEngine) russianCaseFunction(ctx context.Context, args []interface{}) (interface{}, error) {
    if len(args) < 2 {
        return "", fmt.Errorf("russian_case function requires word and case arguments")
    }
    
    word, ok := args[0].(string)
    if !ok {
        return "", fmt.Errorf("russian_case function first argument must be string")
    }
    
    caseType, ok := args[1].(string)
    if !ok {
        return "", fmt.Errorf("russian_case function second argument must be string")
    }
    
    // Использование библиотеки для морфологии русского языка
    result, err := te.russianMorphology.Decline(word, caseType)
    if err != nil {
        return word, nil // Возвращаем исходное слово при ошибке
    }
    
    return result, nil
}
```

## Обработка ошибок

### Структура ошибок

```go
type ParseError struct {
    Line     int    `json:"line"`
    Column   int    `json:"column"`
    Position int    `json:"position"`
    Message  string `json:"message"`
    Context  string `json:"context"`
    Type     string `json:"type"`
}

type ValidationError struct {
    ParseError
    Rule      string `json:"rule"`
    Expected  string `json:"expected"`
    Actual    string `json:"actual"`
    Severity  string `json:"severity"`
}

type RenderError struct {
    ParseError
    Template string `json:"template"`
    Data     string `json:"data"`
    Function string `json:"function"`
}
```

### Обработчик ошибок

```go
type ErrorHandler struct {
    logger  Logger
    metrics MetricsCollector
}

func (eh *ErrorHandler) HandleParseError(err error, content string) *ParseError {
    var parseErr *ParseError
    if errors.As(err, &parseErr) {
        eh.logger.Error("Parse error", 
            "line", parseErr.Line,
            "column", parseErr.Column,
            "message", parseErr.Message,
        )
        
        eh.metrics.IncrementCounter("prompt_parse_errors", 
            "type", parseErr.Type,
        )
        
        return parseErr
    }
    
    // Создание общей ошибки
    return &ParseError{
        Line:    1,
        Column:  1,
        Message: err.Error(),
        Type:    "general",
    }
}

func (eh *ErrorHandler) HandleValidationError(err error, prompt *SyntaxTree) *ValidationError {
    var validationErr *ValidationError
    if errors.As(err, &validationErr) {
        eh.logger.Warn("Validation error",
            "rule", validationErr.Rule,
            "severity", validationErr.Severity,
            "message", validationErr.Message,
        )
        
        eh.metrics.IncrementCounter("prompt_validation_errors",
            "rule", validationErr.Rule,
            "severity", validationErr.Severity,
        )
        
        return validationErr
    }
    
    return &ValidationError{
        ParseError: ParseError{
            Message: err.Error(),
            Type:    "validation",
        },
        Severity: "error",
    }
}
```

## Кэширование и оптимизация

### Кэш парсинга

```go
type ParseCache struct {
    cache    map[string]*CacheEntry
    mutex    sync.RWMutex
    maxSize  int
    ttl      time.Duration
}

type CacheEntry struct {
    SyntaxTree *SyntaxTree
    Timestamp time.Time
    Hash      string
}

func (pc *ParseCache) Get(content string) (*SyntaxTree, bool) {
    hash := pc.hash(content)
    
    pc.mutex.RLock()
    defer pc.mutex.RUnlock()
    
    entry, exists := pc.cache[hash]
    if !exists || time.Since(entry.Timestamp) > pc.ttl {
        return nil, false
    }
    
    return entry.SyntaxTree, true
}

func (pc *ParseCache) Set(content string, tree *SyntaxTree) {
    hash := pc.hash(content)
    
    pc.mutex.Lock()
    defer pc.mutex.Unlock()
    
    // Очистка при превышении лимита
    if len(pc.cache) >= pc.maxSize {
        pc.evictOldest()
    }
    
    pc.cache[hash] = &CacheEntry{
        SyntaxTree: tree,
        Timestamp: time.Now(),
        Hash:      hash,
    }
}

func (pc *ParseCache) hash(content string) string {
    h := sha256.New()
    h.Write([]byte(content))
    return hex.EncodeToString(h.Sum(nil))
}
```

### Оптимизация AST

```go
func (pp *PromptParser) optimizeAST(tree *SyntaxTree) *SyntaxTree {
    optimizer := &ASTOptimizer{}
    return optimizer.Optimize(tree)
}

type ASTOptimizer struct {
    constantFolding bool
    deadCodeElimination bool
}

func (ao *ASTOptimizer) Optimize(tree *SyntaxTree) *SyntaxTree {
    // 1. Константное сворачивание
    if ao.constantFolding {
        tree = ao.foldConstants(tree)
    }
    
    // 2. Удаление мертвого кода
    if ao.deadCodeElimination {
        tree = ao.eliminateDeadCode(tree)
    }
    
    // 3. Оптимизация ролевых секций
    tree = ao.optimizeRoles(tree)
    
    // 4. Слияние текстовых узлов
    tree = ao.mergeTextNodes(tree)
    
    return tree
}

func (ao *ASTOptimizer) foldConstants(tree *SyntaxTree) *SyntaxTree {
    // Сворачивание константных выражений
    visitor := &ConstantFoldingVisitor{}
    return visitor.Visit(tree)
}

func (ao *ASTOptimizer) eliminateDeadCode(tree *SyntaxTree) *SyntaxTree {
    // Удаление недостижимого кода
    visitor := &DeadCodeEliminationVisitor{}
    return visitor.Visit(tree)
}
```

## Интеграция с PonchoFramework

### Prompt Manager

```go
type PromptManager struct {
    parser         *PromptParser
    validator      *PromptValidator
    templateEngine *TemplateEngine
    cache          *PromptCache
    logger         Logger
}

func (pm *PromptManager) LoadPrompt(name string) (*PromptTemplate, error) {
    // Проверка кэша
    if cached, found := pm.cache.Get(name); found {
        return cached, nil
    }
    
    // Загрузка из файла
    content, err := pm.loadFromFile(name)
    if err != nil {
        return nil, fmt.Errorf("failed to load prompt %s: %w", name, err)
    }
    
    // Парсинг
    syntaxTree, err := pm.parser.Parse(content)
    if err != nil {
        return nil, fmt.Errorf("failed to parse prompt %s: %w", name, err)
    }
    
    // Валидация
    if err := pm.validator.ValidateSyntaxTree(syntaxTree); err != nil {
        return nil, fmt.Errorf("validation failed for prompt %s: %w", name, err)
    }
    
    // Создание шаблона
    template := &PromptTemplate{
        Name:       name,
        Content:    content,
        SyntaxTree: syntaxTree,
        CompiledAt: time.Now(),
    }
    
    // Кэширование
    pm.cache.Set(name, template)
    
    return template, nil
}

func (pm *PromptManager) ExecutePrompt(ctx context.Context, name string, data map[string]interface{}) (*PromptResult, error) {
    template, err := pm.LoadPrompt(name)
    if err != nil {
        return nil, err
    }
    
    // Рендеринг шаблона
    rendered, err := pm.templateEngine.Render(template.Content, data)
    if err != nil {
        return nil, fmt.Errorf("template rendering failed: %w", err)
    }
    
    // Построение запроса для модели
    request, err := pm.buildModelRequest(template, rendered, data)
    if err != nil {
        return nil, fmt.Errorf("failed to build model request: %w", err)
    }
    
    // Выполнение запроса
    response, err := pm.modelClient.Generate(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("model generation failed: %w", err)
    }
    
    // Постобработка ответа
    result, err := pm.postProcessResponse(template, response)
    if err != nil {
        return nil, fmt.Errorf("post-processing failed: %w", err)
    }
    
    return result, nil
}
```

## Метрики и мониторинг

### Сбор метрик

```go
type PromptMetrics struct {
    ParseDuration    prometheus.HistogramVec
    ValidationErrors prometheus.CounterVec
    RenderDuration   prometheus.HistogramVec
    CacheHitRate     prometheus.GaugeVec
    TemplateSize     prometheus.HistogramVec
}

func NewPromptMetrics() *PromptMetrics {
    return &PromptMetrics{
        ParseDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "prompt_parse_duration_seconds",
                Help: "Time spent parsing prompts",
            },
            []string{"status"},
        ),
        ValidationErrors: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "prompt_validation_errors_total",
                Help: "Total number of validation errors",
            },
            []string{"rule", "severity"},
        ),
        RenderDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "prompt_render_duration_seconds", 
                Help: "Time spent rendering templates",
            },
            []string{"template"},
        ),
        CacheHitRate: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "prompt_cache_hit_rate",
                Help: "Cache hit rate for prompts",
            },
            []string{"type"},
        ),
        TemplateSize: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "prompt_template_size_bytes",
                Help: "Size of prompt templates",
            },
            []string{"template"},
        ),
    }
}
```

Это руководство обеспечивает полную реализацию парсера и обработчика для унифицированного формата промптов PonchoFramework с учетом производительности, надежности и расширяемости.