-- Migration: 035_sprint6_enums.sql
-- Description: Enum types for Sprint 6 (Risk Register)
-- Created: 2026-02-21
-- Sprint: 6 — Risk Register

-- ============================================================================
-- NEW ENUM TYPES
-- ============================================================================

-- Risk taxonomy categories (from spec §4.1.1 — customizable risk taxonomy)
DO $$ BEGIN
    CREATE TYPE risk_category AS ENUM (
        'operational',
        'financial',
        'strategic',
        'compliance',
        'technology',
        'legal',
        'reputational',
        'third_party',
        'physical',
        'data_privacy',
        'cyber_security',
        'human_resources',
        'environmental',
        'custom'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE risk_category IS 'Risk taxonomy categories: operational, financial, strategic, compliance, technology, legal, reputational, third_party, physical, data_privacy, cyber_security, human_resources, environmental, custom';

-- Risk lifecycle status
DO $$ BEGIN
    CREATE TYPE risk_status AS ENUM (
        'identified',
        'open',
        'assessing',
        'treating',
        'monitoring',
        'accepted',
        'closed',
        'archived'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE risk_status IS 'Risk lifecycle: identified → open → assessing → treating → monitoring → accepted → closed → archived';

-- 5-level likelihood scale (maps to 1–5 for scoring)
DO $$ BEGIN
    CREATE TYPE likelihood_level AS ENUM (
        'rare',
        'unlikely',
        'possible',
        'likely',
        'almost_certain'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE likelihood_level IS '5-level likelihood scale: rare(1), unlikely(2), possible(3), likely(4), almost_certain(5)';

-- 5-level impact scale (maps to 1–5 for scoring)
DO $$ BEGIN
    CREATE TYPE impact_level AS ENUM (
        'negligible',
        'minor',
        'moderate',
        'major',
        'severe'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE impact_level IS '5-level impact scale: negligible(1), minor(2), moderate(3), major(4), severe(5)';

-- Treatment strategy options (from spec §4.1.3)
DO $$ BEGIN
    CREATE TYPE treatment_type AS ENUM (
        'mitigate',
        'accept',
        'transfer',
        'avoid'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE treatment_type IS 'Treatment strategy: mitigate, accept, transfer, avoid (per spec §4.1.3)';

-- Treatment plan lifecycle
DO $$ BEGIN
    CREATE TYPE treatment_status AS ENUM (
        'planned',
        'in_progress',
        'implemented',
        'verified',
        'ineffective',
        'cancelled'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE treatment_status IS 'Treatment lifecycle: planned → in_progress → implemented → verified | ineffective | cancelled';

-- Assessment scope type
DO $$ BEGIN
    CREATE TYPE risk_assessment_type AS ENUM (
        'inherent',
        'residual',
        'target'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE risk_assessment_type IS 'Assessment scope: inherent (before controls), residual (after), target (desired)';

-- Control effectiveness against a specific risk
DO $$ BEGIN
    CREATE TYPE control_effectiveness AS ENUM (
        'effective',
        'partially_effective',
        'ineffective',
        'not_assessed'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE control_effectiveness IS 'How effective a control is at mitigating a specific risk';

-- ============================================================================
-- EXTEND EXISTING ENUMS
-- ============================================================================

-- Add Sprint 6 actions to audit_action enum
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk.created'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk.updated'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk.status_changed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk.archived'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk.deleted'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk.owner_changed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk.score_recalculated'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_assessment.created'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_assessment.updated'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_assessment.deleted'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_treatment.created'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_treatment.updated'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_treatment.status_changed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_treatment.completed'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_treatment.cancelled'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_control.linked'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_control.unlinked'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN ALTER TYPE audit_action ADD VALUE IF NOT EXISTS 'risk_control.effectiveness_updated'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- Extend evidence_link_target_type to support risks
DO $$ BEGIN ALTER TYPE evidence_link_target_type ADD VALUE IF NOT EXISTS 'risk'; EXCEPTION WHEN duplicate_object THEN NULL; END $$;
