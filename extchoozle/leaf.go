package extchoozle

import (
	"context"
	"fmt"

	"github.com/jdmeyer3/errors"
	"github.com/jdmeyer3/errors/errbase"
	"github.com/jdmeyer3/errors/markers"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// This file demonstrates how to add a wrapper type not otherwise
// known to the rest of the library.

// WithChoozleError is our wrapper type.
type WithChoozleError struct {
	cause error
	code  ChoozleErrorCode
}

// WrapWithChoozleError adds a HTTP code to an existing error.
func WrapWithChoozleError(err error, code ChoozleErrorCode, overrideDefaultMessage string) error {
	if err == nil {
		return nil
	}
	return &WithChoozleError{cause: err, code: code}
}

// GetHTTPCode retrieves the HTTP code from a stack of causes.
func GetHTTPCode(err error, defaultCode int) int {
	if v, ok := markers.If(err, func(err error) (interface{}, bool) {
		if w, ok := err.(*WithChoozleError); ok {
			return w.code, true
		}
		return nil, false
	}); ok {
		return v.(int)
	}
	return defaultCode
}

// it's an error.
func (w *WithChoozleError) Error() string { return w.cause.Error() }

// it's also a wrapper.
func (w *WithChoozleError) Cause() error  { return w.cause }
func (w *WithChoozleError) Unwrap() error { return w.cause }

// it knows how to format itself.
func (w *WithChoozleError) Format(s fmt.State, verb rune) { errors.FormatError(w, s, verb) }

// SafeFormatter implements errors.SafeFormatter.
// Note: see the documentation of errbase.SafeFormatter for details
// on how to implement this. In particular beware of not emitting
// unsafe strings.
func (w *WithChoozleError) SafeFormatError(p errors.Printer) (next error) {
	if p.Detail() {
		p.Printf("http code: %d", w.code)
	}
	return w.cause
}

// it's an encodable error.
func encodeWithHTTPCode(_ context.Context, err error) (string, []string, proto.Message) {
	w := err.(*WithChoozleError)
	details := []string{fmt.Sprintf("HTTP %d", w.code)}
	payload := &EncodedHTTPCode{Code: uint32(w.code)}
	return "", details, payload
}

// it's a decodable error.
func decodeWithHTTPCode(
	_ context.Context, cause error, _ string, _ []string, payload proto.Message,
) error {
	wp := EncodedHTTPCode{}
	err := anypb.UnmarshalTo(payload.(*anypb.Any), &wp, proto.UnmarshalOptions{})
	if err != nil {
		panic(err)
	}
	return &WithChoozleError{cause: cause, code: ChoozleErrorCode(wp.Code)}
}

func init() {
	errbase.RegisterWrapperEncoder(errbase.GetTypeKey((*WithChoozleError)(nil)), encodeWithHTTPCode)
	errbase.RegisterWrapperDecoder(errbase.GetTypeKey((*WithChoozleError)(nil)), decodeWithHTTPCode)
}
