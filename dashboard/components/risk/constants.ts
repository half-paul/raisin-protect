/**
 * Risk Management constants — Sprint 6
 * Labels, colors, and enums for risk UI components
 */

import type {
  RiskCategory,
  RiskStatus,
  RiskSeverity,
  LikelihoodLevel,
  ImpactLevel,
  TreatmentType,
  TreatmentStatus,
  ControlEffectiveness,
  AssessmentType,
  AssessmentStatus,
} from '@/lib/api';

// ---- Risk Status ----

export const RISK_STATUS_LABELS: Record<RiskStatus, string> = {
  identified: 'Identified',
  open: 'Open',
  assessing: 'Assessing',
  treating: 'Treating',
  monitoring: 'Monitoring',
  accepted: 'Accepted',
  closed: 'Closed',
  archived: 'Archived',
};

export const RISK_STATUS_COLORS: Record<RiskStatus, string> = {
  identified: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  open: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  assessing: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  treating: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
  monitoring: 'bg-cyan-100 text-cyan-800 dark:bg-cyan-900/30 dark:text-cyan-400',
  accepted: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
  closed: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  archived: 'bg-gray-100 text-gray-500 dark:bg-gray-900/30 dark:text-gray-500',
};

// ---- Risk Category ----

export const RISK_CATEGORY_LABELS: Record<RiskCategory, string> = {
  cyber_security: 'Cyber Security',
  operational: 'Operational',
  compliance: 'Compliance',
  data_privacy: 'Data Privacy',
  technology: 'Technology',
  third_party: 'Third Party',
  financial: 'Financial',
  legal: 'Legal',
  reputational: 'Reputational',
  hr_personnel: 'HR / Personnel',
  strategic: 'Strategic',
  physical_security: 'Physical Security',
  environmental: 'Environmental',
};

// ---- Severity ----

export const SEVERITY_LABELS: Record<RiskSeverity, string> = {
  critical: 'Critical',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
};

export const SEVERITY_COLORS: Record<RiskSeverity, string> = {
  critical: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  high: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
  medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  low: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
};

export const SEVERITY_BG_COLORS: Record<RiskSeverity, string> = {
  critical: 'bg-red-500',
  high: 'bg-orange-500',
  medium: 'bg-yellow-500',
  low: 'bg-green-500',
};

// ---- Likelihood ----

export const LIKELIHOOD_LABELS: Record<LikelihoodLevel, string> = {
  rare: 'Rare',
  unlikely: 'Unlikely',
  possible: 'Possible',
  likely: 'Likely',
  almost_certain: 'Almost Certain',
};

export const LIKELIHOOD_SCORES: Record<LikelihoodLevel, number> = {
  rare: 1,
  unlikely: 2,
  possible: 3,
  likely: 4,
  almost_certain: 5,
};

export const LIKELIHOOD_ORDER: LikelihoodLevel[] = [
  'rare', 'unlikely', 'possible', 'likely', 'almost_certain',
];

// ---- Impact ----

export const IMPACT_LABELS: Record<ImpactLevel, string> = {
  negligible: 'Negligible',
  minor: 'Minor',
  moderate: 'Moderate',
  major: 'Major',
  severe: 'Severe',
};

export const IMPACT_SCORES: Record<ImpactLevel, number> = {
  negligible: 1,
  minor: 2,
  moderate: 3,
  major: 4,
  severe: 5,
};

export const IMPACT_ORDER: ImpactLevel[] = [
  'negligible', 'minor', 'moderate', 'major', 'severe',
];

// ---- Treatment ----

export const TREATMENT_TYPE_LABELS: Record<TreatmentType, string> = {
  mitigate: 'Mitigate',
  accept: 'Accept',
  transfer: 'Transfer',
  avoid: 'Avoid',
};

export const TREATMENT_STATUS_LABELS: Record<TreatmentStatus, string> = {
  planned: 'Planned',
  in_progress: 'In Progress',
  implemented: 'Implemented',
  verified: 'Verified',
  ineffective: 'Ineffective',
  cancelled: 'Cancelled',
};

export const TREATMENT_STATUS_COLORS: Record<TreatmentStatus, string> = {
  planned: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  in_progress: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  implemented: 'bg-cyan-100 text-cyan-800 dark:bg-cyan-900/30 dark:text-cyan-400',
  verified: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  ineffective: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  cancelled: 'bg-gray-100 text-gray-500 dark:bg-gray-900/30 dark:text-gray-500',
};

// ---- Control Effectiveness ----

export const EFFECTIVENESS_LABELS: Record<ControlEffectiveness, string> = {
  effective: 'Effective',
  partially_effective: 'Partially Effective',
  ineffective: 'Ineffective',
  not_assessed: 'Not Assessed',
};

export const EFFECTIVENESS_COLORS: Record<ControlEffectiveness, string> = {
  effective: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  partially_effective: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  ineffective: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  not_assessed: 'bg-gray-100 text-gray-500 dark:bg-gray-900/30 dark:text-gray-500',
};

// ---- Assessment ----

export const ASSESSMENT_TYPE_LABELS: Record<AssessmentType, string> = {
  inherent: 'Inherent',
  residual: 'Residual',
  target: 'Target',
};

export const ASSESSMENT_STATUS_LABELS: Record<AssessmentStatus, string> = {
  overdue: 'Overdue',
  due_soon: 'Due Soon',
  on_track: 'On Track',
  no_schedule: 'No Schedule',
};

export const ASSESSMENT_STATUS_COLORS: Record<AssessmentStatus, string> = {
  overdue: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  due_soon: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  on_track: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  no_schedule: 'bg-gray-100 text-gray-500 dark:bg-gray-900/30 dark:text-gray-500',
};

// ---- Helper: Score to severity ----

export function scoreToSeverity(score: number): RiskSeverity {
  if (score >= 20) return 'critical';
  if (score >= 12) return 'high';
  if (score >= 6) return 'medium';
  return 'low';
}

export function severityToColor(severity: RiskSeverity): string {
  switch (severity) {
    case 'critical': return '#ef4444';
    case 'high': return '#f97316';
    case 'medium': return '#eab308';
    case 'low': return '#22c55e';
  }
}

// ---- Heat Map Cell Color ----

export function heatMapCellColor(score: number): string {
  if (score >= 20) return 'bg-red-500/80 dark:bg-red-600/80';
  if (score >= 15) return 'bg-red-400/70 dark:bg-red-500/70';
  if (score >= 12) return 'bg-orange-400/70 dark:bg-orange-500/70';
  if (score >= 10) return 'bg-orange-300/60 dark:bg-orange-400/60';
  if (score >= 6) return 'bg-yellow-300/60 dark:bg-yellow-400/60';
  if (score >= 4) return 'bg-yellow-200/50 dark:bg-yellow-300/50';
  if (score >= 2) return 'bg-green-200/50 dark:bg-green-300/50';
  return 'bg-green-100/40 dark:bg-green-200/40';
}

// ---- Priority Labels ----

export const PRIORITY_LABELS: Record<string, string> = {
  critical: 'Critical',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
};

export const PRIORITY_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  high: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
  medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  low: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
};
