package preparation

import (
	"github.com/stackfoundation/sandbox/core/pkg/workflows/expansion"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/validation"
	"github.com/stackfoundation/sandbox/log"
)

type stepExpansionError struct {
	err  error
	step string
}

func (e *stepExpansionError) Error() string {
	return "Error replacing variable placeholders in step " + e.step + ":\n" + e.err.Error()
}

func expandStep(workflow *v1.Workflow, step *v1.WorkflowStep, stepSelector []int) error {
	stepName := step.StepName(stepSelector)
	log.Debugf("Expanding variable placeholders in step %v", stepName)

	err := expansion.ExpandStep(step, workflow.Spec.State.Variables)
	if err != nil {
		if step.IgnoreMissing() == nil {
			if !workflow.Spec.IgnoreMissing {
				return &stepExpansionError{err: err, step: stepName}
			}
		} else if !*step.IgnoreMissing() {
			return &stepExpansionError{err: err, step: stepName}
		}

		log.Debugf("Ignoring missing variable placeholders in step %v:\n%v", stepName, err)
	}

	return nil
}

func shouldIgnoreValidation(workflow *v1.Workflow, step *v1.WorkflowStep, stepSelector []int, err error) error {
	if step.IgnoreValidation() == nil {
		if !workflow.Spec.IgnoreValidation {
			return err
		}
	} else if !*step.IgnoreValidation() {
		return err
	}

	log.Debugf("Ignoring validation errors in step %v:\n%v", step.StepName(stepSelector), err)
	return nil
}

// PrepareStepIfNecessary Prepare the step for execution
func PrepareStepIfNecessary(workflow *v1.Workflow, step *v1.WorkflowStep, stepSelector []int) error {
	if step != nil && !step.State.Prepared {
		err := expandStep(workflow, step, stepSelector)
		if err != nil {
			return err
		}

		err = validation.ValidateStep(step, stepSelector)
		if err != nil {
			err = shouldIgnoreValidation(workflow, step, stepSelector, err)
			if err != nil {
				return err
			}
		}

		step.State.Prepared = true
	}

	return nil
}
