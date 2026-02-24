package loader

type expectFile struct {
	Connections []connection `yaml:"connections" json:"connections"`
	Scenarios   []scenario   `yaml:"scenarios"   json:"scenarios"`
}

type connection struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"`
	URL  string `yaml:"url"  json:"url"`
}

type scenario struct {
	Name  string `yaml:"name"  json:"name"`
	Steps []step `yaml:"steps" json:"steps"`
}

type step struct {
	Request *request     `yaml:"request" json:"request"`
	Expect  *expectation `yaml:"expect"  json:"expect"`
}

type request struct {
	Connection string            `yaml:"connection" json:"connection"`
	Method     string            `yaml:"method"     json:"method"`
	Endpoint   string            `yaml:"endpoint"   json:"endpoint"`
	Body       any               `yaml:"body"       json:"body"`
	Header     map[string]string `yaml:"header"     json:"header"`
	Query      map[string]string `yaml:"query"      json:"query"`
}

type expectation struct {
	Status int               `yaml:"status" json:"status"`
	Code   string            `yaml:"code"   json:"code"`
	Header map[string]string `yaml:"header" json:"header"`
	Body   any               `yaml:"body"   json:"body"`
	Save   []saveEntry       `yaml:"save"   json:"save"`
}

type saveEntry struct {
	Field string `yaml:"field" json:"field"`
	As    string `yaml:"as"    json:"as"`
}
