/**
 * E2E Integration Tests: Compensating Controls Worksheet
 *
 * Coverage:
 *   - Creating a compensating control (is_compensating=true + worksheet fields)
 *   - Retrieving compensating control worksheets
 *   - Updating worksheet fields (business_constraint, objective_met, risk_maintained)
 *   - Validation: required worksheet fields when is_compensating=true
 *   - Listing controls filtered by is_compensating
 *   - Org-scoping: cannot access another org's compensating controls
 *   - Regression: existing controls (is_compensating=false) are unaffected
 *
 * Depends on: LEX-86 (Compensating Controls API) + LEX-83 schema (is_compensating
 * flag + worksheet fields on controls table). Tests are spec-first.
 *
 * PCI DSS coverage: Appendix B (Compensating Controls)
 *
 * FIELD NAME NOTE: The API uses `compensating_worksheet` (not `compensating_control_worksheet`).
 * Matches the `json:"compensating_worksheet"` tag on both the Control model and
 * UpdateControlRequest. CreateControlRequest also needs `is_compensating` + `compensating_worksheet`
 * added by Logan (LEX-86) before create-path tests will pass.
 *
 * BACKEND GAPS (Logan — LEX-86):
 *   1. Add `is_compensating` and `compensating_worksheet` to CreateControlRequest struct
 *   2. Add `is_compensating`, `compensating_worksheet` to ListControls SELECT query
 *   3. Add `is_compensating`, `compensating_worksheet` to GetControl SELECT query
 *   4. Implement validation: require worksheet when is_compensating=true (create + update)
 */

import { test, expect } from '@playwright/test';

const BASE_URL = 'http://localhost:8090';

// ---- helpers ----------------------------------------------------------------

async function registerOrg(request: any, label: string) {
  const res = await request.post(`${BASE_URL}/api/v1/auth/register`, {
    data: {
      email: `cc-test-${label}-${Date.now()}@example.com`,
      password: 'Test1234!',
      first_name: 'QA',
      last_name: 'Tester',
      org_name: `Compensating Controls QA Org ${label}`,
      role: 'compliance_manager',
    },
  });
  expect(res.ok(), `register ${label}: ${await res.text()}`).toBeTruthy();
  const body = await res.json();
  return {
    token: body.data.access_token as string,
    orgId: body.data.organization.id as string,
  };
}

function authHeaders(token: string) {
  return { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' };
}

// ---- shared state -----------------------------------------------------------

let orgA: { token: string; orgId: string };
let orgB: { token: string; orgId: string };

let standardControlId: string;  // is_compensating = false
let compensatingControlId: string; // is_compensating = true

test.beforeAll(async ({ request }) => {
  [orgA, orgB] = await Promise.all([
    registerOrg(request, 'ccA'),
    registerOrg(request, 'ccB'),
  ]);
});

// =============================================================================
// COMPENSATING CONTROL CREATION
// =============================================================================

test.describe('Compensating Controls — creation', () => {
  test('should create a standard control (is_compensating defaults to false)', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/controls`, {
      headers: authHeaders(orgA.token),
      data: {
        identifier: `CTRL-STD-${Date.now()}`,
        title: 'Standard Network Firewall Control',
        description: 'Firewall policy enforcing CDE perimeter segmentation.',
        category: 'technical',
      },
    });

    expect(res.status()).toBe(201);
    const body = await res.json();
    // is_compensating must default to false — no worksheet fields populated
    expect(body.data.is_compensating).toBe(false);
    expect(body.data.compensating_worksheet).toBeNull();
    standardControlId = body.data.id;
  });

  test('should create a compensating control with worksheet', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/controls`, {
      headers: authHeaders(orgA.token),
      data: {
        identifier: `CTRL-CC-${Date.now()}`,
        title: 'Compensating: Enhanced Monitoring Instead of Patch',
        description: 'Compensating control using enhanced SIEM monitoring in lieu of applying unsupported legacy OS patch.',
        category: 'technical',
        is_compensating: true,
        compensating_worksheet: {
          // PCI DSS Appendix B fields
          business_constraint: 'Legacy POS system cannot be patched without full hardware replacement; replacement scheduled for Q4 2026.',
          objective_met: 'Continuous SIEM monitoring detects unauthorized access within 5 minutes, meeting the intent of Req 6.3.3.',
          additional_security: 'Network micro-segmentation, IDS on the legacy segment, monthly penetration testing.',
          risk_maintained: true,
          review_date: '2026-06-15',
        },
      },
    });

    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.is_compensating).toBe(true);

    const worksheet = body.data.compensating_worksheet;
    expect(worksheet).not.toBeNull();
    expect(worksheet.business_constraint).toContain('Legacy POS');
    expect(worksheet.risk_maintained).toBe(true);
    expect(worksheet.review_date).toBe('2026-06-15');

    compensatingControlId = body.data.id;
    expect(compensatingControlId).toBeTruthy();
  });

  test('should require worksheet when is_compensating=true', async ({ request }) => {
    // is_compensating=true without worksheet fields must fail validation
    const res = await request.post(`${BASE_URL}/api/v1/controls`, {
      headers: authHeaders(orgA.token),
      data: {
        identifier: `CTRL-CC-INVALID-${Date.now()}`,
        title: 'Missing Worksheet',
        description: 'Should fail without worksheet.',
        category: 'technical',
        is_compensating: true,
        // no compensating_worksheet provided
      },
    });

    expect(res.status()).toBe(400);
    const body = await res.json();
    expect(body.error).toMatch(/worksheet|compensating/i);
  });

  test('should require business_constraint in worksheet', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/controls`, {
      headers: authHeaders(orgA.token),
      data: {
        identifier: `CTRL-CC-PARTIAL-${Date.now()}`,
        title: 'Partial Worksheet',
        description: 'Missing business_constraint.',
        category: 'technical',
        is_compensating: true,
        compensating_worksheet: {
          // missing business_constraint — required per PCI Appendix B
          objective_met: 'Some objective',
          risk_maintained: true,
        },
      },
    });

    expect(res.status()).toBe(400);
  });
});

// =============================================================================
// COMPENSATING CONTROL RETRIEVAL
// =============================================================================

test.describe('Compensating Controls — retrieval', () => {
  test('should get compensating control with worksheet fields', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/controls/${compensatingControlId}`,
      { headers: authHeaders(orgA.token) },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.is_compensating).toBe(true);
    expect(body.data.compensating_worksheet).not.toBeNull();
    expect(body.data.compensating_worksheet.business_constraint).toBeTruthy();
  });

  test('should list controls filtered by is_compensating=true', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/controls?is_compensating=true`,
      { headers: authHeaders(orgA.token) },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(Array.isArray(body.data)).toBeTruthy();
    expect(body.data.length).toBeGreaterThanOrEqual(1);

    // Every returned control must have is_compensating=true
    for (const ctrl of body.data) {
      expect(ctrl.is_compensating).toBe(true);
    }
  });

  test('should list controls filtered by is_compensating=false', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/controls?is_compensating=false`,
      { headers: authHeaders(orgA.token) },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    for (const ctrl of body.data) {
      expect(ctrl.is_compensating).toBe(false);
    }
  });
});

