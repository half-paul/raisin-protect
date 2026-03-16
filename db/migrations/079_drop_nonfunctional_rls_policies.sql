-- Migration 079: Drop non-functional RLS policies from migrations 074–077
--
-- Background (LEX-98):
-- The RLS policies added in 074–077 are completely non-functional for two
-- independent reasons:
--
--   1. Owner bypass: the app connects as the table owner ('rp'), which skips
--      RLS entirely unless FORCE ROW LEVEL SECURITY is set on the table.
--
--   2. Unset session variable: the app never executes
--      SET app.current_org_id = '<uuid>' before queries, so every
--      current_setting('app.current_org_id', TRUE) call returns NULL, causing
--      the USING clause to evaluate to FALSE for all rows.
--
-- Together these mean the policies neither enforce isolation (owner bypass) nor
-- filter correctly when they would be applied (NULL comparison).
--
-- Fix: Remove the policies and disable RLS on all affected tables.
-- Tenant isolation is enforced at the application layer (org_id filter on
-- every query + JWT-scoped middleware). A proper DB-level RLS implementation
-- requires a dedicated low-privilege role and explicit SET LOCAL calls inside
-- transactions — tracked as a follow-up task.
--
-- Tables affected (from migration 074):
--   cde_assets, cde_network_segments, cde_data_flows,
--   cde_segmentation_tests, cde_scope_history
--
-- Tables affected (from migration 075):
--   document_templates, compliance_documents, document_sections,
--   document_attestations, document_approvals
--
-- Tables affected (from migration 076):
--   service_providers, sp_compliance_documents, sp_responsibility_matrix
--
-- Tables affected (from migration 077):
--   asv_scans

-- ── 074: CDE scoping ─────────────────────────────────────────────────────────

DROP POLICY IF EXISTS cde_assets_org_isolation          ON cde_assets;
DROP POLICY IF EXISTS cde_network_segments_org_isolation ON cde_network_segments;
DROP POLICY IF EXISTS cde_data_flows_org_isolation       ON cde_data_flows;
DROP POLICY IF EXISTS cde_segmentation_tests_org_isolation ON cde_segmentation_tests;
DROP POLICY IF EXISTS cde_scope_history_org_isolation    ON cde_scope_history;

ALTER TABLE cde_assets             DISABLE ROW LEVEL SECURITY;
ALTER TABLE cde_network_segments   DISABLE ROW LEVEL SECURITY;
ALTER TABLE cde_data_flows         DISABLE ROW LEVEL SECURITY;
ALTER TABLE cde_segmentation_tests DISABLE ROW LEVEL SECURITY;
ALTER TABLE cde_scope_history      DISABLE ROW LEVEL SECURITY;

-- ── 075: AOC/ROC documents ───────────────────────────────────────────────────

DROP POLICY IF EXISTS doc_templates_visibility           ON document_templates;
DROP POLICY IF EXISTS compliance_documents_org_isolation ON compliance_documents;
DROP POLICY IF EXISTS document_sections_org_isolation    ON document_sections;
DROP POLICY IF EXISTS document_attestations_org_isolation ON document_attestations;
DROP POLICY IF EXISTS document_approvals_org_isolation   ON document_approvals;

ALTER TABLE document_templates    DISABLE ROW LEVEL SECURITY;
ALTER TABLE compliance_documents  DISABLE ROW LEVEL SECURITY;
ALTER TABLE document_sections     DISABLE ROW LEVEL SECURITY;
ALTER TABLE document_attestations DISABLE ROW LEVEL SECURITY;
ALTER TABLE document_approvals    DISABLE ROW LEVEL SECURITY;

-- ── 076: Service providers ───────────────────────────────────────────────────

DROP POLICY IF EXISTS service_providers_org_isolation         ON service_providers;
DROP POLICY IF EXISTS sp_compliance_documents_org_isolation   ON sp_compliance_documents;
DROP POLICY IF EXISTS sp_responsibility_matrix_org_isolation  ON sp_responsibility_matrix;

ALTER TABLE service_providers          DISABLE ROW LEVEL SECURITY;
ALTER TABLE sp_compliance_documents    DISABLE ROW LEVEL SECURITY;
ALTER TABLE sp_responsibility_matrix   DISABLE ROW LEVEL SECURITY;

-- ── 077: ASV scans ───────────────────────────────────────────────────────────

DROP POLICY IF EXISTS asv_scans_org_isolation ON asv_scans;

ALTER TABLE asv_scans DISABLE ROW LEVEL SECURITY;
