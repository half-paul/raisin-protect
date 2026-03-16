'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  ArrowLeft,
  RefreshCw,
  Play,
  Pause,
  Trash2,
  CheckCircle2,
  XCircle,
  Clock,
  AlertCircle,
  Activity,
  Webhook,
  FileText,
  Loader2,
  Plus,
  RotateCw,
  Copy,
} from 'lucide-react';
import {
  IntegrationConnection,
  IntegrationRun,
  IntegrationLog,
  IntegrationWebhook,
  IntegrationHealthData,
  getIntegrationConnection,
  getIntegrationConnectionHealth,
  listIntegrationRuns,
  getIntegrationRunLogs,
  listIntegrationWebhooks,
  testIntegrationConnection,
  enableIntegrationConnection,
  disableIntegrationConnection,
  triggerIntegrationSync,
  cancelIntegrationRun,
  deleteIntegrationConnection,
  createIntegrationWebhook,
  deleteIntegrationWebhook,
  rotateIntegrationWebhookSecret,
  triggerIntegrationHealthCheck,
} from '@/lib/api';

const STATUS_COLORS: Record<string, string> = {
  connected: 'bg-green-100 text-green-800',
  pending: 'bg-yellow-100 text-yellow-800',
  disconnected: 'bg-gray-100 text-gray-800',
  error: 'bg-red-100 text-red-800',
  disabled: 'bg-gray-100 text-gray-500',
};

const HEALTH_COLORS: Record<string, string> = {
  healthy: 'bg-green-100 text-green-800',
  degraded: 'bg-yellow-100 text-yellow-800',
  unhealthy: 'bg-red-100 text-red-800',
  unknown: 'bg-gray-100 text-gray-500',
};

const RUN_STATUS_COLORS: Record<string, string> = {
  completed: 'bg-green-100 text-green-800',
  running: 'bg-blue-100 text-blue-800',
  pending: 'bg-yellow-100 text-yellow-800',
  partial: 'bg-orange-100 text-orange-800',
  failed: 'bg-red-100 text-red-800',
  cancelled: 'bg-gray-100 text-gray-500',
};

const LOG_LEVEL_COLORS: Record<string, string> = {
  info: 'text-blue-600',
  warn: 'text-yellow-600',
  error: 'text-red-600',
  debug: 'text-gray-500',
};

