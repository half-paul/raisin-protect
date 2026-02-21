'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
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
  ArrowLeft, Edit, Archive, Shield, AlertTriangle, TrendingDown, Link2,
  Plus, CheckCircle2, XCircle, Clock, Target, Search,
} from 'lucide-react';
import {
  Risk, RiskAssessment, RiskTreatment, RiskControl, Control,
  getRisk, archiveRisk, changeRiskStatus,
  listRiskAssessments, createRiskAssessment,
  listRiskTreatments, createRiskTreatment, updateRiskTreatment, completeRiskTreatment,
  listRiskControls, linkRiskControl, unlinkRiskControl, updateRiskControlEffectiveness,
  listControls,
} from '@/lib/api';
import {
  RISK_STATUS_LABELS, RISK_STATUS_COLORS,
  RISK_CATEGORY_LABELS,
  SEVERITY_LABELS, SEVERITY_COLORS,
  LIKELIHOOD_LABELS, LIKELIHOOD_ORDER,
  IMPACT_LABELS, IMPACT_ORDER,
  TREATMENT_TYPE_LABELS, TREATMENT_STATUS_LABELS, TREATMENT_STATUS_COLORS,
  EFFECTIVENESS_LABELS, EFFECTIVENESS_COLORS,
  ASSESSMENT_TYPE_LABELS, ASSESSMENT_STATUS_LABELS, ASSESSMENT_STATUS_COLORS,
  PRIORITY_LABELS, PRIORITY_COLORS,
  scoreToSeverity,
} from '@/components/risk/constants';

