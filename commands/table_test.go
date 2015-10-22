package commands_test

import (
	"os/exec"

	. "github.com/concourse/fly/commands"
	"github.com/fatih/color"
	"github.com/kr/pty"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Table", func() {
	var table Table

	BeforeEach(func() {
		table = Table{
			Headers: TableRow{
				{Contents: "column1", Color: color.New(color.Bold)},
				{Contents: "column2", Color: color.New(color.Bold)},
			},
			Data: []TableRow{
				{
					{Contents: "r1c1"},
					{Contents: "r1c2"},
				},
				{
					{Contents: "r2c1"},
					{Contents: "r2c2"},
				},
				{
					{Contents: "r3c1"},
					{Contents: "r3c2"},
				},
			},
		}
	})

	Context("when the render method is called without a TTY", func() {
		It("prints the headers and the data", func() {
			expectedOutput := "" +
				"column1  column2\n" +
				"r1c1     r1c2   \n" +
				"r2c1     r2c2   \n" +
				"r3c1     r3c2   \n"
			actualOutput := table.Render()
			Expect(actualOutput).To(Equal(expectedOutput))
		})
	})

	Context("when the render method is called in a TTY", func() {
		It("prints the headers and the data in color", func() {
			pty, tty, err := pty.Open()
			Expect(err).NotTo(HaveOccurred())

			tableCmd := exec.Command(pathToTableRunner)
			tableCmd.Stdin = tty

			sess, err := gexec.Start(tableCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess.Out).Should(gbytes.Say("column1"))

			err = pty.Close()
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))
		})
	})
})
