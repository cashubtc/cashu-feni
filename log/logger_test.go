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

func TestNewRotateFileHook(t *testing.T) {
	type args struct {
		config RotateFileConfig
	}
	tests := []struct {
		name    string
		args    args
		want    logrus.Hook
		wantErr bool
	}{
		{name: "NewRotateFileHook", want: &RotateFileHook{Config: RotateFileConfig{Filename: "out.log"}}, args: args{RotateFileConfig{Filename: "out.log"}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRotateFileHook(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRotateFileHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.(*RotateFileHook).Config, tt.want.(*RotateFileHook).Config) {
				t.Errorf("NewRotateFileHook() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRotateFileHook_Levels(t *testing.T) {
	tests := []struct {
		name string
		want []logrus.Level
	}{
		{name: "allLevels", want: logrus.AllLevels},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook, err := NewRotateFileHook(RotateFileConfig{Level: logrus.TraceLevel})
			if err != nil {
				panic(err)
			}
			if got := hook.Levels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Levels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToMap(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{name: "toMap", args: args{i: struct {
			Test string
		}{
			Test: "hi",
		}}, want: map[string]interface{}{"Test": "hi"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToMap(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	type args struct {
		logLevel string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "hookTest", args: args{logLevel: "trace"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preConfigureHooks := len(logrus.StandardLogger().Hooks)
			Configure(tt.args.logLevel)
			if len(logrus.StandardLogger().Hooks) == preConfigureHooks {
				t.Errorf("invalid log hook configuration")
			}
		})
	}
}
