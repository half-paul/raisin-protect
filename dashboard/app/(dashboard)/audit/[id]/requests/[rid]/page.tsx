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
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger,
} from '@/components/ui/dialog';
import {
  ArrowLeft, FileText, Upload, CheckCircle, XCircle, Clock, AlertTriangle,
  Link2, Eye, Trash2, MessageSquare,
} from 'lucide-react';
import {
  AuditRequest, AuditEvidenceLink, AuditComment,
  getAuditRequest, listRequestEvidence, submitRequestEvidence,
  reviewRequestEvidence, removeRequestEvidence,
  submitAuditRequest, reviewAuditRequest,
  listAuditComments, createAuditComment,
} from '@/lib/api';
import {
  REQUEST_STATUS_LABELS, REQUEST_STATUS_COLORS,
  REQUEST_PRIORITY_LABELS, REQUEST_PRIORITY_COLORS,
  EVIDENCE_SUBMISSION_STATUS_LABELS, EVIDENCE_SUBMISSION_STATUS_COLORS,
} from '@/components/audit/constants';

export default function AuditRequestDetailPage() {
  const params = useParams();
  const auditId = params.id as string;
  const requestId = params.rid as string;
  const { user, hasRole } = useAuth();
  const canSubmitEvidence = hasRole('ciso', 'compliance_manager', 'security_engineer', 'it_admin');
  const canReviewEvidence = hasRole('auditor');
  const canSubmitRequest = hasRole('ciso', 'compliance_manager', 'security_engineer', 'it_admin');
  const canReviewRequest = hasRole('auditor');

  const [request, setRequest] = useState<AuditRequest | null>(null);
  const [evidence, setEvidence] = useState<AuditEvidenceLink[]>([]);
  const [comments, setComments] = useState<AuditComment[]>([]);
  const [loading, setLoading] = useState(true);

  // Submit evidence dialog
  const [linkOpen, setLinkOpen] = useState(false);
  const [artifactId, setArtifactId] = useState('');
  const [linkNotes, setLinkNotes] = useState('');
  const [linking, setLinking] = useState(false);

  // Review evidence dialog
  const [reviewOpen, setReviewOpen] = useState(false);
  const [reviewTarget, setReviewTarget] = useState<string>('');
  const [reviewDecision, setReviewDecision] = useState('accepted');
  const [reviewNotes, setReviewNotes] = useState('');
  const [reviewing, setReviewing] = useState(false);

  // Request review dialog
  const [reqReviewOpen, setReqReviewOpen] = useState(false);
  const [reqDecision, setReqDecision] = useState('accepted');
  const [reqNotes, setReqNotes] = useState('');
  const [reqReviewing, setReqReviewing] = useState(false);

  // New comment
  const [commentBody, setCommentBody] = useState('');
  const [commentInternal, setCommentInternal] = useState(false);
  const [commenting, setCommenting] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [reqRes, evRes, comRes] = await Promise.all([
        getAuditRequest(auditId, requestId),
        listRequestEvidence(auditId, requestId).catch(() => ({ data: [] })),
        listAuditComments(auditId, { target_type: 'request', target_id: requestId }).catch(() => ({ data: [] })),
      ]);
      setRequest(reqRes.data);
      setEvidence(evRes.data);
      setComments(comRes.data);
    } catch (err) {
      console.error('Failed to fetch request:', err);
    } finally {
      setLoading(false);
    }
  }, [auditId, requestId]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleLinkEvidence = async () => {
    if (!artifactId.trim()) return;
    try {
      setLinking(true);
      await submitRequestEvidence(auditId, requestId, { artifact_id: artifactId, notes: linkNotes || undefined });
      setLinkOpen(false);
      setArtifactId('');
      setLinkNotes('');
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to submit evidence');
    } finally {
      setLinking(false);
    }
  };

  const handleReviewEvidence = async () => {
    if (!reviewTarget) return;
    try {
      setReviewing(true);
      await reviewRequestEvidence(auditId, requestId, reviewTarget, { status: reviewDecision, notes: reviewNotes || undefined });
      setReviewOpen(false);
      setReviewTarget('');
      setReviewDecision('accepted');
      setReviewNotes('');
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Review failed');
    } finally {
      setReviewing(false);
    }
  };

  const handleRemoveEvidence = async (linkId: string) => {
    if (!confirm('Remove this evidence from the request?')) return;
    try {
      await removeRequestEvidence(auditId, requestId, linkId);
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Remove failed');
    }
  };

  const handleSubmitRequest = async () => {
    if (!confirm('Submit this request for auditor review?')) return;
    try {
      await submitAuditRequest(auditId, requestId);
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Submit failed');
    }
  };

  const handleReviewRequest = async () => {
    try {
      setReqReviewing(true);
      await reviewAuditRequest(auditId, requestId, { decision: reqDecision, notes: reqNotes || undefined });
      setReqReviewOpen(false);
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Review failed');
    } finally {
      setReqReviewing(false);
    }
  };

  const handleComment = async () => {
    if (!commentBody.trim()) return;
    try {
      setCommenting(true);
      await createAuditComment(auditId, {
        target_type: 'request',
        target_id: requestId,
        body: commentBody,
        is_internal: commentInternal,
      });
      setCommentBody('');
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to add comment');
    } finally {
      setCommenting(false);
    }
  };

  if (loading) return <div className="flex items-center justify-center py-16 text-muted-foreground">Loading...</div>;
  if (!request) return <div className="flex items-center justify-center py-16 text-muted-foreground">Request not found</div>;

  const isOverdue = request.due_date && !['accepted', 'closed'].includes(request.status) && new Date(request.due_date) < new Date();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href={`/audit/${auditId}/requests`}>
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-bold">{request.title}</h1>
            <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${REQUEST_STATUS_COLORS[request.status]}`}>
              {REQUEST_STATUS_LABELS[request.status]}
            </span>
            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${REQUEST_PRIORITY_COLORS[request.priority]}`}>
              {request.priority}
            </span>
            {isOverdue && <Badge variant="destructive" className="text-xs">Overdue</Badge>}
          </div>
          <div className="flex items-center gap-3 text-sm text-muted-foreground mt-1">
            {request.reference_number && <span className="font-mono">{request.reference_number}</span>}
            {request.assigned_to_name && <span>Assigned: {request.assigned_to_name}</span>}
            {request.due_date && <span>Due: {new Date(request.due_date).toLocaleDateString()}</span>}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {canSubmitEvidence && (
            <Dialog open={linkOpen} onOpenChange={setLinkOpen}>
              <DialogTrigger asChild>
                <Button variant="outline" size="sm"><Link2 className="h-4 w-4 mr-1" /> Link Evidence</Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Submit Evidence</DialogTitle>
                  <DialogDescription>Link an existing evidence artifact to this request.</DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div>
                    <Label>Artifact ID *</Label>
                    <Input value={artifactId} onChange={(e) => setArtifactId(e.target.value)} placeholder="Paste evidence artifact UUID" />
                    <p className="text-xs text-muted-foreground mt-1">Find artifacts in the <Link href="/evidence" className="text-primary hover:underline">Evidence Library</Link></p>
                  </div>
                  <div>
                    <Label>Submission Notes</Label>
                    <Textarea value={linkNotes} onChange={(e) => setLinkNotes(e.target.value)} rows={2} placeholder="What does this evidence demonstrate?" />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setLinkOpen(false)}>Cancel</Button>
                  <Button onClick={handleLinkEvidence} disabled={linking || !artifactId.trim()}>{linking ? 'Linking...' : 'Submit Evidence'}</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          )}
          {canSubmitRequest && ['open', 'in_progress'].includes(request.status) && evidence.length > 0 && (
            <Button size="sm" onClick={handleSubmitRequest}>
              <CheckCircle className="h-4 w-4 mr-1" /> Submit for Review
            </Button>
          )}
          {canReviewRequest && request.status === 'submitted' && (
            <Button size="sm" onClick={() => setReqReviewOpen(true)}>
              <Eye className="h-4 w-4 mr-1" /> Review Request
            </Button>
          )}
        </div>
      </div>

      {/* Request details */}
      <Card>
        <CardHeader><CardTitle className="text-base">Request Details</CardTitle></CardHeader>
        <CardContent className="space-y-3 text-sm">
          <p>{request.description}</p>
          <div className="grid grid-cols-2 gap-3 text-sm">
            {request.control_title && <div><span className="text-muted-foreground">Control:</span> {request.control_title}</div>}
            {request.requirement_title && <div><span className="text-muted-foreground">Requirement:</span> {request.requirement_title}</div>}
            {request.requested_by_name && <div><span className="text-muted-foreground">Requested By:</span> {request.requested_by_name}</div>}
            {request.submitted_at && <div><span className="text-muted-foreground">Submitted:</span> {new Date(request.submitted_at).toLocaleString()}</div>}
            {request.reviewed_at && <div><span className="text-muted-foreground">Reviewed:</span> {new Date(request.reviewed_at).toLocaleString()}</div>}
            {request.reviewer_notes && <div className="col-span-2"><span className="text-muted-foreground">Reviewer Notes:</span> {request.reviewer_notes}</div>}
          </div>
          {request.tags.length > 0 && (
            <div className="flex gap-1 flex-wrap">
              {request.tags.map(t => <Badge key={t} variant="secondary" className="text-xs">{t}</Badge>)}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Chain of Custody — Evidence submissions */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <FileText className="h-4 w-4" /> Evidence Submissions ({evidence.length})
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {evidence.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground text-sm">No evidence submitted yet</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Artifact</TableHead>
                  <TableHead>Submitted By</TableHead>
                  <TableHead>Submitted At</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Reviewer</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {evidence.map((ev) => (
                  <TableRow key={ev.link_id}>
                    <TableCell>
                      <div>
                        <p className="font-medium text-sm">{ev.artifact_title}</p>
                        {ev.file_name && <p className="text-xs text-muted-foreground">{ev.file_name} ({ev.file_size ? `${(ev.file_size / 1024).toFixed(0)} KB` : ''})</p>}
                        {ev.submission_notes && <p className="text-xs text-muted-foreground mt-1">{ev.submission_notes}</p>}
                      </div>
                    </TableCell>
                    <TableCell className="text-sm">{ev.submitted_by_name}</TableCell>
                    <TableCell className="text-sm">{new Date(ev.submitted_at).toLocaleString()}</TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${EVIDENCE_SUBMISSION_STATUS_COLORS[ev.status]}`}>
                        {EVIDENCE_SUBMISSION_STATUS_LABELS[ev.status]}
                      </span>
                    </TableCell>
                    <TableCell className="text-sm">
                      {ev.reviewed_by_name ? (
                        <div>
                          <p>{ev.reviewed_by_name}</p>
                          {ev.review_notes && <p className="text-xs text-muted-foreground">{ev.review_notes}</p>}
                        </div>
                      ) : '—'}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        {canReviewEvidence && ev.status === 'pending_review' && (
                          <Button variant="outline" size="sm" onClick={() => { setReviewTarget(ev.link_id); setReviewOpen(true); }}>
                            Review
                          </Button>
                        )}
                        {canSubmitEvidence && (
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-red-500" onClick={() => handleRemoveEvidence(ev.link_id)}>
                            <Trash2 className="h-3 w-3" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Comments */}
      <Card>
        <CardHeader><CardTitle className="text-base flex items-center gap-2"><MessageSquare className="h-4 w-4" /> Comments</CardTitle></CardHeader>
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
          {/* New comment */}
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

      {/* Evidence Review Dialog */}
      <Dialog open={reviewOpen} onOpenChange={setReviewOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Review Evidence</DialogTitle>
            <DialogDescription>Accept, reject, or request clarification on this evidence submission.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <Label>Decision *</Label>
              <Select value={reviewDecision} onValueChange={setReviewDecision}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="accepted">Accept</SelectItem>
                  <SelectItem value="rejected">Reject</SelectItem>
                  <SelectItem value="needs_clarification">Needs Clarification</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Notes {reviewDecision !== 'accepted' ? '*' : ''}</Label>
              <Textarea value={reviewNotes} onChange={(e) => setReviewNotes(e.target.value)} rows={3} placeholder="Feedback on the evidence..." />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setReviewOpen(false)}>Cancel</Button>
            <Button onClick={handleReviewEvidence} disabled={reviewing}>{reviewing ? 'Reviewing...' : 'Submit Review'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Request Review Dialog */}
      <Dialog open={reqReviewOpen} onOpenChange={setReqReviewOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Review Request</DialogTitle>
            <DialogDescription>Accept or reject this evidence request submission.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <Label>Decision *</Label>
              <Select value={reqDecision} onValueChange={setReqDecision}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="accepted">Accept</SelectItem>
                  <SelectItem value="rejected">Reject</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Notes {reqDecision === 'rejected' ? '*' : ''}</Label>
              <Textarea value={reqNotes} onChange={(e) => setReqNotes(e.target.value)} rows={3} placeholder="Feedback..." />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setReqReviewOpen(false)}>Cancel</Button>
            <Button onClick={handleReviewRequest} disabled={reqReviewing}>{reqReviewing ? 'Reviewing...' : 'Submit Review'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
