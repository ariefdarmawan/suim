package suim

import (
	"reflect"
	"testing"
)

func TestArrangeFormConfigFieldsAccountsForColumnSpans(t *testing.T) {
	for _, tc := range []struct {
		name    string
		columns int
		fields  []FormField
		rows    [][]string
	}{
		{
			name: "full width reason followed by two columns", columns: 2,
			fields: []FormField{{Field: "Reason", ColSpan: 2}, {Field: "CancelledAt"}, {Field: "CancelledBy"}},
			rows:   [][]string{{"Reason"}, {"CancelledAt", "CancelledBy"}},
		},
		{
			name: "wide field moves to next row when it does not fit", columns: 2,
			fields: []FormField{{Field: "First"}, {Field: "Wide", ColSpan: 2}, {Field: "Next"}, {Field: "Last"}},
			rows:   [][]string{{"First"}, {"Wide"}, {"Next", "Last"}},
		},
		{
			name: "mixed spans fit in three columns", columns: 3,
			fields: []FormField{{Field: "Wide", ColSpan: 2}, {Field: "Single"}, {Field: "Next"}},
			rows:   [][]string{{"Wide", "Single"}, {"Next"}},
		},
		{
			name: "legacy width remains a fallback", columns: 2,
			fields: []FormField{{Field: "Wide", Width: "2"}, {Field: "Next"}, {Field: "Last"}},
			rows:   [][]string{{"Wide"}, {"Next", "Last"}},
		},
		{
			name: "colspan takes precedence over legacy width", columns: 2,
			fields: []FormField{{Field: "First", ColSpan: 1, Width: "2"}, {Field: "Second"}},
			rows:   [][]string{{"First", "Second"}},
		},
		{
			name: "non-numeric widths use one column", columns: 2,
			fields: []FormField{{Field: "First", Width: "100%"}, {Field: "Second", Width: "-1"}},
			rows:   [][]string{{"First", "Second"}},
		},
		{
			name: "oversized span consumes one full row", columns: 2,
			fields: []FormField{{Field: "Wide", ColSpan: 3}, {Field: "Next"}, {Field: "Last"}},
			rows:   [][]string{{"Wide"}, {"Next", "Last"}},
		},
		{
			name: "explicit positions remain unchanged", columns: 2,
			fields: []FormField{{Field: "Manual", Row: 1, Col: 2}, {Field: "First"}, {Field: "Second"}},
			rows:   [][]string{{"Manual"}, {"First", "Second"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &FormConfig{SectionGroups: []FormSectionGroup{{Sections: []FormSection{{Title: "General", AutoCol: tc.columns}}}}}
			if _, err := ArrangeFormConfigFields(cfg, tc.fields...); err != nil {
				t.Fatal(err)
			}
			var rows [][]string
			for _, row := range cfg.SectionGroups[0].Sections[0].Rows {
				var fields []string
				for _, field := range row {
					fields = append(fields, field.Field)
				}
				rows = append(rows, fields)
			}
			if !reflect.DeepEqual(rows, tc.rows) {
				t.Fatalf("rows = %v, want %v", rows, tc.rows)
			}
		})
	}
}

type formColSpanFixture struct {
	Address string `form_col_span:"2"`
}

type autoColCancellationFixture struct {
	Reason      string `form_section:"Cancellation" form_section_auto_col:"2" form_col_span:"2"`
	CancelledAt string `form_section:"Cancellation"`
	CancelledBy string `form_section:"Cancellation"`
}

func TestAutoColCancellationLayoutFromTags(t *testing.T) {
	cfg, err := CreateFormConfig(new(autoColCancellationFixture))
	if err != nil {
		t.Fatal(err)
	}
	section := cfg.SectionGroups[0].Sections[0]
	if section.AutoCol != 2 || len(section.Rows) != 2 {
		t.Fatalf("expected a two-column section with two rows: %#v", section)
	}
	if len(section.Rows[0]) != 1 || section.Rows[0][0].Field != "Reason" || section.Rows[0][0].ColSpan != 2 {
		t.Fatalf("expected Reason to occupy the entire first row: %#v", section.Rows[0])
	}
	if len(section.Rows[1]) != 2 || section.Rows[1][0].Field != "CancelledAt" || section.Rows[1][1].Field != "CancelledBy" {
		t.Fatalf("expected cancellation time and user in the second row: %#v", section.Rows[1])
	}
}

type formInsertFixture struct {
	CustomerID string `form_lookup:"/customer/find|_id|Name" form_insert_config:"/customer/formconfig" form_insert_size:"480" form_insert_api:"/customer/insert"`
	NoSizeID   string `form_lookup:"/customer/find|_id|Name" form_insert_config:"/customer/formconfig" form_insert_api:"/customer/insert"`
	WideID     string `form_lookup:"/customer/find|_id|Name" form_insert_config:"/customer/formconfig" form_insert_size:"80%" form_insert_api:"/customer/insert"`
}

func TestFormColSpanTag(t *testing.T) {
	_, fields, err := ObjToFields(new(formColSpanFixture))
	if err != nil {
		t.Fatalf("load form fields: %v", err)
	}
	if fields[0].Form.ColSpan != 2 {
		t.Fatalf("expected form_col_span to be 2, got %d", fields[0].Form.ColSpan)
	}
}

func TestFormInsertTags(t *testing.T) {
	_, fields, err := ObjToFields(new(formInsertFixture))
	if err != nil {
		t.Fatalf("load form fields: %v", err)
	}

	insert := fields[0].Form
	if insert.FormInsertConfig != "/customer/formconfig" {
		t.Fatalf("expected form insert config, got %q", insert.FormInsertConfig)
	}
	if insert.FormInsertAPI != "/customer/insert" {
		t.Fatalf("expected form insert API, got %q", insert.FormInsertAPI)
	}
	if insert.FormInsertSize != "480" {
		t.Fatalf("expected form insert size 480, got %q", insert.FormInsertSize)
	}
	if fields[1].Form.FormInsertSize != "240" {
		t.Fatalf("expected default form insert size 240, got %q", fields[1].Form.FormInsertSize)
	}
	if fields[2].Form.FormInsertSize != "80%" {
		t.Fatalf("expected CSS form insert size 80%%, got %q", fields[2].Form.FormInsertSize)
	}
}
