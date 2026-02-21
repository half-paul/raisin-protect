import { test, expect } from '@playwright/test';

let authToken: string;
let userId: string;
let orgId: string;
let riskId: string;

test.beforeAll(async ({ request }) => {
  // Register a test user
  const response = await request.post('http://localhost:8090/api/v1/auth/register', {
    data: {
      email: `qatest-risk-${Date.now()}@example.com`,
      password: 'Test1234!',
      first_name: 'QA',
      last_name: 'Tester',
      org_name: 'QA Org Risk Tests',
      role: 'compliance_manager'
    }
  });
  
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  authToken = body.data.access_token;
  userId = body.data.user.id;
  orgId = body.data.organization.id;
});

test.describe('Risk CRUD Operations', () => {
  
  test('should create a new risk with initial assessment', async ({ request }) => {
    const response = await request.post('http://localhost:8090/api/v1/risks', {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        title: 'E2E Test Risk: Ransomware Attack',
        description: 'Risk of ransomware compromising critical systems',
        category: 'cybersecurity',
        owner_id: userId,
        initial_likelihood: 'high',
        initial_impact: 'critical'
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.risk.title).toBe('E2E Test Risk: Ransomware Attack');
    expect(body.data.risk.category).toBe('cybersecurity');
    expect(body.data.risk.status).toBe('identified');
    
    // Should have created initial assessment
    expect(body.data.initial_assessment).toBeDefined();
    expect(body.data.initial_assessment.likelihood).toBe('high');
    expect(body.data.initial_assessment.impact).toBe('critical');
    expect(body.data.initial_assessment.score).toBe(20); // high(4) × critical(5) = 20
    
    riskId = body.data.risk.id;
  });

  test('should list risks with filtering', async ({ request }) => {
    const response = await request.get('http://localhost:8090/api/v1/risks?category=cybersecurity', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.risks.length).toBeGreaterThan(0);
    expect(body.data.risks[0].category).toBe('cybersecurity');
  });

  test('should get risk by ID', async ({ request }) => {
    const response = await request.get(`http://localhost:8090/api/v1/risks/${riskId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.id).toBe(riskId);
    expect(body.data.title).toBe('E2E Test Risk: Ransomware Attack');
  });

  test('should update risk', async ({ request }) => {
    const response = await request.put(`http://localhost:8090/api/v1/risks/${riskId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        title: 'Updated: Ransomware Attack Risk',
        description: 'Updated description with more details'
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.title).toBe('Updated: Ransomware Attack Risk');
  });

  test('should archive risk', async ({ request }) => {
    const response = await request.post(`http://localhost:8090/api/v1/risks/${riskId}/archive`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        reason: 'No longer applicable'
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.status).toBe('archived');
  });

  test('should enforce multi-tenancy isolation', async ({ request }) => {
    // Register another user in different org
    const otherUserResponse = await request.post('http://localhost:8090/api/v1/auth/register', {
      data: {
        email: `other-${Date.now()}@example.com`,
        password: 'Test1234!',
        first_name: 'Other',
        last_name: 'User',
        org_name: 'Other Org',
        role: 'compliance_manager'
      }
    });

    const otherUserBody = await otherUserResponse.json();
    const otherToken = otherUserBody.data.access_token;

    // Try to access original risk with different org's token
    const response = await request.get(`http://localhost:8090/api/v1/risks/${riskId}`, {
      headers: {
        'Authorization': `Bearer ${otherToken}`
      }
    });

    expect(response.status()).toBe(404); // Should not find risk from different org
  });
});
