import { test, expect } from '@playwright/test';

// Sprint 5: Policy-to-Control Mapping tests
// Tests linking policies to controls and gap detection

test.describe('Policy-to-Control Mapping', () => {
  let authToken: string;
  let policyId: string;
  let controlId: string;

  test.beforeAll(async ({ request }) => {
    // Login
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
        identifier: 'MAPPING-TEST-001',
        title: 'Control Mapping Test Policy',
        description: 'Policy for testing control mappings',
        category: 'access_control',
        content: '<h1>Access Control</h1><p>Controls user access.</p>',
        content_format: 'html',
        review_frequency_days: 365
      }
    });
    
    const policyData = await policyResponse.json();
    policyId = policyData.data.id;

    // Get an existing control to link
    const controlsResponse = await request.get('http://localhost:8090/api/v1/controls?limit=1', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const controlsData = await controlsResponse.json();
    if (controlsData.data && controlsData.data.length > 0) {
      controlId = controlsData.data[0].id;
    }
  });

  test('should link policy to control', async ({ request }) => {
    test.skip(!controlId, 'No control available to link');

    const response = await request.post(`http://localhost:8090/api/v1/policies/${policyId}/controls`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        control_id: controlId,
        coverage_level: 'full',
        notes: 'This policy fully covers the control requirements'
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.control_id).toBe(controlId);
    expect(data.data.coverage_level).toBe('full');
  });

  test('should list policy controls', async ({ request }) => {
    const response = await request.get(`http://localhost:8090/api/v1/policies/${policyId}/controls`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(Array.isArray(data.data)).toBeTruthy();
  });

  test('should bulk link multiple controls', async ({ request }) => {
    // Get multiple controls
    const controlsResponse = await request.get('http://localhost:8090/api/v1/controls?limit=3', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const controlsData = await controlsResponse.json();
    
    if (!controlsData.data || controlsData.data.length < 2) {
      console.log('[INFO] Not enough controls for bulk link test');
      return;
    }

    const controlIds = controlsData.data.slice(0, 2).map((c: any) => c.id);

    const response = await request.post(`http://localhost:8090/api/v1/policies/${policyId}/controls/bulk`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        control_ids: controlIds,
        coverage_level: 'partial',
        notes: 'Bulk linked for testing'
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.created_count).toBe(controlIds.length);
  });

  test('should unlink policy from control', async ({ request }) => {
    test.skip(!controlId, 'No control available to unlink');

    const response = await request.delete(
      `http://localhost:8090/api/v1/policies/${policyId}/controls/${controlId}`,
      {
        headers: {
          'Authorization': `Bearer ${authToken}`
        }
      }
    );

    expect(response.ok()).toBeTruthy();
  });

  test('should detect policy gaps by control', async ({ request }) => {
    const response = await request.get('http://localhost:8090/api/v1/policy-gap', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data).toBeDefined();
    expect(data.data.total_controls).toBeGreaterThan(0);
    expect(data.data.controls_without_policy).toBeDefined();
  });

  test('should detect policy gaps by framework', async ({ request }) => {
    // Get a framework first
    const frameworksResponse = await request.get('http://localhost:8090/api/v1/frameworks?limit=1', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const frameworksData = await frameworksResponse.json();
    
    if (!frameworksData.data || frameworksData.data.length === 0) {
      console.log('[INFO] No frameworks available for gap analysis');
      return;
    }

    const frameworkId = frameworksData.data[0].id;

    const response = await request.get(`http://localhost:8090/api/v1/policy-gap/by-framework?framework_id=${frameworkId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data).toBeDefined();
    expect(Array.isArray(data.data.gaps)).toBeTruthy();
  });

  test('should include coverage level in mapping', async ({ request }) => {
    test.skip(!controlId, 'No control available');

    // Re-link with different coverage levels
    const coverageLevels = ['full', 'partial', 'minimal'];
    
    for (const level of coverageLevels) {
      // First unlink if exists
      await request.delete(`http://localhost:8090/api/v1/policies/${policyId}/controls/${controlId}`, {
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      // Link with specific coverage level
      const response = await request.post(`http://localhost:8090/api/v1/policies/${policyId}/controls`, {
        headers: {
          'Authorization': `Bearer ${authToken}`,
          'Content-Type': 'application/json'
        },
        data: {
          control_id: controlId,
          coverage_level: level,
          notes: `Testing ${level} coverage`
        }
      });

      expect(response.ok()).toBeTruthy();
      const data = await response.json();
      expect(data.data.coverage_level).toBe(level);
    }
  });
});
