package acl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/template"
	"time"

	"codeberg.org/uhppoted/uhppoted-core/types"
	api "codeberg.org/uhppoted/uhppoted-lib/acl"
	"codeberg.org/uhppoted/uhppoted-lib/uhppoted"
	"codeberg.org/uhppoted/uhppoted-mqtt/common"
)

var templates = struct {
	report string
}{
	report: `ACL DIFF REPORT {{ .DateTime }}
{{range $id,$value := .Diffs}}
  DEVICE {{ $id }}{{if or $value.Updated $value.Added $value.Deleted}}{{else}} OK{{end}}{{if $value.Updated}}
    Incorrect:  {{range $value.Updated}}{{.}}
                {{end}}{{end}}{{if $value.Added}}
    Missing:    {{range $value.Added}}{{.}}
                {{end}}{{end}}{{if $value.Deleted}}
    Unexpected: {{range $value.Deleted}}{{.}}
                {{end}}{{end}}{{end}}
`,
}

type Report struct {
	DateTime types.DateTime
	Diffs    map[uint32]diff
}

type diff struct {
	Unchanged []string
	Updated   []string
	Added     []string
	Deleted   []string
}

func (a *ACL) Compare(impl uhppoted.IUHPPOTED, request []byte) (any, error) {
	body := struct {
		URL struct {
			ACL    *string `json:"acl"`
			Report *string `json:"report"`
		} `json:"url"`
		MimeType string `json:"mime-type"`
	}{
		MimeType: "application/tar+gzip",
	}

	if err := json.Unmarshal(request, &body); err != nil {
		return common.MakeError(StatusBadRequest, "Cannot parse request", err), fmt.Errorf("%w: %v", uhppoted.ErrBadRequest, err)
	}

	if body.URL.ACL == nil {
		return common.MakeError(StatusBadRequest, "Missing/invalid download URL", nil), fmt.Errorf("missing/invalid download URL")
	}

	uri, err := url.Parse(*body.URL.ACL)
	if err != nil {
		return common.MakeError(StatusBadRequest, "Missing/invalid download URL", err), fmt.Errorf("invalid download URL '%v' (%w)", body.URL.ACL, err)
	}

	if body.URL.Report == nil {
		return common.MakeError(StatusBadRequest, "Missing/invalid report URL", nil), fmt.Errorf("missing/invalid report URL")
	}

	rpt, err := url.Parse(*body.URL.Report)
	if err != nil {
		return common.MakeError(StatusBadRequest, "Missing/invalid report URL", err), fmt.Errorf("invalid report URL '%v' (%w)", body.URL.Report, err)
	}

	acl, err := a.fetch("acl:compare", uri.String(), body.MimeType)
	if err != nil {
		return common.MakeError(StatusBadRequest, "Error downloading ACL", err), err
	}

	if acl == nil {
		return common.MakeError(StatusBadRequest, "Error downloading ACL", nil), fmt.Errorf("Download return nil ACL")
	}

	for k, l := range *acl {
		infof("acl:compare", "%v  Retrieved %v records", k, len(l))
	}

	current, errors := api.GetACL(a.UHPPOTE, a.Devices)
	if len(errors) > 0 {
		err := fmt.Errorf("%v", errors)
		return common.MakeError(StatusInternalServerError, "Error retrieving current ACL", err), err
	}

	diff, err := api.Compare(current, *acl)
	if err != nil {
		return common.MakeError(StatusInternalServerError, "Error comparing current and downloaded ACL's", err), err
	}

	var w strings.Builder
	if err := a.report(diff, templates.report, &w); err != nil {
		return common.MakeError(StatusInternalServerError, "Error generating ACL compare report", err), err
	}

	filename := time.Now().Format("acl-2006-01-02T150405.rpt")
	if err = a.store("acl:compare", rpt.String(), filename, []byte(w.String())); err != nil {
		return common.MakeError(StatusBadRequest, "Error uploading report", err), err
	}

	summary := map[uint32]struct {
		Unchanged  int `json:"unchanged"`
		Different  int `json:"different"`
		Missing    int `json:"missing"`
		Extraneous int `json:"extraneous"`
	}{}

	for k, v := range diff {
		infof("acl:compare", "%v  SUMMARY  unchanged:%v  different:%v  missing:%v  extraneous:%v", k, len(v.Unchanged), len(v.Updated), len(v.Added), len(v.Deleted))

		summary[k] = struct {
			Unchanged  int `json:"unchanged"`
			Different  int `json:"different"`
			Missing    int `json:"missing"`
			Extraneous int `json:"extraneous"`
		}{
			Unchanged:  len(v.Unchanged),
			Different:  len(v.Updated),
			Missing:    len(v.Added),
			Extraneous: len(v.Deleted),
		}
	}

	return struct {
		URL    string `json:"url"`
		Report map[uint32]struct {
			Unchanged  int `json:"unchanged"`
			Different  int `json:"different"`
			Missing    int `json:"missing"`
			Extraneous int `json:"extraneous"`
		} `json:"report"`
	}{
		URL:    rpt.String(),
		Report: summary,
	}, nil
}

func (a *ACL) report(diffs map[uint32]api.Diff, format string, w io.Writer) error {
	t, err := template.New("report").Parse(format)
	if err != nil {
		return err
	}

	permission := func(p uint8) string {
		switch {
		case p == 0:
			return "N"

		case p == 1:
			return "Y"

		case p >= 2 && p <= 254:
			return fmt.Sprintf("%v", p)

		default:
			return "N"
		}
	}

	date := func(d types.Date) string {
		if d.IsZero() {
			return fmt.Sprintf("%-10v", "-")
		} else {
			return fmt.Sprintf("%-10v", d)
		}
	}

	stringify := func(card types.Card) string {
		s := fmt.Sprintf("%-8v %v %v %v %v %v %v",
			card.CardNumber,
			date(card.From),
			date(card.To),
			permission(card.Doors[1]),
			permission(card.Doors[2]),
			permission(card.Doors[3]),
			permission(card.Doors[4]))

		if a.WithPINs {
			if card.PIN == 0 || card.PIN > 999999 {
				s = fmt.Sprintf("%v -", s)
			} else {
				s = fmt.Sprintf("%v %-6v", s, card.PIN)
			}
		}

		if a.WithFirstCard {
			if !a.WithPINs {
				s = fmt.Sprintf("%v -", s)
			}

			if card.FirstCard.IsZero() {
				s = fmt.Sprintf("%v -", s)
			} else {
				s = fmt.Sprintf("%v %v", s, card.FirstCard)
			}
		}

		return s
	}

	data := map[uint32]diff{}

	for k, v := range diffs {
		unchanged := []string{}
		updated := []string{}
		added := []string{}
		deleted := []string{}

		for _, card := range v.Unchanged {
			unchanged = append(unchanged, stringify(card))
		}

		for _, card := range v.Updated {
			updated = append(updated, stringify(card))
		}

		for _, card := range v.Added {
			added = append(added, stringify(card))
		}

		for _, card := range v.Deleted {
			deleted = append(deleted, stringify(card))
		}

		data[k] = diff{
			Unchanged: unchanged,
			Updated:   updated,
			Added:     added,
			Deleted:   deleted,
		}
	}

	rpt := Report{
		DateTime: types.DateTime(time.Now()),
		Diffs:    data,
	}

	return t.Execute(w, rpt)
}
