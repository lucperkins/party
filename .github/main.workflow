workflow "Test party library" {
  on = "push"
  resolves = ["test"]
}

action "test" {
  uses = "actions-contrib/go@master"
  args = "test -v -short -race ./..."
}