package tenant

// Demo identifies the seeded showcase store ("SYSTEM ARCHIVE"). Every service seeds its demo data
// under this fixed tenant id so the docker-compose smoke test and the original storefront keep working
// after the multi-tenant pivot. The storefront resolves it by DemoSlug.
const (
	DemoID   = "000000000000000000000001"
	DemoSlug = "system-archive"
	DemoName = "SYSTEM ARCHIVE"
)
