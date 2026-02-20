'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Separator } from '@/components/ui/separator';
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
  ArrowLeft, FileText, Link2, Clock, CheckCircle2, AlertTriangle, Download,
  Upload, Shield, User, XCircle, Eye, History, Plus, Trash2, ClipboardCheck,
  Star, ChevronRight,
} from 'lucide-react';
import {
  EvidenceArtifact, EvidenceLink, EvidenceEvaluation, EvidenceVersion, Control,
  getEvidence, listEvidenceLinks, listEvidenceEvaluations, listEvidenceVersions,
  getDownloadURL, createEvidenceLinks, deleteEvidenceLink, createEvidenceEvaluation,
  changeEvidenceStatus, createEvidenceVersion, confirmEvidenceUpload,
  listControls,
} from '@/lib/api';
import {
  FreshnessBadge,
  EVIDENCE_TYPE_LABELS, EVIDENCE_STATUS_LABELS, EVIDENCE_STATUS_COLORS,
  COLLECTION_METHOD_LABELS, VERDICT_CONFIG, formatFileSize,
} from '@/components/evidence/freshness-badge';

const STRENGTH_LABELS: Record<string, string> = {
  primary: 'Primary',
  supporting: 'Supporting',
  supplementary: 'Supplementary',
};

const STRENGTH_COLORS: Record<string, string> = {
  primary: 'bg-green-500/10 text-green-700 dark:text-green-400',
  supporting: 'bg-blue-500/10 text-blue-700 dark:text-blue-400',
  supplementary: 'bg-amber-500/10 text-amber-700 dark:text-amber-400',
};

