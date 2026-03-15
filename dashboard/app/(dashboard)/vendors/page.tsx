'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
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
import {
  Search, Plus, ChevronLeft, ChevronRight, Eye, Truck,
  CheckCircle2, AlertCircle, Clock, ShieldAlert,
} from 'lucide-react';
import {
  ServiceProvider, SPComplianceSummary, SPType, SPComplianceStatus, SPRiskLevel,
  listServiceProviders, getSPComplianceSummary, createServiceProvider,
} from '@/lib/api';
import { WikiHelpLink } from '@/components/wiki-help-link';
import {
  SP_TYPE_LABELS,
  SP_COMPLIANCE_LABELS, SP_COMPLIANCE_COLORS,
  SP_RISK_LABELS, SP_RISK_COLORS,
} from '@/components/vendor/constants';

export default function VendorsPage() {
  const { hasRole } = useAuth();
  const canManage = hasRole('ciso', 'compliance_manager', 'vendor_manager');

  const [providers, setProviders] = useState<ServiceProvider[]>([]);
  const [summary, setSummary] = useState<SPComplianceSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [riskFilter, setRiskFilter] = useState('');
  const [activeFilter, setActiveFilter] = useState('');

  // Create dialog
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState('');
  const [newType, setNewType] = useState<SPType>('other');
  const [newContactName, setNewContactName] = useState('');
  const [newContactEmail, setNewContactEmail] = useState('');
  const [newRiskLevel, setNewRiskLevel] = useState<SPRiskLevel>('medium');
  const [newComplianceStatus, setNewComplianceStatus] = useState<SPComplianceStatus>('unknown');
  const [newServicesProvided, setNewServicesProvided] = useState('');

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
      };
      if (search) params.search = search;
      if (typeFilter) params.type = typeFilter;
      if (statusFilter) params.pci_compliance_status = statusFilter;
      if (riskFilter) params.risk_level = riskFilter;
      if (activeFilter) params.is_active = activeFilter;

      const [listRes, summaryRes] = await Promise.all([
        listServiceProviders(params),
        getSPComplianceSummary(),
      ]);
      setProviders(listRes.data);
      setTotal(listRes.meta?.total ?? listRes.data.length);
      setSummary(summaryRes.data);
    } catch (err) {
      console.error('Failed to fetch service providers:', err);
    } finally {
      setLoading(false);
    }
  }, [page, search, typeFilter, statusFilter, riskFilter, activeFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCreate = async () => {
    if (!newName.trim() || !newType) return;
    try {
      setCreating(true);
      await createServiceProvider({
        name: newName.trim(),
        type: newType,
        contact_name: newContactName.trim() || undefined,
        contact_email: newContactEmail.trim() || undefined,
        risk_level: newRiskLevel,
        pci_compliance_status: newComplianceStatus,
        services_provided: newServicesProvided.trim() || undefined,
      });
      setShowCreate(false);
      setNewName('');
      setNewType('other');
      setNewContactName('');
      setNewContactEmail('');
      setNewRiskLevel('medium');
      setNewComplianceStatus('unknown');
      setNewServicesProvided('');
      fetchData();
    } catch (err) {
      console.error('Create failed:', err);
      alert(err instanceof Error ? err.message : 'Failed to create service provider');
    } finally {
      setCreating(false);
    }
  };

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <Truck className="h-6 w-6" /> Vendor Management
          </h1>
          <WikiHelpLink path="pci-dss/vendor-management/" />
          <p className="text-sm text-muted-foreground">
            Service provider registry — PCI DSS v4.0.1 Req 12.8 &amp; 12.9
          </p>
        </div>
        {canManage && (
          <Button onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4 mr-2" /> Add Provider
          </Button>
        )}
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid gap-4 md:grid-cols-5">
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{summary.total}</div>
              <p className="text-xs text-muted-foreground">Total Providers</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-1">
                <CheckCircle2 className="h-4 w-4 text-green-600" />
                <div className="text-2xl font-bold text-green-600">{summary.compliant}</div>
              </div>
              <p className="text-xs text-muted-foreground">Compliant</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-1">
                <AlertCircle className="h-4 w-4 text-red-600" />
                <div className="text-2xl font-bold text-red-600">{summary.non_compliant}</div>
              </div>
              <p className="text-xs text-muted-foreground">Non-Compliant</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-1">
                <Clock className="h-4 w-4 text-yellow-600" />
                <div className="text-2xl font-bold text-yellow-600">
                  {summary.expiring_soon?.within_30_days ?? 0}
                </div>
              </div>
              <p className="text-xs text-muted-foreground">Expiring in 30d</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-1">
                <ShieldAlert className="h-4 w-4 text-orange-600" />
                <div className="text-2xl font-bold text-orange-600">
                  {(summary.by_risk_level?.critical ?? 0) + (summary.by_risk_level?.high ?? 0)}
                </div>
              </div>
              <p className="text-xs text-muted-foreground">Critical / High Risk</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Needs Attention Banner */}
      {summary && summary.needs_attention && summary.needs_attention.length > 0 && (
        <Card className="border-orange-200 bg-orange-50 dark:border-orange-900/50 dark:bg-orange-900/10">
          <CardContent className="p-4">
            <p className="text-sm font-semibold text-orange-800 dark:text-orange-400 mb-2">
              Providers Needing Attention ({summary.needs_attention.length})
            </p>
            <div className="flex flex-wrap gap-2">
              {summary.needs_attention.slice(0, 8).map((sp) => (
                <Link key={sp.id} href={`/vendors/${sp.id}`}>
                  <span className="inline-flex items-center gap-1 text-xs bg-white dark:bg-gray-900 border border-orange-200 dark:border-orange-800 rounded px-2 py-1 hover:bg-orange-50 dark:hover:bg-orange-900/20 transition-colors">
                    {sp.name}
                    <span className="text-muted-foreground">— {sp.reason}</span>
                  </span>
                </Link>
              ))}
              {summary.needs_attention.length > 8 && (
                <span className="text-xs text-muted-foreground self-center">
                  +{summary.needs_attention.length - 8} more
                </span>
              )}
            </div>
          </CardContent>
        </Card>
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
                  placeholder="Search providers..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="w-[180px]">
              <Label className="text-xs">Type</Label>
              <Select value={typeFilter} onValueChange={(v) => { setTypeFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All types" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All types</SelectItem>
                  {(Object.entries(SP_TYPE_LABELS) as [SPType, string][]).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[170px]">
              <Label className="text-xs">PCI Status</Label>
              <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All statuses" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All statuses</SelectItem>
                  {(Object.entries(SP_COMPLIANCE_LABELS) as [SPComplianceStatus, string][]).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[140px]">
              <Label className="text-xs">Risk Level</Label>
              <Select value={riskFilter} onValueChange={(v) => { setRiskFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  {(Object.entries(SP_RISK_LABELS) as [SPRiskLevel, string][]).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[120px]">
              <Label className="text-xs">Active</Label>
              <Select value={activeFilter} onValueChange={(v) => { setActiveFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  <SelectItem value="true">Active</SelectItem>
                  <SelectItem value="false">Inactive</SelectItem>
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
                <TableHead>Provider</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>PCI Compliance</TableHead>
                <TableHead>Risk Level</TableHead>
                <TableHead>Contact</TableHead>
                <TableHead>Next Review</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                    Loading...
                  </TableCell>
                </TableRow>
              ) : providers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                    No service providers found
                  </TableCell>
                </TableRow>
              ) : (
                providers.map((sp) => (
                  <TableRow key={sp.id}>
                    <TableCell>
                      <Link
                        href={`/vendors/${sp.id}`}
                        className="text-primary hover:underline font-medium"
                      >
                        {sp.name}
                      </Link>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {SP_TYPE_LABELS[sp.type] ?? sp.type}
                    </TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SP_COMPLIANCE_COLORS[sp.pci_compliance_status]}`}>
                        {SP_COMPLIANCE_LABELS[sp.pci_compliance_status]}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SP_RISK_COLORS[sp.risk_level]}`}>
                        {SP_RISK_LABELS[sp.risk_level]}
                      </span>
                    </TableCell>
                    <TableCell className="text-sm">
                      {sp.contact_name
                        ? <span>{sp.contact_name}{sp.contact_email ? <span className="text-muted-foreground ml-1">({sp.contact_email})</span> : null}</span>
                        : <span className="text-muted-foreground">—</span>
                      }
                    </TableCell>
                    <TableCell className="text-sm">
                      {sp.next_review_date
                        ? new Date(sp.next_review_date).toLocaleDateString()
                        : <span className="text-muted-foreground">—</span>
                      }
                    </TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${sp.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                        {sp.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </TableCell>
                    <TableCell className="text-right">
                      <Link href={`/vendors/${sp.id}`}>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <Eye className="h-4 w-4" />
                        </Button>
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

      {/* Create Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Add Service Provider</DialogTitle>
            <DialogDescription>
              Register a new vendor or service provider (PCI DSS Req 12.8)
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Provider Name *</Label>
              <Input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="e.g. Stripe, AWS, Salesforce"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Type *</Label>
                <Select value={newType} onValueChange={(v) => setNewType(v as SPType)}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {(Object.entries(SP_TYPE_LABELS) as [SPType, string][]).map(([k, v]) => (
                      <SelectItem key={k} value={k}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Risk Level</Label>
                <Select value={newRiskLevel} onValueChange={(v) => setNewRiskLevel(v as SPRiskLevel)}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {(Object.entries(SP_RISK_LABELS) as [SPRiskLevel, string][]).map(([k, v]) => (
                      <SelectItem key={k} value={k}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div>
              <Label>PCI Compliance Status</Label>
              <Select value={newComplianceStatus} onValueChange={(v) => setNewComplianceStatus(v as SPComplianceStatus)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {(Object.entries(SP_COMPLIANCE_LABELS) as [SPComplianceStatus, string][]).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Contact Name</Label>
                <Input
                  value={newContactName}
                  onChange={(e) => setNewContactName(e.target.value)}
                  placeholder="Jane Smith"
                />
              </div>
              <div>
                <Label>Contact Email</Label>
                <Input
                  type="email"
                  value={newContactEmail}
                  onChange={(e) => setNewContactEmail(e.target.value)}
                  placeholder="jane@vendor.com"
                />
              </div>
            </div>
            <div>
              <Label>Services Provided</Label>
              <Input
                value={newServicesProvided}
                onChange={(e) => setNewServicesProvided(e.target.value)}
                placeholder="Brief description of services..."
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={creating || !newName.trim()}>
              {creating ? 'Adding...' : 'Add Provider'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
