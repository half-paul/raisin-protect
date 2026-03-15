'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
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
  ArrowLeft, Download, Eye, FileCheck, FileCog, ScrollText,
  Pencil, CheckCircle2, AlertTriangle, Play, Lock,
  User, Building2, Calendar, Shield,
} from 'lucide-react';
import {
  ComplianceDocument, DocumentSection, DocumentAttestation, DocumentRequirementSnapshot,
  DocumentSectionComplianceStatus, AttestationRole, SignatureMethod,
  TemplateSection,
  getComplianceDocument, updateComplianceDocument, generateDocument, finalizeDocument,
  getDocumentAttestations, upsertDocumentAttestation,
  getDocumentRequirements,
} from '@/lib/api';
import {
  DOC_TYPE_LABELS,
  DOC_STATUS_LABELS, DOC_STATUS_COLORS, DOC_TERMINAL_STATUSES,
  SECTION_STATUS_LABELS, SECTION_STATUS_COLORS,
  ATTESTATION_ROLE_LABELS,
} from '@/components/aoc/constants';

// Parse raw JSONB sections string safely
function parseSections(raw: string | undefined): TemplateSection[] {
  if (!raw) return [];
  try {
    return JSON.parse(raw) as TemplateSection[];
  } catch {
    return [];
  }
}

