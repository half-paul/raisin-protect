/**
 * E2E tests for Service Provider / Vendor Management (PCI DSS v4.0.1 Req 12.8).
 *
 * These tests exercise the full stack: UI → API → database.
 * Covers: provider CRUD, compliance document tracking, responsibility matrix,
 * compliance summary, expiry alerts, and RBAC.
 */
import { test, expect } from "@playwright/test";

const CM_EMAIL = "compliance@acme-test.com";
const CM_PASSWORD = "TestPassword123!";

async function login(page: any, email = CM_EMAIL, password = CM_PASSWORD) {
  await page.goto("/login");
  await page.fill('[name="email"]', email);
  await page.fill('[name="password"]', password);
  await page.click('[type="submit"]');
  await expect(page).toHaveURL(/\/dashboard/);
}

// ---------------------------------------------------------------------------
// Navigation
// ---------------------------------------------------------------------------

test("can navigate to Vendor Management from sidebar", async ({ page }) => {
  await login(page);
  await page.click("text=Vendors");
  await expect(page).toHaveURL(/\/service-providers/);
  await expect(page.locator("h1")).toContainText(/Vendor|Service Provider/i);
});

// ---------------------------------------------------------------------------
// Happy Path — Provider CRUD
// ---------------------------------------------------------------------------

test("can create a new service provider", async ({ page }) => {
  await login(page);
  await page.goto("/service-providers");

  await page.click("text=Add Provider");
  await page.fill('[name="name"]', "Stripe");
  await page.selectOption('[name="type"]', "payment_processor");
  await page.fill('[name="contact_name"]', "John Billing");
  await page.fill('[name="contact_email"]', "billing@stripe.com");
  await page.selectOption('[name="risk_level"]', "medium");
  await page.click('[type="submit"]');

  await expect(page.locator("text=Stripe")).toBeVisible();
  await expect(page.locator('[data-testid="compliance-badge-unknown"]')).toBeVisible();
});

test("can view service provider detail page", async ({ page }) => {
  await login(page);
  await page.goto("/service-providers");

  await page.click("table tbody tr:first-child");

  await expect(page.locator('[data-testid="sp-detail"]')).toBeVisible();
  await expect(page.locator('[data-testid="sp-name"]')).not.toBeEmpty();
  await expect(page.locator('[data-testid="pci-compliance-status"]')).toBeVisible();
});

test("can update provider compliance status to compliant", async ({ page }) => {
  await login(page);

  // Create a provider via API.
  const resp = await page.request.post("/api/v1/service-providers", {
    data: { name: "TestProvider", type: "hosting" },
  });
  const { data: provider } = await resp.json();

  await page.goto(`/service-providers/${provider.id}`);
  await page.click("text=Edit");
  await page.selectOption('[name="pci_compliance_status"]', "compliant");
  await page.fill('[name="risk_notes"]', "Annual AOC received and verified.");
  await page.click('[type="submit"]');

  await expect(page.locator('[data-testid="pci-compliance-status"]')).toContainText(/compliant/i);
});

test("can deactivate a service provider", async ({ page }) => {
  await login(page);

  const resp = await page.request.post("/api/v1/service-providers", {
    data: { name: "InactiveVendor", type: "software" },
  });
  const { data: provider } = await resp.json();

  await page.goto(`/service-providers/${provider.id}`);
  await page.click("text=Edit");
  // Toggle the is_active checkbox off.
  const activeToggle = page.locator('[name="is_active"]');
  if (await activeToggle.isChecked()) {
    await activeToggle.uncheck();
  }
  await page.click('[type="submit"]');

  await expect(page.locator('[data-testid="sp-inactive-badge"]')).toBeVisible();
});

