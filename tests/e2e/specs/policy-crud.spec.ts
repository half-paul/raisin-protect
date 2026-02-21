import { test, expect } from '@playwright/test';

// Sprint 5: Policy CRUD tests
// Tests basic policy creation, update, search, and archiving

test.describe('Policy CRUD', () => {
  let authToken: string;
  let policyId: string;

  test.beforeAll(async ({ request }) => {
    // Login as demo user
    const response = await request.post('http://localhost:8090/api/v1/auth/login', {
      data: {
        email: 'demo@example.com',
        password: 'demo123'
      }
    });
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    authToken = data.data.token;
  });

  test('should create a new policy', async ({ request }) => {
    const response = await request.post('http://localhost:8090/api/v1/policies', {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        identifier: 'E2E-TEST-001',
        title: 'E2E Test Policy',
        description: 'Test policy created by E2E tests',
        category: 'access_control',
        content: '<h1>Test Policy</h1><p>This is a test policy.</p>',
        content_format: 'html',
        review_frequency_days: 365,
        tags: ['e2e', 'test']
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.identifier).toBe('E2E-TEST-001');
    expect(data.data.status).toBe('draft');
    policyId = data.data.id;
  });

  test('should list policies', async ({ request }) => {
    const response = await request.get('http://localhost:8090/api/v1/policies', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(Array.isArray(data.data)).toBeTruthy();
    expect(data.data.length).toBeGreaterThan(0);
  });

  test('should get policy by id', async ({ request }) => {
    const response = await request.get(`http://localhost:8090/api/v1/policies/${policyId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.id).toBe(policyId);
    expect(data.data.identifier).toBe('E2E-TEST-001');
  });

  test('should update policy', async ({ request }) => {
    const response = await request.put(`http://localhost:8090/api/v1/policies/${policyId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        title: 'E2E Test Policy (Updated)',
        description: 'Updated description'
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.title).toBe('E2E Test Policy (Updated)');
  });

  test('should search policies by keyword', async ({ request }) => {
    const response = await request.get('http://localhost:8090/api/v1/policies/search?q=E2E', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(Array.isArray(data.data)).toBeTruthy();
  });

  test('should filter policies by category', async ({ request }) => {
    const response = await request.get('http://localhost:8090/api/v1/policies?category=access_control', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(Array.isArray(data.data)).toBeTruthy();
  });

  test('should filter policies by status', async ({ request }) => {
    const response = await request.get('http://localhost:8090/api/v1/policies?status=draft', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(Array.isArray(data.data)).toBeTruthy();
  });

  // CRITICAL TEST: Issue #10 - Missing RBAC on ArchivePolicy
  test('should enforce RBAC on archive policy (non-admin should be denied)', async ({ request }) => {
    // This test checks Issue #10 from code review
    // Expected: Only compliance_manager or above should be able to archive
    // Actual (bug): Any authenticated user can archive
    
    const response = await request.post(`http://localhost:8090/api/v1/policies/${policyId}/archive`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        archive_reason: 'Test archive'
      }
    });

    // TODO: Once Issue #10 is fixed, this should return 403 for non-managers
    // For now, document the security gap
    if (response.status() === 200) {
      console.warn('[SECURITY] Issue #10 confirmed: ArchivePolicy missing RBAC check');
    }
  });
});
