stack {
  name        = "worker"
  description = "Worker function"
  runtime     = "go1.x"
  handler     = "."
  after       = ["../shared"]
}