export default function ConnectionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { hasRole } = useAuth();
  const canManage = hasRole('ciso', 'compliance_manager', 'it_admin', 'devops_engineer');
  const canDelete = hasRole('ciso', 'it_admin');

  const [connection, setConnection] = useState<IntegrationConnection | null>(null);
  const [health, setHealth] = useState<IntegrationHealthData | null>(null);
  const [runs, setRuns] = useState<IntegrationRun[]>([]);
  const [logs, setLogs] = useState<IntegrationLog[]>([]);
  const [webhooks, setWebhooks] = useState<IntegrationWebhook[]>([]);
  const [selectedRunId, setSelectedRunId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('overview');

  // Webhook creation
  const [showWebhookDialog, setShowWebhookDialog] = useState(false);
  const [webhookName, setWebhookName] = useState('');
  const [newWebhookSecret, setNewWebhookSecret] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);

  const fetchConnection = useCallback(async () => {
    if (!id) return;
    try {
      setLoading(true);
      const res = await getIntegrationConnection(id);
      setConnection(res.data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load connection');
    } finally {
      setLoading(false);
    }
  }, [id]);

  const fetchHealth = useCallback(async () => {
    if (!id) return;
    try {
      const res = await getIntegrationConnectionHealth(id);
      setHealth(res.data);
    } catch (err) { console.error('Failed to fetch health:', err); }
  }, [id]);

  const fetchRuns = useCallback(async () => {
    if (!id) return;
    try {
      const res = await listIntegrationRuns(id);
      setRuns(res.data || []);
    } catch (err) { console.error('Failed to fetch runs:', err); }
  }, [id]);

  const fetchWebhooks = useCallback(async () => {
    if (!id) return;
    try {
      const res = await listIntegrationWebhooks(id);
      setWebhooks(res.data || []);
    } catch (err) { console.error('Failed to fetch webhooks:', err); }
  }, [id]);

  const fetchLogs = useCallback(async (runId: string) => {
    if (!id) return;
    try {
      const res = await getIntegrationRunLogs(id, runId);
      setLogs(res.data || []);
    } catch (err) { console.error('Failed to fetch logs:', err); }
  }, [id]);

  useEffect(() => { fetchConnection(); }, [fetchConnection]);
  useEffect(() => {
    if (activeTab === 'overview') fetchHealth();
    if (activeTab === 'runs') fetchRuns();
    if (activeTab === 'webhooks') fetchWebhooks();
  }, [activeTab, fetchHealth, fetchRuns, fetchWebhooks]);

  useEffect(() => {
    if (selectedRunId) fetchLogs(selectedRunId);
  }, [selectedRunId, fetchLogs]);

  const handleTest = async () => {
    try { await testIntegrationConnection(id); fetchConnection(); fetchHealth(); } catch (err) { setError(err instanceof Error ? err.message : 'Test failed'); }
  };
  const handleEnable = async () => {
    try { await enableIntegrationConnection(id); fetchConnection(); } catch (err) { setError(err instanceof Error ? err.message : 'Enable failed'); }
  };
  const handleDisable = async () => {
    try { await disableIntegrationConnection(id); fetchConnection(); } catch (err) { setError(err instanceof Error ? err.message : 'Disable failed'); }
  };
  const handleSync = async () => {
    try { await triggerIntegrationSync(id); fetchRuns(); fetchConnection(); } catch (err) { setError(err instanceof Error ? err.message : 'Sync failed'); }
  };
  const handleCancelRun = async (runId: string) => {
    try { await cancelIntegrationRun(id, runId); fetchRuns(); } catch (err) { setError(err instanceof Error ? err.message : 'Cancel failed'); }
  };
  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this connection? This action cannot be undone.')) return;
    try { await deleteIntegrationConnection(id); router.push('/integrations'); } catch (err) { setError(err instanceof Error ? err.message : 'Delete failed'); }
  };
  const handleHealthCheck = async () => {
    try { await triggerIntegrationHealthCheck(id); fetchHealth(); fetchConnection(); } catch (err) { setError(err instanceof Error ? err.message : 'Health check failed'); }
  };
  const handleCreateWebhook = async () => {
    if (!webhookName.trim()) return;
    try {
      setCreating(true);
      const res = await createIntegrationWebhook(id, { name: webhookName.trim() });
      setNewWebhookSecret(res.data.webhook_secret);
      setWebhookName('');
      fetchWebhooks();
    } catch (err) { setError(err instanceof Error ? err.message : 'Failed to create webhook'); } finally { setCreating(false); }
  };
  const handleDeleteWebhook = async (webhookId: string) => {
    if (!confirm('Delete this webhook?')) return;
    try { await deleteIntegrationWebhook(id, webhookId); fetchWebhooks(); } catch (err) { setError(err instanceof Error ? err.message : 'Delete failed'); }
  };
  const handleRotateSecret = async (webhookId: string) => {
    if (!confirm('Rotate webhook secret? The old secret will stop working immediately.')) return;
    try {
      const res = await rotateIntegrationWebhookSecret(id, webhookId);
      setNewWebhookSecret(res.data.webhook_secret);
    } catch (err) { setError(err instanceof Error ? err.message : 'Rotate failed'); }
  };

  if (loading) {
    return <div className="flex items-center justify-center py-12"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>;
  }

  if (!connection) {
    return <div className="text-center py-12 text-muted-foreground">Connection not found</div>;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Link href="/integrations">
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{connection.name}</h1>
            <Badge className={STATUS_COLORS[connection.status] || ''}>{connection.status}</Badge>
            <Badge className={HEALTH_COLORS[connection.health] || ''}>{connection.health}</Badge>
          </div>
          <p className="text-muted-foreground">{connection.definition_name} &middot; {connection.description || 'No description'}</p>
        </div>
        {canManage && (
          <div className="flex gap-2">
            {connection.status === 'connected' && (
              <>
                <Button variant="outline" size="sm" onClick={handleSync}><RefreshCw className="mr-1 h-3 w-3" /> Sync</Button>
                <Button variant="outline" size="sm" onClick={handleDisable}><Pause className="mr-1 h-3 w-3" /> Disable</Button>
              </>
            )}
            {connection.status !== 'connected' && connection.status !== 'disabled' && (
              <>
                <Button variant="outline" size="sm" onClick={handleTest}>Test</Button>
                <Button variant="outline" size="sm" onClick={handleEnable}><Play className="mr-1 h-3 w-3" /> Enable</Button>
              </>
            )}
            {connection.status === 'disabled' && (
              <Button variant="outline" size="sm" onClick={handleEnable}><Play className="mr-1 h-3 w-3" /> Enable</Button>
            )}
            {canDelete && (
              <Button variant="destructive" size="sm" onClick={handleDelete}><Trash2 className="mr-1 h-3 w-3" /> Delete</Button>
            )}
          </div>
        )}
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-700">
          <AlertCircle className="inline h-4 w-4 mr-1" />{error}
          <button className="ml-2 underline" onClick={() => setError(null)}>Dismiss</button>
        </div>
      )}

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="overview"><Activity className="mr-1 h-4 w-4" />Overview</TabsTrigger>
          <TabsTrigger value="runs"><RefreshCw className="mr-1 h-4 w-4" />Runs</TabsTrigger>
          <TabsTrigger value="logs"><FileText className="mr-1 h-4 w-4" />Logs</TabsTrigger>
          <TabsTrigger value="webhooks"><Webhook className="mr-1 h-4 w-4" />Webhooks</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Connection Details</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <div className="flex justify-between"><span className="text-muted-foreground">Provider</span><span>{connection.definition_name}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Category</span><span>{connection.definition_category}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Sync Enabled</span><span>{connection.sync_enabled ? 'Yes' : 'No'}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Sync Interval</span><span>{connection.sync_interval_mins} min</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Last Sync</span><span>{connection.last_sync_at ? new Date(connection.last_sync_at).toLocaleString() : 'Never'}</span></div>
                <div className="flex justify-between"><span className="text-muted-foreground">Created</span><span>{new Date(connection.created_at).toLocaleString()}</span></div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">Health</CardTitle>
                  {canManage && <Button variant="outline" size="sm" onClick={handleHealthCheck}><RefreshCw className="mr-1 h-3 w-3" /> Check</Button>}
                </div>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                {health ? (
                  <>
                    <div className="flex justify-between"><span className="text-muted-foreground">Health</span><Badge className={HEALTH_COLORS[health.health]}>{health.health}</Badge></div>
                    <div className="flex justify-between"><span className="text-muted-foreground">Uptime</span><span>{health.uptime_pct.toFixed(1)}%</span></div>
                    <div className="flex justify-between"><span className="text-muted-foreground">Total Runs</span><span>{health.total_runs}</span></div>
                    <div className="flex justify-between"><span className="text-muted-foreground">Successful</span><span className="text-green-600">{health.successful_runs}</span></div>
                    <div className="flex justify-between"><span className="text-muted-foreground">Failed</span><span className="text-red-600">{health.failed_runs}</span></div>
                    <div className="flex justify-between"><span className="text-muted-foreground">Consecutive Failures</span><span>{health.consecutive_failures}</span></div>
                    <div className="flex justify-between"><span className="text-muted-foreground">Last Check</span><span>{health.last_health_check_at ? new Date(health.last_health_check_at).toLocaleString() : 'Never'}</span></div>
                  </>
                ) : (
                  <p className="text-muted-foreground">Loading health data...</p>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Config (masked) */}
          <Card>
            <CardHeader><CardTitle className="text-base">Configuration</CardTitle></CardHeader>
            <CardContent>
              <pre className="bg-muted p-4 rounded-md text-xs overflow-auto max-h-[200px]">
                {JSON.stringify(connection.config, null, 2)}
              </pre>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Runs Tab */}
        <TabsContent value="runs" className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium">Sync Runs</h3>
            {canManage && connection.status === 'connected' && (
              <Button size="sm" onClick={handleSync}><RefreshCw className="mr-1 h-3 w-3" /> Trigger Sync</Button>
            )}
          </div>
          {runs.length === 0 ? (
            <p className="text-muted-foreground text-center py-8">No runs yet</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableHead>Trigger</TableHead>
                  <TableHead>Started</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Stats</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {runs.map((run) => (
                  <TableRow key={run.id} className="cursor-pointer" onClick={() => { setSelectedRunId(run.id); setActiveTab('logs'); }}>
                    <TableCell><Badge className={RUN_STATUS_COLORS[run.status] || ''}>{run.status}</Badge></TableCell>
                    <TableCell>{run.trigger}</TableCell>
                    <TableCell>{run.started_at ? new Date(run.started_at).toLocaleString() : '-'}</TableCell>
                    <TableCell>{run.duration_ms ? `${(run.duration_ms / 1000).toFixed(1)}s` : '-'}</TableCell>
                    <TableCell className="text-xs max-w-[200px] truncate">{Object.entries(run.stats || {}).map(([k, v]) => `${k}: ${v}`).join(', ') || '-'}</TableCell>
                    <TableCell>
                      {canManage && (run.status === 'pending' || run.status === 'running') && (
                        <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleCancelRun(run.id); }}>Cancel</Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </TabsContent>

        {/* Logs Tab */}
        <TabsContent value="logs" className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium">Logs {selectedRunId && <span className="text-sm text-muted-foreground ml-2">(Run: {selectedRunId.slice(0, 8)}...)</span>}</h3>
            {runs.length > 0 && !selectedRunId && (
              <p className="text-sm text-muted-foreground">Select a run from the Runs tab to view its logs</p>
            )}
          </div>
          {selectedRunId && logs.length === 0 && (
            <p className="text-muted-foreground text-center py-8">No logs for this run</p>
          )}
          {logs.length > 0 && (
            <div className="bg-muted rounded-md p-4 font-mono text-xs space-y-1 max-h-[400px] overflow-auto">
              {logs.map((log) => (
                <div key={log.id} className="flex gap-2">
                  <span className="text-muted-foreground whitespace-nowrap">{new Date(log.created_at).toLocaleTimeString()}</span>
                  <span className={`font-semibold w-12 ${LOG_LEVEL_COLORS[log.level] || ''}`}>{log.level.toUpperCase()}</span>
                  <span>{log.message}</span>
                  {log.source && <span className="text-muted-foreground">({log.source})</span>}
                </div>
              ))}
            </div>
          )}
          {!selectedRunId && runs.length > 0 && (
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Recent runs:</p>
              {runs.slice(0, 5).map((run) => (
                <button key={run.id} className="block w-full text-left p-2 rounded hover:bg-muted text-sm" onClick={() => setSelectedRunId(run.id)}>
                  <Badge className={`mr-2 ${RUN_STATUS_COLORS[run.status] || ''}`}>{run.status}</Badge>
                  {run.started_at ? new Date(run.started_at).toLocaleString() : 'Queued'} - {run.trigger}
                </button>
              ))}
            </div>
          )}
        </TabsContent>

        {/* Webhooks Tab */}
        <TabsContent value="webhooks" className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium">Webhooks</h3>
            {canManage && (
              <Button size="sm" onClick={() => setShowWebhookDialog(true)}>
                <Plus className="mr-1 h-3 w-3" /> Add Webhook
              </Button>
            )}
          </div>

          {newWebhookSecret && (
            <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
              <p className="text-sm font-medium text-yellow-800 mb-2">Webhook Secret (shown only once):</p>
              <div className="flex items-center gap-2">
                <code className="bg-white px-3 py-1 rounded text-sm font-mono flex-1 truncate">{newWebhookSecret}</code>
                <Button variant="outline" size="sm" onClick={() => { navigator.clipboard.writeText(newWebhookSecret); }}>
                  <Copy className="h-3 w-3" />
                </Button>
              </div>
              <Button variant="ghost" size="sm" className="mt-2" onClick={() => setNewWebhookSecret(null)}>Dismiss</Button>
            </div>
          )}

          {webhooks.length === 0 ? (
            <p className="text-muted-foreground text-center py-8">No webhooks configured</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Events</TableHead>
                  <TableHead>Received</TableHead>
                  <TableHead>Last Received</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {webhooks.map((wh) => (
                  <TableRow key={wh.id}>
                    <TableCell className="font-medium">{wh.name}</TableCell>
                    <TableCell><Badge variant="outline">{wh.status}</Badge></TableCell>
                    <TableCell className="text-xs">{wh.event_types.join(', ') || 'All'}</TableCell>
                    <TableCell>{wh.total_received} ({wh.total_errors} errors)</TableCell>
                    <TableCell>{wh.last_received_at ? new Date(wh.last_received_at).toLocaleString() : 'Never'}</TableCell>
                    <TableCell>
                      {canManage && (
                        <div className="flex gap-1">
                          <Button variant="ghost" size="sm" onClick={() => handleRotateSecret(wh.id)}>
                            <RotateCw className="h-3 w-3" />
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => handleDeleteWebhook(wh.id)}>
                            <Trash2 className="h-3 w-3 text-red-500" />
                          </Button>
                        </div>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </TabsContent>
      </Tabs>

      {/* Add Webhook Dialog */}
      <Dialog open={showWebhookDialog} onOpenChange={setShowWebhookDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Webhook</DialogTitle>
            <DialogDescription>Create a webhook endpoint to receive events from this integration</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="wh-name">Webhook Name</Label>
              <Input id="wh-name" placeholder="e.g., Push Events" value={webhookName} onChange={(e) => setWebhookName(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowWebhookDialog(false)}>Cancel</Button>
            <Button onClick={handleCreateWebhook} disabled={!webhookName.trim() || creating}>
              {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create Webhook
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
