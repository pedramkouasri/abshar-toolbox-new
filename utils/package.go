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

	return diff
}

func hasDiff(branch1 string, branch2 string) bool {
	return branch1 != branch2
}
