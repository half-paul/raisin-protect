/**
 * API client helpers for Raisin Protect Dashboard
 * Sprint 2: Frameworks & Controls
 * Sprint 3: Evidence Management
 * Sprint 4: Continuous Monitoring Engine
 * Sprint 5: Policy Management
 * Sprint 6: Risk Management
 * Sprint 7: Audit Hub
 */

import { authFetch } from './auth';

const API_BASE = '';

// ========== Types ==========

export interface Framework {
  id: string;
  identifier: string;
  name: string;
  description: string;
  category: string;
  website_url: string | null;
  logo_url: string | null;
  versions_count: number;
  created_at: string;
}

export interface FrameworkVersion {
  id: string;
  framework_id?: string;
  framework_identifier?: string;
  framework_name?: string;
  version: string;
  display_name: string;
  status: string;
  effective_date: string | null;
  sunset_date: string | null;
  changelog?: string | null;
  total_requirements: number;
  created_at: string;
}

export interface FrameworkDetail extends Framework {
  versions: FrameworkVersion[];
}

export interface Requirement {
  id: string;
  identifier: string;
  title: string;
  description?: string;
  guidance?: string;
  parent_id: string | null;
  depth: number;
  section_order: number;
  is_assessable: boolean;
  created_at: string;
  children?: Requirement[];
}

export interface OrgFrameworkStats {
  total_requirements: number;
  in_scope: number;
  out_of_scope: number;
  mapped: number;
  unmapped: number;
  coverage_pct: number;
}

export interface OrgFramework {
  id: string;
  framework: {
    id: string;
    identifier: string;
    name: string;
    category: string;
  };
  active_version: {
    id: string;
    version: string;
    display_name: string;
    total_requirements: number;
  };
  status: string;
  target_date: string | null;
  notes: string | null;
  stats: OrgFrameworkStats;
  activated_at: string;
  created_at: string;
}

export interface CoverageRequirement {
  id: string;
  identifier: string;
  title: string;
  depth: number;
  in_scope: boolean;
  status: 'covered' | 'gap';
  controls: {
    id: string;
    identifier: string;
    title: string;
    strength: string;
    status: string;
  }[];
}

export interface CoverageSummary {
  total_requirements: number;
  assessable_requirements: number;
  in_scope: number;
  out_of_scope: number;
  covered: number;
  gaps: number;
  coverage_pct: number;
}

export interface CoverageData {
  framework: {
    identifier: string;
    name: string;
    version: string;
  };
  summary: CoverageSummary;
  requirements: CoverageRequirement[];
}

export interface ScopingDecision {
  id: string;
  requirement: {
    id: string;
    identifier: string;
    title: string;
  };
  in_scope: boolean;
  justification: string | null;
  scoped_by: {
    id: string;
    name: string;
  };
  updated_at: string;
}

export interface Control {
  id: string;
  identifier: string;
  title: string;
  description: string;
  implementation_guidance?: string;
  category: string;
  status: string;
  is_custom: boolean;
  owner: { id: string; name: string; email?: string } | null;
  secondary_owner: { id: string; name: string } | null;
  evidence_requirements?: string;
  test_criteria?: string;
  metadata?: Record<string, unknown>;
  mappings_count: number;
  frameworks?: string[];
  source_template_id?: string;
  created_at: string;
  updated_at: string;
}

export interface ControlMapping {
  id: string;
  requirement: {
    id: string;
    identifier: string;
    title: string;
    framework?: { identifier: string; name: string };
    version?: string;
  };
  strength: string;
  notes: string | null;
  mapped_by?: { id: string; name: string };
  created_at: string;
}

export interface ControlDetail extends Control {
  mappings: ControlMapping[];
}

export interface MappingMatrixFramework {
  id: string;
  identifier: string;
  name: string;
  version: string;
}

export interface MappingMatrixControl {
  id: string;
  identifier: string;
  title: string;
  category: string;
  status: string;
  mappings_by_framework: Record<
    string,
    { requirement_id: string; identifier: string; strength: string }[]
  >;
}

export interface MappingMatrixData {
  frameworks: MappingMatrixFramework[];
  controls: MappingMatrixControl[];
}

export interface ControlStats {
  total: number;
  by_status: Record<string, number>;
  by_category: Record<string, number>;
  custom_count: number;
  library_count: number;
  unowned_count: number;
  unmapped_count: number;
  frameworks_coverage: {
    framework: string;
    version: string;
    in_scope: number;
    covered: number;
    gaps: number;
    coverage_pct: number;
  }[];
}

export interface PaginatedMeta {
  total: number;
  total_controls?: number;
  page?: number;
  per_page?: number;
  request_id?: string;
}

// ========== API Functions ==========

async function apiGet<T>(path: string, params?: Record<string, string>): Promise<{ data: T; meta?: PaginatedMeta }> {
  const url = new URL(`${API_BASE}${path}`, window.location.origin);
  if (params) {
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') url.searchParams.set(k, v);
    });
  }
  const res = await authFetch(url.toString());
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err?.error?.message || `API error ${res.status}`);
  }
  return res.json();
}

async function apiPost<T>(path: string, body: unknown): Promise<{ data: T }> {
  const res = await authFetch(`${API_BASE}${path}`, {
    method: 'POST',
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err?.error?.message || `API error ${res.status}`);
  }
  return res.json();
}

