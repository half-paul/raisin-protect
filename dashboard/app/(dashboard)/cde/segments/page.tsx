'use client';

import { useState, useEffect, useCallback } from 'react';
import { Plus, Search, MoreVertical, AlertCircle, Network } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
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
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import {
  NetworkSegment,
  IsolationMethod,
  listNetworkSegments,
  createNetworkSegment,
  updateNetworkSegment,
  deleteNetworkSegment,
} from '@/lib/api';
import {
  ISOLATION_METHOD_LABELS,
  ISOLATION_METHOD_COLORS,
} from '@/components/cde/constants';
import { useAuth } from '@/lib/auth-context';

const ISOLATION_METHODS: IsolationMethod[] = [
  'firewall', 'vlan', 'dmz', 'air_gap', 'cloud_vpc', 'microsegmentation', 'other',
];

const emptyForm = {
  name: '',
  isolation_method: 'firewall' as IsolationMethod,
  description: '',
  cidr: '',
  vlan_id: '',
  notes: '',
};

export default function NetworkSegmentsPage() {
  const { hasRole } = useAuth();
  const canWrite = hasRole('ciso', 'compliance_manager', 'security_engineer', 'it_admin');

  const [segments, setSegments] = useState<NetworkSegment[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');
  const [methodFilter, setMethodFilter] = useState('');

  const [showDialog, setShowDialog] = useState(false);
  const [editingSegment, setEditingSegment] = useState<NetworkSegment | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState('');

  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = { page: String(page), per_page: String(perPage) };
      if (search) params.search = search;
      if (methodFilter) params.isolation_method = methodFilter;
      const res = await listNetworkSegments(params);
      setSegments(res.data);
      setTotal(res.meta?.total ?? res.data.length);
    } catch (err) {
      console.error('Failed to fetch network segments:', err);
    } finally {
      setLoading(false);
    }
  }, [page, search, methodFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  function openCreate() {
    setEditingSegment(null);
    setForm(emptyForm);
    setFormError('');
    setShowDialog(true);
  }

  function openEdit(seg: NetworkSegment) {
    setEditingSegment(seg);
    setForm({
      name: seg.name,
      isolation_method: seg.isolation_method,
      description: seg.description ?? '',
      cidr: seg.cidr ?? '',
      vlan_id: seg.vlan_id ?? '',
      notes: seg.notes ?? '',
    });
    setFormError('');
    setShowDialog(true);
  }

  async function handleSubmit() {
    setFormError('');
    setFormLoading(true);
    try {
      if (editingSegment) {
        await updateNetworkSegment(editingSegment.id, form);
      } else {
        await createNetworkSegment(form);
      }
      setShowDialog(false);
      fetchData();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to save segment');
    } finally {
      setFormLoading(false);
    }
  }

  async function handleDelete() {
    if (!deletingId) return;
    setDeleteLoading(true);
    try {
      await deleteNetworkSegment(deletingId);
      setDeletingId(null);
      fetchData();
    } catch (err) {
      console.error('Failed to delete segment:', err);
    } finally {
      setDeleteLoading(false);
    }
  }

  // Count by isolation method
  const methodCounts = ISOLATION_METHODS.reduce<Record<string, number>>((acc, m) => {
    acc[m] = segments.filter((s) => s.isolation_method === m).length;
    return acc;
  }, {});

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Network Segments</h1>
          <p className="text-muted-foreground text-sm mt-1">
            Document and review network segments isolating the CDE per PCI DSS Req 1.3
          </p>
        </div>
        {canWrite && (
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4 mr-2" />
            Add Segment
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold">{total}</div>
            <p className="text-xs text-muted-foreground">Total Segments</p>
          </CardContent>
        </Card>
        {(['firewall', 'vlan', 'cloud_vpc'] as IsolationMethod[]).map((m) => (
          <Card key={m}>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{methodCounts[m] ?? 0}</div>
              <p className="text-xs text-muted-foreground">{ISOLATION_METHOD_LABELS[m]}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-4">
            <div className="flex-1 min-w-[200px]">
              <Label className="text-xs">Search</Label>
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search segments..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="w-[180px]">
              <Label className="text-xs">Isolation Method</Label>
              <Select value={methodFilter || 'all'} onValueChange={(v) => { setMethodFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All methods</SelectItem>
                  {ISOLATION_METHODS.map((m) => (
                    <SelectItem key={m} value={m}>{ISOLATION_METHOD_LABELS[m]}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button variant="outline" size="sm" onClick={() => { setSearch(searchInput); setPage(1); }}>
              Search
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
                <TableHead>Name</TableHead>
                <TableHead>Isolation Method</TableHead>
                <TableHead>CIDR</TableHead>
                <TableHead>VLAN ID</TableHead>
                <TableHead>Description</TableHead>
                {canWrite && <TableHead className="w-[50px]" />}
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={canWrite ? 6 : 5} className="text-center py-8 text-muted-foreground">
                    Loading segments...
                  </TableCell>
                </TableRow>
              ) : segments.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={canWrite ? 6 : 5} className="text-center py-8">
                    <Network className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                    <p className="text-muted-foreground text-sm">No network segments defined</p>
                  </TableCell>
                </TableRow>
              ) : (
                segments.map((seg) => (
                  <TableRow key={seg.id}>
                    <TableCell className="font-medium">{seg.name}</TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${ISOLATION_METHOD_COLORS[seg.isolation_method]}`}>
                        {ISOLATION_METHOD_LABELS[seg.isolation_method]}
                      </span>
                    </TableCell>
                    <TableCell className="text-sm font-mono text-muted-foreground">
                      {seg.cidr || '—'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {seg.vlan_id || '—'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground max-w-[240px] truncate">
                      {seg.description || '—'}
                    </TableCell>
                    {canWrite && (
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon">
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => openEdit(seg)}>Edit</DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => setDeletingId(seg.id)}
                            >
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    )}
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} of {total}
          </p>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              Previous
            </Button>
            <span className="text-sm">Page {page} of {totalPages}</span>
            <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              Next
            </Button>
          </div>
        </div>
      )}

      {/* Create / Edit Dialog */}
      <Dialog open={showDialog} onOpenChange={setShowDialog}>
        <DialogContent className="sm:max-w-[520px]">
          <DialogHeader>
            <DialogTitle>{editingSegment ? 'Edit Segment' : 'Add Network Segment'}</DialogTitle>
            <DialogDescription>
              Define a network segment and its isolation method.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            {formError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
                <AlertCircle className="h-4 w-4 shrink-0" />
                {formError}
              </div>
            )}
            <div className="space-y-2">
              <Label>Name <span className="text-destructive">*</span></Label>
              <Input
                placeholder="e.g. CDE Network Zone"
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label>Isolation Method <span className="text-destructive">*</span></Label>
              <Select value={form.isolation_method} onValueChange={(v) => setForm((f) => ({ ...f, isolation_method: v as IsolationMethod }))}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {ISOLATION_METHODS.map((m) => (
                    <SelectItem key={m} value={m}>{ISOLATION_METHOD_LABELS[m]}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>CIDR Range</Label>
                <Input
                  placeholder="10.0.1.0/24"
                  value={form.cidr}
                  onChange={(e) => setForm((f) => ({ ...f, cidr: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label>VLAN ID</Label>
                <Input
                  placeholder="100"
                  value={form.vlan_id}
                  onChange={(e) => setForm((f) => ({ ...f, vlan_id: e.target.value }))}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Description</Label>
              <Textarea
                placeholder="Describe this segment's purpose and boundaries..."
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                rows={2}
              />
            </div>
            <div className="space-y-2">
              <Label>Notes</Label>
              <Textarea
                placeholder="Any additional notes for QSA review..."
                value={form.notes}
                onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
                rows={2}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDialog(false)}>Cancel</Button>
            <Button onClick={handleSubmit} disabled={formLoading || !form.name}>
              {formLoading ? 'Saving...' : editingSegment ? 'Save Changes' : 'Add Segment'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deletingId} onOpenChange={() => setDeletingId(null)}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>Delete Segment</DialogTitle>
            <DialogDescription>
              This will permanently remove this network segment. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeletingId(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteLoading}>
              {deleteLoading ? 'Deleting...' : 'Delete Segment'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
