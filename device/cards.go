package device

import (
	"fmt"

	"codeberg.org/uhppoted/uhppoted-core/types"
	"codeberg.org/uhppoted/uhppoted-lib/uhppoted"
	"codeberg.org/uhppoted/uhppoted-mqtt/common"
)

func (d *Device) GetCards(impl uhppoted.IUHPPOTED, request []byte) (any, error) {
	body := struct {
		DeviceID *uhppoted.DeviceID `json:"device-id"`
	}{}

	if response, err := unmarshal(request, &body); err != nil {
		return response, err
	}

	if body.DeviceID == nil {
		return common.MakeError(uhppoted.StatusBadRequest, "Invalid/missing device ID", nil), fmt.Errorf("invalid/missing device ID")
	}

	rq := uhppoted.GetCardsRequest{
		DeviceID: *body.DeviceID,
	}

	response, err := impl.GetCards(rq)
	if err != nil {
		return common.MakeError(uhppoted.StatusInternalServerError, fmt.Sprintf("Could not retrieve cards from %d", *body.DeviceID), err), err
	}

	return response, nil
}

func (d *Device) DeleteCards(impl uhppoted.IUHPPOTED, request []byte) (any, error) {
	body := struct {
		DeviceID *uhppoted.DeviceID `json:"device-id"`
	}{}

	if response, err := unmarshal(request, &body); err != nil {
		return response, err
	}

	if body.DeviceID == nil {
		return common.MakeError(uhppoted.StatusBadRequest, "Invalid/missing device ID", nil), fmt.Errorf("invalid/missing device ID")
	}

	rq := uhppoted.DeleteCardsRequest{
		DeviceID: *body.DeviceID,
	}

	response, err := impl.DeleteCards(rq)
	if err != nil {
		return common.MakeError(uhppoted.StatusInternalServerError, fmt.Sprintf("Could not delete cards on %d", *body.DeviceID), err), err
	}

	return response, nil
}

func (d *Device) GetCard(impl uhppoted.IUHPPOTED, request []byte) (any, error) {
	body := struct {
		DeviceID   *uhppoted.DeviceID `json:"device-id"`
		CardNumber *uint32            `json:"card-number"`
	}{}

	if response, err := unmarshal(request, &body); err != nil {
		return response, err
	}

	if body.DeviceID == nil {
		return common.MakeError(uhppoted.StatusBadRequest, "Invalid/missing device ID", nil), fmt.Errorf("invalid/missing device ID")
	}

	if body.CardNumber == nil {
		return common.MakeError(uhppoted.StatusBadRequest, "Invalid/missing card number", nil), fmt.Errorf("invalid/missing card number")
	}

	rq := uhppoted.GetCardRequest{
		DeviceID:   *body.DeviceID,
		CardNumber: *body.CardNumber,
	}

	response, err := impl.GetCard(rq)
	if err != nil {
		return common.MakeError(uhppoted.StatusInternalServerError, fmt.Sprintf("Could not retrieve card %v from %d", *body.CardNumber, *body.DeviceID), err), err
	}

	type card struct {
		CardNumber uint32          `json:"card-number"`
		From       types.Date      `json:"start-date"`
		To         types.Date      `json:"end-date"`
		Doors      map[uint8]uint8 `json:"doors"`
		PIN        uint32          `json:"PIN,omitempty"`
		FirstCard  bool            `json:"first-card,omitempty"`
	}

	c := card{
		CardNumber: response.Card.CardNumber,
		From:       response.Card.From,
		To:         response.Card.To,
		Doors:      response.Card.Doors,
	}

	if d.WithPINs {
		c.PIN = uint32(response.Card.PIN)
	}

	if d.WithFirstCard {
		c.FirstCard = response.Card.FirstCard.Door1 || response.Card.FirstCard.Door2 || response.Card.FirstCard.Door3 || response.Card.FirstCard.Door4
	}

	return struct {
		DeviceID uint32 `json:"device-id"`
		Card     card   `json:"card"`
	}{
		DeviceID: uint32(response.DeviceID),
		Card:     c,
	}, nil
}