export default function RiskDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { hasRole } = useAuth();
  const riskId = params.id as string;

  const canEdit = hasRole('ciso', 'compliance_manager', 'security_engineer');
  const canArchive = hasRole('ciso', 'compliance_manager');

  const [risk, setRisk] = useState<Risk | null>(null);
  const [assessments, setAssessments] = useState<RiskAssessment[]>([]);
  const [treatments, setTreatments] = useState<RiskTreatment[]>([]);
  const [controls, setControls] = useState<RiskControl[]>([]);
  const [loading, setLoading] = useState(true);

  // Assessment dialog
  const [showAssessment, setShowAssessment] = useState(false);
  const [assessForm, setAssessForm] = useState({
    assessment_type: 'residual' as string,
    likelihood: '' as string,
    impact: '' as string,
    justification: '',
    assumptions: '',
    valid_until: '',
  });
  const [assessSaving, setAssessSaving] = useState(false);

  // Treatment dialog
  const [showTreatment, setShowTreatment] = useState(false);
  const [treatForm, setTreatForm] = useState({
    treatment_type: 'mitigate' as string,
    title: '',
    description: '',
    priority: 'medium',
    due_date: '',
    estimated_effort_hours: '',
    expected_residual_likelihood: '' as string,
    expected_residual_impact: '' as string,
    notes: '',
  });
  const [treatSaving, setTreatSaving] = useState(false);

  // Control linking dialog
  const [showLinkControl, setShowLinkControl] = useState(false);
  const [availableControls, setAvailableControls] = useState<Control[]>([]);
  const [controlSearch, setControlSearch] = useState('');
  const [selectedControlId, setSelectedControlId] = useState('');
  const [linkEffectiveness, setLinkEffectiveness] = useState('not_assessed');
  const [linkMitigation, setLinkMitigation] = useState('');
  const [linkNotes, setLinkNotes] = useState('');
  const [linkSaving, setLinkSaving] = useState(false);

  // Complete treatment dialog
  const [showComplete, setShowComplete] = useState<string | null>(null);
  const [completeForm, setCompleteForm] = useState({
    actual_effort_hours: '',
    effectiveness_rating: '' as string,
    effectiveness_notes: '',
  });
  const [completeSaving, setCompleteSaving] = useState(false);

  const fetchAll = useCallback(async () => {
    try {
      setLoading(true);
      const [riskRes, assessRes, treatRes, ctrlRes] = await Promise.all([
        getRisk(riskId),
        listRiskAssessments(riskId),
        listRiskTreatments(riskId),
        listRiskControls(riskId),
      ]);
      setRisk(riskRes.data);
      setAssessments(assessRes.data);
      setTreatments(treatRes.data);
      setControls(ctrlRes.data);
    } catch (err) {
      console.error('Failed to fetch risk:', err);
    } finally {
      setLoading(false);
    }
  }, [riskId]);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  const handleArchive = async () => {
    if (!risk || !confirm('Archive this risk? Active treatments will be cancelled.')) return;
    try {
      await archiveRisk(risk.id);
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to archive');
    }
  };

  // Assessment submission
  const handleCreateAssessment = async () => {
    if (!assessForm.likelihood || !assessForm.impact) return;
    try {
      setAssessSaving(true);
      await createRiskAssessment(riskId, {
        assessment_type: assessForm.assessment_type,
        likelihood: assessForm.likelihood,
        impact: assessForm.impact,
        justification: assessForm.justification || undefined,
        assumptions: assessForm.assumptions || undefined,
        valid_until: assessForm.valid_until || undefined,
      });
      setShowAssessment(false);
      setAssessForm({ assessment_type: 'residual', likelihood: '', impact: '', justification: '', assumptions: '', valid_until: '' });
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create assessment');
    } finally {
      setAssessSaving(false);
    }
  };

  // Computed preview score
  const previewScore = assessForm.likelihood && assessForm.impact
    ? (LIKELIHOOD_ORDER.indexOf(assessForm.likelihood as any) + 1) * (IMPACT_ORDER.indexOf(assessForm.impact as any) + 1)
    : null;

  // Treatment submission
  const handleCreateTreatment = async () => {
    if (!treatForm.title || !treatForm.treatment_type) return;
    try {
      setTreatSaving(true);
      await createRiskTreatment(riskId, {
        treatment_type: treatForm.treatment_type,
        title: treatForm.title,
        description: treatForm.description || undefined,
        priority: treatForm.priority || undefined,
        due_date: treatForm.due_date || undefined,
        estimated_effort_hours: treatForm.estimated_effort_hours ? Number(treatForm.estimated_effort_hours) : undefined,
        expected_residual_likelihood: treatForm.expected_residual_likelihood || undefined,
        expected_residual_impact: treatForm.expected_residual_impact || undefined,
        notes: treatForm.notes || undefined,
      });
      setShowTreatment(false);
      setTreatForm({ treatment_type: 'mitigate', title: '', description: '', priority: 'medium', due_date: '', estimated_effort_hours: '', expected_residual_likelihood: '', expected_residual_impact: '', notes: '' });
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to create treatment');
    } finally {
      setTreatSaving(false);
    }
  };

  // Link control
  const handleOpenLinkControl = async () => {
    try {
      const res = await listControls({ status: 'active', per_page: '100' });
      setAvailableControls(res.data);
    } catch (err) {
      console.error(err);
    }
    setShowLinkControl(true);
  };

  const handleLinkControl = async () => {
    if (!selectedControlId) return;
    try {
      setLinkSaving(true);
      await linkRiskControl(riskId, {
        control_id: selectedControlId,
        effectiveness: linkEffectiveness || undefined,
        mitigation_percentage: linkMitigation ? Number(linkMitigation) : undefined,
        notes: linkNotes || undefined,
      });
      setShowLinkControl(false);
      setSelectedControlId('');
      setLinkEffectiveness('not_assessed');
      setLinkMitigation('');
      setLinkNotes('');
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to link control');
    } finally {
      setLinkSaving(false);
    }
  };

  const handleUnlinkControl = async (controlId: string) => {
    if (!confirm('Unlink this control from the risk?')) return;
    try {
      await unlinkRiskControl(riskId, controlId);
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to unlink');
    }
  };

  // Complete treatment
  const handleCompleteTreatment = async () => {
    if (!showComplete) return;
    try {
      setCompleteSaving(true);
      await completeRiskTreatment(riskId, showComplete, {
        actual_effort_hours: completeForm.actual_effort_hours ? Number(completeForm.actual_effort_hours) : undefined,
        effectiveness_rating: completeForm.effectiveness_rating || undefined,
        effectiveness_notes: completeForm.effectiveness_notes || undefined,
      });
      setShowComplete(null);
      setCompleteForm({ actual_effort_hours: '', effectiveness_rating: '', effectiveness_notes: '' });
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to complete treatment');
    } finally {
      setCompleteSaving(false);
    }
  };

  // Treatment status update
  const handleTreatmentStatusChange = async (treatmentId: string, newStatus: string) => {
    try {
      await updateRiskTreatment(riskId, treatmentId, { status: newStatus });
      fetchAll();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update treatment');
    }
  };

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-muted-foreground">Loading risk...</div>;
  }

  if (!risk) {
    return <div className="flex items-center justify-center h-64 text-muted-foreground">Risk not found</div>;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => router.push('/risks')}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold">{risk.title}</h1>
            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${RISK_STATUS_COLORS[risk.status]}`}>
              {RISK_STATUS_LABELS[risk.status]}
            </span>
            {risk.appetite_breached && (
              <Badge variant="destructive" className="text-xs">Appetite Breached</Badge>
            )}
          </div>
          <p className="text-sm text-muted-foreground">{risk.identifier} · {RISK_CATEGORY_LABELS[risk.category] || risk.category}</p>
        </div>
        <div className="flex gap-2">
          {canEdit && (
            <Link href={`/risks/${risk.id}/edit`}>
              <Button variant="outline" size="sm"><Edit className="h-4 w-4 mr-1" /> Edit</Button>
            </Link>
          )}
          {canArchive && risk.status !== 'archived' && (
            <Button variant="outline" size="sm" className="text-red-500" onClick={handleArchive}>
              <Archive className="h-4 w-4 mr-1" /> Archive
            </Button>
          )}
        </div>
      </div>

      {/* Score Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground mb-1">Inherent Risk</p>
            {risk.inherent_score ? (
              <>
                <div className="text-2xl font-bold">{risk.inherent_score.score}</div>
                <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[risk.inherent_score.severity]}`}>
                  {SEVERITY_LABELS[risk.inherent_score.severity]}
                </span>
                <p className="text-xs text-muted-foreground mt-1">
                  {LIKELIHOOD_LABELS[risk.inherent_score.likelihood!]} × {IMPACT_LABELS[risk.inherent_score.impact!]}
                </p>
              </>
            ) : <div className="text-muted-foreground">Not assessed</div>}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground mb-1">Residual Risk</p>
            {risk.residual_score ? (
              <>
                <div className="text-2xl font-bold">{risk.residual_score.score}</div>
                <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[risk.residual_score.severity]}`}>
                  {SEVERITY_LABELS[risk.residual_score.severity]}
                </span>
                <p className="text-xs text-muted-foreground mt-1">
                  {LIKELIHOOD_LABELS[risk.residual_score.likelihood!]} × {IMPACT_LABELS[risk.residual_score.impact!]}
                </p>
              </>
            ) : <div className="text-muted-foreground">Not assessed</div>}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground mb-1">Risk Appetite</p>
            <div className="text-2xl font-bold">{risk.risk_appetite_threshold ?? '—'}</div>
            {risk.appetite_breached ? (
              <Badge variant="destructive" className="text-xs">Breached</Badge>
            ) : risk.risk_appetite_threshold ? (
              <Badge variant="outline" className="text-xs text-green-600">Within Appetite</Badge>
            ) : null}
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground mb-1">Assessment Status</p>
            {risk.assessment_status ? (
              <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${ASSESSMENT_STATUS_COLORS[risk.assessment_status]}`}>
                {ASSESSMENT_STATUS_LABELS[risk.assessment_status]}
              </span>
            ) : <div className="text-muted-foreground">—</div>}
            {risk.next_assessment_at && (
              <p className="text-xs text-muted-foreground mt-1">Next: {new Date(risk.next_assessment_at).toLocaleDateString()}</p>
            )}
            {risk.last_assessed_at && (
              <p className="text-xs text-muted-foreground">Last: {new Date(risk.last_assessed_at).toLocaleDateString()}</p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Description & Metadata */}
      {(risk.description || risk.owner || risk.affected_assets?.length) && (
        <Card>
          <CardContent className="p-4 space-y-3">
            {risk.description && <p className="text-sm">{risk.description}</p>}
            <div className="flex flex-wrap gap-4 text-sm">
              {risk.owner && <div><span className="text-muted-foreground">Owner:</span> {risk.owner.name}</div>}
              {risk.secondary_owner && <div><span className="text-muted-foreground">Secondary:</span> {risk.secondary_owner.name}</div>}
              {risk.source && <div><span className="text-muted-foreground">Source:</span> {risk.source}</div>}
            </div>
            {risk.affected_assets && risk.affected_assets.length > 0 && (
              <div className="flex flex-wrap gap-1">
                <span className="text-xs text-muted-foreground mr-1">Assets:</span>
                {risk.affected_assets.map(a => <Badge key={a} variant="outline" className="text-xs">{a}</Badge>)}
              </div>
            )}
            {risk.tags && risk.tags.length > 0 && (
              <div className="flex flex-wrap gap-1">
                <span className="text-xs text-muted-foreground mr-1">Tags:</span>
                {risk.tags.map(t => <Badge key={t} variant="secondary" className="text-xs">{t}</Badge>)}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Acceptance info */}
      {risk.acceptance && (
        <Card className="border-yellow-500/50">
          <CardContent className="p-4">
            <h3 className="font-semibold text-sm mb-2">Risk Accepted</h3>
            <p className="text-sm">{risk.acceptance.justification}</p>
            <div className="flex gap-4 mt-2 text-xs text-muted-foreground">
              <span>By: {risk.acceptance.accepted_by.name}</span>
              <span>Expiry: {new Date(risk.acceptance.expiry).toLocaleDateString()}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Tabs */}
      <Tabs defaultValue="assessments">
        <TabsList>
          <TabsTrigger value="assessments">Assessments ({assessments.length})</TabsTrigger>
          <TabsTrigger value="treatments">Treatments ({treatments.length})</TabsTrigger>
          <TabsTrigger value="controls">Controls ({controls.length})</TabsTrigger>
        </TabsList>

        {/* Assessments Tab */}
        <TabsContent value="assessments" className="space-y-4">
          <div className="flex justify-between items-center">
            <h3 className="text-lg font-semibold">Risk Assessments</h3>
            {canEdit && (
              <Button size="sm" onClick={() => setShowAssessment(true)}>
                <Plus className="h-4 w-4 mr-1" /> New Assessment
              </Button>
            )}
          </div>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Type</TableHead>
                    <TableHead>Likelihood</TableHead>
                    <TableHead>Impact</TableHead>
                    <TableHead>Score</TableHead>
                    <TableHead>Severity</TableHead>
                    <TableHead>Assessor</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Current</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {assessments.length === 0 ? (
                    <TableRow><TableCell colSpan={8} className="text-center py-6 text-muted-foreground">No assessments yet</TableCell></TableRow>
                  ) : (
                    assessments.map(a => (
                      <TableRow key={a.id}>
                        <TableCell>
                          <Badge variant="outline" className="text-xs">{ASSESSMENT_TYPE_LABELS[a.assessment_type] || a.assessment_type}</Badge>
                        </TableCell>
                        <TableCell className="text-sm">{LIKELIHOOD_LABELS[a.likelihood] || a.likelihood}</TableCell>
                        <TableCell className="text-sm">{IMPACT_LABELS[a.impact] || a.impact}</TableCell>
                        <TableCell className="font-mono font-medium">{a.overall_score}</TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[a.severity]}`}>
                            {SEVERITY_LABELS[a.severity]}
                          </span>
                        </TableCell>
                        <TableCell className="text-sm">{a.assessed_by?.name || '—'}</TableCell>
                        <TableCell className="text-sm">{new Date(a.assessment_date).toLocaleDateString()}</TableCell>
                        <TableCell>{a.is_current ? <CheckCircle2 className="h-4 w-4 text-green-500" /> : null}</TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Treatments Tab */}
        <TabsContent value="treatments" className="space-y-4">
          <div className="flex justify-between items-center">
            <h3 className="text-lg font-semibold">Treatment Plans</h3>
            {canEdit && (
              <Button size="sm" onClick={() => setShowTreatment(true)}>
                <Plus className="h-4 w-4 mr-1" /> New Treatment
              </Button>
            )}
          </div>
          <div className="space-y-3">
            {treatments.length === 0 ? (
              <Card><CardContent className="p-6 text-center text-muted-foreground">No treatment plans yet</CardContent></Card>
            ) : (
              treatments.map(t => (
                <Card key={t.id}>
                  <CardContent className="p-4">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h4 className="font-semibold text-sm">{t.title}</h4>
                          <Badge variant="outline" className="text-xs">{TREATMENT_TYPE_LABELS[t.treatment_type] || t.treatment_type}</Badge>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${TREATMENT_STATUS_COLORS[t.status]}`}>
                            {TREATMENT_STATUS_LABELS[t.status]}
                          </span>
                          {t.priority && (
                            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${PRIORITY_COLORS[t.priority] || ''}`}>
                              {PRIORITY_LABELS[t.priority] || t.priority}
                            </span>
                          )}
                        </div>
                        {t.description && <p className="text-sm text-muted-foreground mb-2">{t.description}</p>}
                        <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
                          {t.owner && <span>Owner: {t.owner.name}</span>}
                          {t.due_date && <span>Due: {new Date(t.due_date).toLocaleDateString()}</span>}
                          {t.started_at && <span>Started: {new Date(t.started_at).toLocaleDateString()}</span>}
                          {t.completed_at && <span>Completed: {new Date(t.completed_at).toLocaleDateString()}</span>}
                          {t.estimated_effort_hours != null && <span>Est: {t.estimated_effort_hours}h</span>}
                          {t.actual_effort_hours != null && <span>Actual: {t.actual_effort_hours}h</span>}
                        </div>
                        {t.expected_residual && (
                          <p className="text-xs mt-1">Expected residual: {t.expected_residual.score} ({scoreToSeverity(t.expected_residual.score)})</p>
                        )}
                        {t.effectiveness_rating && (
                          <p className="text-xs mt-1">Effectiveness: {t.effectiveness_rating} — {t.effectiveness_notes}</p>
                        )}
                      </div>
                      <div className="flex gap-1">
                        {t.status === 'planned' && canEdit && (
                          <Button variant="outline" size="sm" onClick={() => handleTreatmentStatusChange(t.id, 'in_progress')}>Start</Button>
                        )}
                        {(t.status === 'in_progress' || t.status === 'implemented') && canEdit && (
                          <Button variant="outline" size="sm" onClick={() => { setShowComplete(t.id); setCompleteForm({ actual_effort_hours: String(t.actual_effort_hours || ''), effectiveness_rating: '', effectiveness_notes: '' }); }}>
                            <CheckCircle2 className="h-4 w-4 mr-1" /> Complete
                          </Button>
                        )}
                        {(t.status === 'planned' || t.status === 'in_progress') && canEdit && (
                          <Button variant="ghost" size="sm" className="text-red-500" onClick={() => handleTreatmentStatusChange(t.id, 'cancelled')}>
                            <XCircle className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))
            )}
          </div>
        </TabsContent>

        {/* Controls Tab */}
        <TabsContent value="controls" className="space-y-4">
          <div className="flex justify-between items-center">
            <h3 className="text-lg font-semibold">Linked Controls</h3>
            {canEdit && (
              <Button size="sm" onClick={handleOpenLinkControl}>
                <Link2 className="h-4 w-4 mr-1" /> Link Control
              </Button>
            )}
          </div>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Identifier</TableHead>
                    <TableHead>Title</TableHead>
                    <TableHead>Effectiveness</TableHead>
                    <TableHead>Mitigation %</TableHead>
                    <TableHead>Last Review</TableHead>
                    <TableHead>Frameworks</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {controls.length === 0 ? (
                    <TableRow><TableCell colSpan={7} className="text-center py-6 text-muted-foreground">No controls linked</TableCell></TableRow>
                  ) : (
                    controls.map(c => (
                      <TableRow key={c.id}>
                        <TableCell className="font-mono text-xs">
                          <Link href={`/controls/${c.id}`} className="text-primary hover:underline">{c.identifier}</Link>
                        </TableCell>
                        <TableCell className="text-sm">{c.title}</TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${EFFECTIVENESS_COLORS[c.effectiveness]}`}>
                            {EFFECTIVENESS_LABELS[c.effectiveness]}
                          </span>
                        </TableCell>
                        <TableCell className="text-sm">{c.mitigation_percentage != null ? `${c.mitigation_percentage}%` : '—'}</TableCell>
                        <TableCell className="text-sm">{c.last_effectiveness_review ? new Date(c.last_effectiveness_review).toLocaleDateString() : '—'}</TableCell>
                        <TableCell>
                          <div className="flex flex-wrap gap-1">
                            {c.frameworks?.map(f => <Badge key={f} variant="outline" className="text-[10px]">{f}</Badge>)}
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          {canEdit && (
                            <Button variant="ghost" size="sm" className="text-red-500" onClick={() => handleUnlinkControl(c.id)}>
                              <XCircle className="h-4 w-4" />
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* ===== DIALOGS ===== */}

      {/* Assessment Dialog */}
      <Dialog open={showAssessment} onOpenChange={setShowAssessment}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>New Risk Assessment</DialogTitle>
            <DialogDescription>Score likelihood and impact for this risk</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Assessment Type</Label>
              <Select value={assessForm.assessment_type} onValueChange={v => setAssessForm(f => ({ ...f, assessment_type: v }))}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {Object.entries(ASSESSMENT_TYPE_LABELS).map(([k, v]) => <SelectItem key={k} value={k}>{v}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Likelihood *</Label>
                <Select value={assessForm.likelihood} onValueChange={v => setAssessForm(f => ({ ...f, likelihood: v }))}>
                  <SelectTrigger><SelectValue placeholder="Select" /></SelectTrigger>
                  <SelectContent>
                    {LIKELIHOOD_ORDER.map(l => (
                      <SelectItem key={l} value={l}>{LIKELIHOOD_LABELS[l]} ({LIKELIHOOD_ORDER.indexOf(l) + 1})</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Impact *</Label>
                <Select value={assessForm.impact} onValueChange={v => setAssessForm(f => ({ ...f, impact: v }))}>
                  <SelectTrigger><SelectValue placeholder="Select" /></SelectTrigger>
                  <SelectContent>
                    {IMPACT_ORDER.map(i => (
                      <SelectItem key={i} value={i}>{IMPACT_LABELS[i]} ({IMPACT_ORDER.indexOf(i) + 1})</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            {previewScore !== null && (
              <div className="p-3 rounded-md bg-muted text-center">
                <span className="text-sm text-muted-foreground">Score: </span>
                <span className="text-lg font-bold">{previewScore}</span>
                <span className={`ml-2 inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[scoreToSeverity(previewScore)]}`}>
                  {SEVERITY_LABELS[scoreToSeverity(previewScore)]}
                </span>
              </div>
            )}
            <div>
              <Label>Justification</Label>
              <Textarea value={assessForm.justification} onChange={e => setAssessForm(f => ({ ...f, justification: e.target.value }))} rows={3} placeholder="Why these levels were chosen..." />
            </div>
            <div>
              <Label>Assumptions</Label>
              <Input value={assessForm.assumptions} onChange={e => setAssessForm(f => ({ ...f, assumptions: e.target.value }))} placeholder="Key assumptions..." />
            </div>
            <div>
              <Label>Valid Until</Label>
              <Input type="date" value={assessForm.valid_until} onChange={e => setAssessForm(f => ({ ...f, valid_until: e.target.value }))} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAssessment(false)}>Cancel</Button>
            <Button onClick={handleCreateAssessment} disabled={assessSaving || !assessForm.likelihood || !assessForm.impact}>
              {assessSaving ? 'Saving...' : 'Create Assessment'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Treatment Dialog */}
      <Dialog open={showTreatment} onOpenChange={setShowTreatment}>
        <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>New Treatment Plan</DialogTitle>
            <DialogDescription>Define how to address this risk</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Treatment Type *</Label>
                <Select value={treatForm.treatment_type} onValueChange={v => setTreatForm(f => ({ ...f, treatment_type: v }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {Object.entries(TREATMENT_TYPE_LABELS).map(([k, v]) => <SelectItem key={k} value={k}>{v}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Priority</Label>
                <Select value={treatForm.priority} onValueChange={v => setTreatForm(f => ({ ...f, priority: v }))}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {Object.entries(PRIORITY_LABELS).map(([k, v]) => <SelectItem key={k} value={k}>{v}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div>
              <Label>Title *</Label>
              <Input value={treatForm.title} onChange={e => setTreatForm(f => ({ ...f, title: e.target.value }))} placeholder="Treatment plan title" />
            </div>
            <div>
              <Label>Description</Label>
              <Textarea value={treatForm.description} onChange={e => setTreatForm(f => ({ ...f, description: e.target.value }))} rows={3} placeholder="Detailed plan..." />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Due Date</Label>
                <Input type="date" value={treatForm.due_date} onChange={e => setTreatForm(f => ({ ...f, due_date: e.target.value }))} />
              </div>
              <div>
                <Label>Estimated Hours</Label>
                <Input type="number" value={treatForm.estimated_effort_hours} onChange={e => setTreatForm(f => ({ ...f, estimated_effort_hours: e.target.value }))} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Expected Residual Likelihood</Label>
                <Select value={treatForm.expected_residual_likelihood} onValueChange={v => setTreatForm(f => ({ ...f, expected_residual_likelihood: v }))}>
                  <SelectTrigger><SelectValue placeholder="Optional" /></SelectTrigger>
                  <SelectContent>
                    {LIKELIHOOD_ORDER.map(l => <SelectItem key={l} value={l}>{LIKELIHOOD_LABELS[l]}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Expected Residual Impact</Label>
                <Select value={treatForm.expected_residual_impact} onValueChange={v => setTreatForm(f => ({ ...f, expected_residual_impact: v }))}>
                  <SelectTrigger><SelectValue placeholder="Optional" /></SelectTrigger>
                  <SelectContent>
                    {IMPACT_ORDER.map(i => <SelectItem key={i} value={i}>{IMPACT_LABELS[i]}</SelectItem>)}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div>
              <Label>Notes</Label>
              <Textarea value={treatForm.notes} onChange={e => setTreatForm(f => ({ ...f, notes: e.target.value }))} rows={2} placeholder="Additional context..." />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowTreatment(false)}>Cancel</Button>
            <Button onClick={handleCreateTreatment} disabled={treatSaving || !treatForm.title}>
              {treatSaving ? 'Saving...' : 'Create Treatment'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Link Control Dialog */}
      <Dialog open={showLinkControl} onOpenChange={setShowLinkControl}>
        <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Link Control to Risk</DialogTitle>
            <DialogDescription>Search and select a control to link</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Search Controls</Label>
              <Input placeholder="Search by identifier or title..." value={controlSearch} onChange={e => setControlSearch(e.target.value)} />
            </div>
            <div className="max-h-48 overflow-y-auto border rounded-md">
              {availableControls
                .filter(c => {
                  const q = controlSearch.toLowerCase();
                  return !q || c.identifier.toLowerCase().includes(q) || c.title.toLowerCase().includes(q);
                })
                .filter(c => !controls.some(linked => linked.id === c.id))
                .map(c => (
                  <div
                    key={c.id}
                    onClick={() => setSelectedControlId(c.id)}
                    className={`p-2 cursor-pointer text-sm hover:bg-accent ${selectedControlId === c.id ? 'bg-primary/10' : ''}`}
                  >
                    <span className="font-mono text-xs">{c.identifier}</span> — {c.title}
                  </div>
                ))}
            </div>
            <div>
              <Label>Effectiveness</Label>
              <Select value={linkEffectiveness} onValueChange={setLinkEffectiveness}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {Object.entries(EFFECTIVENESS_LABELS).map(([k, v]) => <SelectItem key={k} value={k}>{v}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Mitigation Percentage (0-100)</Label>
              <Input type="number" min={0} max={100} value={linkMitigation} onChange={e => setLinkMitigation(e.target.value)} placeholder="25" />
            </div>
            <div>
              <Label>Notes</Label>
              <Textarea value={linkNotes} onChange={e => setLinkNotes(e.target.value)} rows={2} placeholder="Why this control mitigates this risk..." />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowLinkControl(false)}>Cancel</Button>
            <Button onClick={handleLinkControl} disabled={linkSaving || !selectedControlId}>
              {linkSaving ? 'Linking...' : 'Link Control'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Complete Treatment Dialog */}
      <Dialog open={!!showComplete} onOpenChange={() => setShowComplete(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Complete Treatment</DialogTitle>
            <DialogDescription>Mark treatment as complete and optionally record effectiveness</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label>Actual Effort (hours)</Label>
              <Input type="number" value={completeForm.actual_effort_hours} onChange={e => setCompleteForm(f => ({ ...f, actual_effort_hours: e.target.value }))} />
            </div>
            <div>
              <Label>Effectiveness Rating</Label>
              <Select value={completeForm.effectiveness_rating} onValueChange={v => setCompleteForm(f => ({ ...f, effectiveness_rating: v }))}>
                <SelectTrigger><SelectValue placeholder="Optional" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="highly_effective">Highly Effective</SelectItem>
                  <SelectItem value="effective">Effective</SelectItem>
                  <SelectItem value="partially_effective">Partially Effective</SelectItem>
                  <SelectItem value="ineffective">Ineffective</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Effectiveness Notes</Label>
              <Textarea value={completeForm.effectiveness_notes} onChange={e => setCompleteForm(f => ({ ...f, effectiveness_notes: e.target.value }))} rows={3} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowComplete(null)}>Cancel</Button>
            <Button onClick={handleCompleteTreatment} disabled={completeSaving}>
              {completeSaving ? 'Saving...' : 'Mark Complete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
