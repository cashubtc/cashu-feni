package cashu

import (
	"encoding/hex"
	"fmt"
	"github.com/gohumble/cashu-feni/lightning"
	"github.com/gohumble/cashu-feni/lightning/lnbits"
	"reflect"
	"testing"
	"time"
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

func TestWithCode(t *testing.T) {
	type args struct {
		code int
	}
	tests := []struct {
		name string
		args args
		want ErrorResponse
	}{
		{name: "withCode", want: ErrorResponse{Code: 1, Err: "test"}, args: args{code: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorResponse(fmt.Errorf("test"), WithCode(tt.args.code))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	type args struct {
		err     error
		options []ErrorOptions
	}
	tests := []struct {
		name string
		args args
		want ErrorResponse
	}{
		{name: "withCode", want: ErrorResponse{Code: 1, Err: "test"},
			args: args{err: fmt.Errorf("test"),
				options: []ErrorOptions{WithCode(1)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewErrorResponse(tt.args.err, tt.args.options...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorResponse_String(t *testing.T) {
	type fields struct {
		Err  string
		Code int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "string", fields: fields{Err: "test", Code: 1}, want: "{\"error\":\"test\",\"code\":1}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ErrorResponse{
				Err:  tt.fields.Err,
				Code: tt.fields.Code,
			}
			if got := e.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorResponse_Error(t *testing.T) {
	type fields struct {
		Err  string
		Code int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "err", want: "test", fields: fields{Err: "test"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ErrorResponse{
				Err:  tt.fields.Err,
				Code: tt.fields.Code,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateInvoice(t *testing.T) {
	tests := []struct {
		name string
		want lightning.Invoice
	}{
		{name: "createNoInvoice", want: nil},
		{name: "createInvoice", want: &lnbits.Invoice{}},
		{name: "lightningOnly", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "createInvoice":
				lightning.Config.Lightning.Lnbits = &lightning.LnbitsConfig{}
				lightning.Config.Lightning.Enabled = true
			case "lightningOnly":
				lightning.Config.Lightning.Enabled = true
				lightning.Config.Lightning.Lnbits = nil
			}
			if got := CreateInvoice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestProof_Decode(t *testing.T) {
	type fields struct {
		Id           string
		Amount       uint64
		Secret       string
		C            string
		reserved     bool
		Script       *P2SHScript
		sendId       string
		timeCreated  time.Time
		timeReserved time.Time
	}
	msg := hex.EncodeToString([]byte("hello"))
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{name: "decode", want: []byte("hello"), fields: fields{C: msg}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Proof{
				Id:           tt.fields.Id,
				Amount:       tt.fields.Amount,
				Secret:       tt.fields.Secret,
				C:            tt.fields.C,
				reserved:     tt.fields.reserved,
				Script:       tt.fields.Script,
				sendId:       tt.fields.sendId,
				timeCreated:  tt.fields.timeCreated,
				timeReserved: tt.fields.timeReserved,
			}
			got, err := p.Decode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() got = %v, want %v", string(got), tt.want)
			}
		})
	}
}
