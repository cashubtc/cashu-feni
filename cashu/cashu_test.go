package cashu

import (
	"encoding/hex"
	"fmt"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/cashubtc/cashu-feni/lightning/invoice"
	cashuLog "github.com/cashubtc/cashu-feni/log"
	"github.com/google/uuid"
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
			if got := cashuLog.ToJson(tt.args.i); got != tt.want {
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
		want lightning.Invoicer
	}{
		{name: "createNoInvoice", want: nil},
		{name: "createInvoice", want: &invoice.Invoice{}},
		{name: "lightningOnly", want: &invoice.Invoice{}},
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
				Reserved:     tt.fields.reserved,
				Script:       tt.fields.Script,
				SendId:       uuid.New(),
				TimeCreated:  tt.fields.timeCreated,
				TimeReserved: tt.fields.timeReserved,
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

func TestPromise_Log(t *testing.T) {
	type fields struct {
		B_b    string
		C_c    string
		Amount uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{name: "promiseLog", want: map[string]interface{}{"B_b": "1234a", "C_c": "1234", "Amount": uint64(1)}, fields: fields{Amount: 1, C_c: "1234", B_b: "1234a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Promise{
				B_b:    tt.fields.B_b,
				C_c:    tt.fields.C_c,
				Amount: tt.fields.Amount,
			}
			if got := p.Log(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProof_Log(t *testing.T) {
	type fields struct {
		Id           string
		Amount       uint64
		Secret       string
		C            string
		reserved     bool
		Script       *P2SHScript
		sendId       string
		status       ProofStatus
		timeCreated  time.Time
		timeReserved time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		{name: "proofLog", want: map[string]interface{}{"Id": "1234a", "Amount": uint64(1), "Secret": "1", "C": "1234", "Reserved": false, "Status": ProofStatusPending},
			fields: fields{Amount: 1, C: "1234", Id: "1234a", Secret: "1", Script: nil, status: ProofStatusPending}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Proof{
				Id:           tt.fields.Id,
				Amount:       tt.fields.Amount,
				Secret:       tt.fields.Secret,
				C:            tt.fields.C,
				Reserved:     tt.fields.reserved,
				Script:       tt.fields.Script,
				TimeCreated:  tt.fields.timeCreated,
				TimeReserved: tt.fields.timeReserved,
				Status:       tt.fields.status,
			}
			if got := p.Log(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Log() = %v, want %v", got, tt.want)
			}
		})
	}
}
