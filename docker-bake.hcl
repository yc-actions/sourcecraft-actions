// docker-bake.hcl
// This file configures docker buildx bake to build all images in the repository

// Define variables with default values
variable "REGISTRY" {
  default = ""
}

variable "TAG" {
  default = "latest"
}

// Define the group that includes all targets
group "default" {
  targets = ["apigw", "coi", "container", "function", "obj-storage-upload"]
}

// Common configuration for all targets
target "common" {
  dockerfile = "Dockerfile"
  platforms = ["linux/amd64"]
  pull = true
}

// Target for apigw
target "apigw" {
  inherits = ["common"]
  args = {
    APP_NAME = "apigw"
  }
  tags = [
    "${REGISTRY}apigw:${TAG}",
    notequal("", REGISTRY) ? "${REGISTRY}apigw:latest" : ""
  ]
}

// Target for coi
target "coi" {
  inherits = ["common"]
  args = {
    APP_NAME = "coi"
  }
  tags = [
    "${REGISTRY}coi:${TAG}",
    notequal("", REGISTRY) ? "${REGISTRY}coi:latest" : ""
  ]
}

// Target for container
target "container" {
  inherits = ["common"]
  args = {
    APP_NAME = "container"
  }
  tags = [
    "${REGISTRY}container:${TAG}",
    notequal("", REGISTRY) ? "${REGISTRY}container:latest" : ""
  ]
}

// Target for function
target "function" {
  inherits = ["common"]
  args = {
    APP_NAME = "function"
  }
  tags = [
    "${REGISTRY}function:${TAG}",
    notequal("", REGISTRY) ? "${REGISTRY}function:latest" : ""
  ]
}

// Target for obj-storage-upload
target "obj-storage-upload" {
  inherits = ["common"]
  args = {
    APP_NAME = "obj-storage-upload"
  }
  tags = [
    "${REGISTRY}obj-storage-upload:${TAG}",
    notequal("", REGISTRY) ? "${REGISTRY}obj-storage-upload:latest" : ""
  ]
}