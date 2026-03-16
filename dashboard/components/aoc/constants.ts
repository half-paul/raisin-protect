/**
 * AOC / ROC document builder constants — Sprint 12
 * Labels and colors for PCI DSS Req 12.4 compliance document UI
 */

import type {
  ComplianceDocumentType,
  ComplianceDocumentStatus,
  DocumentSectionComplianceStatus,
  AttestationRole,
} from '@/lib/api';

// ---- Document Type ----

export const DOC_TYPE_LABELS: Record<ComplianceDocumentType, string> = {
  aoc_saq_a:    'AOC — SAQ A',
  aoc_saq_a_ep: 'AOC — SAQ A-EP',
  aoc_saq_b:    'AOC — SAQ B',
  aoc_saq_b_ip: 'AOC — SAQ B-IP',
  aoc_saq_c_vt: 'AOC — SAQ C-VT',
  aoc_saq_c:    'AOC — SAQ C',
  aoc_saq_d:    'AOC — SAQ D',
  aoc_saq_d_sp: 'AOC — SAQ D (SP)',
  roc:          'Report on Compliance (ROC)',
};

export const DOC_TYPE_DESCRIPTIONS: Record<ComplianceDocumentType, string> = {
  aoc_saq_a:    'Fully-outsourced card-not-present merchants',
  aoc_saq_a_ep: 'E-commerce merchants with payment redirect',
  aoc_saq_b:    'Imprint machines / standalone terminals',
  aoc_saq_b_ip: 'IP-connected standalone terminals',
  aoc_saq_c_vt: 'Virtual payment terminals',
  aoc_saq_c:    'Payment application systems',
  aoc_saq_d:    'All other merchants',
  aoc_saq_d_sp: 'All service providers',
  roc:          'Level 1 merchants (over 6M transactions/year)',
};

// ---- Document Status ----

export const DOC_STATUS_LABELS: Record<ComplianceDocumentStatus, string> = {
  draft:      'Draft',
  generating: 'Generating',
  review:     'In Review',
  approved:   'Approved',
  final:      'Final',
  signed:     'Signed',
  superseded: 'Superseded',
  cancelled:  'Cancelled',
};

export const DOC_STATUS_COLORS: Record<ComplianceDocumentStatus, string> = {
  draft:      'bg-gray-100 text-gray-600 dark:bg-gray-900/30 dark:text-gray-400',
  generating: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  review:     'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  approved:   'bg-teal-100 text-teal-800 dark:bg-teal-900/30 dark:text-teal-400',
  final:      'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  signed:     'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400',
  superseded: 'bg-gray-100 text-gray-500 dark:bg-gray-900/30 dark:text-gray-500',
  cancelled:  'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400',
};

// Which statuses are terminal (document locked)
export const DOC_TERMINAL_STATUSES: ComplianceDocumentStatus[] = ['final', 'signed', 'cancelled'];

// ---- Section Compliance Status ----

export const SECTION_STATUS_LABELS: Record<DocumentSectionComplianceStatus, string> = {
  compliant:            'Compliant',
  non_compliant:        'Non-Compliant',
  partially_compliant:  'Partial',
  not_applicable:       'N/A',
  compensating_control: 'Compensating Control',
  customized_approach:  'Customized Approach',
};

export const SECTION_STATUS_COLORS: Record<DocumentSectionComplianceStatus, string> = {
  compliant:            'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  non_compliant:        'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  partially_compliant:  'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  not_applicable:       'bg-gray-100 text-gray-500 dark:bg-gray-900/30 dark:text-gray-400',
  compensating_control: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
  customized_approach:  'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
};

// ---- Attestation Role ----

export const ATTESTATION_ROLE_LABELS: Record<AttestationRole, string> = {
  merchant_signatory: 'Merchant Signatory',
  qsa_signatory:      'QSA Signatory',
  isac_signatory:     'Internal Security Assessor (ISAC)',
  sp_signatory:       'Service Provider Signatory',
};
