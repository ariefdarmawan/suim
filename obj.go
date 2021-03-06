package suim

import (
	"reflect"

	"github.com/sebarcode/codekit"
)

/*
type SuimModel interface {
	HandleChange(fieldName string, v1, v2, vOld interface{})
}
*/

type ObjMeta struct {
	Grid               GridSetting
	Form               FormSetting
	GoCustomValidator  string
	HandleChangeFields []string
}

var (
	sections = map[string][]FormSection{}
)

func autoFormSections(obj interface{}) ([]FormSection, error) {
	v := reflect.Indirect(reflect.ValueOf(obj))
	typeString := v.Type().String()

	res, has := sections[typeString]
	if has {
		return res, nil
	}

	_, fields, err := ObjToFields(obj)
	if err != nil {
		return res, err
	}

	lastSection := ""
	for _, f := range fields {
		if lastSection != f.Form.Section {
			if !codekit.HasMember(f.Form.Section, res) {
				res = append(res, FormSection{Title: f.Form.Section, Name: f.Form.Section, AutoCol: 1, ShowTitle: false})
				lastSection = f.Form.Section
			}
		}
	}

	sections[typeString] = res
	return res, nil
}
