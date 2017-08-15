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
    name = "week",
    srcs = ["week.go"],
)

go_test(
    name = "week_test",
    srcs = ["week_test.go"],
    library = "week",
)

##
## Protos
##
go_proto_library(
    name = "rssdl_proto",
    srcs = ["rssdl.proto"],
)
