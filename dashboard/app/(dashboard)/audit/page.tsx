'use client';

import { useState, useEffect, useCallback } from 'react';
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
  Search, Plus, ChevronLeft, ChevronRight, Eye, ClipboardCheck,
  AlertTriangle, Clock, FileCheck, BarChart3,
} from 'lucide-react';
import {
  Audit, AuditDashboard,
  listAudits, getAuditDashboard, createAudit,
} from '@/lib/api';
import {
  AUDIT_STATUS_LABELS, AUDIT_STATUS_COLORS,
  AUDIT_TYPE_LABELS,
} from '@/components/audit/constants';

export default function AuditHubPage() {
  const { hasRole } = useAuth();
  const canCreate = hasRole('ciso', 'compliance_manager');

  const [audits, setAudits] = useState<Audit[]>([]);
  const [dashboard, setDashboard] = useState<AuditDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');

  // Create dialog
  const [createOpen, setCreateOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newAudit, setNewAudit] = useState({
    title: '',
    description: '',
    audit_type: 'soc2_type2',
    audit_firm: '',
    planned_start: '',
    planned_end: '',
    period_start: '',
    period_end: '',
    report_type: '',
  });

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
      };
      if (search) params.search = search;
      if (statusFilter) params.status = statusFilter;
      if (typeFilter) params.audit_type = typeFilter;

      const [auditsRes, dashRes] = await Promise.all([
        listAudits(params),
        getAuditDashboard().catch(() => ({ data: null })),
      ]);
      setAudits(auditsRes.data);
      setTotal(auditsRes.meta?.total || auditsRes.data.length);
      if (dashRes.data) setDashboard(dashRes.data);
    } catch (err) {
      console.error('Failed to fetch audits:', err);
    } finally {
      setLoading(false);
    }
  }, [page, search, statusFilter, typeFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCreate = async () => {
    if (!newAudit.title.trim() || !newAudit.audit_type) return;
    try {
      setCreating(true);
      const body: Record<string, unknown> = {
        title: newAudit.title,
        audit_type: newAudit.audit_type,
      };
      if (newAudit.description) body.description = newAudit.description;
      if (newAudit.audit_firm) body.audit_firm = newAudit.audit_firm;
      if (newAudit.planned_start) body.planned_start = newAudit.planned_start;
      if (newAudit.planned_end) body.planned_end = newAudit.planned_end;
      if (newAudit.period_start) body.period_start = newAudit.period_start;
      if (newAudit.period_end) body.period_end = newAudit.period_end;
      if (newAudit.report_type) body.report_type = newAudit.report_type;
      await createAudit(body as Parameters<typeof createAudit>[0]);
      setCreateOpen(false);
      setNewAudit({ title: '', description: '', audit_type: 'soc2_type2', audit_firm: '', planned_start: '', planned_end: '', period_start: '', period_end: '', report_type: '' });
      fetchData();
    } catch (err) {
      console.error('Create failed:', err);
      alert(err instanceof Error ? err.message : 'Failed to create audit');
    } finally {
      setCreating(false);
    }
  };

  const totalPages = Math.ceil(total / perPage);
  const ds = dashboard?.summary;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Audit Hub</h1>
          <p className="text-sm text-muted-foreground">Manage audit engagements, evidence requests, and findings</p>
        </div>
        <div className="flex items-center gap-2">
          <Link href="/audit-templates">
            <Button variant="outline" size="sm">
              <FileCheck className="h-4 w-4 mr-2" /> PBC Templates
            </Button>
          </Link>
          <Link href="/audit-readiness">
            <Button variant="outline" size="sm">
              <BarChart3 className="h-4 w-4 mr-2" /> Readiness
            </Button>
          </Link>
          {canCreate && (
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
              <DialogTrigger asChild>
                <Button><Plus className="h-4 w-4 mr-2" /> New Audit</Button>
              </DialogTrigger>
              <DialogContent className="max-w-lg">
                <DialogHeader>
                  <DialogTitle>Create Audit Engagement</DialogTitle>
                  <DialogDescription>Start a new audit engagement. You can add auditors and milestones after creation.</DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  <div>
                    <Label>Title *</Label>
                    <Input
                      placeholder="SOC 2 Type II — 2026 Annual"
                      value={newAudit.title}
                      onChange={(e) => setNewAudit({ ...newAudit, title: e.target.value })}
                    />
                  </div>
                  <div>
                    <Label>Audit Type *</Label>
                    <Select value={newAudit.audit_type} onValueChange={(v) => setNewAudit({ ...newAudit, audit_type: v })}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {Object.entries(AUDIT_TYPE_LABELS).map(([k, v]) => (
                          <SelectItem key={k} value={k}>{v}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <Label>Description</Label>
                    <Textarea
                      placeholder="Describe the audit scope and objectives..."
                      value={newAudit.description}
                      onChange={(e) => setNewAudit({ ...newAudit, description: e.target.value })}
                      rows={3}
                    />
                  </div>
                  <div>
                    <Label>Audit Firm</Label>
                    <Input
                      placeholder="Deloitte & Touche LLP"
                      value={newAudit.audit_firm}
                      onChange={(e) => setNewAudit({ ...newAudit, audit_firm: e.target.value })}
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label>Audit Period Start</Label>
                      <Input type="date" value={newAudit.period_start} onChange={(e) => setNewAudit({ ...newAudit, period_start: e.target.value })} />
                    </div>
                    <div>
                      <Label>Audit Period End</Label>
                      <Input type="date" value={newAudit.period_end} onChange={(e) => setNewAudit({ ...newAudit, period_end: e.target.value })} />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label>Planned Start</Label>
                      <Input type="date" value={newAudit.planned_start} onChange={(e) => setNewAudit({ ...newAudit, planned_start: e.target.value })} />
                    </div>
                    <div>
                      <Label>Planned End</Label>
                      <Input type="date" value={newAudit.planned_end} onChange={(e) => setNewAudit({ ...newAudit, planned_end: e.target.value })} />
                    </div>
                  </div>
                  <div>
                    <Label>Report Type</Label>
                    <Input
                      placeholder="SOC 2 Type II"
                      value={newAudit.report_type}
                      onChange={(e) => setNewAudit({ ...newAudit, report_type: e.target.value })}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                  <Button onClick={handleCreate} disabled={creating || !newAudit.title.trim()}>
                    {creating ? 'Creating...' : 'Create Engagement'}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          )}
        </div>
      </div>

      {/* Dashboard Stats Cards */}
      {ds && (
        <div className="grid gap-4 md:grid-cols-5">
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{ds.active_audits}</div>
              <p className="text-xs text-muted-foreground">Active Audits</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-green-600">{ds.completed_audits}</div>
              <p className="text-xs text-muted-foreground">Completed</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-blue-600">{ds.total_open_requests}</div>
              <p className="text-xs text-muted-foreground">Open Requests</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-orange-600">{ds.total_overdue_requests}</div>
              <p className="text-xs text-muted-foreground">Overdue Requests</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-red-600">{ds.critical_findings + ds.high_findings}</div>
              <p className="text-xs text-muted-foreground">Critical/High Findings</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Active Audit Cards (from dashboard) */}
      {dashboard?.active_audits && dashboard.active_audits.length > 0 && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {dashboard.active_audits.map((a) => (
            <Link key={a.id} href={`/audit/${a.id}`}>
              <Card className="hover:border-primary/50 transition-colors cursor-pointer">
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-medium truncate">{a.title}</CardTitle>
                    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium ${AUDIT_STATUS_COLORS[a.status]}`}>
                      {AUDIT_STATUS_LABELS[a.status]}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground">{AUDIT_TYPE_LABELS[a.audit_type] || a.audit_type}</p>
                </CardHeader>
                <CardContent className="pb-4">
                  {/* Readiness bar */}
                  <div className="mb-3">
                    <div className="flex items-center justify-between text-xs mb-1">
                      <span className="text-muted-foreground">Readiness</span>
                      <span className="font-medium">{a.readiness_pct}%</span>
                    </div>
                    <div className="h-2 bg-muted rounded-full overflow-hidden">
                      <div
                        className={`h-full rounded-full transition-all ${a.readiness_pct >= 75 ? 'bg-green-500' : a.readiness_pct >= 50 ? 'bg-yellow-500' : 'bg-red-500'}`}
                        style={{ width: `${a.readiness_pct}%` }}
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-2 text-xs">
                    <div className="flex items-center gap-1">
                      <ClipboardCheck className="h-3 w-3 text-muted-foreground" />
                      <span>{a.open_requests}/{a.total_requests} requests open</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <AlertTriangle className="h-3 w-3 text-muted-foreground" />
                      <span>{a.open_findings}/{a.total_findings} findings open</span>
                    </div>
                    {a.days_remaining != null && (
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3 text-muted-foreground" />
                        <span>{a.days_remaining}d remaining</span>
                      </div>
                    )}
                    {a.next_milestone && (
                      <div className="flex items-center gap-1 text-muted-foreground">
                        <span>Next: {a.next_milestone.name} ({a.next_milestone.days_until}d)</span>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-4">
            <div className="flex-1 min-w-[200px]">
              <Label className="text-xs">Search</Label>
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search audits..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="w-[160px]">
              <Label className="text-xs">Status</Label>
              <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All statuses</SelectItem>
                  {Object.entries(AUDIT_STATUS_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[200px]">
              <Label className="text-xs">Audit Type</Label>
              <Select value={typeFilter} onValueChange={(v) => { setTypeFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All types</SelectItem>
                  {Object.entries(AUDIT_TYPE_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Audit Table */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Firm</TableHead>
                <TableHead>Planned End</TableHead>
                <TableHead>Requests</TableHead>
                <TableHead>Findings</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Loading...</TableCell></TableRow>
              ) : audits.length === 0 ? (
                <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">No audits found</TableCell></TableRow>
              ) : (
                audits.map((audit) => (
                  <TableRow key={audit.id}>
                    <TableCell>
                      <Link href={`/audit/${audit.id}`} className="text-primary hover:underline font-medium">
                        {audit.title}
                      </Link>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {AUDIT_TYPE_LABELS[audit.audit_type] || audit.audit_type}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${AUDIT_STATUS_COLORS[audit.status]}`}>
                        {AUDIT_STATUS_LABELS[audit.status]}
                      </span>
                    </TableCell>
                    <TableCell className="text-sm">{audit.audit_firm || '—'}</TableCell>
                    <TableCell className="text-sm">{audit.planned_end ? new Date(audit.planned_end).toLocaleDateString() : '—'}</TableCell>
                    <TableCell>
                      <span className="text-sm">
                        {audit.open_requests}/{audit.total_requests}
                        {audit.open_requests > 0 && <span className="text-orange-500 ml-1">open</span>}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className="text-sm">
                        {audit.open_findings}/{audit.total_findings}
                        {audit.open_findings > 0 && <span className="text-red-500 ml-1">open</span>}
                      </span>
                    </TableCell>
                    <TableCell className="text-right">
                      <Link href={`/audit/${audit.id}`}>
                        <Button variant="ghost" size="icon" className="h-8 w-8"><Eye className="h-4 w-4" /></Button>
                      </Link>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} of {total}
          </p>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm">Page {page} of {totalPages}</span>
            <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}

      {/* Overdue Requests & Critical Findings (from dashboard) */}
      {dashboard && (
        <div className="grid gap-6 md:grid-cols-2">
          {/* Overdue Requests */}
          {dashboard.overdue_requests.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <Clock className="h-4 w-4 text-orange-500" />
                  Overdue Evidence Requests
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <Table>
                  <TableBody>
                    {dashboard.overdue_requests.slice(0, 5).map((req) => (
                      <TableRow key={req.id}>
                        <TableCell>
                          <p className="font-medium text-sm">{req.title}</p>
                          <p className="text-xs text-muted-foreground">{req.audit_title}</p>
                        </TableCell>
                        <TableCell className="text-right">
                          <Badge variant="destructive" className="text-xs">{req.days_overdue}d overdue</Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {/* Critical Findings */}
          {dashboard.critical_findings.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-red-500" />
                  Critical / High Findings
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <Table>
                  <TableBody>
                    {dashboard.critical_findings.slice(0, 5).map((f) => (
                      <TableRow key={f.id}>
                        <TableCell>
                          <p className="font-medium text-sm">{f.title}</p>
                          <p className="text-xs text-muted-foreground">{f.audit_title}</p>
                        </TableCell>
                        <TableCell className="text-right">
                          <Badge variant={f.severity === 'critical' ? 'destructive' : 'default'} className="text-xs">
                            {f.severity}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </div>
  );
}
