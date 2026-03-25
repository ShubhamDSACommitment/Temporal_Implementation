import axios from 'axios'
import type { WorkflowDefinition } from '../types/workflow'

const api = axios.create({
  baseURL: '/api',
})

export async function listDefinitions(): Promise<WorkflowDefinition[]> {
  const { data } = await api.get('/definitions')
  return data
}

export async function getDefinition(id: string): Promise<WorkflowDefinition> {
  const { data } = await api.get(`/definitions/${id}`)
  return data
}

export async function saveDefinition(def: WorkflowDefinition): Promise<WorkflowDefinition> {
  const { data } = await api.post('/definitions', def)
  return data
}

export async function updateDefinition(def: WorkflowDefinition): Promise<WorkflowDefinition> {
  const { data } = await api.put(`/definitions/${def.id}`, def)
  return data
}

export async function deleteDefinition(id: string): Promise<void> {
  await api.delete(`/definitions/${id}`)
}

export interface StartWorkflowRequest {
  definition_id: string
  input: Record<string, any>
}

export interface StartWorkflowResponse {
  workflow_id: string
  run_id: string
}

export async function startWorkflow(req: StartWorkflowRequest): Promise<StartWorkflowResponse> {
  const { data } = await api.post('/workflows/start', req)
  return data
}

export interface WorkflowStatus {
  workflow_id: string
  run_id: string
  status: string
  result?: any
  error?: string
}

export async function getWorkflowStatus(workflowId: string): Promise<WorkflowStatus> {
  const { data } = await api.get(`/workflows/${workflowId}/status`)
  return data
}

export interface InputFieldSchema {
  name: string
  type: string
  required: boolean
  description: string
}

export interface RegisteredActivity {
  activity_name: string
  display_name: string
  description: string
  task_queue: string
  service_name: string
  category: string
  input_schema: InputFieldSchema[]
  default_mapping: Record<string, string>
}

export interface ActivityCategory {
  name: string
  activities: RegisteredActivity[]
}

export async function listActivities(): Promise<ActivityCategory[]> {
  const { data } = await api.get('/activities')
  return data.categories ?? []
}
