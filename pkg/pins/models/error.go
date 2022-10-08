package models

import (
	"fmt"
)

const (
	// Error Codes
	ErrMaxPinLimitCode = iota + 1000
	ErrMaxNameLimitCode
	ErrFailedConversionCode
	ErrUnknownStatusStringCode
	ErrUnknownTextMatchStringCode
	ErrUnknownSortStringCode
	ErrUnknownFieldStringCode
	ErrRequiredParamCode
	ErrUnauthorizedCode
	ErrProcessingRequestCode
	ErrNotFoundCode
)

var ErrorCodeStrings = map[int]string{
	ErrMaxPinLimitCode:            "MAX_PIN_LIMIT",
	ErrMaxNameLimitCode:           "MAX_NAME_LIMIT",
	ErrFailedConversionCode:       "FAILED_CONVERSION",
	ErrUnknownStatusStringCode:    "UNKNOWN_STATUS_STRING",
	ErrUnknownSortStringCode:      "UNKNOWN_SORT_STRING",
	ErrUnknownFieldStringCode:     "UNKNOWN_FIELD_STRING",
	ErrUnknownTextMatchStringCode: "UNKNOWN_MATCH_STRING",
	ErrRequiredParamCode:          "REQUIRED_PARAM",
	ErrUnauthorizedCode:           "UNAUTHORIZED",
	ErrProcessingRequestCode:      "PROCESSING_REQUEST",
	ErrNotFoundCode:               "NOT_FOUND",
}

type Error struct {
	Reason  string
	Details string
}

func (err Error) Error() string {
	return fmt.Sprintf("%s: %s", err.Reason, err.Details)
}

type Failure struct {
	Error Error
}

func NewFailure(reasonCode int, details string) Failure {
	return Failure{
		Error: Error{
			Reason:  ErrorCodeStrings[reasonCode],
			Details: details,
		},
	}
}
