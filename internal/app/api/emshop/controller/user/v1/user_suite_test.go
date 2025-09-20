package user

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUserController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EMShop API User Controller Suite")
}
