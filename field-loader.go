package suim

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sebarcode/codekit"
)

type objConfig struct {
	Obj    *ObjMeta
	Fields []Field
}

var (
	objConfigs = map[string]objConfig{}
)

func ObjToFields(obj interface{}) (*ObjMeta, []Field, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, []Field{}, errors.New("object should be a struct or pointer of a struct")
	}
	t := v.Type()

	// if already inside repo then get from them
	if res, has := objConfigs[t.String()]; has {
		return res.Obj, res.Fields, nil
	}

	// if not processing it and save into memory repo at the end
	fieldNum := v.NumField()
	fields := []Field{}
	meta := new(ObjMeta)

	gs := GridSetting{}

	fs := FormSetting{
		ShowTitle:        false,
		HideButtons:      false,
		HideEditButton:   false,
		HideSubmitButton: false,
		HideCancelButton: false,
		InitialMode:      "edit",
		SubmitText:       "Save",
		AutoCol:          1,
		SectionDirection: "col",
		SectionSize:      1,
	}

	for i := 0; i < fieldNum; i++ {
		ft := t.Field(i)
		tag := ft.Tag
		alias := tag.Get(codekit.TagName())
		if alias == "-" {
			continue
		}
		if alias == "" {
			alias = ft.Name
		}

		field, e := toField(ft)
		if e != nil {
			return nil, []Field{}, fmt.Errorf("%s: %s", alias, e)
		}
		fields = append(fields, field)

		//-- FormSetting
		SetIfStruct(&fs, "IDField", fs.IDField == "" && TagExist(tag, "key"), alias)
		SetIfStruct(&fs, "Title", fs.Title == "", Label(TagValue(tag, "obj_title", t.Name()), ""))
		SetIfStruct(&fs, "ShowTitle", TagExist(tag, "form_hide_title"), TagValue(tag, "form_hide_title", "") == "1")
		SetIfStruct(&fs, "HideButtons", TagExist(tag, "form_hide_buttons"), TagValue(tag, "form_hide_buttons", "") == "1")
		SetIfStruct(&fs, "HideEditButton", TagExist(tag, "form_hide_edit_button"), TagValue(tag, "form_hide_edit_button", "") == "1")
		SetIfStruct(&fs, "HideSubmitButton", TagExist(tag, "form_hide_submit_button"), TagValue(tag, "form_hide_submit_button", "") == "1")
		SetIfStruct(&fs, "HideCancelButton", TagExist(tag, "form_hide_cancel_button"), TagValue(tag, "form_hide_cancel_button", "") == "1")
		SetIfStruct(&fs, "InitialMode", TagExist(tag, "form_initial_mode"), TagValue(tag, "form_initial_mode", "edit"))
		SetIfStruct(&fs, "SubmitText", TagExist(tag, "form_submit_text"), TagValue(tag, "form_submit_text", "Save"))
		SetIfStruct(&fs, "AutoCol", TagExist(tag, "form_auto_col"), DefInt(TagValue(tag, "form_auto_col", "1"), 1))
		SetIfStruct(&fs, "SectionDirection", TagExist(tag, "form_section_direction"), TagValue(tag, "form_section_direction", ""))
		SetIfStruct(&fs, "SectionSize", TagExist(tag, "form_section_size"), DefInt(TagValue(tag, "form_section_size", "1"), 1))

		//-- GridSetting
		SetIfStruct(&gs, "IDField", gs.IDField == "" && TagValue(tag, "key", "") == "1", alias)
		if TagValue(tag, "grid_keyword", "0") == "1" {
			gs.KeywordFields = append(gs.KeywordFields, alias)
		}
		if TagValue(tag, "grid_sortable", "0") == "1" {
			gs.SortableFields = append(gs.SortableFields, alias)
		}
		if TagValue(tag, "grid_searchable", "0") == "1" {
			gs.SearchableFields = append(gs.SearchableFields, alias)
		}

		//-- main obj
		SetIfStruct(meta, "GoCustomValidator", meta.GoCustomValidator == "", TagValue(tag, "obj_go_validator", GoCustomValidator))
	}

	if len(gs.KeywordFields) == 0 {
		gs.KeywordFields = []string{"_id", "Name"}
	}
	if len(gs.SortableFields) == 0 {
		gs.SortableFields = []string{"_id"}
	}
	meta.Grid = gs
	meta.Form = fs

	mtx.Lock()
	defer mtx.Unlock()

	objCfg := objConfig{Obj: meta, Fields: fields}
	objConfigs[t.String()] = objCfg

	return meta, fields, nil
}

