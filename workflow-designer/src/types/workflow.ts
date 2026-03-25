export type NodeKind = 'startEvent' | 'endEvent' | 'activity' | 'exclusiveGateway'

export interface PaletteCategory {
  name: string
  collapsed: boolean
  items: PaletteItem[]
}

export interface InputFieldSchema {
  name: string
  type: string
  required: boolean
  description: string
}

export interface PaletteItem {
  kind: NodeKind
  displayName: string
  activityName?: string
  defaultTaskQueue?: string
  defaultInputMapping?: Record<string, string>
  inputSchema?: InputFieldSchema[]
  icon: string
}

export const PALETTE_CATEGORIES: PaletteCategory[] = [
  {
    name: 'Events',
    collapsed: false,
    items: [
      { kind: 'startEvent', displayName: 'Start Event', icon: 'start' },
      { kind: 'endEvent', displayName: 'End Event', icon: 'end' },
      { kind: 'exclusiveGateway', displayName: 'Exclusive Gateway', icon: 'gateway' },
    ],
  },
]

export interface EdgeCondition {
  expression: string
  isDefault: boolean
}

export interface GatewayDefinition {
  id: string
  type: 'exclusiveGateway'
  label: string
  position: { x: number; y: number }
}

export interface EdgeDefinition {
  id: string
  source: string
  target: string
  condition?: EdgeCondition
  label?: string
}

export interface RetryConfig {
  max_attempts: number
  initial_interval_sec: number
  backoff_coefficient: number
}

export interface StepDefinition {
  id: string
  name: string
  activity_name: string
  task_queue: string
  timeout_seconds: number
  retry_policy?: RetryConfig
  input_mapping: Record<string, string>
  depends_on: string[]
}

// Start Event
export type StartTriggerType = 'manual' | 'timer' | 'webhook'

export interface TimerConfig {
  type: 'cron' | 'delay'
  cron?: string
  delay?: string
}

export interface WebhookConfig {
  endpointPath?: string
  eventName?: string
}

export interface FormField {
  name: string
  type: 'string' | 'number' | 'boolean' | 'json'
  label: string
  required: boolean
  defaultValue?: string
}

export interface StartEventConfig {
  triggerType: StartTriggerType
  timer?: TimerConfig
  webhook?: WebhookConfig
  formFields: FormField[]
}

// End Event
export type EndEventType = 'none' | 'error' | 'terminate'

export interface ErrorEndConfig {
  errorCode: string
  errorMessage: string
}

export interface OutputVariable {
  name: string
  expression: string
}

export interface EndEventConfig {
  endType: EndEventType
  error?: ErrorEndConfig
  outputVariables: OutputVariable[]
}

// Event persistence
export interface EventDefinition {
  id: string
  type: 'startEvent' | 'endEvent'
  label: string
  position: { x: number; y: number }
  start_config?: StartEventConfig
  end_config?: EndEventConfig
}

export interface WorkflowDefinition {
  id: string
  name: string
  description: string
  version: number
  steps: StepDefinition[]
  events?: EventDefinition[]
  gateways?: GatewayDefinition[]
  edges?: EdgeDefinition[]
}

