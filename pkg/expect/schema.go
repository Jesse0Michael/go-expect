package expect

type expectFile struct {
	Connections []fileConnection `yaml:"connections" json:"connections"`
	Scenarios   []fileScenario   `yaml:"scenarios"   json:"scenarios"`
}

type fileConnection struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"`
	URL  string `yaml:"url"  json:"url"`
}

type fileScenario struct {
	Name  string     `yaml:"name"  json:"name"`
	Steps []fileStep `yaml:"steps" json:"steps"`
}

type fileStep struct {
	Request *fileRequest     `yaml:"request" json:"request"`
	Expect  *fileExpectation `yaml:"expect"  json:"expect"`
}

type fileRequest struct {
	Connection string            `yaml:"connection" json:"connection"`
	Method     string            `yaml:"method"     json:"method"`
	Endpoint   string            `yaml:"endpoint"   json:"endpoint"`
	Body       any               `yaml:"body"       json:"body"`
	Header     map[string]string `yaml:"header"     json:"header"`
	Query      map[string]string `yaml:"query"      json:"query"`

	// SQL-specific fields
	Statement string `yaml:"statement" json:"statement"`
	Params    []any  `yaml:"params"    json:"params"`
	Exec      bool   `yaml:"exec"      json:"exec"`
}

type fileExpectation struct {
	Status int               `yaml:"status" json:"status"`
	Code   string            `yaml:"code"   json:"code"`
	Header map[string]string `yaml:"header" json:"header"`
	Body   any               `yaml:"body"   json:"body"`
	Save   []fileSaveEntry   `yaml:"save"   json:"save"`

	// SQL-specific fields
	RowCount     *int  `yaml:"row_count"     json:"row_count"`
	RowsAffected *int  `yaml:"rows_affected" json:"rows_affected"`
	Rows         []any `yaml:"rows"          json:"rows"`
}

type fileSaveEntry struct {
	Field string `yaml:"field" json:"field"`
	As    string `yaml:"as"    json:"as"`
}
