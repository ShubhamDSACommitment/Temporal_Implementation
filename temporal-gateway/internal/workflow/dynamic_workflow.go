package workflow

import (
	"fmt"
	"strings"
	"time"

	"temporal-shared"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DynamicWorkflowInput is the input for a dynamic workflow execution.
type DynamicWorkflowInput struct {
	Definition shared.WorkflowDefinition `json:"definition"`
	Input      map[string]interface{}    `json:"input"`
}

// DynamicWorkflow executes a user-designed workflow definition as a DAG.
func DynamicWorkflow(ctx workflow.Context, dwInput DynamicWorkflowInput) (map[string]interface{}, error) {
	def := dwInput.Definition
	input := dwInput.Input

	// Build a map of step ID → step for quick lookup
	stepMap := make(map[string]shared.StepDefinition, len(def.Steps))
	for _, s := range def.Steps {
		stepMap[s.ID] = s
	}

	// Results from completed steps
	stepResults := make(map[string]map[string]interface{})

	// Track which steps are done
	completed := make(map[string]bool)

	// Topological execution: keep running until all steps are done
	for len(completed) < len(def.Steps) {
		// Find steps whose dependencies are all satisfied
		var ready []shared.StepDefinition
		for _, s := range def.Steps {
			if completed[s.ID] {
				continue
			}
			allDepsMet := true
			for _, dep := range s.DependsOn {
				if !completed[dep] {
					allDepsMet = false
					break
				}
			}
			if allDepsMet {
				ready = append(ready, s)
			}
		}

		if len(ready) == 0 {
			return nil, fmt.Errorf("deadlock: no steps ready but %d remain", len(def.Steps)-len(completed))
		}

		// Execute all ready steps in parallel
		futures := make(map[string]workflow.Future, len(ready))
		for _, step := range ready {
			activityInput := resolveInputMapping(step.InputMapping, input, stepResults)

			timeout := time.Duration(step.TimeoutSeconds) * time.Second
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			actOpts := workflow.ActivityOptions{
				TaskQueue:           step.TaskQueue,
				StartToCloseTimeout: timeout,
			}

			if step.RetryPolicy != nil {
				rp := step.RetryPolicy
				actOpts.RetryPolicy = &temporal.RetryPolicy{
					MaximumAttempts: int32(rp.MaxAttempts),
				}
				if rp.InitialIntervalSec > 0 {
					actOpts.RetryPolicy.InitialInterval = time.Duration(rp.InitialIntervalSec) * time.Second
				}
				if rp.BackoffCoefficient > 0 {
					actOpts.RetryPolicy.BackoffCoefficient = rp.BackoffCoefficient
				}
			}

			actCtx := workflow.WithActivityOptions(ctx, actOpts)
			futures[step.ID] = workflow.ExecuteActivity(actCtx, step.ActivityName, activityInput)
		}

		// Wait for all parallel steps to complete
		for _, step := range ready {
			var result map[string]interface{}
			if err := futures[step.ID].Get(ctx, &result); err != nil {
				return nil, fmt.Errorf("step %s (%s) failed: %w", step.ID, step.ActivityName, err)
			}
			stepResults[step.ID] = result
			completed[step.ID] = true
		}
	}

	// Return the result from the last step(s) merged
	finalResult := make(map[string]interface{})
	for k, v := range stepResults {
		finalResult[k] = v
	}
	finalResult["_input"] = input
	return finalResult, nil
}

// resolveInputMapping builds the activity input by resolving expressions like
// $input.field and $steps.stepId.field.
func resolveInputMapping(mapping map[string]string, wfInput map[string]interface{}, stepResults map[string]map[string]interface{}) map[string]interface{} {
	if len(mapping) == 0 {
		return wfInput
	}

	resolved := make(map[string]interface{}, len(mapping))
	for key, expr := range mapping {
		resolved[key] = resolveExpression(expr, wfInput, stepResults)
	}
	return resolved
}

func resolveExpression(expr string, wfInput map[string]interface{}, stepResults map[string]map[string]interface{}) interface{} {
	expr = strings.TrimSpace(expr)

	// $input.fieldName → value from workflow input
	if strings.HasPrefix(expr, "$input.") {
		field := strings.TrimPrefix(expr, "$input.")
		if v, ok := wfInput[field]; ok {
			return v
		}
		return nil
	}

	// $steps.stepId.fieldName → value from a prior step's result
	if strings.HasPrefix(expr, "$steps.") {
		parts := strings.SplitN(strings.TrimPrefix(expr, "$steps."), ".", 2)
		if len(parts) == 2 {
			if stepResult, ok := stepResults[parts[0]]; ok {
				if v, ok := stepResult[parts[1]]; ok {
					return v
				}
			}
		}
		return nil
	}

	// Literal value
	return expr
}
