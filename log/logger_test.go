package cashuLog

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
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
		{name: "error", args: args{i: struct {
			Err  string `json:"error"`
			Code int    `json:"code"`
		}{
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

func TestWithLoggable(t *testing.T) {
	type args struct {
		l    Loggable
		more []interface{}
	}
	tests := []struct {
		name string
		args args
		want logrus.Fields
	}{
		{name: "loggable", args: args{more: []interface{}{fmt.Errorf("test")}}, want: logrus.Fields{"error.message": fmt.Errorf("test")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithLoggable(tt.args.l, tt.args.more...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithLoggable() = %v, want %v", got, tt.want)
			}
		})
	}
}
