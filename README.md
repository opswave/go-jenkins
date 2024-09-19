# go-jenkins
Jenkins API Client in Go, support The Blue Ocean REST API.

## Installation
go get github.com/opswave/go-jenkins

## About

Jenkins is the most popular Open Source Continuous Integration system. This Library will help you interact with Jenkins in a more developer-friendly way.

Fork From 
* https://github.com/bndr/gojenkins
* https://github.com/kubesphere/ks-devops

These are some of the features that are currently implemented:

* Get information on test-results of completed/failed build
* Ability to query Nodes, and manipulate them. Start, Stop, set Offline.
* Ability to query Jobs, and manipulate them.
* Get Plugins, Builds, Artifacts, Fingerprints
* Validate Fingerprints of Artifacts
* Get Current Queue, Cancel Tasks
* etc. For all methods go to GoDoc Reference.

Add some features:

* Credentials Management
* Pipeline Model Converter
* RBAC control
