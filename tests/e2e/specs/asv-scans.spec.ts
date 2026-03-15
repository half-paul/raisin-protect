/**
 * E2E tests for ASV Scan Management (PCI DSS v4.0.1 Req 11.3.2).
 *
 * Prerequisites: test org seeded with compliance_manager user.
 * These tests exercise the full stack: UI → API → database.
 */
import { test, expect } from "@playwright/test";

const LOGIN_EMAIL = "compliance@acme-test.com";
const LOGIN_PASSWORD = "TestPassword123!";

async function login(page: any) {
  await page.goto("/login");
  await page.fill('[name="email"]', LOGIN_EMAIL);
  await page.fill('[name="password"]', LOGIN_PASSWORD);
  await page.click('[type="submit"]');
  await expect(page).toHaveURL(/\/dashboard/);
}

// ---------------------------------------------------------------------------
// Navigation
// ---------------------------------------------------------------------------

test("can navigate to ASV Scans from sidebar", async ({ page }) => {
  await login(page);
  await page.click("text=ASV Scans");
  await expect(page).toHaveURL(/\/asv-scans/);
  await expect(page.locator("h1")).toContainText("ASV Scans");
});

// ---------------------------------------------------------------------------
// Happy Path — CRUD
// ---------------------------------------------------------------------------

test("can create an ASV scan record manually", async ({ page }) => {
  await login(page);
  await page.goto("/asv-scans");

  await page.click("text=Add Scan");
  await page.fill('[name="asv_vendor"]', "Trustwave");
  await page.selectOption('[name="scan_type"]', "external");
  await page.selectOption('[name="quarter"]', "1");
  await page.fill('[name="year"]', "2026");
  await page.fill('[name="scan_date"]', "2026-03-01");
  await page.click('[type="submit"]');

  await expect(page.locator("text=Trustwave")).toBeVisible();
  await expect(page.locator("text=Q1 2026")).toBeVisible();
});

test("can view ASV scan details", async ({ page }) => {
  await login(page);
  await page.goto("/asv-scans");

  // Click first scan in the list.
  await page.click("table tbody tr:first-child");

  await expect(page.locator('[data-testid="scan-detail"]')).toBeVisible();
  await expect(page.locator('[data-testid="scan-vendor"]')).not.toBeEmpty();
  await expect(page.locator('[data-testid="scan-status"]')).toBeVisible();
});

test("can update scan status to pass", async ({ page }) => {
  await login(page);
  await page.goto("/asv-scans");
  await page.click("table tbody tr:first-child");

  await page.click("text=Edit");
  await page.selectOption('[name="status"]', "pass");
  await page.fill('[name="findings_count"]', "15");
  await page.fill('[name="critical_count"]', "0");
  await page.fill('[name="high_count"]', "3");
  await page.click('[type="submit"]');

  await expect(page.locator("text=Pass")).toBeVisible();
});

test("can delete an ASV scan record", async ({ page }) => {
  await login(page);

  // Create a disposable scan first via API to avoid test interdependency.
  const resp = await page.request.post("/api/v1/asv-scans", {
    data: {
      asv_vendor: "Disposable Vendor",
      scan_type: "external",
      quarter: 4,
      year: 2024,
      scan_date: "2024-12-01T00:00:00Z",
    },
  });
  expect(resp.ok()).toBeTruthy();
  const { data: created } = await resp.json();

  await page.goto(`/asv-scans/${created.id}`);
  await page.click("text=Delete");
  await page.click("text=Confirm");

  await expect(page).toHaveURL(/\/asv-scans$/);
  await expect(page.locator(`text=Disposable Vendor`)).not.toBeVisible();
});

// ---------------------------------------------------------------------------
// Quarterly Status View
// ---------------------------------------------------------------------------

test("quarterly status page shows all 4 quarters for current year", async ({
  page,
}) => {
  await login(page);
  await page.goto("/asv-scans?view=quarterly");

  // Should display Q1-Q4 for current year.
  await expect(page.locator("text=Q1")).toBeVisible();
  await expect(page.locator("text=Q2")).toBeVisible();
  await expect(page.locator("text=Q3")).toBeVisible();
  await expect(page.locator("text=Q4")).toBeVisible();
});

