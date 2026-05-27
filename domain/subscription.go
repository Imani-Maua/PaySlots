package domain

import (
	"fmt"
	"time"
)


type SubscriptionStatus string


const (
	StatusFree SubscriptionStatus = "free"
	StatusTrialing SubscriptionStatus = "trialing"
	StatusActive SubscriptionStatus = "active"
	StatusPastDue SubscriptionStatus = "past_due"
	StatusSuspended SubscriptionStatus = "suspended"
	StatusCancelled SubscriptionStatus = "cancelled"
)


type Subscription struct {
	ID int64
	VenueID int64
	PlanID int64
	PendingPlanID *int64 
	//this starts as nil since a user starts at free and there's no plan to "downgrade to"
	//once we are on any other status, this needs to be updated.
	//if we are trialing, the pending plan is the plan we want to be on once trialing is over
	//if we are on an active plan, the pending plan is free if payment is not received.
	Status SubscriptionStatus
	PeriodEnd time.Time
	RetryCount int 
	CreatedAt time.Time
	UpdatedAt time.Time
}



// ---- state machine methods -----

// -- Move Subscription into a trialing state on a new plan

func (sub *Subscription) BeginTrial(newPlanID int64, now time.Time) error {

	switch sub.Status {
	case StatusFree, StatusActive:
		// valid states - fall through
	default:
		return fmt.Errorf("upgrade not allowed from status :%q", sub.Status)
	}

	sub.PlanID = newPlanID
	sub.PendingPlanID = nil // -- this is why PendingPlanID is actually a pointer since it can hold nil (or null, but int64 cannot)
	sub.Status = StatusTrialing
	sub.PeriodEnd = now.AddDate(0, 0, 30)

	return nil
}


// -- Check that we are still within a valid trial period for a user to change plans
// we should be able to update the trial plan only from those plans that are already in trial mode
// the status does not change, we change the plans

func (sub *Subscription) UpdateTrialPlan(newPlanID int64, now time.Time) error {
	
	switch sub.Status {
		case StatusTrialing:
			// valid therefore fall through
		default:
			return fmt.Errorf("updating not allowed from status: %q", sub.Status)
	}
	if now.After(sub.PeriodEnd){
		return fmt.Errorf("trial period has expired")
	}

	sub.PlanID = newPlanID
	return nil

}

// ScheduleDowngradePlan records a pending plan change to take effect at period end.
func (sub *Subscription) ScheduleDowngradePlan(newPlanID int64) error {

	if sub.Status != StatusActive{
		return fmt.Errorf("downgrade only allowed from active status, got %q", sub.Status)
	}

	sub.PendingPlanID = &newPlanID
	return nil
}

// ActivatePlanAfterPayment moves trialing → active (trial ended, payment succeeded).

func (sub *Subscription) ActivatePlan() error {
	if sub.Status != StatusTrialing{
		return fmt.Errorf("activate called on non-trialing subscription, %q", sub.Status)
	}

	sub.Status = StatusActive
	if sub.PendingPlanID != nil {
		sub.PlanID = * sub.PendingPlanID
		sub.PendingPlanID = nil
	}

	return nil
}

// MarkPastDue moves active → past_due when a renewal payment fails.
func (sub *Subscription) MarkPastDue() error {
	if sub.Status != StatusActive{
		return fmt.Errorf("cannot mark past due from status %q", sub.Status)
	}

	sub.Status = StatusPastDue
	sub.RetryCount = 0
	return nil
}


// Retry records a failed retry attempt. Returns true if retries are exhausted.
// You are supposed to be able to be able to call retry logic 
func ( sub *Subscription) Retry() (exhausted bool, err error) {

	if sub.Status != StatusPastDue{
		return false, fmt.Errorf("retry called on a subscription that is not past due.")
	}
	sub.RetryCount ++

	return sub.RetryCount >= 3, nil
}

// Suspend moves the subscription to suspended (retries exhausted).

func (sub *Subscription) Suspend()  error{

	if sub.Status != StatusPastDue{
		return fmt.Errorf("suspend called from invalid status")
	}

	sub.Status = StatusSuspended
	return  nil
}

// Cancel moves active → cancelled.

func (sub *Subscription) Cancel() error {
	switch sub.Status{
	case StatusActive, StatusTrialing:
			//fall through
	default:
		return fmt.Errorf("cancel not allowed from status %q", sub.Status)
	}
	sub.Status = StatusCancelled
	return nil
}