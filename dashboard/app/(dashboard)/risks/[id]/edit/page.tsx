'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { ArrowLeft, Save } from 'lucide-react';
import { Risk, getRisk, createRisk, updateRisk } from '@/lib/api';
import {
  RISK_CATEGORY_LABELS,
  LIKELIHOOD_LABELS, LIKELIHOOD_ORDER,
  IMPACT_LABELS, IMPACT_ORDER,
  SEVERITY_LABELS, SEVERITY_COLORS, scoreToSeverity,
} from '@/components/risk/constants';

export default function RiskEditorPage() {
  const params = useParams();
  const router = useRouter();
  const { hasRole } = useAuth();
  const riskId = params.id as string;
  const isNew = riskId === 'new';

  const canEdit = hasRole('ciso', 'compliance_manager', 'security_engineer');

  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({
    identifier: '',
    title: '',
    description: '',
    category: '' as string,
    risk_appetite_threshold: '',
    assessment_frequency_days: '90',
    source: '',
    affected_assets: '',
    tags: '',
    // Initial assessment (create only)
    inherent_likelihood: '' as string,
    inherent_impact: '' as string,
    residual_likelihood: '' as string,
    residual_impact: '' as string,
    assessment_justification: '',
  });

  useEffect(() => {
    if (!isNew) {
      getRisk(riskId).then(res => {
        const r = res.data;
        setForm({
          identifier: r.identifier,
          title: r.title,
          description: r.description || '',
          category: r.category,
          risk_appetite_threshold: r.risk_appetite_threshold != null ? String(r.risk_appetite_threshold) : '',
          assessment_frequency_days: r.assessment_frequency_days != null ? String(r.assessment_frequency_days) : '90',
          source: r.source || '',
          affected_assets: r.affected_assets?.join(', ') || '',
          tags: r.tags?.join(', ') || '',
          inherent_likelihood: '',
          inherent_impact: '',
          residual_likelihood: '',
          residual_impact: '',
          assessment_justification: '',
        });
        setLoading(false);
      }).catch(() => {
        setLoading(false);
      });
    }
  }, [isNew, riskId]);

  const inherentScore = form.inherent_likelihood && form.inherent_impact
    ? (LIKELIHOOD_ORDER.indexOf(form.inherent_likelihood as any) + 1) * (IMPACT_ORDER.indexOf(form.inherent_impact as any) + 1)
    : null;

  const residualScore = form.residual_likelihood && form.residual_impact
    ? (LIKELIHOOD_ORDER.indexOf(form.residual_likelihood as any) + 1) * (IMPACT_ORDER.indexOf(form.residual_impact as any) + 1)
    : null;

  const handleSave = async () => {
    if (!form.identifier || !form.title || !form.category) return;
    try {
      setSaving(true);
      if (isNew) {
        const body: Parameters<typeof createRisk>[0] = {
          identifier: form.identifier,
          title: form.title,
          description: form.description || undefined,
          category: form.category,
          risk_appetite_threshold: form.risk_appetite_threshold ? Number(form.risk_appetite_threshold) : undefined,
          assessment_frequency_days: form.assessment_frequency_days ? Number(form.assessment_frequency_days) : undefined,
          source: form.source || undefined,
          affected_assets: form.affected_assets ? form.affected_assets.split(',').map(s => s.trim()).filter(Boolean) : undefined,
          tags: form.tags ? form.tags.split(',').map(s => s.trim()).filter(Boolean) : undefined,
        };
        if (form.inherent_likelihood && form.inherent_impact) {
          body.initial_assessment = {
            inherent_likelihood: form.inherent_likelihood,
            inherent_impact: form.inherent_impact,
            residual_likelihood: form.residual_likelihood || undefined,
            residual_impact: form.residual_impact || undefined,
            justification: form.assessment_justification || undefined,
          };
        }
        const res = await createRisk(body);
        router.push(`/risks/${res.data.id}`);
      } else {
        await updateRisk(riskId, {
          title: form.title,
          description: form.description || undefined,
          category: form.category,
          risk_appetite_threshold: form.risk_appetite_threshold ? Number(form.risk_appetite_threshold) : undefined,
          assessment_frequency_days: form.assessment_frequency_days ? Number(form.assessment_frequency_days) : undefined,
          source: form.source || undefined,
          affected_assets: form.affected_assets ? form.affected_assets.split(',').map(s => s.trim()).filter(Boolean) : undefined,
          tags: form.tags ? form.tags.split(',').map(s => s.trim()).filter(Boolean) : undefined,
        });
        router.push(`/risks/${riskId}`);
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to save risk');
    } finally {
      setSaving(false);
    }
  };

  if (!canEdit) {
    return <div className="flex items-center justify-center h-64 text-muted-foreground">You don&apos;t have permission to edit risks.</div>;
  }

  if (loading) {
    return <div className="flex items-center justify-center h-64 text-muted-foreground">Loading...</div>;
  }

  return (
    <div className="space-y-6 max-w-3xl">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => router.back()}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-2xl font-bold">{isNew ? 'Create New Risk' : 'Edit Risk'}</h1>
          <p className="text-sm text-muted-foreground">{isNew ? 'Register a new risk in the inventory' : `Editing ${form.identifier}`}</p>
        </div>
      </div>

      {/* Basic Info */}
      <Card>
        <CardHeader>
          <CardTitle>Risk Information</CardTitle>
          <CardDescription>Core risk identification and metadata</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>Identifier *</Label>
              <Input
                value={form.identifier}
                onChange={e => setForm(f => ({ ...f, identifier: e.target.value }))}
                placeholder="RISK-CY-001"
                disabled={!isNew}
              />
            </div>
            <div>
              <Label>Category *</Label>
              <Select value={form.category} onValueChange={v => setForm(f => ({ ...f, category: v }))}>
                <SelectTrigger><SelectValue placeholder="Select category" /></SelectTrigger>
                <SelectContent>
                  {Object.entries(RISK_CATEGORY_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div>
            <Label>Title *</Label>
            <Input value={form.title} onChange={e => setForm(f => ({ ...f, title: e.target.value }))} placeholder="Risk title" />
          </div>
          <div>
            <Label>Description</Label>
            <Textarea value={form.description} onChange={e => setForm(f => ({ ...f, description: e.target.value }))} rows={4} placeholder="Detailed risk description..." />
          </div>
          <div className="grid grid-cols-3 gap-4">
            <div>
              <Label>Risk Appetite Threshold (1-25)</Label>
              <Input type="number" min={1} max={25} value={form.risk_appetite_threshold} onChange={e => setForm(f => ({ ...f, risk_appetite_threshold: e.target.value }))} placeholder="10" />
            </div>
            <div>
              <Label>Assessment Frequency (days)</Label>
              <Input type="number" value={form.assessment_frequency_days} onChange={e => setForm(f => ({ ...f, assessment_frequency_days: e.target.value }))} placeholder="90" />
            </div>
            <div>
              <Label>Source</Label>
              <Input value={form.source} onChange={e => setForm(f => ({ ...f, source: e.target.value }))} placeholder="threat_assessment" />
            </div>
          </div>
          <div>
            <Label>Affected Assets (comma-separated)</Label>
            <Input value={form.affected_assets} onChange={e => setForm(f => ({ ...f, affected_assets: e.target.value }))} placeholder="payment-api, customer-db" />
          </div>
          <div>
            <Label>Tags (comma-separated)</Label>
            <Input value={form.tags} onChange={e => setForm(f => ({ ...f, tags: e.target.value }))} placeholder="critical, q1-review" />
          </div>
        </CardContent>
      </Card>

      {/* Initial Assessment (create only) */}
      {isNew && (
        <Card>
          <CardHeader>
            <CardTitle>Initial Assessment (Optional)</CardTitle>
            <CardDescription>Score the risk at creation time</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Inherent Likelihood</Label>
                <Select value={form.inherent_likelihood} onValueChange={v => setForm(f => ({ ...f, inherent_likelihood: v }))}>
                  <SelectTrigger><SelectValue placeholder="Select" /></SelectTrigger>
                  <SelectContent>
                    {LIKELIHOOD_ORDER.map(l => (
                      <SelectItem key={l} value={l}>{LIKELIHOOD_LABELS[l]} ({LIKELIHOOD_ORDER.indexOf(l) + 1})</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Inherent Impact</Label>
                <Select value={form.inherent_impact} onValueChange={v => setForm(f => ({ ...f, inherent_impact: v }))}>
                  <SelectTrigger><SelectValue placeholder="Select" /></SelectTrigger>
                  <SelectContent>
                    {IMPACT_ORDER.map(i => (
                      <SelectItem key={i} value={i}>{IMPACT_LABELS[i]} ({IMPACT_ORDER.indexOf(i) + 1})</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            {inherentScore !== null && (
              <div className="p-3 rounded-md bg-muted">
                <span className="text-sm text-muted-foreground">Inherent Score: </span>
                <span className="text-lg font-bold">{inherentScore}</span>
                <span className={`ml-2 inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[scoreToSeverity(inherentScore)]}`}>
                  {SEVERITY_LABELS[scoreToSeverity(inherentScore)]}
                </span>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>Residual Likelihood</Label>
                <Select value={form.residual_likelihood} onValueChange={v => setForm(f => ({ ...f, residual_likelihood: v }))}>
                  <SelectTrigger><SelectValue placeholder="Optional" /></SelectTrigger>
                  <SelectContent>
                    {LIKELIHOOD_ORDER.map(l => (
                      <SelectItem key={l} value={l}>{LIKELIHOOD_LABELS[l]} ({LIKELIHOOD_ORDER.indexOf(l) + 1})</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Residual Impact</Label>
                <Select value={form.residual_impact} onValueChange={v => setForm(f => ({ ...f, residual_impact: v }))}>
                  <SelectTrigger><SelectValue placeholder="Optional" /></SelectTrigger>
                  <SelectContent>
                    {IMPACT_ORDER.map(i => (
                      <SelectItem key={i} value={i}>{IMPACT_LABELS[i]} ({IMPACT_ORDER.indexOf(i) + 1})</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            {residualScore !== null && (
              <div className="p-3 rounded-md bg-muted">
                <span className="text-sm text-muted-foreground">Residual Score: </span>
                <span className="text-lg font-bold">{residualScore}</span>
                <span className={`ml-2 inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[scoreToSeverity(residualScore)]}`}>
                  {SEVERITY_LABELS[scoreToSeverity(residualScore)]}
                </span>
              </div>
            )}

            <div>
              <Label>Justification</Label>
              <Textarea value={form.assessment_justification} onChange={e => setForm(f => ({ ...f, assessment_justification: e.target.value }))} rows={3} placeholder="Why these scoring levels were chosen..." />
            </div>
          </CardContent>
        </Card>
      )}

      {/* Save */}
      <div className="flex justify-end gap-3">
        <Button variant="outline" onClick={() => router.back()}>Cancel</Button>
        <Button onClick={handleSave} disabled={saving || !form.identifier || !form.title || !form.category}>
          <Save className="h-4 w-4 mr-2" />
          {saving ? 'Saving...' : isNew ? 'Create Risk' : 'Save Changes'}
        </Button>
      </div>
    </div>
  );
}
