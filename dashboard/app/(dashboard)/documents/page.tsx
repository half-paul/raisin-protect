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
  Search, Plus, ChevronLeft, ChevronRight, Eye, ScrollText,
  FileCheck, FileClock, FileX, Download,
} from 'lucide-react';
import {
  ComplianceDocument, ComplianceDocumentType, ComplianceDocumentStatus,
  listComplianceDocuments, createComplianceDocument,
} from '@/lib/api';
import { WikiHelpLink } from '@/components/wiki-help-link';
import {
  DOC_TYPE_LABELS, DOC_TYPE_DESCRIPTIONS,
  DOC_STATUS_LABELS, DOC_STATUS_COLORS,
  DOC_TERMINAL_STATUSES,
} from '@/components/aoc/constants';

export default function DocumentsPage() {
  const { hasRole } = useAuth();
  const canCreate = hasRole('ciso', 'compliance_manager');

  const [documents, setDocuments] = useState<ComplianceDocument[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [searchInput, setSearchInput] = useState('');
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  // Stats (derived)
  const [allDocs, setAllDocs] = useState<ComplianceDocument[]>([]);

  // Create wizard
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [step, setStep] = useState<1 | 2>(1);
  const [newDocType, setNewDocType] = useState<ComplianceDocumentType>('aoc_saq_a');
  const [newTitle, setNewTitle] = useState('');
  const [newPeriodStart, setNewPeriodStart] = useState('');
  const [newPeriodEnd, setNewPeriodEnd] = useState('');
  const [newMerchantName, setNewMerchantName] = useState('');

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
      };
      if (search) params.search = search;
      if (typeFilter) params.document_type = typeFilter;
      if (statusFilter) params.doc_status = statusFilter;

      const [listRes, allRes] = await Promise.all([
        listComplianceDocuments(params),
        // Fetch unfiltered for stats
        listComplianceDocuments({ per_page: '200' }),
      ]);
      setDocuments(listRes.data);
      setTotal(listRes.meta?.total ?? listRes.data.length);
      setAllDocs(allRes.data);
    } catch (err) {
      console.error('Failed to fetch documents:', err);
    } finally {
      setLoading(false);
    }
  }, [page, search, typeFilter, statusFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleCreate = async () => {
    if (!newTitle.trim() || !newPeriodStart || !newPeriodEnd) return;
    try {
      setCreating(true);
      const doc = await createComplianceDocument({
        document_type: newDocType,
        title: newTitle.trim(),
        assessment_period_start: newPeriodStart,
        assessment_period_end: newPeriodEnd,
        merchant_name: newMerchantName.trim() || undefined,
        pci_dss_version: '4.0.1',
      });
      setShowCreate(false);
      setStep(1);
      setNewTitle('');
      setNewPeriodStart('');
      setNewPeriodEnd('');
      setNewMerchantName('');
      // Navigate to the builder
      window.location.href = `/documents/${doc.data.id}`;
    } catch (err) {
      console.error('Create failed:', err);
      alert(err instanceof Error ? err.message : 'Failed to create document');
    } finally {
      setCreating(false);
    }
  };

  const openCreate = () => {
    setStep(1);
    setNewDocType('aoc_saq_a');
    setNewTitle('');
    setNewPeriodStart('');
    setNewPeriodEnd('');
    setNewMerchantName('');
    setShowCreate(true);
  };

  const totalPages = Math.ceil(total / perPage);

  // Derived stats
  const statFinal = allDocs.filter(d => d.doc_status === 'final' || d.doc_status === 'signed').length;
  const statInProgress = allDocs.filter(d => !DOC_TERMINAL_STATUSES.includes(d.doc_status) && d.doc_status !== 'draft').length;
  const statDraft = allDocs.filter(d => d.doc_status === 'draft').length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold flex items-center gap-2">
            <ScrollText className="h-6 w-6" /> AOC / ROC Documents
          </h1>
          <WikiHelpLink path="pci-dss/aoc-roc-documents/" />
          <p className="text-sm text-muted-foreground">
            Attestation of Compliance and Report on Compliance — PCI DSS v4.0.1 Req 12.4
          </p>
        </div>
        {canCreate && (
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4 mr-2" /> New Document
          </Button>
        )}
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-2xl font-bold">{allDocs.length}</div>
            <p className="text-xs text-muted-foreground">Total Documents</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-1">
              <FileCheck className="h-4 w-4 text-green-600" />
              <div className="text-2xl font-bold text-green-600">{statFinal}</div>
            </div>
            <p className="text-xs text-muted-foreground">Final / Signed</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-1">
              <FileClock className="h-4 w-4 text-yellow-600" />
              <div className="text-2xl font-bold text-yellow-600">{statInProgress}</div>
            </div>
            <p className="text-xs text-muted-foreground">In Progress</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-1">
              <FileX className="h-4 w-4 text-gray-400" />
              <div className="text-2xl font-bold text-gray-500">{statDraft}</div>
            </div>
            <p className="text-xs text-muted-foreground">Draft</p>
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
                  placeholder="Search documents..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="w-[200px]">
              <Label className="text-xs">Document Type</Label>
              <Select value={typeFilter} onValueChange={(v) => { setTypeFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All types" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All types</SelectItem>
                  {(Object.entries(DOC_TYPE_LABELS) as [ComplianceDocumentType, string][]).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[160px]">
              <Label className="text-xs">Status</Label>
              <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All statuses" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All statuses</SelectItem>
                  {(Object.entries(DOC_STATUS_LABELS) as [ComplianceDocumentStatus, string][]).map(([k, v]) => (
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
                <TableHead>Title</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Assessment Period</TableHead>
                <TableHead>Merchant</TableHead>
                <TableHead>Version</TableHead>
                <TableHead>Created</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Loading...</TableCell>
                </TableRow>
              ) : documents.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                    No compliance documents found
                  </TableCell>
                </TableRow>
              ) : (
                documents.map((doc) => (
                  <TableRow key={doc.id}>
                    <TableCell>
                      <Link href={`/documents/${doc.id}`} className="text-primary hover:underline font-medium">
                        {doc.title}
                      </Link>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground whitespace-nowrap">
                      {DOC_TYPE_LABELS[doc.document_type] ?? doc.document_type}
                    </TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${DOC_STATUS_COLORS[doc.doc_status]}`}>
                        {DOC_STATUS_LABELS[doc.doc_status]}
                      </span>
                    </TableCell>
                    <TableCell className="text-sm whitespace-nowrap">
                      {new Date(doc.assessment_period_start).toLocaleDateString()}
                      {' — '}
                      {new Date(doc.assessment_period_end).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-sm">{doc.merchant_name || '—'}</TableCell>
                    <TableCell className="text-sm text-center">v{doc.version}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(doc.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Link href={`/documents/${doc.id}`}>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <Eye className="h-4 w-4" />
                          </Button>
                        </Link>
                        {doc.pdf_path && (
                          <Button variant="ghost" size="icon" className="h-8 w-8" title="Download PDF">
                            <Download className="h-4 w-4" />
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

      {/* Create Wizard Dialog */}
      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Generate Compliance Document</DialogTitle>
            <DialogDescription>
              {step === 1
                ? 'Step 1 of 2 — Select document type'
                : 'Step 2 of 2 — Assessment details'}
            </DialogDescription>
          </DialogHeader>

          {step === 1 && (
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">Select the type of compliance document to generate:</p>
              <div className="grid gap-2">
                {(Object.entries(DOC_TYPE_LABELS) as [ComplianceDocumentType, string][]).map(([k, v]) => (
                  <button
                    key={k}
                    type="button"
                    onClick={() => setNewDocType(k)}
                    className={`flex items-start gap-3 p-3 rounded-md border text-left transition-colors ${
                      newDocType === k
                        ? 'border-primary bg-primary/5'
                        : 'border-border hover:bg-muted/50'
                    }`}
                  >
                    <div className="flex-1">
                      <div className="font-medium text-sm">{v}</div>
                      <div className="text-xs text-muted-foreground mt-0.5">
                        {DOC_TYPE_DESCRIPTIONS[k]}
                      </div>
                    </div>
                    {newDocType === k && (
                      <div className="h-4 w-4 rounded-full bg-primary mt-0.5 shrink-0" />
                    )}
                  </button>
                ))}
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4">
              <div>
                <Label>Document Title *</Label>
                <Input
                  value={newTitle}
                  onChange={(e) => setNewTitle(e.target.value)}
                  placeholder={`e.g. ${new Date().getFullYear()} PCI DSS ${DOC_TYPE_LABELS[newDocType]}`}
                />
              </div>
              <div>
                <Label>Merchant / Company Name</Label>
                <Input
                  value={newMerchantName}
                  onChange={(e) => setNewMerchantName(e.target.value)}
                  placeholder="Your organization name"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label>Assessment Period Start *</Label>
                  <Input
                    type="date"
                    value={newPeriodStart}
                    onChange={(e) => setNewPeriodStart(e.target.value)}
                  />
                </div>
                <div>
                  <Label>Assessment Period End *</Label>
                  <Input
                    type="date"
                    value={newPeriodEnd}
                    onChange={(e) => setNewPeriodEnd(e.target.value)}
                  />
                </div>
              </div>
              <p className="text-xs text-muted-foreground">
                PCI DSS Version: 4.0.1 (current)
              </p>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => {
              if (step === 2) setStep(1);
              else setShowCreate(false);
            }}>
              {step === 1 ? 'Cancel' : 'Back'}
            </Button>
            {step === 1 ? (
              <Button onClick={() => setStep(2)}>
                Next →
              </Button>
            ) : (
              <Button
                onClick={handleCreate}
                disabled={creating || !newTitle.trim() || !newPeriodStart || !newPeriodEnd}
              >
                {creating ? 'Creating...' : 'Create & Open Builder'}
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
