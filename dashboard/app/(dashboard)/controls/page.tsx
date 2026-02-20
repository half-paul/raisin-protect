'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Shield,
  Search,
  Plus,
  MoreVertical,
  ChevronLeft,
  ChevronRight,
  Filter,
  AlertCircle,
  CheckCircle2,
  Layers,
  Zap,
} from 'lucide-react';
import {
  Control,
  ControlStats,
  listControls,
  getControlStats,
  createControl,
  changeControlStatus,
  deprecateControl,
  bulkControlStatus,
} from '@/lib/api';

const STATUS_LABELS: Record<string, string> = {
  draft: 'Draft',
  active: 'Active',
  under_review: 'Under Review',
  deprecated: 'Deprecated',
};

const STATUS_COLORS: Record<string, string> = {
  draft: 'secondary',
  active: 'default',
  under_review: 'outline',
  deprecated: 'destructive',
};

const CATEGORY_LABELS: Record<string, string> = {
  technical: 'Technical',
  administrative: 'Administrative',
  physical: 'Physical',
  operational: 'Operational',
};

const CATEGORY_ICONS: Record<string, typeof Shield> = {
  technical: Zap,
  administrative: Layers,
  physical: Shield,
  operational: Layers,
};

export default function ControlsPage() {
  const { hasRole } = useAuth();
  const canCreate = hasRole('ciso', 'compliance_manager', 'security_engineer');
  const canBulk = hasRole('ciso', 'compliance_manager');

  const [controls, setControls] = useState<Control[]>([]);
  const [stats, setStats] = useState<ControlStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [sortField, setSortField] = useState('identifier');

  // Bulk selection
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [showBulkDialog, setShowBulkDialog] = useState(false);
  const [bulkStatus, setBulkStatus] = useState('active');
  const [bulkLoading, setBulkLoading] = useState(false);

  // Create dialog
  const [showCreate, setShowCreate] = useState(false);
  const [createForm, setCreateForm] = useState({
    identifier: '',
    title: '',
    description: '',
    category: 'technical',
    status: 'draft',
  });
  const [createLoading, setCreateLoading] = useState(false);
  const [createError, setCreateError] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
        sort: sortField,
      };
      if (search) params.search = search;
      if (statusFilter) params.status = statusFilter;
      if (categoryFilter) params.category = categoryFilter;

      const [ctrlRes, statsRes] = await Promise.all([
        listControls(params),
        getControlStats(),
      ]);

      setControls(ctrlRes.data || []);
      setTotal(ctrlRes.meta?.total || 0);
      setStats(statsRes.data);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, [page, search, statusFilter, categoryFilter, sortField]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Debounced search
  const [searchInput, setSearchInput] = useState('');
  useEffect(() => {
    const timer = setTimeout(() => {
      setSearch(searchInput);
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput]);

  function toggleSelected(id: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function toggleAll() {
    if (selected.size === controls.length) {
      setSelected(new Set());
    } else {
      setSelected(new Set(controls.map((c) => c.id)));
    }
  }

  async function handleBulkStatus() {
    setBulkLoading(true);
    try {
      await bulkControlStatus(Array.from(selected), bulkStatus);
      setSelected(new Set());
      setShowBulkDialog(false);
      fetchData();
    } catch {
      // handle
    } finally {
      setBulkLoading(false);
    }
  }

  async function handleStatusChange(id: string, newStatus: string) {
    try {
      await changeControlStatus(id, newStatus);
      fetchData();
    } catch {
      // handle
    }
  }

  async function handleDeprecate(id: string) {
    if (!confirm('Deprecate this control? Mappings will be preserved.')) return;
    try {
      await deprecateControl(id);
      fetchData();
    } catch {
      // handle
    }
  }

  async function handleCreate() {
    setCreateError('');
    setCreateLoading(true);
    try {
      await createControl(createForm);
      setShowCreate(false);
      setCreateForm({ identifier: '', title: '', description: '', category: 'technical', status: 'draft' });
      fetchData();
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create');
    } finally {
      setCreateLoading(false);
    }
  }

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Shield className="h-8 w-8" />
            Control Library
          </h1>
          <p className="text-muted-foreground mt-1">
            Browse, search, and manage your organization&apos;s controls
          </p>
        </div>
        {canCreate && (
          <Button onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4 mr-2" />
            New Control
          </Button>
        )}
      </div>

      {/* Stats cards */}
      {stats && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold">{stats.total}</div>
              <p className="text-xs text-muted-foreground">Total Controls</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                {stats.by_status?.active || 0}
              </div>
              <p className="text-xs text-muted-foreground">Active</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-amber-600 dark:text-amber-400">
                {stats.by_status?.draft || 0}
              </div>
              <p className="text-xs text-muted-foreground">Draft</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-red-600 dark:text-red-400">
                {stats.unmapped_count}
              </div>
              <p className="text-xs text-muted-foreground">Unmapped</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                {stats.custom_count}
              </div>
              <p className="text-xs text-muted-foreground">Custom</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search controls..."
            className="pl-10"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </div>
        <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Statuses</SelectItem>
            {Object.entries(STATUS_LABELS).map(([k, v]) => (
              <SelectItem key={k} value={k}>{v}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={categoryFilter} onValueChange={(v) => { setCategoryFilter(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Category" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Categories</SelectItem>
            {Object.entries(CATEGORY_LABELS).map(([k, v]) => (
              <SelectItem key={k} value={k}>{v}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <span className="text-sm text-muted-foreground">{total} controls</span>

        {/* Bulk actions */}
        {canBulk && selected.size > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowBulkDialog(true)}
          >
            <Layers className="h-4 w-4 mr-1" />
            Bulk Status ({selected.size})
          </Button>
        )}
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : controls.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              No controls found matching your filters
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  {canBulk && (
                    <TableHead className="w-[40px]">
                      <Checkbox
                        checked={selected.size === controls.length && controls.length > 0}
                        onCheckedChange={toggleAll}
                      />
                    </TableHead>
                  )}
                  <TableHead className="w-[130px]">Identifier</TableHead>
                  <TableHead>Title</TableHead>
                  <TableHead className="w-[120px]">Category</TableHead>
                  <TableHead className="w-[110px]">Status</TableHead>
                  <TableHead className="w-[80px] text-center">Mappings</TableHead>
                  <TableHead className="w-[160px]">Frameworks</TableHead>
                  <TableHead className="w-[50px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {controls.map((ctrl) => (
                  <TableRow key={ctrl.id} className="group">
                    {canBulk && (
                      <TableCell>
                        <Checkbox
                          checked={selected.has(ctrl.id)}
                          onCheckedChange={() => toggleSelected(ctrl.id)}
                        />
                      </TableCell>
                    )}
                    <TableCell>
                      <Link
                        href={`/controls/${ctrl.id}`}
                        className="font-mono text-sm font-medium hover:text-primary"
                      >
                        {ctrl.identifier}
                      </Link>
                    </TableCell>
                    <TableCell className="max-w-xs">
                      <Link href={`/controls/${ctrl.id}`} className="hover:text-primary">
                        <span className="line-clamp-1 text-sm">{ctrl.title}</span>
                      </Link>
                      {ctrl.is_custom && (
                        <Badge variant="outline" className="text-[10px] ml-1">Custom</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {CATEGORY_LABELS[ctrl.category] || ctrl.category}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={STATUS_COLORS[ctrl.status] as 'default' | 'secondary' | 'destructive' | 'outline'}>
                        {STATUS_LABELS[ctrl.status] || ctrl.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-center">
                      <span className="text-sm font-medium">{ctrl.mappings_count}</span>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {ctrl.frameworks?.slice(0, 3).map((fw) => (
                          <Badge key={fw} variant="secondary" className="text-[10px]">
                            {fw}
                          </Badge>
                        ))}
                        {(ctrl.frameworks?.length || 0) > 3 && (
                          <Badge variant="secondary" className="text-[10px]">
                            +{(ctrl.frameworks?.length || 0) - 3}
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8 opacity-0 group-hover:opacity-100">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem asChild>
                            <Link href={`/controls/${ctrl.id}`}>View Details</Link>
                          </DropdownMenuItem>
                          {ctrl.status === 'draft' && canCreate && (
                            <DropdownMenuItem onClick={() => handleStatusChange(ctrl.id, 'active')}>
                              Activate
                            </DropdownMenuItem>
                          )}
                          {ctrl.status === 'active' && canCreate && (
                            <DropdownMenuItem onClick={() => handleStatusChange(ctrl.id, 'under_review')}>
                              Mark for Review
                            </DropdownMenuItem>
                          )}
                          {ctrl.status !== 'deprecated' && canBulk && (
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => handleDeprecate(ctrl.id)}
                            >
                              Deprecate
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Create Control Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Create New Control</DialogTitle>
            <DialogDescription>
              Add a custom control to your organization&apos;s library
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {createError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
                <AlertCircle className="h-4 w-4" />
                {createError}
              </div>
            )}
            <div className="space-y-2">
              <Label>Identifier</Label>
              <Input
                placeholder="CTRL-XX-001"
                value={createForm.identifier}
                onChange={(e) => setCreateForm((f) => ({ ...f, identifier: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label>Title</Label>
              <Input
                placeholder="Control title"
                value={createForm.title}
                onChange={(e) => setCreateForm((f) => ({ ...f, title: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label>Description</Label>
              <textarea
                className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                placeholder="Describe what this control does..."
                value={createForm.description}
                onChange={(e) => setCreateForm((f) => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Category</Label>
                <Select
                  value={createForm.category}
                  onValueChange={(v) => setCreateForm((f) => ({ ...f, category: v }))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {Object.entries(CATEGORY_LABELS).map(([k, v]) => (
                      <SelectItem key={k} value={k}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Initial Status</Label>
                <Select
                  value={createForm.status}
                  onValueChange={(v) => setCreateForm((f) => ({ ...f, status: v }))}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="draft">Draft</SelectItem>
                    <SelectItem value="active">Active</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>Cancel</Button>
            <Button
              onClick={handleCreate}
              disabled={createLoading || !createForm.identifier || !createForm.title || !createForm.description}
            >
              {createLoading ? 'Creating...' : 'Create Control'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Bulk Status Dialog */}
      <Dialog open={showBulkDialog} onOpenChange={setShowBulkDialog}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>Bulk Status Change</DialogTitle>
            <DialogDescription>
              Change status for {selected.size} selected control{selected.size !== 1 ? 's' : ''}
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Label>New Status</Label>
            <Select value={bulkStatus} onValueChange={setBulkStatus}>
              <SelectTrigger className="mt-2">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(STATUS_LABELS).map(([k, v]) => (
                  <SelectItem key={k} value={k}>{v}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBulkDialog(false)}>Cancel</Button>
            <Button onClick={handleBulkStatus} disabled={bulkLoading}>
              {bulkLoading ? 'Updating...' : `Update ${selected.size} Controls`}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
