package suim

type GridField struct {
	Field      string `json:"field"`
	Kind       string `json:"kind"`
	Label      string `json:"label"`
	Halign     string `json:"halign"`
	Valign     string `json:"valign"`
	LabelField string `json:"labelField"`
	Length     int    `json:"length"`
	Width      string `json:"width"`
	Pos        int    `json:"pos"`
	ReadType   string `json:"readType"`

	//formatting attr
	Decimal    int    `json:"decimal"`
	DateFormat string `json:"dateFormat"`
	Unit       string `json:"unit"`

	Form FormField `json:"input"`
}

type GridSetting struct {
	IDField          string   `json:"idField"`
	KeywordFields    []string `json:"keywordFields"`
	SortableFields   []string `json:"sortable"`
	SearchableFields []string `json:"searchable"`
}

type GridConfig struct {
	Setting GridSetting `json:"setting"`
	Fields  []GridField `json:"fields"`
}
