import { test, expect } from '@playwright/test';

let authToken: string;
let userId: string;
let riskId: string;
let assessmentId: string;

test.beforeAll(async ({ request }) => {
  // Register and create a test risk
  const authResponse = await request.post('http://localhost:8090/api/v1/auth/register', {
    data: {
      email: `qatest-assessment-${Date.now()}@example.com`,
      password: 'Test1234!',
      first_name: 'QA',
      last_name: 'Tester',
      org_name: 'QA Org Assessment Tests',
      role: 'compliance_manager'
    }
  });
  
  const authBody = await authResponse.json();
  authToken = authBody.data.access_token;
  userId = authBody.data.user.id;

  // Create a risk
  const riskResponse = await request.post('http://localhost:8090/api/v1/risks', {
    headers: {
      'Authorization': `Bearer ${authToken}`,
      'Content-Type': 'application/json'
    },
    data: {
      title: 'Test Risk for Assessments',
      description: 'Testing risk assessment workflow',
      category: 'operational',
      owner_id: userId,
      initial_likelihood: 'medium',
      initial_impact': 'medium'
    }
  });

  const riskBody = await riskResponse.json();
  riskId = riskBody.data.risk.id;
});

test.describe('Risk Assessment Workflow', () => {
  
  test('should create a risk assessment', async ({ request }) => {
    const response = await request.post(`http://localhost:8090/api/v1/risks/${riskId}/assessments`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        assessment_type: 'residual',
        likelihood: 'low',
        impact: 'medium',
        rationale: 'After implementing controls, likelihood reduced'
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.assessment_type).toBe('residual');
    expect(body.data.likelihood).toBe('low');
    expect(body.data.impact).toBe('medium');
    expect(body.data.score).toBe(6); // low(2) × medium(3) = 6
    expect(body.data.severity).toBe('low'); // score 6 = low severity
    
    assessmentId = body.data.id;
  });

  test('should list assessments for a risk', async ({ request }) => {
    const response = await request.get(`http://localhost:8090/api/v1/risks/${riskId}/assessments`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.assessments.length).toBeGreaterThan(0);
    
    // Should have both inherent (initial) and residual
    const types = body.data.assessments.map((a: any) => a.assessment_type);
    expect(types).toContain('inherent');
    expect(types).toContain('residual');
  });

  test('should calculate correct risk scores', async ({ request }) => {
    const testCases = [
      { likelihood: 'very_low', impact: 'very_low', expectedScore: 1, expectedSeverity: 'low' },
      { likelihood: 'low', impact: 'medium', expectedScore: 6, expectedSeverity: 'low' },
      { likelihood: 'medium', impact: 'high', expectedScore: 12, expectedSeverity: 'medium' },
      { likelihood: 'high', impact: 'critical', expectedScore: 20, expectedSeverity: 'high' },
      { likelihood: 'very_high', impact: 'critical', expectedScore: 25, expectedSeverity: 'critical' }
    ];

    for (const testCase of testCases) {
      const response = await request.post(`http://localhost:8090/api/v1/risks/${riskId}/assessments`, {
        headers: {
          'Authorization': `Bearer ${authToken}`,
          'Content-Type': 'application/json'
        },
        data: {
          assessment_type: 'residual',
          likelihood: testCase.likelihood,
          impact: testCase.impact,
          rationale: `Testing ${testCase.likelihood} × ${testCase.impact}`
        }
      });

      expect(response.ok()).toBeTruthy();
      const body = await response.json();
      expect(body.data.score).toBe(testCase.expectedScore);
      expect(body.data.severity).toBe(testCase.expectedSeverity);
    }
  });

  test('should recalculate all risk scores', async ({ request }) => {
    const response = await request.post('http://localhost:8090/api/v1/risks/recalculate-scores', {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.recalculated_count).toBeGreaterThan(0);
  });

  test('should require rationale for assessments', async ({ request }) => {
    const response = await request.post(`http://localhost:8090/api/v1/risks/${riskId}/assessments`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        assessment_type: 'residual',
        likelihood: 'medium',
        impact: 'high'
        // Missing rationale
      }
    });

    expect(response.status()).toBe(400); // Should fail validation
  });
});
