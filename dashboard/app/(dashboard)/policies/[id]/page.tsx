'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import DOMPurify from 'dompurify';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
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
  ArrowLeft, FileText, Edit, Send, CheckCircle2, XCircle, Clock, Shield,
  Link2, History, Users, Archive, BookOpen, AlertTriangle, Trash2, Plus,
} from 'lucide-react';
import {
  Policy, PolicyVersion, PolicySignoff, PolicyControlLink, Control,
  getPolicy, updatePolicy, archivePolicy, submitPolicyForReview, publishPolicy,
  listPolicyVersions, listPolicySignoffs, listPolicyControls,
  linkPolicyControl, unlinkPolicyControl, listControls,
  approvePolicySignoff, rejectPolicySignoff, withdrawPolicySignoff,
} from '@/lib/api';

const STATUS_LABELS: Record<string, string> = {
  draft: 'Draft', in_review: 'In Review', approved: 'Approved',
  published: 'Published', archived: 'Archived',
};
const STATUS_COLORS: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300',
  in_review: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  approved: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  published: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  archived: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
};
const CATEGORY_LABELS: Record<string, string> = {
  information_security: 'Information Security', access_control: 'Access Control',
  incident_response: 'Incident Response', data_privacy: 'Data Privacy',
  network_security: 'Network Security', encryption: 'Encryption',
  vulnerability_management: 'Vulnerability Management', change_management: 'Change Management',
  business_continuity: 'Business Continuity', secure_development: 'Secure Development',
  vendor_management: 'Vendor Management', acceptable_use: 'Acceptable Use',
  physical_security: 'Physical Security', hr_security: 'HR Security',
  asset_management: 'Asset Management',
};
const SIGNOFF_COLORS: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  approved: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  rejected: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  withdrawn: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400',
};

