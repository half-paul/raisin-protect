import { test, expect } from '@playwright/test';

// Sprint 5: Policy Version Management tests
// Tests version creation, history, and comparison

test.describe('Policy Versioning', () => {
  let authToken: string;
  let policyId: string;
  let version1Id: string;
  let version2Id: string;

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

    // Create a test policy
    const policyResponse = await request.post('http://localhost:8090/api/v1/policies', {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        identifier: 'VERSION-TEST-001',
        title: 'Version Test Policy',
        description: 'Policy for testing versioning',
        category: 'data_protection',
        content: '<h1>Version 1</h1><p>Initial content.</p>',
        content_format: 'html',
        review_frequency_days: 90,
        tags: ['versioning', 'test']
      }
    });
    
    expect(policyResponse.ok()).toBeTruthy();
    const policyData = await policyResponse.json();
    policyId = policyData.data.id;
    version1Id = policyData.data.current_version.id;
  });

  test('should create a new policy version', async ({ request }) => {
    const response = await request.post(`http://localhost:8090/api/v1/policies/${policyId}/versions`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        content: '<h1>Version 2</h1><p>Updated content with new requirements.</p>',
        content_format: 'html',
        change_summary: 'Added new security requirements per NIST 800-53',
        change_type: 'minor'
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.version_number).toBe(2);
    expect(data.data.change_summary).toContain('NIST 800-53');
    version2Id = data.data.id;
  });

  test('should list all policy versions', async ({ request }) => {
    const response = await request.get(`http://localhost:8090/api/v1/policies/${policyId}/versions`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(Array.isArray(data.data)).toBeTruthy();
    expect(data.data.length).toBeGreaterThanOrEqual(2);
  });

  test('should get specific version by number', async ({ request }) => {
    const response = await request.get(`http://localhost:8090/api/v1/policies/${policyId}/versions/1`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.version_number).toBe(1);
    expect(data.data.content).toContain('Version 1');
  });

  test('should compare two versions', async ({ request }) => {
    const response = await request.get(
      `http://localhost:8090/api/v1/policies/${policyId}/versions/compare?from=1&to=2`,
      {
        headers: {
          'Authorization': `Bearer ${authToken}`
        }
      }
    );

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.from.version_number).toBe(1);
    expect(data.data.to.version_number).toBe(2);
    expect(data.data.diff).toBeDefined();
  });

  test('should track word count changes between versions', async ({ request }) => {
    const v1Response = await request.get(`http://localhost:8090/api/v1/policies/${policyId}/versions/1`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const v1Data = await v1Response.json();
    const v1WordCount = v1Data.data.word_count;

    const v2Response = await request.get(`http://localhost:8090/api/v1/policies/${policyId}/versions/2`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const v2Data = await v2Response.json();
    const v2WordCount = v2Data.data.word_count;

    expect(v2WordCount).toBeGreaterThan(v1WordCount);
  });

  test('should update policy to new version (current_version_id should change)', async ({ request }) => {
    const beforeResponse = await request.get(`http://localhost:8090/api/v1/policies/${policyId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const beforeData = await beforeResponse.json();
    const beforeVersionId = beforeData.data.current_version.id;

    // Create version 3
    const createResponse = await request.post(`http://localhost:8090/api/v1/policies/${policyId}/versions`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        content: '<h1>Version 3</h1><p>Latest content.</p>',
        content_format: 'html',
        change_summary: 'Version 3',
        change_type: 'patch'
      }
    });
    expect(createResponse.ok()).toBeTruthy();

    const afterResponse = await request.get(`http://localhost:8090/api/v1/policies/${policyId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const afterData = await afterResponse.json();
    const afterVersionId = afterData.data.current_version.id;

    expect(afterVersionId).not.toBe(beforeVersionId);
    expect(afterData.data.current_version.version_number).toBe(3);
  });
});
