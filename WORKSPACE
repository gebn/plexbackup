load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "e6a6c016b0663e06fa5fccf1cd8152eab8aa8180c583ec20c872f4f9953a7ac5",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.22.1/rules_go-v0.22.1.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.22.1/rules_go-v0.22.1.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

http_archive(
    name = "bazel_gazelle",
    sha256 = "d8c45ee70ec39a57e7a05e5027c32b1576cc7f16d9dd37135b0eddde45cf1b10",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/bazel-gazelle/releases/download/v0.20.0/bazel-gazelle-v0.20.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.20.0/bazel-gazelle-v0.20.0.tar.gz",
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
    tag = "v1.29.29",
)

go_repository(
    name = "com_github_jmespath_go_jmespath",
    commit = "c2b33e8439af944379acbdd9c3a5fe0bc44bd8a5",
    importpath = "github.com/jmespath/go-jmespath",
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
