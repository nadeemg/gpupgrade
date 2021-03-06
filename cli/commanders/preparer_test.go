package commanders_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/greenplum-db/gpupgrade/cli/commanders"
	"github.com/greenplum-db/gpupgrade/hub/configutils"
	pb "github.com/greenplum-db/gpupgrade/idl"
	mockpb "github.com/greenplum-db/gpupgrade/mock_idl"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"

	"github.com/greenplum-db/gpupgrade/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("preparer", func() {

	var (
		client *mockpb.MockCliToHubClient
		ctrl   *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		client = mockpb.NewMockCliToHubClient(ctrl)
	})

	AfterEach(func() {
		utils.System = utils.InitializeSystemFunctions()
		defer ctrl.Finish()
	})

	Describe("VerifyConnectivity", func() {
		It("returns nil when hub answers PingRequest", func() {
			testhelper.SetupTestLogger()

			client.EXPECT().Ping(
				gomock.Any(),
				&pb.PingRequest{},
			).Return(&pb.PingReply{}, nil)

			preparer := commanders.Preparer{}
			err := preparer.VerifyConnectivity(client)
			Expect(err).To(BeNil())
		})

		It("returns err when hub doesn't answer PingRequest", func() {
			testhelper.SetupTestLogger()
			commanders.NumberOfConnectionAttempt = 1

			client.EXPECT().Ping(
				gomock.Any(),
				&pb.PingRequest{},
			).Return(&pb.PingReply{}, errors.New("not answering ping")).Times(commanders.NumberOfConnectionAttempt + 1)

			preparer := commanders.Preparer{}
			err := preparer.VerifyConnectivity(client)
			Expect(err).ToNot(BeNil())
		})
		It("returns success if Ping eventually answers", func() {
			testhelper.SetupTestLogger()

			client.EXPECT().Ping(
				gomock.Any(),
				&pb.PingRequest{},
			).Return(&pb.PingReply{}, errors.New("not answering ping"))

			client.EXPECT().Ping(
				gomock.Any(),
				&pb.PingRequest{},
			).Return(&pb.PingReply{}, nil)

			preparer := commanders.Preparer{}
			err := preparer.VerifyConnectivity(client)
			Expect(err).To(BeNil())
		})
	})

	Describe("PrepareInitCluster", func() {
		It("returns successfully if hub gets the request", func() {
			testStdout, _, _ := testhelper.SetupTestLogger()
			client.EXPECT().PrepareInitCluster(
				gomock.Any(),
				&pb.PrepareInitClusterRequest{DbPort: int32(11111), NewBinDir: "/tmp"},
			).Return(&pb.PrepareInitClusterReply{}, nil)
			preparer := commanders.NewPreparer(client)
			err := preparer.InitCluster(11111, "/tmp")
			Expect(err).To(BeNil())
			Eventually(testStdout).Should(gbytes.Say("Gleaning the new cluster config"))
		})
	})
	Describe("PrepareShutdownCluster", func() {
		It("returns successfully", func() {
			testStdout, _, _ := testhelper.SetupTestLogger()

			client.EXPECT().PrepareShutdownClusters(
				gomock.Any(),
				&pb.PrepareShutdownClustersRequest{OldBinDir: "/old", NewBinDir: "/new"},
			).Return(&pb.PrepareShutdownClustersReply{}, nil)
			preparer := commanders.NewPreparer(client)
			err := preparer.ShutdownClusters("/old", "/new")
			Expect(err).To(BeNil())
			Eventually(testStdout).Should(gbytes.Say("request to shutdown clusters sent to hub"))
		})
	})
	Describe("PrepareStartAgents", func() {
		It("returns successfully", func() {
			testStdout, _, _ := testhelper.SetupTestLogger()

			client.EXPECT().PrepareStartAgents(
				gomock.Any(),
				&pb.PrepareStartAgentsRequest{},
			).Return(&pb.PrepareStartAgentsReply{}, nil)
			preparer := commanders.NewPreparer(client)
			err := preparer.StartAgents()
			Expect(err).To(BeNil())
			Eventually(testStdout).Should(gbytes.Say("Started Agents in progress, check gpupgrade_agent logs for details"))
		})
	})
	Describe("Prepare init", func() {
		It("creates dir when none exists", func() {
			dir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dir)

			stateDir := filepath.Join(dir, "foo")
			err = commanders.DoInit(stateDir, "/does/not/exist")
			Expect(err).To(BeNil())

			reader := configutils.NewReader()
			reader.OfOldClusterConfig(stateDir)
			reader.Read()
			Expect(reader.GetBinDir()).To(Equal("/does/not/exist"))
		})
		It("errs out when dir exists", func() {
			dir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dir)

			err = commanders.DoInit(dir, "/does/not/exist")
			Expect(err).ToNot(BeNil())
		})
	})
})
