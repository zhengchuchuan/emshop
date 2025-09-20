package sentinel_test

import (
	"github.com/alibaba/sentinel-golang/core/flow"

	sentinel "emshop/pkg/sentinel"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BusinessRules", func() {
	var br *sentinel.BusinessRules

	BeforeEach(func() {
		br = sentinel.DefaultBusinessRules()
	})

	Describe("Flash sale service", func() {
		It("exposes the expected QPS ceilings", func() {
			Expect(br.FlashSale).NotTo(BeNil())
			Expect(br.FlashSale.FlashSaleQPS).To(BeNumerically("==", 1000), "flash sale entry QPS")
			Expect(br.FlashSale.ProductDetailQPS).To(BeNumerically("==", 2000), "flash sale product detail QPS")
			Expect(br.FlashSale.OrderQPS).To(BeNumerically("==", 500), "flash sale ordering QPS")
		})
	})

	Describe("Payment service", func() {
		It("caps payment interfaces according to defaults", func() {
			Expect(br.Payment).NotTo(BeNil())
			Expect(br.Payment.PaymentQPS).To(BeNumerically("==", 200), "payment create QPS")
			Expect(br.Payment.PaymentQueryQPS).To(BeNumerically("==", 1000), "payment query QPS")
			Expect(br.Payment.RefundQPS).To(BeNumerically("==", 100), "refund QPS")
		})
	})

	Describe("Inventory service", func() {
		It("keeps stock operations within the configured thresholds", func() {
			Expect(br.Inventory).NotTo(BeNil())
			Expect(br.Inventory.DeductQPS).To(BeNumerically("==", 300), "inventory deduction QPS")
			Expect(br.Inventory.QueryQPS).To(BeNumerically("==", 2000), "inventory query QPS")
			Expect(br.Inventory.RestoreQPS).To(BeNumerically("==", 200), "inventory restore QPS")
		})
	})

	Describe("Coupon service", func() {
		It("defines baseline QPS for issue, use, and query endpoints", func() {
			Expect(br.Coupon).NotTo(BeNil())
			Expect(br.Coupon.IssueQPS).To(BeNumerically("==", 500), "coupon issue QPS")
			Expect(br.Coupon.UseQPS).To(BeNumerically("==", 800), "coupon use QPS")
			Expect(br.Coupon.QueryQPS).To(BeNumerically("==", 1500), "coupon query QPS")
		})
	})

	Describe("User service", func() {
		It("restricts login, registration, and lookup throughput", func() {
			Expect(br.User).NotTo(BeNil())
			Expect(br.User.LoginQPS).To(BeNumerically("==", 300), "user login QPS")
			Expect(br.User.RegisterQPS).To(BeNumerically("==", 100), "user registration QPS")
			Expect(br.User.QueryQPS).To(BeNumerically("==", 1000), "user lookup QPS")
		})
	})

	Describe("Goods service", func() {
		It("documents the read-heavy QPS ceilings", func() {
			Expect(br.Goods).NotTo(BeNil())
			Expect(br.Goods.ListQPS).To(BeNumerically("==", 2000), "goods list QPS")
			Expect(br.Goods.DetailQPS).To(BeNumerically("==", 3000), "goods detail QPS")
			Expect(br.Goods.SearchQPS).To(BeNumerically("==", 1500), "goods search QPS")
		})
	})

	Describe("Order service", func() {
		It("keeps core order workflow QPS aligned with defaults", func() {
			Expect(br.Order).NotTo(BeNil())
			Expect(br.Order.CreateQPS).To(BeNumerically("==", 500), "order create QPS")
			Expect(br.Order.QueryQPS).To(BeNumerically("==", 1500), "order query QPS")
			Expect(br.Order.CancelQPS).To(BeNumerically("==", 200), "order cancel QPS")
		})
	})

	Describe("Flow rule export", func() {
		It("maps coupon service thresholds to flow control rules", func() {
			rules, err := br.GenerateFlowRules("emshop-coupon-srv")
			Expect(err).NotTo(HaveOccurred())
			Expect(rules).To(HaveLen(3))

			thresholds := resourceThresholds(rules)
			Expect(thresholds).To(HaveKeyWithValue("coupon-srv:IssueCoupon", BeNumerically("==", br.Coupon.IssueQPS)))
			Expect(thresholds).To(HaveKeyWithValue("coupon-srv:UseCoupon", BeNumerically("==", br.Coupon.UseQPS)))
			Expect(thresholds).To(HaveKeyWithValue("coupon-srv:GetUserCoupons", BeNumerically("==", br.Coupon.QueryQPS)))
		})

		It("maps user service thresholds to flow control rules", func() {
			rules, err := br.GenerateFlowRules("emshop-user-srv")
			Expect(err).NotTo(HaveOccurred())
			Expect(rules).To(HaveLen(3))

			thresholds := resourceThresholds(rules)
			Expect(thresholds).To(HaveKeyWithValue("user-srv:CreateUser", BeNumerically("==", br.User.RegisterQPS)))
			Expect(thresholds).To(HaveKeyWithValue("user-srv:GetUserByMobile", BeNumerically("==", br.User.LoginQPS)))
			Expect(thresholds).To(HaveKeyWithValue("user-srv:GetUserById", BeNumerically("==", br.User.QueryQPS)))
		})

		It("maps inventory service thresholds to flow control rules", func() {
			rules, err := br.GenerateFlowRules("emshop-inventory-srv")
			Expect(err).NotTo(HaveOccurred())
			Expect(rules).To(HaveLen(2))

			thresholds := resourceThresholds(rules)
			Expect(thresholds).To(HaveKeyWithValue("inventory-srv:Sell", BeNumerically("==", br.Inventory.DeductQPS)))
			Expect(thresholds).To(HaveKeyWithValue("inventory-srv:InvDetail", BeNumerically("==", br.Inventory.QueryQPS)))
		})
	})
})

func resourceThresholds(rules []*flow.Rule) map[string]float64 {
	thresholds := make(map[string]float64, len(rules))
	for _, rule := range rules {
		thresholds[rule.Resource] = rule.Threshold
	}
	return thresholds
}
