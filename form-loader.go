package suim

import (
	"fmt"
	"reflect"

	"errors"

	"git.kanosolution.net/koloni/crowd"
)

func CreateFormConfig(obj interface{}) (*FormConfig, error) {
	vo := reflect.ValueOf(obj)
	to := vo.Type()

	v := vo.Elem()
	t := v.Type()
	cfgName := t.Name()
	cfg, ok := formConfigs[cfgName]
	if ok && cfg != nil {
		return cfg, nil
	}

	meta, fields, err := ObjToFields(obj)
	if err != nil {
		return nil, err
	}

	cfg = new(FormConfig)
	cfg.Setting = meta.Form

	//-- sections
	_, hasSectionFn := to.MethodByName("FormSections")
	if hasSectionFn {
		sectionFn := vo.MethodByName("FormSections")
		outs := sectionFn.Call([]reflect.Value{})
		if len(outs) != 1 {
			return nil, errors.New("invalid FormSections")
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("error generating sections. %v", r)
				}
			}()
			cfg.SectionGroups = outs[0].Interface().([]FormSectionGroup)
		}()
		if err != nil {
			return nil, err
		}
	}

	//-- assign auto section if no section found
	if len(cfg.SectionGroups) == 0 {
		if cfg.SectionGroups, err = autoFormSections(obj); err != nil {
			return nil, fmt.Errorf("error generating section. %s", err.Error())
		}
	}

	//-- filter shown field only
	formFields := []FormField{}
	for _, field := range fields {
		if field.FormElement == "show" {
			formFields = append(formFields, field.Form)
		}
	}

	_, err = ArrangeFormConfigFields(cfg, formFields...)
	if err != nil {
		return nil, err
	}

	mtx.Lock()
	defer mtx.Unlock()
	formConfigs[cfgName] = cfg

	return cfg, nil
}

func ArrangeFormConfigFields(cfg *FormConfig, formFields ...FormField) (*FormConfig, error) {
	var err error
	//-- for each eaction arrange the fields
	for gindex, sg := range cfg.SectionGroups {
		for idx, section := range sg.Sections {
			//-- get section fields
			var sectionFields []FormField
			if err = crowd.FromSlice(formFields).Filter(func(f FormField) bool {
				if f.Section == "" && section.Title == "General" {
					return true
				}
				return f.Section == section.Title
			}).Collect().Run(&sectionFields); err != nil {
				return nil, errors.New("fail retrieve fields for section " + section.Title + ". " + err.Error())
			}

			//-- calc before and after
			newFields := []FormField{}
			for _, f := range sectionFields {
				if f.SpaceBefore > 0 {
					for idx := 0; idx < f.SpaceBefore; idx++ {
						newFields = append(newFields, FormField{Kind: "space"})
					}
				}
				newFields = append(newFields, f)
				if f.SpaceAfter > 0 {
					for idx := 0; idx < f.SpaceAfter; idx++ {
						newFields = append(newFields, FormField{Kind: "space"})
					}
				}
			}
			sectionFields = newFields

			//-- assign row and col to empty field based on autocol
			if section.AutoCol > 0 {
				rowIndex := 1001
				colIndex := 0
				for idx, f := range sectionFields {
					if f.Row == 0 {
						// Match the frontend: ColSpan takes precedence over the
						// legacy numeric Width when consuming grid columns.
						columnSpan := f.ColSpan
						if columnSpan <= 0 {
							columnSpan = DefInt(f.Width, 1)
						}
						if columnSpan < 1 {
							columnSpan = 1
						}
						if columnSpan > section.AutoCol {
							columnSpan = section.AutoCol
						}
						if colIndex > 0 && colIndex+columnSpan > section.AutoCol {
							colIndex = 0
							rowIndex++
						}
						f.Row = rowIndex
						f.Col = colIndex + 1
						colIndex += columnSpan
						if colIndex >= section.AutoCol {
							colIndex = 0
							rowIndex++
						}
					}
					sectionFields[idx] = f
				}
			}

			//-- arrange the field
			type formRow struct {
				Row    int
				Fields []FormField
			}

			var arrangeFields [][]FormField
			if err = crowd.FromSlice(sectionFields).Group(func(f FormField) int {
				return f.Row
			}).Map(func(row int, fs []FormField) formRow {
				var sortedFs []FormField
				if e := crowd.FromSlice(fs).Sort(func(f1, f2 FormField) bool {
					return f1.Col < f2.Col
				}).Collect().Run(&sortedFs); e != nil {
					return formRow{row, fs}
				}
				return formRow{row, sortedFs}
			}).Sort(func(f1, f2 formRow) bool {
				return f1.Row < f2.Row
			}).Map(func(fr formRow) []FormField {
				return fr.Fields
			}).Collect().Run(&arrangeFields); err != nil {
				return nil, errors.New("fail processing section " + section.Title + ". " + err.Error())
			}
			section.Rows = arrangeFields

			sg.Sections[idx] = section
		}
		cfg.SectionGroups[gindex] = sg
	}

	return cfg, nil
}
