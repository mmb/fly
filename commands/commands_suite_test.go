package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var pathToTableRunner string

var _ = SynchronizedBeforeSuite(func() []byte {
	binPath, err := gexec.Build("github.com/concourse/fly/commands/tablerunner")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binPath)
}, func(data []byte) {
	pathToTableRunner = string(data)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}
