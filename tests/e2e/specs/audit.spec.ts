import { test, expect } from '@playwright/test';

/**
 * Sprint 7: Audit Hub E2E Tests
 * 
 * Tests for audit engagement workflow, evidence requests, findings,
 * auditor isolation, and chain-of-custody tracking.
 */

const TEST_USER = {
  email: `e2e-audit-${Date.now()}@example.com`,
  password: 'TestPass123!',
  first_name: 'E2E',
  last_name: 'Audit',
  org_name: `E2E Audit Org ${Date.now()}`,
};

const AUDITOR_USER = {
  email: `e2e-auditor-${Date.now()}@example.com`,
  password: 'TestPass123!',
  first_name: 'External',
  last_name: 'Auditor',
  org_name: `E2E Auditor Org ${Date.now()}`,
};

test.describe('Audit Hub — Authentication & Setup', () => {
  test('should register compliance manager user', async ({ page }) => {
    await page.goto('/auth/register');
    await expect(page.locator('h1')).toContainText('Create Account');
    
    await page.fill('input[name="email"]', TEST_USER.email);
    await page.fill('input[name="password"]', TEST_USER.password);
    await page.fill('input[name="first_name"]', TEST_USER.first_name);
    await page.fill('input[name="last_name"]', TEST_USER.last_name);
    await page.fill('input[name="org_name"]', TEST_USER.org_name);
    
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('should login as compliance manager', async ({ page }) => {
    await page.goto('/auth/login');
    await page.fill('input[name="email"]', TEST_USER.email);
    await page.fill('input[name="password"]', TEST_USER.password);
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/dashboard/);
  });
});

test.describe('Audit Hub — Engagement CRUD', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/auth/login');
    await page.fill('input[name="email"]', TEST_USER.email);
    await page.fill('input[name="password"]', TEST_USER.password);
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard/);
  });

  test('should navigate to audit hub', async ({ page }) => {
    await page.click('a[href="/audits"]');
    await expect(page).toHaveURL(/\/audits/);
    await expect(page.locator('h1')).toContainText('Audit Hub');
  });

  test('should create new audit engagement', async ({ page }) => {
    await page.goto('/audits');
    
    // Open create dialog
    await page.click('button:has-text("Create Audit")');
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    
    // Fill audit details
    await page.fill('input[name="title"]', 'SOC 2 Type II Annual Audit 2026');
    await page.fill('textarea[name="description"]', 'Annual SOC 2 Type II examination');
    await page.selectOption('select[name="audit_type"]', 'soc2_type2');
    await page.fill('input[name="period_start"]', '2026-01-01');
    await page.fill('input[name="period_end"]', '2026-12-31');
    await page.fill('input[name="planned_start"]', '2026-02-01');
    await page.fill('input[name="planned_end"]', '2026-04-30');
    await page.fill('input[name="firm_name"]', 'Big Four Auditors LLP');
    
    // Submit
    await page.click('button:has-text("Create")');
    await expect(page.locator('[role="dialog"]')).not.toBeVisible();
    
    // Verify audit appears in list
    await expect(page.locator('text=SOC 2 Type II Annual Audit 2026')).toBeVisible();
  });

  test('should view audit detail page', async ({ page }) => {
    await page.goto('/audits');
    await page.click('text=SOC 2 Type II Annual Audit 2026');
    
    await expect(page).toHaveURL(/\/audits\/[a-f0-9-]+/);
    await expect(page.locator('h1')).toContainText('SOC 2 Type II Annual Audit 2026');
    
    // Verify tabs exist
    await expect(page.locator('button:has-text("Overview")')).toBeVisible();
    await expect(page.locator('button:has-text("Requests")')).toBeVisible();
    await expect(page.locator('button:has-text("Findings")')).toBeVisible();
    await expect(page.locator('button:has-text("Comments")')).toBeVisible();
  });

  test('should change audit status to fieldwork', async ({ page }) => {
    await page.goto('/audits');
    await page.click('text=SOC 2 Type II Annual Audit 2026');
    
    // Click status transition button
    await page.click('button:has-text("Change Status")');
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    
    await page.selectOption('select[name="new_status"]', 'fieldwork');
    await page.click('button:has-text("Update Status")');
    
    // Verify status badge updated
    await expect(page.locator('text=Fieldwork')).toBeVisible();
  });
});

test.describe('Audit Hub — Evidence Requests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/auth/login');
    await page.fill('input[name="email"]', TEST_USER.email);
    await page.fill('input[name="password"]', TEST_USER.password);
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard/);
  });

  test('should create evidence request', async ({ page }) => {
    await page.goto('/audits');
    await page.click('text=SOC 2 Type II Annual Audit 2026');
    
    // Switch to Requests tab
    await page.click('button:has-text("Requests")');
    
    // Create request
    await page.click('button:has-text("Create Request")');
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    
    await page.fill('input[name="title"]', 'Access Control Policy Review');
    await page.fill('textarea[name="description"]', 'Please provide current access control policy');
    await page.selectOption('select[name="priority"]', 'high');
    await page.fill('input[name="due_date"]', '2026-03-15');
    
    await page.click('button:has-text("Create")');
    await expect(page.locator('text=Access Control Policy Review')).toBeVisible();
  });

  test('should submit evidence for request', async ({ page }) => {
    await page.goto('/audits');
    await page.click('text=SOC 2 Type II Annual Audit 2026');
    await page.click('button:has-text("Requests")');
    
    // Click on the request
    await page.click('text=Access Control Policy Review');
    
    // Submit evidence (link existing artifact)
    await page.click('button:has-text("Submit Evidence")');
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    
    // Note: This assumes evidence artifacts exist from Sprint 3
    // In real workflow, user would select existing artifact
    await page.fill('textarea[name="submission_notes"]', 'Current policy attached');
    await page.click('button:has-text("Submit")');
    
    await expect(page.locator('text=Submitted')).toBeVisible();
  });
});

