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
import { Textarea } from '@/components/ui/textarea';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  ArrowLeft, AlertTriangle, CheckCircle, Clock, FileText, MessageSquare, ShieldCheck,
} from 'lucide-react';
import {
  AuditFinding, AuditComment,
  getAuditFinding, transitionFindingStatus, submitManagementResponse,
  listAuditComments, createAuditComment,
} from '@/lib/api';
import {
  FINDING_SEVERITY_LABELS, FINDING_SEVERITY_COLORS,
  FINDING_STATUS_LABELS, FINDING_STATUS_COLORS,
  FINDING_CATEGORY_LABELS,
} from '@/components/audit/constants';

export default function FindingDetailPage() {
  const params = useParams();
  const auditId = params.id as string;
  const findingId = params.fid as string;
  const { hasRole } = useAuth();

  const [finding, setFinding] = useState<AuditFinding | null>(null);
  const [comments, setComments] = useState<AuditComment[]>([]);
  const [loading, setLoading] = useState(true);

  // Management response
  const [mgmtOpen, setMgmtOpen] = useState(false);
  const [mgmtResponse, setMgmtResponse] = useState('');
  const [mgmtSaving, setMgmtSaving] = useState(false);

  // Status transition
  const [transOpen, setTransOpen] = useState(false);
  const [transStatus, setTransStatus] = useState('');
  const [transNotes, setTransNotes] = useState('');
  const [transPlan, setTransPlan] = useState('');
  const [transitioning, setTransitioning] = useState(false);

  // Comments
  const [commentBody, setCommentBody] = useState('');
  const [commentInternal, setCommentInternal] = useState(false);
  const [commenting, setCommenting] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [findRes, comRes] = await Promise.all([
        getAuditFinding(auditId, findingId),
        listAuditComments(auditId, { target_type: 'finding', target_id: findingId }).catch(() => ({ data: [] })),
      ]);
      setFinding(findRes.data);
      setComments(comRes.data);
    } catch (err) {
      console.error('Failed to load finding:', err);
    } finally {
      setLoading(false);
    }
  }, [auditId, findingId]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleMgmtResponse = async () => {
    if (!mgmtResponse.trim()) return;
    try {
      setMgmtSaving(true);
      await submitManagementResponse(auditId, findingId, { management_response: mgmtResponse });
      setMgmtOpen(false);
      setMgmtResponse('');
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to submit response');
    } finally {
      setMgmtSaving(false);
    }
  };

  const handleTransition = async () => {
    if (!transStatus) return;
    try {
      setTransitioning(true);
      const body: Record<string, string | undefined> = { status: transStatus };
      if (transNotes) body.notes = transNotes;
      if (transPlan) body.remediation_plan = transPlan;
      await transitionFindingStatus(auditId, findingId, body as Parameters<typeof transitionFindingStatus>[2]);
      setTransOpen(false);
      setTransStatus('');
      setTransNotes('');
      setTransPlan('');
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Transition failed');
    } finally {
      setTransitioning(false);
    }
  };

  const handleComment = async () => {
    if (!commentBody.trim()) return;
    try {
      setCommenting(true);
      await createAuditComment(auditId, { target_type: 'finding', target_id: findingId, body: commentBody, is_internal: commentInternal });
      setCommentBody('');
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to add comment');
    } finally {
      setCommenting(false);
    }
  };

  if (loading) return <div className="flex items-center justify-center py-16 text-muted-foreground">Loading...</div>;
  if (!finding) return <div className="flex items-center justify-center py-16 text-muted-foreground">Finding not found</div>;

  const nextStatuses: Record<string, string[]> = {
    identified: ['acknowledged'],
    acknowledged: ['remediation_planned'],
    remediation_planned: ['remediation_in_progress'],
    remediation_in_progress: ['remediation_complete'],
    remediation_complete: ['verified', 'remediation_in_progress'],
    verified: ['closed'],
    risk_accepted: ['closed'],
  };
  const available = nextStatuses[finding.status] || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href={`/audit/${auditId}/findings`}>
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-bold">{finding.title}</h1>
            <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${FINDING_SEVERITY_COLORS[finding.severity]}`}>
              {FINDING_SEVERITY_LABELS[finding.severity]}
            </span>
            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${FINDING_STATUS_COLORS[finding.status]}`}>
              {FINDING_STATUS_LABELS[finding.status]}
            </span>
          </div>
          <div className="flex items-center gap-3 text-sm text-muted-foreground mt-1">
            {finding.reference_number && <span className="font-mono">{finding.reference_number}</span>}
            <Badge variant="outline" className="text-xs">{FINDING_CATEGORY_LABELS[finding.category] || finding.category}</Badge>
            {finding.found_by_name && <span>Found by: {finding.found_by_name}</span>}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {hasRole('ciso', 'compliance_manager') && (
            <Button variant="outline" size="sm" onClick={() => { setMgmtResponse(finding.management_response || ''); setMgmtOpen(true); }}>
              <FileText className="h-4 w-4 mr-1" /> Mgmt Response
            </Button>
          )}
          {available.length > 0 && (
            <Select onValueChange={(v) => { setTransStatus(v); setTransOpen(true); }}>
              <SelectTrigger className="h-8 w-auto text-xs"><SelectValue placeholder="Advance →" /></SelectTrigger>
              <SelectContent>
                {available.map(s => (<SelectItem key={s} value={s}>{FINDING_STATUS_LABELS[s]}</SelectItem>))}
              </SelectContent>
            </Select>
          )}
        </div>
      </div>

      {/* Details */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader><CardTitle className="text-base">Finding Details</CardTitle></CardHeader>
          <CardContent className="space-y-3 text-sm">
            <p>{finding.description}</p>
            {finding.recommendation && (
              <div className="p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg">
                <p className="text-xs font-semibold text-blue-700 dark:text-blue-300 mb-1">Recommendation</p>
                <p className="text-sm">{finding.recommendation}</p>
              </div>
            )}
            {finding.management_response && (
              <div className="p-3 bg-green-50 dark:bg-green-950/20 rounded-lg">
                <p className="text-xs font-semibold text-green-700 dark:text-green-300 mb-1">Management Response</p>
                <p className="text-sm">{finding.management_response}</p>
              </div>
            )}
            {finding.risk_accepted && finding.risk_acceptance_reason && (
              <div className="p-3 bg-amber-50 dark:bg-amber-950/20 rounded-lg">
                <p className="text-xs font-semibold text-amber-700 dark:text-amber-300 mb-1">Risk Acceptance Reason</p>
                <p className="text-sm">{finding.risk_acceptance_reason}</p>
              </div>
            )}
            <div className="grid grid-cols-2 gap-3">
              {finding.control_title && <div><span className="text-muted-foreground">Control:</span> {finding.control_title}</div>}
              {finding.requirement_title && <div><span className="text-muted-foreground">Requirement:</span> {finding.requirement_title}</div>}
            </div>
            {finding.tags.length > 0 && (
              <div className="flex gap-1 flex-wrap">{finding.tags.map(t => <Badge key={t} variant="secondary" className="text-xs">{t}</Badge>)}</div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle className="text-base">Remediation Tracking</CardTitle></CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-3">
              <div><span className="text-muted-foreground">Owner:</span> {finding.remediation_owner_name || 'Unassigned'}</div>
              <div><span className="text-muted-foreground">Due Date:</span> {finding.remediation_due_date ? new Date(finding.remediation_due_date).toLocaleDateString() : '—'}</div>
              <div><span className="text-muted-foreground">Started:</span> {finding.remediation_started_at ? new Date(finding.remediation_started_at).toLocaleString() : '—'}</div>
              <div><span className="text-muted-foreground">Completed:</span> {finding.remediation_completed_at ? new Date(finding.remediation_completed_at).toLocaleString() : '—'}</div>
              {finding.verified_at && <div><span className="text-muted-foreground">Verified:</span> {new Date(finding.verified_at).toLocaleString()}</div>}
              {finding.verified_by_name && <div><span className="text-muted-foreground">Verified By:</span> {finding.verified_by_name}</div>}
            </div>
            {finding.remediation_plan && (
              <div className="p-3 bg-muted rounded-lg">
                <p className="text-xs font-semibold mb-1">Remediation Plan</p>
                <p className="text-sm">{finding.remediation_plan}</p>
              </div>
            )}
            {finding.verification_notes && (
              <div className="p-3 bg-muted rounded-lg">
                <p className="text-xs font-semibold mb-1">Verification Notes</p>
                <p className="text-sm">{finding.verification_notes}</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Comments */}
      <Card>
        <CardHeader><CardTitle className="text-base flex items-center gap-2"><MessageSquare className="h-4 w-4" /> Discussion</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          {comments.length > 0 && (
            <div className="space-y-3">
              {comments.filter(c => !c.parent_comment_id).map((c) => (
                <div key={c.id} className={`border rounded-lg p-3 ${c.is_internal ? 'border-yellow-300 bg-yellow-50/50 dark:bg-yellow-950/10' : ''}`}>
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-medium text-sm">{c.author_name}</span>
                    <Badge variant="outline" className="text-[10px]">{c.author_role}</Badge>
                    {c.is_internal && <Badge variant="secondary" className="text-[10px]">Internal</Badge>}
                    <span className="text-[10px] text-muted-foreground">{new Date(c.created_at).toLocaleString()}</span>
                  </div>
                  <p className="text-sm">{c.body}</p>
                  {c.replies && c.replies.length > 0 && (
                    <div className="ml-4 mt-2 space-y-2 border-l-2 pl-3">
                      {c.replies.map((r) => (
                        <div key={r.id}>
                          <div className="flex items-center gap-2 mb-0.5">
                            <span className="font-medium text-xs">{r.author_name}</span>
                            <span className="text-[10px] text-muted-foreground">{new Date(r.created_at).toLocaleString()}</span>
                          </div>
                          <p className="text-sm">{r.body}</p>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
          <div className="flex gap-2">
            <div className="flex-1">
              <Textarea value={commentBody} onChange={(e) => setCommentBody(e.target.value)} rows={2} placeholder="Add a comment..." />
            </div>
            <div className="flex flex-col gap-2">
              {!hasRole('auditor') && (
                <label className="flex items-center gap-1 text-xs cursor-pointer">
                  <input type="checkbox" checked={commentInternal} onChange={(e) => setCommentInternal(e.target.checked)} className="rounded" />
                  Internal
                </label>
              )}
              <Button size="sm" onClick={handleComment} disabled={commenting || !commentBody.trim()}>
                {commenting ? '...' : 'Send'}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Management Response Dialog */}
      <Dialog open={mgmtOpen} onOpenChange={setMgmtOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Management Response</DialogTitle>
            <DialogDescription>Submit the organization&apos;s formal response to this finding.</DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Textarea value={mgmtResponse} onChange={(e) => setMgmtResponse(e.target.value)} rows={5} placeholder="Management agrees with the finding..." />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setMgmtOpen(false)}>Cancel</Button>
            <Button onClick={handleMgmtResponse} disabled={mgmtSaving || !mgmtResponse.trim()}>
              {mgmtSaving ? 'Saving...' : 'Submit Response'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Status Transition Dialog */}
      <Dialog open={transOpen} onOpenChange={setTransOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update Finding Status</DialogTitle>
            <DialogDescription>Move to <strong>{FINDING_STATUS_LABELS[transStatus]}</strong></DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {transStatus === 'remediation_planned' && (
              <div>
                <Label>Remediation Plan *</Label>
                <Textarea value={transPlan} onChange={(e) => setTransPlan(e.target.value)} rows={3} placeholder="Describe the plan..." />
              </div>
            )}
            <div>
              <Label>Notes</Label>
              <Input value={transNotes} onChange={(e) => setTransNotes(e.target.value)} placeholder="Optional..." />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTransOpen(false)}>Cancel</Button>
            <Button onClick={handleTransition} disabled={transitioning}>{transitioning ? 'Updating...' : 'Update'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
