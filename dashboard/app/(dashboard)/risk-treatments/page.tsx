'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import {
  ArrowLeft, Clock, CheckCircle2, XCircle, Loader2, AlertTriangle,
  TrendingDown, Calendar, Timer,
} from 'lucide-react';
import { RiskStats, Risk, getRiskStats, listRisks } from '@/lib/api';
import {
  TREATMENT_STATUS_LABELS, TREATMENT_STATUS_COLORS,
  TREATMENT_TYPE_LABELS,
  RISK_STATUS_LABELS, RISK_STATUS_COLORS,
  SEVERITY_LABELS, SEVERITY_COLORS,
  PRIORITY_LABELS, PRIORITY_COLORS,
  RISK_CATEGORY_LABELS,
} from '@/components/risk/constants';

interface TreatmentOverview {
  risk: Risk;
}

export default function TreatmentProgressPage() {
  const [stats, setStats] = useState<RiskStats | null>(null);
  const [risks, setRisks] = useState<Risk[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState('treating');

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {
        per_page: '50',
        has_treatments: 'true',
      };
      if (statusFilter) params.status = statusFilter;
      const [statsRes, risksRes] = await Promise.all([
        getRiskStats(),
        listRisks(params),
      ]);
      setStats(statsRes.data);
      setRisks(risksRes.data);
    } catch (err) {
      console.error('Failed to fetch treatment data:', err);
    } finally {
      setLoading(false);
    }
  }, [statusFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const treatmentTotal = stats?.treatment_summary;
  const totalActive = treatmentTotal
    ? treatmentTotal.planned + treatmentTotal.in_progress + treatmentTotal.implemented
    : 0;
  const totalComplete = treatmentTotal
    ? treatmentTotal.verified + treatmentTotal.ineffective
    : 0;
  const totalAll = treatmentTotal ? treatmentTotal.total_treatments : 0;
  const completionPct = totalAll > 0 ? Math.round(((totalComplete) / totalAll) * 100) : 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/risks">
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div>
          <h1 className="text-2xl font-bold">Treatment Progress</h1>
          <p className="text-sm text-muted-foreground">Track risk treatment plans, timelines, and completion</p>
        </div>
      </div>

      {/* Summary Cards */}
      {treatmentTotal && (
        <>
          <div className="grid gap-4 md:grid-cols-6">
            <Card>
              <CardContent className="p-4">
                <div className="text-2xl font-bold">{treatmentTotal.total_treatments}</div>
                <p className="text-xs text-muted-foreground">Total Treatments</p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-blue-500" />
                  <div className="text-2xl font-bold text-blue-600">{treatmentTotal.planned}</div>
                </div>
                <p className="text-xs text-muted-foreground">Planned</p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <Loader2 className="h-4 w-4 text-yellow-500" />
                  <div className="text-2xl font-bold text-yellow-600">{treatmentTotal.in_progress}</div>
                </div>
                <p className="text-xs text-muted-foreground">In Progress</p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-green-500" />
                  <div className="text-2xl font-bold text-green-600">{treatmentTotal.verified}</div>
                </div>
                <p className="text-xs text-muted-foreground">Verified</p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <XCircle className="h-4 w-4 text-red-500" />
                  <div className="text-2xl font-bold text-red-600">{treatmentTotal.ineffective}</div>
                </div>
                <p className="text-xs text-muted-foreground">Ineffective</p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-orange-500" />
                  <div className="text-2xl font-bold text-orange-600">{treatmentTotal.overdue}</div>
                </div>
                <p className="text-xs text-muted-foreground">Overdue</p>
              </CardContent>
            </Card>
          </div>

          {/* Overall Progress */}
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium">Overall Treatment Completion</span>
                <span className="text-sm text-muted-foreground">{completionPct}% ({totalComplete} of {totalAll})</span>
              </div>
              <Progress value={completionPct} className="h-3" />
            </CardContent>
          </Card>
        </>
      )}

      {/* Filter */}
      <div className="flex gap-4">
        <div className="w-[180px]">
          <Label className="text-xs">Risk Status</Label>
          <Select value={statusFilter} onValueChange={v => setStatusFilter(v === 'all' ? '' : v)}>
            <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              {Object.entries(RISK_STATUS_LABELS).map(([k, v]) => (
                <SelectItem key={k} value={k}>{v}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Risk Treatment Cards */}
      {loading ? (
        <div className="text-center py-12 text-muted-foreground">Loading treatments...</div>
      ) : risks.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <TrendingDown className="h-12 w-12 mx-auto mb-3 text-muted-foreground/30" />
          <p>No risks with treatments found</p>
        </div>
      ) : (
        <div className="space-y-4">
          {risks.map(risk => {
            const summary = risk.treatment_summary;
            const total = summary?.total || 0;
            const complete = (summary?.verified || 0) + (summary?.cancelled || 0);
            const pct = total > 0 ? Math.round((complete / total) * 100) : 0;

            return (
              <Card key={risk.id}>
                <CardContent className="p-4">
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <Link href={`/risks/${risk.id}`} className="text-primary hover:underline font-medium">
                        <span className="font-mono text-xs">{risk.identifier}</span> — {risk.title}
                      </Link>
                      <div className="flex items-center gap-2 mt-1">
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${RISK_STATUS_COLORS[risk.status]}`}>
                          {RISK_STATUS_LABELS[risk.status]}
                        </span>
                        {risk.residual_score && (
                          <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[risk.residual_score.severity]}`}>
                            Score: {risk.residual_score.score}
                          </span>
                        )}
                        <Badge variant="outline" className="text-xs">{RISK_CATEGORY_LABELS[risk.category] || risk.category}</Badge>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-sm font-medium">{pct}% complete</div>
                      <div className="text-xs text-muted-foreground">{total} treatment{total !== 1 ? 's' : ''}</div>
                    </div>
                  </div>

                  <Progress value={pct} className="h-2 mb-3" />

                  {/* Treatment status breakdown */}
                  {summary && (
                    <div className="flex flex-wrap gap-3 text-xs">
                      {summary.planned > 0 && (
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 font-medium ${TREATMENT_STATUS_COLORS.planned}`}>
                          {summary.planned} Planned
                        </span>
                      )}
                      {summary.in_progress > 0 && (
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 font-medium ${TREATMENT_STATUS_COLORS.in_progress}`}>
                          {summary.in_progress} In Progress
                        </span>
                      )}
                      {summary.implemented > 0 && (
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 font-medium ${TREATMENT_STATUS_COLORS.implemented}`}>
                          {summary.implemented} Implemented
                        </span>
                      )}
                      {summary.verified > 0 && (
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 font-medium ${TREATMENT_STATUS_COLORS.verified}`}>
                          {summary.verified} Verified
                        </span>
                      )}
                      {summary.cancelled > 0 && (
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 font-medium ${TREATMENT_STATUS_COLORS.cancelled}`}>
                          {summary.cancelled} Cancelled
                        </span>
                      )}
                    </div>
                  )}

                  {risk.owner && (
                    <div className="text-xs text-muted-foreground mt-2">Owner: {risk.owner.name}</div>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
