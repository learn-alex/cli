package featureflag_test

import (
	"errors"

	fakeflag "github.com/cloudfoundry/cli/cf/api/feature_flags/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/featureflag"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("feature-flags command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		flagRepo            *fakeflag.FakeFeatureFlagRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		flagRepo = &fakeflag.FakeFeatureFlagRepository{}
	})

	runCommand := func(args ...string) bool {
		cmd := NewListFeatureFlags(ui, testconfig.NewRepositoryWithDefaults(), flagRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			Expect(runCommand("blahblah")).To(BeFalse())
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			flags := []models.FeatureFlag{
				models.FeatureFlag{
					Name:         "user_org_creation",
					Enabled:      true,
					ErrorMessage: "error",
				},
				models.FeatureFlag{
					Name:    "private_domain_creation",
					Enabled: false,
				},
				models.FeatureFlag{
					Name:    "app_bits_upload",
					Enabled: true,
				},
				models.FeatureFlag{
					Name:    "app_scaling",
					Enabled: true,
				},
				models.FeatureFlag{
					Name:    "route_creation",
					Enabled: false,
				},
			}
			flagRepo.ListReturns(flags, nil)
		})

		It("lists the state of all feature flags", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Retrieving status of all flagged features as my-user..."},
				[]string{"Feature", "State"},
				[]string{"user_org_creation", "enabled"},
				[]string{"private_domain_creation", "disabled"},
				[]string{"app_bits_upload", "enabled"},
				[]string{"app_scaling", "enabled"},
				[]string{"route_creation", "disabled"},
			))
		})

		Context("when an error occurs", func() {
			BeforeEach(func() {
				flagRepo.ListReturns(nil, errors.New("An error occurred."))
			})

			It("fails with an error", func() {
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"An error occurred."},
				))
			})
		})
	})
})