export default function EvidenceDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { hasRole } = useAuth();
  const evidenceId = params.id as string;

  const canLink = hasRole('ciso', 'compliance_manager', 'security_engineer');
  const canManage = hasRole('ciso', 'compliance_manager');
  const canEvaluate = hasRole('ciso', 'compliance_manager', 'auditor');
  const canVersion = hasRole('ciso', 'compliance_manager', 'security_engineer', 'it_admin', 'devops_engineer');

  const [evidence, setEvidence] = useState<EvidenceArtifact | null>(null);
  const [links, setLinks] = useState<EvidenceLink[]>([]);
  const [evaluations, setEvaluations] = useState<EvidenceEvaluation[]>([]);
  const [versions, setVersions] = useState<EvidenceVersion[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');

  // Link dialog
  const [showLinkDialog, setShowLinkDialog] = useState(false);
  const [linkControls, setLinkControls] = useState<Control[]>([]);
  const [linkControlSearch, setLinkControlSearch] = useState('');
  const [linkSelectedControl, setLinkSelectedControl] = useState('');
  const [linkStrength, setLinkStrength] = useState('primary');
  const [linkNotes, setLinkNotes] = useState('');
  const [linkLoading, setLinkLoading] = useState(false);

  // Evaluation dialog
  const [showEvalDialog, setShowEvalDialog] = useState(false);
  const [evalForm, setEvalForm] = useState({
    verdict: 'sufficient' as string,
    confidence: 'medium' as string,
    comments: '',
    missing_elements: '',
    remediation_notes: '',
  });
  const [evalLoading, setEvalLoading] = useState(false);

  // Version upload dialog
  const [showVersionDialog, setShowVersionDialog] = useState(false);
  const [versionFile, setVersionFile] = useState<File | null>(null);
  const [versionForm, setVersionForm] = useState({
    title: '',
    description: '',
    collection_date: new Date().toISOString().split('T')[0],
    freshness_period_days: '',
  });
  const [versionLoading, setVersionLoading] = useState(false);
  const [versionError, setVersionError] = useState('');

  // Version comparison
  const [compareVersions, setCompareVersions] = useState<[string | null, string | null]>([null, null]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [evRes, linksRes, evalRes, versRes] = await Promise.all([
        getEvidence(evidenceId),
        listEvidenceLinks(evidenceId),
        listEvidenceEvaluations(evidenceId),
        listEvidenceVersions(evidenceId),
      ]);
      setEvidence(evRes.data);
      setLinks(linksRes.data || []);
      setEvaluations(evalRes.data || []);
      setVersions(versRes.data || []);
    } catch {
      router.push('/evidence');
    } finally {
      setLoading(false);
    }
  }, [evidenceId, router]);

  useEffect(() => { fetchData(); }, [fetchData]);

  // Fetch controls for linking
  useEffect(() => {
    if (showLinkDialog) {
      const params: Record<string, string> = { per_page: '50', status: 'active' };
      if (linkControlSearch) params.search = linkControlSearch;
      listControls(params).then(res => setLinkControls(res.data || [])).catch(() => {});
    }
  }, [showLinkDialog, linkControlSearch]);

  async function handleDownload() {
    if (!evidence) return;
    try {
      const res = await getDownloadURL(evidence.id);
      const url = res.data.download?.presigned_url;
      if (url) window.open(url, '_blank');
    } catch { /* handle */ }
  }

  async function handleCreateLink() {
    if (!linkSelectedControl) return;
    setLinkLoading(true);
    try {
      await createEvidenceLinks(evidenceId, [{
        target_type: 'control',
        control_id: linkSelectedControl,
        strength: linkStrength,
        notes: linkNotes || undefined,
      }]);
      setShowLinkDialog(false);
      setLinkSelectedControl('');
      setLinkNotes('');
      fetchData();
    } catch { /* handle */ }
    finally { setLinkLoading(false); }
  }

  async function handleDeleteLink(linkId: string) {
    if (!confirm('Remove this evidence link?')) return;
    try {
      await deleteEvidenceLink(evidenceId, linkId);
      fetchData();
    } catch { /* handle */ }
  }

  async function handleEvaluation() {
    setEvalLoading(true);
    try {
      const missing = evalForm.missing_elements
        ? evalForm.missing_elements.split(',').map(s => s.trim()).filter(Boolean)
        : [];
      await createEvidenceEvaluation(evidenceId, {
        verdict: evalForm.verdict,
        confidence: evalForm.confidence,
        comments: evalForm.comments,
        missing_elements: missing.length > 0 ? missing : undefined,
        remediation_notes: evalForm.remediation_notes || undefined,
      });
      setShowEvalDialog(false);
      setEvalForm({ verdict: 'sufficient', confidence: 'medium', comments: '', missing_elements: '', remediation_notes: '' });
      fetchData();
    } catch { /* handle */ }
    finally { setEvalLoading(false); }
  }

  async function handleNewVersion() {
    if (!versionFile || !evidence) return;
    setVersionError('');
    setVersionLoading(true);
    try {
      const res = await createEvidenceVersion(evidenceId, {
        title: versionForm.title || evidence.title,
        file_name: versionFile.name,
        file_size: versionFile.size,
        mime_type: versionFile.type || 'application/octet-stream',
        collection_date: versionForm.collection_date,
        freshness_period_days: versionForm.freshness_period_days ? parseInt(versionForm.freshness_period_days) : undefined,
      });
      // Upload to presigned URL
      if (res.data.upload?.presigned_url) {
        await fetch(res.data.upload.presigned_url, {
          method: res.data.upload.method || 'PUT',
          headers: { 'Content-Type': versionFile.type || 'application/octet-stream' },
          body: versionFile,
        });
      }
      await confirmEvidenceUpload(res.data.id);
      setShowVersionDialog(false);
      setVersionFile(null);
      // Navigate to the new version
      router.push(`/evidence/${res.data.id}`);
    } catch (err) {
      setVersionError(err instanceof Error ? err.message : 'Upload failed');
    } finally { setVersionLoading(false); }
  }

  async function handleStatusChange(status: string) {
    try {
      await changeEvidenceStatus(evidenceId, status);
      fetchData();
    } catch { /* handle */ }
  }

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center py-24">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  if (!evidence) return null;

  return (
    <div className="p-6 space-y-6">
      {/* Back + Header */}
      <div>
        <Button variant="ghost" size="sm" className="mb-2" onClick={() => router.push('/evidence')}>
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Evidence Library
        </Button>
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-3xl font-bold tracking-tight">{evidence.title}</h1>
              <Badge variant="outline" className="text-sm">
                {EVIDENCE_TYPE_LABELS[evidence.evidence_type] || evidence.evidence_type}
              </Badge>
              {evidence.version > 1 && (
                <Badge variant="secondary" className="text-sm">v{evidence.version}</Badge>
              )}
            </div>
            <p className="text-sm text-muted-foreground mt-1 flex items-center gap-2">
              <FileText className="h-4 w-4" />
              {evidence.file_name} · {formatFileSize(evidence.file_size)}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className={`text-sm ${EVIDENCE_STATUS_COLORS[evidence.status] || ''}`}>
              {EVIDENCE_STATUS_LABELS[evidence.status] || evidence.status}
            </Badge>
            <FreshnessBadge
              status={evidence.freshness_status}
              daysUntilExpiry={evidence.days_until_expiry}
            />
          </div>
        </div>
      </div>

      {/* Action buttons */}
      <div className="flex flex-wrap gap-2">
        <Button variant="outline" size="sm" onClick={handleDownload}>
          <Download className="h-4 w-4 mr-2" />Download
        </Button>
        {canLink && (
          <Button variant="outline" size="sm" onClick={() => setShowLinkDialog(true)}>
            <Link2 className="h-4 w-4 mr-2" />Link to Control
          </Button>
        )}
        {canEvaluate && (
          <Button variant="outline" size="sm" onClick={() => setShowEvalDialog(true)}>
            <Star className="h-4 w-4 mr-2" />Evaluate
          </Button>
        )}
        {canVersion && (
          <Button variant="outline" size="sm" onClick={() => setShowVersionDialog(true)}>
            <Upload className="h-4 w-4 mr-2" />Upload New Version
          </Button>
        )}
        {evidence.status === 'draft' && canManage && (
          <Button size="sm" onClick={() => handleStatusChange('pending_review')}>
            <Clock className="h-4 w-4 mr-2" />Submit for Review
          </Button>
        )}
        {evidence.status === 'pending_review' && canManage && (
          <>
            <Button size="sm" variant="default" onClick={() => handleStatusChange('approved')}>
              <CheckCircle2 className="h-4 w-4 mr-2" />Approve
            </Button>
            <Button size="sm" variant="destructive" onClick={() => handleStatusChange('rejected')}>
              <XCircle className="h-4 w-4 mr-2" />Reject
            </Button>
          </>
        )}
      </div>

      {/* Quick stats */}
      <div className="grid gap-4 md:grid-cols-5">
        <Card>
          <CardContent className="pt-6 flex items-center gap-3">
            <Link2 className="h-5 w-5 text-muted-foreground" />
            <div>
              <div className="text-2xl font-bold">{links.length}</div>
              <p className="text-xs text-muted-foreground">Linked Controls</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6 flex items-center gap-3">
            <Star className="h-5 w-5 text-muted-foreground" />
            <div>
              <div className="text-2xl font-bold">{evaluations.length}</div>
              <p className="text-xs text-muted-foreground">Evaluations</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6 flex items-center gap-3">
            <History className="h-5 w-5 text-muted-foreground" />
            <div>
              <div className="text-2xl font-bold">{evidence.total_versions || versions.length}</div>
              <p className="text-xs text-muted-foreground">Versions</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6 flex items-center gap-3">
            <User className="h-5 w-5 text-muted-foreground" />
            <div>
              <div className="text-sm font-medium">{evidence.uploaded_by?.name || 'Unknown'}</div>
              <p className="text-xs text-muted-foreground">Uploaded By</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6 flex items-center gap-3">
            <Clock className="h-5 w-5 text-muted-foreground" />
            <div>
              <div className="text-sm font-medium">
                {evidence.expires_at ? new Date(evidence.expires_at).toLocaleDateString() : 'No expiry'}
              </div>
              <p className="text-xs text-muted-foreground">Expires</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="overview">
            <FileText className="h-4 w-4 mr-1" />Overview
          </TabsTrigger>
          <TabsTrigger value="links">
            <Link2 className="h-4 w-4 mr-1" />Links ({links.length})
          </TabsTrigger>
          <TabsTrigger value="evaluations">
            <Star className="h-4 w-4 mr-1" />Evaluations ({evaluations.length})
          </TabsTrigger>
          <TabsTrigger value="versions">
            <History className="h-4 w-4 mr-1" />Versions ({versions.length})
          </TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6 mt-4">
          {evidence.description && (
            <Card>
              <CardHeader><CardTitle>Description</CardTitle></CardHeader>
              <CardContent>
                <p className="text-sm whitespace-pre-wrap">{evidence.description}</p>
              </CardContent>
            </Card>
          )}

          <Card>
            <CardHeader><CardTitle className="text-base">Details</CardTitle></CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="text-muted-foreground">Collection Method</p>
                  <p className="font-medium">{COLLECTION_METHOD_LABELS[evidence.collection_method] || evidence.collection_method}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Collection Date</p>
                  <p className="font-medium">{new Date(evidence.collection_date).toLocaleDateString()}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Freshness Period</p>
                  <p className="font-medium">{evidence.freshness_period_days ? `${evidence.freshness_period_days} days` : 'Not set'}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Source System</p>
                  <p className="font-medium">{evidence.source_system || 'Not specified'}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">MIME Type</p>
                  <p className="font-mono text-xs">{evidence.mime_type}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Created</p>
                  <p className="font-medium">{new Date(evidence.created_at).toLocaleDateString()}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Updated</p>
                  <p className="font-medium">{new Date(evidence.updated_at).toLocaleDateString()}</p>
                </div>
                {evidence.checksum_sha256 && (
                  <div>
                    <p className="text-muted-foreground">Checksum (SHA-256)</p>
                    <p className="font-mono text-xs truncate">{evidence.checksum_sha256}</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {evidence.tags && evidence.tags.length > 0 && (
            <Card>
              <CardHeader><CardTitle className="text-base">Tags</CardTitle></CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {evidence.tags.map(tag => (
                    <Badge key={tag} variant="secondary">{tag}</Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Latest evaluation */}
          {evidence.latest_evaluation && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Latest Evaluation</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-4">
                  <Badge variant="outline" className={VERDICT_CONFIG[evidence.latest_evaluation.verdict]?.className || ''}>
                    {VERDICT_CONFIG[evidence.latest_evaluation.verdict]?.label || evidence.latest_evaluation.verdict}
                  </Badge>
                  <span className="text-sm text-muted-foreground">
                    Confidence: {evidence.latest_evaluation.confidence}
                  </span>
                  {evidence.latest_evaluation.evaluated_by && (
                    <span className="text-sm text-muted-foreground">
                      by {evidence.latest_evaluation.evaluated_by.name}
                    </span>
                  )}
                </div>
                {evidence.latest_evaluation.comments && (
                  <p className="text-sm mt-2">{evidence.latest_evaluation.comments}</p>
                )}
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* Links Tab (Task 4: Evidence-to-control linking UI) */}
        <TabsContent value="links" className="space-y-6 mt-4">
          {canLink && (
            <div className="flex justify-end">
              <Button size="sm" onClick={() => setShowLinkDialog(true)}>
                <Plus className="h-4 w-4 mr-2" />Link to Control
              </Button>
            </div>
          )}

          {links.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <Link2 className="h-12 w-12 mb-4 opacity-50" />
                <p className="text-lg font-medium">No links</p>
                <p className="text-sm">Link this evidence to controls or requirements</p>
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[80px]">Type</TableHead>
                      <TableHead className="w-[120px]">Identifier</TableHead>
                      <TableHead>Title</TableHead>
                      <TableHead className="w-[100px]">Strength</TableHead>
                      <TableHead>Notes</TableHead>
                      <TableHead className="w-[100px]">Linked By</TableHead>
                      {canLink && <TableHead className="w-[50px]" />}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {links.map(link => (
                      <TableRow key={link.id}>
                        <TableCell>
                          <Badge variant="outline" className="text-xs">
                            {link.target_type === 'control' ? 'Control' : 'Requirement'}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono text-sm">
                          {link.target_type === 'control' ? (
                            <Link href={`/controls/${link.control?.id}`} className="hover:text-primary">
                              {link.control?.identifier}
                            </Link>
                          ) : (
                            link.requirement?.identifier
                          )}
                        </TableCell>
                        <TableCell className="max-w-xs">
                          <span className="line-clamp-1 text-sm">
                            {link.target_type === 'control' ? link.control?.title : link.requirement?.title}
                          </span>
                        </TableCell>
                        <TableCell>
                          <span className={`text-xs font-medium px-2 py-1 rounded-full ${STRENGTH_COLORS[link.strength] || ''}`}>
                            {STRENGTH_LABELS[link.strength] || link.strength}
                          </span>
                        </TableCell>
                        <TableCell className="max-w-xs">
                          <span className="line-clamp-1 text-xs text-muted-foreground">
                            {link.notes || '—'}
                          </span>
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {link.linked_by?.name || '—'}
                        </TableCell>
                        {canLink && (
                          <TableCell>
                            <Button
                              variant="ghost" size="icon" className="h-8 w-8"
                              onClick={() => handleDeleteLink(link.id)}
                            >
                              <Trash2 className="h-4 w-4 text-destructive" />
                            </Button>
                          </TableCell>
                        )}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* Evaluations Tab (Task 8) */}
        <TabsContent value="evaluations" className="space-y-6 mt-4">
          {canEvaluate && (
            <div className="flex justify-end">
              <Button size="sm" onClick={() => setShowEvalDialog(true)}>
                <Plus className="h-4 w-4 mr-2" />New Evaluation
              </Button>
            </div>
          )}

          {evaluations.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <Star className="h-12 w-12 mb-4 opacity-50" />
                <p className="text-lg font-medium">No evaluations</p>
                <p className="text-sm">Submit an evaluation to assess this evidence</p>
              </CardContent>
            </Card>
          ) : (
            evaluations.map(ev => (
              <Card key={ev.id}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <Badge variant="outline" className={VERDICT_CONFIG[ev.verdict]?.className || ''}>
                        {VERDICT_CONFIG[ev.verdict]?.label || ev.verdict}
                      </Badge>
                      <Badge variant="secondary" className="text-xs">
                        Confidence: {ev.confidence}
                      </Badge>
                      {ev.evidence_link?.control_identifier && (
                        <Badge variant="outline" className="text-xs">
                          {ev.evidence_link.control_identifier}
                        </Badge>
                      )}
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {new Date(ev.created_at).toLocaleString()}
                    </span>
                  </div>
                  <CardDescription>
                    Evaluated by {ev.evaluated_by?.name}
                    {ev.evaluated_by?.role && ` (${ev.evaluated_by.role})`}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="text-sm whitespace-pre-wrap">{ev.comments}</p>
                  {ev.missing_elements && ev.missing_elements.length > 0 && (
                    <div>
                      <p className="text-sm font-medium text-amber-600 dark:text-amber-400 mb-1">Missing Elements:</p>
                      <ul className="list-disc list-inside text-sm text-muted-foreground">
                        {ev.missing_elements.map((el, i) => <li key={i}>{el}</li>)}
                      </ul>
                    </div>
                  )}
                  {ev.remediation_notes && (
                    <div>
                      <p className="text-sm font-medium text-blue-600 dark:text-blue-400 mb-1">Remediation Notes:</p>
                      <p className="text-sm text-muted-foreground">{ev.remediation_notes}</p>
                    </div>
                  )}
                </CardContent>
              </Card>
            ))
          )}
        </TabsContent>

        {/* Versions Tab (Task 9: version comparison) */}
        <TabsContent value="versions" className="space-y-6 mt-4">
          {canVersion && (
            <div className="flex justify-end">
              <Button size="sm" onClick={() => setShowVersionDialog(true)}>
                <Upload className="h-4 w-4 mr-2" />Upload New Version
              </Button>
            </div>
          )}

          {versions.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <History className="h-12 w-12 mb-4 opacity-50" />
                <p className="text-lg font-medium">No version history</p>
              </CardContent>
            </Card>
          ) : (
            <>
              {/* Version comparison view */}
              {versions.length >= 2 && (
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Version Comparison</CardTitle>
                    <CardDescription>Compare metadata between versions</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 gap-4 mb-4">
                      <div>
                        <Label className="text-xs">Version A</Label>
                        <Select
                          value={compareVersions[0] || ''}
                          onValueChange={v => setCompareVersions([v, compareVersions[1]])}
                        >
                          <SelectTrigger><SelectValue placeholder="Select version" /></SelectTrigger>
                          <SelectContent>
                            {versions.map(v => (
                              <SelectItem key={v.id} value={v.id}>
                                v{v.version} — {v.title} ({new Date(v.collection_date).toLocaleDateString()})
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      <div>
                        <Label className="text-xs">Version B</Label>
                        <Select
                          value={compareVersions[1] || ''}
                          onValueChange={v => setCompareVersions([compareVersions[0], v])}
                        >
                          <SelectTrigger><SelectValue placeholder="Select version" /></SelectTrigger>
                          <SelectContent>
                            {versions.map(v => (
                              <SelectItem key={v.id} value={v.id}>
                                v{v.version} — {v.title} ({new Date(v.collection_date).toLocaleDateString()})
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                    {compareVersions[0] && compareVersions[1] && (() => {
                      const va = versions.find(v => v.id === compareVersions[0]);
                      const vb = versions.find(v => v.id === compareVersions[1]);
                      if (!va || !vb) return null;
                      const fields: { label: string; a: string; b: string }[] = [
                        { label: 'Version', a: `v${va.version}`, b: `v${vb.version}` },
                        { label: 'Title', a: va.title, b: vb.title },
                        { label: 'Status', a: EVIDENCE_STATUS_LABELS[va.status] || va.status, b: EVIDENCE_STATUS_LABELS[vb.status] || vb.status },
                        { label: 'File Name', a: va.file_name, b: vb.file_name },
                        { label: 'File Size', a: formatFileSize(va.file_size), b: formatFileSize(vb.file_size) },
                        { label: 'Collection Date', a: new Date(va.collection_date).toLocaleDateString(), b: new Date(vb.collection_date).toLocaleDateString() },
                        { label: 'Uploaded By', a: va.uploaded_by?.name || '—', b: vb.uploaded_by?.name || '—' },
                        { label: 'Created', a: new Date(va.created_at).toLocaleString(), b: new Date(vb.created_at).toLocaleString() },
                      ];
                      return (
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead className="w-[140px]">Field</TableHead>
                              <TableHead>Version A (v{va.version})</TableHead>
                              <TableHead>Version B (v{vb.version})</TableHead>
                              <TableHead className="w-[80px]">Changed</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {fields.map(f => (
                              <TableRow key={f.label} className={f.a !== f.b ? 'bg-amber-500/5' : ''}>
                                <TableCell className="font-medium text-sm">{f.label}</TableCell>
                                <TableCell className="text-sm">{f.a}</TableCell>
                                <TableCell className="text-sm">{f.b}</TableCell>
                                <TableCell className="text-center">
                                  {f.a !== f.b ? (
                                    <Badge variant="outline" className="text-xs bg-amber-500/10 text-amber-700 dark:text-amber-400">
                                      Changed
                                    </Badge>
                                  ) : (
                                    <span className="text-xs text-muted-foreground">—</span>
                                  )}
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      );
                    })()}
                  </CardContent>
                </Card>
              )}

              {/* Version list */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Version History</CardTitle>
                </CardHeader>
                <CardContent className="p-0">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[80px]">Version</TableHead>
                        <TableHead>Title</TableHead>
                        <TableHead className="w-[100px]">Status</TableHead>
                        <TableHead className="w-[100px]">File</TableHead>
                        <TableHead className="w-[100px]">Collected</TableHead>
                        <TableHead className="w-[100px]">Uploaded By</TableHead>
                        <TableHead className="w-[120px]">Created</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {versions.map(v => (
                        <TableRow key={v.id} className={v.is_current ? 'bg-primary/5' : ''}>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <span className="font-mono font-medium">v{v.version}</span>
                              {v.is_current && (
                                <Badge variant="default" className="text-[10px]">Current</Badge>
                              )}
                            </div>
                          </TableCell>
                          <TableCell>
                            {v.id === evidenceId ? (
                              <span className="text-sm font-medium">{v.title}</span>
                            ) : (
                              <Link href={`/evidence/${v.id}`} className="text-sm font-medium hover:text-primary">
                                {v.title}
                              </Link>
                            )}
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline" className={`text-xs ${EVIDENCE_STATUS_COLORS[v.status] || ''}`}>
                              {EVIDENCE_STATUS_LABELS[v.status] || v.status}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-xs text-muted-foreground">
                            {formatFileSize(v.file_size)}
                          </TableCell>
                          <TableCell className="text-xs">
                            {new Date(v.collection_date).toLocaleDateString()}
                          </TableCell>
                          <TableCell className="text-xs text-muted-foreground">
                            {v.uploaded_by?.name || '—'}
                          </TableCell>
                          <TableCell className="text-xs text-muted-foreground">
                            {new Date(v.created_at).toLocaleString()}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
            </>
          )}
        </TabsContent>
      </Tabs>

      {/* Link to Control Dialog (Task 4) */}
      <Dialog open={showLinkDialog} onOpenChange={setShowLinkDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Link Evidence to Control</DialogTitle>
            <DialogDescription>
              Associate this evidence artifact with a control from your library
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Search Controls</Label>
              <Input
                placeholder="Search by identifier or title..."
                value={linkControlSearch}
                onChange={(e) => setLinkControlSearch(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>Select Control</Label>
              <Select value={linkSelectedControl} onValueChange={setLinkSelectedControl}>
                <SelectTrigger><SelectValue placeholder="Choose a control..." /></SelectTrigger>
                <SelectContent>
                  {linkControls.map(c => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.identifier} — {c.title}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Strength</Label>
              <Select value={linkStrength} onValueChange={setLinkStrength}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="primary">Primary — Direct evidence</SelectItem>
                  <SelectItem value="supporting">Supporting — Supplementary evidence</SelectItem>
                  <SelectItem value="supplementary">Supplementary — Indirect evidence</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Notes (optional)</Label>
              <Textarea
                placeholder="How does this evidence support the control?"
                value={linkNotes}
                onChange={(e) => setLinkNotes(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowLinkDialog(false)}>Cancel</Button>
            <Button onClick={handleCreateLink} disabled={linkLoading || !linkSelectedControl}>
              {linkLoading ? 'Linking...' : 'Link Evidence'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Evaluation Dialog (Task 8) */}
      <Dialog open={showEvalDialog} onOpenChange={setShowEvalDialog}>
        <DialogContent className="sm:max-w-[550px]">
          <DialogHeader>
            <DialogTitle>Evaluate Evidence</DialogTitle>
            <DialogDescription>
              Assess the sufficiency of this evidence artifact
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Verdict *</Label>
                <Select value={evalForm.verdict} onValueChange={v => setEvalForm(f => ({ ...f, verdict: v }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="sufficient">Sufficient</SelectItem>
                    <SelectItem value="partial">Partial</SelectItem>
                    <SelectItem value="insufficient">Insufficient</SelectItem>
                    <SelectItem value="needs_update">Needs Update</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Confidence</Label>
                <Select value={evalForm.confidence} onValueChange={v => setEvalForm(f => ({ ...f, confidence: v }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="high">High</SelectItem>
                    <SelectItem value="medium">Medium</SelectItem>
                    <SelectItem value="low">Low</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="space-y-2">
              <Label>Comments *</Label>
              <Textarea
                placeholder="Describe your assessment of this evidence..."
                value={evalForm.comments}
                onChange={(e) => setEvalForm(f => ({ ...f, comments: e.target.value }))}
                className="min-h-[100px]"
              />
            </div>
            <div className="space-y-2">
              <Label>Missing Elements (comma-separated)</Label>
              <Input
                placeholder="e.g. timestamps, user names, scope coverage"
                value={evalForm.missing_elements}
                onChange={(e) => setEvalForm(f => ({ ...f, missing_elements: e.target.value }))}
              />
            </div>
            {(evalForm.verdict === 'insufficient' || evalForm.verdict === 'partial' || evalForm.verdict === 'needs_update') && (
              <div className="space-y-2">
                <Label>Remediation Notes</Label>
                <Textarea
                  placeholder="What needs to be done to make this evidence sufficient?"
                  value={evalForm.remediation_notes}
                  onChange={(e) => setEvalForm(f => ({ ...f, remediation_notes: e.target.value }))}
                />
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowEvalDialog(false)}>Cancel</Button>
            <Button onClick={handleEvaluation} disabled={evalLoading || !evalForm.comments}>
              {evalLoading ? 'Submitting...' : 'Submit Evaluation'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* New Version Dialog */}
      <Dialog open={showVersionDialog} onOpenChange={setShowVersionDialog}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Upload New Version</DialogTitle>
            <DialogDescription>
              Upload a new version of &quot;{evidence.title}&quot;. The current version will be superseded.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            {versionError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive flex items-center gap-2">
                <AlertTriangle className="h-4 w-4" />{versionError}
              </div>
            )}
            <div className="space-y-2">
              <Label>File *</Label>
              <Input
                type="file"
                onChange={(e) => setVersionFile(e.target.files?.[0] || null)}
              />
            </div>
            <div className="space-y-2">
              <Label>Title (optional, defaults to current)</Label>
              <Input
                placeholder={evidence.title}
                value={versionForm.title}
                onChange={(e) => setVersionForm(f => ({ ...f, title: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Collection Date *</Label>
                <Input
                  type="date"
                  value={versionForm.collection_date}
                  onChange={(e) => setVersionForm(f => ({ ...f, collection_date: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label>Freshness Period (days)</Label>
                <Input
                  type="number"
                  placeholder={String(evidence.freshness_period_days || '')}
                  value={versionForm.freshness_period_days}
                  onChange={(e) => setVersionForm(f => ({ ...f, freshness_period_days: e.target.value }))}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowVersionDialog(false)}>Cancel</Button>
            <Button onClick={handleNewVersion} disabled={versionLoading || !versionFile}>
              {versionLoading ? 'Uploading...' : 'Upload Version'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
