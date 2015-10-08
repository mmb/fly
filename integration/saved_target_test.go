package integration_test

import (
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Calling ATC with a saved target", func() {
	var (
		flyPath   string
		atcServer *ghttp.Server
	)

	BeforeEach(func() {
		var err error

		flyPath, err = gexec.Build("github.com/concourse/fly")
		Expect(err).NotTo(HaveOccurred())

		atcServer = ghttp.NewServer()
	})

	Context("with a github authentication tokent", func() {
		FIt("sets the autorization header to Token", func() {
			atcServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeader(http.Header{"Authorization": {"Token some-token"}}),
					ghttp.RespondWithJSONEncoded(200, nil, http.Header{}),
				),
			)

			flyCmd := exec.Command(flyPath, "save-target",
				"--api", atcServer.URL()+"/",
				"--github-personal-access-token", "some-token",
				"some-target",
			)

			sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(0))
			Eventually(sess).Should(gbytes.Say("successfully saved target some-target\n"))

			flyCmd = exec.Command(flyPath, "-t", "some-target", "configure", "some-pipeline")

			sess, err = gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))
		})
	})

})
