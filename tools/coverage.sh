#!/bin/bash
go test -cover -coverprofile=./tools/coverage.out ./...

TEST_ERROR=$?

function cleanUp {
  rm -f ./tools/coverage.out
  rm -f ./tools/coverage_cleaned.out
}

if [ $TEST_ERROR -ne 0 ]; then
    exit $TEST_ERROR
fi

# ignore mocks and main.go
grep -v 'mock_' ./tools/coverage.out | grep -v 'main.go' |grep -v "/models/" > ./tools/coverage_cleaned.out

go tool cover -func=./tools/coverage_cleaned.out
go tool cover -html=./tools/coverage_cleaned.out -o ./tools/coverage.html

trap cleanUp EXIT
