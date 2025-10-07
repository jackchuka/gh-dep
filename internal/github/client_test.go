package github

import "testing"

func TestDeriveCIState(t *testing.T) {
	tests := []struct {
		name   string
		suites checkSuiteResponse
		status statusResponse
		want   string
	}{
		{
			name: "check suites success",
			suites: checkSuiteResponse{
				CheckSuites: []checkSuite{
					{Status: "completed", Conclusion: strPtr("success")},
				},
			},
			status: statusResponse{},
			want:   "success",
		},
		{
			name: "check suites failure",
			suites: checkSuiteResponse{
				CheckSuites: []checkSuite{
					{Status: "completed", Conclusion: strPtr("failure")},
				},
			},
			status: statusResponse{},
			want:   "failure",
		},
		{
			name: "queued suite ignored when another succeeds",
			suites: checkSuiteResponse{
				CheckSuites: []checkSuite{
					{Status: "queued"},
					{Status: "completed", Conclusion: strPtr("success")},
				},
			},
			status: statusResponse{},
			want:   "success",
		},
		{
			name: "status pending overrides success",
			suites: checkSuiteResponse{
				CheckSuites: []checkSuite{
					{Status: "completed", Conclusion: strPtr("success")},
				},
			},
			status: statusResponse{
				Statuses: []statusContext{
					{State: "pending"},
				},
			},
			want: "pending",
		},
		{
			name: "status failure overrides suite success",
			suites: checkSuiteResponse{
				CheckSuites: []checkSuite{
					{Status: "completed", Conclusion: strPtr("success")},
				},
			},
			status: statusResponse{
				Statuses: []statusContext{
					{State: "failure"},
				},
			},
			want: "failure",
		},
		{
			name: "status success without suites",
			status: statusResponse{
				Statuses: []statusContext{
					{State: "success"},
				},
			},
			want: "success",
		},
		{
			name: "status failure without suites",
			status: statusResponse{
				Statuses: []statusContext{
					{State: "failure"},
				},
			},
			want: "failure",
		},
		{
			name: "status pending without suites",
			status: statusResponse{
				Statuses: []statusContext{
					{State: "pending"},
				},
			},
			want: "pending",
		},
		{
			name: "suite in progress is pending",
			suites: checkSuiteResponse{
				CheckSuites: []checkSuite{
					{Status: "in_progress"},
				},
			},
			status: statusResponse{},
			want:   "pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveCIState(tt.suites, tt.status)
			if got != tt.want {
				t.Fatalf("deriveCIState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
