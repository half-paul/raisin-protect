import { PolicyStatus, PolicyCategory, PolicyReviewStatus, SignoffStatus } from '@/lib/api';

export const POLICY_STATUS_LABELS: Record<string, string> = {
  draft: 'Draft',
  in_review: 'In Review',
  approved: 'Approved',
  published: 'Published',
  archived: 'Archived',
};

export const POLICY_STATUS_COLORS: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300',
  in_review: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  approved: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  published: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  archived: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
};

export const POLICY_CATEGORY_LABELS: Record<string, string> = {
  information_security: 'Information Security',
  access_control: 'Access Control',
  incident_response: 'Incident Response',
  data_privacy: 'Data Privacy',
  network_security: 'Network Security',
  encryption: 'Encryption',
  vulnerability_management: 'Vulnerability Management',
  change_management: 'Change Management',
  business_continuity: 'Business Continuity',
  secure_development: 'Secure Development',
  vendor_management: 'Vendor Management',
  acceptable_use: 'Acceptable Use',
  physical_security: 'Physical Security',
  hr_security: 'HR Security',
  asset_management: 'Asset Management',
};

export const REVIEW_STATUS_LABELS: Record<string, string> = {
  overdue: 'Overdue',
  due_soon: 'Due Soon',
  on_track: 'On Track',
  no_schedule: 'No Schedule',
};

export const REVIEW_STATUS_COLORS: Record<string, string> = {
  overdue: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  due_soon: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  on_track: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  no_schedule: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400',
};

export const SIGNOFF_STATUS_COLORS: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  approved: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  rejected: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  withdrawn: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400',
};

export const URGENCY_COLORS: Record<string, string> = {
  overdue: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  due_soon: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  on_time: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
};
