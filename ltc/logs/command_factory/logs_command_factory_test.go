package command_factory_test

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/fake_app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/fake_exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
)

var _ = Describe("CommandFactory", func() {
	var (
		appExaminer             *fake_app_examiner.FakeAppExaminer
		outputBuffer            *gbytes.Buffer
		terminalUI              terminal.UI
		fakeTailedLogsOutputter *fake_tailed_logs_outputter.FakeTailedLogsOutputter
		signalChan              chan os.Signal
		exitHandler             exit_handler.ExitHandler
	)

	BeforeEach(func() {
		appExaminer = &fake_app_examiner.FakeAppExaminer{}
		outputBuffer = gbytes.NewBuffer()
		terminalUI = terminal.NewUI(nil, outputBuffer, nil)
		fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
		signalChan = make(chan os.Signal)
		exitHandler = &fake_exit_handler.FakeExitHandler{}
	})

	Describe("LogsCommand", func() {

		var logsCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewLogsCommandFactory(appExaminer, terminalUI, fakeTailedLogsOutputter, exitHandler)
			logsCommand = commandFactory.MakeLogsCommand()
		})

		It("tails logs", func() {
			args := []string{
				"my-app-guid",
			}
			appExaminer.AppExistsReturns(true, nil)

			test_helpers.AsyncExecuteCommandWithArgs(logsCommand, args)

			Eventually(fakeTailedLogsOutputter.OutputTailedLogsCallCount).Should(Equal(1))
			Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("my-app-guid"))
		})

		It("handles invalid appguids", func() {
			test_helpers.ExecuteCommandWithArgs(logsCommand, []string{})

			Expect(outputBuffer).To(test_helpers.Say("Incorrect Usage"))
			Expect(fakeTailedLogsOutputter.OutputTailedLogsCallCount()).To(Equal(0))
		})

		It("handles non existent application", func() {
			args := []string{
				"non_existent_app",
			}
			appExaminer.AppExistsReturns(false, nil)

			test_helpers.AsyncExecuteCommandWithArgs(logsCommand, args)

			Eventually(fakeTailedLogsOutputter.OutputTailedLogsCallCount).Should(Equal(1))
			Expect(fakeTailedLogsOutputter.OutputTailedLogsArgsForCall(0)).To(Equal("non_existent_app"))
			Expect(outputBuffer).To(test_helpers.Say("Application non_existent_app not found."))
			Expect(outputBuffer).To(test_helpers.Say("Tailing logs and waiting for non_existent_app to appear..."))
		})

		Context("when the receptor returns an error", func() {
			It("displays an error and exits", func() {
				args := []string{
					"non_existent_app",
				}
				appExaminer.AppExistsReturns(false, errors.New("can't log this"))

				commandDone := test_helpers.AsyncExecuteCommandWithArgs(logsCommand, args)

				Eventually(commandDone).Should(BeClosed())
				Expect(outputBuffer).To(test_helpers.Say("Error: can't log this"))
				Expect(fakeTailedLogsOutputter.OutputTailedLogsCallCount()).To(BeZero())
			})
		})
	})

	Describe("DebugLogsCommand", func() {

		var debugLogsCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewLogsCommandFactory(appExaminer, terminalUI, fakeTailedLogsOutputter, exitHandler)
			debugLogsCommand = commandFactory.MakeDebugLogsCommand()
		})

		It("tails logs from the lattice-debug stream", func() {
			test_helpers.AsyncExecuteCommandWithArgs(debugLogsCommand, []string{})

			Eventually(fakeTailedLogsOutputter.OutputDebugLogsCallCount).Should(Equal(1))
			Expect(fakeTailedLogsOutputter.OutputDebugLogsArgsForCall(0)).To(Equal(true))
		})

		Context("when the --raw flag is passed", func() {
			It("tails the debug logs without pretty print", func() {
				test_helpers.AsyncExecuteCommandWithArgs(debugLogsCommand, []string{"--raw"})

				Eventually(fakeTailedLogsOutputter.OutputDebugLogsCallCount).Should(Equal(1))
				Expect(fakeTailedLogsOutputter.OutputDebugLogsArgsForCall(0)).To(Equal(false))
			})
		})

	})

})
