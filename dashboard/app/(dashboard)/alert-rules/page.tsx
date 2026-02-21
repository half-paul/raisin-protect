'use client';

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Settings, Plus, MoreVertical, AlertTriangle, Bell, Power,
  PowerOff, Trash2, ChevronLeft, ChevronRight, AlertCircle,
} from 'lucide-react';
import {
  AlertRule, listAlertRules, createAlertRule, updateAlertRule, deleteAlertRule,
  testAlertDelivery,
} from '@/lib/api';
import { cn } from '@/lib/utils';

const SEVERITY_COLORS: Record<string, string> = {
  critical: 'bg-red-600 text-white',
  high: 'bg-orange-500 text-white',
  medium: 'bg-amber-500 text-white',
  low: 'bg-blue-500 text-white',
};

const DELIVERY_LABELS: Record<string, string> = {
  slack: 'Slack',
  email: 'Email',
  webhook: 'Webhook',
  in_app: 'In-App',
};

export default function AlertRulesPage() {
  const { hasRole } = useAuth();
  const canCreate = hasRole('ciso', 'compliance_manager');
  const canView = hasRole('ciso', 'compliance_manager', 'security_engineer');

  const [rules, setRules] = useState<AlertRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  const [showCreate, setShowCreate] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const [createError, setCreateError] = useState('');
  const [form, setForm] = useState({
    name: '',
    description: '',
    enabled: true,
    match_severities: [] as string[],
    match_result_statuses: ['fail'] as string[],
    consecutive_failures: 1,
    cooldown_minutes: 60,
    alert_severity: 'high' as string,
    sla_hours: '',
    delivery_channels: ['in_app'] as string[],
    slack_webhook_url: '',
    email_recipients: '',
    priority: 100,
  });

  const [showTest, setShowTest] = useState(false);
  const [testChannel, setTestChannel] = useState('slack');
  const [testUrl, setTestUrl] = useState('');
  const [testEmail, setTestEmail] = useState('');
  const [testLoading, setTestLoading] = useState(false);
  const [testResult, setTestResult] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
        sort: 'priority',
        order: 'asc',
      };
      const res = await listAlertRules(params);
      setRules(res.data || []);
      setTotal(res.meta?.total || 0);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => { fetchData(); }, [fetchData]);

  async function handleCreate() {
    setCreateError('');
    setCreateLoading(true);
    try {
      const body: Record<string, unknown> = {
        name: form.name,
        description: form.description || undefined,
        enabled: form.enabled,
        match_severities: form.match_severities.length > 0 ? form.match_severities : null,
        match_result_statuses: form.match_result_statuses.length > 0 ? form.match_result_statuses : ['fail'],
        consecutive_failures: form.consecutive_failures,
        cooldown_minutes: form.cooldown_minutes,
        alert_severity: form.alert_severity,
        sla_hours: form.sla_hours ? parseInt(form.sla_hours) : null,
        delivery_channels: form.delivery_channels,
        priority: form.priority,
      };
      if (form.delivery_channels.includes('slack') && form.slack_webhook_url) {
        body.slack_webhook_url = form.slack_webhook_url;
      }
      if (form.delivery_channels.includes('email') && form.email_recipients) {
        body.email_recipients = form.email_recipients.split(',').map(e => e.trim()).filter(Boolean);
      }
      await createAlertRule(body as Partial<AlertRule>);
      setShowCreate(false);
      resetForm();
      fetchData();
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create');
    } finally {
      setCreateLoading(false);
    }
  }

  function resetForm() {
    setForm({
      name: '', description: '', enabled: true,
      match_severities: [], match_result_statuses: ['fail'],
      consecutive_failures: 1, cooldown_minutes: 60,
      alert_severity: 'high', sla_hours: '',
      delivery_channels: ['in_app'], slack_webhook_url: '',
      email_recipients: '', priority: 100,
    });
  }

  async function handleToggle(rule: AlertRule) {
    try {
      await updateAlertRule(rule.id, { enabled: !rule.enabled } as Partial<AlertRule>);
      fetchData();
    } catch {
      // handle
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this alert rule? Existing alerts will be preserved.')) return;
    try {
      await deleteAlertRule(id);
      fetchData();
    } catch {
      // handle
    }
  }

  async function handleTestDelivery() {
    setTestLoading(true);
    setTestResult(null);
    try {
      const body: Record<string, unknown> = { channel: testChannel };
      if (testChannel === 'slack') body.slack_webhook_url = testUrl;
      if (testChannel === 'email') body.email_recipients = testEmail.split(',').map(e => e.trim()).filter(Boolean);
      if (testChannel === 'webhook') body.webhook_url = testUrl;

      const res = await testAlertDelivery(body as Parameters<typeof testAlertDelivery>[0]);
      setTestResult(res.data.success ? '✅ Delivered successfully' : '❌ Delivery failed');
    } catch (err) {
      setTestResult(`❌ ${err instanceof Error ? err.message : 'Failed'}`);
    } finally {
      setTestLoading(false);
    }
  }

  function toggleSeverityMatch(sev: string) {
    setForm(f => ({
      ...f,
      match_severities: f.match_severities.includes(sev)
        ? f.match_severities.filter(s => s !== sev)
        : [...f.match_severities, sev],
    }));
  }

  function toggleDeliveryChannel(ch: string) {
    setForm(f => ({
      ...f,
      delivery_channels: f.delivery_channels.includes(ch)
        ? f.delivery_channels.filter(c => c !== ch)
        : [...f.delivery_channels, ch],
    }));
  }

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Settings className="h-8 w-8" />
            Alert Rules
          </h1>
          <p className="text-muted-foreground mt-1">
            Configure how alerts are generated from test failures
          </p>
        </div>
        <div className="flex gap-2">
          {canView && (
            <Button variant="outline" onClick={() => setShowTest(true)}>
              <Bell className="h-4 w-4 mr-2" />
              Test Delivery
            </Button>
          )}
          {canCreate && (
            <Button onClick={() => setShowCreate(true)}>
              <Plus className="h-4 w-4 mr-2" />
              New Rule
            </Button>
          )}
        </div>
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : rules.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Settings className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">No alert rules configured</p>
              <p className="text-sm">Create rules to automatically generate alerts from test failures</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[50px]">On</TableHead>
                  <TableHead className="w-[60px]">Priority</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead className="w-[100px]">Severity</TableHead>
                  <TableHead className="w-[100px]">Matches</TableHead>
                  <TableHead className="w-[100px]">Consecutive</TableHead>
                  <TableHead className="w-[100px]">Cooldown</TableHead>
                  <TableHead className="w-[80px]">SLA</TableHead>
                  <TableHead>Channels</TableHead>
                  <TableHead className="w-[90px]">Alerts</TableHead>
                  <TableHead className="w-[50px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {rules.map((rule) => (
                  <TableRow key={rule.id} className={!rule.enabled ? 'opacity-50' : ''}>
                    <TableCell>
                      {canCreate ? (
                        <button onClick={() => handleToggle(rule)}>
                          {rule.enabled ? (
                            <Power className="h-4 w-4 text-green-500" />
                          ) : (
                            <PowerOff className="h-4 w-4 text-muted-foreground" />
                          )}
                        </button>
                      ) : (
                        rule.enabled ? (
                          <Power className="h-4 w-4 text-green-500" />
                        ) : (
                          <PowerOff className="h-4 w-4 text-muted-foreground" />
                        )
                      )}
                    </TableCell>
                    <TableCell className="font-mono text-sm">{rule.priority}</TableCell>
                    <TableCell>
                      <div>
                        <p className="text-sm font-medium">{rule.name}</p>
                        {rule.description && (
                          <p className="text-xs text-muted-foreground line-clamp-1">{rule.description}</p>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge className={cn('text-[10px]', SEVERITY_COLORS[rule.alert_severity])}>
                        {rule.alert_severity}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-0.5">
                        {rule.match_severities?.map(s => (
                          <Badge key={s} variant="outline" className="text-[10px]">{s}</Badge>
                        )) || <span className="text-xs text-muted-foreground">all</span>}
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-center">{rule.consecutive_failures}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {rule.cooldown_minutes > 0 ? `${rule.cooldown_minutes}m` : 'None'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {rule.sla_hours ? `${rule.sla_hours}h` : '—'}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-0.5">
                        {rule.delivery_channels.map(ch => (
                          <Badge key={ch} variant="secondary" className="text-[10px]">
                            {DELIVERY_LABELS[ch] || ch}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="text-sm font-medium">{rule.alerts_generated || 0}</TableCell>
                    <TableCell>
                      {canCreate && (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleToggle(rule)}>
                              {rule.enabled ? 'Disable' : 'Enable'}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => handleDelete(rule.id)}
                            >
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-muted-foreground">Page {page} of {totalPages}</span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Create Rule Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent className="max-w-lg max-h-[85vh] overflow-auto">
          <DialogHeader>
            <DialogTitle>Create Alert Rule</DialogTitle>
            <DialogDescription>
              Define conditions for automatic alert generation from test failures
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {createError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
                <AlertCircle className="h-4 w-4" /> {createError}
              </div>
            )}

            <div className="space-y-2">
              <Label>Name *</Label>
              <Input placeholder="e.g. Critical Test Failures"
                value={form.name}
                onChange={(e) => setForm(f => ({ ...f, name: e.target.value }))} />
            </div>

            <div className="space-y-2">
              <Label>Description</Label>
              <textarea
                className="flex min-h-[60px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                placeholder="Describe this rule..."
                value={form.description}
                onChange={(e) => setForm(f => ({ ...f, description: e.target.value }))} />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Alert Severity *</Label>
                <Select value={form.alert_severity} onValueChange={(v) => setForm(f => ({ ...f, alert_severity: v }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="critical">Critical</SelectItem>
                    <SelectItem value="high">High</SelectItem>
                    <SelectItem value="medium">Medium</SelectItem>
                    <SelectItem value="low">Low</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Priority (lower = higher)</Label>
                <Input type="number" value={form.priority}
                  onChange={(e) => setForm(f => ({ ...f, priority: parseInt(e.target.value) || 100 }))} />
              </div>
            </div>

            <div className="space-y-2">
              <Label>Match Test Severities (empty = all)</Label>
              <div className="flex flex-wrap gap-2">
                {['critical', 'high', 'medium', 'low', 'informational'].map(sev => (
                  <label key={sev} className="flex items-center gap-1.5 text-sm">
                    <Checkbox checked={form.match_severities.includes(sev)}
                      onCheckedChange={() => toggleSeverityMatch(sev)} />
                    {sev}
                  </label>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Consecutive Failures</Label>
                <Input type="number" min={1} max={100} value={form.consecutive_failures}
                  onChange={(e) => setForm(f => ({ ...f, consecutive_failures: parseInt(e.target.value) || 1 }))} />
              </div>
              <div className="space-y-2">
                <Label>Cooldown (minutes)</Label>
                <Input type="number" min={0} value={form.cooldown_minutes}
                  onChange={(e) => setForm(f => ({ ...f, cooldown_minutes: parseInt(e.target.value) || 0 }))} />
              </div>
            </div>

            <div className="space-y-2">
              <Label>SLA Hours (empty = no SLA)</Label>
              <Input type="number" min={1} placeholder="e.g. 4"
                value={form.sla_hours}
                onChange={(e) => setForm(f => ({ ...f, sla_hours: e.target.value }))} />
            </div>

            <div className="space-y-2">
              <Label>Delivery Channels *</Label>
              <div className="flex flex-wrap gap-3">
                {Object.entries(DELIVERY_LABELS).map(([ch, label]) => (
                  <label key={ch} className="flex items-center gap-1.5 text-sm">
                    <Checkbox checked={form.delivery_channels.includes(ch)}
                      onCheckedChange={() => toggleDeliveryChannel(ch)} />
                    {label}
                  </label>
                ))}
              </div>
            </div>

            {form.delivery_channels.includes('slack') && (
              <div className="space-y-2">
                <Label>Slack Webhook URL</Label>
                <Input type="url" placeholder="https://hooks.slack.com/services/..."
                  value={form.slack_webhook_url}
                  onChange={(e) => setForm(f => ({ ...f, slack_webhook_url: e.target.value }))} />
              </div>
            )}

            {form.delivery_channels.includes('email') && (
              <div className="space-y-2">
                <Label>Email Recipients (comma-separated)</Label>
                <Input placeholder="security@acme.com, ciso@acme.com"
                  value={form.email_recipients}
                  onChange={(e) => setForm(f => ({ ...f, email_recipients: e.target.value }))} />
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setShowCreate(false); resetForm(); }}>Cancel</Button>
            <Button onClick={handleCreate}
              disabled={createLoading || !form.name || form.delivery_channels.length === 0}>
              {createLoading ? 'Creating...' : 'Create Rule'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Test Delivery Dialog */}
      <Dialog open={showTest} onOpenChange={setShowTest}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Test Alert Delivery</DialogTitle>
            <DialogDescription>
              Send a test notification to verify channel configuration
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Channel</Label>
              <Select value={testChannel} onValueChange={(v) => { setTestChannel(v); setTestResult(null); }}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="slack">Slack</SelectItem>
                  <SelectItem value="email">Email</SelectItem>
                  <SelectItem value="webhook">Webhook</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {testChannel === 'slack' && (
              <div className="space-y-2">
                <Label>Webhook URL</Label>
                <Input type="url" placeholder="https://hooks.slack.com/services/..."
                  value={testUrl} onChange={(e) => setTestUrl(e.target.value)} />
              </div>
            )}
            {testChannel === 'email' && (
              <div className="space-y-2">
                <Label>Recipients</Label>
                <Input placeholder="test@acme.com"
                  value={testEmail} onChange={(e) => setTestEmail(e.target.value)} />
              </div>
            )}
            {testChannel === 'webhook' && (
              <div className="space-y-2">
                <Label>Webhook URL</Label>
                <Input type="url" placeholder="https://api.example.com/hooks/..."
                  value={testUrl} onChange={(e) => setTestUrl(e.target.value)} />
              </div>
            )}
            {testResult && (
              <p className="text-sm">{testResult}</p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowTest(false)}>Close</Button>
            <Button onClick={handleTestDelivery} disabled={testLoading}>
              {testLoading ? 'Sending...' : 'Send Test'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
