package integrations_test

import (
	"errors"

	"github.com/greenplum-db/gpupgrade/hub/cluster"
	"github.com/greenplum-db/gpupgrade/hub/configutils"
	"github.com/greenplum-db/gpupgrade/hub/services"
	"github.com/greenplum-db/gpupgrade/testutils"

	"google.golang.org/grpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("upgrade reconfigure ports", func() {

	var (
		hub       *services.Hub
		hubExecer *testutils.FakeCommandExecer
		agentPort int

		outChan            chan []byte
		errChan            chan error
		stubRemoteExecutor *testutils.StubRemoteExecutor
	)

	BeforeEach(func() {
		var err error

		config := `[{
			"dbid": 1,
			"port": 5432,
			"host": "localhost"
		}]`

		testutils.WriteOldConfig(testStateDir, config)
		testutils.WriteNewConfig(testStateDir, config)

		agentPort, err = testutils.GetOpenPort()
		Expect(err).ToNot(HaveOccurred())

		port, err = testutils.GetOpenPort()
		Expect(err).ToNot(HaveOccurred())

		conf := &services.HubConfig{
			CliToHubPort:   port,
			HubToAgentPort: agentPort,
			StateDir:       testStateDir,
		}

		reader := configutils.NewReader()

		outChan = make(chan []byte, 10)
		errChan = make(chan error, 10)
		hubExecer = &testutils.FakeCommandExecer{}
		hubExecer.SetOutput(&testutils.FakeCommand{
			Out: outChan,
			Err: errChan,
		})

		stubRemoteExecutor = testutils.NewStubRemoteExecutor()
		hub = services.NewHub(&cluster.Pair{}, &reader, grpc.DialContext, hubExecer.Exec, conf, stubRemoteExecutor)
		go hub.Start()
	})

	AfterEach(func() {
		hub.Stop()

		Expect(checkPortIsAvailable(port)).To(BeTrue())
		Expect(checkPortIsAvailable(agentPort)).To(BeTrue())
	})

	It("updates status PENDING to COMPLETE if successful", func() {
		Expect(runStatusUpgrade()).To(ContainSubstring("PENDING - Adjust upgrade cluster ports"))

		upgradeReconfigurePortsSession := runCommand("upgrade", "reconfigure-ports")
		Eventually(upgradeReconfigurePortsSession).Should(Exit(0))

		Expect(hubExecer.Calls()[0]).To(ContainSubstring("sed"))

		Expect(runStatusUpgrade()).To(ContainSubstring("COMPLETE - Adjust upgrade cluster ports"))

	})

	It("updates status to FAILED if it fails to run", func() {
		Expect(runStatusUpgrade()).To(ContainSubstring("PENDING - Adjust upgrade cluster ports"))

		errChan <- errors.New("fake test error, reconfigure-ports failed")

		upgradeShareOidsSession := runCommand("upgrade", "reconfigure-ports")
		Eventually(upgradeShareOidsSession).Should(Exit(1))

		Eventually(runStatusUpgrade()).Should(ContainSubstring("FAILED - Adjust upgrade cluster ports"))
	})
})
