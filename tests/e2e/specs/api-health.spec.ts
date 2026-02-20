import { test, expect } from '@playwright/test';

const API_BASE_URL = 'http://localhost:8090';

test.describe('API Health Checks', () => {
  test('GET /health should return 200 OK', async ({ request }) => {
    const response = await request.get(`${API_BASE_URL}/health`);
    expect(response.ok()).toBeTruthy();
    
    const data = await response.json();
    expect(data.status).toBe('ok');
    expect(data.version).toBeDefined();
  });

  test('Auth endpoints should be accessible', async ({ request }) => {
    const response = await request.post(`${API_BASE_URL}/api/v1/auth/login`, {
      data: {
        email: 'nonexistent@example.com',
        password: 'wrong',
      },
    });
    
    // Should return 401 for invalid credentials (not 404 or 500)
    expect(response.status()).toBe(401);
  });
});

test.describe('Evidence API', () => {
  let token: string;
  let orgId: string;

  test.beforeAll(async ({ request }) => {
    // Register a test user
    const response = await request.post(`${API_BASE_URL}/api/v1/auth/register`, {
      data: {
        email: `e2e-${Date.now()}@example.com`,
        password: 'TestP@ss123',
        first_name: 'E2E',
        last_name: 'Test',
        org_name: 'E2E Test Org',
      },
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    token = data.data.access_token;
    orgId = data.data.organization.id;
  });

  test('GET /api/v1/evidence should return empty list for new org', async ({ request }) => {
    const response = await request.get(`${API_BASE_URL}/api/v1/evidence`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data).toEqual([]);
    expect(data.meta.total).toBe(0);
  });

  test('GET /api/v1/evidence/staleness should return empty alerts', async ({ request }) => {
    const response = await request.get(`${API_BASE_URL}/api/v1/evidence/staleness?days=30`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.summary.total_alerts).toBe(0);
    expect(data.data.summary.expired).toBe(0);
    expect(data.data.summary.expiring_soon).toBe(0);
  });

  test('GET /api/v1/evidence without auth should return 401', async ({ request }) => {
    const response = await request.get(`${API_BASE_URL}/api/v1/evidence`);
    expect(response.status()).toBe(401);
  });
});
