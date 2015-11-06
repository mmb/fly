package integration_test

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/vito/go-sse/sse"

	"github.com/concourse/atc"
	"github.com/concourse/atc/event"
)

var _ = FDescribe("Fly CLI", func() {
	var tmpdir string
	var buildDir string
	var taskConfigPath string

	var atcServer *ghttp.Server
	var streaming chan struct{}
	var events chan atc.Event
	var uploadingBits <-chan struct{}

	var expectedPlan atc.Plan

	BeforeEach(func() {
		var err error
		tmpdir, err = ioutil.TempDir("", "fly-build-dir")
		Expect(err).NotTo(HaveOccurred())

		buildDir = filepath.Join(tmpdir, "fixture")

		err = os.Mkdir(buildDir, 0755)
		Expect(err).NotTo(HaveOccurred())

		taskConfigPath = filepath.Join(buildDir, "task.yml")

		err = ioutil.WriteFile(
			taskConfigPath,
			[]byte(`---
platform: some-platform

image: ubuntu

outputs:
- name: fixture 

params:
  FOO: bar
  BAZ: buzz
  X: 1

run:
  path: find
  args: [.]
`),
			0644,
		)
		Expect(err).NotTo(HaveOccurred())

		atcServer = ghttp.NewServer()

		streaming = make(chan struct{})
		events = make(chan atc.Event)

		expectedPlan = atc.Plan{
			OnSuccess: &atc.OnSuccessPlan{
				Step: atc.Plan{
					Aggregate: &atc.AggregatePlan{},
				},
				Next: atc.Plan{
					Ensure: &atc.EnsurePlan{
						Step: atc.Plan{
							Location: &atc.Location{
								ParallelGroup: 0,
								ParentID:      0,
								ID:            2,
							},
							Task: &atc.TaskPlan{
								Name: "one-off",
								Config: &atc.TaskConfig{
									Platform: "some-platform",
									Image:    "ubuntu",
									Outputs: []atc.TaskOutputConfig{
										{Name: "fixture"},
									},
									Params: map[string]string{
										"FOO": "bar",
										"BAZ": "buzz",
										"X":   "1",
									},
									Run: atc.TaskRunConfig{
										Path: "find",
										Args: []string{"."},
									},
								},
							},
						},
						Next: atc.Plan{
							Aggregate: &atc.AggregatePlan{
								atc.Plan{
									Location: &atc.Location{
										ID:            4,
										ParentID:      0,
										ParallelGroup: 3,
									},
									Put: &atc.PutPlan{
										Name: "some-output",
										Type: "archive",
										Source: atc.Source{
											"uri": atcServer.URL() + "/api/v1/pipes/some-pipe-id",
										},
										Params: atc.Params{
											"directory": "some-output",
										},
									},
								},
							},
						},
					},
				},
			},
		}
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	JustBeforeEach(func() {
		uploading := make(chan struct{})
		uploadingBits = uploading

		atcServer.RouteToHandler("POST", "/api/v1/pipes",
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v1/pipes"),
				ghttp.RespondWithJSONEncoded(http.StatusCreated, atc.Pipe{
					ID: "some-pipe-id",
				}),
			),
		)
		atcServer.RouteToHandler("POST", "/api/v1/builds",
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/v1/builds"),
				ghttp.VerifyJSONRepresenting(expectedPlan),
				func(w http.ResponseWriter, r *http.Request) {
					http.SetCookie(w, &http.Cookie{
						Name:    "Some-Cookie",
						Value:   "some-cookie-data",
						Path:    "/",
						Expires: time.Now().Add(1 * time.Minute),
					})
				},
				ghttp.RespondWith(201, `{"id":128}`),
			),
		)
		atcServer.RouteToHandler("GET", "/api/v1/builds/128/events",
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/builds/128/events"),
				func(w http.ResponseWriter, r *http.Request) {
					flusher := w.(http.Flusher)

					w.Header().Add("Content-Type", "text/event-stream; charset=utf-8")
					w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
					w.Header().Add("Connection", "keep-alive")

					w.WriteHeader(http.StatusOK)

					flusher.Flush()

					close(streaming)

					id := 0

					for e := range events {
						payload, err := json.Marshal(event.Message{Event: e})
						Expect(err).NotTo(HaveOccurred())

						event := sse.Event{
							ID:   fmt.Sprintf("%d", id),
							Name: "event",
							Data: payload,
						}

						err = event.Write(w)
						Expect(err).NotTo(HaveOccurred())

						flusher.Flush()

						id++
					}

					err := sse.Event{
						Name: "end",
					}.Write(w)
					Expect(err).NotTo(HaveOccurred())
				},
			),
		)
		atcServer.RouteToHandler("PUT", "/api/v1/pipes/some-pipe-id",
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v1/pipes/some-pipe-id"),
				func(w http.ResponseWriter, req *http.Request) {
					close(uploading)

					gr, err := gzip.NewReader(req.Body)
					Expect(err).NotTo(HaveOccurred())

					tr := tar.NewReader(gr)

					hdr, err := tr.Next()
					Expect(err).NotTo(HaveOccurred())

					Expect(hdr.Name).To(Equal("./"))

					hdr, err = tr.Next()
					Expect(err).NotTo(HaveOccurred())

					Expect(hdr.Name).To(MatchRegexp("(./)?task.yml$"))
				},
				ghttp.RespondWith(200, ""),
			),
		)
		atcServer.RouteToHandler("GET", "/api/v1/pipes/some-pipe-id",
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/api/v1/pipes/some-pipe-id"),
				func(w http.ResponseWriter, req *http.Request) {
					//TODO do I need to check anything here?
				},
				ghttp.RespondWith(200, ""),
			),
		)
	})

	Context("when the user specifies outputs", func() {
		Context("when the task specifies those outputs", func() {
			FIt("tries to download output", func() {
				flyCmd := exec.Command(flyPath, "-t", atcServer.URL(), "e", "-c", taskConfigPath, "-o", "some-output=fixture")
				flyCmd.Dir = buildDir

				sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(sess.Out).Should(gbytes.Say("executing build 128\n"))

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(0))
			})
		})

		Context("when the task does not specify those outputs", func() {
			It("exits 1", func() {
				flyCmd := exec.Command(flyPath, "-t", atcServer.URL(), "e", "-c", taskConfigPath, "-o", "wrong-output=wrong-path")
				flyCmd.Dir = buildDir

				sess, err := gexec.Start(flyCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(sess.Err).Should(gbytes.Say("unknown output `wrong-output'"))

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(1))
			})
		})
	})
})
