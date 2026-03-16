'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
  ArrowLeft, Edit2, Trash2, Plus, FileText, Shield, Grid3x3,
  Calendar, Mail, Phone, User, Building2, AlertTriangle,
} from 'lucide-react';
import {
  ServiceProvider, SPComplianceDocument, SPResponsibilityMatrix,
  SPType, SPComplianceStatus, SPRiskLevel, SPDocType, SPResponsibleParty,
  getServiceProvider, updateServiceProvider, deleteServiceProvider,
  listSPComplianceDocs, createSPComplianceDoc, deleteSPComplianceDoc,
  getSPResponsibilityMatrix, upsertSPResponsibility, deleteSPResponsibility,
} from '@/lib/api';
import {
  SP_TYPE_LABELS,
  SP_COMPLIANCE_LABELS, SP_COMPLIANCE_COLORS,
  SP_RISK_LABELS, SP_RISK_COLORS,
  SP_DOC_TYPE_LABELS,
  SP_PARTY_LABELS, SP_PARTY_COLORS,
} from '@/components/vendor/constants';

export default function VendorDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { hasRole } = useAuth();
  const canManage = hasRole('ciso', 'compliance_manager', 'vendor_manager');

  const [provider, setProvider] = useState<ServiceProvider | null>(null);
  const [docs, setDocs] = useState<SPComplianceDocument[]>([]);
  const [matrix, setMatrix] = useState<SPResponsibilityMatrix[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');

  // Edit state
  const [showEdit, setShowEdit] = useState(false);
  const [saving, setSaving] = useState(false);
  const [editName, setEditName] = useState('');
  const [editType, setEditType] = useState<SPType>('other');
  const [editContactName, setEditContactName] = useState('');
  const [editContactEmail, setEditContactEmail] = useState('');
  const [editContactPhone, setEditContactPhone] = useState('');
  const [editServices, setEditServices] = useState('');
  const [editComplianceStatus, setEditComplianceStatus] = useState<SPComplianceStatus>('unknown');
  const [editRiskLevel, setEditRiskLevel] = useState<SPRiskLevel>('medium');
  const [editRiskNotes, setEditRiskNotes] = useState('');
  const [editLastAOC, setEditLastAOC] = useState('');
  const [editNextReview, setEditNextReview] = useState('');
  const [editContractStart, setEditContractStart] = useState('');
  const [editContractEnd, setEditContractEnd] = useState('');
  const [editIsActive, setEditIsActive] = useState(true);

  // Add document state
  const [showAddDoc, setShowAddDoc] = useState(false);
  const [addingDoc, setAddingDoc] = useState(false);
  const [docType, setDocType] = useState<SPDocType>('aoc');
  const [docTitle, setDocTitle] = useState('');
  const [docVersion, setDocVersion] = useState('');
  const [docValidFrom, setDocValidFrom] = useState('');
  const [docValidUntil, setDocValidUntil] = useState('');
  const [docReviewNotes, setDocReviewNotes] = useState('');

  // Add responsibility state
  const [showAddResp, setShowAddResp] = useState(false);
  const [addingResp, setAddingResp] = useState(false);
  const [respCode, setRespCode] = useState('');
  const [respParty, setRespParty] = useState<SPResponsibleParty>('merchant');
  const [respNotes, setRespNotes] = useState('');

  const fetchAll = useCallback(async () => {
    try {
      setLoading(true);
      const [spRes, docsRes, matrixRes] = await Promise.all([
        getServiceProvider(id),
        listSPComplianceDocs(id),
        getSPResponsibilityMatrix(id),
      ]);
      setProvider(spRes.data);
      setDocs(docsRes.data);
      setMatrix(matrixRes.data);
    } catch (err) {
      console.error('Failed to load provider:', err);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  const openEdit = () => {
    if (!provider) return;
    setEditName(provider.name);
    setEditType(provider.type);
    setEditContactName(provider.contact_name ?? '');
    setEditContactEmail(provider.contact_email ?? '');
    setEditContactPhone(provider.contact_phone ?? '');
    setEditServices(provider.services_provided ?? '');
    setEditComplianceStatus(provider.pci_compliance_status);
    setEditRiskLevel(provider.risk_level);
    setEditRiskNotes(provider.risk_notes ?? '');
    setEditLastAOC(provider.last_aoc_date ? provider.last_aoc_date.slice(0, 10) : '');
    setEditNextReview(provider.next_review_date ? provider.next_review_date.slice(0, 10) : '');
    setEditContractStart(provider.contract_start_date ? provider.contract_start_date.slice(0, 10) : '');
    setEditContractEnd(provider.contract_end_date ? provider.contract_end_date.slice(0, 10) : '');
    setEditIsActive(provider.is_active);
    setShowEdit(true);
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      await updateServiceProvider(id, {
        name: editName.trim(),
        type: editType,
        contact_name: editContactName.trim() || undefined,
        contact_email: editContactEmail.trim() || undefined,
        contact_phone: editContactPhone.trim() || undefined,
        services_provided: editServices.trim() || undefined,
        pci_compliance_status: editComplianceStatus,
        risk_level: editRiskLevel,
        risk_notes: editRiskNotes.trim() || undefined,
        last_aoc_date: editLastAOC || undefined,
        next_review_date: editNextReview || undefined,
        contract_start_date: editContractStart || undefined,
        contract_end_date: editContractEnd || undefined,
        is_active: editIsActive,
      });
      setShowEdit(false);
      fetchAll();
    } catch (err) {
      console.error('Save failed:', err);
      alert(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm(`Delete "${provider?.name}"? This action cannot be undone.`)) return;
    try {
      await deleteServiceProvider(id);
      router.push('/vendors');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete');
    }
  };

  const handleAddDoc = async () => {
    try {
      setAddingDoc(true);
      await createSPComplianceDoc(id, {
        document_type: docType,
        title: docTitle.trim() || undefined,
        document_version: docVersion.trim() || undefined,
        valid_from: docValidFrom || undefined,
        valid_until: docValidUntil || undefined,
        review_notes: docReviewNotes.trim() || undefined,
      });
      setShowAddDoc(false);
      setDocTitle('');
      setDocVersion('');
      setDocValidFrom('');
      setDocValidUntil('');
      setDocReviewNotes('');
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to add document');
    } finally {
      setAddingDoc(false);
    }
  };

  const handleDeleteDoc = async (docId: string) => {
    if (!confirm('Remove this compliance document?')) return;
    try {
      await deleteSPComplianceDoc(id, docId);
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete document');
    }
  };

  const handleAddResp = async () => {
    if (!respCode.trim()) return;
    try {
      setAddingResp(true);
      await upsertSPResponsibility(id, respCode.trim().toUpperCase(), {
        responsible_party: respParty,
        notes: respNotes.trim() || undefined,
      });
      setShowAddResp(false);
      setRespCode('');
      setRespParty('merchant');
      setRespNotes('');
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to save responsibility');
    } finally {
      setAddingResp(false);
    }
  };

  const handleDeleteResp = async (reqCode: string) => {
    if (!confirm(`Remove responsibility assignment for ${reqCode}?`)) return;
    try {
      await deleteSPResponsibility(id, reqCode);
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete');
    }
  };

  if (loading) {
    return <div className="text-muted-foreground py-12 text-center">Loading...</div>;
  }
  if (!provider) {
    return <div className="text-muted-foreground py-12 text-center">Provider not found.</div>;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push('/vendors')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SP_COMPLIANCE_COLORS[provider.pci_compliance_status]}`}>
                {SP_COMPLIANCE_LABELS[provider.pci_compliance_status]}
              </span>
              <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SP_RISK_COLORS[provider.risk_level]}`}>
                {SP_RISK_LABELS[provider.risk_level]} Risk
              </span>
              {!provider.is_active && (
                <Badge variant="outline" className="text-xs text-muted-foreground">Inactive</Badge>
              )}
            </div>
            <h1 className="text-2xl font-bold">{provider.name}</h1>
            <p className="text-sm text-muted-foreground mt-0.5">
              {SP_TYPE_LABELS[provider.type]}
            </p>
          </div>
        </div>
        {canManage && (
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={openEdit}>
              <Edit2 className="h-4 w-4 mr-1" /> Edit
            </Button>
            <Button variant="outline" size="sm" className="text-red-500 hover:text-red-600" onClick={handleDelete}>
              <Trash2 className="h-4 w-4 mr-1" /> Delete
            </Button>
          </div>
        )}
      </div>

      {/* Metadata Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <User className="h-3 w-3" /> Contact
            </div>
            <div className="font-medium text-sm">{provider.contact_name || '—'}</div>
            {provider.contact_email && (
              <div className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
                <Mail className="h-3 w-3" /> {provider.contact_email}
              </div>
            )}
            {provider.contact_phone && (
              <div className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
                <Phone className="h-3 w-3" /> {provider.contact_phone}
              </div>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <Calendar className="h-3 w-3" /> Last AOC
            </div>
            <div className="font-medium text-sm">
              {provider.last_aoc_date
                ? new Date(provider.last_aoc_date).toLocaleDateString()
                : '—'}
            </div>
            <div className="text-xs text-muted-foreground mt-0.5">
              Next review:{' '}
              {provider.next_review_date
                ? new Date(provider.next_review_date).toLocaleDateString()
                : '—'}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <Building2 className="h-3 w-3" /> Contract
            </div>
            <div className="font-medium text-sm">
              {provider.contract_start_date
                ? new Date(provider.contract_start_date).toLocaleDateString()
                : '—'}
              {' → '}
              {provider.contract_end_date
                ? new Date(provider.contract_end_date).toLocaleDateString()
                : 'Ongoing'}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2 text-muted-foreground text-xs mb-1">
              <Shield className="h-3 w-3" /> Compliance Docs
            </div>
            <div className="text-2xl font-bold">{docs.filter(d => d.is_current).length}</div>
            <div className="text-xs text-muted-foreground">current document{docs.filter(d => d.is_current).length !== 1 ? 's' : ''}</div>
          </CardContent>
        </Card>
      </div>

      {/* Services Provided */}
      {provider.services_provided && (
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground mb-1">Services Provided</p>
            <p className="text-sm">{provider.services_provided}</p>
          </CardContent>
        </Card>
      )}

      {/* Risk Notes */}
      {provider.risk_notes && (
        <Card className="border-orange-200 bg-orange-50/50 dark:border-orange-900/30 dark:bg-orange-900/5">
          <CardContent className="p-4 flex gap-2">
            <AlertTriangle className="h-4 w-4 text-orange-500 mt-0.5 shrink-0" />
            <div>
              <p className="text-xs font-medium text-orange-700 dark:text-orange-400 mb-0.5">Risk Notes</p>
              <p className="text-sm">{provider.risk_notes}</p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="overview">
            <Building2 className="h-4 w-4 mr-1" /> Overview
          </TabsTrigger>
          <TabsTrigger value="documents">
            <FileText className="h-4 w-4 mr-1" /> Compliance Docs
            {docs.length > 0 && (
              <span className="ml-1.5 rounded-full bg-muted px-1.5 py-0.5 text-xs">{docs.length}</span>
            )}
          </TabsTrigger>
          <TabsTrigger value="matrix">
            <Grid3x3 className="h-4 w-4 mr-1" /> Responsibility Matrix
            {matrix.length > 0 && (
              <span className="ml-1.5 rounded-full bg-muted px-1.5 py-0.5 text-xs">{matrix.length}</span>
            )}
          </TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Provider Details</CardTitle>
            </CardHeader>
            <CardContent>
              <dl className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
                <div>
                  <dt className="text-muted-foreground text-xs">Provider Type</dt>
                  <dd className="font-medium">{SP_TYPE_LABELS[provider.type]}</dd>
                </div>
                <div>
                  <dt className="text-muted-foreground text-xs">PCI Compliance Status</dt>
                  <dd>
                    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SP_COMPLIANCE_COLORS[provider.pci_compliance_status]}`}>
                      {SP_COMPLIANCE_LABELS[provider.pci_compliance_status]}
                    </span>
                  </dd>
                </div>
                <div>
                  <dt className="text-muted-foreground text-xs">Risk Level</dt>
                  <dd>
                    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SP_RISK_COLORS[provider.risk_level]}`}>
                      {SP_RISK_LABELS[provider.risk_level]}
                    </span>
                  </dd>
                </div>
                <div>
                  <dt className="text-muted-foreground text-xs">Active</dt>
                  <dd className="font-medium">{provider.is_active ? 'Yes' : 'No'}</dd>
                </div>
                <div>
                  <dt className="text-muted-foreground text-xs">Last AOC Date</dt>
                  <dd className="font-medium">
                    {provider.last_aoc_date ? new Date(provider.last_aoc_date).toLocaleDateString() : '—'}
                  </dd>
                </div>
                <div>
                  <dt className="text-muted-foreground text-xs">Next Review Date</dt>
                  <dd className="font-medium">
                    {provider.next_review_date ? new Date(provider.next_review_date).toLocaleDateString() : '—'}
                  </dd>
                </div>
                <div>
                  <dt className="text-muted-foreground text-xs">Contract Start</dt>
                  <dd className="font-medium">
                    {provider.contract_start_date ? new Date(provider.contract_start_date).toLocaleDateString() : '—'}
                  </dd>
                </div>
                <div>
                  <dt className="text-muted-foreground text-xs">Contract End</dt>
                  <dd className="font-medium">
                    {provider.contract_end_date ? new Date(provider.contract_end_date).toLocaleDateString() : 'Ongoing'}
                  </dd>
                </div>
                <div>
                  <dt className="text-muted-foreground text-xs">Added</dt>
                  <dd className="font-medium">{new Date(provider.created_at).toLocaleDateString()}</dd>
                </div>
                <div>
                  <dt className="text-muted-foreground text-xs">Last Updated</dt>
                  <dd className="font-medium">{new Date(provider.updated_at).toLocaleDateString()}</dd>
                </div>
              </dl>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Compliance Documents Tab */}
        <TabsContent value="documents">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between py-4">
              <CardTitle className="text-base">Compliance Documents</CardTitle>
              {canManage && (
                <Button size="sm" onClick={() => setShowAddDoc(true)}>
                  <Plus className="h-4 w-4 mr-1" /> Add Document
                </Button>
              )}
            </CardHeader>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Type</TableHead>
                    <TableHead>Title</TableHead>
                    <TableHead>Version</TableHead>
                    <TableHead>Valid From</TableHead>
                    <TableHead>Valid Until</TableHead>
                    <TableHead>Current</TableHead>
                    {canManage && <TableHead className="text-right">Actions</TableHead>}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {docs.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={canManage ? 7 : 6} className="text-center py-8 text-muted-foreground">
                        No compliance documents on file
                      </TableCell>
                    </TableRow>
                  ) : (
                    docs.map((doc) => (
                      <TableRow key={doc.id}>
                        <TableCell>
                          <Badge variant="outline" className="text-xs whitespace-nowrap">
                            {SP_DOC_TYPE_LABELS[doc.document_type] ?? doc.document_type}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm">{doc.title || '—'}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{doc.document_version || '—'}</TableCell>
                        <TableCell className="text-sm">
                          {doc.valid_from ? new Date(doc.valid_from).toLocaleDateString() : '—'}
                        </TableCell>
                        <TableCell className="text-sm">
                          {doc.valid_until
                            ? (
                              <span className={new Date(doc.valid_until) < new Date() ? 'text-red-600 font-medium' : ''}>
                                {new Date(doc.valid_until).toLocaleDateString()}
                              </span>
                            )
                            : '—'}
                        </TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${doc.is_current ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                            {doc.is_current ? 'Yes' : 'No'}
                          </span>
                        </TableCell>
                        {canManage && (
                          <TableCell className="text-right">
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-red-500"
                              onClick={() => handleDeleteDoc(doc.id)}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        )}
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Responsibility Matrix Tab */}
        <TabsContent value="matrix">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between py-4">
              <div>
                <CardTitle className="text-base">Responsibility Matrix</CardTitle>
                <p className="text-xs text-muted-foreground mt-0.5">
                  PCI DSS Req 12.9.2 — documents which party handles each requirement
                </p>
              </div>
              {canManage && (
                <Button size="sm" onClick={() => setShowAddResp(true)}>
                  <Plus className="h-4 w-4 mr-1" /> Assign Requirement
                </Button>
              )}
            </CardHeader>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Requirement</TableHead>
                    <TableHead>Responsible Party</TableHead>
                    <TableHead>Notes</TableHead>
                    {canManage && <TableHead className="text-right">Actions</TableHead>}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {matrix.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={canManage ? 4 : 3} className="text-center py-8 text-muted-foreground">
                        No responsibility assignments yet
                      </TableCell>
                    </TableRow>
                  ) : (
                    matrix.map((row) => (
                      <TableRow key={row.id}>
                        <TableCell className="font-mono text-sm font-medium">
                          {row.requirement_code}
                        </TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SP_PARTY_COLORS[row.responsible_party]}`}>
                            {SP_PARTY_LABELS[row.responsible_party]}
                          </span>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground max-w-xs truncate">
                          {row.notes || '—'}
                        </TableCell>
                        {canManage && (
                          <TableCell className="text-right">
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-red-500"
                              onClick={() => handleDeleteResp(row.requirement_code)}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        )}
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Edit Dialog */}
      <Dialog open={showEdit} onOpenChange={setShowEdit}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Edit Provider</DialogTitle>
            <DialogDescription>Update service provider details and compliance information</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-2">
                <Label>Provider Name *</Label>
                <Input value={editName} onChange={(e) => setEditName(e.target.value)} />
              </div>
              <div>
                <Label>Type</Label>
                <Select value={editType} onValueChange={(v) => setEditType(v as SPType)}>
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
                <Select value={editRiskLevel} onValueChange={(v) => setEditRiskLevel(v as SPRiskLevel)}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {(Object.entries(SP_RISK_LABELS) as [SPRiskLevel, string][]).map(([k, v]) => (
                      <SelectItem key={k} value={k}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="col-span-2">
                <Label>PCI Compliance Status</Label>
                <Select value={editComplianceStatus} onValueChange={(v) => setEditComplianceStatus(v as SPComplianceStatus)}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {(Object.entries(SP_COMPLIANCE_LABELS) as [SPComplianceStatus, string][]).map(([k, v]) => (
                      <SelectItem key={k} value={k}>{v}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Contact Name</Label>
                <Input value={editContactName} onChange={(e) => setEditContactName(e.target.value)} />
              </div>
              <div>
                <Label>Contact Email</Label>
                <Input type="email" value={editContactEmail} onChange={(e) => setEditContactEmail(e.target.value)} />
              </div>
              <div>
                <Label>Contact Phone</Label>
                <Input value={editContactPhone} onChange={(e) => setEditContactPhone(e.target.value)} />
              </div>
              <div>
                <Label>Active</Label>
                <Select value={editIsActive ? 'true' : 'false'} onValueChange={(v) => setEditIsActive(v === 'true')}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="true">Active</SelectItem>
                    <SelectItem value="false">Inactive</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="col-span-2">
                <Label>Services Provided</Label>
                <Input value={editServices} onChange={(e) => setEditServices(e.target.value)} placeholder="Brief description..." />
              </div>
              <div className="col-span-2">
                <Label>Risk Notes</Label>
                <Input value={editRiskNotes} onChange={(e) => setEditRiskNotes(e.target.value)} placeholder="Risk context notes..." />
              </div>
              <div>
                <Label>Last AOC Date</Label>
                <Input type="date" value={editLastAOC} onChange={(e) => setEditLastAOC(e.target.value)} />
              </div>
              <div>
                <Label>Next Review Date</Label>
                <Input type="date" value={editNextReview} onChange={(e) => setEditNextReview(e.target.value)} />
              </div>
              <div>
                <Label>Contract Start</Label>
                <Input type="date" value={editContractStart} onChange={(e) => setEditContractStart(e.target.value)} />
              </div>
              <div>
                <Label>Contract End</Label>
                <Input type="date" value={editContractEnd} onChange={(e) => setEditContractEnd(e.target.value)} />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEdit(false)}>Cancel</Button>
            <Button onClick={handleSave} disabled={saving || !editName.trim()}>
              {saving ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Document Dialog */}
      <Dialog open={showAddDoc} onOpenChange={setShowAddDoc}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Add Compliance Document</DialogTitle>
            <DialogDescription>Record a compliance document received from this provider</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Document Type *</Label>
              <Select value={docType} onValueChange={(v) => setDocType(v as SPDocType)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {(Object.entries(SP_DOC_TYPE_LABELS) as [SPDocType, string][]).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Title</Label>
                <Input value={docTitle} onChange={(e) => setDocTitle(e.target.value)} placeholder="e.g. Stripe AOC 2025" />
              </div>
              <div>
                <Label>Version / Standard</Label>
                <Input value={docVersion} onChange={(e) => setDocVersion(e.target.value)} placeholder="e.g. PCI DSS 4.0.1" />
              </div>
              <div>
                <Label>Valid From</Label>
                <Input type="date" value={docValidFrom} onChange={(e) => setDocValidFrom(e.target.value)} />
              </div>
              <div>
                <Label>Valid Until</Label>
                <Input type="date" value={docValidUntil} onChange={(e) => setDocValidUntil(e.target.value)} />
              </div>
            </div>
            <div>
              <Label>Review Notes</Label>
              <Input value={docReviewNotes} onChange={(e) => setDocReviewNotes(e.target.value)} placeholder="Notes from reviewer..." />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAddDoc(false)}>Cancel</Button>
            <Button onClick={handleAddDoc} disabled={addingDoc}>
              {addingDoc ? 'Adding...' : 'Add Document'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Responsibility Dialog */}
      <Dialog open={showAddResp} onOpenChange={setShowAddResp}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Assign PCI Requirement</DialogTitle>
            <DialogDescription>
              Map a PCI DSS requirement to a responsible party (Req 12.9.2)
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Requirement Code *</Label>
              <Input
                value={respCode}
                onChange={(e) => setRespCode(e.target.value.toUpperCase())}
                placeholder="e.g. 1.3.2, 8.2.1"
                className="font-mono"
              />
            </div>
            <div>
              <Label>Responsible Party *</Label>
              <Select value={respParty} onValueChange={(v) => setRespParty(v as SPResponsibleParty)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="merchant">Merchant — our org is fully responsible</SelectItem>
                  <SelectItem value="provider">Provider — SP is fully responsible</SelectItem>
                  <SelectItem value="shared">Shared — both parties share responsibility</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Notes</Label>
              <Input value={respNotes} onChange={(e) => setRespNotes(e.target.value)} placeholder="How responsibility is divided..." />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAddResp(false)}>Cancel</Button>
            <Button onClick={handleAddResp} disabled={addingResp || !respCode.trim()}>
              {addingResp ? 'Saving...' : 'Save Assignment'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
