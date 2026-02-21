/**
 * Audit Hub constants — Sprint 7
 * Labels, colors, and enums for audit engagement UI
 */

// ========== Audit Status ==========

export const AUDIT_STATUS_LABELS: Record<string, string> = {
  planning: 'Planning',
  fieldwork: 'Fieldwork',
  review: 'Review',
  draft_report: 'Draft Report',
  management_response: 'Mgmt Response',
  final_report: 'Final Report',
  completed: 'Completed',
  cancelled: 'Cancelled',
};

export const AUDIT_STATUS_COLORS: Record<string, string> = {
  planning: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300',
  fieldwork: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-300',
  review: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300',
  draft_report: 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300',
  management_response: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300',
  final_report: 'bg-teal-100 text-teal-800 dark:bg-teal-900/30 dark:text-teal-300',
  completed: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
  cancelled: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-300',
};

// ========== Audit Type ==========

export const AUDIT_TYPE_LABELS: Record<string, string> = {
  soc2_type1: 'SOC 2 Type I',
  soc2_type2: 'SOC 2 Type II',
  iso27001_certification: 'ISO 27001 Certification',
  iso27001_surveillance: 'ISO 27001 Surveillance',
  pci_dss_roc: 'PCI DSS ROC',
  pci_dss_saq: 'PCI DSS SAQ',
  gdpr_dpia: 'GDPR DPIA',
  hipaa_assessment: 'HIPAA Assessment',
  nist_assessment: 'NIST Assessment',
  internal_audit: 'Internal Audit',
  vendor_assessment: 'Vendor Assessment',
  custom: 'Custom',
};

// ========== Request Status ==========

export const REQUEST_STATUS_LABELS: Record<string, string> = {
  open: 'Open',
  in_progress: 'In Progress',
  submitted: 'Submitted',
  accepted: 'Accepted',
  rejected: 'Rejected',
  closed: 'Closed',
};

export const REQUEST_STATUS_COLORS: Record<string, string> = {
  open: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-300',
  in_progress: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300',
  submitted: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300',
  accepted: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
  rejected: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
  closed: 'bg-gray-100 text-gray-600 dark:bg-gray-900/30 dark:text-gray-400',
};

// ========== Request Priority ==========

export const REQUEST_PRIORITY_LABELS: Record<string, string> = {
  critical: 'Critical',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
};

export const REQUEST_PRIORITY_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
  high: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300',
  medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
  low: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
};

// ========== Finding Severity ==========

export const FINDING_SEVERITY_LABELS: Record<string, string> = {
  critical: 'Critical',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
  informational: 'Informational',
};

export const FINDING_SEVERITY_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
  high: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300',
  medium: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
  low: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300',
  informational: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-300',
};

// ========== Finding Status ==========

export const FINDING_STATUS_LABELS: Record<string, string> = {
  identified: 'Identified',
  acknowledged: 'Acknowledged',
  remediation_planned: 'Remediation Planned',
  remediation_in_progress: 'Remediation In Progress',
  remediation_complete: 'Remediation Complete',
  verified: 'Verified',
  risk_accepted: 'Risk Accepted',
  closed: 'Closed',
};

export const FINDING_STATUS_COLORS: Record<string, string> = {
  identified: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
  acknowledged: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300',
  remediation_planned: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300',
  remediation_in_progress: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-300',
  remediation_complete: 'bg-teal-100 text-teal-800 dark:bg-teal-900/30 dark:text-teal-300',
  verified: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
  risk_accepted: 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300',
  closed: 'bg-gray-100 text-gray-600 dark:bg-gray-900/30 dark:text-gray-400',
};

// ========== Finding Category ==========

export const FINDING_CATEGORY_LABELS: Record<string, string> = {
  access_control: 'Access Control',
  change_management: 'Change Management',
  data_protection: 'Data Protection',
  incident_management: 'Incident Management',
  monitoring: 'Monitoring',
  network_security: 'Network Security',
  physical_security: 'Physical Security',
  policy_governance: 'Policy & Governance',
  risk_management: 'Risk Management',
  vendor_management: 'Vendor Management',
  other: 'Other',
};

// ========== Evidence Submission Status ==========

export const EVIDENCE_SUBMISSION_STATUS_LABELS: Record<string, string> = {
  pending_review: 'Pending Review',
  accepted: 'Accepted',
  rejected: 'Rejected',
  needs_clarification: 'Needs Clarification',
};

export const EVIDENCE_SUBMISSION_STATUS_COLORS: Record<string, string> = {
  pending_review: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
  accepted: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
  rejected: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
  needs_clarification: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300',
};

// ========== Comment Visibility ==========

export const COMMENT_VISIBILITY_LABELS: Record<string, string> = {
  external: 'Visible to Auditors',
  internal: 'Internal Only',
};

// ========== Template Framework ==========

export const TEMPLATE_FRAMEWORK_LABELS: Record<string, string> = {
  soc2: 'SOC 2',
  pci_dss: 'PCI DSS',
  iso27001: 'ISO 27001',
  gdpr: 'GDPR',
};
