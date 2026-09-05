package suim

import (
	"reflect"
	"strconv"
	"strings"
)

func Label(name, kind string) string {
	if name == "ID" || name == "_id" {
		return name
	}
	// ID is an implementation suffix for reference fields (for example
	// ClientID). It should not be exposed in the default human-facing label.
	if strings.HasSuffix(name, "ID") {
		withoutID := strings.TrimSuffix(strings.TrimSuffix(name, "ID"), "_")
		if withoutID != "" {
			name = withoutID
		}
	}
	IDs := strings.Split(name, "ID")
	words := []string{}
	for idIndex, idText := range IDs {
		if idIndex > 0 {
			words = append(words, "ID")
		}
		if idText == "" {
			continue
		}
		texts := strings.Split(idText, "_")
		for idx, text := range texts {
			words = append(words, labelWord(text, kind, idx == 0)...)
		}
	}
	return strings.Join(words, " ")
}

func labelWord(name string, kind string, start bool) []string {
	words := []string{}
	word := ""
	for idx, c := range name {
		if idx == 0 {
			word += string(c)
		} else {
			if c >= 'A' && c <= 'Z' {
				words = append(words, word)
				word = string(c)
			} else {
				word += string(c)
			}
		}
	}
	words = append(words, word)

	switch kind {
	case "u":
		for idx, w := range words {
			words[idx] = strings.ToUpper(w)
		}

	case "l":
		for idx, w := range words {
			if idx == 0 && start {
				if len(w) == 1 {
					words[idx] = strings.ToUpper(w)
				} else {
					words[idx] = strings.ToUpper(string(w[0])) + w[1:]
				}
			} else {
				words[idx] = strings.ToLower(w)
			}
		}
	}

	return words
}

func LabelToID(txt, joiner, kind string) string {
	texts := strings.Split(txt, " ")

	switch kind {
	case "l":
		for idx, text := range texts {
			text = strings.ToLower(text)
			texts[idx] = text
		}

	case "u":
		for idx, text := range texts {
			text = strings.ToUpper(text)
			texts[idx] = text
		}

	case "c", "":
		for idx, text := range texts {
			if len(text) > 1 {
				text = strings.ToUpper(string(text[0])) + text[1:]
			} else {
				text = strings.ToUpper(text)
			}
			texts[idx] = text
		}
	}

	return strings.Join(texts, joiner)
}

func TagValue(tag reflect.StructTag, name string, def string) string {
	v := tag.Get(name)
	if v == "" {
		return def
	}
	return v
}

func TagExist(tag reflect.StructTag, name string) bool {
	_, b := tag.Lookup(name)
	return b
}

func DefTxt(txt, def string) string {
	if txt == "" {
		return def
	}
	return txt
}

func DefInt(txt string, def int) int {
	if txt == "" {
		return def
	}
	if v, e := strconv.Atoi(txt); e == nil {
		return v
	}
	return def
}

func DefSliceItem(a []string, index int, def string) string {
	if index >= len(a) {
		return def
	}

	return a[index]
}

func SetIfStruct(obj interface{}, el string, change bool, val interface{}) {
	if change {
		rv := reflect.ValueOf(obj)
		if rv.Kind() != reflect.Ptr {
			return
		}
		rv = rv.Elem().FieldByName(el)
		//target = val
		rv.Set(reflect.ValueOf(val))
	}
}

func SetIf(obj interface{}, change bool, val interface{}) {
	if change {
		rv := reflect.ValueOf(obj)
		if rv.Kind() != reflect.Ptr {
			return
		}
		rv = rv.Elem()
		//target = val
		rv.Set(reflect.ValueOf(val))
	}
}

func SplitNonEmpty(txt, splitter string) []string {
	parts := strings.Split(txt, splitter)
	res := []string{}
	for _, p := range parts {
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}
