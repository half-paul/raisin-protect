/**
 * E2E Integration Tests: CDE Scoping Module
 *
 * Coverage:
 *   - CDE Asset CRUD (POST/GET/PUT/DELETE /api/v1/cde/assets)
 *   - Network Segments CRUD (/api/v1/cde/segments)
 *   - Data Flows CRUD (/api/v1/cde/data-flows)
 *   - Segmentation Tests CRUD (/api/v1/cde/segmentation-tests)
 *   - Scope Summary (/api/v1/cde/scope-summary)
 *   - Org-scoping enforcement via JWT (cross-tenant isolation)
 *   - Input validation and error handling
 *   - Edge cases: empty states, max field lengths, invalid scope_status values
 *
 * Depends on: LEX-84 (CDE API) — tests are spec-first, will pass once
 * handlers are implemented and migrated (LEX-83 schema + LEX-84 handlers).
 *
 * PCI DSS coverage: Requirement 1 (Network Segmentation), Req 11.4 (Segmentation Testing)
 *
 * URL note: CDE routes live at /api/v1/cde/* (NOT /api/v1/orgs/:orgId/cde/*).
 * Org isolation is enforced via the JWT — the org_id in the token scopes all queries.
 * Cross-tenant isolation: accessing another org's resources by ID returns 404 (not 403).
 */

import { test, expect } from '@playwright/test';

const BASE_URL = 'http://localhost:8090';

// ---- helpers ----------------------------------------------------------------

