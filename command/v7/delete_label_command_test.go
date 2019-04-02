package v7_test

import (
	"errors"
	"regexp"

	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-label command", func() {
	var (
		cmd             DeleteLabelCommand
		fakeConfig      *commandfakes.FakeConfig
		testUI          *ui.UI
		fakeSharedActor *commandfakes.FakeSharedActor

		executeErr error
	)
	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		cmd = DeleteLabelCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("doesn't error", func() {
		Expect(executeErr).ToNot(HaveOccurred())
	})

	It("checks that the user is logged in and targeted to an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(checkOrg).To(BeTrue())
		Expect(checkSpace).To(BeTrue())
	})

	When("checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("Target not found"))
		})

		It("we expect an error to be returned", func() {
			Expect(executeErr).To(MatchError("Target not found"))
		})
	})

	When("checking the target succeeds", func() {
		var appName string

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})
			appName = generateAppName()
		})

		It("informs the user that labels are being deleted", func() {
			Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Deleting label(s) for app %s in org fake-org / space fake-space as some-user...`), appName))

		})

		It("collects the label keys from the command", func() {

		})

		It("calls the command to update labels on an application", func() {

		})
	})
})

func generateAppName() string {
	u, _ := uuid.NewV4()
	return u.String()
}
