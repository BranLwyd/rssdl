load("@io_bazel_rules_go//go:def.bzl", "go_prefix", "go_binary", "go_library", "go_test")
load("@io_bazel_rules_go//proto:go_proto_library.bzl", "go_proto_library")

go_prefix("github.com/BranLwyd/rssdl")

##
## Binaries
##
go_binary(
    name = "rssdld",
    srcs = ["rssdld.go"],
)

##
## Libraries
##
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
