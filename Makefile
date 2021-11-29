GOTEST_MIN = go test 
GOTEST = $(GOTEST_MIN) 
GOTEST_WITH_COVERAGE = $(GOTEST) -coverprofile cover.out ./...

.PHONY: cover
cover:
		$(GOTEST_WITH_COVERAGE)