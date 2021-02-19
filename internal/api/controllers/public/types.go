package public

import (
	"time"
)

// workaround for https://github.com/deepmap/oapi-codegen/issues/187
func (this CreatedAt) MarshalJSON() ([]byte, error) {
	return time.Time(this).MarshalJSON()
}

// workaround for https://github.com/deepmap/oapi-codegen/issues/187
func (this UpdatedAt) MarshalJSON() ([]byte, error) {
	return time.Time(this).MarshalJSON()
}
