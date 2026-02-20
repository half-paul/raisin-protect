'use client';

import { Badge } from '@/components/ui/badge';
import { CheckCircle2, AlertTriangle, XCircle, Clock } from 'lucide-react';
import { FreshnessStatus } from '@/lib/api';

interface FreshnessBadgeProps {
  status: FreshnessStatus | null | undefined;
  daysUntilExpiry?: number | null;
  daysOverdue?: number | null;
  showIcon?: boolean;
  className?: string;
}

const FRESHNESS_CONFIG: Record<string, {
  label: string;
  variant: 'default' | 'secondary' | 'destructive' | 'outline';
  icon: typeof CheckCircle2;
  className: string;
}> = {
  fresh: {
    label: 'Fresh',
    variant: 'default',
    icon: CheckCircle2,
    className: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
  },
  expiring_soon: {
    label: 'Expiring Soon',
    variant: 'outline',
    icon: AlertTriangle,
    className: 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20',
  },
  expired: {
    label: 'Expired',
    variant: 'destructive',
    icon: XCircle,
    className: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
  },
};

export function FreshnessBadge({ status, daysUntilExpiry, daysOverdue, showIcon = true, className }: FreshnessBadgeProps) {
  if (!status) {
    return (
      <Badge variant="secondary" className={`text-xs ${className || ''}`}>
        <Clock className="h-3 w-3 mr-1" />
        No Expiry
      </Badge>
    );
  }

  const config = FRESHNESS_CONFIG[status] || FRESHNESS_CONFIG.fresh;
  const Icon = config.icon;

  let detail = '';
  if (status === 'expired' && daysOverdue != null) {
    detail = ` (${daysOverdue}d overdue)`;
  } else if (status === 'expiring_soon' && daysUntilExpiry != null) {
    detail = ` (${daysUntilExpiry}d)`;
  } else if (status === 'fresh' && daysUntilExpiry != null) {
    detail = ` (${daysUntilExpiry}d)`;
  }

  return (
    <Badge variant="outline" className={`text-xs ${config.className} ${className || ''}`}>
      {showIcon && <Icon className="h-3 w-3 mr-1" />}
      {config.label}{detail}
    </Badge>
  );
}

export const EVIDENCE_TYPE_LABELS: Record<string, string> = {
  screenshot: 'Screenshot',
  api_response: 'API Response',
  configuration_export: 'Config Export',
  log_sample: 'Log Sample',
  policy_document: 'Policy Document',
  access_list: 'Access List',
  vulnerability_report: 'Vuln Report',
  certificate: 'Certificate',
  training_record: 'Training Record',
  penetration_test: 'Pen Test',
  audit_report: 'Audit Report',
  other: 'Other',
};

export const EVIDENCE_STATUS_LABELS: Record<string, string> = {
  draft: 'Draft',
  pending_review: 'Pending Review',
  approved: 'Approved',
  rejected: 'Rejected',
  expired: 'Expired',
  superseded: 'Superseded',
};

export const EVIDENCE_STATUS_COLORS: Record<string, string> = {
  draft: 'bg-gray-500/10 text-gray-700 dark:text-gray-400 border-gray-500/20',
  pending_review: 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20',
  approved: 'bg-green-500/10 text-green-700 dark:text-green-400 border-green-500/20',
  rejected: 'bg-red-500/10 text-red-700 dark:text-red-400 border-red-500/20',
  expired: 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/20',
  superseded: 'bg-gray-500/10 text-gray-500 dark:text-gray-500 border-gray-500/20',
};

export const COLLECTION_METHOD_LABELS: Record<string, string> = {
  manual_upload: 'Manual Upload',
  automated_pull: 'Automated Pull',
  api_ingestion: 'API Ingestion',
  screenshot_capture: 'Screenshot Capture',
  system_export: 'System Export',
};

export const VERDICT_CONFIG: Record<string, { label: string; className: string }> = {
  sufficient: { label: 'Sufficient', className: 'bg-green-500/10 text-green-700 dark:text-green-400' },
  partial: { label: 'Partial', className: 'bg-amber-500/10 text-amber-700 dark:text-amber-400' },
  insufficient: { label: 'Insufficient', className: 'bg-red-500/10 text-red-700 dark:text-red-400' },
  needs_update: { label: 'Needs Update', className: 'bg-blue-500/10 text-blue-700 dark:text-blue-400' },
};

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
