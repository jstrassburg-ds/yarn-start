package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testYarnBerry(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually
		pack       occam.Pack
		docker     occam.Docker

		pullPolicy       = "never"
		extenderBuildStr = ""
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()

		if settings.Extensions.UbiNodejsExtension.Online != "" {
			pullPolicy = "always"
			extenderBuildStr = "[extender (build)] "
		}
	})

	context("when building a Yarn Berry app", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
			source    string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("creates a working OCI image with start command", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "yarn_berry_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy(pullPolicy).
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.Yarn.Online,
					settings.Buildpacks.YarnInstall.Online,
					settings.Buildpacks.YarnStart.Online,
				).
				WithEnv(map[string]string{
					"BP_LOG_LEVEL": "DEBUG",
				}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred(), logs.String())

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s\d+\.\d+\.\d+`, extenderBuildStr)),
				"  Executing build process",
				MatchRegexp(`    Running 'yarn install( --production --cache-folder /layers/[\w\-_]+/yarn-cache)?'`),
				"",
				extenderBuildStr+"Paketo Buildpack for Yarn Start",
				MatchRegexp(`%s  yarn-start \d+\.\d+\.\d+`, extenderBuildStr),
				extenderBuildStr+"  Assigning launch processes:",
				extenderBuildStr+"    web (default): bash -c \"node server.js\"",
			))

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable())
			Eventually(container).Should(Serve(ContainSubstring("hello yarn berry")).OnPort(8080))
		})

		context("when BP_LIVE_RELOAD_ENABLED=true", func() {
			it("creates a working OCI image with reloadable process", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "yarn_berry_app"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy(pullPolicy).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.Yarn.Online,
						settings.Buildpacks.YarnInstall.Online,
						settings.Buildpacks.YarnStart.Online,
						settings.Buildpacks.Watchexec.Online,
					).
					WithEnv(map[string]string{
						"BP_LIVE_RELOAD_ENABLED": "true",
						"BP_LOG_LEVEL":           "DEBUG",
					}).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred(), logs.String())

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s\d+\.\d+\.\d+`, extenderBuildStr)),
					"  Executing build process",
					MatchRegexp(`    Running 'yarn install( --production --cache-folder /layers/[\w\-_]+/yarn-cache)?'`),
					"",

					extenderBuildStr+"Paketo Buildpack for Yarn Start",
					MatchRegexp(`%s  yarn-start \d+\.\d+\.\d+`, extenderBuildStr),
					extenderBuildStr+"  Assigning launch processes:",
					MatchRegexp(`%s    web \(default\): watchexec --restart --shell none --watch .* --ignore .*/package\.json --ignore .*/yarn\.lock --ignore .*/node_modules -- bash -c "node server\.js"`, extenderBuildStr),
					MatchRegexp(`%s    no-reload: bash -c "node server\.js"`, extenderBuildStr),
				))

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())
				Eventually(container).Should(Serve(ContainSubstring("hello yarn berry")).OnPort(8080))
			})
		})
	})
}
