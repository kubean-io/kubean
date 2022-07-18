.PHONY: test
test:
	bash hack/unit-test.sh

.PHONY: staticcheck
staticcheck:
	hack/staticcheck.sh