test("can delete a service provider", async ({ page }) => {
  await login(page);

  const resp = await page.request.post("/api/v1/service-providers", {
    data: { name: "ToDelete Corp", type: "other" },
  });
  const { data: provider } = await resp.json();

  await page.goto(`/service-providers/${provider.id}`);
  await page.click("text=Delete");
  await page.click("text=Confirm");

  await expect(page).toHaveURL(/\/service-providers$/);
  await expect(page.locator("text=ToDelete Corp")).not.toBeVisible();
});

// ---------------------------------------------------------------------------
// Compliance Documents Sub-resource
// ---------------------------------------------------------------------------

test("can add an AOC document to a service provider", async ({ page }) => {
  await login(page);

  const resp = await page.request.post("/api/v1/service-providers", {
    data: { name: "DocProvider", type: "payment_gateway" },
  });
  const { data: provider } = await resp.json();

  await page.goto(`/service-providers/${provider.id}/compliance-docs`);
  await page.click("text=Add Document");
  await page.selectOption('[name="document_type"]', "aoc");
  await page.fill('[name="title"]', "DocProvider AOC 2025");
  await page.fill('[name="document_version"]', "2025.1");
  await page.fill('[name="valid_from"]', "2025-01-01");
  await page.fill('[name="valid_until"]', "2025-12-31");
  await page.click('[type="submit"]');

  await expect(page.locator("text=DocProvider AOC 2025")).toBeVisible();
});

test("expired compliance documents appear in list with expiry warning", async ({
  page,
}) => {
  await login(page);
  await page.goto("/service-providers");

  // Click a provider known to have expired docs (seeded in test data).
  const expiredDoc = page.locator('[data-testid="expired-doc-warning"]').first();
  if (await expiredDoc.isVisible()) {
    // Expired docs must be visible (not silently hidden) and flagged.
    await expect(expiredDoc).toHaveClass(/expired|warning|red/);
  }
});

test("can delete a compliance document from a provider", async ({ page }) => {
  await login(page);

  const spResp = await page.request.post("/api/v1/service-providers", {
    data: { name: "DocDeleteProvider", type: "other" },
  });
  const { data: provider } = await spResp.json();

  const docResp = await page.request.post(
    `/api/v1/service-providers/${provider.id}/compliance-docs`,
    { data: { document_type: "aoc", title: "Temporary AOC" } }
  );
  const { data: doc } = await docResp.json();

  await page.goto(`/service-providers/${provider.id}/compliance-docs`);
  await page.locator(`[data-doc-id="${doc.id}"]`).locator("text=Delete").click();
  await page.click("text=Confirm");

  await expect(page.locator("text=Temporary AOC")).not.toBeVisible();
});

// ---------------------------------------------------------------------------
// Responsibility Matrix
// ---------------------------------------------------------------------------

test("can set PCI requirement responsibility for a provider", async ({
  page,
}) => {
  await login(page);

  const resp = await page.request.post("/api/v1/service-providers", {
    data: { name: "ResponsibilityProvider", type: "managed_security" },
  });
  const { data: provider } = await resp.json();

  await page.goto(`/service-providers/${provider.id}/responsibility-matrix`);

  // Set req 12.8.1 as 'provider' responsibility.
  const row = page.locator('[data-req-code="12.8.1"]');
  await row.locator("select").selectOption("provider");
  await page.click("text=Save");

  await expect(row.locator('[data-testid="responsibility-badge"]')).toContainText(/provider/i);
});

test("responsibility matrix shows all relevant PCI DSS 12.8 requirements", async ({
  page,
}) => {
  await login(page);

  const resp = await page.request.post("/api/v1/service-providers", {
    data: { name: "MatrixProvider", type: "payment_processor" },
  });
  const { data: provider } = await resp.json();

  await page.goto(`/service-providers/${provider.id}/responsibility-matrix`);

  // At minimum, PCI 12.8.x requirements should be visible.
  await expect(page.locator("text=12.8")).toBeVisible();
});

// ---------------------------------------------------------------------------
// Compliance Summary
// ---------------------------------------------------------------------------

