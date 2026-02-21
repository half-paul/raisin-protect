'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  ClipboardCheck, Clock, CheckCircle2, XCircle, AlertTriangle,
  ChevronLeft, ChevronRight, Eye, FileText,
} from 'lucide-react';
import {
  PendingSignoff, PolicySignoff,
  getPendingSignoffs, approvePolicySignoff, rejectPolicySignoff,
  listPolicies,
  Policy,
} from '@/lib/api';

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

const URGENCY_COLORS: Record<string, string> = {
  overdue: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  due_soon: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  on_time: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
};

export default function PolicyApprovalsPage() {
  const { user } = useAuth();

  const [pending, setPending] = useState<PendingSignoff[]>([]);
  const [recentPolicies, setRecentPolicies] = useState<Policy[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('pending');

  // Filters
  const [urgencyFilter, setUrgencyFilter] = useState('');

  // Action dialog
  const [actionSignoff, setActionSignoff] = useState<{
    signoff: PendingSignoff;
    action: 'approve' | 'reject';
  } | null>(null);
  const [comments, setComments] = useState('');
  const [processing, setProcessing] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = { per_page: '50' };
      if (urgencyFilter) params.urgency = urgencyFilter;

      const [pendRes, polRes] = await Promise.all([
        getPendingSignoffs(params),
        listPolicies({ status: 'in_review', per_page: '10', order: 'desc', sort: 'updated_at' }),
      ]);
      setPending(pendRes.data);
      setRecentPolicies(polRes.data);
    } catch (err) {
      console.error('Failed to fetch:', err);
    } finally {
      setLoading(false);
    }
  }, [urgencyFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleAction = async () => {
    if (!actionSignoff) return;
    try {
      setProcessing(true);
      const policyId = actionSignoff.signoff.policy.id;
      const signoffId = actionSignoff.signoff.id;

      if (actionSignoff.action === 'approve') {
        await approvePolicySignoff(policyId, signoffId, {
          comments: comments || undefined,
        });
      } else {
        if (!comments.trim()) { alert('Comments are required for rejection'); return; }
        await rejectPolicySignoff(policyId, signoffId, { comments });
      }
      setActionSignoff(null);
      setComments('');
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setProcessing(false);
    }
  };

  const overdueCount = pending.filter(p => p.urgency === 'overdue').length;
  const dueSoonCount = pending.filter(p => p.urgency === 'due_soon').length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <ClipboardCheck className="h-6 w-6" /> Policy Approvals
        </h1>
        <p className="text-sm text-muted-foreground">Review and sign-off on policy documents</p>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold">{pending.length}</div>
            <p className="text-xs text-muted-foreground">Pending Approvals</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold text-red-600">{overdueCount}</div>
            <p className="text-xs text-muted-foreground">Overdue</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold text-yellow-600">{dueSoonCount}</div>
            <p className="text-xs text-muted-foreground">Due Soon</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold">{recentPolicies.length}</div>
            <p className="text-xs text-muted-foreground">Policies In Review</p>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="pending">
            <Clock className="h-4 w-4 mr-1" /> My Pending ({pending.length})
          </TabsTrigger>
          <TabsTrigger value="in-review">
            <FileText className="h-4 w-4 mr-1" /> All In Review ({recentPolicies.length})
          </TabsTrigger>
        </TabsList>

        {/* Pending Approvals */}
        <TabsContent value="pending">
          {/* Urgency Filter */}
          <div className="mb-4">
            <Select value={urgencyFilter} onValueChange={(v) => setUrgencyFilter(v === 'all' ? '' : v)}>
              <SelectTrigger className="w-[180px]"><SelectValue placeholder="All urgencies" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All urgencies</SelectItem>
                <SelectItem value="overdue">Overdue</SelectItem>
                <SelectItem value="due_soon">Due Soon</SelectItem>
                <SelectItem value="on_time">On Time</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <Card>
            <CardContent className="p-0">
              {loading ? (
                <div className="p-6 text-center text-muted-foreground">Loading...</div>
              ) : pending.length === 0 ? (
                <div className="p-6 text-center">
                  <CheckCircle2 className="h-8 w-8 text-green-500 mx-auto mb-2" />
                  <p className="text-muted-foreground">No pending approvals</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Policy</TableHead>
                      <TableHead>Category</TableHead>
                      <TableHead>Version</TableHead>
                      <TableHead>Requested By</TableHead>
                      <TableHead>Due Date</TableHead>
                      <TableHead>Urgency</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {pending.map((p) => (
                      <TableRow key={p.id}>
                        <TableCell>
                          <Link href={`/policies/${p.policy.id}`} className="text-primary hover:underline">
                            <span className="font-mono text-xs mr-2">{p.policy.identifier}</span>
                            <span className="text-sm font-medium">{p.policy.title}</span>
                          </Link>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className="text-xs">
                            {CATEGORY_LABELS[p.policy.category] || p.policy.category}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm">v{p.policy_version.version_number}</TableCell>
                        <TableCell className="text-sm">{p.requested_by?.name || '—'}</TableCell>
                        <TableCell className="text-sm">
                          {p.due_date ? new Date(p.due_date).toLocaleDateString() : '—'}
                        </TableCell>
                        <TableCell>
                          {p.urgency && (
                            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${URGENCY_COLORS[p.urgency]}`}>
                              {p.urgency === 'overdue' && <AlertTriangle className="h-3 w-3 mr-1" />}
                              {p.urgency}
                            </span>
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Link href={`/policies/${p.policy.id}`}>
                              <Button variant="ghost" size="sm" className="h-7">
                                <Eye className="h-3 w-3 mr-1" /> View
                              </Button>
                            </Link>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="h-7 text-green-600"
                              onClick={() => { setActionSignoff({ signoff: p, action: 'approve' }); setComments(''); }}
                            >
                              <CheckCircle2 className="h-3 w-3 mr-1" /> Approve
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="h-7 text-red-600"
                              onClick={() => { setActionSignoff({ signoff: p, action: 'reject' }); setComments(''); }}
                            >
                              <XCircle className="h-3 w-3 mr-1" /> Reject
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* All In Review */}
        <TabsContent value="in-review">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Policies Currently In Review</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {recentPolicies.length === 0 ? (
                <div className="p-6 text-center text-muted-foreground">No policies currently in review</div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Identifier</TableHead>
                      <TableHead>Title</TableHead>
                      <TableHead>Category</TableHead>
                      <TableHead>Owner</TableHead>
                      <TableHead>Sign-offs</TableHead>
                      <TableHead>Updated</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {recentPolicies.map((pol) => (
                      <TableRow key={pol.id}>
                        <TableCell className="font-mono text-xs">{pol.identifier}</TableCell>
                        <TableCell>
                          <Link href={`/policies/${pol.id}`} className="text-primary hover:underline font-medium">
                            {pol.title}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className="text-xs">
                            {CATEGORY_LABELS[pol.category] || pol.category}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm">{pol.owner?.name || '—'}</TableCell>
                        <TableCell className="text-sm">
                          {pol.signoff_summary ? (
                            <span>
                              {pol.signoff_summary.approved}/{pol.signoff_summary.total}
                              {pol.signoff_summary.pending > 0 && (
                                <span className="text-yellow-600 ml-1">({pol.signoff_summary.pending} pending)</span>
                              )}
                            </span>
                          ) : '—'}
                        </TableCell>
                        <TableCell className="text-sm">{new Date(pol.updated_at).toLocaleDateString()}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Approve/Reject Dialog */}
      <Dialog open={!!actionSignoff} onOpenChange={() => setActionSignoff(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {actionSignoff?.action === 'approve' ? 'Approve' : 'Reject'} Sign-off
            </DialogTitle>
            <DialogDescription>
              {actionSignoff?.signoff.policy.identifier} — {actionSignoff?.signoff.policy.title}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>
                Comments {actionSignoff?.action === 'reject' ? '(required)' : '(optional)'}
              </Label>
              <Textarea
                placeholder={actionSignoff?.action === 'reject'
                  ? 'Explain why this policy needs changes...'
                  : 'Optional approval comments...'
                }
                value={comments}
                onChange={(e) => setComments(e.target.value)}
                rows={4}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setActionSignoff(null)}>Cancel</Button>
            <Button
              onClick={handleAction}
              disabled={processing || (actionSignoff?.action === 'reject' && !comments.trim())}
              variant={actionSignoff?.action === 'approve' ? 'default' : 'destructive'}
            >
              {processing ? 'Processing...' : actionSignoff?.action === 'approve' ? 'Approve' : 'Reject'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
