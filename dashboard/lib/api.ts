/**
 * API client helpers for Raisin Protect Dashboard
 * Sprint 2: Frameworks & Controls
 * Sprint 3: Evidence Management
 * Sprint 4: Continuous Monitoring Engine
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
