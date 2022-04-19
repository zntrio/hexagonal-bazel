load("@bazel_gazelle//:def.bzl", "gazelle")
load("@com_github_bazelbuild_buildtools//buildifier:def.bzl", "buildifier")

buildifier(
    name = "buildifier",
)

# gazelle:prefix zntr.io/hexagonal-bazel
#
# gazelle:exclude bin
# gazelle:exclude vendor
# gazelle:exclude **/**.proto
# gazelle:exclude **/*_generated.go
# gazelle:exclude **/*_generated_test.go
gazelle(
    name = "gazelle",
    prefix = "zntr.io/hexagonal-bazel",
)

gazelle(
    name = "gazelle-update-repos",
    args = [
        "-from_file=go.mod",
        "-to_macro=go_repositories.bzl%go_repositories",
        "-prune",
        "-build_file_proto_mode=disable_global",
    ],
    command = "update-repos",
)