async function registerOrg(request: any, label: string) {
  const res = await request.post(`${BASE_URL}/api/v1/auth/register`, {
    data: {
      email: `cde-test-${label}-${Date.now()}@example.com`,
      password: 'Test1234!',
      first_name: 'QA',
      last_name: 'Tester',
      org_name: `CDE QA Org ${label}`,
      role: 'compliance_manager',
    },
  });
  expect(res.ok(), `register ${label} failed: ${await res.text()}`).toBeTruthy();
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
let orgB: { token: string; orgId: string }; // used for cross-tenant isolation tests

let assetId: string;
let destAssetId: string;  // second asset for data flow src→dest pair
let segmentId: string;
let dataFlowId: string;
let segTestId: string;

test.beforeAll(async ({ request }) => {
  [orgA, orgB] = await Promise.all([
    registerOrg(request, 'orgA'),
    registerOrg(request, 'orgB'),
  ]);
});

// =============================================================================
// CDE ASSETS
// =============================================================================

test.describe('CDE Assets — CRUD', () => {
  test('should create a CDE asset (happy path)', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/cde/assets`, {
      headers: authHeaders(orgA.token),
      data: {
        name: 'Payment Processing Server',
        type: 'server',
        scope_status: 'in_scope',
        environment: 'production',
        description: 'Primary server handling card authorization',
        data_classification: 'pan',
        ip_address: '10.0.1.50',
        hostname: 'pay-svc-01.internal',
      },
    });

    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.name).toBe('Payment Processing Server');
    expect(body.data.type).toBe('server');
    expect(body.data.scope_status).toBe('in_scope');
    expect(body.data.org_id).toBe(orgA.orgId);

    assetId = body.data.id;
    expect(assetId).toBeTruthy();
  });

  test('should reject asset creation with missing required fields', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/cde/assets`, {
      headers: authHeaders(orgA.token),
      data: {
        // missing name, type, environment, and data_classification (all required)
        scope_status: 'in_scope',
      },
    });
    expect(res.status()).toBe(400);
  });

  test('should reject invalid scope_status value', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/cde/assets`, {
      headers: authHeaders(orgA.token),
      data: {
        name: 'Test Asset',
        type: 'server',
        environment: 'production',
        data_classification: 'none',
        scope_status: 'maybe_in_scope', // invalid
      },
    });
    expect(res.status()).toBe(400);
  });

  test('should list CDE assets for the org (non-empty)', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/v1/cde/assets`, {
      headers: authHeaders(orgA.token),
    });

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(Array.isArray(body.data)).toBeTruthy();
    expect(body.data.length).toBeGreaterThanOrEqual(1);
    expect(body.meta.total).toBeGreaterThanOrEqual(1);
  });

  test('should return empty list when no assets exist', async ({ request }) => {
    // orgB has no assets yet — JWT-scoped list returns orgB's empty set
    const res = await request.get(`${BASE_URL}/api/v1/cde/assets`, {
      headers: authHeaders(orgB.token),
    });

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data).toEqual([]);
    expect(body.meta.total).toBe(0);
  });

  test('should filter assets by scope_status', async ({ request }) => {
    // Create an out-of-scope asset first
    await request.post(`${BASE_URL}/api/v1/cde/assets`, {
      headers: authHeaders(orgA.token),
      data: {
        name: 'Dev Server',
        type: 'server',
        environment: 'development',
        data_classification: 'none',
        scope_status: 'out_of_scope',
      },
    });

    const res = await request.get(
      `${BASE_URL}/api/v1/cde/assets?scope_status=in_scope`,
      { headers: authHeaders(orgA.token) },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    // Every returned asset must be in_scope
    for (const asset of body.data) {
      expect(asset.scope_status).toBe('in_scope');
    }
  });

  test('should get a single CDE asset by ID', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/assets/${assetId}`,
      { headers: authHeaders(orgA.token) },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.id).toBe(assetId);
    expect(body.data.name).toBe('Payment Processing Server');
  });

  test('should return 404 for non-existent asset', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/assets/00000000-0000-0000-0000-000000000000`,
      { headers: authHeaders(orgA.token) },
    );
    expect(res.status()).toBe(404);
  });

  test('should update a CDE asset', async ({ request }) => {
    const res = await request.put(
      `${BASE_URL}/api/v1/cde/assets/${assetId}`,
      {
        headers: authHeaders(orgA.token),
        data: {
          description: 'Updated: Primary server handling card authorization v2',
          scope_status: 'in_scope',
        },
      },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.description).toBe('Updated: Primary server handling card authorization v2');
  });

  test('should reject name exceeding max length', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/cde/assets`, {
      headers: authHeaders(orgA.token),
      data: {
        name: 'A'.repeat(256), // exceeds typical 255-char limit
        type: 'server',
        environment: 'production',
        data_classification: 'pan',
        scope_status: 'in_scope',
      },
    });
    expect(res.status()).toBe(400);
  });

  test('should delete a CDE asset', async ({ request }) => {
    // Create a disposable asset
    const createRes = await request.post(
      `${BASE_URL}/api/v1/cde/assets`,
      {
        headers: authHeaders(orgA.token),
        data: {
          name: 'Disposable Asset',
          type: 'workstation',
          environment: 'staging',
          data_classification: 'none',
          scope_status: 'in_scope',
        },
      },
    );
    const disposableId = (await createRes.json()).data.id;

    const deleteRes = await request.delete(
      `${BASE_URL}/api/v1/cde/assets/${disposableId}`,
      { headers: authHeaders(orgA.token) },
    );
    expect(deleteRes.status()).toBe(204);

    // Verify gone
    const getRes = await request.get(
      `${BASE_URL}/api/v1/cde/assets/${disposableId}`,
      { headers: authHeaders(orgA.token) },
    );
    expect(getRes.status()).toBe(404);
  });
});

// =============================================================================
// ORG-SCOPING SECURITY TESTS (cross-tenant isolation)
//
// CDE routes use JWT-scoped org_id — there is no :orgId in the URL path.
// All queries are filtered by the org_id extracted from the JWT, so:
//   - List endpoints return the caller's own data only (200, not 403)
//   - Accessing another org's resource by ID returns 404 (not 403)
// =============================================================================

test.describe('CDE Assets — Org-scoping enforcement (no cross-tenant leakage)', () => {
  test('orgB list returns only orgB assets (not orgA assets)', async ({ request }) => {
    // orgB calls the list — JWT scopes it to orgB's org, returns empty list
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/assets`,
      { headers: authHeaders(orgB.token) },
    );
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    // orgA's assetId must not appear in orgB's list
    const ids = body.data.map((a: any) => a.id);
    expect(ids).not.toContain(assetId);
  });

  test('orgB cannot read a specific orgA asset (returns 404)', async ({ request }) => {
    // assetId belongs to orgA; orgB's JWT scopes queries to orgB — returns 404 not found
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/assets/${assetId}`,
      { headers: authHeaders(orgB.token) },
    );
    expect(res.status()).toBe(404);
  });

  test('orgB creating an asset creates it in orgB org, not orgA', async ({ request }) => {
    const res = await request.post(
      `${BASE_URL}/api/v1/cde/assets`,
      {
        headers: authHeaders(orgB.token),
        data: {
          name: 'OrgB Asset',
          type: 'server',
          environment: 'production',
          data_classification: 'none',
          scope_status: 'in_scope',
        },
      },
    );
    expect(res.status()).toBe(201);
    const body = await res.json();
    // Must be scoped to orgB, not orgA
    expect(body.data.org_id).toBe(orgB.orgId);
    expect(body.data.org_id).not.toBe(orgA.orgId);
  });

  test('orgB cannot update orgA assets (returns 404)', async ({ request }) => {
    const res = await request.put(
      `${BASE_URL}/api/v1/cde/assets/${assetId}`,
      {
        headers: authHeaders(orgB.token),
        data: { name: 'Hijacked Asset' },
      },
    );
    expect(res.status()).toBe(404);
  });

  test('orgB cannot delete orgA assets (returns 404)', async ({ request }) => {
    const res = await request.delete(
      `${BASE_URL}/api/v1/cde/assets/${assetId}`,
      { headers: authHeaders(orgB.token) },
    );
    expect(res.status()).toBe(404);
  });

  test('unauthenticated request returns 401', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/assets`,
    );
    expect(res.status()).toBe(401);
  });
});

