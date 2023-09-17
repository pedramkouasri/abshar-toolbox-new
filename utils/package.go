package utils

import "github.com/pedramkousari/abshar-toolbox-new/types"

func GetPackageDiff(pkg []types.Packages) []types.CreatePackageParams {
	package1 := pkg[len(pkg)-2].PackageService

	package2 := pkg[len(pkg)-1].PackageService

	diff := []types.CreatePackageParams{}

	if hasDiff(package1.Baadbaan, package2.Baadbaan) {
		diff = append(diff, types.CreatePackageParams{
			ServiceName: "baadbaan",
			Tag1:        package1.Baadbaan,
			Tag2:        package2.Baadbaan,
		})
	}

	if hasDiff(package1.Technical, package2.Technical) {
		diff = append(diff, types.CreatePackageParams{
			ServiceName: "technical",
			Tag1:        package1.Technical,
			Tag2:        package2.Technical,
		})
	}

	if hasDiff(package1.Docker, package2.Docker) {
		diff = append(diff, types.CreatePackageParams{
			ServiceName: "docker",
			Tag1:        package1.Docker,
			Tag2:        package2.Docker,
		})
	}

	if hasDiff(package1.Toolbox, package2.Toolbox) {
		diff = append(diff, types.CreatePackageParams{
			ServiceName: "toolbox",
			Tag1:        package1.Toolbox,
			Tag2:        package2.Toolbox,
		})
	}

	if hasDiff(package1.Discovery, package2.Discovery) {
		diff = append(diff, types.CreatePackageParams{
			ServiceName: "discovery",
			Tag1:        package1.Discovery,
			Tag2:        package2.Discovery,
		})
	}

	return diff
}

func hasDiff(branch1 string, branch2 string) bool {
	return branch1 != branch2
}
