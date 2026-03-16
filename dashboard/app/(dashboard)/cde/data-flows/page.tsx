'use client';

import { useState, useEffect, useCallback } from 'react';
import { Plus, Search, MoreVertical, AlertCircle, ArrowRightLeft } from 'lucide-react';
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
  DataFlow,
  DataFlowProtocol,
  listCdeAssets,
  listDataFlows,
  createDataFlow,
  updateDataFlow,
  deleteDataFlow,
} from '@/lib/api';
import { PROTOCOL_LABELS } from '@/components/cde/constants';
import { useAuth } from '@/lib/auth-context';

const PROTOCOLS: DataFlowProtocol[] = ['https', 'http', 'tls', 'ssh', 'sftp', 'smb', 'ftp', 'other'];

const emptyForm = {
  source_asset_id: '',
  dest_asset_id: '',
  protocol: 'https' as DataFlowProtocol,
  port: '',
  data_type: '',
  encryption_method: '',
};

export default function DataFlowsPage() {
  const { hasRole } = useAuth();
  const canWrite = hasRole('ciso', 'compliance_manager', 'security_engineer', 'it_admin');

  const [flows, setFlows] = useState<DataFlow[]>([]);
  const [assets, setAssets] = useState<CdeAsset[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');
  const [protocolFilter, setProtocolFilter] = useState('');

  const [showDialog, setShowDialog] = useState(false);
  const [editingFlow, setEditingFlow] = useState<DataFlow | null>(null);
  const [form, setForm] = useState(emptyForm);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState('');

  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);

  // Build a map for quick asset name lookup
  const assetMap = Object.fromEntries(assets.map((a) => [a.id, a]));

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = { page: String(page), per_page: String(perPage) };
      if (search) params.search = search;
      if (protocolFilter) params.protocol = protocolFilter;
      const res = await listDataFlows(params);
      setFlows(res.data);
      setTotal(res.meta?.total ?? res.data.length);
    } catch (err) {
      console.error('Failed to fetch data flows:', err);
    } finally {
      setLoading(false);
    }
  }, [page, search, protocolFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  // Load all assets once for the pickers
  useEffect(() => {
    listCdeAssets({ per_page: '500' })
      .then((res) => setAssets(res.data))
      .catch((err) => console.error('Failed to fetch assets:', err));
  }, []);

  function openCreate() {
    setEditingFlow(null);
    setForm(emptyForm);
    setFormError('');
    setShowDialog(true);
  }

  function openEdit(flow: DataFlow) {
    setEditingFlow(flow);
    setForm({
      source_asset_id: flow.source_asset_id,
      dest_asset_id: flow.dest_asset_id,
      protocol: flow.protocol ?? 'https',
      port: flow.port != null ? String(flow.port) : '',
      data_type: flow.data_type ?? '',
      encryption_method: flow.encryption_method ?? '',
    });
    setFormError('');
    setShowDialog(true);
  }

  async function handleSubmit() {
    setFormError('');
    setFormLoading(true);
    try {
      const portNum = form.port ? parseInt(form.port, 10) : undefined;
      const payload = {
        source_asset_id: form.source_asset_id,
        dest_asset_id: form.dest_asset_id,
        protocol: form.protocol || undefined,
        port: Number.isFinite(portNum) ? portNum : undefined,
        data_type: form.data_type || undefined,
        encryption_method: form.encryption_method || undefined,
      };
      if (editingFlow) {
        await updateDataFlow(editingFlow.id, payload);
      } else {
        await createDataFlow(payload);
      }
      setShowDialog(false);
      fetchData();
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to save data flow');
    } finally {
      setFormLoading(false);
    }
  }

  async function handleDelete() {
    if (!deletingId) return;
    setDeleteLoading(true);
    try {
      await deleteDataFlow(deletingId);
      setDeletingId(null);
      fetchData();
    } catch (err) {
      console.error('Failed to delete data flow:', err);
    } finally {
      setDeleteLoading(false);
    }
  }

  const encryptedCount = flows.filter((f) => !!f.encryption_method).length;
  const unencryptedCount = flows.filter((f) => !f.encryption_method).length;
  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Data Flows</h1>
          <p className="text-muted-foreground text-sm mt-1">
            Track cardholder data flows between systems per PCI DSS Req 1.2 & 4.2
          </p>
        </div>
        {canWrite && (
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4 mr-2" />
            Add Data Flow
          </Button>
        )}
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold">{total}</div>
            <p className="text-xs text-muted-foreground">Total Data Flows</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className={`text-2xl font-bold ${encryptedCount === total && total > 0 ? 'text-green-600' : 'text-muted-foreground'}`}>
              {encryptedCount}
            </div>
            <p className="text-xs text-muted-foreground">Flows with Encryption Method</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className={`text-2xl font-bold ${unencryptedCount > 0 ? 'text-yellow-600' : 'text-muted-foreground'}`}>
              {unencryptedCount}
            </div>
            <p className="text-xs text-muted-foreground">Flows without Encryption Method</p>
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
                  placeholder="Search flows..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="w-[150px]">
              <Label className="text-xs">Protocol</Label>
              <Select value={protocolFilter || 'all'} onValueChange={(v) => { setProtocolFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All protocols</SelectItem>
                  {PROTOCOLS.map((p) => (
                    <SelectItem key={p} value={p}>{PROTOCOL_LABELS[p]}</SelectItem>
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
                <TableHead>Source Asset</TableHead>
                <TableHead>Destination Asset</TableHead>
                <TableHead>Protocol</TableHead>
                <TableHead>Port</TableHead>
                <TableHead>Data Type</TableHead>
                <TableHead>Encryption Method</TableHead>
                {canWrite && <TableHead className="w-[50px]" />}
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={canWrite ? 7 : 6} className="text-center py-8 text-muted-foreground">
                    Loading data flows...
                  </TableCell>
                </TableRow>
              ) : flows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={canWrite ? 7 : 6} className="text-center py-8">
                    <ArrowRightLeft className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                    <p className="text-muted-foreground text-sm">No data flows documented</p>
                  </TableCell>
                </TableRow>
              ) : (
                flows.map((flow) => (
                  <TableRow key={flow.id}>
                    <TableCell className="text-sm font-medium max-w-[140px] truncate">
                      {assetMap[flow.source_asset_id]?.name ?? flow.source_asset_id}
                    </TableCell>
                    <TableCell className="text-sm font-medium max-w-[140px] truncate">
                      {assetMap[flow.dest_asset_id]?.name ?? flow.dest_asset_id}
                    </TableCell>
                    <TableCell>
                      {flow.protocol ? (
                        <Badge variant="outline" className="text-xs font-mono">
                          {PROTOCOL_LABELS[flow.protocol]}
                        </Badge>
                      ) : (
                        <span className="text-muted-foreground text-sm">—</span>
                      )}
                    </TableCell>
                    <TableCell className="text-sm font-mono text-muted-foreground">
                      {flow.port ?? '—'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {flow.data_type || '—'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {flow.encryption_method || (
                        <span className="text-yellow-600 font-medium">None</span>
                      )}
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
                            <DropdownMenuItem onClick={() => openEdit(flow)}>Edit</DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className="text-destructive"
                              onClick={() => setDeletingId(flow.id)}
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
        <DialogContent className="sm:max-w-[540px]">
          <DialogHeader>
            <DialogTitle>{editingFlow ? 'Edit Data Flow' : 'Add Data Flow'}</DialogTitle>
            <DialogDescription>
              Document a cardholder data flow between two CDE assets.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            {formError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
                <AlertCircle className="h-4 w-4 shrink-0" />
                {formError}
              </div>
            )}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Source Asset <span className="text-destructive">*</span></Label>
                <Select
                  value={form.source_asset_id || 'none'}
                  onValueChange={(v) => setForm((f) => ({ ...f, source_asset_id: v === 'none' ? '' : v }))}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select source asset" />
                  </SelectTrigger>
                  <SelectContent>
                    {assets.length === 0 && (
                      <SelectItem value="none" disabled>No assets available</SelectItem>
                    )}
                    {assets.map((a) => (
                      <SelectItem key={a.id} value={a.id}>{a.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Destination Asset <span className="text-destructive">*</span></Label>
                <Select
                  value={form.dest_asset_id || 'none'}
                  onValueChange={(v) => setForm((f) => ({ ...f, dest_asset_id: v === 'none' ? '' : v }))}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select destination asset" />
                  </SelectTrigger>
                  <SelectContent>
                    {assets.length === 0 && (
                      <SelectItem value="none" disabled>No assets available</SelectItem>
                    )}
                    {assets.map((a) => (
                      <SelectItem key={a.id} value={a.id}>{a.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Protocol</Label>
                <Select value={form.protocol} onValueChange={(v) => setForm((f) => ({ ...f, protocol: v as DataFlowProtocol }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {PROTOCOLS.map((p) => (
                      <SelectItem key={p} value={p}>{PROTOCOL_LABELS[p]}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Port</Label>
                <Input
                  type="number"
                  placeholder="443"
                  value={form.port}
                  onChange={(e) => setForm((f) => ({ ...f, port: e.target.value }))}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Data Type</Label>
              <Input
                placeholder="e.g. PAN, SAD, Cardholder Data"
                value={form.data_type}
                onChange={(e) => setForm((f) => ({ ...f, data_type: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label>Encryption Method</Label>
              <Input
                placeholder="e.g. TLS 1.3, AES-256"
                value={form.encryption_method}
                onChange={(e) => setForm((f) => ({ ...f, encryption_method: e.target.value }))}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowDialog(false)}>Cancel</Button>
            <Button
              onClick={handleSubmit}
              disabled={formLoading || !form.source_asset_id || !form.dest_asset_id}
            >
              {formLoading ? 'Saving...' : editingFlow ? 'Save Changes' : 'Add Flow'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deletingId} onOpenChange={() => setDeletingId(null)}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>Delete Data Flow</DialogTitle>
            <DialogDescription>
              This will permanently remove this data flow record. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeletingId(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteLoading}>
              {deleteLoading ? 'Deleting...' : 'Delete Flow'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
