load("@io_bazel_rules_go//go:def.bzl", "go_prefix", "go_binary", "go_library", "go_test")
load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")

go_prefix("github.com/BranLwyd/rssdl")

##
## Binaries
##
go_binary(
    name = "rssdld",
    srcs = ["rssdld.go"],
    deps = [
        ":config",
        ":rssdl_proto",
        ":weekly",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_mmcdole_gofeed//:go_default_library",
    ],
)

##
## Libraries
##
go_library(
    name = "config",
    srcs = ["config.go"],
    deps = [
        ":rssdl_proto",
        ":weekly",
        "@com_github_golang_protobuf//proto:go_default_library",
    ],
)

go_library(
    name = "weekly",
    srcs = ["weekly.go"],
)

go_test(
    name = "weekly_test",
    srcs = ["weekly_test.go"],
    library = "weekly",
)

##
## Protos
##
go_proto_library(
    name = "rssdl_proto",
    srcs = ["rssdl.proto"],
)
