-- Migration: 078_requirement_scope_customized_approach_description.sql
-- Description: Add customized_approach_description text column to requirement_scopes
-- Created: 2026-03-15
-- Sprint: 13 — PCI Regression Fixes (LEX-100 / TC4)
--
-- Context:
--   Migration 053 added the customized_approach BOOLEAN flag but omitted the
--   free-text description field required for PCI DSS v4.0 Customized Approach
--   documentation. This migration adds that missing column.

ALTER TABLE requirement_scopes
    ADD COLUMN IF NOT EXISTS customized_approach_description TEXT;

COMMENT ON COLUMN requirement_scopes.customized_approach_description IS
    'Narrative description of the customized approach implemented for this requirement '
    'per PCI DSS v4.0. Required when customized_approach = TRUE. Documents how the '
    'organization achieves the stated objective through alternative means.';
