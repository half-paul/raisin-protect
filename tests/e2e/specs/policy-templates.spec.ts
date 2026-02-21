import { test, expect } from '@playwright/test';

// Sprint 5: Policy Template tests
// Tests template library and cloning functionality

test.describe('Policy Templates', () => {
  let authToken: string;
  let templateId: string;

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
  });

  test('should list policy templates', async ({ request }) => {
    const response = await request.get('http://localhost:8090/api/v1/policy-templates', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(Array.isArray(data.data)).toBeTruthy();
    
    if (data.data.length > 0) {
      templateId = data.data[0].id;
      expect(data.data[0].is_template).toBe(true);
    }
  });

  test('should filter templates by framework', async ({ request }) => {
    // Get a framework first
    const frameworksResponse = await request.get('http://localhost:8090/api/v1/frameworks?limit=1', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const frameworksData = await frameworksResponse.json();
    
    if (!frameworksData.data || frameworksData.data.length === 0) {
      console.log('[INFO] No frameworks available');
      return;
    }

    const frameworkId = frameworksData.data[0].id;

    const response = await request.get(`http://localhost:8090/api/v1/policy-templates?framework_id=${frameworkId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(Array.isArray(data.data)).toBeTruthy();
  });

  test('should filter templates by category', async ({ request }) => {
    const response = await request.get('http://localhost:8090/api/v1/policy-templates?category=access_control', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(Array.isArray(data.data)).toBeTruthy();
  });

  test('should clone template to new policy', async ({ request }) => {
    test.skip(!templateId, 'No template available to clone');

    const response = await request.post(`http://localhost:8090/api/v1/policy-templates/${templateId}/clone`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        identifier: 'CLONED-001',
        title: 'Cloned Test Policy',
        description: 'Policy cloned from template for testing'
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.data.identifier).toBe('CLONED-001');
    expect(data.data.cloned_from_policy_id).toBe(templateId);
    expect(data.data.is_template).toBe(false);
    expect(data.data.status).toBe('draft');
  });

  test('cloned policy should preserve content from template', async ({ request }) => {
    test.skip(!templateId, 'No template available');

    // Get template content
    const templateResponse = await request.get(`http://localhost:8090/api/v1/policies/${templateId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const templateData = await templateResponse.json();
    const templateContent = templateData.data.current_version.content;

    // Clone template
    const cloneResponse = await request.post(`http://localhost:8090/api/v1/policy-templates/${templateId}/clone`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        identifier: 'PRESERVE-TEST-001',
        title: 'Content Preservation Test',
        description: 'Testing content preservation'
      }
    });

    const cloneData = await cloneResponse.json();
    const clonedPolicyId = cloneData.data.id;

    // Get cloned policy content
    const clonedResponse = await request.get(`http://localhost:8090/api/v1/policies/${clonedPolicyId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const clonedData = await clonedResponse.json();
    const clonedContent = clonedData.data.current_version.content;

    // Content should match
    expect(clonedContent).toBe(templateContent);
  });

  test('should verify template has standard fields', async ({ request }) => {
    test.skip(!templateId, 'No template available');

    const response = await request.get(`http://localhost:8090/api/v1/policies/${templateId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    
    // Templates should have these properties
    expect(data.data.is_template).toBe(true);
    expect(data.data.identifier).toBeDefined();
    expect(data.data.title).toBeDefined();
    expect(data.data.category).toBeDefined();
    expect(data.data.current_version).toBeDefined();
    expect(data.data.current_version.content).toBeDefined();
  });

  test('should list templates for all major frameworks', async ({ request }) => {
    const frameworks = ['SOC 2', 'ISO 27001', 'PCI DSS', 'GDPR', 'CCPA'];
    
    for (const frameworkName of frameworks) {
      // Get framework by name (simplified - in reality would need to search)
      const frameworksResponse = await request.get('http://localhost:8090/api/v1/frameworks', {
        headers: {
          'Authorization': `Bearer ${authToken}`
        }
      });
      const frameworksData = await frameworksResponse.json();
      
      const framework = frameworksData.data?.find((f: any) => f.name.includes(frameworkName));
      
      if (!framework) {
        console.log(`[INFO] Framework not found: ${frameworkName}`);
        continue;
      }

      const response = await request.get(`http://localhost:8090/api/v1/policy-templates?framework_id=${framework.id}`, {
        headers: {
          'Authorization': `Bearer ${authToken}`
        }
      });

      expect(response.ok()).toBeTruthy();
      const data = await response.json();
      
      console.log(`[INFO] Framework ${frameworkName} has ${data.data.length} templates`);
    }
  });

  test('should handle XSS in cloned policy content', async ({ request }) => {
    // CRITICAL: Test Issue #12 - XSS vulnerability from dangerouslySetInnerHTML
    test.skip(!templateId, 'No template available');

    const xssPayload = '<script>alert("XSS")</script><h1>Test</h1>';

    const response = await request.post(`http://localhost:8090/api/v1/policy-templates/${templateId}/clone`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        identifier: 'XSS-TEST-001',
        title: 'XSS Test Policy',
        description: 'Testing XSS sanitization',
        content_override: xssPayload  // If API allows content override on clone
      }
    });

    if (!response.ok()) {
      console.log('[INFO] API does not allow content override on clone');
      return;
    }

    const data = await response.json();
    const policyId = data.data.id;

    // Get the policy content
    const policyResponse = await request.get(`http://localhost:8090/api/v1/policies/${policyId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    const policyData = await policyResponse.json();
    const content = policyData.data.current_version.content;

    // Script tags should be stripped by HTML sanitization
    expect(content).not.toContain('<script>');
    expect(content).not.toContain('alert("XSS")');
    
    if (content.includes('<script>')) {
      console.error('[SECURITY] Issue #12 confirmed: XSS vulnerability - script tags not sanitized');
    }
  });
});