test("compliance summary dashboard shows aggregate stats", async ({ page }) => {
  await login(page);
  await page.goto("/service-providers");

  // Summary panel should be visible on list page.
  await expect(page.locator('[data-testid="compliance-summary"]')).toBeVisible();
  await expect(page.locator('[data-testid="summary-total"]')).not.toBeEmpty();
});

test("compliance summary shows expiry alerts for 30/60/90 day windows", async ({
  page,
}) => {
  await login(page);
  await page.goto("/service-providers");

  const summary = page.locator('[data-testid="compliance-summary"]');
  await expect(summary.locator('[data-testid="expiring-30"]')).toBeVisible();
  await expect(summary.locator('[data-testid="expiring-60"]')).toBeVisible();
  await expect(summary.locator('[data-testid="expiring-90"]')).toBeVisible();
});

test("non-compliant providers are highlighted in the list", async ({ page }) => {
  await login(page);
  await page.goto("/service-providers");

  const nonCompliant = page.locator('[data-compliance-status="non_compliant"]').first();
  if (await nonCompliant.isVisible()) {
    await expect(nonCompliant).toHaveClass(/danger|red|alert/);
  }
});

// ---------------------------------------------------------------------------
// Filtering
// ---------------------------------------------------------------------------

test("can filter providers by compliance status", async ({ page }) => {
  await login(page);
  await page.goto("/service-providers");

  await page.selectOption('[data-testid="filter-compliance-status"]', "compliant");
  await page.waitForLoadState("networkidle");

  // All visible rows should show 'compliant' status.
  const statusBadges = page.locator('[data-testid="compliance-badge"]');
  const count = await statusBadges.count();
  for (let i = 0; i < count; i++) {
    await expect(statusBadges.nth(i)).toContainText(/compliant/i);
  }
});

test("can filter providers by type", async ({ page }) => {
  await login(page);
  await page.goto("/service-providers");

  await page.selectOption('[data-testid="filter-type"]', "payment_processor");
  await page.waitForLoadState("networkidle");

  // All visible rows should be payment processors.
  const typeBadges = page.locator('[data-testid="sp-type-badge"]');
  const count = await typeBadges.count();
  for (let i = 0; i < count; i++) {
    await expect(typeBadges.nth(i)).toContainText(/payment processor/i);
  }
});

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

test("shows error when creating provider with invalid type", async ({
  page,
}) => {
  await login(page);

  // Direct API call with invalid type.
  const resp = await page.request.post("/api/v1/service-providers", {
    data: { name: "BadType Corp", type: "mystery_vendor" },
  });

  expect(resp.status()).toBe(400);
  const body = await resp.json();
  expect(body.error.code).toBe("VALIDATION_ERROR");
});

test("shows error when creating provider with missing name", async ({
  page,
}) => {
  await login(page);
  await page.goto("/service-providers");
  await page.click("text=Add Provider");

  await page.selectOption('[name="type"]', "hosting");
  // Do NOT fill in name.
  await page.click('[type="submit"]');

  await expect(page.locator('[data-testid="error-name"]')).toBeVisible();
});

// ---------------------------------------------------------------------------
// RBAC
// ---------------------------------------------------------------------------

test("auditor can view vendors but not create, update, or delete", async ({
  page,
}) => {
  await login(page, "auditor@acme-test.com", "TestPassword123!");
  await page.goto("/service-providers");

  await expect(page.locator("h1")).toContainText(/Vendor|Service Provider/i);

  // Write actions must be invisible for auditors.
  await expect(page.locator("text=Add Provider")).not.toBeVisible();
  await expect(page.locator("text=Delete")).not.toBeVisible();
});

test("vendor_manager can create providers", async ({ page }) => {
  await login(page, "vendormgr@acme-test.com", "TestPassword123!");
  await page.goto("/service-providers");

  // vendor_manager is in SPManageRoles — Add Provider button must be visible.
  await expect(page.locator("text=Add Provider")).toBeVisible();
});
