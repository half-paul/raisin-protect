import { test, expect } from '@playwright/test';

let authToken: string;
let userId: string;
let riskId: string;
let treatmentId: string;

test.beforeAll(async ({ request }) => {
  // Register and create a test risk
  const authResponse = await request.post('http://localhost:8090/api/v1/auth/register', {
    data: {
      email: `qatest-treatment-${Date.now()}@example.com`,
      password: 'Test1234!',
      first_name: 'QA',
      last_name: 'Tester',
      org_name: 'QA Org Treatment Tests',
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
      identifier: 'RISK-TEST-001',
      title: 'Test Risk for Treatments',
      description: 'Testing risk treatment workflow',
      category: 'cyber_security',
      owner_id: userId,
      initial_assessment: {
        inherent_likelihood: 'likely',
        inherent_impact: 'severe'
      }
    }
  });

  const riskBody = await riskResponse.json();
  riskId = riskBody.data.id;
});

test.describe('Risk Treatment Workflow', () => {
  
  test('should create a risk treatment plan', async ({ request }) => {
    const response = await request.post(`http://localhost:8090/api/v1/risks/${riskId}/treatments`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        treatment_type: 'mitigate',
        title: 'Implement EDR Solution',
        description: 'Deploy endpoint detection and response across all endpoints',
        owner_id: userId,
        target_likelihood: 'unlikely',
        target_impact: 'moderate',
        estimated_cost: 50000,
        start_date: '2026-03-01',
        target_completion_date: '2026-06-30'
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.treatment_type).toBe('mitigate');
    expect(body.data.status).toBe('planned');
    expect(body.data.title).toBe('Implement EDR Solution');
    expect(body.data.expected_residual_score).toBe(6); // unlikely(2) × moderate(3) = 6
    
    treatmentId = body.data.id;
  });

  test('should list treatments for a risk', async ({ request }) => {
    const response = await request.get(`http://localhost:8090/api/v1/risks/${riskId}/treatments`, {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.treatments.length).toBeGreaterThan(0);
    expect(body.data.treatments[0].treatment_type).toBe('mitigate');
  });

  test('should update treatment status to in_progress', async ({ request }) => {
    const response = await request.put(`http://localhost:8090/api/v1/risks/${riskId}/treatments/${treatmentId}`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        status: 'in_progress',
        progress_notes: 'EDR vendor selected, deployment starting next week'
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.status).toBe('in_progress');
  });

  test('should complete treatment with effectiveness review', async ({ request }) => {
    const response = await request.post(`http://localhost:8090/api/v1/risks/${riskId}/treatments/${treatmentId}/complete`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        actual_cost: 48000,
        effectiveness_rating: 'effective',
        completion_notes: 'EDR deployed to 100% of endpoints, blocking tested successfully'
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.status).toBe('completed');
    expect(body.data.effectiveness_rating).toBe('effective');
    expect(body.data.actual_cost).toBe(48000);
  });

  test('should create acceptance treatment for low-impact risk', async ({ request }) => {
    const response = await request.post(`http://localhost:8090/api/v1/risks/${riskId}/treatments`, {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        treatment_type: 'accept',
        title: 'Accept Residual Risk',
        description: 'After implementing EDR, residual risk is within appetite',
        owner_id: userId,
        acceptance_reason: 'Cost of further mitigation exceeds residual impact',
        acceptance_valid_until: '2026-12-31'
      }
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();
    expect(body.data.treatment_type).toBe('accept');
  });

  test('should detect treatment gaps', async ({ request }) => {
    // Create a new high risk without treatments
    const riskResponse = await request.post('http://localhost:8090/api/v1/risks', {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
      },
      data: {
        identifier: 'RISK-GAP-001',
        title: 'Unmitigated High Risk',
        description: 'High risk with no treatment plan',
        category: 'operational',
        owner_id: userId,
        initial_assessment: {
          inherent_likelihood: 'likely',
          inherent_impact: 'major'
        }
      }
    });

    expect(riskResponse.ok()).toBeTruthy();

    // Check gap detection
    const gapResponse = await request.get('http://localhost:8090/api/v1/risks/gaps', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });

    expect(gapResponse.ok()).toBeTruthy();
    const gapBody = await gapResponse.json();
    
    // Should have at least one gap (the new risk without treatments)
    expect(gapBody.data.gaps.length).toBeGreaterThan(0);
    const highRiskNoTreatment = gapBody.data.gaps.find((g: any) => 
      g.gap_type === 'high_risk_no_treatments'
    );
    expect(highRiskNoTreatment).toBeDefined();
  });
});
