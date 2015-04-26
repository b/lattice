package app_examiner_test

import (
	"github.com/cloudfoundry/noaa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/cli_app_factory"
)

var _ = Describe("NoaaConsumer", func() {

	It("consumes the noaa", func() {
		noaaConsumer := noaa.NewConsumer(cli_app_factory.LoggregatorUrl(""), nil, nil)

		consumer := app_examiner.NewNoaaConsumer(noaaConsumer)

		Expect(consumer).NotTo(BeNil())
	})

	It("gets container metrics", func() {
		noaaConsumer := noaa.NewConsumer(cli_app_factory.LoggregatorUrl(""), nil, nil)

		consumer := app_examiner.NewNoaaConsumer(noaaConsumer)

		containerMetrics, _ := consumer.GetContainerMetrics("", "")

		// Expect(err).NotTo(HaveOccurred())
		Expect(containerMetrics).To(BeZero())
	})

})
