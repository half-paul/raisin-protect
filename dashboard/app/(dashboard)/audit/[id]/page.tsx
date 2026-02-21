'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Progress } from '@/components/ui/progress';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  ArrowLeft, Clock, AlertTriangle, CheckCircle, XCircle, FileText,
  MessageSquare, ClipboardCheck, BarChart3, Calendar, Building2, Users,
} from 'lucide-react';
import {
  Audit, AuditStats, AuditRequest, AuditFinding, AuditComment,
  getAudit, getAuditStats, listAuditRequests, listAuditFindings, listAuditComments,
  transitionAuditStatus,
} from '@/lib/api';
import {
  AUDIT_STATUS_LABELS, AUDIT_STATUS_COLORS, AUDIT_TYPE_LABELS,
  REQUEST_STATUS_LABELS, REQUEST_STATUS_COLORS, REQUEST_PRIORITY_COLORS,
  FINDING_SEVERITY_LABELS, FINDING_SEVERITY_COLORS, FINDING_STATUS_LABELS, FINDING_STATUS_COLORS,
} from '@/components/audit/constants';

const VALID_TRANSITIONS: Record<string, string[]> = {
  planning: ['fieldwork', 'cancelled'],
  fieldwork: ['review', 'cancelled'],
  review: ['draft_report', 'fieldwork', 'cancelled'],
  draft_report: ['management_response', 'cancelled'],
  management_response: ['final_report', 'draft_report'],
  final_report: ['completed'],
};