test("quarterly status shows missing quarters as Not Recorded", async ({
  page,
}) => {
  await login(page);
  await page.goto("/asv-scans?view=quarterly");

  // Quarters with no scan should show a clear 'Not Recorded' indicator.
  const notRecorded = page.locator('[data-testid="quarter-status-missing"]');
  // At least one quarter should be unrecorded in a fresh environment.
  await expect(notRecorded.first()).toBeVisible();
});

test("passing scan is shown in green on quarterly view", async ({ page }) => {
  await login(page);
  await page.goto("/asv-scans?view=quarterly");

  const passIndicator = page.locator('[data-status="pass"]').first();
  if (await passIndicator.isVisible()) {
    // Verify it has the passing visual indicator.
    await expect(passIndicator).toHaveClass(/pass|green|success/);
  }
});

test("failing scan is shown with alert on quarterly view", async ({ page }) => {
  await login(page);
  await page.goto("/asv-scans?view=quarterly");

  const failIndicator = page.locator('[data-status="fail"]').first();
  if (await failIndicator.isVisible()) {
    await expect(failIndicator).toHaveClass(/fail|red|danger/);
  }
});

// ---------------------------------------------------------------------------
// Import Flow
// ---------------------------------------------------------------------------

test("can import scan via manual entry form", async ({ page }) => {
  await login(page);
  await page.goto("/asv-scans");

  await page.click("text=Import Scan");
  await page.fill('[name="asv_vendor"]', "Qualys");
  await page.selectOption('[name="import_format"]', "manual");
  await page.selectOption('[name="quarter"]', "2");
  await page.fill('[name="year"]', "2026");
  await page.fill('[name="scan_date"]', "2026-06-15");
  await page.click('[type="submit"]');

  await expect(page.locator("text=Qualys")).toBeVisible();
});

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

test("shows validation error for invalid quarter", async ({ page }) => {
  await login(page);
  await page.goto("/asv-scans");
  await page.click("text=Add Scan");

  await page.fill('[name="asv_vendor"]', "X");
  await page.fill('[name="year"]', "2026");
  await page.fill('[name="scan_date"]', "2026-01-01");
  // Leave quarter blank / set to invalid.
  await page.click('[type="submit"]');

  await expect(page.locator('[data-testid="error-quarter"]')).toBeVisible();
});

test("shows validation error for duplicate quarter scan", async ({ page }) => {
  await login(page);

  // Create a scan for Q3 2026 first.
  await page.request.post("/api/v1/asv-scans", {
    data: {
      asv_vendor: "Duplicate Vendor",
      scan_type: "external",
      quarter: 3,
      year: 2026,
      scan_date: "2026-09-01T00:00:00Z",
    },
  });

  await page.goto("/asv-scans");
  await page.click("text=Add Scan");
  await page.fill('[name="asv_vendor"]', "Another Vendor");
  await page.selectOption('[name="scan_type"]', "external");
  await page.selectOption('[name="quarter"]', "3");
  await page.fill('[name="year"]', "2026");
  await page.fill('[name="scan_date"]', "2026-09-15");
  await page.click('[type="submit"]');

  await expect(
    page.locator("text=scan already exists for this quarter")
  ).toBeVisible();
});

// ---------------------------------------------------------------------------
// RBAC
// ---------------------------------------------------------------------------

test("auditor can view ASV scans but not create", async ({ page }) => {
  // Log in as auditor.
  await page.goto("/login");
  await page.fill('[name="email"]', "auditor@acme-test.com");
  await page.fill('[name="password"]', "TestPassword123!");
  await page.click('[type="submit"]');
  await expect(page).toHaveURL(/\/dashboard/);

  await page.goto("/asv-scans");
  await expect(page.locator("h1")).toContainText("ASV Scans");

  // 'Add Scan' button must not be visible for auditors.
  await expect(page.locator("text=Add Scan")).not.toBeVisible();
});