// =============================================================================
// NETWORK SEGMENTS
// =============================================================================

test.describe('CDE Network Segments — CRUD', () => {
  test('should create a network segment', async ({ request }) => {
    const res = await request.post(
      `${BASE_URL}/api/v1/cde/segments`,
      {
        headers: authHeaders(orgA.token),
        data: {
          name: 'CDE VLAN',
          cidr: '10.0.1.0/24',
          segment_type: 'cde',
          description: 'Isolated VLAN containing PAN data systems',
        },
      },
    );

    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.name).toBe('CDE VLAN');
    expect(body.data.segment_type).toBe('cde');
    segmentId = body.data.id;
  });

  test('should list network segments', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/segments`,
      { headers: authHeaders(orgA.token) },
    );
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.length).toBeGreaterThanOrEqual(1);
  });

  test('should update a network segment', async ({ request }) => {
    const res = await request.put(
      `${BASE_URL}/api/v1/cde/segments/${segmentId}`,
      {
        headers: authHeaders(orgA.token),
        data: { description: 'Updated VLAN description' },
      },
    );
    expect(res.ok()).toBeTruthy();
  });

  test('orgB list returns only orgB segments (not orgA segments)', async ({ request }) => {
    // JWT-scoped: orgB's query returns orgB's (empty) segments
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/segments`,
      { headers: authHeaders(orgB.token) },
    );
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    const ids = body.data.map((s: any) => s.id);
    expect(ids).not.toContain(segmentId);
  });
});

// =============================================================================
// DATA FLOWS
//
// Data flows link two CDE assets via source_asset_id / dest_asset_id (UUIDs).
// Fields: source_asset_id (required), dest_asset_id (required), protocol,
//         port, data_type, encryption_method.
// =============================================================================

