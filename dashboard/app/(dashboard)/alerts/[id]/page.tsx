'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  ArrowLeft, AlertTriangle, Bell, Shield, Clock, User, CheckCircle2,
  XCircle, Play, Pause, Send, MessageSquare,
} from 'lucide-react';
import {
  Alert, getAlert, changeAlertStatus, assignAlert, resolveAlert,
  suppressAlert, closeAlert, redeliverAlert,
} from '@/lib/api';
import { cn } from '@/lib/utils';
import Link from 'next/link';

const SEVERITY_COLORS: Record<string, string> = {
  critical: 'bg-red-600 text-white',
  high: 'bg-orange-500 text-white',
  medium: 'bg-amber-500 text-white',
  low: 'bg-blue-500 text-white',
};

const STATUS_STYLES: Record<string, { variant: 'default' | 'secondary' | 'destructive' | 'outline'; label: string }> = {
  open: { variant: 'destructive', label: 'Open' },
  acknowledged: { variant: 'outline', label: 'Acknowledged' },
  in_progress: { variant: 'default', label: 'In Progress' },
  resolved: { variant: 'secondary', label: 'Resolved' },
  suppressed: { variant: 'outline', label: 'Suppressed' },
  closed: { variant: 'secondary', label: 'Closed' },
};

export default function AlertDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { hasRole } = useAuth();
  const alertId = params.id as string;

  const canManage = hasRole('ciso', 'compliance_manager', 'security_engineer', 'it_admin', 'devops_engineer');
  const canAssign = hasRole('ciso', 'compliance_manager', 'security_engineer');
  const canSuppress = hasRole('ciso', 'compliance_manager');

  const [alert, setAlert] = useState<Alert | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionLoading, setActionLoading] = useState(false);

  // Resolve dialog
  const [showResolve, setShowResolve] = useState(false);
  const [resolutionNotes, setResolutionNotes] = useState('');

  // Suppress dialog
  const [showSuppress, setShowSuppress] = useState(false);
  const [suppressUntil, setSuppressUntil] = useState('');
  const [suppressReason, setSuppressReason] = useState('');

  // Assign dialog
  const [showAssign, setShowAssign] = useState(false);
  const [assignUserId, setAssignUserId] = useState('');

  // Close dialog
  const [showClose, setShowClose] = useState(false);
  const [closeNotes, setCloseNotes] = useState('');

  const fetchAlert = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getAlert(alertId);
      setAlert(res.data);
    } catch {
      setError('Failed to load alert');
    } finally {
      setLoading(false);
    }
  }, [alertId]);

  useEffect(() => { fetchAlert(); }, [fetchAlert]);

  async function handleStatusChange(newStatus: string) {
    setActionLoading(true);
    try {
      await changeAlertStatus(alertId, newStatus);
      await fetchAlert();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to update status');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleResolve() {
    if (!resolutionNotes.trim()) return;
    setActionLoading(true);
    try {
      await resolveAlert(alertId, resolutionNotes);
      setShowResolve(false);
      setResolutionNotes('');
      await fetchAlert();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to resolve');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleSuppress() {
    if (!suppressUntil || !suppressReason.trim()) return;
    setActionLoading(true);
    try {
      await suppressAlert(alertId, new Date(suppressUntil).toISOString(), suppressReason);
      setShowSuppress(false);
      setSuppressUntil('');
      setSuppressReason('');
      await fetchAlert();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to suppress');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleAssign() {
    if (!assignUserId.trim()) return;
    setActionLoading(true);
    try {
      await assignAlert(alertId, assignUserId);
      setShowAssign(false);
      setAssignUserId('');
      await fetchAlert();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to assign');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleClose() {
    setActionLoading(true);
    try {
      await closeAlert(alertId, closeNotes || undefined);
      setShowClose(false);
      setCloseNotes('');
      await fetchAlert();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to close');
    } finally {
      setActionLoading(false);
    }
  }

  async function handleRedeliver() {
    setActionLoading(true);
    try {
      await redeliverAlert(alertId);
      await fetchAlert();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to re-deliver');
    } finally {
      setActionLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  if (!alert) {
    return (
      <div className="p-6">
        <Button variant="ghost" onClick={() => router.push('/alerts')}>
          <ArrowLeft className="h-4 w-4 mr-1" /> Back
        </Button>
        <div className="text-center py-12 text-muted-foreground">
          <AlertTriangle className="h-12 w-12 mx-auto mb-4 opacity-50" />
          <p className="text-lg">{error || 'Alert not found'}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <Button variant="ghost" size="sm" className="mb-2" onClick={() => router.push('/alerts')}>
          <ArrowLeft className="h-4 w-4 mr-1" /> Back to Alerts
        </Button>
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-3 mb-1">
              <h1 className="text-2xl font-bold">Alert #{alert.alert_number}</h1>
              <Badge className={cn('text-xs', SEVERITY_COLORS[alert.severity])}>
                {alert.severity}
              </Badge>
              <Badge variant={STATUS_STYLES[alert.status]?.variant || 'outline'}>
                {STATUS_STYLES[alert.status]?.label || alert.status}
              </Badge>
            </div>
            <p className="text-lg text-muted-foreground">{alert.title}</p>
          </div>
        </div>
      </div>

      {error && (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Action buttons */}
      {canManage && (
        <div className="flex flex-wrap gap-2">
          {alert.status === 'open' && (
            <Button variant="outline" size="sm" disabled={actionLoading}
              onClick={() => handleStatusChange('acknowledged')}>
              <CheckCircle2 className="h-4 w-4 mr-1" /> Acknowledge
            </Button>
          )}
          {['open', 'acknowledged'].includes(alert.status) && (
            <Button variant="outline" size="sm" disabled={actionLoading}
              onClick={() => handleStatusChange('in_progress')}>
              <Play className="h-4 w-4 mr-1" /> Start Working
            </Button>
          )}
          {['open', 'acknowledged', 'in_progress'].includes(alert.status) && (
            <Button variant="default" size="sm" disabled={actionLoading}
              onClick={() => setShowResolve(true)}>
              <CheckCircle2 className="h-4 w-4 mr-1" /> Resolve
            </Button>
          )}
          {canAssign && !['closed', 'resolved'].includes(alert.status) && (
            <Button variant="outline" size="sm" disabled={actionLoading}
              onClick={() => setShowAssign(true)}>
              <User className="h-4 w-4 mr-1" /> Assign
            </Button>
          )}
          {canSuppress && alert.status !== 'closed' && (
            <Button variant="outline" size="sm" disabled={actionLoading}
              onClick={() => setShowSuppress(true)}>
              <Pause className="h-4 w-4 mr-1" /> Suppress
            </Button>
          )}
          {canSuppress && (
            <Button variant="outline" size="sm" disabled={actionLoading}
              onClick={() => setShowClose(true)}>
              <XCircle className="h-4 w-4 mr-1" /> Close
            </Button>
          )}
          <Button variant="ghost" size="sm" disabled={actionLoading}
            onClick={handleRedeliver}>
            <Send className="h-4 w-4 mr-1" /> Re-deliver
          </Button>
        </div>
      )}

      <div className="grid gap-6 md:grid-cols-3">
        {/* Main content */}
        <div className="md:col-span-2 space-y-6">
          {/* Description */}
          {alert.description && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Description</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm whitespace-pre-wrap">{alert.description}</p>
              </CardContent>
            </Card>
          )}

          {/* Test Result */}
          {alert.test_result && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Test Result</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex items-center gap-2">
                  <Badge variant={alert.test_result.status === 'fail' ? 'destructive' : 'secondary'}>
                    {alert.test_result.status}
                  </Badge>
                  {alert.test_result.tested_at && (
                    <span className="text-xs text-muted-foreground">
                      Tested: {new Date(alert.test_result.tested_at).toLocaleString()}
                    </span>
                  )}
                </div>
                {alert.test_result.message && (
                  <p className="text-sm">{alert.test_result.message}</p>
                )}
                {alert.test_result.details && (
                  <pre className="text-xs bg-muted p-3 rounded-md overflow-auto max-h-64">
                    {JSON.stringify(alert.test_result.details, null, 2)}
                  </pre>
                )}
              </CardContent>
            </Card>
          )}

          {/* Resolution */}
          {alert.resolution_notes && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-green-500" />
                  Resolution
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <p className="text-sm whitespace-pre-wrap">{alert.resolution_notes}</p>
                {alert.resolved_by && (
                  <p className="text-xs text-muted-foreground">
                    Resolved by {alert.resolved_by.name} on {alert.resolved_at ? new Date(alert.resolved_at).toLocaleString() : '—'}
                  </p>
                )}
              </CardContent>
            </Card>
          )}

          {/* Suppression info */}
          {alert.status === 'suppressed' && alert.suppression_reason && (
            <Card className="border-amber-500/20">
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <Pause className="h-4 w-4 text-amber-500" />
                  Suppressed
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <p className="text-sm">{alert.suppression_reason}</p>
                {alert.suppressed_until && (
                  <p className="text-xs text-muted-foreground">
                    Until: {new Date(alert.suppressed_until).toLocaleString()}
                  </p>
                )}
              </CardContent>
            </Card>
          )}
        </div>

        {/* Sidebar info */}
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm">
              <div>
                <p className="text-muted-foreground text-xs mb-1">Control</p>
                {alert.control ? (
                  <Link href={`/controls/${alert.control.id}`} className="hover:text-primary">
                    <div className="flex items-center gap-1">
                      <Shield className="h-3 w-3" />
                      <span className="font-mono text-xs">{alert.control.identifier}</span>
                    </div>
                    <p className="text-xs text-muted-foreground">{alert.control.title}</p>
                  </Link>
                ) : <span className="text-muted-foreground">—</span>}
              </div>

              <Separator />

              <div>
                <p className="text-muted-foreground text-xs mb-1">Test</p>
                {alert.test ? (
                  <div>
                    <span className="font-mono text-xs">{alert.test.identifier}</span>
                    <p className="text-xs text-muted-foreground">{alert.test.title}</p>
                  </div>
                ) : <span className="text-muted-foreground">—</span>}
              </div>

              <Separator />

              <div>
                <p className="text-muted-foreground text-xs mb-1">Assigned To</p>
                {alert.assigned_to ? (
                  <div className="flex items-center gap-2">
                    <User className="h-3 w-3" />
                    <span>{alert.assigned_to.name}</span>
                  </div>
                ) : <span className="text-muted-foreground">Unassigned</span>}
                {alert.assigned_at && (
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {new Date(alert.assigned_at).toLocaleString()}
                    {alert.assigned_by && ` by ${alert.assigned_by.name}`}
                  </p>
                )}
              </div>

              <Separator />

              <div>
                <p className="text-muted-foreground text-xs mb-1">SLA</p>
                {alert.sla_deadline ? (
                  <div className={cn(
                    'text-sm font-medium',
                    alert.sla_breached ? 'text-red-600 dark:text-red-400' : 'text-muted-foreground'
                  )}>
                    <Clock className="h-3 w-3 inline mr-1" />
                    {alert.sla_breached ? 'BREACHED — ' : ''}
                    {new Date(alert.sla_deadline).toLocaleString()}
                    {alert.hours_remaining != null && (
                      <p className="text-xs mt-0.5">
                        {alert.hours_remaining < 0
                          ? `${Math.abs(alert.hours_remaining).toFixed(1)}h overdue`
                          : `${alert.hours_remaining.toFixed(1)}h remaining`}
                      </p>
                    )}
                  </div>
                ) : <span className="text-muted-foreground">No SLA</span>}
              </div>

              <Separator />

              <div>
                <p className="text-muted-foreground text-xs mb-1">Alert Rule</p>
                {alert.alert_rule ? (
                  <Link href="/alert-rules" className="text-xs hover:text-primary">
                    {alert.alert_rule.name}
                  </Link>
                ) : <span className="text-muted-foreground">—</span>}
              </div>

              <Separator />

              <div>
                <p className="text-muted-foreground text-xs mb-1">Delivery</p>
                <div className="flex flex-wrap gap-1">
                  {alert.delivery_channels?.map((ch) => (
                    <Badge key={ch} variant="secondary" className="text-[10px]">
                      {ch}
                      {alert.delivered_at?.[ch] && (
                        <CheckCircle2 className="h-2.5 w-2.5 ml-0.5 text-green-500" />
                      )}
                    </Badge>
                  ))}
                  {(!alert.delivery_channels || alert.delivery_channels.length === 0) && (
                    <span className="text-muted-foreground text-xs">None</span>
                  )}
                </div>
              </div>

              <Separator />

              <div>
                <p className="text-muted-foreground text-xs mb-1">Created</p>
                <span className="text-xs">{new Date(alert.created_at).toLocaleString()}</span>
              </div>
              <div>
                <p className="text-muted-foreground text-xs mb-1">Updated</p>
                <span className="text-xs">{new Date(alert.updated_at).toLocaleString()}</span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Resolve Dialog */}
      <Dialog open={showResolve} onOpenChange={setShowResolve}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Resolve Alert #{alert.alert_number}</DialogTitle>
            <DialogDescription>
              Describe what was done to fix this issue
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Resolution Notes *</Label>
              <textarea
                className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                placeholder="Describe the resolution..."
                value={resolutionNotes}
                onChange={(e) => setResolutionNotes(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowResolve(false)}>Cancel</Button>
            <Button onClick={handleResolve} disabled={actionLoading || !resolutionNotes.trim()}>
              {actionLoading ? 'Resolving...' : 'Resolve Alert'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Suppress Dialog */}
      <Dialog open={showSuppress} onOpenChange={setShowSuppress}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Suppress Alert #{alert.alert_number}</DialogTitle>
            <DialogDescription>
              Snooze this alert with mandatory justification (max 90 days)
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Suppress Until *</Label>
              <Input
                type="datetime-local"
                value={suppressUntil}
                onChange={(e) => setSuppressUntil(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Reason * (min 20 chars)</Label>
              <textarea
                className="flex min-h-[100px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                placeholder="Explain why this alert is being suppressed..."
                value={suppressReason}
                onChange={(e) => setSuppressReason(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowSuppress(false)}>Cancel</Button>
            <Button onClick={handleSuppress} disabled={actionLoading || !suppressUntil || suppressReason.length < 20}>
              {actionLoading ? 'Suppressing...' : 'Suppress Alert'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Assign Dialog */}
      <Dialog open={showAssign} onOpenChange={setShowAssign}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Assign Alert #{alert.alert_number}</DialogTitle>
            <DialogDescription>
              Enter the user ID to assign this alert to
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>User ID *</Label>
              <Input
                placeholder="Enter user UUID"
                value={assignUserId}
                onChange={(e) => setAssignUserId(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAssign(false)}>Cancel</Button>
            <Button onClick={handleAssign} disabled={actionLoading || !assignUserId.trim()}>
              {actionLoading ? 'Assigning...' : 'Assign Alert'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Close Dialog */}
      <Dialog open={showClose} onOpenChange={setShowClose}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Close Alert #{alert.alert_number}</DialogTitle>
            <DialogDescription>
              Mark as verified fixed or accepted as risk
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Notes (optional)</Label>
              <textarea
                className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                placeholder="Optional resolution notes..."
                value={closeNotes}
                onChange={(e) => setCloseNotes(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowClose(false)}>Cancel</Button>
            <Button onClick={handleClose} disabled={actionLoading}>
              {actionLoading ? 'Closing...' : 'Close Alert'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
