'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useAuth } from '@/lib/auth-context';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import {
  Search, Plus, ChevronLeft, ChevronRight, Eye, Archive, AlertTriangle,
  ShieldAlert, TrendingDown, BarChart3,
} from 'lucide-react';
import {
  Risk, RiskStats, HeatMapData, RiskSeverity, LikelihoodLevel, ImpactLevel,
  listRisks, getRiskStats, getRiskHeatMap, archiveRisk,
} from '@/lib/api';
import {
  RISK_STATUS_LABELS, RISK_STATUS_COLORS,
  RISK_CATEGORY_LABELS,
  SEVERITY_LABELS, SEVERITY_COLORS,
  LIKELIHOOD_LABELS, LIKELIHOOD_ORDER,
  IMPACT_LABELS, IMPACT_ORDER,
  heatMapCellColor, scoreToSeverity,
} from '@/components/risk/constants';

export default function RiskRegisterPage() {
  const { hasRole } = useAuth();
  const canCreate = hasRole('ciso', 'compliance_manager', 'security_engineer');
  const canArchive = hasRole('ciso', 'compliance_manager');

  const [risks, setRisks] = useState<Risk[]>([]);
  const [stats, setStats] = useState<RiskStats | null>(null);
  const [heatMap, setHeatMap] = useState<HeatMapData | null>(null);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const perPage = 20;

  // Filters
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [severityFilter, setSeverityFilter] = useState('');

  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const params: Record<string, string> = {
        page: String(page),
        per_page: String(perPage),
      };
      if (search) params.search = search;
      if (statusFilter) params.status = statusFilter;
      if (categoryFilter) params.category = categoryFilter;
      if (severityFilter) params.severity = severityFilter;

      const [risksRes, statsRes, heatMapRes] = await Promise.all([
        listRisks(params),
        getRiskStats(),
        getRiskHeatMap({ score_type: 'residual' }),
      ]);
      setRisks(risksRes.data);
      setTotal(risksRes.meta?.total || risksRes.data.length);
      setStats(statsRes.data);
      setHeatMap(heatMapRes.data);
    } catch (err) {
      console.error('Failed to fetch risks:', err);
    } finally {
      setLoading(false);
    }
  }, [page, search, statusFilter, categoryFilter, severityFilter]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleArchive = async (id: string) => {
    if (!confirm('Archive this risk? Active treatments will be cancelled.')) return;
    try {
      await archiveRisk(id);
      fetchData();
    } catch (err) {
      console.error('Archive failed:', err);
      alert(err instanceof Error ? err.message : 'Failed to archive');
    }
  };

  const totalPages = Math.ceil(total / perPage);

  // Build heat map grid lookup
  const heatMapLookup = new Map<string, { count: number; severity: RiskSeverity; risks: { id: string; identifier: string; title: string }[] }>();
  if (heatMap) {
    for (const cell of heatMap.grid) {
      const key = `${cell.likelihood_score}-${cell.impact_score}`;
      heatMapLookup.set(key, { count: cell.count, severity: cell.severity, risks: cell.risks });
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Risk Register</h1>
          <p className="text-sm text-muted-foreground">Organizational risk inventory with scoring, treatments, and controls</p>
        </div>
        <div className="flex items-center gap-2">
          <Link href="/risk-heatmap">
            <Button variant="outline" size="sm">
              <BarChart3 className="h-4 w-4 mr-2" /> Heat Map
            </Button>
          </Link>
          <Link href="/risk-gaps">
            <Button variant="outline" size="sm">
              <AlertTriangle className="h-4 w-4 mr-2" /> Gaps
            </Button>
          </Link>
          <Link href="/risk-treatments">
            <Button variant="outline" size="sm">
              <TrendingDown className="h-4 w-4 mr-2" /> Treatments
            </Button>
          </Link>
          {canCreate && (
            <Link href="/risks/new/edit">
              <Button>
                <Plus className="h-4 w-4 mr-2" /> New Risk
              </Button>
            </Link>
          )}
        </div>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid gap-4 md:grid-cols-5">
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{stats.total_risks}</div>
              <p className="text-xs text-muted-foreground">Total Risks</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-red-600">{stats.by_severity?.critical || 0}</div>
              <p className="text-xs text-muted-foreground">Critical</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-orange-600">{stats.by_severity?.high || 0}</div>
              <p className="text-xs text-muted-foreground">High</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold text-orange-500">{stats.appetite_summary?.breaching_appetite || 0}</div>
              <p className="text-xs text-muted-foreground">Appetite Breaches</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-2xl font-bold">{stats.scoring_summary?.average_risk_reduction?.toFixed(0) || 0}%</div>
              <p className="text-xs text-muted-foreground">Avg Risk Reduction</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Mini Heat Map Preview */}
      {heatMap && (
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-semibold">Residual Risk Heat Map</h3>
              <Link href="/risk-heatmap" className="text-xs text-primary hover:underline">View Full →</Link>
            </div>
            <TooltipProvider>
              <div className="grid grid-cols-[auto_repeat(5,1fr)] gap-1 max-w-md">
                {/* Header row */}
                <div />
                {IMPACT_ORDER.map(imp => (
                  <div key={imp} className="text-[10px] text-center text-muted-foreground truncate px-0.5">
                    {IMPACT_LABELS[imp].slice(0, 3)}
                  </div>
                ))}
                {/* Grid rows — likelihood top to bottom (almost_certain at top) */}
                {[...LIKELIHOOD_ORDER].reverse().map(lik => (
                  <>
                    <div key={`label-${lik}`} className="text-[10px] text-muted-foreground flex items-center pr-1 truncate">
                      {LIKELIHOOD_LABELS[lik].slice(0, 4)}
                    </div>
                    {IMPACT_ORDER.map(imp => {
                      const lScore = LIKELIHOOD_ORDER.indexOf(lik) + 1;
                      const iScore = IMPACT_ORDER.indexOf(imp) + 1;
                      const cellScore = lScore * iScore;
                      const cellKey = `${lScore}-${iScore}`;
                      const cellData = heatMapLookup.get(cellKey);
                      const count = cellData?.count || 0;
                      return (
                        <Tooltip key={`${lik}-${imp}`}>
                          <TooltipTrigger asChild>
                            <div
                              className={`h-8 rounded-sm flex items-center justify-center text-xs font-medium cursor-default ${heatMapCellColor(cellScore)} ${count > 0 ? 'text-white dark:text-white font-bold' : 'text-muted-foreground/50'}`}
                            >
                              {count > 0 ? count : ''}
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p className="font-medium">{LIKELIHOOD_LABELS[lik]} × {IMPACT_LABELS[imp]} = {cellScore}</p>
                            <p className="text-xs">{count} risk{count !== 1 ? 's' : ''}</p>
                            {cellData?.risks?.slice(0, 3).map(r => (
                              <p key={r.id} className="text-xs text-muted-foreground">{r.identifier}: {r.title}</p>
                            ))}
                          </TooltipContent>
                        </Tooltip>
                      );
                    })}
                  </>
                ))}
              </div>
            </TooltipProvider>
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
                  placeholder="Search risks..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') { setSearch(searchInput); setPage(1); } }}
                  className="pl-8"
                />
              </div>
            </div>
            <div className="w-[150px]">
              <Label className="text-xs">Status</Label>
              <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All statuses</SelectItem>
                  {Object.entries(RISK_STATUS_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[170px]">
              <Label className="text-xs">Category</Label>
              <Select value={categoryFilter} onValueChange={(v) => { setCategoryFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All categories</SelectItem>
                  {Object.entries(RISK_CATEGORY_LABELS).map(([k, v]) => (
                    <SelectItem key={k} value={k}>{v}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-[140px]">
              <Label className="text-xs">Severity</Label>
              <Select value={severityFilter} onValueChange={(v) => { setSeverityFilter(v === 'all' ? '' : v); setPage(1); }}>
                <SelectTrigger><SelectValue placeholder="All" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  {Object.entries(SEVERITY_LABELS).map(([k, v]) => (
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
                <TableHead>Identifier</TableHead>
                <TableHead>Title</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Inherent</TableHead>
                <TableHead>Residual</TableHead>
                <TableHead>Owner</TableHead>
                <TableHead>Controls</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow><TableCell colSpan={9} className="text-center py-8 text-muted-foreground">Loading...</TableCell></TableRow>
              ) : risks.length === 0 ? (
                <TableRow><TableCell colSpan={9} className="text-center py-8 text-muted-foreground">No risks found</TableCell></TableRow>
              ) : (
                risks.map((risk) => (
                  <TableRow key={risk.id}>
                    <TableCell className="font-mono text-xs">{risk.identifier}</TableCell>
                    <TableCell>
                      <Link href={`/risks/${risk.id}`} className="text-primary hover:underline font-medium">
                        {risk.title}
                      </Link>
                      {risk.appetite_breached && (
                        <ShieldAlert className="inline ml-1 h-3.5 w-3.5 text-red-500" />
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {RISK_CATEGORY_LABELS[risk.category] || risk.category}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${RISK_STATUS_COLORS[risk.status]}`}>
                        {RISK_STATUS_LABELS[risk.status]}
                      </span>
                    </TableCell>
                    <TableCell>
                      {risk.inherent_score ? (
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[risk.inherent_score.severity]}`}>
                          {risk.inherent_score.score}
                        </span>
                      ) : '—'}
                    </TableCell>
                    <TableCell>
                      {risk.residual_score ? (
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_COLORS[risk.residual_score.severity]}`}>
                          {risk.residual_score.score}
                        </span>
                      ) : '—'}
                    </TableCell>
                    <TableCell className="text-sm">{risk.owner?.name || '—'}</TableCell>
                    <TableCell className="text-sm">{risk.linked_controls_count ?? 0}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Link href={`/risks/${risk.id}`}>
                          <Button variant="ghost" size="icon" className="h-8 w-8"><Eye className="h-4 w-4" /></Button>
                        </Link>
                        {canArchive && risk.status !== 'archived' && (
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-red-500" onClick={() => handleArchive(risk.id)}>
                            <Archive className="h-4 w-4" />
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
    </div>
  );
}
