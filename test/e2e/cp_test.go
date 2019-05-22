// +build !remoteclient

package integration

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/containers/libpod/test/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Podman cp", func() {
	var (
		tempdir    string
		err        error
		podmanTest *PodmanTestIntegration
	)

	BeforeEach(func() {
		tempdir, err = CreateTempDirInTempDir()
		if err != nil {
			os.Exit(1)
		}
		podmanTest = PodmanTestCreate(tempdir)
		podmanTest.Setup()
		podmanTest.RestoreAllArtifacts()
	})

	AfterEach(func() {
		podmanTest.Cleanup()
		f := CurrentGinkgoTestDescription()
		processTestResult(f)

	})

	It("podman cp file", func() {
		srcPath := filepath.Join(podmanTest.RunRoot, "cp_test.txt")
		dstPath := filepath.Join(podmanTest.RunRoot, "cp_from_container")
		fromHostToContainer := []byte("copy from host to container")
		err := ioutil.WriteFile(srcPath, fromHostToContainer, 0644)
		Expect(err).To(BeNil())

		session := podmanTest.Podman([]string{"create", ALPINE, "cat", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))
		name := session.OutputToString()

		session = podmanTest.Podman([]string{"cp", srcPath, name + ":foo"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))

		session = podmanTest.Podman([]string{"cp", name + ":foo", dstPath})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))
	})

	It("podman cp file to dir", func() {
		srcPath := filepath.Join(podmanTest.RunRoot, "cp_test.txt")
		dstDir := filepath.Join(podmanTest.RunRoot, "receive")
		fromHostToContainer := []byte("copy from host to container directory")
		err := ioutil.WriteFile(srcPath, fromHostToContainer, 0644)
		Expect(err).To(BeNil())
		err = os.Mkdir(dstDir, 0755)
		Expect(err).To(BeNil())

		session := podmanTest.Podman([]string{"create", ALPINE, "ls", "foodir/"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))
		name := session.OutputToString()

		session = podmanTest.Podman([]string{"cp", srcPath, name + ":foodir/"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))

		session = podmanTest.Podman([]string{"cp", name + ":foodir/cp_test.txt", dstDir})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))

		os.Remove("cp_test.txt")
		os.RemoveAll("receive")
	})

	It("podman cp dir to dir", func() {
		testDirPath := filepath.Join(podmanTest.RunRoot, "TestDir")
		err := os.Mkdir(testDirPath, 0755)
		Expect(err).To(BeNil())

		session := podmanTest.Podman([]string{"create", ALPINE, "ls", "/foodir"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))
		name := session.OutputToString()

		session = podmanTest.Podman([]string{"cp", testDirPath, name + ":/foodir"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))

		session = podmanTest.Podman([]string{"cp", testDirPath, name + ":/foodir"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))
	})

	It("podman cp stdin/stdout", func() {
		testDirPath := filepath.Join(podmanTest.RunRoot, "TestDir")
		err := os.Mkdir(testDirPath, 0755)
		Expect(err).To(BeNil())
		cmd := exec.Command("tar", "-zcvf", "file.tar.gz", testDirPath)
		_, err = cmd.Output()
		Expect(err).To(BeNil())

		session := podmanTest.Podman([]string{"create", ALPINE, "ls", "foo"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))
		name := session.OutputToString()

		data, err := ioutil.ReadFile("foo.tar.gz")
		reader := strings.NewReader(string(data))
		cmd.Stdin = reader
		session = podmanTest.Podman([]string{"cp", "-", name + ":/foo"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))

		session = podmanTest.Podman([]string{"cp", "file.tar.gz", name + ":/foo.tar.gz"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))
		session = podmanTest.Podman([]string{"cp", name + ":/foo.tar.gz", "-"})
		session.WaitWithDefaultTimeout()
		Expect(session.ExitCode()).To(Equal(0))
	})
})
