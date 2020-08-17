load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "2697f6bc7c529ee5e6a2d9799870b9ec9eaeb3ee7d70ed50b87a2c2c97e13d9e",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.23.8/rules_go-v0.23.8.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.23.8/rules_go-v0.23.8.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

http_archive(
    name = "bazel_gazelle",
    sha256 = "cdb02a887a7187ea4d5a27452311a75ed8637379a1287d8eeb952138ea485f7d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "com_github_gebn_go_stamp",
    importpath = "github.com/gebn/go-stamp",
    tag = "v2.0.2",
)

go_repository(
    name = "com_github_aws_aws_sdk_go",
    importpath = "github.com/aws/aws-sdk-go",
    tag = "v1.34.6",
)

go_repository(
    name = "com_github_jmespath_go_jmespath",
    importpath = "github.com/jmespath/go-jmespath",
    tag = "v0.3.0",
)

go_repository(
    name = "com_github_alecthomas_kingpin",
    importpath = "github.com/alecthomas/kingpin",
    tag = "v2.2.6",
)

go_repository(
    name = "com_github_alecthomas_units",
    commit = "f65c72e2690dc4b403c8bd637baf4611cd4c069b",
    importpath = "github.com/alecthomas/units",
)

go_repository(
    name = "com_github_alecthomas_template",
    commit = "fb15b899a75114aa79cc930e33c46b577cc664b1",
    importpath = "github.com/alecthomas/template",
)

go_repository(
    name = "org_golang_x_sys",
    commit = "9781c653f44323e2079cb1123720c32f9626be2c",
    importpath = "golang.org/x/sys",
)