// =============================================================================
// WORKSHEET UPDATE
// =============================================================================

test.describe('Compensating Controls — worksheet update', () => {
  test('should update worksheet fields on a compensating control', async ({ request }) => {
    const res = await request.put(
      `${BASE_URL}/api/v1/controls/${compensatingControlId}`,
      {
        headers: authHeaders(orgA.token),
        data: {
          compensating_worksheet: {
            business_constraint: 'Updated: Full system replacement now scheduled Q2 2026.',
            objective_met: 'Updated objective met statement.',
            additional_security: 'Added quarterly pen testing to compensating controls.',
            risk_maintained: true,
            review_date: '2026-04-15',
          },
        },
      },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.compensating_worksheet.review_date).toBe('2026-04-15');
    expect(body.data.compensating_worksheet.business_constraint).toContain('Q2 2026');
  });

  test('should convert standard control to compensating with worksheet', async ({ request }) => {
    // An existing standard control can be upgraded to compensating
    const res = await request.put(
      `${BASE_URL}/api/v1/controls/${standardControlId}`,
      {
        headers: authHeaders(orgA.token),
        data: {
          is_compensating: true,
          compensating_worksheet: {
            business_constraint: 'Converted to compensating due to new risk assessment.',
            objective_met: 'Enhanced monitoring meets control objective.',
            risk_maintained: true,
            review_date: '2026-12-31',
          },
        },
      },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.is_compensating).toBe(true);
    expect(body.data.compensating_worksheet).not.toBeNull();
  });
});

// =============================================================================
// ORG-SCOPING SECURITY TESTS
// =============================================================================

test.describe('Compensating Controls — org-scoping enforcement', () => {
  test('orgB cannot read orgA compensating control worksheet', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/controls/${compensatingControlId}`,
      { headers: authHeaders(orgB.token) },
    );
    // Must return 403 — cross-org access blocked
    expect(res.status()).toBe(403);
  });

  test('orgB worksheet list does not include orgA controls', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/controls?is_compensating=true`,
      { headers: authHeaders(orgB.token) },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();

    // orgB has no compensating controls — should return empty
    expect(body.data).toEqual([]);

    // Verify orgA's compensating control ID is not present
    const ids = body.data.map((c: any) => c.id);
    expect(ids).not.toContain(compensatingControlId);
  });
});

// =============================================================================
// EDGE CASES
// =============================================================================

test.describe('Compensating Controls — edge cases', () => {
  test('should handle empty business_constraint string as invalid', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/controls`, {
      headers: authHeaders(orgA.token),
      data: {
        identifier: `CTRL-CC-EMPTY-${Date.now()}`,
        title: 'Empty Worksheet',
        description: 'Test empty string validation.',
        category: 'technical',
        is_compensating: true,
        compensating_worksheet: {
          business_constraint: '', // empty string — should fail
          objective_met: 'Objective met',
          risk_maintained: true,
        },
      },
    });
    expect(res.status()).toBe(400);
  });

  test('should cap worksheet text fields at max length', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/controls`, {
      headers: authHeaders(orgA.token),
      data: {
        identifier: `CTRL-CC-LONG-${Date.now()}`,
        title: 'Long Worksheet',
        description: 'Test max length.',
        category: 'technical',
        is_compensating: true,
        compensating_worksheet: {
          business_constraint: 'X'.repeat(5001), // over 5000-char limit
          objective_met: 'Objective met.',
          risk_maintained: true,
          review_date: '2026-12-31',
        },
      },
    });
    expect(res.status()).toBe(400);
  });

  test('should accept review_date as null (not yet scheduled)', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/controls`, {
      headers: authHeaders(orgA.token),
      data: {
        identifier: `CTRL-CC-NODATE-${Date.now()}`,
        title: 'No Review Date',
        description: 'Review date is optional.',
        category: 'technical',
        is_compensating: true,
        compensating_worksheet: {
          business_constraint: 'Technical constraint preventing full implementation.',
          objective_met: 'Objective met through alternative means.',
          risk_maintained: true,
          review_date: null,
        },
      },
    });
    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.compensating_worksheet.review_date).toBeNull();
  });
});
