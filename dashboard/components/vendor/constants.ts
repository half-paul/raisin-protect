/**
 * Vendor / Service Provider constants — Sprint 12
 * Labels and colors for PCI DSS Req 12.8 vendor management UI
 */

import type {
  SPType,
  SPComplianceStatus,
  SPRiskLevel,
  SPDocType,
  SPResponsibleParty,
} from '@/lib/api';

// ---- SP Type ----

export const SP_TYPE_LABELS: Record<SPType, string> = {
  payment_processor:  'Payment Processor',
  payment_gateway:    'Payment Gateway',
  acquirer:           'Acquirer',
  tokenization:       'Tokenization',
  hosting:            'Hosting / IaaS',
  managed_security:   'Managed Security',
  software:           'Payment Software',
  network:            'Network / SD-WAN',
  cloud_storage:      'Cloud Storage',
  third_party_agent:  'Third-Party Agent',
  other:              'Other',
};

// ---- PCI Compliance Status ----

export const SP_COMPLIANCE_LABELS: Record<SPComplianceStatus, string> = {
  compliant:                  'Compliant',
  compliance_in_progress:     'In Progress',
  compliance_not_validated:   'Not Validated',
  non_compliant:              'Non-Compliant',
  not_applicable:             'N/A',
  unknown:                    'Unknown',
};

export const SP_COMPLIANCE_COLORS: Record<SPComplianceStatus, string> = {
  compliant:                  'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  compliance_in_progress:     'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  compliance_not_validated:   'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  non_compliant:              'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  not_applicable:             'bg-gray-100 text-gray-600 dark:bg-gray-900/30 dark:text-gray-400',
  unknown:                    'bg-gray-100 text-gray-500 dark:bg-gray-900/30 dark:text-gray-500',
};

// ---- Risk Level ----

export const SP_RISK_LABELS: Record<SPRiskLevel, string> = {
  critical: 'Critical',
  high:     'High',
  medium:   'Medium',
  low:      'Low',
};

export const SP_RISK_COLORS: Record<SPRiskLevel, string> = {
  critical: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  high:     'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
  medium:   'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  low:      'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
};

// ---- Document Type ----

export const SP_DOC_TYPE_LABELS: Record<SPDocType, string> = {
  aoc:           'AOC (Attestation of Compliance)',
  soc2_type1:    'SOC 2 Type I',
  soc2_type2:    'SOC 2 Type II',
  iso27001:      'ISO 27001',
  csa_star:      'CSA STAR',
  pentest:       'Penetration Test',
  questionnaire: 'Security Questionnaire',
  other:         'Other',
};

// ---- Responsible Party ----

export const SP_PARTY_LABELS: Record<SPResponsibleParty, string> = {
  merchant: 'Merchant',
  provider: 'Provider',
  shared:   'Shared',
};

export const SP_PARTY_COLORS: Record<SPResponsibleParty, string> = {
  merchant: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  provider: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  shared:   'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
};
