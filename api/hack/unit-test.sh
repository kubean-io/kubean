#!/usr/bin/env bash

go test github.com/kubean-io/kubean-api/generated
go test github.com/kubean-io/kubean-api/apis
go test github.com/kubean-io/kubean-api/apis/manifest/v1alpha1