func toField(rt reflect.StructField) (Field, error) {
	f := Field{}
	f.Field = rt.Name
	f.DataType = rt.Type.String()

	tag := rt.Tag

	f.GridElement = TagValue(tag, "grid", "show")
	f.FormElement = TagValue(tag, "form", "show")

	//if f.FormElement != "hide" {
	form := FormField{}
	form.AllowAdd = TagValue(tag, "form_allow_add", "") == "1"
	form.Decimal = DefInt(TagValue(tag, "form_decimal", "0"), 0)
	form.DateFormat = TagValue(tag, "form_date_format", "DD-MMM-YYYY hh:mm:ss Z")
	form.Field = TagValue(tag, codekit.TagName(), rt.Name)
	form.Hide = f.FormElement != "show"
	pos := strings.Split(TagValue(tag, "form_pos", ","), ",")
	rowStr := DefTxt(pos[0], "0")
	colStr := "0"
	if len(pos) > 1 {
		colStr = DefTxt(pos[1], "0")
	}
	form.Row, _ = strconv.Atoi(rowStr)
	form.Col, _ = strconv.Atoi(colStr)

	form.Section = TagValue(tag, "form_section", "General")
	form.SectionWidth = TagValue(tag, "form_section_width", "")
	form.SectionShowTitle = TagValue(tag, "form_section_show_title", "0") == "1"
	form.SectionAutoCol = DefInt(TagValue(tag, "form_section_auto_col", "1"), 1)
	form.Unit = TagValue(tag, "form_unit", "")

	form.Kind = TagValue(tag, "form_kind", "")
	if form.Kind == "" {
		//fmt.Println(f.Field, form.Kind, f.DataType)
		switch f.DataType {
		case "int", "float32", "float64":
			form.Kind = "number"
		case "time.Time", "*time.Time":
			form.Kind = "datetime"
		case "bool":
			form.Kind = "checkbox"
		case "[]string":
			form.Multiple = true
		default:
			form.Kind = "text"
		}
	}

	form.Disable = TagExist(tag, "form_disable")
	form.FixDetail = TagExist(tag, "form_fix_detail")
	form.FixTitle = TagExist(tag, "form_fix_title")
	form.Hint = TagValue(tag, "form_hint", "")
	items := strings.Split(TagValue(tag, "form_items", ""), "|")
	form.Items = []FormListItem{}
	for _, item := range items {
		parts := strings.Split(item, ":")
		if parts[0] == "" {
			continue
		}
		if len(parts) > 1 {
			form.Items = append(form.Items, FormListItem{Key: parts[0], Text: parts[1]})
		} else if len(parts) == 1 {
			form.Items = append(form.Items, FormListItem{Key: parts[0], Text: parts[0]})
		}
	}
	form.LabelField = TagValue(tag, "obj_label_field", "")
	form.Label = TagValue(tag, "form_label", TagValue(tag, "label", Label(rt.Name, "l")))
	SetIfStruct(form, "Multiple", !form.Multiple && TagExist(tag, "form_multiple"), TagValue(tag, "form_multiple", "") == "1")
	form.UseList = len(form.Items) > 0 || TagExist(tag, "form_use_list")
	if TagValue(tag, "form_lookup", "") != "" {
		form.UseList = true
		form.UseLookup = true
		lookups := strings.Split(TagValue(tag, "form_lookup", ""), "|")
		if len(lookups) < 2 {
			return f, errors.New("lookup should contains at least 2 elements: url and fieldof key")
		}
		if lookups[0] == "" {
			return f, errors.New("lookup url can not be blank")
		}
		form.LookupUrl = lookups[0]
		form.LookupKey = lookups[1]
		form.LookupLabels = []string{form.LookupKey}

		if len(lookups) > 2 {
			form.LookupLabels = SplitNonEmpty(lookups[2], ",")
		}

		if len(lookups) > 3 {
			form.LookupFormat1 = lookups[3]
			form.LookupFormat2 = lookups[3]
		}

		if len(lookups) > 4 {
			form.LookupFormat2 = lookups[4]
		}

		form.LookupSearchs = SplitNonEmpty(TagValue(tag, "form_lookup_search", ""), ",")
		if len(form.LookupSearchs) == 0 {
			form.LookupSearchs = form.LookupLabels
		}
	}
	form.Placeholder = TagValue(tag, "form_placeholder", form.Label)
	lengths := strings.Split(TagValue(tag, "form_length", "0,999"), ",")
	form.MinLength = DefInt(DefSliceItem(lengths, 0, "0"), 0)
	form.MaxLength = DefInt(DefSliceItem(lengths, 1, "9999"), 9999)
	form.MultiRow = DefInt(TagValue(tag, "form_multi_row", "1"), 1)
	form.Required = TagExist(tag, "form_required")
	form.ReadOnly = TagValue(tag, "form_read_only", "0") == "1"
	form.ReadOnlyOnEdit = TagValue(tag, "form_read_only_edit", "0") == "1"
	form.ReadOnlyOnNew = TagValue(tag, "form_read_only_new", "0") == "1"
	form.ShowDetail = TagExist(tag, "form_hide_detail")
	form.ShowHint = TagExist(tag, "form_hide_hint")
	form.ShowTitle = TagExist(tag, "form_hide_title")
	form.Width = TagValue(tag, "form_width", "")
	form.SpaceBefore = DefInt(TagValue(tag, "form_space_before", "0"), 0)
	form.SpaceAfter = DefInt(TagValue(tag, "form_space_after", "0"), 0)
	f.Form = form
	//}

	if f.GridElement != "hide" {
		grid := GridField{}
		grid.Field = f.Form.Field
		grid.Kind = f.Form.Kind
		grid.Halign = TagValue(tag, "grid_halign", "start")
		grid.Valign = TagValue(tag, "grid_valign", "start")
		grid.Label = TagValue(tag, "grid_label", TagValue(tag, "label", Label(rt.Name, "l")))
		grid.LabelField = TagValue(tag, "obj_label_field", "")
		grid.Length = DefInt(TagValue(tag, "grid_length", "0"), 0)
		grid.Pos = DefInt(TagValue(tag, "grid_pos", "0"), 0)
		grid.Width = TagValue(tag, "grid_width", "")
		grid.ReadType = f.GridElement
		grid.Decimal = f.Form.Decimal
		grid.DateFormat = f.Form.DateFormat
		grid.Unit = f.Form.Unit
		f.Grid = grid
	}

	return f, nil
}
