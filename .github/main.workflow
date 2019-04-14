workflow "Release" {
  on = "push"
  resolves = ["goreleaser"]
}

action "is-tag" {
  uses = "actions/bin/filter@master"
  args = "tag"
}

action "goreleaser" {
  uses = "docker://goreleaser/goreleaser"
  secrets = [
    "GITHUB_TOKEN",
    "DOCKER_USERNAME",
    "DOCKER_PASSWORD",
  ]
  args = "release"
  needs = ["is-tag"]
}

workflow "linter" {
  on = "push"
  resolves = ["golangci-lint"]
}

action "golangci-lint" {
  uses = "docker://golangci/golangci-lint:v1.16"
  args = "run --deadline=55m"
  runs = "golangci-lint"
}
