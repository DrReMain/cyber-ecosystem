package biz

// region[rgba(239,83,80,0.15)] 🔴 Model -------------------------------------------------------------------------------

type ResourceMethod struct {
	Name             string
	FullName         string
	RequestName      string
	RequestFullName  string
	ResponseName     string
	ResponseFullName string
	HttpMethod       string
	HttpPath         string
	Comment          string
}

type ResourceService struct {
	Name       string
	FullName   string
	Package    string
	SourceFile string
	Comment    string
	Methods    []*ResourceMethod
}
