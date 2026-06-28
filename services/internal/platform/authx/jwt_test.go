package authx

import (
	"testing"
	"time"
)

// TestTokenAudiences pins the buyer↔merchant isolation: each issuer method stamps the right audience and
// binds the right scope. The gateway rejects a customer-audience token on merchant/admin routes and
// requires it on customer routes, so getting these claims wrong would let a shopper's token reach a
// merchant API (or vice versa). A round-trip through Verify guards the encoding too.
func TestTokenAudiences(t *testing.T) {
	iss := NewIssuer("test-secret", time.Hour)

	merchTok, err := iss.IssueMerchant("m1", RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	storeTok, err := iss.IssueStore("m1", "store1", "pro")
	if err != nil {
		t.Fatal(err)
	}
	custTok, err := iss.IssueCustomer("c1", "store1")
	if err != nil {
		t.Fatal(err)
	}

	merch, err := iss.Verify(merchTok)
	if err != nil || merch.Aud != AudMerchant || merch.Role != RoleAdmin {
		t.Fatalf("merchant token: aud=%q role=%q err=%v", merch.Aud, merch.Role, err)
	}
	store, err := iss.Verify(storeTok)
	if err != nil || store.Aud != AudMerchant || store.StoreID != "store1" || store.Plan != "pro" {
		t.Fatalf("store token: aud=%q store=%q plan=%q err=%v", store.Aud, store.StoreID, store.Plan, err)
	}
	cust, err := iss.Verify(custTok)
	if err != nil || cust.Aud != AudCustomer || cust.TenantID != "store1" || cust.MerchantID != "c1" {
		t.Fatalf("customer token: aud=%q tnt=%q sub=%q err=%v", cust.Aud, cust.TenantID, cust.MerchantID, err)
	}

	// The decisive invariant the gateway relies on: a customer token is never the merchant audience, so
	// the `claims.Aud == AudCustomer` reject in the merchant middlewares fires.
	if cust.Aud == AudMerchant {
		t.Fatal("customer token must not carry the merchant audience")
	}
}
