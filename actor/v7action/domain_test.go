package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _ = NewTestActor()
	})

	Describe("create shared domain", func() {
		It("delegates to the cloud controller client", func() {
			fakeCloudControllerClient.CreateDomainReturns(ccv3.Domain{}, ccv3.Warnings{"create-warning-1", "create-warning-2"}, errors.New("create-error"))

			warnings, executeErr := actor.CreateSharedDomain("the-domain-name", true)
			Expect(executeErr).To(MatchError("create-error"))
			Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2"))

			Expect(fakeCloudControllerClient.CreateDomainCallCount()).To(Equal(1))
			passedDomain := fakeCloudControllerClient.CreateDomainArgsForCall(0)

			Expect(passedDomain).To(Equal(
				ccv3.Domain{
					Name:     "the-domain-name",
					Internal: types.NullBool{IsSet: true, Value: true},
				},
			))
		})
	})

	Describe("create private domain", func() {

		BeforeEach(func() {
			fakeCloudControllerClient.GetOrganizationsReturns(
				[]ccv3.Organization{
					{GUID: "org-guid"},
				},
				ccv3.Warnings{"get-orgs-warning"},
				nil,
			)

			fakeCloudControllerClient.CreateDomainReturns(
				ccv3.Domain{},
				ccv3.Warnings{"create-warning-1", "create-warning-2"},
				errors.New("create-error"),
			)
		})

		It("delegates to the cloud controller client", func() {
			warnings, executeErr := actor.CreatePrivateDomain("private-domain-name", "org-name")
			Expect(executeErr).To(MatchError("create-error"))
			Expect(warnings).To(ConsistOf("get-orgs-warning", "create-warning-1", "create-warning-2"))

			Expect(fakeCloudControllerClient.CreateDomainCallCount()).To(Equal(1))
			passedDomain := fakeCloudControllerClient.CreateDomainArgsForCall(0)

			Expect(passedDomain).To(Equal(
				ccv3.Domain{
					Name:             "private-domain-name",
					OrganizationGUID: "org-guid",
				},
			))
		})
	})

	Describe("list domains for org", func() {
		var (
			ccv3Domains []ccv3.Domain
			domains     []Domain

			domain1Name string
			domain1Guid string

			domain2Name string
			domain2Guid string

			domain3Name string
			domain3Guid string

			org1Guid          = "some-org-guid"
			sharedFromOrgGuid = "another-org-guid"
			warnings          Warnings
			executeErr        error
		)

		BeforeEach(func() {
			ccv3Domains = []ccv3.Domain{
				{Name: domain1Name, GUID: domain1Guid, OrganizationGUID: org1Guid},
				{Name: domain2Name, GUID: domain2Guid, OrganizationGUID: org1Guid},
				{Name: domain3Name, GUID: domain3Guid, OrganizationGUID: sharedFromOrgGuid},
			}
		})

		JustBeforeEach(func() {
			domains, warnings, executeErr = actor.GetOrganizationDomains(org1Guid)
		})

		When("the API layer call is successful", func() {
			expectedOrgGUID := "some-org-guid"

			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationDomainsReturns(
					ccv3Domains,
					ccv3.Warnings{"some-domain-warning"},
					nil,
				)
			})

			It("returns back the domains and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationDomainsCallCount()).To(Equal(1))
				actualOrgGuid, actualQuery := fakeCloudControllerClient.GetOrganizationDomainsArgsForCall(0)
				Expect(actualOrgGuid).To(Equal(expectedOrgGUID))
				Expect(actualQuery).To(BeNil())

				Expect(domains).To(ConsistOf(
					Domain{Name: domain1Name, GUID: domain1Guid, OrganizationGUID: org1Guid},
					Domain{Name: domain2Name, GUID: domain2Guid, OrganizationGUID: org1Guid},
					Domain{Name: domain3Name, GUID: domain3Guid, OrganizationGUID: sharedFromOrgGuid},
				))
				Expect(warnings).To(ConsistOf("some-domain-warning"))

			})
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationDomainsReturns(
					[]ccv3.Domain{},
					ccv3.Warnings{"some-domain-warning"},
					errors.New("list-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(MatchError("list-error"))
				Expect(warnings).To(ConsistOf("some-domain-warning"))
				Expect(domains).To(ConsistOf([]Domain{}))

				Expect(fakeCloudControllerClient.GetOrganizationDomainsCallCount()).To(Equal(1))
			})

		})
	})

	Describe("get domain by name", func() {
		var (
			ccv3Domains []ccv3.Domain
			domain      Domain

			domain1Name string
			domain1Guid string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			ccv3Domains = []ccv3.Domain{
				{Name: domain1Name, GUID: domain1Guid},
			}
		})

		JustBeforeEach(func() {
			domain, warnings, executeErr = actor.GetDomainByName(domain1Name)
		})

		When("the API layer call is successful", func() {
			expectedQuery := []ccv3.Query{{Key: ccv3.NameFilter, Values: []string{domain1Name}}}
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					ccv3Domains,
					ccv3.Warnings{"some-domain-warning"},
					nil,
				)
			})

			It("returns back the domains and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				actualQuery := fakeCloudControllerClient.GetDomainsArgsForCall(0)
				Expect(actualQuery).To(Equal(expectedQuery))

				Expect(domain).To(Equal(
					Domain{Name: domain1Name, GUID: domain1Guid},
				))
				Expect(warnings).To(ConsistOf("some-domain-warning"))

			})
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]ccv3.Domain{},
					ccv3.Warnings{"some-domain-warning"},
					errors.New("list-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(MatchError("list-error"))
				Expect(warnings).To(ConsistOf("some-domain-warning"))
				Expect(domain).To(Equal(Domain{}))

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
			})
		})
	})

	Describe("share private domain to org", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetOrganizationsReturns(
				[]ccv3.Organization{
					{GUID: "org-guid"},
				},
				ccv3.Warnings{"get-orgs-warning"},
				nil,
			)

			fakeCloudControllerClient.GetDomainsReturns(
				[]ccv3.Domain{
					{Name: "private-domain.com", GUID: "private-domain-guid"},
				},
				ccv3.Warnings{"get-domains-warning"},
				nil,
			)

			fakeCloudControllerClient.SharePrivateDomainToOrgsReturns(
				ccv3.Warnings{"share-domain-warning"},
				nil,
			)
		})

		It("delegates to the cloud controller client", func() {
			warnings, executeErr := actor.SharePrivateDomain("private-domain.com", "org-name")
			Expect(executeErr).To(BeNil())
			Expect(warnings).To(ConsistOf("share-domain-warning", "get-orgs-warning", "get-domains-warning"))

			Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
			actualQuery := fakeCloudControllerClient.GetDomainsArgsForCall(0)
			Expect(actualQuery).To(Equal([]ccv3.Query{{Key: ccv3.NameFilter, Values: []string{"private-domain.com"}}}))

			Expect(fakeCloudControllerClient.SharePrivateDomainToOrgsCallCount()).To(Equal(1))
			domainGuid, sharedOrgs := fakeCloudControllerClient.SharePrivateDomainToOrgsArgsForCall(0)
			Expect(domainGuid).To(Equal("private-domain-guid"))
			Expect(sharedOrgs).To(Equal(ccv3.SharedOrgs{GUIDs: []string{"org-guid"}}))
		})
	})
})
