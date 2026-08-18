package suim_test

import (
	"testing"
	"time"

	"github.com/ariefdarmawan/suim"
	"github.com/smartystreets/goconvey/convey"
)

func TestLabel(t *testing.T) {
	type labelTest struct {
		Kind   string
		Result string
	}
	data := map[string]labelTest{
		"FullName":     {Kind: "", Result: "Full Name"},
		"EmailAddress": {Kind: "l", Result: "Email address"},
		"FamilyName":   {Kind: "u", Result: "FAMILY NAME"},
		"First_Name":   {Kind: "l", Result: "First name"},
		"RememberMe":   {Kind: "", Result: "Remember Me"},
		"EmailID":      {Kind: "", Result: "Email ID"},
		"_id":          {Kind: "", Result: "_id"},
		"ID":           {Kind: "", Result: "ID"},
	}

	for k, d := range data {
		label := suim.Label(k, d.Kind)
		if label != d.Result {
			t.Errorf("Label for %s should be %s, got %s", k, d.Result, label)
		}
	}
}

func TestLabelToID(t *testing.T) {
	type labelIDTest struct {
		Joiner string
		Kind   string
		Result string
	}
	data := map[string]labelIDTest{
		"Full Name":     {Kind: "", Joiner: "", Result: "FullName"},
		"Email address": {Kind: "l", Joiner: "_", Result: "email_address"},
		"Country name":  {Kind: "u", Joiner: "-", Result: "COUNTRY-NAME"},
	}

	for k, d := range data {
		id := suim.LabelToID(k, d.Joiner, d.Kind)
		if id != d.Result {
			t.Errorf("ID for %s should be %s, got %s", k, d.Result, id)
		}
	}
}

func TestSuimForm(t *testing.T) {
	convey.Convey("Form", t, func() {
		cfgMain, e := suim.CreateFormConfig(new(LoginModel))
		convey.So(e, convey.ShouldBeNil)
		convey.So(len(cfgMain.SectionGroups), convey.ShouldBeGreaterThan, 0)
		convey.Convey("Check Sections", func() {
			cfg := cfgMain.SectionGroups[0]
			convey.So(len(cfg.Sections), convey.ShouldEqual, 2)
			convey.So(len(cfg.Sections[0].Rows), convey.ShouldEqual, 5)

			convey.Convey("Check Elements", func() {
				convey.So(cfg.Sections[0].Rows[0][0].Label, convey.ShouldEqual, "Remember me")
				convey.So(cfg.Sections[0].Rows[0][0].Kind, convey.ShouldEqual, "checkbox")
				convey.So(cfg.Sections[0].Rows[2][0].Kind, convey.ShouldEqual, "password")
			})
		})
	})
}

func BenchmarkMemory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		suim.CreateFormConfig(new(LoginModel))
	}
}

func TestSuimGrid(t *testing.T) {
	convey.Convey("Form", t, func() {
		cfg, e := suim.CreateGridConfig(new(LoginModel))
		convey.So(e, convey.ShouldBeNil)

		convey.Convey("Check", func() {
			convey.So(len(cfg.Fields), convey.ShouldEqual, 6)
			convey.So(cfg.Fields[0].Label, convey.ShouldEqual, "Pin code")
		})
	})
}

func TestSuimGridCustomKey(t *testing.T) {
	convey.Convey("Grid with custom primary key", t, func() {
		cfg, e := suim.CreateGridConfig(new(ProductModel))
		convey.So(e, convey.ShouldBeNil)
		convey.So(cfg.Setting.IDField, convey.ShouldEqual, "code")
		convey.So(cfg.Setting.SortableFields, convey.ShouldResemble, []string{"code"})
		convey.So(cfg.Setting.KeywordFields, convey.ShouldResemble, []string{"code", "name"})
	})
}

type ProductModel struct {
	Code string `json:"code" key:"1" grid:"include"`
	Name string `json:"name" grid:"include"`
}

func TestValidate(t *testing.T) {
	joinDate := time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC)
	convey.Convey("valid data", t, func() {
		user := UserModel{
			ID:       "userid1",
			Name:     "john doe",
			Email:    "john-doe@mail.com",
			Age:      29,
			BirthDay: time.Date(1999, 02, 01, 0, 0, 0, 0, time.UTC),
			JoinDate: &joinDate,
			Company:  "C1",
			Address: &Address{
				Street:  "123 Main St",
				City:    "Anytown",
				State:   "CA",
				ZipCode: "12345",
			},
			Phones:  []string{"123-456-7890", "987-654-3210"},
			Domains: []Domains{{ID: 1, Name: "A1"}, {ID: 2, Name: "A2"}},
		}
		err := suim.Validate(&user)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("not valid data", t, func() {
		user := UserModel{
			ID:       "userid1",
			Name:     "",
			Email:    "john-doe@mail.com",
			Age:      29,
			BirthDay: time.Date(1999, 02, 01, 0, 0, 0, 0, time.UTC),
			JoinDate: &joinDate,
			Company:  "C1",
			Address: &Address{
				Street:  "",
				City:    "Anytown",
				State:   "CA",
				ZipCode: "12345",
			},
			Phones:  []string{"123-456-7890", "987-654-3210"},
			Domains: []Domains{{ID: 1, Name: "A1"}, {ID: 2, Name: "A2"}},
		}
		err := suim.Validate(&user)
		convey.So(err, convey.ShouldBeError)
	})
}

type LoginModel struct {
	LoginId    string   `form_required:"1" form_length:"5,8"`
	Password   string   `form_required:"1" form_kind:"password"`
	RememberMe bool     `form_pos:"1,1"`
	PinCode    int      `form_section:"Setting" grid_pos:"1"`
	Data       []string `form_section:"Setting" grid:"hide"`
	Company    string   `form_use_list:"1" form_items:"C1|C2|C3"`
	Timeout    string   `form_use_list:"1" form_items:"30s|1m|5m|30m|60m|6h|12h"`
}

func (l *LoginModel) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{{
		Sections: []suim.FormSection{
			{Title: "General", AutoCol: 1},
			{Title: "Setting", AutoCol: 1}},
	}}
}

type UserModel struct {
	ID       string     `form_required:"1" form_length:"5,8"`
	Name     string     `form_required:"1"`
	Email    string     `form_required:"1"`
	Age      int        `form_required:"1"`
	BirthDay time.Time  `form_required:"1"`
	JoinDate *time.Time `form_required:"1"`
	Company  string     `form_required:"1" form_use_list:"1" form_items:"C1|C2|C3"`
	Address  *Address
	Phones   []string `form_required:"1"`
	Domains  []Domains
}
type Address struct {
	Street  string `form_required:"1"`
	City    string `form_required:"1"`
	State   string `form_required:"1"`
	ZipCode string `form_required:"1"`
}
type Domains struct {
	ID   int    `form_required:"1"`
	Name string `form_required:"1"`
}

func (l *UserModel) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{{
		Sections: []suim.FormSection{
			{Title: "General", AutoCol: 1}},
	}}
}
