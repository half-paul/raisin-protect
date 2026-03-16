'use client';

import { useState, useEffect, useCallback } from 'react';
import { Plus, Search, MoreVertical, AlertCircle, Server } from 'lucide-react';
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
  CdeAsset,
  CdeAssetScope,
  CdeAssetType,
  listCdeAssets,
  createCdeAsset,
  updateCdeAsset,
  deleteCdeAsset,
} from '@/lib/api';
import {
  ASSET_SCOPE_LABELS,
  ASSET_SCOPE_COLORS,
  ASSET_TYPE_LABELS,
} from '@/components/cde/constants';
import { useAuth } from '@/lib/auth-context';

const ASSET_TYPES: CdeAssetType[] = [
  'server', 'database', 'network_device', 'application', 'endpoint', 'cloud_service', 'storage', 'other',
];
const SCOPE_STATUSES: CdeAssetScope[] = ['in_scope', 'out_of_scope', 'connected_to'];

const emptyForm = {
  name: '',
  asset_type: 'server' as CdeAssetType,
  scope_status: 'in_scope' as CdeAssetScope,
  description: '',
  ip_address: '',
  hostname: '',
  owner: '',
  notes: '',
};

export default function CdeAssetsPage() {
  const { hasRole } = useAuth();
  const canWrite = hasRole('ciso', 'compliance_manager', 'security_engineer', 'it_admin');

  const [assets, setAssets] = useState<CdeAsset[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');
  const [scopeFilter, setScopeFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');

  // Stats
  const [inScopeCount, setInScopeCount] = useState(0);
  const [outOfScopeCount, setOutOfScopeCount] = useState(0);
  const [connectedToCount, setConnectedToCount] = useState(0);

  // Create/Edit dialog
  const [showDialog, setShowDialog] = useState(false);
  const [editingAsset, setEditingAsset] = useState<CdeAsset | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState('');

  // Delete confirm
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
      };
      if (search) params.search = search;
      if (scopeFilter) params.scope_status = scopeFilter;
      if (typeFilter) params.asset_type = typeFilter;
      const res = await listCdeAssets(params);
      setAssets(res.data);
      setTotal(res.meta?.total ?? res.data.length);
      // Compute local stats when no filter applied
      if (!scopeFilter && !search && !typeFilter) {
        setInScopeCount(res.data.filter((a) => a.scope_status === 'in_scope').length);
        setOutOfScopeCount(res.data.filter((a) => a.scope_status === 'out_of_scope').length);
        setConnectedToCount(res.data.filter((a) => a.scope_status === 'connected_to').length);
      }
    } catch (err) {
      console.error('Failed to fetch CDE assets:', err);
    } finally {
      setLoading(false);
    }
  }, [page, search, scopeFilter, typeFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  function openCreate() {
    setEditingAsset(null);
    setForm(emptyForm);
    setFormError('');
    setShowDialog(true);
  }

  function openEdit(asset: CdeAsset) {
    setEditingAsset(asset);
    setForm({
      name: asset.name,
      asset_type: asset.asset_type,
      scope_status: asset.scope_status,
      description: asset.description ?? '',
      ip_address: asset.ip_address ?? '',
      hostname: asset.hostname ?? '',
      owner: asset.owner ?? '',
      notes: asset.notes ?? '',
    });
    setFormError('');
    setShowDialog(true);
  }

  async function handleSubmit() {
    setFormError('');
    setFormLoading(true);
    try {
      if (editingAsset) {
        await updateCdeAsset(editingAsset.id, form);
      } else {
        await createCdeAsset(form);
      }
      setShowDialog(false);
      fetchData();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to save asset');
    } finally {
      setFormLoading(false);
    }
  }

  async function handleDelete() {
    if (!deletingId) return;
    setDeleteLoading(true);
    try {
      await deleteCdeAsset(deletingId);
      setDeletingId(null);
      fetchData();
    } catch (err) {
      console.error('Failed to delete asset:', err);
    } finally {
      setDeleteLoading(false);
    }
  }

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">CDE Asset Inventory</h1>
          <p className="text-muted-foreground text-sm mt-1">
            Cardholder Data Environment assets — track scope classification per PCI DSS Req 1
          </p>
        </div>
        {canWrite && (
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4 mr-2" />
            Add Asset
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold text-red-600">{inScopeCount}</div>
            <p className="text-xs text-muted-foreground">In Scope</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold text-orange-600">{connectedToCount}</div>
            <p className="text-xs text-muted-foreground">Connected To CDE</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold text-green-600">{outOfScopeCount}</div>
            <p className="text-xs text-muted-foreground">Out of Scope</p>
          </CardContent>
        </Card>
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
                  placeholder="Search assets..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="w-[160px]">
              <Label className="text-xs">Scope Status</Label>
              <Select value={scopeFilter || 'all'} onValueChange={(v) => { setScopeFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All statuses</SelectItem>
                  {SCOPE_STATUSES.map((s) => (
                    <SelectItem key={s} value={s}>{ASSET_SCOPE_LABELS[s]}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[160px]">
              <Label className="text-xs">Asset Type</Label>
              <Select value={typeFilter || 'all'} onValueChange={(v) => { setTypeFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All types</SelectItem>
                  {ASSET_TYPES.map((t) => (
                    <SelectItem key={t} value={t}>{ASSET_TYPE_LABELS[t]}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => { setSearch(searchInput); setPage(1); }}
            >
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
                <TableHead>Type</TableHead>
                <TableHead>Scope Status</TableHead>
                <TableHead>IP / Hostname</TableHead>
                <TableHead>Owner</TableHead>
                {canWrite && <TableHead className="w-[50px]" />}
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={canWrite ? 6 : 5} className="text-center py-8 text-muted-foreground">
                    Loading assets...
                  </TableCell>
                </TableRow>
              ) : assets.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={canWrite ? 6 : 5} className="text-center py-8">
                    <Server className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                    <p className="text-muted-foreground text-sm">No CDE assets found</p>
                  </TableCell>
                </TableRow>
              ) : (
                assets.map((asset) => (
                  <TableRow key={asset.id}>
                    <TableCell>
                      <div className="font-medium">{asset.name}</div>
                      {asset.description && (
                        <div className="text-xs text-muted-foreground line-clamp-1">{asset.description}</div>
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {ASSET_TYPE_LABELS[asset.asset_type]}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${ASSET_SCOPE_COLORS[asset.scope_status]}`}>
                        {ASSET_SCOPE_LABELS[asset.scope_status]}
                      </span>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {asset.ip_address || asset.hostname || '—'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {asset.owner || '—'}
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
                            <DropdownMenuItem onClick={() => openEdit(asset)}>Edit</DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => setDeletingId(asset.id)}
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

      {/* Pagination */}
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
            <DialogTitle>{editingAsset ? 'Edit Asset' : 'Add CDE Asset'}</DialogTitle>
            <DialogDescription>
              {editingAsset ? 'Update asset details and scope classification.' : 'Add a system, device, or service to the CDE inventory.'}
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
                placeholder="e.g. Payment Processing Server"
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Asset Type <span className="text-destructive">*</span></Label>
                <Select value={form.asset_type} onValueChange={(v) => setForm((f) => ({ ...f, asset_type: v as CdeAssetType }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {ASSET_TYPES.map((t) => (
                      <SelectItem key={t} value={t}>{ASSET_TYPE_LABELS[t]}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Scope Status <span className="text-destructive">*</span></Label>
                <Select value={form.scope_status} onValueChange={(v) => setForm((f) => ({ ...f, scope_status: v as CdeAssetScope }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {SCOPE_STATUSES.map((s) => (
                      <SelectItem key={s} value={s}>{ASSET_SCOPE_LABELS[s]}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>IP Address</Label>
                <Input
                  placeholder="192.168.1.10"
                  value={form.ip_address}
                  onChange={(e) => setForm((f) => ({ ...f, ip_address: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label>Hostname</Label>
                <Input
                  placeholder="pay-srv-01.internal"
                  value={form.hostname}
                  onChange={(e) => setForm((f) => ({ ...f, hostname: e.target.value }))}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Owner</Label>
              <Input
                placeholder="e.g. infrastructure@company.com"
                value={form.owner}
                onChange={(e) => setForm((f) => ({ ...f, owner: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label>Description</Label>
              <Textarea
                placeholder="Brief description of this asset and its role..."
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
              {formLoading ? 'Saving...' : editingAsset ? 'Save Changes' : 'Add Asset'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <Dialog open={!!deletingId} onOpenChange={() => setDeletingId(null)}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>Delete Asset</DialogTitle>
            <DialogDescription>
              This will permanently remove this asset from the CDE inventory. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeletingId(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteLoading}>
              {deleteLoading ? 'Deleting...' : 'Delete Asset'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
