package suim

import "testing"

type formColSpanFixture struct {
	Address string `form_col_span:"2"`
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
