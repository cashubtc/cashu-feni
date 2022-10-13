package cashu

import (
	"testing"
)

func TestToJson(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "json", args: args{i: &struct {
			Test    string `json:"test"`
			Integer int    `json:"integer"`
		}{Integer: 1, Test: "cashu"}}, want: `{"test":"cashu","integer":1}`},
		{name: "annotate", args: args{i: &struct {
			Test    string
			Integer int
		}{Integer: 1, Test: "cashu"}}, want: `{"Test":"cashu","Integer":1}`},
		{name: "error", args: args{i: ErrorResponse{
			Err:  "exception",
			Code: 200,
		}}, want: `{"error":"exception","code":200}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToJson(tt.args.i); got != tt.want {
				t.Errorf("ToJson() = %v, want %v", got, tt.want)
			}
		})
	}
}
