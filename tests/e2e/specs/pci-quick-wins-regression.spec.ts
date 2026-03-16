/**
 * E2E Regression Tests: PCI DSS Quick-Win Schema Changes
 *
 * Verifies that Sprint 11 quick-win schema additions do NOT break existing
 * functionality in controls, requirement_scopes, and audits.
 *
 * Quick wins covered (from LEX-78 plan):
 *   1. is_compensating flag on controls table
 *   2. customized_approach flag on requirement_scopes table
 *   3. New audit types: pci_dss_saq_a, pci_dss_saq_d, pci_dss_aoc
 *   4. vendor_management placeholder routes
 *
 * Each section follows the pattern:
 *   a) Existing behavior still works (true regression)
 *   b) New field/behavior works correctly
 *
 * Depends on: LEX-83 (schema migrations for quick-win fields)
 */

import { test, expect } from '@playwright/test';

const BASE_URL = 'http://localhost:8090';

// ---- helpers ----------------------------------------------------------------

async function registerOrg(request: any, label: string) {
  const res = await request.post(`${BASE_URL}/api/v1/auth/register`, {
    data: {
      email: `qw-test-${label}-${Date.now()}@example.com`,
      password: 'Test1234!',
      first_name: 'QA',
      last_name: 'Tester',
      org_name: `Quick Wins Regression Org ${label}`,
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

let org: { token: string; orgId: string };

test.beforeAll(async ({ request }) => {
  org = await registerOrg(request, 'qw');
});

// =============================================================================
// CONTROLS: is_compensating field regression
// =============================================================================

test.describe('Controls regression — is_compensating field addition', () => {
  let controlId: string;

  test('existing control creation still works (is_compensating omitted = false)', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/controls`, {
      headers: authHeaders(org.token),
      data: {
        identifier: `CTRL-REG-${Date.now()}`,
        title: 'Pre-existing Style Control',
        description: 'Created without new PCI fields — must still succeed.',
        category: 'technical',
      },
    });

    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.identifier).toBeTruthy();
    expect(body.data.status).toBe('draft');
    // New field must default to false and not break response schema
    expect(body.data.is_compensating).toBe(false);
    expect(body.data.compensating_control_worksheet).toBeNull();
    controlId = body.data.id;
  });

  test('existing control GET still returns all original fields intact', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/v1/controls/${controlId}`, {
      headers: authHeaders(org.token),
    });

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    // Original fields
    expect(body.data.id).toBe(controlId);
    expect(body.data.title).toBe('Pre-existing Style Control');
    expect(body.data.category).toBe('technical');
    expect(body.data.status).toBe('draft');
    expect(body.data.created_at).toBeTruthy();
    // New fields present but benign
    expect(body.data.is_compensating).toBe(false);
  });

  test('existing control LIST still works and includes new field', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/v1/controls`, {
      headers: authHeaders(org.token),
    });

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(Array.isArray(body.data)).toBeTruthy();
    expect(body.meta.total).toBeGreaterThan(0);

    // Every control in the list must have is_compensating field
    for (const ctrl of body.data) {
      expect(typeof ctrl.is_compensating).toBe('boolean');
    }
  });

  test('existing control UPDATE still works without new fields', async ({ request }) => {
    const res = await request.put(`${BASE_URL}/api/v1/controls/${controlId}`, {
      headers: authHeaders(org.token),
      data: {
        description: 'Updated description — regression check for schema change.',
      },
    });

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.description).toBe('Updated description — regression check for schema change.');
    // is_compensating must remain untouched
    expect(body.data.is_compensating).toBe(false);
  });

  test('existing control status transitions still work', async ({ request }) => {
    // Activate the control
    const res = await request.put(`${BASE_URL}/api/v1/controls/${controlId}/status`, {
      headers: authHeaders(org.token),
      data: { status: 'active' },
    });

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.status).toBe('active');
    // New field unaffected by status transition
    expect(body.data.is_compensating).toBe(false);
  });
});

// =============================================================================
// REQUIREMENT SCOPES: customized_approach flag regression
//
// BACKEND GAPS (Logan — LEX-86 / Dana — schema):
//   1. `customized_approach` column added by migration 053, but NOT yet included
//      in the ListScopes or SetScope SELECT / response maps. Logan must add it.
//   2. `customized_approach_description` column does NOT exist in the DB schema yet.
//      Dana must add it as a migration before these tests can pass.
// =============================================================================

test.describe('Requirement Scopes regression — customized_approach field addition', () => {
  // We need a framework requirement ID to scope. Fetch one from the org's framework.
  let requirementId: string;
  let frameworkId: string;

  test.beforeAll(async ({ request }) => {
    // Enroll org in PCI DSS framework
    const frameworksRes = await request.get(`${BASE_URL}/api/v1/frameworks`, {
      headers: authHeaders(org.token),
    });
    const body = await frameworksRes.json();
    const pci = body.data.find((f: any) =>
      f.name?.toLowerCase().includes('pci') || f.short_name?.toLowerCase().includes('pci'),
    );
    if (pci) {
      frameworkId = pci.id;

      // Enroll org in PCI framework
      await request.post(`${BASE_URL}/api/v1/org-frameworks`, {
        headers: authHeaders(org.token),
        data: { framework_id: frameworkId },
      });

      // Grab first requirement
      const reqRes = await request.get(
        `${BASE_URL}/api/v1/frameworks/${frameworkId}/requirements?page=1&per_page=1`,
        { headers: authHeaders(org.token) },
      );
      const reqBody = await reqRes.json();
      if (reqBody.data?.length > 0) {
        requirementId = reqBody.data[0].id;
      }
    }
  });

  test('existing scope set/get still works (customized_approach omitted = false)', async ({ request }) => {
    if (!requirementId) test.skip();

    const res = await request.put(
      `${BASE_URL}/api/v1/requirements/${requirementId}/scope`,
      {
        headers: authHeaders(org.token),
        data: {
          in_scope: true,
          justification: 'Requirement applies to our CDE environment.',
        },
      },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.in_scope).toBe(true);
    // New field must default to false and not break response schema
    expect(body.data.customized_approach).toBe(false);
  });

  test('can set customized_approach=true on a requirement scope', async ({ request }) => {
    if (!requirementId) test.skip();

    const res = await request.put(
      `${BASE_URL}/api/v1/requirements/${requirementId}/scope`,
      {
        headers: authHeaders(org.token),
        data: {
          in_scope: true,
          justification: 'Using PCI DSS v4 customized approach.',
          customized_approach: true,
          customized_approach_description:
            'We implement a custom control set that achieves the stated objective through alternative means as documented in our targeted risk analysis.',
        },
      },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.customized_approach).toBe(true);
    expect(body.data.customized_approach_description).toContain('targeted risk analysis');
  });

  test('existing scope listing still works and includes new field', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/requirement-scopes`,
      { headers: authHeaders(org.token) },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(Array.isArray(body.data)).toBeTruthy();

    // Every scope row must include customized_approach (even if false)
    for (const scope of body.data) {
      expect(typeof scope.customized_approach).toBe('boolean');
    }
  });
});

// =============================================================================
// AUDITS: new PCI DSS audit types
// =============================================================================

test.describe('Audits regression — new PCI DSS audit types', () => {
  test('existing audit creation still works (standard type)', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/audits`, {
      headers: authHeaders(org.token),
      data: {
        title: 'Annual SOC 2 Type II Audit',
        audit_type: 'external',
        framework_id: null,
        start_date: '2026-01-01',
        target_end_date: '2026-03-31',
      },
    });

    // Should succeed with original audit_type values
    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.audit_type).toBe('external');
  });

  test('should create audit with pci_dss_saq_a type', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/audits`, {
      headers: authHeaders(org.token),
      data: {
        title: 'PCI DSS SAQ-A Self Assessment',
        audit_type: 'pci_dss_saq_a',
        start_date: '2026-03-01',
        target_end_date: '2026-03-31',
      },
    });

    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.audit_type).toBe('pci_dss_saq_a');
  });

  test('should create audit with pci_dss_saq_d type', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/audits`, {
      headers: authHeaders(org.token),
      data: {
        title: 'PCI DSS SAQ-D Merchant Self Assessment',
        audit_type: 'pci_dss_saq_d',
        start_date: '2026-03-01',
        target_end_date: '2026-04-30',
      },
    });

    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.audit_type).toBe('pci_dss_saq_d');
  });

  test('should create audit with pci_dss_aoc type', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/audits`, {
      headers: authHeaders(org.token),
      data: {
        title: 'PCI DSS Attestation of Compliance',
        audit_type: 'pci_dss_aoc',
        start_date: '2026-03-01',
        target_end_date: '2026-06-30',
      },
    });

    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.audit_type).toBe('pci_dss_aoc');
  });

  test('existing audit LIST includes new types without breaking pagination', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/v1/audits?page=1&per_page=20`, {
      headers: authHeaders(org.token),
    });

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(Array.isArray(body.data)).toBeTruthy();
    expect(typeof body.meta.total).toBe('number');

    // Should have at least our 4 created audits
    expect(body.meta.total).toBeGreaterThanOrEqual(4);
  });

  test('can filter audits by pci_dss_aoc type', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/audits?audit_type=pci_dss_aoc`,
      { headers: authHeaders(org.token) },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    for (const audit of body.data) {
      expect(audit.audit_type).toBe('pci_dss_aoc');
    }
  });

  test('invalid audit type still returns 400', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/audits`, {
      headers: authHeaders(org.token),
      data: {
        title: 'Invalid Type Audit',
        audit_type: 'not_a_real_type',
        start_date: '2026-03-01',
        target_end_date: '2026-04-30',
      },
    });
    expect(res.status()).toBe(400);
  });
});

// =============================================================================
// SERVICE PROVIDERS (Vendor Management): routes
//
// The vendor management quick-win was implemented as /api/v1/service-providers
// (not /api/v1/vendor-management — that path is not registered).
// =============================================================================

test.describe('Service Providers (Vendor Management) — routes reachable', () => {
  test('service providers list endpoint is reachable and returns 200', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/v1/service-providers`, {
      headers: authHeaders(org.token),
    });

    // Route must be registered and return an authenticated response
    expect(res.status()).not.toBe(404);
    // Accept 200 (empty list or populated)
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(Array.isArray(body.data)).toBeTruthy();
  });
});
