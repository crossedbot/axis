package match

import (
	"reflect"

	"github.com/golang/mock/gomock"

	"github.com/crossedbot/axis/pkg/pins/models"
)

type pinStatus struct{ inner models.PinStatus }

func PinStatus(pin models.Pin) gomock.Matcher {
	return &pinStatus{models.PinStatus{Pin: pin}}
}

func (ps *pinStatus) Matches(x interface{}) bool {
	ps2, ok := x.(models.PinStatus)
	if ok {
		return ps2.Status == models.StatusInit.String() &&
			reflect.DeepEqual(ps.inner.Pin, ps2.Pin)
	}
	return false
}

func (ps *pinStatus) String() string {
	return "is of type PinStatus"
}
