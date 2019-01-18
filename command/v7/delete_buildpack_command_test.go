package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = FDescribe("delete-buildpack Command", func() {

	var (
		cmd             DeleteBuildpackCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeDeleteBuildpackActor
		input           *Buffer
		binaryName    string
		buildpackName string
		executeErr    error
	)

	BeforeEach(func() {
		fakeActor = new(v7fakes.FakeDeleteBuildpackActor)
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())

		cmd = DeleteBuildpackCommand{
			Actor:       fakeActor,
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
		}
		binaryName = "faceman"
		buildpackName = "the-buildpack"
		fakeConfig.BinaryNameReturns(binaryName)
		cmd.RequiredArgs.Buildpack = buildpackName
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error if the check fails", func() {
			executeErr = cmd.Execute(nil)

			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			shouldCheckTargetedOrg, shouldCheckTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(shouldCheckTargetedOrg).To(BeFalse())
			Expect(shouldCheckTargetedSpace).To(BeFalse())
		})
	})

	When("the DeleteBuildpack actor completes successfully", func() {
		BeforeEach(func() {
			fakeActor.DeleteBuildpackReturns(nil, nil)
		})
		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		When("--force is specified", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			When("a stack is not specified", func() {
				BeforeEach(func() {
					cmd.Stack = ""
				})

				It("prints appropriate output", func() {
					Expect(testUI.Out).To(Say("Deleting buildpack the-buildpack..."))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("a stack is specified", func() {
				BeforeEach(func() {
					cmd.Stack = "a-stack"
				})

				It("prints appropriate output that includes the stack name", func() {
					Expect(testUI.Out).To(Say("Deleting buildpack the-buildpack with stack a-stack..."))
					Expect(testUI.Out).To(Say("OK"))
				})
			})
		})

		When("--force is not specified", func() {
			BeforeEach(func() {
				cmd.Force = false
			})

			When("the user inputs yes", func() {
				BeforeEach(func() {
					_, err := input.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("prompted the user for confirmation", func() {
					Expect(testUI.Out).To(Say("Deleting buildpack the-buildpack with stack a-stack..."))
				})
			})
		})
	})

	It("delegates to the actor", func() {
		cmd.Stack = "the-stack"
		fakeActor.DeleteBuildpackReturns(nil, nil)

		executeErr = cmd.Execute(nil)

		Expect(executeErr).ToNot(HaveOccurred())
		actualBuildpack, actualStack := fakeActor.DeleteBuildpackArgsForCall(0)
		Expect(actualBuildpack).To(Equal("the-buildpack"))
		Expect(actualStack).To(Equal("the-stack"))
	})

	It("prints warnings", func() {
		cmd.Stack = "a-stack"
		fakeActor.DeleteBuildpackReturns(v7action.Warnings{"a-warning"}, nil)

		executeErr = cmd.Execute(nil)

		Expect(executeErr).ToNot(HaveOccurred())
		Expect(testUI.Err).To(Say("a-warning"))
	})

	It("returns error from the actor and prints the errors", func() {
		cmd.Stack = "a-stack"

		fakeActor.DeleteBuildpackReturns(v7action.Warnings{"a-warning"}, errors.New("some-error"))

		executeErr = cmd.Execute(nil)

		Expect(executeErr).To(MatchError("some-error"))
		Expect(testUI.Err).To(Say("a-warning"))
	})
})
