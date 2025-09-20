package sentinel_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSentinel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sentinel Rules Suite")
}
