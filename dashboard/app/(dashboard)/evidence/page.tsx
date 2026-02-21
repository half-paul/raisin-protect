'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
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
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  ClipboardCheck, Search, Plus, MoreVertical, ChevronLeft, ChevronRight,
  Upload, FileText, AlertTriangle, CheckCircle2, XCircle, Download,
  Eye, Trash2, Clock,
} from 'lucide-react';
import {
  EvidenceArtifact, FreshnessSummary,
  listEvidence, getFreshnessSummary, createEvidence, confirmEvidenceUpload,
  deleteEvidence, changeEvidenceStatus,
} from '@/lib/api';
import {
  FreshnessBadge,
  EVIDENCE_TYPE_LABELS, EVIDENCE_STATUS_LABELS, EVIDENCE_STATUS_COLORS,
  COLLECTION_METHOD_LABELS, formatFileSize,
} from '@/components/evidence/freshness-badge';

const ALLOWED_MIME_TYPES: Record<string, string> = {
  'application/pdf': '.pdf',
  'application/json': '.json',
  'text/csv': '.csv',
  'text/plain': '.txt',
  'image/png': '.png',
  'image/jpeg': '.jpg',
  'image/gif': '.gif',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': '.xlsx',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': '.docx',
  'application/xml': '.xml',
  'text/xml': '.xml',
  'application/zip': '.zip',
};