export default function DocumentBuilderPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { hasRole } = useAuth();
  const canEdit = hasRole('ciso', 'compliance_manager');
  const canFinalize = hasRole('ciso');

  const [doc, setDoc] = useState<ComplianceDocument | null>(null);
  const [attestations, setAttestations] = useState<DocumentAttestation[]>([]);
  const [requirements, setRequirements] = useState<DocumentRequirementSnapshot[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('sections');

  // Section state — keyed by section_key
  const [sections, setSections] = useState<Record<string, DocumentSection>>({});

  // Edit doc metadata
  const [showEditMeta, setShowEditMeta] = useState(false);
  const [savingMeta, setSavingMeta] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [editMerchantName, setEditMerchantName] = useState('');
  const [editMerchantDba, setEditMerchantDba] = useState('');
  const [editMerchantUrl, setEditMerchantUrl] = useState('');
  const [editQsaName, setEditQsaName] = useState('');
  const [editQsaCompany, setEditQsaCompany] = useState('');

  // Section edit
  const [editingSectionKey, setEditingSectionKey] = useState<string | null>(null);
  const [sectionContent, setSectionContent] = useState('');
  const [sectionStatus, setSectionStatus] = useState<DocumentSectionComplianceStatus | ''>('');
  const [savingSection, setSavingSection] = useState(false);

  // Attestation edit
  const [editingAttestRole, setEditingAttestRole] = useState<AttestationRole | null>(null);
  const [attestFullName, setAttestFullName] = useState('');
  const [attestTitle, setAttestTitle] = useState('');
  const [attestCompany, setAttestCompany] = useState('');
  const [attestEmail, setAttestEmail] = useState('');
  const [attestPhone, setAttestPhone] = useState('');
  const [attestQsaCompany, setAttestQsaCompany] = useState('');
  const [attestQsaNumber, setAttestQsaNumber] = useState('');
  const [attestSignedAt, setAttestSignedAt] = useState('');
  const [attestMethod, setAttestMethod] = useState<SignatureMethod>('manual');
  const [savingAttest, setSavingAttest] = useState(false);

  // Generate / finalize state
  const [generating, setGenerating] = useState(false);
  const [finalizing, setFinalizing] = useState(false);
  const [showFinalizeConfirm, setShowFinalizeConfirm] = useState(false);

  const fetchAll = useCallback(async () => {
    try {
      setLoading(true);
      const [docRes, attestRes, reqRes] = await Promise.all([
        getComplianceDocument(id),
        getDocumentAttestations(id),
        getDocumentRequirements(id),
      ]);
      setDoc(docRes.data);
      setAttestations(attestRes.data);
      setRequirements(reqRes.data);
    } catch (err) {
      console.error('Failed to load document:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  const isLocked = doc ? DOC_TERMINAL_STATUSES.includes(doc.doc_status) : false;

  const openEditMeta = () => {
    if (!doc) return;
    setEditTitle(doc.title);
    setEditMerchantName(doc.merchant_name ?? '');
    setEditMerchantDba(doc.merchant_dba ?? '');
    setEditMerchantUrl(doc.merchant_url ?? '');
    setEditQsaName(doc.qsa_name ?? '');
    setEditQsaCompany(doc.qsa_company ?? '');
    setShowEditMeta(true);
  };

  const handleSaveMeta = async () => {
    try {
      setSavingMeta(true);
      await updateComplianceDocument(id, {
        title: editTitle.trim() || undefined,
        merchant_name: editMerchantName.trim() || undefined,
        merchant_dba: editMerchantDba.trim() || undefined,
        merchant_url: editMerchantUrl.trim() || undefined,
        qsa_name: editQsaName.trim() || undefined,
        qsa_company: editQsaCompany.trim() || undefined,
      });
      setShowEditMeta(false);
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSavingMeta(false);
    }
  };

  const openSectionEdit = (section: TemplateSection, saved: DocumentSection | undefined) => {
    setEditingSectionKey(section.key);
    setSectionContent(saved?.content ?? '');
    setSectionStatus((saved?.compliance_status ?? '') as DocumentSectionComplianceStatus | '');
  };

  const handleSaveSection = async (sectionTitle: string) => {
    if (!editingSectionKey) return;
    try {
      setSavingSection(true);
      const { upsertDocumentSection } = await import('@/lib/api');
      const saved = await upsertDocumentSection(id, editingSectionKey, {
        title: sectionTitle,
        content: sectionContent || undefined,
        compliance_status: sectionStatus || undefined,
      });
      setSections(prev => ({ ...prev, [editingSectionKey]: saved.data }));
      setEditingSectionKey(null);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to save section');
    } finally {
      setSavingSection(false);
    }
  };

  const openAttestEdit = (role: AttestationRole) => {
    const existing = attestations.find(a => a.attestation_role === role);
    setAttestFullName(existing?.full_name ?? '');
    setAttestTitle(existing?.title ?? '');
    setAttestCompany(existing?.company_name ?? '');
    setAttestEmail(existing?.email ?? '');
    setAttestPhone(existing?.phone ?? '');
    setAttestQsaCompany(existing?.qsa_company ?? '');
    setAttestQsaNumber(existing?.qsa_number ?? '');
    setAttestSignedAt(existing?.signed_at ? existing.signed_at.slice(0, 10) : '');
    setAttestMethod((existing?.signature_method as SignatureMethod) ?? 'manual');
    setEditingAttestRole(role);
  };

  const handleSaveAttest = async () => {
    if (!editingAttestRole) return;
    try {
      setSavingAttest(true);
      const saved = await upsertDocumentAttestation(id, editingAttestRole, {
        full_name: attestFullName.trim(),
        title: attestTitle.trim(),
        company_name: attestCompany.trim() || undefined,
        email: attestEmail.trim() || undefined,
        phone: attestPhone.trim() || undefined,
        qsa_company: attestQsaCompany.trim() || undefined,
        qsa_number: attestQsaNumber.trim() || undefined,
        signed_at: attestSignedAt || undefined,
        signature_method: attestMethod,
      });
      setAttestations(prev => {
        const filtered = prev.filter(a => a.attestation_role !== editingAttestRole);
        return [...filtered, saved.data];
      });
      setEditingAttestRole(null);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to save attestation');
    } finally {
      setSavingAttest(false);
    }
  };

  const handleGenerate = async () => {
    if (!confirm('Generate PDF for this document? The document will move to review status.')) return;
    try {
      setGenerating(true);
      await generateDocument(id);
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to generate document');
    } finally {
      setGenerating(false);
    }
  };

  const handleFinalize = async () => {
    try {
      setFinalizing(true);
      await finalizeDocument(id);
      setShowFinalizeConfirm(false);
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to finalize document');
    } finally {
      setFinalizing(false);
    }
  };

  if (loading) {
    return <div className="text-muted-foreground py-12 text-center">Loading...</div>;
  }
  if (!doc) {
    return <div className="text-muted-foreground py-12 text-center">Document not found.</div>;
  }

  // We don't have template sections from this fetch — we infer from the document type.
  // The builder shows a generic section list from requirements snapshot when available,
  // otherwise shows free-form content entry.

  const requirementRows = requirements.slice().sort((a, b) =>
    a.requirement_code.localeCompare(b.requirement_code, undefined, { numeric: true })
  );

  // Attestation roles relevant to this document type
  const relevantRoles: AttestationRole[] = doc.document_type === 'roc'
    ? ['merchant_signatory', 'qsa_signatory', 'sp_signatory']
    : ['merchant_signatory', 'isac_signatory'];

  const compliantCount = requirements.filter(r => r.status === 'compliant').length;
  const nonCompliantCount = requirements.filter(r => r.status === 'non_compliant' || r.status === 'partially_compliant').length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push('/documents')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${DOC_STATUS_COLORS[doc.doc_status]}`}>
                {DOC_STATUS_LABELS[doc.doc_status]}
              </span>
              <Badge variant="outline" className="text-xs">
                {DOC_TYPE_LABELS[doc.document_type]}
              </Badge>
              <span className="text-xs text-muted-foreground">v{doc.version}</span>
            </div>
            <h1 className="text-2xl font-bold">{doc.title}</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              PCI DSS {doc.pci_dss_version} ·{' '}
              {new Date(doc.assessment_period_start).toLocaleDateString()} — {new Date(doc.assessment_period_end).toLocaleDateString()}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {canEdit && !isLocked && (
            <Button variant="outline" size="sm" onClick={openEditMeta}>
              <Pencil className="h-4 w-4 mr-1" /> Edit Details
            </Button>
          )}
          {canEdit && doc.doc_status === 'draft' && (
            <Button variant="outline" size="sm" onClick={handleGenerate} disabled={generating}>
              <Play className="h-4 w-4 mr-1" />
              {generating ? 'Generating...' : 'Generate PDF'}
            </Button>
          )}
          {doc.pdf_path && (
            <Button variant="outline" size="sm">
              <Download className="h-4 w-4 mr-1" /> Download PDF
            </Button>
          )}
          {canFinalize && doc.doc_status === 'approved' && (
            <Button size="sm" onClick={() => setShowFinalizeConfirm(true)}>
              <Lock className="h-4 w-4 mr-1" /> Finalize
            </Button>
          )}
        </div>
      </div>

      {/* Generation Error Banner */}
      {doc.generation_error && (
        <Card className="border-red-200 bg-red-50/50 dark:border-red-900/30">
          <CardContent className="p-4 flex gap-2">
            <AlertTriangle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
            <div>
              <p className="text-xs font-medium text-red-700 dark:text-red-400">Generation Error</p>
              <p className="text-sm text-red-600">{doc.generation_error}</p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Metadata Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <Building2 className="h-3 w-3" /> Merchant
            </div>
            <div className="font-medium text-sm">{doc.merchant_name || '—'}</div>
            {doc.merchant_dba && <div className="text-xs text-muted-foreground">DBA: {doc.merchant_dba}</div>}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <User className="h-3 w-3" /> QSA
            </div>
            <div className="font-medium text-sm">{doc.qsa_name || '—'}</div>
            {doc.qsa_company && <div className="text-xs text-muted-foreground">{doc.qsa_company}</div>}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <Calendar className="h-3 w-3" /> Generated
            </div>
            <div className="font-medium text-sm">
              {doc.generated_at ? new Date(doc.generated_at).toLocaleString() : '—'}
            </div>
            {doc.file_size_bytes && (
              <div className="text-xs text-muted-foreground">
                {(doc.file_size_bytes / 1024).toFixed(0)} KB
              </div>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <Shield className="h-3 w-3" /> Compliance
            </div>
            {requirements.length > 0 ? (
              <>
                <div className="text-2xl font-bold text-green-600">{compliantCount}</div>
                <div className="text-xs text-muted-foreground">
                  compliant · {nonCompliantCount} gap{nonCompliantCount !== 1 ? 's' : ''}
                </div>
              </>
            ) : (
              <div className="text-sm text-muted-foreground">No snapshot yet</div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="sections">
            <FileCog className="h-4 w-4 mr-1" /> Document Sections
          </TabsTrigger>
          <TabsTrigger value="attestations">
            <FileCheck className="h-4 w-4 mr-1" /> Attestations
            {attestations.length > 0 && (
              <span className="ml-1.5 rounded-full bg-muted px-1.5 py-0.5 text-xs">{attestations.length}</span>
            )}
          </TabsTrigger>
          <TabsTrigger value="requirements">
            <Eye className="h-4 w-4 mr-1" /> Requirements Snapshot
            {requirements.length > 0 && (
              <span className="ml-1.5 rounded-full bg-muted px-1.5 py-0.5 text-xs">{requirements.length}</span>
            )}
          </TabsTrigger>
          <TabsTrigger value="preview">
            <ScrollText className="h-4 w-4 mr-1" /> Preview
          </TabsTrigger>
        </TabsList>

        {/* Document Sections Tab */}
        <TabsContent value="sections">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Section Editor</CardTitle>
              <p className="text-xs text-muted-foreground">
                {isLocked
                  ? 'This document is finalized and locked.'
                  : 'Fill in each section to build the compliance document. Sections marked with compliance status will appear in the generated PDF.'}
              </p>
            </CardHeader>
            <CardContent className="space-y-3">
              {/* If we have section data from the snapshot, show those; otherwise show
                  a generic editor for custom content */}
              {Object.keys(sections).length === 0 && requirements.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground text-sm">
                  <FileCog className="h-8 w-8 mx-auto mb-2 opacity-40" />
                  <p>No sections yet. Sections will be created as you fill them in.</p>
                  {!isLocked && canEdit && (
                    <Button
                      variant="outline"
                      size="sm"
                      className="mt-3"
                      onClick={() => openSectionEdit(
                        { key: 'merchant_info', title: 'Merchant Information', required: true, order: 1, fields: [] },
                        undefined,
                      )}
                    >
                      Start filling sections
                    </Button>
                  )}
                </div>
              ) : (
                // Show editable sections from saved sections
                Object.entries(sections).map(([key, sec]) => (
                  <div key={key} className="border rounded-md p-4 space-y-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-sm">{sec.title}</span>
                        {sec.compliance_status && (
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SECTION_STATUS_COLORS[sec.compliance_status]}`}>
                            {SECTION_STATUS_LABELS[sec.compliance_status]}
                          </span>
                        )}
                      </div>
                      {canEdit && !isLocked && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            setSectionContent(sec.content ?? '');
                            setSectionStatus((sec.compliance_status ?? '') as DocumentSectionComplianceStatus | '');
                            setEditingSectionKey(key);
                          }}
                        >
                          <Pencil className="h-3.5 w-3.5 mr-1" /> Edit
                        </Button>
                      )}
                    </div>
                    {sec.content && (
                      <p className="text-sm text-muted-foreground line-clamp-2">{sec.content}</p>
                    )}
                  </div>
                ))
              )}

              {/* Quick-add common sections for AOC documents */}
              {canEdit && !isLocked && (
                <div className="pt-2 border-t">
                  <p className="text-xs text-muted-foreground mb-2">Quick-add section:</p>
                  <div className="flex flex-wrap gap-2">
                    {[
                      { key: 'merchant_info', title: 'Merchant Information' },
                      { key: 'assessment_period', title: 'Assessment Period' },
                      { key: 'scope_description', title: 'Scope Description' },
                      { key: 'merchant_attestation', title: 'Merchant Attestation' },
                      ...(doc.document_type === 'roc' ? [
                        { key: 'executive_summary', title: 'Executive Summary' },
                        { key: 'scope', title: 'Scope of Assessment' },
                      ] : []),
                    ].filter(s => !sections[s.key]).map(s => (
                      <Button
                        key={s.key}
                        variant="outline"
                        size="sm"
                        className="text-xs"
                        onClick={() => openSectionEdit(
                          { key: s.key, title: s.title, required: false, order: 0, fields: [] },
                          sections[s.key],
                        )}
                      >
                        + {s.title}
                      </Button>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Attestations Tab */}
        <TabsContent value="attestations">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Attestations &amp; Signatures</CardTitle>
              <p className="text-xs text-muted-foreground">
                QSA attestation fields — required before finalization
              </p>
            </CardHeader>
            <CardContent className="space-y-4">
              {relevantRoles.map((role) => {
                const attest = attestations.find(a => a.attestation_role === role);
                return (
                  <div key={role} className="border rounded-md p-4">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-sm">{ATTESTATION_ROLE_LABELS[role]}</span>
                        {attest ? (
                          <span className="inline-flex items-center gap-1 text-xs text-green-700">
                            <CheckCircle2 className="h-3 w-3" /> Filled
                          </span>
                        ) : (
                          <span className="text-xs text-muted-foreground">Not yet filled</span>
                        )}
                      </div>
                      {canEdit && !isLocked && (
                        <Button variant="outline" size="sm" onClick={() => openAttestEdit(role)}>
                          <Pencil className="h-3.5 w-3.5 mr-1" /> {attest ? 'Edit' : 'Fill In'}
                        </Button>
                      )}
                    </div>
                    {attest && (
                      <dl className="grid grid-cols-2 md:grid-cols-3 gap-x-4 gap-y-1 text-sm">
                        <div>
                          <dt className="text-xs text-muted-foreground">Name</dt>
                          <dd>{attest.full_name}</dd>
                        </div>
                        <div>
                          <dt className="text-xs text-muted-foreground">Title</dt>
                          <dd>{attest.title}</dd>
                        </div>
                        {attest.company_name && (
                          <div>
                            <dt className="text-xs text-muted-foreground">Company</dt>
                            <dd>{attest.company_name}</dd>
                          </div>
                        )}
                        {attest.email && (
                          <div>
                            <dt className="text-xs text-muted-foreground">Email</dt>
                            <dd>{attest.email}</dd>
                          </div>
                        )}
                        {attest.signed_at && (
                          <div>
                            <dt className="text-xs text-muted-foreground">Signed</dt>
                            <dd>{new Date(attest.signed_at).toLocaleDateString()}</dd>
                          </div>
                        )}
                        {attest.qsa_number && (
                          <div>
                            <dt className="text-xs text-muted-foreground">QSA #</dt>
                            <dd className="font-mono">{attest.qsa_number}</dd>
                          </div>
                        )}
                      </dl>
                    )}
                  </div>
                );
              })}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Requirements Snapshot Tab */}
        <TabsContent value="requirements">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Requirements Snapshot</CardTitle>
              <p className="text-xs text-muted-foreground">
                Point-in-time compliance posture captured at document generation
              </p>
            </CardHeader>
            <CardContent className="p-0">
              {requirements.length === 0 ? (
                <div className="text-center py-10 text-muted-foreground text-sm">
                  <Shield className="h-8 w-8 mx-auto mb-2 opacity-40" />
                  <p>No snapshot yet. Generate the document to capture compliance posture.</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Requirement</TableHead>
                      <TableHead>Title</TableHead>
                      <TableHead>In Scope</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Controls</TableHead>
                      <TableHead>Evidence</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {requirementRows.map((req) => (
                      <TableRow key={req.id}>
                        <TableCell className="font-mono text-sm font-medium">{req.requirement_code}</TableCell>
                        <TableCell className="text-sm max-w-xs truncate">{req.requirement_title}</TableCell>
                        <TableCell>
                          <span className={`text-xs font-medium ${req.in_scope ? 'text-green-700' : 'text-muted-foreground'}`}>
                            {req.in_scope ? 'Yes' : 'No'}
                          </span>
                        </TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SECTION_STATUS_COLORS[req.status]}`}>
                            {SECTION_STATUS_LABELS[req.status]}
                          </span>
                        </TableCell>
                        <TableCell className="text-sm">
                          {req.passing_controls}/{req.control_count}
                        </TableCell>
                        <TableCell className="text-sm">{req.evidence_count}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Preview Tab */}
        <TabsContent value="preview">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Document Preview</CardTitle>
            </CardHeader>
            <CardContent>
              {doc.pdf_path ? (
                <div className="text-center py-6">
                  <FileCheck className="h-12 w-12 mx-auto mb-3 text-green-500" />
                  <p className="text-sm font-medium">PDF generated successfully</p>
                  <p className="text-xs text-muted-foreground mb-4">
                    Generated {doc.generated_at ? new Date(doc.generated_at).toLocaleString() : '—'}
                    {doc.file_size_bytes ? ` · ${(doc.file_size_bytes / 1024).toFixed(0)} KB` : ''}
                  </p>
                  <Button>
                    <Download className="h-4 w-4 mr-2" /> Download PDF
                  </Button>
                </div>
              ) : (
                <div className="border rounded-md p-8 bg-muted/30 space-y-4">
                  <div className="text-center mb-6">
                    <ScrollText className="h-10 w-10 mx-auto mb-2 text-muted-foreground/50" />
                    <p className="font-semibold text-lg">{doc.title}</p>
                    <p className="text-sm text-muted-foreground">
                      {DOC_TYPE_LABELS[doc.document_type]} · PCI DSS {doc.pci_dss_version}
                    </p>
                  </div>
                  <div className="border-t pt-4 grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <p className="text-xs text-muted-foreground">Merchant Name</p>
                      <p className="font-medium">{doc.merchant_name || '—'}</p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Assessment Period</p>
                      <p className="font-medium">
                        {new Date(doc.assessment_period_start).toLocaleDateString()} — {new Date(doc.assessment_period_end).toLocaleDateString()}
                      </p>
                    </div>
                    {doc.qsa_name && (
                      <div>
                        <p className="text-xs text-muted-foreground">QSA Name</p>
                        <p className="font-medium">{doc.qsa_name}</p>
                      </div>
                    )}
                    {doc.qsa_company && (
                      <div>
                        <p className="text-xs text-muted-foreground">QSA Company</p>
                        <p className="font-medium">{doc.qsa_company}</p>
                      </div>
                    )}
                  </div>
                  {Object.values(sections).length > 0 && (
                    <div className="border-t pt-4 space-y-3">
                      {Object.values(sections).sort((a, b) => a.sort_order - b.sort_order).map(sec => (
                        <div key={sec.id}>
                          <div className="flex items-center gap-2 mb-1">
                            <p className="font-medium text-sm">{sec.title}</p>
                            {sec.compliance_status && (
                              <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SECTION_STATUS_COLORS[sec.compliance_status]}`}>
                                {SECTION_STATUS_LABELS[sec.compliance_status]}
                              </span>
                            )}
                          </div>
                          {sec.content && (
                            <p className="text-sm text-muted-foreground whitespace-pre-line">{sec.content}</p>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                  {doc.doc_status === 'draft' && canEdit && (
                    <div className="border-t pt-4 text-center">
                      <p className="text-sm text-muted-foreground mb-3">
                        PDF not yet generated. Fill in sections and attestations, then generate.
                      </p>
                      <Button onClick={handleGenerate} disabled={generating}>
                        <Play className="h-4 w-4 mr-2" />
                        {generating ? 'Generating...' : 'Generate PDF'}
                      </Button>
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Edit Metadata Dialog */}
      <Dialog open={showEditMeta} onOpenChange={setShowEditMeta}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Edit Document Details</DialogTitle>
            <DialogDescription>Update merchant and QSA information for this document</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Document Title</Label>
              <Input value={editTitle} onChange={(e) => setEditTitle(e.target.value)} />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Merchant Name</Label>
                <Input value={editMerchantName} onChange={(e) => setEditMerchantName(e.target.value)} />
              </div>
              <div>
                <Label>DBA / Brand Name</Label>
                <Input value={editMerchantDba} onChange={(e) => setEditMerchantDba(e.target.value)} />
              </div>
              <div>
                <Label>Merchant URL</Label>
                <Input value={editMerchantUrl} onChange={(e) => setEditMerchantUrl(e.target.value)} placeholder="https://..." />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>QSA Name</Label>
                <Input value={editQsaName} onChange={(e) => setEditQsaName(e.target.value)} placeholder="QSA individual name" />
              </div>
              <div>
                <Label>QSA Company</Label>
                <Input value={editQsaCompany} onChange={(e) => setEditQsaCompany(e.target.value)} placeholder="QSA firm name" />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEditMeta(false)}>Cancel</Button>
            <Button onClick={handleSaveMeta} disabled={savingMeta}>
              {savingMeta ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Section Edit Dialog */}
      {editingSectionKey && (
        <Dialog open={!!editingSectionKey} onOpenChange={() => setEditingSectionKey(null)}>
          <DialogContent className="max-w-xl">
            <DialogHeader>
              <DialogTitle>Edit Section</DialogTitle>
              <DialogDescription>
                Section key: <code className="font-mono text-xs">{editingSectionKey}</code>
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div>
                <Label>Compliance Status</Label>
                <Select
                  value={sectionStatus}
                  onValueChange={(v) => setSectionStatus(v as DocumentSectionComplianceStatus | '')}
                >
                  <SelectTrigger><SelectValue placeholder="Select status..." /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">Not set</SelectItem>
                    {(Object.entries(SECTION_STATUS_LABELS) as [DocumentSectionComplianceStatus, string][]).map(([k, v]) => (
                      <SelectItem key={k} value={k}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Content</Label>
                <Textarea
                  value={sectionContent}
                  onChange={(e) => setSectionContent(e.target.value)}
                  placeholder="Section content, findings, or notes..."
                  className="min-h-[120px]"
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setEditingSectionKey(null)}>Cancel</Button>
              <Button
                onClick={() => handleSaveSection(
                  sections[editingSectionKey]?.title ?? editingSectionKey
                )}
                disabled={savingSection}
              >
                {savingSection ? 'Saving...' : 'Save Section'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {/* Attestation Edit Dialog */}
      {editingAttestRole && (
        <Dialog open={!!editingAttestRole} onOpenChange={() => setEditingAttestRole(null)}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>{ATTESTATION_ROLE_LABELS[editingAttestRole]}</DialogTitle>
              <DialogDescription>Fill in signatory details for this attestation</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label>Full Name *</Label>
                  <Input value={attestFullName} onChange={(e) => setAttestFullName(e.target.value)} />
                </div>
                <div>
                  <Label>Title *</Label>
                  <Input value={attestTitle} onChange={(e) => setAttestTitle(e.target.value)} placeholder="e.g. CISO, QSA" />
                </div>
                <div>
                  <Label>Company</Label>
                  <Input value={attestCompany} onChange={(e) => setAttestCompany(e.target.value)} />
                </div>
                <div>
                  <Label>Email</Label>
                  <Input type="email" value={attestEmail} onChange={(e) => setAttestEmail(e.target.value)} />
                </div>
                <div>
                  <Label>Phone</Label>
                  <Input value={attestPhone} onChange={(e) => setAttestPhone(e.target.value)} />
                </div>
                <div>
                  <Label>Signature Date</Label>
                  <Input type="date" value={attestSignedAt} onChange={(e) => setAttestSignedAt(e.target.value)} />
                </div>
                <div>
                  <Label>Signature Method</Label>
                  <Select value={attestMethod} onValueChange={(v) => setAttestMethod(v as SignatureMethod)}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="manual">Manual / Wet Signature</SelectItem>
                      <SelectItem value="digital_signature">Digital Signature</SelectItem>
                      <SelectItem value="docusign_ref">DocuSign Reference</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                {(editingAttestRole === 'qsa_signatory') && (
                  <>
                    <div>
                      <Label>QSA Company</Label>
                      <Input value={attestQsaCompany} onChange={(e) => setAttestQsaCompany(e.target.value)} />
                    </div>
                    <div>
                      <Label>QSA Certificate #</Label>
                      <Input value={attestQsaNumber} onChange={(e) => setAttestQsaNumber(e.target.value)} className="font-mono" />
                    </div>
                  </>
                )}
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setEditingAttestRole(null)}>Cancel</Button>
              <Button
                onClick={handleSaveAttest}
                disabled={savingAttest || !attestFullName.trim() || !attestTitle.trim()}
              >
                {savingAttest ? 'Saving...' : 'Save Attestation'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {/* Finalize Confirmation Dialog */}
      <Dialog open={showFinalizeConfirm} onOpenChange={setShowFinalizeConfirm}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Finalize Document</DialogTitle>
            <DialogDescription>
              Finalizing locks this document permanently. It cannot be edited after finalization.
              Make sure all sections, attestations, and signatures are complete.
            </DialogDescription>
          </DialogHeader>
          <div className="py-2">
            <p className="text-sm font-medium">{doc.title}</p>
            <p className="text-xs text-muted-foreground">{DOC_TYPE_LABELS[doc.document_type]}</p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowFinalizeConfirm(false)}>Cancel</Button>
            <Button onClick={handleFinalize} disabled={finalizing}>
              <Lock className="h-4 w-4 mr-1" />
              {finalizing ? 'Finalizing...' : 'Finalize Document'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