export default function PolicyDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { user, hasRole } = useAuth();
  const canEdit = hasRole('ciso', 'compliance_manager', 'security_engineer');
  const canPublish = hasRole('ciso', 'compliance_manager');
  const canArchive = hasRole('ciso', 'compliance_manager');

  const [policy, setPolicy] = useState<Policy | null>(null);
  const [versions, setVersions] = useState<PolicyVersion[]>([]);
  const [signoffs, setSignoffs] = useState<PolicySignoff[]>([]);
  const [linkedControls, setLinkedControls] = useState<PolicyControlLink[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('content');

  // Submit for review dialog
  const [showSubmit, setShowSubmit] = useState(false);
  const [submitForm, setSubmitForm] = useState({ signer_ids: '', due_date: '', message: '' });
  const [submitting, setSubmitting] = useState(false);

  // Link control dialog
  const [showLinkControl, setShowLinkControl] = useState(false);
  const [availableControls, setAvailableControls] = useState<Control[]>([]);
  const [controlSearch, setControlSearch] = useState('');
  const [selectedControlId, setSelectedControlId] = useState('');
  const [linkCoverage, setLinkCoverage] = useState('full');
  const [linkNotes, setLinkNotes] = useState('');
  const [linking, setLinking] = useState(false);

  // Signoff action
  const [showSignoffAction, setShowSignoffAction] = useState<{ id: string; action: 'approve' | 'reject' } | null>(null);
  const [signoffComments, setSignoffComments] = useState('');
  const [signingOff, setSigningOff] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const [polRes, versRes, sigRes, ctrlRes] = await Promise.all([
        getPolicy(id),
        listPolicyVersions(id),
        listPolicySignoffs(id),
        listPolicyControls(id),
      ]);
      setPolicy(polRes.data);
      setVersions(versRes.data);
      setSignoffs(sigRes.data);
      setLinkedControls(ctrlRes.data);
    } catch (err) {
      console.error('Failed to fetch policy:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleSubmitForReview = async () => {
    if (!submitForm.signer_ids.trim()) return;
    try {
      setSubmitting(true);
      const signerIds = submitForm.signer_ids.split(',').map(s => s.trim()).filter(Boolean);
      await submitPolicyForReview(id, {
        signer_ids: signerIds,
        due_date: submitForm.due_date || undefined,
        message: submitForm.message || undefined,
      });
      setShowSubmit(false);
      setSubmitForm({ signer_ids: '', due_date: '', message: '' });
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to submit');
    } finally {
      setSubmitting(false);
    }
  };

  const handlePublish = async () => {
    if (!confirm('Publish this policy? It will become active and in-effect.')) return;
    try {
      await publishPolicy(id);
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to publish');
    }
  };

  const handleArchive = async () => {
    if (!confirm('Archive this policy?')) return;
    try {
      await archivePolicy(id);
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to archive');
    }
  };

  const handleLinkControl = async () => {
    if (!selectedControlId) return;
    try {
      setLinking(true);
      await linkPolicyControl(id, {
        control_id: selectedControlId,
        coverage: linkCoverage,
        notes: linkNotes || undefined,
      });
      setShowLinkControl(false);
      setSelectedControlId('');
      setLinkNotes('');
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to link');
    } finally {
      setLinking(false);
    }
  };

  const handleUnlinkControl = async (controlId: string) => {
    if (!confirm('Remove this control link?')) return;
    try {
      await unlinkPolicyControl(id, controlId);
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to unlink');
    }
  };

  const handleSignoffAction = async () => {
    if (!showSignoffAction) return;
    try {
      setSigningOff(true);
      if (showSignoffAction.action === 'approve') {
        await approvePolicySignoff(id, showSignoffAction.id, {
          comments: signoffComments || undefined,
        });
      } else {
        if (!signoffComments.trim()) { alert('Comments are required for rejection'); return; }
        await rejectPolicySignoff(id, showSignoffAction.id, { comments: signoffComments });
      }
      setShowSignoffAction(null);
      setSignoffComments('');
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed');
    } finally {
      setSigningOff(false);
    }
  };

  const openLinkDialog = async () => {
    setShowLinkControl(true);
    try {
      const res = await listControls({ per_page: '200', status: 'active' });
      setAvailableControls(res.data);
    } catch (err) {
      console.error('Failed to load controls:', err);
    }
  };

  const filteredControls = availableControls.filter(c =>
    !controlSearch || c.identifier.toLowerCase().includes(controlSearch.toLowerCase()) || c.title.toLowerCase().includes(controlSearch.toLowerCase())
  );

  if (loading) {
    return <div className="flex items-center justify-center py-20 text-muted-foreground">Loading policy...</div>;
  }
  if (!policy) {
    return <div className="flex items-center justify-center py-20 text-muted-foreground">Policy not found</div>;
  }

  const myPendingSignoffs = signoffs.filter(s => s.status === 'pending' && s.signer?.id === user?.id);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push('/policies')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="font-mono text-sm text-muted-foreground">{policy.identifier}</span>
              <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[policy.status]}`}>
                {STATUS_LABELS[policy.status]}
              </span>
              <Badge variant="outline" className="text-xs">{CATEGORY_LABELS[policy.category] || policy.category}</Badge>
            </div>
            <h1 className="text-2xl font-bold">{policy.title}</h1>
            {policy.description && <p className="text-sm text-muted-foreground mt-1">{policy.description}</p>}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {canEdit && policy.status === 'draft' && (
            <Button variant="outline" onClick={() => setShowSubmit(true)}>
              <Send className="h-4 w-4 mr-2" /> Submit for Review
            </Button>
          )}
          {canPublish && policy.status === 'approved' && (
            <Button onClick={handlePublish}>
              <CheckCircle2 className="h-4 w-4 mr-2" /> Publish
            </Button>
          )}
          {canEdit && (
            <Link href={`/policies/${id}/edit`}>
              <Button variant="outline"><Edit className="h-4 w-4 mr-2" /> Edit Content</Button>
            </Link>
          )}
          {canArchive && policy.status !== 'archived' && (
            <Button variant="outline" className="text-red-500" onClick={handleArchive}>
              <Archive className="h-4 w-4 mr-2" /> Archive
            </Button>
          )}
        </div>
      </div>

      {/* Pending signoffs for current user */}
      {myPendingSignoffs.length > 0 && (
        <Card className="border-yellow-500 bg-yellow-50 dark:bg-yellow-950/20">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <AlertTriangle className="h-5 w-5 text-yellow-600" />
                <span className="font-medium">Your sign-off is requested for this policy</span>
              </div>
              <div className="flex gap-2">
                <Button size="sm" variant="outline" className="text-green-600" onClick={() => setShowSignoffAction({ id: myPendingSignoffs[0].id, action: 'approve' })}>
                  <CheckCircle2 className="h-4 w-4 mr-1" /> Approve
                </Button>
                <Button size="sm" variant="outline" className="text-red-600" onClick={() => setShowSignoffAction({ id: myPendingSignoffs[0].id, action: 'reject' })}>
                  <XCircle className="h-4 w-4 mr-1" /> Reject
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Metadata Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Owner</div>
            <div className="font-medium">{policy.owner?.name || 'Unassigned'}</div>
            {policy.secondary_owner && <div className="text-xs text-muted-foreground mt-1">Secondary: {policy.secondary_owner.name}</div>}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Current Version</div>
            <div className="font-medium">v{policy.current_version?.version_number || 1}</div>
            <div className="text-xs text-muted-foreground mt-1">{policy.current_version?.word_count || 0} words</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Linked Controls</div>
            <div className="font-medium">{linkedControls.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Sign-offs</div>
            <div className="font-medium">
              {policy.signoff_summary ? `${policy.signoff_summary.approved}/${policy.signoff_summary.total} approved` : 'None'}
            </div>
            {policy.signoff_summary && policy.signoff_summary.pending > 0 && (
              <div className="text-xs text-yellow-600 mt-1">{policy.signoff_summary.pending} pending</div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="content"><FileText className="h-4 w-4 mr-1" /> Content</TabsTrigger>
          <TabsTrigger value="versions"><History className="h-4 w-4 mr-1" /> Versions ({versions.length})</TabsTrigger>
          <TabsTrigger value="signoffs"><Users className="h-4 w-4 mr-1" /> Sign-offs ({signoffs.length})</TabsTrigger>
          <TabsTrigger value="controls"><Shield className="h-4 w-4 mr-1" /> Controls ({linkedControls.length})</TabsTrigger>
        </TabsList>

        {/* Content Tab */}
        <TabsContent value="content">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Policy Content</CardTitle>
              {policy.current_version?.change_summary && (
                <CardDescription>Last change: {policy.current_version.change_summary}</CardDescription>
              )}
            </CardHeader>
            <CardContent>
              {policy.current_version?.content ? (
                <div
                  className="prose dark:prose-invert max-w-none"
                  dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(policy.current_version.content) }}
                />
              ) : (
                <p className="text-muted-foreground">No content available.</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Versions Tab */}
        <TabsContent value="versions">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="text-lg">Version History</CardTitle>
              {versions.length >= 2 && (
                <Link href={`/policies/${id}/versions`}>
                  <Button variant="outline" size="sm">Compare Versions</Button>
                </Link>
              )}
            </CardHeader>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Version</TableHead>
                    <TableHead>Change</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Words</TableHead>
                    <TableHead>Author</TableHead>
                    <TableHead>Date</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {versions.map((v) => (
                    <TableRow key={v.id}>
                      <TableCell className="font-medium">
                        v{v.version_number}
                        {v.is_current && <Badge className="ml-2 text-xs" variant="secondary">Current</Badge>}
                      </TableCell>
                      <TableCell className="text-sm">{v.change_summary || '—'}</TableCell>
                      <TableCell><Badge variant="outline" className="text-xs">{v.change_type || 'initial'}</Badge></TableCell>
                      <TableCell className="text-sm">{v.word_count || '—'}</TableCell>
                      <TableCell className="text-sm">{v.created_by?.name || '—'}</TableCell>
                      <TableCell className="text-sm">{v.created_at ? new Date(v.created_at).toLocaleDateString() : '—'}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Signoffs Tab */}
        <TabsContent value="signoffs">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Sign-off Requests</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {signoffs.length === 0 ? (
                <div className="p-6 text-center text-muted-foreground">No sign-off requests yet</div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Signer</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead>Version</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Due Date</TableHead>
                      <TableHead>Comments</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {signoffs.map((s) => (
                      <TableRow key={s.id}>
                        <TableCell className="font-medium">{s.signer?.name || '—'}</TableCell>
                        <TableCell className="text-sm">{s.signer_role || s.signer?.role || '—'}</TableCell>
                        <TableCell className="text-sm">v{s.policy_version?.version_number || '—'}</TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SIGNOFF_COLORS[s.status]}`}>
                            {s.status}
                          </span>
                        </TableCell>
                        <TableCell className="text-sm">{s.due_date ? new Date(s.due_date).toLocaleDateString() : '—'}</TableCell>
                        <TableCell className="text-sm max-w-[200px] truncate">{s.comments || '—'}</TableCell>
                        <TableCell>
                          {s.status === 'pending' && s.signer?.id === user?.id && (
                            <div className="flex gap-1">
                              <Button size="sm" variant="ghost" className="text-green-600 h-7" onClick={() => setShowSignoffAction({ id: s.id, action: 'approve' })}>
                                Approve
                              </Button>
                              <Button size="sm" variant="ghost" className="text-red-600 h-7" onClick={() => setShowSignoffAction({ id: s.id, action: 'reject' })}>
                                Reject
                              </Button>
                            </div>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Controls Tab */}
        <TabsContent value="controls">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="text-lg">Linked Controls</CardTitle>
              {canEdit && (
                <Button size="sm" onClick={openLinkDialog}>
                  <Plus className="h-4 w-4 mr-1" /> Link Control
                </Button>
              )}
            </CardHeader>
            <CardContent className="p-0">
              {linkedControls.length === 0 ? (
                <div className="p-6 text-center text-muted-foreground">No controls linked to this policy</div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Identifier</TableHead>
                      <TableHead>Title</TableHead>
                      <TableHead>Category</TableHead>
                      <TableHead>Coverage</TableHead>
                      <TableHead>Frameworks</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {linkedControls.map((c) => (
                      <TableRow key={c.id}>
                        <TableCell className="font-mono text-xs">{c.identifier}</TableCell>
                        <TableCell>
                          <Link href={`/controls/${c.id}`} className="text-primary hover:underline">{c.title}</Link>
                        </TableCell>
                        <TableCell><Badge variant="outline" className="text-xs">{c.category}</Badge></TableCell>
                        <TableCell>
                          <Badge variant={c.coverage === 'full' ? 'default' : 'secondary'} className="text-xs">
                            {c.coverage || 'full'}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-xs">{c.frameworks?.join(', ') || '—'}</TableCell>
                        <TableCell>
                          {canEdit && (
                            <Button size="sm" variant="ghost" className="text-red-500 h-7" onClick={() => handleUnlinkControl(c.id)}>
                              <Trash2 className="h-3 w-3" />
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Submit for Review Dialog */}
      <Dialog open={showSubmit} onOpenChange={setShowSubmit}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Submit for Review</DialogTitle>
            <DialogDescription>Request sign-offs from approvers before publishing.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Signer IDs * (comma-separated UUIDs)</Label>
              <Input
                placeholder="uuid-1, uuid-2"
                value={submitForm.signer_ids}
                onChange={(e) => setSubmitForm(f => ({ ...f, signer_ids: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground mt-1">Enter user IDs of required approvers</p>
            </div>
            <div>
              <Label>Due Date</Label>
              <Input type="date" value={submitForm.due_date} onChange={(e) => setSubmitForm(f => ({ ...f, due_date: e.target.value }))} />
            </div>
            <div>
              <Label>Message</Label>
              <Textarea placeholder="Optional message to reviewers..." value={submitForm.message} onChange={(e) => setSubmitForm(f => ({ ...f, message: e.target.value }))} rows={3} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowSubmit(false)}>Cancel</Button>
            <Button onClick={handleSubmitForReview} disabled={submitting || !submitForm.signer_ids.trim()}>
              {submitting ? 'Submitting...' : 'Submit for Review'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Link Control Dialog */}
      <Dialog open={showLinkControl} onOpenChange={setShowLinkControl}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Link Control to Policy</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Search Controls</Label>
              <Input placeholder="Search by identifier or title..." value={controlSearch} onChange={(e) => setControlSearch(e.target.value)} />
            </div>
            <div className="max-h-[200px] overflow-y-auto border rounded-md">
              {filteredControls.slice(0, 20).map(c => (
                <div
                  key={c.id}
                  className={`p-2 cursor-pointer hover:bg-accent text-sm ${selectedControlId === c.id ? 'bg-primary/10' : ''}`}
                  onClick={() => setSelectedControlId(c.id)}
                >
                  <span className="font-mono text-xs mr-2">{c.identifier}</span>
                  {c.title}
                </div>
              ))}
              {filteredControls.length === 0 && <div className="p-4 text-center text-muted-foreground text-sm">No controls found</div>}
            </div>
            <div>
              <Label>Coverage</Label>
              <Select value={linkCoverage} onValueChange={setLinkCoverage}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="full">Full</SelectItem>
                  <SelectItem value="partial">Partial</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Notes</Label>
              <Textarea placeholder="Why does this policy govern this control?" value={linkNotes} onChange={(e) => setLinkNotes(e.target.value)} rows={2} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowLinkControl(false)}>Cancel</Button>
            <Button onClick={handleLinkControl} disabled={linking || !selectedControlId}>
              {linking ? 'Linking...' : 'Link Control'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Signoff Action Dialog */}
      <Dialog open={!!showSignoffAction} onOpenChange={() => setShowSignoffAction(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{showSignoffAction?.action === 'approve' ? 'Approve' : 'Reject'} Sign-off</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Comments {showSignoffAction?.action === 'reject' ? '*' : '(optional)'}</Label>
              <Textarea
                placeholder={showSignoffAction?.action === 'reject' ? 'Explain why this policy needs changes...' : 'Optional approval comments...'}
                value={signoffComments}
                onChange={(e) => setSignoffComments(e.target.value)}
                rows={4}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowSignoffAction(null)}>Cancel</Button>
            <Button
              onClick={handleSignoffAction}
              disabled={signingOff || (showSignoffAction?.action === 'reject' && !signoffComments.trim())}
              variant={showSignoffAction?.action === 'approve' ? 'default' : 'destructive'}
            >
              {signingOff ? 'Processing...' : showSignoffAction?.action === 'approve' ? 'Approve' : 'Reject'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
