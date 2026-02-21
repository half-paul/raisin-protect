'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
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
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  Search, Plus, ChevronLeft, ChevronRight, FileText, BookOpen, Eye,
  Archive, Clock, CheckCircle2, XCircle, AlertTriangle, Send,
} from 'lucide-react';
import {
  Policy, PolicyStatus, PolicyCategory, PolicyStats,
  listPolicies, getPolicyStats, createPolicy, archivePolicy,
} from '@/lib/api';
import {
  POLICY_STATUS_LABELS as STATUS_LABELS,
  POLICY_STATUS_COLORS as STATUS_COLORS,
  POLICY_CATEGORY_LABELS as CATEGORY_LABELS,
  REVIEW_STATUS_LABELS,
  REVIEW_STATUS_COLORS,
} from '@/components/policy/constants';

export default function PolicyLibraryPage() {
  const { hasRole } = useAuth();
  const canCreate = hasRole('ciso', 'compliance_manager', 'security_engineer');
  const canArchive = hasRole('ciso', 'compliance_manager');

  const [policies, setPolicies] = useState<Policy[]>([]);
  const [stats, setStats] = useState<PolicyStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [reviewFilter, setReviewFilter] = useState('');

  // Create dialog
  const [showCreate, setShowCreate] = useState(false);
  const [createForm, setCreateForm] = useState({
    identifier: '',
    title: '',
    description: '',
    category: '' as string,
    content: '<h1>Policy Title</h1>\n<h2>1. Purpose</h2>\n<p>Describe the purpose of this policy.</p>\n<h2>2. Scope</h2>\n<p>Define who this policy applies to.</p>\n<h2>3. Policy Statement</h2>\n<p>State the policy requirements.</p>',
    review_frequency_days: '365',
    tags: '',
  });
  const [creating, setCreating] = useState(false);

  const fetchPolicies = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
      };
      if (search) params.search = search;
      if (statusFilter) params.status = statusFilter;
      if (categoryFilter) params.category = categoryFilter;
      if (reviewFilter) params.review_status = reviewFilter;

      const [polRes, statsRes] = await Promise.all([
        listPolicies(params),
        getPolicyStats(),
      ]);
      setPolicies(polRes.data);
      setTotal(polRes.meta?.total || polRes.data.length);
      setStats(statsRes.data);
    } catch (err) {
      console.error('Failed to fetch policies:', err);
    } finally {
      setLoading(false);
    }
  }, [page, search, statusFilter, categoryFilter, reviewFilter]);

  useEffect(() => { fetchPolicies(); }, [fetchPolicies]);

  const handleCreate = async () => {
    if (!createForm.identifier || !createForm.title || !createForm.category || !createForm.content) return;
    try {
      setCreating(true);
      await createPolicy({
        identifier: createForm.identifier,
        title: createForm.title,
        description: createForm.description || undefined,
        category: createForm.category,
        content: createForm.content,
        content_format: 'html',
        review_frequency_days: createForm.review_frequency_days ? parseInt(createForm.review_frequency_days) : undefined,
        tags: createForm.tags ? createForm.tags.split(',').map(t => t.trim()).filter(Boolean) : undefined,
      });
      setShowCreate(false);
      setCreateForm({
        identifier: '', title: '', description: '', category: '',
        content: '<h1>Policy Title</h1>\n<h2>1. Purpose</h2>\n<p>Describe the purpose of this policy.</p>\n<h2>2. Scope</h2>\n<p>Define who this policy applies to.</p>\n<h2>3. Policy Statement</h2>\n<p>State the policy requirements.</p>',
        review_frequency_days: '365', tags: '',
      });
      fetchPolicies();
    } catch (err) {
      console.error('Create policy failed:', err);
      alert(err instanceof Error ? err.message : 'Failed to create policy');
    } finally {
      setCreating(false);
    }
  };

  const handleArchive = async (id: string) => {
    if (!confirm('Are you sure you want to archive this policy?')) return;
    try {
      await archivePolicy(id);
      fetchPolicies();
    } catch (err) {
      console.error('Archive failed:', err);
      alert(err instanceof Error ? err.message : 'Failed to archive policy');
    }
  };

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Policy Library</h1>
          <p className="text-sm text-muted-foreground">Manage organizational policies and governance documents</p>
        </div>
        {canCreate && (
          <Button onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4 mr-2" /> New Policy
          </Button>
        )}
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid gap-4 md:grid-cols-5">
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{stats.total_policies}</div>
              <p className="text-xs text-muted-foreground">Total Policies</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-green-600">{stats.by_status?.published || 0}</div>
              <p className="text-xs text-muted-foreground">Published</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-yellow-600">{stats.by_status?.in_review || 0}</div>
              <p className="text-xs text-muted-foreground">In Review</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-red-600">{stats.review_status?.overdue || 0}</div>
              <p className="text-xs text-muted-foreground">Review Overdue</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{stats.gap_summary?.coverage_percentage?.toFixed(0) || 0}%</div>
              <p className="text-xs text-muted-foreground">Policy Coverage</p>
            </CardContent>
          </Card>
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
                  placeholder="Search policies..."
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
                <SelectTrigger><SelectValue placeholder="All statuses" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All statuses</SelectItem>
                  {Object.entries(STATUS_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[180px]">
              <Label className="text-xs">Category</Label>
              <Select value={categoryFilter} onValueChange={(v) => { setCategoryFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All categories" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All categories</SelectItem>
                  {Object.entries(CATEGORY_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[160px]">
              <Label className="text-xs">Review Status</Label>
              <Select value={reviewFilter} onValueChange={(v) => { setReviewFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  {Object.entries(REVIEW_STATUS_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
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
                <TableHead>Identifier</TableHead>
                <TableHead>Title</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Review</TableHead>
                <TableHead>Owner</TableHead>
                <TableHead>Controls</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Loading...</TableCell></TableRow>
              ) : policies.length === 0 ? (
                <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">No policies found</TableCell></TableRow>
              ) : (
                policies.map((pol) => (
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
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[pol.status]}`}>
                        {STATUS_LABELS[pol.status]}
                      </span>
                    </TableCell>
                    <TableCell>
                      {pol.review_status && (
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${REVIEW_STATUS_COLORS[pol.review_status]}`}>
                          {REVIEW_STATUS_LABELS[pol.review_status]}
                        </span>
                      )}
                    </TableCell>
                    <TableCell className="text-sm">{pol.owner?.name || '—'}</TableCell>
                    <TableCell className="text-sm">{pol.linked_controls_count ?? 0}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Link href={`/policies/${pol.id}`}>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <Eye className="h-4 w-4" />
                          </Button>
                        </Link>
                        {canArchive && pol.status !== 'archived' && (
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-red-500" onClick={() => handleArchive(pol.id)}>
                            <Archive className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
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

      {/* Create Policy Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Create New Policy</DialogTitle>
            <DialogDescription>Create a new policy document. You can also start from a template.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Identifier *</Label>
                <Input placeholder="POL-XX-001" value={createForm.identifier} onChange={(e) => setCreateForm(f => ({ ...f, identifier: e.target.value }))} />
              </div>
              <div>
                <Label>Category *</Label>
                <Select value={createForm.category} onValueChange={(v) => setCreateForm(f => ({ ...f, category: v }))}>
                  <SelectTrigger><SelectValue placeholder="Select category" /></SelectTrigger>
                  <SelectContent>
                    {Object.entries(CATEGORY_LABELS).map(([k, v]) => (
                      <SelectItem key={k} value={k}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div>
              <Label>Title *</Label>
              <Input placeholder="Policy title" value={createForm.title} onChange={(e) => setCreateForm(f => ({ ...f, title: e.target.value }))} />
            </div>
            <div>
              <Label>Description</Label>
              <Textarea placeholder="Brief description..." value={createForm.description} onChange={(e) => setCreateForm(f => ({ ...f, description: e.target.value }))} rows={2} />
            </div>
            <div>
              <Label>Content (HTML) *</Label>
              <Textarea placeholder="Policy content..." value={createForm.content} onChange={(e) => setCreateForm(f => ({ ...f, content: e.target.value }))} rows={8} className="font-mono text-sm" />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Review Frequency (days)</Label>
                <Input type="number" placeholder="365" value={createForm.review_frequency_days} onChange={(e) => setCreateForm(f => ({ ...f, review_frequency_days: e.target.value }))} />
              </div>
              <div>
                <Label>Tags (comma-separated)</Label>
                <Input placeholder="annual, security" value={createForm.tags} onChange={(e) => setCreateForm(f => ({ ...f, tags: e.target.value }))} />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={creating || !createForm.identifier || !createForm.title || !createForm.category || !createForm.content}>
              {creating ? 'Creating...' : 'Create Policy'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
