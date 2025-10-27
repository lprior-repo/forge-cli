stack {
  name        = "api"
  description = "API function"
  runtime     = "go1.x"
  handler     = "."
  after       = ["../shared"]
}