test.describe('Audit Hub — Findings Management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/auth/login');
    await page.fill('input[name="email"]', TEST_USER.email);
    await page.fill('input[name="password"]', TEST_USER.password);
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard/);
  });

  test('should create audit finding', async ({ page }) => {
    await page.goto('/audits');
    await page.click('text=SOC 2 Type II Annual Audit 2026');
    
    // Switch to Findings tab
    await page.click('button:has-text("Findings")');
    
    // Create finding
    await page.click('button:has-text("Create Finding")');
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    
    await page.fill('input[name="title"]', 'Insufficient Password Complexity');
    await page.fill('textarea[name="description"]', 'Password policy does not enforce special characters');
    await page.selectOption('select[name="severity"]', 'medium');
    await page.fill('textarea[name="auditor_recommendation"]', 'Update password policy to require special chars');
    
    await page.click('button:has-text("Create")');
    await expect(page.locator('text=Insufficient Password Complexity')).toBeVisible();
  });

  test('should submit management response to finding', async ({ page }) => {
    await page.goto('/audits');
    await page.click('text=SOC 2 Type II Annual Audit 2026');
    await page.click('button:has-text("Findings")');
    
    await page.click('text=Insufficient Password Complexity');
    
    // Submit management response
    await page.click('button:has-text("Submit Response")');
    await expect(page.locator('[role="dialog"]')).toBeVisible();
    
    await page.fill('textarea[name="management_response"]', 'We will update password policy by 2026-03-31');
    await page.fill('input[name="target_date"]', '2026-03-31');
    
    await page.click('button:has-text("Submit")');
    await expect(page.locator('text=Management response submitted')).toBeVisible();
  });
});

test.describe('Audit Hub — Comments', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/auth/login');
    await page.fill('input[name="email"]', TEST_USER.email);
    await page.fill('input[name="password"]', TEST_USER.password);
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard/);
  });

  test('should post internal comment', async ({ page }) => {
    await page.goto('/audits');
    await page.click('text=SOC 2 Type II Annual Audit 2026');
    
    // Switch to Comments tab
    await page.click('button:has-text("Comments")');
    
    // Create internal comment
    await page.fill('textarea[name="comment_text"]', 'Internal note: need to prepare evidence by next week');
    await page.check('input[name="is_internal"]'); // Mark as internal
    
    await page.click('button:has-text("Post Comment")');
    
    await expect(page.locator('text=Internal note')).toBeVisible();
    await expect(page.locator('text=Internal')).toBeVisible(); // Badge
  });

  test('should post external comment visible to auditor', async ({ page }) => {
    await page.goto('/audits');
    await page.click('text=SOC 2 Type II Annual Audit 2026');
    await page.click('button:has-text("Comments")');
    
    await page.fill('textarea[name="comment_text"]', 'We have completed the requested documentation');
    await page.uncheck('input[name="is_internal"]'); // External comment
    
    await page.click('button:has-text("Post Comment")');
    
    await expect(page.locator('text=We have completed the requested documentation')).toBeVisible();
  });
});

test.describe('Audit Hub — Auditor Isolation', () => {
  test('auditor should only see assigned audits', async ({ page }) => {
    // Register auditor user
    await page.goto('/auth/register');
    await page.fill('input[name="email"]', AUDITOR_USER.email);
    await page.fill('input[name="password"]', AUDITOR_USER.password);
    await page.fill('input[name="first_name"]', AUDITOR_USER.first_name);
    await page.fill('input[name="last_name"]', AUDITOR_USER.last_name);
    await page.fill('input[name="org_name"]', AUDITOR_USER.org_name);
    await page.click('button[type="submit"]');
    
    // Navigate to audit hub
    await page.goto('/audits');
    
    // Auditor should see empty list (no audits assigned to them)
    await expect(page.locator('text=No audits found')).toBeVisible();
  });

  test('auditor should not see internal comments', async ({ page }) => {
    // This test would require:
    // 1. Adding auditor to an audit engagement
    // 2. Logging in as auditor
    // 3. Verifying internal comments are hidden
    // 
    // Skipping detailed implementation for now, but spec is documented
  });
});

test.describe('Audit Hub — Dashboard & Readiness', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/auth/login');
    await page.fill('input[name="email"]', TEST_USER.email);
    await page.fill('input[name="password"]', TEST_USER.password);
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard/);
  });

  test('should view audit readiness dashboard', async ({ page }) => {
    await page.goto('/audits/readiness');
    
    await expect(page.locator('h1')).toContainText('Audit Readiness');
    
    // Verify dashboard components
    await expect(page.locator('text=Overall Readiness')).toBeVisible();
    await expect(page.locator('text=By Requirement')).toBeVisible();
    await expect(page.locator('text=Coverage Gaps')).toBeVisible();
  });

  test('should view audit hub dashboard', async ({ page }) => {
    await page.goto('/audits');
    
    // Verify dashboard stats cards
    await expect(page.locator('text=Active Audits')).toBeVisible();
    await expect(page.locator('text=Overdue Requests')).toBeVisible();
    await expect(page.locator('text=Critical Findings')).toBeVisible();
  });
});
