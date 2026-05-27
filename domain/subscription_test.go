package domain


import "testing"


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