async function apiPut<T>(path: string, body: unknown): Promise<{ data: T }> {
  const res = await authFetch(`${API_BASE}${path}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err?.error?.message || `API error ${res.status}`);
  }
  return res.json();
}

async function apiDelete<T>(path: string): Promise<{ data: T }> {
  const res = await authFetch(`${API_BASE}${path}`, { method: 'DELETE' });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err?.error?.message || `API error ${res.status}`);
  }
  return res.json();
}

// ---- Frameworks ----

export function listFrameworks(params?: { category?: string; search?: string }) {
  return apiGet<Framework[]>('/api/v1/frameworks', params as Record<string, string>);
}

export function getFramework(id: string) {
  return apiGet<FrameworkDetail>(`/api/v1/frameworks/${id}`);
}

export function getFrameworkVersion(frameworkId: string, versionId: string) {
  return apiGet<FrameworkVersion>(`/api/v1/frameworks/${frameworkId}/versions/${versionId}`);
}

export function listRequirements(
  frameworkId: string,
  versionId: string,
  params?: { format?: string; assessable_only?: string; parent_id?: string; search?: string; page?: string; per_page?: string }
) {
  return apiGet<Requirement[]>(
    `/api/v1/frameworks/${frameworkId}/versions/${versionId}/requirements`,
    params as Record<string, string>
  );
}

// ---- Org Frameworks ----

export function listOrgFrameworks(params?: { status?: string }) {
  return apiGet<OrgFramework[]>('/api/v1/org-frameworks', params as Record<string, string>);
}

export function activateFramework(body: {
  framework_id: string;
  version_id: string;
  target_date?: string;
  notes?: string;
  seed_controls?: boolean;
}) {
  return apiPost<OrgFramework>('/api/v1/org-frameworks', body);
}

export function updateOrgFramework(id: string, body: {
  version_id?: string;
  target_date?: string | null;
  notes?: string;
  status?: string;
}) {
  return apiPut<OrgFramework>(`/api/v1/org-frameworks/${id}`, body);
}

export function deactivateOrgFramework(id: string) {
  return apiDelete<{ id: string; status: string; message: string }>(`/api/v1/org-frameworks/${id}`);
}

export function getOrgFrameworkCoverage(
  id: string,
  params?: { status?: string; page?: string; per_page?: string }
) {
  return apiGet<CoverageData>(`/api/v1/org-frameworks/${id}/coverage`, params as Record<string, string>);
}

// ---- Requirement Scoping ----

export function listScoping(orgFrameworkId: string, params?: { in_scope?: string; page?: string; per_page?: string }) {
  return apiGet<ScopingDecision[]>(
    `/api/v1/org-frameworks/${orgFrameworkId}/scoping`,
    params as Record<string, string>
  );
}

export function setRequirementScope(
  orgFrameworkId: string,
  requirementId: string,
  body: { in_scope: boolean; justification?: string }
) {
  return apiPut<ScopingDecision>(
    `/api/v1/org-frameworks/${orgFrameworkId}/requirements/${requirementId}/scope`,
    body
  );
}

export function resetRequirementScope(orgFrameworkId: string, requirementId: string) {
  return apiDelete<{ message: string }>(
    `/api/v1/org-frameworks/${orgFrameworkId}/requirements/${requirementId}/scope`
  );
}

// ---- Controls ----

export function listControls(params?: Record<string, string>) {
  return apiGet<Control[]>('/api/v1/controls', params);
}

export function getControl(id: string) {
  return apiGet<ControlDetail>(`/api/v1/controls/${id}`);
}

export function createControl(body: Partial<Control>) {
  return apiPost<Control>('/api/v1/controls', body);
}

export function updateControl(id: string, body: Partial<Control>) {
  return apiPut<Control>(`/api/v1/controls/${id}`, body);
}

export function deprecateControl(id: string) {
  return apiDelete<{ id: string; status: string; message: string }>(`/api/v1/controls/${id}`);
}

export function changeControlStatus(id: string, status: string) {
  return apiPut<{ id: string; identifier: string; status: string; previous_status: string; message: string }>(
    `/api/v1/controls/${id}/status`,
    { status }
  );
}

export function bulkControlStatus(controlIds: string[], status: string) {
  return apiPost<{
    updated: number;
    failed: number;
    results: { id: string; identifier: string; status: string; success: boolean }[];
  }>('/api/v1/controls/bulk-status', { control_ids: controlIds, status });
}

export function getControlStats() {
  return apiGet<ControlStats>('/api/v1/controls/stats');
}

// ---- Control Mappings ----

export function listControlMappings(controlId: string) {
  return apiGet<ControlMapping[]>(`/api/v1/controls/${controlId}/mappings`);
}

export function createControlMappings(
  controlId: string,
  mappings: { requirement_id: string; strength?: string; notes?: string }[]
) {
  if (mappings.length === 1) {
    return apiPost<{ created: number }>(`/api/v1/controls/${controlId}/mappings`, mappings[0]);
  }
  return apiPost<{ created: number }>(`/api/v1/controls/${controlId}/mappings`, { mappings });
}

export function deleteControlMapping(controlId: string, mappingId: string) {
  return apiDelete<{ message: string }>(`/api/v1/controls/${controlId}/mappings/${mappingId}`);
}

// ---- Mapping Matrix ----

export function getMappingMatrix(params?: Record<string, string>) {
  return apiGet<MappingMatrixData>('/api/v1/mapping-matrix', params);
}

// ========== Evidence Types ==========

export type EvidenceType =
  | 'screenshot'
  | 'api_response'
  | 'configuration_export'
  | 'log_sample'
  | 'policy_document'
  | 'access_list'
  | 'vulnerability_report'
  | 'certificate'
  | 'training_record'
  | 'penetration_test'
  | 'audit_report'
  | 'other';

export type EvidenceStatus =
  | 'draft'
  | 'pending_review'
  | 'approved'
  | 'rejected'
  | 'expired'
  | 'superseded';

export type CollectionMethod =
  | 'manual_upload'
  | 'automated_pull'
  | 'api_ingestion'
  | 'screenshot_capture'
  | 'system_export';

export type FreshnessStatus = 'fresh' | 'expiring_soon' | 'expired';

export type EvidenceVerdict = 'sufficient' | 'partial' | 'insufficient' | 'needs_update';

export type LinkStrength = 'primary' | 'supporting' | 'supplementary';

export interface EvidenceArtifact {
  id: string;
  title: string;
  description?: string;
  evidence_type: EvidenceType;
  status: EvidenceStatus;
  collection_method: CollectionMethod;
  file_name: string;
  file_size: number;
  mime_type: string;
  object_key?: string;
  checksum_sha256?: string;
  version: number;
  is_current: boolean;
  total_versions?: number;
  collection_date: string;
  expires_at: string | null;
  freshness_period_days: number | null;
  freshness_status: FreshnessStatus | null;
  days_until_expiry?: number | null;
  source_system: string | null;
  uploaded_by: { id: string; name: string; email?: string };
  tags: string[];
  metadata?: Record<string, unknown>;
  links_count?: number;
  evaluations_count?: number;
  links?: EvidenceLink[];
  latest_evaluation?: {
    id?: string;
    verdict: EvidenceVerdict;
    confidence: string;
    comments?: string;
    evaluated_by?: { id: string; name: string };
    evaluated_at?: string;
    created_at?: string;
  } | null;
  parent_artifact_id?: string;
  created_at: string;
  updated_at: string;
}

export interface EvidenceLink {
  id: string;
  target_type: 'control' | 'requirement';
  control?: {
    id: string;
    identifier: string;
    title: string;
    category?: string;
    status?: string;
  } | null;
  requirement?: {
    id: string;
    identifier: string;
    title: string;
    framework?: string;
    framework_version?: string;
  } | null;
  strength: LinkStrength;
  notes: string | null;
  linked_by?: { id: string; name: string };
  created_at: string;
}

export interface EvidenceEvaluation {
  id: string;
  artifact_id?: string;
  evidence_link_id?: string;
  verdict: EvidenceVerdict;
  confidence: string;
  comments: string;
  missing_elements: string[];
  remediation_notes: string | null;
  evidence_link?: {
    id: string;
    target_type: string;
    control_identifier?: string;
  } | null;
  evaluated_by: { id: string; name: string; role?: string };
  created_at: string;
}

export interface EvidenceVersion {
  id: string;
  version: number;
  is_current: boolean;
  title: string;
  status: EvidenceStatus;
  file_name: string;
  file_size: number;
  collection_date: string;
  uploaded_by: { id: string; name: string };
  created_at: string;
}

export interface UploadInfo {
  presigned_url: string;
  method: string;
  expires_in: number;
  max_size: number;
  content_type: string;
}

export interface StalenessAlert {
  id: string;
  title: string;
  evidence_type: EvidenceType;
  status: EvidenceStatus;
  collection_date: string;
  expires_at: string;
  freshness_period_days: number;
  alert_level: 'expired' | 'expiring_soon';
  days_overdue?: number;
  days_until_expiry?: number;
  linked_controls: { id: string; identifier: string; title: string }[];
  linked_controls_count: number;
  uploaded_by: { id: string; name: string };
}

export interface StalenessSummary {
  total_alerts: number;
  expired: number;
  expiring_soon: number;
  affected_controls: number;
}

export interface FreshnessSummary {
  total_evidence: number;
  by_freshness: { fresh: number; expiring_soon: number; expired: number; no_expiry: number };
  by_status: Record<string, number>;
  by_type: Record<string, number>;
  coverage: {
    total_active_controls: number;
    controls_with_evidence: number;
    controls_without_evidence: number;
    evidence_coverage_pct: number;
  };
}

export interface ControlEvidence {
  control: { id: string; identifier: string; title: string };
  evidence_summary: {
    total: number;
    approved: number;
    pending_review: number;
    fresh: number;
    expiring_soon: number;
    expired: number;
  };
  evidence: (EvidenceArtifact & {
    link: { id: string; strength: string; notes: string | null };
  })[];
}

// ========== Evidence API Functions ==========

export function listEvidence(params?: Record<string, string>) {
  return apiGet<EvidenceArtifact[]>('/api/v1/evidence', params);
}

export function getEvidence(id: string) {
  return apiGet<EvidenceArtifact>(`/api/v1/evidence/${id}`);
}

export function createEvidence(body: {
  title: string;
  description?: string;
  evidence_type: string;
  collection_method?: string;
  file_name: string;
  file_size: number;
  mime_type: string;
  collection_date: string;
  freshness_period_days?: number;
  source_system?: string;
  tags?: string[];
}) {
  return apiPost<EvidenceArtifact & { upload: UploadInfo }>('/api/v1/evidence', body);
}

export function updateEvidence(id: string, body: Partial<{
  title: string;
  description: string;
  evidence_type: string;
  collection_date: string;
  freshness_period_days: number;
  source_system: string;
  tags: string[];
}>) {
  return apiPut<EvidenceArtifact>(`/api/v1/evidence/${id}`, body);
}

export function deleteEvidence(id: string) {
  return apiDelete<{ id: string; status: string; message: string }>(`/api/v1/evidence/${id}`);
}

export function confirmEvidenceUpload(id: string, body?: { checksum_sha256?: string }) {
  return apiPost<{ id: string; status: string; file_verified: boolean; message: string }>(
    `/api/v1/evidence/${id}/confirm`,
    body || {}
  );
}

export function getUploadURL(id: string) {
  return apiPost<{ id: string; upload: UploadInfo }>(`/api/v1/evidence/${id}/upload`, {});
}

export function getDownloadURL(id: string, version?: number) {
  const params: Record<string, string> = {};
  if (version) params.version = String(version);
  return apiGet<{ id: string; file_name: string; file_size: number; mime_type: string; download: { presigned_url: string; method: string; expires_in: number } }>(
    `/api/v1/evidence/${id}/download`,
    params
  );
}

export function changeEvidenceStatus(id: string, status: string) {
  return apiPut<{ id: string; status: string; previous_status: string; message: string }>(
    `/api/v1/evidence/${id}/status`,
    { status }
  );
}

export function createEvidenceVersion(id: string, body: {
  title?: string;
  description?: string;
  file_name: string;
  file_size: number;
  mime_type: string;
  collection_date: string;
  freshness_period_days?: number;
  tags?: string[];
}) {
  return apiPost<EvidenceArtifact & { previous_version: { id: string; version: number; status: string }; upload: UploadInfo }>(
    `/api/v1/evidence/${id}/versions`,
    body
  );
}

export function listEvidenceVersions(id: string) {
  return apiGet<EvidenceVersion[]>(`/api/v1/evidence/${id}/versions`);
}

export function listEvidenceLinks(id: string) {
  return apiGet<EvidenceLink[]>(`/api/v1/evidence/${id}/links`);
}

export function createEvidenceLinks(id: string, links: {
  target_type: string;
  control_id?: string;
  requirement_id?: string;
  strength?: string;
  notes?: string;
}[]) {
  if (links.length === 1) {
    return apiPost<{ created: number; links: EvidenceLink[] }>(`/api/v1/evidence/${id}/links`, links[0]);
  }
  return apiPost<{ created: number; links: EvidenceLink[] }>(`/api/v1/evidence/${id}/links`, { links });
}

export function deleteEvidenceLink(evidenceId: string, linkId: string) {
  return apiDelete<{ message: string }>(`/api/v1/evidence/${evidenceId}/links/${linkId}`);
}

export function listControlEvidence(controlId: string, params?: Record<string, string>) {
  return apiGet<ControlEvidence>(`/api/v1/controls/${controlId}/evidence`, params);
}

export function getStalenessAlerts(params?: Record<string, string>) {
  return apiGet<{ summary: StalenessSummary; alerts: StalenessAlert[] }>('/api/v1/evidence/staleness', params);
}

export function getFreshnessSummary() {
  return apiGet<FreshnessSummary>('/api/v1/evidence/freshness-summary');
}

export function listEvidenceEvaluations(id: string, params?: Record<string, string>) {
  return apiGet<EvidenceEvaluation[]>(`/api/v1/evidence/${id}/evaluations`, params);
}

export function createEvidenceEvaluation(id: string, body: {
  evidence_link_id?: string;
  verdict: string;
  confidence?: string;
  comments: string;
  missing_elements?: string[];
  remediation_notes?: string;
}) {
  return apiPost<EvidenceEvaluation>(`/api/v1/evidence/${id}/evaluations`, body);
}

export function searchEvidence(params?: Record<string, string>) {
  return apiGet<EvidenceArtifact[]>('/api/v1/evidence/search', params);
}

// ========== Sprint 4: Continuous Monitoring Types ==========

export type TestType =
  | 'configuration'
  | 'access_control'
  | 'endpoint'
  | 'vulnerability'
  | 'data_protection'
  | 'network'
  | 'logging'
  | 'custom';

export type TestStatus = 'draft' | 'active' | 'paused' | 'deprecated';
export type TestSeverity = 'critical' | 'high' | 'medium' | 'low' | 'informational';
export type TestResultStatus = 'pass' | 'fail' | 'error' | 'skip' | 'warning';
export type TestRunStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
export type TestRunTrigger = 'scheduled' | 'manual' | 'on_change' | 'webhook';

export type AlertSeverity = 'critical' | 'high' | 'medium' | 'low';
export type AlertStatus = 'open' | 'acknowledged' | 'in_progress' | 'resolved' | 'suppressed' | 'closed';
export type AlertDeliveryChannel = 'slack' | 'email' | 'webhook' | 'in_app';

export interface TestDefinition {
  id: string;
  identifier: string;
  title: string;
  description?: string;
  test_type: TestType;
  severity: TestSeverity;
  status: TestStatus;
  control?: {
    id: string;
    identifier: string;
    title: string;
    category?: string;
    status?: string;
  };
  schedule_cron?: string | null;
  schedule_interval_min?: number | null;
  next_run_at?: string | null;
  last_run_at?: string | null;
  timeout_seconds?: number;
  retry_count?: number;
  retry_delay_seconds?: number;
  test_config?: Record<string, unknown>;
  test_script?: string | null;
  test_script_language?: string | null;
  tags: string[];
  created_by?: { id: string; name: string };
  latest_result?: {
    id?: string;
    status: TestResultStatus;
    message?: string;
    tested_at?: string;
    duration_ms?: number;
  } | null;
  result_summary?: {
    total_runs: number;
    last_24h: Record<TestResultStatus, number>;
  };
  active_alerts?: number;
  created_at: string;
  updated_at: string;
}

export interface TestRun {
  id: string;
  run_number: number;
  status: TestRunStatus;
  trigger_type: TestRunTrigger;
  started_at?: string | null;
  completed_at?: string | null;
  duration_ms?: number | null;
  total_tests: number;
  passed: number;
  failed: number;
  errors: number;
  skipped: number;
  warnings: number;
  triggered_by?: { id: string; name: string } | null;
  trigger_metadata?: Record<string, unknown>;
  worker_id?: string | null;
  error_message?: string | null;
  created_at: string;
  updated_at?: string;
}

export interface TestResult {
  id: string;
  test_run_id?: string;
  run_number?: number;
  test: {
    id: string;
    identifier: string;
    title: string;
    test_type: TestType;
    severity?: TestSeverity;
  };
  control?: {
    id: string;
    identifier: string;
    title: string;
  };
  status: TestResultStatus;
  severity: TestSeverity;
  message?: string;
  details?: Record<string, unknown>;
  output_log?: string;
  error_message?: string | null;
  duration_ms?: number;
  alert_generated?: boolean;
  alert_id?: string | null;
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

export interface Alert {
  id: string;
  alert_number: number;
  title: string;
  description?: string;
  severity: AlertSeverity;
  status: AlertStatus;
  control?: {
    id: string;
    identifier: string;
    title: string;
    category?: string;
  };
  test?: {
    id: string;
    identifier: string;
    title: string;
    test_type?: TestType;
  };
  test_result?: {
    id: string;
    status: TestResultStatus;
    message?: string;
    details?: Record<string, unknown>;
    tested_at?: string;
  };
  alert_rule?: { id: string; name: string } | null;
  assigned_to?: { id: string; name: string; email?: string } | null;
  assigned_at?: string | null;
  assigned_by?: { id: string; name: string } | null;
  sla_deadline?: string | null;
  sla_breached?: boolean;
  hours_remaining?: number | null;
  resolved_by?: { id: string; name: string } | null;
  resolved_at?: string | null;
  resolution_notes?: string | null;
  suppressed_until?: string | null;
  suppression_reason?: string | null;
  delivery_channels?: AlertDeliveryChannel[];
  delivered_at?: Record<string, string>;
  tags?: string[];
  metadata?: Record<string, unknown>;
  control_identifier?: string;
  test_identifier?: string;
  assigned_to_name?: string | null;
  created_at: string;
  updated_at: string;
}

export interface AlertRule {
  id: string;
  name: string;
  description?: string;
  enabled: boolean;
  match_test_types?: TestType[] | null;
  match_severities?: TestSeverity[] | null;
  match_result_statuses?: TestResultStatus[] | null;
  match_control_ids?: string[] | null;
  match_tags?: string[] | null;
  consecutive_failures: number;
  cooldown_minutes: number;
  alert_severity: AlertSeverity;
  alert_title_template?: string | null;
  auto_assign_to?: string | null;
  sla_hours?: number | null;
  delivery_channels: AlertDeliveryChannel[];
  slack_webhook_url?: string;
  email_recipients?: string[];
  webhook_url?: string;
  webhook_headers?: Record<string, string>;
  priority: number;
  alerts_generated?: number;
  created_by?: { id: string; name: string };
  created_at: string;
  updated_at: string;
}

export interface HeatmapControl {
  id: string;
  identifier: string;
  title: string;
  category: string;
  health_status: 'healthy' | 'failing' | 'error' | 'warning' | 'untested';
  latest_result?: {
    status: TestResultStatus;
    severity: TestSeverity;
    message?: string;
    tested_at?: string;
  } | null;
  active_alerts: number;
  tests_count: number;
}

export interface HeatmapData {
  summary: {
    total_controls: number;
    healthy: number;
    failing: number;
    error: number;
    warning: number;
    untested: number;
  };
  controls: HeatmapControl[];
}

export interface PostureFramework {
  framework_id: string;
  framework_name: string;
  framework_version: string;
  org_framework_id: string;
  total_mapped_controls: number;
  passing: number;
  failing: number;
  untested: number;
  posture_score: number;
  trend?: {
    '7d_ago': number;
    '30d_ago': number;
    direction: 'improving' | 'declining' | 'stable';
  };
}

export interface PostureData {
  overall_score: number;
  frameworks: PostureFramework[];
}

export interface MonitoringSummary {
  overall_posture_score: number;
  controls: {
    total_active: number;
    healthy: number;
    failing: number;
    untested: number;
    health_rate: number;
  };
  tests: {
    total_active: number;
    last_run?: {
      run_number: number;
      status: string;
      completed_at?: string;
      passed: number;
      failed: number;
      errors: number;
    } | null;
    pass_rate_24h: number;
  };
  alerts: {
    open: number;
    acknowledged: number;
    in_progress: number;
    sla_breached: number;
    resolved_today: number;
    by_severity: Record<AlertSeverity, number>;
  };
  recent_activity: {
    type: string;
    alert_number?: number;
    run_number?: number;
    title?: string;
    severity?: string;
    passed?: number;
    failed?: number;
    resolved_by?: string;
    timestamp: string;
  }[];
}

export interface AlertQueueData {
  queue_summary: {
    active: number;
    resolved: number;
    suppressed: number;
    closed: number;
    sla_breached: number;
  };
  alerts: Alert[];
}

// ========== Sprint 4: Monitoring API Functions ==========

// ---- Tests ----

export function listTests(params?: Record<string, string>) {
  return apiGet<TestDefinition[]>('/api/v1/tests', params);
}

export function getTest(id: string) {
  return apiGet<TestDefinition>(`/api/v1/tests/${id}`);
}

export function createTest(body: Partial<TestDefinition>) {
  return apiPost<TestDefinition>('/api/v1/tests', body);
}

export function updateTest(id: string, body: Partial<TestDefinition>) {
  return apiPut<TestDefinition>(`/api/v1/tests/${id}`, body);
}

export function changeTestStatus(id: string, status: string) {
  return apiPut<{ id: string; status: string; previous_status: string; next_run_at?: string; message: string }>(
    `/api/v1/tests/${id}/status`,
    { status }
  );
}

export function deleteTest(id: string) {
  return apiDelete<{ id: string; status: string; message: string }>(`/api/v1/tests/${id}`);
}

export function getTestResults(testId: string, params?: Record<string, string>) {
  return apiGet<{ test: { id: string; identifier: string; title: string }; results: TestResult[] }>(
    `/api/v1/tests/${testId}/results`,
    params
  );
}

// ---- Test Runs ----

export function createTestRun(body?: { test_ids?: string[]; trigger_metadata?: Record<string, unknown> }) {
  return apiPost<TestRun>('/api/v1/test-runs', body || {});
}

export function listTestRuns(params?: Record<string, string>) {
  return apiGet<TestRun[]>('/api/v1/test-runs', params);
}

export function getTestRun(id: string) {
  return apiGet<TestRun>(`/api/v1/test-runs/${id}`);
}

export function cancelTestRun(id: string) {
  return apiPost<{ id: string; status: string; previous_status: string; message: string }>(
    `/api/v1/test-runs/${id}/cancel`,
    {}
  );
}

export function listTestRunResults(runId: string, params?: Record<string, string>) {
  return apiGet<TestResult[]>(`/api/v1/test-runs/${runId}/results`, params);
}

export function getTestRunResult(runId: string, resultId: string) {
  return apiGet<TestResult>(`/api/v1/test-runs/${runId}/results/${resultId}`);
}

// ---- Control Test Results ----

export function getControlTestResults(controlId: string, params?: Record<string, string>) {
  return apiGet<{
    control: { id: string; identifier: string; title: string };
    health_status: string;
    tests_count: number;
    results: TestResult[];
  }>(`/api/v1/controls/${controlId}/test-results`, params);
}

// ---- Alerts ----

export function listAlerts(params?: Record<string, string>) {
  return apiGet<Alert[]>('/api/v1/alerts', params);
}

export function getAlert(id: string) {
  return apiGet<Alert>(`/api/v1/alerts/${id}`);
}

export function changeAlertStatus(id: string, status: string) {
  return apiPut<{ id: string; alert_number: number; status: string; previous_status: string; message: string }>(
    `/api/v1/alerts/${id}/status`,
    { status }
  );
}

export function assignAlert(id: string, assignedTo: string) {
  return apiPut<Alert>(`/api/v1/alerts/${id}/assign`, { assigned_to: assignedTo });
}

export function resolveAlert(id: string, resolutionNotes: string) {
  return apiPut<Alert>(`/api/v1/alerts/${id}/resolve`, { resolution_notes: resolutionNotes });
}

export function suppressAlert(id: string, suppressedUntil: string, suppressionReason: string) {
  return apiPut<Alert>(`/api/v1/alerts/${id}/suppress`, {
    suppressed_until: suppressedUntil,
    suppression_reason: suppressionReason,
  });
}

export function closeAlert(id: string, resolutionNotes?: string) {
  return apiPut<Alert>(`/api/v1/alerts/${id}/close`, { resolution_notes: resolutionNotes });
}

export function redeliverAlert(id: string, channels?: string[]) {
  return apiPost<{ id: string; alert_number: number; delivery_results: Record<string, unknown>; message: string }>(
    `/api/v1/alerts/${id}/deliver`,
    channels ? { channels } : {}
  );
}

export function testAlertDelivery(body: {
  channel: string;
  slack_webhook_url?: string;
  email_recipients?: string[];
  webhook_url?: string;
  webhook_headers?: Record<string, string>;
}) {
  return apiPost<{ channel: string; success: boolean; message: string }>(
    '/api/v1/alerts/test-delivery',
    body
  );
}

// ---- Alert Rules ----

export function listAlertRules(params?: Record<string, string>) {
  return apiGet<AlertRule[]>('/api/v1/alert-rules', params);
}

export function getAlertRule(id: string) {
  return apiGet<AlertRule>(`/api/v1/alert-rules/${id}`);
}

export function createAlertRule(body: Partial<AlertRule>) {
  return apiPost<AlertRule>('/api/v1/alert-rules', body);
}

export function updateAlertRule(id: string, body: Partial<AlertRule>) {
  return apiPut<AlertRule>(`/api/v1/alert-rules/${id}`, body);
}

export function deleteAlertRule(id: string) {
  return apiDelete<{ id: string; message: string }>(`/api/v1/alert-rules/${id}`);
}

// ---- Monitoring Dashboard ----

export function getMonitoringHeatmap(params?: Record<string, string>) {
  return apiGet<HeatmapData>('/api/v1/monitoring/heatmap', params);
}

export function getMonitoringPosture() {
  return apiGet<PostureData>('/api/v1/monitoring/posture');
}

export function getMonitoringSummary() {
  return apiGet<MonitoringSummary>('/api/v1/monitoring/summary');
}

export function getAlertQueue(params?: Record<string, string>) {
  return apiGet<AlertQueueData>('/api/v1/monitoring/alert-queue', params);
}

// ========== Sprint 5: Policy Management Types ==========

export type PolicyStatus = 'draft' | 'in_review' | 'approved' | 'published' | 'archived';

export type PolicyCategory =
  | 'information_security'
  | 'access_control'
  | 'incident_response'
  | 'data_privacy'
  | 'network_security'
  | 'encryption'
  | 'vulnerability_management'
  | 'change_management'
  | 'business_continuity'
  | 'secure_development'
  | 'vendor_management'
  | 'acceptable_use'
  | 'physical_security'
  | 'hr_security'
  | 'asset_management';

export type SignoffStatus = 'pending' | 'approved' | 'rejected' | 'withdrawn';

export type PolicyChangeType = 'initial' | 'major' | 'minor' | 'patch';

export type PolicyContentFormat = 'html' | 'markdown' | 'plain_text';

export type PolicyReviewStatus = 'overdue' | 'due_soon' | 'on_track' | 'no_schedule';

export interface PolicyOwner {
  id: string;
  name: string;
  email?: string;
}

export interface PolicyCurrentVersion {
  id: string;
  version_number: number;
  content?: string;
  content_format?: PolicyContentFormat;
  content_summary?: string;
  change_summary?: string;
  change_type?: PolicyChangeType;
  word_count?: number;
  character_count?: number;
  created_by?: PolicyOwner;
  created_at?: string;
}

export interface PolicyLinkedControl {
  id: string;
  identifier: string;
  title: string;
  category?: string;
  coverage?: string;
}

export interface PolicySignoffSummary {
  total: number;
  approved: number;
  pending: number;
  rejected: number;
}

export interface Policy {
  id: string;
  identifier: string;
  title: string;
  description?: string;
  category: PolicyCategory;
  status: PolicyStatus;
  owner?: PolicyOwner | null;
  secondary_owner?: PolicyOwner | null;
  current_version?: PolicyCurrentVersion | null;
  review_frequency_days?: number | null;
  next_review_at?: string | null;
  last_reviewed_at?: string | null;
  review_status?: PolicyReviewStatus;
  approved_at?: string | null;
  approved_version?: number | null;
  published_at?: string | null;
  is_template?: boolean;
  template_framework?: { id: string; identifier: string; name: string } | null;
  cloned_from_policy_id?: string | null;
  linked_controls?: PolicyLinkedControl[];
  linked_controls_count?: number;
  pending_signoffs_count?: number;
  signoff_summary?: PolicySignoffSummary;
  tags?: string[];
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface PolicyVersion {
  id: string;
  policy_id?: string;
  version_number: number;
  is_current: boolean;
  content?: string;
  content_format?: PolicyContentFormat;
  content_summary?: string;
  change_summary?: string;
  change_type?: PolicyChangeType;
  word_count?: number;
  character_count?: number;
  created_by?: PolicyOwner;
  signoff_summary?: PolicySignoffSummary;
  signoffs?: PolicySignoff[];
  created_at: string;
}

export interface PolicySignoff {
  id: string;
  policy_id?: string;
  policy_version?: { id: string; version_number: number };
  signer: PolicyOwner & { role?: string };
  signer_role?: string;
  requested_by?: PolicyOwner;
  requested_at?: string;
  due_date?: string | null;
  status: SignoffStatus;
  decided_at?: string | null;
  comments?: string | null;
  reminder_count?: number;
  reminder_sent_at?: string | null;
}

export interface PendingSignoff {
  id: string;
  policy: {
    id: string;
    identifier: string;
    title: string;
    category: PolicyCategory;
  };
  policy_version: {
    id: string;
    version_number: number;
    content_summary?: string;
    word_count?: number;
  };
  requested_by: PolicyOwner;
  requested_at: string;
  due_date?: string | null;
  urgency?: 'overdue' | 'due_soon' | 'on_time';
  reminder_count?: number;
}

export interface PolicyControlLink {
  id: string;
  policy_control_id?: string;
  identifier: string;
  title: string;
  description?: string;
  category?: string;
  status?: string;
  coverage?: string;
  notes?: string | null;
  linked_by?: PolicyOwner;
  linked_at?: string;
  frameworks?: string[];
}

export interface PolicyTemplate {
  id: string;
  identifier: string;
  title: string;
  description?: string;
  category: PolicyCategory;
  framework?: { id: string; identifier: string; name: string } | null;
  current_version?: {
    id: string;
    version_number: number;
    word_count?: number;
    content_summary?: string;
  };
  review_frequency_days?: number | null;
  tags?: string[];
}

export interface PolicyGapSummary {
  total_active_controls: number;
  controls_with_full_coverage: number;
  controls_with_partial_coverage: number;
  controls_without_coverage: number;
  coverage_percentage: number;
}

export interface PolicyGapItem {
  control: {
    id: string;
    identifier: string;
    title: string;
    category: string;
    status: string;
    owner?: PolicyOwner | null;
  };
  mapped_frameworks: string[];
  mapped_requirements_count: number;
  policy_coverage: 'none' | 'partial';
  suggested_categories: string[];
}

export interface PolicyGapByFramework {
  framework: {
    id: string;
    identifier: string;
    name: string;
    version: string;
  };
  total_requirements: number;
  requirements_with_controls: number;
  controls_with_policy_coverage: number;
  controls_without_policy_coverage: number;
  policy_coverage_percentage: number;
  gap_count: number;
}

export interface PolicyStats {
  total_policies: number;
  by_status: Record<string, number>;
  by_category: Record<string, number>;
  review_status: {
    overdue: number;
    due_within_30_days: number;
    on_track: number;
    no_schedule: number;
  };
  signoff_summary: {
    total_pending: number;
    overdue_signoffs: number;
  };
  gap_summary: {
    total_active_controls: number;
    controls_with_policy_coverage: number;
    coverage_percentage: number;
  };
  templates_available: number;
  recent_activity: {
    policy_identifier: string;
    action: string;
    actor: string;
    timestamp: string;
  }[];
}

export interface PolicyVersionCompare {
  policy_id: string;
  policy_identifier: string;
  versions: PolicyVersion[];
  word_count_delta: number;
}

export interface PolicySearchResult {
  id: string;
  identifier: string;
  title: string;
  description?: string;
  category: PolicyCategory;
  status: PolicyStatus;
  match_context?: string;
  match_source?: string;
  current_version_number?: number;
  owner?: PolicyOwner;
}

// ========== Sprint 5: Policy Management API Functions ==========

// ---- Policies ----

export function listPolicies(params?: Record<string, string>) {
  return apiGet<Policy[]>('/api/v1/policies', params);
}

export function getPolicy(id: string) {
  return apiGet<Policy>(`/api/v1/policies/${id}`);
}

export function createPolicy(body: {
  identifier: string;
  title: string;
  description?: string;
  category: string;
  owner_id?: string;
  secondary_owner_id?: string;
  review_frequency_days?: number;
  tags?: string[];
  content: string;
  content_format?: string;
  content_summary?: string;
}) {
  return apiPost<Policy>('/api/v1/policies', body);
}

export function updatePolicy(id: string, body: Partial<{
  title: string;
  description: string;
  category: string;
  owner_id: string;
  secondary_owner_id: string | null;
  review_frequency_days: number;
  next_review_at: string;
  tags: string[];
}>) {
  return apiPut<Policy>(`/api/v1/policies/${id}`, body);
}

export function archivePolicy(id: string) {
  return apiPost<{ id: string; identifier: string; status: string }>(`/api/v1/policies/${id}/archive`, {});
}

export function submitPolicyForReview(id: string, body: {
  signer_ids: string[];
  due_date?: string;
  message?: string;
}) {
  return apiPost<Policy & { signoffs_created: number; signoffs: PolicySignoff[] }>(
    `/api/v1/policies/${id}/submit-for-review`,
    body
  );
}

export function publishPolicy(id: string) {
  return apiPost<Policy>(`/api/v1/policies/${id}/publish`, {});
}

// ---- Policy Versions ----

export function listPolicyVersions(policyId: string, params?: Record<string, string>) {
  return apiGet<PolicyVersion[]>(`/api/v1/policies/${policyId}/versions`, params);
}

export function getPolicyVersion(policyId: string, versionNumber: number) {
  return apiGet<PolicyVersion>(`/api/v1/policies/${policyId}/versions/${versionNumber}`);
}

export function createPolicyVersion(policyId: string, body: {
  content: string;
  content_format?: string;
  content_summary?: string;
  change_summary: string;
  change_type?: string;
}) {
  return apiPost<PolicyVersion>(`/api/v1/policies/${policyId}/versions`, body);
}

export function comparePolicyVersions(policyId: string, v1: number, v2: number) {
  return apiGet<PolicyVersionCompare>(
    `/api/v1/policies/${policyId}/versions/compare`,
    { v1: String(v1), v2: String(v2) }
  );
}

// ---- Policy Signoffs ----

export function listPolicySignoffs(policyId: string, params?: Record<string, string>) {
  return apiGet<PolicySignoff[]>(`/api/v1/policies/${policyId}/signoffs`, params);
}

export function approvePolicySignoff(policyId: string, signoffId: string, body?: { comments?: string }) {
  return apiPost<PolicySignoff & { policy_status: string; all_signoffs_complete: boolean }>(
    `/api/v1/policies/${policyId}/signoffs/${signoffId}/approve`,
    body || {}
  );
}

export function rejectPolicySignoff(policyId: string, signoffId: string, body: { comments: string }) {
  return apiPost<PolicySignoff & { policy_status: string }>(
    `/api/v1/policies/${policyId}/signoffs/${signoffId}/reject`,
    body
  );
}

export function withdrawPolicySignoff(policyId: string, signoffId: string) {
  return apiPost<PolicySignoff>(
    `/api/v1/policies/${policyId}/signoffs/${signoffId}/withdraw`,
    {}
  );
}

export function getPendingSignoffs(params?: Record<string, string>) {
  return apiGet<PendingSignoff[]>('/api/v1/signoffs/pending', params);
}

export function remindPolicySignoffs(policyId: string, body?: {
  signoff_ids?: string[];
  message?: string;
}) {
  return apiPost<{ reminders_sent: number; signers: { id: string; name: string; reminder_count: number }[] }>(
    `/api/v1/policies/${policyId}/signoffs/remind`,
    body || {}
  );
}

// ---- Policy Controls ----

export function listPolicyControls(policyId: string) {
  return apiGet<PolicyControlLink[]>(`/api/v1/policies/${policyId}/controls`);
}

export function linkPolicyControl(policyId: string, body: {
  control_id: string;
  coverage?: string;
  notes?: string;
}) {
  return apiPost<PolicyControlLink>(`/api/v1/policies/${policyId}/controls`, body);
}

export function unlinkPolicyControl(policyId: string, controlId: string) {
  return apiDelete<{ message: string }>(`/api/v1/policies/${policyId}/controls/${controlId}`);
}

export function bulkLinkPolicyControls(policyId: string, links: {
  control_id: string;
  coverage?: string;
  notes?: string;
}[]) {
  return apiPost<{ created: number; skipped: number; errors: string[] }>(
    `/api/v1/policies/${policyId}/controls/bulk`,
    { links }
  );
}

// ---- Policy Templates ----

export function listPolicyTemplates(params?: Record<string, string>) {
  return apiGet<PolicyTemplate[]>('/api/v1/policy-templates', params);
}

export function clonePolicyTemplate(templateId: string, body: {
  identifier: string;
  title?: string;
  description?: string;
  owner_id?: string;
  review_frequency_days?: number;
  tags?: string[];
}) {
  return apiPost<Policy>(`/api/v1/policy-templates/${templateId}/clone`, body);
}

// ---- Policy Gap Detection ----

export function getPolicyGap(params?: Record<string, string>) {
  return apiGet<{ summary: PolicyGapSummary; gaps: PolicyGapItem[] }>('/api/v1/policy-gap', params);
}

export function getPolicyGapByFramework() {
  return apiGet<PolicyGapByFramework[]>('/api/v1/policy-gap/by-framework');
}

// ---- Policy Search ----

export function searchPolicies(params: Record<string, string>) {
  return apiGet<PolicySearchResult[]>('/api/v1/policies/search', params);
}

// ---- Policy Stats ----

export function getPolicyStats() {
  return apiGet<PolicyStats>('/api/v1/policies/stats');
}

// ========== Sprint 6: Risk Management Types ==========

export type RiskCategory =
  | 'cyber_security'
  | 'operational'
  | 'compliance'
  | 'data_privacy'
  | 'technology'
  | 'third_party'
  | 'financial'
  | 'legal'
  | 'reputational'
  | 'hr_personnel'
  | 'strategic'
  | 'physical_security'
  | 'environmental';

export type RiskStatus =
  | 'identified'
  | 'open'
  | 'assessing'
  | 'treating'
  | 'monitoring'
  | 'accepted'
  | 'closed'
  | 'archived';

export type LikelihoodLevel = 'rare' | 'unlikely' | 'possible' | 'likely' | 'almost_certain';
export type ImpactLevel = 'negligible' | 'minor' | 'moderate' | 'major' | 'severe';
export type RiskSeverity = 'critical' | 'high' | 'medium' | 'low';

export type TreatmentType = 'mitigate' | 'accept' | 'transfer' | 'avoid';
export type TreatmentStatus = 'planned' | 'in_progress' | 'implemented' | 'verified' | 'ineffective' | 'cancelled';
export type ControlEffectiveness = 'effective' | 'partially_effective' | 'ineffective' | 'not_assessed';

export type AssessmentType = 'inherent' | 'residual' | 'target';
export type AssessmentStatus = 'overdue' | 'due_soon' | 'on_track' | 'no_schedule';

export interface RiskScore {
  likelihood?: LikelihoodLevel;
  likelihood_score?: number;
  impact?: ImpactLevel;
  impact_score?: number;
  score: number;
  severity: RiskSeverity;
}

export interface RiskOwner {
  id: string;
  name: string;
  email?: string;
}

export interface Risk {
  id: string;
  identifier: string;
  title: string;
  description?: string;
  category: RiskCategory;
  status: RiskStatus;
  owner?: RiskOwner | null;
  secondary_owner?: RiskOwner | null;
  inherent_score?: RiskScore | null;
  residual_score?: RiskScore | null;
  risk_appetite_threshold?: number | null;
  appetite_breached?: boolean;
  acceptance?: {
    accepted_at: string;
    accepted_by: RiskOwner;
    expiry: string;
    justification: string;
  } | null;
  assessment_frequency_days?: number | null;
  next_assessment_at?: string | null;
  last_assessed_at?: string | null;
  assessment_status?: AssessmentStatus;
  source?: string | null;
  affected_assets?: string[];
  is_template?: boolean;
  template_source?: string | null;
  linked_controls?: {
    id: string;
    identifier: string;
    title: string;
    effectiveness: ControlEffectiveness;
    mitigation_percentage: number;
  }[];
  linked_controls_count?: number;
  treatment_summary?: {
    total: number;
    planned: number;
    in_progress: number;
    implemented: number;
    verified: number;
    cancelled: number;
  };
  active_treatments_count?: number;
  latest_assessments?: {
    inherent?: {
      id: string;
      assessment_date: string;
      assessor: string;
      justification: string;
      valid_until: string;
    } | null;
    residual?: {
      id: string;
      assessment_date: string;
      assessor: string;
      justification: string;
      valid_until: string;
    } | null;
  };
  tags?: string[];
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface RiskAssessment {
  id: string;
  risk_id: string;
  assessment_type: AssessmentType;
  likelihood: LikelihoodLevel;
  impact: ImpactLevel;
  likelihood_score: number;
  impact_score: number;
  overall_score: number;
  scoring_formula: string;
  severity: RiskSeverity;
  justification?: string;
  assumptions?: string | null;
  data_sources?: string[];
  assessed_by: RiskOwner;
  assessment_date: string;
  valid_until?: string | null;
  is_current: boolean;
  superseded_by?: string | null;
  created_at: string;
}

export interface RiskTreatment {
  id: string;
  risk_id: string;
  treatment_type: TreatmentType;
  title: string;
  description?: string;
  status: TreatmentStatus;
  owner?: RiskOwner | null;
  priority?: string;
  due_date?: string | null;
  started_at?: string | null;
  completed_at?: string | null;
  estimated_effort_hours?: number | null;
  actual_effort_hours?: number | null;
  effectiveness_rating?: string | null;
  effectiveness_notes?: string | null;
  expected_residual?: {
    likelihood: LikelihoodLevel;
    impact: ImpactLevel;
    score: number;
  } | null;
  target_control?: {
    id: string;
    identifier: string;
    title: string;
  } | null;
  notes?: string | null;
  created_by?: RiskOwner;
  created_at: string;
  updated_at: string;
}

export interface RiskControl {
  id: string;
  risk_control_id?: string;
  identifier: string;
  title: string;
  description?: string;
  category?: string;
  status?: string;
  effectiveness: ControlEffectiveness;
  mitigation_percentage: number;
  notes?: string | null;
  last_effectiveness_review?: string | null;
  reviewed_by?: RiskOwner | null;
  linked_by?: RiskOwner | null;
  linked_at?: string;
  latest_test_status?: string | null;
  frameworks?: string[];
}

export interface HeatMapCell {
  likelihood: LikelihoodLevel;
  likelihood_score: number;
  impact: ImpactLevel;
  impact_score: number;
  score: number;
  severity: RiskSeverity;
  count: number;
  risks: {
    id: string;
    identifier: string;
    title: string;
    status: RiskStatus;
  }[];
}

export interface HeatMapData {
  score_type: string;
  grid: HeatMapCell[];
  summary: {
    total_risks: number;
    by_severity: Record<RiskSeverity, number>;
    average_score: number;
    appetite_breaches: number;
  };
}

export interface RiskGapItem {
  risk: {
    id: string;
    identifier: string;
    title: string;
    category: RiskCategory;
    status: RiskStatus;
    residual_score: number;
    severity: RiskSeverity;
    owner?: RiskOwner | null;
  };
  gap_types: string[];
  days_open: number;
  recommendation: string;
  acceptance_expiry?: string;
  days_until_expiry?: number;
}

export interface RiskGapData {
  summary: {
    total_active_risks: number;
    risks_without_treatments: number;
    risks_without_controls: number;
    high_risks_without_controls: number;
    overdue_assessments: number;
    expired_acceptances: number;
  };
  gaps: RiskGapItem[];
}

export interface RiskStats {
  total_risks: number;
  by_status: Record<string, number>;
  by_category: Record<string, number>;
  by_severity: Record<string, number>;
  scoring_summary: {
    average_inherent_score: number;
    average_residual_score: number;
    average_risk_reduction: number;
    highest_residual?: {
      id: string;
      identifier: string;
      title: string;
      score: number;
      severity: RiskSeverity;
    };
  };
  treatment_summary: {
    total_treatments: number;
    planned: number;
    in_progress: number;
    implemented: number;
    verified: number;
    ineffective: number;
    cancelled: number;
    overdue: number;
  };
  control_coverage: {
    risks_with_controls: number;
    risks_without_controls: number;
    average_controls_per_risk: number;
  };
  assessment_health: {
    overdue_assessments: number;
    due_within_30_days: number;
    expired_acceptances: number;
  };
  appetite_summary: {
    within_appetite: number;
    breaching_appetite: number;
    no_threshold_set: number;
  };
  templates_available: number;
  recent_activity: {
    risk_identifier: string;
    action: string;
    actor: string;
    timestamp: string;
  }[];
}

// ========== Sprint 6: Risk Management API Functions ==========

// ---- Risks ----

export function listRisks(params?: Record<string, string>) {
  return apiGet<Risk[]>('/api/v1/risks', params);
}

export function getRisk(id: string) {
  return apiGet<Risk>(`/api/v1/risks/${id}`);
}

export function createRisk(body: {
  identifier: string;
  title: string;
  description?: string;
  category: string;
  owner_id?: string;
  secondary_owner_id?: string;
  risk_appetite_threshold?: number;
  assessment_frequency_days?: number;
  source?: string;
  affected_assets?: string[];
  tags?: string[];
  initial_assessment?: {
    inherent_likelihood: string;
    inherent_impact: string;
    residual_likelihood?: string;
    residual_impact?: string;
    justification?: string;
  };
}) {
  return apiPost<Risk>('/api/v1/risks', body);
}

export function updateRisk(id: string, body: Partial<{
  title: string;
  description: string;
  category: string;
  owner_id: string;
  secondary_owner_id: string | null;
  risk_appetite_threshold: number;
  assessment_frequency_days: number;
  next_assessment_at: string;
  source: string;
  affected_assets: string[];
  tags: string[];
}>) {
  return apiPut<Risk>(`/api/v1/risks/${id}`, body);
}

export function archiveRisk(id: string) {
  return apiPost<{ id: string; identifier: string; status: string }>(`/api/v1/risks/${id}/archive`, {});
}

export function changeRiskStatus(id: string, body: {
  status: string;
  justification?: string;
  acceptance_expiry?: string;
}) {
  return apiPut<Risk>(`/api/v1/risks/${id}/status`, body);
}

// ---- Risk Assessments ----

export function listRiskAssessments(riskId: string, params?: Record<string, string>) {
  return apiGet<RiskAssessment[]>(`/api/v1/risks/${riskId}/assessments`, params);
}

export function createRiskAssessment(riskId: string, body: {
  assessment_type: string;
  likelihood: string;
  impact: string;
  scoring_formula?: string;
  justification?: string;
  assumptions?: string;
  data_sources?: string[];
  valid_until?: string;
}) {
  return apiPost<RiskAssessment>(`/api/v1/risks/${riskId}/assessments`, body);
}

export function recalculateRisk(riskId: string) {
  return apiPost<Risk>(`/api/v1/risks/${riskId}/recalculate`, {});
}

// ---- Risk Treatments ----

export function listRiskTreatments(riskId: string, params?: Record<string, string>) {
  return apiGet<RiskTreatment[]>(`/api/v1/risks/${riskId}/treatments`, params);
}

export function createRiskTreatment(riskId: string, body: {
  treatment_type: string;
  title: string;
  description?: string;
  owner_id?: string;
  priority?: string;
  due_date?: string;
  estimated_effort_hours?: number;
  expected_residual_likelihood?: string;
  expected_residual_impact?: string;
  target_control_id?: string;
  notes?: string;
}) {
  return apiPost<RiskTreatment>(`/api/v1/risks/${riskId}/treatments`, body);
}

export function updateRiskTreatment(riskId: string, treatmentId: string, body: Partial<{
  title: string;
  description: string;
  status: string;
  owner_id: string;
  priority: string;
  due_date: string;
  started_at: string;
  estimated_effort_hours: number;
  actual_effort_hours: number;
  notes: string;
}>) {
  return apiPut<RiskTreatment>(`/api/v1/risks/${riskId}/treatments/${treatmentId}`, body);
}

export function completeRiskTreatment(riskId: string, treatmentId: string, body?: {
  actual_effort_hours?: number;
  effectiveness_rating?: string;
  effectiveness_notes?: string;
}) {
  return apiPost<RiskTreatment>(`/api/v1/risks/${riskId}/treatments/${treatmentId}/complete`, body || {});
}

// ---- Risk Controls ----

export function listRiskControls(riskId: string) {
  return apiGet<RiskControl[]>(`/api/v1/risks/${riskId}/controls`);
}

export function linkRiskControl(riskId: string, body: {
  control_id: string;
  effectiveness?: string;
  mitigation_percentage?: number;
  notes?: string;
}) {
  return apiPost<RiskControl>(`/api/v1/risks/${riskId}/controls`, body);
}

export function updateRiskControlEffectiveness(riskId: string, controlId: string, body: {
  effectiveness?: string;
  mitigation_percentage?: number;
  notes?: string;
}) {
  return apiPut<RiskControl>(`/api/v1/risks/${riskId}/controls/${controlId}`, body);
}

export function unlinkRiskControl(riskId: string, controlId: string) {
  return apiDelete<{ message: string }>(`/api/v1/risks/${riskId}/controls/${controlId}`);
}

// ---- Risk Heat Map ----

export function getRiskHeatMap(params?: Record<string, string>) {
  return apiGet<HeatMapData>('/api/v1/risks/heat-map', params);
}

// ---- Risk Gaps ----

export function getRiskGaps(params?: Record<string, string>) {
  return apiGet<RiskGapData>('/api/v1/risks/gaps', params);
}

// ---- Risk Search ----

export function searchRisks(params: Record<string, string>) {
  return apiGet<Risk[]>('/api/v1/risks/search', params);
}

// ---- Risk Stats ----

export function getRiskStats() {
  return apiGet<RiskStats>('/api/v1/risks/stats');
}

// ========== Sprint 7: Audit Hub Types ==========

export type AuditStatus = 'planning' | 'fieldwork' | 'review' | 'draft_report' | 'management_response' | 'final_report' | 'completed' | 'cancelled';
export type AuditType = 'soc2_type1' | 'soc2_type2' | 'iso27001_certification' | 'iso27001_surveillance' | 'pci_dss_roc' | 'pci_dss_saq' | 'gdpr_dpia' | 'hipaa_assessment' | 'nist_assessment' | 'internal_audit' | 'vendor_assessment' | 'custom';
export type RequestStatus = 'open' | 'in_progress' | 'submitted' | 'accepted' | 'rejected' | 'closed';
export type RequestPriority = 'critical' | 'high' | 'medium' | 'low';
export type FindingSeverity = 'critical' | 'high' | 'medium' | 'low' | 'informational';
export type FindingStatus = 'identified' | 'acknowledged' | 'remediation_planned' | 'remediation_in_progress' | 'remediation_complete' | 'verified' | 'risk_accepted' | 'closed';
export type EvidenceSubmissionStatus = 'pending_review' | 'accepted' | 'rejected' | 'needs_clarification';
export type CommentTargetType = 'audit' | 'request' | 'finding';
export type CommentVisibility = 'external' | 'internal';

export interface AuditMilestone {
  name: string;
  target_date: string;
  completed_at: string | null;
}

export interface Audit {
  id: string;
  title: string;
  description: string;
  audit_type: AuditType;
  status: AuditStatus;
  org_framework_id: string | null;
  framework_name: string | null;
  period_start: string | null;
  period_end: string | null;
  planned_start: string | null;
  planned_end: string | null;
  actual_start: string | null;
  actual_end: string | null;
  audit_firm: string | null;
  lead_auditor_id: string | null;
  lead_auditor_name: string | null;
  internal_lead_id: string | null;
  internal_lead_name: string | null;
  auditor_ids: string[];
  milestones: AuditMilestone[];
  report_type: string | null;
  report_url: string | null;
  report_issued_at: string | null;
  total_requests: number;
  open_requests: number;
  total_findings: number;
  open_findings: number;
  tags: string[];
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface AuditRequest {
  id: string;
  audit_id: string;
  title: string;
  description: string;
  priority: RequestPriority;
  status: RequestStatus;
  control_id: string | null;
  control_title: string | null;
  requirement_id: string | null;
  requirement_title: string | null;
  requested_by: string | null;
  requested_by_name: string | null;
  assigned_to: string | null;
  assigned_to_name: string | null;
  due_date: string | null;
  submitted_at: string | null;
  reviewed_at: string | null;
  reviewer_notes: string | null;
  reference_number: string | null;
  evidence_count: number;
  evidence?: AuditEvidenceLink[];
  tags: string[];
  created_at: string;
  updated_at: string;
}

export interface AuditEvidenceLink {
  link_id: string;
  artifact_id: string;
  artifact_title: string;
  file_name?: string;
  file_size?: number;
  mime_type?: string;
  evidence_type?: string;
  evidence_status?: string;
  submitted_by: string;
  submitted_by_name: string;
  submitted_at: string;
  submission_notes: string | null;
  status: EvidenceSubmissionStatus;
  reviewed_by: string | null;
  reviewed_by_name: string | null;
  reviewed_at: string | null;
  review_notes: string | null;
}

export interface AuditFinding {
  id: string;
  audit_id: string;
  title: string;
  description: string;
  severity: FindingSeverity;
  category: string;
  status: FindingStatus;
  control_id: string | null;
  control_title: string | null;
  requirement_id: string | null;
  requirement_title: string | null;
  found_by: string | null;
  found_by_name: string | null;
  remediation_owner_id: string | null;
  remediation_owner_name: string | null;
  remediation_plan: string | null;
  remediation_due_date: string | null;
  remediation_started_at: string | null;
  remediation_completed_at: string | null;
  verified_at: string | null;
  verified_by: string | null;
  verified_by_name: string | null;
  verification_notes: string | null;
  reference_number: string | null;
  recommendation: string | null;
  management_response: string | null;
  risk_accepted: boolean;
  risk_acceptance_reason: string | null;
  comment_count: number;
  tags: string[];
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface AuditComment {
  id: string;
  audit_id: string;
  target_type: CommentTargetType;
  target_id: string;
  author_id: string;
  author_name: string;
  author_role: string;
  body: string;
  parent_comment_id: string | null;
  is_internal: boolean;
  edited_at: string | null;
  replies?: AuditComment[];
  created_at: string;
  updated_at: string;
}

export interface AuditRequestTemplate {
  id: string;
  title: string;
  description: string;
  audit_type: string;
  framework: string;
  category: string;
  default_priority: RequestPriority;
  tags: string[];
}

export interface AuditDashboard {
  summary: {
    active_audits: number;
    completed_audits: number;
    total_open_requests: number;
    total_overdue_requests: number;
    total_open_findings: number;
    critical_findings: number;
    high_findings: number;
  };
  active_audits: {
    id: string;
    title: string;
    audit_type: AuditType;
    status: AuditStatus;
    planned_end: string | null;
    days_remaining: number | null;
    readiness_pct: number;
    total_requests: number;
    open_requests: number;
    total_findings: number;
    open_findings: number;
    next_milestone: { name: string; target_date: string; days_until: number } | null;
  }[];
  overdue_requests: {
    id: string;
    title: string;
    audit_title: string;
    due_date: string;
    days_overdue: number;
    assigned_to_name: string | null;
    priority: RequestPriority;
  }[];
  critical_findings: {
    id: string;
    title: string;
    audit_title: string;
    severity: FindingSeverity;
    status: FindingStatus;
    remediation_due_date: string | null;
    remediation_owner_name: string | null;
  }[];
  recent_activity: {
    type: string;
    title: string;
    audit_title: string;
    actor_name: string;
    timestamp: string;
    old_status?: string;
    new_status?: string;
  }[];
}

export interface AuditStats {
  audit_id: string;
  title: string;
  status: AuditStatus;
  readiness: {
    total_requests: number;
    accepted: number;
    submitted: number;
    in_progress: number;
    open: number;
    rejected: number;
    overdue: number;
    readiness_pct: number;
  };
  findings: {
    total: number;
    by_severity: Record<string, number>;
    by_status: Record<string, number>;
    overdue_remediation: number;
  };
  evidence: {
    total_submitted: number;
    accepted: number;
    pending_review: number;
    rejected: number;
  };
  timeline: {
    planned_start: string | null;
    planned_end: string | null;
    actual_start: string | null;
    days_elapsed: number;
    days_remaining: number | null;
    milestones_completed: number;
    milestones_total: number;
    next_milestone: { name: string; target_date: string; days_until: number } | null;
  };
  activity: {
    comments_count: number;
    last_activity_at: string | null;
  };
}

export interface AuditReadiness {
  audit_id: string;
  overall_readiness_pct: number;
  by_requirement: {
    requirement_id: string;
    requirement_title: string;
    total_requests: number;
    accepted_requests: number;
    readiness_pct: number;
  }[];
  by_control: {
    control_id: string;
    control_title: string;
    total_requests: number;
    accepted_requests: number;
    readiness_pct: number;
  }[];
  gaps: {
    requirement_id: string;
    requirement_title: string;
    issue: string;
    description: string;
  }[];
}

// ========== Sprint 7: Audit Hub API Functions ==========

// ---- Audits ----

export function listAudits(params?: Record<string, string>) {
  return apiGet<Audit[]>('/api/v1/audits', params);
}

export function getAudit(id: string) {
  return apiGet<Audit>(`/api/v1/audits/${id}`);
}

export function createAudit(body: {
  title: string;
  description?: string;
  audit_type: string;
  org_framework_id?: string;
  period_start?: string;
  period_end?: string;
  planned_start?: string;
  planned_end?: string;
  audit_firm?: string;
  lead_auditor_id?: string;
  auditor_ids?: string[];
  internal_lead_id?: string;
  milestones?: { name: string; target_date: string }[];
  report_type?: string;
  tags?: string[];
}) {
  return apiPost<Audit>('/api/v1/audits', body);
}

export function updateAudit(id: string, body: Partial<{
  title: string;
  description: string;
  org_framework_id: string;
  period_start: string;
  period_end: string;
  planned_start: string;
  planned_end: string;
  audit_firm: string;
  lead_auditor_id: string;
  internal_lead_id: string;
  milestones: { name: string; target_date: string }[];
  report_type: string;
  tags: string[];
}>) {
  return apiPut<Audit>(`/api/v1/audits/${id}`, body);
}

export function transitionAuditStatus(id: string, body: { status: string; notes?: string }) {
  return apiPut<Audit>(`/api/v1/audits/${id}/status`, body);
}

export function addAuditor(auditId: string, userId: string) {
  return apiPost<Audit>(`/api/v1/audits/${auditId}/auditors`, { user_id: userId });
}

export function removeAuditor(auditId: string, userId: string) {
  return apiDelete<Audit>(`/api/v1/audits/${auditId}/auditors/${userId}`);
}

// ---- Audit Requests ----

export function listAuditRequests(auditId: string, params?: Record<string, string>) {
  return apiGet<AuditRequest[]>(`/api/v1/audits/${auditId}/requests`, params);
}

export function getAuditRequest(auditId: string, requestId: string) {
  return apiGet<AuditRequest>(`/api/v1/audits/${auditId}/requests/${requestId}`);
}

export function createAuditRequest(auditId: string, body: {
  title: string;
  description: string;
  priority?: string;
  control_id?: string;
  requirement_id?: string;
  assigned_to?: string;
  due_date?: string;
  reference_number?: string;
  tags?: string[];
}) {
  return apiPost<AuditRequest>(`/api/v1/audits/${auditId}/requests`, body);
}

export function updateAuditRequest(auditId: string, requestId: string, body: Partial<{
  title: string;
  description: string;
  priority: string;
  due_date: string;
  reference_number: string;
  tags: string[];
}>) {
  return apiPut<AuditRequest>(`/api/v1/audits/${auditId}/requests/${requestId}`, body);
}

export function assignAuditRequest(auditId: string, requestId: string, assignedTo: string) {
  return apiPut<AuditRequest>(`/api/v1/audits/${auditId}/requests/${requestId}/assign`, { assigned_to: assignedTo });
}

export function submitAuditRequest(auditId: string, requestId: string, notes?: string) {
  return apiPut<AuditRequest>(`/api/v1/audits/${auditId}/requests/${requestId}/submit`, { notes });
}

export function reviewAuditRequest(auditId: string, requestId: string, body: { decision: string; notes?: string }) {
  return apiPut<AuditRequest>(`/api/v1/audits/${auditId}/requests/${requestId}/review`, body);
}

export function closeAuditRequest(auditId: string, requestId: string, reason?: string) {
  return apiPut<AuditRequest>(`/api/v1/audits/${auditId}/requests/${requestId}/close`, { reason });
}

export function bulkCreateAuditRequests(auditId: string, body: {
  requests: {
    title: string;
    description: string;
    priority?: string;
    due_date?: string;
    reference_number?: string;
  }[];
}) {
  return apiPost<{ created: number; requests: AuditRequest[] }>(`/api/v1/audits/${auditId}/requests/bulk`, body);
}

export function createRequestsFromTemplate(auditId: string, body: {
  template_ids: string[];
  default_due_date?: string;
  auto_number?: boolean;
  number_prefix?: string;
}) {
  return apiPost<{ created: number; requests: AuditRequest[] }>(`/api/v1/audits/${auditId}/requests/from-template`, body);
}

// ---- Evidence Submission ----

export function listRequestEvidence(auditId: string, requestId: string) {
  return apiGet<AuditEvidenceLink[]>(`/api/v1/audits/${auditId}/requests/${requestId}/evidence`);
}

export function submitRequestEvidence(auditId: string, requestId: string, body: { artifact_id: string; notes?: string }) {
  return apiPost<AuditEvidenceLink>(`/api/v1/audits/${auditId}/requests/${requestId}/evidence`, body);
}

export function reviewRequestEvidence(auditId: string, requestId: string, linkId: string, body: { status: string; notes?: string }) {
  return apiPut<AuditEvidenceLink>(`/api/v1/audits/${auditId}/requests/${requestId}/evidence/${linkId}/review`, body);
}

export function removeRequestEvidence(auditId: string, requestId: string, linkId: string) {
  return apiDelete<{ message: string }>(`/api/v1/audits/${auditId}/requests/${requestId}/evidence/${linkId}`);
}

// ---- Audit Findings ----

export function listAuditFindings(auditId: string, params?: Record<string, string>) {
  return apiGet<AuditFinding[]>(`/api/v1/audits/${auditId}/findings`, params);
}

export function getAuditFinding(auditId: string, findingId: string) {
  return apiGet<AuditFinding>(`/api/v1/audits/${auditId}/findings/${findingId}`);
}

export function createAuditFinding(auditId: string, body: {
  title: string;
  description: string;
  severity: string;
  category: string;
  control_id?: string;
  requirement_id?: string;
  remediation_owner_id?: string;
  remediation_due_date?: string;
  reference_number?: string;
  recommendation?: string;
  tags?: string[];
}) {
  return apiPost<AuditFinding>(`/api/v1/audits/${auditId}/findings`, body);
}

export function updateAuditFinding(auditId: string, findingId: string, body: Partial<{
  title: string;
  description: string;
  severity: string;
  category: string;
  recommendation: string;
  reference_number: string;
  tags: string[];
}>) {
  return apiPut<AuditFinding>(`/api/v1/audits/${auditId}/findings/${findingId}`, body);
}

export function transitionFindingStatus(auditId: string, findingId: string, body: {
  status: string;
  remediation_plan?: string;
  remediation_due_date?: string;
  remediation_owner_id?: string;
  management_response?: string;
  verification_notes?: string;
  risk_acceptance_reason?: string;
  notes?: string;
}) {
  return apiPut<AuditFinding>(`/api/v1/audits/${auditId}/findings/${findingId}/status`, body);
}

export function submitManagementResponse(auditId: string, findingId: string, body: { management_response: string }) {
  return apiPut<AuditFinding>(`/api/v1/audits/${auditId}/findings/${findingId}/management-response`, body);
}

// ---- Audit Comments ----

export function listAuditComments(auditId: string, params?: Record<string, string>) {
  return apiGet<AuditComment[]>(`/api/v1/audits/${auditId}/comments`, params);
}

export function createAuditComment(auditId: string, body: {
  target_type: string;
  target_id: string;
  body: string;
  parent_comment_id?: string;
  is_internal?: boolean;
}) {
  return apiPost<AuditComment>(`/api/v1/audits/${auditId}/comments`, body);
}

export function editAuditComment(auditId: string, commentId: string, body: { body: string }) {
  return apiPut<AuditComment>(`/api/v1/audits/${auditId}/comments/${commentId}`, body);
}

export function deleteAuditComment(auditId: string, commentId: string) {
  return apiDelete<{ message: string }>(`/api/v1/audits/${auditId}/comments/${commentId}`);
}

// ---- PBC Templates ----

export function listAuditRequestTemplates(params?: Record<string, string>) {
  return apiGet<AuditRequestTemplate[]>('/api/v1/audit-request-templates', params);
}

// ---- Audit Dashboard & Analytics ----

export function getAuditDashboard() {
  return apiGet<AuditDashboard>('/api/v1/audits/dashboard');
}

export function getAuditStats(auditId: string) {
  return apiGet<AuditStats>(`/api/v1/audits/${auditId}/stats`);
}

export function getAuditReadiness(auditId: string) {
  return apiGet<AuditReadiness>(`/api/v1/audits/${auditId}/readiness`);
}
