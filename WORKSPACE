git_repository(
    name = "io_bazel_rules_go",
    remote = "https://github.com/bazelbuild/rules_go.git",
    tag = "0.5.3",
)

load("@io_bazel_rules_go//go:def.bzl", "go_repositories", "go_repository")
load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_repositories")

go_repository(
    name = "com_github_andybalholm_cascadia",
    commit = "349dd0209470eabd9514242c688c403c0926d266",
    importpath = "github.com/andybalholm/cascadia",
)

go_repository(
    name = "com_github_golang_protobuf",
    commit = "ab9f9a6dab164b7d1246e0e688b0ab7b94d8553e",
    importpath = "github.com/golang/protobuf",
)

go_repository(
    name = "com_github_mmcdole_gofeed",
    commit = "042c0a9121581210fc8ef106d8ad1b0bdf931ae2",
    importpath = "github.com/mmcdole/gofeed",
)

go_repository(
    name = "com_github_mmcdole_goxpp",
    commit = "77e4a51a73ed99ee3f33c1474dc166866304acbd",
    importpath = "github.com/mmcdole/goxpp",
)

go_repository(
    name = "com_github_PuerkitoBio_goquery",
    commit = "e1271ee34c6a305e38566ecd27ae374944907ee9",
    importpath = "github.com/PuerkitoBio/goquery",
)

go_repository(
    name = "org_golang_x_text",
    commit = "e56139fd9c5bc7244c76116c68e500765bb6db6b",
    importpath = "golang.org/x/text",
)

go_repositories()

go_proto_repositories()
