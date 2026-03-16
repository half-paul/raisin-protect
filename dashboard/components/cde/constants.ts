import type { CdeAssetScope, CdeAssetType, IsolationMethod, DataFlowProtocol, EncryptionStatus, SegmentationTestResult } from '@/lib/api';

// ---- Asset Scope ----

export const ASSET_SCOPE_LABELS: Record<CdeAssetScope, string> = {
  in_scope: 'In Scope',
  out_of_scope: 'Out of Scope',
  connected_to: 'Connected To',
};

export const ASSET_SCOPE_COLORS: Record<CdeAssetScope, string> = {
  in_scope: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  out_of_scope: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  connected_to: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
};

// ---- Asset Type ----

export const ASSET_TYPE_LABELS: Record<CdeAssetType, string> = {
  server: 'Server',
  database: 'Database',
  network_device: 'Network Device',
  application: 'Application',
  endpoint: 'Endpoint',
  cloud_service: 'Cloud Service',
  storage: 'Storage',
  other: 'Other',
};

// ---- Isolation Method ----

export const ISOLATION_METHOD_LABELS: Record<IsolationMethod, string> = {
  firewall: 'Firewall',
  vlan: 'VLAN',
  dmz: 'DMZ',
  air_gap: 'Air Gap',
  cloud_vpc: 'Cloud VPC',
  microsegmentation: 'Microsegmentation',
  other: 'Other',
};

export const ISOLATION_METHOD_COLORS: Record<IsolationMethod, string> = {
  firewall: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  vlan: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
  dmz: 'bg-cyan-100 text-cyan-800 dark:bg-cyan-900/30 dark:text-cyan-400',
  air_gap: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  cloud_vpc: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400',
  microsegmentation: 'bg-teal-100 text-teal-800 dark:bg-teal-900/30 dark:text-teal-400',
  other: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300',
};

// ---- Protocol ----

export const PROTOCOL_LABELS: Record<DataFlowProtocol, string> = {
  https: 'HTTPS',
  http: 'HTTP',
  tls: 'TLS',
  ssh: 'SSH',
  sftp: 'SFTP',
  smb: 'SMB',
  ftp: 'FTP',
  other: 'Other',
};

// ---- Encryption Status ----

export const ENCRYPTION_LABELS: Record<EncryptionStatus, string> = {
  encrypted: 'Encrypted',
  unencrypted: 'Unencrypted',
  partial: 'Partial',
};

export const ENCRYPTION_COLORS: Record<EncryptionStatus, string> = {
  encrypted: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  unencrypted: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  partial: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
};

// ---- Test Result ----

export const TEST_RESULT_LABELS: Record<SegmentationTestResult, string> = {
  pass: 'Pass',
  fail: 'Fail',
  pending: 'Pending',
};

export const TEST_RESULT_COLORS: Record<SegmentationTestResult, string> = {
  pass: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  fail: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
};
