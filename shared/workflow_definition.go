package shared

// WorkflowDefinition represents a user-designed workflow stored as JSON.
type WorkflowDefinition struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Version     int               `json:"version"`
	Steps       []StepDefinition  `json:"steps"`
	Events      []EventDefinition `json:"events,omitempty"`
	Gateways    []GatewayDefinition `json:"gateways,omitempty"`
	Edges       []EdgeDefinition    `json:"edges,omitempty"`
}

// Position represents the x/y coordinates of a node on the canvas.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// EventDefinition represents a start or end event in the workflow.
type EventDefinition struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Label       string            `json:"label"`
	Position    Position          `json:"position"`
	StartConfig *StartEventConfig `json:"start_config,omitempty"`
	EndConfig   *EndEventConfig   `json:"end_config,omitempty"`
}

// StartEventConfig holds configuration for start events.
type StartEventConfig struct {
	TriggerType string         `json:"triggerType"`
	Timer       *TimerConfig   `json:"timer,omitempty"`
	Webhook     *WebhookConfig `json:"webhook,omitempty"`
	FormFields  []FormField    `json:"formFields"`
}

// TimerConfig defines timer trigger settings.
type TimerConfig struct {
	Type  string `json:"type"`
	Cron  string `json:"cron,omitempty"`
	Delay string `json:"delay,omitempty"`
}

// WebhookConfig defines webhook trigger settings.
type WebhookConfig struct {
	EndpointPath string `json:"endpointPath,omitempty"`
	EventName    string `json:"eventName,omitempty"`
}

// FormField defines a form input field for start events.
type FormField struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Label        string `json:"label"`
	Required     bool   `json:"required"`
	DefaultValue string `json:"defaultValue,omitempty"`
}

// EndEventConfig holds configuration for end events.
type EndEventConfig struct {
	EndType         string           `json:"endType"`
	Error           *ErrorEndConfig  `json:"error,omitempty"`
	OutputVariables []OutputVariable `json:"outputVariables"`
}

// ErrorEndConfig defines error end event settings.
type ErrorEndConfig struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// OutputVariable defines an output variable mapping for end events.
type OutputVariable struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
}

// GatewayDefinition represents an exclusive gateway (decision point) in the workflow.
type GatewayDefinition struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Label    string   `json:"label"`
	Position Position `json:"position"`
}

// EdgeCondition holds the condition expression for a gateway outgoing edge.
type EdgeCondition struct {
	Expression string `json:"expression"`
	IsDefault  bool   `json:"isDefault"`
}

// EdgeDefinition represents a connection between nodes, optionally with a condition.
type EdgeDefinition struct {
	ID        string         `json:"id"`
	Source    string         `json:"source"`
	Target   string         `json:"target"`
	Condition *EdgeCondition `json:"condition,omitempty"`
	Label     string         `json:"label,omitempty"`
}

// StepDefinition represents a single activity step within a workflow.
type StepDefinition struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	ActivityName   string            `json:"activity_name"`
	TaskQueue      string            `json:"task_queue"`
	TimeoutSeconds int               `json:"timeout_seconds"`
	RetryPolicy    *RetryConfig      `json:"retry_policy,omitempty"`
	InputMapping   map[string]string `json:"input_mapping"`
	DependsOn      []string          `json:"depends_on"`
}

// RetryConfig defines retry behavior for an activity step.
type RetryConfig struct {
	MaxAttempts        int     `json:"max_attempts"`
	InitialIntervalSec int     `json:"initial_interval_sec"`
	BackoffCoefficient float64 `json:"backoff_coefficient"`
}
