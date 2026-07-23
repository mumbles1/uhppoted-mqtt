package acl

import (
	"encoding/json"
	"fmt"

	api "codeberg.org/uhppoted/uhppoted-lib/acl"
	"codeberg.org/uhppoted/uhppoted-lib/uhppoted"
	"codeberg.org/uhppoted/uhppoted-mqtt/common"
)

func (a *ACL) Revoke(impl uhppoted.IUHPPOTED, request []byte) (any, error) {
	body := struct {
		CardNumber *uint32  `json:"card-number"`
		Doors      []string `json:"doors"`
	}{}

	if err := json.Unmarshal(request, &body); err != nil {
		return common.MakeError(StatusBadRequest, "Cannot parse request", err), fmt.Errorf("%w: %v", uhppoted.ErrBadRequest, err)
	}

	if body.CardNumber == nil {
		return common.MakeError(StatusBadRequest, "Missing/invalid card number", nil), fmt.Errorf("missing/invalid card number")
	}

	err := api.Revoke(a.UHPPOTE, a.Devices, *body.CardNumber, body.Doors)
	if err != nil {
		return common.MakeError(StatusInternalServerError, err.Error(), nil), err
	}

	return struct {
		Revoked bool `json:"revoked"`
	}{
		Revoked: true,
	}, nil
}
