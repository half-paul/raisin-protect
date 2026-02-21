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
  ArrowLeft, Plus, Search, ChevronLeft, ChevronRight, AlertTriangle, Clock, CheckCircle,
} from 'lucide-react';
import {
  AuditRequest, listAuditRequests, createAuditRequest, submitAuditRequest,
} from '@/lib/api';
import {
  REQUEST_STATUS_LABELS, REQUEST_STATUS_COLORS,
  REQUEST_PRIORITY_LABELS, REQUEST_PRIORITY_COLORS,
} from '@/components/audit/constants';

export default function AuditRequestQueuePage() {
  const params = useParams();
  const auditId = params.id as string;
  const { hasRole } = useAuth();
  const canCreate = hasRole('ciso', 'compliance_manager', 'auditor');
  const canSubmit = hasRole('ciso', 'compliance_manager', 'security_engineer', 'it_admin');

  const [requests, setRequests] = useState<AuditRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [priorityFilter, setPriorityFilter] = useState('');
  const [overdueOnly, setOverdueOnly] = useState(false);

  // Create dialog
  const [createOpen, setCreateOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newReq, setNewReq] = useState({ title: '', description: '', priority: 'medium', due_date: '', reference_number: '' });

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const p: Record<string, string> = { page: String(page), per_page: String(perPage) };
      if (search) p.search = search;
      if (statusFilter) p.status = statusFilter;
      if (priorityFilter) p.priority = priorityFilter;
      if (overdueOnly) p.overdue = 'true';

      const res = await listAuditRequests(auditId, p);
      setRequests(res.data);
      setTotal(res.meta?.total || res.data.length);
    } catch (err) {
      console.error('Failed to fetch requests:', err);
    } finally {
      setLoading(false);
    }
  }, [auditId, page, search, statusFilter, priorityFilter, overdueOnly]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCreate = async () => {
    if (!newReq.title.trim() || !newReq.description.trim()) return;
    try {
      setCreating(true);
      await createAuditRequest(auditId, {
        title: newReq.title,
        description: newReq.description,
        priority: newReq.priority || undefined,
        due_date: newReq.due_date || undefined,
        reference_number: newReq.reference_number || undefined,
      });
      setCreateOpen(false);
      setNewReq({ title: '', description: '', priority: 'medium', due_date: '', reference_number: '' });
      fetchData();
    } catch (err) {
      console.error('Create failed:', err);
      alert(err instanceof Error ? err.message : 'Failed to create request');
    } finally {
      setCreating(false);
    }
  };

  const handleSubmit = async (reqId: string) => {
    if (!confirm('Mark this request as submitted? Auditor will be notified to review.')) return;
    try {
      await submitAuditRequest(auditId, reqId);
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Submit failed');
    }
  };

  const totalPages = Math.ceil(total / perPage);

  // Summary counts
  const openCount = requests.filter(r => r.status === 'open').length;
  const inProgressCount = requests.filter(r => r.status === 'in_progress').length;
  const submittedCount = requests.filter(r => r.status === 'submitted').length;
  const acceptedCount = requests.filter(r => r.status === 'accepted').length;
  const overdueCount = requests.filter(r => r.due_date && !['accepted', 'closed'].includes(r.status) && new Date(r.due_date) < new Date()).length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href={`/audit/${auditId}`}>
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">Evidence Request Queue</h1>
          <p className="text-sm text-muted-foreground">Manage evidence requests and track SLA compliance</p>
        </div>
        {canCreate && (
          <Dialog open={createOpen} onOpenChange={setCreateOpen}>
            <DialogTrigger asChild>
              <Button><Plus className="h-4 w-4 mr-2" /> New Request</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Evidence Request</DialogTitle>
                <DialogDescription>Request evidence from the internal team.</DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div>
                  <Label>Title *</Label>
                  <Input value={newReq.title} onChange={(e) => setNewReq({ ...newReq, title: e.target.value })} placeholder="Information Security Policy (current, approved)" />
                </div>
                <div>
                  <Label>Description *</Label>
                  <Textarea value={newReq.description} onChange={(e) => setNewReq({ ...newReq, description: e.target.value })} rows={3} placeholder="Describe what evidence is needed..." />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label>Priority</Label>
                    <Select value={newReq.priority} onValueChange={(v) => setNewReq({ ...newReq, priority: v })}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {Object.entries(REQUEST_PRIORITY_LABELS).map(([k, v]) => (
                          <SelectItem key={k} value={k}>{v}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label>Due Date</Label>
                    <Input type="date" value={newReq.due_date} onChange={(e) => setNewReq({ ...newReq, due_date: e.target.value })} />
                  </div>
                </div>
                <div>
                  <Label>Reference Number</Label>
                  <Input value={newReq.reference_number} onChange={(e) => setNewReq({ ...newReq, reference_number: e.target.value })} placeholder="PBC-001" />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                <Button onClick={handleCreate} disabled={creating || !newReq.title.trim() || !newReq.description.trim()}>
                  {creating ? 'Creating...' : 'Create Request'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-5">
        <Card><CardContent className="p-4"><div className="text-2xl font-bold">{openCount}</div><p className="text-xs text-muted-foreground">Open</p></CardContent></Card>
        <Card><CardContent className="p-4"><div className="text-2xl font-bold text-blue-600">{inProgressCount}</div><p className="text-xs text-muted-foreground">In Progress</p></CardContent></Card>
        <Card><CardContent className="p-4"><div className="text-2xl font-bold text-purple-600">{submittedCount}</div><p className="text-xs text-muted-foreground">Submitted</p></CardContent></Card>
        <Card><CardContent className="p-4"><div className="text-2xl font-bold text-green-600">{acceptedCount}</div><p className="text-xs text-muted-foreground">Accepted</p></CardContent></Card>
        <Card><CardContent className="p-4"><div className="text-2xl font-bold text-red-600">{overdueCount}</div><p className="text-xs text-muted-foreground">Overdue</p></CardContent></Card>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-4">
            <div className="flex-1 min-w-[200px]">
              <Label className="text-xs">Search</Label>
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input placeholder="Search requests..." value={searchInput} onChange={(e) => setSearchInput(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }} className="pl-8" />
              </div>
            </div>
            <div className="w-[150px]">
              <Label className="text-xs">Status</Label>
              <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  {Object.entries(REQUEST_STATUS_LABELS).map(([k, v]) => (<SelectItem key={k} value={k}>{v}</SelectItem>))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[140px]">
              <Label className="text-xs">Priority</Label>
              <Select value={priorityFilter} onValueChange={(v) => { setPriorityFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  {Object.entries(REQUEST_PRIORITY_LABELS).map(([k, v]) => (<SelectItem key={k} value={k}>{v}</SelectItem>))}
                </SelectContent>
              </Select>
            </div>
            <Button variant={overdueOnly ? 'default' : 'outline'} size="sm" onClick={() => { setOverdueOnly(!overdueOnly); setPage(1); }}>
              <AlertTriangle className="h-4 w-4 mr-1" /> Overdue Only
            </Button>
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
                <TableHead>Priority</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Assigned To</TableHead>
                <TableHead>Due Date</TableHead>
                <TableHead>Evidence</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Loading...</TableCell></TableRow>
              ) : requests.length === 0 ? (
                <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">No requests found</TableCell></TableRow>
              ) : (
                requests.map((req) => {
                  const isOverdue = req.due_date && !['accepted', 'closed'].includes(req.status) && new Date(req.due_date) < new Date();
                  const daysOverdue = isOverdue ? Math.ceil((Date.now() - new Date(req.due_date!).getTime()) / 86400000) : 0;
                  return (
                    <TableRow key={req.id} className={isOverdue ? 'bg-red-50/50 dark:bg-red-950/10' : ''}>
                      <TableCell className="font-mono text-xs">{req.reference_number || '—'}</TableCell>
                      <TableCell>
                        <Link href={`/audit/${auditId}/requests/${req.id}`} className="text-primary hover:underline font-medium text-sm">{req.title}</Link>
                      </TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${REQUEST_PRIORITY_COLORS[req.priority]}`}>{req.priority}</span>
                      </TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${REQUEST_STATUS_COLORS[req.status]}`}>{REQUEST_STATUS_LABELS[req.status]}</span>
                      </TableCell>
                      <TableCell className="text-sm">{req.assigned_to_name || '—'}</TableCell>
                      <TableCell className="text-sm">
                        {req.due_date ? new Date(req.due_date).toLocaleDateString() : '—'}
                        {isOverdue && <Badge variant="destructive" className="ml-1 text-[10px]">{daysOverdue}d</Badge>}
                      </TableCell>
                      <TableCell className="text-sm">{req.evidence_count}</TableCell>
                      <TableCell className="text-right">
                        {canSubmit && ['open', 'in_progress'].includes(req.status) && req.evidence_count > 0 && (
                          <Button variant="outline" size="sm" onClick={() => handleSubmit(req.id)}>
                            <CheckCircle className="h-3 w-3 mr-1" /> Submit
                          </Button>
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
    </div>
  );
}
