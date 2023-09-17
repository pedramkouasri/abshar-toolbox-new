package types

type PackageService struct {
	Baadbaan  string `json:"baadbaan"`
	Technical string `json:"technical"`
	Discovery string `json:"discovery"`
	Toolbox   string `json:"toolbox"`
	Docker    string `json:"docker"`
}

type Packages struct {
	Version        string `json:"version"`
	PackageService `json:"package_version"`
}

type CreatePackageParams struct {
	ServiceName string
	Tag1        string
	Tag2        string
}
