package suim

import "testing"

type formColSpanFixture struct {
	Address string `form_col_span:"2"`
}

type formInsertFixture struct {
	CustomerID string `form_lookup:"/customer/find|_id|Name" form_insert_config:"/customer/formconfig" form_insert_size:"480" form_insert_api:"/customer/insert"`
	NoSizeID   string `form_lookup:"/customer/find|_id|Name" form_insert_config:"/customer/formconfig" form_insert_api:"/customer/insert"`
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
	if insert.FormInsertSize != 480 {
		t.Fatalf("expected form insert size 480, got %d", insert.FormInsertSize)
	}
	if fields[1].Form.FormInsertSize != 240 {
		t.Fatalf("expected default form insert size 240, got %d", fields[1].Form.FormInsertSize)
	}
}