export default function AuditDetailPage() {
  const params = useParams();
  const auditId = params.id as string;
  const { hasRole } = useAuth();
  const canManage = hasRole('ciso', 'compliance_manager');

  const [audit, setAudit] = useState<Audit | null>(null);
  const [stats, setStats] = useState<AuditStats | null>(null);
  const [requests, setRequests] = useState<AuditRequest[]>([]);
  const [findings, setFindings] = useState<AuditFinding[]>([]);
  const [comments, setComments] = useState<AuditComment[]>([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState('overview');

  // Status transition dialog
  const [transitionOpen, setTransitionOpen] = useState(false);
  const [transitionTarget, setTransitionTarget] = useState('');
  const [transitionNotes, setTransitionNotes] = useState('');
  const [transitioning, setTransitioning] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [auditRes, statsRes, reqRes, findRes, comRes] = await Promise.all([
        getAudit(auditId),
        getAuditStats(auditId).catch(() => ({ data: null })),
        listAuditRequests(auditId, { per_page: '100' }).catch(() => ({ data: [] })),
        listAuditFindings(auditId, { per_page: '100' }).catch(() => ({ data: [] })),
        listAuditComments(auditId, { per_page: '100' }).catch(() => ({ data: [] })),
      ]);
      setAudit(auditRes.data);
      if (statsRes.data) setStats(statsRes.data);
      setRequests(reqRes.data);
      setFindings(findRes.data);
      setComments(comRes.data);
    } catch (err) {
      console.error('Failed to fetch audit:', err);
    } finally {
      setLoading(false);
    }
  }, [auditId]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleTransition = async () => {
    if (!transitionTarget) return;
    try {
      setTransitioning(true);
      await transitionAuditStatus(auditId, { status: transitionTarget, notes: transitionNotes || undefined });
      setTransitionOpen(false);
      setTransitionTarget('');
      setTransitionNotes('');
      fetchData();
    } catch (err) {
      console.error('Transition failed:', err);
      alert(err instanceof Error ? err.message : 'Transition failed');
    } finally {
      setTransitioning(false);
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center py-16 text-muted-foreground">Loading audit...</div>;
  }

  if (!audit) {
    return <div className="flex items-center justify-center py-16 text-muted-foreground">Audit not found</div>;
  }

  const availableTransitions = VALID_TRANSITIONS[audit.status] || [];
  const rd = stats?.readiness;
  const fd = stats?.findings;
  const tl = stats?.timeline;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/audit">
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{audit.title}</h1>
            <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${AUDIT_STATUS_COLORS[audit.status]}`}>
              {AUDIT_STATUS_LABELS[audit.status]}
            </span>
          </div>
          <div className="flex items-center gap-4 text-sm text-muted-foreground mt-1">
            <span className="flex items-center gap-1"><FileText className="h-3 w-3" />{AUDIT_TYPE_LABELS[audit.audit_type]}</span>
            {audit.audit_firm && <span className="flex items-center gap-1"><Building2 className="h-3 w-3" />{audit.audit_firm}</span>}
            {audit.lead_auditor_name && <span className="flex items-center gap-1"><Users className="h-3 w-3" />Lead: {audit.lead_auditor_name}</span>}
          </div>
        </div>
        {canManage && availableTransitions.length > 0 && (
          <div className="flex items-center gap-2">
            {availableTransitions.filter(t => t !== 'cancelled').map((t) => (
              <Button
                key={t}
                variant="outline"
                size="sm"
                onClick={() => { setTransitionTarget(t); setTransitionOpen(true); }}
              >
                {AUDIT_STATUS_LABELS[t]}
              </Button>
            ))}
            {availableTransitions.includes('cancelled') && (
              <Button
                variant="destructive"
                size="sm"
                onClick={() => { setTransitionTarget('cancelled'); setTransitionOpen(true); }}
              >
                Cancel Audit
              </Button>
            )}
          </div>
        )}
      </div>

      {/* Stats Overview */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-2xl font-bold">{rd?.readiness_pct ?? 0}%</div>
                <p className="text-xs text-muted-foreground">Evidence Readiness</p>
              </div>
              <BarChart3 className="h-8 w-8 text-muted-foreground/50" />
            </div>
            {rd && <Progress value={rd.readiness_pct} className="mt-2" />}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold">{rd?.open ?? 0 + (rd?.in_progress ?? 0)}</div>
            <p className="text-xs text-muted-foreground">Open Requests ({rd?.total_requests ?? 0} total)</p>
            {(rd?.overdue ?? 0) > 0 && <p className="text-xs text-red-500 mt-1">{rd?.overdue} overdue</p>}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold">{fd?.total ?? 0}</div>
            <p className="text-xs text-muted-foreground">Findings</p>
            <div className="flex gap-2 mt-1">
              {(fd?.by_severity?.critical ?? 0) > 0 && <Badge variant="destructive" className="text-[10px]">{fd?.by_severity?.critical} Critical</Badge>}
              {(fd?.by_severity?.high ?? 0) > 0 && <Badge variant="default" className="text-[10px] bg-orange-500">{fd?.by_severity?.high} High</Badge>}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            {tl ? (
              <>
                <div className="text-2xl font-bold">{tl.days_remaining ?? '—'}</div>
                <p className="text-xs text-muted-foreground">Days Remaining</p>
                <p className="text-xs text-muted-foreground mt-1">
                  {tl.milestones_completed}/{tl.milestones_total} milestones complete
                </p>
                {tl.next_milestone && (
                  <p className="text-xs mt-1">Next: {tl.next_milestone.name} ({tl.next_milestone.days_until}d)</p>
                )}
              </>
            ) : (
              <>
                <div className="text-2xl font-bold">—</div>
                <p className="text-xs text-muted-foreground">Timeline</p>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="requests">Requests ({requests.length})</TabsTrigger>
          <TabsTrigger value="findings">Findings ({findings.length})</TabsTrigger>
          <TabsTrigger value="comments">Comments ({comments.length})</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            {/* Audit Details */}
            <Card>
              <CardHeader><CardTitle className="text-base">Engagement Details</CardTitle></CardHeader>
              <CardContent className="space-y-3 text-sm">
                {audit.description && <p>{audit.description}</p>}
                <div className="grid grid-cols-2 gap-3">
                  <div><span className="text-muted-foreground">Period:</span> {audit.period_start ? `${new Date(audit.period_start).toLocaleDateString()} — ${audit.period_end ? new Date(audit.period_end).toLocaleDateString() : 'TBD'}` : '—'}</div>
                  <div><span className="text-muted-foreground">Planned:</span> {audit.planned_start ? `${new Date(audit.planned_start).toLocaleDateString()} — ${audit.planned_end ? new Date(audit.planned_end).toLocaleDateString() : 'TBD'}` : '—'}</div>
                  {audit.actual_start && <div><span className="text-muted-foreground">Actual Start:</span> {new Date(audit.actual_start).toLocaleDateString()}</div>}
                  {audit.actual_end && <div><span className="text-muted-foreground">Actual End:</span> {new Date(audit.actual_end).toLocaleDateString()}</div>}
                  {audit.internal_lead_name && <div><span className="text-muted-foreground">Internal Lead:</span> {audit.internal_lead_name}</div>}
                  {audit.report_type && <div><span className="text-muted-foreground">Report Type:</span> {audit.report_type}</div>}
                </div>
                {audit.tags.length > 0 && (
                  <div className="flex gap-1 flex-wrap pt-2">
                    {audit.tags.map(t => <Badge key={t} variant="secondary" className="text-xs">{t}</Badge>)}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Milestones */}
            <Card>
              <CardHeader><CardTitle className="text-base">Milestones</CardTitle></CardHeader>
              <CardContent>
                {audit.milestones && audit.milestones.length > 0 ? (
                  <div className="space-y-3">
                    {audit.milestones.map((m, i) => (
                      <div key={i} className="flex items-center gap-3">
                        {m.completed_at ? (
                          <CheckCircle className="h-4 w-4 text-green-500 shrink-0" />
                        ) : (
                          <Clock className="h-4 w-4 text-muted-foreground shrink-0" />
                        )}
                        <div className="flex-1 min-w-0">
                          <p className={`text-sm font-medium ${m.completed_at ? 'line-through text-muted-foreground' : ''}`}>{m.name}</p>
                          <p className="text-xs text-muted-foreground">
                            Target: {new Date(m.target_date).toLocaleDateString()}
                            {m.completed_at && ` · Completed: ${new Date(m.completed_at).toLocaleDateString()}`}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">No milestones defined</p>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Evidence Readiness by Status */}
          {rd && (
            <Card>
              <CardHeader><CardTitle className="text-base">Evidence Request Breakdown</CardTitle></CardHeader>
              <CardContent>
                <div className="grid grid-cols-6 gap-4 text-center">
                  {[
                    { label: 'Open', value: rd.open, color: 'text-gray-600' },
                    { label: 'In Progress', value: rd.in_progress, color: 'text-blue-600' },
                    { label: 'Submitted', value: rd.submitted, color: 'text-purple-600' },
                    { label: 'Accepted', value: rd.accepted, color: 'text-green-600' },
                    { label: 'Rejected', value: rd.rejected, color: 'text-red-600' },
                    { label: 'Overdue', value: rd.overdue, color: 'text-orange-600' },
                  ].map((s) => (
                    <div key={s.label}>
                      <div className={`text-xl font-bold ${s.color}`}>{s.value}</div>
                      <p className="text-xs text-muted-foreground">{s.label}</p>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* Requests Tab */}
        <TabsContent value="requests" className="space-y-4">
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Ref</TableHead>
                    <TableHead>Title</TableHead>
                    <TableHead>Priority</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Assigned To</TableHead>
                    <TableHead>Due Date</TableHead>
                    <TableHead>Evidence</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {requests.length === 0 ? (
                    <TableRow><TableCell colSpan={7} className="text-center py-8 text-muted-foreground">No evidence requests yet</TableCell></TableRow>
                  ) : (
                    requests.map((req) => {
                      const isOverdue = req.due_date && !['accepted', 'closed'].includes(req.status) && new Date(req.due_date) < new Date();
                      return (
                        <TableRow key={req.id} className={isOverdue ? 'bg-red-50/50 dark:bg-red-950/10' : ''}>
                          <TableCell className="font-mono text-xs">{req.reference_number || '—'}</TableCell>
                          <TableCell>
                            <Link href={`/audit/${auditId}/requests/${req.id}`} className="text-primary hover:underline font-medium text-sm">
                              {req.title}
                            </Link>
                          </TableCell>
                          <TableCell>
                            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${REQUEST_PRIORITY_COLORS[req.priority]}`}>
                              {req.priority}
                            </span>
                          </TableCell>
                          <TableCell>
                            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${REQUEST_STATUS_COLORS[req.status]}`}>
                              {REQUEST_STATUS_LABELS[req.status]}
                            </span>
                          </TableCell>
                          <TableCell className="text-sm">{req.assigned_to_name || '—'}</TableCell>
                          <TableCell className="text-sm">
                            {req.due_date ? new Date(req.due_date).toLocaleDateString() : '—'}
                            {isOverdue && <AlertTriangle className="inline ml-1 h-3 w-3 text-red-500" />}
                          </TableCell>
                          <TableCell className="text-sm">{req.evidence_count}</TableCell>
                        </TableRow>
                      );
                    })
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Findings Tab */}
        <TabsContent value="findings" className="space-y-4">
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Ref</TableHead>
                    <TableHead>Title</TableHead>
                    <TableHead>Severity</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Owner</TableHead>
                    <TableHead>Remediation Due</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {findings.length === 0 ? (
                    <TableRow><TableCell colSpan={6} className="text-center py-8 text-muted-foreground">No findings yet</TableCell></TableRow>
                  ) : (
                    findings.map((f) => (
                      <TableRow key={f.id}>
                        <TableCell className="font-mono text-xs">{f.reference_number || '—'}</TableCell>
                        <TableCell>
                          <Link href={`/audit/${auditId}/findings/${f.id}`} className="text-primary hover:underline font-medium text-sm">
                            {f.title}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${FINDING_SEVERITY_COLORS[f.severity]}`}>
                            {FINDING_SEVERITY_LABELS[f.severity]}
                          </span>
                        </TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${FINDING_STATUS_COLORS[f.status]}`}>
                            {FINDING_STATUS_LABELS[f.status]}
                          </span>
                        </TableCell>
                        <TableCell className="text-sm">{f.remediation_owner_name || '—'}</TableCell>
                        <TableCell className="text-sm">{f.remediation_due_date ? new Date(f.remediation_due_date).toLocaleDateString() : '—'}</TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Comments Tab */}
        <TabsContent value="comments" className="space-y-4">
          <Card>
            <CardContent className="pt-6">
              {comments.length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-8">No comments yet</p>
              ) : (
                <div className="space-y-4">
                  {comments.filter(c => !c.parent_comment_id).map((comment) => (
                    <div key={comment.id} className="border rounded-lg p-4">
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-sm">{comment.author_name}</span>
                          <Badge variant="outline" className="text-[10px]">{comment.author_role}</Badge>
                          {comment.is_internal && <Badge variant="secondary" className="text-[10px]">Internal</Badge>}
                        </div>
                        <span className="text-xs text-muted-foreground">{new Date(comment.created_at).toLocaleString()}</span>
                      </div>
                      <p className="text-sm">{comment.body}</p>
                      {/* Replies */}
                      {comment.replies && comment.replies.length > 0 && (
                        <div className="ml-6 mt-3 space-y-3 border-l-2 pl-4">
                          {comment.replies.map((reply) => (
                            <div key={reply.id}>
                              <div className="flex items-center gap-2 mb-1">
                                <span className="font-medium text-xs">{reply.author_name}</span>
                                <Badge variant="outline" className="text-[10px]">{reply.author_role}</Badge>
                                <span className="text-[10px] text-muted-foreground">{new Date(reply.created_at).toLocaleString()}</span>
                              </div>
                              <p className="text-sm">{reply.body}</p>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Status Transition Dialog */}
      <Dialog open={transitionOpen} onOpenChange={setTransitionOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Transition Audit Status</DialogTitle>
            <DialogDescription>
              Move audit from <strong>{AUDIT_STATUS_LABELS[audit.status]}</strong> to <strong>{AUDIT_STATUS_LABELS[transitionTarget] || transitionTarget}</strong>
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Label>Notes (optional)</Label>
            <Input
              placeholder="Reason for transition..."
              value={transitionNotes}
              onChange={(e) => setTransitionNotes(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTransitionOpen(false)}>Cancel</Button>
            <Button
              onClick={handleTransition}
              disabled={transitioning}
              variant={transitionTarget === 'cancelled' ? 'destructive' : 'default'}
            >
              {transitioning ? 'Transitioning...' : `Move to ${AUDIT_STATUS_LABELS[transitionTarget] || transitionTarget}`}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
