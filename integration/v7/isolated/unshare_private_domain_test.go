package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Unshare Private Domain", func() {
	BeforeEach(func() {
		Skip("waiting for implementation")
	})
	var (
		domainName      string
		owningOrgName   string
		sharedToOrgName string
	)

	BeforeEach(func() {
		domainName = helpers.NewDomainName()
		sharedToOrgName = helpers.NewOrgName()
	})

	Describe("Help Text", func() {
		It("Displays the help text", func() {
			session := helpers.CF("unshare-private-domain", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("unshare-private-domain - Unshare a private domain with an org"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf unshare-private-domain ORG DOMAIN"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("delete-domain, domains"))
		})
	})

	Describe("When the environment is not set up correctly", func() {
		When("The user is not logged in", func() {
			It("lets the user know", func() {
				session := helpers.CF("unshare-private-domain", sharedToOrgName, domainName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Not logged in. Use 'cf login' to log in."))
			})
		})
	})

	Describe("When the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			owningOrgName = helpers.CreateAndTargetOrg()
			helpers.CreateOrg(sharedToOrgName)
			domain := helpers.NewDomain(owningOrgName, domainName)
			domain.CreatePrivate()
			domain.V7Share(sharedToOrgName)
		})

		It("unshares the domain from the org", func() {
			session := helpers.CF("unshare-private-domain", sharedToOrgName, domainName)
			Eventually(session).Should(Say("Unsharing domain %s from org %s as admin...", domainName, sharedToOrgName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))

			helpers.TargetOrg(sharedToOrgName)
			session = helpers.CF("domains")
			Consistently(session).Should(Not(Say("%s", domainName)))
		})
	})
})