func (d *Device) PutCard(impl uhppoted.IUHPPOTED, request []byte) (any, error) {
	type card struct {
		CardNumber uint32        `json:"card-number"`
		From       types.Date    `json:"start-date"`
		To         types.Date    `json:"end-date"`
		Doors      map[uint8]any `json:"doors"`
		PIN        uint32        `json:"PIN,omitempty"`
		FirstCard  bool          `json:"first-card,omitempty"`
	}

	body := struct {
		DeviceID *uhppoted.DeviceID `json:"device-id"`
		Card     *card              `json:"card"`
	}{}

	if response, err := unmarshal(request, &body); err != nil {
		return response, err
	}

	if body.DeviceID == nil {
		return common.MakeError(uhppoted.StatusBadRequest, "Invalid/missing device ID", nil), fmt.Errorf("invalid/missing device ID")
	}

	if body.Card == nil {
		return common.MakeError(uhppoted.StatusBadRequest, "Invalid/missing card", nil), fmt.Errorf("invalid/missing card")
	}

	if body.Card.PIN > 999999 {
		return common.MakeError(uhppoted.StatusBadRequest,
			"Invalid card PIN",
			fmt.Errorf("PIN %v out of range", body.Card.PIN)), fmt.Errorf("invalid card PIN (%v)", body.Card.PIN)
	}

	deviceID := uint32(*body.DeviceID)
	c := types.Card{
		CardNumber: body.Card.CardNumber,
		From:       body.Card.From,
		To:         body.Card.To,
		Doors:      map[uint8]uint8{1: 0, 2: 0, 3: 0, 4: 0},
	}

	for _, k := range []uint8{1, 2, 3, 4} {
		switch v := body.Card.Doors[k].(type) {
		case bool:
			if v {
				c.Doors[k] = 1
			}

		case int:
			if v >= 2 && v < 254 {
				c.Doors[k] = uint8(v)
			}

		case float64:
			if int(v) >= 2 && int(v) < 254 {
				c.Doors[k] = uint8(v)
			}
		}
	}

	if d.WithPINs {
		c.PIN = types.PIN(body.Card.PIN)
	}

	if d.WithFirstCard && body.Card.FirstCard {
		c.FirstCard = types.FirstCardPrivileges{
			Door1: c.Doors[1] != 0,
			Door2: c.Doors[2] != 0,
			Door3: c.Doors[3] != 0,
			Door4: c.Doors[4] != 0,
		}
	}

	if ok, err := impl.PutCard(deviceID, c); err != nil {
		return common.MakeError(uhppoted.StatusInternalServerError,
			fmt.Sprintf("Could not store card %v to %d", c.CardNumber, deviceID),
			err), err
	} else if !ok {
		return common.MakeError(
				uhppoted.StatusInternalServerError,
				fmt.Sprintf("Failed to store card %v to %d", c.CardNumber, deviceID),
				fmt.Errorf("put-card returned %v", ok)),
			fmt.Errorf("put-card for card %v:%v returned %v", deviceID, c.CardNumber, ok)
	}

	PIN := uint32(0)
	if d.WithPINs {
		PIN = uint32(c.PIN)
	}

	firstcard := false
	if d.WithFirstCard {
		firstcard = c.FirstCard.Door1 || c.FirstCard.Door2 || c.FirstCard.Door3 || c.FirstCard.Door4
	}

	response := struct {
		DeviceID uhppoted.DeviceID `json:"device-id"`
		Card     card              `json:"card"`
	}{
		DeviceID: uhppoted.DeviceID(deviceID),
		Card: card{
			CardNumber: c.CardNumber,
			From:       c.From,
			To:         c.To,
			Doors:      map[uint8]any{1: false, 2: false, 3: false, 4: false},
			PIN:        PIN,
			FirstCard:  firstcard,
		},
	}

	for _, k := range []uint8{1, 2, 3, 4} {
		v := c.Doors[k]
		switch v {
		case 0:
			response.Card.Doors[k] = false
		case 1:
			response.Card.Doors[k] = true
		default:
			response.Card.Doors[k] = v
		}
	}

	return response, nil
}

func (d *Device) DeleteCard(impl uhppoted.IUHPPOTED, request []byte) (any, error) {
	body := struct {
		DeviceID   *uhppoted.DeviceID `json:"device-id"`
		CardNumber *uint32            `json:"card-number"`
	}{}

	if response, err := unmarshal(request, &body); err != nil {
		return response, err
	}

	if body.DeviceID == nil {
		return common.MakeError(uhppoted.StatusBadRequest, "Invalid/missing device ID", nil), fmt.Errorf("invalid/missing device ID")
	}

	if body.CardNumber == nil {
		return common.MakeError(uhppoted.StatusBadRequest, "Invalid/missing card number", nil), fmt.Errorf("invalid/missing card number")
	}

	rq := uhppoted.DeleteCardRequest{
		DeviceID:   *body.DeviceID,
		CardNumber: *body.CardNumber,
	}

	response, err := impl.DeleteCard(rq)
	if err != nil {
		return common.MakeError(uhppoted.StatusInternalServerError, fmt.Sprintf("Could not store card %v to %d", *body.CardNumber, *body.DeviceID), err), err
	}

	return response, nil
}