test.describe('CDE Data Flows — CRUD', () => {
  // Create a destination asset to pair with assetId (created in CDE Assets describe above)
  test.beforeAll(async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/v1/cde/assets`, {
      headers: authHeaders(orgA.token),
      data: {
        name: 'Payment Gateway Server',
        type: 'server',
        environment: 'production',
        data_classification: 'pan',
        scope_status: 'in_scope',
        description: 'Downstream payment gateway endpoint',
      },
    });
    expect(res.ok(), `dest asset create failed: ${await res.text()}`).toBeTruthy();
    destAssetId = (await res.json()).data.id;
  });

  test('should create a data flow', async ({ request }) => {
    const res = await request.post(
      `${BASE_URL}/api/v1/cde/data-flows`,
      {
        headers: authHeaders(orgA.token),
        data: {
          source_asset_id: assetId,
          dest_asset_id: destAssetId,
          data_type: 'pan',
          protocol: 'HTTPS/TLS 1.3',
        },
      },
    );

    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.source_asset_id).toBe(assetId);
    expect(body.data.dest_asset_id).toBe(destAssetId);
    expect(body.data.data_type).toBe('pan');
    dataFlowId = body.data.id;
  });

  test('should list data flows', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/data-flows`,
      { headers: authHeaders(orgA.token) },
    );
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.length).toBeGreaterThanOrEqual(1);
  });

  test('orgB list returns only orgB data flows (not orgA flows)', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/data-flows`,
      { headers: authHeaders(orgB.token) },
    );
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    const ids = body.data.map((f: any) => f.id);
    expect(ids).not.toContain(dataFlowId);
  });
});

// =============================================================================
// SEGMENTATION TESTS
// =============================================================================

test.describe('CDE Segmentation Tests — CRUD', () => {
  test('should create a segmentation test record', async ({ request }) => {
    const res = await request.post(
      `${BASE_URL}/api/v1/cde/segmentation-tests`,
      {
        headers: authHeaders(orgA.token),
        data: {
          test_date: '2026-03-15',
          performed_by: 'External Pentest Vendor',
          methodology: 'Network port scanning + firewall rule review',
          result: 'pass',
          findings_summary: 'No unauthorized cross-segment paths identified',
          next_test_due: '2026-09-15',
        },
      },
    );

    expect(res.status()).toBe(201);
    const body = await res.json();
    expect(body.data.result).toBe('pass');
    segTestId = body.data.id;
  });

  test('should reject invalid result value', async ({ request }) => {
    const res = await request.post(
      `${BASE_URL}/api/v1/cde/segmentation-tests`,
      {
        headers: authHeaders(orgA.token),
        data: {
          test_date: '2026-03-15',
          performed_by: 'Vendor',
          result: 'maybe', // invalid
        },
      },
    );
    expect(res.status()).toBe(400);
  });

  test('should list segmentation tests', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/segmentation-tests`,
      { headers: authHeaders(orgA.token) },
    );
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.data.length).toBeGreaterThanOrEqual(1);
  });
});

// =============================================================================
// SCOPE SUMMARY
// =============================================================================

test.describe('CDE Scope Summary', () => {
  test('should return aggregated scope status', async ({ request }) => {
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/scope-summary`,
      { headers: authHeaders(orgA.token) },
    );

    expect(res.ok()).toBeTruthy();
    const body = await res.json();

    // Verify expected summary fields
    expect(typeof body.data.total_assets).toBe('number');
    expect(typeof body.data.in_scope_assets).toBe('number');
    expect(typeof body.data.out_of_scope_assets).toBe('number');
    expect(typeof body.data.total_segments).toBe('number');
    expect(typeof body.data.total_data_flows).toBe('number');
    expect(body.data.last_segmentation_test).toBeDefined();
  });

  test('orgB scope summary returns only orgB data', async ({ request }) => {
    // orgB's summary should reflect their own (empty) state
    const res = await request.get(
      `${BASE_URL}/api/v1/cde/scope-summary`,
      { headers: authHeaders(orgB.token) },
    );
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    // orgB has no CDE assets yet — totals must be 0
    expect(body.data.total_assets).toBe(0);
    expect(body.data.total_data_flows).toBe(0);
  });
});
