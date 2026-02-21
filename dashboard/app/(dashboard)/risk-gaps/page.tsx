'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import {
  ArrowLeft, AlertTriangle, ShieldOff, Clock, FileWarning, ShieldAlert,
  ChevronLeft, ChevronRight,
} from 'lucide-react';
import { RiskGapData, getRiskGaps } from '@/lib/api';
import {
  SEVERITY_LABELS, SEVERITY_COLORS,
  RISK_CATEGORY_LABELS, RISK_STATUS_LABELS,
} from '@/components/risk/constants';

const GAP_TYPE_LABELS: Record<string, string> = {
  no_treatments: 'No Treatments',
  no_controls: 'No Controls',
  high_without_controls: 'High Risk, No Controls',
  overdue_assessment: 'Overdue Assessment',
  expired_acceptance: 'Expired Acceptance',
};

const GAP_TYPE_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  no_treatments: FileWarning,
  no_controls: ShieldOff,
  high_without_controls: ShieldAlert,
  overdue_assessment: Clock,
  expired_acceptance: AlertTriangle,
};

export default function RiskGapDashboardPage() {
  const [data, setData] = useState<RiskGapData | null>(null);
  const [loading, setLoading] = useState(true);
  const [gapTypeFilter, setGapTypeFilter] = useState('');
  const [severityFilter, setSeverityFilter] = useState('');
  const [page, setPage] = useState(1);
  const perPage = 20;

  const fetchGaps = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
      };
      if (gapTypeFilter) params.gap_type = gapTypeFilter;
      if (severityFilter) params.min_severity = severityFilter;
      const res = await getRiskGaps(params);
      setData(res.data);
    } catch (err) {
      console.error('Failed to fetch gaps:', err);
    } finally {
      setLoading(false);
    }
  }, [page, gapTypeFilter, severityFilter]);

  useEffect(() => { fetchGaps(); }, [fetchGaps]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link href="/risks">
          <Button variant="ghost" size="icon"><ArrowLeft className="h-4 w-4" /></Button>
        </Link>
        <div>
          <h1 className="text-2xl font-bold">Risk Gap Dashboard</h1>
          <p className="text-sm text-muted-foreground">Risks missing treatments, controls, or with overdue assessments</p>
        </div>
      </div>

      {/* Summary Cards */}
      {data && (
        <div className="grid gap-4 md:grid-cols-6">
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{data.summary.total_active_risks}</div>
              <p className="text-xs text-muted-foreground">Active Risks</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-2">
                <FileWarning className="h-4 w-4 text-orange-500" />
                <div className="text-2xl font-bold text-orange-600">{data.summary.risks_without_treatments}</div>
              </div>
              <p className="text-xs text-muted-foreground">No Treatments</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-2">
                <ShieldOff className="h-4 w-4 text-red-500" />
                <div className="text-2xl font-bold text-red-600">{data.summary.risks_without_controls}</div>
              </div>
              <p className="text-xs text-muted-foreground">No Controls</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-2">
                <ShieldAlert className="h-4 w-4 text-red-600" />
                <div className="text-2xl font-bold text-red-700">{data.summary.high_risks_without_controls}</div>
              </div>
              <p className="text-xs text-muted-foreground">High+ Without Controls</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4 text-yellow-600" />
                <div className="text-2xl font-bold text-yellow-600">{data.summary.overdue_assessments}</div>
              </div>
              <p className="text-xs text-muted-foreground">Overdue Assessments</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center gap-2">
                <AlertTriangle className="h-4 w-4 text-purple-500" />
                <div className="text-2xl font-bold text-purple-600">{data.summary.expired_acceptances}</div>
              </div>
              <p className="text-xs text-muted-foreground">Expired Acceptances</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-4">
            <div className="w-[200px]">
              <Label className="text-xs">Gap Type</Label>
              <Select value={gapTypeFilter} onValueChange={v => { setGapTypeFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All gap types" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All gap types</SelectItem>
                  {Object.entries(GAP_TYPE_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[160px]">
              <Label className="text-xs">Min Severity</Label>
              <Select value={severityFilter} onValueChange={v => { setSeverityFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="Any" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Any severity</SelectItem>
                  {Object.entries(SEVERITY_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Gaps Table */}
      <Card>
        <CardContent className="p-0">
          {loading ? (
            <div className="text-center py-12 text-muted-foreground">Loading gaps...</div>
          ) : !data || data.gaps.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <ShieldAlert className="h-12 w-12 mx-auto mb-3 text-green-500" />
              <p className="text-lg font-medium">No gaps found</p>
              <p className="text-sm">All risks have appropriate treatments and controls</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Risk</TableHead>
                  <TableHead>Category</TableHead>
                  <TableHead>Score</TableHead>
                  <TableHead>Gap Types</TableHead>
                  <TableHead>Days Open</TableHead>
                  <TableHead>Owner</TableHead>
                  <TableHead>Recommendation</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.gaps.map(gap => (
                  <TableRow key={gap.risk.id}>
                    <TableCell>
                      <Link href={`/risks/${gap.risk.id}`} className="text-primary hover:underline font-medium text-sm">
                        <span className="font-mono text-xs">{gap.risk.identifier}</span>
                        <br />
                        {gap.risk.title}
                      </Link>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">{RISK_CATEGORY_LABELS[gap.risk.category] || gap.risk.category}</Badge>
                    </TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[gap.risk.severity]}`}>
                        {gap.risk.residual_score} — {SEVERITY_LABELS[gap.risk.severity]}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {gap.gap_types.map(gt => (
                          <Badge key={gt} variant="destructive" className="text-[10px]">
                            {GAP_TYPE_LABELS[gt] || gt}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="text-sm">{gap.days_open}d</TableCell>
                    <TableCell className="text-sm">{gap.risk.owner?.name || '—'}</TableCell>
                    <TableCell className="text-xs text-muted-foreground max-w-xs truncate" title={gap.recommendation}>
                      {gap.recommendation}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
