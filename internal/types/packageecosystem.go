package types

type PackageEcosystem uint

const (
	PackageEcosystemBundler       PackageEcosystem = iota // bundler
	PackageEcosystemCargo                                 // cargo
	PackageEcosystemComposer                              // composer
	PackageEcosystemDocker                                // docker
	PackageEcosystemHex                                   // mix
	PackageEcosystemElm                                   // elm
	PackageEcosystemGitSubmodule                          // gitsubmodule
	PackageEcosystemGitHubActions                         // github-actions
	PackageEcosystemGoModules                             // gomod
	PackageEcosystemGradle                                // gradle
	PackageEcosystemMaven                                 // maven
	PackageEcosystemNpm                                   // npm
	PackageEcosystemNuGet                                 // nuget
	PackageEcosystemPip                                   // pip
	PackageEcosystemPub                                   // pub
	PackageEcosystemTerraform                             // terraform
)

//go:generate stringer -type=PackageEcosystem -linecomment