export default function EvidenceLibraryPage() {
  const { hasRole } = useAuth();
  const canUpload = hasRole('ciso', 'compliance_manager', 'security_engineer', 'it_admin', 'devops_engineer');
  const canManage = hasRole('ciso', 'compliance_manager');

  const [evidence, setEvidence] = useState<EvidenceArtifact[]>([]);
  const [summary, setSummary] = useState<FreshnessSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [freshnessFilter, setFreshnessFilter] = useState('');

  // Upload dialog
  const [showUpload, setShowUpload] = useState(false);
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [uploadDragging, setUploadDragging] = useState(false);
  const [uploadForm, setUploadForm] = useState({
    title: '',
    description: '',
    evidence_type: 'other' as string,
    collection_method: 'manual_upload' as string,
    collection_date: new Date().toISOString().split('T')[0],
    freshness_period_days: '',
    source_system: '',
    tags: '',
  });
  const [uploadLoading, setUploadLoading] = useState(false);
  const [uploadError, setUploadError] = useState('');
  const [uploadProgress, setUploadProgress] = useState<string>('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
        sort: 'created_at',
        order: 'desc',
      };
      if (search) params.search = search;
      if (statusFilter) params.status = statusFilter;
      if (typeFilter) params.evidence_type = typeFilter;
      if (freshnessFilter) params.freshness = freshnessFilter;

      const [evRes, sumRes] = await Promise.all([
        listEvidence(params),
        getFreshnessSummary(),
      ]);

      setEvidence(evRes.data || []);
      setTotal(evRes.meta?.total || 0);
      setSummary(sumRes.data);
    } catch {
      // handle
    } finally {
      setLoading(false);
    }
  }, [page, search, statusFilter, typeFilter, freshnessFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  useEffect(() => {
    const timer = setTimeout(() => { setSearch(searchInput); setPage(1); }, 300);
    return () => clearTimeout(timer);
  }, [searchInput]);

  // Drag & drop handlers
  function handleDragOver(e: React.DragEvent) {
    e.preventDefault();
    setUploadDragging(true);
  }
  function handleDragLeave(e: React.DragEvent) {
    e.preventDefault();
    setUploadDragging(false);
  }
  function handleDrop(e: React.DragEvent) {
    e.preventDefault();
    setUploadDragging(false);
    const file = e.dataTransfer.files?.[0];
    if (file) selectFile(file);
  }
  function handleFileInput(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (file) selectFile(file);
  }
  // Maximum file size: 100 MB
  const MAX_FILE_SIZE = 100 * 1024 * 1024;

  function selectFile(file: File) {
    // Validate file size
    if (file.size > MAX_FILE_SIZE) {
      setUploadError(`File size (${formatFileSize(file.size)}) exceeds maximum allowed (100 MB)`);
      return;
    }

    setUploadError(''); // Clear any previous error
    setUploadFile(file);
    if (!uploadForm.title) {
      setUploadForm(f => ({ ...f, title: file.name.replace(/\.[^.]+$/, '').replace(/[-_]/g, ' ') }));
    }
    // Auto-detect type
    if (file.name.match(/\.pdf$/i) && uploadForm.evidence_type === 'other') {
      setUploadForm(f => ({ ...f, evidence_type: 'policy_document' }));
    } else if (file.name.match(/\.(png|jpg|jpeg|gif)$/i) && uploadForm.evidence_type === 'other') {
      setUploadForm(f => ({ ...f, evidence_type: 'screenshot' }));
    } else if (file.name.match(/\.json$/i) && uploadForm.evidence_type === 'other') {
      setUploadForm(f => ({ ...f, evidence_type: 'configuration_export' }));
    }
  }

  async function handleUpload() {
    if (!uploadFile) return;
    setUploadError('');
    setUploadLoading(true);
    setUploadProgress('Creating record...');

    try {
      // Step 1: Create evidence record
      const mime = uploadFile.type || 'application/octet-stream';
      const tags = uploadForm.tags ? uploadForm.tags.split(',').map(t => t.trim()).filter(Boolean) : [];
      const res = await createEvidence({
        title: uploadForm.title,
        description: uploadForm.description || undefined,
        evidence_type: uploadForm.evidence_type,
        collection_method: uploadForm.collection_method,
        file_name: uploadFile.name,
        file_size: uploadFile.size,
        mime_type: mime,
        collection_date: uploadForm.collection_date,
        freshness_period_days: uploadForm.freshness_period_days ? parseInt(uploadForm.freshness_period_days) : undefined,
        source_system: uploadForm.source_system || undefined,
        tags: tags.length > 0 ? tags : undefined,
      });

      // Step 2: Upload file to MinIO via presigned URL
      setUploadProgress('Uploading file...');
      const upload = res.data.upload;
      if (upload?.presigned_url) {
        const uploadRes = await fetch(upload.presigned_url, {
          method: upload.method || 'PUT',
          headers: { 'Content-Type': mime },
          body: uploadFile,
        });
        if (!uploadRes.ok) throw new Error('File upload failed');
      }

      // Step 3: Confirm upload
      setUploadProgress('Confirming...');
      await confirmEvidenceUpload(res.data.id);

      setShowUpload(false);
      resetUploadForm();
      fetchData();
    } catch (err) {
      setUploadError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploadLoading(false);
      setUploadProgress('');
    }
  }

  function resetUploadForm() {
    setUploadFile(null);
    setUploadForm({
      title: '', description: '', evidence_type: 'other',
      collection_method: 'manual_upload',
      collection_date: new Date().toISOString().split('T')[0],
      freshness_period_days: '', source_system: '', tags: '',
    });
    setUploadError('');
  }

  async function handleDelete(id: string) {
    if (!confirm('Remove this evidence from active view? Files are retained for audit.')) return;
    try {
      await deleteEvidence(id);
      fetchData();
    } catch { /* handle */ }
  }

  async function handleStatusChange(id: string, status: string) {
    try {
      await changeEvidenceStatus(id, status);
      fetchData();
    } catch { /* handle */ }
  }

  const totalPages = Math.ceil(total / perPage);

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <ClipboardCheck className="h-8 w-8" />
            Evidence Library
          </h1>
          <p className="text-muted-foreground mt-1">
            Upload, manage, and track compliance evidence artifacts
          </p>
        </div>
        <div className="flex gap-2">
          <Link href="/staleness">
            <Button variant="outline">
              <AlertTriangle className="h-4 w-4 mr-2" />
              Staleness Alerts
            </Button>
          </Link>
          {canUpload && (
            <Button onClick={() => setShowUpload(true)}>
              <Upload className="h-4 w-4 mr-2" />
              Upload Evidence
            </Button>
          )}
        </div>
      </div>

      {/* Summary cards */}
      {summary && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold">{summary.total_evidence}</div>
              <p className="text-xs text-muted-foreground">Total Artifacts</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                {summary.by_freshness.fresh}
              </div>
              <p className="text-xs text-muted-foreground">Fresh</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-amber-600 dark:text-amber-400">
                {summary.by_freshness.expiring_soon}
              </div>
              <p className="text-xs text-muted-foreground">Expiring Soon</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-red-600 dark:text-red-400">
                {summary.by_freshness.expired}
              </div>
              <p className="text-xs text-muted-foreground">Expired</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                {summary.coverage.evidence_coverage_pct.toFixed(0)}%
              </div>
              <p className="text-xs text-muted-foreground">Control Coverage</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search evidence..."
            className="pl-10"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </div>
        <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Statuses</SelectItem>
            {Object.entries(EVIDENCE_STATUS_LABELS).map(([k, v]) => (
              <SelectItem key={k} value={k}>{v}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={typeFilter} onValueChange={(v) => { setTypeFilter(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            {Object.entries(EVIDENCE_TYPE_LABELS).map(([k, v]) => (
              <SelectItem key={k} value={k}>{v}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={freshnessFilter} onValueChange={(v) => { setFreshnessFilter(v === 'all' ? '' : v); setPage(1); }}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Freshness" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="fresh">Fresh</SelectItem>
            <SelectItem value="expiring_soon">Expiring Soon</SelectItem>
            <SelectItem value="expired">Expired</SelectItem>
          </SelectContent>
        </Select>
        <span className="text-sm text-muted-foreground">{total} artifacts</span>
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
            </div>
          ) : evidence.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <ClipboardCheck className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">No evidence found</p>
              <p className="text-sm">Upload your first evidence artifact to get started</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Title</TableHead>
                  <TableHead className="w-[120px]">Type</TableHead>
                  <TableHead className="w-[110px]">Status</TableHead>
                  <TableHead className="w-[130px]">Freshness</TableHead>
                  <TableHead className="w-[80px]">Size</TableHead>
                  <TableHead className="w-[60px] text-center">Links</TableHead>
                  <TableHead className="w-[100px]">Collected</TableHead>
                  <TableHead className="w-[50px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {evidence.map((ev) => (
                  <TableRow key={ev.id} className="group">
                    <TableCell className="max-w-xs">
                      <Link href={`/evidence/${ev.id}`} className="hover:text-primary">
                        <div className="flex items-center gap-2">
                          <FileText className="h-4 w-4 text-muted-foreground shrink-0" />
                          <div>
                            <span className="line-clamp-1 text-sm font-medium">{ev.title}</span>
                            <span className="text-xs text-muted-foreground line-clamp-1">{ev.file_name}</span>
                          </div>
                        </div>
                      </Link>
                      {ev.version > 1 && (
                        <Badge variant="secondary" className="text-[10px] ml-6 mt-0.5">v{ev.version}</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {EVIDENCE_TYPE_LABELS[ev.evidence_type] || ev.evidence_type}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className={`text-xs ${EVIDENCE_STATUS_COLORS[ev.status] || ''}`}>
                        {EVIDENCE_STATUS_LABELS[ev.status] || ev.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <FreshnessBadge
                        status={ev.freshness_status}
                        daysUntilExpiry={ev.days_until_expiry}
                      />
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {formatFileSize(ev.file_size)}
                    </TableCell>
                    <TableCell className="text-center">
                      <span className="text-sm font-medium">{ev.links_count || 0}</span>
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {new Date(ev.collection_date).toLocaleDateString()}
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
                            <Link href={`/evidence/${ev.id}`}>
                              <Eye className="h-4 w-4 mr-2" />View Details
                            </Link>
                          </DropdownMenuItem>
                          {ev.status === 'draft' && canManage && (
                            <DropdownMenuItem onClick={() => handleStatusChange(ev.id, 'pending_review')}>
                              <Clock className="h-4 w-4 mr-2" />Submit for Review
                            </DropdownMenuItem>
                          )}
                          {ev.status === 'pending_review' && canManage && (
                            <>
                              <DropdownMenuItem onClick={() => handleStatusChange(ev.id, 'approved')}>
                                <CheckCircle2 className="h-4 w-4 mr-2" />Approve
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => handleStatusChange(ev.id, 'rejected')}>
                                <XCircle className="h-4 w-4 mr-2" />Reject
                              </DropdownMenuItem>
                            </>
                          )}
                          {canManage && (
                            <DropdownMenuItem className="text-destructive" onClick={() => handleDelete(ev.id)}>
                              <Trash2 className="h-4 w-4 mr-2" />Remove
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
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-muted-foreground">Page {page} of {totalPages}</span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Upload Evidence Dialog */}
      <Dialog open={showUpload} onOpenChange={(o) => { if (!o) resetUploadForm(); setShowUpload(o); }}>
        <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Upload Evidence</DialogTitle>
            <DialogDescription>
              Upload a file and provide metadata for your evidence artifact
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {uploadError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
                <AlertTriangle className="h-4 w-4" />
                {uploadError}
              </div>
            )}

            {/* Drag & drop zone */}
            <div
              className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer ${
                uploadDragging ? 'border-primary bg-primary/5' : 'border-muted-foreground/25 hover:border-primary/50'
              }`}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
            >
              <input
                ref={fileInputRef}
                type="file"
                className="hidden"
                onChange={handleFileInput}
                accept={Object.values(ALLOWED_MIME_TYPES).join(',')}
              />
              {uploadFile ? (
                <div className="flex items-center justify-center gap-3">
                  <FileText className="h-8 w-8 text-primary" />
                  <div className="text-left">
                    <p className="font-medium">{uploadFile.name}</p>
                    <p className="text-sm text-muted-foreground">{formatFileSize(uploadFile.size)}</p>
                  </div>
                  <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); setUploadFile(null); }}>
                    <XCircle className="h-4 w-4" />
                  </Button>
                </div>
              ) : (
                <>
                  <Upload className="h-10 w-10 mx-auto mb-3 text-muted-foreground" />
                  <p className="font-medium">Drag & drop a file here</p>
                  <p className="text-sm text-muted-foreground mt-1">or click to browse (max 100 MB)</p>
                </>
              )}
            </div>

            <div className="space-y-2">
              <Label>Title *</Label>
              <Input
                placeholder="Evidence title"
                value={uploadForm.title}
                onChange={(e) => setUploadForm(f => ({ ...f, title: e.target.value }))}
              />
            </div>

            <div className="space-y-2">
              <Label>Description</Label>
              <Textarea
                placeholder="Describe this evidence..."
                value={uploadForm.description}
                onChange={(e) => setUploadForm(f => ({ ...f, description: e.target.value }))}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Evidence Type *</Label>
                <Select value={uploadForm.evidence_type} onValueChange={(v) => setUploadForm(f => ({ ...f, evidence_type: v }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {Object.entries(EVIDENCE_TYPE_LABELS).map(([k, v]) => (
                      <SelectItem key={k} value={k}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Collection Method</Label>
                <Select value={uploadForm.collection_method} onValueChange={(v) => setUploadForm(f => ({ ...f, collection_method: v }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {Object.entries(COLLECTION_METHOD_LABELS).map(([k, v]) => (
                      <SelectItem key={k} value={k}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Collection Date *</Label>
                <Input
                  type="date"
                  value={uploadForm.collection_date}
                  onChange={(e) => setUploadForm(f => ({ ...f, collection_date: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label>Freshness Period (days)</Label>
                <Input
                  type="number"
                  placeholder="e.g. 90"
                  value={uploadForm.freshness_period_days}
                  onChange={(e) => setUploadForm(f => ({ ...f, freshness_period_days: e.target.value }))}
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Source System</Label>
                <Input
                  placeholder="e.g. Okta, AWS, Jira"
                  value={uploadForm.source_system}
                  onChange={(e) => setUploadForm(f => ({ ...f, source_system: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label>Tags (comma-separated)</Label>
                <Input
                  placeholder="mfa, okta, q1-2026"
                  value={uploadForm.tags}
                  onChange={(e) => setUploadForm(f => ({ ...f, tags: e.target.value }))}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { resetUploadForm(); setShowUpload(false); }}>Cancel</Button>
            <Button
              onClick={handleUpload}
              disabled={uploadLoading || !uploadFile || !uploadForm.title || !uploadForm.collection_date}
            >
              {uploadLoading ? (
                <>{uploadProgress || 'Uploading...'}</>
              ) : (
                <><Upload className="h-4 w-4 mr-2" />Upload Evidence</>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
