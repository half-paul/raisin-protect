-- Migration: 036_sprint6_functions.sql
-- Description: Helper functions for risk scoring (Sprint 6 — Risk Register)
-- Created: 2026-02-21
-- Sprint: 6 — Risk Register

-- ============================================================================
-- SCORING HELPER FUNCTIONS
-- ============================================================================

-- Map likelihood_level enum to numeric score (1–5)
CREATE OR REPLACE FUNCTION likelihood_to_score(level likelihood_level)
RETURNS INT AS $$
BEGIN
    RETURN CASE level
        WHEN 'rare' THEN 1
        WHEN 'unlikely' THEN 2
        WHEN 'possible' THEN 3
        WHEN 'likely' THEN 4
        WHEN 'almost_certain' THEN 5
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION likelihood_to_score(likelihood_level) IS 'Maps likelihood enum to 1–5 numeric score';

-- Map impact_level enum to numeric score (1–5)
CREATE OR REPLACE FUNCTION impact_to_score(level impact_level)
RETURNS INT AS $$
BEGIN
    RETURN CASE level
        WHEN 'negligible' THEN 1
        WHEN 'minor' THEN 2
        WHEN 'moderate' THEN 3
        WHEN 'major' THEN 4
        WHEN 'severe' THEN 5
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION impact_to_score(impact_level) IS 'Maps impact enum to 1–5 numeric score';

-- Classify a numeric risk score (1–25) into a severity band
CREATE OR REPLACE FUNCTION risk_score_severity(score NUMERIC)
RETURNS TEXT AS $$
BEGIN
    RETURN CASE
        WHEN score >= 20 THEN 'critical'   -- 20–25: red zone
        WHEN score >= 12 THEN 'high'       -- 12–19: orange zone
        WHEN score >= 6  THEN 'medium'     -- 6–11: yellow zone
        WHEN score >= 1  THEN 'low'        -- 1–5: green zone
        ELSE 'none'
    END;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON FUNCTION risk_score_severity(NUMERIC) IS 'Classifies numeric risk score (1–25) into severity band: critical/high/medium/low/none';
