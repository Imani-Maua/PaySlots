package domain


import (
	"testing"
	"time"
)


func TestCancel(test *testing.T){

	tests := []struct{
		name string
		initialStatus SubscriptionStatus
		wantErr bool
	}{
		{"from active", StatusActive, false},
		{"from trialing", StatusTrialing, false},
		{"from suspended", StatusSuspended, true},
		{"from past_due", StatusPastDue, true},
	}

	for _, tt := range tests{
		test.Run(tt.name, func(t *testing.T){
			sub := &Subscription{Status: tt.initialStatus}
			err := sub.Cancel()

			if (err != nil) != tt.wantErr {
				t.Errorf("got err = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr && sub.Status != StatusCancelled {
				t.Errorf("expected status %q, got %q", StatusCancelled, sub.Status)
			}
		})
	}
}
func TestBeginTrial(t *testing.T){


	tests := []struct{
		name string
		initialStatus SubscriptionStatus
		newPlanID int64
		wantErr bool
	}{
		{"from free", StatusFree, 12, false},
		{"from active", StatusActive, 12, false},
		{"from cancelled", StatusCancelled, 12, true},
		{"from trialing", StatusTrialing, 12, true},
		{"from pastdue", StatusPastDue, 12, true},
		{"from suspended", StatusSuspended, 12, true},
		
	}
	

	now := time.Now()
	for _, tt := range tests{
		t.Run(tt.name, func(t *testing.T) {

			sub := &Subscription{
				Status: tt.initialStatus,
			}

			err := sub.BeginTrial(tt.newPlanID, now)

			if (err != nil) != tt.wantErr{
				t.Errorf("got err = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr{
				if sub.Status != StatusTrialing {
					t.Errorf("expected status %q, got %q", StatusTrialing, sub.Status)
				}

				if sub.PlanID != tt.newPlanID {
					t.Errorf("expected planID %d, got %d", tt.newPlanID, sub.PlanID)
				}

				if sub.PendingPlanID != nil {
					t.Errorf("expected pendingPlanID to be nil, got %v", sub.PendingPlanID)
				}

				if sub.PeriodEnd != now.AddDate(0,0, 30){
					t.Errorf("expected PeriodEnd %v, got %v", now.AddDate(0, 0, 30), sub.PeriodEnd)
				}
			}
		})
	}
}
func TestUpdateTrialPlan(t *testing.T) {

	now := time.Now()

	tests := []struct {
		name          string
		initialStatus SubscriptionStatus
		periodEnd     time.Time
		newPlan       int64
		wantErr       bool
	}{
		{"from trialing, valid period", StatusTrialing, now.AddDate(0, 0, 1), 12, false},
		{"from trialing, expired", StatusTrialing, now.AddDate(0, 0, -1), 12, true},
		{"from active", StatusActive, now.AddDate(0, 0, 1), 12, true},
		{"from free", StatusFree, now.AddDate(0, 0, 1), 12, true},
		{"from past_due", StatusPastDue, now.AddDate(0, 0, 1), 12, true},
		{"from suspended", StatusSuspended, now.AddDate(0, 0, 1), 12, true},
		{"from cancelled", StatusCancelled, now.AddDate(0, 0, 1), 12, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sub := &Subscription{
				Status:    tt.initialStatus,
				PeriodEnd: tt.periodEnd,
			}

			err := sub.UpdateTrialPlan(tt.newPlan, now)

			if (err != nil) != tt.wantErr {
				t.Errorf("got err = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if sub.PlanID != tt.newPlan {
					t.Errorf("expected planID %d, got %d", tt.newPlan, sub.PlanID)
				}
			}
		})
	}
}


func TestScheduleDowngrade(t *testing.T){

	tests := []struct{
		name string
		initialStatus SubscriptionStatus
		newPlanID int64
		wantErr bool
	}{
		{"from active", StatusActive, 12, false},
		{"from trialing", StatusTrialing, 12, true},
		{"from suspended", StatusSuspended, 12, true},
		{"from past_due", StatusPastDue, 12, true},
		{"from free", StatusFree, 12, true},
		{"from cancelled", StatusCancelled, 12, true},
	}

	for _, tt := range tests{
		t.Run(tt.name, func(t *testing.T) {

			sub := &Subscription{Status: tt.initialStatus}

			err := sub.ScheduleDowngradePlan(tt.newPlanID)

			if (err != nil) != tt.wantErr{
				t.Errorf("err = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr {
    			if sub.PendingPlanID == nil {
        			t.Errorf("expected pendingPlanID to be set, got nil")
    			} else if *sub.PendingPlanID != tt.newPlanID {
        			t.Errorf("expected pendingPlanID %d, got %d", tt.newPlanID, *sub.PendingPlanID)
    			}
			}
		})
	}
}

func TestMarkPastDue(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus SubscriptionStatus
		wantErr       bool
	}{
		{"from active", StatusActive, false},
		{"from trialing", StatusTrialing, true},
		{"from free", StatusFree, true},
		{"from past_due", StatusPastDue, true},
		{"from suspended", StatusSuspended, true},
		{"from cancelled", StatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscription{Status: tt.initialStatus}

			err := sub.MarkPastDue()

			if (err != nil) != tt.wantErr {
				t.Errorf("got err = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if sub.Status != StatusPastDue {
					t.Errorf("expected status %q, got %q", StatusPastDue, sub.Status)
				}
				if sub.RetryCount != 0 {
					t.Errorf("expected RetryCount 0, got %d", sub.RetryCount)
				}
			}
		})
	}
}

func TestRetry(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  SubscriptionStatus
		initialRetries int
		wantExhausted  bool
		wantErr        bool
	}{
		{"first retry", StatusPastDue, 0, false, false},
		{"second retry", StatusPastDue, 1, false, false},
		{"third retry, exhausted", StatusPastDue, 2, true, false},
		{"from active", StatusActive, 0, false, true},
		{"from trialing", StatusTrialing, 0, false, true},
		{"from suspended", StatusSuspended, 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscription{
				Status:     tt.initialStatus,
				RetryCount: tt.initialRetries,
			}

			exhausted, err := sub.Retry()

			if (err != nil) != tt.wantErr {
				t.Errorf("got err = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if exhausted != tt.wantExhausted {
					t.Errorf("got exhausted = %v, want %v", exhausted, tt.wantExhausted)
				}
				if sub.RetryCount != tt.initialRetries+1 {
					t.Errorf("expected RetryCount %d, got %d", tt.initialRetries+1, sub.RetryCount)
				}
			}
		})
	}
}

func TestSuspend(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus SubscriptionStatus
		wantErr       bool
	}{
		{"from past_due", StatusPastDue, false},
		{"from active", StatusActive, true},
		{"from trialing", StatusTrialing, true},
		{"from free", StatusFree, true},
		{"from suspended", StatusSuspended, true},
		{"from cancelled", StatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscription{Status: tt.initialStatus}

			err := sub.Suspend()

			if (err != nil) != tt.wantErr {
				t.Errorf("got err = %v, wantErr = %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if sub.Status != StatusSuspended {
					t.Errorf("expected status %q, got %q", StatusSuspended, sub.Status)
				}
			}
		})
	}
}
