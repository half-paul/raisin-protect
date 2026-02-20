/**
 * API client helpers for Sprint 2: Frameworks & Controls
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
