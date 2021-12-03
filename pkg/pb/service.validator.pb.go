// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: service.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/mwitkow/go-proto-validators"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
	_ "google.golang.org/protobuf/types/known/emptypb"
	_ "google.golang.org/protobuf/types/known/wrapperspb"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *Count) Validate() error {
	return nil
}
func (this *QuoteRequest) Validate() error {
	if this.Code == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Code", fmt.Errorf(`value '%v' must not be an empty string`, this.Code))
	}
	if this.Date == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Date", fmt.Errorf(`value '%v' must not be an empty string`, this.Date))
	}
	return nil
}
func (this *Stock) Validate() error {
	if this.Code == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Code", fmt.Errorf(`value '%v' must not be an empty string`, this.Code))
	}
	if this.Name == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Name", fmt.Errorf(`value '%v' must not be an empty string`, this.Name))
	}
	if this.Suspend == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Suspend", fmt.Errorf(`value '%v' must not be an empty string`, this.Suspend))
	}
	return nil
}
func (this *Quote) Validate() error {
	if this.Code == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Code", fmt.Errorf(`value '%v' must not be an empty string`, this.Code))
	}
	if !(this.Open >= 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Open", fmt.Errorf(`value '%v' must be greater than or equal to '0'`, this.Open))
	}
	if !(this.Close >= 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Close", fmt.Errorf(`value '%v' must be greater than or equal to '0'`, this.Close))
	}
	if !(this.High >= 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("High", fmt.Errorf(`value '%v' must be greater than or equal to '0'`, this.High))
	}
	if !(this.Low >= 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Low", fmt.Errorf(`value '%v' must be greater than or equal to '0'`, this.Low))
	}
	if !(this.YesterdayClosed >= 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("YesterdayClosed", fmt.Errorf(`value '%v' must be greater than or equal to '0'`, this.YesterdayClosed))
	}
	if !(this.Account >= 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Account", fmt.Errorf(`value '%v' must be greater than or equal to '0'`, this.Account))
	}
	if this.Date == "" {
		return github_com_mwitkow_go_proto_validators.FieldError("Date", fmt.Errorf(`value '%v' must not be an empty string`, this.Date))
	}
	return nil
}
