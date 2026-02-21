'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
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
  ArrowLeft, Plus, Search, ChevronLeft, ChevronRight, AlertTriangle, ShieldAlert,
} from 'lucide-react';
import {
  AuditFinding, listAuditFindings, createAuditFinding, transitionFindingStatus,
} from '@/lib/api';
import {
  FINDING_SEVERITY_LABELS, FINDING_SEVERITY_COLORS,
  FINDING_STATUS_LABELS, FINDING_STATUS_COLORS,
  FINDING_CATEGORY_LABELS,
} from '@/components/audit/constants';

export default function AuditFindingsPage() {
  const params = useParams();
  const auditId = params.id as string;
  const { hasRole } = useAuth();
  const canCreate = hasRole('auditor');
  const canTransition = hasRole('ciso', 'compliance_manager', 'security_engineer', 'auditor');

  const [findings, setFindings] = useState<AuditFinding[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [severityFilter, setSeverityFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  // Create dialog
  const [createOpen, setCreateOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newFinding, setNewFinding] = useState({
    title: '', description: '', severity: 'medium', category: 'access_control',
    recommendation: '', reference_number: '', remediation_due_date: '',
  });

  // Status transition dialog
  const [transOpen, setTransOpen] = useState(false);
  const [transTarget, setTransTarget] = useState<AuditFinding | null>(null);
  const [transStatus, setTransStatus] = useState('');
  const [transNotes, setTransNotes] = useState('');
  const [transPlan, setTransPlan] = useState('');
  const [transitioning, setTransitioning] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const p: Record<string, string> = { page: String(page), per_page: String(perPage) };
      if (search) p.search = search;
      if (severityFilter) p.severity = severityFilter;
      if (statusFilter) p.status = statusFilter;
      const res = await listAuditFindings(auditId, p);
      setFindings(res.data);
      setTotal(res.meta?.total || res.data.length);
    } catch (err) {
      console.error('Failed to fetch findings:', err);
    } finally {
      setLoading(false);
    }
  }, [auditId, page, search, severityFilter, statusFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCreate = async () => {
    if (!newFinding.title.trim() || !newFinding.description.trim()) return;
    try {
      setCreating(true);
      await createAuditFinding(auditId, {
        title: newFinding.title,
        description: newFinding.description,
        severity: newFinding.severity,
        category: newFinding.category,
        recommendation: newFinding.recommendation || undefined,
        reference_number: newFinding.reference_number || undefined,
        remediation_due_date: newFinding.remediation_due_date || undefined,
      });
      setCreateOpen(false);
      setNewFinding({ title: '', description: '', severity: 'medium', category: 'access_control', recommendation: '', reference_number: '', remediation_due_date: '' });
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create finding');
    } finally {
      setCreating(false);
    }
  };

  const handleTransition = async () => {
    if (!transTarget || !transStatus) return;
    try {
      setTransitioning(true);
      const body: Record<string, string | undefined> = { status: transStatus };
      if (transNotes) body.notes = transNotes;
      if (transPlan && transStatus === 'remediation_planned') body.remediation_plan = transPlan;
      await transitionFindingStatus(auditId, transTarget.id, body as Parameters<typeof transitionFindingStatus>[2]);
      setTransOpen(false);
      setTransTarget(null);
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

  const totalPages = Math.ceil(total / perPage);

  // Summary counts
  const criticalCount = findings.filter(f => f.severity === 'critical').length;
  const highCount = findings.filter(f => f.severity === 'high').length;
  const openCount = findings.filter(f => !['verified', 'closed', 'risk_accepted'].includes(f.status)).length;
  const remediatedCount = findings.filter(f => ['remediation_complete', 'verified', 'closed'].includes(f.status)).length;

  const getNextStatuses = (status: string): string[] => {
    const map: Record<string, string[]> = {
      identified: ['acknowledged'],
      acknowledged: ['remediation_planned'],
      remediation_planned: ['remediation_in_progress'],
      remediation_in_progress: ['remediation_complete'],
      remediation_complete: ['verified', 'remediation_in_progress'],
      verified: ['closed'],
      risk_accepted: ['closed'],
    };
    return map[status] || [];
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href={`/audit/${auditId}`}>
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">Audit Findings</h1>
          <p className="text-sm text-muted-foreground">Track deficiencies and remediation progress</p>
        </div>
        {canCreate && (
          <Dialog open={createOpen} onOpenChange={setCreateOpen}>
            <DialogTrigger asChild>
              <Button><Plus className="h-4 w-4 mr-2" /> New Finding</Button>
            </DialogTrigger>
            <DialogContent className="max-w-lg">
              <DialogHeader>
                <DialogTitle>Create Finding</DialogTitle>
                <DialogDescription>Report a new audit finding or deficiency.</DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div>
                  <Label>Title *</Label>
                  <Input value={newFinding.title} onChange={(e) => setNewFinding({ ...newFinding, title: e.target.value })} placeholder="Missing MFA on admin accounts" />
                </div>
                <div>
                  <Label>Description *</Label>
                  <Textarea value={newFinding.description} onChange={(e) => setNewFinding({ ...newFinding, description: e.target.value })} rows={3} placeholder="Describe the deficiency..." />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label>Severity *</Label>
                    <Select value={newFinding.severity} onValueChange={(v) => setNewFinding({ ...newFinding, severity: v })}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {Object.entries(FINDING_SEVERITY_LABELS).map(([k, v]) => (<SelectItem key={k} value={k}>{v}</SelectItem>))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label>Category *</Label>
                    <Select value={newFinding.category} onValueChange={(v) => setNewFinding({ ...newFinding, category: v })}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {Object.entries(FINDING_CATEGORY_LABELS).map(([k, v]) => (<SelectItem key={k} value={k}>{v}</SelectItem>))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div>
                  <Label>Recommendation</Label>
                  <Textarea value={newFinding.recommendation} onChange={(e) => setNewFinding({ ...newFinding, recommendation: e.target.value })} rows={2} placeholder="Recommend how to fix..." />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label>Reference Number</Label>
                    <Input value={newFinding.reference_number} onChange={(e) => setNewFinding({ ...newFinding, reference_number: e.target.value })} placeholder="F-001" />
                  </div>
                  <div>
                    <Label>Remediation Due Date</Label>
                    <Input type="date" value={newFinding.remediation_due_date} onChange={(e) => setNewFinding({ ...newFinding, remediation_due_date: e.target.value })} />
                  </div>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                <Button onClick={handleCreate} disabled={creating || !newFinding.title.trim() || !newFinding.description.trim()}>
                  {creating ? 'Creating...' : 'Create Finding'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card><CardContent className="p-4"><div className="text-2xl font-bold text-red-600">{criticalCount}</div><p className="text-xs text-muted-foreground">Critical</p></CardContent></Card>
        <Card><CardContent className="p-4"><div className="text-2xl font-bold text-orange-600">{highCount}</div><p className="text-xs text-muted-foreground">High</p></CardContent></Card>
        <Card><CardContent className="p-4"><div className="text-2xl font-bold">{openCount}</div><p className="text-xs text-muted-foreground">Open</p></CardContent></Card>
        <Card><CardContent className="p-4"><div className="text-2xl font-bold text-green-600">{remediatedCount}</div><p className="text-xs text-muted-foreground">Remediated</p></CardContent></Card>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-4">
            <div className="flex-1 min-w-[200px]">
              <Label className="text-xs">Search</Label>
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input placeholder="Search findings..." value={searchInput} onChange={(e) => setSearchInput(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }} className="pl-8" />
              </div>
            </div>
            <div className="w-[150px]">
              <Label className="text-xs">Severity</Label>
              <Select value={severityFilter} onValueChange={(v) => { setSeverityFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  {Object.entries(FINDING_SEVERITY_LABELS).map(([k, v]) => (<SelectItem key={k} value={k}>{v}</SelectItem>))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[170px]">
              <Label className="text-xs">Status</Label>
              <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  {Object.entries(FINDING_STATUS_LABELS).map(([k, v]) => (<SelectItem key={k} value={k}>{v}</SelectItem>))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Ref</TableHead>
                <TableHead>Title</TableHead>
                <TableHead>Severity</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Owner</TableHead>
                <TableHead>Rem. Due</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Loading...</TableCell></TableRow>
              ) : findings.length === 0 ? (
                <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">No findings found</TableCell></TableRow>
              ) : (
                findings.map((f) => {
                  const nextStatuses = getNextStatuses(f.status);
                  return (
                    <TableRow key={f.id}>
                      <TableCell className="font-mono text-xs">{f.reference_number || '—'}</TableCell>
                      <TableCell>
                        <Link href={`/audit/${auditId}/findings/${f.id}`} className="text-primary hover:underline font-medium text-sm">{f.title}</Link>
                        {f.comment_count > 0 && <span className="text-xs text-muted-foreground ml-1">💬 {f.comment_count}</span>}
                      </TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${FINDING_SEVERITY_COLORS[f.severity]}`}>
                          {FINDING_SEVERITY_LABELS[f.severity]}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline" className="text-xs">{FINDING_CATEGORY_LABELS[f.category] || f.category}</Badge>
                      </TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${FINDING_STATUS_COLORS[f.status]}`}>
                          {FINDING_STATUS_LABELS[f.status]}
                        </span>
                      </TableCell>
                      <TableCell className="text-sm">{f.remediation_owner_name || '—'}</TableCell>
                      <TableCell className="text-sm">{f.remediation_due_date ? new Date(f.remediation_due_date).toLocaleDateString() : '—'}</TableCell>
                      <TableCell className="text-right">
                        {canTransition && nextStatuses.length > 0 && (
                          <Select onValueChange={(v) => { setTransTarget(f); setTransStatus(v); setTransOpen(true); }}>
                            <SelectTrigger className="h-7 w-auto text-xs"><SelectValue placeholder="Advance →" /></SelectTrigger>
                            <SelectContent>
                              {nextStatuses.map(s => (<SelectItem key={s} value={s}>{FINDING_STATUS_LABELS[s]}</SelectItem>))}
                            </SelectContent>
                          </Select>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">Showing {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} of {total}</p>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}><ChevronLeft className="h-4 w-4" /></Button>
            <span className="text-sm">Page {page} of {totalPages}</span>
            <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}><ChevronRight className="h-4 w-4" /></Button>
          </div>
        </div>
      )}

      {/* Transition Dialog */}
      <Dialog open={transOpen} onOpenChange={setTransOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update Finding Status</DialogTitle>
            <DialogDescription>
              {transTarget && <>Move &quot;{transTarget.title}&quot; to <strong>{FINDING_STATUS_LABELS[transStatus]}</strong></>}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {transStatus === 'remediation_planned' && (
              <div>
                <Label>Remediation Plan *</Label>
                <Textarea value={transPlan} onChange={(e) => setTransPlan(e.target.value)} rows={3} placeholder="Describe the remediation plan..." />
              </div>
            )}
            <div>
              <Label>Notes</Label>
              <Input value={transNotes} onChange={(e) => setTransNotes(e.target.value)} placeholder="Optional notes..." />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTransOpen(false)}>Cancel</Button>
            <Button onClick={handleTransition} disabled={transitioning}>
              {transitioning ? 'Updating...' : 'Update Status'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
